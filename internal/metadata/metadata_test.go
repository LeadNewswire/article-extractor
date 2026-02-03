package metadata

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "og:title",
			html: `<html><head><meta property="og:title" content="OG Title"></head><body></body></html>`,
			expected: "OG Title",
		},
		{
			name: "twitter:title",
			html: `<html><head><meta name="twitter:title" content="Twitter Title"></head><body></body></html>`,
			expected: "Twitter Title",
		},
		{
			name: "h1 in article",
			html: `<html><body><article><h1>Article H1 Title</h1></article></body></html>`,
			expected: "Article H1 Title",
		},
		{
			name: "first h1",
			html: `<html><body><h1>First H1</h1></body></html>`,
			expected: "First H1",
		},
		{
			name: "title tag",
			html: `<html><head><title>Page Title</title></head><body></body></html>`,
			expected: "Page Title",
		},
		{
			name: "title with site name",
			html: `<html><head><title>A Very Long Article Title About Something Interesting - Site</title></head><body></body></html>`,
			expected: "A Very Long Article Title About Something Interesting",
		},
		{
			name: "schema.org headline",
			html: `<html><head><script type="application/ld+json">{"@type":"Article","headline":"Schema Headline"}</script></head><body></body></html>`,
			expected: "Schema Headline",
		},
		{
			name: "itemprop headline",
			html: `<html><body><h1 itemprop="headline">Itemprop Headline</h1></body></html>`,
			expected: "Itemprop Headline",
		},
		{
			name: "empty",
			html: `<html><body></body></html>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatal(err)
			}

			result := ExtractTitle(doc)
			if result != tt.expected {
				t.Errorf("ExtractTitle = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractAuthor(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "meta author",
			html: `<html><head><meta name="author" content="John Doe"></head><body></body></html>`,
			expected: "John Doe",
		},
		{
			name: "schema.org author",
			html: `<html><head><script type="application/ld+json">{"@type":"Article","author":{"name":"Jane Smith"}}</script></head><body></body></html>`,
			expected: "Jane Smith",
		},
		{
			name: "itemprop author",
			html: `<html><body><span itemprop="author"><span itemprop="name">Author Name</span></span></body></html>`,
			expected: "Author Name",
		},
		{
			name: "author class",
			html: `<html><body><span class="author">Class Author</span></body></html>`,
			expected: "Class Author",
		},
		{
			name: "byline",
			html: `<html><body><div class="byline">By Writer Name</div></body></html>`,
			expected: "Writer Name",
		},
		{
			name: "rel author",
			html: `<html><body><a rel="author">Link Author</a></body></html>`,
			expected: "Link Author",
		},
		{
			name: "empty",
			html: `<html><body></body></html>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatal(err)
			}

			result := ExtractAuthor(doc)
			if result != tt.expected {
				t.Errorf("ExtractAuthor = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractDate(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		expectNil   bool
		expectYear  int
		expectMonth int
		expectDay   int
	}{
		{
			name:        "article:published_time",
			html:        `<html><head><meta property="article:published_time" content="2024-01-15T10:00:00Z"></head><body></body></html>`,
			expectNil:   false,
			expectYear:  2024,
			expectMonth: 1,
			expectDay:   15,
		},
		{
			name:        "datePublished meta",
			html:        `<html><head><meta name="datePublished" content="2023-12-25"></head><body></body></html>`,
			expectNil:   false,
			expectYear:  2023,
			expectMonth: 12,
			expectDay:   25,
		},
		{
			name:        "time element",
			html:        `<html><body><time datetime="2024-06-10">June 10, 2024</time></body></html>`,
			expectNil:   false,
			expectYear:  2024,
			expectMonth: 6,
			expectDay:   10,
		},
		{
			name:        "schema.org datePublished",
			html:        `<html><head><script type="application/ld+json">{"@type":"Article","datePublished":"2024-03-20"}</script></head><body></body></html>`,
			expectNil:   false,
			expectYear:  2024,
			expectMonth: 3,
			expectDay:   20,
		},
		{
			name:        "itemprop datePublished",
			html:        `<html><body><span itemprop="datePublished" content="2024-08-15"></span></body></html>`,
			expectNil:   false,
			expectYear:  2024,
			expectMonth: 8,
			expectDay:   15,
		},
		{
			name:      "empty",
			html:      `<html><body></body></html>`,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatal(err)
			}

			result := ExtractDate(doc)

			if tt.expectNil {
				if result != nil {
					t.Errorf("ExtractDate = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("ExtractDate returned nil, expected a date")
			}

			if result.Year() != tt.expectYear {
				t.Errorf("Year = %d, want %d", result.Year(), tt.expectYear)
			}
			if int(result.Month()) != tt.expectMonth {
				t.Errorf("Month = %d, want %d", result.Month(), tt.expectMonth)
			}
			if result.Day() != tt.expectDay {
				t.Errorf("Day = %d, want %d", result.Day(), tt.expectDay)
			}
		})
	}
}

func TestExtractAuthors(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected int // Number of authors
	}{
		{
			name: "single author",
			html: `<html><head><meta name="author" content="John Doe"></head><body></body></html>`,
			expected: 1,
		},
		{
			name: "multiple authors with and",
			html: `<html><head><meta name="author" content="John Doe and Jane Smith"></head><body></body></html>`,
			expected: 2,
		},
		{
			name: "multiple authors with comma",
			html: `<html><head><meta name="author" content="John Doe, Jane Smith, Bob Wilson"></head><body></body></html>`,
			expected: 3,
		},
		{
			name: "no author",
			html: `<html><body></body></html>`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatal(err)
			}

			result := ExtractAuthors(doc)
			if len(result) != tt.expected {
				t.Errorf("ExtractAuthors count = %d, want %d", len(result), tt.expected)
			}
		})
	}
}
