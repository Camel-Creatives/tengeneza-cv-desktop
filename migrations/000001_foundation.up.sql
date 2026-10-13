CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE agencies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL, seats_purchased INT NOT NULL DEFAULT 0,
  brand_color TEXT, brand_logo_url TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), phone VARCHAR(20) UNIQUE, email VARCHAR(255) UNIQUE,
  password_hash TEXT, full_name TEXT NOT NULL, preferred_language TEXT NOT NULL DEFAULT 'en',
  target_region TEXT NOT NULL DEFAULT 'eac', plan TEXT NOT NULL DEFAULT 'free', export_credits INT NOT NULL DEFAULT 0,
  ai_token_balance BIGINT NOT NULL DEFAULT 20000, voice_seconds_balance INT NOT NULL DEFAULT 0,
  role TEXT NOT NULL DEFAULT 'user', agency_id UUID REFERENCES agencies(id), phone_verified_at TIMESTAMPTZ,
  email_verified_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (preferred_language IN ('en','sw')), CHECK (plan IN ('free','payg','career_pro','agency'))
);
CREATE TABLE otp_codes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID REFERENCES users(id), destination TEXT NOT NULL,
  channel TEXT NOT NULL, code_hash TEXT NOT NULL, purpose TEXT NOT NULL, expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ, attempts SMALLINT NOT NULL DEFAULT 0, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL REFERENCES users(id), token_hash TEXT NOT NULL,
  revoked BOOLEAN NOT NULL DEFAULT false, expires_at TIMESTAMPTZ NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE templates (
  id TEXT PRIMARY KEY, name TEXT NOT NULL, family TEXT NOT NULL, is_free BOOLEAN NOT NULL DEFAULT false,
  supports_photo BOOLEAN NOT NULL DEFAULT false, preview_image_url TEXT, html_template_path TEXT NOT NULL,
  accent_color_hex TEXT NOT NULL DEFAULT '#0F5132', active BOOLEAN NOT NULL DEFAULT true
);
CREATE TABLE cvs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL REFERENCES users(id), title TEXT NOT NULL DEFAULT 'Untitled CV',
  document_style TEXT NOT NULL DEFAULT 'resume', template_id TEXT NOT NULL REFERENCES templates(id), target_region TEXT NOT NULL DEFAULT 'eac',
  language TEXT NOT NULL DEFAULT 'en', content JSONB NOT NULL DEFAULT '{}'::jsonb, ats_score INT, ats_report JSONB,
  is_archived BOOLEAN NOT NULL DEFAULT false, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE cv_versions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), cv_id UUID NOT NULL REFERENCES cvs(id), content JSONB NOT NULL,
  changed_by TEXT NOT NULL, change_summary TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE cv_exports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), cv_id UUID NOT NULL REFERENCES cvs(id), cv_version_id UUID NOT NULL REFERENCES cv_versions(id),
  format TEXT NOT NULL, content_hash TEXT NOT NULL, file_url TEXT NOT NULL, credit_consumed BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE pricing_config (
  code TEXT PRIMARY KEY, label TEXT NOT NULL, amount_tzs BIGINT NOT NULL, export_credits INT NOT NULL DEFAULT 0,
  ai_token_grant BIGINT NOT NULL DEFAULT 0, voice_seconds_grant INT NOT NULL DEFAULT 0, is_subscription BOOLEAN NOT NULL DEFAULT false,
  subscription_period_days INT, active BOOLEAN NOT NULL DEFAULT true
);
CREATE TABLE payment_intents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL REFERENCES users(id), pack_code TEXT NOT NULL REFERENCES pricing_config(code),
  amount_tzs BIGINT NOT NULL, idempotency_key TEXT NOT NULL UNIQUE, provider_session_reference TEXT UNIQUE,
  status TEXT NOT NULL DEFAULT 'pending', raw_webhook_payload JSONB, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), completed_at TIMESTAMPTZ
);
CREATE TABLE subscriptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL REFERENCES users(id), pack_code TEXT NOT NULL REFERENCES pricing_config(code),
  status TEXT NOT NULL DEFAULT 'active', renews_at TIMESTAMPTZ NOT NULL, grace_ends_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE ai_usage_log (
  id BIGSERIAL PRIMARY KEY, user_id UUID NOT NULL REFERENCES users(id), cv_id UUID REFERENCES cvs(id), kind TEXT NOT NULL,
  model TEXT, prompt_tokens INT, completion_tokens INT, total_tokens INT, audio_seconds INT, ai_credits_charged BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE email_log (
  id BIGSERIAL PRIMARY KEY, user_id UUID REFERENCES users(id), mailafrica_message_id BIGINT, kind TEXT NOT NULL,
  to_address TEXT NOT NULL, status TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE webhook_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), provider TEXT NOT NULL, provider_event_id TEXT NOT NULL, payload JSONB NOT NULL,
  processed_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE(provider, provider_event_id)
);
CREATE INDEX idx_cvs_user_active ON cvs(user_id) WHERE NOT is_archived;
CREATE INDEX idx_payment_intents_pending ON payment_intents(status) WHERE status = 'pending';
