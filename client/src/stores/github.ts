import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useFetch } from '@/composables/useFetch'
import type { GitHubUser, Repo, PR, GitHubEvent } from '@/types'

interface PrCreate {
  title: string
  body: string
  headBranch: string
  baseBranch: string
}

export const useGithubStore = defineStore('github', () => {
  const { get, post, del } = useFetch()

  const connected = ref(false)
  const user = ref<GitHubUser | null>(null)
  const repos = ref<Repo[]>([])
  const reposPage = ref(1)
  const repoSearch = ref('')
  const activity = ref<GitHubEvent[]>([])
  const prs = ref<Record<string, PR[]>>({})
  const isLoading = ref(false)

  async function fetchStatus() {
    const { data, error } = await get<{ connected: boolean; user?: GitHubUser }>(
      '/api/github/status'
    )
    if (!error && data) {
      connected.value = data.connected
      user.value = data.user ?? null
    }
    return { error }
  }

  async function fetchRepos(page = 1, search = '') {
    isLoading.value = true
    const params = new URLSearchParams({ page: String(page), search })
    const { data, error } = await get<{ repos: Repo[]; total: number }>(
      `/api/github/repos?${params}`
    )
    isLoading.value = false
    if (!error && data) {
      if (page === 1) {
        repos.value = data.repos
      } else {
        repos.value.push(...data.repos)
      }
      reposPage.value = page
    }
    return { error }
  }

  async function fetchActivity() {
    const { data, error } = await get<{ events: GitHubEvent[] }>(
      '/api/github/activity'
    )
    if (!error && data) {
      activity.value = data.events
    }
    return { error }
  }

  async function fetchPRs(owner: string, repo: string) {
    const { data, error } = await get<{ prs: PR[] }>(
      `/api/github/repos/${owner}/${repo}/prs`
    )
    if (!error && data) {
      prs.value[`${owner}/${repo}`] = data.prs
    }
    return { error }
  }

  async function createPR(owner: string, repo: string, params: PrCreate) {
    const { data, error } = await post<{ pr: PR }>(
      `/api/github/repos/${owner}/${repo}/prs`,
      params
    )
    return { data: data?.pr ?? null, error }
  }

  function importRepo(repoUrl: string, name: string, path: string): EventSource {
    const params = new URLSearchParams({ repoUrl, name, path })
    return new EventSource(`/api/github/import?${params}`)
  }

  async function importRepoPost(repoUrl: string, name: string, path: string) {
    const { data, error } = await post<{ projectId: string }>(
      '/api/github/import',
      { repoUrl, name, path }
    )
    return { data, error }
  }

  async function disconnect() {
    const { error } = await del('/api/github/disconnect')
    if (!error) {
      connected.value = false
      user.value = null
    }
    return { error }
  }

  return {
    connected,
    user,
    repos,
    reposPage,
    repoSearch,
    activity,
    prs,
    isLoading,
    fetchStatus,
    fetchRepos,
    fetchActivity,
    fetchPRs,
    createPR,
    importRepo,
    importRepoPost,
    disconnect,
  }
})
