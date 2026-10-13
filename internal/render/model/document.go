// Package model defines the normalized document data shared by all renderers.
package model

import (
	"fmt"
	htmltemplate "html/template"
	"strings"

	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
)

type Document struct {
	Personal       Personal
	Language       string
	Summary        string
	Experience     []Experience
	Education      []Education
	Certifications []Certification
	HardSkills     []string
	SoftSkills     []string
	Languages      []Language
	ShowPhoto      bool
	ShowReferees   bool
	TemplateID     string
	TemplateFamily string
	TargetRegion   string
	Watermark      bool
}
type Personal struct {
	FullName, Headline, Email, Phone, Location, LinkedIn, Portfolio string
	PhotoURL                                                        htmltemplate.URL
}
type Experience struct {
	Employer, Title, Location, StartDate, EndDate string
	Current                                       bool
	Bullets                                       []string
}
type Education struct{ Institution, Credential, Classification, StartDate, EndDate, Location string }
type Certification struct{ Name, Issuer, Year string }
type Language struct{ Name, Proficiency string }

func Normalize(cv domain.CV, template domain.Template, watermark bool) Document {
	content := cv.Content
	p := mapValue(content["personal"])
	doc := Document{Personal: Personal{FullName: stringValue(p["full_name"]), Headline: stringValue(p["headline"]), Email: stringValue(p["email"]), Phone: stringValue(p["phone"]), Location: stringValue(p["location"]), LinkedIn: stringValue(p["linkedin_url"]), Portfolio: stringValue(p["portfolio_url"]), PhotoURL: htmltemplate.URL(stringValue(p["photo_url"]))}, Language: cv.Language, Summary: stringValue(content["summary"]), TemplateID: template.ID, TemplateFamily: template.Family, TargetRegion: cv.TargetRegion, Watermark: watermark}
	settings := mapValue(content["settings"])
	showPhoto := true
	if v, ok := settings["show_photo"].(bool); ok {
		showPhoto = v
	}
	if v, ok := p["show_photo"].(bool); ok {
		showPhoto = v
	}
	doc.ShowPhoto = showPhoto && doc.Personal.PhotoURL != ""
	doc.ShowReferees = boolValue(settings["show_referees"])
	for _, raw := range sliceValue(content["experience"]) {
		item := mapValue(raw)
		doc.Experience = append(doc.Experience, Experience{Employer: stringValue(item["employer"]), Title: stringValue(item["title"]), Location: stringValue(item["location"]), StartDate: formatDate(stringValue(item["start_date"]), cv.Language), EndDate: formatDate(stringValue(item["end_date"]), cv.Language), Current: boolValue(item["is_current"]), Bullets: stringSlice(item["bullets"])})
	}
	for _, raw := range sliceValue(content["education"]) {
		item := mapValue(raw)
		doc.Education = append(doc.Education, Education{Institution: stringValue(item["institution"]), Credential: stringValue(item["credential"]), Classification: firstNonBlank(stringValue(item["classification_local"]), stringValue(item["classification_international_equiv"])), StartDate: formatDate(stringValue(item["start_date"]), cv.Language), EndDate: formatDate(stringValue(item["end_date"]), cv.Language), Location: stringValue(item["location"])})
	}
	for _, raw := range sliceValue(content["certifications"]) {
		item := mapValue(raw)
		doc.Certifications = append(doc.Certifications, Certification{Name: stringValue(item["name"]), Issuer: stringValue(item["issuer"]), Year: stringValue(item["year"])})
	}
	skills := mapValue(content["skills"])
	doc.HardSkills = stringSlice(skills["hard_skills"])
	doc.SoftSkills = stringSlice(skills["soft_skills"])
	for _, raw := range sliceValue(skills["languages"]) {
		item := mapValue(raw)
		doc.Languages = append(doc.Languages, Language{Name: stringValue(item["language"]), Proficiency: stringValue(item["proficiency"])})
	}
	return doc
}
func FileStem(doc Document) string {
	value := strings.TrimSpace(strings.ReplaceAll(doc.Personal.FullName, " ", "_"))
	if value == "" {
		return "Tengeneza_CV"
	}
	clean := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			return r
		}
		return -1
	}, value)
	return fmt.Sprintf("%s_CV", clean)
}
func mapValue(v any) map[string]any {
	if value, ok := v.(map[string]any); ok {
		return value
	}
	return map[string]any{}
}
func sliceValue(v any) []any {
	if value, ok := v.([]any); ok {
		return value
	}
	return []any{}
}
func stringValue(v any) string {
	if value, ok := v.(string); ok {
		return value
	}
	return ""
}
func boolValue(v any) bool { value, _ := v.(bool); return value }
func stringSlice(v any) []string {
	items := sliceValue(v)
	out := make([]string, 0, len(items))
	for _, item := range items {
		if value := stringValue(item); value != "" {
			out = append(out, value)
		}
	}
	return out
}
func formatDate(value, language string) string {
	if value == "" {
		return Translate(language, "Present")
	}
	if len(value) == 7 && value[4] == '-' {
		return value[5:7] + "/" + value[:4]
	}
	return value
}
func firstNonBlank(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
