ALTER TABLE users ADD COLUMN IF NOT EXISTS external_subject TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS users_external_subject_unique ON users(external_subject) WHERE external_subject IS NOT NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_photo_url TEXT;
ALTER TABLE users ALTER COLUMN full_name SET DEFAULT '';

ALTER TABLE cv_exports ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id);
ALTER TABLE cv_exports ADD COLUMN IF NOT EXISTS local_export_date DATE;
ALTER TABLE cv_exports ADD COLUMN IF NOT EXISTS object_key TEXT;
ALTER TABLE cv_exports ADD COLUMN IF NOT EXISTS byte_size BIGINT NOT NULL DEFAULT 0;
ALTER TABLE cv_exports ADD COLUMN IF NOT EXISTS cached BOOLEAN NOT NULL DEFAULT false;

UPDATE cv_exports e SET user_id = c.user_id FROM cvs c WHERE e.cv_id = c.id AND e.user_id IS NULL;
UPDATE cv_exports SET local_export_date = (created_at AT TIME ZONE 'Africa/Dar_es_Salaam')::date WHERE local_export_date IS NULL;
ALTER TABLE cv_exports ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE cv_exports ALTER COLUMN local_export_date SET NOT NULL;

CREATE INDEX IF NOT EXISTS cv_exports_user_day_idx ON cv_exports(user_id, local_export_date) WHERE cached = false;
CREATE UNIQUE INDEX IF NOT EXISTS cv_exports_unique_artifact_idx ON cv_exports(cv_id, format, content_hash);

CREATE TABLE IF NOT EXISTS daily_export_usage (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  export_date DATE NOT NULL,
  export_count SMALLINT NOT NULL DEFAULT 0 CHECK (export_count BETWEEN 0 AND 5),
  PRIMARY KEY (user_id, export_date)
);

INSERT INTO templates (id, name, family, is_free, supports_photo, html_template_path, accent_color_hex, active)
VALUES
  ('ats_classic','ATS Classic','ats',true,false,'ats_classic','#0F5132',true),
  ('ats_modern','ATS Modern','ats',true,false,'ats_modern','#1E3A5F',true),
  ('global_minimal','Global Minimal','ats',true,false,'global_minimal','#334155',true),
  ('executive_ats','Executive ATS','ats',true,false,'executive_ats','#4A3225',true),
  ('visual_local','Visual Local','visual',true,true,'visual_local','#0F5132',true),
  ('hospitality_local','Hospitality Local','visual',true,true,'hospitality_local','#0D5B57',true),
  ('academic_cv','Academic CV','academic',true,false,'academic_cv','#4A3B6A',true),
  ('career_change','Career Change','ats',true,false,'career_change','#A14D2F',true),
  ('creative_portfolio','Creative Portfolio','visual',true,true,'creative_portfolio','#C6337B',true),
  ('tech_product','Tech Product','visual',true,true,'tech_product','#36D1A4',true),
  ('finance_audit','Finance Audit','visual',true,true,'finance_audit','#183B31',true),
  ('legal_chambers','Legal Chambers','visual',true,true,'legal_chambers','#7C5675',true),
  ('clinical_care','Clinical Care','visual',true,true,'clinical_care','#267DAF',true),
  ('educator_classroom','Educator Classroom','visual',true,true,'educator_classroom','#AF4D3D',true),
  ('sales_growth','Sales Growth','visual',true,true,'sales_growth','#EF632B',true),
  ('field_operations','Field Operations','visual',true,true,'field_operations','#304A48',true),
  ('consulting_strategy','Consulting Strategy','visual',true,true,'consulting_strategy','#263A59',true),
  ('engineering_blueprint','Engineering Blueprint','visual',true,true,'engineering_blueprint','#163C57',true),
  ('public_service','Public Service','visual',true,true,'public_service','#542737',true),
  ('nonprofit_impact','Nonprofit Impact','visual',true,true,'nonprofit_impact','#17705B',true),
  ('graduate_launch','Graduate Launch','visual',true,true,'graduate_launch','#E05C42',true),
  ('skilled_trades','Skilled Trades','visual',true,true,'skilled_trades','#F6C34B',true),
  ('remote_async','Remote Async','visual',true,true,'remote_async','#6456A7',true),
  ('creative_producer','Creative Producer','visual',true,true,'creative_producer','#FF8364',true)
ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, family=EXCLUDED.family,
  is_free=EXCLUDED.is_free, supports_photo=EXCLUDED.supports_photo,
  html_template_path=EXCLUDED.html_template_path, accent_color_hex=EXCLUDED.accent_color_hex, active=true;
