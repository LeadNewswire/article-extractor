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

	// Remove known non-content widgets (AI widgets, chatbots, etc.)
	RemoveKnownWidgets(doc)

	// Strip unlikely candidates
	StripUnlikelyCandidates(doc)

	// Convert data-articlebody content to paragraphs (for sites like Times of India)
	ConvertDataArticleBodyToParagraphs(doc)

	// Convert divs to paragraphs where appropriate
	ConvertToParagraphs(doc)
}

// ConvertDataArticleBodyToParagraphs converts [data-articlebody] elements to proper article structure.
// This handles sites like Times of India that use data attributes for article content.
// Times of India places text content as direct text nodes in divs, not wrapped in <p> tags.
func ConvertDataArticleBodyToParagraphs(doc *goquery.Document) {
	doc.Find("[data-articlebody]").Each(func(_ int, articleBody *goquery.Selection) {
		text := strings.TrimSpace(articleBody.Text())
		html, _ := articleBody.Html()

		// If it has substantial content but no <p> tags, we need to convert
		if len(text) > 200 && !strings.Contains(html, "<p") {
			// Find the main content wrapper (commonly has class like _s30J for TOI)
			// Look for a div that contains direct text nodes
			articleBody.Find("div").Each(func(_ int, div *goquery.Selection) {
				// Check if this div has direct text node children
				hasDirectText := false
				var textParts []string

				div.Contents().Each(func(_ int, content *goquery.Selection) {
					nodeName := goquery.NodeName(content)
					if nodeName == "#text" {
						textContent := strings.TrimSpace(content.Text())
						if len(textContent) > 30 {
							hasDirectText = true
							textParts = append(textParts, textContent)
						}
					}
				})

				// If this div has significant direct text, convert text nodes to paragraphs
				if hasDirectText && len(textParts) >= 3 {
					// Build new HTML with text wrapped in paragraphs
					var newHTML strings.Builder

					div.Contents().Each(func(_ int, content *goquery.Selection) {
						nodeName := goquery.NodeName(content)
						if nodeName == "#text" {
							textContent := strings.TrimSpace(content.Text())
							if len(textContent) > 30 {
								newHTML.WriteString("<p>")
								// Escape HTML entities
								escaped := strings.ReplaceAll(textContent, "&", "&amp;")
								escaped = strings.ReplaceAll(escaped, "<", "&lt;")
								escaped = strings.ReplaceAll(escaped, ">", "&gt;")
								newHTML.WriteString(escaped)
								newHTML.WriteString("</p>\n")
							}
						} else {
							// Keep non-text nodes (like h2 for headings)
							childHTML, _ := goquery.OuterHtml(content)
							// Skip ad-related divs
							class, _ := content.Attr("class")
							if !strings.Contains(class, "taboola") &&
								!strings.Contains(class, "trc_") &&
								!strings.Contains(class, "_ad") &&
								!strings.Contains(class, "mgid") {
								newHTML.WriteString(childHTML)
								newHTML.WriteString("\n")
							}
						}
					})

					if newHTML.Len() > 0 {
						div.SetHtml(newHTML.String())
					}
				}
			})
		}
	})
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

// widgetClassPatterns are CSS class patterns that identify non-content widgets.
// These patterns are designed to match widget-specific class names, not article tags.
// Use word boundaries to avoid matching partial words like "tag-chatbot-controversy".
var widgetClassPatterns = regexp.MustCompile(`(?i)(^|[\s"'-])(dd-widget|deeperdive|ai-widget|chatbot-widget|ask-ai|genai-widget|ai-assistant|ai-answer|ai-summary)([\s"'-]|$)`)

// widgetExactClasses are exact class names that indicate widgets.
var widgetExactClasses = []string{
	"dd-widget-wrapper",
	"dd-widget-input-wrapper",
	"deeperdive-widget",
	"ai-chatbot",
	"chatbot-container",
	"ai-assistant-widget",
}

// RemoveKnownWidgets removes known non-content widgets like AI assistants, chatbots, etc.
// These are removed unconditionally because they never contain article content.
func RemoveKnownWidgets(doc *goquery.Document) {
	// First, remove elements with exact widget class names
	for _, className := range widgetExactClasses {
		doc.Find("." + className).Remove()
	}

	// Then, check for widget patterns but be careful not to remove article elements
	doc.Find("*").Each(func(_ int, sel *goquery.Selection) {
		// Don't remove article, main, or body elements
		tag := dom.GetTagName(sel)
		if tag == "article" || tag == "main" || tag == "body" || tag == "html" {
			return
		}

		class := dom.GetAttribute(sel, "class")
		id := dom.GetAttribute(sel, "id")

		// Check for exact matches in class list
		for _, cls := range widgetExactClasses {
			if strings.Contains(" "+class+" ", " "+cls+" ") {
				sel.Remove()
				return
			}
		}

		// Check pattern matches on id only (more restrictive)
		if id != "" && widgetClassPatterns.MatchString(id) {
			sel.Remove()
			return
		}
	})
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
