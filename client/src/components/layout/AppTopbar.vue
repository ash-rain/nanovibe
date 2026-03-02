<script setup lang="ts">
import {
  XMarkIcon,
  PlusIcon,
} from '@heroicons/vue/24/outline'
import StatusDot from '@/components/ui/StatusDot.vue'

const props = defineProps<{
  projectName?: string
  branch?: string
  connected?: boolean
}>()

const emit = defineEmits<{
  killSession: []
  newSession: []
}>()
</script>

<template>
  <div
    class="h-10 flex items-center px-4 border-b border-surface-600 bg-surface-800 gap-3 flex-shrink-0"
  >
    <!-- Project name + branch -->
    <div class="flex items-center gap-2 min-w-0">
      <span class="text-sm font-medium text-text truncate">
        {{ projectName || 'No project' }}
      </span>
      <div
        v-if="branch"
        class="flex items-center gap-1 px-2 py-0.5 rounded bg-surface-700 text-xs text-muted font-mono"
      >
        <svg class="w-3 h-3" fill="none" viewBox="0 0 16 16">
          <path
            d="M5 3.25a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0zm0 2.122a2.25 2.25 0 1 0-1.5 0v.878A2.25 2.25 0 0 0 5.75 8.5h1.5v2.128a2.251 2.251 0 1 0 1.5 0V8.5h1.5a2.25 2.25 0 0 0 2.25-2.25v-.878a2.25 2.25 0 1 0-1.5 0v.878a.75.75 0 0 1-.75.75h-4.5A.75.75 0 0 1 5 6.25v-.878z"
            fill="currentColor"
          />
        </svg>
        {{ branch }}
      </div>
    </div>

    <!-- Status -->
    <div class="flex items-center gap-1.5">
      <StatusDot :status="connected ? 'running' : 'stopped'" size="xs" />
      <span class="text-xs text-muted">
        {{ connected ? 'Connected' : 'Disconnected' }}
      </span>
    </div>

    <div class="flex-1" />

    <!-- Actions -->
    <div class="flex items-center gap-2">
      <button
        @click="emit('newSession')"
        class="flex items-center gap-1.5 px-2.5 py-1 rounded text-xs text-muted hover:text-text hover:bg-surface-700 transition-all duration-120"
      >
        <PlusIcon class="w-3 h-3" />
        New
      </button>
      <button
        @click="emit('killSession')"
        class="flex items-center gap-1.5 px-2.5 py-1 rounded text-xs text-danger/70 hover:text-danger hover:bg-danger/10 transition-all duration-120"
      >
        <XMarkIcon class="w-3 h-3" />
        Kill
      </button>
    </div>
  </div>
</template>
