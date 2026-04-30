package metadata

import (
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractSiteName extracts the publishing site / media name from a document.
//
// The order mirrors how publishers most reliably advertise their canonical
// brand name:
//  1. <meta property="og:site_name">          — Open Graph standard
//  2. JSON-LD publisher.name                  — schema.org NewsArticle / Article
//  3. <meta name="application-name">          — PWA / Microsoft tile metadata
//  4. <meta name="twitter:site">              — Twitter handle, with leading "@" stripped
//
// Returns empty string when none are present so callers can apply their own
// fallback (e.g. derive from URL host).
func ExtractSiteName(doc *goquery.Document) string {
	if name := getMetaContent(doc, "og:site_name"); name != "" {
		return cleanSiteName(name)
	}

	if name := getJSONLDPublisherName(doc); name != "" {
		return cleanSiteName(name)
	}

	if name := getMetaContent(doc, "application-name"); name != "" {
		return cleanSiteName(name)
	}

	if name := getMetaContent(doc, "twitter:site"); name != "" {
		name = strings.TrimPrefix(strings.TrimSpace(name), "@")
		if name != "" {
			return cleanSiteName(name)
		}
	}

	return ""
}

// getJSONLDPublisherName scans script[type='application/ld+json'] blocks for
// schema.org publisher.name. Each block is parsed permissively — both objects
// and arrays of objects are common in the wild.
func getJSONLDPublisherName(doc *goquery.Document) string {
	var found string

	doc.Find("script[type='application/ld+json']").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		text := strings.TrimSpace(sel.Text())
		if text == "" {
			return true
		}

		if name := scanJSONForPublisherName([]byte(text)); name != "" {
			found = name
			return false
		}
		return true
	})

	return found
}

// scanJSONForPublisherName decodes a JSON-LD payload (object or array) and
// walks it looking for a `publisher.name` string.
func scanJSONForPublisherName(raw []byte) string {
	var any interface{}
	if err := json.Unmarshal(raw, &any); err != nil {
		return ""
	}
	return walkForPublisherName(any)
}

func walkForPublisherName(node interface{}) string {
	switch v := node.(type) {
	case map[string]interface{}:
		if pub, ok := v["publisher"]; ok {
			if name := publisherName(pub); name != "" {
				return name
			}
		}
		for _, child := range v {
			if name := walkForPublisherName(child); name != "" {
				return name
			}
		}
	case []interface{}:
		for _, child := range v {
			if name := walkForPublisherName(child); name != "" {
				return name
			}
		}
	}
	return ""
}

func publisherName(pub interface{}) string {
	switch v := pub.(type) {
	case map[string]interface{}:
		if n, ok := v["name"].(string); ok {
			return strings.TrimSpace(n)
		}
	case []interface{}:
		for _, child := range v {
			if name := publisherName(child); name != "" {
				return name
			}
		}
	case string:
		return strings.TrimSpace(v)
	}
	return ""
}

// cleanSiteName collapses whitespace and trims. Site names are typically
// short and clean; we keep this minimal to avoid mangling legitimate names.
func cleanSiteName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	return strings.Join(strings.Fields(name), " ")
}
