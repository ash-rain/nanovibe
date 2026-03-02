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

pnpm workspace — two packages:

```
/server    → Fastify (Node.js + TypeScript) backend
/client    → Vue 3 + Vite + Tailwind CSS frontend
```

Root scripts:
```bash
pnpm dev          # Both packages in watch mode (concurrently)
pnpm build        # Client → server/public/; then tsc for server
pnpm typecheck    # tsc --noEmit in both packages
pnpm lint         # ESLint across both
```

---

## Server (`/server`)

### Stack
- **Framework**: Fastify v5 + TypeScript
- **Database**: `better-sqlite3` (synchronous)
- **WebSocket**: `@fastify/websocket`
- **Static files**: `@fastify/static` serves built client at `/`
- **Security**: `@fastify/helmet` (CSP, HSTS, etc.)
- **CORS**: `@fastify/cors` (dev only — Vite proxy in prod)
- **Terminal**: `node-pty` spawns opencode, piped over WS
- **Git**: `simple-git` for per-project git operations
- **GitHub API**: `@octokit/rest` with stored OAuth token
- **Metrics**: reads `/proc/stat`, `/proc/meminfo`, `/sys/class/thermal/` directly

### Conventions
- Routes in `server/src/routes/` — one file per domain, thin (validate → service → respond)
- Business logic in `server/src/services/` — never in routes
- Never `import Database from 'better-sqlite3'` outside `db/index.ts`
- All SQLite access via the singleton from `db/index.ts`
- Use synchronous `db.*` calls throughout (better-sqlite3 is sync by design)
- All AI keys and GitHub token via `keystore.ts` — never read `settings` table directly in routes
- Error responses: `reply.code(N).send({ error: 'message', detail?: 'more' })`
- SSE streams: set `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`; write `data: ...\n\n` frames

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

### Key Encryption (`db/crypto.ts`)

Machine key = `SHA256(os.hostname() + ':' + primaryMacAddress())`.
This key is derived at runtime, never stored, never logged.

```typescript
// Usage via keystore.ts only:
export function set(name: string, value: string): void
export function get(name: string): string | null
export function del(name: string): void
export function exists(name: string): boolean
export function mask(value: string): string  // "sk-ant-••••••••1234"
```

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
GET  /api/settings/system                → { hostname, ip, localUrl, nodeVersion, dockerVersion, appVersion }
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

### `system-check.ts`
```typescript
interface SystemCheck {
  id: 'node' | 'docker' | 'docker-daemon' | 'ram' | 'disk' | 'internet' | 'git'
  label: string
  critical: boolean
  status: 'pending' | 'running' | 'pass' | 'fail' | 'warning'
  detail?: string
  fixable: boolean
}
runAllChecks(): Promise<SystemCheck[]>
runCheck(id: string): Promise<SystemCheck>
// Fix actions: each streams log lines via AsyncIterable<string>
fixNode(): AsyncIterable<string>
fixDocker(): AsyncIterable<string>
fixDockerDaemon(): AsyncIterable<string>
fixGit(): AsyncIterable<string>
```

### `opencode.ts`
```typescript
isInstalled(): Promise<boolean>
install(): AsyncIterable<string>          // npm install -g opencode stdout
getVersion(): Promise<string>
writeConfig(providers: ProviderConfig[]): void   // writes ~/.config/opencode/config.json
startSession(projectId: string, cwd: string): IPty
killSession(projectId: string): void
resizeSession(projectId: string, cols: number, rows: number): void
getSession(projectId: string): IPty | undefined
```

### `nanoclaw.ts`
```typescript
isCloned(): boolean
clone(): AsyncIterable<string>
isRunning(): Promise<boolean>
start(): Promise<void>
stop(): Promise<void>
writeEnv(providers: ProviderConfig[]): void
insertUserMessage(content: string, projectId?: string): void
watchForResponses(since: number): AgentMessage[]
```

### `cloudflare.ts`
```typescript
download(arch: 'arm64' | 'arm' | 'amd64'): Promise<void>
startQuickTunnel(): ChildProcess
startNamedTunnel(token: string): ChildProcess
stopTunnel(): void
getStatus(): TunnelStatus
// TunnelStatus: { mode: 'quick'|'named'|'none', connected, tunnelUrl, localUrl, uptimeS }
validateToken(token: string): Promise<{ valid: boolean, url?: string }>
```

### `github.ts`
```typescript
getClient(): Octokit             // throws if not authenticated
isAuthenticated(): boolean
getUser(): Promise<GitHubUser>
listRepos(page: number, search?: string): Promise<Repo[]>
listPRs(owner: string, repo: string): Promise<PR[]>
createPR(owner: string, repo: string, params: PrCreate): Promise<PR>
getActivity(login: string): Promise<GitHubEvent[]>
importRepo(repoUrl: string, destPath: string): AsyncIterable<string>
```

### `git.ts`
```typescript
// All methods take an absolute project path
status(path: string): Promise<GitStatus>
diff(path: string): Promise<string>
commit(path: string, message: string): Promise<void>
push(path: string): Promise<void>
pull(path: string): Promise<void>
branches(path: string): Promise<Branch[]>
checkout(path: string, branch: string): Promise<void>
clone(url: string, dest: string): AsyncIterable<string>
// Uses GIT_CONFIG=~/.vibecodepc/.gitconfig for credential injection
```

### `metrics.ts`
```typescript
interface SystemMetrics {
  cpu: number        // 0–100
  ramUsedMb: number
  ramTotalMb: number
  diskUsedGb: number
  diskTotalGb: number
  tempC: number      // from /sys/class/thermal (0 if unavailable)
  uptimeS: number
}
read(): Promise<SystemMetrics>
// Caller streams via setInterval + SSE
```

### `keystore.ts`
```typescript
set(name: string, value: string): void
get(name: string): string | null
del(name: string): void
exists(name: string): boolean
mask(value: string): string
```

---

## Environment & Running

### Development
```bash
pnpm dev
# Server: nodemon + ts-node on :3000
# Client: Vite on :5173 (proxies /api/*, /ws/*, /auth/* to :3000)
```

### Production (Raspberry Pi)
```bash
pnpm build
node server/dist/index.js
# Or via systemd: vibecodepc.service
```

### Environment Variables

```env
PORT=3000
HOST=0.0.0.0
NODE_ENV=production
DATA_DIR=/home/pi/.vibecodepc/data

# GitHub OAuth (register at github.com/settings/developers)
GITHUB_CLIENT_ID=your_client_id
GITHUB_CLIENT_SECRET=your_client_secret
GITHUB_REDIRECT_URI=http://localhost:3000/auth/github/callback
```

`GITHUB_REDIRECT_URI` must match the app's public URL in production (the Cloudflare tunnel URL or local URL). The wizard Settings → GitHub tab provides a reminder to update this.

---

## Working with NanoClaw

- Cloned into `~/.vibecodepc/nanoclaw/` during setup wizard Step 7
- SQLite at `~/.vibecodepc/nanoclaw/data/nanoclaw.db` (volume-mounted into Docker)
- `scripts/nanoclaw-web-bridge.ts` applied post-clone: registers `web` as a virtual platform
- Never modify nanoclaw's source in-repo — apply as a patch after cloning

---

## Common Pitfalls

1. **node-pty on ARM**: Native module — run `npm rebuild node-pty` after `pnpm install` on Pi.
2. **better-sqlite3 on ARM**: Same native compile requirement.
3. **opencode PATH**: Use `which opencode` at runtime; don't assume `/usr/local/bin`.
4. **cloudflared arch**: `process.arch` → `arm64` / `arm` / `x64`; map to cloudflare release names (`linux-arm64`, `linux-arm`, `linux-amd64`).
5. **SSE and quick tunnel URL**: Parse stdout of `cloudflared` for lines matching `trycloudflare.com` to extract the assigned URL.
6. **GitHub OAuth redirect URI**: Must exactly match what's registered in the GitHub OAuth App. In dev this is `http://localhost:3000/auth/github/callback`; in production it must be the Cloudflare tunnel URL. Surfaced clearly in Settings.
7. **simple-git credentials**: Inject via `GIT_CONFIG` env pointing to `~/.vibecodepc/.gitconfig` which uses `url.<base>.insteadOf` + credential helper storing the GitHub token.
8. **Metrics on non-Pi Linux**: `/sys/class/thermal/` may not exist — `tempC` returns 0 gracefully.
9. **SPA catch-all**: Use `@fastify/static` + a `GET *` route that sends `index.html` for all non-API, non-WS, non-auth paths.

---

## Testing

- Unit: Vitest in both packages; colocated `*.test.ts`
- Server integration: `fastify.inject()` (no real HTTP)
- Client: `@vue/test-utils` + Vitest + happy-dom
- E2E: Playwright (Phase 8+)
- CI: GitHub Actions on `main` — lint + typecheck + unit tests

---

## Git Conventions

- Branches: `feature/<name>`, `fix/<name>`, `chore/<name>`
- Commits: Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`)
- No force-push to `main`
