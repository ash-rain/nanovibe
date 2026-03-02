<script setup lang="ts">
import { computed } from 'vue'
import type { GitHubEvent } from '@/types'

const props = defineProps<{
  events: GitHubEvent[]
  loading?: boolean
}>()

const eventIcon = (type: string) => ({
  PushEvent: '↑',
  PullRequestEvent: '⎇',
  IssuesEvent: '◎',
  WatchEvent: '★',
  ForkEvent: '⑂',
  CreateEvent: '✦',
  DeleteEvent: '✕',
}[type] ?? '●')

const eventColor = (type: string) => ({
  PushEvent: 'text-primary-glow',
  PullRequestEvent: 'text-success',
  IssuesEvent: 'text-warning',
  WatchEvent: 'text-warning',
  ForkEvent: 'text-text',
  CreateEvent: 'text-success',
  DeleteEvent: 'text-danger',
}[type] ?? 'text-muted')

function timeAgo(dateStr: string): string {
  const date = new Date(dateStr)
  const now = Date.now()
  const diff = now - date.getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}
</script>

<template>
  <div class="bg-surface-800 border border-surface-600 rounded-lg p-5 shadow-card">
    <h3 class="text-sm font-semibold text-text mb-4">GitHub Activity</h3>

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <div v-for="i in 4" :key="i" class="flex gap-3 animate-pulse">
        <div class="w-5 h-5 rounded bg-surface-600" />
        <div class="flex-1 space-y-1.5">
          <div class="h-3 bg-surface-600 rounded w-3/4" />
          <div class="h-2 bg-surface-600 rounded w-1/2" />
        </div>
      </div>
    </div>

    <!-- Empty -->
    <div v-else-if="!events.length" class="py-6 text-center">
      <p class="text-sm text-muted">No recent activity</p>
    </div>

    <!-- Events -->
    <div v-else class="space-y-3">
      <div
        v-for="(event, i) in events.slice(0, 8)"
        :key="i"
        class="flex gap-3"
      >
        <span
          :class="['text-sm font-mono flex-shrink-0 w-5 text-center mt-0.5', eventColor(event.type)]"
        >
          {{ eventIcon(event.type) }}
        </span>
        <div class="flex-1 min-w-0">
          <p class="text-sm text-text truncate">{{ event.description }}</p>
          <div class="flex items-center gap-2 mt-0.5">
            <span class="text-xs text-muted truncate">{{ event.repoName }}</span>
            <span class="text-xs text-subtle">·</span>
            <span class="text-xs text-muted flex-shrink-0">{{ timeAgo(event.createdAt) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
