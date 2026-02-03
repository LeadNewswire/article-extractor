package extractor

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/example/article-extractor/internal/cleaner"
	"github.com/example/article-extractor/internal/dom"
	"github.com/example/article-extractor/internal/fetcher"
	"github.com/example/article-extractor/internal/metadata"
	"github.com/example/article-extractor/internal/scorer"
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

	// Get cleaned HTML and text
	contentHTML := cleaner.GetCleanHTML(contentClone)
	textContent := cleaner.GetCleanText(contentClone)

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
		Title:       title,
		Content:     contentHTML,
		TextContent: textContent,
		Excerpt:     excerpt,
		Author:      author,
		PublishedAt: publishedAt,
		LeadImage:   leadImage,
		URL:         baseURL,
		WordCount:   wordCount,
		Score:       topCandidate.GetScore(),
		Confidence:  confidence,
	}, nil
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

		src, _ := sel.Attr("src")
		if src == "" {
			// Try data-src for lazy-loaded images
			src, _ = sel.Attr("data-src")
		}
		if src == "" {
			return
		}

		img := &Image{
			URL: src,
			Alt: sel.AttrOr("alt", ""),
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
