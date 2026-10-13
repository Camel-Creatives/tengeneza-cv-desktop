# Verified Provider Contracts


## MailAfrica webhooks

The official [MailAfrica Webhooks documentation](https://docs.mailafrica.online/webhooks) documents `inbound.message_received` events carrying `webhook_id`, `address_id`, `message_id`, and `timestamp`. The complete message is fetched from `GET /api/inbound/messages/{message_id}` rather than supplied in the event body.

MailAfrica signs the exact raw body with an HMAC-SHA256 hexadecimal signature in both `X-Signature` and `X-Webhook-Signature`. Webhook delivery retries up to three times with 1s, 3s, and 9s backoff, so handling must be idempotent and return a 2xx response promptly. The service will use a stored webhook secret from webhook registration and deduplicate provider event IDs.

## Upstream design consequence

The product API at `api.tengenezacv.online` remains distinct from the upstream Ngamia AI gateway at `api.ngamia.cc/v1`. All code and configuration use the `NGAMIA_AI_*` prefix for the upstream gateway to preserve that separation.
