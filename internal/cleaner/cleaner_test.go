package cleaner

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestRemoveUnwantedTags(t *testing.T) {
	html := `
<html>
<head>
	<script>var x = 1;</script>
	<style>.test { color: red; }</style>
	<link rel="stylesheet" href="style.css">
</head>
<body>
	<p>Content</p>
	<noscript>No script content</noscript>
	<iframe src="frame.html"></iframe>
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	RemoveUnwantedTags(doc)

	// Check that scripts are removed
	if doc.Find("script").Length() > 0 {
		t.Error("script tags should be removed")
	}

	// Check that styles are removed
	if doc.Find("style").Length() > 0 {
		t.Error("style tags should be removed")
	}

	// Check that content is preserved
	if !strings.Contains(doc.Text(), "Content") {
		t.Error("Content should be preserved")
	}

	// Check that noscript is removed
	if doc.Find("noscript").Length() > 0 {
		t.Error("noscript tags should be removed")
	}

	// Check that iframe is removed
	if doc.Find("iframe").Length() > 0 {
		t.Error("iframe tags should be removed")
	}
}

func TestRemoveHiddenElements(t *testing.T) {
	html := `
<html>
<body>
	<p>Visible</p>
	<div style="display: none;">Hidden 1</div>
	<div style="visibility: hidden;">Hidden 2</div>
	<div hidden>Hidden 3</div>
	<div aria-hidden="true">Hidden 4</div>
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	RemoveHiddenElements(doc)

	text := doc.Text()

	if !strings.Contains(text, "Visible") {
		t.Error("Visible content should be preserved")
	}

	if strings.Contains(text, "Hidden 1") {
		t.Error("display:none element should be removed")
	}

	if strings.Contains(text, "Hidden 2") {
		t.Error("visibility:hidden element should be removed")
	}

	if strings.Contains(text, "Hidden 3") {
		t.Error("hidden attribute element should be removed")
	}

	if strings.Contains(text, "Hidden 4") {
		t.Error("aria-hidden element should be removed")
	}
}

func TestStripUnlikelyCandidates(t *testing.T) {
	html := `
<html>
<body>
	<article class="content">
		<p>Main content here that should be preserved in the output.</p>
	</article>
	<nav>Navigation links</nav>
	<aside>Sidebar content</aside>
	<footer>Footer content</footer>
	<div class="advertisement">Ad content</div>
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	StripUnlikelyCandidates(doc)

	text := doc.Text()

	// Main content should be preserved
	if !strings.Contains(text, "Main content") {
		t.Error("Main content should be preserved")
	}

	// Navigation should be removed
	if strings.Contains(text, "Navigation links") {
		t.Error("Navigation should be removed")
	}

	// Aside should be removed
	if strings.Contains(text, "Sidebar content") {
		t.Error("Aside should be removed")
	}

	// Footer should be removed
	if strings.Contains(text, "Footer content") {
		t.Error("Footer should be removed")
	}
}

func TestConvertToParagraphs(t *testing.T) {
	html := `
<html>
<body>
	<div>Simple div text that should become a paragraph.</div>
	<div>
		<p>Already a paragraph</p>
	</div>
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	ConvertToParagraphs(doc)

	// Should still contain the text
	text := doc.Text()
	if !strings.Contains(text, "Simple div text") {
		t.Error("Converted text should be preserved")
	}

	if !strings.Contains(text, "Already a paragraph") {
		t.Error("Original paragraph should be preserved")
	}
}

func TestCleanAttributes(t *testing.T) {
	html := `
<html>
<body>
	<a href="https://example.com" class="link" data-tracking="123" onclick="track()">Link</a>
	<img src="image.jpg" alt="Image" class="image" width="100" height="100" data-lazy="true">
	<p class="paragraph" id="p1" style="color: red;">Text</p>
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	CleanAttributes(doc.Selection)

	// Link should keep href and title only
	link := doc.Find("a")
	if _, exists := link.Attr("href"); !exists {
		t.Error("Link href should be preserved")
	}
	if _, exists := link.Attr("class"); exists {
		t.Error("Link class should be removed")
	}
	if _, exists := link.Attr("onclick"); exists {
		t.Error("Link onclick should be removed")
	}

	// Image should keep src, alt, title, width, height
	img := doc.Find("img")
	if _, exists := img.Attr("src"); !exists {
		t.Error("Image src should be preserved")
	}
	if _, exists := img.Attr("alt"); !exists {
		t.Error("Image alt should be preserved")
	}
	if _, exists := img.Attr("class"); exists {
		t.Error("Image class should be removed")
	}
}

func TestConvertRelativeURLs(t *testing.T) {
	html := `
<html>
<body>
	<a href="/page">Relative link</a>
	<a href="https://example.com/absolute">Absolute link</a>
	<img src="/images/photo.jpg">
	<img src="//cdn.example.com/img.jpg">
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	ConvertRelativeURLs(doc.Selection, "https://example.com/article")

	// Check relative link
	link := doc.Find("a").First()
	href, _ := link.Attr("href")
	if !strings.HasPrefix(href, "https://example.com/page") {
		t.Errorf("Relative link not converted: %s", href)
	}

	// Check absolute link (should remain unchanged)
	absLink := doc.Find("a").Eq(1)
	absHref, _ := absLink.Attr("href")
	if absHref != "https://example.com/absolute" {
		t.Errorf("Absolute link should not change: %s", absHref)
	}

	// Check relative image
	img := doc.Find("img").First()
	src, _ := img.Attr("src")
	if !strings.HasPrefix(src, "https://example.com/images/photo.jpg") {
		t.Errorf("Relative image not converted: %s", src)
	}
}

func TestRemoveEmptyElements(t *testing.T) {
	html := `
<html>
<body>
	<div></div>
	<p></p>
	<span>  </span>
	<p>Content</p>
	<br>
	<img src="test.jpg">
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	RemoveEmptyElements(doc.Selection)

	// Empty div should be removed
	if doc.Find("div").Length() > 0 {
		text := doc.Find("div").Text()
		if strings.TrimSpace(text) == "" {
			t.Error("Empty div should be removed")
		}
	}

	// Content paragraph should remain
	if !strings.Contains(doc.Text(), "Content") {
		t.Error("Non-empty paragraph should be preserved")
	}

	// Self-closing elements should remain
	if doc.Find("br").Length() == 0 {
		t.Error("br element should be preserved")
	}

	if doc.Find("img").Length() == 0 {
		t.Error("img element should be preserved")
	}
}

func TestGetCleanHTML(t *testing.T) {
	html := `<div>  <p>Test content</p>  </div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	result := GetCleanHTML(doc.Find("div"))
	if !strings.Contains(result, "Test content") {
		t.Error("GetCleanHTML should contain the content")
	}
}

func TestGetCleanText(t *testing.T) {
	html := `<div>  <p>Test   content</p>  <p>More text</p>  </div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	result := GetCleanText(doc.Find("div"))
	if !strings.Contains(result, "Test content") {
		t.Error("GetCleanText should normalize whitespace")
	}
}
