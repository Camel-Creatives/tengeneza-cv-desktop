// Package httpapi implements the product's canonical response envelope.
package httpapi

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Meta struct { RequestID string `json:"request_id"`; Timestamp time.Time `json:"timestamp"` }
type Success struct { Success bool `json:"success"`; Data any `json:"data"`; Meta Meta `json:"meta"` }
type APIError struct { Code string `json:"code"`; Message string `json:"message"`; Details any `json:"details,omitempty"` }
type Failure struct { Success bool `json:"success"`; Error APIError `json:"error"`; Meta Meta `json:"meta"` }

func requestMeta(c *fiber.Ctx) Meta { return Meta{RequestID: c.Locals("request_id").(string), Timestamp: time.Now().UTC()} }
func OK(c *fiber.Ctx, data any) error { return c.Status(fiber.StatusOK).JSON(Success{Success: true, Data: data, Meta: requestMeta(c)}) }
func Created(c *fiber.Ctx, data any) error { return c.Status(fiber.StatusCreated).JSON(Success{Success: true, Data: data, Meta: requestMeta(c)}) }
func Fail(c *fiber.Ctx, status int, code, message string, details any) error { return c.Status(status).JSON(Failure{Success: false, Error: APIError{Code: code, Message: message, Details: details}, Meta: requestMeta(c)}) }
