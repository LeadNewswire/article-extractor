# article-extractor

A Golang library for extracting article content from HTML pages. Inspired by [Postlight Parser (Mercury)](https://github.com/postlight/parser).

## Features

- Extract article content from HTML
- Extract metadata (title, author, publication date, lead image)
- Remove ads, navigation, sidebars, and other non-content elements
- Support for Open Graph, Schema.org, and common HTML patterns
- Configurable extraction parameters
- URL fetching with charset detection

## Installation

```bash
go get github.com/example/article-extractor
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    extractor "github.com/example/article-extractor"
)

func main() {
    html := `<html>...</html>`

    ext := extractor.New()
    article, err := ext.Extract(html)
    if err != nil {
        panic(err)
    }

    fmt.Println("Title:", article.Title)
    fmt.Println("Author:", article.Author)
    fmt.Println("Content:", article.TextContent)
}
```

### Extract from URL

```go
package main

import (
    "context"
    "fmt"
    extractor "github.com/example/article-extractor"
)

func main() {
    ctx := context.Background()

    ext := extractor.New()
    article, err := ext.ExtractFromURL(ctx, "https://example.com/article")
    if err != nil {
        panic(err)
    }

    fmt.Println("Title:", article.Title)
    fmt.Println("Word Count:", article.WordCount)
}
```

### With Options

```go
ext := extractor.New(
    extractor.WithMinContentLength(200),
    extractor.WithMinParagraphLength(50),
    extractor.WithHTTPTimeout(10 * time.Second),
    extractor.WithUserAgent("MyBot/1.0"),
    extractor.WithDebug(true),
)
```

## Article Structure

```go
type Article struct {
    Title       string     // Article title
    Content     string     // Cleaned HTML content
    TextContent string     // Plain text content
    Excerpt     string     // Short excerpt
    Author      string     // Author name
    PublishedAt *time.Time // Publication date
    LeadImage   *Image     // Main image
    URL         string     // Source URL
    WordCount   int        // Word count
    Score       float64    // Extraction score
    Confidence  float64    // Confidence level (0-1)
}
```

## Algorithm

The extraction algorithm is based on content scoring:

1. **Preprocessing**: Remove scripts, styles, hidden elements, and unlikely candidates
2. **Scoring**: Score paragraphs based on:
   - Text length
   - Comma count
   - Class/ID weight (positive for content-related, negative for ads/navigation)
   - hNews microformat bonus
3. **Propagation**: Propagate scores to parent and grandparent elements
4. **Selection**: Find the highest-scoring element as the main content
5. **Sibling Merging**: Merge qualifying sibling elements
6. **Postprocessing**: Clean attributes, remove empty elements

## Scoring System

| Scoring Type | Points |
|-------------|--------|
| hNews microformat | +80 |
| Positive class/id (article, content, post) | +25 |
| Negative class/id (ad, sidebar, comment) | -25 |
| Paragraph base score | +1 |
| Per comma | +1 |
| Per 50 characters length | +1 (max 3) |
| div tag | +5 |
| td/blockquote | +3 |
| form/address | -3 |
| Parent propagation | 100% |
| Grandparent propagation | 50% |

## License

MIT
