<script setup lang="ts">
import { ref, computed } from 'vue'
import type { SystemCheck } from '@/types'

const props = defineProps<{
  check: SystemCheck
}>()

const emit = defineEmits<{
  fix: [id: string]
}>()

const expanded = ref(false)
const fixLogs = ref<string[]>([])
const fixing = ref(false)

const borderColor = computed(() => ({
  pending: 'border-l-surface-500',
  running: 'border-l-warning',
  pass: 'border-l-success',
  fail: 'border-l-danger',
  warning: 'border-l-warning',
}[props.check.status]))

const statusIcon = computed(() => ({
  pending: '○',
  running: '◌',
  pass: '✓',
  fail: '✗',
  warning: '⚠',
}[props.check.status]))

const statusColor = computed(() => ({
  pending: 'text-muted',
  running: 'text-warning',
  pass: 'text-success',
  fail: 'text-danger',
  warning: 'text-warning',
}[props.check.status]))

function handleFix() {
  emit('fix', props.check.id)
}
</script>

<template>
  <div
    :class="[
      'rounded-md bg-surface-800 border-l-2 px-4 py-3 transition-all duration-200',
      borderColor,
    ]"
  >
    <div class="flex items-center gap-3">
      <!-- Status icon -->
      <span
        :class="[
          'text-sm font-mono font-bold flex-shrink-0 w-5 text-center',
          statusColor,
          check.status === 'running' ? 'animate-spin-slow' : '',
        ]"
      >
        {{ statusIcon }}
      </span>

      <!-- Label + detail -->
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2">
          <span class="text-sm font-medium text-text">{{ check.label }}</span>
          <span
            v-if="check.critical"
            class="text-[10px] font-medium uppercase tracking-wider px-1.5 py-0.5 rounded bg-danger/15 text-danger"
          >
            Required
          </span>
        </div>
        <p v-if="check.detail" class="text-xs text-muted mt-0.5 truncate">
          {{ check.detail }}
        </p>
      </div>

      <!-- Fix button -->
      <button
        v-if="check.fixable && (check.status === 'fail' || check.status === 'warning')"
        @click="handleFix"
        :disabled="fixing"
        class="flex-shrink-0 px-3 py-1.5 rounded text-xs font-medium bg-primary/15 text-primary-glow hover:bg-primary/25 transition-all duration-120 disabled:opacity-50"
      >
        {{ fixing ? 'Fixing...' : 'Fix it' }}
      </button>
    </div>
  </div>
</template>
