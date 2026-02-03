package scorer

// Scoring constants based on reference implementation.
const (
	// HNewsBonus is the score bonus for hNews microformat elements.
	HNewsBonus = 80

	// PositiveClassBonus is the score bonus for positive class/id matches.
	PositiveClassBonus = 25

	// NegativeClassPenalty is the score penalty for negative class/id matches.
	NegativeClassPenalty = -25

	// ParagraphBaseScore is the base score for a paragraph.
	ParagraphBaseScore = 1

	// CommaBonus is the score bonus per comma.
	CommaBonus = 1

	// LengthBonusPerChunk is the score bonus per 50 characters.
	LengthBonusPerChunk = 1

	// LengthChunkSize is the character count per length chunk.
	LengthChunkSize = 50

	// MaxLengthBonus is the maximum score bonus from length.
	MaxLengthBonus = 3

	// DivBonus is the score bonus for div elements.
	DivBonus = 5

	// TdBlockquoteBonus is the score bonus for td and blockquote elements.
	TdBlockquoteBonus = 3

	// FormAddressPenalty is the score penalty for form and address elements.
	FormAddressPenalty = -3

	// ParentScoreProportion is the proportion of score passed to parent (100%).
	ParentScoreProportion = 1.0

	// GrandparentScoreProportion is the proportion of score passed to grandparent (50%).
	GrandparentScoreProportion = 0.5
)

// Threshold constants.
const (
	// MinParagraphLength is the minimum paragraph length to score.
	MinParagraphLength = 25

	// MinContentLength is the minimum total content length.
	MinContentLength = 100

	// LowWeightLinkDensityMax is the maximum link density for low-weight elements.
	LowWeightLinkDensityMax = 0.2

	// HighWeightLinkDensityMax is the maximum link density for high-weight elements.
	HighWeightLinkDensityMax = 0.5

	// SiblingScoreThresholdBase is the base threshold for sibling merging.
	SiblingScoreThresholdBase = 10.0

	// SiblingScoreThresholdFactor is the factor of top score for sibling threshold.
	SiblingScoreThresholdFactor = 0.25
)

// Tag scoring map for quick lookup.
var tagScores = map[string]float64{
	"div":        DivBonus,
	"td":         TdBlockquoteBonus,
	"blockquote": TdBlockquoteBonus,
	"form":       FormAddressPenalty,
	"address":    FormAddressPenalty,
}

// GetTagScore returns the score adjustment for a tag.
func GetTagScore(tag string) float64 {
	if score, ok := tagScores[tag]; ok {
		return score
	}
	return 0
}
