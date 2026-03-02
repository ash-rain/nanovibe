<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useTerminal } from '@/composables/useTerminal'

const props = defineProps<{
  projectId?: string
}>()

const { terminalRef, connected, connecting, connect, disconnect, focus } =
  useTerminal()

const containerRef = ref<HTMLElement | null>(null)

watch(
  () => props.projectId,
  (id) => {
    if (id) {
      connect(id)
    } else {
      disconnect()
    }
  }
)

onMounted(() => {
  if (props.projectId) {
    connect(props.projectId)
  }
})

onUnmounted(() => {
  disconnect()
})
</script>

<template>
  <div class="relative flex-1 flex flex-col min-h-0 bg-surface-900" ref="containerRef">
    <!-- Terminal mount -->
    <div
      ref="terminalRef"
      class="flex-1 min-h-0"
      :class="{ 'opacity-50': !connected && !connecting }"
    />

    <!-- Connection overlay -->
    <Transition name="fade">
      <div
        v-if="!connected && !connecting && !projectId"
        class="absolute inset-0 flex flex-col items-center justify-center gap-4 bg-surface-900/90"
      >
        <div class="w-12 h-12 rounded-xl bg-surface-700 flex items-center justify-center">
          <svg class="w-6 h-6 text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
        <div class="text-center">
          <p class="text-sm font-medium text-text">No active session</p>
          <p class="text-xs text-muted mt-1">Select a project to open a terminal</p>
        </div>
      </div>
    </Transition>

    <!-- Connecting overlay -->
    <Transition name="fade">
      <div
        v-if="connecting"
        class="absolute inset-0 flex flex-col items-center justify-center gap-3 bg-surface-900/80"
      >
        <div class="w-5 h-5 rounded-full border-2 border-primary border-t-transparent animate-spin" />
        <p class="text-sm text-muted">Connecting...</p>
      </div>
    </Transition>
  </div>
</template>
