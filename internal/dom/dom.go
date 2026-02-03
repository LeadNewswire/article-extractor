package dom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Document wraps a goquery document with additional functionality.
type Document struct {
	*goquery.Document
}

// NewDocument creates a new Document from HTML string.
func NewDocument(html string) (*Document, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	return &Document{Document: doc}, nil
}

// NewDocumentFromSelection creates a Document wrapper from a selection.
// Note: This returns nil if the selection doesn't have a valid document.
func NewDocumentFromSelection(sel *goquery.Selection) *Document {
	// goquery doesn't expose the document directly from a selection
	// so we return nil and callers should use NewDocument instead
	return nil
}

// Clone creates a deep clone of a selection.
func Clone(sel *goquery.Selection) *goquery.Selection {
	return sel.Clone()
}

// GetAttribute gets an attribute value from a selection.
func GetAttribute(sel *goquery.Selection, attr string) string {
	val, _ := sel.Attr(attr)
	return val
}

// HasAttribute checks if a selection has an attribute.
func HasAttribute(sel *goquery.Selection, attr string) bool {
	_, exists := sel.Attr(attr)
	return exists
}

// GetTagName returns the tag name of the first element in selection.
func GetTagName(sel *goquery.Selection) string {
	if sel.Length() == 0 {
		return ""
	}
	return goquery.NodeName(sel)
}

// IsTag checks if the selection is a specific tag.
func IsTag(sel *goquery.Selection, tag string) bool {
	return strings.EqualFold(GetTagName(sel), tag)
}

// InlineElements are elements that are typically inline.
var InlineElements = map[string]bool{
	"a":      true,
	"abbr":   true,
	"b":      true,
	"br":     true,
	"cite":   true,
	"code":   true,
	"em":     true,
	"i":      true,
	"img":    true,
	"kbd":    true,
	"mark":   true,
	"q":      true,
	"s":      true,
	"small":  true,
	"span":   true,
	"strong": true,
	"sub":    true,
	"sup":    true,
	"time":   true,
	"u":      true,
}

// BlockElements are elements that are typically block-level.
var BlockElements = map[string]bool{
	"address":    true,
	"article":    true,
	"aside":      true,
	"blockquote": true,
	"div":        true,
	"dl":         true,
	"fieldset":   true,
	"figcaption": true,
	"figure":     true,
	"footer":     true,
	"form":       true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"header":     true,
	"hgroup":     true,
	"hr":         true,
	"li":         true,
	"main":       true,
	"nav":        true,
	"noscript":   true,
	"ol":         true,
	"p":          true,
	"pre":        true,
	"section":    true,
	"table":      true,
	"ul":         true,
}

// IsInlineElement checks if a tag is an inline element.
func IsInlineElement(tag string) bool {
	return InlineElements[strings.ToLower(tag)]
}

// IsBlockElement checks if a tag is a block element.
func IsBlockElement(tag string) bool {
	return BlockElements[strings.ToLower(tag)]
}

// RemoveEmptyElements removes empty elements from a selection.
func RemoveEmptyElements(sel *goquery.Selection) {
	sel.Find("*").Each(func(_ int, s *goquery.Selection) {
		// Skip elements that should not be removed even if empty
		tag := GetTagName(s)
		if tag == "br" || tag == "hr" || tag == "img" || tag == "input" {
			return
		}

		// Check if empty
		html, _ := s.Html()
		if strings.TrimSpace(html) == "" && strings.TrimSpace(s.Text()) == "" {
			s.Remove()
		}
	})
}

// UnwrapElement unwraps an element, keeping its children.
func UnwrapElement(sel *goquery.Selection) {
	sel.Each(func(_ int, s *goquery.Selection) {
		s.ReplaceWithSelection(s.Contents())
	})
}
