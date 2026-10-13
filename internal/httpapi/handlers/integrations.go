package handlers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/apiresponse"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/billing"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/clients/ngamiaai"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/render/lint"
)

func (h *Core) BillingPlans(c *fiber.Ctx) error {
	if h.FreePlatform {
		return apiresponse.OK(c, []fiber.Map{{
			"code": "free", "label": "Local desktop", "amount_tzs": 0,
			"export_credits": 0, "ai_tokens": 0,
			"description": "Unlimited local CV creation and exports. No subscription or sign-in required.",
			"features": []string{
				"All 24 professional templates", "Unlimited CVs and version history",
				"Unlimited PDF, DOCX and CSV exports", "Unlimited ATS reviews",
				"Optional AI drafting when a provider is configured",
			},
		}})
	}
	return apiresponse.OK(c, billing.Plans())
}
func (h *Core) MailAfricaWebhook(c *fiber.Ctx) error {
	raw := c.Body()
	signature := c.Get("X-Webhook-Signature")
	if signature == "" {
		signature = c.Get("X-Signature")
	}
	if err := h.MailAfrica.VerifyWebhook(raw, signature); err != nil {
		return apiresponse.Fail(c, 400, "VALIDATION_ERROR", "Webhook signature verification failed.", nil)
	}
	return apiresponse.OK(c, fiber.Map{"received": true})
}
func (h *Core) AutofillSection(c *fiber.Ctx) error {
	var body struct{ Section, Instruction string }
	if err := c.BodyParser(&body); err != nil || body.Section == "" {
		return apiresponse.Fail(c, 400, "VALIDATION_ERROR", "Provide the CV section to draft.", nil)
	}
	cv, err := h.Store.CV(userID(c), c.Params("id"))
	if err != nil {
		return apiresponse.Fail(c, 404, "NOT_FOUND", "CV not found.", nil)
	}
	prompt := "Draft only the " + body.Section + " section for an ATS-safe CV. Do not invent facts. User instruction: " + body.Instruction
	response, err := h.NgamiaAI.Chat(c.Context(), ngamiaai.CompletionRequest{Model: "", Messages: []ngamiaai.Message{{Role: "system", Content: "You are the Tengeneza CV Agent. Never invent employers, dates, metrics, or personal data."}, {Role: "user", Content: prompt + "\nCV content: " + stringify(cv.Content)}}})
	if err != nil {
		return apiresponse.Fail(c, 502, "UPSTREAM_AI_ERROR", "AI drafting is temporarily unavailable.", nil)
	}
	draft := ""
	if len(response.Choices) > 0 {
		draft = response.Choices[0].Message.Content
	}
	return apiresponse.OK(c, fiber.Map{"section": body.Section, "draft": draft, "usage": response.Usage})
}
func (h *Core) ReScore(c *fiber.Ctx) error {
	cv, err := h.Store.CV(userID(c), c.Params("id"))
	if err != nil {
		return apiresponse.Fail(c, 404, "NOT_FOUND", "CV not found.", nil)
	}
	template, err := h.Store.Template(cv.TemplateID)
	if err != nil {
		return apiresponse.Fail(c, 400, "VALIDATION_ERROR", "Template is unavailable.", nil)
	}
	return apiresponse.OK(c, lint.Score(cv, template))
}
func (h *Core) Examples(c *fiber.Ctx) error {
	entries, err := os.ReadDir(h.ExamplesDir)
	if err != nil {
		return apiresponse.OK(c, []fiber.Map{})
	}
	out := []fiber.Map{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(h.ExamplesDir, entry.Name()))
		if err != nil {
			continue
		}
		var doc struct {
			Title      string `json:"title"`
			TemplateID string `json:"template_id"`
			Content    struct {
				Personal struct {
					FullName string `json:"full_name"`
					Headline string `json:"headline"`
				} `json:"personal"`
			} `json:"content"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			continue
		}
		slug := strings.TrimSuffix(entry.Name(), ".json")
		out = append(out, fiber.Map{"slug": slug, "title": doc.Title, "template_id": doc.TemplateID, "full_name": doc.Content.Personal.FullName, "headline": doc.Content.Personal.Headline})
	}
	return apiresponse.OK(c, out)
}
func (h *Core) Example(c *fiber.Ctx) error {
	slug := filepath.Base(c.Params("slug"))
	if !strings.HasSuffix(slug, ".json") {
		slug += ".json"
	}
	raw, err := os.ReadFile(filepath.Join(h.ExamplesDir, slug))
	if err != nil {
		return apiresponse.Fail(c, 404, "NOT_FOUND", "Example CV not found.", nil)
	}
	var payload fiber.Map
	if err := json.Unmarshal(raw, &payload); err != nil {
		return apiresponse.Fail(c, 500, "INTERNAL_ERROR", "The example CV could not be read.", nil)
	}
	return apiresponse.OK(c, payload)
}
func stringify(value any) string { raw, _ := json.Marshal(value); return string(raw) }
