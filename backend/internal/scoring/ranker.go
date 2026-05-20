package scoring

import "sort"

func metricValue(a Actuals, metricKey string) float64 {
	switch metricKey {
	case "ms_absolute_vol":
		return a.MSVolKL
	case "ms_growth_pct":
		return GrowthPct(a.MSVolKL, a.MSLYKl)
	case "hsd_absolute_vol":
		return a.HSDVolKL
	case "hsd_growth_pct":
		return GrowthPct(a.HSDVolKL, a.HSDLYKl)
	case "oil_change_count":
		return a.QOCCount
	case "mak_ge_sales":
		if !a.MAKGEHasData {
			return -1
		}
		return a.MAKGESalesKL
	case "ufill_txns":
		return a.UFillCount
	case "speed_vol":
		return a.SpeedVolKL
	case "sangam_certs":
		return a.SangamCerts
	case "google_rating":
		return a.GoogleComposite
	default:
		return 0
	}
}

func Rank(actuals []Actuals, metricKey string, n int) map[string]int {
	type entry struct {
		ccCode string
		value  float64
	}

	entries := make([]entry, len(actuals))
	for i, a := range actuals {
		entries[i] = entry{a.CCCode, metricValue(a, metricKey)}
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].value > entries[j].value
	})

	ranks := make(map[string]int, len(entries))
	for i, e := range entries {
		ranks[e.ccCode] = i + 1
	}

	return ranks
}