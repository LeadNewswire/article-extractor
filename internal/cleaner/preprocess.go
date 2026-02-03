package cleaner

import (
	"regexp"
	"strings"

	"github.com/LeadNewswire/article-extractor/internal/dom"
	"github.com/LeadNewswire/article-extractor/internal/keywords"
	"github.com/PuerkitoBio/goquery"
)

// Tags that should be removed during preprocessing.
var removeTagsList = []string{
	"script",
	"style",
	"noscript",
	"iframe",
	"object",
	"embed",
	"applet",
	"link",
	"meta",
}

// Tags that are unlikely to contain content.
var unlikelyTags = []string{
	"footer",
	"header",
	"nav",
	"aside",
	"menu",
	"menuitem",
}

// Regex for checking visibility styles.
var hiddenStyleRegex = regexp.MustCompile(`(?i)(display\s*:\s*none|visibility\s*:\s*hidden)`)

// Preprocess performs initial cleanup on the document.
func Preprocess(doc *goquery.Document) {
	// Remove script, style, and other non-content tags
	RemoveUnwantedTags(doc)

	// Remove hidden elements
	RemoveHiddenElements(doc)

	// Strip unlikely candidates
	StripUnlikelyCandidates(doc)

	// Convert divs to paragraphs where appropriate
	ConvertToParagraphs(doc)
}

// RemoveUnwantedTags removes script, style, and other non-content tags.
func RemoveUnwantedTags(doc *goquery.Document) {
	for _, tag := range removeTagsList {
		doc.Find(tag).Remove()
	}
}

// RemoveHiddenElements removes elements that are hidden via CSS.
func RemoveHiddenElements(doc *goquery.Document) {
	doc.Find("[style]").Each(func(_ int, sel *goquery.Selection) {
		style, _ := sel.Attr("style")
		if hiddenStyleRegex.MatchString(style) {
			sel.Remove()
		}
	})

	// Remove elements with hidden attribute
	doc.Find("[hidden]").Remove()

	// Remove aria-hidden elements
	doc.Find("[aria-hidden='true']").Remove()
}

// StripUnlikelyCandidates removes elements unlikely to contain content.
func StripUnlikelyCandidates(doc *goquery.Document) {
	// Remove unlikely tags
	for _, tag := range unlikelyTags {
		doc.Find(tag).Each(func(_ int, sel *goquery.Selection) {
			// Check if it has a positive class/id that might indicate content
			class := dom.GetAttribute(sel, "class")
			id := dom.GetAttribute(sel, "id")

			if keywords.IsWhitelisted(class) || keywords.IsWhitelisted(id) {
				return // Keep this element
			}

			sel.Remove()
		})
	}

	// Remove elements with negative class/id patterns
	doc.Find("*").Each(func(_ int, sel *goquery.Selection) {
		// Don't remove body, html, or article-like elements
		tag := dom.GetTagName(sel)
		if tag == "body" || tag == "html" || tag == "article" || tag == "main" {
			return
		}

		class := dom.GetAttribute(sel, "class")
		id := dom.GetAttribute(sel, "id")
		combined := class + " " + id

		// Skip if whitelisted
		if keywords.IsWhitelisted(combined) {
			return
		}

		// Remove if blacklisted and not containing much text
		if keywords.IsBlacklisted(combined) {
			textLen := dom.GetTextLength(sel)
			linkDensity := dom.CalculateLinkDensity(sel)

			// Remove if short text or high link density
			if textLen < 200 || linkDensity > 0.5 {
				sel.Remove()
			}
		}
	})
}

// ConvertToParagraphs converts div and span elements that look like paragraphs.
func ConvertToParagraphs(doc *goquery.Document) {
	// Find divs that have no block-level children
	doc.Find("div, span").Each(func(_ int, sel *goquery.Selection) {
		if !hasBlockChild(sel) {
			// Convert to p if it has meaningful text
			text := dom.GetText(sel)
			if len(text) > 0 {
				convertToP(sel)
			}
		}
	})

	// Handle br-separated content in divs
	doc.Find("div").Each(func(_ int, sel *goquery.Selection) {
		html, _ := sel.Html()
		if strings.Contains(html, "<br") {
			replaceBrWithP(sel)
		}
	})
}

// hasBlockChild checks if a selection has any block-level children.
func hasBlockChild(sel *goquery.Selection) bool {
	hasBlock := false
	sel.Children().Each(func(_ int, child *goquery.Selection) {
		tag := dom.GetTagName(child)
		if dom.IsBlockElement(tag) {
			hasBlock = true
		}
	})
	return hasBlock
}

// convertToP converts an element to a paragraph.
func convertToP(sel *goquery.Selection) {
	// This is a simplified conversion - in goquery we can't easily change tag names
	// Instead, we wrap the content in a p tag
	html, _ := sel.Html()
	if html != "" {
		sel.SetHtml("<p>" + html + "</p>")
	}
}

// replaceBrWithP replaces br-separated text with paragraph elements.
func replaceBrWithP(sel *goquery.Selection) {
	html, _ := sel.Html()

	// Split by br tags
	brRegex := regexp.MustCompile(`<br\s*/?>`)
	parts := brRegex.Split(html, -1)

	if len(parts) <= 1 {
		return
	}

	var newHTML strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			newHTML.WriteString("<p>")
			newHTML.WriteString(part)
			newHTML.WriteString("</p>")
		}
	}

	if newHTML.Len() > 0 {
		sel.SetHtml(newHTML.String())
	}
}

// RemoveEmptyParagraphs removes empty p elements.
func RemoveEmptyParagraphs(doc *goquery.Document) {
	doc.Find("p").Each(func(_ int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		if text == "" {
			sel.Remove()
		}
	})
}

// CleanExcessiveWhitespace removes excessive whitespace in text nodes.
func CleanExcessiveWhitespace(doc *goquery.Document) {
	// This is handled during text extraction, not in DOM manipulation
}
