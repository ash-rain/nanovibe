<script setup lang="ts">
import { computed } from 'vue'
import type { Project, GitStatus } from '@/types'

const props = defineProps<{
  project: Project
  gitStatus?: GitStatus
}>()

const emit = defineEmits<{
  openIDE: [id: string]
  openAgent: [id: string]
}>()

const languageColors: Record<string, string> = {
  typescript: '#3178c6',
  javascript: '#f7df1e',
  python: '#3776ab',
  rust: '#ce422b',
  go: '#00add8',
  java: '#b07219',
  cpp: '#f34b7d',
  c: '#555555',
  ruby: '#cc342d',
  php: '#4F5D95',
}

const langColor = computed(() => {
  const lang = props.project.language?.toLowerCase()
  return lang ? (languageColors[lang] ?? '#64748b') : '#64748b'
})

const langLabel = computed(() => {
  if (!props.project.language) return null
  return props.project.language.charAt(0).toUpperCase() + props.project.language.slice(1)
})

function timeAgo(ts?: number): string {
  if (!ts) return 'never'
  const diff = Date.now() - ts
  const days = Math.floor(diff / 86400000)
  if (days === 0) return 'today'
  if (days === 1) return 'yesterday'
  return `${days}d ago`
}
</script>

<template>
  <div
    class="relative group bg-surface-800 border border-surface-600 rounded-lg p-5 shadow-card hover:border-surface-500 hover:-translate-y-0.5 transition-all duration-200 cursor-pointer"
  >
    <!-- Language dot -->
    <div class="flex items-center gap-2 mb-3">
      <div
        class="w-2.5 h-2.5 rounded-full flex-shrink-0"
        :style="`background-color: ${langColor}`"
      />
      <span v-if="langLabel" class="text-xs text-muted">{{ langLabel }}</span>
      <div class="flex-1" />
      <span class="text-xs text-subtle">{{ timeAgo(project.lastOpenedAt) }}</span>
    </div>

    <!-- Name + path -->
    <h3 class="text-sm font-semibold text-text mb-1">{{ project.name }}</h3>
    <p class="text-xs text-muted truncate font-mono mb-3">{{ project.path }}</p>

    <!-- Git info -->
    <div v-if="gitStatus" class="flex items-center gap-3 text-xs text-muted">
      <div class="flex items-center gap-1">
        <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 16 16">
          <path d="M5 3.25a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0zm0 2.122a2.25 2.25 0 1 0-1.5 0v.878A2.25 2.25 0 0 0 5.75 8.5h1.5v2.128a2.251 2.251 0 1 0 1.5 0V8.5h1.5a2.25 2.25 0 0 0 2.25-2.25v-.878a2.25 2.25 0 1 0-1.5 0v.878a.75.75 0 0 1-.75.75h-4.5A.75.75 0 0 1 5 6.25v-.878z"/>
        </svg>
        <span class="font-mono">{{ gitStatus.branch }}</span>
      </div>
      <div v-if="gitStatus.ahead > 0 || gitStatus.behind > 0" class="flex items-center gap-1">
        <span v-if="gitStatus.ahead > 0" class="text-success">↑{{ gitStatus.ahead }}</span>
        <span v-if="gitStatus.behind > 0" class="text-warning">↓{{ gitStatus.behind }}</span>
      </div>
      <div
        v-if="gitStatus.staged.length + gitStatus.unstaged.length + gitStatus.untracked.length > 0"
        class="flex items-center gap-1"
      >
        <span class="text-warning">
          {{ gitStatus.staged.length + gitStatus.unstaged.length + gitStatus.untracked.length }} changed
        </span>
      </div>
    </div>

    <!-- Hover actions -->
    <div
      class="absolute inset-0 flex items-center justify-center gap-3 bg-surface-800/95 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity duration-200"
    >
      <button
        @click="emit('openIDE', project.id)"
        class="px-3 py-2 rounded-md bg-primary text-white text-xs font-medium hover:bg-primary-dim transition-colors duration-120"
      >
        Open IDE
      </button>
      <button
        @click="emit('openAgent', project.id)"
        class="px-3 py-2 rounded-md bg-surface-600 text-text text-xs font-medium hover:bg-surface-500 transition-colors duration-120"
      >
        Chat
      </button>
    </div>
  </div>
</template>
