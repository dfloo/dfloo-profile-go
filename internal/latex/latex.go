package latex

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/dfloo/dfloo-profile-go/internal/model"
)

type TemplateData struct {
	model.Resume
	FullName   string
	Date       string
	Address    string
	FontFamily string
	Experience []model.Experience
	Education  []model.Education
	Skills     []string
}

func GenerateFromResume(resume *model.Resume) (string, error) {
	template, err := template.ParseFiles("./internal/latex/templates/default.tex")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return "", err
	}

	tempDir, err := os.MkdirTemp("", "resume")
	if err != nil {
		log.Printf("Error creating temp dir: %v", err)
		return "", err
	}

	filePath := tempDir + "/resume.tex"
	output, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating output file: %v", err)
		return "", err
	}
	defer output.Close()

	err = template.Execute(output, FormatTemplateData(resume))
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return "", err
	}

	return filePath, nil
}

func ConvertToPDF(filePath string) ([]byte, error) {
	dir := ""
	if idx := len(filePath) - len("/resume.tex"); idx > 0 {
		dir = filePath[:idx]
	}

	cmd := exec.Command("lualatex", "-output-directory", dir, filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Error converting LaTeX to PDF: %v", err)
		return nil, err
	}

	pdfPath := dir + "/resume.pdf"
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		log.Printf("Error reading PDF file: %v", err)
		return nil, err
	}
	return pdfBytes, nil
}

func FormatTemplateData(resume *model.Resume) TemplateData {
	return TemplateData{
		Resume:     *resume,
		FullName:   FormatFullName(&resume.Profile),
		Date:       resume.Updated.Format("01-02-2006"),
		Address:    FormatAddress(&resume.Profile),
		Experience: FormatExperience(resume.Experience),
		Education:  FormatEducation(resume.Education),
		Skills:     FormatSkills(resume.Skills),
		FontFamily: resume.TemplateSettings.FontFamily,
	}
}

func EscapeSpecialChars(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\textbackslash{}")
	s = strings.ReplaceAll(s, "&", "\\&")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "$", "\\$")
	s = strings.ReplaceAll(s, "#", "\\#")
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")
	s = strings.ReplaceAll(s, "~", "\\textasciitilde{}")
	s = strings.ReplaceAll(s, "^", "\\textasciicircum{}")
	s = strings.ReplaceAll(s, "®", "\\textregistered{}")
	return s
}

func FormatFullName(profile *model.Profile) string {
	parts := []string{}
	if profile.FirstName != "" {
		parts = append(parts, profile.FirstName)
	}
	if profile.MiddleName != "" {
		parts = append(parts, profile.MiddleName)
	}
	if profile.LastName != "" {
		parts = append(parts, profile.LastName)
	}
	return strings.Join(parts, " ")
}

func FormatAddress(profile *model.Profile) string {
	parts := []string{}
	if profile.Address1 != "" {
		parts = append(parts, profile.Address1)
	}
	if profile.Address2 != "" {
		parts = append(parts, profile.Address2)
	}
	if profile.City != "" {
		parts = append(parts, profile.City)
	}
	if profile.State != "" {
		parts = append(parts, GetStateShortCode(profile.State))
	}
	if profile.ZipCode != "" {
		parts = append(parts, profile.ZipCode)
	}
	if profile.Country != "" {
		parts = append(parts, profile.Country)
	}
	return EscapeSpecialChars(strings.Join(parts, ", "))
}

func GetStateShortCode(stateCode string) string {
	codeComponents := strings.Split(stateCode, "-")
	var shortCode string
	if len(codeComponents) > 1 {
		shortCode = codeComponents[1]
	}

	return shortCode
}

func FormatExperience(experience []model.Experience) []model.Experience {
	escaped := make([]model.Experience, len(experience))
	for i, exp := range experience {
		exp.Title = EscapeSpecialChars(exp.Title)
		exp.Employer = EscapeSpecialChars(exp.Employer)
		exp.Location = EscapeSpecialChars(exp.Location)
		exp.Description = EscapeSpecialChars(exp.Description)
		exp.StartDate = EscapeSpecialChars(exp.StartDate)
		exp.EndDate = EscapeSpecialChars(exp.EndDate)
		for j, bp := range exp.BulletPoints {
			exp.BulletPoints[j] = EscapeSpecialChars(bp)
		}
		escaped[i] = exp
	}
	return escaped
}

func FormatEducation(education []model.Education) []model.Education {
	escaped := make([]model.Education, len(education))
	for i, edu := range education {
		edu.Type = EscapeSpecialChars(edu.Type)
		edu.Name = EscapeSpecialChars(edu.Name)
		edu.Location = EscapeSpecialChars(edu.Location)
		edu.CompletionDate = EscapeSpecialChars(edu.CompletionDate)
		escaped[i] = edu
	}
	return escaped
}

func FormatSkills(skills []string) []string {
	escaped := make([]string, len(skills))
	for i, skill := range skills {
		escaped[i] = EscapeSpecialChars(skill)
	}
	return escaped
}
