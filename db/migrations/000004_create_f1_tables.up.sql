CREATE TABLE IF NOT EXISTS f1_circuits (
    circuit_id INTEGER PRIMARY KEY,
    circuit_ref TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    location TEXT NOT NULL,
    country TEXT NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    alt INTEGER,
    url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS f1_constructors (
    constructor_id INTEGER PRIMARY KEY,
    constructor_ref TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    nationality TEXT NOT NULL,
    url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS f1_drivers (
    driver_id INTEGER PRIMARY KEY,
    driver_ref TEXT NOT NULL UNIQUE,
    number INTEGER,
    code TEXT,
    forename TEXT NOT NULL,
    surname TEXT NOT NULL,
    dob DATE NOT NULL,
    nationality TEXT NOT NULL,
    url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS f1_seasons (
    year INTEGER PRIMARY KEY,
    url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS f1_status (
    status_id INTEGER PRIMARY KEY,
    status TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS f1_races (
    race_id INTEGER PRIMARY KEY,
    year INTEGER NOT NULL REFERENCES f1_seasons(year),
    round INTEGER NOT NULL,
    circuit_id INTEGER NOT NULL REFERENCES f1_circuits(circuit_id),
    name TEXT NOT NULL,
    race_date DATE NOT NULL,
    race_time TIME,
    url TEXT NOT NULL,
    fp1_date DATE,
    fp1_time TIME,
    fp2_date DATE,
    fp2_time TIME,
    fp3_date DATE,
    fp3_time TIME,
    quali_date DATE,
    quali_time TIME,
    sprint_date DATE,
    sprint_time TIME,
    UNIQUE (year, round)
);

CREATE TABLE IF NOT EXISTS f1_constructor_results (
    constructor_results_id INTEGER PRIMARY KEY,
    race_id INTEGER NOT NULL REFERENCES f1_races(race_id),
    constructor_id INTEGER NOT NULL REFERENCES f1_constructors(constructor_id),
    points NUMERIC(8, 2) NOT NULL,
    status TEXT
);

CREATE TABLE IF NOT EXISTS f1_constructor_standings (
    constructor_standings_id INTEGER PRIMARY KEY,
    race_id INTEGER NOT NULL REFERENCES f1_races(race_id),
    constructor_id INTEGER NOT NULL REFERENCES f1_constructors(constructor_id),
    points NUMERIC(8, 2) NOT NULL,
    position INTEGER,
    position_text TEXT,
    wins INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS f1_driver_standings (
    driver_standings_id INTEGER PRIMARY KEY,
    race_id INTEGER NOT NULL REFERENCES f1_races(race_id),
    driver_id INTEGER NOT NULL REFERENCES f1_drivers(driver_id),
    points NUMERIC(8, 2) NOT NULL,
    position INTEGER,
    position_text TEXT,
    wins INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS f1_results (
    result_id INTEGER PRIMARY KEY,
    race_id INTEGER NOT NULL REFERENCES f1_races(race_id),
    driver_id INTEGER NOT NULL REFERENCES f1_drivers(driver_id),
    constructor_id INTEGER NOT NULL REFERENCES f1_constructors(constructor_id),
    number INTEGER,
    grid INTEGER NOT NULL,
    position INTEGER,
    position_text TEXT NOT NULL,
    position_order INTEGER NOT NULL,
    points NUMERIC(8, 2) NOT NULL,
    laps INTEGER NOT NULL,
    result_time TEXT,
    milliseconds INTEGER,
    fastest_lap INTEGER,
    rank INTEGER,
    fastest_lap_time TEXT,
    fastest_lap_speed NUMERIC(8, 3),
    status_id INTEGER NOT NULL REFERENCES f1_status(status_id)
);

CREATE TABLE IF NOT EXISTS f1_sprint_results (
    result_id INTEGER PRIMARY KEY,
    race_id INTEGER NOT NULL REFERENCES f1_races(race_id),
    driver_id INTEGER NOT NULL REFERENCES f1_drivers(driver_id),
    constructor_id INTEGER NOT NULL REFERENCES f1_constructors(constructor_id),
    number INTEGER,
    grid INTEGER NOT NULL,
    position INTEGER,
    position_text TEXT NOT NULL,
    position_order INTEGER NOT NULL,
    points NUMERIC(8, 2) NOT NULL,
    laps INTEGER NOT NULL,
    result_time TEXT,
    milliseconds INTEGER,
    fastest_lap INTEGER,
    fastest_lap_time TEXT,
    status_id INTEGER NOT NULL REFERENCES f1_status(status_id)
);

CREATE TABLE IF NOT EXISTS f1_qualifying (
    qualify_id INTEGER PRIMARY KEY,
    race_id INTEGER NOT NULL REFERENCES f1_races(race_id),
    driver_id INTEGER NOT NULL REFERENCES f1_drivers(driver_id),
    constructor_id INTEGER NOT NULL REFERENCES f1_constructors(constructor_id),
    number INTEGER,
    position INTEGER NOT NULL,
    q1 TEXT,
    q2 TEXT,
    q3 TEXT
);

CREATE TABLE IF NOT EXISTS f1_pit_stops (
    race_id INTEGER NOT NULL REFERENCES f1_races(race_id),
    driver_id INTEGER NOT NULL REFERENCES f1_drivers(driver_id),
    stop INTEGER NOT NULL,
    lap INTEGER NOT NULL,
    stop_time TIME,
    duration TEXT,
    milliseconds INTEGER,
    PRIMARY KEY (race_id, driver_id, stop)
);

CREATE TABLE IF NOT EXISTS f1_lap_times (
    race_id INTEGER NOT NULL REFERENCES f1_races(race_id),
    driver_id INTEGER NOT NULL REFERENCES f1_drivers(driver_id),
    lap INTEGER NOT NULL,
    position INTEGER NOT NULL,
    lap_time TEXT NOT NULL,
    milliseconds INTEGER NOT NULL,
    PRIMARY KEY (race_id, driver_id, lap),
    UNIQUE (race_id, driver_id, lap, position)
);

CREATE INDEX IF NOT EXISTS idx_f1_races_circuit_id ON f1_races(circuit_id);
CREATE INDEX IF NOT EXISTS idx_f1_constructor_results_race_id ON f1_constructor_results(race_id);
CREATE INDEX IF NOT EXISTS idx_f1_constructor_results_constructor_id ON f1_constructor_results(constructor_id);
CREATE INDEX IF NOT EXISTS idx_f1_constructor_standings_race_id ON f1_constructor_standings(race_id);
CREATE INDEX IF NOT EXISTS idx_f1_constructor_standings_constructor_id ON f1_constructor_standings(constructor_id);
CREATE INDEX IF NOT EXISTS idx_f1_driver_standings_race_id ON f1_driver_standings(race_id);
CREATE INDEX IF NOT EXISTS idx_f1_driver_standings_driver_id ON f1_driver_standings(driver_id);
CREATE INDEX IF NOT EXISTS idx_f1_results_race_id ON f1_results(race_id);
CREATE INDEX IF NOT EXISTS idx_f1_results_driver_id ON f1_results(driver_id);
CREATE INDEX IF NOT EXISTS idx_f1_results_constructor_id ON f1_results(constructor_id);
CREATE INDEX IF NOT EXISTS idx_f1_results_status_id ON f1_results(status_id);
CREATE INDEX IF NOT EXISTS idx_f1_sprint_results_race_id ON f1_sprint_results(race_id);
CREATE INDEX IF NOT EXISTS idx_f1_sprint_results_driver_id ON f1_sprint_results(driver_id);
CREATE INDEX IF NOT EXISTS idx_f1_sprint_results_constructor_id ON f1_sprint_results(constructor_id);
CREATE INDEX IF NOT EXISTS idx_f1_sprint_results_status_id ON f1_sprint_results(status_id);
CREATE INDEX IF NOT EXISTS idx_f1_qualifying_race_id ON f1_qualifying(race_id);
CREATE INDEX IF NOT EXISTS idx_f1_qualifying_driver_id ON f1_qualifying(driver_id);
CREATE INDEX IF NOT EXISTS idx_f1_qualifying_constructor_id ON f1_qualifying(constructor_id);
CREATE INDEX IF NOT EXISTS idx_f1_pit_stops_driver_id ON f1_pit_stops(driver_id);
CREATE INDEX IF NOT EXISTS idx_f1_pit_stops_race_id ON f1_pit_stops(race_id);
CREATE INDEX IF NOT EXISTS idx_f1_lap_times_race_id ON f1_lap_times(race_id);
CREATE INDEX IF NOT EXISTS idx_f1_lap_times_driver_id ON f1_lap_times(driver_id);
