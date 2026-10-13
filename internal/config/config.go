// Package config owns environment loading and production fail-fast validation.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment             string
	Port                    string
	PublicAPIBaseURL        string
	AllowedOrigins          []string
	DatabaseURL             string
	RedisURL                string
	JWTSigningKey           string
	JWTIssuer               string
	AdminLoginEmail         string
	AdminLoginPassword      string
	AdminTOTPSecretFile     string
	CamelAccountsIssuer     string
	CamelAccountsJWKSURL    string
	CamelAccountsAudience   string
	NgamiaAIBaseURL         string
	NgamiaAIAPIKey          string
	MailAfricaBaseURL       string
	MailAfricaAPIKey        string
	MailAfricaWebhookSecret string
	MailAfricaDefaultFrom   string
	MailAfricaFromDomainID  int64
	S3Endpoint              string
	S3Bucket                string
	S3AccessKey             string
	S3SecretKey             string
	S3Secure                bool
	LocalExportDir          string
	LocalUploadDir          string
	ExamplesDir             string
	ChromiumExecutable      string
	LogLevel                slog.Level
}

func Load() (Config, error) {
	_ = godotenv.Load()
	cfg := Config{
		Environment:             envOr("APP_ENV", "development"),
		Port:                    envOr("PORT", "8080"),
		PublicAPIBaseURL:        envOr("PUBLIC_API_BASE_URL", "https://api.tengenezacv.online"),
		AllowedOrigins:          splitCSV(envOr("ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000,https://tengenezacv.online,https://www.tengenezacv.online")),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		RedisURL:                os.Getenv("REDIS_URL"),
		JWTSigningKey:           os.Getenv("JWT_SIGNING_KEY"),
		JWTIssuer:               envOr("JWT_ISSUER", "tengenezacv-api"),
		AdminLoginEmail:         strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_LOGIN_EMAIL"))),
		AdminLoginPassword:      os.Getenv("ADMIN_LOGIN_PASSWORD"),
		AdminTOTPSecretFile:     envOr("ADMIN_TOTP_SECRET_FILE", "/var/lib/tengenezacv/admin-totp.secret"),
		CamelAccountsIssuer:     envOr("CAMEL_ACCOUNTS_ISSUER", "https://accounts.camelcreatives.com"),
		CamelAccountsJWKSURL:    envOr("CAMEL_ACCOUNTS_JWKS_URL", "https://accounts.camelcreatives.com/oauth/jwks"),
		CamelAccountsAudience:   envOr("CAMEL_ACCOUNTS_AUDIENCE", "tengenezacv-web"),
		NgamiaAIBaseURL:         envOr("NGAMIA_AI_BASE_URL", "https://api.ngamia.cc/v1"),
		NgamiaAIAPIKey:          os.Getenv("NGAMIA_AI_API_KEY"),
		MailAfricaBaseURL:       envOr("MAILAFRICA_BASE_URL", "https://api.mailafrica.online"),
		MailAfricaAPIKey:        os.Getenv("MAILAFRICA_API_KEY"),
		MailAfricaWebhookSecret: os.Getenv("MAILAFRICA_WEBHOOK_SECRET"),
		MailAfricaDefaultFrom:   envOr("MAILAFRICA_DEFAULT_FROM", "noreply@mail.tengenezacv.online"),
		MailAfricaFromDomainID:  parseInt64(os.Getenv("MAILAFRICA_FROM_DOMAIN_ID")),
		S3Endpoint:              os.Getenv("S3_ENDPOINT"),
		S3Bucket:                os.Getenv("S3_BUCKET"),
		S3AccessKey:             os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:             os.Getenv("S3_SECRET_KEY"),
		S3Secure:                envOr("S3_SECURE", "false") == "true",
		LocalExportDir:          envOr("LOCAL_EXPORT_DIR", "/tmp/tengenezacv-exports"),
		LocalUploadDir:          envOr("LOCAL_UPLOAD_DIR", "/tmp/tengenezacv-uploads"),
		ExamplesDir:             envOr("EXAMPLES_DIR", "examples/sample-cvs"),
		ChromiumExecutable:      envOr("CHROMIUM_EXECUTABLE", "/usr/bin/chromium"),
		LogLevel:                parseLogLevel(envOr("LOG_LEVEL", "info")),
	}
	if (cfg.AdminLoginEmail == "") != (cfg.AdminLoginPassword == "") {
		return Config{}, fmt.Errorf("ADMIN_LOGIN_EMAIL and ADMIN_LOGIN_PASSWORD must be configured together")
	}
	if cfg.AdminLoginPassword != "" && len(cfg.AdminLoginPassword) < 12 {
		return Config{}, fmt.Errorf("ADMIN_LOGIN_PASSWORD must contain at least 12 characters")
	}
	if cfg.Environment == "production" {
		if err := cfg.ValidateProduction(); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}

func (c Config) ValidateProduction() error {
	required := map[string]string{
		"DATABASE_URL": c.DatabaseURL, "REDIS_URL": c.RedisURL, "JWT_SIGNING_KEY": c.JWTSigningKey, "CAMEL_ACCOUNTS_ISSUER": c.CamelAccountsIssuer, "CAMEL_ACCOUNTS_AUDIENCE": c.CamelAccountsAudience, "CAMEL_ACCOUNTS_JWKS_URL": c.CamelAccountsJWKSURL,
		"NGAMIA_AI_API_KEY": c.NgamiaAIAPIKey, "MAILAFRICA_API_KEY": c.MailAfricaAPIKey,
		"MAILAFRICA_WEBHOOK_SECRET": c.MailAfricaWebhookSecret,
		"S3_ENDPOINT":               c.S3Endpoint, "S3_BUCKET": c.S3Bucket,
		"S3_ACCESS_KEY": c.S3AccessKey, "S3_SECRET_KEY": c.S3SecretKey,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("missing required production environment variable %s", name)
		}
	}
	if len(c.JWTSigningKey) < 32 {
		return fmt.Errorf("JWT_SIGNING_KEY must contain at least 32 characters in production")
	}
	return nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func splitCSV(value string) []string {
	raw := strings.Split(value, ",")
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
func parseLogLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
func parseInt64(value string) int64 {
	number, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return number
}
