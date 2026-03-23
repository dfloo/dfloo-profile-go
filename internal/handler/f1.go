package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/dfloo/dfloo-profile-go/internal/repository"
)

type F1Handler struct {
	Repo repository.F1Repository
}

func NewF1Handler(repo repository.F1Repository) *F1Handler {
	return &F1Handler{Repo: repo}
}

func (h *F1Handler) GetChampionships(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	yearParam := r.URL.Query().Get("year")
	year, hasYear, statusCode, message := parseOptionalYear(yearParam)
	if statusCode != 0 {
		h.writeError(w, statusCode, message)
		return
	}

	availableYears, err := h.Repo.GetAvailableYears(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	if !hasYear {
		year, statusCode, message = parseDefaultYear(availableYears)
		if statusCode != 0 {
			h.writeError(w, statusCode, message)
			return
		}
	}

	championshipData, err := h.Repo.GetChampionshipByYear(r.Context(), year)
	if err != nil {
		if errors.Is(err, repository.ErrF1YearNotFound) {
			h.writeError(w, http.StatusNotFound, "Unsupported championship year.")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.F1ChampionshipsResponse{
		AvailableYears: availableYears,
		Data:           *championshipData,
	})
}

func (h *F1Handler) GetDrivers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	yearParam := r.URL.Query().Get("year")
	if yearParam == "" {
		h.writeError(w, http.StatusBadRequest, "Missing year query parameter.")
		return
	}

	year, _, statusCode, message := parseOptionalYear(yearParam)
	if statusCode != 0 {
		h.writeError(w, statusCode, message)
		return
	}

	availableYears, err := h.Repo.GetAvailableYears(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	drivers, err := h.Repo.GetDriversByYear(r.Context(), year)
	if err != nil {
		if errors.Is(err, repository.ErrF1YearNotFound) {
			h.writeError(w, http.StatusNotFound, "Unsupported championship year.")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.F1DriversResponse{
		AvailableYears: availableYears,
		Data: model.F1DriversData{
			Year:    year,
			Drivers: drivers,
		},
	})
}

func (h *F1Handler) GetConstructors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	yearParam := r.URL.Query().Get("year")
	if yearParam == "" {
		h.writeError(w, http.StatusBadRequest, "Missing year query parameter.")
		return
	}

	year, _, statusCode, message := parseOptionalYear(yearParam)
	if statusCode != 0 {
		h.writeError(w, statusCode, message)
		return
	}

	availableYears, err := h.Repo.GetAvailableYears(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	constructors, err := h.Repo.GetConstructorsByYear(r.Context(), year)
	if err != nil {
		if errors.Is(err, repository.ErrF1YearNotFound) {
			h.writeError(w, http.StatusNotFound, "Unsupported championship year.")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.F1ConstructorsResponse{
		AvailableYears: availableYears,
		Data: model.F1ConstructorsData{
			Year:         year,
			Constructors: constructors,
		},
	})
}

func (h *F1Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	yearParam := r.URL.Query().Get("year")
	if yearParam == "" {
		h.writeError(w, http.StatusBadRequest, "Missing year query parameter.")
		return
	}

	year, _, statusCode, message := parseOptionalYear(yearParam)
	if statusCode != 0 {
		h.writeError(w, statusCode, message)
		return
	}

	availableYears, err := h.Repo.GetAvailableYears(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	events, err := h.Repo.GetEventsByYear(r.Context(), year)
	if err != nil {
		if errors.Is(err, repository.ErrF1YearNotFound) {
			h.writeError(w, http.StatusNotFound, "Unsupported championship year.")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.F1EventsResponse{
		AvailableYears: availableYears,
		Data: model.F1EventsData{
			Year:   year,
			Events: events,
		},
	})
}

func parseOptionalYear(yearParam string) (int, bool, int, string) {
	if yearParam == "" {
		return 0, false, 0, ""
	}

	year, err := strconv.Atoi(yearParam)
	if err != nil {
		return 0, false, http.StatusBadRequest, "Invalid year format."
	}

	return year, true, 0, ""
}

func parseDefaultYear(availableYears []int) (int, int, string) {
	if len(availableYears) == 0 {
		return 0, http.StatusNotFound, "Unsupported championship year."
	}

	return availableYears[0], 0, ""
}

func (h *F1Handler) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(model.APIErrorResponse{Message: message})
}
