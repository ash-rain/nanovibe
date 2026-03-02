# CLAUDE.md — VibeCodePC.com Development Context

This file is the authoritative guide for Claude Code when working on this project.
Read this before touching any file.

---

## Project Identity

**Name**: VibeCodePC.com
**Purpose**: Self-hosted web app that turns a Raspberry Pi into an AI coding station.
**Primary CWD**: `~/Documents/_DEV/nanovibe` (repo root)
**Automation principle**: The app does the work. Users watch progress bars, not type commands.

---

## Monorepo Structure

Go module at the root + pnpm for the frontend only:

```
/server    → Go backend (Chi router, SQLite, WebSockets, SSE, PTY)
/client    → Vue 3 + Vite + Tailwind CSS frontend
```

Root scripts via `Makefile`:
```bash
make dev          # air (Go hot reload) + Vite in parallel
make build        # vite build → server/public/; then go build -o dist/vibecodepc .
make check        # go vet ./... + vue-tsc --noEmit
make lint         # golangci-lint + eslint
make cross        # GOOS=linux GOARCH=arm64 go build -o dist/vibecodepc-arm64 .
```

---

## Server (`/server`)

### Stack
- **Language**: Go 1.22+
- **Router**: `github.com/go-chi/chi/v5` — lightweight, `net/http`-compatible, composable middleware
- **Database**: `modernc.org/sqlite` — pure Go SQLite (no CGO; cross-compiles to ARM cleanly)
- **WebSocket**: `github.com/gorilla/websocket`
- **Static files**: `net/http.FileServer` + `//go:embed` to bake the client build into the binary
- **Security headers**: custom middleware (CSP, HSTS, X-Frame-Options, etc.)
- **CORS**: `github.com/go-chi/cors` middleware (dev only)
- **Terminal**: `github.com/creack/pty` spawns opencode, output piped over WebSocket
- **Git**: `os/exec` shells to system `git` for reliability with auth/config
- **GitHub API**: `github.com/google/go-github/v60` with stored OAuth token
- **Metrics**: reads `/proc/stat`, `/proc/meminfo`, `/sys/class/thermal/` via `os` package

### Conventions
- Handlers in `server/routes/` — one file per domain, thin (validate → service → respond)
- Business logic in `server/services/` — never in handlers
- All SQLite access via the singleton from `server/db/db.go`
- Use the `db` package's prepared statements; never write raw SQL outside `db/`
- All AI keys and GitHub token via `services/keystore.go` — never read `settings` table directly in routes
- Error responses via helper: `httputil.WriteError(w, http.StatusBadRequest, "message")`
- SSE streams: set headers then write `fmt.Fprintf(w, "data: %s\n\n", payload)` and call `flusher.Flush()`

### SSE Pattern (Go)

```go
w.Header().Set("Content-Type", "text/event-stream")
w.Header().Set("Cache-Control", "no-cache")
w.Header().Set("Connection", "keep-alive")
flusher, ok := w.(http.Flusher)
if !ok {
    http.Error(w, "streaming not supported", http.StatusInternalServerError)
    return
}
for line := range logCh {
    fmt.Fprintf(w, "data: %s\n\n", line)
    flusher.Flush()
}
```

### WebSocket Pattern (Go)

```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}
conn, err := upgrader.Upgrade(w, r, nil)
```

### Database Schema (SQLite)

```sql
-- Setup wizard state (singleton row, id always 1)
CREATE TABLE setup_state (
  id INTEGER PRIMARY KEY,
  current_step TEXT NOT NULL DEFAULT 'welcome',
  completed_steps TEXT NOT NULL DEFAULT '[]',  -- JSON array
  updated_at INTEGER NOT NULL
);

-- Projects
CREATE TABLE projects (
  id TEXT PRIMARY KEY,           -- UUIDv4
  name TEXT NOT NULL,
  path TEXT NOT NULL UNIQUE,     -- Absolute filesystem path
  language TEXT,                 -- 'typescript' | 'python' | 'rust' | etc.
  github_url TEXT,               -- e.g. https://github.com/user/repo
  git_remote TEXT,               -- origin remote URL
  default_provider TEXT,         -- 'anthropic' | 'openai' | 'google' | 'ollama'
  created_at INTEGER NOT NULL,
  last_opened_at INTEGER
);

-- Encrypted key-value settings
CREATE TABLE settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,           -- AES-256-GCM encrypted, base64
  updated_at INTEGER NOT NULL
);
-- Known keys: anthropic_key, openai_key, google_key, cf_tunnel_token,
--             github_token, github_user_login, github_user_avatar

-- Agent chat messages
CREATE TABLE agent_messages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  role TEXT NOT NULL,            -- 'user' | 'agent'
  content TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  project_id TEXT                -- nullable FK to projects.id
);

-- GitHub OAuth state
CREATE TABLE github_auth (
  id INTEGER PRIMARY KEY,        -- always 1
  login TEXT,
  avatar_url TEXT,
  public_repos INTEGER,
  connected_at INTEGER
);
```

### Key Encryption (`db/crypto.go`)

Machine key = `SHA256(hostname + ":" + primaryMACAddress())`.
Derived at runtime via `crypto/sha256` + `net` interfaces. Never stored, never logged.

```go
// Usage via services/keystore.go only:
func Set(name, value string) error
func Get(name string) (string, bool)
func Del(name string) error
func Exists(name string) bool
func Mask(value string) string  // "sk-ant-••••••••1234"
```

AES-256-GCM via standard `crypto/aes` + `crypto/cipher`. Nonce is prepended to ciphertext, base64-encoded for storage.

---

## Client (`/client`)

### Stack
- **Framework**: Vue 3, `<script setup>` Composition API exclusively (no Options API)
- **Build**: Vite
- **Styling**: Tailwind CSS v4 — utility-first, no `<style>` blocks except third-party overrides
- **State**: Pinia — one store per domain
- **Routing**: Vue Router v4 with typed route definitions
- **Icons**: `@heroicons/vue` (outline set as default, solid for active states)
- **Terminal**: `@xterm/xterm` + `@xterm/addon-fit` + `@xterm/addon-web-links`
- **Markdown**: `markdown-it` for agent responses
- **Syntax highlight**: `shiki` (lazy-loaded) in code blocks
- **Charts**: `uplot` (tiny, fast) for sparklines in vitals
- **QR codes**: `qrcode` (canvas-based, tiny)
- **Confetti**: `canvas-confetti` (setup complete step)
- **Git UI**: custom — talks to server API, not client-side git

### Conventions
- Always `<script setup lang="ts">`
- Props: `defineProps<{ foo: string }>()`
- Emits: `defineEmits<{ change: [value: string] }>()`
- Store access at the top of `<script setup>`: `const store = useMyStore()`
- All HTTP: through store actions or the `useFetch` composable — never raw `fetch()` in components
- All WebSocket/SSE: through composables (`useTerminal`, `useMetrics`, `useAgentStream`)
- No direct `localStorage` access in components — use store with persistence plugin

### Design System

**Palette** (Tailwind CSS v4 `@theme` block in `client/src/main.css`):
```css
--color-surface-900: #0f0f1a;   /* page background */
--color-surface-800: #16162a;   /* card background */
--color-surface-700: #1e1e35;   /* elevated / input */
--color-surface-600: #2a2a45;   /* border / divider */
--color-primary:     #7c3aed;   /* violet-600 — CTAs, active nav */
--color-primary-glow:#a855f7;   /* violet-400 — hover, accent */
--color-success:     #34d399;   /* emerald-400 */
--color-warning:     #fbbf24;   /* amber-400 */
--color-danger:      #f87171;   /* red-400 */
--color-text:        #e2e8f0;   /* slate-200 */
--color-muted:       #94a3b8;   /* slate-400 */
```

**Typography**:
- `font-sans` → Inter (`@fontsource/inter`)
- `font-mono` → JetBrains Mono (`@fontsource/jetbrains-mono`)

**Dark mode**: Default. `prefers-color-scheme` fallback. Toggle in localStorage via `useColorMode()`.

**Motion**: Tailwind `transition-*` for property transitions. Vue `<Transition name="page">` for route changes (fade + 8 px translateY). `<TransitionGroup>` for list animations. Never block on animation — use `@after-enter` hooks.

**StatusBadge**: Pulsing dot + text. Dot color maps: `running → success`, `starting → warning`, `error → danger`, `stopped → muted`.

---

## Routing Structure

```
/                            → redirect: setup incomplete → /setup/welcome, else → /app/dashboard
/setup
  /setup/welcome             → StepWelcome
  /setup/system-check        → StepSystemCheck
  /setup/cloudflare          → StepCloudflare
  /setup/github              → StepGitHub
  /setup/providers           → StepProviders
  /setup/opencode            → StepOpenCode
  /setup/nanoclaw            → StepNanoClaw
  /setup/complete            → StepComplete
/app
  /app/dashboard             → DashboardView
  /app/ide/:projectId?       → IDEView
  /app/agent                 → AgentView
  /app/projects              → ProjectsView
  /app/github                → GitHubView
  /app/settings              → SettingsView
/auth
  /auth/github/callback      → handled server-side, redirects back to app
```

Navigation guards in `router/index.ts`:
- Root `beforeEach`: fetch `GET /api/setup/state` once on first load; cache in setup store
- `/app/*` guard: if `setupState.currentStep !== 'complete'` → push `/setup/welcome`
- `/setup/*` guard: if `setupState.currentStep === 'complete'` → push `/app/dashboard`

---

## Pinia Stores

### `setup.ts`
```typescript
state: {
  currentStep: string          // from server
  completedSteps: string[]
  checks: SystemCheck[]        // populated by system-check step
  tunnelStatus: TunnelStatus
  isLoading: boolean
}
actions:
  fetchState()        // GET /api/setup/state
  advanceTo(step)     // POST /api/setup/state
  runChecks()         // GET /api/setup/check/system
  fixCheck(id)        // POST /api/setup/fix/:id — returns SSE log stream
```

### `projects.ts`
```typescript
state: {
  list: Project[]
  active: Project | null
  gitStatus: Record<string, GitStatus>  // keyed by projectId
}
actions:
  fetchAll()
  create(data: ProjectCreate)
  remove(id)
  setActive(id)
  fetchGitStatus(id)
  gitCommit(id, message)
  gitPush(id)
  gitPull(id)
  gitCheckout(id, branch)
```

### `github.ts`
```typescript
state: {
  connected: boolean
  user: { login, avatar, publicRepos } | null
  repos: Repo[]
  reposPage: number
  repoSearch: string
  activity: GitHubEvent[]
  prs: Record<string, PR[]>   // keyed by "owner/repo"
}
actions:
  fetchStatus()
  fetchRepos(page?, search?)
  fetchActivity()
  importRepo(repoUrl, name, path)
  createPR(owner, repo, params)
```

### `metrics.ts`
```typescript
state: {
  current: SystemMetrics | null
  history: SystemMetrics[]      // last 60 readings
  connected: boolean
}
// Composable useMetrics() opens SSE on DashboardView mount
// Pushes every 2 s: { cpu, ramUsedMb, ramTotalMb, diskUsedGb, diskTotalGb, tempC, uptimeS }
```

### `settings.ts`
```typescript
state: {
  providers: ProviderConfig[]   // masked keys
  tunnel: TunnelStatus
  system: SystemInfo
}
```

### `agent.ts`
```typescript
state: {
  messages: AgentMessage[]
  connected: boolean
  typing: boolean
  activeProjectId: string | null
}
```

---

## API Contract

### Setup API
```
GET  /api/setup/state                    → { currentStep, completedSteps }
POST /api/setup/state                    → { step } → 204
GET  /api/setup/check/system             → { checks: SystemCheck[] }
POST /api/setup/fix/:checkId             → SSE stream of fix log lines
POST /api/setup/cloudflare/validate      → { token } → { valid, tunnelUrl }
POST /api/setup/providers/detect         → scan env → { detected: ProviderKey[] }
POST /api/setup/providers/test           → { provider, key } → { valid, models? }
POST /api/setup/opencode/install         → SSE stream (install stdout)
POST /api/setup/nanoclaw/setup           → SSE stream (clone + docker build)
```

### Projects API
```
GET    /api/projects                     → { data: Project[] }
POST   /api/projects                     → ProjectCreate → { data: Project }
GET    /api/projects/:id                 → { data: Project }
DELETE /api/projects/:id                 → 204
GET    /api/projects/:id/git/status      → GitStatus
POST   /api/projects/:id/git/commit      → { message } → 204
POST   /api/projects/:id/git/push        → 204
POST   /api/projects/:id/git/pull        → 204
POST   /api/projects/:id/git/checkout    → { branch } → 204
GET    /api/projects/:id/git/diff        → { diff: string }
GET    /api/projects/:id/git/branches    → { branches: Branch[] }
```

### GitHub API
```
GET    /api/github/status                → { connected, user? }
GET    /api/github/repos                 → { repos: Repo[], total }  ?page=1&search=
GET    /api/github/repos/:owner/:repo/prs → { prs: PR[] }
POST   /api/github/repos/:owner/:repo/prs → PrCreate → { pr: PR }
GET    /api/github/activity              → { events: GitHubEvent[] }
POST   /api/github/import               → { repoUrl, name, path } → SSE clone progress + { projectId }
DELETE /api/github/disconnect            → 204 (removes token)
```

### Settings API
```
GET  /api/settings/providers             → { providers: ProviderConfig[] }
POST /api/settings/providers             → { provider, key } → 204
GET  /api/settings/tunnel                → { mode, connected, tunnelUrl, localUrl, uptimeS }
POST /api/settings/tunnel/restart        → 204
POST /api/settings/tunnel/upgrade        → { token } → 204
GET  /api/settings/system                → { hostname, ip, localUrl, goVersion, dockerVersion, appVersion }
POST /api/settings/system/update         → SSE stream (update.sh progress)
```

### Agent API
```
GET    /api/agent/stream                 → SSE (event: message | ping)
POST   /api/agent/message               → { content, projectId? } → 204
GET    /api/agent/messages              → { data: AgentMessage[] }  (last 50)
DELETE /api/agent/messages              → 204
```

### Metrics API
```
GET /api/metrics/stream                  → SSE every 2 s
  event: metrics
  data: { cpu: number, ramUsedMb: number, ramTotalMb: number,
          diskUsedGb: number, diskTotalGb: number, tempC: number, uptimeS: number }
```

### Terminal WebSocket
```
WS /ws/terminal/:projectId
  C→S: { type: 'input', data: string }
  C→S: { type: 'resize', cols: number, rows: number }
  S→C: { type: 'output', data: string }
  S→C: { type: 'exit', code: number }
```

### Auth
```
GET /auth/github/start     → 302 to GitHub OAuth
GET /auth/github/callback  → exchanges code, stores token, 302 to /app/dashboard
```

---

## Services Reference

### `system_check.go`
```go
type CheckID string
const (
    CheckDocker       CheckID = "docker"
    CheckDockerDaemon CheckID = "docker-daemon"
    CheckRAM          CheckID = "ram"
    CheckDisk         CheckID = "disk"
    CheckInternet     CheckID = "internet"
    CheckGit          CheckID = "git"
)

type SystemCheck struct {
    ID       CheckID
    Label    string
    Critical bool
    Status   string  // "pending" | "running" | "pass" | "fail" | "warning"
    Detail   string
    Fixable  bool
}

func RunAllChecks(ctx context.Context) ([]SystemCheck, error)
func RunCheck(ctx context.Context, id CheckID) (SystemCheck, error)
// Fix actions return a channel of log lines:
func FixDocker(ctx context.Context) <-chan string
func FixDockerDaemon(ctx context.Context) <-chan string
func FixGit(ctx context.Context) <-chan string
```

### `opencode.go`
```go
func IsInstalled() bool
func Install(ctx context.Context) <-chan string    // streams stdout of npm install -g opencode
func GetVersion() (string, error)
func WriteConfig(providers []ProviderConfig) error // writes ~/.config/opencode/config.json
func StartSession(projectID, cwd string) (*os.File, error)  // returns PTY master (creack/pty)
func KillSession(projectID string) error
func ResizeSession(projectID string, cols, rows uint16) error
func GetSession(projectID string) (*os.File, bool)
```

### `nanoclaw.go`
```go
func IsCloned() bool
func Clone(ctx context.Context) <-chan string
func IsRunning(ctx context.Context) (bool, error)
func Start(ctx context.Context) error
func Stop(ctx context.Context) error
func WriteEnv(providers []ProviderConfig) error
func InsertUserMessage(content, projectID string) error
func WatchForResponses(ctx context.Context, since int64) <-chan AgentMessage
```

### `cloudflare.go`
```go
type TunnelMode string
const (
    ModeNone  TunnelMode = "none"
    ModeQuick TunnelMode = "quick"
    ModeNamed TunnelMode = "named"
)

type TunnelStatus struct {
    Mode      TunnelMode
    Connected bool
    TunnelURL string
    LocalURL  string
    UptimeS   int64
}

func Download(arch string) error  // arch: "arm64" | "arm" | "amd64"
func StartQuickTunnel() (*exec.Cmd, error)
func StartNamedTunnel(token string) (*exec.Cmd, error)
func StopTunnel() error
func GetStatus() TunnelStatus
func ValidateToken(ctx context.Context, token string) (bool, string, error)
```

### `github.go`
```go
func GetClient() (*github.Client, error)  // returns error if not authenticated
func IsAuthenticated() bool
func GetUser(ctx context.Context) (*GitHubUser, error)
func ListRepos(ctx context.Context, page int, search string) ([]Repo, error)
func ListPRs(ctx context.Context, owner, repo string) ([]PR, error)
func CreatePR(ctx context.Context, owner, repo string, params PRCreate) (*PR, error)
func GetActivity(ctx context.Context, login string) ([]GitHubEvent, error)
func ImportRepo(ctx context.Context, repoURL, destPath string) <-chan string
```

### `git.go`
```go
// All functions shell out to system git via os/exec.
// GIT_CONFIG is set to ~/.vibecodepc/.gitconfig for credential injection.

type GitStatus struct {
    Branch    string
    Ahead     int
    Behind    int
    Staged    []string
    Unstaged  []string
    Untracked []string
}

type Branch struct {
    Name    string
    Current bool
    Remote  bool
}

func Status(projectPath string) (GitStatus, error)
func Diff(projectPath string) (string, error)
func Commit(projectPath, message string) error
func Push(projectPath string) error
func Pull(projectPath string) error
func Branches(projectPath string) ([]Branch, error)
func Checkout(projectPath, branch string) error
func Clone(ctx context.Context, url, destPath string) <-chan string
```

### `metrics.go`
```go
type SystemMetrics struct {
    CPU         float64  // 0–100
    RAMUsedMB   int64
    RAMTotalMB  int64
    DiskUsedGB  float64
    DiskTotalGB float64
    TempC       float64  // from /sys/class/thermal (0 if unavailable)
    UptimeS     int64
}

func Read() (SystemMetrics, error)
// Caller streams via time.Ticker + SSE
```

### `keystore.go`
```go
func Set(name, value string) error
func Get(name string) (string, bool)
func Del(name string) error
func Exists(name string) bool
func Mask(value string) string  // "sk-ant-••••••••1234"
```

---

## Environment & Running

### Development
```bash
make dev
# air: Go server with hot reload on :3000
# Vite: client dev server on :5173 (proxies /api/*, /ws/*, /auth/* to :3000)
```

`air` config (`.air.toml` at root):
```toml
[build]
  cmd = "go build -o ./tmp/vibecodepc ./server/..."
  bin = "./tmp/vibecodepc"
  include_ext = ["go"]
  exclude_dir = ["client", "tmp", "dist"]
```

### Production (Raspberry Pi)
```bash
make build
./dist/vibecodepc
# Or via systemd: vibecodepc.service
```

The binary embeds the built client assets via `//go:embed`:
```go
//go:embed public/*
var staticFiles embed.FS
```

### Cross-Compilation for ARM64
```bash
GOOS=linux GOARCH=arm64 go build -o dist/vibecodepc-arm64 ./server/...
```

No CGO — `modernc.org/sqlite` is pure Go so this works without a cross-toolchain.

### Environment Variables

```env
PORT=3000
HOST=0.0.0.0
APP_ENV=production
DATA_DIR=/home/pi/.vibecodepc/data

# GitHub OAuth App (register at github.com/settings/developers)
GITHUB_CLIENT_ID=your_client_id
GITHUB_CLIENT_SECRET=your_client_secret
GITHUB_REDIRECT_URI=http://localhost:3000/auth/github/callback
```

`GITHUB_REDIRECT_URI` must match the app's public URL in production (the Cloudflare tunnel URL or local URL). The wizard Settings → GitHub tab provides a reminder to update this.

---

## Working with NanoClaw

- Cloned into `~/.vibecodepc/nanoclaw/` during setup wizard Step 7
- SQLite at `~/.vibecodepc/nanoclaw/data/nanoclaw.db` (volume-mounted into Docker)
- `scripts/nanoclaw-web-bridge.patch` applied post-clone: registers `web` as a virtual platform
- Never modify nanoclaw's source in-repo — apply as a patch after cloning

---

## Common Pitfalls

1. **creack/pty on ARM**: Pure Go — no native compile needed. Works out of the box.
2. **modernc.org/sqlite on ARM**: Pure Go — cross-compiles cleanly with `GOARCH=arm64`. No CGO.
3. **opencode PATH**: opencode is a Node.js app; use `exec.LookPath("opencode")` at runtime.
4. **cloudflared arch**: `runtime.GOARCH` → `arm64` / `arm` / `amd64`; map to cloudflare release names (`linux-arm64`, `linux-arm`, `linux-amd64`).
5. **SSE and quick tunnel URL**: Parse stdout of `cloudflared` for lines matching `trycloudflare.com` to extract the assigned URL.
6. **GitHub OAuth redirect URI**: Must exactly match what's registered in the GitHub OAuth App. In dev this is `http://localhost:3000/auth/github/callback`; in production it must be the Cloudflare tunnel URL. Surfaced clearly in Settings.
7. **git credentials**: Inject via `GIT_CONFIG` env pointing to `~/.vibecodepc/.gitconfig` which uses `url.<base>.insteadOf` + credential helper storing the GitHub token.
8. **Metrics on non-Pi Linux**: `/sys/class/thermal/` may not exist — `TempC` returns 0 gracefully.
9. **SPA catch-all**: Register a catch-all route after all API routes that serves `index.html` from the embedded FS for all non-API, non-WS, non-auth paths.
10. **Context cancellation**: All long-running goroutines (SSE, PTY piping, metrics tickers) must respect `r.Context().Done()` to clean up on client disconnect.

---

## Testing

- Unit: `go test ./...` with table-driven tests, colocated `*_test.go`
- Handler tests: `net/http/httptest` with the Chi router (no real HTTP listener)
- Client: `@vue/test-utils` + Vitest + happy-dom
- E2E: Playwright (Phase 8+)
- CI: GitHub Actions on `main` — `go vet`, `golangci-lint`, `go test`, vue-tsc, ESLint

---

## Git Conventions

- Branches: `feature/<name>`, `fix/<name>`, `chore/<name>`
- Commits: Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`)
- No force-push to `main`
