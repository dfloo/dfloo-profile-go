package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrF1YearNotFound = errors.New("f1 year not found")

const f1FallbackColor = "#6B7280"

var constructorColors = map[int]string{
	1:   "#00D2BE",
	3:   "#005AFF",
	4:   "#DC0000",
	6:   "#FF8700",
	9:   "#3671C6",
	51:  "#52E252",
	117: "#B6BABD",
	131: "#64C4FF",
	210: "#6692FF",
	213: "#229971",
	214: "#B12039",
}

type F1Repository interface {
	GetAvailableYears(ctx context.Context) ([]int, error)
	GetChampionshipByYear(ctx context.Context, year int) (*model.F1ChampionshipData, error)
	GetDriversByYear(ctx context.Context, year int) ([]model.F1DriverStanding, error)
	GetConstructorsByYear(ctx context.Context, year int) ([]model.F1ConstructorStanding, error)
	GetEventsByYear(ctx context.Context, year int) ([]model.F1Event, error)
}

type DBF1Repository struct {
	Pool *pgxpool.Pool
}

type championshipDriverAccumulator struct {
	id               int
	name             string
	color            string
	cumulativePoints []float64
	raceResults      []model.F1RaceResult
	standingsByRace  map[int]float64
}

func NewDBF1Repository(pool *pgxpool.Pool) *DBF1Repository {
	return &DBF1Repository{Pool: pool}
}

func (r *DBF1Repository) GetAvailableYears(ctx context.Context) ([]int, error) {
	if r.Pool == nil {
		return nil, errors.New("database pool is nil")
	}

	rows, err := r.Pool.Query(ctx, `SELECT year FROM f1_seasons ORDER BY year DESC`)
	if err != nil {
		return nil, fmt.Errorf("query available years: %w", err)
	}
	defer rows.Close()

	var years []int
	for rows.Next() {
		var year int
		if scanErr := rows.Scan(&year); scanErr != nil {
			return nil, fmt.Errorf("scan available year: %w", scanErr)
		}
		years = append(years, year)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate available years: %w", rows.Err())
	}

	return years, nil
}

func (r *DBF1Repository) GetChampionshipByYear(
	ctx context.Context,
	year int,
) (*model.F1ChampionshipData, error) {
	if r.Pool == nil {
		return nil, errors.New("database pool is nil")
	}

	events, raceIDs, raceIndexByID, err := r.getRacesByYear(ctx, year)
	if err != nil {
		return nil, err
	}

	driverAccumulators, err := r.getDriverStandingsByYear(ctx, year, raceIDs, raceIndexByID)
	if err != nil {
		return nil, err
	}

	if err = r.populateRaceResults(ctx, year, raceIndexByID, driverAccumulators); err != nil {
		return nil, err
	}

	drivers := finalizeChampionshipDrivers(driverAccumulators)

	return &model.F1ChampionshipData{
		Year:    year,
		Events:  events,
		Drivers: drivers,
	}, nil
}

func (r *DBF1Repository) GetDriversByYear(
	ctx context.Context,
	year int,
) ([]model.F1DriverStanding, error) {
	if r.Pool == nil {
		return nil, errors.New("database pool is nil")
	}

	rows, err := r.Pool.Query(
		ctx,
		`SELECT DISTINCT ON (ds.driver_id)
			ds.driver_id,
			d.forename,
			d.surname,
			ds.points
		 FROM f1_driver_standings ds
		 JOIN f1_races r ON r.race_id = ds.race_id
		 JOIN f1_drivers d ON d.driver_id = ds.driver_id
		 WHERE r.year = $1
		 ORDER BY ds.driver_id, r.round DESC`,
		year,
	)
	if err != nil {
		return nil, fmt.Errorf("query drivers by year: %w", err)
	}
	defer rows.Close()

	drivers := make([]model.F1DriverStanding, 0)
	for rows.Next() {
		var driverID int
		var firstName string
		var lastName string
		var points float64

		if scanErr := rows.Scan(&driverID, &firstName, &lastName, &points); scanErr != nil {
			return nil, fmt.Errorf("scan drivers by year: %w", scanErr)
		}

		drivers = append(drivers, model.F1DriverStanding{
			ID:           strconv.Itoa(driverID),
			Name:         strings.TrimSpace(firstName + " " + lastName),
			LatestPoints: points,
		})
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate drivers by year: %w", rows.Err())
	}

	if len(drivers) == 0 {
		return nil, ErrF1YearNotFound
	}

	sort.SliceStable(drivers, func(i, j int) bool {
		if drivers[i].LatestPoints == drivers[j].LatestPoints {
			return drivers[i].Name < drivers[j].Name
		}
		return drivers[i].LatestPoints > drivers[j].LatestPoints
	})

	return drivers, nil
}

func (r *DBF1Repository) GetConstructorsByYear(
	ctx context.Context,
	year int,
) ([]model.F1ConstructorStanding, error) {
	if r.Pool == nil {
		return nil, errors.New("database pool is nil")
	}

	rows, err := r.Pool.Query(
		ctx,
		`SELECT DISTINCT ON (cs.constructor_id)
			cs.constructor_id,
			c.name,
			cs.points
		 FROM f1_constructor_standings cs
		 JOIN f1_races r ON r.race_id = cs.race_id
		 JOIN f1_constructors c ON c.constructor_id = cs.constructor_id
		 WHERE r.year = $1
		 ORDER BY cs.constructor_id, r.round DESC`,
		year,
	)
	if err != nil {
		return nil, fmt.Errorf("query constructors by year: %w", err)
	}
	defer rows.Close()

	constructors := make([]model.F1ConstructorStanding, 0)
	for rows.Next() {
		var constructorID int
		var name string
		var points float64

		if scanErr := rows.Scan(&constructorID, &name, &points); scanErr != nil {
			return nil, fmt.Errorf("scan constructors by year: %w", scanErr)
		}

		constructors = append(constructors, model.F1ConstructorStanding{
			ID:           strconv.Itoa(constructorID),
			Name:         strings.TrimSpace(name),
			Color:        getConstructorColor(constructorID),
			LatestPoints: points,
		})
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate constructors by year: %w", rows.Err())
	}

	if len(constructors) == 0 {
		return nil, ErrF1YearNotFound
	}

	return finalizeConstructorStandings(constructors), nil
}

func (r *DBF1Repository) GetEventsByYear(ctx context.Context, year int) ([]model.F1Event, error) {
	if r.Pool == nil {
		return nil, errors.New("database pool is nil")
	}

	events, _, _, err := r.getRacesByYear(ctx, year)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *DBF1Repository) getRacesByYear(
	ctx context.Context,
	year int,
) ([]model.F1Event, []int, map[int]int, error) {
	rows, err := r.Pool.Query(
		ctx,
		`SELECT race_id, round, name
		 FROM f1_races
		 WHERE year = $1
		 ORDER BY round ASC`,
		year,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("query races by year: %w", err)
	}
	defer rows.Close()

	events := make([]model.F1Event, 0)
	raceIDs := make([]int, 0)
	raceIndexByID := make(map[int]int)

	for rows.Next() {
		var raceID int
		var event model.F1Event
		if scanErr := rows.Scan(&raceID, &event.Round, &event.Name); scanErr != nil {
			return nil, nil, nil, fmt.Errorf("scan race by year: %w", scanErr)
		}
		raceIndexByID[raceID] = len(events)
		raceIDs = append(raceIDs, raceID)
		event.RaceID = strconv.Itoa(raceID)
		events = append(events, event)
	}

	if rows.Err() != nil {
		return nil, nil, nil, fmt.Errorf("iterate races by year: %w", rows.Err())
	}

	if len(events) == 0 {
		return nil, nil, nil, ErrF1YearNotFound
	}

	return events, raceIDs, raceIndexByID, nil
}

func (r *DBF1Repository) getDriverStandingsByYear(
	ctx context.Context,
	year int,
	raceIDs []int,
	raceIndexByID map[int]int,
) (map[int]*championshipDriverAccumulator, error) {
	rows, err := r.Pool.Query(
		ctx,
		`SELECT
			ds.driver_id,
			d.forename,
			d.surname,
			r.race_id,
			ds.points
		 FROM f1_driver_standings ds
		 JOIN f1_races r ON r.race_id = ds.race_id
		 JOIN f1_drivers d ON d.driver_id = ds.driver_id
		 WHERE r.year = $1
		 ORDER BY ds.driver_id, r.round ASC`,
		year,
	)
	if err != nil {
		return nil, fmt.Errorf("query driver standings by year: %w", err)
	}
	defer rows.Close()

	driverAccumulators := make(map[int]*championshipDriverAccumulator)

	for rows.Next() {
		var driverID int
		var firstName string
		var lastName string
		var raceID int
		var cumulativePoints float64

		if scanErr := rows.Scan(&driverID, &firstName, &lastName, &raceID, &cumulativePoints); scanErr != nil {
			return nil, fmt.Errorf("scan driver standings by year: %w", scanErr)
		}

		accumulator, ok := driverAccumulators[driverID]
		if !ok {
			accumulator = initializeChampionshipDriverAccumulator(driverID, firstName, lastName, len(raceIDs))
			driverAccumulators[driverID] = accumulator
		}

		raceIndex, found := raceIndexByID[raceID]
		if !found {
			continue
		}

		accumulator.standingsByRace[raceIndex] = cumulativePoints
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate driver standings by year: %w", rows.Err())
	}

	for _, accumulator := range driverAccumulators {
		running := 0.0
		for idx := range accumulator.cumulativePoints {
			if points, ok := accumulator.standingsByRace[idx]; ok {
				running = points
			}
			accumulator.cumulativePoints[idx] = running
		}
	}

	return driverAccumulators, nil
}

func (r *DBF1Repository) populateRaceResults(
	ctx context.Context,
	year int,
	raceIndexByID map[int]int,
	driverAccumulators map[int]*championshipDriverAccumulator,
) error {
	rows, err := r.Pool.Query(
		ctx,
		`SELECT
			fr.driver_id,
			r.race_id,
			fr.points,
			fc.constructor_id,
			fc.name,
			d.forename,
			d.surname
		 FROM f1_results fr
		 JOIN f1_races r ON r.race_id = fr.race_id
		 JOIN f1_drivers d ON d.driver_id = fr.driver_id
		 LEFT JOIN f1_constructors fc ON fc.constructor_id = fr.constructor_id
		 WHERE r.year = $1
		 ORDER BY fr.driver_id, r.round ASC`,
		year,
	)
	if err != nil {
		return fmt.Errorf("query race results by year: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var driverID int
		var raceID int
		var racePoints float64
		var constructorID *int
		var constructorName *string
		var firstName string
		var lastName string

		if scanErr := rows.Scan(
			&driverID,
			&raceID,
			&racePoints,
			&constructorID,
			&constructorName,
			&firstName,
			&lastName,
		); scanErr != nil {
			return fmt.Errorf("scan race results by year: %w", scanErr)
		}

		accumulator, ok := driverAccumulators[driverID]
		if !ok {
			accumulator = initializeChampionshipDriverAccumulator(
				driverID,
				firstName,
				lastName,
				len(raceIndexByID),
			)
			driverAccumulators[driverID] = accumulator
		}

		raceIndex, found := raceIndexByID[raceID]
		if !found {
			continue
		}

		constructorIDStr := (*string)(nil)
		constructorColor := f1FallbackColor
		if constructorID != nil {
			value := strconv.Itoa(*constructorID)
			constructorIDStr = &value
			constructorColor = getConstructorColor(*constructorID)
			if accumulator.color == f1FallbackColor {
				accumulator.color = constructorColor
			}
		}

		accumulator.raceResults[raceIndex] = model.F1RaceResult{
			ConstructorID:    constructorIDStr,
			ConstructorName:  constructorName,
			ConstructorColor: constructorColor,
			RacePoints:       racePoints,
		}
	}

	if rows.Err() != nil {
		return fmt.Errorf("iterate race results by year: %w", rows.Err())
	}

	return nil
}

func initializeChampionshipDriverAccumulator(
	driverID int,
	firstName string,
	lastName string,
	raceCount int,
) *championshipDriverAccumulator {
	raceResults := make([]model.F1RaceResult, raceCount)
	for idx := range raceResults {
		raceResults[idx] = model.F1RaceResult{
			ConstructorID:    nil,
			ConstructorName:  nil,
			ConstructorColor: f1FallbackColor,
			RacePoints:       0,
		}
	}

	return &championshipDriverAccumulator{
		id:               driverID,
		name:             strings.TrimSpace(firstName + " " + lastName),
		color:            f1FallbackColor,
		cumulativePoints: make([]float64, raceCount),
		raceResults:      raceResults,
		standingsByRace:  make(map[int]float64),
	}
}

func finalizeChampionshipDrivers(
	driverAccumulators map[int]*championshipDriverAccumulator,
) []model.F1ChampionshipDriver {
	drivers := make([]model.F1ChampionshipDriver, 0, len(driverAccumulators))
	for _, accumulator := range driverAccumulators {
		drivers = append(drivers, model.F1ChampionshipDriver{
			ID:               strconv.Itoa(accumulator.id),
			Name:             accumulator.name,
			Color:            accumulator.color,
			CumulativePoints: accumulator.cumulativePoints,
			RaceResults:      accumulator.raceResults,
		})
	}

	sort.SliceStable(drivers, func(i, j int) bool {
		iFinal := 0.0
		jFinal := 0.0
		if len(drivers[i].CumulativePoints) > 0 {
			iFinal = drivers[i].CumulativePoints[len(drivers[i].CumulativePoints)-1]
		}
		if len(drivers[j].CumulativePoints) > 0 {
			jFinal = drivers[j].CumulativePoints[len(drivers[j].CumulativePoints)-1]
		}

		if iFinal == jFinal {
			return drivers[i].Name < drivers[j].Name
		}
		return iFinal > jFinal
	})

	return drivers
}

func getConstructorColor(constructorID int) string {
	if color, ok := constructorColors[constructorID]; ok {
		return color
	}
	return f1FallbackColor
}

func finalizeConstructorStandings(
	constructors []model.F1ConstructorStanding,
) []model.F1ConstructorStanding {
	sort.SliceStable(constructors, func(i, j int) bool {
		if constructors[i].LatestPoints == constructors[j].LatestPoints {
			return constructors[i].Name < constructors[j].Name
		}
		return constructors[i].LatestPoints > constructors[j].LatestPoints
	})

	return constructors
}
