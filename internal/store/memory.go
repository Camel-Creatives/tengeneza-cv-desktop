// Package store provides the repository contract. MemoryStore powers local and
// contract tests; production adapters must retain the same ownership semantics.
package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/auth"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

type MemoryStore struct {
	mu        sync.RWMutex
	users     map[string]domain.User
	byEmail   map[string]string
	byPhone   map[string]string
	otps      map[string]auth.OTP
	cvs       map[string]domain.CV
	versions  map[string][]domain.CVVersion
	exports   map[string][]domain.Export
	templates map[string]domain.Template
}

func (s *MemoryStore) AdminAnalytics(context.Context) (AdminAnalytics, error) {
	return AdminAnalytics{Templates: len(s.templates)}, nil
}

const DefaultUserID = "default-user"

func (s *MemoryStore) Ping(context.Context) error { return nil }

func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{users: map[string]domain.User{}, byEmail: map[string]string{}, byPhone: map[string]string{}, otps: map[string]auth.OTP{}, cvs: map[string]domain.CV{}, versions: map[string][]domain.CVVersion{}, exports: map[string][]domain.Export{}, templates: map[string]domain.Template{}}
	s.seedTemplates()
	s.seedDefaultUser()
	return s
}
func (s *MemoryStore) seedDefaultUser() {
	// Retained only for non-production demo compatibility. Production auth never selects this identity.
	u := domain.User{ID: DefaultUserID, FullName: "Guest User", Email: "guest@tengenezacv.online", PreferredLanguage: "en", TargetRegion: "eac", Plan: "free", ExportCredits: DailyExportLimit, AITokenBalance: 20000, Role: domain.RoleUser, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	s.users[DefaultUserID] = u
	s.byEmail[u.Email] = DefaultUserID
}
func (s *MemoryStore) seedTemplates() {
	for _, t := range []domain.Template{{ID: "ats_classic", Name: "ATS Classic", Family: "ats", IsFree: true, SupportsPhoto: false, Active: true, AccentColorHex: "#0F5132"}, {ID: "ats_modern", Name: "ATS Modern", Family: "ats", IsFree: true, SupportsPhoto: false, Active: true, AccentColorHex: "#1E3A5F"}, {ID: "global_minimal", Name: "Global Minimal", Family: "ats", IsFree: true, SupportsPhoto: false, Active: true, AccentColorHex: "#334155"}, {ID: "executive_ats", Name: "Executive ATS", Family: "ats", IsFree: true, SupportsPhoto: false, Active: true, AccentColorHex: "#4A3225"}, {ID: "visual_local", Name: "Visual Local", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#0F5132"}, {ID: "hospitality_local", Name: "Hospitality Local", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#0D5B57"}, {ID: "academic_cv", Name: "Academic CV", Family: "academic", IsFree: true, SupportsPhoto: false, Active: true, AccentColorHex: "#4A3B6A"}, {ID: "career_change", Name: "Career Change", Family: "ats", IsFree: true, SupportsPhoto: false, Active: true, AccentColorHex: "#A14D2F"}, {ID: "creative_portfolio", Name: "Creative Portfolio", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#C6337B"}, {ID: "tech_product", Name: "Tech Product", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#36D1A4"}, {ID: "finance_audit", Name: "Finance Audit", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#183B31"}, {ID: "legal_chambers", Name: "Legal Chambers", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#7C5675"}, {ID: "clinical_care", Name: "Clinical Care", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#267DAF"}, {ID: "educator_classroom", Name: "Educator Classroom", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#AF4D3D"}, {ID: "sales_growth", Name: "Sales Growth", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#EF632B"}, {ID: "field_operations", Name: "Field Operations", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#304A48"}, {ID: "consulting_strategy", Name: "Consulting Strategy", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#263A59"}, {ID: "engineering_blueprint", Name: "Engineering Blueprint", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#163C57"}, {ID: "public_service", Name: "Public Service", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#542737"}, {ID: "nonprofit_impact", Name: "Nonprofit Impact", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#17705B"}, {ID: "graduate_launch", Name: "Graduate Launch", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#E05C42"}, {ID: "skilled_trades", Name: "Skilled Trades", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#F6C34B"}, {ID: "remote_async", Name: "Remote Async", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#6456A7"}, {ID: "creative_producer", Name: "Creative Producer", Family: "visual", IsFree: true, SupportsPhoto: true, Active: true, AccentColorHex: "#FF8364"}} {
		s.templates[t.ID] = t
	}
}
func (s *MemoryStore) CreateUser(user domain.User) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if user.Email != "" {
		if _, ok := s.byEmail[user.Email]; ok {
			return domain.User{}, ErrConflict
		}
	}
	if user.Phone != "" {
		if _, ok := s.byPhone[user.Phone]; ok {
			return domain.User{}, ErrConflict
		}
	}
	now := time.Now().UTC()
	user.ID = uuid.NewString()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.Plan = "free"
	user.PreferredLanguage = defaultStr(user.PreferredLanguage, "en")
	user.TargetRegion = defaultStr(user.TargetRegion, "eac")
	user.Role = domain.RoleUser
	user.AITokenBalance = 20000
	user.ExportCredits = DailyExportLimit
	s.users[user.ID] = user
	if user.Email != "" {
		s.byEmail[user.Email] = user.ID
	}
	if user.Phone != "" {
		s.byPhone[user.Phone] = user.ID
	}
	return user, nil
}
func (s *MemoryStore) GetUser(id string) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return domain.User{}, ErrNotFound
	}
	return u, nil
}

// EnsureExternalUser provisions the product-side profile the first time a
// trusted CamelAccounts subject reaches the API. The subject remains the
// stable identity key; profile claims can be enriched later through /me.
func (s *MemoryStore) EnsureExternalUser(id string, claims ...domain.User) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if user, ok := s.users[id]; ok {
		if len(claims) > 0 {
			if claims[0].FullName != "" {
				user.FullName = claims[0].FullName
			}
			if claims[0].Email != "" {
				user.Email = claims[0].Email
				s.byEmail[user.Email] = id
			}
			if claims[0].Phone != "" {
				user.Phone = claims[0].Phone
				s.byPhone[user.Phone] = id
			}
			user.UpdatedAt = time.Now().UTC()
			s.users[id] = user
		}
		return user, nil
	}
	now := time.Now().UTC()
	user := domain.User{ID: id, ExternalSubject: id, FullName: "", PreferredLanguage: "en", TargetRegion: "eac", Plan: "free", ExportCredits: 5, AITokenBalance: 20000, Role: domain.RoleUser, CreatedAt: now, UpdatedAt: now}
	if len(claims) > 0 {
		user.FullName = claims[0].FullName
		user.Email = claims[0].Email
		user.Phone = claims[0].Phone
	}
	if user.Email != "" {
		s.byEmail[user.Email] = id
	}
	if user.Phone != "" {
		s.byPhone[user.Phone] = id
	}
	s.users[id] = user
	return user, nil
}
func (s *MemoryStore) FindUserByEmail(email string) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byEmail[email]
	if !ok {
		return domain.User{}, ErrNotFound
	}
	return s.users[id], nil
}
func (s *MemoryStore) UpdateUser(id string, mutate func(*domain.User)) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return domain.User{}, ErrNotFound
	}
	mutate(&u)
	u.UpdatedAt = time.Now().UTC()
	s.users[id] = u
	return u, nil
}
func (s *MemoryStore) SaveOTP(otp auth.OTP) { s.mu.Lock(); defer s.mu.Unlock(); s.otps[otp.ID] = otp }
func (s *MemoryStore) GetOTP(id string) (auth.OTP, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	otp, ok := s.otps[id]
	if !ok {
		return auth.OTP{}, ErrNotFound
	}
	return otp, nil
}
func (s *MemoryStore) ConsumeOTP(id string) { s.mu.Lock(); defer s.mu.Unlock(); delete(s.otps, id) }
func (s *MemoryStore) Templates() []domain.Template {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Template, 0, len(s.templates))
	for _, t := range s.templates {
		if t.Active {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (s *MemoryStore) Template(id string) (domain.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.templates[id]
	if !ok || !t.Active {
		return domain.Template{}, ErrNotFound
	}
	return t, nil
}
func (s *MemoryStore) CreateCV(cv domain.CV) (domain.CV, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.templates[cv.TemplateID]; !ok {
		return domain.CV{}, ErrNotFound
	}
	if _, ok := s.users[cv.UserID]; !ok {
		return domain.CV{}, ErrNotFound
	}
	cv.ID = uuid.NewString()
	now := time.Now().UTC()
	cv.CreatedAt = now
	cv.UpdatedAt = now
	cv.Content = cloneContent(cv.Content)
	s.cvs[cv.ID] = cv
	s.addVersionLocked(cv, "user", "Created CV")
	return cv, nil
}
func (s *MemoryStore) CV(userID, id string) (domain.CV, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cv, ok := s.cvs[id]
	if !ok || cv.UserID != userID || cv.Archived {
		return domain.CV{}, ErrNotFound
	}
	cv.Content = cloneContent(cv.Content)
	return cv, nil
}
func (s *MemoryStore) CVs(userID string) []domain.CV {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.CV{}
	for _, cv := range s.cvs {
		if cv.UserID == userID && !cv.Archived {
			cv.Content = cloneContent(cv.Content)
			out = append(out, cv)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}
func (s *MemoryStore) PatchCV(userID, id string, content map[string]any, summary string) (domain.CV, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cv, ok := s.cvs[id]
	if !ok || cv.UserID != userID || cv.Archived {
		return domain.CV{}, ErrNotFound
	}
	for key, value := range content {
		cv.Content[key] = value
	}
	cv.UpdatedAt = time.Now().UTC()
	s.cvs[id] = cv
	s.addVersionLocked(cv, "user", summary)
	cv.Content = cloneContent(cv.Content)
	return cv, nil
}
func (s *MemoryStore) DuplicateCV(userID, id string) (domain.CV, error) {
	cv, err := s.CV(userID, id)
	if err != nil {
		return domain.CV{}, err
	}
	cv.ID = ""
	cv.Title = cv.Title + " (copy)"
	return s.CreateCV(cv)
}
func (s *MemoryStore) ArchiveCV(userID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cv, ok := s.cvs[id]
	if !ok || cv.UserID != userID {
		return ErrNotFound
	}
	cv.Archived = true
	cv.UpdatedAt = time.Now().UTC()
	s.cvs[id] = cv
	return nil
}
func (s *MemoryStore) Versions(userID, id string) ([]domain.CVVersion, error) {
	if _, err := s.CV(userID, id); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]domain.CVVersion(nil), s.versions[id]...)
	return out, nil
}
func (s *MemoryStore) RestoreVersion(userID, cvID, versionID string) (domain.CV, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cv, ok := s.cvs[cvID]
	if !ok || cv.UserID != userID || cv.Archived {
		return domain.CV{}, ErrNotFound
	}
	for _, v := range s.versions[cvID] {
		if v.ID == versionID {
			cv.Content = cloneContent(v.Content)
			cv.UpdatedAt = time.Now().UTC()
			s.cvs[cvID] = cv
			s.addVersionLocked(cv, "user", "Restored version")
			return cv, nil
		}
	}
	return domain.CV{}, ErrNotFound
}

var ErrDailyExportLimit = errors.New("daily export limit reached")

const DailyExportLimit = 5

func (s *MemoryStore) DailyExportsRemaining(userID string, now time.Time) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return dailyExportsRemaining(s.exports, userID, now)
}

func dailyExportsRemaining(exports map[string][]domain.Export, userID string, now time.Time) int {
	location, err := time.LoadLocation("Africa/Dar_es_Salaam")
	if err != nil {
		location = time.FixedZone("EAT", 3*60*60)
	}
	start := time.Date(now.In(location).Year(), now.In(location).Month(), now.In(location).Day(), 0, 0, 0, 0, location).UTC()
	count := 0
	for _, entries := range exports {
		for _, entry := range entries {
			if entry.UserID == userID && !entry.Cached && !entry.CreatedAt.Before(start) {
				count++
			}
		}
	}
	if count >= DailyExportLimit {
		return 0
	}
	return DailyExportLimit - count
}

func (s *MemoryStore) ConsumeExportCredit(userID, cvID, format string) (domain.Export, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cv, ok := s.cvs[cvID]
	if !ok || cv.UserID != userID || cv.Archived {
		return domain.Export{}, ErrNotFound
	}
	hash := contentHash(cv.Content)
	for i := len(s.exports[cvID]) - 1; i >= 0; i-- {
		e := s.exports[cvID][i]
		if e.ContentHash == hash && e.Format == format && e.CreatedAt.After(time.Now().UTC().Add(-30*24*time.Hour)) {
			e.Cached = true
			s.exports[cvID][i] = e
			return e, nil
		}
	}
	user := s.users[userID]
	if user.Plan != LocalDesktopPlan && dailyExportsRemaining(s.exports, userID, time.Now().UTC()) == 0 {
		return domain.Export{}, ErrDailyExportLimit
	}

	versions := s.versions[cvID]
	versionID := ""
	if len(versions) > 0 {
		versionID = versions[len(versions)-1].ID
	}
	e := domain.Export{
		ID:             uuid.NewString(),
		CVID:           cvID,
		VersionID:      versionID,
		Format:         format,
		ContentHash:    hash,
		FileURL:        "https://files.tengenezacv.online/exports/" + cvID + "/" + hash + "." + format,
		CreditConsumed: false,
		UserID:         userID,
		CreatedAt:      time.Now().UTC(),
	}
	s.exports[cvID] = append(s.exports[cvID], e)
	return e, nil
}
func (s *MemoryStore) addVersionLocked(cv domain.CV, changedBy, summary string) {
	s.versions[cv.ID] = append(s.versions[cv.ID], domain.CVVersion{ID: uuid.NewString(), CVID: cv.ID, Content: cloneContent(cv.Content), ChangedBy: changedBy, ChangeSummary: summary, CreatedAt: time.Now().UTC()})
}
func cloneContent(content map[string]any) map[string]any {
	if content == nil {
		return map[string]any{}
	}
	raw, _ := json.Marshal(content)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out
}
func contentHash(content map[string]any) string {
	raw, _ := json.Marshal(content)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func defaultStr(v, f string) string {
	if v == "" {
		return f
	}
	return v
}

func (s *MemoryStore) CompleteExport(userID, exportID, objectKey string, size int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for cvID, entries := range s.exports {
		for i, entry := range entries {
			if entry.ID == exportID && entry.UserID == userID {
				entry.FileURL = objectKey
				entry.ObjectKey = objectKey
				entry.ByteSize = size
				s.exports[cvID][i] = entry
				return nil
			}
		}
	}
	return ErrNotFound
}
func (s *MemoryStore) ExportForUser(userID, filename string) (domain.Export, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, entries := range s.exports {
		for _, entry := range entries {
			if entry.UserID == userID && (entry.FileURL == filename || filepath.Base(entry.ObjectKey) == filename || filepath.Base(entry.FileURL) == filename) {
				return entry, nil
			}
		}
	}
	return domain.Export{}, ErrNotFound
}

func (s *MemoryStore) FailExport(userID, exportID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for cvID, entries := range s.exports {
		for i, entry := range entries {
			if entry.ID == exportID && entry.UserID == userID {
				s.exports[cvID] = append(entries[:i], entries[i+1:]...)
				return nil
			}
		}
	}
	return ErrNotFound
}
