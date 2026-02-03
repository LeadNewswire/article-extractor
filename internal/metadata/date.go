package metadata

import (
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Common date formats to try.
var dateFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05-07:00",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"January 2, 2006",
	"Jan 2, 2006",
	"02 January 2006",
	"02 Jan 2006",
	"2006/01/02",
	"01/02/2006",
	"02/01/2006",
}

// ExtractDate extracts the publication date from a document.
func ExtractDate(doc *goquery.Document) *time.Time {
	// Try meta article:published_time (Open Graph)
	if date := parseMetaDate(doc, "article:published_time"); date != nil {
		return date
	}

	// Try meta datePublished
	if date := parseMetaDate(doc, "datePublished"); date != nil {
		return date
	}

	// Try meta date
	if date := parseMetaDate(doc, "date"); date != nil {
		return date
	}

	// Try meta DC.date
	if date := parseMetaDate(doc, "DC.date"); date != nil {
		return date
	}

	// Try schema.org datePublished
	if date := getSchemaDate(doc); date != nil {
		return date
	}

	// Try time element
	if date := getTimeElement(doc); date != nil {
		return date
	}

	// Try common date selectors
	if date := getDateBySelector(doc); date != nil {
		return date
	}

	return nil
}

// parseMetaDate parses a date from a meta tag.
func parseMetaDate(doc *goquery.Document, property string) *time.Time {
	content := getMetaContent(doc, property)
	if content == "" {
		return nil
	}
	return parseDate(content)
}

// getSchemaDate gets date from schema.org markup.
func getSchemaDate(doc *goquery.Document) *time.Time {
	var dateStr string

	// Try JSON-LD
	doc.Find("script[type='application/ld+json']").Each(func(_ int, sel *goquery.Selection) {
		if dateStr != "" {
			return
		}
		text := sel.Text()

		// Look for datePublished
		patterns := []string{`"datePublished"`, `"dateCreated"`}
		for _, pattern := range patterns {
			if idx := strings.Index(text, pattern); idx != -1 {
				rest := text[idx+len(pattern):]
				if colonIdx := strings.Index(rest, ":"); colonIdx != -1 {
					valueRest := strings.TrimSpace(rest[colonIdx+1:])
					if len(valueRest) > 0 && valueRest[0] == '"' {
						valueRest = valueRest[1:]
						if endIdx := strings.Index(valueRest, `"`); endIdx != -1 {
							dateStr = valueRest[:endIdx]
							return
						}
					}
				}
			}
		}
	})

	if dateStr != "" {
		return parseDate(dateStr)
	}

	// Try itemprop datePublished
	doc.Find("[itemprop='datePublished']").Each(func(_ int, sel *goquery.Selection) {
		if dateStr == "" {
			// Check content attribute first
			if content, exists := sel.Attr("content"); exists {
				dateStr = content
			} else if datetime, exists := sel.Attr("datetime"); exists {
				dateStr = datetime
			} else {
				dateStr = strings.TrimSpace(sel.Text())
			}
		}
	})

	if dateStr != "" {
		return parseDate(dateStr)
	}

	return nil
}

// getTimeElement gets date from time elements.
func getTimeElement(doc *goquery.Document) *time.Time {
	var date *time.Time

	doc.Find("time[datetime]").Each(func(_ int, sel *goquery.Selection) {
		if date != nil {
			return
		}
		datetime, _ := sel.Attr("datetime")
		if datetime != "" {
			date = parseDate(datetime)
		}
	})

	return date
}

// getDateBySelector tries common date CSS selectors.
func getDateBySelector(doc *goquery.Document) *time.Time {
	selectors := []string{
		".post-date",
		".entry-date",
		".article-date",
		".published-date",
		".publish-date",
		".date-published",
		".meta-date",
		".timestamp",
		"[class*='date']",
	}

	for _, selector := range selectors {
		sel := doc.Find(selector)
		if sel.Length() > 0 {
			text := strings.TrimSpace(sel.First().Text())
			if date := parseDate(text); date != nil {
				return date
			}
		}
	}

	return nil
}

// parseDate attempts to parse a date string in various formats.
func parseDate(dateStr string) *time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return nil
	}

	// Try all known formats
	for _, format := range dateFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return &t
		}
	}

	// Try with timezone stripping
	noTZ := stripTimezone(dateStr)
	for _, format := range dateFormats {
		if t, err := time.Parse(format, noTZ); err == nil {
			return &t
		}
	}

	return nil
}

// stripTimezone removes timezone information for simpler parsing.
func stripTimezone(dateStr string) string {
	// Remove trailing timezone like "EST", "PST", etc.
	tzRegex := regexp.MustCompile(`\s+[A-Z]{3,4}$`)
	return tzRegex.ReplaceAllString(dateStr, "")
}

// ExtractModifiedDate extracts the last modified date.
func ExtractModifiedDate(doc *goquery.Document) *time.Time {
	// Try meta article:modified_time
	if date := parseMetaDate(doc, "article:modified_time"); date != nil {
		return date
	}

	// Try meta dateModified
	if date := parseMetaDate(doc, "dateModified"); date != nil {
		return date
	}

	// Try schema.org dateModified
	var dateStr string
	doc.Find("[itemprop='dateModified']").Each(func(_ int, sel *goquery.Selection) {
		if dateStr == "" {
			if content, exists := sel.Attr("content"); exists {
				dateStr = content
			} else if datetime, exists := sel.Attr("datetime"); exists {
				dateStr = datetime
			}
		}
	})

	if dateStr != "" {
		return parseDate(dateStr)
	}

	return nil
}

// FormatDate formats a date for display.
func FormatDate(t *time.Time, format string) string {
	if t == nil {
		return ""
	}
	return t.Format(format)
}
