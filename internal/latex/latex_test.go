package latex

import (
	"testing"
	"time"

	"github.com/dfloo/dfloo-profile-go/internal/model"
)

func TestEscapeSpecialChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "backslash", input: "\\", want: "\\textbackslash\\{\\}"},
		{name: "ampersand", input: "&", want: "\\&"},
		{name: "percent", input: "%", want: "\\%"},
		{name: "dollar", input: "$", want: "\\$"},
		{name: "hash", input: "#", want: "\\#"},
		{name: "underscore", input: "_", want: "\\_"},
		{name: "left brace", input: "{", want: "\\{"},
		{name: "right brace", input: "}", want: "\\}"},
		{name: "tilde", input: "~", want: "\\textasciitilde{}"},
		{name: "caret", input: "^", want: "\\textasciicircum{}"},
		{name: "registered", input: "®", want: "\\textregistered{}"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EscapeSpecialChars(tc.input)
			if got != tc.want {
				t.Errorf("EscapeSpecialChars(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestFormatFullName(t *testing.T) {
	tests := []struct {
		name    string
		profile model.Profile
		want    string
	}{
		{
			name: "all parts",
			profile: model.Profile{
				FirstName:  "Ada",
				MiddleName: "Lovelace",
				LastName:   "Byron",
			},
			want: "Ada Lovelace Byron",
		},
		{
			name: "first and last only",
			profile: model.Profile{
				FirstName: "Ada",
				LastName:  "Byron",
			},
			want: "Ada Byron",
		},
		{
			name:    "empty",
			profile: model.Profile{},
			want:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatFullName(&tc.profile)
			if got != tc.want {
				t.Errorf("FormatFullName() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGetStateShortCode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "valid", input: "US-CA", want: "CA"},
		{name: "invalid no dash", input: "CA", want: ""},
		{name: "empty", input: "", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetStateShortCode(tc.input)
			if got != tc.want {
				t.Errorf("GetStateShortCode(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestFormatAddress(t *testing.T) {
	profile := &model.Profile{
		Address1: "123 Main St #5",
		Address2: "Apt_2",
		City:     "San Francisco",
		State:    "US-CA",
		ZipCode:  "95110",
		Country:  "U$A",
	}

	got := FormatAddress(profile)
	want := "123 Main St \\#5, Apt\\_2, San Francisco, CA, 95110, U\\$A"
	if got != want {
		t.Errorf("FormatAddress() = %q, want %q", got, want)
	}
}

func TestFormatExperienceEducationAndSkills(t *testing.T) {
	experience := []model.Experience{
		{
			Title:        "Engineer & Architect",
			Employer:     "ACME_Inc",
			Location:     "San Jose #1",
			Description:  "Built 100% of $core",
			StartDate:    "2020-01",
			EndDate:      "2022-01",
			BulletPoints: []string{"Reduced cost by 20%", "C# + Go"},
		},
	}
	education := []model.Education{
		{
			Type:           "B.Sc",
			Name:           "Uni & College",
			Location:       "NY_1",
			CompletionDate: "2020",
		},
	}
	skills := []string{"Go", "C#", "Infra_1"}

	formattedExperience := FormatExperience(experience)
	if formattedExperience[0].Title != "Engineer \\& Architect" {
		t.Errorf("unexpected escaped experience title: %q", formattedExperience[0].Title)
	}
	if formattedExperience[0].Employer != "ACME\\_Inc" {
		t.Errorf("unexpected escaped employer: %q", formattedExperience[0].Employer)
	}
	if formattedExperience[0].Description != "Built 100\\% of \\$core" {
		t.Errorf("unexpected escaped description: %q", formattedExperience[0].Description)
	}
	if formattedExperience[0].BulletPoints[1] != "C\\# + Go" {
		t.Errorf("unexpected escaped bullet point: %q", formattedExperience[0].BulletPoints[1])
	}

	formattedEducation := FormatEducation(education)
	if formattedEducation[0].Name != "Uni \\& College" {
		t.Errorf("unexpected escaped education name: %q", formattedEducation[0].Name)
	}
	if formattedEducation[0].Location != "NY\\_1" {
		t.Errorf("unexpected escaped education location: %q", formattedEducation[0].Location)
	}

	formattedSkills := FormatSkills(skills)
	if formattedSkills[1] != "C\\#" {
		t.Errorf("unexpected escaped skill: %q", formattedSkills[1])
	}
	if formattedSkills[2] != "Infra\\_1" {
		t.Errorf("unexpected escaped skill: %q", formattedSkills[2])
	}
}

func TestResolvedFontSpec(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string defaults", input: "", want: ""},
		{name: "lmodern alias defaults", input: "lmodern", want: ""},
		{name: "times", input: "times", want: `\usepackage{mathptmx}`},
		{name: "palatino", input: "palatino", want: `\usepackage{mathpazo}`},
		{name: "charter", input: "charter", want: `\usepackage{charter}`},
		{name: "helvetica", input: "helvetica", want: "\\usepackage{helvet}\n\\renewcommand{\\familydefault}{\\sfdefault}"},
		{name: "unknown key falls back", input: "comic-sans", want: ""},
		{name: "injection attempt falls back", input: `}\usepackage{shellesc}\directlua{os.execute("id")}%`, want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolvedFontSpec(tc.input)
			if got != tc.want {
				t.Errorf("resolvedFontSpec(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestFormatTemplateData_FontFamily(t *testing.T) {
	base := &model.Resume{
		Profile: model.Profile{FirstName: "Jane", LastName: "Doe"},
	}
	tests := []struct {
		name       string
		fontFamily string
		want       string
	}{
		{name: "known font maps to package snippet", fontFamily: "times", want: `\usepackage{mathptmx}`},
		{name: "lmodern resolves to empty", fontFamily: "lmodern", want: ""},
		{name: "unknown font resolves to empty", fontFamily: "not-a-font", want: ""},
		{name: "empty resolves to empty", fontFamily: "", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := *base
			r.TemplateSettings = model.TemplateSettings{FontFamily: tc.fontFamily}
			got := FormatTemplateData(&r)
			if got.FontFamily != tc.want {
				t.Errorf("FontFamily = %q, want %q", got.FontFamily, tc.want)
			}
		})
	}
}

func TestFormatTemplateData(t *testing.T) {
	updated := time.Date(2026, time.February, 23, 12, 30, 0, 0, time.UTC)
	resume := &model.Resume{
		Profile: model.Profile{
			FirstName: "Jane",
			LastName:  "Doe",
			Address1:  "1 Main St",
			City:      "Austin",
			State:     "US-TX",
			ZipCode:   "78701",
			Country:   "USA",
		},
		Summary: "Built 100% of core",
		Experience: []model.Experience{{
			Title: "Engineer",
		}},
		Education: []model.Education{{
			Name: "State University",
		}},
		Skills: []string{"Go", "C#"},
		TemplateSettings: model.TemplateSettings{
			FontFamily: "lmodern",
		},
		Updated: updated,
	}

	got := FormatTemplateData(resume)

	if got.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want %q", got.FullName, "Jane Doe")
	}
	if got.Date != "02-23-2026" {
		t.Errorf("Date = %q, want %q", got.Date, "02-23-2026")
	}
	if got.Address != "1 Main St, Austin, TX, 78701, USA" {
		t.Errorf("Address = %q, want %q", got.Address, "1 Main St, Austin, TX, 78701, USA")
	}
	if got.Summary != "Built 100\\% of core" {
		t.Errorf("Summary = %q, want escaped summary", got.Summary)
	}
	if got.FontFamily != "" {
		t.Errorf("FontFamily = %q, want %q", got.FontFamily, "")
	}
	if len(got.Skills) != 2 || got.Skills[1] != "C\\#" {
		t.Errorf("Skills not formatted as expected: %#v", got.Skills)
	}
}
