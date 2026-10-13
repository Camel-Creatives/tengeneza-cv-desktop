package render

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render/docx"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render/html"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render/model"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/store"
)

func TestHTMLAndDOCXRenderersProduceSemanticDocumentOutput(t *testing.T) {
	document := model.Normalize(domain.CV{TargetRegion: "eac", Content: map[string]any{"personal": map[string]any{"full_name": "Amina Juma", "email": "amina@example.test"}, "summary": "Compliance professional.", "experience": []any{map[string]any{"title": "Analyst", "employer": "Demo Bank", "start_date": "2022-01", "is_current": true, "bullets": []any{"Led a 12-person review, reducing exceptions by 25%."}}}, "education": []any{}, "skills": map[string]any{"hard_skills": []any{"IFRS", "SQL"}}, "settings": map[string]any{"show_photo": false}}}, domain.Template{Family: "ats"}, false)
	htmlOutput, err := html.Renderer{}.Render(document)
	require.NoError(t, err)
	require.Contains(t, htmlOutput, "Professional Summary")
	require.Contains(t, htmlOutput, "Work Experience")

	docxBytes, err := docx.Renderer{}.Render(document)
	require.NoError(t, err)
	archive, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	require.NoError(t, err)
	entries := map[string]bool{}
	for _, file := range archive.File {
		entries[file.Name] = true
	}
	require.True(t, entries["word/document.xml"])
	require.True(t, entries["word/styles.xml"])
	require.True(t, entries["word/numbering.xml"])
}

func TestSwahiliLocaleTranslatesHTMLAndDOCXHeadings(t *testing.T) {
	cv := domain.CV{Language: "sw", TargetRegion: "eac", Content: map[string]any{
		"personal":   map[string]any{"full_name": "Amina Juma"},
		"summary":    "Mtaalamu wa uzingatiaji.",
		"experience": []any{map[string]any{"title": "Mchambuzi", "employer": "Benki", "start_date": "2022-01", "is_current": true}},
	}}
	document := model.Normalize(cv, domain.Template{Family: "ats"}, false)
	htmlOutput, err := html.Renderer{}.Render(document)
	require.NoError(t, err)
	require.Contains(t, htmlOutput, "Muhtasari wa Kitaaluma")
	require.Contains(t, htmlOutput, "Uzoefu wa Kazi")
	require.Contains(t, htmlOutput, "Sasa")
	require.NotContains(t, htmlOutput, "Professional Summary")

	docxBytes, err := docx.Renderer{}.Render(document)
	require.NoError(t, err)
	archive, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	require.NoError(t, err)
	var documentXML []byte
	for _, file := range archive.File {
		if file.Name != "word/document.xml" {
			continue
		}
		reader, openErr := file.Open()
		require.NoError(t, openErr)
		documentXML, err = io.ReadAll(reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
	}
	require.NotEmpty(t, documentXML)
	require.Contains(t, string(documentXML), "Muhtasari wa Kitaaluma")
	require.Contains(t, string(documentXML), "Uzoefu wa Kazi")
	require.Contains(t, string(documentXML), "Sasa")
}

func TestRenderServiceExportsCSVAndDOCXForEveryActiveTemplate(t *testing.T) {
	catalog := store.NewMemoryStore()
	service := New(t.TempDir(), "", t.TempDir(), t.TempDir())
	for _, template := range catalog.Templates() {
		cv := domain.CV{ID: "cv-" + template.ID, UserID: "user", Title: "Export QA", TemplateID: template.ID, Content: map[string]any{"personal": map[string]any{"full_name": "Test Person", "email": "person@example.test"}, "summary": "CSV and DOCX QA"}}
		for _, format := range []string{"csv", "docx"} {
			artifact, err := service.Render(context.Background(), cv, template, format, false)
			require.NoError(t, err, template.ID+" "+format)
			raw, err := os.ReadFile(artifact.Path)
			require.NoError(t, err)
			require.NotEmpty(t, raw)
			if format == "csv" {
				rows, err := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))).ReadAll()
				require.NoError(t, err)
				require.Greater(t, len(rows), 1)
				require.Equal(t, filepath.Base(artifact.Path), artifact.Filename)
			}
		}
	}
}
