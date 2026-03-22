package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type tableSpec struct {
	Table   string
	File    string
	Columns []string
}

var tableSpecs = []tableSpec{
	{Table: "f1_seasons", File: "seasons.csv", Columns: []string{"year", "url"}},
	{Table: "f1_circuits", File: "circuits.csv", Columns: []string{"circuit_id", "circuit_ref", "name", "location", "country", "lat", "lng", "alt", "url"}},
	{Table: "f1_constructors", File: "constructors.csv", Columns: []string{"constructor_id", "constructor_ref", "name", "nationality", "url"}},
	{Table: "f1_drivers", File: "drivers.csv", Columns: []string{"driver_id", "driver_ref", "number", "code", "forename", "surname", "dob", "nationality", "url"}},
	{Table: "f1_status", File: "status.csv", Columns: []string{"status_id", "status"}},
	{Table: "f1_races", File: "races.csv", Columns: []string{"race_id", "year", "round", "circuit_id", "name", "race_date", "race_time", "url", "fp1_date", "fp1_time", "fp2_date", "fp2_time", "fp3_date", "fp3_time", "quali_date", "quali_time", "sprint_date", "sprint_time"}},
	{Table: "f1_constructor_results", File: "constructor_results.csv", Columns: []string{"constructor_results_id", "race_id", "constructor_id", "points", "status"}},
	{Table: "f1_constructor_standings", File: "constructor_standings.csv", Columns: []string{"constructor_standings_id", "race_id", "constructor_id", "points", "position", "position_text", "wins"}},
	{Table: "f1_driver_standings", File: "driver_standings.csv", Columns: []string{"driver_standings_id", "race_id", "driver_id", "points", "position", "position_text", "wins"}},
	{Table: "f1_results", File: "results.csv", Columns: []string{"result_id", "race_id", "driver_id", "constructor_id", "number", "grid", "position", "position_text", "position_order", "points", "laps", "result_time", "milliseconds", "fastest_lap", "rank", "fastest_lap_time", "fastest_lap_speed", "status_id"}},
	{Table: "f1_sprint_results", File: "sprint_results.csv", Columns: []string{"result_id", "race_id", "driver_id", "constructor_id", "number", "grid", "position", "position_text", "position_order", "points", "laps", "result_time", "milliseconds", "fastest_lap", "fastest_lap_time", "status_id"}},
	{Table: "f1_qualifying", File: "qualifying.csv", Columns: []string{"qualify_id", "race_id", "driver_id", "constructor_id", "number", "position", "q1", "q2", "q3"}},
	{Table: "f1_pit_stops", File: "pit_stops.csv", Columns: []string{"race_id", "driver_id", "stop", "lap", "stop_time", "duration", "milliseconds"}},
	{Table: "f1_lap_times", File: "lap_times.csv", Columns: []string{"race_id", "driver_id", "lap", "position", "lap_time", "milliseconds"}},
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	dataDirFlag := flag.String("data-dir", "db/f1-data", "directory containing source CSV files")
	flag.Parse()

	dataDir := strings.TrimSpace(*dataDirFlag)
	if dataDir == "" {
		log.Fatal("data-dir cannot be empty")
	}

	pool, err := pgxpool.New(ctx, buildDSN())
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	state, err := populationState(ctx, pool)
	if err != nil {
		log.Fatalf("check table population: %v", err)
	}

	switch state {
	case "full":
		log.Print("F1 tables already populated; skipping load")
		return
	case "partial":
		log.Fatal("partial F1 dataset detected; refusing to continue")
	}

	if err := loadAll(ctx, pool, dataDir); err != nil {
		log.Fatalf("load F1 data: %v", err)
	}

	log.Print("F1 data load complete")
}

func buildDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGPORT"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
		os.Getenv("PGDATABASE"),
	)
}

func populationState(ctx context.Context, pool *pgxpool.Pool) (string, error) {
	populated := 0

	for _, spec := range tableSpecs {
		hasRows, err := tableHasRows(ctx, pool, spec.Table)
		if err != nil {
			return "", err
		}
		if hasRows {
			populated++
		}
	}

	switch {
	case populated == len(tableSpecs):
		return "full", nil
	case populated == 0:
		return "empty", nil
	default:
		return "partial", nil
	}
}

func tableHasRows(ctx context.Context, pool *pgxpool.Pool, tableName string) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s LIMIT 1)", tableName)
	var hasRows bool
	if err := pool.QueryRow(ctx, query).Scan(&hasRows); err != nil {
		return false, fmt.Errorf("check rows in %s: %w", tableName, err)
	}
	return hasRows, nil
}

func loadAll(ctx context.Context, pool *pgxpool.Pool, dataDir string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	for _, spec := range tableSpecs {
		if err := loadTable(ctx, conn, dataDir, spec); err != nil {
			return err
		}
	}

	return nil
}

func loadTable(ctx context.Context, conn *pgxpool.Conn, dataDir string, spec tableSpec) error {
	csvPath := filepath.Join(dataDir, spec.File)
	file, err := os.Open(csvPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("missing csv file: %s", csvPath)
		}
		return fmt.Errorf("open csv file %s: %w", csvPath, err)
	}
	defer file.Close()

	copySQL := fmt.Sprintf(
		"COPY %s (%s) FROM STDIN WITH (FORMAT csv, HEADER true, NULL '\\N')",
		spec.Table,
		strings.Join(spec.Columns, ", "),
	)

	start := time.Now()
	tag, err := conn.Conn().PgConn().CopyFrom(ctx, file, copySQL)
	if err != nil {
		return fmt.Errorf("copy into %s from %s: %w", spec.Table, csvPath, err)
	}

	log.Printf("loaded %s from %s (%d rows in %s)", spec.Table, spec.File, tag.RowsAffected(), time.Since(start).Round(time.Millisecond))
	return nil
}
