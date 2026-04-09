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

func TestBuildDriverSelectedRaceChartData_ComputesSeriesInLapOrder(t *testing.T) {
	lapStats := []driverLapStatsPoint{
		{Lap: 2, Time: intPtrForRepoTest(97000), MinTime: intPtrForRepoTest(96000), MaxTime: intPtrForRepoTest(99000), AvgTime: floatPtrForRepoTest(97367), Position: intPtrForRepoTest(1)},
		{Lap: 1, Time: intPtrForRepoTest(98120), MinTime: intPtrForRepoTest(96500), MaxTime: intPtrForRepoTest(99800), AvgTime: floatPtrForRepoTest(98200), Position: intPtrForRepoTest(1)},
		{Lap: 3, Time: intPtrForRepoTest(96900), MinTime: intPtrForRepoTest(96050), MaxTime: intPtrForRepoTest(99050), AvgTime: floatPtrForRepoTest(97000), Position: intPtrForRepoTest(2)},
	}
	teammatePositions := map[int]*int{
		1: intPtrForRepoTest(4),
		2: intPtrForRepoTest(3),
	}

	got := buildDriverSelectedRaceChartData(lapStats, teammatePositions)

	if got == nil {
		t.Fatalf("chart data = nil, want non-nil")
	}

	if len(got.PaceDeltaVsAverageMs) != 3 {
		t.Fatalf("paceDeltaVsAverageMs length = %d, want 3", len(got.PaceDeltaVsAverageMs))
	}
	if got.PaceDeltaVsAverageMs[0].Lap != 1 || valueOrDefault(got.PaceDeltaVsAverageMs[0].Value) != -80 {
		t.Fatalf("lap1 paceDelta = %+v, want lap=1 value=-80", got.PaceDeltaVsAverageMs[0])
	}
	if got.PaceDeltaVsAverageMs[1].Lap != 2 || valueOrDefault(got.PaceDeltaVsAverageMs[1].Value) != -367 {
		t.Fatalf("lap2 paceDelta = %+v, want lap=2 value=-367", got.PaceDeltaVsAverageMs[1])
	}

	if len(got.GapToFastestMs) != 3 {
		t.Fatalf("gapToFastestMs length = %d, want 3", len(got.GapToFastestMs))
	}
	if got.GapToFastestMs[0].Lap != 1 || valueOrDefault(got.GapToFastestMs[0].Value) != 1620 {
		t.Fatalf("lap1 gapToFastest = %+v, want lap=1 value=1620", got.GapToFastestMs[0])
	}

	if len(got.Rolling3LapPaceMs) != 3 {
		t.Fatalf("rolling3LapPaceMs length = %d, want 3", len(got.Rolling3LapPaceMs))
	}
	if got.Rolling3LapPaceMs[0].Value != nil {
		t.Fatalf("lap1 rolling3LapPace value = %+v, want nil", got.Rolling3LapPaceMs[0].Value)
	}
	if got.Rolling3LapPaceMs[1].Value != nil {
		t.Fatalf("lap2 rolling3LapPace value = %+v, want nil", got.Rolling3LapPaceMs[1].Value)
	}
	if got.Rolling3LapPaceMs[2].Lap != 3 || valueOrDefault(got.Rolling3LapPaceMs[2].Value) != 97340 {
		t.Fatalf("lap3 rolling3LapPace = %+v, want lap=3 value=97340", got.Rolling3LapPaceMs[2])
	}

	if len(got.PositionsByLap) != 3 {
		t.Fatalf("positionsByLap length = %d, want 3", len(got.PositionsByLap))
	}
	if got.PositionsByLap[0].Lap != 1 || valueOrDefault(got.PositionsByLap[0].DriverPosition) != 1 || valueOrDefault(got.PositionsByLap[0].TeammatePosition) != 4 {
		t.Fatalf("lap1 positionsByLap = %+v, want lap=1 driver=1 teammate=4", got.PositionsByLap[0])
	}
	if got.PositionsByLap[2].Lap != 3 || valueOrDefault(got.PositionsByLap[2].DriverPosition) != 2 || got.PositionsByLap[2].TeammatePosition != nil {
		t.Fatalf("lap3 positionsByLap = %+v, want lap=3 driver=2 teammate=nil", got.PositionsByLap[2])
	}
}

func TestBuildDriverSelectedRaceChartData_NoLapDataReturnsNullSeries(t *testing.T) {
	got := buildDriverSelectedRaceChartData(nil, nil)

	if got == nil {
		t.Fatalf("chart data = nil, want non-nil")
	}
	if got.PaceDeltaVsAverageMs != nil {
		t.Fatalf("paceDeltaVsAverageMs = %+v, want nil", got.PaceDeltaVsAverageMs)
	}
	if got.GapToFastestMs != nil {
		t.Fatalf("gapToFastestMs = %+v, want nil", got.GapToFastestMs)
	}
	if got.Rolling3LapPaceMs != nil {
		t.Fatalf("rolling3LapPaceMs = %+v, want nil", got.Rolling3LapPaceMs)
	}
	if got.PositionsByLap != nil {
		t.Fatalf("positionsByLap = %+v, want nil", got.PositionsByLap)
	}
}

func intPtrForRepoTest(v int) *int {
	return &v
}

func floatPtrForRepoTest(v float64) *float64 {
	return &v
}

func valueOrDefault(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
