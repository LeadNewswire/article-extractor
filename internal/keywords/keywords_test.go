package keywords

import "testing"

func TestIsBlacklisted(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"sidebar", true},
		{"ad", true},
		{"advertisement", true},
		{"comment", true},
		{"footer", true},
		{"header", true},
		{"navigation", true},
		{"social", true},
		{"related", true},
		{"promo", true},
		{"article", false},
		{"content", false},
		{"post", false},
		{"main", false},
		{"", false},
		{"random-class", false},
		{"SIDEBAR", true},  // Case insensitive
		{"AdVeRtIsEmEnT", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsBlacklisted(tt.input)
			if result != tt.expected {
				t.Errorf("IsBlacklisted(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsWhitelisted(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"article", true},
		{"content", true},
		{"post", true},
		{"main", true},
		{"entry", true},
		{"blog", true},
		{"story", true},
		{"hentry", true},
		{"h-entry", true},
		{"entry-content", true},
		{"article-body", true},
		{"sidebar", false},
		{"ad", false},
		{"comment", false},
		{"", false},
		{"random-class", false},
		{"ARTICLE", true},  // Case insensitive
		{"CoNtEnT", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsWhitelisted(tt.input)
			if result != tt.expected {
				t.Errorf("IsWhitelisted(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetWeight(t *testing.T) {
	tests := []struct {
		class    string
		id       string
		expected int
	}{
		{"article", "", 25},
		{"", "content", 25},
		{"sidebar", "", -25},
		{"", "ad", -25},
		{"article", "content", 50},          // Both positive
		{"sidebar", "ad", -50},              // Both negative
		{"article", "sidebar", 0},           // Mixed
		{"article-sidebar", "", 0},          // Contains both
		{"", "", 0},                         // Empty
		{"random", "random", 0},             // No match
	}

	for _, tt := range tests {
		name := "class=" + tt.class + ",id=" + tt.id
		t.Run(name, func(t *testing.T) {
			result := GetWeight(tt.class, tt.id)
			if result != tt.expected {
				t.Errorf("GetWeight(%q, %q) = %v, want %v", tt.class, tt.id, result, tt.expected)
			}
		})
	}
}

func TestGetBlacklistPattern(t *testing.T) {
	pattern := GetBlacklistPattern()
	if pattern == nil {
		t.Fatal("GetBlacklistPattern returned nil")
	}

	// Should match known blacklist words
	if !pattern.MatchString("sidebar") {
		t.Error("Pattern should match 'sidebar'")
	}

	if !pattern.MatchString("advertisement") {
		t.Error("Pattern should match 'advertisement'")
	}

	// Should not match content words
	if pattern.MatchString("article") {
		t.Error("Pattern should not match 'article'")
	}
}

func TestGetWhitelistPattern(t *testing.T) {
	pattern := GetWhitelistPattern()
	if pattern == nil {
		t.Fatal("GetWhitelistPattern returned nil")
	}

	// Should match known whitelist words
	if !pattern.MatchString("article") {
		t.Error("Pattern should match 'article'")
	}

	if !pattern.MatchString("content") {
		t.Error("Pattern should match 'content'")
	}

	// Should not match sidebar/ad words
	if pattern.MatchString("sidebar") {
		t.Error("Pattern should not match 'sidebar'")
	}
}
