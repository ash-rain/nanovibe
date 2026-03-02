package services

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"vibecodepc/server/config"
)

// TunnelMode describes how the cloudflare tunnel is running.
type TunnelMode string

const (
	ModeNone  TunnelMode = "none"
	ModeQuick TunnelMode = "quick"
	ModeNamed TunnelMode = "named"
)

// TunnelStatus holds the current state of the cloudflare tunnel.
type TunnelStatus struct {
	Mode      TunnelMode `json:"mode"`
	Connected bool       `json:"connected"`
	TunnelURL string     `json:"tunnelUrl"`
	LocalURL  string     `json:"localUrl"`
	UptimeS   int64      `json:"uptimeS"`
}

var (
	tunnelMu      sync.Mutex
	tunnelCmd     *exec.Cmd
	tunnelMode    TunnelMode = ModeNone
	tunnelURL     string
	tunnelStarted time.Time
)

// StartQuickTunnel starts cloudflared in quick-tunnel mode and parses the assigned URL
// from its output. The URL is written to stderr by cloudflared.
func StartQuickTunnel() error {
	tunnelMu.Lock()
	defer tunnelMu.Unlock()

	if tunnelCmd != nil {
		return nil // already running
	}

	cfg := config.Get()
	target := fmt.Sprintf("http://localhost:%s", cfg.Port)

	cmd := exec.Command("cloudflared", "tunnel", "--url", target)

	// Use os.Pipe so we get both stdout and stderr in one reader.
	pr, pw, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("cloudflare: pipe: %w", err)
	}
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return fmt.Errorf("cloudflare: start: %w", err)
	}
	pw.Close() // write end belongs to child process now

	tunnelCmd = cmd
	tunnelMode = ModeQuick
	tunnelStarted = time.Now()

	go parseTunnelOutput(pr)

	return nil
}

// StartNamedTunnel starts cloudflared with the given tunnel token.
func StartNamedTunnel(token string) error {
	tunnelMu.Lock()
	defer tunnelMu.Unlock()

	if tunnelCmd != nil {
		_ = tunnelCmd.Process.Kill()
		tunnelCmd = nil
		tunnelURL = ""
	}

	cmd := exec.Command("cloudflared", "tunnel", "--no-autoupdate", "run", "--token", token)

	pr, pw, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("cloudflare: pipe: %w", err)
	}
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return fmt.Errorf("cloudflare: start named: %w", err)
	}
	pw.Close()

	tunnelCmd = cmd
	tunnelMode = ModeNamed
	tunnelStarted = time.Now()

	go parseTunnelOutput(pr)

	return nil
}

// parseTunnelOutput reads combined output from cloudflared and extracts the tunnel URL.
func parseTunnelOutput(pr *os.File) {
	defer pr.Close()
	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		line := scanner.Text()
		if url := findTunnelURL(line); url != "" {
			tunnelMu.Lock()
			tunnelURL = url
			tunnelMu.Unlock()
		}
	}
}

// findTunnelURL extracts a cloudflare URL from a log line.
func findTunnelURL(line string) string {
	for _, word := range strings.Fields(line) {
		word = strings.Trim(word, "|\"'")
		if strings.Contains(word, "trycloudflare.com") {
			if !strings.HasPrefix(word, "http") {
				word = "https://" + word
			}
			return word
		}
		if strings.HasPrefix(word, "https://") && strings.Contains(word, ".cloudflare") {
			return word
		}
	}
	return ""
}

// StopTunnel kills the running cloudflared process.
func StopTunnel() error {
	tunnelMu.Lock()
	defer tunnelMu.Unlock()

	if tunnelCmd == nil {
		return nil
	}
	if err := tunnelCmd.Process.Kill(); err != nil {
		return fmt.Errorf("cloudflare: kill: %w", err)
	}
	tunnelCmd = nil
	tunnelMode = ModeNone
	tunnelURL = ""
	return nil
}

// GetStatus returns the current tunnel status.
func GetStatus() TunnelStatus {
	tunnelMu.Lock()
	defer tunnelMu.Unlock()

	cfg := config.Get()
	localURL := fmt.Sprintf("http://localhost:%s", cfg.Port)

	if tunnelCmd == nil {
		return TunnelStatus{
			Mode:     ModeNone,
			LocalURL: localURL,
		}
	}

	var uptimeS int64
	if !tunnelStarted.IsZero() {
		uptimeS = int64(time.Since(tunnelStarted).Seconds())
	}

	return TunnelStatus{
		Mode:      tunnelMode,
		Connected: tunnelURL != "",
		TunnelURL: tunnelURL,
		LocalURL:  localURL,
		UptimeS:   uptimeS,
	}
}

// ValidateToken verifies that the cloudflare tunnel token can start a tunnel.
func ValidateToken(ctx context.Context, token string) (bool, string, error) {
	if token == "" {
		return false, "", fmt.Errorf("cloudflare: empty token")
	}
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "--no-autoupdate", "run", "--token", token, "--dry-run")
	out, err := cmd.CombinedOutput()
	output := string(out)

	if err != nil {
		if strings.Contains(output, "tunnel") || strings.Contains(output, "Successfully") {
			return true, findTunnelURL(output), nil
		}
		return false, "", fmt.Errorf("cloudflare: validate token: %w", err)
	}
	return true, findTunnelURL(output), nil
}
