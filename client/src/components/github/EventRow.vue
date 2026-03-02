<script setup lang="ts">
import type { GitHubEvent } from '@/types'

const props = defineProps<{
  event: GitHubEvent
}>()

const eventIcon = (type: string) =>
  ({
    PushEvent: '↑',
    PullRequestEvent: '⎇',
    IssuesEvent: '◎',
    WatchEvent: '★',
    ForkEvent: '⑂',
    CreateEvent: '✦',
    DeleteEvent: '✕',
  }[type] ?? '●')

const eventColor = (type: string) =>
  ({
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
  const diff = Date.now() - date.getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}
</script>

<template>
  <div class="flex gap-3 py-3 border-b border-surface-600 last:border-b-0">
    <!-- Icon -->
    <span :class="['text-sm font-mono flex-shrink-0 w-6 text-center mt-0.5', eventColor(event.type)]">
      {{ eventIcon(event.type) }}
    </span>

    <!-- Content -->
    <div class="flex-1 min-w-0">
      <p class="text-sm text-text">{{ event.description }}</p>
      <div class="flex items-center gap-2 mt-0.5 text-xs text-muted">
        <span class="font-mono truncate">{{ event.repoName }}</span>
        <span v-if="event.branch" class="flex-shrink-0 font-mono">@ {{ event.branch }}</span>
        <span class="flex-shrink-0 ml-auto">{{ timeAgo(event.createdAt) }}</span>
      </div>
    </div>
  </div>
</template>
