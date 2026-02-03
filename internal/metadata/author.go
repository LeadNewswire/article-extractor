package metadata

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Author extraction patterns.
var (
	bylineRegex    = regexp.MustCompile(`(?i)^[\s]*by[\s]+(.+?)[\s]*$`)
	authorPrefixes = []string{"By ", "by ", "BY ", "Written by ", "Author: ", "Posted by "}
)

// ExtractAuthor extracts the article author from a document.
func ExtractAuthor(doc *goquery.Document) string {
	// Try meta author
	if author := getMetaContent(doc, "author"); author != "" {
		return cleanAuthor(author)
	}

	// Try og:article:author
	if author := getMetaContent(doc, "article:author"); author != "" {
		return cleanAuthor(author)
	}

	// Try schema.org author
	if author := getSchemaAuthor(doc); author != "" {
		return cleanAuthor(author)
	}

	// Try common author selectors
	if author := getAuthorBySelector(doc); author != "" {
		return cleanAuthor(author)
	}

	// Try byline patterns
	if author := getAuthorByByline(doc); author != "" {
		return cleanAuthor(author)
	}

	return ""
}

// getSchemaAuthor gets author from schema.org markup.
func getSchemaAuthor(doc *goquery.Document) string {
	var author string

	// Try JSON-LD
	doc.Find("script[type='application/ld+json']").Each(func(_ int, sel *goquery.Selection) {
		if author != "" {
			return
		}
		text := sel.Text()

		// Look for author name in various formats
		patterns := []string{`"author"`, `"creator"`}
		for _, pattern := range patterns {
			if idx := strings.Index(text, pattern); idx != -1 {
				rest := text[idx:]
				// Try to find name field
				if nameIdx := strings.Index(rest, `"name"`); nameIdx != -1 {
					nameRest := rest[nameIdx+len(`"name"`):]
					if colonIdx := strings.Index(nameRest, ":"); colonIdx != -1 {
						valueRest := strings.TrimSpace(nameRest[colonIdx+1:])
						if len(valueRest) > 0 && valueRest[0] == '"' {
							valueRest = valueRest[1:]
							if endIdx := strings.Index(valueRest, `"`); endIdx != -1 {
								author = valueRest[:endIdx]
								return
							}
						}
					}
				}
			}
		}
	})

	if author != "" {
		return author
	}

	// Try itemprop author
	doc.Find("[itemprop='author']").Each(func(_ int, sel *goquery.Selection) {
		if author == "" {
			// Check for nested name
			if nameSel := sel.Find("[itemprop='name']"); nameSel.Length() > 0 {
				author = strings.TrimSpace(nameSel.Text())
			} else {
				author = strings.TrimSpace(sel.Text())
			}
		}
	})

	return author
}

// getAuthorBySelector tries common author CSS selectors.
func getAuthorBySelector(doc *goquery.Document) string {
	selectors := []string{
		".author-name",
		".author",
		".byline-name",
		".byline__name",
		"[rel='author']",
		".entry-author-name",
		".post-author-name",
		".article-author",
		".article__author",
		".author__name",
		"a.author",
		"span.author",
	}

	for _, selector := range selectors {
		sel := doc.Find(selector)
		if sel.Length() > 0 {
			text := strings.TrimSpace(sel.First().Text())
			if text != "" && len(text) < 100 { // Sanity check
				return text
			}
		}
	}

	return ""
}

// getAuthorByByline looks for byline patterns.
func getAuthorByByline(doc *goquery.Document) string {
	bylineSelectors := []string{
		".byline",
		".by-line",
		".post-byline",
		".article-byline",
		".meta-author",
	}

	for _, selector := range bylineSelectors {
		sel := doc.Find(selector)
		if sel.Length() > 0 {
			text := strings.TrimSpace(sel.First().Text())
			if author := extractAuthorFromByline(text); author != "" {
				return author
			}
		}
	}

	return ""
}

// extractAuthorFromByline extracts author name from a byline string.
func extractAuthorFromByline(byline string) string {
	// Try regex pattern
	if matches := bylineRegex.FindStringSubmatch(byline); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try removing common prefixes
	for _, prefix := range authorPrefixes {
		if strings.HasPrefix(byline, prefix) {
			author := strings.TrimPrefix(byline, prefix)
			// Clean up - stop at common delimiters
			for _, delim := range []string{",", "|", "·", " on ", " - "} {
				if idx := strings.Index(author, delim); idx != -1 {
					author = author[:idx]
				}
			}
			return strings.TrimSpace(author)
		}
	}

	return ""
}

// cleanAuthor cleans up an author string.
func cleanAuthor(author string) string {
	// Trim whitespace
	author = strings.TrimSpace(author)

	// Normalize whitespace
	whitespaceRegex := regexp.MustCompile(`\s+`)
	author = whitespaceRegex.ReplaceAllString(author, " ")

	// Remove "By " prefix if present
	for _, prefix := range authorPrefixes {
		if strings.HasPrefix(author, prefix) {
			author = strings.TrimPrefix(author, prefix)
			break
		}
	}

	return strings.TrimSpace(author)
}

// ExtractAuthors extracts multiple authors if present.
func ExtractAuthors(doc *goquery.Document) []string {
	author := ExtractAuthor(doc)
	if author == "" {
		return nil
	}

	// Check for multiple authors
	separators := []string{" and ", ", ", " & "}
	for _, sep := range separators {
		if strings.Contains(author, sep) {
			parts := strings.Split(author, sep)
			var authors []string
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					authors = append(authors, part)
				}
			}
			return authors
		}
	}

	return []string{author}
}
