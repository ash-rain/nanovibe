package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/creack/pty"
)

// ProviderConfig holds an AI provider name and its API key.
type ProviderConfig struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// session represents a running opencode PTY session.
type session struct {
	pty *os.File
	cmd *exec.Cmd
}

var (
	sessionsMu sync.Mutex
	sessions   = make(map[string]*session)
)

// IsInstalled returns true if opencode is found in PATH.
func IsInstalled() bool {
	_, err := exec.LookPath("opencode")
	return err == nil
}

// Install runs `npm install -g opencode` and streams stdout/stderr lines.
func Install(ctx context.Context) <-chan string {
	ch := make(chan string, 64)
	go func() {
		defer close(ch)
		cmd := exec.CommandContext(ctx, "npm", "install", "-g", "opencode")
		pr, pw, err := os.Pipe()
		if err != nil {
			ch <- fmt.Sprintf("error: %v", err)
			return
		}
		cmd.Stdout = pw
		cmd.Stderr = pw

		if err := cmd.Start(); err != nil {
			pw.Close()
			pr.Close()
			ch <- fmt.Sprintf("error: start npm: %v", err)
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
			ch <- fmt.Sprintf("error: npm install failed: %v", err)
		} else {
			ch <- "done: opencode installed successfully"
		}
	}()
	return ch
}

// GetVersion returns the installed opencode version string.
func GetVersion() (string, error) {
	path, err := exec.LookPath("opencode")
	if err != nil {
		return "", fmt.Errorf("opencode: not found in PATH")
	}
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("opencode: --version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// opencodeConfig is the structure written to ~/.config/opencode/config.json.
type opencodeConfig struct {
	Providers map[string]opencodeProvider `json:"providers"`
}

type opencodeProvider struct {
	APIKey string `json:"apiKey"`
}

// WriteConfig writes the opencode configuration file with the given providers.
func WriteConfig(providers []ProviderConfig) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("opencode: home dir: %w", err)
	}
	cfgDir := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return fmt.Errorf("opencode: mkdir config dir: %w", err)
	}

	cfg := opencodeConfig{Providers: make(map[string]opencodeProvider)}
	for _, p := range providers {
		cfg.Providers[p.Name] = opencodeProvider{APIKey: p.Key}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("opencode: marshal config: %w", err)
	}

	cfgPath := filepath.Join(cfgDir, "config.json")
	if err := os.WriteFile(cfgPath, data, 0o600); err != nil {
		return fmt.Errorf("opencode: write config: %w", err)
	}
	return nil
}

// StartSession starts an opencode PTY session for the given project.
// Returns the PTY master file. If a session is already running for projectID it is returned.
func StartSession(projectID, cwd string) (*os.File, error) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if s, ok := sessions[projectID]; ok {
		return s.pty, nil
	}

	path, err := exec.LookPath("opencode")
	if err != nil {
		return nil, fmt.Errorf("opencode: not found in PATH")
	}

	cmd := exec.Command(path)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("opencode: start pty: %w", err)
	}

	sessions[projectID] = &session{pty: ptmx, cmd: cmd}

	// Clean up session when the process exits.
	go func() {
		_ = cmd.Wait()
		sessionsMu.Lock()
		delete(sessions, projectID)
		sessionsMu.Unlock()
	}()

	return ptmx, nil
}

// KillSession terminates the opencode session for the given project.
func KillSession(projectID string) error {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	s, ok := sessions[projectID]
	if !ok {
		return nil
	}
	if err := s.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("opencode: kill session %q: %w", projectID, err)
	}
	_ = s.pty.Close()
	delete(sessions, projectID)
	return nil
}

// ResizeSession resizes the PTY for the given project session.
func ResizeSession(projectID string, cols, rows uint16) error {
	sessionsMu.Lock()
	s, ok := sessions[projectID]
	sessionsMu.Unlock()

	if !ok {
		return fmt.Errorf("opencode: session %q not found", projectID)
	}
	return pty.Setsize(s.pty, &pty.Winsize{Cols: cols, Rows: rows})
}

// GetSession returns the PTY master for the given project session.
func GetSession(projectID string) (*os.File, bool) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	s, ok := sessions[projectID]
	if !ok {
		return nil, false
	}
	return s.pty, true
}
