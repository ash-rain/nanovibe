# CLAUDE.md — VibeCodePC.com Development Context

This file is the authoritative guide for Claude Code when working on this project.
Read this before touching any file.

---

## Project Identity

**Name**: VibeCodePC.com
**Purpose**: Self-hosted web app that turns a Raspberry Pi into an AI coding station.
**Primary CWD**: `~/Documents/_DEV/nanovibe` (repo root)

---

## Monorepo Structure

This is a pnpm workspace with two packages:

```
/server    → Fastify (Node.js + TypeScript) backend
/client    → Vue 3 + Vite + Tailwind CSS frontend
```

Run everything from the repo root:

```bash
pnpm dev          # Starts both server and client in watch mode
pnpm build        # Builds client, then server
pnpm typecheck    # tsc --noEmit in both packages
pnpm lint         # ESLint across both packages
```

---

## Server (`/server`)

### Stack
- **Framework**: Fastify v5 with TypeScript
- **Database**: `better-sqlite3` (synchronous, no async needed)
- **WebSocket**: `@fastify/websocket`
- **Static files**: `@fastify/static` serves the built client at `/`
- **CORS**: `@fastify/cors` (dev only — Vite proxy handles prod)
- **Terminal**: `node-pty` spawns opencode, xterm.js connects via WS

### Conventions
- All routes live in `server/src/routes/` — one file per domain
- All business logic lives in `server/src/services/` — never in routes
- Routes are thin: validate input → call service → return result
- Never throw raw errors from routes — use Fastify's `reply.code(x).send({ error: msg })`
- Database access only through `server/src/db/index.ts` singleton — never import `better-sqlite3` directly in routes or services
- Use synchronous `db.*` calls (better-sqlite3 is sync)
- All AI keys fetched via `keystore.ts` — never read settings table directly from routes

### Database Schema (SQLite)

```sql
-- Setup wizard state
CREATE TABLE setup_state (
  id INTEGER PRIMARY KEY,        -- always 1 (singleton)
  current_step TEXT NOT NULL,    -- 'welcome' | 'system_check' | ... | 'complete'
  completed_steps TEXT NOT NULL, -- JSON array of step names
  updated_at INTEGER NOT NULL    -- Unix timestamp
);

-- Projects
CREATE TABLE projects (
  id TEXT PRIMARY KEY,           -- UUID v4
  name TEXT NOT NULL,
  path TEXT NOT NULL UNIQUE,     -- Absolute path on filesystem
  language TEXT,                 -- 'typescript' | 'python' | etc.
  git_url TEXT,
  default_provider TEXT,         -- 'anthropic' | 'openai' | 'google' | 'ollama'
  created_at INTEGER NOT NULL,
  last_opened_at INTEGER
);

-- Settings (encrypted key-value)
CREATE TABLE settings (
  key TEXT PRIMARY KEY,          -- e.g. 'anthropic_key', 'cf_tunnel_token'
  value TEXT NOT NULL,           -- AES-256-GCM encrypted, base64
  updated_at INTEGER NOT NULL
);

-- Agent chat messages
CREATE TABLE agent_messages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  role TEXT NOT NULL,            -- 'user' | 'agent'
  content TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  project_id TEXT                -- optional project context
);
```

### Key Encryption Pattern (`db/crypto.ts`)

```typescript
// Machine key is derived from: SHA256(hostname + ':' + primaryMacAddress)
// This key never leaves the machine
import { deriveKey, encrypt, decrypt } from '../db/crypto'

// In keystore.ts:
export function setKey(name: string, value: string): void
export function getKey(name: string): string | null
```

---

## Client (`/client`)

### Stack
- **Framework**: Vue 3 with `<script setup>` Composition API — no Options API
- **Build**: Vite
- **Styling**: Tailwind CSS v4 — utility-first, no custom CSS unless unavoidable
- **State**: Pinia — one store per domain (`setup`, `projects`, `settings`, `agent`)
- **Routing**: Vue Router v4 with typed routes
- **Icons**: Heroicons via `@heroicons/vue`
- **Terminal**: `@xterm/xterm` + `@xterm/addon-fit` + `@xterm/addon-web-links`
- **Markdown**: `markdown-it` for agent responses
- **Syntax highlight**: `shiki` (lazy-loaded) for code blocks in chat

### Conventions
- Always use `<script setup lang="ts">` — never `<script>` alone
- Props typed with `defineProps<{ ... }>()`
- Emits typed with `defineEmits<{ ... }>()`
- Store access: `const store = useMyStore()` at top of `<script setup>`
- Fetch pattern: `useFetch` composable wrapping `fetch` with base URL from `import.meta.env`
- No direct `fetch()` calls in components — always go through a store action or composable
- Tailwind only — no scoped `<style>` blocks unless for third-party lib overrides
- Component naming: PascalCase files, `<MyComponent />` in templates

### Routing Structure

```
/                          → redirect to /setup or /app based on setup completion
/setup
  /setup/welcome           → StepWelcome
  /setup/system-check      → StepSystemCheck
  /setup/cloudflare        → StepCloudflare
  /setup/providers         → StepProviders
  /setup/opencode          → StepOpenCode
  /setup/nanoclaw          → StepNanoClaw
  /setup/complete          → StepComplete
/app
  /app/dashboard           → DashboardView
  /app/ide/:projectId?     → IDEView
  /app/agent               → AgentView
  /app/projects            → ProjectsView
  /app/settings            → SettingsView
```

Navigation guard in `router/index.ts`:
- If setup not complete → redirect any `/app/*` to `/setup/welcome`
- If setup complete → redirect `/setup/*` to `/app/dashboard`
- Setup state fetched from `GET /api/setup/state` on app boot

### Design System

**Colors** (defined in `tailwind.config.ts`):
```
primary:   indigo-900 base, violet-500 accent
success:   emerald-400
warning:   amber-400
danger:    rose-500
surface:   zinc-900, zinc-800, zinc-700 (dark mode layering)
text:      zinc-100, zinc-400 (muted)
```

**Typography**:
- UI text: `font-sans` → Inter (loaded via `@fontsource/inter`)
- Code/terminal: `font-mono` → JetBrains Mono (loaded via `@fontsource/jetbrains-mono`)

**Dark mode**: Default. System `prefers-color-scheme` respected. Toggle stored in localStorage.

**Animations**: Use `transition-*` Tailwind utilities. Page transitions via Vue `<Transition name="page">` with fade-and-slide.

---

## API Conventions

### Request/Response shape

```typescript
// Success
{ data: T }

// Error
{ error: string, detail?: string }
```

### Setup API
```
GET  /api/setup/state              → { currentStep, completedSteps }
POST /api/setup/state              → { step: string }
GET  /api/setup/check/system       → { checks: SystemCheck[] }
POST /api/setup/cloudflare/validate → { token: string } → { valid, tunnelUrl }
POST /api/setup/providers/test     → { provider, key } → { valid, models? }
POST /api/setup/opencode/install   → SSE stream of install stdout lines
POST /api/setup/nanoclaw/setup     → SSE stream of git clone + docker build
```

### Projects API
```
GET    /api/projects               → { data: Project[] }
POST   /api/projects               → body: ProjectCreate → { data: Project }
GET    /api/projects/:id           → { data: Project }
DELETE /api/projects/:id           → 204
GET    /api/projects/:id/status    → { gitBranch, fileCount, diskUsage, lastCommit }
```

### Settings API
```
GET  /api/settings/providers       → { providers: ProviderConfig[] }  (keys masked)
POST /api/settings/providers       → { provider, key }
GET  /api/settings/tunnel          → { mode, status, tunnelUrl, localUrl, connected }
POST /api/settings/tunnel/restart  → 204
GET  /api/settings/system          → { hostname, ip, localUrl, ramMb, diskGb, nodeVersion, dockerVersion }
// localUrl = "http://<hostname>.local:3000"
```

### Agent API
```
GET  /api/agent/stream             → SSE (text/event-stream)
POST /api/agent/message            → { content: string, projectId?: string }
GET  /api/agent/messages           → { data: AgentMessage[] }  (last 50)
DELETE /api/agent/messages         → 204  (clear history)
```

### Terminal WebSocket
```
WS   /ws/terminal/:projectId
  Client → Server frames:
    { type: 'input', data: string }
    { type: 'resize', cols: number, rows: number }
  Server → Client frames:
    { type: 'output', data: string }
    { type: 'exit', code: number }
```

---

## Services Reference

### `system-check.ts`
```typescript
interface SystemCheck {
  id: string
  label: string
  status: 'pending' | 'pass' | 'fail' | 'warning'
  detail?: string
}
runAllChecks(): Promise<SystemCheck[]>
```

### `opencode.ts`
```typescript
startSession(projectId: string, cwd: string): IPty
killSession(projectId: string): void
resizeSession(projectId: string, cols: number, rows: number): void
getSession(projectId: string): IPty | undefined
installOpenCode(): AsyncIterable<string>  // stdout lines
```

### `nanoclaw.ts`
```typescript
setup(config: NanoclawConfig): AsyncIterable<string>  // setup log lines
start(): Promise<void>    // docker start
stop(): Promise<void>     // docker stop
isRunning(): Promise<boolean>
insertUserMessage(content: string, projectId?: string): void  // write to nanoclaw SQLite
watchForResponses(since: number): AgentMessage[]  // poll nanoclaw outbox
```

### `cloudflare.ts`
```typescript
download(arch: 'arm64' | 'arm' | 'amd64'): Promise<void>
// Quick tunnel (no account): cloudflared tunnel --url http://localhost:3000
startQuickTunnel(): ChildProcess
// Named tunnel (account token): cloudflared tunnel --token <t>
startNamedTunnel(token: string): ChildProcess
stopTunnel(): void
getStatus(): TunnelStatus  // { mode: 'quick'|'named'|'none', connected, tunnelUrl, uptimeSeconds }
validateToken(token: string): Promise<{ valid: boolean, url?: string }>
getLocalUrl(): string  // e.g. "http://vibecodepc.local:3000"
```

### `keystore.ts`
```typescript
set(name: string, value: string): void
get(name: string): string | null
delete(name: string): void
exists(name: string): boolean
maskValue(value: string): string  // "sk-ant-••••••••1234"
```

---

## Environment & Running

### Development

```bash
# Root — starts both in parallel
pnpm dev

# Separately
pnpm --filter server dev   # nodemon + ts-node, port 3000
pnpm --filter client dev   # Vite, port 5173 (proxies /api/* and /ws/* to :3000)
```

### Production (Raspberry Pi)

```bash
pnpm build                 # Builds client → server/public/, then tsc server
node server/dist/index.js  # Or via systemd service
```

The app always runs on **port 3000**. No elevated privileges needed.

### Key Environment Variables

```env
PORT=3000
HOST=0.0.0.0
NODE_ENV=production
DATA_DIR=/home/pi/.vibecodepc/data
```

---

## Working with NanoClaw

NanoClaw is cloned into `~/.vibecodepc/nanoclaw/` and run via Docker.
Its SQLite database is at `~/.vibecodepc/nanoclaw/data/nanoclaw.db`.

To add the "web" platform to nanoclaw, a skill script is run post-clone:
`scripts/nanoclaw-web-bridge.ts` — this modifies nanoclaw's source to accept
messages from a virtual `web` source, inserted directly by our Fastify server.

Never modify the nanoclaw source directly in the repo. Instead apply the bridge
as a patch after cloning.

---

## Common Pitfalls

1. **node-pty on ARM**: Must be compiled for target arch. Always `npm rebuild node-pty` after installing on Raspberry Pi.
2. **better-sqlite3 on ARM**: Same — needs native compilation. Handled by `pnpm install` with node-gyp.
3. **opencode PATH**: After `npm install -g opencode`, the binary may not be in PATH when spawning via node-pty. Use full path: `/usr/local/bin/opencode` or detect via `which opencode`.
4. **Cloudflare tunnel arch**: The `cloudflared` binary download URL varies by arch. Always detect with `process.arch` → map to cloudflare release name.
5. **Quick tunnel URL is ephemeral**: The `*.trycloudflare.com` URL changes on every cloudflared restart. It's only for initial setup access. The wizard's Cloudflare step upgrades to a named (stable) tunnel.
6. **SSE and no reverse proxy**: Fastify serves directly on port 3000. No nginx in front. Do not introduce a reverse proxy.
7. **Vue Router and SPA**: Fastify must serve `index.html` for all non-API routes. Use `@fastify/static` with `wildcard: false` and a catch-all route.

---

## Testing

- Unit tests: Vitest for both `server/` and `client/`
- Server integration tests: Use `fastify.inject()` (no real HTTP)
- Client component tests: `@vue/test-utils` + Vitest + happy-dom
- E2E: Playwright (future — not in Phase 1–3)
- Test files colocated: `*.test.ts` next to source files

---

## Git Conventions

- Branch: `feature/<name>`, `fix/<name>`, `chore/<name>`
- Commits: Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`)
- No commits to `main` directly after Phase 1
