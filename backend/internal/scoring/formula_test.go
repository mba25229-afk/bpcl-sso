package scoring

import (
	"testing"
)

func TestStandardScore(t *testing.T) {
	cases := []struct {
		rank, n  int
		max      float64
		expected float64
	}{
		{1, 40, 10, 10.0},  // rank 1 always gets full marks
		{40, 40, 10, 0.0},  // rank 40 always gets zero
		{4, 40, 10, 9.23},  // MANN — verified against v4 template
		{2, 40, 10, 9.74},  // AUTO CARE — verified against v4 template
	}
	for _, c := range cases {
		got := StandardScore(c.rank, c.n, c.max)
		if got != c.expected {
			t.Errorf("StandardScore(%d,%d,%.0f) = %.2f, want %.2f",
				c.rank, c.n, c.max, got, c.expected)
		}
	}
}

func TestNegativeScaleScore(t *testing.T) {
	cases := []struct {
		rank, n  int
		max      float64
		expected float64
	}{
		{1, 40, 5, 5.0},   // best growth → full +5
		{40, 40, 5, -5.0}, // worst growth → full -5
		{20, 40, 5, 0.13}, // median → near zero
	}
	for _, c := range cases {
		got := NegativeScaleScore(c.rank, c.n, c.max)
		if got != c.expected {
			t.Errorf("NegativeScaleScore(%d,%d,%.0f) = %.2f, want %.2f",
				c.rank, c.n, c.max, got, c.expected)
		}
	}
}

func TestCleanlinessScore(t *testing.T) {
	tests := []struct {
		grade    string
		expected float64
	}{
		{"Excellent", 5},
		{"Good", 4},
		{"Average", 3},
		{"Below Average", 1},
		{"Poor", 0},
		{"", 0},
		{"Invalid", 0},
	}
	for _, tt := range tests {
		got := CleanlinessScore(tt.grade)
		if got != tt.expected {
			t.Errorf("CleanlinessScore(%q) = %.2f, want %.2f", tt.grade, got, tt.expected)
		}
	}
}

func TestGrowthPct(t *testing.T) {
	tests := []struct {
		actual, ly   float64
		expected     float64
	}{
		{552.5, 507.5, 8.87},
		{100, 50, 100.0},
		{100, 0, 0},     // no division by zero
		{0, 100, -100},  // negative growth
		{110, 100, 10.0},
	}
	for _, tt := range tests {
		got := GrowthPct(tt.actual, tt.ly)
		if got != tt.expected {
			t.Errorf("GrowthPct(%.2f, %.2f) = %.2f, want %.2f", tt.actual, tt.ly, got, tt.expected)
		}
	}
}