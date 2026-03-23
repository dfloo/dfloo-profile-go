package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
)

type MockF1Repository struct {
	GetAvailableYearsFunc   func(ctx context.Context) ([]int, error)
	GetChampionshipByYearFn func(ctx context.Context, year int) (*model.F1ChampionshipData, error)
	GetDriversByYearFn      func(ctx context.Context, year int) ([]model.F1DriverStanding, error)
	GetDriverDetailsFn      func(ctx context.Context, driverID int, year int, raceID *int) (*model.DriverDetailData, error)
	GetConstructorsByYearFn func(ctx context.Context, year int) ([]model.F1ConstructorStanding, error)
	GetEventsByYearFn       func(ctx context.Context, year int) ([]model.F1Event, error)
}

func (m *MockF1Repository) GetAvailableYears(ctx context.Context) ([]int, error) {
	if m.GetAvailableYearsFunc != nil {
		return m.GetAvailableYearsFunc(ctx)
	}
	return nil, errors.New("GetAvailableYears not implemented in mock")
}

func (m *MockF1Repository) GetChampionshipByYear(
	ctx context.Context,
	year int,
) (*model.F1ChampionshipData, error) {
	if m.GetChampionshipByYearFn != nil {
		return m.GetChampionshipByYearFn(ctx, year)
	}
	return nil, errors.New("GetChampionshipByYear not implemented in mock")
}

func (m *MockF1Repository) GetDriversByYear(
	ctx context.Context,
	year int,
) ([]model.F1DriverStanding, error) {
	if m.GetDriversByYearFn != nil {
		return m.GetDriversByYearFn(ctx, year)
	}
	return nil, errors.New("GetDriversByYear not implemented in mock")
}

func (m *MockF1Repository) GetDriverDetails(
	ctx context.Context,
	driverID int,
	year int,
	raceID *int,
) (*model.DriverDetailData, error) {
	if m.GetDriverDetailsFn != nil {
		return m.GetDriverDetailsFn(ctx, driverID, year, raceID)
	}
	return nil, errors.New("GetDriverDetails not implemented in mock")
}

func (m *MockF1Repository) GetConstructorsByYear(
	ctx context.Context,
	year int,
) ([]model.F1ConstructorStanding, error) {
	if m.GetConstructorsByYearFn != nil {
		return m.GetConstructorsByYearFn(ctx, year)
	}
	return nil, errors.New("GetConstructorsByYear not implemented in mock")
}

func (m *MockF1Repository) GetEventsByYear(
	ctx context.Context,
	year int,
) ([]model.F1Event, error) {
	if m.GetEventsByYearFn != nil {
		return m.GetEventsByYearFn(ctx, year)
	}
	return nil, errors.New("GetEventsByYear not implemented in mock")
}

func TestGetChampionships_DefaultLatestYearSuccess(t *testing.T) {
	repo := &MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024, 2023, 2022}, nil
		},
		GetChampionshipByYearFn: func(ctx context.Context, year int) (*model.F1ChampionshipData, error) {
			if year != 2024 {
				t.Fatalf("year = %d, want %d", year, 2024)
			}

			return &model.F1ChampionshipData{
				Year: 2024,
				Events: []model.F1Event{
					{RaceID: "1", Round: 1, Name: "Bahrain GP"},
					{RaceID: "2", Round: 2, Name: "Saudi Arabian GP"},
				},
				Drivers: []model.F1ChampionshipDriver{
					{
						ID:               "1",
						Name:             "Max Verstappen",
						Color:            "#3671C6",
						CumulativePoints: []float64{25, 43},
						RaceResults: []model.F1RaceResult{
							{
								ConstructorID:    stringPtr("9"),
								ConstructorName:  stringPtr("Red Bull"),
								ConstructorColor: "#3671C6",
								RacePoints:       25,
							},
							{
								ConstructorID:    stringPtr("9"),
								ConstructorName:  stringPtr("Red Bull"),
								ConstructorColor: "#3671C6",
								RacePoints:       18,
							},
						},
					},
				},
			}, nil
		},
	}

	h := NewF1Handler(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/f1/championships", nil)
	w := httptest.NewRecorder()

	h.GetChampionships(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var got model.F1ChampionshipsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Data.Year != 2024 {
		t.Fatalf("data.year = %d, want 2024", got.Data.Year)
	}

	raceCount := len(got.Data.Events)
	for _, driver := range got.Data.Drivers {
		if len(driver.CumulativePoints) != raceCount {
			t.Fatalf("cumulativePoints length = %d, want %d", len(driver.CumulativePoints), raceCount)
		}
		if len(driver.RaceResults) != raceCount {
			t.Fatalf("raceResults length = %d, want %d", len(driver.RaceResults), raceCount)
		}
	}
}

func TestGetChampionships_InvalidYearFormat(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/championships?year=abc", nil)
	w := httptest.NewRecorder()

	h.GetChampionships(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Invalid year format.")
}

func TestGetChampionships_InvalidYearFormat_DoesNotDependOnDB(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/championships?year=abc", nil)
	w := httptest.NewRecorder()

	h.GetChampionships(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Invalid year format.")
}

func TestGetChampionships_UnsupportedYear(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024, 2023}, nil
		},
		GetChampionshipByYearFn: func(ctx context.Context, year int) (*model.F1ChampionshipData, error) {
			return nil, repository.ErrF1YearNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/championships?year=1999", nil)
	w := httptest.NewRecorder()

	h.GetChampionships(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Unsupported championship year.")
}

func TestGetDrivers_MissingYear(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers", nil)
	w := httptest.NewRecorder()

	h.GetDrivers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Missing year query parameter.")
}

func TestGetDrivers_InvalidYearFormat_DoesNotDependOnDB(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers?year=abc", nil)
	w := httptest.NewRecorder()

	h.GetDrivers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Invalid year format.")
}

func TestGetDrivers_Success(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024, 2023}, nil
		},
		GetDriversByYearFn: func(ctx context.Context, year int) ([]model.F1DriverStanding, error) {
			if year != 2024 {
				t.Fatalf("year = %d, want %d", year, 2024)
			}
			return []model.F1DriverStanding{
				{ID: "1", Name: "Max Verstappen", LatestPoints: 437},
				{ID: "16", Name: "Charles Leclerc", LatestPoints: 356},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers?year=2024", nil)
	w := httptest.NewRecorder()

	h.GetDrivers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var got model.F1DriversResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Data.Year != 2024 {
		t.Fatalf("data.year = %d, want %d", got.Data.Year, 2024)
	}
	if len(got.Data.Drivers) != 2 {
		t.Fatalf("drivers length = %d, want 2", len(got.Data.Drivers))
	}
}

func TestGetDrivers_UnsupportedYear(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024}, nil
		},
		GetDriversByYearFn: func(ctx context.Context, year int) ([]model.F1DriverStanding, error) {
			return nil, repository.ErrF1YearNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers?year=1999", nil)
	w := httptest.NewRecorder()

	h.GetDrivers(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Unsupported championship year.")
}

func TestGetDrivers_InternalFailure(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers?year=2024", nil)
	w := httptest.NewRecorder()

	h.GetDrivers(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Internal server error.")
}

func TestGetDriverDetails_MissingDriverID(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers/?year=2024", nil)
	w := httptest.NewRecorder()

	h.GetDriverDetails(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Missing driver id path parameter.")
}

func TestGetDriverDetails_InvalidDriverID_DoesNotDependOnDB(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers/abc?year=2024", nil)
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()

	h.GetDriverDetails(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Invalid driver id format.")
}

func TestGetDriverDetails_MissingYear(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.GetDriverDetails(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Missing year query parameter.")
}

func TestGetDriverDetails_InvalidYearFormat_DoesNotDependOnDB(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers/1?year=abc", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.GetDriverDetails(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Invalid year format.")
}

func TestGetDriverDetails_InvalidRaceID_DoesNotDependOnDB(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers/1?year=2024&raceId=abc", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.GetDriverDetails(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Invalid raceId format.")
}

func TestGetDriverDetails_Success_DefaultRace(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024, 2023}, nil
		},
		GetDriverDetailsFn: func(ctx context.Context, driverID int, year int, raceID *int) (*model.DriverDetailData, error) {
			if driverID != 1 {
				t.Fatalf("driverID = %d, want 1", driverID)
			}
			if year != 2024 {
				t.Fatalf("year = %d, want 2024", year)
			}
			if raceID != nil {
				t.Fatalf("raceID = %v, want nil", *raceID)
			}

			return &model.DriverDetailData{
				Year: 2024,
				Driver: model.DriverDetailHeader{
					ID:              "1",
					Name:            "Max Verstappen",
					Number:          stringPtr("1"),
					ConstructorName: stringPtr("Red Bull"),
					CurrentPoints:   43,
				},
				Races: []model.DriverRaceOption{
					{ID: "1", Round: 1, Name: "Bahrain GP"},
					{ID: "2", Round: 2, Name: "Saudi Arabian GP"},
				},
				SelectedRace: &model.DriverSelectedRaceContext{
					ID:               "2",
					Round:            2,
					Name:             "Saudi Arabian GP",
					RacePoints:       18,
					CumulativePoints: 43,
					StartingPosition: intPtr(1),
					EndingPosition:   intPtr(1),
					RaceScore:        97.4,
					SeasonScore:      100,
					Qualifying: &model.DriverQualifyingBreakdown{
						Q1: stringPtr("1:28.171"),
						Q2: stringPtr("1:27.567"),
						Q3: stringPtr("1:27.241"),
					},
					LapTimes: []model.DriverLapTimePoint{
						{Lap: 1, Time: intPtr(95111), MinTime: intPtr(94300), MaxTime: intPtr(98900), AvgTime: floatPtr(95600.5)},
					},
				},
				SeasonPoints: []model.DriverSeasonPointsPoint{
					{RaceID: "1", Round: 1, Name: "Bahrain GP", RacePoints: 25, CumulativePoints: 25},
					{RaceID: "2", Round: 2, Name: "Saudi Arabian GP", RacePoints: 18, CumulativePoints: 43},
				},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers/1?year=2024", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.GetDriverDetails(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var got model.DriverDetailResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Data.Driver.ID != "1" {
		t.Fatalf("driver.id = %s, want 1", got.Data.Driver.ID)
	}
	if got.Data.SelectedRace == nil || got.Data.SelectedRace.ID != "2" {
		t.Fatalf("selectedRace = %+v, want id=2", got.Data.SelectedRace)
	}
	if len(got.AvailableYears) != 2 {
		t.Fatalf("availableYears length = %d, want 2", len(got.AvailableYears))
	}
}

func TestGetDriverDetails_UnsupportedYear(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024}, nil
		},
		GetDriverDetailsFn: func(ctx context.Context, driverID int, year int, raceID *int) (*model.DriverDetailData, error) {
			return nil, repository.ErrF1YearNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers/1?year=1999", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.GetDriverDetails(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Unsupported championship year.")
}

func TestGetDriverDetails_DriverNotFound(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024}, nil
		},
		GetDriverDetailsFn: func(ctx context.Context, driverID int, year int, raceID *int) (*model.DriverDetailData, error) {
			return nil, repository.ErrF1DriverNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers/1?year=2024", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.GetDriverDetails(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Driver not found for championship year.")
}

func TestGetDriverDetails_RaceNotFound(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024}, nil
		},
		GetDriverDetailsFn: func(ctx context.Context, driverID int, year int, raceID *int) (*model.DriverDetailData, error) {
			return nil, repository.ErrF1RaceNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers/1?year=2024&raceId=9999", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.GetDriverDetails(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Race not found for championship year.")
}

func TestGetDriverDetails_InternalFailure(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/drivers/1?year=2024", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.GetDriverDetails(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Internal server error.")
}

func TestGetConstructors_MissingYear(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/constructors", nil)
	w := httptest.NewRecorder()

	h.GetConstructors(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Missing year query parameter.")
}

func TestGetConstructors_InvalidYearFormat_DoesNotDependOnDB(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/constructors?year=abc", nil)
	w := httptest.NewRecorder()

	h.GetConstructors(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Invalid year format.")
}

func TestGetConstructors_Success(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024, 2023}, nil
		},
		GetConstructorsByYearFn: func(ctx context.Context, year int) ([]model.F1ConstructorStanding, error) {
			if year != 2024 {
				t.Fatalf("year = %d, want %d", year, 2024)
			}
			return []model.F1ConstructorStanding{
				{ID: "9", Name: "Red Bull", Color: "#3671C6", LatestPoints: 860},
				{ID: "6", Name: "Ferrari", Color: "#FF8700", LatestPoints: 652},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/constructors?year=2024", nil)
	w := httptest.NewRecorder()

	h.GetConstructors(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var got model.F1ConstructorsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Data.Year != 2024 {
		t.Fatalf("data.year = %d, want %d", got.Data.Year, 2024)
	}
	if len(got.Data.Constructors) != 2 {
		t.Fatalf("constructors length = %d, want 2", len(got.Data.Constructors))
	}
}

func TestGetConstructors_UnsupportedYear(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024}, nil
		},
		GetConstructorsByYearFn: func(ctx context.Context, year int) ([]model.F1ConstructorStanding, error) {
			return nil, repository.ErrF1YearNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/constructors?year=1999", nil)
	w := httptest.NewRecorder()

	h.GetConstructors(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Unsupported championship year.")
}

func TestGetConstructors_InternalFailure(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/constructors?year=2024", nil)
	w := httptest.NewRecorder()

	h.GetConstructors(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Internal server error.")
}

func TestGetEvents_MissingYear(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/events", nil)
	w := httptest.NewRecorder()

	h.GetEvents(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Missing year query parameter.")
}

func TestGetEvents_InvalidYearFormat_DoesNotDependOnDB(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/events?year=abc", nil)
	w := httptest.NewRecorder()

	h.GetEvents(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Invalid year format.")
}

func TestGetEvents_Success(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024, 2023}, nil
		},
		GetEventsByYearFn: func(ctx context.Context, year int) ([]model.F1Event, error) {
			if year != 2024 {
				t.Fatalf("year = %d, want %d", year, 2024)
			}
			return []model.F1Event{
				{RaceID: "1", Round: 1, Name: "Bahrain Grand Prix"},
				{RaceID: "2", Round: 2, Name: "Saudi Arabian Grand Prix"},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/events?year=2024", nil)
	w := httptest.NewRecorder()

	h.GetEvents(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var got model.F1EventsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Data.Year != 2024 {
		t.Fatalf("data.year = %d, want %d", got.Data.Year, 2024)
	}

	if len(got.Data.Events) != 2 {
		t.Fatalf("events length = %d, want 2", len(got.Data.Events))
	}

	if got.Data.Events[0].RaceID != "1" || got.Data.Events[0].Round != 1 || got.Data.Events[0].Name != "Bahrain Grand Prix" {
		t.Fatalf("first event = %+v, want raceId=1 round=1 name=Bahrain Grand Prix", got.Data.Events[0])
	}
}

func TestGetEvents_UnsupportedYear(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return []int{2024}, nil
		},
		GetEventsByYearFn: func(ctx context.Context, year int) ([]model.F1Event, error) {
			return nil, repository.ErrF1YearNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/events?year=1999", nil)
	w := httptest.NewRecorder()

	h.GetEvents(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Unsupported championship year.")
}

func TestGetEvents_InternalFailure(t *testing.T) {
	h := NewF1Handler(&MockF1Repository{
		GetAvailableYearsFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/f1/events?year=2024", nil)
	w := httptest.NewRecorder()

	h.GetEvents(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	assertJSONErrorMessage(t, w.Body.Bytes(), "Internal server error.")
}

func assertJSONErrorMessage(t *testing.T, body []byte, want string) {
	t.Helper()

	var got model.APIErrorResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if got.Message != want {
		t.Fatalf("message = %q, want %q", got.Message, want)
	}
}

func stringPtr(v string) *string {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}
