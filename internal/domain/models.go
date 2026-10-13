// Package domain contains provider-agnostic business models.
package domain

import "time"

type Role string

const (
	RoleUser        Role = "user"
	RoleAgencyAdmin Role = "agency_admin"
	RoleSupport     Role = "support"
	RoleSuperAdmin  Role = "superadmin"
)

type User struct {
	ID, FullName                          string
	Phone, Email, PasswordHash            string
	PreferredLanguage, TargetRegion, Plan string
	ExportCredits                         int
	AITokenBalance                        int64
	VoiceSecondsBalance                   int
	ProfilePhotoURL                       string
	ExternalSubject                       string
	Role                                  Role
	PhoneVerified                         bool
	CreatedAt, UpdatedAt                  time.Time
}

type Template struct {
	ID, Name, Family              string
	IsFree, SupportsPhoto, Active bool
	AccentColorHex                string
}

type CV struct {
	ID, UserID, Title, DocumentStyle, TemplateID, TargetRegion, Language string
	Content                                                              map[string]any
	ATSScore                                                             int
	Archived                                                             bool
	CreatedAt, UpdatedAt                                                 time.Time
}

type CVVersion struct {
	ID, CVID                 string
	Content                  map[string]any
	ChangedBy, ChangeSummary string
	CreatedAt                time.Time
}
type Export struct {
	ID, CVID, VersionID, Format, ContentHash, FileURL string
	ObjectKey                                         string
	ByteSize                                          int64
	CreditConsumed                                    bool
	UserID                                            string
	Cached                                            bool
	CreatedAt                                         time.Time
}

type LintResult struct{ Check, Severity, Message string }
type ATSReport struct {
	Score     int          `json:"score"`
	Breakdown []LintResult `json:"breakdown"`
	CoachNote string       `json:"coach_note"`
}
