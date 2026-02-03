package cleaner

import (
	"github.com/LeadNewswire/article-extractor/internal/dom"
	"github.com/LeadNewswire/article-extractor/internal/keywords"
	"github.com/LeadNewswire/article-extractor/internal/scorer"
	"github.com/PuerkitoBio/goquery"
)

// MergeSiblings merges qualifying sibling elements into the content.
func MergeSiblings(topCandidate *goquery.Selection, topScore float64, minParagraphLength int) *goquery.Selection {
	if topCandidate == nil || topCandidate.Length() == 0 {
		return topCandidate
	}

	// Calculate threshold
	threshold := scorer.GetSiblingThreshold(topScore)

	// Get parent
	parent := topCandidate.Parent()
	if parent.Length() == 0 {
		return topCandidate
	}

	// Create a new div to hold the content
	contentDiv := goquery.NewDocumentFromNode(topCandidate.Nodes[0]).Find("body")
	if contentDiv.Length() == 0 {
		return topCandidate
	}

	// Check siblings
	siblings := parent.Children()
	var mergedContent []string

	siblings.Each(func(_ int, sibling *goquery.Selection) {
		// If this is the top candidate, include it
		if sibling.Nodes[0] == topCandidate.Nodes[0] {
			html, _ := sibling.Html()
			mergedContent = append(mergedContent, html)
			return
		}

		// Check if sibling should be merged
		if shouldMergeSibling(sibling, threshold, minParagraphLength) {
			html, _ := sibling.Html()
			mergedContent = append(mergedContent, html)
		}
	})

	// If we have merged content, update the selection
	if len(mergedContent) > 1 {
		// We need to wrap merged siblings in a container
		// For simplicity, we return the original but mark siblings for inclusion
		return topCandidate.Parent()
	}

	return topCandidate
}

// shouldMergeSibling determines if a sibling element should be merged.
func shouldMergeSibling(sibling *goquery.Selection, threshold float64, minParagraphLength int) bool {
	tag := dom.GetTagName(sibling)

	// Always consider merging paragraphs
	if tag == "p" {
		textLen := dom.GetTextLength(sibling)
		linkDensity := dom.CalculateLinkDensity(sibling)

		// Merge if has substantial text and low link density
		if textLen >= minParagraphLength && linkDensity < 0.2 {
			return true
		}
		return false
	}

	// For other elements, check class/id weight
	class := dom.GetAttribute(sibling, "class")
	id := dom.GetAttribute(sibling, "id")
	weight := keywords.GetWeight(class, id)

	// Don't merge negatively weighted elements
	if weight < 0 {
		return false
	}

	// Check link density
	linkDensity := dom.CalculateLinkDensity(sibling)
	if linkDensity > 0.25 {
		return false
	}

	// Score the sibling's content
	score := scoreSibling(sibling, minParagraphLength)

	// Merge if score meets threshold
	return score >= threshold
}

// scoreSibling calculates a score for a sibling element.
func scoreSibling(sibling *goquery.Selection, minParagraphLength int) float64 {
	score := 0.0

	// Score paragraphs within the sibling
	sibling.Find("p, pre").Each(func(_ int, p *goquery.Selection) {
		score += scorer.ScoreParagraph(p, minParagraphLength)
	})

	return score
}

// GetSiblingCandidates returns sibling elements that might contain content.
func GetSiblingCandidates(topCandidate *goquery.Selection) []*goquery.Selection {
	if topCandidate == nil || topCandidate.Length() == 0 {
		return nil
	}

	parent := topCandidate.Parent()
	if parent.Length() == 0 {
		return nil
	}

	var candidates []*goquery.Selection
	siblings := parent.Children()

	siblings.Each(func(_ int, sibling *goquery.Selection) {
		// Skip the top candidate itself
		if sibling.Nodes[0] == topCandidate.Nodes[0] {
			return
		}

		// Skip obviously bad candidates
		tag := dom.GetTagName(sibling)
		if tag == "script" || tag == "style" || tag == "nav" || tag == "aside" {
			return
		}

		// Check if has meaningful text
		textLen := dom.GetTextLength(sibling)
		if textLen >= 25 {
			candidates = append(candidates, sibling)
		}
	})

	return candidates
}

// MergeContent merges multiple selections into a single HTML string.
func MergeContent(selections []*goquery.Selection) string {
	var result string
	for _, sel := range selections {
		html, _ := sel.Html()
		result += html
	}
	return result
}

// WrapInContainer wraps content in a container element.
func WrapInContainer(html, tag string) string {
	return "<" + tag + ">" + html + "</" + tag + ">"
}
