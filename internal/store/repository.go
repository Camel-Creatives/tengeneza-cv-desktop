package store

import (
	"context"
	"time"

	"github.com/troubleman96/tengeneza-cv-desktop/internal/auth"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
)

type Repository interface {
	Ping(context.Context) error
	GetUser(string) (domain.User, error)
	EnsureExternalUser(string, ...domain.User) (domain.User, error)
	UpdateUser(string, func(*domain.User)) (domain.User, error)
	SaveOTP(auth.OTP)
	GetOTP(string) (auth.OTP, error)
	ConsumeOTP(string)
	Templates() []domain.Template
	Template(string) (domain.Template, error)
	CreateCV(domain.CV) (domain.CV, error)
	CV(string, string) (domain.CV, error)
	CVs(string) []domain.CV
	PatchCV(string, string, map[string]any, string) (domain.CV, error)
	DuplicateCV(string, string) (domain.CV, error)
	ArchiveCV(string, string) error
	Versions(string, string) ([]domain.CVVersion, error)
	RestoreVersion(string, string, string) (domain.CV, error)
	DailyExportsRemaining(string, time.Time) int
	ConsumeExportCredit(string, string, string) (domain.Export, error)
	CompleteExport(string, string, string, int64) error
	FailExport(string, string) error
	ExportForUser(string, string) (domain.Export, error)
	AdminAnalytics(context.Context) (AdminAnalytics, error)
}

type AdminAnalytics struct {
	Users, ActiveCVs, Templates, NewUsers30d, NewCVs30d, ActiveUsers7d int
	Daily                                                              []AnalyticsDay
	Regions                                                            []RegionMetric
	TemplateUsage                                                      []TemplateMetric
}
type AnalyticsDay struct {
	Date       string
	Users, CVs int
}
type RegionMetric struct {
	Region string
	Users  int
}
type TemplateMetric struct {
	ID, Name, Family string
	Usage            int
}

var _ Repository = (*MemoryStore)(nil)
