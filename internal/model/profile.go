package model

import "time"

type Profile struct {
	ProfileID      string          `json:"profileId"`
	PhoneNumber    string          `json:"phoneNumber"`
	Email          string          `json:"email"`
	FirstName      string          `json:"firstName"`
	MiddleName     string          `json:"middleName"`
	LastName       string          `json:"lastName"`
	Address1       string          `json:"address1"`
	Address2       string          `json:"address2"`
	City           string          `json:"city"`
	State          string          `json:"state"`
	ZipCode        string          `json:"zipCode"`
	Country        string          `json:"country"`
	SocialAccounts []SocialAccount `json:"socialAccounts"`
	Created        time.Time       `json:"created"`
	Updated        time.Time       `json:"updated"`
}

type SocialAccount struct {
	Href  string `json:"href"`
	Label string `json:"label"`
}
