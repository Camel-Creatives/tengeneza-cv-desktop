package lint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
)

func TestFictionalSampleCVsMeetCoreATSChecks(t *testing.T) {
	files, err := filepath.Glob("../../../examples/sample-cvs/*.json")
	require.NoError(t, err)
	require.Len(t, files, 24)

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			raw, err := os.ReadFile(file)
			require.NoError(t, err)
			var fixture struct {
				Title        string         `json:"title"`
				TemplateID   string         `json:"template_id"`
				TargetRegion string         `json:"target_region"`
				Content      map[string]any `json:"content"`
			}
			require.NoError(t, json.Unmarshal(raw, &fixture))
			family := "ats"
			supportsPhoto := false
			switch fixture.TemplateID {
			case "visual_local", "hospitality_local", "creative_portfolio", "tech_product", "finance_audit", "legal_chambers", "clinical_care", "educator_classroom", "sales_growth", "field_operations", "consulting_strategy", "engineering_blueprint", "public_service", "nonprofit_impact", "graduate_launch", "skilled_trades", "remote_async", "creative_producer":
				family = "visual"
				supportsPhoto = true
			}
			report := Score(domain.CV{Title: fixture.Title, TemplateID: fixture.TemplateID, TargetRegion: fixture.TargetRegion, Content: fixture.Content}, domain.Template{ID: fixture.TemplateID, Family: family, SupportsPhoto: supportsPhoto})
			require.GreaterOrEqual(t, report.Score, 90)
			for _, result := range report.Breakdown {
				require.NotEqual(t, "error", result.Severity)
			}
		})
	}
}
