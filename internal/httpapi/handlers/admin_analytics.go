package handlers

import (
	"github.com/gofiber/fiber/v2"
	httpapi "github.com/troubleman96/tengeneza-cv-desktop/internal/apiresponse"
	"log/slog"
	"time"
)

func (h *Core) AdminAnalytics(c *fiber.Ctx) error {
	a, err := h.Store.AdminAnalytics(c.UserContext())
	if err != nil {
		slog.Error("admin analytics query failed", "error", err)
		return httpapi.Fail(c, 500, "ANALYTICS_FAILED", "Analytics could not be loaded.", nil)
	}
	daily := make([]fiber.Map, 0, len(a.Daily))
	for _, x := range a.Daily {
		daily = append(daily, fiber.Map{"date": x.Date, "users": x.Users, "cvs": x.CVs})
	}
	regions := make([]fiber.Map, 0, len(a.Regions))
	for _, x := range a.Regions {
		regions = append(regions, fiber.Map{"region": x.Region, "users": x.Users})
	}
	templates := make([]fiber.Map, 0, len(a.TemplateUsage))
	for _, x := range a.TemplateUsage {
		templates = append(templates, fiber.Map{"id": x.ID, "name": x.Name, "family": x.Family, "usage": x.Usage})
	}
	return httpapi.OK(c, fiber.Map{"generated_at": time.Now().UTC(), "data_source": "postgres", "totals": fiber.Map{"users": a.Users, "active_cvs": a.ActiveCVs, "templates": a.Templates, "new_users_30d": a.NewUsers30d, "new_cvs_30d": a.NewCVs30d, "active_users_7d": a.ActiveUsers7d}, "daily": daily, "regions": regions, "templates": templates})
}
