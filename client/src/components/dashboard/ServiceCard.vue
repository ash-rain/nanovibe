<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Component } from 'vue'
import StatusDot from '@/components/ui/StatusDot.vue'
import type { ServiceStatus } from '@/types'

const props = defineProps<{
  title: string
  icon?: Component
  status: ServiceStatus
  metric?: string | number
  metricLabel?: string
  primaryAction?: string
  logs?: string[]
}>()

const emit = defineEmits<{
  action: []
}>()

const expanded = ref(false)

const statusLabel = computed(() => ({
  running: 'Running',
  starting: 'Starting',
  error: 'Error',
  stopped: 'Stopped',
  connecting: 'Connecting',
}[props.status]))

const statusTextColor = computed(() => ({
  running: 'text-success',
  starting: 'text-warning',
  error: 'text-danger',
  stopped: 'text-muted',
  connecting: 'text-primary-glow',
}[props.status]))
</script>

<template>
  <div
    class="bg-surface-800 border border-surface-600 rounded-lg p-5 shadow-card hover:border-surface-500 transition-all duration-200 hover:-translate-y-0.5"
    :class="status === 'running' ? 'hover:shadow-glow' : ''"
  >
    <!-- Header -->
    <div class="flex items-start justify-between mb-4">
      <div class="flex items-center gap-3">
        <div
          class="w-9 h-9 rounded-md bg-surface-700 border border-surface-600 flex items-center justify-center flex-shrink-0"
        >
          <component v-if="icon" :is="icon" class="w-5 h-5 text-muted" />
          <div v-else class="w-4 h-4 rounded bg-surface-500" />
        </div>
        <div>
          <h3 class="text-sm font-semibold text-text">{{ title }}</h3>
          <div class="flex items-center gap-1.5 mt-0.5">
            <StatusDot :status="status" size="xs" />
            <span :class="['text-xs', statusTextColor]">{{ statusLabel }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Metric -->
    <div v-if="metric !== undefined" class="mb-4">
      <div class="text-2xl font-bold text-text tabular-nums">{{ metric }}</div>
      <div v-if="metricLabel" class="text-xs text-muted mt-0.5">{{ metricLabel }}</div>
    </div>

    <!-- Primary action -->
    <div v-if="primaryAction" class="flex items-center justify-between">
      <button
        @click="emit('action')"
        :disabled="status === 'starting'"
        class="text-xs font-medium px-3 py-1.5 rounded bg-primary/15 text-primary-glow hover:bg-primary/25 transition-all duration-120 disabled:opacity-50"
      >
        {{ primaryAction }}
      </button>

      <button
        v-if="logs && logs.length"
        @click="expanded = !expanded"
        class="text-xs text-muted hover:text-text transition-colors duration-120"
      >
        {{ expanded ? 'Hide logs' : 'View logs' }}
      </button>
    </div>

    <!-- Log expansion -->
    <Transition name="slide-up">
      <div
        v-if="expanded && logs && logs.length"
        class="mt-3 pt-3 border-t border-surface-600"
      >
        <div
          class="bg-surface-900 rounded-md p-3 max-h-32 overflow-y-auto font-mono text-xs text-muted space-y-0.5"
        >
          <div v-for="(line, i) in logs" :key="i" class="leading-5">{{ line }}</div>
        </div>
      </div>
    </Transition>
  </div>
</template>
