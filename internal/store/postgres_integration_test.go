package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
)

func TestPostgresWorkspacePersistsAndEnforcesDailyQuota(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	s, err := OpenPostgres(context.Background(), dsn)
	require.NoError(t, err)
	subject := "integration-" + time.Now().UTC().Format("20060102150405.000000000")
	user, err := s.EnsureExternalUser(subject, domain.User{FullName: "Integration User", Email: subject + "@example.test", Phone: "+255700000001"})
	require.NoError(t, err)
	refreshed, err := s.EnsureExternalUser(subject, domain.User{FullName: "Updated Integration User", Email: subject + "@example.test", Phone: "+255700000001"})
	require.NoError(t, err)
	require.Equal(t, user.ID, refreshed.ID, "repeat CamelAccounts provisioning must return the same workspace user")
	require.Equal(t, "Updated Integration User", refreshed.FullName, "latest CamelAccounts claims should update the profile")
	cv, err := s.CreateCV(domain.CV{UserID: user.ID, Title: "Quota", TemplateID: "ats_classic", Content: map[string]any{"summary": "before"}})
	require.NoError(t, err)
	for i := 0; i < 5; i++ {
		_, err = s.ConsumeExportCredit(user.ID, cv.ID, "fmt"+string(rune('a'+i)))
		require.NoError(t, err)
	}
	_, err = s.ConsumeExportCredit(user.ID, cv.ID, "sixth")
	require.ErrorIs(t, err, ErrDailyExportLimit)
	require.Equal(t, 0, s.DailyExportsRemaining(user.ID, time.Now()))
	cv, err = s.PatchCV(user.ID, cv.ID, map[string]any{"summary": "persisted"}, "test")
	require.NoError(t, err)
	s.Close()
	s, err = OpenPostgres(context.Background(), dsn)
	require.NoError(t, err)
	defer s.Close()
	restored, err := s.CV(user.ID, cv.ID)
	require.NoError(t, err)
	require.Equal(t, "persisted", restored.Content["summary"])
	require.Len(t, s.CVs(user.ID), 1)
}
