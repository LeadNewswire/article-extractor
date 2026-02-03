package metadata

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// TitleSeparators are common title separators.
var titleSeparators = []string{" | ", " - ", " :: ", " / ", " » ", " — ", " · "}

// ExtractTitle extracts the article title from a document.
func ExtractTitle(doc *goquery.Document) string {
	// Try og:title first
	if title := getMetaContent(doc, "og:title"); title != "" {
		return cleanTitle(title)
	}

	// Try twitter:title
	if title := getMetaContent(doc, "twitter:title"); title != "" {
		return cleanTitle(title)
	}

	// Try schema.org headline
	if title := getSchemaHeadline(doc); title != "" {
		return cleanTitle(title)
	}

	// Try h1 in article
	if title := getArticleH1(doc); title != "" {
		return cleanTitle(title)
	}

	// Try first h1
	if title := getFirstH1(doc); title != "" {
		return cleanTitle(title)
	}

	// Try title tag
	if title := getTitleTag(doc); title != "" {
		return cleanTitle(title)
	}

	return ""
}

// getMetaContent gets content from a meta tag.
func getMetaContent(doc *goquery.Document, property string) string {
	var content string

	// Try property attribute (Open Graph)
	doc.Find("meta[property='" + property + "']").Each(func(_ int, sel *goquery.Selection) {
		if c, exists := sel.Attr("content"); exists && c != "" {
			content = c
		}
	})

	if content != "" {
		return content
	}

	// Try name attribute
	doc.Find("meta[name='" + property + "']").Each(func(_ int, sel *goquery.Selection) {
		if c, exists := sel.Attr("content"); exists && c != "" {
			content = c
		}
	})

	return content
}

// getSchemaHeadline gets headline from schema.org markup.
func getSchemaHeadline(doc *goquery.Document) string {
	var headline string

	// Try JSON-LD
	doc.Find("script[type='application/ld+json']").Each(func(_ int, sel *goquery.Selection) {
		if headline != "" {
			return
		}
		text := sel.Text()
		// Simple extraction - look for "headline"
		if idx := strings.Index(text, `"headline"`); idx != -1 {
			// Find the value
			rest := text[idx+len(`"headline"`):]
			if colonIdx := strings.Index(rest, ":"); colonIdx != -1 {
				rest = rest[colonIdx+1:]
				rest = strings.TrimSpace(rest)
				if len(rest) > 0 && rest[0] == '"' {
					rest = rest[1:]
					if endIdx := strings.Index(rest, `"`); endIdx != -1 {
						headline = rest[:endIdx]
					}
				}
			}
		}
	})

	if headline != "" {
		return headline
	}

	// Try itemprop
	doc.Find("[itemprop='headline']").Each(func(_ int, sel *goquery.Selection) {
		if headline == "" {
			headline = strings.TrimSpace(sel.Text())
		}
	})

	return headline
}

// getArticleH1 gets h1 from within an article element.
func getArticleH1(doc *goquery.Document) string {
	var title string

	doc.Find("article h1, [role='article'] h1, .article h1, .post h1").Each(func(_ int, sel *goquery.Selection) {
		if title == "" {
			title = strings.TrimSpace(sel.Text())
		}
	})

	return title
}

// getFirstH1 gets the first h1 element.
func getFirstH1(doc *goquery.Document) string {
	var title string

	doc.Find("h1").Each(func(_ int, sel *goquery.Selection) {
		if title == "" {
			title = strings.TrimSpace(sel.Text())
		}
	})

	return title
}

// getTitleTag gets the title from the title tag.
func getTitleTag(doc *goquery.Document) string {
	return strings.TrimSpace(doc.Find("title").Text())
}

// cleanTitle cleans up a title string.
func cleanTitle(title string) string {
	// Trim whitespace
	title = strings.TrimSpace(title)

	// Normalize whitespace
	whitespaceRegex := regexp.MustCompile(`\s+`)
	title = whitespaceRegex.ReplaceAllString(title, " ")

	// Try to remove site name suffix
	for _, sep := range titleSeparators {
		if idx := strings.LastIndex(title, sep); idx != -1 {
			// Keep the longer part
			before := title[:idx]
			after := title[idx+len(sep):]

			// Usually the article title is the longer part
			if len(before) > len(after)*2 {
				title = strings.TrimSpace(before)
			} else if len(after) > len(before)*2 {
				title = strings.TrimSpace(after)
			}
			// If similar lengths, keep the original
		}
	}

	return title
}

// ExtractTitleWithFallback extracts title with a fallback.
func ExtractTitleWithFallback(doc *goquery.Document, fallback string) string {
	title := ExtractTitle(doc)
	if title == "" {
		return fallback
	}
	return title
}
