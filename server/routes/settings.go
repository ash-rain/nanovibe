package routes

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"vibecodepc/server/config"
	"vibecodepc/server/httputil"
	"vibecodepc/server/services"
)

// appVersion is the application version, overridden at build time via -ldflags.
var appVersion = "dev"

// RegisterSettingsRoutes registers all settings API routes.
func RegisterSettingsRoutes(r chi.Router) {
	r.Get("/api/settings/providers", getProviders)
	r.Post("/api/settings/providers", postProviders)
	r.Get("/api/settings/tunnel", getTunnel)
	r.Post("/api/settings/tunnel/restart", postTunnelRestart)
	r.Post("/api/settings/tunnel/upgrade", postTunnelUpgrade)
	r.Get("/api/settings/system", getSystemInfo)
	r.Post("/api/settings/system/update", postSystemUpdate)
}

// ProviderConfigResponse is the masked provider config returned to the client.
type ProviderConfigResponse struct {
	Name   string `json:"name"`
	Key    string `json:"key"`
	Exists bool   `json:"exists"`
}

func getProviders(w http.ResponseWriter, r *http.Request) {
	providerKeys := []struct {
		name    string
		keyName string
	}{
		{"anthropic", "anthropic_key"},
		{"openai", "openai_key"},
		{"google", "google_key"},
		{"ollama", "ollama_key"},
	}

	var providers []ProviderConfigResponse
	for _, pk := range providerKeys {
		val, ok := services.Get(pk.keyName)
		masked := ""
		if ok && val != "" {
			masked = services.Mask(val)
		}
		providers = append(providers, ProviderConfigResponse{
			Name:   pk.name,
			Key:    masked,
			Exists: ok && val != "",
		})
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"providers": providers})
}

func postProviders(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Provider string `json:"provider"`
		Key      string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Provider == "" {
		httputil.WriteError(w, http.StatusBadRequest, "provider is required")
		return
	}

	keyName := body.Provider + "_key"
	if body.Key == "" {
		if err := services.Del(keyName); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to remove key")
			return
		}
	} else {
		if err := services.Set(keyName, body.Key); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to save key")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func getTunnel(w http.ResponseWriter, r *http.Request) {
	status := services.GetStatus()
	httputil.WriteJSON(w, http.StatusOK, status)
}

func postTunnelRestart(w http.ResponseWriter, r *http.Request) {
	_ = services.StopTunnel()
	time.Sleep(500 * time.Millisecond)
	if err := services.StartQuickTunnel(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to restart tunnel: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func postTunnelUpgrade(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" {
		httputil.WriteError(w, http.StatusBadRequest, "token is required")
		return
	}

	if err := services.Set("cf_tunnel_token", body.Token); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to save token")
		return
	}

	_ = services.StopTunnel()
	time.Sleep(500 * time.Millisecond)

	if err := services.StartNamedTunnel(body.Token); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to start named tunnel: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getSystemInfo(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	hostname, _ := os.Hostname()
	localIP := getLocalIP()

	dockerVer := getDockerVersion()

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"hostname":      hostname,
		"ip":            localIP,
		"localUrl":      fmt.Sprintf("http://%s.local:%s", hostname, cfg.Port),
		"goVersion":     runtime.Version(),
		"dockerVersion": dockerVer,
		"appVersion":    appVersion,
	})
}

func postSystemUpdate(w http.ResponseWriter, r *http.Request) {
	ch := runUpdateScript(r.Context())
	streamSSE(w, r, ch)
}

// runUpdateScript streams the output of the update.sh script.
func runUpdateScript(ctx context.Context) <-chan string {
	ch := make(chan string, 64)
	go func() {
		defer close(ch)

		home, _ := os.UserHomeDir()
		updateScript := home + "/.vibecodepc/update.sh"

		if _, err := os.Stat(updateScript); err != nil {
			ch <- "error: update script not found at " + updateScript
			return
		}

		cmd := exec.CommandContext(ctx, "sh", updateScript)
		pr, pw, err := os.Pipe()
		if err != nil {
			ch <- fmt.Sprintf("error: pipe: %v", err)
			return
		}
		cmd.Stdout = pw
		cmd.Stderr = pw

		if err := cmd.Start(); err != nil {
			pw.Close()
			pr.Close()
			ch <- fmt.Sprintf("error: start update: %v", err)
			return
		}
		pw.Close()

		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				pr.Close()
				return
			case ch <- scanner.Text():
			}
		}
		pr.Close()

		if err := cmd.Wait(); err != nil {
			ch <- "error: update failed: " + err.Error()
		} else {
			ch <- "done: update complete"
		}
	}()
	return ch
}

// getLocalIP returns the first non-loopback local IPv4 address.
func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return "127.0.0.1"
}

// getDockerVersion returns the Docker version string or empty string.
func getDockerVersion() string {
	out, err := exec.Command("docker", "--version").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
