package keywords

import "regexp"

// Whitelist keywords that indicate content elements.
// These are patterns that suggest an element is likely to be the main content.
var whitelistKeywords = []string{
	"article",
	"body",
	"content",
	"entry",
	"main",
	"page",
	"post",
	"text",
	"blog",
	"story",
	"hentry",
	"h-entry",
	"entry-content",
	"article-body",
	"article-content",
}

var whitelistPattern *regexp.Regexp

func init() {
	// Build the combined pattern
	pattern := buildWhitelistPattern(whitelistKeywords)
	whitelistPattern = regexp.MustCompile(pattern)
}

// IsWhitelisted checks if a string matches any whitelist keyword.
func IsWhitelisted(s string) bool {
	if s == "" {
		return false
	}
	return whitelistPattern.MatchString(s)
}

// GetWhitelistPattern returns the compiled whitelist pattern.
func GetWhitelistPattern() *regexp.Regexp {
	return whitelistPattern
}

// buildWhitelistPattern builds a regex pattern from keywords.
func buildWhitelistPattern(keywords []string) string {
	if len(keywords) == 0 {
		return "(?!)" // Never matches
	}
	pattern := "(?i)("
	for i, kw := range keywords {
		if i > 0 {
			pattern += "|"
		}
		pattern += regexp.QuoteMeta(kw)
	}
	pattern += ")"
	return pattern
}

// GetWeight returns the weight for a given class/id combination.
// Positive weight indicates likely content, negative indicates likely non-content.
func GetWeight(class, id string) int {
	weight := 0

	// Check class
	if class != "" {
		if IsWhitelisted(class) {
			weight += 25
		}
		if IsBlacklisted(class) {
			weight -= 25
		}
	}

	// Check id
	if id != "" {
		if IsWhitelisted(id) {
			weight += 25
		}
		if IsBlacklisted(id) {
			weight -= 25
		}
	}

	return weight
}
