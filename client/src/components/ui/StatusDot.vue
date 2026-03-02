<script setup lang="ts">
import { computed } from 'vue'
import type { ServiceStatus } from '@/types'

const props = defineProps<{
  status: ServiceStatus
  size?: 'xs' | 'sm' | 'md'
}>()

const size = computed(() => props.size ?? 'sm')

const sizeClasses = computed(() => ({
  xs: 'w-1.5 h-1.5',
  sm: 'w-2 h-2',
  md: 'w-2.5 h-2.5',
}[size.value]))

const colorClass = computed(() => ({
  running: 'bg-success',
  starting: 'bg-warning',
  error: 'bg-danger',
  stopped: 'bg-muted',
  connecting: 'bg-primary-glow',
}[props.status]))

const isAnimated = computed(() => props.status === 'running' || props.status === 'starting' || props.status === 'connecting')
</script>

<template>
  <span class="relative inline-flex">
    <span
      v-if="isAnimated"
      :class="[
        'absolute inset-0 rounded-full opacity-75',
        colorClass,
        status === 'running' ? 'animate-ping' : 'animate-breathe',
      ]"
    />
    <span
      :class="[
        'relative inline-block rounded-full flex-shrink-0',
        sizeClasses,
        colorClass,
      ]"
    />
  </span>
</template>
