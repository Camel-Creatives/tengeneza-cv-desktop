// render-samples creates reviewable PDF and DOCX files from the fictional JSON fixtures.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render"
)

func main() {
	input := flag.String("input", "examples/sample-cvs", "directory containing sample JSON fixtures")
	output := flag.String("output", "/tmp/tengenezacv-rendered-samples", "directory for rendered documents")
	flag.Parse()
	files, err := filepath.Glob(filepath.Join(*input, "*.json"))
	if err != nil || len(files) == 0 {
		panic("no sample fixtures found")
	}
	service := render.New(*output, "/usr/bin/chromium")
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		var fixture struct {
			Title         string         `json:"title"`
			TemplateID    string         `json:"template_id"`
			TargetRegion  string         `json:"target_region"`
			DocumentStyle string         `json:"document_style"`
			Language      string         `json:"language"`
			Content       map[string]any `json:"content"`
		}
		if err := json.Unmarshal(raw, &fixture); err != nil {
			panic(err)
		}
		template := domain.Template{ID: fixture.TemplateID, Family: "ats"}
		switch fixture.TemplateID {
		case "visual_local", "hospitality_local", "creative_portfolio", "tech_product", "finance_audit", "legal_chambers", "clinical_care", "educator_classroom", "sales_growth", "field_operations", "consulting_strategy", "engineering_blueprint", "public_service", "nonprofit_impact", "graduate_launch", "skilled_trades", "remote_async", "creative_producer":
			template.Family = "visual"
			template.SupportsPhoto = true
		}
		cv := domain.CV{Title: fixture.Title, TemplateID: fixture.TemplateID, TargetRegion: fixture.TargetRegion, DocumentStyle: fixture.DocumentStyle, Language: fixture.Language, Content: fixture.Content}
		for _, format := range []string{"pdf", "docx"} {
			artifact, err := service.Render(context.Background(), cv, template, format, false)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s\n", artifact.Path)
		}
	}
}
