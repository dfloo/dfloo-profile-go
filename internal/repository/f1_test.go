package repository

import (
	"context"
	"testing"

	"github.com/dfloo/dfloo-profile-go/internal/model"
)

func TestGetEventsByYear_NilPool(t *testing.T) {
	repo := &DBF1Repository{}

	_, err := repo.GetEventsByYear(context.Background(), 2024)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "database pool is nil" {
		t.Fatalf("error = %v, want database pool is nil", err)
	}
}

func TestGetDriverDetails_NilPool(t *testing.T) {
	repo := &DBF1Repository{}

	_, err := repo.GetDriverDetails(context.Background(), 1, 2024, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "database pool is nil" {
		t.Fatalf("error = %v, want database pool is nil", err)
	}
}

func TestF1DriverAndRaceErrors_AreDistinct(t *testing.T) {
	if ErrF1DriverNotFound == ErrF1RaceNotFound {
		t.Fatalf("expected distinct sentinel errors")
	}
}

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

func TestFinalizeConstructorStandings_SortsByPointsThenName(t *testing.T) {
	input := []model.F1ConstructorStanding{
		{ID: "6", Name: "Ferrari", Color: "#FF8700", LatestPoints: 652},
		{ID: "9", Name: "Red Bull", Color: "#3671C6", LatestPoints: 860},
		{ID: "117", Name: "Aston Martin", Color: "#52E252", LatestPoints: 652},
	}

	got := finalizeConstructorStandings(input)

	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3", len(got))
	}

	if got[0].Name != "Red Bull" {
		t.Fatalf("first constructor = %q, want %q", got[0].Name, "Red Bull")
	}

	if got[1].Name != "Aston Martin" {
		t.Fatalf("second constructor = %q, want %q", got[1].Name, "Aston Martin")
	}

	if got[2].Name != "Ferrari" {
		t.Fatalf("third constructor = %q, want %q", got[2].Name, "Ferrari")
	}
}
