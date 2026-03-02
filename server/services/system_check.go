package services

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// CheckID identifies a system check.
type CheckID string

const (
	CheckDocker       CheckID = "docker"
	CheckDockerDaemon CheckID = "docker-daemon"
	CheckRAM          CheckID = "ram"
	CheckDisk         CheckID = "disk"
	CheckInternet     CheckID = "internet"
	CheckGit          CheckID = "git"
)

// SystemCheck represents the result of a system readiness check.
type SystemCheck struct {
	ID       CheckID `json:"id"`
	Label    string  `json:"label"`
	Critical bool    `json:"critical"`
	Status   string  `json:"status"` // pending | running | pass | fail | warning
	Detail   string  `json:"detail"`
	Fixable  bool    `json:"fixable"`
}

// RunAllChecks executes all system checks concurrently and returns the results.
func RunAllChecks(ctx context.Context) ([]SystemCheck, error) {
	ids := []CheckID{CheckDocker, CheckDockerDaemon, CheckRAM, CheckDisk, CheckInternet, CheckGit}

	type result struct {
		idx   int
		check SystemCheck
		err   error
	}

	results := make([]SystemCheck, len(ids))
	ch := make(chan result, len(ids))
	var wg sync.WaitGroup

	for i, id := range ids {
		wg.Add(1)
		go func(i int, id CheckID) {
			defer wg.Done()
			c, err := RunCheck(ctx, id)
			ch <- result{idx: i, check: c, err: err}
		}(i, id)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for r := range ch {
		if r.err != nil {
			results[r.idx] = SystemCheck{
				ID:     ids[r.idx],
				Status: "fail",
				Detail: r.err.Error(),
			}
		} else {
			results[r.idx] = r.check
		}
	}
	return results, nil
}

// RunCheck executes a single system check and returns the result.
func RunCheck(ctx context.Context, id CheckID) (SystemCheck, error) {
	switch id {
	case CheckDocker:
		return checkDockerInstalled(ctx)
	case CheckDockerDaemon:
		return checkDockerDaemon(ctx)
	case CheckRAM:
		return checkRAM()
	case CheckDisk:
		return checkDisk()
	case CheckInternet:
		return checkInternet(ctx)
	case CheckGit:
		return checkGit(ctx)
	default:
		return SystemCheck{}, fmt.Errorf("system_check: unknown check id %q", id)
	}
}

func checkDockerInstalled(_ context.Context) (SystemCheck, error) {
	c := SystemCheck{
		ID:       CheckDocker,
		Label:    "Docker installed",
		Critical: true,
		Fixable:  true,
	}
	_, err := exec.LookPath("docker")
	if err != nil {
		c.Status = "fail"
		c.Detail = "Docker not found in PATH"
		return c, nil
	}
	out, err := exec.Command("docker", "--version").Output()
	if err != nil {
		c.Status = "fail"
		c.Detail = fmt.Sprintf("docker --version failed: %v", err)
		return c, nil
	}
	c.Status = "pass"
	c.Detail = strings.TrimSpace(string(out))
	return c, nil
}

func checkDockerDaemon(ctx context.Context) (SystemCheck, error) {
	c := SystemCheck{
		ID:       CheckDockerDaemon,
		Label:    "Docker daemon running",
		Critical: true,
		Fixable:  true,
	}
	cmd := exec.CommandContext(ctx, "docker", "info")
	out, err := cmd.CombinedOutput()
	if err != nil {
		c.Status = "fail"
		c.Detail = "Docker daemon not running — run 'sudo systemctl start docker'"
		return c, nil
	}
	c.Status = "pass"
	// Extract server version from docker info output
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Server Version") {
			c.Detail = strings.TrimSpace(line)
			break
		}
	}
	return c, nil
}

func checkRAM() (SystemCheck, error) {
	c := SystemCheck{
		ID:       CheckRAM,
		Label:    "Available RAM",
		Critical: false,
		Fixable:  false,
	}
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		c.Status = "warning"
		c.Detail = "Cannot read /proc/meminfo"
		return c, nil
	}
	defer f.Close()

	vals := make(map[string]int64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 2 {
			key := strings.TrimSuffix(parts[0], ":")
			v, _ := strconv.ParseInt(parts[1], 10, 64)
			vals[key] = v
		}
	}

	totalMB := vals["MemTotal"] / 1024
	availMB := vals["MemAvailable"] / 1024

	c.Detail = fmt.Sprintf("%d MB total, %d MB available", totalMB, availMB)
	if totalMB < 512 {
		c.Status = "warning"
		c.Detail += " (minimum 512 MB recommended)"
	} else {
		c.Status = "pass"
	}
	return c, nil
}

func checkDisk() (SystemCheck, error) {
	c := SystemCheck{
		ID:       CheckDisk,
		Label:    "Free disk space",
		Critical: false,
		Fixable:  false,
	}
	usedGB, totalGB, err := readDisk()
	if err != nil {
		c.Status = "warning"
		c.Detail = fmt.Sprintf("Cannot read disk info: %v", err)
		return c, nil
	}
	freeGB := totalGB - usedGB
	c.Detail = fmt.Sprintf("%.1f GB free of %.1f GB", freeGB, totalGB)
	if freeGB < 1.0 {
		c.Status = "warning"
		c.Detail += " (minimum 1 GB recommended)"
	} else {
		c.Status = "pass"
	}
	return c, nil
}

func checkInternet(ctx context.Context) (SystemCheck, error) {
	c := SystemCheck{
		ID:       CheckInternet,
		Label:    "Internet connectivity",
		Critical: true,
		Fixable:  false,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, "https://cloudflare.com", nil)
	if err != nil {
		c.Status = "fail"
		c.Detail = fmt.Sprintf("Cannot create request: %v", err)
		return c, nil
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.Status = "fail"
		c.Detail = fmt.Sprintf("No internet: %v", err)
		return c, nil
	}
	resp.Body.Close()
	c.Status = "pass"
	c.Detail = "Connected to internet"
	return c, nil
}

func checkGit(_ context.Context) (SystemCheck, error) {
	c := SystemCheck{
		ID:       CheckGit,
		Label:    "Git installed",
		Critical: true,
		Fixable:  true,
	}
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		c.Status = "fail"
		c.Detail = "git not found in PATH"
		return c, nil
	}
	c.Status = "pass"
	c.Detail = strings.TrimSpace(string(out))
	return c, nil
}

// FixDocker installs Docker and streams progress lines.
func FixDocker(ctx context.Context) <-chan string {
	ch := make(chan string, 64)
	go func() {
		defer close(ch)
		ch <- "Downloading Docker installation script..."
		cmd := exec.CommandContext(ctx, "sh", "-c", "curl -fsSL https://get.docker.com | sh")
		streamCmd(ctx, cmd, ch)
	}()
	return ch
}

// FixDockerDaemon starts the Docker daemon and streams progress.
func FixDockerDaemon(ctx context.Context) <-chan string {
	ch := make(chan string, 64)
	go func() {
		defer close(ch)
		ch <- "Starting Docker daemon..."
		cmd := exec.CommandContext(ctx, "sudo", "systemctl", "start", "docker")
		streamCmd(ctx, cmd, ch)
		if ctx.Err() != nil {
			return
		}
		ch <- "Enabling Docker to start on boot..."
		cmd2 := exec.CommandContext(ctx, "sudo", "systemctl", "enable", "docker")
		streamCmd(ctx, cmd2, ch)
	}()
	return ch
}

// FixGit installs git and streams progress.
func FixGit(ctx context.Context) <-chan string {
	ch := make(chan string, 64)
	go func() {
		defer close(ch)
		ch <- "Installing git..."
		cmd := exec.CommandContext(ctx, "sudo", "apt-get", "install", "-y", "git")
		streamCmd(ctx, cmd, ch)
	}()
	return ch
}

// streamCmd runs the command and streams combined output lines to ch.
func streamCmd(ctx context.Context, cmd *exec.Cmd, ch chan<- string) {
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
		ch <- fmt.Sprintf("error: %v", err)
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
		ch <- fmt.Sprintf("error: command failed: %v", err)
	} else {
		ch <- "done"
	}
}
