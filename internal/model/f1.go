package model

type F1ChampionshipsResponse struct {
	AvailableYears []int              `json:"availableYears"`
	Data           F1ChampionshipData `json:"data"`
}

type F1ChampionshipData struct {
	Year    int                    `json:"year"`
	Races   []F1Race               `json:"races"`
	Drivers []F1ChampionshipDriver `json:"drivers"`
}

type F1Race struct {
	Round int    `json:"round"`
	Name  string `json:"name"`
}

type F1ChampionshipDriver struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Color            string         `json:"color"`
	CumulativePoints []float64      `json:"cumulativePoints"`
	RaceResults      []F1RaceResult `json:"raceResults"`
}

type F1RaceResult struct {
	ConstructorID    *string `json:"constructorId"`
	ConstructorName  *string `json:"constructorName"`
	ConstructorColor string  `json:"constructorColor"`
	RacePoints       float64 `json:"racePoints"`
}

type F1DriversResponse struct {
	AvailableYears []int         `json:"availableYears"`
	Data           F1DriversData `json:"data"`
}

type F1DriversData struct {
	Year    int                `json:"year"`
	Drivers []F1DriverStanding `json:"drivers"`
}

type F1DriverStanding struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	LatestPoints float64 `json:"latestPoints"`
}

type F1ConstructorsResponse struct {
	AvailableYears []int              `json:"availableYears"`
	Data           F1ConstructorsData `json:"data"`
}

type F1ConstructorsData struct {
	Year         int                     `json:"year"`
	Constructors []F1ConstructorStanding `json:"constructors"`
}

type F1ConstructorStanding struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Color        string  `json:"color"`
	LatestPoints float64 `json:"latestPoints"`
}

type APIErrorResponse struct {
	Message string `json:"message"`
}
