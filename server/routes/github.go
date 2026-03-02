package routes

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"vibecodepc/server/config"
	"vibecodepc/server/db"
	"vibecodepc/server/httputil"
	"vibecodepc/server/services"
)

// RegisterGitHubRoutes registers all GitHub API routes and OAuth handlers.
func RegisterGitHubRoutes(r chi.Router) {
	r.Get("/api/github/status", getGitHubStatus)
	r.Get("/api/github/repos", listGitHubRepos)
	r.Get("/api/github/repos/{owner}/{repo}/prs", listGitHubPRs)
	r.Post("/api/github/repos/{owner}/{repo}/prs", createGitHubPR)
	r.Get("/api/github/activity", getGitHubActivity)
	r.Post("/api/github/import", postGitHubImport)
	r.Delete("/api/github/disconnect", deleteGitHubDisconnect)

	r.Get("/auth/github/start", getGitHubAuthStart)
	r.Get("/auth/github/callback", getGitHubAuthCallback)
}

func getGitHubStatus(w http.ResponseWriter, r *http.Request) {
	if !services.IsAuthenticated() {
		httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"connected": false,
			"user":      nil,
		})
		return
	}

	user, err := services.GetUser(r.Context())
	if err != nil {
		httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"connected": false,
			"user":      nil,
		})
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"connected": true,
		"user":      user,
	})
}

func listGitHubRepos(w http.ResponseWriter, r *http.Request) {
	if !services.IsAuthenticated() {
		httputil.WriteError(w, http.StatusUnauthorized, "not authenticated with GitHub")
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	search := r.URL.Query().Get("search")

	repos, total, err := services.ListRepos(r.Context(), page, search)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if repos == nil {
		repos = []services.Repo{}
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"repos": repos,
		"total": total,
	})
}

func listGitHubPRs(w http.ResponseWriter, r *http.Request) {
	if !services.IsAuthenticated() {
		httputil.WriteError(w, http.StatusUnauthorized, "not authenticated with GitHub")
		return
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	prs, err := services.ListPRs(r.Context(), owner, repo)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if prs == nil {
		prs = []services.PR{}
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"prs": prs})
}

func createGitHubPR(w http.ResponseWriter, r *http.Request) {
	if !services.IsAuthenticated() {
		httputil.WriteError(w, http.StatusUnauthorized, "not authenticated with GitHub")
		return
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	var params services.PRCreate
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if params.Title == "" || params.Head == "" || params.Base == "" {
		httputil.WriteError(w, http.StatusBadRequest, "title, head, and base are required")
		return
	}

	pr, err := services.CreatePR(r.Context(), owner, repo, params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"pr": pr})
}

func getGitHubActivity(w http.ResponseWriter, r *http.Request) {
	if !services.IsAuthenticated() {
		httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"events": []interface{}{}})
		return
	}

	var login string
	_ = db.DB().QueryRowContext(r.Context(), `SELECT COALESCE(login, '') FROM github_auth WHERE id = 1`).Scan(&login)

	if login == "" {
		user, err := services.GetUser(r.Context())
		if err != nil {
			httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"events": []interface{}{}})
			return
		}
		login = user.Login
	}

	events, err := services.GetActivity(r.Context(), login)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if events == nil {
		events = []services.GitHubEvent{}
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"events": events})
}

func postGitHubImport(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RepoURL string `json:"repoUrl"`
		Name    string `json:"name"`
		Path    string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.RepoURL == "" || body.Path == "" {
		httputil.WriteError(w, http.StatusBadRequest, "repoUrl and path are required")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	logCh := services.ImportRepo(r.Context(), body.RepoURL, body.Path)
	name := body.Name
	if name == "" {
		// Derive name from URL
		parts := strings.Split(strings.TrimSuffix(body.RepoURL, ".git"), "/")
		name = parts[len(parts)-1]
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case line, open := <-logCh:
			if !open {
				// Create project record after clone completes
				id := uuid.New().String()
				now := time.Now().Unix()
				_, _ = db.DB().ExecContext(r.Context(),
					`INSERT INTO projects (id, name, path, github_url, created_at) VALUES (?, ?, ?, ?, ?)
                     ON CONFLICT(path) DO UPDATE SET name = excluded.name, github_url = excluded.github_url`,
					id, name, body.Path, body.RepoURL, now,
				)
				fmt.Fprintf(w, "data: {\"projectId\":\"%s\"}\n\n", id)
				flusher.Flush()
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", line)
			flusher.Flush()
		}
	}
}

func deleteGitHubDisconnect(w http.ResponseWriter, r *http.Request) {
	if err := services.Del("github_token"); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to remove token")
		return
	}
	_, _ = db.DB().ExecContext(r.Context(),
		`UPDATE github_auth SET login = NULL, avatar_url = NULL, public_repos = 0, connected_at = NULL WHERE id = 1`,
	)
	w.WriteHeader(http.StatusNoContent)
}

// --- OAuth handlers ---

var (
	oauthStateMu sync.Mutex
	oauthStates  = make(map[string]time.Time)
)

func getGitHubAuthStart(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	if cfg.GitHubClientID == "" {
		httputil.WriteError(w, http.StatusServiceUnavailable, "GitHub OAuth not configured")
		return
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to generate state")
		return
	}
	state := hex.EncodeToString(b)

	oauthStateMu.Lock()
	oauthStates[state] = time.Now()
	oauthStateMu.Unlock()

	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=repo,read:user&state=%s",
		cfg.GitHubClientID,
		url.QueryEscape(cfg.GitHubRedirectURI),
		state,
	)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func getGitHubAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if !consumeOAuthState(state) {
		httputil.WriteError(w, http.StatusBadRequest, "invalid or expired OAuth state")
		return
	}
	if code == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing OAuth code")
		return
	}

	cfg := config.Get()
	token, err := exchangeGitHubCode(r.Context(), cfg.GitHubClientID, cfg.GitHubClientSecret, code, cfg.GitHubRedirectURI)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to exchange OAuth code")
		return
	}

	if err := services.Set("github_token", token); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to store token")
		return
	}

	// Fetch and cache user info
	user, err := services.GetUser(r.Context())
	if err == nil {
		now := time.Now().Unix()
		_, _ = db.DB().ExecContext(r.Context(),
			`UPDATE github_auth SET login = ?, avatar_url = ?, public_repos = ?, connected_at = ? WHERE id = 1`,
			user.Login, user.AvatarURL, user.PublicRepos, now,
		)
		_ = services.Set("github_user_login", user.Login)
		_ = services.Set("github_user_avatar", user.AvatarURL)
	}

	http.Redirect(w, r, "/app/dashboard", http.StatusFound)
}

func consumeOAuthState(state string) bool {
	oauthStateMu.Lock()
	defer oauthStateMu.Unlock()
	t, ok := oauthStates[state]
	if !ok {
		return false
	}
	if time.Since(t) > 10*time.Minute {
		delete(oauthStates, state)
		return false
	}
	delete(oauthStates, state)
	return true
}

// exchangeGitHubCode exchanges an OAuth authorization code for an access token.
func exchangeGitHubCode(ctx context.Context, clientID, clientSecret, code, redirectURI string) (string, error) {
	vals := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(vals.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("oauth exchange: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("oauth exchange: parse response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("oauth exchange: %s: %s", result.Error, result.Description)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("oauth exchange: empty access token")
	}
	return result.AccessToken, nil
}

