package services

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// nanoclawDir returns the path where nanoclaw is cloned.
func nanoclawDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".vibecodepc", "nanoclaw")
}

// nanoclawDBPath returns the path to nanoclaw's SQLite database.
func nanoclawDBPath() string {
	return filepath.Join(nanoclawDir(), "data", "nanoclaw.db")
}

// IsCloned returns true if nanoclaw has been cloned to the expected directory.
func IsCloned() bool {
	_, err := os.Stat(filepath.Join(nanoclawDir(), ".git"))
	return err == nil
}

// Clone clones the nanoclaw repository and streams progress lines.
// CloneNanoclaw clones the nanoclaw repository and streams progress lines.
func CloneNanoclaw(ctx context.Context) <-chan string {
	ch := make(chan string, 32)
	go func() {
		defer close(ch)
		repoURL := "https://github.com/qwibitai/nanoclaw.git"
		destPath := nanoclawDir()

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			ch <- fmt.Sprintf("error: mkdir: %v", err)
			return
		}

		cmd := exec.CommandContext(ctx, "git", "clone", "--progress", repoURL, destPath)
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
			ch <- fmt.Sprintf("error: clone start: %v", err)
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
			ch <- fmt.Sprintf("error: clone failed: %v", err)
			return
		}

		ch <- "cloned: applying patch..."
		if err := applyPatch(ctx); err != nil {
			ch <- fmt.Sprintf("warning: patch failed: %v", err)
		}
		ch <- "done"
	}()
	return ch
}

// applyPatch applies the web bridge patch after cloning.
func applyPatch(ctx context.Context) error {
	home, _ := os.UserHomeDir()
	patchPath := filepath.Join(home, "Documents", "_DEV", "nanovibe", "scripts", "nanoclaw-web-bridge.patch")
	if _, err := os.Stat(patchPath); err != nil {
		return fmt.Errorf("patch file not found: %w", err)
	}
	cmd := exec.CommandContext(ctx, "git", "apply", patchPath)
	cmd.Dir = nanoclawDir()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("patch: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// IsRunning returns true if the nanoclaw Docker container is running.
func IsRunning(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "--filter", "name=nanoclaw", "--format", "{{.Names}}")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("nanoclaw: docker ps: %w", err)
	}
	return strings.Contains(string(out), "nanoclaw"), nil
}

// Start starts the nanoclaw Docker container via docker compose.
func Start(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "up", "-d")
	cmd.Dir = nanoclawDir()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nanoclaw: docker compose up: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Stop stops the nanoclaw Docker container.
func Stop(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "down")
	cmd.Dir = nanoclawDir()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nanoclaw: docker compose down: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// WriteEnv writes provider keys to nanoclaw's .env file.
func WriteEnv(providers []ProviderConfig) error {
	var sb strings.Builder
	for _, p := range providers {
		switch strings.ToLower(p.Name) {
		case "anthropic":
			sb.WriteString(fmt.Sprintf("ANTHROPIC_API_KEY=%s\n", p.Key))
		case "openai":
			sb.WriteString(fmt.Sprintf("OPENAI_API_KEY=%s\n", p.Key))
		case "google":
			sb.WriteString(fmt.Sprintf("GOOGLE_API_KEY=%s\n", p.Key))
		default:
			sb.WriteString(fmt.Sprintf("%s_API_KEY=%s\n", strings.ToUpper(p.Name), p.Key))
		}
	}
	envPath := filepath.Join(nanoclawDir(), ".env")
	return os.WriteFile(envPath, []byte(sb.String()), 0o600)
}

// AgentMessage represents a message in the nanoclaw agent conversation.
type AgentMessage struct {
	ID        int64  `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"createdAt"`
	ProjectID string `json:"projectId"`
}

// openNanoclawDB opens a read-only connection to nanoclaw's SQLite database.
func openNanoclawDB() (*sql.DB, error) {
	dbPath := nanoclawDBPath()
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("nanoclaw: db not found: %w", err)
	}
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("nanoclaw: open db: %w", err)
	}
	return db, nil
}

// openNanoclawDBRW opens a read-write connection to nanoclaw's SQLite database.
func openNanoclawDBRW() (*sql.DB, error) {
	dbPath := nanoclawDBPath()
	if _, err := os.Stat(dbPath); err != nil {
		// Try to create the data directory
		if err2 := os.MkdirAll(filepath.Dir(dbPath), 0o755); err2 != nil {
			return nil, fmt.Errorf("nanoclaw: mkdir db dir: %w", err2)
		}
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("nanoclaw: open db rw: %w", err)
	}
	return db, nil
}

// InsertUserMessage inserts a user message into nanoclaw's database.
func InsertUserMessage(content, projectID string) error {
	db, err := openNanoclawDBRW()
	if err != nil {
		return err
	}
	defer db.Close()

	now := time.Now().Unix()
	_, err = db.Exec(
		`INSERT INTO messages (role, content, created_at, project_id) VALUES (?, ?, ?, ?)`,
		"user", content, now, nullableString(projectID),
	)
	if err != nil {
		return fmt.Errorf("nanoclaw: insert message: %w", err)
	}
	return nil
}

// WatchForResponses polls nanoclaw's database for new agent messages since the given ID.
// The returned channel is closed when ctx is cancelled.
func WatchForResponses(ctx context.Context, since int64) <-chan AgentMessage {
	ch := make(chan AgentMessage, 16)
	go func() {
		defer close(ch)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		lastID := since

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				msgs, err := fetchMessagesSince(lastID)
				if err != nil {
					continue
				}
				for _, m := range msgs {
					if m.Role == "agent" || m.Role == "assistant" {
						select {
						case <-ctx.Done():
							return
						case ch <- m:
							if m.ID > lastID {
								lastID = m.ID
							}
						}
					}
				}
			}
		}
	}()
	return ch
}

func fetchMessagesSince(since int64) ([]AgentMessage, error) {
	db, err := openNanoclawDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT id, role, content, created_at, COALESCE(project_id, '') FROM messages WHERE id > ? ORDER BY id ASC LIMIT 50`,
		since,
	)
	if err != nil {
		return nil, fmt.Errorf("nanoclaw: query messages: %w", err)
	}
	defer rows.Close()

	var msgs []AgentMessage
	for rows.Next() {
		var m AgentMessage
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt, &m.ProjectID); err != nil {
			continue
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
