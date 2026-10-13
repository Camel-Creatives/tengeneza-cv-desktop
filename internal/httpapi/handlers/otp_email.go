// OTP verification delivery over MailAfrica transactional email.
package handlers

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/apiresponse"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/auth"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/clients/mailafrica"
)

const otpExpiresInSeconds = 300

func otpEmail(destination, code string) mailafrica.Email {
	return mailafrica.Email{
		To:          []string{destination},
		FromAddress: "mail@mailafrica.online",
		Subject:     "Your Tengeneza CV verification code",
		TextBody:    "Your Tengeneza CV verification code is " + code + ". It expires in 5 minutes. If you did not request this code, you can ignore this email.",
		HTMLBody:    `<div style="font-family:Arial,Helvetica,sans-serif;background:#F7FAFF;padding:32px 16px"><div style="max-width:520px;margin:auto;background:#ffffff;border:1px solid #DCE9FF;border-radius:16px;overflow:hidden"><div style="background:#102A56;padding:20px 28px;color:#ffffff;font-size:14px;font-weight:800;letter-spacing:.18em">TENGENEZA <span style="color:#6EA8FF">CV</span></div><div style="padding:28px"><p style="margin:0 0 14px;color:#48617D;font-size:15px;line-height:1.6">Habari! Use this six-digit code to verify your Tengeneza CV account:</p><p style="margin:0 0 18px;font-size:34px;font-weight:800;letter-spacing:.35em;color:#155EEF">` + code + `</p><p style="margin:0;color:#48617D;font-size:13px;line-height:1.6">This code expires in 5 minutes. If you did not request it, you can safely ignore this email.</p></div></div></div>`,
	}
}

func deliverOTP(h *Core, destination, code string) (bool, error) {
	if !strings.Contains(destination, "@") || h.MailAfrica == nil || h.MailAfrica.APIKey == "" {
		return false, nil
	}
	email := otpEmail(destination, code)
	if _, err := h.MailAfrica.Send(context.Background(), email); err != nil {
		return true, err
	}
	return true, nil
}

func (h *Core) ResendOTP(c *fiber.Ctx) error {
	var body struct {
		VerificationID string `json:"verification_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.VerificationID == "" {
		return apiresponse.Fail(c, 400, "VALIDATION_ERROR", "Provide a verification ID.", nil)
	}
	previous, err := h.Store.GetOTP(body.VerificationID)
	if err != nil {
		return apiresponse.Fail(c, 404, "NOT_FOUND", "Verification challenge was not found. Please sign up again.", nil)
	}
	otp, code, err := auth.NewOTP(previous.UserID, previous.Destination, previous.Purpose)
	if err != nil {
		return err
	}
	h.Store.SaveOTP(otp)
	delivered, err := deliverOTP(h, previous.Destination, code)
	if err != nil {
		return apiresponse.Fail(c, 502, "OTP_DELIVERY_FAILED", "The verification email could not be sent right now. Please try again in a moment.", nil)
	}
	result := fiber.Map{"verification_id": otp.ID, "destination": previous.Destination, "expires_in_seconds": otpExpiresInSeconds}
	if !delivered && !h.IsProduction {
		result["development_otp"] = code
	}
	return apiresponse.OK(c, result)
}
