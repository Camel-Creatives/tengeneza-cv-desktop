// Package mailafrica keeps outbound email and inbound support-webhook handling isolated from HTTP handlers.
package mailafrica

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrInvalidSignature = errors.New("invalid MailAfrica webhook signature")

type Client struct {
	BaseURL, APIKey, WebhookSecret, DefaultFrom string
	FromDomainID                                int64
	HTTP                                        *http.Client
}
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
	DataBase64  string `json:"data_base64"`
}
type Email struct {
	To                          []string `json:"to"`
	Subject, HTMLBody, TextBody string
	FromAddress                 string       `json:"from_address"`
	FromDomainID                int64        `json:"from_domain_id"`
	Attachments                 []Attachment `json:"attachments,omitempty"`
}

func New(baseURL, apiKey, secret, from string, domainID int64) *Client {
	return &Client{BaseURL: strings.TrimSuffix(baseURL, "/"), APIKey: apiKey, WebhookSecret: secret, DefaultFrom: from, FromDomainID: domainID, HTTP: &http.Client{Timeout: 15 * time.Second}}
}
func (c *Client) Send(ctx context.Context, email Email) (int64, error) {
	if c.APIKey == "" {
		return 0, nil
	}
	if email.FromAddress == "" {
		email.FromAddress = c.DefaultFrom
	}
	if email.FromDomainID == 0 {
		email.FromDomainID = c.FromDomainID
	}
	raw, err := json.Marshal(email)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/outbound/emails", bytes.NewReader(raw))
	if err != nil {
		return 0, err
	}
	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return 0, fmt.Errorf("MailAfrica send failed with status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var parsed struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return 0, err
	}
	return parsed.Data.ID, nil
}
func (c *Client) VerifyWebhook(raw []byte, signature string) error {
	if c.WebhookSecret == "" || signature == "" {
		return ErrInvalidSignature
	}
	mac := hmac.New(sha256.New, []byte(c.WebhookSecret))
	_, _ = mac.Write(raw)
	expected := mac.Sum(nil)
	provided, err := hex.DecodeString(signature)
	if err != nil || !hmac.Equal(provided, expected) {
		return ErrInvalidSignature
	}
	return nil
}
