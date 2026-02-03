package extractor

import "time"

// Article represents the extracted article data.
type Article struct {
	// Title is the article title
	Title string `json:"title"`

	// Content is the cleaned HTML content
	Content string `json:"content"`

	// TextContent is the plain text content
	TextContent string `json:"textContent"`

	// Excerpt is a short summary/excerpt of the article
	Excerpt string `json:"excerpt"`

	// Author is the article author
	Author string `json:"author,omitempty"`

	// PublishedAt is the article publication date
	PublishedAt *time.Time `json:"publishedAt,omitempty"`

	// LeadImage is the main article image
	LeadImage *Image `json:"leadImage,omitempty"`

	// URL is the source URL
	URL string `json:"url,omitempty"`

	// WordCount is the number of words in the article
	WordCount int `json:"wordCount"`

	// Score is the extraction score of the content
	Score float64 `json:"score"`

	// Confidence is the confidence level (0-1)
	Confidence float64 `json:"confidence"`
}

// Image represents an image in the article.
type Image struct {
	// URL is the image source URL
	URL string `json:"url"`

	// Width is the image width in pixels
	Width int `json:"width,omitempty"`

	// Height is the image height in pixels
	Height int `json:"height,omitempty"`

	// Alt is the alternative text
	Alt string `json:"alt,omitempty"`
}
