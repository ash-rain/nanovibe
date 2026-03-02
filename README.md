# VibeCodePC.com

**Turn any Raspberry Pi into a personal AI coding powerhouse.**
Set it up once. Code from anywhere.

---

## What is this?

VibeCodePC is a self-hosted web app that runs on your Raspberry Pi (or any Linux machine). It gives you:

- A **guided setup wizard** that configures everything from scratch — Docker, Cloudflare tunnel, AI provider keys — no command line skills needed.
- An **embedded AI coding environment** powered by [opencode](https://github.com/anomalyco/opencode) — a terminal-based AI coding agent that works with Claude, OpenAI, Gemini, or local models.
- A **web chat interface** for [nanoclaw](https://github.com/qwibitai/nanoclaw) — a lightweight Claude-powered AI agent that can also connect to WhatsApp, Telegram, and more.
- **Project management** — create, switch, and track all your coding projects in one place.
- **AI provider key management** — securely store and rotate keys for every provider, system-wide.
- **Remote access** via a free Cloudflare tunnel — no port forwarding, no static IP required.

---

## Screenshots

> _Coming soon — wizard, IDE, and agent chat views._

---

## Requirements

| Requirement | Minimum |
|---|---|
| Hardware | Raspberry Pi 4 (2 GB RAM+) or any Linux x86/ARM machine |
| OS | Raspberry Pi OS (64-bit), Ubuntu 22.04+, Debian 12+ |
| Node.js | 20 LTS or later |
| Docker | 24+ (for NanoClaw agent container) |
| Disk space | 5 GB free |
| Network | Internet connection for setup |

---

## Quick Install

On your Raspberry Pi (or Linux machine), run:

```bash
curl -fsSL https://vibecodepc.com/install.sh | bash
```

The installer sets up everything — including a Cloudflare tunnel — and prints your access URLs at the end:

```
✓ VibeCodePC installed!

  Local:   http://<your-pi-hostname>.local:3000
  Remote:  https://<random>.trycloudflare.com   ← open this to run the setup wizard
```

Open the remote URL from any device to run the setup wizard. No need to be on the same network.

---

## Manual Install

```bash
# Clone the repo
git clone https://github.com/vibecodepc/app.git ~/.vibecodepc/app
cd ~/.vibecodepc/app

# Install dependencies
npm install -g pnpm
pnpm install

# Build
pnpm build

# Start
node server/dist/index.js
```

Open `http://localhost:3000` and follow the wizard.

---

## Setup Wizard

The first time you open the app, you'll be guided through 7 steps:

| Step | What happens |
|---|---|
| Welcome | Overview of what VibeCodePC does |
| System Check | Verifies Docker, Node.js, RAM, disk, and internet |
| Cloudflare Tunnel | Optionally upgrade to a **stable named tunnel** — the quick tunnel is already running |
| AI Providers | Enter API keys (Anthropic, OpenAI, Google, or Ollama) |
| OpenCode | Installs the AI coding agent and picks your default model |
| NanoClaw | (Optional) Sets up your personal AI assistant |
| Complete | Both access URLs with copy buttons and QR codes |

After setup, you'll never need to touch the command line again.

---

## Features

### AI Coding IDE

The IDE view embeds a full terminal running [opencode](https://github.com/anomalyco/opencode) — an AI coding agent that:

- Works with any AI provider (Claude, GPT-4, Gemini, local Ollama models)
- Has full access to your project files
- Can write, edit, run, and debug code with you
- Runs in your browser — accessible from any device on your network (or remotely via your Cloudflare URL)

Each project gets its own isolated opencode session.

### AI Agent Chat

The Agent view gives you a clean web chat interface to [nanoclaw](https://github.com/qwibitai/nanoclaw) — a personal Claude-powered agent that:

- Answers questions, helps you plan features, reviews code
- Can also connect to WhatsApp, Telegram, Discord, and more
- Runs in a secure Docker container
- Remembers context per conversation

### Project Management

- Create new projects with one click (optionally clone a Git repo)
- See git branch, last commit, and disk usage at a glance
- Switch between projects instantly — each gets its own AI coding session

### AI Key Management

- Store API keys for Anthropic, OpenAI, Google, and Ollama
- Keys are encrypted with AES-256-GCM using a machine-derived key
- Test any key with a live provider check
- Rotate or remove keys from the UI anytime

### Dual Access — LAN + Internet

The app runs on **port 3000** and is reachable two ways from day one:

| Method | URL | Notes |
|---|---|---|
| Local network | `http://<hostname>.local:3000` | Zero-latency on your home/office LAN |
| Cloudflare tunnel | `https://<random>.trycloudflare.com` | Internet access, set up by the installer automatically |

The installer starts a **Cloudflare quick tunnel** (no account needed) and prints the URL. You can optionally upgrade to a **named tunnel** in the wizard for a stable URL on your own domain.

Both URLs are live simultaneously and shown on the dashboard with copy buttons and QR codes.

---

## Architecture

```
http://<hostname>.local:3000              (LAN)
https://<tunnel>.trycloudflare.com        (internet, via cloudflared quick/named tunnel)
           │
           │  HTTP / WebSocket / SSE
           ▼
     Fastify Server :3000
           │
           ├── OpenCode (node-pty) ──► terminal sessions per project
           ├── NanoClaw (Docker) ───► AI agent in isolated container
           ├── cloudflared ────────► Cloudflare tunnel → :3000
           └── SQLite ─────────────► projects, settings, messages
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Vue 3, Vite, Tailwind CSS v4, Pinia |
| Backend | Node.js 20, Fastify v5, TypeScript |
| Terminal | xterm.js + node-pty |
| Database | SQLite (better-sqlite3) |
| AI Coding | opencode (anomalyco/opencode) |
| AI Agent | nanoclaw (qwibitai/nanoclaw) |
| Tunneling | cloudflared (Cloudflare free tier) |
| Container | Docker |

---

## Security

- **Keys never leave your machine** — encrypted at rest with a machine-derived key
- **Cloudflare tunnel** provides end-to-end encryption — no open ports needed
- **Terminal isolation** — each project runs in its own sandboxed process
- **NanoClaw in Docker** — agent executes in an isolated container
- **Optional password lock** — protect the web UI with a bcrypt-hashed password

---

## Updating

```bash
~/.vibecodepc/app/scripts/update.sh
```

Or from the Settings view → System → "Check for Updates".

---

## Uninstalling

```bash
~/.vibecodepc/app/scripts/uninstall.sh
```

This removes the app and systemd service but **does not delete your projects**.

---

## Roadmap

- [ ] Password-protected web UI
- [ ] Multiple Cloudflare tunnels per project
- [ ] VS Code extension mode (opencode as LSP)
- [ ] Mobile-optimized chat view (PWA)
- [ ] Agent-to-agent collaboration (nanoclaw swarms)
- [ ] One-click project templates (Python, TypeScript, Rust, etc.)
- [ ] GitHub/GitLab integration (PR creation, issue linking)
- [ ] Usage dashboard (token counts, costs per provider)

---

## Contributing

PRs welcome. See [PLAN.md](./PLAN.md) for the development roadmap and [CLAUDE.md](./CLAUDE.md) for coding conventions.

```bash
# Development setup (no sudo needed — dev defaults to port 3000)
pnpm install
pnpm dev        # Server on :3000, Vite client on :5173
```

---

## License

MIT — build something amazing.

---

_Built with ❤️ for makers, hackers, and anyone who wants to vibe-code from their Pi._
