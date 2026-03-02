# VibeCodePC.com

**Turn any Raspberry Pi into a personal AI coding powerhouse.**
One command installs everything. The app does the rest.

---

## What is this?

VibeCodePC is a self-hosted web application that transforms a Raspberry Pi (or any Linux machine) into a complete AI-assisted coding station. It installs itself, configures itself, and hands you a live URL — then guides you through the rest with an automated setup wizard.

---

## Install

On your Raspberry Pi (or any Debian/Ubuntu machine):

```bash
curl -fsSL https://vibecodepc.com/install.sh | bash
```

The installer runs silently, then prints:

```
╔═══════════════════════════════════════════════════════╗
║  ✓ VibeCodePC is ready!                               ║
║                                                       ║
║  Local   http://raspberrypi.local:3000                ║
║  Remote  https://random-name.trycloudflare.com        ║
║                                                       ║
║  Open the Remote URL to run the setup wizard.         ║
╚═══════════════════════════════════════════════════════╝
```

Open the Remote URL from **any device, anywhere** — your phone, laptop, a friend's computer. The setup wizard is already live.

---

## Requirements

| Requirement | Minimum |
|---|---|
| Hardware | Raspberry Pi 4 (2 GB RAM) or any Linux x86/ARM machine |
| OS | Raspberry Pi OS 64-bit, Ubuntu 22.04+, Debian 12+ |
| Network | Internet connection for install |

Everything else (Node.js, Docker, git, cloudflared) is installed automatically.

---

## Setup Wizard

Eight steps, almost entirely automated. Watch the progress bars — you're only needed for your API keys and GitHub login.

| Step | What the app does automatically | What you do |
|---|---|---|
| Welcome | Detects hostname, personalizes the greeting | Click "Start" |
| System Check | Runs checks; offers one-click "Fix" for any failures (installs Docker, Node, git) | Watch / approve fixes |
| Cloudflare Tunnel | Already running — shows live URL and QR code | Optionally upgrade to a stable named tunnel |
| GitHub | OAuth connect flow | Click "Connect GitHub", approve in popup |
| AI Providers | Scans env vars for existing keys, auto-tests each one | Paste any missing keys |
| OpenCode | Auto-installs, auto-configures from your keys | Choose default AI provider |
| NanoClaw | Auto-clones, auto-builds Docker image, auto-writes config | Optionally pair with WhatsApp/Telegram |
| Complete | Shows both URLs with QR codes, confetti | Click "Open Dashboard" |

After the wizard, you never need the command line again.

---

## Features

### Real-Time Dashboard

A live command center that updates every 2 seconds — no page refresh:

- **Service cards**: opencode, NanoClaw, Cloudflare, Docker — each showing live status, key metrics, and quick-action buttons. Click any card to tail its live logs.
- **System vitals**: CPU %, RAM, disk, Pi temperature, uptime — visualised as sparkline graphs
- **GitHub activity feed**: recent pushes, PR opens, issue closes — straight from your account
- **Quick launch**: your 4 most recently opened projects — one click to open in the IDE
- **Access panel**: both your local and Cloudflare URLs with copy and QR code buttons

### Embedded AI Coding IDE

Runs [opencode](https://github.com/anomalyco/opencode) — a full terminal AI coding agent:

- In-browser terminal (xterm.js) connected over WebSocket to a real PTY on your Pi
- Works with Claude, GPT-4, Gemini, or local Ollama models
- Each project gets its own isolated session
- Inline git panel: branch selector, staged/unstaged files, diff preview, commit & push

### GitHub Integration

GitHub is a first-class citizen:

- **Repo browser**: search and browse all your GitHub repos with language badges, stars, last-pushed
- **One-click import**: clone any repo directly from the UI → auto-creates a project
- **Activity feed**: recent events from your account — pushes, PR opens/merges, issues
- **PR creation**: create pull requests from the current branch without leaving the app
- **In-IDE git panel**: full git workflow (status, diff, commit, push, pull, branch switch) in a collapsible sidebar

### AI Agent Chat

A web chat interface backed by [nanoclaw](https://github.com/qwibitai/nanoclaw) — a Claude-powered agent:

- Chat UI with markdown rendering and syntax-highlighted code blocks
- Context-aware: associate chats with a specific project
- Optionally also connects to WhatsApp, Telegram, Discord, or Slack
- Runs in an isolated Docker container

### Project Management

- Project grid: name, language, git branch, ahead/behind count, last opened
- New projects: create from a local path or import from GitHub
- Instant switch: click any project to jump straight into the IDE

### AI Key Management

- Support for Anthropic, OpenAI, Google Gemini, and Ollama
- Keys auto-detected from environment variables and tested immediately
- AES-256-GCM encrypted at rest with a machine-derived key
- Rotate or remove any key from Settings anytime

---

## Architecture

```
http://raspberrypi.local:3000        (LAN, instant)
https://<tunnel>.trycloudflare.com   (internet, via cloudflared — set up by installer)
              │
              │  HTTP / WebSocket / SSE
              ▼
        Fastify Server :3000
              │
              ├── opencode (node-pty)  ─► per-project terminal sessions
              ├── NanoClaw (Docker)    ─► Claude agent, SQLite bridge
              ├── GitHub Service       ─► Octokit OAuth + repo/PR API
              ├── Git Service          ─► simple-git per-project operations
              ├── Metrics Service      ─► CPU/RAM/disk/temp SSE stream
              ├── cloudflared          ─► quick or named tunnel → :3000
              └── SQLite              ─► projects, settings, messages, auth
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Vue 3, Vite, Tailwind CSS v4, Pinia |
| Backend | Node.js 20, Fastify v5, TypeScript |
| Terminal | xterm.js + node-pty |
| Git | simple-git |
| GitHub API | Octokit (`@octokit/rest`) |
| Database | SQLite (better-sqlite3) |
| AI Coding | opencode (anomalyco/opencode) |
| AI Agent | nanoclaw (qwibitai/nanoclaw) |
| Tunneling | cloudflared (Cloudflare free tier) |
| Containers | Docker |

---

## Dual Access — LAN + Internet

The app is reachable two ways simultaneously, from the moment the installer completes:

| Method | URL | Notes |
|---|---|---|
| Local network | `http://<hostname>.local:3000` | Zero-latency on your LAN |
| Cloudflare tunnel | `https://<random>.trycloudflare.com` | Internet access, set up automatically |

The installer starts a **Cloudflare quick tunnel** (no account required). During the wizard you can optionally upgrade to a **named tunnel** (free Cloudflare account) for a stable URL across reboots.

Both URLs are shown on the dashboard with copy buttons and QR codes.

---

## Security

- **Keys encrypted**: AES-256-GCM with a machine-derived key (never stored or logged)
- **GitHub token**: minimal scopes (`repo`, `read:user`); stored encrypted
- **Cloudflare tunnel**: end-to-end TLS; zero open ports on your Pi
- **Terminal isolation**: each project's opencode runs as the app user in its own PTY
- **Docker isolation**: NanoClaw agent runs in a Docker container
- **Optional password lock**: protect the web UI with a bcrypt password (Settings → System)

---

## Updating

```bash
~/.vibecodepc/app/scripts/update.sh
```

Or from the app: Settings → System → "Check for Updates" → live progress log.

## Uninstalling

```bash
~/.vibecodepc/app/scripts/uninstall.sh
```

Removes the app and systemd services. Does **not** delete your project directories.

---

## Contributing

```bash
# Development (no sudo needed — uses port 3000)
pnpm install
pnpm dev        # Server :3000, Vite client :5173
```

See [PLAN.md](./PLAN.md) for the full development roadmap and [CLAUDE.md](./CLAUDE.md) for coding conventions.

---

## Roadmap

- [ ] Password-protected web UI (bcrypt + session cookie)
- [ ] VS Code extension mode (opencode LSP integration)
- [ ] Mobile-optimised PWA (add-to-homescreen)
- [ ] Agent swarms (nanoclaw multi-agent collaboration)
- [ ] One-click project templates (Next.js, FastAPI, Rust, etc.)
- [ ] GitHub Actions status in dashboard
- [ ] Token usage + cost tracking per provider
- [ ] Multi-user support (per-user API keys + project isolation)

---

## License

MIT — build something amazing.

---

_Built for makers, hackers, and anyone who wants to vibe-code from their Pi._
