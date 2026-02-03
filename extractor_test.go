package extractor

import (
	"strings"
	"testing"
)

func TestExtract_SimpleArticle(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<head>
	<title>Test Article - Test Site</title>
	<meta property="og:title" content="Test Article">
	<meta name="author" content="John Doe">
</head>
<body>
	<header>
		<nav>Navigation</nav>
	</header>
	<article>
		<h1>Test Article Title</h1>
		<p class="byline">By John Doe</p>
		<p>This is the first paragraph of the article. It contains enough text to be considered meaningful content for extraction purposes. We need to have substantial content here.</p>
		<p>This is the second paragraph with more content. The extractor should be able to identify this as the main content area based on the scoring algorithm that evaluates text density and structure.</p>
		<p>The third paragraph continues the article with additional information. Good articles typically have multiple paragraphs with substantial content that tells a complete story.</p>
	</article>
	<aside>
		<h3>Related Articles</h3>
		<ul>
			<li><a href="#">Link 1</a></li>
			<li><a href="#">Link 2</a></li>
		</ul>
	</aside>
	<footer>Footer content</footer>
</body>
</html>`

	ext := New()
	article, err := ext.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Check title
	if article.Title != "Test Article" {
		t.Errorf("Expected title 'Test Article', got '%s'", article.Title)
	}

	// Check author
	if article.Author != "John Doe" {
		t.Errorf("Expected author 'John Doe', got '%s'", article.Author)
	}

	// Check content contains paragraphs
	if !strings.Contains(article.TextContent, "first paragraph") {
		t.Error("Content should contain 'first paragraph'")
	}

	// Check that navigation/sidebar is removed
	if strings.Contains(article.TextContent, "Navigation") {
		t.Error("Content should not contain navigation")
	}

	if strings.Contains(article.TextContent, "Related Articles") {
		t.Error("Content should not contain sidebar content")
	}

	// Check word count
	if article.WordCount == 0 {
		t.Error("Word count should be greater than 0")
	}

	// Check confidence
	if article.Confidence <= 0 || article.Confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", article.Confidence)
	}
}

func TestExtract_ArticleWithDivs(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<head>
	<title>Div Article</title>
</head>
<body>
	<div class="content">
		<div class="article-body">
			<p>First paragraph of the div-based article. This should be detected as content despite being nested in divs rather than an article element.</p>
			<p>Second paragraph with more information. The scoring algorithm should recognize this div as containing the main content based on text density.</p>
			<p>Third paragraph to ensure we have enough content for proper extraction and testing of the algorithm.</p>
		</div>
	</div>
	<div class="sidebar">
		<a href="#">Link 1</a>
		<a href="#">Link 2</a>
		<a href="#">Link 3</a>
	</div>
</body>
</html>`

	ext := New()
	article, err := ext.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Check content
	if !strings.Contains(article.TextContent, "div-based article") {
		t.Error("Content should contain article text")
	}

	// Check that sidebar with high link density is not included
	if strings.Contains(article.TextContent, "Link 1") && strings.Contains(article.TextContent, "Link 2") && strings.Contains(article.TextContent, "Link 3") {
		t.Error("Content should not contain sidebar links")
	}
}

func TestExtract_NoContent(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<head>
	<title>Empty Page</title>
</head>
<body>
	<nav>Navigation only</nav>
</body>
</html>`

	ext := New()
	_, err := ext.Extract(html)
	if err == nil {
		t.Error("Expected error for page with no content")
	}
}

func TestExtract_SchemaOrgArticle(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<head>
	<title>Schema.org Article</title>
	<script type="application/ld+json">
	{
		"@type": "Article",
		"headline": "Schema Article Title",
		"author": {"@type": "Person", "name": "Jane Smith"},
		"datePublished": "2024-01-15T10:00:00Z"
	}
	</script>
</head>
<body>
	<article itemscope itemtype="http://schema.org/Article">
		<h1 itemprop="headline">Schema Article Title</h1>
		<p>This is an article with schema.org markup. The extractor should be able to extract metadata from the JSON-LD script and recognize the article structure.</p>
		<p>The second paragraph provides more content for the extraction algorithm to work with. Schema.org markup should boost the confidence score.</p>
		<p>Third paragraph ensures we have adequate content length for successful extraction.</p>
	</article>
</body>
</html>`

	ext := New()
	article, err := ext.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Check title from schema
	if article.Title != "Schema Article Title" {
		t.Errorf("Expected title 'Schema Article Title', got '%s'", article.Title)
	}

	// Check author from schema
	if article.Author != "Jane Smith" {
		t.Errorf("Expected author 'Jane Smith', got '%s'", article.Author)
	}

	// Check date was extracted
	if article.PublishedAt == nil {
		t.Error("Expected publication date to be extracted")
	}
}

func TestExtract_WithMinContentLength(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
	<article>
		<p>Short content.</p>
	</article>
</body>
</html>`

	ext := New(WithMinContentLength(500))
	_, err := ext.Extract(html)
	if err == nil {
		t.Error("Expected error when content is shorter than minimum")
	}
}

func TestExtract_LeadImage(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<head>
	<meta property="og:image" content="https://example.com/image.jpg">
	<meta property="og:image:width" content="800">
	<meta property="og:image:height" content="600">
</head>
<body>
	<article>
		<p>This is an article with a lead image specified via Open Graph tags. The extractor should be able to identify and extract this metadata.</p>
		<p>Second paragraph for adequate content length. The lead image is an important piece of metadata for article display.</p>
		<p>Third paragraph to meet content requirements.</p>
	</article>
</body>
</html>`

	ext := New()
	article, err := ext.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Check lead image
	if article.LeadImage == nil {
		t.Fatal("Expected lead image to be extracted")
	}

	if article.LeadImage.URL != "https://example.com/image.jpg" {
		t.Errorf("Expected image URL 'https://example.com/image.jpg', got '%s'", article.LeadImage.URL)
	}

	if article.LeadImage.Width != 800 {
		t.Errorf("Expected image width 800, got %d", article.LeadImage.Width)
	}

	if article.LeadImage.Height != 600 {
		t.Errorf("Expected image height 600, got %d", article.LeadImage.Height)
	}
}

func TestExtract_RemovesHiddenElements(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
	<article>
		<p>Visible content that should be extracted. This is the main article content.</p>
		<div style="display: none;">Hidden content that should be removed.</div>
		<p>More visible content in the article.</p>
		<div hidden>Another hidden element.</div>
		<p>Final paragraph of visible content.</p>
	</article>
</body>
</html>`

	ext := New()
	article, err := ext.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Check that hidden content is removed
	if strings.Contains(article.TextContent, "Hidden content") {
		t.Error("Content should not contain hidden elements")
	}

	if strings.Contains(article.TextContent, "Another hidden") {
		t.Error("Content should not contain hidden attribute elements")
	}

	// Check that visible content is present
	if !strings.Contains(article.TextContent, "Visible content") {
		t.Error("Content should contain visible content")
	}
}

func TestExtract_Excerpt(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
	<article>
		<p>This is a long article with many paragraphs. The extractor should generate a meaningful excerpt from the beginning of the content.</p>
		<p>Second paragraph with more content.</p>
		<p>Third paragraph continues the story.</p>
	</article>
</body>
</html>`

	ext := New()
	article, err := ext.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Check excerpt exists
	if article.Excerpt == "" {
		t.Error("Expected excerpt to be generated")
	}

	// Check excerpt is not too long
	if len(article.Excerpt) > 250 {
		t.Errorf("Excerpt too long: %d characters", len(article.Excerpt))
	}
}
