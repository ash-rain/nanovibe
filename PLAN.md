# VibeCodePC.com — Development Plan

> Your Pi. Your AI. Your code.

---

## Vision

VibeCodePC.com is a self-hosted AI coding station built on a single principle: **complexity belongs to the machine, not the user**. It installs in one command, configures itself, and hands you a live URL. From that moment forward, the experience is a dashboard — not a terminal, not a config file, not a stack overflow search.

The product is opinionated. It makes choices so users don't have to. It auto-detects, auto-installs, auto-configures, and auto-advances. Every transition is earned, every animation purposeful, every empty state an invitation. The app should feel like it was designed by someone who cared deeply about every pixel — because it was.

---

## Design Philosophy

### Complexity is invisible. Beauty is not optional. Every moment is designed.

**1. The machine does the work.**
The user should never feel like they are operating software. They should feel like they are *watching something happen for them*. Progress bars aren't loading states — they're theatre. Every automated action earns trust.

**2. One primary action per screen.**
At any moment, there is exactly one thing the user should do next. That action gets violet. Everything else recedes. The eye should never have to search.

**3. Every transition tells a story.**
A panel doesn't appear — it arrives. A check doesn't pass — it resolves. The terminal doesn't connect — it opens. Motion is not decoration; it communicates causality and sequence.

**4. Live data should feel alive.**
Numbers don't jump — they count. Lines don't redraw — they flow. Status dots don't blink — they breathe. The dashboard is a living thing, and users should feel its pulse.

**5. Silence is a feature.**
No success toasts. No notification banners. When something works, the interface just changes — confidently, cleanly. Alerts are reserved for situations that genuinely require human attention.

**6. Empty states are invitations.**
An empty projects list is not a bug — it's the beginning of something. Every empty state has a headline, a subtext, and a single action. They should make the user want to fill them.

**7. Errors are helpers, not accusations.**
When something fails, the UI stays calm. The error message explains what happened and what to do next. The color is muted red, never alarming. The tone is a colleague, not a warning siren.

---

## Design System

### Color

The palette is dark by nature — like black anodized aluminum. Violet is the only warm color and it appears exactly once per screen: on the single primary action. Everything else is built from cool, receding surfaces.

```
--color-surface-900: #0a0a14   /* deepest background — the void */
--color-surface-800: #12121f   /* card background */
--color-surface-700: #1a1a2e   /* elevated surface / input background */
--color-surface-600: #24243d   /* border / divider */
--color-surface-500: #32325a   /* hover surface */

--color-primary:     #7c3aed   /* violet-600 — one primary action per screen */
--color-primary-dim: #5b21b6   /* violet-800 — pressed state */
--color-primary-glow:#a78bfa   /* violet-400 — glow / accent text */

--color-success:     #34d399   /* emerald-400 — completion, passing checks */
--color-warning:     #fbbf24   /* amber-400 — in-progress, optional */
--color-danger:      #f87171   /* red-400 — failures, errors */
--color-text:        #e2e8f0   /* slate-200 — primary text */
--color-muted:       #64748b   /* slate-500 — secondary text, icons at rest */
--color-subtle:      #334155   /* slate-700 — disabled, placeholder */
```

**Background texture**: A barely-perceptible radial gradient behind surface-900 — violet at 2.5% opacity, centered at the top of the viewport. Invisible until you notice it. Then you can't un-see the depth it adds.

**Color rules (enforced)**:
- Violet only on: the primary CTA button, active nav item, focused inputs, progress fills, the status dot of a running service
- Success green only on: completed steps, passing checks, successful commits — never as a decorative tint
- Warning amber only on: in-progress states, optional steps, non-critical warnings
- Danger red only on: failing checks, error states — calm and steady, never flashing

### Typography

Inter for UI. JetBrains Mono for code and terminal. Both loaded via `@fontsource` — no FOUT, no CDN dependency.

```
Display:  48px / line-height 56px / letter-spacing -0.04em / weight 700
H1:       32px / line-height 40px / letter-spacing -0.025em / weight 700
H2:       24px / line-height 32px / letter-spacing -0.015em / weight 600
H3:       18px / line-height 28px / letter-spacing -0.01em / weight 600
Body:     16px / line-height 26px / letter-spacing 0 / weight 400
Body-sm:  14px / line-height 22px / letter-spacing 0 / weight 400
Label:    12px / line-height 16px / letter-spacing +0.06em / weight 500  (UPPERCASE)
Mono:     13px / line-height 20px / letter-spacing 0 / weight 400 (JetBrains Mono)
```

Numbers in the dashboard (CPU %, RAM, temperature) use `font-variant-numeric: tabular-nums` so they don't jump in width as they change.

### Motion

Motion communicates. It is never gratuitous. All interactive elements respond within one frame (16ms) even if the animation takes longer.

**Duration tiers:**
```
Micro:    120ms  — property changes within a single element (color, opacity, border)
Short:    200ms  — small elements appearing/disappearing (badges, tooltips, icons)
Medium:   300ms  — panels, modals, drawers, route transitions
Long:     500ms  — hero animations, wizard step entrance, confetti fade-in
Stagger:  40ms   — delay between list items entering
          60ms   — delay between cards entering
Exit:     65% of entry duration — exits are snappier than entries
```

**Easing:**
```
Snap:     cubic-bezier(0.4, 0, 1, 1)     — exits, things leaving screen
Enter:    cubic-bezier(0, 0, 0.2, 1)     — things arriving, settling
Spring:   spring(stiffness: 320, damping: 28) — interactive elements, cards
Smooth:   cubic-bezier(0.4, 0, 0.2, 1)  — general purpose, property transitions
```

**Motion rules:**
- Entering elements: translate from +12px below → 0 + opacity 0 → 1
- Exiting elements: translate to -6px above + opacity → 0 (lighter than entry)
- List items stagger from top to bottom, never simultaneously
- Route changes: outgoing fades up (-8px, 200ms), incoming rises from +16px (300ms, 100ms delay)
- Never animate layout — only transform and opacity for performance

### Spacing (4px grid, strictly enforced)

```
4, 8, 12, 16, 20, 24, 32, 40, 48, 64, 80, 96px
```

Page margins: 24px (mobile) → 48px (tablet) → 64px (desktop)
Card padding: 24px
Section gap: 40px
Component gap: 12px–16px
Inline gap: 8px

### Border Radii

```
XL:   20px  — modals, bottom sheets, feature cards
LG:   16px  — standard cards, panels
MD:   10px  — inputs, dropdowns, code blocks
SM:   6px   — badges, inline pills
Full: 9999px — status dots, button pills, URL chips
```

### Elevation & Shadow

Shadows convey depth, not decoration. Three levels only:

```
Flat:     none  — sidebar, toolbar — these are surfaces, not cards
Card:     0 1px 3px rgba(0,0,0,0.5), 0 4px 16px rgba(0,0,0,0.3)
Float:    0 2px 8px rgba(0,0,0,0.6), 0 16px 48px rgba(0,0,0,0.4)
Glow:     0 0 0 1px rgba(124,58,237,0.4), 0 0 24px rgba(124,58,237,0.2)
```

The violet glow shadow is used exclusively on: focused inputs, the active service card, the primary CTA button on hover.

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
│       ├── main.css           # @theme tokens, global resets
│       ├── App.vue
│       ├── router/index.ts
│       ├── composables/
│       │   ├── useFetch.ts        # Base HTTP, auto JSON, error normalisation
│       │   ├── useTerminal.ts     # WS connection, xterm init, reconnect loop
│       │   ├── useMetrics.ts      # SSE subscription, history ring buffer
│       │   └── useAgentStream.ts  # SSE agent messages, connection state
│       ├── stores/
│       │   ├── setup.ts
│       │   ├── projects.ts
│       │   ├── settings.ts
│       │   ├── agent.ts
│       │   ├── github.ts
│       │   └── metrics.ts
│       ├── views/
│       │   ├── setup/
│       │   │   ├── WizardLayout.vue
│       │   │   ├── StepWelcome.vue
│       │   │   ├── StepSystemCheck.vue
│       │   │   ├── StepCloudflare.vue
│       │   │   ├── StepGitHub.vue
│       │   │   ├── StepProviders.vue
│       │   │   ├── StepOpenCode.vue
│       │   │   ├── StepNanoClaw.vue
│       │   │   └── StepComplete.vue
│       │   ├── DashboardView.vue
│       │   ├── IDEView.vue
│       │   ├── AgentView.vue
│       │   ├── ProjectsView.vue
│       │   ├── GitHubView.vue
│       │   └── SettingsView.vue
│       └── components/
│           ├── layout/
│           │   ├── AppShell.vue
│           │   ├── AppSidebar.vue
│           │   ├── AppTopbar.vue
│           │   └── MobileNav.vue
│           ├── dashboard/
│           │   ├── ServiceCard.vue
│           │   ├── VitalsPanel.vue
│           │   ├── ActivityFeed.vue
│           │   ├── QuickLaunch.vue
│           │   └── AccessPanel.vue
│           ├── terminal/
│           │   └── TerminalPane.vue
│           ├── agent/
│           │   ├── ChatBubble.vue
│           │   ├── ChatInput.vue
│           │   └── TypingIndicator.vue
│           ├── projects/
│           │   ├── ProjectCard.vue
│           │   ├── NewProjectModal.vue
│           │   └── GitPanel.vue
│           ├── github/
│           │   ├── RepoCard.vue
│           │   ├── PrCard.vue
│           │   └── EventRow.vue
│           └── ui/
│               ├── StatusDot.vue         # Pulsing dot — runs/starting/error/stopped
│               ├── KeyInput.vue          # Masked key field with show/hide
│               ├── CheckRow.vue          # Check row with animated states + fix inline
│               ├── StepRail.vue          # Wizard progress rail
│               ├── Sparkline.vue         # SVG sparkline, animated new points
│               ├── QrCode.vue            # Canvas QR from URL
│               ├── CopyButton.vue        # Copies text, shows ✓ for 1.5s
│               └── EmptyState.vue        # Shared empty state: illustration + cta
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

---

### Phase 1 — Foundation (Week 1)

**Goal**: Go server + Vue shell running. Design system in place. Auth skeleton wired.

**Backend:**
- [ ] `go mod init vibecodepc`, dependencies: chi, gorilla/websocket, modernc.org/sqlite, go-github, creack/pty
- [ ] `Makefile`: `dev`, `build`, `check`, `lint`, `cross` targets
- [ ] `.air.toml` for Go hot reload, watching `server/`
- [ ] Chi router with middleware stack: security headers, CORS (dev), request logger, recovery
- [ ] SQLite: all 5 tables, migrations run on first startup
- [ ] `//go:embed public/*` for serving built client assets
- [ ] `/auth/github` routes + keystore token storage

**Frontend:**
- [ ] Vue 3 + Vite + Tailwind CSS v4 scaffold
- [ ] `main.css`: full `@theme` block with every design token
- [ ] `@fontsource/inter` + `@fontsource/jetbrains-mono` — both preloaded, no FOUT
- [ ] `App.vue` with `<RouterView>` wrapped in `<Transition name="page">` (fade + translate)
- [ ] `router/index.ts`: `/setup/*` guards + `/app/*` guards, setup check on first load
- [ ] All 6 Pinia stores: state shapes, empty actions stubs
- [ ] `useFetch` composable
- [ ] `vite.config.ts`: proxy `/api/*`, `/ws/*`, `/auth/*` → `:3000`
- [ ] golangci-lint + ESLint + Prettier

**Deliverable**: `make dev` shows the wizard welcome screen. OAuth round-trip works. The design tokens produce the correct dark palette.

---

### Phase 2 — Setup Wizard (Week 2)

**Goal**: A fully automated wizard that feels like setting up a piece of premium hardware — effortless, inevitable, beautiful.

The wizard is full-screen. The sidebar nav is hidden during setup. There is no chrome, no distraction — just the step, the progress rail, and the action.

---

#### WizardLayout

A vertical progress rail on the left (desktop) or a horizontal step bar at the top (mobile). Eight numbered nodes connected by a line. Completed nodes fill with emerald and get a drawn checkmark. The current node pulses with violet. Future nodes are surface-600 and muted.

The rail line between nodes fills as the user advances — not instantly, but with a 600ms ease-out animation left to right.

The step content area slides: on advance, the current step translates left and fades (-24px, 250ms), the next step arrives from the right (+24px → 0, 300ms, 50ms after exit begins).

---

#### Step 1 — Welcome

**Choreography (auto-plays on mount, no user interaction until complete):**

1. **0ms**: Screen is surface-900. Nothing.
2. **300ms**: A single 6px violet point appears at center-screen, blurs to a 40px soft orb. Spring in.
3. **700ms**: The orb expands into a 280px radial gradient wash. Four tool icons (Docker, GitHub, Claude mark, opencode) materialize 120px from center, one at a time (60ms stagger), each at 40% opacity.
4. **1.1s**: The icons begin a slow orbital drift — each on a slightly different radius and speed. 3% scale oscillation, 4–6s per orbit, organic, not mechanical.
5. **1.4s**: The Raspberry Pi icon fades into center (replacing the orb), 300ms.
6. **1.8s**: The hostname appears — typed letter by letter, 28ms per character, with a blinking cursor. `"Let's set up raspberrypi.local"` in H2 weight, violet for the hostname part.
7. **2.6s**: Tagline fades in: `"Your Pi. Your AI. Your code."` in Display size, white, centered. Arrives from +12px below, 400ms spring.
8. **3.2s**: "Start Setup →" button rises into place. Violet, pill shape. A faint violet border glow pulses gently (opacity 0.4 → 0.7, 2s loop).

The button is the only interactive element. Pressing it instantly transitions to Step 2.

---

#### Step 2 — System Check

`system_check.go` fires all checks concurrently on mount. The UI shows each check arriving and resolving.

**Layout**: A centered list, max-width 560px. Each row is 56px tall, 16px radius, surface-800 background, 1px border (surface-600).

**Row entrance**: Checks slide in from +20px right, opacity 0 → 1, 40ms stagger, 200ms each. By the time the last row is visible, the first check has already started running.

**Row states:**

```
Pending   — left border: surface-600 (2px)
            status dot: 8px, muted
            label: muted

Running   — left border: amber, 2px, glowing (box-shadow amber at 40%)
            status dot: breathing scale (0.85 → 1.15, 1.2s loop, ease-in-out)
            label: text (white) — "Checking Docker..."

Pass      — left border sweeps to emerald (width 0 → 100% of height, 400ms ease-out)
            checkmark draws itself: SVG path stroke-dashoffset 1 → 0, 350ms ease-out
            label: text — "Docker 24.0.7"
            Row gently flashes (surface-700 → surface-800, 300ms) to mark resolution

Fail      — left border: danger, steady (no throb — calm authority, not alarm)
            X icon fades in, 200ms
            label: danger text — "Docker not installed"
            "Fix" pill button appears 200ms after fail state settles
```

**Fix flow:**
- "Fix" button: pill shape, danger-border, body-sm text. On click: the button becomes a spinner + "Fixing..." label.
- Below the row, a log terminal expands (max-height: 160px, overflow-y: auto, 300ms height spring) — JetBrains Mono, 12px, surface-900 background.
- Log lines appear one by one as SSE frames arrive. Each line fades from muted → text as it settles.
- On success: log collapses (300ms), row transitions to Pass state.
- On failure: log stays open, a "Retry" button replaces "Fix".

**Checks:**

| Check | Auto-fix |
|---|---|
| Docker installed | Yes — streams Docker Engine install |
| Docker daemon running | Yes — `systemctl start docker` |
| RAM ≥ 1 GB | No — warning badge, doesn't block advance |
| Disk ≥ 5 GB free | No — shows usage bar, doesn't block advance |
| Internet reachable | No — "Check your network connection" |
| `git` installed | Yes — `apt-get install -y git` |

**Auto-advance**: When all critical checks pass, a 1-second countdown bar fills under the "Continue" button. A subtle "Auto-advancing in 1s — cancel" caption appears. The bar fill is violet. Cancel restores the button to a manual state.

---

#### Step 3 — Cloudflare Tunnel

The tunnel is already running from the installer. This step is confirmation and optional upgrade.

**Layout**: Centered, max-width 600px.

**Animated data-flow diagram** (SVG, 320px wide):
```
[Pi]  ─────────────────────►  [CF]  ─────────────────────►  [Globe]
```
- Nodes: 48px circles, icons inside (device, cloud, globe)
- Connecting lines: 2px, surface-600
- Traveling packets: 5px dots that depart every 800ms, travel the full path in 1.2s (ease-in-out)
- When connected: each node gets a 16px emerald ring that pulses out (scale 1 → 1.5, opacity 0.6 → 0, 2s loop)

**URL pills** (two of them, stacked):
- 400px wide, 48px tall, surface-700 background, 12px radius, 1px border
- URL text: JetBrains Mono, 13px, truncated with ellipsis
- Right side: copy icon + QR icon, separated by a 1px divider
- On copy: icon transitions to a checkmark for 1.5s, a green flash sweeps the pill (opacity 0.15, 300ms)
- On QR click: a 200×200 QR pops over the pill (scale 0.9 → 1.0, 200ms spring, backdrop blur)

**Optional upgrade section** (collapsed by default):
- A disclosure row: "Get a stable URL →" with a chevron. Clicking opens a panel below (height spring, 300ms).
- Inside: a 3-step mini-guide with numbered circles, then a token input + "Connect" button.
- Token validation: the input border transitions to amber while validating, then emerald on success.

**Auto-advance**: 3 seconds after tunnel shows connected. A faint countdown bar in the bottom of the step card. "Stay here" text link cancels it.

---

#### Step 4 — GitHub

**Layout**: Centered, max-width 480px. Single purpose — one action.

A centered card: 80px GitHub mark, H2 "Connect GitHub", body-sm subtext explaining why. Below: the primary "Connect with GitHub →" button — violet, full-width, pill.

On click: a popup opens (centered, 500×600px). While the popup is open, the button becomes "Waiting for authorization..." with a rotating arc loader. The arc is a single segment, not a full circle — lighter feeling.

On completion (popup sends `postMessage`):
1. Popup closes
2. Button morphs into a success state: emerald background, checkmark icon + "Connected" (300ms crossfade)
3. Below the button, a card slides down (height spring, 400ms): GitHub avatar (40px, circle) + username (H3) + `N repos ready` in muted body-sm
4. After 1.5s, the "Continue →" button appears below the identity card

"Skip for now" — a muted text link, always visible below the main card. Never hidden — GitHub is optional.

---

#### Step 5 — AI Providers

On mount, the server scans env for `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GOOGLE_API_KEY`, and probes `localhost:11434`. The UI begins showing results before the user has touched anything.

**Provider cards**: 4 cards in a 2×2 grid (desktop), stacked (mobile). Each 280px × 160px, surface-800, 16px radius.

Card anatomy:
```
┌──────────────────────────────────┐
│  [logo]  Provider Name           │
│          "Powers the AI agent"   │
│                                  │
│  ┌────────────────────────────┐  │
│  │ sk-ant-••••••••••••1234   │  │
│  └────────────────────────────┘  │
│  [Test Key]          ✓ Valid     │
└──────────────────────────────────┘
```

**Auto-detected state**: The card's top border transitions to violet (shimmer left to right, 400ms) as detection begins. A "Scanning..." caption with a small orbital spinner (two dots orbiting, not a circle). Then the masked key fills in character by character (30ms each, appearing from the right). A "Detected from environment" badge slides in from the right: emerald background, 10px, rounded. Immediately, the test fires — border goes amber ("Testing...") → emerald ("Valid"). The whole sequence takes under 2 seconds and requires zero user input.

**Manual entry**: The key input has a `type="password"` -like masking but with the last 4 characters visible. On focus: the input border glows violet. On blur with content: auto-triggers the test. The "Test Key" button is the secondary option.

**Anthropic card**: Shown first, larger, with an additional note "Required for NanoClaw". The card gets a subtle violet left border by default (not just on hover) — it's special.

**Ollama**: If `localhost:11434` responds, the Ollama card shows "Running locally" with a green dot and auto-configures. No key needed.

**Auto-advance**: Once any provider is Valid, a "Continue →" button appears at the bottom of the step (slides up from +12px, 300ms spring). But users can keep adding more providers — the step doesn't auto-advance aggressively here.

---

#### Step 6 — OpenCode

On mount: the server checks for `opencode` in PATH. The result arrives within 200ms.

**If already installed:**
A terminal block (surface-900, 16px radius, JetBrains Mono) fades in with `opencode --version` output. Green dot + version number. A provider selector pill appears below (pre-filled from Step 5). The step is complete in under 1 second.

**If not installed:**
1. A terminal block appears (empty, with a pulsing cursor)
2. After 300ms: install auto-starts. Lines stream in from SSE.
3. Each line: opacity 0 → 1, translate +4px → 0, 80ms each. Older lines dim to muted as new ones arrive — the focus is always on the latest output.
4. A thin violet progress bar sits above the terminal block, fills 0 → 100% over the install duration. It doesn't know the true percentage — it uses a timed estimation curve (fast at first, slows near 100%, then snaps to 100% on completion).
5. On completion: the terminal output fades out (300ms), the progress bar flashes violet then fills to emerald. The version badge bounces in (scale 1.15 → 1.0, spring).

**Provider selector**: a horizontal pill group — one pill per configured provider. Active pill: violet background. Others: surface-700, muted text. Click switches instantly.

---

#### Step 7 — NanoClaw (Optional)

The step header has a muted "(Optional)" label — a small badge, not a disclaimer wall of text.

**Setup timeline** (vertical, left-aligned, 4 stages):

```
  ◉  Cloning nanoclaw...          ← violet dot, pulsing
  │
  ○  Writing configuration        ← surface-600 dot
  │
  ○  Building Docker image        ← surface-600 dot
  │
  ○  Starting container           ← surface-600 dot
```

As each stage completes:
- Its dot fills with emerald and gets a drawn checkmark inside (SVG, 250ms)
- The line below it becomes solid emerald (fills top to bottom, 400ms)
- The next stage's dot transitions to violet and begins pulsing

Inline log output appears below the active stage node, not in a separate box — it's part of the timeline, indented 24px. The log lines appear one at a time, fading in. When the stage completes, the log collapses (height spring, 400ms) and shows a summary line ("Docker image built in 3m 22s").

**Messaging platform selector** (appears after container starts):
Four cards in a 2×2 grid: WhatsApp, Telegram, Discord, Web Only. Web Only is pre-selected (emerald border). Clicking another highlights it instead.

WhatsApp: shows a QR code immediately (auto-refreshes every 20s with a 500ms crossfade).
Others: show a token input that validates inline on blur.

"Skip messaging" — a muted text link. Clicking it advances without selecting a platform. The container is running regardless; only the platform pairing is optional.

---

#### Step 8 — Complete

The moment the user lands here, confetti fires. Not an explosion — a composed, elegant cascade. Particles: medium density, varying sizes (4–12px), all in the app's palette (violet, emerald, white, slate). They drift downward with slight lateral drift and rotation. They fade out before reaching the fold. Duration: 5 seconds. Cannot be retriggered.

**Layout** (max-width 640px, centered):

1. **"Your station is ready."** — Display size, white, fades in from +8px below, 400ms.
2. **Summary grid**: 3-column list of configured services. Each row: service icon (24px) + name + status pill ("Active", "Connected", etc.). Items stagger in 80ms apart. Each status pill has its color (emerald/amber) and a drawn checkmark or checkmark icon.
3. **Access panel**: Two large URL pills side by side. 80px tall, surface-700 background, 20px radius. Each has a copy button (right side) and a QR button. On hover: a subtle prismatic border (gradient shifts, 300ms transition).
4. **"Open Dashboard →"** — The only violet element on screen. Full-width, pill, 56px tall. Below it, muted caption: "Auto-opening in 8s". A thin emerald progress bar fills beneath the button, counting down.

If GitHub was connected: a card with the user's GitHub avatar + "**username** — N repos ready to import" sits between the summary grid and the access panel. It slides in 400ms after the summary grid settles.

Auto-redirect at 8 seconds. Page transition: the entire complete step fades out (400ms) as the dashboard fade-in plays simultaneously (300ms delay).

---

### Phase 3 — OpenCode IDE (Week 3)

**Goal**: A professional-grade embedded terminal that feels like a portal into the machine.

#### Layout

`IDEView.vue` is the fullest-bleed view in the app — it uses the entire content area minus the sidebar. No topbar visible in IDE mode (collapsed to a slim 40px strip showing only project name, branch, and a ×). This gives maximum terminal real-estate.

```
┌─ Session Strip (40px) ────────────────────────────────────────────┐
│  ← [my-saas-app]  [main ↑1]  [● Connected]      [Kill] [New]   │
├─ Git Panel (240px, collapsible) ─────┬─ Terminal ─────────────────┤
│                                      │                            │
│  BRANCH                              │  ▌                         │
│  ◉ main  ↑1 ↓0          ⌄           │  (xterm.js)               │
│                                      │                            │
│  CHANGES                             │                            │
│  M  src/App.vue                      │                            │
│  M  server/routes/api.go             │                            │
│  ?  notes.md                         │                            │
│                                      │                            │
│  STAGED (0)                          │                            │
│  ─────────────────────               │                            │
│                                      │                            │
│  ┌────────────────────────┐          │                            │
│  │ commit message         │          │                            │
│  └────────────────────────┘          │                            │
│  [Stage All & Commit]                │                            │
│  [Push →]  ↑1 ready                  │                            │
└──────────────────────────────────────┴────────────────────────────┘
```

#### TerminalPane

xterm.js theme precisely matching the design system:
```js
{
  background:  '#0a0a14',
  foreground:  '#e2e8f0',
  cursor:      '#7c3aed',
  cursorAccent:'#0a0a14',
  selection:   'rgba(124,58,237,0.25)',
  black:       '#0a0a14',
  brightBlack: '#334155',
  // ... full 16-color scheme aligned to the palette
}
```

Cursor: block style, blinks at 530ms interval (slightly slower than default — more comfortable for long sessions).

**Connection states:**
- Connecting: terminal is empty, a centered `Connecting...` caption in muted body-sm with a small orbital spinner. The xterm canvas is 50% opacity.
- Connected: the opacity animates to 100% (200ms), the status indicator in the session strip turns emerald, the caption fades out.
- Disconnected: the terminal dims to 35% opacity. An overlay appears: blurred backdrop + centered card with `Reconnecting...` spinner and elapsed time. Exponential backoff: 500ms, 1s, 2s, 4s, 8s, then every 30s. On reconnect: overlay fades, brightness restores, the terminal emits a subtle `\x07` bell (if audio enabled).

**Terminal entrance**: When a terminal session first opens, the xterm canvas scales from 0.98 → 1.0 and fades from 0 → 1 over 200ms. It's barely perceptible but makes the "portal opening" feel intentional.

**ResizeObserver**: Attached to the pane container. On any size change, `fit()` is called and a resize frame sent to the server. Debounced 50ms.

#### GitPanel

The panel slides in from the left: translate(-240px) → translate(0), 300ms spring. The terminal resizes to fill the vacated space simultaneously (no layout pop).

**Branch selector**: A pill showing the current branch. Click opens a dropdown with all branches — local and remote, separated. Branches are sorted: current first, then recent-last-checkout order. Click to checkout; if the tree is dirty, a modal asks to stash first.

**File badges**: `M` (modified) in amber, `A` (added) in emerald, `D` (deleted) in muted red, `?` (untracked) in muted. Each is a 20px pill with 2px padding. File path is truncated from the left (`...routes/api.go`).

**Diff preview**: Clicking a file shows a diff view in a slide-over panel (right side of the git panel, 300ms spring). Syntax-highlighted with line-level +/- coloring. The diff is scrollable independently from the file list. Close with Escape or clicking outside.

**Commit flow**:
1. The commit message field starts at 1 line height. On focus: expands to 3 lines (height spring, 250ms).
2. Placeholder: "Describe your changes..." in subtle.
3. Below: "Stage All & Commit" button (primary, violet) + "Push →" (secondary, surface-700) side by side.
4. On commit: button shows `Committing...` with a spinner arc. On success: `Committed ✓` with a green flash (150ms). Returns to default state after 2s.
5. On push: shows `Pushing...`. On success: the ↑N counter resets to ↑0.

**Ahead/behind**: Small indicator row: `↑1 ↓0` — using JetBrains Mono, 12px. The arrows animate when the number changes (brief scale 1.2 → 1.0, 150ms).

---

### Phase 4 — Agent Chat (Week 4)

**Goal**: A conversation interface that feels like talking to something genuinely intelligent — not a chatbot widget.

#### Layout

Full-height view. Two columns at ≥1024px: a 280px sidebar (project context selector + conversation list) and the main chat area. Single column on mobile.

```
┌─ Context ──────────────┬─ Chat ──────────────────────────────────┐
│                        │                                         │
│  Chatting with         │    ◉  nanoclaw                         │
│  ◉ nanoclaw agent      │       "I've analyzed your codebase..."  │
│                        │       [code block]                      │
│  Project context:      │                                         │
│  ◉ my-saas-app    ×    │    ─────────────────────────────────    │
│                        │                                         │
│  ─────────────         │    You                                  │
│                        │    "Can you refactor the auth module?"  │
│  History               │                                         │
│  Today                 │    ◉  nanoclaw                         │
│  "Refactor auth..."    │       ●●●  (typing indicator)           │
│  "Fix the DB query"    │                                         │
│                        │ ────────────────────────────────────── │
│                        │ ┌──────────────────────────────────┐   │
│                        │ │ Message nanoclaw...               │   │
│                        │ │                                   │   │
│                        │ └──────────────────────────────────┘   │
│                        │  Shift+Enter for new line    [Send ↑]   │
└────────────────────────┴─────────────────────────────────────────┘
```

#### Messages

**Agent bubble**: Left-aligned, max-width 75%. Avatar: a 28px circle with a glowing violet dot (the dot's box-shadow pulses gently when the agent is active: 0 0 0 0 → 0 0 8px violet, 2s loop). Agent name in Label style above the first bubble of each turn.

Messages slide up from +12px as they arrive: translate(0, 12px) → translate(0, 0) + opacity 0 → 1, 250ms spring.

Code blocks in agent responses:
- surface-900 background, 12px radius, 1px border (surface-600)
- Language badge: top-right corner, Label style, surface-700 pill
- Copy button: top-right, appears on hover (200ms fade), shows ✓ for 1.5s on click
- Syntax highlighting: Shiki with a custom theme matching the palette

**User bubble**: Right-aligned, surface-700 background, primary-glow border at 30% opacity on the right side.

**Typing indicator** (`TypingIndicator.vue`):
Three dots, 6px each, violet fill. They animate in a wave: each dot delays 120ms from the previous. Scale: 0.6 → 1.0 → 0.6, ease-in-out, 700ms loop. The overall indicator fades in over 200ms when the agent starts processing, fades out when the response arrives.

#### Input

A `<textarea>` that auto-grows from 1 line to a maximum of 5 lines (height transitions smoothly, spring animation). On focus: a faint violet glow (box-shadow: 0 0 0 1px violet at 40%, 0 0 12px violet at 15%, 200ms transition). Placeholder: "Ask anything about your code..." in subtle.

When the agent is processing: the input is `disabled`, opacity transitions to 40%, and a "Thinking..." caption appears below in muted body-sm.

Send button (↑ arrow): appears only when the textarea has content (fade in, 150ms). On click: message animates upward out of the input area (translate -20px, opacity 0, 200ms), the textarea clears, the input re-focuses.

#### Project Context

A pill at the top of the chat showing the active project context. Click to switch. When a project is active, the agent's responses are implicitly aware of its language, framework, and recent git activity (passed as metadata with each message).

#### Empty State

When there are no messages: a centered layout with a glowing 48px violet orb (the "nanoclaw" avatar at rest), headline "Ask anything", subtext "Your AI coding companion has full context of your project. Try asking it to review a file, suggest refactors, or explain a concept." No buttons — just the input at the bottom, ready.

---

### Phase 5 — GitHub Integration (Week 4–5)

**Goal**: GitHub as a first-class citizen — not a settings page, but a workspace.

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
- `Push(projectPath)` → push current branch (GitHub token via git credential helper)
- `Pull(projectPath)` → pull with rebase
- `Branches(projectPath)` → list local + remote branches
- `Checkout(projectPath, branch)` → switch branch (stash if dirty)
- `Clone(ctx, url, destPath)` → progress lines emitted on returned channel

#### GitHubView Layout

Two tabs: **Repos** | **Activity** — horizontal pill tabs, not underline tabs. Active tab: violet background, white text. Inactive: surface-700, muted text. Tab content transitions with a 200ms crossfade.

**Repos tab:**

Search input at the top: 100% width, 44px tall, surface-700 background, 10px radius. A search icon inside (left), a clear button (right, only visible when non-empty). As the user types, results filter client-side instantly. After 400ms of no typing, a background refetch fires to search the full GitHub API.

Filter pills below the search: `All` `TypeScript` `Python` `Go` `Rust` `...` — language filter. Sort: `Recently pushed` `Stars` `Name`. These are small pill groups, not a select dropdown.

`RepoCard.vue` grid: 3 columns (desktop), 2 (tablet), 1 (mobile). Each card:
- Repo name (H3) + private/public badge (if private: a small lock icon, muted)
- One-line description in body-sm, muted, truncated
- Row of tags: language dot + name, star count (⭐ N), last pushed ("2 days ago")
- "Import" button: bottom-right, appears on card hover (slides up from bottom edge, 200ms)
- Cards load in skeleton state first (shimmer animation, surface-700/surface-600 blocks) then crossfade to content

**Activity tab:**

A vertical timeline. Each event:
- Left: a 32px circle icon — the event type icon in the relevant color (push→violet, PR→emerald, issue→amber, review→muted)
- Center: repo name (bold) + event description + branch/PR number
- Right: relative time ("2h ago") in JetBrains Mono, muted, 12px

New events that arrive on the 60s poll slide in from the top with a soft green flash (the row background briefly flashes success at 10% opacity, then settles).

#### Not-Connected State

When GitHub is not connected, the `GitHubView` shows an `EmptyState`:
- Centered GitHub mark (48px, muted)
- Headline: "Connect GitHub to see your repos"
- Subtext: "Browse, import, and manage pull requests without leaving the app."
- A single "Connect GitHub →" button (violet)

---

### Phase 6 — Real-Time Dashboard (Week 5–6)

**Goal**: A command center that communicates system health at a glance. Form and function as one.

The dashboard is the home screen. It should feel like the instrument panel of something powerful — data-dense but visually quiet, animated but not distracting.

#### Layout

```
┌─ AppShell ─────────────────────────────────────────────────────────┐
│                                                                    │
│  ┌─ Sidebar ──┐  ┌─ Main Content ────────────────────────────────┐ │
│  │            │  │                                               │ │
│  │  ◉ Dash    │  │  ┌── SERVICES ─────────────────────────────┐ │ │
│  │  ○ IDE     │  │  │  [opencode] [NanoClaw] [Cloudflare] [Docker]│ │
│  │  ○ Agent   │  │  └─────────────────────────────────────────┘ │ │
│  │  ○ Projects│  │                                               │ │
│  │  ○ GitHub  │  │  ┌── QUICK LAUNCH ──────┐ ┌── VITALS ──────┐ │ │
│  │  ○ Settings│  │  │  [proj] [proj] [new] │ │ CPU ████░ 42%  │ │ │
│  │            │  │  └─────────────────────┘ │ RAM ████░ 67%  │ │ │
│  │            │  │                           │ Disk███░ 45%   │ │ │
│  │            │  │  ┌── ACTIVITY ──────────┐ │ 52°C  3d 14h  │ │ │
│  │            │  │  │  [events timeline]   │ └───────────────┘ │ │
│  │            │  │  └─────────────────────┘                     │ │
│  │            │  │                          ┌── ACCESS ────────┐ │ │
│  │            │  │                          │ Local  [copy][QR]│ │ │
│  │            │  │                          │ Remote [copy][QR]│ │ │
│  │            │  │                          └──────────────────┘ │ │
│  └────────────┘  └───────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────┘
```

All dashboard panels load with skeleton shimmer simultaneously. As data arrives, panels crossfade from skeleton to content (150ms each). Panels don't block each other — each resolves independently.

#### Sidebar (`AppSidebar.vue`)

Width: 220px (desktop), collapses to icons-only at 64px on tablet, becomes a bottom nav on mobile.

Each nav item: 40px tall, 10px radius, full-width. Text + icon. Active item: violet background at 15%, a 3px violet bar on the left edge, white text. Inactive: muted icon + muted text. On hover: background transitions to surface-700 (150ms).

Below the nav items, at the bottom of the sidebar: live service status dots. Four rows — opencode, NanoClaw, Cloudflare, Docker. Each: a 6px `StatusDot` + service name in body-sm. No additional info — just the health at a glance.

`StatusDot.vue`:
- Running: emerald, with a 12px ghost ring that expands and fades (scale 1→2, opacity 0.5→0, 2s loop)
- Starting: amber, dot pulses (opacity 0.6↔1.0, 1s loop)
- Error: danger, steady — no pulse (calm authority)
- Stopped: surface-500, no animation

#### Service Cards (`ServiceCard.vue`)

Four cards in a horizontal row. Each: 270px × 148px, 16px radius, surface-800, 1px border (surface-600).

Card anatomy:
```
┌──────────────────────────────────────────┐
│ [icon 32px]  Service Name    ● Running   │ ← top row
│                                          │
│         2 active sessions                │ ← key metric (H2 size)
│                                          │
│ [Primary Action]              [···]      │ ← bottom row
└──────────────────────────────────────────┘
```

The key metric (session count, container count, tunnel uptime) uses `font-variant-numeric: tabular-nums`. When the number changes, it counts (not jumps): e.g., 1→2 counts 1, 1.5, 2 in 200ms.

On hover: card border gains a violet tint at 30% opacity (200ms), a `box-shadow: 0 4px 24px rgba(0,0,0,0.4)` expands, the card translates up 3px. The primary action button, which was visible but subtle, brightens.

On click (card body): a log panel drops down from below the card (not in a modal — inline). The panel is surface-900, 160px tall, scrollable, showing the last 50 log lines via SSE. Clicking again collapses it.

#### System Vitals (`VitalsPanel.vue`)

A compact panel with four metrics. Each has a label, current value, and a sparkline graph.

The sparkline (`Sparkline.vue`):
- SVG, 100% wide × 40px tall
- A filled area chart with a 1.5px stroke line above
- Line color: `--color-primary-glow` (violet-400)
- Fill: gradient from violet at 15% opacity → transparent
- No axes, no grid — just the shape of the data
- When a new point arrives: the rightmost point animates in (the line extends right, 200ms smooth)
- The entire graph shifts left as new points arrive — existing points don't jump, they translate

The percentage values beside each metric use a number-morph transition: the digits reel to the new value at 200ms, using `font-variant-numeric: tabular-nums`.

Temperature: the sparkline color transitions from violet → amber → danger as temperature rises (25°C = violet, 55°C = amber, 75°C = danger). This transition is smooth — not a threshold jump.

Uptime: shows as `3d 14h 22m`, updates every minute. The format changes when crossing day/hour boundaries.

#### GitHub Activity (`ActivityFeed.vue`)

A vertical feed, max 8 events visible, scrollable. Polled every 60 seconds with exponential fallback if offline.

Each event row: 48px. Icon left (32px circle), content center, time right.

When new events arrive on poll: they insert at the top with a brief emerald shimmer (the row background flashes `--color-success` at 8% opacity, fades in 400ms). Older events push down smoothly — no layout jump, the list height doesn't change because the oldest event exits at the bottom.

**Not-connected state**: A placeholder card with a muted GitHub mark and "Connect GitHub to see your activity" — single action button.

#### Quick Launch (`QuickLaunch.vue`)

Up to 4 project cards + a "New Project" card. Horizontal row on desktop, 2×2 grid on tablet.

Each card: 160px × 100px, surface-800, 16px radius. Project name (H3, truncated), language icon (16px), branch name (mono, 12px, muted). Last-opened in the bottom-right corner: relative time in muted body-sm.

On hover: the card lifts 4px, a violet border replaces the default (200ms), an "Open →" label appears in the bottom-right corner (fades in, 150ms). Clicking navigates to `/app/ide/:projectId` with a page transition.

The "+ New Project" card: dashed border (surface-600), centered `+` icon (24px, muted). On hover: the border becomes violet, the `+` icon transitions to violet (150ms).

#### Access Panel (`AccessPanel.vue`)

Two rows. Local URL, Remote URL. Each: a 48px pill — the full URL (truncated, JetBrains Mono) with a copy icon (right) and QR icon (right of copy). A small badge indicates tunnel mode: `Quick` (amber) or `Stable` (emerald).

If in quick mode: a muted "Upgrade for a stable URL →" text link below. Clicking opens the Cloudflare settings page.

On copy: a subtle toast-less success state — the copy icon morphs into a checkmark (SVG crossfade, 200ms) for 1.5s, then morphs back.

On QR: a 200×200 QR popover appears above the row (scale 0.9 → 1.0, spring, 200ms). It auto-closes on click-outside or Escape.

---

### Phase 7 — Projects & Settings (Week 6)

#### Projects View

**Header**: "Projects" (H1) left. "New Project +" button right (violet). A search input between them (desktop: 280px wide, mobile: full-width below header).

**Card grid**: Masonry, 3 columns (desktop) → 2 (tablet) → 1 (mobile). 24px gap.

`ProjectCard.vue`:
```
┌─────────────────────────────────────┐
│  ◉ python  my-saas-app          [⋮] │  ← header: language dot + name + menu
│                                     │
│  main  ↑2 ↓0                        │  ← branch + ahead/behind
│  Last opened 3 hours ago            │  ← relative time
│                                     │
│  [Open IDE]    [Chat about this]    │  ← actions (revealed on hover)
└─────────────────────────────────────┘
```

Actions (`Open IDE`, `Chat about this`) are hidden by default. On hover: they slide up from below the card content (translate +8px → 0, opacity 0 → 1, 200ms spring). This reveals them without shifting the card's size.

The language dot: a 10px circle, colored by language:
```
TypeScript → #3178c6 (TS blue)
Python     → #f7cb3c (Python yellow)
Go         → #00add8 (Go cyan)
Rust       → #ce412b (Rust orange)
JavaScript → #f7df1e (JS yellow)
Other      → --color-muted
```

**`NewProjectModal.vue`**: Opens with a scale+fade (0.96 → 1.0, 300ms spring). Full backdrop blur behind it.

Two tabs inside: "Local Path" | "From GitHub". Tab switching crossfades content (200ms).

Local Path: A filesystem path input with a "Browse..." button (opens a server-side directory picker that streams directory listings via a simple API). Language auto-detected from the path (looking for `package.json`, `go.mod`, `Cargo.toml`, `requirements.txt`).

From GitHub: A search field + repo list (reuses the github store). Clicking a repo pre-fills name + sets the suggested clone path. "Import" button triggers the clone SSE stream inline in the modal — the button becomes a progress view (streaming log lines) then transitions to a success state with a "Go to Project →" button.

**Empty state**: An SVG illustration — a simple folder with a small sparkle beside it. Headline "Your first project awaits." Subtext "Create one from a local path or import directly from GitHub." Two buttons side by side.

#### Settings View

A tabbed settings page. Tabs are a vertical list in a left sidebar (desktop), horizontal pills on mobile.

Each settings section is a stack of cards. Each card: a label (H3), a description (body-sm, muted), then the control. Generous spacing — 32px between cards.

**AI Providers tab**: Same provider cards as in the wizard, but here they are full-width and persistent. Each shows: status badge (Configured / Not set), masked key, "Rotate Key" button (replaces the key input), "Test" button (fires the live test). Key rotation: clicking "Rotate Key" transforms the card into an input state. The old key stays until the new one is validated.

**System tab**: A read-only info block showing hostname, LAN IP, Go version, docker version, app version. Below: an "Update" section. Clicking "Check for Updates" fetches the latest release. If available: "Update to v2.1.0 →" button. Clicking streams `update.sh` into an inline log terminal (same pattern as the wizard install log). When complete: "Restart required" badge appears. A "Restart now" button fires a server restart — the client shows a "Reconnecting..." overlay and auto-reconnects.

---

### Phase 8 — Installer & Packaging (Week 7)

**Goal**: One command that works. The output is beautiful. The result is instant.

```bash
curl -fsSL https://vibecodepc.com/install.sh | bash
```

The installer is designed as carefully as the app. It uses ANSI colors and Unicode to create a terminal experience that matches the app's aesthetic — because the installer *is* the first impression.

```
  VibeCodePC Installer

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  [1/7]  Detecting system...                   arm64, Raspberry Pi OS
  [2/7]  Installing dependencies...            git ✓  docker ✓
  [3/7]  Downloading VibeCodePC...             v1.0.0 (arm64)  ━━━━━━━━━━ 100%
  [4/7]  Downloading cloudflared...            2024.3.0 (arm64) ━━━━━━━━━━ 100%
  [5/7]  Creating directories...               ~/.vibecodepc/  ✓
  [6/7]  Registering services...               vibecodepc  ✓  vibecodepc-tunnel  ✓
  [7/7]  Starting & waiting for tunnel URL...  ●●●

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄

  ✓  VibeCodePC is running

     Local   →  http://raspberrypi.local:3000
     Remote  →  https://random-words.trycloudflare.com

     Open Remote from any device to run the setup wizard.
     The Remote URL changes on reboot until you set up
     a named tunnel — the wizard guides you through it.

  ┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄
```

Styling rules for the installer output:
- Section headers: bold white
- Step numbers: muted
- Step labels: white
- Results: muted green for ✓, dimmed for skipped
- The progress bars: `━━━━━━░░` Unicode blocks — determinate where known (file download), indeterminate (spinning `●●●`) while waiting
- The final output box: a single horizontal rule above and below, not a full box-drawing border — cleaner

**No Go compiler on the Pi.** The installer downloads a prebuilt binary from GitHub Releases. Cross-compiled via CI for `linux/arm64`, `linux/arm`, `linux/amd64`. The binary includes the entire frontend via `//go:embed`.

`update.sh`:
1. Downloads the latest binary to a temp path
2. Stops `vibecodepc.service`
3. Replaces the binary
4. Starts the service
5. Tails the service log for 5 seconds to confirm healthy startup
6. Prints "Updated to v2.1.0 ✓"

`uninstall.sh`:
1. Warns: "This will remove VibeCodePC and its config. Your project directories will not be touched."
2. Confirms with `[y/N]`
3. Stops and disables both services
4. Removes the binary and `~/.vibecodepc/app/`
5. Keeps project directories intact
6. Prints "VibeCodePC removed. Your projects are at ~/... — safe."

---

## Integration Details

### OpenCode Integration

```
Browser (xterm.js) ──WS──► /ws/terminal/:projectId ──► creack/pty ──► opencode
                   ◄──WS──                          ◄──── stdout
```

- Config auto-written to `~/.config/opencode/config.json` from stored provider keys
- Each project uses its own CWD — opencode loads that project's context automatically
- Multiple browser tabs share the same PTY session for the same project (all see the same output)
- opencode is a Node.js app; Node.js is installed in wizard Step 6

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

Scopes: `repo`, `read:user`. Token refresh not needed.
Git push authentication: token written to `~/.vibecodepc/.gitconfig`, referenced via `GIT_CONFIG` env.

### Cloudflare Tunnel

```
Internet ──► CF Edge ──► cloudflared ──► localhost:3000
```

Two modes in `cloudflare.go`:
- **Quick**: `cloudflared tunnel --url http://localhost:3000` — ephemeral URL, no account
- **Named**: `cloudflared tunnel --token <token>` — stable URL, free Cloudflare account

Mode selected on startup based on whether `cf_tunnel_token` exists in SQLite.

Quick tunnel URL parsed from cloudflared stdout: lines matching `trycloudflare.com`.

### Real-Time Metrics

```
GET /api/metrics/stream  →  SSE (text/event-stream), every 2s
  data: { cpu, ramUsedMb, ramTotalMb, diskUsedGb, diskTotalGb, tempC, uptimeS }
```

Go reads `/proc/stat` (CPU delta), `/proc/meminfo` (RAM), `df` output (disk), `/sys/class/thermal/thermal_zone0/temp` (Pi temp, 0 if unavailable), `/proc/uptime` (uptime). `time.Ticker` every 2s. Handler respects `r.Context().Done()`.

---

## Environment Variables

```env
PORT=3000
HOST=0.0.0.0
APP_ENV=production
DATA_DIR=/home/pi/.vibecodepc/data

GITHUB_CLIENT_ID=<your_client_id>
GITHUB_CLIENT_SECRET=<your_client_secret>
GITHUB_REDIRECT_URI=http://localhost:3000/auth/github/callback
```

All AI provider keys stored AES-256-GCM encrypted in SQLite. Never in `.env`.

---

## Security

- **Key encryption**: Machine-derived key (SHA256 of hostname + primary MAC). Never logged. Standard Go `crypto` package.
- **GitHub token**: `repo` + `read:user` scopes only. Stored encrypted. No org admin, no delete.
- **Terminal isolation**: Each project's opencode PTY runs as the app user — not root.
- **NanoClaw**: Runs in Docker container — isolated from host filesystem except the data volume.
- **Cloudflare tunnel**: End-to-end TLS. Zero open ports on the Pi.
- **Security headers**: Custom Chi middleware — CSP, HSTS, X-Frame-Options, X-Content-Type-Options, Referrer-Policy.
- **Optional password lock**: bcrypt hash in SQLite. Enabled in Settings → System. Protects the web UI with a session cookie (HttpOnly, Secure, SameSite=Strict).
- **Git credentials**: Written to `~/.vibecodepc/.gitconfig` — not the global `~/.gitconfig`. Referenced via `GIT_CONFIG` env on all git operations.

---

## Milestones

| # | Milestone | Target |
|---|---|---|
| M1 | Foundation: Go server + Vue shell + GitHub OAuth + design system | End of Week 1 |
| M2 | Full automated setup wizard (all 8 steps, all animations) | End of Week 2 |
| M3 | OpenCode terminal embedded, per-project sessions, git panel | End of Week 3 |
| M4 | NanoClaw chat + GitHub integration (repos, activity, import) | End of Week 4–5 |
| M5 | Real-time dashboard: service cards, vitals, activity, access panel | End of Week 5–6 |
| M6 | Projects view, Settings, empty states, polish pass | End of Week 6 |
| M7 | Prebuilt binaries via CI, installer, systemd, update/uninstall scripts | End of Week 7 |
