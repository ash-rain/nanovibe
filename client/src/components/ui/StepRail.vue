<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  steps: Array<{ id: string; label: string }>
  currentStep: string
  completedSteps: string[]
  orientation?: 'vertical' | 'horizontal'
}>()

const orientation = computed(() => props.orientation ?? 'vertical')

function getStepStatus(stepId: string): 'completed' | 'current' | 'future' {
  if (props.completedSteps.includes(stepId)) return 'completed'
  if (props.currentStep === stepId) return 'current'
  return 'future'
}
</script>

<template>
  <!-- Vertical rail (desktop) -->
  <nav v-if="orientation === 'vertical'" class="space-y-1">
    <div
      v-for="(step, index) in steps"
      :key="step.id"
      class="relative"
    >
      <!-- Connector line -->
      <div
        v-if="index < steps.length - 1"
        class="absolute left-[11px] top-7 w-0.5 h-5"
        :class="
          getStepStatus(step.id) === 'completed'
            ? 'bg-success/40'
            : 'bg-surface-600'
        "
      />

      <div class="flex items-center gap-3 px-3 py-2">
        <!-- Node -->
        <div class="relative flex-shrink-0">
          <!-- Completed -->
          <div
            v-if="getStepStatus(step.id) === 'completed'"
            class="w-5 h-5 rounded-full bg-success/20 border border-success flex items-center justify-center"
          >
            <svg class="w-3 h-3 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
            </svg>
          </div>

          <!-- Current -->
          <div
            v-else-if="getStepStatus(step.id) === 'current'"
            class="w-5 h-5 rounded-full border-2 border-primary bg-primary/15 flex items-center justify-center"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-primary animate-pulse" />
          </div>

          <!-- Future -->
          <div
            v-else
            class="w-5 h-5 rounded-full border border-surface-500 bg-surface-700"
          />
        </div>

        <!-- Label -->
        <span
          class="text-sm transition-colors duration-200"
          :class="{
            'text-success font-medium': getStepStatus(step.id) === 'completed',
            'text-text font-medium': getStepStatus(step.id) === 'current',
            'text-muted': getStepStatus(step.id) === 'future',
          }"
        >
          {{ step.label }}
        </span>
      </div>
    </div>
  </nav>

  <!-- Horizontal rail (mobile) -->
  <nav v-else class="flex items-center gap-0">
    <div
      v-for="(step, index) in steps"
      :key="step.id"
      class="flex items-center"
    >
      <!-- Node -->
      <div class="relative flex-shrink-0">
        <div
          v-if="getStepStatus(step.id) === 'completed'"
          class="w-4 h-4 rounded-full bg-success/20 border border-success flex items-center justify-center"
        >
          <svg class="w-2.5 h-2.5 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <div
          v-else-if="getStepStatus(step.id) === 'current'"
          class="w-4 h-4 rounded-full border-2 border-primary bg-primary/15"
        />
        <div
          v-else
          class="w-4 h-4 rounded-full border border-surface-500 bg-surface-700"
        />
      </div>

      <!-- Connector -->
      <div
        v-if="index < steps.length - 1"
        class="w-8 h-0.5 mx-1"
        :class="
          getStepStatus(step.id) === 'completed'
            ? 'bg-success/40'
            : 'bg-surface-600'
        "
      />
    </div>
  </nav>
</template>
