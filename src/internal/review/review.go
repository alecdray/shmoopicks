package review

import "math"

const (
	// Controls how much extreme answers (1 or 5) are amplified.
	// Lower values = more powerful extremes, higher values = closer to linear.
	// Valid range 0.0–1.0, where 1.0 is fully linear.
	scoreCurveExponent = 0.8

	// Minimum possible score — awarded when all answers are 1.
	scoreFloor = 2.0

	// Maximum possible score — awarded when all answers are 5.
	scoreCeiling = 10.0

	// Weighting for Q1: track-by-track consistency.
	// Controls how much skips and weak tracks pull the score down.
	ScoreWeightConsistency = 0.3

	// Weighting for Q2: emotional impact while listening.
	// Highest weight — the primary differentiator between good and great.
	ScoreWeightImpact = 0.4

	// Weighting for Q3: immediate gut reaction when the album ended.
	// Captures the overall impression beyond individual tracks.
	ScoreWeightGutCheck = 0.3
)

type QuestionScore struct {
	Value  int
	Weight float64
}

func curveAnswer(answer int) float64 {
	normalized := (float64(answer) - 3.0) / 2.0
	curved := math.Copysign(math.Pow(math.Abs(normalized), scoreCurveExponent), normalized)
	return curved*2.0 + 3.0
}

func CalculateResponseScore(scores ...QuestionScore) float64 {
	var raw float64
	for _, score := range scores {
		raw += curveAnswer(score.Value) * score.Weight
	}
	score := scoreFloor + ((raw-1.0)/4.0)*(scoreCeiling-scoreFloor)
	return math.Round(score*10) / 10
}

type ScoreLabel string

const (
	ScoreLabelDNR                ScoreLabel = "DNR"
	ScoreLabelNope               ScoreLabel = "Nope"
	ScoreLabelNotForMe           ScoreLabel = "Not For Me"
	ScoreLabelHasItsM0ments      ScoreLabel = "Has Its Moments"
	ScoreLabelGoodNotGreat       ScoreLabel = "Good Not Great"
	ScoreLabelWouldRecommend     ScoreLabel = "Would Recommend"
	ScoreLabelEssentialListening ScoreLabel = "Essential Listening"
	ScoreLabelInstantClassic     ScoreLabel = "Instant Classic"
	ScoreLabelMasterpiece        ScoreLabel = "Masterpiece"
)

func GetScoreLabel(score float64) ScoreLabel {
	switch {
	case score < 2.0:
		return ScoreLabelDNR
	case score < 4.0:
		return ScoreLabelNope
	case score < 6.0:
		return ScoreLabelNotForMe
	case score < 6.6:
		return ScoreLabelHasItsM0ments
	case score < 7.0:
		return ScoreLabelGoodNotGreat
	case score < 8.0:
		return ScoreLabelWouldRecommend
	case score < 9.0:
		return ScoreLabelEssentialListening
	case score < 10.0:
		return ScoreLabelInstantClassic
	default:
		return ScoreLabelMasterpiece
	}
}
