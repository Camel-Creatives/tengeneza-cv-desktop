package render

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render/model"
)

func TestRenderCSVIncludesSectionsAndEscapesSpreadsheetContent(t *testing.T) {
	doc := model.Document{Personal: model.Personal{FullName: "Asha, Mollel", Email: "asha@example.test"}, Summary: "=1+1", Experience: []model.Experience{{Employer: "Clinic", Title: "Nurse", Bullets: []string{"Care, triage"}}}, HardSkills: []string{"Excel"}}
	out, err := renderCSV(doc)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(out), "\ufeff"))
	rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(out), "\ufeff"))).ReadAll()
	require.NoError(t, err)
	require.Equal(t, []string{"section", "field", "value"}, rows[0])
	require.Contains(t, rows, []string{"personal", "full_name", "Asha, Mollel"})
	require.Contains(t, rows, []string{"profile", "summary", "=1+1"})
	require.Contains(t, rows, []string{"experience_1", "achievement_1", "Care, triage"})
	require.Contains(t, rows, []string{"skills", "hard_skill_1", "Excel"})
}
