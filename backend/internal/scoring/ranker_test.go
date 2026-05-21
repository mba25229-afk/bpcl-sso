package scoring

import (
	"testing"
)

func TestRank(t *testing.T) {
	actuals := []Actuals{
		{CCCode: "A", MSVolKL: 100},
		{CCCode: "B", MSVolKL: 200},
		{CCCode: "C", MSVolKL: 50},
		{CCCode: "D", MSVolKL: 150},
	}

	ranks := Rank(actuals, "ms_absolute_vol", 40)

	if ranks["A"] != 3 {
		t.Errorf("rank A = %d, want 3", ranks["A"])
	}
	if ranks["B"] != 1 {
		t.Errorf("rank B = %d, want 1", ranks["B"])
	}
	if ranks["C"] != 4 {
		t.Errorf("rank C = %d, want 4", ranks["C"])
	}
	if ranks["D"] != 2 {
		t.Errorf("rank D = %d, want 2", ranks["D"])
	}
}

func TestRank_GrowthPct(t *testing.T) {
	actuals := []Actuals{
		{CCCode: "A", MSVolKL: 110, MSLYKl: 100},  // +10%
		{CCCode: "B", MSVolKL: 150, MSLYKl: 100},  // +50%
		{CCCode: "C", MSVolKL: 100, MSLYKl: 100},  // 0%
	}

	ranks := Rank(actuals, "ms_growth_pct", 40)

	if ranks["A"] != 2 {
		t.Errorf("rank A = %d, want 2", ranks["A"])
	}
	if ranks["B"] != 1 {
		t.Errorf("rank B = %d, want 1", ranks["B"])
	}
	if ranks["C"] != 3 {
		t.Errorf("rank C = %d, want 3", ranks["C"])
	}
}

func TestRank_MAKGE_InsufficientData(t *testing.T) {
	actuals := []Actuals{
		{CCCode: "A", MAKGESalesKL: 10, MAKGEHasData: true},
		{CCCode: "B", MAKGESalesKL: 20, MAKGEHasData: true},
		{CCCode: "C", MAKGESalesKL: 0, MAKGEHasData: false}, // insufficient
	}

	ranks := Rank(actuals, "mak_ge_sales", 40)

	if ranks["A"] != 2 {
		t.Errorf("rank A = %d, want 2", ranks["A"])
	}
	if ranks["B"] != 1 {
		t.Errorf("rank B = %d, want 1", ranks["B"])
	}
	// C has MAKGEHasData=false, so gets -1 sentinel value, ranks last
	if ranks["C"] != 3 {
		t.Errorf("rank C = %d, want 3 (insufficient data ranks last)", ranks["C"])
	}
}