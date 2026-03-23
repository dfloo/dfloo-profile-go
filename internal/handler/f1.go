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

func (h *F1Handler) GetDriverDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	driverID, statusCode, message := parseRequiredNumericPathParam(r.PathValue("id"), "Missing driver id path parameter.", "Invalid driver id format.")
	if statusCode != 0 {
		h.writeError(w, statusCode, message)
		return
	}

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

	raceID, statusCode, message := parseOptionalNumericQueryParam(r.URL.Query().Get("raceId"), "Invalid raceId format.")
	if statusCode != 0 {
		h.writeError(w, statusCode, message)
		return
	}

	availableYears, err := h.Repo.GetAvailableYears(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	data, err := h.Repo.GetDriverDetails(r.Context(), driverID, year, raceID)
	if err != nil {
		if errors.Is(err, repository.ErrF1YearNotFound) {
			h.writeError(w, http.StatusNotFound, "Unsupported championship year.")
			return
		}
		if errors.Is(err, repository.ErrF1DriverNotFound) {
			h.writeError(w, http.StatusNotFound, "Driver not found for championship year.")
			return
		}
		if errors.Is(err, repository.ErrF1RaceNotFound) {
			h.writeError(w, http.StatusNotFound, "Race not found for championship year.")
			return
		}

		h.writeError(w, http.StatusInternalServerError, "Internal server error.")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.DriverDetailResponse{
		AvailableYears: availableYears,
		Data:           *data,
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

func parseOptionalNumericQueryParam(value string, invalidMessage string) (*int, int, string) {
	if value == "" {
		return nil, 0, ""
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, http.StatusBadRequest, invalidMessage
	}

	return &parsed, 0, ""
}

func parseRequiredNumericPathParam(value string, missingMessage string, invalidMessage string) (int, int, string) {
	if value == "" {
		return 0, http.StatusBadRequest, missingMessage
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, http.StatusBadRequest, invalidMessage
	}

	return parsed, 0, ""
}

func (h *F1Handler) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(model.APIErrorResponse{Message: message})
}
