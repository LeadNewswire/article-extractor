package dom

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello world"},
		{"  hello   world  ", "hello world"},
		{"hello\n\nworld", "hello world"},
		{"hello\t\tworld", "hello world"},
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeText(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"hello world", 2},
		{"one two three four five", 5},
		{"", 0},
		{"   ", 0},
		{"word", 1},
		{"  multiple   spaces   between  words  ", 4},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CountWords(tt.input)
			if result != tt.expected {
				t.Errorf("CountWords(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountCommas(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"hello, world", 1},
		{"one, two, three", 2},
		{"no commas here", 0},
		{"", 0},
		{"中文，逗號", 1}, // Chinese comma
		{"mixed, commas，here", 2},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CountCommas(tt.input)
			if result != tt.expected {
				t.Errorf("CountCommas(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetExcerpt(t *testing.T) {
	tests := []struct {
		input     string
		maxLength int
		checkLen  bool
	}{
		{"Short text", 100, true},
		{"This is a longer text that should be truncated at some point.", 30, true},
		{"Sentence one. Sentence two. Sentence three.", 25, true},
		{"", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.input[:min(len(tt.input), 20)], func(t *testing.T) {
			result := GetExcerpt(tt.input, tt.maxLength)
			if tt.checkLen && len(result) > tt.maxLength+3 { // Allow for "..."
				t.Errorf("GetExcerpt length = %d, want <= %d", len(result), tt.maxLength+3)
			}
		})
	}
}

func TestCalculateLinkDensity(t *testing.T) {
	html := `<div>This is regular text. <a href="#">This is a link</a>. More text here.</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	sel := doc.Find("div")
	density := CalculateLinkDensity(sel)

	// Link text is "This is a link" (14 chars)
	// Total text is longer
	// Density should be > 0 and < 1
	if density <= 0 || density >= 1 {
		t.Errorf("CalculateLinkDensity = %f, expected between 0 and 1", density)
	}
}

func TestCalculateLinkDensity_NoLinks(t *testing.T) {
	html := `<div>This is text without any links.</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	sel := doc.Find("div")
	density := CalculateLinkDensity(sel)

	if density != 0 {
		t.Errorf("CalculateLinkDensity with no links = %f, want 0", density)
	}
}

func TestCalculateLinkDensity_AllLinks(t *testing.T) {
	html := `<div><a href="#">All link text</a></div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	sel := doc.Find("div")
	density := CalculateLinkDensity(sel)

	if density != 1.0 {
		t.Errorf("CalculateLinkDensity with all links = %f, want 1.0", density)
	}
}

func TestGetText(t *testing.T) {
	html := `<div>  Hello   <span>World</span>  </div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	sel := doc.Find("div")
	text := GetText(sel)

	if text != "Hello World" {
		t.Errorf("GetText = %q, want %q", text, "Hello World")
	}
}

func TestIsWhitespaceOnly(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"   ", true},
		{"\n\t\r", true},
		{"text", false},
		{" text ", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsWhitespaceOnly(tt.input)
			if result != tt.expected {
				t.Errorf("IsWhitespaceOnly(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
