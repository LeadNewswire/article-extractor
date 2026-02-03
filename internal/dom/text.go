package dom

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

var (
	// whitespaceRegex matches multiple whitespace characters.
	whitespaceRegex = regexp.MustCompile(`\s+`)

	// newlineRegex matches multiple newlines.
	newlineRegex = regexp.MustCompile(`\n{3,}`)
)

// GetText extracts and normalizes text from a selection.
func GetText(sel *goquery.Selection) string {
	text := sel.Text()
	return NormalizeText(text)
}

// GetTextLength returns the length of normalized text.
func GetTextLength(sel *goquery.Selection) int {
	return utf8.RuneCountInString(GetText(sel))
}

// NormalizeText normalizes whitespace in text.
func NormalizeText(text string) string {
	// Replace multiple whitespace with single space
	text = whitespaceRegex.ReplaceAllString(text, " ")
	// Trim leading/trailing whitespace
	text = strings.TrimSpace(text)
	return text
}

// NormalizeTextPreserveNewlines normalizes text but preserves paragraph breaks.
func NormalizeTextPreserveNewlines(text string) string {
	// Replace multiple spaces (but not newlines) with single space
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(whitespaceRegex.ReplaceAllString(line, " "))
	}
	text = strings.Join(lines, "\n")
	// Collapse multiple newlines to double newline
	text = newlineRegex.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

// CountWords counts the number of words in text.
func CountWords(text string) int {
	text = NormalizeText(text)
	if text == "" {
		return 0
	}

	count := 0
	inWord := false

	for _, r := range text {
		if unicode.IsSpace(r) {
			if inWord {
				count++
				inWord = false
			}
		} else {
			inWord = true
		}
	}

	if inWord {
		count++
	}

	return count
}

// CountCommas counts the number of commas in text.
func CountCommas(text string) int {
	count := 0
	for _, r := range text {
		if r == ',' || r == '，' { // Include Chinese comma
			count++
		}
	}
	return count
}

// CountSentences estimates the number of sentences in text.
func CountSentences(text string) int {
	count := 0
	for _, r := range text {
		if r == '.' || r == '!' || r == '?' || r == '。' || r == '！' || r == '？' {
			count++
		}
	}
	if count == 0 && len(text) > 0 {
		count = 1
	}
	return count
}

// GetLinkText extracts text from all links in a selection.
func GetLinkText(sel *goquery.Selection) string {
	var texts []string
	sel.Find("a").Each(func(_ int, s *goquery.Selection) {
		text := GetText(s)
		if text != "" {
			texts = append(texts, text)
		}
	})
	return strings.Join(texts, " ")
}

// GetLinkTextLength returns the total length of text in links.
func GetLinkTextLength(sel *goquery.Selection) int {
	totalLen := 0
	sel.Find("a").Each(func(_ int, s *goquery.Selection) {
		totalLen += GetTextLength(s)
	})
	return totalLen
}

// CalculateLinkDensity calculates the ratio of link text to total text.
// Returns a value between 0 and 1.
func CalculateLinkDensity(sel *goquery.Selection) float64 {
	textLen := GetTextLength(sel)
	if textLen == 0 {
		return 0
	}

	linkTextLen := GetLinkTextLength(sel)
	return float64(linkTextLen) / float64(textLen)
}

// GetExcerpt extracts a short excerpt from text.
func GetExcerpt(text string, maxLength int) string {
	text = NormalizeText(text)
	if utf8.RuneCountInString(text) <= maxLength {
		return text
	}

	// Find a good break point
	runes := []rune(text)
	if len(runes) <= maxLength {
		return text
	}

	// Look for sentence end within the limit
	for i := maxLength - 1; i >= maxLength/2; i-- {
		if runes[i] == '.' || runes[i] == '。' {
			return string(runes[:i+1])
		}
	}

	// Look for word break
	for i := maxLength - 1; i >= maxLength/2; i-- {
		if unicode.IsSpace(runes[i]) {
			return string(runes[:i]) + "..."
		}
	}

	// Hard cut
	return string(runes[:maxLength-3]) + "..."
}

// IsWhitespaceOnly checks if text contains only whitespace.
func IsWhitespaceOnly(text string) bool {
	return strings.TrimSpace(text) == ""
}

// HasSubstantialText checks if selection has substantial text content.
func HasSubstantialText(sel *goquery.Selection, minLength int) bool {
	return GetTextLength(sel) >= minLength
}
