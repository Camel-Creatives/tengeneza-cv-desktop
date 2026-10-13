package migrations

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.up.sql
var files embed.FS

func Apply(ctx context.Context, pool *pgxpool.Pool) error {
	// Serialize startup migrations across API replicas using a session-level advisory lock.
	lockConn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	if _, err = lockConn.Exec(ctx, `SELECT pg_advisory_lock(742_019_663_2026)`); err != nil {
		lockConn.Release()
		return err
	}
	defer func() {
		_, _ = lockConn.Exec(context.Background(), `SELECT pg_advisory_unlock(742_019_663_2026)`)
		lockConn.Release()
	}()
	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS app_schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`)
	if err != nil {
		return err
	}
	var hasFoundation bool
	if err = pool.QueryRow(ctx, `SELECT to_regclass('public.users') IS NOT NULL`).Scan(&hasFoundation); err != nil {
		return err
	}
	var tracked bool
	if err = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM app_schema_migrations WHERE version='000001_foundation')`).Scan(&tracked); err != nil {
		return err
	}
	if hasFoundation && !tracked {
		if _, err = pool.Exec(ctx, `INSERT INTO app_schema_migrations(version) VALUES('000001_foundation') ON CONFLICT DO NOTHING`); err != nil {
			return err
		}
	}
	entries, err := files.ReadDir(".")
	if err != nil {
		return err
	}
	versions := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			versions = append(versions, strings.TrimSuffix(name, ".up.sql"))
		}
	}
	sort.Strings(versions)
	for _, version := range versions {
		var applied bool
		if err = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM app_schema_migrations WHERE version=$1)`, version).Scan(&applied); err != nil {
			return err
		}
		if applied {
			continue
		}
		raw, readErr := files.ReadFile(version + ".up.sql")
		if readErr != nil {
			return readErr
		}
		tx, beginErr := pool.Begin(ctx)
		if beginErr != nil {
			return beginErr
		}
		if _, err = tx.Exec(ctx, string(raw)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO app_schema_migrations(version) VALUES($1)`, version); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}
