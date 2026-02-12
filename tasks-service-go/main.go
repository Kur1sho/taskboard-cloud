package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	chicors "github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type ctxKey string

const userEmailKey ctxKey = "userEmail"

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type TaskUpdate struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-me"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	dsn = normalizePostgresDSN(dsn)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := waitForDB(db, 30*time.Second); err != nil {
		log.Fatalf("db not ready: %v", err)
	}

	if err := ensureSchema(db); err != nil {
		log.Fatalf("ensure schema: %v", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.StripSlashes)
	r.Use(middleware.Logger)

	origins := strings.Split(getEnv("CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"), ",")
	r.Use(chicors.Handler(chicors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/tasks", func(rt chi.Router) {
		rt.Use(authMiddleware(jwtSecret))

		listHandler := func(w http.ResponseWriter, r *http.Request) {
			email := mustUserEmail(r.Context())
			tasks, err := listTasks(r.Context(), db, email)
			if err != nil {
				writeErr(w, http.StatusInternalServerError, "Failed to load tasks")
				return
			}
			writeJSON(w, http.StatusOK, tasks)
		}

		createHandler := func(w http.ResponseWriter, r *http.Request) {
			email := mustUserEmail(r.Context())
			title := strings.TrimSpace(r.URL.Query().Get("title"))
			if title == "" {
				writeErr(w, http.StatusBadRequest, "Task title cannot be empty")
				return
			}
			created, err := createTask(r.Context(), db, email, title)
			if err != nil {
				log.Printf("createTask failed: %v (email=%q title=%q)", err, email, title)
				writeErr(w, http.StatusInternalServerError, "Create task failed")
				return
			}
			writeJSON(w, http.StatusOK, created)
		}

		rt.Get("/", listHandler)
		rt.Post("/", createHandler)

		rt.Route("/{taskID}", func(rid chi.Router) {
			updateHandler := func(w http.ResponseWriter, r *http.Request) {
				email := mustUserEmail(r.Context())
				id, err := strconv.Atoi(chi.URLParam(r, "taskID"))
				if err != nil {
					writeErr(w, http.StatusBadRequest, "Invalid task id")
					return
				}

				var patch TaskUpdate
				if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
					writeErr(w, http.StatusBadRequest, "Invalid JSON")
					return
				}

				if patch.Title != nil {
					t := strings.TrimSpace(*patch.Title)
					if t == "" {
						writeErr(w, http.StatusBadRequest, "Title cannot be empty")
						return
					}
					patch.Title = &t
				}

				updated, err := updateTask(r.Context(), db, email, id, patch)
				if errors.Is(err, sql.ErrNoRows) {
					writeErr(w, http.StatusNotFound, "Task not found")
					return
				}
				if err != nil {
					writeErr(w, http.StatusInternalServerError, "Update task failed")
					return
				}
				writeJSON(w, http.StatusOK, updated)
			}

			deleteHandler := func(w http.ResponseWriter, r *http.Request) {
				email := mustUserEmail(r.Context())
				id, err := strconv.Atoi(chi.URLParam(r, "taskID"))
				if err != nil {
					writeErr(w, http.StatusBadRequest, "Invalid task id")
					return
				}

				err = deleteTask(r.Context(), db, email, id)
				if errors.Is(err, sql.ErrNoRows) {
					writeErr(w, http.StatusNotFound, "Task not found")
					return
				}
				if err != nil {
					writeErr(w, http.StatusInternalServerError, "Delete task failed")
					return
				}
				writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
			}

			rid.Put("/", updateHandler)
			rid.Delete("/", deleteHandler)
		})
	})

	addr := getEnv("ADDR", ":8000")
	log.Printf("tasks-service-go listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func getEnv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
func normalizePostgresDSN(dsn string) string {
	if strings.HasPrefix(dsn, "postgresql+psycopg://") {
		dsn = strings.Replace(dsn, "postgresql+psycopg://", "postgres://", 1)
	}
	// local/dev/CI: force sslmode=disable if not specified
	if !strings.Contains(dsn, "sslmode=") {
		if strings.Contains(dsn, "?") {
			dsn += "&sslmode=disable"
		} else {
			dsn += "?sslmode=disable"
		}
	}
	return dsn
}

func waitForDB(db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := db.Ping(); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout after %s", timeout)
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS tasks (
  id SERIAL PRIMARY KEY,
  owner_email VARCHAR(320) NOT NULL,
  title VARCHAR(200) NOT NULL,
  done BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_tasks_owner_email ON tasks(owner_email);
`)
	return err
}

func listTasks(ctx context.Context, db *sql.DB, owner string) ([]Task, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, title, done
FROM tasks
WHERE owner_email = $1
ORDER BY id DESC
`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func createTask(ctx context.Context, db *sql.DB, owner, title string) (Task, error) {
	var t Task
	err := db.QueryRowContext(ctx, `
INSERT INTO tasks (owner_email, title, done, created_at)
VALUES ($1, $2, FALSE, NOW())
RETURNING id, title, done
`, owner, title).Scan(&t.ID, &t.Title, &t.Done)
	return t, err
}

func updateTask(ctx context.Context, db *sql.DB, owner string, id int, patch TaskUpdate) (Task, error) {
	// fetch existing (and enforce owner)
	var cur Task
	err := db.QueryRowContext(ctx, `
SELECT id, title, done
FROM tasks
WHERE id = $1 AND owner_email = $2
`, id, owner).Scan(&cur.ID, &cur.Title, &cur.Done)
	if err != nil {
		return Task{}, err
	}

	newTitle := cur.Title
	newDone := cur.Done

	if patch.Title != nil {
		newTitle = *patch.Title
	}
	if patch.Done != nil {
		newDone = *patch.Done
	}

	var out Task
	err = db.QueryRowContext(ctx, `
UPDATE tasks
SET title = $1, done = $2
WHERE id = $3 AND owner_email = $4
RETURNING id, title, done
`, newTitle, newDone, id, owner).Scan(&out.ID, &out.Title, &out.Done)
	return out, err
}

func deleteTask(ctx context.Context, db *sql.DB, owner string, id int) error {
	res, err := db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1 AND owner_email = $2`, id, owner)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func authMiddleware(secret string) func(http.Handler) http.Handler {
	key := []byte(secret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				writeErr(w, http.StatusUnauthorized, "Invalid token")
				return
			}
			tokenStr := strings.TrimPrefix(h, "Bearer ")

			tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, fmt.Errorf("unexpected alg: %s", t.Method.Alg())
				}
				return key, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
			if err != nil || !tok.Valid {
				writeErr(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			claims, ok := tok.Claims.(jwt.MapClaims)
			if !ok {
				writeErr(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			sub, _ := claims["sub"].(string)
			if sub == "" {
				writeErr(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), userEmailKey, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func mustUserEmail(ctx context.Context) string {
	v := ctx.Value(userEmailKey)
	s, _ := v.(string)
	return s
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"detail": msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
