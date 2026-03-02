package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"

	"vibecodepc/server/db"
	"vibecodepc/server/httputil"
	"vibecodepc/server/services"
)

// RegisterSetupRoutes registers all setup wizard API routes.
func RegisterSetupRoutes(r chi.Router) {
	r.Get("/api/setup/state", getSetupState)
	r.Post("/api/setup/state", postSetupState)
	r.Get("/api/setup/check/system", getSystemChecks)
	r.Post("/api/setup/fix/{checkId}", postFixCheck)
	r.Post("/api/setup/cloudflare/validate", postCloudflareValidate)
	r.Post("/api/setup/providers/detect", postProvidersDetect)
	r.Post("/api/setup/providers/test", postProvidersTest)
	r.Post("/api/setup/opencode/install", postOpencodeInstall)
	r.Post("/api/setup/nanoclaw/setup", postNanoclawSetup)
}

type setupStateResponse struct {
	CurrentStep    string   `json:"currentStep"`
	CompletedSteps []string `json:"completedSteps"`
}

func getSetupState(w http.ResponseWriter, r *http.Request) {
	var (
		step      string
		stepsJSON string
		updatedAt int64
	)
	err := db.DB().QueryRowContext(r.Context(),
		`SELECT current_step, completed_steps, updated_at FROM setup_state WHERE id = 1`,
	).Scan(&step, &stepsJSON, &updatedAt)
	if err == sql.ErrNoRows {
		httputil.WriteJSON(w, http.StatusOK, setupStateResponse{
			CurrentStep:    "welcome",
			CompletedSteps: []string{},
		})
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read setup state")
		return
	}

	var completed []string
	if err := json.Unmarshal([]byte(stepsJSON), &completed); err != nil {
		completed = []string{}
	}

	httputil.WriteJSON(w, http.StatusOK, setupStateResponse{
		CurrentStep:    step,
		CompletedSteps: completed,
	})
}

func postSetupState(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Step string `json:"step"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Step == "" {
		httputil.WriteError(w, http.StatusBadRequest, "step is required")
		return
	}

	// Fetch current state
	var (
		currentStep string
		stepsJSON   string
	)
	_ = db.DB().QueryRowContext(r.Context(),
		`SELECT current_step, completed_steps FROM setup_state WHERE id = 1`,
	).Scan(&currentStep, &stepsJSON)

	var completed []string
	if stepsJSON != "" {
		_ = json.Unmarshal([]byte(stepsJSON), &completed)
	}

	// Add current step to completed if advancing to a new step
	if currentStep != "" && currentStep != body.Step {
		found := false
		for _, s := range completed {
			if s == currentStep {
				found = true
				break
			}
		}
		if !found {
			completed = append(completed, currentStep)
		}
	}

	newStepsJSON, _ := json.Marshal(completed)
	now := time.Now().Unix()

	_, err := db.DB().ExecContext(r.Context(),
		`UPDATE setup_state SET current_step = ?, completed_steps = ?, updated_at = ? WHERE id = 1`,
		body.Step, string(newStepsJSON), now,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update setup state")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getSystemChecks(w http.ResponseWriter, r *http.Request) {
	checks, err := services.RunAllChecks(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"checks": checks})
}

func postFixCheck(w http.ResponseWriter, r *http.Request) {
	checkID := chi.URLParam(r, "checkId")

	var logCh <-chan string
	switch services.CheckID(checkID) {
	case services.CheckDocker:
		logCh = services.FixDocker(r.Context())
	case services.CheckDockerDaemon:
		logCh = services.FixDockerDaemon(r.Context())
	case services.CheckGit:
		logCh = services.FixGit(r.Context())
	default:
		httputil.WriteError(w, http.StatusBadRequest, fmt.Sprintf("check %q is not fixable", checkID))
		return
	}

	streamSSE(w, r, logCh)
}

func postCloudflareValidate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Token == "" {
		httputil.WriteError(w, http.StatusBadRequest, "token is required")
		return
	}

	valid, tunnelURL, err := services.ValidateToken(r.Context(), body.Token)
	if err != nil {
		httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"valid":     false,
			"tunnelUrl": "",
		})
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"valid":     valid,
		"tunnelUrl": tunnelURL,
	})
}

func postProvidersDetect(w http.ResponseWriter, r *http.Request) {
	type detectedProvider struct {
		Provider string `json:"provider"`
		Key      string `json:"key"` // masked
	}

	envKeys := map[string]string{
		"ANTHROPIC_API_KEY": "anthropic",
		"OPENAI_API_KEY":    "openai",
		"GOOGLE_API_KEY":    "google",
	}

	var found []detectedProvider
	for envVar, provider := range envKeys {
		if val := os.Getenv(envVar); val != "" {
			found = append(found, detectedProvider{
				Provider: provider,
				Key:      services.Mask(val),
			})
		}
	}

	if found == nil {
		found = []detectedProvider{}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"detected": found})
}

func postProvidersTest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Provider string `json:"provider"`
		Key      string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Provider == "" || body.Key == "" {
		httputil.WriteError(w, http.StatusBadRequest, "provider and key are required")
		return
	}

	valid, detail := validateProviderKey(body.Provider, body.Key)
	if !valid {
		httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"valid":  false,
			"detail": detail,
		})
		return
	}

	if err := services.Set(body.Provider+"_key", body.Key); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to save key")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"valid":  true,
		"detail": detail,
	})
}

func postOpencodeInstall(w http.ResponseWriter, r *http.Request) {
	logCh := services.Install(r.Context())
	streamSSE(w, r, logCh)
}

func postNanoclawSetup(w http.ResponseWriter, r *http.Request) {
	logCh := services.CloneNanoclaw(r.Context())
	streamSSE(w, r, logCh)
}

// streamSSE writes a channel of strings as Server-Sent Events.
func streamSSE(w http.ResponseWriter, r *http.Request, ch <-chan string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case line, open := <-ch:
			if !open {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", line)
			flusher.Flush()
		}
	}
}

// validateProviderKey performs simple format validation for AI provider keys.
func validateProviderKey(provider, key string) (bool, string) {
	switch provider {
	case "anthropic":
		if len(key) < 10 {
			return false, "Anthropic API key appears too short"
		}
		return true, "Anthropic key format looks valid"
	case "openai":
		if len(key) < 10 {
			return false, "OpenAI API key appears too short"
		}
		return true, "OpenAI key format looks valid"
	case "google":
		if len(key) < 10 {
			return false, "Google API key appears too short"
		}
		return true, "Google key format looks valid"
	default:
		if len(key) < 5 {
			return false, "Key appears too short"
		}
		return true, "Key format looks valid"
	}
}
