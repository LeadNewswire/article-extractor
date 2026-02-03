package scorer

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestScoreParagraph(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		minLength int
		minScore  float64
		maxScore  float64
	}{
		{
			name:      "Short paragraph (too short)",
			text:      "Short text",
			minLength: 25,
			minScore:  0,
			maxScore:  0,
		},
		{
			name:      "Medium paragraph",
			text:      "This is a medium length paragraph with some content.",
			minLength: 25,
			minScore:  1,
			maxScore:  10,
		},
		{
			name:      "Paragraph with commas",
			text:      "First, second, third, fourth, and fifth items here.",
			minLength: 25,
			minScore:  4, // Base + commas
			maxScore:  15,
		},
		{
			name:      "Long paragraph",
			text:      strings.Repeat("This is a long paragraph with lots of content. ", 5),
			minLength: 25,
			minScore:  3, // Base + length bonus
			maxScore:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := "<p>" + tt.text + "</p>"
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatal(err)
			}

			sel := doc.Find("p")
			score := ScoreParagraph(sel, tt.minLength)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("ScoreParagraph = %f, want between %f and %f", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestGetTagScore(t *testing.T) {
	tests := []struct {
		tag      string
		expected float64
	}{
		{"div", DivBonus},
		{"td", TdBlockquoteBonus},
		{"blockquote", TdBlockquoteBonus},
		{"form", FormAddressPenalty},
		{"address", FormAddressPenalty},
		{"p", 0},
		{"span", 0},
		{"article", 0},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			result := GetTagScore(tt.tag)
			if result != tt.expected {
				t.Errorf("GetTagScore(%q) = %f, want %f", tt.tag, result, tt.expected)
			}
		})
	}
}

func TestGetSiblingThreshold(t *testing.T) {
	tests := []struct {
		topScore  float64
		minExpect float64
	}{
		{100, 25},                        // 100 * 0.25 = 25
		{40, SiblingScoreThresholdBase},  // 40 * 0.25 = 10, use base
		{20, SiblingScoreThresholdBase},  // 20 * 0.25 = 5 < 10, use base
		{200, 50},                        // 200 * 0.25 = 50
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := GetSiblingThreshold(tt.topScore)
			if result < tt.minExpect {
				t.Errorf("GetSiblingThreshold(%f) = %f, want >= %f", tt.topScore, result, tt.minExpect)
			}
		})
	}
}

func TestNodeScore(t *testing.T) {
	html := "<div>Test content</div>"
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	sel := doc.Find("div")
	ns := NewNodeScore(sel)

	// Test initial values
	if ns.ContentScore != 0 {
		t.Errorf("Initial ContentScore = %f, want 0", ns.ContentScore)
	}

	// Test AddScore
	ns.AddScore(10)
	if ns.ContentScore != 10 {
		t.Errorf("ContentScore after AddScore(10) = %f, want 10", ns.ContentScore)
	}

	ns.AddScore(5)
	if ns.ContentScore != 15 {
		t.Errorf("ContentScore after AddScore(5) = %f, want 15", ns.ContentScore)
	}

	// Test SetWeight
	ns.SetWeight(25)
	if ns.Weight != 25 {
		t.Errorf("Weight = %d, want 25", ns.Weight)
	}

	// Test GetWeightedScore
	weighted := ns.GetWeightedScore()
	if weighted != 40 { // 15 + 25
		t.Errorf("GetWeightedScore = %f, want 40", weighted)
	}
}

func TestScoreMap(t *testing.T) {
	html := "<div><p>Paragraph 1</p><p>Paragraph 2</p></div>"
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	sm := NewScoreMap()

	// Get scores for elements
	div := doc.Find("div")
	divScore := sm.Get(div)
	divScore.AddScore(50)

	p1 := doc.Find("p").First()
	p1Score := sm.Get(p1)
	p1Score.AddScore(20)

	// Test Size
	if sm.Size() != 2 {
		t.Errorf("ScoreMap Size = %d, want 2", sm.Size())
	}

	// Test GetTopCandidate
	top := sm.GetTopCandidate()
	if top != divScore {
		t.Error("GetTopCandidate should return div score")
	}

	// Test GetCandidatesByScore
	candidates := sm.GetCandidatesByScore()
	if len(candidates) != 2 {
		t.Errorf("GetCandidatesByScore length = %d, want 2", len(candidates))
	}
	if candidates[0].ContentScore != 50 {
		t.Error("First candidate should have highest score")
	}
}

func TestScorer_Score(t *testing.T) {
	html := `
<html>
<body>
<article>
	<p>This is the first paragraph with enough content to be scored properly.</p>
	<p>This is the second paragraph with additional meaningful content.</p>
	<p>This is the third paragraph completing the article structure.</p>
</article>
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	scorer := NewScorer(25, 100, false)
	topCandidate, scoreMap := scorer.Score(doc)

	if topCandidate == nil {
		t.Fatal("TopCandidate should not be nil")
	}

	if scoreMap.Size() == 0 {
		t.Error("ScoreMap should contain scored elements")
	}

	if topCandidate.GetScore() <= 0 {
		t.Error("TopCandidate should have positive score")
	}
}
