<script setup lang="ts">
import type { PR } from '@/types'

const props = defineProps<{
  pr: PR
}>()

const stateColor = {
  open: 'text-success bg-success/10 border-success/30',
  closed: 'text-danger bg-danger/10 border-danger/30',
  merged: 'text-primary-glow bg-primary/10 border-primary/30',
}
</script>

<template>
  <div class="flex gap-3 px-4 py-3 bg-surface-800 border border-surface-600 rounded-lg hover:border-surface-500 transition-all duration-120">
    <!-- PR number -->
    <span class="text-xs font-mono text-muted flex-shrink-0 mt-0.5">#{{ pr.number }}</span>

    <!-- Title + meta -->
    <div class="flex-1 min-w-0">
      <a
        :href="pr.htmlUrl"
        target="_blank"
        class="text-sm font-medium text-text hover:text-primary-glow transition-colors duration-120 truncate block"
      >
        {{ pr.title }}
      </a>
      <div class="flex items-center gap-2 mt-1 text-xs text-muted">
        <span class="font-mono">{{ pr.headBranch }}</span>
        <span>→</span>
        <span class="font-mono">{{ pr.baseBranch }}</span>
      </div>
    </div>

    <!-- State badge -->
    <span
      :class="[
        'flex-shrink-0 self-start text-[10px] font-medium px-1.5 py-0.5 rounded border capitalize',
        stateColor[pr.state as keyof typeof stateColor] ?? 'text-muted bg-surface-600',
      ]"
    >
      {{ pr.state }}
    </span>
  </div>
</template>
