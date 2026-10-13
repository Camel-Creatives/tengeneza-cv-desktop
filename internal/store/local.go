package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/troubleman96/tengeneza-cv-desktop/internal/auth"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
)

const LocalDesktopPlan = "desktop-unlimited"

// LocalStore persists the existing repository contract to a private JSON file.
// The Wails app is single-user and offline-first, so no account or database is
// needed; writes are serialized and atomically replaced on disk.
type LocalStore struct {
	mu     sync.Mutex
	memory *MemoryStore
	path   string
}

type localSnapshot struct {
	Version   int                           `json:"version"`
	Users     map[string]domain.User        `json:"users"`
	ByEmail   map[string]string             `json:"by_email"`
	ByPhone   map[string]string             `json:"by_phone"`
	OTPs      map[string]auth.OTP           `json:"otps,omitempty"`
	CVs       map[string]domain.CV          `json:"cvs"`
	Versions  map[string][]domain.CVVersion `json:"versions"`
	Exports   map[string][]domain.Export    `json:"exports"`
	Templates map[string]domain.Template    `json:"templates"`
}

// OpenLocalStore opens or creates a local-only workspace. The file and parent
// directory are restricted to the current OS user because CVs contain personal
// information.
func OpenLocalStore(path string) (*LocalStore, error) {
	if path == "" {
		return nil, errors.New("local store path is required")
	}
	store := &LocalStore{memory: NewMemoryStore(), path: path}
	loaded := false
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	if raw, err := os.ReadFile(path); err == nil {
		loaded = true
		var snapshot localSnapshot
		if err := json.Unmarshal(raw, &snapshot); err != nil {
			return nil, errors.New("local workspace file is invalid; preserve it and restore from a backup")
		}
		if snapshot.Version != 1 {
			return nil, errors.New("local workspace file version is unsupported")
		}
		store.memory.users = nonNil(snapshot.Users)
		store.memory.byEmail = nonNil(snapshot.ByEmail)
		store.memory.byPhone = nonNil(snapshot.ByPhone)
		store.memory.otps = nonNil(snapshot.OTPs)
		store.memory.cvs = nonNil(snapshot.CVs)
		store.memory.versions = nonNil(snapshot.Versions)
		store.memory.exports = nonNil(snapshot.Exports)
		if len(snapshot.Templates) > 0 {
			store.memory.templates = snapshot.Templates
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	user, ok := store.memory.users[DefaultUserID]
	if !ok {
		user = domain.User{ID: DefaultUserID, PreferredLanguage: "en", TargetRegion: "eac", Role: domain.RoleUser, CreatedAt: time.Now().UTC()}
	}
	if !loaded {
		user.FullName = ""
		user.Email = ""
		user.Phone = ""
		store.memory.byEmail = map[string]string{}
		store.memory.byPhone = map[string]string{}
	}
	user.Plan = LocalDesktopPlan
	user.ExportCredits = 0
	user.AITokenBalance = 0
	store.memory.users[DefaultUserID] = user
	if err := store.persistLocked(); err != nil {
		return nil, err
	}
	return store, nil
}

func nonNil[T any](value map[string]T) map[string]T {
	if value == nil {
		return map[string]T{}
	}
	return value
}

func (s *LocalStore) snapshot() localSnapshot {
	s.memory.mu.RLock()
	defer s.memory.mu.RUnlock()
	return localSnapshot{
		Version: 1,
		Users:   s.memory.users, ByEmail: s.memory.byEmail, ByPhone: s.memory.byPhone,
		OTPs: s.memory.otps, CVs: s.memory.cvs, Versions: s.memory.versions,
		Exports: s.memory.exports, Templates: s.memory.templates,
	}
}

func (s *LocalStore) persistLocked() error {
	raw, err := json.Marshal(s.snapshot())
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(s.path), ".workspace-*.tmp")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer os.Remove(temp)
	if err := file.Chmod(0600); err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(raw); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temp, s.path)
}

func (s *LocalStore) write(fn func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := fn(); err != nil {
		return err
	}
	return s.persistLocked()
}

func (s *LocalStore) Ping(ctx context.Context) error         { return s.memory.Ping(ctx) }
func (s *LocalStore) GetUser(id string) (domain.User, error) { return s.memory.GetUser(id) }
func (s *LocalStore) EnsureExternalUser(id string, claims ...domain.User) (domain.User, error) {
	var result domain.User
	err := s.write(func() error { var err error; result, err = s.memory.EnsureExternalUser(id, claims...); return err })
	return result, err
}
func (s *LocalStore) UpdateUser(id string, mutate func(*domain.User)) (domain.User, error) {
	var result domain.User
	err := s.write(func() error { var err error; result, err = s.memory.UpdateUser(id, mutate); return err })
	return result, err
}
func (s *LocalStore) SaveOTP(otp auth.OTP) {
	_ = s.write(func() error { s.memory.SaveOTP(otp); return nil })
}
func (s *LocalStore) GetOTP(id string) (auth.OTP, error) { return s.memory.GetOTP(id) }
func (s *LocalStore) ConsumeOTP(id string) {
	_ = s.write(func() error { s.memory.ConsumeOTP(id); return nil })
}
func (s *LocalStore) Templates() []domain.Template                { return s.memory.Templates() }
func (s *LocalStore) Template(id string) (domain.Template, error) { return s.memory.Template(id) }
func (s *LocalStore) CreateCV(cv domain.CV) (domain.CV, error) {
	var result domain.CV
	err := s.write(func() error { var err error; result, err = s.memory.CreateCV(cv); return err })
	return result, err
}
func (s *LocalStore) CV(userID, id string) (domain.CV, error) { return s.memory.CV(userID, id) }
func (s *LocalStore) CVs(userID string) []domain.CV           { return s.memory.CVs(userID) }
func (s *LocalStore) PatchCV(userID, id string, content map[string]any, summary string) (domain.CV, error) {
	var result domain.CV
	err := s.write(func() error { var err error; result, err = s.memory.PatchCV(userID, id, content, summary); return err })
	return result, err
}
func (s *LocalStore) DuplicateCV(userID, id string) (domain.CV, error) {
	var result domain.CV
	err := s.write(func() error { var err error; result, err = s.memory.DuplicateCV(userID, id); return err })
	return result, err
}
func (s *LocalStore) ArchiveCV(userID, id string) error {
	return s.write(func() error { return s.memory.ArchiveCV(userID, id) })
}
func (s *LocalStore) Versions(userID, id string) ([]domain.CVVersion, error) {
	return s.memory.Versions(userID, id)
}
func (s *LocalStore) RestoreVersion(userID, cvID, versionID string) (domain.CV, error) {
	var result domain.CV
	err := s.write(func() error {
		var err error
		result, err = s.memory.RestoreVersion(userID, cvID, versionID)
		return err
	})
	return result, err
}
func (s *LocalStore) DailyExportsRemaining(userID string, now time.Time) int {
	user, err := s.memory.GetUser(userID)
	if err == nil && user.Plan == LocalDesktopPlan {
		return -1
	}
	return s.memory.DailyExportsRemaining(userID, now)
}
func (s *LocalStore) ConsumeExportCredit(userID, cvID, format string) (domain.Export, error) {
	var result domain.Export
	err := s.write(func() error {
		var err error
		result, err = s.memory.ConsumeExportCredit(userID, cvID, format)
		return err
	})
	return result, err
}
func (s *LocalStore) CompleteExport(userID, exportID, objectKey string, size int64) error {
	return s.write(func() error { return s.memory.CompleteExport(userID, exportID, objectKey, size) })
}
func (s *LocalStore) FailExport(userID, exportID string) error {
	return s.write(func() error { return s.memory.FailExport(userID, exportID) })
}
func (s *LocalStore) ExportForUser(userID, filename string) (domain.Export, error) {
	return s.memory.ExportForUser(userID, filename)
}
func (s *LocalStore) AdminAnalytics(ctx context.Context) (AdminAnalytics, error) {
	return s.memory.AdminAnalytics(ctx)
}

var _ Repository = (*LocalStore)(nil)
