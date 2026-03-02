<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import StepRail from '@/components/ui/StepRail.vue'
import { useSetupStore } from '@/stores/setup'

const route = useRoute()
const setupStore = useSetupStore()

const STEPS = [
  { id: 'welcome', label: 'Welcome' },
  { id: 'system-check', label: 'System Check' },
  { id: 'cloudflare', label: 'Cloudflare' },
  { id: 'github', label: 'GitHub' },
  { id: 'providers', label: 'AI Providers' },
  { id: 'opencode', label: 'OpenCode' },
  { id: 'nanoclaw', label: 'NanoClaw' },
  { id: 'complete', label: 'Complete' },
]

const currentStepId = computed(() => {
  const path = route.path
  const match = path.match(/\/setup\/(.+)/)
  return match ? match[1] : 'welcome'
})
</script>

<template>
  <div class="min-h-screen bg-surface-900 flex">
    <!-- Left rail (desktop) -->
    <div class="hidden lg:flex flex-col w-64 bg-surface-800 border-r border-surface-600 p-6">
      <!-- Logo -->
      <div class="flex items-center gap-3 mb-10">
        <div class="w-8 h-8 rounded-lg bg-primary flex items-center justify-center flex-shrink-0">
          <svg width="16" height="16" viewBox="0 0 14 14" fill="none">
            <path d="M2 4L7 1L12 4V10L7 13L2 10V4Z" stroke="white" stroke-width="1.5" stroke-linejoin="round" />
            <circle cx="7" cy="7" r="1.5" fill="white" />
          </svg>
        </div>
        <div>
          <div class="text-sm font-semibold text-text">VibeCodePC</div>
          <div class="text-[10px] text-muted uppercase tracking-wider">Setup</div>
        </div>
      </div>

      <StepRail
        :steps="STEPS"
        :current-step="currentStepId"
        :completed-steps="setupStore.completedSteps"
        orientation="vertical"
      />
    </div>

    <!-- Main content -->
    <div class="flex-1 flex flex-col min-h-screen">
      <!-- Mobile step indicator -->
      <div class="lg:hidden px-6 pt-6 pb-4 flex items-center justify-between border-b border-surface-600">
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded bg-primary flex items-center justify-center">
            <svg width="12" height="12" viewBox="0 0 14 14" fill="none">
              <path d="M2 4L7 1L12 4V10L7 13L2 10V4Z" stroke="white" stroke-width="1.5" stroke-linejoin="round" />
              <circle cx="7" cy="7" r="1.5" fill="white" />
            </svg>
          </div>
          <span class="text-sm font-semibold text-text">Setup</span>
        </div>
        <StepRail
          :steps="STEPS"
          :current-step="currentStepId"
          :completed-steps="setupStore.completedSteps"
          orientation="horizontal"
        />
      </div>

      <!-- Step content with transition -->
      <div class="flex-1 flex items-start justify-center p-6 lg:p-12 overflow-y-auto">
        <div class="w-full max-w-xl">
          <RouterView v-slot="{ Component, route: currentRoute }">
            <Transition name="page" mode="out-in">
              <component :is="Component" :key="currentRoute.path" />
            </Transition>
          </RouterView>
        </div>
      </div>
    </div>
  </div>
</template>
