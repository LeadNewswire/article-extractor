package extractor_test

import (
	"fmt"
	"os"

	extractor "github.com/LeadNewswire/article-extractor"
)

func Example() {
	html := `
<!DOCTYPE html>
<html>
<head>
	<title>Sample Article - News Site</title>
	<meta property="og:title" content="Sample Article">
	<meta name="author" content="John Doe">
</head>
<body>
	<article>
		<h1>Sample Article</h1>
		<p>This is a sample article with multiple paragraphs of content. The extractor will identify this as the main content area based on the text density and structure.</p>
		<p>The second paragraph provides additional information and context. Good articles typically have multiple paragraphs that tell a complete story.</p>
		<p>Finally, the third paragraph wraps up the article with concluding remarks and summary.</p>
	</article>
</body>
</html>`

	ext := extractor.New()
	article, err := ext.Extract(html)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Title: %s\n", article.Title)
	fmt.Printf("Author: %s\n", article.Author)
	fmt.Printf("Word Count: %d\n", article.WordCount)
	fmt.Printf("Confidence: %.2f\n", article.Confidence)
	// Output:
	// Title: Sample Article
	// Author: John Doe
	// Word Count: 61
	// Confidence: 0.30
}

func ExampleExtractor_ExtractWithURL() {
	html := `
<!DOCTYPE html>
<html>
<head>
	<title>News Article</title>
</head>
<body>
	<article>
		<p>This is a news article about important events. It contains substantial content that the extractor will process.</p>
		<p>The second paragraph has more details about the events described in the article.</p>
		<p>A third paragraph concludes the article with additional context.</p>
		<img src="/images/photo.jpg" alt="Photo">
	</article>
</body>
</html>`

	ext := extractor.New()
	article, err := ext.ExtractWithURL(html, "https://example.com/article")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("URL: %s\n", article.URL)
	fmt.Printf("Has Content: %v\n", len(article.TextContent) > 0)
	// Output:
	// URL: https://example.com/article
	// Has Content: true
}

func ExampleExtractor_Extract_withOptions() {
	html := `
<!DOCTYPE html>
<html>
<body>
	<article>
		<p>A short paragraph.</p>
		<p>Another brief paragraph with some text content for the extraction algorithm to process and evaluate.</p>
		<p>Third paragraph to ensure adequate content.</p>
	</article>
</body>
</html>`

	// Create extractor with custom options
	ext := extractor.New(
		extractor.WithMinContentLength(50),
		extractor.WithMinParagraphLength(10),
		extractor.WithDebug(false),
	)

	article, err := ext.Extract(html)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Extracted: %v\n", article.Title != "" || len(article.TextContent) > 0)
	// Output:
	// Extracted: true
}

func ExampleNew() {
	// Create a basic extractor
	ext := extractor.New()
	_ = ext

	// Create an extractor with options
	extWithOpts := extractor.New(
		extractor.WithMinContentLength(200),
		extractor.WithDebug(true),
	)
	_ = extWithOpts

	fmt.Println("Extractors created successfully")
	// Output:
	// Extractors created successfully
}

// This example shows how to extract from a fixture file.
func Example_fromFile() {
	// Read HTML from fixture file
	html, err := os.ReadFile("testdata/fixtures/sample_article.html")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	ext := extractor.New()
	article, err := ext.Extract(string(html))
	if err != nil {
		fmt.Printf("Error extracting: %v\n", err)
		return
	}

	fmt.Printf("Title: %s\n", article.Title)
	fmt.Printf("Author: %s\n", article.Author)
	fmt.Printf("Has Date: %v\n", article.PublishedAt != nil)
	fmt.Printf("Has Image: %v\n", article.LeadImage != nil)
	fmt.Printf("Word Count > 100: %v\n", article.WordCount > 100)
	// Output:
	// Title: Breaking News: Major Scientific Discovery Announced
	// Author: Dr. Sarah Johnson
	// Has Date: true
	// Has Image: true
	// Word Count > 100: true
}
