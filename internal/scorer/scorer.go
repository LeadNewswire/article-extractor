package scorer

import (
	"github.com/example/article-extractor/internal/dom"
	"github.com/example/article-extractor/internal/keywords"
	"github.com/PuerkitoBio/goquery"
)

// Scorer is the main scoring engine.
type Scorer struct {
	minParagraphLength int
	minContentLength   int
	debug              bool
}

// NewScorer creates a new Scorer.
func NewScorer(minParagraphLength, minContentLength int, debug bool) *Scorer {
	return &Scorer{
		minParagraphLength: minParagraphLength,
		minContentLength:   minContentLength,
		debug:              debug,
	}
}

// Score scores a document and returns the top candidate.
func (s *Scorer) Score(doc *goquery.Document) (*NodeScore, *ScoreMap) {
	// Build score map
	scoreMap := ScoreAndPropagate(doc, s.minParagraphLength)

	// Refine scores
	RefineScores(scoreMap)

	// Get top candidate
	topCandidate := scoreMap.GetTopCandidate()

	return topCandidate, scoreMap
}

// FindTopCandidate finds the best content candidate from a document.
func (s *Scorer) FindTopCandidate(doc *goquery.Document) *goquery.Selection {
	topCandidate, _ := s.Score(doc)
	if topCandidate == nil {
		return nil
	}
	return topCandidate.Selection
}

// ScoreSelection scores a single selection.
func (s *Scorer) ScoreSelection(sel *goquery.Selection) *NodeScore {
	ns := NewNodeScore(sel)

	// Get tag name
	tag := dom.GetTagName(sel)

	// Add tag-based score
	ns.AddScore(GetTagScore(tag))

	// Add weight from class/id
	class := dom.GetAttribute(sel, "class")
	id := dom.GetAttribute(sel, "id")
	weight := keywords.GetWeight(class, id)
	ns.SetWeight(weight)

	// Calculate link density
	linkDensity := dom.CalculateLinkDensity(sel)
	ns.SetLinkDensity(linkDensity)

	// Set text length
	textLen := dom.GetTextLength(sel)
	ns.SetTextLength(textLen)

	// Score child paragraphs
	sel.Find("p, pre").Each(func(_ int, p *goquery.Selection) {
		score := ScoreParagraph(p, s.minParagraphLength)
		ns.AddScore(score)
	})

	return ns
}

// GetSiblingThreshold calculates the threshold for sibling merging.
func GetSiblingThreshold(topScore float64) float64 {
	threshold := topScore * SiblingScoreThresholdFactor
	if threshold < SiblingScoreThresholdBase {
		return SiblingScoreThresholdBase
	}
	return threshold
}

// ShouldMergeSibling determines if a sibling should be merged.
func ShouldMergeSibling(sibling *goquery.Selection, threshold float64, minParagraphLength int) bool {
	// Check tag
	tag := dom.GetTagName(sibling)

	// Always merge paragraphs with enough content
	if tag == "p" {
		textLen := dom.GetTextLength(sibling)
		linkDensity := dom.CalculateLinkDensity(sibling)
		return textLen >= minParagraphLength && linkDensity < LowWeightLinkDensityMax
	}

	// For other elements, calculate score
	class := dom.GetAttribute(sibling, "class")
	id := dom.GetAttribute(sibling, "id")
	weight := keywords.GetWeight(class, id)

	// Don't merge elements with negative weight
	if weight < 0 {
		return false
	}

	// Check link density
	linkDensity := dom.CalculateLinkDensity(sibling)
	if linkDensity > LowWeightLinkDensityMax {
		return false
	}

	// Score the sibling's paragraphs
	score := 0.0
	sibling.Find("p, pre").Each(func(_ int, p *goquery.Selection) {
		score += ScoreParagraph(p, minParagraphLength)
	})

	return score >= threshold
}
