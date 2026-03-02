package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"vibecodepc/server/db"
	"vibecodepc/server/httputil"
	"vibecodepc/server/services"
)

// RegisterProjectRoutes registers all project and git API routes.
func RegisterProjectRoutes(r chi.Router) {
	r.Get("/api/projects", listProjects)
	r.Post("/api/projects", createProject)
	r.Get("/api/projects/{id}", getProject)
	r.Delete("/api/projects/{id}", deleteProject)

	r.Get("/api/projects/{id}/git/status", getGitStatus)
	r.Post("/api/projects/{id}/git/commit", postGitCommit)
	r.Post("/api/projects/{id}/git/push", postGitPush)
	r.Post("/api/projects/{id}/git/pull", postGitPull)
	r.Post("/api/projects/{id}/git/checkout", postGitCheckout)
	r.Get("/api/projects/{id}/git/diff", getGitDiff)
	r.Get("/api/projects/{id}/git/branches", getGitBranches)
}

// Project mirrors the projects table row.
type Project struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Path            string  `json:"path"`
	Language        *string `json:"language"`
	GithubURL       *string `json:"githubUrl"`
	GitRemote       *string `json:"gitRemote"`
	DefaultProvider *string `json:"defaultProvider"`
	CreatedAt       int64   `json:"createdAt"`
	LastOpenedAt    *int64  `json:"lastOpenedAt"`
}

// ProjectCreate holds the fields required to create a new project.
type ProjectCreate struct {
	Name            string  `json:"name"`
	Path            string  `json:"path"`
	Language        *string `json:"language"`
	GithubURL       *string `json:"githubUrl"`
	GitRemote       *string `json:"gitRemote"`
	DefaultProvider *string `json:"defaultProvider"`
}

func listProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB().QueryContext(r.Context(),
		`SELECT id, name, path, language, github_url, git_remote, default_provider, created_at, last_opened_at
         FROM projects ORDER BY COALESCE(last_opened_at, created_at) DESC`,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		p := scanProject(rows)
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []Project{}
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": projects})
}

func createProject(w http.ResponseWriter, r *http.Request) {
	var body ProjectCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" || body.Path == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name and path are required")
		return
	}

	id := uuid.New().String()
	now := time.Now().Unix()

	_, err := db.DB().ExecContext(r.Context(),
		`INSERT INTO projects (id, name, path, language, github_url, git_remote, default_provider, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, body.Name, body.Path, body.Language, body.GithubURL, body.GitRemote, body.DefaultProvider, now,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	p, err := fetchProject(r, id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to fetch created project")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"data": p})
}

func getProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := fetchProject(r, id)
	if err == sql.ErrNoRows {
		httputil.WriteError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to fetch project")
		return
	}

	// Update last_opened_at
	now := time.Now().Unix()
	_, _ = db.DB().ExecContext(r.Context(),
		`UPDATE projects SET last_opened_at = ? WHERE id = ?`, now, id,
	)
	p.LastOpenedAt = &now

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": p})
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := db.DB().ExecContext(r.Context(), `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete project")
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		httputil.WriteError(w, http.StatusNotFound, "project not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Git operations ---

func getGitStatus(w http.ResponseWriter, r *http.Request) {
	p, err := requireProject(w, r)
	if err != nil {
		return
	}
	status, err := services.Status(p.Path)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, status)
}

func postGitCommit(w http.ResponseWriter, r *http.Request) {
	p, err := requireProject(w, r)
	if err != nil {
		return
	}
	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Message == "" {
		httputil.WriteError(w, http.StatusBadRequest, "message is required")
		return
	}
	if err := services.Commit(p.Path, body.Message); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func postGitPush(w http.ResponseWriter, r *http.Request) {
	p, err := requireProject(w, r)
	if err != nil {
		return
	}
	if err := services.Push(p.Path); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func postGitPull(w http.ResponseWriter, r *http.Request) {
	p, err := requireProject(w, r)
	if err != nil {
		return
	}
	if err := services.Pull(p.Path); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func postGitCheckout(w http.ResponseWriter, r *http.Request) {
	p, err := requireProject(w, r)
	if err != nil {
		return
	}
	var body struct {
		Branch string `json:"branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Branch == "" {
		httputil.WriteError(w, http.StatusBadRequest, "branch is required")
		return
	}
	if err := services.Checkout(p.Path, body.Branch); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getGitDiff(w http.ResponseWriter, r *http.Request) {
	p, err := requireProject(w, r)
	if err != nil {
		return
	}
	diff, err := services.Diff(p.Path)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"diff": diff})
}

func getGitBranches(w http.ResponseWriter, r *http.Request) {
	p, err := requireProject(w, r)
	if err != nil {
		return
	}
	branches, err := services.Branches(p.Path)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if branches == nil {
		branches = []services.Branch{}
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"branches": branches})
}

// --- helpers ---

func requireProject(w http.ResponseWriter, r *http.Request) (Project, error) {
	id := chi.URLParam(r, "id")
	p, err := fetchProject(r, id)
	if err == sql.ErrNoRows {
		httputil.WriteError(w, http.StatusNotFound, "project not found")
		return Project{}, err
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to fetch project")
		return Project{}, err
	}
	return p, nil
}

func fetchProject(r *http.Request, id string) (Project, error) {
	var p Project
	err := db.DB().QueryRowContext(r.Context(),
		`SELECT id, name, path, language, github_url, git_remote, default_provider, created_at, last_opened_at
         FROM projects WHERE id = ?`, id,
	).Scan(
		&p.ID, &p.Name, &p.Path,
		&p.Language, &p.GithubURL, &p.GitRemote, &p.DefaultProvider,
		&p.CreatedAt, &p.LastOpenedAt,
	)
	return p, err
}

// rowScanner abstracts over *sql.Rows for scanProjectFromRows.
type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanProject(s rowScanner) Project {
	var p Project
	_ = s.Scan(
		&p.ID, &p.Name, &p.Path,
		&p.Language, &p.GithubURL, &p.GitRemote, &p.DefaultProvider,
		&p.CreatedAt, &p.LastOpenedAt,
	)
	return p
}
