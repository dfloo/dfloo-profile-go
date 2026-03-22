package repository

import "testing"

func TestGetConstructorColor_UsesFallbackForUnknown(t *testing.T) {
	got := getConstructorColor(999999)
	if got != f1FallbackColor {
		t.Fatalf("color = %q, want %q", got, f1FallbackColor)
	}
}

func TestFinalizeChampionshipDrivers_SortsByFinalPointsThenName(t *testing.T) {
	input := map[int]*championshipDriverAccumulator{
		2: {
			id:               2,
			name:             "Alex Albon",
			color:            f1FallbackColor,
			cumulativePoints: []float64{5, 10},
			raceResults:      nil,
			standingsByRace:  nil,
		},
		1: {
			id:               1,
			name:             "Max Verstappen",
			color:            f1FallbackColor,
			cumulativePoints: []float64{8, 20},
			raceResults:      nil,
			standingsByRace:  nil,
		},
		3: {
			id:               3,
			name:             "Charles Leclerc",
			color:            f1FallbackColor,
			cumulativePoints: []float64{5, 10},
			raceResults:      nil,
			standingsByRace:  nil,
		},
	}

	got := finalizeChampionshipDrivers(input)

	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3", len(got))
	}

	if got[0].Name != "Max Verstappen" {
		t.Fatalf("first driver = %q, want %q", got[0].Name, "Max Verstappen")
	}

	if got[1].Name != "Alex Albon" {
		t.Fatalf("second driver = %q, want %q", got[1].Name, "Alex Albon")
	}

	if got[2].Name != "Charles Leclerc" {
		t.Fatalf("third driver = %q, want %q", got[2].Name, "Charles Leclerc")
	}
}
