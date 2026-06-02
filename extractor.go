package extractor

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/LeadNewswire/article-extractor/internal/cleaner"
	"github.com/LeadNewswire/article-extractor/internal/dom"
	"github.com/LeadNewswire/article-extractor/internal/fetcher"
	"github.com/LeadNewswire/article-extractor/internal/metadata"
	"github.com/LeadNewswire/article-extractor/internal/scorer"
)

// Extractor is the main article extraction engine.
type Extractor struct {
	config *Config
	client *fetcher.Client
}

// New creates a new Extractor with the given options.
func New(opts ...Option) *Extractor {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	client := fetcher.NewClient(
		config.HTTPTimeout,
		config.UserAgent,
		config.MaxContentLength,
	)

	return &Extractor{
		config: config,
		client: client,
	}
}

// Extract extracts an article from HTML content.
func (e *Extractor) Extract(html string) (*Article, error) {
	return e.ExtractWithURL(html, "")
}

// ExtractWithURL extracts an article from HTML content with a base URL.
func (e *Extractor) ExtractWithURL(html, baseURL string) (*Article, error) {
	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, NewExtractionError("parse", baseURL, ErrInvalidHTML)
	}

	return e.extractFromDocument(doc, baseURL)
}

// ExtractFromURL fetches and extracts an article from a URL.
func (e *Extractor) ExtractFromURL(ctx context.Context, url string) (*Article, error) {
	// Validate URL
	if !fetcher.IsValidURL(url) {
		url = fetcher.NormalizeURL(url)
		if !fetcher.IsValidURL(url) {
			return nil, NewExtractionError("validate", url, ErrInvalidURL)
		}
	}

	// Fetch HTML
	html, err := e.client.Fetch(ctx, url)
	if err != nil {
		return nil, NewExtractionError("fetch", url, err)
	}

	// Extract article
	article, err := e.ExtractWithURL(html, url)
	if err != nil {
		return nil, err
	}

	article.URL = url
	return article, nil
}

// extractFromDocument extracts an article from a goquery document.
func (e *Extractor) extractFromDocument(doc *goquery.Document, baseURL string) (*Article, error) {
	// Extract metadata first (before preprocessing removes elements)
	title := metadata.ExtractTitle(doc)
	author := metadata.ExtractAuthor(doc)
	siteName := metadata.ExtractSiteName(doc)
	publishedAt := metadata.ExtractDate(doc)
	leadImage := e.extractLeadImage(doc, baseURL)

	// Preprocess document
	cleaner.Preprocess(doc)

	// Score content
	s := scorer.NewScorer(
		e.config.MinParagraphLength,
		e.config.MinContentLength,
		e.config.Debug,
	)

	topCandidate, scoreMap := s.Score(doc)

	// Check if we found content
	if topCandidate == nil || topCandidate.Selection == nil {
		return nil, NewExtractionError("extract", baseURL, ErrNoContent)
	}

	// Get the content selection
	contentSel := topCandidate.Selection

	// Try to merge siblings
	contentSel = cleaner.MergeSiblings(
		contentSel,
		topCandidate.GetScore(),
		e.config.MinParagraphLength,
	)

	// Clone the content for cleaning
	contentClone := contentSel.Clone()

	// Postprocess content
	cleaner.Postprocess(contentClone)

	// Convert relative URLs if base URL provided
	if baseURL != "" {
		cleaner.ConvertRelativeURLs(contentClone, baseURL)
	}

	// Collect inline images BEFORE GetCleanHTML — we need <figure>/<figcaption>
	// structure preserved so captions can be paired with their images. URLs at
	// this point are already absolute (ConvertRelativeURLs ran above).
	inlineImages := extractInlineImages(contentClone)

	// Get cleaned HTML and text
	contentHTML := cleaner.GetCleanHTML(contentClone)
	textContent := cleaner.GetCleanText(contentClone)

	// Ancestor expansion: if the extracted content is suspiciously short, walk up the
	// DOM tree to find a larger container with more paragraph content. This handles
	// slideshow/listicle layouts where content is scattered across deeply nested sibling
	// containers that the scorer treats as separate low-scoring candidates.
	if len(textContent) < minExpandContentLength {
		expandedHTML, expandedText := e.tryAncestorExpansion(contentSel, baseURL)
		if len(expandedText) > len(textContent) {
			contentHTML = expandedHTML
			textContent = expandedText
		}
	}

	// Check content length
	if len(textContent) < e.config.MinContentLength {
		return nil, NewExtractionError("validate", baseURL, ErrContentTooShort)
	}

	// Calculate word count
	wordCount := dom.CountWords(textContent)

	// Calculate confidence based on score and content quality
	confidence := e.calculateConfidence(topCandidate, scoreMap, wordCount)

	// Generate excerpt
	excerpt := dom.GetExcerpt(textContent, 200)

	return &Article{
		Title:        title,
		Content:      contentHTML,
		TextContent:  textContent,
		Excerpt:      excerpt,
		Author:       author,
		SiteName:     siteName,
		PublishedAt:  publishedAt,
		LeadImage:    leadImage,
		InlineImages: inlineImages,
		URL:          baseURL,
		WordCount:    wordCount,
		Score:        topCandidate.GetScore(),
		Confidence:   confidence,
	}, nil
}

// minExpandContentLength is the text length threshold below which ancestor expansion
// is attempted. Articles shorter than this are likely truncated (e.g., only the intro
// of a slideshow was captured).
const minExpandContentLength = 1500

// maxExpansionLevels is how many parent levels to try during ancestor expansion.
const maxExpansionLevels = 5

// tryAncestorExpansion walks up the DOM from contentSel, cleans each ancestor,
// and returns the one that produces the most paragraph text.
func (e *Extractor) tryAncestorExpansion(contentSel *goquery.Selection, baseURL string) (string, string) {
	bestHTML := ""
	bestText := ""

	candidate := contentSel
	for i := 0; i < maxExpansionLevels; i++ {
		parent := candidate.Parent()
		if parent.Length() == 0 {
			break
		}

		tag := dom.GetTagName(parent)
		if tag == "body" || tag == "html" {
			break
		}

		// Skip parents with high link density (nav bars, footers)
		linkDensity := dom.CalculateLinkDensity(parent)
		if linkDensity > 0.3 {
			candidate = parent
			continue
		}

		// Extract text from <p>, <h2>, <h3> children only — this is more robust than
		// running full Postprocess, which can cascade-remove content in complex layouts.
		text := extractParagraphText(parent)
		if len(text) > len(bestText) {
			bestText = text
			bestHTML = "" // HTML not available in paragraph-only extraction
		}

		candidate = parent
	}

	return bestHTML, bestText
}

// extractParagraphText extracts text from paragraph and heading elements,
// producing clean article text without the aggressive cleanup of Postprocess.
func extractParagraphText(sel *goquery.Selection) string {
	var parts []string
	sel.Find("p, h2, h3, h4, li, blockquote").Each(func(_ int, el *goquery.Selection) {
		text := strings.TrimSpace(el.Text())
		if len(text) > 20 {
			parts = append(parts, text)
		}
	})
	return strings.Join(parts, "\n")
}

// extractLeadImage extracts the main image from the document.
func (e *Extractor) extractLeadImage(doc *goquery.Document, baseURL string) *Image {
	// Try og:image first
	ogImage := doc.Find("meta[property='og:image']").AttrOr("content", "")
	if ogImage != "" {
		img := &Image{URL: ogImage}

		// Try to get dimensions
		if width := doc.Find("meta[property='og:image:width']").AttrOr("content", ""); width != "" {
			img.Width = parseInt(width)
		}
		if height := doc.Find("meta[property='og:image:height']").AttrOr("content", ""); height != "" {
			img.Height = parseInt(height)
		}

		return img
	}

	// Try twitter:image
	twitterImage := doc.Find("meta[name='twitter:image']").AttrOr("content", "")
	if twitterImage != "" {
		return &Image{URL: twitterImage}
	}

	// Try to find a large image in article
	var leadImage *Image
	doc.Find("article img, .article img, .post img, main img").Each(func(_ int, sel *goquery.Selection) {
		if leadImage != nil {
			return
		}

		src := imageSrc(sel)
		if src == "" {
			return
		}

		img := &Image{
			URL:     src,
			Alt:     sel.AttrOr("alt", ""),
			Caption: inferCaption(sel),
		}

		// Get dimensions
		if width := sel.AttrOr("width", ""); width != "" {
			img.Width = parseInt(width)
		}
		if height := sel.AttrOr("height", ""); height != "" {
			img.Height = parseInt(height)
		}

		// Only use images that seem like article images
		if img.Width >= 200 || img.Height >= 200 || (img.Width == 0 && img.Height == 0) {
			leadImage = img
		}
	})

	return leadImage
}

// extractInlineImages collects every image embedded in the article body.
// Captions are paired from <figcaption> within the same <figure>, falling
// back to the WordPress / Substack convention of an italic sibling line
// (<em>, <i>, <small>, or <p class~="caption">) immediately after the <img>.
// URLs already absolute when the caller ran ConvertRelativeURLs on the
// selection. Output order matches document order; identical URLs are
// emitted only once.
func extractInlineImages(sel *goquery.Selection) []*Image {
	var out []*Image
	seen := make(map[string]bool)

	emit := func(img *Image) {
		if img == nil || img.URL == "" || seen[img.URL] {
			return
		}
		seen[img.URL] = true
		out = append(out, img)
	}

	sel.Find("img").Each(func(_ int, imgSel *goquery.Selection) {
		src := imageSrc(imgSel)
		if src == "" {
			return
		}

		img := &Image{
			URL: src,
			Alt: imgSel.AttrOr("alt", ""),
		}
		if w := imgSel.AttrOr("width", ""); w != "" {
			img.Width = parseInt(w)
		}
		if h := imgSel.AttrOr("height", ""); h != "" {
			img.Height = parseInt(h)
		}
		img.Caption = inferCaption(imgSel)
		emit(img)
	})

	return out
}

// imageSrc returns the best-effort source URL for an <img>, falling through
// common lazy-load attributes and the first entry of a srcset when src is
// missing.
func imageSrc(sel *goquery.Selection) string {
	for _, attr := range []string{"src", "data-src", "data-lazy-src", "data-original"} {
		if v, _ := sel.Attr(attr); strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	if srcset, _ := sel.Attr("srcset"); strings.TrimSpace(srcset) != "" {
		first := strings.SplitN(srcset, ",", 2)[0]
		fields := strings.Fields(strings.TrimSpace(first))
		if len(fields) > 0 {
			return fields[0]
		}
	}
	return ""
}

// inferCaption finds a visible caption for an <img>.
//
// Priority order:
//  1. <figcaption> within the nearest enclosing <figure>
//  2. The first italic-style sibling immediately following the image
//     (<em>, <i>, <small>) — the WordPress / Substack convention
//  3. A sibling element whose class name contains "caption"
//  4. A <p> sibling whose only child is <em>/<i> (e.g. <p><em>...</em></p>)
//
// Returns "" when nothing fits — callers fall back to alt text if they want.
func inferCaption(img *goquery.Selection) string {
	// 1. <figure><figcaption>
	if fig := img.Closest("figure"); fig.Length() > 0 {
		if cap := fig.Find("figcaption").First(); cap.Length() > 0 {
			if t := strings.TrimSpace(cap.Text()); t != "" {
				return t
			}
		}
	}

	// Walk forward through siblings until we hit a substantive paragraph
	// (which signals "back to article body, no caption"). We cap at two
	// siblings so we don't reach across the layout.
	next := img.Next()
	for i := 0; i < 2 && next.Length() > 0; i++ {
		tag := strings.ToLower(goquery.NodeName(next))
		text := strings.TrimSpace(next.Text())

		switch tag {
		case "em", "i", "small":
			if text != "" {
				return text
			}
		}

		// class="...caption..." on any wrapper
		if class, _ := next.Attr("class"); class != "" &&
			strings.Contains(strings.ToLower(class), "caption") && text != "" {
			return text
		}

		// <p><em>...</em></p> pattern
		if tag == "p" && text != "" {
			emText := strings.TrimSpace(next.Find("em, i").First().Text())
			if emText != "" && emText == text {
				return text
			}
			// A full-paragraph of prose breaks the caption hunt.
			break
		}

		next = next.Next()
	}

	return ""
}

// calculateConfidence calculates a confidence score for the extraction.
func (e *Extractor) calculateConfidence(topCandidate *scorer.NodeScore, scoreMap *scorer.ScoreMap, wordCount int) float64 {
	confidence := 0.0

	// Base confidence from score
	score := topCandidate.GetScore()
	if score > 100 {
		confidence += 0.4
	} else if score > 50 {
		confidence += 0.3
	} else if score > 20 {
		confidence += 0.2
	} else {
		confidence += 0.1
	}

	// Confidence from word count
	if wordCount > 500 {
		confidence += 0.3
	} else if wordCount > 200 {
		confidence += 0.2
	} else if wordCount > 100 {
		confidence += 0.1
	}

	// Confidence from link density
	if topCandidate.LinkDensity < 0.1 {
		confidence += 0.2
	} else if topCandidate.LinkDensity < 0.2 {
		confidence += 0.1
	}

	// Confidence from score dominance
	candidates := scoreMap.GetCandidatesByScore()
	if len(candidates) >= 2 {
		secondScore := candidates[1].GetScore()
		if score > secondScore*2 {
			confidence += 0.1
		}
	}

	// Clamp to [0, 1]
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// parseInt parses a string to int, returning 0 on error.
func parseInt(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	return n
}
