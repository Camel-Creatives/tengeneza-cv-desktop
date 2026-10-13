package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	httpapi "github.com/troubleman96/tengeneza-cv-desktop/internal/apiresponse"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/auth"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
)

// AdminPasswordLogin authenticates the single operator configured by the
// deployment's ADMIN_LOGIN_EMAIL and ADMIN_LOGIN_PASSWORD environment values.
func (h *Core) AdminPasswordLogin(c *fiber.Ctx) error {
	if h.AdminLoginEmail == "" || h.AdminPasswordHash == "" {
		return httpapi.Fail(c, fiber.StatusServiceUnavailable, "ADMIN_LOGIN_NOT_CONFIGURED", "Administrator password sign-in is not configured.", nil)
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil || body.Email == "" || body.Password == "" {
		return httpapi.Fail(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Provide your administrator email and password.", nil)
	}
	providedEmail := strings.ToLower(strings.TrimSpace(body.Email))
	passwordValid := auth.VerifyPassword(h.AdminPasswordHash, body.Password)
	if providedEmail != h.AdminLoginEmail || !passwordValid {
		return httpapi.Fail(c, fiber.StatusUnauthorized, "INVALID_CREDENTIALS", "The email or password is incorrect.", nil)
	}
	secret, enrolled, err := adminSecret(h.AdminTOTPSecretFile)
	if err != nil {
		return httpapi.Fail(c, fiber.StatusInternalServerError, "TOTP_SETUP_FAILED", "Administrator authenticator setup could not be prepared.", nil)
	}
	if h.AdminChallenges == nil {
		h.AdminChallenges = newAdminChallengeStore()
	}
	challengeID := h.AdminChallenges.put(AdminChallenge{Secret: secret, Enrollment: !enrolled, ExpiresAt: time.Now().Add(10 * time.Minute)})
	if !enrolled {
		return httpapi.OK(c, fiber.Map{"challenge_id": challengeID, "status": "enrollment_required", "method": "totp", "expires_in_seconds": 600, "setup": fiber.Map{"secret": secret, "otpauth_url": adminOTPURI(h.AdminLoginEmail, secret)}})
	}
	return httpapi.OK(c, fiber.Map{"challenge_id": challengeID, "status": "verification_required", "method": "totp", "expires_in_seconds": 600, "masked_target": maskEmail(h.AdminLoginEmail)})
}

func (h *Core) AdminTOTPVerify(c *fiber.Ctx) error {
	var body struct {
		ChallengeID string `json:"challenge_id"`
		Code        string `json:"code"`
	}
	if err := c.BodyParser(&body); err != nil || body.ChallengeID == "" || body.Code == "" {
		return httpapi.Fail(c, 400, "VALIDATION_ERROR", "Provide the authenticator code.", nil)
	}
	if h.AdminChallenges == nil {
		return httpapi.Fail(c, 401, "CHALLENGE_EXPIRED", "Sign in again to continue.", nil)
	}
	challenge, ok := h.AdminChallenges.take(body.ChallengeID)
	if !ok || time.Now().After(challenge.ExpiresAt) || !auth.VerifyTOTP(challenge.Secret, body.Code, time.Now()) {
		return httpapi.Fail(c, 401, "INVALID_TOTP", "That authenticator code is invalid or expired.", nil)
	}
	userID := "admin:" + h.AdminLoginEmail
	accessToken, refreshToken, err := h.Auth.IssueTokens(userID, string(domain.RoleSuperAdmin))
	if err != nil {
		return httpapi.Fail(c, fiber.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "Administrator sign-in could not be completed.", nil)
	}
	return httpapi.OK(c, fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          fiber.Map{"id": userID, "full_name": "Administrator", "email": h.AdminLoginEmail, "preferred_language": "en", "target_region": "eac", "plan": "admin", "role": string(domain.RoleSuperAdmin)},
	})
}

func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return email
	}
	local := parts[0]
	if len(local) > 2 {
		local = local[:2] + "***"
	}
	return local + "@" + parts[1]
}
