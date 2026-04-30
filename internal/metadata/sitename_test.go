package metadata

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractSiteName(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "og:site_name",
			html:     `<html><head><meta property="og:site_name" content="The Motley Fool"></head><body></body></html>`,
			expected: "The Motley Fool",
		},
		{
			name:     "json-ld publisher.name (object)",
			html:     `<html><head><script type="application/ld+json">{"@type":"NewsArticle","publisher":{"@type":"Organization","name":"Reuters"}}</script></head><body></body></html>`,
			expected: "Reuters",
		},
		{
			name:     "json-ld publisher.name (array root)",
			html:     `<html><head><script type="application/ld+json">[{"@type":"WebPage"},{"@type":"NewsArticle","publisher":{"name":"Bloomberg"}}]</script></head><body></body></html>`,
			expected: "Bloomberg",
		},
		{
			name:     "application-name",
			html:     `<html><head><meta name="application-name" content="ExampleSite"></head><body></body></html>`,
			expected: "ExampleSite",
		},
		{
			name:     "twitter:site strips at-prefix",
			html:     `<html><head><meta name="twitter:site" content="@cnn"></head><body></body></html>`,
			expected: "cnn",
		},
		{
			name:     "og:site_name takes precedence over json-ld",
			html:     `<html><head><meta property="og:site_name" content="Primary"><script type="application/ld+json">{"publisher":{"name":"Secondary"}}</script></head><body></body></html>`,
			expected: "Primary",
		},
		{
			name:     "whitespace collapsed",
			html:     `<html><head><meta property="og:site_name" content="  Foo   Bar  "></head><body></body></html>`,
			expected: "Foo Bar",
		},
		{
			name:     "empty when no metadata",
			html:     `<html><body><p>nothing</p></body></html>`,
			expected: "",
		},
		{
			name:     "malformed json-ld is ignored",
			html:     `<html><head><script type="application/ld+json">{not json</script><meta property="og:site_name" content="Fallback"></head><body></body></html>`,
			expected: "Fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatal(err)
			}
			got := ExtractSiteName(doc)
			if got != tt.expected {
				t.Errorf("ExtractSiteName = %q, want %q", got, tt.expected)
			}
		})
	}
}
