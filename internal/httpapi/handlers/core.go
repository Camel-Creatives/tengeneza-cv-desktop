// Package handlers provides HTTP boundaries around the CV product services.
package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	httpapi "github.com/troubleman96/tengeneza-cv-desktop/internal/apiresponse"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/auth"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/blob"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/clients/mailafrica"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/clients/ngamiaai"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/profile"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render/lint"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/store"
)

type Core struct {
	Store               store.Repository
	Blobs               blob.Store
	Auth                *auth.Manager
	CamelAuth           *auth.CamelVerifier
	IsProduction        bool
	MailAfrica          *mailafrica.Client
	NgamiaAI            *ngamiaai.Client
	Renderer            *render.Service
	LocalUploadDir      string
	PublicAPIBaseURL    string
	ExamplesDir         string
	FreePlatform        bool
	AdminLoginEmail     string
	AdminPasswordHash   string
	AdminTOTPSecretFile string
	AdminChallenges     *AdminChallengeStore
}

type AdminChallenge struct {
	Secret     string
	Enrollment bool
	ExpiresAt  time.Time
}
type AdminChallengeStore struct {
	sync.Mutex
	Items map[string]AdminChallenge
}

func newAdminChallengeStore() *AdminChallengeStore {
	return &AdminChallengeStore{Items: map[string]AdminChallenge{}}
}
func NewAdminChallengeStore() *AdminChallengeStore { return newAdminChallengeStore() }
func (s *AdminChallengeStore) put(v AdminChallenge) string {
	raw := make([]byte, 18)
	_, _ = rand.Read(raw)
	id := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)
	s.Lock()
	s.Items[id] = v
	s.Unlock()
	return id
}
func (s *AdminChallengeStore) take(id string) (AdminChallenge, bool) {
	s.Lock()
	defer s.Unlock()
	v, ok := s.Items[id]
	if ok {
		delete(s.Items, id)
	}
	return v, ok
}

func adminSecret(path string) (string, bool, error) {
	raw, err := os.ReadFile(path)
	if err == nil {
		return strings.TrimSpace(string(raw)), true, nil
	}
	if !os.IsNotExist(err) {
		return "", false, err
	}
	bytes := make([]byte, 20)
	if _, err = rand.Read(bytes); err != nil {
		return "", false, err
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", false, err
	}
	if err = os.WriteFile(path, []byte(secret+"\n"), 0600); err != nil {
		return "", false, err
	}
	return secret, false, nil
}

func adminOTPURI(email, secret string) string {
	return fmt.Sprintf("otpauth://totp/Tengeneza%%20CV:%s?secret=%s&issuer=Tengeneza%%20CV&algorithm=SHA1&digits=6&period=30", url.QueryEscape(email), secret)
}

func (h *Core) RequireAuth(c *fiber.Ctx) error {
	header := strings.TrimSpace(c.Get("Authorization"))
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token != "" && h.Auth != nil {
		claims, err := h.Auth.ParseAccessToken(token)
		if err == nil && claims.Role == string(domain.RoleSuperAdmin) && strings.HasPrefix(claims.UserID, "admin:") {
			c.Locals("user_id", claims.UserID)
			c.Locals("role", claims.Role)
			c.Locals("admin_email", strings.TrimPrefix(claims.UserID, "admin:"))
			return c.Next()
		}
	}
	if h.CamelAuth != nil && token != "" {
		claims, err := h.CamelAuth.Verify(token)
		if err == nil && claims.Subject != "" {
			name := strings.TrimSpace(claims.FullName)
			if name == "" {
				name = strings.TrimSpace(claims.Name)
			}
			if name == "" {
				name = strings.TrimSpace(claims.GivenName + " " + claims.FamilyName)
			}
			phone := claims.Phone
			if phone == "" {
				phone = claims.PhoneAlt
			}
			if phone == "" {
				phone = claims.PhoneNumber
			}
			user, provisionErr := h.Store.EnsureExternalUser(claims.Subject, domain.User{ExternalSubject: claims.Subject, FullName: name, Email: claims.Email, Phone: phone})
			if provisionErr != nil {
				return httpapi.Fail(c, fiber.StatusInternalServerError, "USER_PROVISIONING_FAILED", "Your account could not be prepared. Please try again.", nil)
			}
			c.Locals("user_id", user.ID)
			c.Locals("role", claims.Role)
			return c.Next()
		}
		if h.IsProduction {
			return httpapi.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Your CamelAccounts session is invalid or expired.", nil)
		}
	}
	if h.IsProduction || token != "" {
		return httpapi.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Sign in with CamelAccounts to continue.", nil)
	}
	c.Locals("user_id", store.DefaultUserID)
	c.Locals("role", "user")
	return c.Next()
}
func userID(c *fiber.Ctx) string {
	if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
		return uid
	}
	return store.DefaultUserID
}

func (h *Core) SignUp(c *fiber.Ctx) error {
	return httpapi.Fail(c, fiber.StatusUnauthorized, "CAMELACCOUNTS_REQUIRED", "Create your account with CamelAccounts to continue.", nil)
}
func (h *Core) VerifyOTP(c *fiber.Ctx) error {
	return httpapi.Fail(c, fiber.StatusUnauthorized, "CAMELACCOUNTS_REQUIRED", "Verify your account with CamelAccounts to continue.", nil)
}
func (h *Core) LoginPassword(c *fiber.Ctx) error {
	return httpapi.Fail(c, fiber.StatusUnauthorized, "CAMELACCOUNTS_REQUIRED", "Sign in with CamelAccounts to continue.", nil)
}
func (h *Core) Me(c *fiber.Ctx) error {
	if role, _ := c.Locals("role").(string); role == string(domain.RoleSuperAdmin) {
		email, _ := c.Locals("admin_email").(string)
		return httpapi.OK(c, fiber.Map{"id": userID(c), "full_name": "Administrator", "email": email, "preferred_language": "en", "target_region": "eac", "plan": "admin", "role": role})
	}
	u, err := h.Store.GetUser(userID(c))
	if err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "User profile not found.", nil)
	}
	return httpapi.OK(c, publicUser(u))
}
func (h *Core) UpdateMe(c *fiber.Ctx) error {
	var body struct {
		FullName          string `json:"full_name"`
		Email             string `json:"email"`
		Phone             string `json:"phone"`
		PreferredLanguage string `json:"preferred_language"`
		TargetRegion      string `json:"target_region"`
	}
	if err := c.BodyParser(&body); err != nil {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", "Provide valid profile fields.", nil)
	}
	u, err := h.Store.UpdateUser(userID(c), func(u *domain.User) {
		if body.FullName != "" {
			u.FullName = body.FullName
		}
		if body.Email != "" {
			u.Email = body.Email
		}
		if body.Phone != "" {
			u.Phone = body.Phone
		}
		if body.PreferredLanguage == "en" || body.PreferredLanguage == "sw" {
			u.PreferredLanguage = body.PreferredLanguage
		}
		if body.TargetRegion != "" {
			u.TargetRegion = body.TargetRegion
		}
	})
	if err != nil {
		return err
	}
	return httpapi.OK(c, publicUser(u))
}
func (h *Core) UploadProfilePhoto(c *fiber.Ctx) error {
	file, err := c.FormFile("photo")
	if err != nil {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", "Attach a profile image in the photo field.", nil)
	}
	raw, extension, mimeType, err := profile.ReadUpload(file)
	if err != nil {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", err.Error(), nil)
	}
	name := uuid.NewString() + extension
	url := "/uploads/profiles/" + name
	if h.Blobs != nil {
		if err = h.Blobs.Put(c.UserContext(), "profiles/"+name, raw, mimeType); err != nil {
			return httpapi.Fail(c, 503, "STORAGE_UNAVAILABLE", "The profile image could not be stored. Please retry.", nil)
		}
	} else {
		if err = os.MkdirAll(filepath.Join(h.LocalUploadDir, "profiles"), 0750); err == nil {
			err = os.WriteFile(filepath.Join(h.LocalUploadDir, "profiles", name), raw, 0640)
		}
		if err != nil {
			return httpapi.Fail(c, 500, "STORAGE_UNAVAILABLE", "The profile image could not be stored. Please retry.", nil)
		}
	}
	u, err := h.Store.UpdateUser(userID(c), func(user *domain.User) { user.ProfilePhotoURL = url })
	if err != nil {
		return err
	}
	return httpapi.Created(c, fiber.Map{"profile_photo_url": u.ProfilePhotoURL, "max_bytes": profile.MaxPhotoBytes, "allowed_mime_types": []string{"image/jpeg", "image/png", "image/webp"}})
}
func (h *Core) ApplyProfilePhoto(c *fiber.Ctx) error {
	user, err := h.Store.GetUser(userID(c))
	if err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "User not found.", nil)
	}
	if user.ProfilePhotoURL == "" {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", "Upload a profile image before applying it to a CV.", nil)
	}
	cv, err := h.Store.CV(userID(c), c.Params("id"))
	if err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "CV not found.", nil)
	}
	personal, ok := cv.Content["personal"].(map[string]any)
	if !ok {
		personal = map[string]any{}
	}
	personal["photo_url"] = user.ProfilePhotoURL
	personal["show_photo"] = true
	cv.Content["personal"] = personal
	updated, err := h.Store.PatchCV(userID(c), cv.ID, cv.Content, "Applied profile photo")
	if err != nil {
		return err
	}
	return httpapi.OK(c, updated)
}
func (h *Core) Wallet(c *fiber.Ctx) error {
	u, err := h.Store.GetUser(userID(c))
	if err != nil {
		return err
	}
	exportLimit, exportsRemaining := h.exportUsage(userID(c))
	return httpapi.OK(c, fiber.Map{"export_credits": u.ExportCredits, "daily_export_limit": exportLimit, "daily_exports_remaining": exportsRemaining, "ai_token_balance": u.AITokenBalance, "voice_seconds_balance": u.VoiceSecondsBalance})
}
func (h *Core) Templates(c *fiber.Ctx) error {
	out := []fiber.Map{}
	for _, t := range h.Store.Templates() {
		if !t.Active {
			continue
		}
		out = append(out, fiber.Map{"id": t.ID, "name": t.Name, "family": t.Family, "supports_photo": t.SupportsPhoto, "locked": false, "accent_color_hex": t.AccentColorHex})
	}
	return httpapi.OK(c, out)
}
func (h *Core) CreateCV(c *fiber.Ctx) error {
	var body struct {
		Title         string         `json:"title"`
		TemplateID    string         `json:"template_id"`
		DocumentStyle string         `json:"document_style"`
		TargetRegion  string         `json:"target_region"`
		Language      string         `json:"language"`
		Content       map[string]any `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", "Provide valid CV details.", nil)
	}
	if body.TemplateID == "" {
		body.TemplateID = "ats_classic"
	}
	cv, err := h.Store.CreateCV(domain.CV{UserID: userID(c), Title: defaultString(body.Title, "Untitled CV"), TemplateID: body.TemplateID, DocumentStyle: defaultString(body.DocumentStyle, "resume"), TargetRegion: defaultString(body.TargetRegion, "eac"), Language: defaultString(body.Language, "en"), Content: body.Content})
	if err != nil {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", "The selected template is unavailable.", nil)
	}
	return httpapi.Created(c, cv)
}
func (h *Core) ListCVs(c *fiber.Ctx) error { return httpapi.OK(c, h.Store.CVs(userID(c))) }
func (h *Core) GetCV(c *fiber.Ctx) error {
	cv, err := h.Store.CV(userID(c), c.Params("id"))
	if err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "CV not found.", nil)
	}
	return httpapi.OK(c, cv)
}
func (h *Core) PatchCV(c *fiber.Ctx) error {
	var body struct {
		Content       map[string]any `json:"content"`
		ChangeSummary string         `json:"change_summary"`
	}
	if err := c.BodyParser(&body); err != nil {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", "Provide valid CV content.", nil)
	}
	cv, err := h.Store.PatchCV(userID(c), c.Params("id"), body.Content, defaultString(body.ChangeSummary, "Updated CV"))
	if err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "CV not found.", nil)
	}
	return httpapi.OK(c, cv)
}
func (h *Core) DeleteCV(c *fiber.Ctx) error {
	if err := h.Store.ArchiveCV(userID(c), c.Params("id")); err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "CV not found.", nil)
	}
	return httpapi.OK(c, fiber.Map{"archived": true})
}
func (h *Core) DuplicateCV(c *fiber.Ctx) error {
	cv, err := h.Store.DuplicateCV(userID(c), c.Params("id"))
	if err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "CV not found.", nil)
	}
	return httpapi.Created(c, cv)
}
func (h *Core) Versions(c *fiber.Ctx) error {
	versions, err := h.Store.Versions(userID(c), c.Params("id"))
	if err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "CV not found.", nil)
	}
	return httpapi.OK(c, versions)
}
func (h *Core) RestoreVersion(c *fiber.Ctx) error {
	cv, err := h.Store.RestoreVersion(userID(c), c.Params("id"), c.Params("versionID"))
	if err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "CV or version not found.", nil)
	}
	return httpapi.OK(c, cv)
}
func (h *Core) ATSScore(c *fiber.Ctx) error {
	cv, err := h.Store.CV(userID(c), c.Params("id"))
	if err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "CV not found.", nil)
	}
	t, err := h.Store.Template(cv.TemplateID)
	if err != nil {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", "Template is unavailable.", nil)
	}
	return httpapi.OK(c, lint.Score(cv, t))
}
func (h *Core) ExportCV(c *fiber.Ctx) error {
	var body struct {
		Format   string `json:"format"`
		Language string `json:"language"`
	}
	if err := c.BodyParser(&body); err != nil || (body.Format != "pdf" && body.Format != "docx" && body.Format != "csv") || (body.Language != "" && body.Language != "en" && body.Language != "sw") {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", "Export format must be pdf, docx, or csv.", nil)
	}
	cv, err := h.Store.CV(userID(c), c.Params("id"))
	if err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "CV not found.", nil)
	}
	if body.Language != "" {
		cv.Language = body.Language
	}
	template, err := h.Store.Template(cv.TemplateID)
	if err != nil {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", "Template is unavailable.", nil)
	}
	exportLimit, exportsRemaining := h.exportUsage(userID(c))
	cacheFormat := body.Format
	if body.Format != "csv" && body.Language != "" {
		cacheFormat += "_" + body.Language
	}
	e, err := h.Store.ConsumeExportCredit(userID(c), c.Params("id"), cacheFormat)
	if err != nil {
		if err == store.ErrDailyExportLimit {
			return httpapi.Fail(c, fiber.StatusTooManyRequests, "DAILY_EXPORT_LIMIT", "You have used all 5 exports for today. Your limit resets at midnight Tanzania time.", fiber.Map{"daily_export_limit": store.DailyExportLimit, "daily_exports_remaining": 0})
		}
		return httpapi.Fail(c, 404, "NOT_FOUND", "CV not found.", nil)
	}
	if e.ObjectKey != "" {
		filename := filepath.Base(e.ObjectKey)
		if filename == "." || filename == "" {
			filename = filepath.Base(e.FileURL)
		}
		return httpapi.Created(c, fiber.Map{"job_id": e.ID, "status": "done", "file_url": e.FileURL, "filename": filename, "bytes": e.ByteSize, "credit_consumed": false, "daily_export_limit": exportLimit, "daily_exports_remaining": exportsRemaining, "download_url": strings.TrimRight(h.PublicAPIBaseURL, "/") + "/exports/" + filename})
	}
	artifact, err := h.Renderer.Render(c.UserContext(), cv, template, body.Format, false)
	if err != nil {
		slog.Error("CV export render failed", "cv_id", cv.ID, "format", body.Format, "error", err)
		_ = h.Store.FailExport(userID(c), e.ID)
		return httpapi.Fail(c, 500, "RENDER_FAILED", "The document could not be rendered. Please retry.", nil)
	}
	// The durable copy is stored through the blob store below; remove the
	// renderer's local working file after this request to avoid disk growth.
	defer os.Remove(artifact.Path)
	raw, err := os.ReadFile(artifact.Path)
	if err != nil {
		_ = h.Store.FailExport(userID(c), e.ID)
		return httpapi.Fail(c, 500, "RENDER_FAILED", "The generated document could not be read. Please retry.", nil)
	}
	objectKey := artifact.Filename
	if h.Blobs != nil {
		if err = h.Blobs.Put(c.UserContext(), objectKey, raw, artifact.MIMEType); err != nil {
			_ = h.Store.FailExport(userID(c), e.ID)
			return httpapi.Fail(c, 503, "STORAGE_UNAVAILABLE", "The export could not be stored. Please retry.", nil)
		}
	}
	if err = h.Store.CompleteExport(userID(c), e.ID, objectKey, int64(len(raw))); err != nil {
		_ = h.Store.FailExport(userID(c), e.ID)
		return httpapi.Fail(c, 500, "EXPORT_RECORD_FAILED", "The export record could not be completed.", nil)
	}
	result := fiber.Map{"job_id": e.ID, "status": "done", "file_url": objectKey, "filename": artifact.Filename, "bytes": artifact.Bytes, "credit_consumed": false, "daily_export_limit": exportLimit, "daily_exports_remaining": exportsRemaining}
	if h.PublicAPIBaseURL != "" {
		result["download_url"] = strings.TrimRight(h.PublicAPIBaseURL, "/") + "/exports/" + artifact.Filename
	}
	return httpapi.Created(c, result)
}

func (h *Core) exportUsage(id string) (limit, remaining int) {
	user, err := h.Store.GetUser(id)
	if err == nil && user.Plan == store.LocalDesktopPlan {
		return 0, -1
	}
	return store.DailyExportLimit, h.Store.DailyExportsRemaining(id, time.Now().UTC())
}

func (h *Core) EmailExport(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Filename string `json:"filename"`
	}
	if err := c.BodyParser(&body); err != nil || !strings.Contains(body.Email, "@") || body.Filename == "" {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", "Provide a valid email and export filename.", nil)
	}
	if h.MailAfrica == nil || h.MailAfrica.APIKey == "" {
		return httpapi.Fail(c, 503, "EMAIL_NOT_CONFIGURED", "Email delivery is not configured.", nil)
	}
	filename := filepath.Base(body.Filename)
	export, err := h.Store.ExportForUser(userID(c), filename)
	if err != nil || export.ObjectKey == "" {
		return httpapi.Fail(c, 404, "NOT_FOUND", "Export file not found. Export the CV first.", nil)
	}
	raw, err := h.Blobs.Get(c.UserContext(), export.ObjectKey)
	if err != nil {
		return httpapi.Fail(c, 404, "NOT_FOUND", "Export file is no longer available.", nil)
	}
	contentType := "application/octet-stream"
	if strings.HasPrefix(export.Format, "pdf") {
		contentType = "application/pdf"
	} else if strings.HasPrefix(export.Format, "docx") {
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	} else if strings.HasPrefix(export.Format, "csv") {
		contentType = "text/csv"
	}
	_, err = h.MailAfrica.Send(context.Background(), mailafrica.Email{To: []string{strings.TrimSpace(body.Email)}, Subject: "Your Tengeneza CV", TextBody: "Your CV is attached. Thank you for using Tengeneza CV.", HTMLBody: "<p>Your CV is attached. Thank you for using <strong>Tengeneza CV</strong>.</p>", Attachments: []mailafrica.Attachment{{Filename: filename, ContentType: contentType, Size: len(raw), DataBase64: base64.StdEncoding.EncodeToString(raw)}}})
	if err != nil {
		slog.Error("CV email delivery failed", "user_id", userID(c), "error", err)
		return httpapi.Fail(c, 502, "EMAIL_DELIVERY_FAILED", "The CV email could not be sent.", nil)
	}
	return httpapi.OK(c, fiber.Map{"sent": true})
}
func publicUser(u domain.User) fiber.Map {
	return fiber.Map{"id": u.ID, "full_name": u.FullName, "email": u.Email, "phone": u.Phone, "profile_photo_url": u.ProfilePhotoURL, "preferred_language": u.PreferredLanguage, "target_region": u.TargetRegion, "plan": u.Plan, "role": u.Role}
}
func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
