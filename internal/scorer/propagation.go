package scorer

import (
	"github.com/LeadNewswire/article-extractor/internal/dom"
	"github.com/LeadNewswire/article-extractor/internal/keywords"
	"github.com/PuerkitoBio/goquery"
)

// PropagateScores propagates paragraph scores to parent and grandparent elements.
func PropagateScores(doc *goquery.Document, scoreMap *ScoreMap, minParagraphLength int) {
	// Score all paragraphs and propagate
	doc.Find("p, pre").Each(func(_ int, sel *goquery.Selection) {
		paragraphScore := ScoreParagraph(sel, minParagraphLength)
		if paragraphScore == 0 {
			return
		}

		// Get parent and grandparent
		parent := GetParagraphParent(sel)
		grandparent := GetParagraphGrandparent(sel)

		// Initialize parent if needed
		if parent != nil && parent.Length() > 0 {
			parentScore := scoreMap.Get(parent)
			if parentScore.ContentScore == 0 {
				initializeNodeScore(parentScore, parent)
			}
			// Add full score to parent
			parentScore.AddScore(paragraphScore * ParentScoreProportion)
		}

		// Initialize grandparent if needed
		if grandparent != nil && grandparent.Length() > 0 {
			grandparentScore := scoreMap.Get(grandparent)
			if grandparentScore.ContentScore == 0 {
				initializeNodeScore(grandparentScore, grandparent)
			}
			// Add half score to grandparent
			grandparentScore.AddScore(paragraphScore * GrandparentScoreProportion)
		}
	})
}

// initializeNodeScore initializes a node's score based on its properties.
func initializeNodeScore(ns *NodeScore, sel *goquery.Selection) {
	// Get tag name
	tag := dom.GetTagName(sel)

	// Add tag-based score
	ns.AddScore(GetTagScore(tag))

	// Add weight from class/id
	class := dom.GetAttribute(sel, "class")
	id := dom.GetAttribute(sel, "id")
	weight := keywords.GetWeight(class, id)
	ns.SetWeight(weight)

	// Check for hNews microformat
	if hasHNews(sel, class) {
		ns.AddScore(HNewsBonus)
	}

	// Calculate link density
	linkDensity := dom.CalculateLinkDensity(sel)
	ns.SetLinkDensity(linkDensity)

	// Set text length
	textLen := dom.GetTextLength(sel)
	ns.SetTextLength(textLen)
}

// hasHNews checks if an element has hNews microformat indicators.
func hasHNews(sel *goquery.Selection, class string) bool {
	// Check for hentry class
	if sel.HasClass("hentry") || sel.HasClass("h-entry") {
		return true
	}

	// Check for entry-content class
	if sel.HasClass("entry-content") {
		return true
	}

	// Check for itemtype schema.org/Article
	itemtype, exists := sel.Attr("itemtype")
	if exists && (itemtype == "http://schema.org/Article" ||
		itemtype == "https://schema.org/Article" ||
		itemtype == "http://schema.org/NewsArticle" ||
		itemtype == "https://schema.org/NewsArticle") {
		return true
	}

	return false
}

// ScoreAndPropagate scores all content and returns the score map.
func ScoreAndPropagate(doc *goquery.Document, minParagraphLength int) *ScoreMap {
	scoreMap := NewScoreMap()
	PropagateScores(doc, scoreMap, minParagraphLength)
	return scoreMap
}

// RefineScores adjusts scores based on link density and other factors.
func RefineScores(scoreMap *ScoreMap) {
	for _, ns := range scoreMap.scores {
		// Penalize high link density
		if ns.IsHighLinkDensity() {
			ns.SetScore(ns.ContentScore * (1 - ns.LinkDensity))
		}
	}
}
