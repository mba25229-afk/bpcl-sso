package scoring

import "math"

func StandardScore(rank, n int, maxMarks float64) float64 {
	if n <= 1 {
		return maxMarks
	}
	return round2((float64(n-rank) / float64(n-1)) * maxMarks)
}

func NegativeScaleScore(rank, n int, maxMarks float64) float64 {
	if n <= 1 {
		return 0
	}
	return round2(((float64(n-rank)/float64(n-1))*2*maxMarks) - maxMarks)
}

func CleanlinessScore(grade string) float64 {
	grades := map[string]float64{
		"Excellent":     5,
		"Good":          4,
		"Average":       3,
		"Below Average": 1,
		"Poor":          0,
	}
	if s, ok := grades[grade]; ok {
		return s
	}
	return 0
}

func GrowthPct(actual, ly float64) float64 {
	if ly == 0 {
		return 0
	}
	return round2(((actual - ly) / ly) * 100)
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}