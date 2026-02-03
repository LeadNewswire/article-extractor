package extractor

import "time"

// Config holds the configuration options for the extractor.
type Config struct {
	// MinContentLength is the minimum content length to consider valid
	MinContentLength int

	// MinParagraphLength is the minimum paragraph length to score
	MinParagraphLength int

	// Debug enables debug logging
	Debug bool

	// HTTPTimeout is the timeout for HTTP requests
	HTTPTimeout time.Duration

	// UserAgent is the User-Agent header for HTTP requests
	UserAgent string

	// MaxContentLength is the maximum HTML content length to process
	MaxContentLength int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		MinContentLength:   100,
		MinParagraphLength: 25,
		Debug:              false,
		HTTPTimeout:        30 * time.Second,
		UserAgent:          "Mozilla/5.0 (compatible; ArticleExtractor/1.0)",
		MaxContentLength:   10 * 1024 * 1024, // 10MB
	}
}

// Option is a function that modifies the config.
type Option func(*Config)

// WithMinContentLength sets the minimum content length.
func WithMinContentLength(length int) Option {
	return func(c *Config) {
		c.MinContentLength = length
	}
}

// WithMinParagraphLength sets the minimum paragraph length.
func WithMinParagraphLength(length int) Option {
	return func(c *Config) {
		c.MinParagraphLength = length
	}
}

// WithDebug enables or disables debug mode.
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

// WithHTTPTimeout sets the HTTP request timeout.
func WithHTTPTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.HTTPTimeout = timeout
	}
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Config) {
		c.UserAgent = ua
	}
}

// WithMaxContentLength sets the maximum content length.
func WithMaxContentLength(length int) Option {
	return func(c *Config) {
		c.MaxContentLength = length
	}
}
