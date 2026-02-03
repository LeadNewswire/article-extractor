package scorer

import (
	"github.com/PuerkitoBio/goquery"
)

// NodeScore holds the score information for a DOM node.
type NodeScore struct {
	// Selection is the goquery selection for this node
	Selection *goquery.Selection

	// ContentScore is the calculated content score
	ContentScore float64

	// Weight is the class/id based weight
	Weight int

	// LinkDensity is the link text to total text ratio
	LinkDensity float64

	// TextLength is the text length of this node
	TextLength int
}

// NewNodeScore creates a new NodeScore for a selection.
func NewNodeScore(sel *goquery.Selection) *NodeScore {
	return &NodeScore{
		Selection:    sel,
		ContentScore: 0,
		Weight:       0,
		LinkDensity:  0,
		TextLength:   0,
	}
}

// AddScore adds to the content score.
func (ns *NodeScore) AddScore(score float64) {
	ns.ContentScore += score
}

// SetScore sets the content score.
func (ns *NodeScore) SetScore(score float64) {
	ns.ContentScore = score
}

// GetScore returns the content score.
func (ns *NodeScore) GetScore() float64 {
	return ns.ContentScore
}

// GetWeightedScore returns the score adjusted by weight.
func (ns *NodeScore) GetWeightedScore() float64 {
	return ns.ContentScore + float64(ns.Weight)
}

// SetWeight sets the weight.
func (ns *NodeScore) SetWeight(weight int) {
	ns.Weight = weight
}

// SetLinkDensity sets the link density.
func (ns *NodeScore) SetLinkDensity(density float64) {
	ns.LinkDensity = density
}

// SetTextLength sets the text length.
func (ns *NodeScore) SetTextLength(length int) {
	ns.TextLength = length
}

// IsHighLinkDensity checks if the node has high link density.
func (ns *NodeScore) IsHighLinkDensity() bool {
	if ns.Weight >= 0 {
		return ns.LinkDensity > HighWeightLinkDensityMax
	}
	return ns.LinkDensity > LowWeightLinkDensityMax
}

// ScoreMap manages scores for multiple nodes.
type ScoreMap struct {
	scores map[*goquery.Selection]*NodeScore
}

// NewScoreMap creates a new ScoreMap.
func NewScoreMap() *ScoreMap {
	return &ScoreMap{
		scores: make(map[*goquery.Selection]*NodeScore),
	}
}

// Get returns the NodeScore for a selection, creating one if it doesn't exist.
func (sm *ScoreMap) Get(sel *goquery.Selection) *NodeScore {
	if sel == nil || sel.Length() == 0 {
		return nil
	}

	// Use the first node as the key
	node := sel.Get(0)
	for key, score := range sm.scores {
		if key.Get(0) == node {
			return score
		}
	}

	// Create new score
	ns := NewNodeScore(sel)
	sm.scores[sel] = ns
	return ns
}

// GetOrNil returns the NodeScore for a selection, or nil if not found.
func (sm *ScoreMap) GetOrNil(sel *goquery.Selection) *NodeScore {
	if sel == nil || sel.Length() == 0 {
		return nil
	}

	node := sel.Get(0)
	for key, score := range sm.scores {
		if key.Get(0) == node {
			return score
		}
	}
	return nil
}

// Set sets the NodeScore for a selection.
func (sm *ScoreMap) Set(sel *goquery.Selection, score *NodeScore) {
	sm.scores[sel] = score
}

// GetTopCandidate returns the selection with the highest score.
func (sm *ScoreMap) GetTopCandidate() *NodeScore {
	var top *NodeScore
	var topScore float64 = -1

	for _, ns := range sm.scores {
		score := ns.GetWeightedScore()
		if score > topScore {
			topScore = score
			top = ns
		}
	}

	return top
}

// GetCandidatesByScore returns candidates sorted by score (descending).
func (sm *ScoreMap) GetCandidatesByScore() []*NodeScore {
	candidates := make([]*NodeScore, 0, len(sm.scores))
	for _, ns := range sm.scores {
		candidates = append(candidates, ns)
	}

	// Simple bubble sort for small lists
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].GetWeightedScore() > candidates[i].GetWeightedScore() {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	return candidates
}

// Size returns the number of scored nodes.
func (sm *ScoreMap) Size() int {
	return len(sm.scores)
}
