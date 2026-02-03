package cleaner

import (
	"regexp"
	"strings"

	"github.com/example/article-extractor/internal/dom"
	"github.com/PuerkitoBio/goquery"
)

// Attributes to keep on elements.
var allowedAttributes = map[string][]string{
	"a":   {"href", "title"},
	"img": {"src", "alt", "title", "width", "height"},
	"*":   {}, // Remove all attributes from other elements
}

// Tags to preserve in output.
var preserveTags = map[string]bool{
	"p":          true,
	"a":          true,
	"img":        true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"ul":         true,
	"ol":         true,
	"li":         true,
	"blockquote": true,
	"pre":        true,
	"code":       true,
	"em":         true,
	"strong":     true,
	"b":          true,
	"i":          true,
	"br":         true,
	"figure":     true,
	"figcaption": true,
	"table":      true,
	"thead":      true,
	"tbody":      true,
	"tr":         true,
	"th":         true,
	"td":         true,
}

// Postprocess cleans up the extracted content.
func Postprocess(sel *goquery.Selection) {
	// Remove unwanted elements
	RemoveUnwantedFromContent(sel)

	// Clean attributes
	CleanAttributes(sel)

	// Remove empty elements
	RemoveEmptyElements(sel)

	// Normalize whitespace (done at text extraction level)
}

// RemoveUnwantedFromContent removes ads, social widgets, etc. from content.
func RemoveUnwantedFromContent(sel *goquery.Selection) {
	// Remove scripts and styles that might have survived
	sel.Find("script, style, noscript").Remove()

	// Remove elements with certain classes/ids
	unwantedPatterns := []string{
		"share",
		"social",
		"comment",
		"related",
		"recommend",
		"newsletter",
		"subscribe",
		"promo",
		"ad-",
		"advertisement",
	}

	for _, pattern := range unwantedPatterns {
		sel.Find("[class*='" + pattern + "'], [id*='" + pattern + "']").Remove()
	}

	// Remove empty links
	sel.Find("a").Each(func(_ int, a *goquery.Selection) {
		text := strings.TrimSpace(a.Text())
		if text == "" && a.Find("img").Length() == 0 {
			a.Remove()
		}
	})
}

// CleanAttributes removes unnecessary attributes from elements.
func CleanAttributes(sel *goquery.Selection) {
	sel.Find("*").Each(func(_ int, el *goquery.Selection) {
		tag := dom.GetTagName(el)
		allowedForTag := allowedAttributes[tag]
		allowedForAll := allowedAttributes["*"]

		// Get all attributes
		if len(el.Nodes) == 0 {
			return
		}

		node := el.Nodes[0]
		var attrsToRemove []string

		for _, attr := range node.Attr {
			// Check if attribute is allowed
			allowed := false
			for _, a := range allowedForTag {
				if attr.Key == a {
					allowed = true
					break
				}
			}
			if !allowed {
				for _, a := range allowedForAll {
					if attr.Key == a {
						allowed = true
						break
					}
				}
			}

			if !allowed {
				attrsToRemove = append(attrsToRemove, attr.Key)
			}
		}

		// Remove disallowed attributes
		for _, attr := range attrsToRemove {
			el.RemoveAttr(attr)
		}
	})
}

// RemoveEmptyElements removes empty elements from the content.
func RemoveEmptyElements(sel *goquery.Selection) {
	// Iterate multiple times to handle nested empty elements
	for i := 0; i < 3; i++ {
		sel.Find("*").Each(func(_ int, el *goquery.Selection) {
			tag := dom.GetTagName(el)

			// Skip self-closing elements
			if tag == "br" || tag == "hr" || tag == "img" {
				return
			}

			// Check if empty
			text := strings.TrimSpace(el.Text())
			html, _ := el.Html()
			html = strings.TrimSpace(html)

			// Remove if both text and html are empty
			if text == "" && (html == "" || isOnlyWhitespace(html)) {
				el.Remove()
			}
		})
	}
}

// isOnlyWhitespace checks if HTML contains only whitespace and empty tags.
func isOnlyWhitespace(html string) bool {
	// Remove all tags
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	text := tagRegex.ReplaceAllString(html, "")
	return strings.TrimSpace(text) == ""
}

// ConvertRelativeURLs converts relative URLs to absolute.
func ConvertRelativeURLs(sel *goquery.Selection, baseURL string) {
	if baseURL == "" {
		return
	}

	// Convert href in links
	sel.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
		href, _ := a.Attr("href")
		if href != "" && !isAbsoluteURL(href) {
			a.SetAttr("href", resolveURL(baseURL, href))
		}
	})

	// Convert src in images
	sel.Find("img[src]").Each(func(_ int, img *goquery.Selection) {
		src, _ := img.Attr("src")
		if src != "" && !isAbsoluteURL(src) {
			img.SetAttr("src", resolveURL(baseURL, src))
		}
	})
}

// isAbsoluteURL checks if a URL is absolute.
func isAbsoluteURL(url string) bool {
	return strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "//")
}

// resolveURL resolves a relative URL against a base URL.
func resolveURL(base, relative string) string {
	// Handle protocol-relative URLs
	if strings.HasPrefix(relative, "//") {
		if strings.HasPrefix(base, "https:") {
			return "https:" + relative
		}
		return "http:" + relative
	}

	// Handle root-relative URLs
	if strings.HasPrefix(relative, "/") {
		// Extract protocol and host from base
		idx := strings.Index(base, "://")
		if idx != -1 {
			hostEnd := strings.Index(base[idx+3:], "/")
			if hostEnd != -1 {
				return base[:idx+3+hostEnd] + relative
			}
			return base + relative
		}
		return relative
	}

	// Handle relative URLs
	idx := strings.LastIndex(base, "/")
	if idx != -1 {
		return base[:idx+1] + relative
	}
	return base + "/" + relative
}

// GetCleanHTML returns cleaned HTML content.
func GetCleanHTML(sel *goquery.Selection) string {
	html, _ := sel.Html()
	return strings.TrimSpace(html)
}

// GetCleanText returns cleaned text content.
func GetCleanText(sel *goquery.Selection) string {
	return dom.NormalizeTextPreserveNewlines(sel.Text())
}
