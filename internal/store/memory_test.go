package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
)

func TestCVWorkspaceIsolation(t *testing.T) {
	s := NewMemoryStore()
	u1, err := s.EnsureExternalUser("user-one")
	require.NoError(t, err)
	_, err = s.EnsureExternalUser("user-two")
	require.NoError(t, err)
	cv, err := s.CreateCV(domain.CV{UserID: u1.ID, TemplateID: "ats_classic", Content: map[string]any{"summary": "private"}})
	require.NoError(t, err)
	require.Len(t, s.CVs("user-two"), 0)
	_, err = s.CV("user-two", cv.ID)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = s.PatchCV("user-two", cv.ID, map[string]any{"summary": "changed"}, "attack")
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, s.ArchiveCV("user-two", cv.ID), ErrNotFound)
	_, err = s.ConsumeExportCredit("user-two", cv.ID, "csv")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestDailyExportLimitUsesTanzaniaDayAndAllowsCachedArtifacts(t *testing.T) {
	s := NewMemoryStore()
	user, err := s.EnsureExternalUser("quota-user")
	require.NoError(t, err)
	cv, err := s.CreateCV(domain.CV{UserID: user.ID, TemplateID: "ats_classic", Content: map[string]any{"summary": "A"}})
	require.NoError(t, err)
	for i := 0; i < DailyExportLimit; i++ {
		_, err = s.ConsumeExportCredit(user.ID, cv.ID, "csv")
		require.NoError(t, err)
		cv, err = s.PatchCV(user.ID, cv.ID, map[string]any{"summary": string(rune('B' + i))}, "change")
		require.NoError(t, err)
	}
	require.Equal(t, 0, s.DailyExportsRemaining(user.ID, time.Now()))
	_, err = s.ConsumeExportCredit(user.ID, cv.ID, "csv")
	require.ErrorIs(t, err, ErrDailyExportLimit)
	local, err := time.LoadLocation("Africa/Dar_es_Salaam")
	require.NoError(t, err)
	tomorrow := time.Now().In(local).AddDate(0, 0, 1)
	tomorrow = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 1, 0, 0, 0, local)
	require.Equal(t, DailyExportLimit, s.DailyExportsRemaining(user.ID, tomorrow))
}

func TestExportDownloadLookupAcceptsObjectKeyAndBasenameOnlyForOwner(t *testing.T) {
	s := NewMemoryStore()
	user, err := s.EnsureExternalUser("export-owner")
	require.NoError(t, err)
	other, err := s.EnsureExternalUser("export-other")
	require.NoError(t, err)
	cv, err := s.CreateCV(domain.CV{UserID: user.ID, TemplateID: "ats_classic", Content: map[string]any{"summary": "Export me"}})
	require.NoError(t, err)
	export, err := s.ConsumeExportCredit(user.ID, cv.ID, "csv")
	require.NoError(t, err)
	objectKey := "nested/path/Person_CV_" + cv.ID + "_" + user.ID + ".csv"
	require.NoError(t, s.CompleteExport(user.ID, export.ID, objectKey, 42))

	byBasename, err := s.ExportForUser(user.ID, filepath.Base(objectKey))
	require.NoError(t, err)
	require.Equal(t, objectKey, byBasename.ObjectKey)
	cached, err := s.ConsumeExportCredit(user.ID, cv.ID, "csv")
	require.NoError(t, err)
	require.True(t, cached.Cached)
	require.Equal(t, filepath.Base(objectKey), filepath.Base(cached.ObjectKey))
	byKey, err := s.ExportForUser(user.ID, objectKey)
	require.NoError(t, err)
	require.Equal(t, objectKey, byKey.ObjectKey)
	_, err = s.ExportForUser(other.ID, filepath.Base(objectKey))
	require.ErrorIs(t, err, ErrNotFound)
}
