package model

import "time"

type Resume struct {
	ResumeID         string           `json:"resumeId"`
	Profile          Profile          `json:"profile"`
	Sections         []string         `json:"sections"`
	Description      string           `json:"description"`
	Summary          string           `json:"summary"`
	Skills           []string         `json:"skills"`
	Experience       []Experience     `json:"experience"`
	Education        []Education      `json:"education"`
	FileName         string           `json:"fileName"`
	TemplateSettings TemplateSettings `json:"templateSettings"`
	Created          time.Time        `json:"created"`
	Updated          time.Time        `json:"updated"`
}

type Experience struct {
	Employer     string   `json:"employer"`
	Location     string   `json:"location"`
	Title        string   `json:"title"`
	StartDate    string   `json:"startDate"`
	EndDate      string   `json:"endDate"`
	Description  string   `json:"description"`
	BulletPoints []string `json:"bulletPoints"`
}

type Education struct {
	Name           string `json:"name"`
	Location       string `json:"location"`
	Type           string `json:"type"`
	CompletionDate string `json:"completionDate"`
}

type TemplateSettings struct {
	Name       string `json:"templateName"`
	FontFamily string `json:"fontFamily"`
}
