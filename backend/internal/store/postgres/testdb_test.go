package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"testing"
)

func isolatedDatabaseURL(t *testing.T, databaseURL string) string {
	t.Helper()
	schema := testSchemaName(t)
	ctx := context.Background()
	db, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA %s`, schema)); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupDB, err := Open(context.Background(), databaseURL)
		if err != nil {
			t.Fatalf("open postgres for cleanup: %v", err)
		}
		defer cleanupDB.Close()
		if _, err := cleanupDB.ExecContext(context.Background(), fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, schema)); err != nil {
			t.Fatalf("drop test schema: %v", err)
		}
	})
	return databaseURLWithSearchPath(t, databaseURL, schema)
}

func databaseURLWithSearchPath(t *testing.T, databaseURL string, schema string) string {
	t.Helper()
	parsed, err := url.Parse(databaseURL)
	if err != nil || parsed.Scheme == "" {
		t.Fatalf("DATABASE_URL must be a postgres URL for isolated tests, got %q", databaseURL)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func testSchemaName(t *testing.T) string {
	t.Helper()
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatalf("generate schema suffix: %v", err)
	}
	return "test_" + hex.EncodeToString(suffix[:])
}
