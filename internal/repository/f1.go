package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/dfloo/dfloo-profile-go/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrF1YearNotFound = errors.New("f1 year not found")
var ErrF1DriverNotFound = errors.New("f1 driver not found")
var ErrF1RaceNotFound = errors.New("f1 race not found")

const f1FallbackColor = "#6B7280"

var constructorColors = map[int]string{
	1:   "#FF8700",
	3:   "#005AFF",
	4:   "#FFEF00",
	6:   "#FF2800",
	9:   "#3671C6",
	15:  "#01C00E",
	51:  "#52E252",
	117: "#229971",
	131: "#64C4FF",
	210: "#F62039",
	213: "#0078C1",
	214: "#B12039",
	215: "#6C98FF",
}

type F1Repository interface {
	GetAvailableYears(ctx context.Context) ([]int, error)
	GetChampionshipByYear(ctx context.Context, year int) (*model.F1ChampionshipData, error)
	GetDriversByYear(ctx context.Context, year int) ([]model.F1DriverStanding, error)
	GetConstructorsByYear(ctx context.Context, year int) ([]model.F1ConstructorStanding, error)
	GetEventsByYear(ctx context.Context, year int) ([]model.F1Event, error)
	GetDriverDetails(ctx context.Context, driverID int, year int, raceID *int) (*model.DriverDetailData, error)
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

type driverLapStatsPoint struct {
	Lap      int
	Time     *int
	MinTime  *int
	MaxTime  *int
	AvgTime  *float64
	Position *int
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

func (r *DBF1Repository) GetDriverDetails(
	ctx context.Context,
	driverID int,
	year int,
	raceID *int,
) (*model.DriverDetailData, error) {
	if r.Pool == nil {
		return nil, errors.New("database pool is nil")
	}

	events, raceIDs, raceIndexByID, err := r.getRacesByYear(ctx, year)
	if err != nil {
		return nil, err
	}

	races := make([]model.DriverRaceOption, 0, len(events))
	for _, event := range events {
		races = append(races, model.DriverRaceOption{
			ID:    event.RaceID,
			Round: event.Round,
			Name:  event.Name,
		})
	}

	selectedRaceID := raceIDs[len(raceIDs)-1]
	if raceID != nil {
		if _, ok := raceIndexByID[*raceID]; !ok {
			return nil, ErrF1RaceNotFound
		}
		selectedRaceID = *raceID
	}

	seasonPoints, seasonPointsByRaceID, err := r.getDriverSeasonPointsByYear(ctx, year, driverID)
	if err != nil {
		return nil, err
	}

	selectedSeasonPoint, ok := seasonPointsByRaceID[selectedRaceID]
	if !ok {
		return nil, ErrF1DriverNotFound
	}

	driverHeader, err := r.getDriverDetailHeaderByRace(ctx, driverID, selectedRaceID, selectedSeasonPoint.CumulativePoints)
	if err != nil {
		return nil, err
	}

	selectedRaceContext := &model.DriverSelectedRaceContext{
		ID:               selectedSeasonPoint.RaceID,
		Round:            selectedSeasonPoint.Round,
		Name:             selectedSeasonPoint.Name,
		RacePoints:       selectedSeasonPoint.RacePoints,
		CumulativePoints: selectedSeasonPoint.CumulativePoints,
		LapTimes:         []model.DriverLapTimePoint{},
	}

	startPosition, endPosition, err := r.getDriverRacePositions(ctx, driverID, selectedRaceID)
	if err != nil {
		return nil, err
	}
	selectedRaceContext.StartingPosition = startPosition
	selectedRaceContext.EndingPosition = endPosition

	qualifying, err := r.getDriverQualifyingByRace(ctx, driverID, selectedRaceID)
	if err != nil {
		return nil, err
	}
	selectedRaceContext.Qualifying = qualifying

	lapStats, err := r.getDriverLapStatsByRace(ctx, driverID, selectedRaceID)
	if err != nil {
		return nil, err
	}
	selectedRaceContext.LapTimes = buildDriverLapTimes(lapStats)

	teammatePositionsByLap, err := r.getTeammatePositionsByLap(ctx, driverID, selectedRaceID)
	if err != nil {
		return nil, err
	}
	selectedRaceContext.ChartData = buildDriverSelectedRaceChartData(lapStats, teammatePositionsByLap)

	raceScore, err := r.calculateRaceScore(ctx, selectedRaceID, driverID, endPosition)
	if err != nil {
		return nil, err
	}
	selectedRaceContext.RaceScore = raceScore

	leaderCumulativePoints, err := r.getLeaderCumulativePointsByRace(ctx, selectedRaceID)
	if err != nil {
		return nil, err
	}
	selectedRaceContext.SeasonScore = calculateSeasonScore(
		selectedSeasonPoint.CumulativePoints,
		leaderCumulativePoints,
	)

	return &model.DriverDetailData{
		Year:         year,
		Driver:       *driverHeader,
		Races:        races,
		SelectedRace: selectedRaceContext,
		SeasonPoints: seasonPoints,
	}, nil
}

func (r *DBF1Repository) getDriverSeasonPointsByYear(
	ctx context.Context,
	year int,
	driverID int,
) ([]model.DriverSeasonPointsPoint, map[int]model.DriverSeasonPointsPoint, error) {
	rows, err := r.Pool.Query(
		ctx,
		`SELECT
			r.race_id,
			r.round,
			r.name,
			COALESCE(fr.points, 0),
			ds.points
		 FROM f1_driver_standings ds
		 JOIN f1_races r ON r.race_id = ds.race_id
		 LEFT JOIN f1_results fr ON fr.race_id = ds.race_id AND fr.driver_id = ds.driver_id
		 WHERE r.year = $1 AND ds.driver_id = $2
		 ORDER BY r.round ASC`,
		year,
		driverID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("query driver season points by year: %w", err)
	}
	defer rows.Close()

	seasonPoints := make([]model.DriverSeasonPointsPoint, 0)
	seasonPointsByRaceID := make(map[int]model.DriverSeasonPointsPoint)

	for rows.Next() {
		var raceID int
		var point model.DriverSeasonPointsPoint
		if scanErr := rows.Scan(&raceID, &point.Round, &point.Name, &point.RacePoints, &point.CumulativePoints); scanErr != nil {
			return nil, nil, fmt.Errorf("scan driver season points by year: %w", scanErr)
		}
		point.RaceID = strconv.Itoa(raceID)
		seasonPoints = append(seasonPoints, point)
		seasonPointsByRaceID[raceID] = point
	}

	if rows.Err() != nil {
		return nil, nil, fmt.Errorf("iterate driver season points by year: %w", rows.Err())
	}

	if len(seasonPoints) == 0 {
		return nil, nil, ErrF1DriverNotFound
	}

	return seasonPoints, seasonPointsByRaceID, nil
}

func (r *DBF1Repository) getDriverDetailHeaderByRace(
	ctx context.Context,
	driverID int,
	raceID int,
	currentPoints float64,
) (*model.DriverDetailHeader, error) {
	var firstName string
	var lastName string
	var number *int
	var constructorName *string

	err := r.Pool.QueryRow(
		ctx,
		`SELECT
			d.forename,
			d.surname,
			d.number,
			c.name
		 FROM f1_drivers d
		 LEFT JOIN f1_results fr ON fr.driver_id = d.driver_id AND fr.race_id = $2
		 LEFT JOIN f1_constructors c ON c.constructor_id = fr.constructor_id
		 WHERE d.driver_id = $1`,
		driverID,
		raceID,
	).Scan(&firstName, &lastName, &number, &constructorName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrF1DriverNotFound
		}
		return nil, fmt.Errorf("query driver detail header by race: %w", err)
	}

	var numberString *string
	if number != nil {
		value := strconv.Itoa(*number)
		numberString = &value
	}

	return &model.DriverDetailHeader{
		ID:              strconv.Itoa(driverID),
		Name:            strings.TrimSpace(firstName + " " + lastName),
		Number:          numberString,
		ConstructorName: constructorName,
		CurrentPoints:   currentPoints,
	}, nil
}

func (r *DBF1Repository) getDriverQualifyingByRace(
	ctx context.Context,
	driverID int,
	raceID int,
) (*model.DriverQualifyingBreakdown, error) {
	var q1 *string
	var q2 *string
	var q3 *string

	err := r.Pool.QueryRow(
		ctx,
		`SELECT q1, q2, q3
		 FROM f1_qualifying
		 WHERE race_id = $1 AND driver_id = $2`,
		raceID,
		driverID,
	).Scan(&q1, &q2, &q3)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query driver qualifying by race: %w", err)
	}

	return &model.DriverQualifyingBreakdown{
		Q1: q1,
		Q2: q2,
		Q3: q3,
	}, nil
}

func (r *DBF1Repository) getDriverLapStatsByRace(
	ctx context.Context,
	driverID int,
	raceID int,
) ([]driverLapStatsPoint, error) {
	rows, err := r.Pool.Query(
		ctx,
		`SELECT
			lt.lap,
			lt.milliseconds,
			lt.position,
			stats.min_ms,
			stats.max_ms,
			stats.avg_ms
		 FROM f1_lap_times lt
		 JOIN (
			SELECT
				lap,
				MIN(milliseconds) AS min_ms,
				MAX(milliseconds) AS max_ms,
				AVG(milliseconds)::float8 AS avg_ms
			FROM f1_lap_times
			WHERE race_id = $1
			GROUP BY lap
		 ) stats ON stats.lap = lt.lap
		 WHERE lt.race_id = $1 AND lt.driver_id = $2
		 ORDER BY lt.lap ASC`,
		raceID,
		driverID,
	)
	if err != nil {
		return nil, fmt.Errorf("query driver lap stats by race: %w", err)
	}
	defer rows.Close()

	lapStats := make([]driverLapStatsPoint, 0)
	for rows.Next() {
		var point driverLapStatsPoint
		var milliseconds int
		var position int
		var minMilliseconds int
		var maxMilliseconds int
		var avgMilliseconds float64
		if scanErr := rows.Scan(
			&point.Lap,
			&milliseconds,
			&position,
			&minMilliseconds,
			&maxMilliseconds,
			&avgMilliseconds,
		); scanErr != nil {
			return nil, fmt.Errorf("scan driver lap stats by race: %w", scanErr)
		}
		point.Position = &position
		point.Time = &milliseconds
		point.MinTime = &minMilliseconds
		point.MaxTime = &maxMilliseconds
		point.AvgTime = &avgMilliseconds
		lapStats = append(lapStats, point)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate driver lap stats by race: %w", rows.Err())
	}

	return lapStats, nil
}

func buildDriverLapTimes(lapStats []driverLapStatsPoint) []model.DriverLapTimePoint {
	lapTimes := make([]model.DriverLapTimePoint, 0, len(lapStats))
	for _, point := range lapStats {
		lapTimes = append(lapTimes, model.DriverLapTimePoint{
			Lap:     point.Lap,
			Time:    point.Time,
			MinTime: point.MinTime,
			MaxTime: point.MaxTime,
			AvgTime: point.AvgTime,
		})
	}

	return lapTimes
}

func buildDriverSelectedRaceChartData(
	lapStats []driverLapStatsPoint,
	teammatePositionsByLap map[int]*int,
) *model.DriverSelectedRaceChartData {
	if len(lapStats) == 0 {
		return &model.DriverSelectedRaceChartData{}
	}

	sortedLapStats := append([]driverLapStatsPoint(nil), lapStats...)
	sort.SliceStable(sortedLapStats, func(i, j int) bool {
		return sortedLapStats[i].Lap < sortedLapStats[j].Lap
	})

	paceDelta := make([]model.DriverMetricByLapPoint, 0, len(sortedLapStats))
	gapToFastest := make([]model.DriverMetricByLapPoint, 0, len(sortedLapStats))
	rollingPace := make([]model.DriverMetricByLapPoint, 0, len(sortedLapStats))
	driverPositionsByLap := make(map[int]*int, len(sortedLapStats))

	for idx, lapStat := range sortedLapStats {
		driverPositionsByLap[lapStat.Lap] = lapStat.Position

		var paceDeltaValue *int
		if lapStat.Time != nil && lapStat.AvgTime != nil {
			delta := int(math.Round(float64(*lapStat.Time) - *lapStat.AvgTime))
			paceDeltaValue = &delta
		}
		paceDelta = append(paceDelta, model.DriverMetricByLapPoint{Lap: lapStat.Lap, Value: paceDeltaValue})

		var gapToFastestValue *int
		if lapStat.Time != nil && lapStat.MinTime != nil {
			gap := *lapStat.Time - *lapStat.MinTime
			gapToFastestValue = &gap
		}
		gapToFastest = append(gapToFastest, model.DriverMetricByLapPoint{Lap: lapStat.Lap, Value: gapToFastestValue})

		var rollingValue *int
		if idx >= 2 {
			first := sortedLapStats[idx-2].Time
			second := sortedLapStats[idx-1].Time
			third := sortedLapStats[idx].Time
			if first != nil && second != nil && third != nil {
				avg := int(math.Round(float64(*first+*second+*third) / 3.0))
				rollingValue = &avg
			}
		}
		rollingPace = append(rollingPace, model.DriverMetricByLapPoint{Lap: lapStat.Lap, Value: rollingValue})
	}

	positionsByLap := buildDriverPositionsByLapSeries(driverPositionsByLap, teammatePositionsByLap)

	return &model.DriverSelectedRaceChartData{
		PaceDeltaVsAverageMs: paceDelta,
		GapToFastestMs:       gapToFastest,
		Rolling3LapPaceMs:    rollingPace,
		PositionsByLap:       positionsByLap,
	}
}

func buildDriverPositionsByLapSeries(
	driverPositionsByLap map[int]*int,
	teammatePositionsByLap map[int]*int,
) []model.DriverPositionsByLapPoint {
	lapsSet := make(map[int]struct{})
	for lap := range driverPositionsByLap {
		lapsSet[lap] = struct{}{}
	}
	for lap := range teammatePositionsByLap {
		lapsSet[lap] = struct{}{}
	}

	if len(lapsSet) == 0 {
		return nil
	}

	laps := make([]int, 0, len(lapsSet))
	for lap := range lapsSet {
		laps = append(laps, lap)
	}
	sort.Ints(laps)

	positions := make([]model.DriverPositionsByLapPoint, 0, len(laps))
	for _, lap := range laps {
		positions = append(positions, model.DriverPositionsByLapPoint{
			Lap:              lap,
			DriverPosition:   driverPositionsByLap[lap],
			TeammatePosition: teammatePositionsByLap[lap],
		})
	}

	return positions
}

func (r *DBF1Repository) getTeammatePositionsByLap(
	ctx context.Context,
	driverID int,
	raceID int,
) (map[int]*int, error) {
	var teammateDriverID int
	err := r.Pool.QueryRow(
		ctx,
		`SELECT teammate.driver_id
		 FROM f1_results own
		 JOIN f1_results teammate
		   ON teammate.race_id = own.race_id
		  AND teammate.constructor_id = own.constructor_id
		  AND teammate.driver_id <> own.driver_id
		 WHERE own.race_id = $1 AND own.driver_id = $2
		 ORDER BY teammate.driver_id ASC
		 LIMIT 1`,
		raceID,
		driverID,
	).Scan(&teammateDriverID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query teammate by race: %w", err)
	}

	rows, err := r.Pool.Query(
		ctx,
		`SELECT lap, position
		 FROM f1_lap_times
		 WHERE race_id = $1 AND driver_id = $2
		 ORDER BY lap ASC`,
		raceID,
		teammateDriverID,
	)
	if err != nil {
		return nil, fmt.Errorf("query teammate lap positions by race: %w", err)
	}
	defer rows.Close()

	teammatePositions := make(map[int]*int)
	for rows.Next() {
		var lap int
		var position int
		if scanErr := rows.Scan(&lap, &position); scanErr != nil {
			return nil, fmt.Errorf("scan teammate lap positions by race: %w", scanErr)
		}
		pos := position
		teammatePositions[lap] = &pos
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate teammate lap positions by race: %w", rows.Err())
	}

	return teammatePositions, nil
}

func (r *DBF1Repository) getDriverRacePositions(
	ctx context.Context,
	driverID int,
	raceID int,
) (*int, *int, error) {
	var grid int
	var finishPosition *int

	err := r.Pool.QueryRow(
		ctx,
		`SELECT grid, position
		 FROM f1_results
		 WHERE race_id = $1 AND driver_id = $2`,
		raceID,
		driverID,
	).Scan(&grid, &finishPosition)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrF1DriverNotFound
		}
		return nil, nil, fmt.Errorf("query driver race positions: %w", err)
	}

	startPosition := grid
	return &startPosition, finishPosition, nil
}

func (r *DBF1Repository) calculateRaceScore(
	ctx context.Context,
	raceID int,
	driverID int,
	finishPosition *int,
) (float64, error) {
	finishComponent := 0.0
	if finishPosition != nil {
		maxPosition, err := r.getMaxFinishingPositionByRace(ctx, raceID)
		if err != nil {
			return 0, err
		}
		if maxPosition > 0 {
			finishComponent = (float64(maxPosition-*finishPosition+1) / float64(maxPosition)) * 100
		}
	}

	paceComponent, err := r.getDriverPaceScoreByRace(ctx, raceID, driverID)
	if err != nil {
		return 0, err
	}

	return clampScore((finishComponent * 0.7) + (paceComponent * 0.3)), nil
}

func (r *DBF1Repository) getMaxFinishingPositionByRace(ctx context.Context, raceID int) (int, error) {
	var maxPosition int
	err := r.Pool.QueryRow(
		ctx,
		`SELECT COALESCE(MAX(position), 0)
		 FROM f1_results
		 WHERE race_id = $1 AND position IS NOT NULL`,
		raceID,
	).Scan(&maxPosition)
	if err != nil {
		return 0, fmt.Errorf("query max finishing position by race: %w", err)
	}

	return maxPosition, nil
}

func (r *DBF1Repository) getDriverPaceScoreByRace(ctx context.Context, raceID int, driverID int) (float64, error) {
	var driverAverage float64
	var fieldAverage float64

	err := r.Pool.QueryRow(
		ctx,
		`WITH driver_laps AS (
			SELECT lap, milliseconds
			FROM f1_lap_times
			WHERE race_id = $1 AND driver_id = $2
		), field_lap_averages AS (
			SELECT lt.lap, AVG(lt.milliseconds)::float8 AS avg_ms
			FROM f1_lap_times lt
			JOIN driver_laps dl ON dl.lap = lt.lap
			WHERE lt.race_id = $1
			GROUP BY lt.lap
		)
		SELECT
			COALESCE((SELECT AVG(milliseconds)::float8 FROM driver_laps), 0),
			COALESCE((SELECT AVG(avg_ms)::float8 FROM field_lap_averages), 0)`,
		raceID,
		driverID,
	).Scan(&driverAverage, &fieldAverage)
	if err != nil {
		return 0, fmt.Errorf("query driver pace score by race: %w", err)
	}

	if driverAverage <= 0 || fieldAverage <= 0 {
		return 0, nil
	}

	return clampScore((fieldAverage / driverAverage) * 100), nil
}

func (r *DBF1Repository) getLeaderCumulativePointsByRace(ctx context.Context, raceID int) (float64, error) {
	var leaderPoints float64
	err := r.Pool.QueryRow(
		ctx,
		`SELECT COALESCE(MAX(points), 0)
		 FROM f1_driver_standings
		 WHERE race_id = $1`,
		raceID,
	).Scan(&leaderPoints)
	if err != nil {
		return 0, fmt.Errorf("query leader cumulative points by race: %w", err)
	}

	return leaderPoints, nil
}

func calculateSeasonScore(driverPoints float64, leaderPoints float64) float64 {
	if driverPoints <= 0 || leaderPoints <= 0 {
		return 0
	}

	return clampScore((driverPoints / leaderPoints) * 100)
}

func clampScore(score float64) float64 {
	clamped := math.Max(0, math.Min(100, score))
	return math.Round(clamped*100) / 100
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
