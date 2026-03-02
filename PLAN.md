# VibeCodePC.com — Development Plan

> Turn any Raspberry Pi into a personal AI coding powerhouse.
> Anyone can set it up. Anyone can vibe-code.

---

## Vision

VibeCodePC.com is a self-hosted web application that transforms a Raspberry Pi (or any Linux machine) into a personal AI coding station. It exposes a polished web UI that walks the user through device setup, then provides a unified dashboard for managing AI-powered coding sessions, a conversational AI agent, projects, and API keys — all accessible remotely via a free Cloudflare tunnel.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Vue 3, Vite, Tailwind CSS v4, Pinia, Vue Router |
| Terminal embed | xterm.js + xterm-addon-fit + xterm-addon-web-links |
| Backend | Node.js 20+, Fastify v5, TypeScript |
| WebSockets | `@fastify/websocket` (terminal I/O, agent chat) |
| Database | SQLite via `better-sqlite3` |
| Process mgmt | `node-pty` (OpenCode terminal sessions) |
| Docker mgmt | Dockerode |
| Key storage | AES-256-GCM encrypted fields in SQLite |
| Tunneling | cloudflared (Cloudflare Tunnel binary) |
| AI Coding | opencode (anomalyco/opencode) |
| AI Agent | nanoclaw (qwibitai/nanoclaw) |
| Installer | Bash script + optional Docker Compose |

---

## Architecture Overview

```
Browser (Vue 3)
     │
     │  HTTP / WebSocket
     ▼
Fastify Server (Node.js)
     ├── /api/setup        ← Setup wizard state machine
     ├── /api/projects     ← Project CRUD
     ├── /api/settings     ← AI keys & config
     ├── /api/agent        ← NanoClaw bridge (SSE + POST)
     ├── /ws/terminal/:id  ← xterm.js ↔ node-pty ↔ opencode
     └── /* (SPA)          ← Serves Vue build
     │
     ├── OpenCode Service
     │     └── node-pty spawns `opencode` per project
     │
     ├── NanoClaw Service
     │     ├── Dockerode manages container lifecycle
     │     └── SQLite bridge for web chat ↔ nanoclaw queue
     │
     ├── Cloudflare Tunnel Service
     │     └── Spawns cloudflared process, monitors status
     │
     └── SQLite DB
           ├── setup_state
           ├── projects
           ├── settings (encrypted keys)
           └── agent_messages
```

---

## Directory Structure

```
vibecodepc/
├── PLAN.md
├── CLAUDE.md
├── README.md
├── package.json              # Workspace root (pnpm workspaces)
├── pnpm-workspace.yaml
│
├── server/                   # Fastify backend
│   ├── package.json
│   ├── tsconfig.json
│   └── src/
│       ├── index.ts          # Server entry point
│       ├── config.ts         # Env vars & defaults
│       ├── db/
│       │   ├── index.ts      # better-sqlite3 singleton
│       │   ├── schema.ts     # Table definitions & migrations
│       │   └── crypto.ts     # AES-256-GCM key encryption helpers
│       ├── routes/
│       │   ├── setup.ts      # GET/POST /api/setup
│       │   ├── projects.ts   # CRUD /api/projects
│       │   ├── settings.ts   # /api/settings (keys, providers)
│       │   ├── agent.ts      # /api/agent (SSE stream + POST message)
│       │   └── terminal.ts   # WS /ws/terminal/:projectId
│       └── services/
│           ├── setup.ts      # Wizard step state machine
│           ├── system-check.ts # Checks: Docker, Node, disk, RAM
│           ├── opencode.ts   # Install, launch, kill opencode per project
│           ├── nanoclaw.ts   # Clone, configure, Docker run nanoclaw
│           ├── cloudflare.ts # Download cloudflared, tunnel lifecycle
│           └── keystore.ts   # Encrypt/decrypt AI provider keys
│
├── client/                   # Vue 3 frontend
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── index.html
│   └── src/
│       ├── main.ts
│       ├── App.vue
│       ├── router/
│       │   └── index.ts      # Vue Router – guards setup completion
│       ├── stores/
│       │   ├── setup.ts      # Pinia: wizard step & state
│       │   ├── projects.ts   # Pinia: project list & active
│       │   ├── settings.ts   # Pinia: provider config
│       │   └── agent.ts      # Pinia: chat messages & connection
│       ├── views/
│       │   ├── setup/
│       │   │   ├── WizardLayout.vue    # Progress bar shell
│       │   │   ├── StepWelcome.vue     # What is VibeCodePC?
│       │   │   ├── StepSystemCheck.vue # Docker / RAM / disk checks
│       │   │   ├── StepCloudflare.vue  # Tunnel token input & status
│       │   │   ├── StepProviders.vue   # AI API key entry
│       │   │   ├── StepOpenCode.vue    # Install opencode, pick default
│       │   │   ├── StepNanoClaw.vue    # Agent name, whatsapp pairing
│       │   │   └── StepComplete.vue    # 🎉 Launch dashboard
│       │   ├── DashboardView.vue       # Service status grid
│       │   ├── IDEView.vue             # Embedded opencode terminal
│       │   ├── AgentView.vue           # NanoClaw web chat
│       │   ├── ProjectsView.vue        # Project browser/creator
│       │   └── SettingsView.vue        # Keys, providers, tunnel
│       └── components/
│           ├── layout/
│           │   ├── AppSidebar.vue
│           │   ├── AppTopbar.vue
│           │   └── ServiceStatusBar.vue
│           ├── terminal/
│           │   └── TerminalPane.vue    # xterm.js wrapper
│           ├── agent/
│           │   ├── ChatBubble.vue
│           │   ├── ChatInput.vue
│           │   └── AgentStatus.vue
│           ├── projects/
│           │   ├── ProjectCard.vue
│           │   └── NewProjectModal.vue
│           └── ui/
│               ├── StatusBadge.vue
│               ├── KeyInput.vue        # Masked API key field
│               ├── CheckItem.vue       # System check row
│               └── StepIndicator.vue
│
├── docker/
│   ├── docker-compose.yml    # Optional: run whole stack in Docker
│   └── nanoclaw/
│       └── Dockerfile        # Nanoclaw container wrapper
│
└── scripts/
    ├── install.sh            # One-line installer for Raspberry Pi
    ├── update.sh             # Pull latest & restart
    └── uninstall.sh
```

---

## Development Phases

### Phase 1 — Foundation (Week 1)

**Goal**: Bare-metal server + Vue shell running locally.

- [ ] Init pnpm workspace with `server/` and `client/` packages
- [ ] Configure TypeScript strict mode in both packages
- [ ] Fastify server with static file serving of Vue build
- [ ] SQLite schema: `setup_state`, `projects`, `settings`, `agent_messages`
- [ ] Vue 3 + Vite + Tailwind CSS v4 scaffold
- [ ] Vue Router with two root guards: `/setup/*` and `/app/*`
- [ ] Setup wizard shell: `WizardLayout` with `StepIndicator`
- [ ] Pinia store for wizard state, persisted to server on each step
- [ ] Dev proxy: Vite → Fastify for API calls during development
- [ ] ESLint + Prettier config (shared rules)

**Deliverable**: `pnpm dev` boots the wizard UI.

---

### Phase 2 — Setup Wizard (Week 2)

**Goal**: Full working wizard that actually configures the device.

#### Step 1 — Welcome
- Animated hero: Raspberry Pi + code terminal visual
- Explains what VibeCodePC does in 30 seconds
- "Let's go" CTA

#### Step 2 — System Check (`system-check.ts`)
- Check: Docker installed & running → `docker info`
- Check: Node.js ≥ 20 → `process.version`
- Check: Available RAM ≥ 1 GB
- Check: Available disk ≥ 5 GB
- Check: Internet connectivity → fetch `1.1.1.1`
- Each check has pass/fail/pending state with animated spinner
- Cannot proceed until all critical checks pass

#### Step 3 — Cloudflare Tunnel
- The quick tunnel (`*.trycloudflare.com`) is **already running** from the installer
- This step shows its status and URL, and offers an optional upgrade to a named tunnel:
  - UI: Link to `dash.cloudflare.com` → guide to create a named free tunnel
  - Token input field → validate with `cloudflared tunnel --token <t> info`
  - Store token encrypted in SQLite; cloudflared restarts in named-tunnel mode
  - Named tunnel gives a stable URL that survives reboots
- Skippable — quick tunnel is sufficient for personal use
- Show both access URLs at all times:
  - Local: `http://<hostname>.local:3000` (LAN, instant)
  - Public: `https://<tunnel>.trycloudflare.com` (internet, Cloudflare-secured)

#### Step 4 — AI Providers
- Provider cards: Anthropic (required for NanoClaw), OpenAI, Google Gemini, Ollama (local)
- Per-provider: API key input (masked), test button → validates via provider's `/models` endpoint
- At least one provider must be configured
- Anthropic key flagged as "Recommended – powers the AI agent"

#### Step 5 — OpenCode
- Install opencode: `npm install -g opencode` (shows live stdout)
- Select default AI provider for opencode sessions
- Verify: `opencode --version`
- Show ASCII opencode logo in terminal preview

#### Step 6 — NanoClaw (Optional)
- Skip-able step for users who only want the IDE
- Clone `qwibitai/nanoclaw` into `~/.vibecodepc/nanoclaw/`
- Configure `.env` from stored API keys
- Show QR code / pairing URL for WhatsApp (via Baileys)
- Or select Telegram/Discord/Slack channel
- Start nanoclaw Docker container

#### Step 7 — Complete
- Confetti animation
- Summary card showing both access methods:
  - Local network: `http://<hostname>.local`
  - Cloudflare tunnel: `https://<tunnel>.trycloudflare.com`
- Copy button + QR code for each URL
- Active providers summary, agent status
- "Open Dashboard" → `/app/dashboard`

---

### Phase 3 — OpenCode IDE Integration (Week 3)

**Goal**: In-browser terminal running opencode per project.

- `opencode.ts` service:
  - `start(projectId, cwd)` → spawn `opencode` via `node-pty` in project directory
  - `kill(projectId)` → terminate process
  - `resize(projectId, cols, rows)` → PTY resize
  - Session registry: `Map<projectId, IPty>`
- WebSocket route `/ws/terminal/:projectId`:
  - On connect: attach to existing session or start new one
  - `data` frames: browser → PTY stdin
  - PTY stdout → broadcast to all connected clients for that project
  - `resize` frames: update PTY dimensions
- `TerminalPane.vue`:
  - `@xterm/xterm` + `@xterm/addon-fit` + `@xterm/addon-web-links`
  - Custom Dracula-inspired theme matching app dark mode
  - Auto-fit on window resize via ResizeObserver
  - Reconnect on WS disconnect with exponential backoff
- `IDEView.vue`:
  - Split layout: file tree sidebar (read-only, future) + terminal
  - Project selector dropdown in topbar
  - "New Session" and "Kill Session" buttons

---

### Phase 4 — NanoClaw Agent Chat (Week 4)

**Goal**: Web chat UI that talks to the nanoclaw agent.

The challenge: nanoclaw is queue-based (SQLite). We bridge it:

- `nanoclaw.ts` service:
  - Writes user messages directly into nanoclaw's `messages` SQLite table as an inbound platform message
  - Polls for outbound responses via `SELECT` with timestamp cursor
  - Emits `agent_message` events via Node.js EventEmitter
- SSE endpoint `GET /api/agent/stream`:
  - Sends `data:` events for new agent messages
  - Heartbeat ping every 15 s
- `POST /api/agent/message` — user sends a message
- `AgentView.vue`:
  - Chat UI: bubbles, timestamps, typing indicator
  - Connects to SSE stream on mount
  - Markdown rendering for agent responses (`markdown-it`)
  - Code blocks with syntax highlighting (`shiki`)
  - "Clear history" and agent status indicator
  - Mobile-first design: full-screen chat on small screens

---

### Phase 5 — Project Management (Week 4–5)

**Goal**: Create, browse, and switch between coding projects.

- `ProjectsView.vue`:
  - Grid of `ProjectCard.vue` components
  - Each card: name, language badge, last-opened, git branch (if applicable)
  - "New Project" FAB → `NewProjectModal.vue`
- `NewProjectModal.vue`:
  - Name, path on filesystem, git clone URL (optional), default AI provider
  - Creates directory, optionally runs `git clone`
  - Writes project row to SQLite
- `projects.ts` route:
  - `GET /api/projects` — list all
  - `POST /api/projects` — create
  - `DELETE /api/projects/:id` — remove (does NOT delete files)
  - `GET /api/projects/:id/status` — git status, file count, disk usage
- Clicking a project → navigates to `/app/ide/:projectId`
- Router guard starts opencode session for that project

---

### Phase 6 — Settings (Week 5)

**Goal**: Manage keys, providers, tunnel, and system info.

- `SettingsView.vue` with tab navigation:
  - **AI Providers**: Re-enter/rotate keys, toggle providers, test connection
  - **Cloudflare Tunnel**: Show status (connected/disconnected), restart tunnel, change domain
  - **OpenCode**: Default provider, opencode version, reinstall
  - **NanoClaw**: Agent config (name, platforms), restart container
  - **System**: Hostname, IP, RAM/disk usage, Node & Docker versions, update button
- `keystore.ts`: All keys AES-256-GCM encrypted with a machine key derived from hostname + MAC address
- `KeyInput.vue`: shows `••••••••` with reveal toggle and copy button

---

### Phase 7 — Dashboard & Polish (Week 6)

**Goal**: Award-winning first impression.

- `DashboardView.vue`:
  - Service health grid: OpenCode, NanoClaw, Cloudflare Tunnel, Docker
  - Quick-launch buttons: "Start Coding" → recent project, "Chat with Agent"
  - Recent projects list (last 5 opened)
  - System vitals: CPU %, RAM %, disk usage (via `/proc` or `os` module)
  - Dual access panel: local URL (`http://<hostname>.local:3000`) + tunnel URL, each with copy + QR code
- Design system:
  - Full dark mode (default), light mode toggle
  - Color palette: deep indigo base, electric violet accent, neon green success
  - Smooth page transitions (Vue's `<Transition>`)
  - Micro-animations: status badges pulse, progress bars animate in
  - Inter font for UI, JetBrains Mono for terminal and code
- PWA: `vite-plugin-pwa` for add-to-homescreen on mobile
- Responsive: sidebar collapses to bottom nav on mobile

---

### Phase 8 — Installer & Packaging (Week 6–7)

**Goal**: One-command install on a fresh Raspberry Pi.

```bash
curl -fsSL https://vibecodepc.com/install.sh | bash
```

`install.sh`:
1. Detect OS / arch (arm64, armv7, amd64)
2. Install Node.js 20 LTS via `nvm` if missing
3. Install pnpm and git if missing
4. Clone this repo into `~/.vibecodepc/app`
5. `pnpm install --prod`
6. `pnpm build`
7. Download `cloudflared` binary for detected arch into `~/.vibecodepc/bin/`
8. Register two systemd services:
   - `vibecodepc.service` — the Node.js app on port 3000
   - `vibecodepc-tunnel.service` — cloudflared quick tunnel pointing to `localhost:3000`
9. Start both services
10. Wait for `vibecodepc-tunnel.service` log to contain the `trycloudflare.com` URL (up to 15 s)
11. Print:
    ```
    ✓ VibeCodePC installed!

      Local:   http://<hostname>.local:3000
      Remote:  https://<random>.trycloudflare.com   ← open this to run the setup wizard

    The remote URL changes on each restart until you configure a named tunnel in the wizard.
    ```

`update.sh`: git pull + rebuild + restart service
`uninstall.sh`: stop service + remove files (keeps project directories)

---

## Integration Details

### OpenCode Integration

```
Browser (xterm.js) ──WS──► Fastify WS route ──► node-pty ──► opencode process
                   ◄──WS──              ◄─────────── PTY stdout
```

- opencode is run as: `opencode` (no flags needed, uses project CWD)
- Provider config via opencode's `~/.config/opencode/config.json` or env vars
- Each browser tab connecting to the same project shares the same PTY session

### NanoClaw Integration

```
AgentView.vue ──POST /api/agent/message──► Fastify ──► nanoclaw SQLite (insert)
              ◄─────── SSE stream ─────── Fastify ◄── SQLite poller (100ms interval)
                                                  (nanoclaw reads → processes → writes response)
```

- nanoclaw runs in a Docker container with its SQLite file volume-mounted to host
- The Fastify server accesses the same SQLite file directly (host-side path)
- Web source ID is registered as a virtual platform in nanoclaw config

### Cloudflare Tunnel

```
Internet ──► Cloudflare Edge ──► cloudflared ──► localhost:3000
```

Two tunnel modes — both managed by `cloudflare.ts`:

**Quick tunnel** (no account, started by installer):
- `cloudflared tunnel --url http://localhost:3000`
- Cloudflare assigns a random `*.trycloudflare.com` URL instantly
- No token, no account needed — works out of the box
- URL changes each restart — suitable for the setup wizard and light use

**Named tunnel** (optional, configured in wizard):
- User creates a tunnel on `dash.cloudflare.com` and pastes the token
- Stable URL, optionally on a custom domain
- `cloudflared` started with `--token <token>` flag

On app startup, `cloudflare.ts` checks which mode is active and starts accordingly.
Both modes: tunnel and local LAN (`http://<hostname>.local:3000`) coexist simultaneously.

---

## Environment Variables

```env
# Server
PORT=3000
HOST=0.0.0.0
DATA_DIR=~/.vibecodepc/data        # SQLite, configs

# Derived at runtime (not user-set)
MACHINE_KEY=<derived>              # For AES key derivation
```

All AI provider keys are stored encrypted in SQLite, not in `.env`.

---

## Security Considerations

- **No public auth by default**: The app is intended to be LAN-only until Cloudflare tunnel is set up. After setup, strongly recommend enabling the built-in password lock (bcrypt hash in SQLite).
- **Key encryption**: AES-256-GCM with a machine-derived key. Keys never logged.
- **Terminal isolation**: Each project gets its own PTY, running as the app user.
- **Docker socket**: Required for NanoClaw management. Warn user of implications.
- **Cloudflare tunnel**: Encrypted end-to-end by Cloudflare. No port forwarding needed.
- **Content Security Policy**: Fastify helmet plugin for HTTP headers.

---

## Milestones

| # | Milestone | Target |
|---|---|---|
| M1 | Dev environment running, wizard shell | End of Week 1 |
| M2 | Full setup wizard, device configured | End of Week 2 |
| M3 | OpenCode terminal embedded and working | End of Week 3 |
| M4 | NanoClaw chat UI live | End of Week 4 |
| M5 | Projects + Settings complete | End of Week 5 |
| M6 | Dashboard, PWA, polish | End of Week 6 |
| M7 | One-line installer, systemd service | End of Week 7 |
