import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useFetch } from '@/composables/useFetch'
import type { Project, GitStatus } from '@/types'

interface ProjectCreate {
  name: string
  path: string
  language?: string
  githubUrl?: string
  defaultProvider?: string
}

export const useProjectsStore = defineStore('projects', () => {
  const { get, post, del } = useFetch()

  const list = ref<Project[]>([])
  const active = ref<Project | null>(null)
  const gitStatus = ref<Record<string, GitStatus>>({})
  const isLoading = ref(false)

  async function fetchAll() {
    isLoading.value = true
    const { data, error } = await get<{ data: Project[] }>('/api/projects')
    isLoading.value = false
    if (!error && data) {
      list.value = data.data
    }
    return { error }
  }

  async function create(projectData: ProjectCreate) {
    const { data, error } = await post<{ data: Project }>(
      '/api/projects',
      projectData
    )
    if (!error && data) {
      list.value.push(data.data)
    }
    return { data: data?.data ?? null, error }
  }

  async function remove(id: string) {
    const { error } = await del(`/api/projects/${id}`)
    if (!error) {
      list.value = list.value.filter((p) => p.id !== id)
      if (active.value?.id === id) {
        active.value = null
      }
    }
    return { error }
  }

  function setActive(id: string) {
    const project = list.value.find((p) => p.id === id) ?? null
    active.value = project
    if (project) {
      project.lastOpenedAt = Date.now()
    }
  }

  async function fetchGitStatus(id: string) {
    const { data, error } = await get<GitStatus>(
      `/api/projects/${id}/git/status`
    )
    if (!error && data) {
      gitStatus.value[id] = data
    }
    return { data, error }
  }

  async function gitCommit(id: string, message: string) {
    const { error } = await post(`/api/projects/${id}/git/commit`, { message })
    if (!error) {
      await fetchGitStatus(id)
    }
    return { error }
  }

  async function gitPush(id: string) {
    const { error } = await post(`/api/projects/${id}/git/push`)
    if (!error) {
      await fetchGitStatus(id)
    }
    return { error }
  }

  async function gitPull(id: string) {
    const { error } = await post(`/api/projects/${id}/git/pull`)
    if (!error) {
      await fetchGitStatus(id)
    }
    return { error }
  }

  async function gitCheckout(id: string, branch: string) {
    const { error } = await post(`/api/projects/${id}/git/checkout`, {
      branch,
    })
    if (!error) {
      await fetchGitStatus(id)
    }
    return { error }
  }

  function getRecentProjects(limit = 4): Project[] {
    return [...list.value]
      .sort((a, b) => (b.lastOpenedAt ?? 0) - (a.lastOpenedAt ?? 0))
      .slice(0, limit)
  }

  return {
    list,
    active,
    gitStatus,
    isLoading,
    fetchAll,
    create,
    remove,
    setActive,
    fetchGitStatus,
    gitCommit,
    gitPush,
    gitPull,
    gitCheckout,
    getRecentProjects,
  }
})
