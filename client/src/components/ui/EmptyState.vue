<script setup lang="ts">
import type { Component } from 'vue'

const props = defineProps<{
  title: string
  subtitle?: string
  icon?: Component
  actionLabel?: string
  actionTo?: string
}>()

const emit = defineEmits<{
  action: []
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-20 px-6 text-center">
    <div
      v-if="icon"
      class="w-14 h-14 rounded-xl bg-surface-700 border border-surface-600 flex items-center justify-center mb-5"
    >
      <component :is="icon" class="w-7 h-7 text-muted" />
    </div>

    <h3 class="text-lg font-semibold text-text mb-2">{{ title }}</h3>
    <p v-if="subtitle" class="text-sm text-muted max-w-sm">{{ subtitle }}</p>

    <div v-if="actionLabel" class="mt-6">
      <RouterLink
        v-if="actionTo"
        :to="actionTo"
        class="inline-flex items-center gap-2 px-4 py-2.5 rounded-md bg-primary text-white text-sm font-medium hover:bg-primary-dim transition-colors duration-120"
      >
        {{ actionLabel }}
      </RouterLink>
      <button
        v-else
        @click="emit('action')"
        class="inline-flex items-center gap-2 px-4 py-2.5 rounded-md bg-primary text-white text-sm font-medium hover:bg-primary-dim transition-colors duration-120"
      >
        {{ actionLabel }}
      </button>
    </div>
  </div>
</template>
