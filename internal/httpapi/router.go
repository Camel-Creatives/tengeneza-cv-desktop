// Package httpapi configures Fiber, common security middleware, and platform endpoints.
package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/auth"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/blob"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/clients/mailafrica"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/clients/ngamiaai"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/config"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/httpapi/handlers"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/httpapi/middleware"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/store"
)

func New(cfg config.Config, logger *slog.Logger) *fiber.App {
	app, err := NewWithError(cfg, logger)
	if err == nil {
		return app
	}
	logger.Error("API bootstrap failed", "error", err)
	app = fiber.New(fiber.Config{AppName: "tengenezacv-api"})
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "unavailable"})
	})
	return app
}

func NewWithError(cfg config.Config, logger *slog.Logger) (*fiber.App, error) {
	return NewWithRepository(cfg, logger, nil)
}

// NewWithRepository allows the Wails host to inject its durable local store
// while the hosted API continues to select PostgreSQL or memory storage.
func NewWithRepository(cfg config.Config, logger *slog.Logger, localRepository store.Repository) (*fiber.App, error) {
	app := fiber.New(fiber.Config{AppName: "tengenezacv-api", ErrorHandler: errorHandler, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second})
	app.Use(recover.New())
	app.Use(func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)
		return c.Next()
	})
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "no-referrer")
		return c.Next()
	})
	app.Use(cors.New(cors.Config{AllowOrigins: strings.Join(cfg.AllowedOrigins, ","), AllowHeaders: "Authorization, Content-Type, Idempotency-Key, X-Request-ID", AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS", MaxAge: 300}))

	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		options, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("parse redis URL: %w", err)
		}
		redisClient = redis.NewClient(options)
		if err = redisClient.Ping(context.Background()).Err(); err != nil {
			_ = redisClient.Close()
			return nil, fmt.Errorf("connect redis: %w", err)
		}
	} else if cfg.Environment == "production" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}
	if cfg.Environment != "desktop" {
		if redisClient != nil {
			app.Use(middleware.NewRedisRateLimiter(redisClient, 120, time.Minute).Handle)
		} else {
			app.Use(middleware.NewRateLimiter(120, time.Minute).Handle)
		}
	}

	app.Get("/healthz", func(c *fiber.Ctx) error { return OK(c, fiber.Map{"status": "ok", "service": "tengenezacv-api"}) })
	app.Get("/metrics", func(c *fiber.Ctx) error {
		c.Type("text/plain")
		return c.SendString("# tengenezacv metrics will be registered in the production platform package\n")
	})
	app.Get("/docs", func(c *fiber.Ctx) error { c.Type("html"); return c.SendString(swaggerHTML) })
	app.Get("/docs/openapi.yaml", func(c *fiber.Ctx) error { c.Type("yaml"); return c.SendString(openAPISpec) })
	var repository store.Repository
	var closeRepository func()
	if localRepository != nil {
		repository = localRepository
	} else if cfg.DatabaseURL != "" {
		postgres, err := store.OpenPostgres(context.Background(), cfg.DatabaseURL)
		if err != nil {
			if redisClient != nil {
				_ = redisClient.Close()
			}
			return nil, fmt.Errorf("connect postgres: %w", err)
		}
		repository, closeRepository = postgres, postgres.Close
	} else if cfg.Environment == "production" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	} else {
		repository = store.NewMemoryStore()
	}
	_ = closeRepository
	var blobStore blob.Store
	if cfg.S3Endpoint != "" {
		remote, err := blob.NewS3(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3Secure)
		if err == nil {
			err = remote.Ready(context.Background(), cfg.Environment != "production")
		}
		if err != nil {
			if closeRepository != nil {
				closeRepository()
			}
			if redisClient != nil {
				_ = redisClient.Close()
			}
			return nil, fmt.Errorf("configure object storage: %w", err)
		}
		blobStore = remote
	} else if cfg.Environment == "production" {
		if closeRepository != nil {
			closeRepository()
		}
		if redisClient != nil {
			_ = redisClient.Close()
		}
		return nil, fmt.Errorf("S3_ENDPOINT is required")
	} else {
		blobStore = blob.FileStore{Root: cfg.LocalExportDir}
	}

	if cfg.Environment != "production" {
		app.Static("/uploads", cfg.LocalUploadDir)
	}
	app.Get("/uploads/profiles/:file", func(c *fiber.Ctx) error {
		filename := filepath.Base(c.Params("file"))
		if blobStore != nil {
			if raw, err := blobStore.Get(c.UserContext(), "profiles/"+filename); err == nil {
				c.Set("Cache-Control", "public, max-age=31536000, immutable")
				c.Set("Content-Type", mime.TypeByExtension(filepath.Ext(filename)))
				return c.Send(raw)
			} else if cfg.Environment == "production" {
				return Fail(c, fiber.StatusNotFound, "NOT_FOUND", "Profile photo not found.", nil)
			}
		}
		candidates := []string{
			filepath.Join(cfg.LocalUploadDir, "profiles", filename),
			filepath.Join(cfg.LocalUploadDir, filename),
			filepath.Join(cfg.ExamplesDir, "profile-assets", "library", filename),
			filepath.Join(cfg.ExamplesDir, "profile-assets", filename),
			filepath.Join(cfg.ExamplesDir, "..", "profile-assets", "library", filename),
			filepath.Join("API", "examples", "profile-assets", "library", filename),
			filepath.Join("UI", "public", "uploads", "profiles", filename),
			filepath.Join("/home/cameltech/Projects/tengeneza-cv/UI/public/uploads/profiles", filename),
		}
		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				return c.SendFile(p)
			}
		}
		return Fail(c, fiber.StatusNotFound, "NOT_FOUND", "Profile photo not found.", nil)
	})
	app.Get("/readyz", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		checks := fiber.Map{"database": "ready", "redis": "ready", "storage": "ready"}
		if repository.Ping(ctx) != nil {
			checks["database"] = "unavailable"
		}
		if redisClient != nil && redisClient.Ping(ctx).Err() != nil {
			checks["redis"] = "unavailable"
		}
		if blobStore.Ping(ctx) != nil {
			checks["storage"] = "unavailable"
		}
		for _, status := range checks {
			if status == "unavailable" {
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"success": false, "error": fiber.Map{"code": "NOT_READY", "message": "A required dependency is unavailable."}, "data": fiber.Map{"checks": checks}})
			}
		}
		return OK(c, fiber.Map{"status": "ready", "checks": checks})
	})
	productAuth := auth.New(cfg.JWTSigningKey, cfg.JWTIssuer)
	adminPasswordHash := ""
	renderDir := cfg.LocalExportDir
	if renderDir != "" {
		renderDir = filepath.Join(renderDir, ".render")
	}
	if cfg.AdminLoginPassword != "" {
		adminPasswordHash, _ = auth.HashPassword(cfg.AdminLoginPassword)
	}
	core := &handlers.Core{Store: repository, Blobs: blobStore, Auth: productAuth, CamelAuth: auth.NewCamelVerifier(cfg.CamelAccountsIssuer, cfg.CamelAccountsAudience, cfg.CamelAccountsJWKSURL), IsProduction: cfg.Environment == "production", MailAfrica: mailafrica.New(cfg.MailAfricaBaseURL, cfg.MailAfricaAPIKey, cfg.MailAfricaWebhookSecret, cfg.MailAfricaDefaultFrom, cfg.MailAfricaFromDomainID), NgamiaAI: ngamiaai.New(cfg.NgamiaAIBaseURL, cfg.NgamiaAIAPIKey), Renderer: render.New(renderDir, cfg.ChromiumExecutable, cfg.LocalUploadDir, cfg.ExamplesDir), LocalUploadDir: cfg.LocalUploadDir, PublicAPIBaseURL: cfg.PublicAPIBaseURL, ExamplesDir: cfg.ExamplesDir, FreePlatform: cfg.Environment != "production", AdminLoginEmail: cfg.AdminLoginEmail, AdminPasswordHash: adminPasswordHash, AdminTOTPSecretFile: cfg.AdminTOTPSecretFile, AdminChallenges: handlers.NewAdminChallengeStore()}
	app.Get("/templates", core.Templates)
	app.Get("/examples", core.Examples)
	app.Get("/examples/:slug", core.Example)
	authRoutes := app.Group("/auth")
	authRoutes.Post("/signup", core.SignUp)
	authRoutes.Post("/otp/verify", core.VerifyOTP)
	authRoutes.Post("/otp/resend", core.ResendOTP)
	authRoutes.Post("/login/password", core.LoginPassword)
	adminLoginLimiter := middleware.NewRateLimiter(10, time.Minute)
	if redisClient != nil {
		adminLoginLimiter = middleware.NewRedisRateLimiter(redisClient, 10, time.Minute)
	}
	authRoutes.Post("/admin/login", adminLoginLimiter.Handle, core.AdminPasswordLogin)
	authRoutes.Post("/admin/verify", adminLoginLimiter.Handle, core.AdminTOTPVerify)
	app.Get("/billing/plans", core.BillingPlans)
	app.Post("/webhooks/mailafrica", core.MailAfricaWebhook)
	app.Get("/exports/:file", core.RequireAuth, func(c *fiber.Ctx) error {
		filename := filepath.Base(c.Params("file"))
		uid, _ := c.Locals("user_id").(string)
		// Export filenames end with _<owner-id>.<format>; the owner id is the
		// authorization boundary, so accept the actual suffix instead of requiring
		// an extra underscore after the UUID.
		ownerSuffix := "_" + uid + "."
		if uid == "" || !strings.Contains(filename, ownerSuffix) {
			slog.Warn("export download rejected", "filename", filename, "user_id", uid)
			return Fail(c, fiber.StatusNotFound, "NOT_FOUND", "Export file not found.", nil)
		}
		export, err := repository.ExportForUser(uid, filename)
		if err != nil || export.ObjectKey == "" {
			slog.Warn("export record not found", "filename", filename, "user_id", uid, "error", err)
			return Fail(c, fiber.StatusNotFound, "NOT_FOUND", "Export file not found.", nil)
		}
		raw, err := blobStore.Get(c.Context(), export.ObjectKey)
		if err != nil {
			slog.Error("export artifact storage read failed", "filename", filename, "object_key", export.ObjectKey, "user_id", uid, "error", err)
			return Fail(c, fiber.StatusNotFound, "NOT_FOUND", "Export file not found.", nil)
		}
		c.Set("Content-Type", exportContentType(export.Format))
		c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		return c.Send(raw)
	})
	api := app.Group("", core.RequireAuth)
	api.Get("/admin/analytics", core.AdminAnalytics)
	api.Get("/me", core.Me)
	api.Patch("/me", core.UpdateMe)
	api.Post("/me/profile-photo", core.UploadProfilePhoto)
	api.Get("/me/wallet", core.Wallet)
	api.Post("/cvs", core.CreateCV)
	api.Get("/cvs", core.ListCVs)
	api.Get("/cvs/:id", core.GetCV)
	api.Patch("/cvs/:id", core.PatchCV)
	api.Delete("/cvs/:id", core.DeleteCV)
	api.Post("/cvs/:id/duplicate", core.DuplicateCV)
	api.Get("/cvs/:id/versions", core.Versions)
	api.Post("/cvs/:id/versions/:versionID/restore", core.RestoreVersion)
	api.Post("/cvs/:id/use-profile-photo", core.ApplyProfilePhoto)
	api.Post("/cvs/:id/agent/ats-score", core.ATSScore)
	api.Post("/cvs/:id/export", core.ExportCV)
	api.Post("/cvs/:id/export/email", core.EmailExport)
	api.Post("/cvs/:id/agent/autofill-section", core.AutofillSection)
	app.Use(func(c *fiber.Ctx) error {
		return Fail(c, fiber.StatusNotFound, "NOT_FOUND", "The requested resource was not found.", nil)
	})
	logger.Info("HTTP router configured", "allowed_origins", cfg.AllowedOrigins)
	return app, nil
}

func errorHandler(c *fiber.Ctx, err error) error {
	if fiberErr, ok := err.(*fiber.Error); ok {
		return Fail(c, fiberErr.Code, "VALIDATION_ERROR", fiberErr.Message, nil)
	}
	return Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong on our end. Try again in a moment.", nil)
}

func exportContentType(format string) string {
	if base, _, ok := strings.Cut(format, "_"); ok {
		format = base
	}
	switch format {
	case "pdf":
		return "application/pdf"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "csv":
		return "text/csv; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}
