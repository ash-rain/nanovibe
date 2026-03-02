<script setup lang="ts">
import { computed } from 'vue'
import Sparkline from '@/components/ui/Sparkline.vue'
import type { SystemMetrics } from '@/types'

const props = defineProps<{
  current: SystemMetrics | null
  history: SystemMetrics[]
}>()

const cpuHistory = computed(() => props.history.map((m) => m.cpu))
const ramHistory = computed(() =>
  props.history.map((m) => (m.ramUsedMb / m.ramTotalMb) * 100)
)
const diskHistory = computed(() =>
  props.history.map((m) => (m.diskUsedGb / m.diskTotalGb) * 100)
)
const tempHistory = computed(() => props.history.map((m) => m.tempC))

const ramPct = computed(() => {
  if (!props.current) return 0
  return Math.round((props.current.ramUsedMb / props.current.ramTotalMb) * 100)
})

const diskPct = computed(() => {
  if (!props.current) return 0
  return Math.round(
    (props.current.diskUsedGb / props.current.diskTotalGb) * 100
  )
})

function formatUptime(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `${h}h ${m}m`
}

const vitals = computed(() => [
  {
    label: 'CPU',
    value: props.current ? `${Math.round(props.current.cpu)}%` : '--',
    history: cpuHistory.value,
    color: 'var(--color-primary-glow)',
  },
  {
    label: 'RAM',
    value: props.current
      ? `${ramPct.value}%`
      : '--',
    sub: props.current
      ? `${Math.round(props.current.ramUsedMb / 1024 * 10) / 10} / ${Math.round(props.current.ramTotalMb / 1024 * 10) / 10} GB`
      : '',
    history: ramHistory.value,
    color: 'var(--color-success)',
  },
  {
    label: 'Disk',
    value: props.current ? `${diskPct.value}%` : '--',
    sub: props.current
      ? `${props.current.diskUsedGb.toFixed(1)} / ${props.current.diskTotalGb.toFixed(1)} GB`
      : '',
    history: diskHistory.value,
    color: 'var(--color-warning)',
  },
  {
    label: 'Temp',
    value: props.current
      ? props.current.tempC > 0
        ? `${props.current.tempC.toFixed(1)}°C`
        : 'N/A'
      : '--',
    history: tempHistory.value,
    color: 'var(--color-danger)',
  },
])
</script>

<template>
  <div class="bg-surface-800 border border-surface-600 rounded-lg p-5 shadow-card">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-sm font-semibold text-text">System Vitals</h3>
      <span v-if="current" class="text-xs text-muted">
        Up {{ formatUptime(current.uptimeS) }}
      </span>
    </div>

    <div class="space-y-4">
      <div v-for="vital in vitals" :key="vital.label" class="flex items-end gap-3">
        <!-- Label + value -->
        <div class="w-24 flex-shrink-0">
          <div class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1">
            {{ vital.label }}
          </div>
          <div class="text-xl font-bold text-text tabular-nums leading-none">
            {{ vital.value }}
          </div>
          <div v-if="vital.sub" class="text-xs text-muted mt-0.5">{{ vital.sub }}</div>
        </div>

        <!-- Sparkline -->
        <div class="flex-1 min-w-0">
          <Sparkline
            :data="vital.history.length ? vital.history : [0, 0]"
            :color="vital.color"
            :height="36"
          />
        </div>
      </div>
    </div>
  </div>
</template>
