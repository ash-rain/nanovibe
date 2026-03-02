export interface SystemCheck {
  id: string
  label: string
  critical: boolean
  status: 'pending' | 'running' | 'pass' | 'fail' | 'warning'
  detail: string
  fixable: boolean
}

export interface Project {
  id: string
  name: string
  path: string
  language?: string
  githubUrl?: string
  gitRemote?: string
  defaultProvider?: string
  createdAt: number
  lastOpenedAt?: number
}

export interface GitStatus {
  branch: string
  ahead: number
  behind: number
  staged: string[]
  unstaged: string[]
  untracked: string[]
}

export interface Branch {
  name: string
  current: boolean
  remote: boolean
}

export interface SystemMetrics {
  cpu: number
  ramUsedMb: number
  ramTotalMb: number
  diskUsedGb: number
  diskTotalGb: number
  tempC: number
  uptimeS: number
}

export interface AgentMessage {
  id: number
  role: 'user' | 'agent'
  content: string
  createdAt: number
  projectId?: string
}

export interface ProviderConfig {
  name: string
  key: string
  valid?: boolean
}

export interface TunnelStatus {
  mode: 'none' | 'quick' | 'named'
  connected: boolean
  tunnelUrl: string
  localUrl: string
  uptimeS: number
}

export interface GitHubUser {
  login: string
  avatarUrl: string
  publicRepos: number
}

export interface Repo {
  fullName: string
  description: string
  language: string
  cloneUrl: string
  htmlUrl: string
  stars: number
  private: boolean
  pushedAt: string
}

export interface PR {
  number: number
  title: string
  state: string
  headBranch: string
  baseBranch: string
  htmlUrl: string
  createdAt: string
}

export interface GitHubEvent {
  type: string
  repoName: string
  description: string
  branch: string
  createdAt: string
}

export interface SystemInfo {
  hostname: string
  ip: string
  localUrl: string
  goVersion: string
  dockerVersion: string
  appVersion: string
}

export type ServiceStatus = 'running' | 'starting' | 'error' | 'stopped' | 'connecting'
