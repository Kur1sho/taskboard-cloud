package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func chiRouter(db *sql.DB, jwtSecret string) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok"})
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
				writeErr(w, http.StatusInternalServerError, "Create task failed")
				return
			}
			writeJSON(w, http.StatusOK, created)
		}

		// Accept BOTH /tasks and /tasks/
		rt.Get("/", listHandler)
		rt.Get("", listHandler)

		rt.Post("/", createHandler)
		rt.Post("", createHandler)

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

			// Accept BOTH /tasks/{id} and /tasks/{id}/
			rid.Put("/", updateHandler)
			rid.Put("", updateHandler)

			rid.Delete("/", deleteHandler)
			rid.Delete("", deleteHandler)
		})
	})

	return r
}
