# VibeCodePC.com — Development Plan

> Turn any Raspberry Pi into a personal AI coding powerhouse.
> Anyone can set it up. Anyone can vibe-code.

---

## Vision

VibeCodePC.com transforms a Raspberry Pi (or any Linux machine) into a fully self-hosted AI coding station. The experience is opinionated and automated: the installer does the heavy lifting, the setup wizard configures itself wherever possible, and the dashboard gives a real-time command center for every aspect of the system. GitHub is a first-class citizen — repos are browsable, projects spawn from a single click, and git operations happen from inside the app. No command line required after install.

---

## Automation Philosophy

Every interaction follows one rule: **the app should do the work, not the user.**

- **Auto-detect**: system capabilities, installed tools, existing API keys in env vars, GitHub identity
- **Auto-install**: missing dependencies (Docker, git) with live terminal progress
- **Auto-configure**: opencode and nanoclaw from stored provider keys — no manual config files
- **Auto-advance**: wizard steps complete themselves when all checks pass; user just watches
- **Auto-fix**: every failing check has a one-click "Fix it" action that runs and re-checks
- **Auto-reconnect**: tunnels, WebSocket connections, and services restart themselves silently

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Vue 3, Vite, Tailwind CSS v4, Pinia, Vue Router |
| Terminal embed | `@xterm/xterm` + `@xterm/addon-fit` + `@xterm/addon-web-links` |
| Backend | Go 1.22+, Chi router |
| WebSockets | `gorilla/websocket` (terminal I/O) |
| SSE | Standard `net/http` with `http.Flusher` (agent chat, log tailing, metrics) |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |
| Process mgmt | `creack/pty` (opencode terminal sessions) |
| Git operations | `os/exec` → system `git` (status, diff, commit, push, pull, branches) |
| GitHub API | `google/go-github` + OAuth flow |
| Key storage | AES-256-GCM encrypted fields in SQLite (standard `crypto` package) |
| Tunneling | cloudflared (Cloudflare Tunnel binary, quick + named modes) |
| AI Coding | opencode (anomalyco/opencode) |
| AI Agent | nanoclaw (qwibitai/nanoclaw) |
| Installer | Bash script + systemd unit files |
| Binary | Single statically-linked Go binary with embedded frontend assets |

---

## Architecture Overview

```
Browser (Vue 3)
     │
     │  HTTP / WebSocket / SSE
     ▼
Go Server (Chi) :3000
     ├── /api/setup          ← Wizard state machine + auto-run actions
     ├── /api/projects       ← Project CRUD + git status
     ├── /api/github         ← OAuth flow, repo browser, PR creation
     ├── /api/settings       ← AI keys, tunnel config, system info
     ├── /api/agent          ← NanoClaw bridge (SSE + POST)
     ├── /api/metrics/stream ← SSE: CPU/RAM/disk/temp every 2 s
     ├── /ws/terminal/:id    ← xterm.js ↔ creack/pty ↔ opencode
     ├── /auth/github        ← GitHub OAuth callback
     └── /* (SPA)            ← Serves embedded Vue build (//go:embed)
     │
     ├── OpenCode Service     — creack/pty sessions per project
     ├── NanoClaw Service     — Docker exec + SQLite bridge
     ├── GitHub Service       — go-github, OAuth token, repo/PR API
     ├── Git Service          — os/exec → system git, per-project
     ├── Metrics Service      — /proc + os package poller
     ├── Cloudflare Service   — quick tunnel / named tunnel process
     └── SQLite DB
           ├── setup_state
           ├── projects
           ├── settings (encrypted keys)
           ├── agent_messages
           └── github_auth
```

---

## Directory Structure

```
vibecodepc/
├── PLAN.md
├── CLAUDE.md
├── README.md
├── go.mod                    # Go module root (module: vibecodepc)
├── go.sum
├── .air.toml                 # air hot-reload config (dev only)
├── Makefile                  # dev, build, check, lint, cross targets
│
├── server/                   # Go source packages
│   ├── main.go               # Entry point: wire up Chi, DB, services, start
│   ├── config/
│   │   └── config.go         # Env var parsing, defaults
│   ├── db/
│   │   ├── db.go             # SQLite singleton, migrations on boot
│   │   ├── schema.go         # CREATE TABLE statements
│   │   └── crypto.go         # AES-256-GCM + machine key derivation
│   ├── routes/
│   │   ├── setup.go          # Wizard state + auto-action SSE streams
│   │   ├── projects.go       # CRUD + git status per project
│   │   ├── github.go         # OAuth, repo list, PR create
│   │   ├── settings.go       # AI keys, tunnel, system info
│   │   ├── agent.go          # NanoClaw SSE stream + POST
│   │   ├── metrics.go        # SSE: real-time system vitals
│   │   └── terminal.go       # WS terminal sessions (gorilla/websocket)
│   └── services/
│       ├── setup.go          # Step state machine + auto-run logic
│       ├── system_check.go   # Checks + auto-installers (Docker, git)
│       ├── opencode.go       # Install, launch, kill sessions (creack/pty)
│       ├── nanoclaw.go       # Clone, configure, Docker lifecycle
│       ├── cloudflare.go     # Quick/named tunnel process manager
│       ├── github.go         # go-github client, OAuth, repo/PR API
│       ├── git.go            # os/exec git per-project operations
│       ├── metrics.go        # CPU/RAM/disk/temp reader
│       └── keystore.go       # AES-256-GCM key store
│
├── client/                   # Vue 3 frontend (standalone pnpm package)
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   └── src/
│       ├── main.ts
│       ├── App.vue
│       ├── router/index.ts
│       ├── stores/
│       │   ├── setup.ts        # Wizard step & auto-run state
│       │   ├── projects.ts     # Projects list, active, git status
│       │   ├── settings.ts     # Provider config, tunnel state
│       │   ├── agent.ts        # Chat messages, SSE connection
│       │   ├── github.ts       # Auth state, repos, PRs, activity
│       │   └── metrics.ts      # Live CPU/RAM/disk/temp
│       ├── views/
│       │   ├── setup/
│       │   │   ├── WizardLayout.vue      # Progress rail + step shell
│       │   │   ├── StepWelcome.vue       # Animated hero
│       │   │   ├── StepSystemCheck.vue   # Auto-running checks + fix buttons
│       │   │   ├── StepCloudflare.vue    # Tunnel status + optional upgrade
│       │   │   ├── StepGitHub.vue        # GitHub OAuth connect
│       │   │   ├── StepProviders.vue     # AI key entry + auto-detect
│       │   │   ├── StepOpenCode.vue      # Auto-install + live progress
│       │   │   ├── StepNanoClaw.vue      # Auto-setup + messaging QR
│       │   │   └── StepComplete.vue      # Confetti + launch
│       │   ├── DashboardView.vue         # Real-time command center
│       │   ├── IDEView.vue               # Embedded opencode terminal
│       │   ├── AgentView.vue             # NanoClaw web chat
│       │   ├── ProjectsView.vue          # Project browser
│       │   ├── GitHubView.vue            # Repo browser + PR management
│       │   └── SettingsView.vue          # All settings tabs
│       └── components/
│           ├── layout/
│           │   ├── AppShell.vue          # Sidebar + topbar wrapper
│           │   ├── AppSidebar.vue        # Nav + live service dots
│           │   ├── AppTopbar.vue         # Tunnel URL, GitHub avatar
│           │   └── MobileNav.vue         # Bottom nav for mobile
│           ├── dashboard/
│           │   ├── ServiceCard.vue       # Per-service status + actions
│           │   ├── VitalsGraph.vue       # Sparkline CPU/RAM over time
│           │   ├── ActivityFeed.vue      # GitHub + agent activity
│           │   ├── QuickLaunch.vue       # Recent projects grid
│           │   └── AccessPanel.vue       # URLs + QR codes
│           ├── terminal/
│           │   └── TerminalPane.vue
│           ├── agent/
│           │   ├── ChatBubble.vue
│           │   ├── ChatInput.vue
│           │   └── AgentStatus.vue
│           ├── projects/
│           │   ├── ProjectCard.vue       # Card with git badge, branch, status
│           │   ├── NewProjectModal.vue   # Local or GitHub import
│           │   └── GitPanel.vue          # Status, diff, commit, push/pull
│           ├── github/
│           │   ├── RepoCard.vue
│           │   ├── PrCard.vue
│           │   └── CommitList.vue
│           └── ui/
│               ├── StatusBadge.vue       # Pulsing dot + label
│               ├── KeyInput.vue          # Masked key field
│               ├── CheckItem.vue         # Check row with auto-fix
│               ├── StepIndicator.vue     # Wizard progress rail
│               ├── Sparkline.vue         # Tiny SVG graph
│               └── QrCode.vue            # Inline QR from URL
│
├── docker/
│   ├── docker-compose.yml
│   └── nanoclaw/Dockerfile
│
└── scripts/
    ├── install.sh
    ├── update.sh
    └── uninstall.sh
```

---

## Development Phases

### Phase 1 — Foundation (Week 1)

**Goal**: Go server + Vue shell with auth skeleton running.

- [ ] Go module init: `go mod init vibecodepc`, add Chi, gorilla/websocket, modernc.org/sqlite, go-github, creack/pty
- [ ] `Makefile` with `dev`, `build`, `check`, `lint`, `cross` targets
- [ ] `.air.toml` for Go hot reload; `vite.config.ts` proxy `/api/*`, `/ws/*`, `/auth/*` to `:3000`
- [ ] Go server: Chi router, static file serving from `//go:embed public/*`, security headers middleware
- [ ] SQLite schema: all 5 tables (see CLAUDE.md), `db.go` runs migrations on startup
- [ ] Vue 3 + Vite + Tailwind CSS v4 scaffold with design tokens
- [ ] Vue Router: `/setup/*` and `/app/*` root guards; setup-completion check on boot
- [ ] Wizard shell: `WizardLayout` with animated step rail
- [ ] Pinia stores: skeleton for all 6 domains
- [ ] GitHub OAuth App registration in config: `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`
- [ ] `/auth/github` routes in Go + token storage in keystore
- [ ] `useFetch` composable (base URL from env, auto JSON, error normalisation)
- [ ] golangci-lint + ESLint + Prettier shared config

**Deliverable**: `make dev` shows the wizard welcome screen; GitHub OAuth round-trip works.

---

### Phase 2 — Setup Wizard (Week 2)

**Goal**: Fully automated wizard that configures the device with minimal user input.

The wizard tracks state server-side in SQLite. On load it restores the user to their last step. Each step fires its checks/installs automatically on mount — the user watches progress, not drives it.

---

#### Step 1 — Welcome

- Full-screen animated hero: orbiting tool icons (Docker, GitHub, Claude, opencode) around a Raspberry Pi
- Three-line pitch: "Your Pi. Your AI. Your code."
- Auto-detect hostname → personalised greeting: "Let's set up **raspberrypi.local**"
- "Start Setup" button — instantly moves to Step 2

---

#### Step 2 — System Check (fully automated)

`system_check.go` runs all checks in parallel on mount. Each check row animates through pending → running → pass/fail.

| Check | Auto-fix available? |
|---|---|
| Docker installed | Yes — install Docker Engine (SSE log stream) |
| Docker daemon running | Yes — `sudo systemctl start docker` |
| RAM ≥ 1 GB | No — display warning only |
| Disk ≥ 5 GB free | No — show usage, proceed with warning |
| Internet (fetch `1.1.1.1`) | No — must fix externally |
| `git` installed | Yes — `apt-get install -y git` |

Note: Go itself is not a system check — the app runs as a prebuilt binary. The installer handles Go if building from source.

**Auto-fix flow**:
- User clicks "Fix" on a failing row
- `POST /api/setup/fix/:checkId` opens an SSE stream
- Terminal-style log scrolls in the check row
- On completion, check re-runs automatically
- Row transitions to pass with a green checkmark

**Auto-advance**: when all critical checks pass, a 1-second countdown starts and the step advances automatically (user can cancel).

---

#### Step 3 — Cloudflare Tunnel (status display)

The quick tunnel is already running from the installer. This step is purely informational + optional upgrade.

- Live status card: `cloudflared` PID, uptime, current URL
- Animated "tunnel connected" graphic: packets flowing Pi → Cloudflare → globe
- Two URL pills with copy + QR:
  - Local: `http://<hostname>.local:3000`
  - Remote: `https://<random>.trycloudflare.com`
- **Optional upgrade section** (collapsed by default):
  - "Get a stable URL" → expands guide to create named tunnel on `dash.cloudflare.com`
  - Token input → validates live → restarts cloudflared → shows new stable URL
- Step auto-advances after 3 seconds if tunnel is connected (override with "Stay here")

---

#### Step 4 — GitHub (optional but recommended)

- Headline: "Connect GitHub to browse and import your repos"
- GitHub OAuth button → opens popup → completes OAuth → popup closes → user identity shown
- On connect: show GitHub avatar, username, public repo count
- Pre-load the first page of repos in the background (used in Step 5 and Projects view)
- "Skip for now" link — GitHub can be connected later from Settings

---

#### Step 5 — AI Providers (auto-detected)

On mount, scan process env for common keys (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GOOGLE_API_KEY`). For each found:
- Auto-import with a "Detected from environment" badge
- Run a live validation test immediately
- Show provider as configured (green) before user touches anything

For providers not found:
- Provider card with masked key input
- "Test key" runs inline — spinner → green tick or red error
- Anthropic shown first and marked "Powers the AI agent (required for NanoClaw)"

Ollama: auto-probe `http://localhost:11434/api/tags` — if running, auto-configure.

**Auto-advance**: once at least one provider is configured and valid.

---

#### Step 6 — OpenCode (auto-install)

On mount:
1. Check if `opencode` binary exists → if yes, skip install, show version
2. If not: auto-start install (`npm install -g opencode`) with live SSE log stream
3. After install: auto-write `~/.config/opencode/config.json` from stored provider keys
4. Show: `opencode --version` output in a styled terminal block
5. Provider selector (pre-filled from Step 5)

Note: opencode is a Node.js application. Node.js is only required for opencode, not for the VibeCodePC server itself.

---

#### Step 7 — NanoClaw (optional, auto-setup)

- Step marked "(Optional)" in progress rail
- On mount if Anthropic key exists: auto-start setup (clone + docker build) with live SSE log
- Setup steps shown as a mini timeline:
  - `git clone qwibitai/nanoclaw` ← animating
  - Writing `.env` from stored keys ← animating
  - `docker build` ← streaming logs
  - Container started ← done
- Once container is up: show messaging platform options
  - **WhatsApp**: QR code from Baileys (auto-refreshed every 20 s)
  - **Telegram / Discord / Slack**: token input → inline validation
  - **Web only**: default, no extra config — the in-app chat is already connected
- "Skip messaging" skips platform pairing but still starts the web bridge

---

#### Step 8 — Complete

- Full-screen confetti burst (canvas-confetti)
- Summary grid: all configured services with green ticks
- Dual access panel: both URLs as large pills with one-click copy + QR side-by-side
- GitHub card (if connected): shows username + "X repos ready to import"
- Animated "Your station is ready" — morphing text effect
- Single CTA: "Open Dashboard" → navigates to `/app/dashboard`
- Auto-redirect after 8 seconds

---

### Phase 3 — OpenCode IDE Integration (Week 3)

**Goal**: In-browser terminal running opencode per project.

- `opencode.go` service:
  - `StartSession(projectID, cwd)` → spawn `opencode` via `creack/pty` in project CWD
  - `KillSession(projectID)` → SIGTERM + cleanup
  - `ResizeSession(projectID, cols, rows)` → `pty.Setsize()`
  - Session registry: `sync.Map` keyed by projectID → `{ ptmx *os.File, clients sync.Map }`
- WebSocket route `/ws/terminal/:projectId` (gorilla/websocket):
  - On connect: attach to or create session
  - `{ type: 'input', data }` → write to PTY master
  - `{ type: 'resize', cols, rows }` → PTY resize
  - PTY stdout → broadcast to all WebSocket clients for that session
  - `{ type: 'exit', code }` on process exit
- `TerminalPane.vue`:
  - `@xterm/xterm` + `@xterm/addon-fit` + `@xterm/addon-web-links`
  - Dracula-inspired dark theme (matches app surface colors)
  - ResizeObserver → auto-fit + send resize frame to server
  - Exponential backoff reconnect on disconnect
- `IDEView.vue`:
  - Two-panel: collapsible `GitPanel.vue` left + `TerminalPane` right
  - Project selector in topbar — switching auto-starts session for new project
  - Session toolbar: active project, branch badge, "New Session" / "Kill" buttons

---

### Phase 4 — NanoClaw Agent Chat (Week 4)

**Goal**: Polished web chat UI bridged to the nanoclaw agent.

- `nanoclaw.go` service:
  - Inserts user messages into nanoclaw's SQLite `messages` table as `source: 'web'`
  - Polls `SELECT` with timestamp cursor (100 ms interval) for outbound responses via `time.Ticker`
  - Emits agent messages via Go channel → consumed by SSE handler
- `GET /api/agent/stream` — SSE, heartbeat ping every 15 s (uses `http.Flusher`)
- `POST /api/agent/message` — `{ content, projectId? }`
- `AgentView.vue`:
  - Chat bubbles: user right, agent left with avatar
  - Animated typing indicator (three dots) while agent is processing
  - Markdown rendering (`markdown-it`) with `shiki` syntax highlighting in code blocks
  - Persistent context selector: "Chatting in context of: **my-saas-app**"
  - Draggable input — expands to multi-line on Shift+Enter
  - "Clear history" and agent restart in overflow menu

---

### Phase 5 — GitHub Integration (Week 4–5)

**Goal**: GitHub as a first-class citizen — repos, projects, git ops, PRs.

#### GitHub Service (`github.go`)

- OAuth: server-side flow (`/auth/github/start` → GitHub → `/auth/github/callback`)
- Token stored encrypted in keystore
- `go-github` client initialized on demand with stored token
- `ListRepos(page, search)` → paginated repo list with language, stars, last push
- `ListPRs(owner, repo)` → open PRs with title, branch, status checks
- `CreatePR(owner, repo, params)` → create PR from current branch
- `GetActivity(username)` → recent events (pushes, PR opens, issues)

#### Git Service (`git.go`)

- `Status(projectPath)` → `{ Branch, Ahead, Behind, Staged, Unstaged, Untracked }`
- `Diff(projectPath)` → unified diff string
- `Commit(projectPath, message)` → stage all + commit
- `Push(projectPath)` → push current branch (uses stored GitHub token via git credential helper)
- `Pull(projectPath)` → pull with rebase
- `Branches(projectPath)` → list local + remote branches
- `Checkout(projectPath, branch)` → switch branch (stash if dirty)
- `Clone(ctx, url, destPath)` → git clone, progress lines emitted on returned channel

#### GitHub Routes (`/api/github`)

```
GET  /api/github/status          → { authenticated, user: { login, avatar, publicRepos } }
GET  /api/github/repos           → { repos: Repo[] }  (paginated, searchable)
GET  /api/github/repos/:owner/:repo/prs  → { prs: PR[] }
POST /api/github/repos/:owner/:repo/prs  → create PR
GET  /api/github/activity        → { events: GitHubEvent[] }  (last 20)
POST /api/github/import          → { repoUrl, name, path } → clone + create project
```

#### `GitHubView.vue`

- Two tabs: **Repos** | **Activity**
- Repos tab:
  - Search input (filters client-side + refetches after 400 ms debounce)
  - Language + sort filters
  - `RepoCard.vue` grid: name, language badge, stars, last pushed, "Import" button
  - Import → `POST /api/github/import` with SSE clone progress → auto-navigates to new project
- Activity tab:
  - Timeline of recent GitHub events with icons (push, PR, issue, review)
  - Click event → opens GitHub URL in new tab

#### `GitPanel.vue` (in IDE view)

- Collapsible left panel in IDEView
- Live git status: M/A/D/? file badges
- Branch selector: dropdown of all branches, click to checkout
- Staged/unstaged file lists with diff preview (click file)
- Commit message input + "Commit & Push" button (or separate)
- Ahead/behind indicator with pull/push arrows
- PR shortcut: "Open PR on GitHub" → pre-filled create-PR modal

---

### Phase 6 — Real-Time Dashboard (Week 5–6)

**Goal**: A live command center that shows everything at a glance.

#### Layout

```
┌─ Topbar ────────────────────────────────────────────────────────┐
│  VibeCodePC    [● tunnel connected]  remote-url  [@user avatar] │
├─ Sidebar ─┬─ Main ────────────────────────────────────────────┤
│           │                                                     │
│  Dashboard│  ┌── SERVICES ──────────────────────────────────┐  │
│  Projects │  │  [opencode]  [NanoClaw]  [Cloudflare]  [Docker]│ │
│  IDE      │  │  ● 2 sessions ● active  ● stable  ● 2 running │ │
│  Agent    │  │  [Open IDE]  [Chat]     [Settings] [Manage]   │ │
│  GitHub   │  └──────────────────────────────────────────────┘  │
│  Settings │                                                     │
│           │  ┌── QUICK LAUNCH ─────────┐ ┌── VITALS ────────┐  │
│           │  │  my-saas-app  ts  main  │ │ CPU [████░] 42%  │  │
│           │  │  api-server   py  feat/ │ │ RAM [████░] 67%  │  │
│           │  │  + New Project          │ │ Disk[████░] 45%  │  │
│           │  └─────────────────────────┘ │ Temp 52°C        │  │
│           │                             │ Uptime 3d 14h     │  │
│           │  ┌── GITHUB ACTIVITY ──────┐ └─────────────────┘  │
│           │  │  PR #42 opened  2h ago  │ ┌── ACCESS ─────────┐ │
│           │  │  3 commits to main 4h   │ │ Local   [copy][QR]│ │
│           │  │  Issue #7 closed  1d    │ │ Remote  [copy][QR]│ │
│           │  │  [View all →]           │ └──────────────────┘  │
│           │  └─────────────────────────┘                       │
└───────────┴─────────────────────────────────────────────────────┘
```

#### Service Cards (`ServiceCard.vue`)

Each of the 4 service cards (opencode, NanoClaw, Cloudflare, Docker) shows:
- Pulsing color dot: green (running) / amber (starting) / red (error) / grey (stopped)
- Key metric (session count, message count, tunnel URL, container count)
- Primary action button (contextual: "Open IDE", "Chat", etc.)
- Secondary "..." menu: restart, logs, settings

Clicking the card body opens an inline log panel (last 50 lines, live-tailed via SSE).

#### System Vitals (`VitalsGraph.vue`)

- CPU %: sparkline graph (last 60 data points at 2 s interval via `/api/metrics/stream` SSE)
- RAM used/total
- Disk used/total
- Pi CPU temperature (reads `/sys/class/thermal/thermal_zone0/temp`)
- Uptime from `/proc/uptime`
- All values update in real time — no page refresh needed

#### GitHub Activity (`ActivityFeed.vue`)

- Fetches from `GET /api/github/activity` on mount and polls every 60 s
- Shows last 10 events: push, PR open/merge, issue open/close, review
- Event icon + repo name + short description + relative time
- Click → opens GitHub in new tab
- Hidden (with placeholder) if GitHub not connected

#### Quick Launch (`QuickLaunch.vue`)

- Last 4 opened projects as cards: name, language icon, branch, time since last opened
- One-click → navigates directly to `/app/ide/:projectId` (starts session)
- "+ New Project" card → `NewProjectModal.vue`

#### Access Panel (`AccessPanel.vue`)

- Two URL rows: Local + Remote
- Each: URL text (truncated) + copy icon + QR code popover
- Cloudflare connection badge: quick vs named tunnel mode, uptime
- "Upgrade to stable URL" link if in quick-tunnel mode

---

### Phase 7 — Projects & Settings (Week 6)

#### Projects View (`ProjectsView.vue`)

- Masonry grid of `ProjectCard.vue` components (3 cols desktop, 2 tablet, 1 mobile)
- Card content: project name, language badge, git branch, ahead/behind count, last opened
- Hover reveals: "Open IDE", "Chat about this", git status dot
- "New Project" FAB (fixed bottom-right)
- `NewProjectModal.vue` has two tabs:
  - **Local**: name + filesystem path picker (browseable via API) + language
  - **From GitHub**: search your repos inline (reuses github store) → one-click import

#### Settings View (`SettingsView.vue`)

Tabs: AI Providers | GitHub | Cloudflare | OpenCode | NanoClaw | System

- **AI Providers**: Provider cards, key rotation, live test, usage display
- **GitHub**: Connect/disconnect OAuth, webhook configuration, default clone path
- **Cloudflare**: Current mode (quick/named), tunnel URL, switch to named, custom domain
- **OpenCode**: Binary path, default provider, version, reinstall button
- **NanoClaw**: Container state, messaging platforms, agent persona name, restart
- **System**: Hostname, IP, Go version, versions table, `update.sh` trigger (shows live progress)

---

### Phase 8 — Installer & Packaging (Week 7)

**Goal**: One-command install that hands off a working URL.

```bash
curl -fsSL https://vibecodepc.com/install.sh | bash
```

`install.sh` sequence:

```
[1/7] Detecting system (OS, arch, hostname)...
[2/7] Installing git and Docker...                 ← skipped if present
[3/7] Downloading VibeCodePC binary (arm64)...     ← prebuilt Go binary from GitHub Releases
[4/7] Downloading cloudflared (arm64)...
[5/7] Creating data directory (~/.vibecodepc/)...
[6/7] Registering systemd services...
      vibecodepc.service       → Go binary :3000
      vibecodepc-tunnel.service → cloudflared quick tunnel → :3000
[7/7] Starting services & waiting for tunnel URL...

╔═══════════════════════════════════════════════════════╗
║  ✓ VibeCodePC is ready!                               ║
║                                                       ║
║  Local   http://raspberrypi.local:3000                ║
║  Remote  https://random-name.trycloudflare.com        ║
║                                                       ║
║  Open the Remote URL to run the setup wizard.         ║
║  The URL changes on restart until you set up a        ║
║  named tunnel in the wizard (free, takes 2 min).      ║
╚═══════════════════════════════════════════════════════╝
```

The installer downloads a prebuilt binary — **no Go compiler needed on the Pi**. Cross-compiled releases for `linux/arm64`, `linux/arm`, and `linux/amd64` are published to GitHub Releases via CI.

`update.sh`: download latest binary + systemctl restart both services + print new version
`uninstall.sh`: stop + disable services + rm app dir (warns before deleting, keeps projects)

---

## Integration Details

### OpenCode Integration

```
Browser (xterm.js) ──WS──► /ws/terminal/:projectId ──► creack/pty ──► opencode
                   ◄──WS──                          ◄──── stdout
```

- Config auto-written to `~/.config/opencode/config.json` from stored provider keys
- Each project uses its own CWD, so opencode loads that project's context
- Multiple browser tabs share the same PTY session for the same project
- opencode is a Node.js app; Node.js is installed by the wizard's Step 6

### NanoClaw Integration

```
AgentView ──POST /api/agent/message──► Go server ──► nanoclaw SQLite (insert)
          ◄─────── SSE stream ──────── Go server ◄── SQLite ticker 100 ms
                                               ↑
                          nanoclaw container reads → Claude → writes response
```

- nanoclaw's SQLite volume-mounted at `~/.vibecodepc/nanoclaw/data/`
- Go server accesses the file directly (same host filesystem)
- `web` registered as a virtual platform in nanoclaw post-clone patch

### GitHub OAuth Flow

```
Browser → GET /auth/github/start
       → 302 to github.com/login/oauth/authorize
       → user approves
       → github.com → GET /auth/github/callback?code=...
       → Go server exchanges code for token
       → token stored in keystore
       → 302 to /app/dashboard (or wizard step if mid-setup)
```

- Scopes: `repo`, `read:user` (no write to user data)
- Token refresh not needed (GitHub classic tokens don't expire unless revoked)
- `git push` uses token via git credential helper written to `~/.vibecodepc/.gitconfig`

### Cloudflare Tunnel

```
Internet ──► CF Edge ──► cloudflared ──► localhost:3000
```

Two modes in `cloudflare.go`:
- **Quick**: `cloudflared tunnel --url http://localhost:3000` (no account, ephemeral URL)
- **Named**: `cloudflared tunnel --token <token>` (account required, stable URL)

Mode is selected on startup based on whether a `cf_tunnel_token` setting exists in SQLite.

### Real-Time Metrics

```
GET /api/metrics/stream   →   SSE (text/event-stream)
  event: metrics
  data: { cpu: 42, ramUsedMb: 1340, ramTotalMb: 2000, diskUsedGb: 12, diskTotalGb: 32, tempC: 52, uptimeS: 123456 }
```

Go server reads `/proc/stat`, `/proc/meminfo`, `df`, `/sys/class/thermal/` every 2 seconds via `time.Ticker`. SSE handler respects `r.Context().Done()` for cleanup on disconnect.

---

## Environment Variables

```env
PORT=3000
HOST=0.0.0.0
APP_ENV=production
DATA_DIR=/home/pi/.vibecodepc/data

# GitHub OAuth App (register at github.com/settings/developers)
GITHUB_CLIENT_ID=<your_client_id>
GITHUB_CLIENT_SECRET=<your_client_secret>
GITHUB_REDIRECT_URI=http://localhost:3000/auth/github/callback
```

All AI provider keys are stored encrypted in SQLite, not in `.env`.

---

## Security Considerations

- **Auth**: App is open on LAN by default (Raspberry Pi use case). After Cloudflare tunnel is active, enable optional password lock (bcrypt hash in SQLite) for internet exposure.
- **GitHub token**: Stored AES-256-GCM encrypted. Only `repo` + `read:user` scopes — no org admin, no delete.
- **Key encryption**: Machine-derived key (SHA256 of hostname + primary MAC). Never logged. Standard `crypto` package.
- **Terminal isolation**: Each project gets its own PTY as the app user — not root.
- **Docker socket**: Required for NanoClaw. Warn user on dashboard that docker group membership is equivalent to root.
- **Cloudflare tunnel**: End-to-end encrypted by Cloudflare. Zero open ports on the Pi.
- **Security headers**: Custom Chi middleware sets CSP, HSTS, X-Frame-Options, X-Content-Type-Options.
- **Git credentials**: GitHub token written to `~/.vibecodepc/.gitconfig` (not global `~/.gitconfig`) and referenced via `GIT_CONFIG` env on git operations.

---

## Milestones

| # | Milestone | Target |
|---|---|---|
| M1 | Foundation: Go server + Vue shell + GitHub OAuth | End of Week 1 |
| M2 | Full automated setup wizard (all 8 steps) | End of Week 2 |
| M3 | OpenCode terminal embedded, per-project sessions | End of Week 3 |
| M4 | NanoClaw chat + GitHub integration + GitPanel | End of Week 4–5 |
| M5 | Real-time dashboard with metrics + activity feed | End of Week 5–6 |
| M6 | Projects, Settings, polish, PWA | End of Week 6 |
| M7 | One-line installer + prebuilt binaries + systemd | End of Week 7 |
