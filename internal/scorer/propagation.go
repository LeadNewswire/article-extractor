package scorer

import (
	"github.com/LeadNewswire/article-extractor/internal/dom"
	"github.com/LeadNewswire/article-extractor/internal/keywords"
	"github.com/PuerkitoBio/goquery"
)

// PropagateScores propagates paragraph scores up to AncestorMaxDepth ancestor levels.
// Each level receives geometrically decaying credit: parent gets 100%, grandparent 50%,
// great-grandparent 25%, and so on. This allows the scorer to recognize content that is
// scattered across deeply nested sibling containers (e.g., slideshow slides).
func PropagateScores(doc *goquery.Document, scoreMap *ScoreMap, minParagraphLength int) {
	doc.Find("p, pre").Each(func(_ int, sel *goquery.Selection) {
		paragraphScore := ScoreParagraph(sel, minParagraphLength)
		if paragraphScore == 0 {
			return
		}

		// Walk up the ancestor chain with decaying score proportion
		proportion := 1.0
		ancestor := sel.Parent()
		for depth := 0; depth < AncestorMaxDepth; depth++ {
			if ancestor == nil || ancestor.Length() == 0 {
				break
			}

			tag := dom.GetTagName(ancestor)
			if tag == "body" || tag == "html" {
				break
			}

			ns := scoreMap.Get(ancestor)
			if ns.ContentScore == 0 {
				initializeNodeScore(ns, ancestor)
			}
			ns.AddScore(paragraphScore * proportion)

			proportion *= 0.5 // Geometric decay
			ancestor = ancestor.Parent()
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
