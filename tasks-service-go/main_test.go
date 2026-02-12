package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const testSecret = "ci-test-secret"

func makeToken(email string) string {
	claims := jwt.MapClaims{
		"sub": email,
		"exp": time.Now().Add(60 * time.Minute).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := t.SignedString([]byte(testSecret))
	return s
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	os.Setenv("JWT_SECRET", testSecret)

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://taskboard:taskboard@localhost:5432/taskboard_test?sslmode=disable"
	}
	dsn = normalizePostgresDSN(dsn)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := waitForDB(db, 30*time.Second); err != nil {
		t.Fatalf("db not ready: %v", err)
	}
	if err := ensureSchema(db); err != nil {
		t.Fatalf("schema: %v", err)
	}

	_, _ = db.Exec(`TRUNCATE TABLE tasks RESTART IDENTITY CASCADE;`)

	jwtSecret := os.Getenv("JWT_SECRET")
	r := buildRouterForTest(db, jwtSecret)

	return httptest.NewServer(r)
}

func buildRouterForTest(db *sql.DB, jwtSecret string) http.Handler {
	mux := chiRouter(db, jwtSecret)
	return mux
}
