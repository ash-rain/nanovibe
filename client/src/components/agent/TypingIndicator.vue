<script setup lang="ts">
// Three dot wave animation for agent typing
</script>

<template>
  <div class="flex items-center gap-3 px-1">
    <!-- Agent avatar -->
    <div
      class="w-7 h-7 rounded-md bg-primary/20 border border-primary/30 flex items-center justify-center flex-shrink-0"
    >
      <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
        <path d="M2 4L7 1L12 4V10L7 13L2 10V4Z" stroke="#a78bfa" stroke-width="1.5" stroke-linejoin="round" />
        <circle cx="7" cy="7" r="1.5" fill="#a78bfa" />
      </svg>
    </div>

    <!-- Dots -->
    <div class="flex items-center gap-1 px-4 py-3 rounded-xl bg-surface-700 border border-surface-600">
      <span
        v-for="i in 3"
        :key="i"
        class="w-1.5 h-1.5 rounded-full bg-primary-glow animate-dot-wave"
        :style="`animation-delay: ${(i - 1) * 160}ms`"
      />
    </div>
  </div>
</template>
