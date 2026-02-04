package keywords

import "regexp"

// Blacklist keywords that indicate non-content elements.
// These are patterns that suggest an element is unlikely to be the main content.
var blacklistKeywords = []string{
	"ad",
	"advertisement",
	"banner",
	"breadcrumbs",
	"combx",
	"comment",
	"community",
	"cover-wrap",
	"disqus",
	"extra",
	"footer",
	"gdpr",
	"header",
	"legends",
	"menu",
	"related",
	"remark",
	"replies",
	"rss",
	"shoutbox",
	"sidebar",
	"skyscraper",
	"social",
	"sponsor",
	"supplemental",
	"widget",
	"agegate",
	"pagination",
	"pager",
	"popup",
	"print",
	"archive",
	"author-info",
	"author-box",
	"bio",
	"carousel",
	"gallery",
	"modal",
	"navigation",
	"newsletter",
	"promo",
	"share",
	"subscribe",
	"tags",
	"toolbar",
	"trending",
	// AI/chatbot widgets
	"dd-widget",      // DeeperDive widget
	"deeperdive",     // DeeperDive
	"ai-widget",      // Generic AI widgets
	"chatbot",        // Chatbot widgets
	"ask-ai",         // AI Q&A widgets
	"genai",          // GenAI widgets
	"ai-assistant",   // AI assistant widgets
	"ai-answer",      // AI answer engines
}

var blacklistPattern *regexp.Regexp

func init() {
	// Build the combined pattern
	pattern := buildPattern(blacklistKeywords)
	blacklistPattern = regexp.MustCompile(pattern)
}

// IsBlacklisted checks if a string matches any blacklist keyword.
func IsBlacklisted(s string) bool {
	if s == "" {
		return false
	}
	return blacklistPattern.MatchString(s)
}

// GetBlacklistPattern returns the compiled blacklist pattern.
func GetBlacklistPattern() *regexp.Regexp {
	return blacklistPattern
}

// buildPattern builds a regex pattern from keywords.
func buildPattern(keywords []string) string {
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
