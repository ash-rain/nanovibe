package services

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// GitStatus represents the current state of a git repository.
type GitStatus struct {
	Branch    string   `json:"branch"`
	Ahead     int      `json:"ahead"`
	Behind    int      `json:"behind"`
	Staged    []string `json:"staged"`
	Unstaged  []string `json:"unstaged"`
	Untracked []string `json:"untracked"`
}

// Branch represents a git branch.
type Branch struct {
	Name    string `json:"name"`
	Current bool   `json:"current"`
	Remote  bool   `json:"remote"`
}

// gitConfigPath returns the path to the vibecodepc gitconfig file.
func gitConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".vibecodepc", ".gitconfig")
}

// gitCmd creates an exec.Cmd for a git command in the given directory.
func gitCmd(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+gitConfigPath())
	return cmd
}

// Status returns the current git status for the repository at projectPath.
func Status(projectPath string) (GitStatus, error) {
	var status GitStatus

	// Get branch and tracking info
	branchOut, err := gitCmd(projectPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return status, fmt.Errorf("git status: get branch: %w", err)
	}
	status.Branch = strings.TrimSpace(string(branchOut))

	// Ahead/behind counts
	abOut, err := gitCmd(projectPath, "rev-list", "--left-right", "--count", "@{u}...HEAD").Output()
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(string(abOut)))
		if len(parts) == 2 {
			status.Behind, _ = strconv.Atoi(parts[0])
			status.Ahead, _ = strconv.Atoi(parts[1])
		}
	}

	// Porcelain status
	out, err := gitCmd(projectPath, "status", "--porcelain").Output()
	if err != nil {
		return status, fmt.Errorf("git status: porcelain: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}
		x := string(line[0]) // index status
		y := string(line[1]) // working-tree status
		path := strings.TrimSpace(line[3:])
		// Handle renames "old -> new"
		if idx := strings.Index(path, " -> "); idx != -1 {
			path = path[idx+4:]
		}

		if x == "?" && y == "?" {
			status.Untracked = append(status.Untracked, path)
		} else {
			if x != " " && x != "?" {
				status.Staged = append(status.Staged, path)
			}
			if y != " " && y != "?" {
				status.Unstaged = append(status.Unstaged, path)
			}
		}
	}

	return status, nil
}

// Diff returns the unified diff of unstaged changes.
func Diff(projectPath string) (string, error) {
	out, err := gitCmd(projectPath, "diff").Output()
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return string(out), nil
}

// Commit stages all changes and creates a commit with the given message.
func Commit(projectPath, message string) error {
	if err := gitCmd(projectPath, "add", "-A").Run(); err != nil {
		return fmt.Errorf("git commit: stage: %w", err)
	}
	out, err := gitCmd(projectPath, "commit", "-m", message).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Push pushes the current branch to the remote.
func Push(projectPath string) error {
	out, err := gitCmd(projectPath, "push").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Pull pulls the latest changes from the remote.
func Pull(projectPath string) error {
	out, err := gitCmd(projectPath, "pull").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Branches returns all local and remote branches for the repository.
func Branches(projectPath string) ([]Branch, error) {
	out, err := gitCmd(projectPath, "branch", "-a", "--format=%(refname:short) %(HEAD)").Output()
	if err != nil {
		return nil, fmt.Errorf("git branches: %w", err)
	}

	var branches []Branch
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		name := parts[0]
		current := len(parts) > 1 && parts[1] == "*"
		remote := strings.HasPrefix(name, "remotes/")
		if remote {
			name = strings.TrimPrefix(name, "remotes/")
		}
		branches = append(branches, Branch{
			Name:    name,
			Current: current,
			Remote:  remote,
		})
	}
	return branches, nil
}

// Checkout switches to the given branch.
func Checkout(projectPath, branch string) error {
	out, err := gitCmd(projectPath, "checkout", branch).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout %q: %s: %w", branch, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Clone clones url into destPath and streams progress lines into the returned channel.
// The channel is closed when the clone completes or fails.
func Clone(ctx context.Context, url, destPath string) <-chan string {
	ch := make(chan string, 32)
	go func() {
		defer close(ch)
		cmd := exec.CommandContext(ctx, "git", "clone", "--progress", url, destPath)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+gitConfigPath())

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
			ch <- fmt.Sprintf("error: %v", err)
			return
		}
		pw.Close()

		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case ch <- scanner.Text():
			}
		}
		pr.Close()

		if err := cmd.Wait(); err != nil {
			ch <- fmt.Sprintf("error: clone failed: %v", err)
		} else {
			ch <- "done"
		}
	}()
	return ch
}
