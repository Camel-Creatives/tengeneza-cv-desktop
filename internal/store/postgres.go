package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/auth"
	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
	"github.com/troubleman96/tengeneza-cv-desktop/migrations"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func (s *PostgresStore) AdminAnalytics(ctx context.Context) (AdminAnalytics, error) {
	var a AdminAnalytics
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*), (SELECT COUNT(*) FROM cvs WHERE NOT is_archived), (SELECT COUNT(*) FROM templates WHERE active), COUNT(*) FILTER (WHERE created_at >= now()-interval '30 days'), (SELECT COUNT(*) FROM cvs WHERE created_at >= now()-interval '30 days'), COUNT(*) FILTER (WHERE updated_at >= now()-interval '7 days') FROM users`).Scan(&a.Users, &a.ActiveCVs, &a.Templates, &a.NewUsers30d, &a.NewCVs30d, &a.ActiveUsers7d)
	if err != nil {
		return a, err
	}
	rows, err := s.pool.Query(ctx, `SELECT COALESCE(target_region,'unknown'), COUNT(*) FROM users GROUP BY 1 ORDER BY 2 DESC LIMIT 10`)
	if err != nil {
		return a, err
	}
	for rows.Next() {
		var x RegionMetric
		if err := rows.Scan(&x.Region, &x.Users); err != nil {
			rows.Close()
			return a, err
		}
		a.Regions = append(a.Regions, x)
	}
	rows.Close()
	rows, err = s.pool.Query(ctx, `SELECT id,name,family,(SELECT COUNT(*) FROM cvs c WHERE c.template_id=t.id AND NOT c.is_archived) FROM templates t WHERE active ORDER BY 4 DESC LIMIT 10`)
	if err != nil {
		return a, err
	}
	for rows.Next() {
		var x TemplateMetric
		if err := rows.Scan(&x.ID, &x.Name, &x.Family, &x.Usage); err != nil {
			rows.Close()
			return a, err
		}
		a.TemplateUsage = append(a.TemplateUsage, x)
	}
	rows.Close()
	rows, err = s.pool.Query(ctx, `SELECT to_char(d,'YYYY-MM-DD'),COALESCE(u.n,0),COALESCE(c.n,0) FROM generate_series(current_date-29,current_date,interval '1 day') d LEFT JOIN (SELECT created_at::date AS created_date,COUNT(*) n FROM users WHERE created_at >= current_date-29 GROUP BY 1) u ON u.created_date=d::date LEFT JOIN (SELECT created_at::date AS created_date,COUNT(*) n FROM cvs WHERE created_at >= current_date-29 GROUP BY 1) c ON c.created_date=d::date ORDER BY d`)
	if err != nil {
		return a, err
	}
	defer rows.Close()
	for rows.Next() {
		var x AnalyticsDay
		if err := rows.Scan(&x.Date, &x.Users, &x.CVs); err != nil {
			return a, err
		}
		a.Daily = append(a.Daily, x)
	}
	return a, nil
}

func OpenPostgres(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if err := migrations.Apply(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}
func (s *PostgresStore) Close()                         { s.pool.Close() }
func (s *PostgresStore) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

func (s *PostgresStore) GetUser(id string) (domain.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return domain.User{}, ErrNotFound
	}
	return scanUser(s.pool.QueryRow(context.Background(), `SELECT id::text, COALESCE(external_subject,''), full_name, COALESCE(email,''), COALESCE(phone,''), COALESCE(preferred_language,'en'), COALESCE(target_region,'eac'), COALESCE(plan,'free'), export_credits, ai_token_balance, voice_seconds_balance, COALESCE(profile_photo_url,''), role, phone_verified_at IS NOT NULL, created_at, updated_at FROM users WHERE id=$1`, uid))
}
func (s *PostgresStore) EnsureExternalUser(subject string, claims ...domain.User) (domain.User, error) {
	if subject == "" {
		return domain.User{}, ErrNotFound
	}
	claim := domain.User{}
	if len(claims) > 0 {
		claim = claims[0]
	}
	row := s.pool.QueryRow(context.Background(), `INSERT INTO users (external_subject,full_name,email,phone,preferred_language,target_region,plan,export_credits,role) VALUES ($1,$2,NULLIF($3,''),NULLIF($4,''),'en','eac','free',5,'user') ON CONFLICT (external_subject) WHERE external_subject IS NOT NULL DO UPDATE SET full_name=CASE WHEN EXCLUDED.full_name<>'' THEN EXCLUDED.full_name ELSE users.full_name END, email=COALESCE(EXCLUDED.email,users.email), phone=COALESCE(EXCLUDED.phone,users.phone), updated_at=now() RETURNING id::text, COALESCE(external_subject,''), full_name, COALESCE(email,''), COALESCE(phone,''), COALESCE(preferred_language,'en'), COALESCE(target_region,'eac'), COALESCE(plan,'free'), export_credits, ai_token_balance, voice_seconds_balance, COALESCE(profile_photo_url,''), role, phone_verified_at IS NOT NULL, created_at, updated_at`, subject, claim.FullName, claim.Email, claim.Phone)
	return scanUser(row)
}
func (s *PostgresStore) UpdateUser(id string, mutate func(*domain.User)) (domain.User, error) {
	u, err := s.GetUser(id)
	if err != nil {
		return domain.User{}, err
	}
	mutate(&u)
	uid, _ := uuid.Parse(id)
	return scanUser(s.pool.QueryRow(context.Background(), `UPDATE users SET full_name=$2, email=NULLIF($3,''), phone=NULLIF($4,''), preferred_language=$5, target_region=$6, profile_photo_url=NULLIF($7,''), updated_at=now() WHERE id=$1 RETURNING id::text, COALESCE(external_subject,''), full_name, COALESCE(email,''), COALESCE(phone,''), COALESCE(preferred_language,'en'), COALESCE(target_region,'eac'), COALESCE(plan,'free'), export_credits, ai_token_balance, voice_seconds_balance, COALESCE(profile_photo_url,''), role, phone_verified_at IS NOT NULL, created_at, updated_at`, uid, u.FullName, u.Email, u.Phone, u.PreferredLanguage, u.TargetRegion, u.ProfilePhotoURL))
}
func (s *PostgresStore) SaveOTP(otp auth.OTP) {
	id, _ := uuid.Parse(otp.ID)
	uid, _ := uuid.Parse(otp.UserID)
	_, _ = s.pool.Exec(context.Background(), `INSERT INTO otp_codes(id,user_id,destination,channel,code_hash,purpose,expires_at,attempts) VALUES($1,$2,$3,'email',$4,$5,$6,$7) ON CONFLICT(id) DO UPDATE SET code_hash=EXCLUDED.code_hash,expires_at=EXCLUDED.expires_at,attempts=EXCLUDED.attempts`, id, uid, otp.Destination, otp.CodeHash, otp.Purpose, otp.ExpiresAt, otp.Attempts)
}
func (s *PostgresStore) GetOTP(id string) (auth.OTP, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return auth.OTP{}, ErrNotFound
	}
	var o auth.OTP
	var userID *string
	err = s.pool.QueryRow(context.Background(), `SELECT id::text,user_id::text,destination,purpose,code_hash,expires_at,attempts FROM otp_codes WHERE id=$1 AND consumed_at IS NULL`, uid).Scan(&o.ID, &userID, &o.Destination, &o.Purpose, &o.CodeHash, &o.ExpiresAt, &o.Attempts)
	if err != nil {
		return auth.OTP{}, mapErr(err)
	}
	if userID != nil {
		o.UserID = *userID
	}
	return o, nil
}
func (s *PostgresStore) ConsumeOTP(id string) {
	uid, err := uuid.Parse(id)
	if err == nil {
		_, _ = s.pool.Exec(context.Background(), `UPDATE otp_codes SET consumed_at=now() WHERE id=$1`, uid)
	}
}
func (s *PostgresStore) Templates() []domain.Template {
	rows, err := s.pool.Query(context.Background(), `SELECT id,name,family,is_free,supports_photo,active,accent_color_hex FROM templates WHERE active ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []domain.Template{}
	for rows.Next() {
		var t domain.Template
		if rows.Scan(&t.ID, &t.Name, &t.Family, &t.IsFree, &t.SupportsPhoto, &t.Active, &t.AccentColorHex) == nil {
			out = append(out, t)
		}
	}
	return out
}
func (s *PostgresStore) Template(id string) (domain.Template, error) {
	var t domain.Template
	err := s.pool.QueryRow(context.Background(), `SELECT id,name,family,is_free,supports_photo,active,accent_color_hex FROM templates WHERE id=$1 AND active`, id).Scan(&t.ID, &t.Name, &t.Family, &t.IsFree, &t.SupportsPhoto, &t.Active, &t.AccentColorHex)
	if err != nil {
		return t, mapErr(err)
	}
	return t, nil
}
func (s *PostgresStore) CreateCV(cv domain.CV) (domain.CV, error) {
	uid, err := uuid.Parse(cv.UserID)
	if err != nil {
		return domain.CV{}, ErrNotFound
	}
	raw, _ := json.Marshal(cv.Content)
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return domain.CV{}, err
	}
	defer tx.Rollback(context.Background())
	err = tx.QueryRow(context.Background(), `INSERT INTO cvs(user_id,title,document_style,template_id,target_region,language,content) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id::text,created_at,updated_at`, uid, cv.Title, cv.DocumentStyle, cv.TemplateID, cv.TargetRegion, cv.Language, raw).Scan(&cv.ID, &cv.CreatedAt, &cv.UpdatedAt)
	if err != nil {
		return domain.CV{}, mapErr(err)
	}
	cv.Content = cloneMap(cv.Content)
	if err = s.addVersionTx(tx, cv, "user", "Created CV"); err != nil {
		return domain.CV{}, err
	}
	if err = tx.Commit(context.Background()); err != nil {
		return domain.CV{}, err
	}
	return cv, nil
}
func (s *PostgresStore) CV(userID, id string) (domain.CV, error) {
	uid, e := uuid.Parse(userID)
	if e != nil {
		return domain.CV{}, ErrNotFound
	}
	cid, e := uuid.Parse(id)
	if e != nil {
		return domain.CV{}, ErrNotFound
	}
	return scanCV(s.pool.QueryRow(context.Background(), `SELECT id::text,user_id::text,title,document_style,template_id,target_region,language,content,COALESCE(ats_score,0),is_archived,created_at,updated_at FROM cvs WHERE id=$1 AND user_id=$2 AND NOT is_archived`, cid, uid))
}
func (s *PostgresStore) CVs(userID string) []domain.CV {
	uid, e := uuid.Parse(userID)
	if e != nil {
		return []domain.CV{}
	}
	rows, e := s.pool.Query(context.Background(), `SELECT id::text,user_id::text,title,document_style,template_id,target_region,language,content,COALESCE(ats_score,0),is_archived,created_at,updated_at FROM cvs WHERE user_id=$1 AND NOT is_archived ORDER BY updated_at DESC`, uid)
	if e != nil {
		return []domain.CV{}
	}
	defer rows.Close()
	out := []domain.CV{}
	for rows.Next() {
		if cv, err := scanCV(rows); err == nil {
			out = append(out, cv)
		}
	}
	return out
}
func (s *PostgresStore) PatchCV(userID, id string, content map[string]any, summary string) (domain.CV, error) {
	cv, e := s.CV(userID, id)
	if e != nil {
		return domain.CV{}, e
	}
	for k, v := range content {
		cv.Content[k] = v
	}
	raw, _ := json.Marshal(cv.Content)
	uid, _ := uuid.Parse(userID)
	cid, _ := uuid.Parse(id)
	tx, e := s.pool.Begin(context.Background())
	if e != nil {
		return domain.CV{}, e
	}
	defer tx.Rollback(context.Background())
	e = tx.QueryRow(context.Background(), `UPDATE cvs SET content=$3,updated_at=now() WHERE id=$1 AND user_id=$2 AND NOT is_archived RETURNING updated_at`, cid, uid, raw).Scan(&cv.UpdatedAt)
	if e != nil {
		return domain.CV{}, mapErr(e)
	}
	if e = s.addVersionTx(tx, cv, "user", summary); e != nil {
		return domain.CV{}, e
	}
	if e = tx.Commit(context.Background()); e != nil {
		return domain.CV{}, e
	}
	return s.CV(userID, id)
}
func (s *PostgresStore) DuplicateCV(userID, id string) (domain.CV, error) {
	cv, e := s.CV(userID, id)
	if e != nil {
		return domain.CV{}, e
	}
	cv.ID = ""
	cv.Title += " (copy)"
	return s.CreateCV(cv)
}
func (s *PostgresStore) ArchiveCV(userID, id string) error {
	uid, e := uuid.Parse(userID)
	if e != nil {
		return ErrNotFound
	}
	cid, e := uuid.Parse(id)
	if e != nil {
		return ErrNotFound
	}
	tag, e := s.pool.Exec(context.Background(), `UPDATE cvs SET is_archived=true,updated_at=now() WHERE id=$1 AND user_id=$2 AND NOT is_archived`, cid, uid)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *PostgresStore) Versions(userID, id string) ([]domain.CVVersion, error) {
	if _, e := s.CV(userID, id); e != nil {
		return nil, e
	}
	cid, _ := uuid.Parse(id)
	rows, e := s.pool.Query(context.Background(), `SELECT id::text,cv_id::text,content,changed_by,COALESCE(change_summary,''),created_at FROM cv_versions WHERE cv_id=$1 ORDER BY created_at`, cid)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.CVVersion{}
	for rows.Next() {
		var v domain.CVVersion
		var raw []byte
		if e = rows.Scan(&v.ID, &v.CVID, &raw, &v.ChangedBy, &v.ChangeSummary, &v.CreatedAt); e != nil {
			return nil, e
		}
		_ = json.Unmarshal(raw, &v.Content)
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *PostgresStore) RestoreVersion(userID, cvID, versionID string) (domain.CV, error) {
	if _, e := s.CV(userID, cvID); e != nil {
		return domain.CV{}, e
	}
	uid, _ := uuid.Parse(userID)
	cid, _ := uuid.Parse(cvID)
	vid, e := uuid.Parse(versionID)
	if e != nil {
		return domain.CV{}, ErrNotFound
	}
	var raw []byte
	e = s.pool.QueryRow(context.Background(), `SELECT v.content FROM cv_versions v JOIN cvs c ON c.id=v.cv_id WHERE v.id=$1 AND c.id=$2 AND c.user_id=$3 AND NOT c.is_archived`, vid, cid, uid).Scan(&raw)
	if e != nil {
		return domain.CV{}, mapErr(e)
	}
	var content map[string]any
	_ = json.Unmarshal(raw, &content)
	return s.PatchCV(userID, cvID, content, "Restored version")
}
func (s *PostgresStore) DailyExportsRemaining(userID string, now time.Time) int {
	uid, e := uuid.Parse(userID)
	if e != nil {
		return 0
	}
	day := now.In(tz()).Format("2006-01-02")
	var count int
	if s.pool.QueryRow(context.Background(), `SELECT export_count FROM daily_export_usage WHERE user_id=$1 AND export_date=$2`, uid, day).Scan(&count) != nil {
		return DailyExportLimit
	}
	remaining := DailyExportLimit - count
	if remaining < 0 {
		return 0
	}
	return remaining
}
func (s *PostgresStore) ConsumeExportCredit(userID, cvID, format string) (domain.Export, error) {
	ctx := context.Background()
	uid, e := uuid.Parse(userID)
	if e != nil {
		return domain.Export{}, ErrNotFound
	}
	cid, e := uuid.Parse(cvID)
	if e != nil {
		return domain.Export{}, ErrNotFound
	}
	tx, e := s.pool.Begin(ctx)
	if e != nil {
		return domain.Export{}, e
	}
	defer tx.Rollback(ctx)
	var cv domain.CV
	var raw []byte
	e = tx.QueryRow(ctx, `SELECT id::text,user_id::text,title,document_style,template_id,target_region,language,content,COALESCE(ats_score,0),is_archived,created_at,updated_at FROM cvs WHERE id=$1 AND user_id=$2 AND NOT is_archived FOR UPDATE`, cid, uid).Scan(&cv.ID, &cv.UserID, &cv.Title, &cv.DocumentStyle, &cv.TemplateID, &cv.TargetRegion, &cv.Language, &raw, &cv.ATSScore, &cv.Archived, &cv.CreatedAt, &cv.UpdatedAt)
	if e != nil {
		return domain.Export{}, mapErr(e)
	}
	_ = json.Unmarshal(raw, &cv.Content)
	hash := contentHash(cv.Content)
	var out domain.Export
	var key string
	e = tx.QueryRow(ctx, `SELECT id::text,cv_id::text,COALESCE(cv_version_id::text,''),format,content_hash,file_url,credit_consumed,user_id::text,cached,created_at,COALESCE(object_key,'') FROM cv_exports WHERE user_id=$1 AND cv_id=$2 AND format=$3 AND content_hash=$4 AND created_at>now()-interval '30 days' AND object_key<>'' ORDER BY created_at DESC LIMIT 1`, uid, cid, format, hash).Scan(&out.ID, &out.CVID, &out.VersionID, &out.Format, &out.ContentHash, &out.FileURL, &out.CreditConsumed, &out.UserID, &out.Cached, &out.CreatedAt, &key)
	if e == nil {
		out.Cached = true
		out.ObjectKey = key
		out.FileURL = pathBase(key)
		if !strings.HasSuffix(strings.ToLower(out.FileURL), "."+format) {
			out.FileURL += "." + format
		}
		_ = tx.QueryRow(ctx, `SELECT byte_size FROM cv_exports WHERE id=$1`, out.ID).Scan(&out.ByteSize)
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return domain.Export{}, commitErr
		}
		return out, nil
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return domain.Export{}, e
	}
	day := time.Now().In(tz()).Format("2006-01-02")
	var count int
	e = tx.QueryRow(ctx, `INSERT INTO daily_export_usage(user_id,export_date,export_count) VALUES($1,$2,1) ON CONFLICT(user_id,export_date) DO UPDATE SET export_count=daily_export_usage.export_count+1 WHERE daily_export_usage.export_count<5 RETURNING export_count`, uid, day).Scan(&count)
	if errors.Is(e, pgx.ErrNoRows) {
		return domain.Export{}, ErrDailyExportLimit
	}
	if e != nil {
		return domain.Export{}, e
	}
	var version string
	_ = tx.QueryRow(ctx, `SELECT id::text FROM cv_versions WHERE cv_id=$1 ORDER BY created_at DESC LIMIT 1`, cid).Scan(&version)
	exportID := uuid.New()
	filename := cv.ID + "_" + cv.UserID + "_" + format + "." + format
	_, e = tx.Exec(ctx, `INSERT INTO cv_exports(id,cv_id,cv_version_id,user_id,format,content_hash,file_url,object_key,credit_consumed,local_export_date,cached) VALUES($1,$2,NULLIF($3,'')::uuid,$4,$5,$6,$7,'',false,$8,false)`, exportID, cid, version, uid, format, hash, filename, day)
	if e != nil {
		return domain.Export{}, e
	}
	if e = tx.Commit(ctx); e != nil {
		return domain.Export{}, e
	}
	return domain.Export{ID: exportID.String(), CVID: cv.ID, VersionID: version, Format: format, ContentHash: hash, FileURL: filename, UserID: userID, CreatedAt: time.Now().UTC()}, nil
}
func (s *PostgresStore) CompleteExport(userID, exportID, objectKey string, size int64) error {
	uid, e := uuid.Parse(userID)
	if e != nil {
		return ErrNotFound
	}
	eid, e := uuid.Parse(exportID)
	if e != nil {
		return ErrNotFound
	}
	tag, e := s.pool.Exec(context.Background(), `UPDATE cv_exports SET object_key=$3,file_url=$3,byte_size=$4 WHERE id=$1 AND user_id=$2`, eid, uid, objectKey, size)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *PostgresStore) ExportForUser(userID, filename string) (domain.Export, error) {
	uid, e := uuid.Parse(userID)
	if e != nil {
		return domain.Export{}, ErrNotFound
	}
	var out domain.Export
	var key string
	e = s.pool.QueryRow(context.Background(), `SELECT id::text,cv_id::text,COALESCE(cv_version_id::text,''),format,content_hash,file_url,credit_consumed,user_id::text,cached,created_at,COALESCE(object_key,'') FROM cv_exports WHERE user_id=$1 AND (file_url=$2 OR file_url || '.' || format=$2 OR object_key=$2 OR regexp_replace(object_key, '^.*/', '')=$2) ORDER BY created_at DESC LIMIT 1`, uid, filename).Scan(&out.ID, &out.CVID, &out.VersionID, &out.Format, &out.ContentHash, &out.FileURL, &out.CreditConsumed, &out.UserID, &out.Cached, &out.CreatedAt, &key)
	if e != nil {
		return domain.Export{}, mapErr(e)
	}
	out.ObjectKey = key
	return out, nil
}
func (s *PostgresStore) addVersionTx(tx pgx.Tx, cv domain.CV, changedBy, summary string) error {
	cid, e := uuid.Parse(cv.ID)
	if e != nil {
		return e
	}
	raw, _ := json.Marshal(cv.Content)
	_, e = tx.Exec(context.Background(), `INSERT INTO cv_versions(cv_id,content,changed_by,change_summary) VALUES($1,$2,$3,$4)`, cid, raw, changedBy, summary)
	return e
}

type scanner interface{ Scan(...any) error }

func scanUser(row scanner) (domain.User, error) {
	var u domain.User
	var role string
	e := row.Scan(&u.ID, &u.ExternalSubject, &u.FullName, &u.Email, &u.Phone, &u.PreferredLanguage, &u.TargetRegion, &u.Plan, &u.ExportCredits, &u.AITokenBalance, &u.VoiceSecondsBalance, &u.ProfilePhotoURL, &role, &u.PhoneVerified, &u.CreatedAt, &u.UpdatedAt)
	u.Role = domain.Role(role)
	return u, mapErr(e)
}
func scanCV(row scanner) (domain.CV, error) {
	var c domain.CV
	var raw []byte
	e := row.Scan(&c.ID, &c.UserID, &c.Title, &c.DocumentStyle, &c.TemplateID, &c.TargetRegion, &c.Language, &raw, &c.ATSScore, &c.Archived, &c.CreatedAt, &c.UpdatedAt)
	if e != nil {
		return c, mapErr(e)
	}
	if e = json.Unmarshal(raw, &c.Content); e != nil {
		return c, e
	}
	return c, nil
}
func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
		return ErrConflict
	}
	return err
}
func tz() *time.Location {
	loc, e := time.LoadLocation("Africa/Dar_es_Salaam")
	if e != nil {
		return time.FixedZone("EAT", 3*60*60)
	}
	return loc
}
func cloneMap(m map[string]any) map[string]any {
	raw, _ := json.Marshal(m)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out
}

var _ Repository = (*PostgresStore)(nil)

func pathBase(value string) string {
	if i := strings.LastIndex(value, "/"); i >= 0 {
		return value[i+1:]
	}
	return value
}

func (s *PostgresStore) FailExport(userID, exportID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrNotFound
	}
	eid, err := uuid.Parse(exportID)
	if err != nil {
		return ErrNotFound
	}
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	var day time.Time
	err = tx.QueryRow(context.Background(), `DELETE FROM cv_exports WHERE id=$1 AND user_id=$2 AND object_key='' RETURNING local_export_date`, eid, uid).Scan(&day)
	if err != nil {
		return mapErr(err)
	}
	_, err = tx.Exec(context.Background(), `UPDATE daily_export_usage SET export_count=GREATEST(export_count-1,0) WHERE user_id=$1 AND export_date=$2`, uid, day)
	if err != nil {
		return err
	}
	return tx.Commit(context.Background())
}
