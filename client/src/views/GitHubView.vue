<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { CodeBracketSquareIcon, MagnifyingGlassIcon } from '@heroicons/vue/24/outline'
import { useGithubStore } from '@/stores/github'
import RepoCard from '@/components/github/RepoCard.vue'
import EventRow from '@/components/github/EventRow.vue'
import EmptyState from '@/components/ui/EmptyState.vue'

const router = useRouter()
const githubStore = useGithubStore()

const activeTab = ref<'repos' | 'activity'>('repos')
const search = ref('')
const importingRepo = ref<string | null>(null)
const importPath = ref('/home/pi/projects')
const showImportPath = ref(false)

const filteredRepos = computed(() => {
  if (!search.value.trim()) return githubStore.repos
  const q = search.value.toLowerCase()
  return githubStore.repos.filter((r) =>
    r.fullName.toLowerCase().includes(q) ||
    r.description?.toLowerCase().includes(q)
  )
})

async function loadMore() {
  await githubStore.fetchRepos(githubStore.reposPage + 1, search.value)
}

async function handleImport(repo: { fullName: string; cloneUrl: string }) {
  if (!importPath.value.trim()) {
    showImportPath.value = true
    return
  }

  const name = repo.fullName.split('/')[1]
  importingRepo.value = repo.fullName

  const es = githubStore.importRepo(repo.cloneUrl, name, importPath.value)

  es.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.projectId) {
        es.close()
        importingRepo.value = null
        router.push(`/app/ide/${data.projectId}`)
      }
    } catch {
      // raw log
    }
  }

  es.onerror = () => {
    es.close()
    importingRepo.value = null
  }
}

function connectGitHub() {
  const w = 600, h = 700
  const left = window.screenX + (window.outerWidth - w) / 2
  const top = window.screenY + (window.outerHeight - h) / 2
  const popup = window.open('/auth/github/start', 'github-auth', `width=${w},height=${h},left=${left},top=${top}`)

  const poll = setInterval(async () => {
    if (!popup || popup.closed) {
      clearInterval(poll)
      await githubStore.fetchStatus()
      if (githubStore.connected) {
        await githubStore.fetchRepos()
        await githubStore.fetchActivity()
      }
    }
  }, 500)
}

onMounted(async () => {
  await githubStore.fetchStatus()
  if (githubStore.connected) {
    await Promise.all([
      githubStore.fetchRepos(),
      githubStore.fetchActivity(),
    ])
  }
})
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <div class="px-6 lg:px-8 pt-6 pb-4 border-b border-surface-600 bg-surface-800">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-3">
          <h1 class="text-xl font-bold text-text">GitHub</h1>
          <div v-if="githubStore.user" class="flex items-center gap-2">
            <img
              :src="githubStore.user.avatarUrl"
              :alt="githubStore.user.login"
              class="w-5 h-5 rounded-full"
            />
            <span class="text-sm text-muted">{{ githubStore.user.login }}</span>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <!-- Import path -->
          <button
            v-if="githubStore.connected"
            @click="showImportPath = !showImportPath"
            class="text-xs text-muted hover:text-text transition-colors duration-120"
          >
            Dest: {{ importPath }}
          </button>
          <button
            v-if="!githubStore.connected"
            @click="connectGitHub"
            class="flex items-center gap-2 px-3 py-2 rounded-md bg-white text-gray-900 text-sm font-medium hover:bg-gray-100 transition-all duration-120"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
            </svg>
            Connect GitHub
          </button>
        </div>
      </div>

      <!-- Import path input (inline) -->
      <Transition name="slide-up">
        <div v-if="showImportPath" class="mb-4">
          <input
            v-model="importPath"
            placeholder="/home/pi/projects"
            class="w-full max-w-xs px-3 py-2 rounded-md bg-surface-700 border border-surface-600 focus:border-primary text-sm text-text font-mono outline-none transition-colors duration-120"
          />
        </div>
      </Transition>

      <!-- Tabs -->
      <div class="flex gap-0 border-b border-surface-600 -mb-4">
        <button
          @click="activeTab = 'repos'"
          :class="[
            'px-4 py-3 text-sm font-medium border-b-2 transition-colors duration-120',
            activeTab === 'repos'
              ? 'text-primary-glow border-primary'
              : 'text-muted border-transparent hover:text-text',
          ]"
        >
          Repositories
        </button>
        <button
          @click="activeTab = 'activity'"
          :class="[
            'px-4 py-3 text-sm font-medium border-b-2 transition-colors duration-120',
            activeTab === 'activity'
              ? 'text-primary-glow border-primary'
              : 'text-muted border-transparent hover:text-text',
          ]"
        >
          Activity
        </button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-6 lg:p-8">
      <!-- Not connected -->
      <div v-if="!githubStore.connected">
        <EmptyState
          title="GitHub not connected"
          subtitle="Connect your account to browse repositories and import projects."
          :icon="CodeBracketSquareIcon"
          action-label="Connect GitHub"
          @action="connectGitHub"
        />
      </div>

      <!-- Repos tab -->
      <div v-else-if="activeTab === 'repos'">
        <!-- Search -->
        <div class="relative mb-5">
          <MagnifyingGlassIcon class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted" />
          <input
            v-model="search"
            type="search"
            placeholder="Search repositories..."
            class="w-full pl-9 pr-4 py-2.5 rounded-md bg-surface-800 border border-surface-600 focus:border-primary text-sm text-text placeholder-muted outline-none transition-colors duration-120"
            @input="githubStore.fetchRepos(1, search)"
          />
        </div>

        <!-- Loading -->
        <div v-if="githubStore.isLoading && !githubStore.repos.length" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div v-for="i in 6" :key="i" class="h-32 rounded-lg bg-surface-800 border border-surface-600 animate-pulse" />
        </div>

        <!-- Grid -->
        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
          <RepoCard
            v-for="repo in filteredRepos"
            :key="repo.fullName"
            :repo="repo"
            :importing="importingRepo === repo.fullName"
            @import="handleImport"
          />
        </div>

        <!-- Load more -->
        <div v-if="filteredRepos.length && !search" class="flex justify-center">
          <button
            @click="loadMore"
            :disabled="githubStore.isLoading"
            class="px-4 py-2 rounded-md text-sm text-muted hover:text-text bg-surface-800 border border-surface-600 hover:bg-surface-700 transition-all duration-120 disabled:opacity-50"
          >
            {{ githubStore.isLoading ? 'Loading...' : 'Load more' }}
          </button>
        </div>
      </div>

      <!-- Activity tab -->
      <div v-else-if="activeTab === 'activity'">
        <div v-if="!githubStore.activity.length" class="py-12 text-center">
          <p class="text-sm text-muted">No recent activity</p>
        </div>
        <div v-else class="bg-surface-800 border border-surface-600 rounded-lg divide-y-0">
          <EventRow
            v-for="(event, i) in githubStore.activity"
            :key="i"
            :event="event"
          />
        </div>
      </div>
    </div>
  </div>
</template>
