<script setup lang="ts">
import type { Repo } from '@/types'

const props = defineProps<{
  repo: Repo
  importing?: boolean
}>()

const emit = defineEmits<{
  import: [repo: Repo]
}>()

const languageColors: Record<string, string> = {
  TypeScript: '#3178c6',
  JavaScript: '#f7df1e',
  Python: '#3776ab',
  Rust: '#ce422b',
  Go: '#00add8',
  Java: '#b07219',
  'C++': '#f34b7d',
  C: '#555555',
  Ruby: '#cc342d',
  PHP: '#4F5D95',
}

function getLangColor(lang: string): string {
  return languageColors[lang] ?? '#64748b'
}

function timeAgo(dateStr: string): string {
  const date = new Date(dateStr)
  const diff = Date.now() - date.getTime()
  const days = Math.floor(diff / 86400000)
  if (days === 0) return 'today'
  if (days < 30) return `${days}d ago`
  if (days < 365) return `${Math.floor(days / 30)}mo ago`
  return `${Math.floor(days / 365)}y ago`
}
</script>

<template>
  <div
    class="group relative bg-surface-800 border border-surface-600 rounded-lg p-4 hover:border-surface-500 hover:-translate-y-0.5 transition-all duration-200 shadow-card"
  >
    <!-- Header -->
    <div class="flex items-start justify-between gap-2 mb-2">
      <a
        :href="repo.htmlUrl"
        target="_blank"
        class="text-sm font-semibold text-text hover:text-primary-glow transition-colors duration-120 truncate"
      >
        {{ repo.fullName.split('/')[1] }}
      </a>
      <div v-if="repo.private" class="flex-shrink-0 text-[10px] px-1.5 py-0.5 rounded bg-surface-600 text-muted">
        Private
      </div>
    </div>

    <p v-if="repo.description" class="text-xs text-muted mb-3 line-clamp-2">
      {{ repo.description }}
    </p>

    <!-- Meta row -->
    <div class="flex items-center gap-3 text-xs text-muted">
      <div v-if="repo.language" class="flex items-center gap-1">
        <span
          class="w-2 h-2 rounded-full"
          :style="`background-color: ${getLangColor(repo.language)}`"
        />
        {{ repo.language }}
      </div>
      <div v-if="repo.stars > 0" class="flex items-center gap-1">
        <span>★</span>
        {{ repo.stars }}
      </div>
      <span class="ml-auto">{{ timeAgo(repo.pushedAt) }}</span>
    </div>

    <!-- Import overlay on hover -->
    <div
      class="absolute inset-0 flex items-center justify-center bg-surface-800/90 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity duration-200"
    >
      <button
        @click="emit('import', repo)"
        :disabled="importing"
        class="px-4 py-2 rounded-md bg-primary text-white text-xs font-medium hover:bg-primary-dim disabled:opacity-50 transition-all duration-120"
      >
        {{ importing ? 'Importing...' : 'Import' }}
      </button>
    </div>
  </div>
</template>
