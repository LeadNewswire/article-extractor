package scorer

import (
	"github.com/example/article-extractor/internal/dom"
	"github.com/PuerkitoBio/goquery"
)

// ScoreParagraph calculates the content score for a paragraph.
func ScoreParagraph(sel *goquery.Selection, minLength int) float64 {
	text := dom.GetText(sel)
	textLen := len([]rune(text))

	// Skip paragraphs that are too short
	if textLen < minLength {
		return 0
	}

	// Base score
	score := float64(ParagraphBaseScore)

	// Bonus for commas
	commas := dom.CountCommas(text)
	score += float64(commas) * CommaBonus

	// Bonus for length (1 point per 50 chars, max 3)
	lengthBonus := textLen / LengthChunkSize
	if lengthBonus > MaxLengthBonus {
		lengthBonus = MaxLengthBonus
	}
	score += float64(lengthBonus)

	return score
}

// ScoreParagraphWithSelection calculates score and returns detailed info.
func ScoreParagraphWithSelection(sel *goquery.Selection, minLength int) *ParagraphScore {
	text := dom.GetText(sel)
	textLen := len([]rune(text))

	ps := &ParagraphScore{
		Selection:  sel,
		Text:       text,
		TextLength: textLen,
		Score:      0,
	}

	// Skip paragraphs that are too short
	if textLen < minLength {
		return ps
	}

	// Base score
	ps.Score = float64(ParagraphBaseScore)

	// Bonus for commas
	commas := dom.CountCommas(text)
	ps.CommaCount = commas
	ps.Score += float64(commas) * CommaBonus

	// Bonus for length
	lengthBonus := textLen / LengthChunkSize
	if lengthBonus > MaxLengthBonus {
		lengthBonus = MaxLengthBonus
	}
	ps.LengthBonus = lengthBonus
	ps.Score += float64(lengthBonus)

	return ps
}

// ParagraphScore holds detailed paragraph scoring information.
type ParagraphScore struct {
	Selection   *goquery.Selection
	Text        string
	TextLength  int
	CommaCount  int
	LengthBonus int
	Score       float64
}

// ScoreAllParagraphs scores all paragraphs in a document.
func ScoreAllParagraphs(doc *goquery.Document, minLength int) []*ParagraphScore {
	var scores []*ParagraphScore

	// Find all paragraph-like elements
	doc.Find("p, pre").Each(func(_ int, sel *goquery.Selection) {
		ps := ScoreParagraphWithSelection(sel, minLength)
		if ps.Score > 0 {
			scores = append(scores, ps)
		}
	})

	return scores
}

// GetParagraphParent returns the parent element that should receive the score.
func GetParagraphParent(sel *goquery.Selection) *goquery.Selection {
	parent := sel.Parent()
	if parent.Length() == 0 {
		return nil
	}
	return parent
}

// GetParagraphGrandparent returns the grandparent element.
func GetParagraphGrandparent(sel *goquery.Selection) *goquery.Selection {
	parent := sel.Parent()
	if parent.Length() == 0 {
		return nil
	}
	grandparent := parent.Parent()
	if grandparent.Length() == 0 {
		return nil
	}
	return grandparent
}
