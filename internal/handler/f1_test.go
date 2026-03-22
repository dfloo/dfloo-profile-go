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
				Races: []model.F1Race{
					{Round: 1, Name: "Bahrain GP"},
					{Round: 2, Name: "Saudi Arabian GP"},
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

	raceCount := len(got.Data.Races)
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
