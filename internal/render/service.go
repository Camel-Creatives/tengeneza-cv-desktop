package render

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	stdhtml "html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render/docx"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render/html"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render/model"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render/pdf"
)

type Service struct{ OutputDir, ChromiumExecutable, LocalUploadDir, ExamplesDir string }
type Artifact struct {
	Path, Filename, MIMEType string
	Bytes                    int
}

func New(outputDir, chromiumExecutable string, uploadDirectories ...string) *Service {
	if outputDir == "" {
		outputDir = "/tmp/tengenezacv-exports"
	}
	if chromiumExecutable == "" {
		chromiumExecutable = "/usr/bin/chromium"
	}
	uploadDir := "/tmp/tengenezacv-uploads"
	if len(uploadDirectories) > 0 && uploadDirectories[0] != "" {
		uploadDir = uploadDirectories[0]
	}
	examplesDir := "examples/sample-cvs"
	if len(uploadDirectories) > 1 && uploadDirectories[1] != "" {
		examplesDir = uploadDirectories[1]
	}
	return &Service{OutputDir: outputDir, ChromiumExecutable: chromiumExecutable, LocalUploadDir: uploadDir, ExamplesDir: examplesDir}
}
func (s *Service) Render(ctx context.Context, cv domain.CV, template domain.Template, format string, watermark bool) (Artifact, error) {
	doc := model.Normalize(cv, template, watermark)
	s.embedLocalPhoto(&doc)
	localeSuffix := ""
	if doc.Language == "sw" {
		localeSuffix = "_sw"
	}
	filename := model.FileStem(doc) + localeSuffix + "_" + cv.ID + "_" + cv.UserID + "." + format
	path := filepath.Join(s.OutputDir, filename)
	if err := os.MkdirAll(s.OutputDir, 0750); err != nil {
		return Artifact{}, err
	}
	var raw []byte
	var err error
	mime := ""
	switch format {
	case "csv":
		raw, err = renderCSV(doc)
		mime = "text/csv; charset=utf-8"
	case "pdf":
		source, renderErr := html.Renderer{}.Render(doc)
		if renderErr != nil {
			return Artifact{}, renderErr
		}
		raw, err = pdf.Renderer{Executable: s.ChromiumExecutable}.Render(ctx, source)
		mime = "application/pdf"
	case "docx":
		raw, err = docx.Renderer{}.Render(doc)
		mime = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return Artifact{}, fmt.Errorf("unsupported format %s", format)
	}
	if err != nil {
		return Artifact{}, err
	}
	if err := os.WriteFile(path, raw, 0640); err != nil {
		return Artifact{}, err
	}
	return Artifact{Path: path, Filename: filename, MIMEType: mime, Bytes: len(raw)}, nil
}
func (s *Service) embedLocalPhoto(doc *model.Document) {
	source := string(doc.Personal.PhotoURL)
	if source == "" || strings.HasPrefix(source, "data:") {
		return
	}

	var raw []byte

	// Handle remote HTTP/HTTPS photo URLs
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, reqErr := client.Get(source)
		if reqErr == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			buf, readErr := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
			if readErr == nil && len(buf) > 0 {
				raw = buf
			}
		}
	}

	// Handle local files if remote fetch wasn't performed or failed
	if len(raw) == 0 {
		filename := filepath.Base(source)
		cleanPath := source
		if strings.HasPrefix(cleanPath, "file://") {
			cleanPath = strings.TrimPrefix(cleanPath, "file://")
		}

		pathsToTry := []string{
			cleanPath,
			filepath.Join(s.LocalUploadDir, strings.TrimPrefix(cleanPath, "/uploads/")),
			filepath.Join(s.LocalUploadDir, "profiles", filename),
			filepath.Join(s.LocalUploadDir, filename),
			filepath.Join(s.ExamplesDir, "profile-assets", filename),
			filepath.Join(s.ExamplesDir, "profile-assets", "library", filename),
			filepath.Join(s.ExamplesDir, "..", "profile-assets", filename),
			filepath.Join(s.ExamplesDir, "..", "profile-assets", "library", filename),
			filepath.Join("API", "examples", "profile-assets", "library", filename),
			filepath.Join("API", "examples", "profile-assets", filename),
			filepath.Join("examples", "profile-assets", "library", filename),
			filepath.Join("examples", "profile-assets", filename),
			filepath.Join("UI", "public", "uploads", "profiles", filename),
			filepath.Join("/home/cameltech/Projects/tengeneza-cv/UI/public/uploads/profiles", filename),
			filepath.Join("/home/cameltech/Projects/tengeneza-cv/UI/public/profiles", filename),
		}

		for _, p := range pathsToTry {
			if p == "" {
				continue
			}
			if data, readErr := os.ReadFile(p); readErr == nil && len(data) > 0 && len(data) <= 5<<20 {
				raw = data
				break
			}
		}
	}

	if len(raw) == 0 {
		return
	}

	mimeType := http.DetectContentType(raw)
	if !strings.HasPrefix(mimeType, "image/") {
		mimeType = "image/jpeg"
	}

	doc.Personal.PhotoURL = stdhtml.URL("data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(raw))
}

// renderCSV emits a single spreadsheet-friendly row per CV fact. Every value is
// quoted by encoding/csv, so user-controlled text cannot break the CSV structure.
func renderCSV(doc model.Document) ([]byte, error) {
	var out strings.Builder
	writer := csv.NewWriter(&out)
	rows := [][]string{{"section", "field", "value"}}
	add := func(section, field, value string) {
		if strings.TrimSpace(value) != "" {
			rows = append(rows, []string{section, field, value})
		}
	}
	add("personal", "full_name", doc.Personal.FullName)
	add("personal", "headline", doc.Personal.Headline)
	add("personal", "email", doc.Personal.Email)
	add("personal", "phone", doc.Personal.Phone)
	add("personal", "location", doc.Personal.Location)
	add("personal", "linkedin_url", doc.Personal.LinkedIn)
	add("personal", "portfolio_url", doc.Personal.Portfolio)
	add("profile", "summary", doc.Summary)
	for i, item := range doc.Experience {
		prefix := fmt.Sprintf("experience_%d", i+1)
		add(prefix, "employer", item.Employer)
		add(prefix, "title", item.Title)
		add(prefix, "location", item.Location)
		add(prefix, "start_date", item.StartDate)
		add(prefix, "end_date", item.EndDate)
		for j, bullet := range item.Bullets {
			add(prefix, fmt.Sprintf("achievement_%d", j+1), bullet)
		}
	}
	for i, item := range doc.Education {
		prefix := fmt.Sprintf("education_%d", i+1)
		add(prefix, "institution", item.Institution)
		add(prefix, "credential", item.Credential)
		add(prefix, "classification", item.Classification)
		add(prefix, "start_date", item.StartDate)
		add(prefix, "end_date", item.EndDate)
		add(prefix, "location", item.Location)
	}
	for i, item := range doc.Certifications {
		prefix := fmt.Sprintf("certification_%d", i+1)
		add(prefix, "name", item.Name)
		add(prefix, "issuer", item.Issuer)
		add(prefix, "year", item.Year)
	}
	for i, value := range doc.HardSkills {
		add("skills", fmt.Sprintf("hard_skill_%d", i+1), value)
	}
	for i, value := range doc.SoftSkills {
		add("skills", fmt.Sprintf("soft_skill_%d", i+1), value)
	}
	for i, item := range doc.Languages {
		prefix := fmt.Sprintf("language_%d", i+1)
		add("languages", prefix, item.Name)
		add("languages", prefix+"_proficiency", item.Proficiency)
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return []byte("\ufeff" + out.String()), nil
}
