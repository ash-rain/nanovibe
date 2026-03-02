<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSetupStore } from '@/stores/setup'
import { useSettingsStore } from '@/stores/settings'
import CopyButton from '@/components/ui/CopyButton.vue'
import QrCode from '@/components/ui/QrCode.vue'
import confetti from 'canvas-confetti'

const router = useRouter()
const setupStore = useSetupStore()
const settingsStore = useSettingsStore()

const countdown = ref(8)
const showQr = ref(false)

function launchConfetti() {
  const duration = 3000
  const end = Date.now() + duration

  const frame = () => {
    confetti({
      particleCount: 3,
      angle: 60,
      spread: 55,
      origin: { x: 0 },
      colors: ['#7c3aed', '#a78bfa', '#34d399'],
    })
    confetti({
      particleCount: 3,
      angle: 120,
      spread: 55,
      origin: { x: 1 },
      colors: ['#7c3aed', '#a78bfa', '#34d399'],
    })

    if (Date.now() < end) {
      requestAnimationFrame(frame)
    }
  }

  requestAnimationFrame(frame)
}

async function goToDashboard() {
  await setupStore.advanceTo('complete')
  router.push('/app/dashboard')
}

const summaryItems = [
  { label: 'System', status: 'pass' },
  { label: 'Cloudflare Tunnel', status: 'pass' },
  { label: 'GitHub', status: setupStore.isStepCompleted('github') ? 'pass' : 'skip' },
  { label: 'AI Providers', status: 'pass' },
  { label: 'OpenCode', status: 'pass' },
  { label: 'NanoClaw', status: setupStore.isStepCompleted('nanoclaw') ? 'pass' : 'skip' },
]

onMounted(async () => {
  launchConfetti()
  await settingsStore.fetchTunnel()
  await settingsStore.fetchSystem()

  // Countdown to dashboard
  const timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(timer)
      goToDashboard()
    }
  }, 1000)
})
</script>

<template>
  <div class="space-y-8 text-center">
    <!-- Hero -->
    <div class="flex flex-col items-center gap-5">
      <div class="relative">
        <div class="w-20 h-20 rounded-full bg-success/15 border-2 border-success flex items-center justify-center animate-glow-pulse">
          <svg class="w-10 h-10 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </div>
      </div>

      <div>
        <h1 class="text-3xl font-bold text-text mb-2" style="letter-spacing: -0.025em">
          You're all set!
        </h1>
        <p class="text-muted text-base">
          VibeCodePC is configured and ready to use.
        </p>
      </div>
    </div>

    <!-- Summary grid -->
    <div class="grid grid-cols-2 gap-2 text-left">
      <div
        v-for="item in summaryItems"
        :key="item.label"
        class="flex items-center gap-2.5 bg-surface-800 border border-surface-600 rounded-lg px-4 py-3"
      >
        <div class="w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0"
          :class="item.status === 'pass' ? 'bg-success/15' : 'bg-surface-700'"
        >
          <svg
            v-if="item.status === 'pass'"
            class="w-3 h-3 text-success"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
          </svg>
          <span v-else class="text-[10px] text-muted">—</span>
        </div>
        <span class="text-sm font-medium text-text">{{ item.label }}</span>
      </div>
    </div>

    <!-- Access panel -->
    <div class="bg-surface-800 border border-surface-600 rounded-xl p-5 text-left space-y-3">
      <h3 class="text-sm font-semibold text-text">Access Your Dashboard</h3>

      <!-- Local URL -->
      <div>
        <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1.5">Local Network</p>
        <div class="flex items-center gap-2">
          <div class="flex-1 px-3 py-2 rounded-md bg-surface-900 border border-surface-600 text-sm font-mono text-text truncate">
            {{ settingsStore.tunnel.localUrl || settingsStore.system.localUrl || 'http://localhost:3000' }}
          </div>
          <CopyButton :text="settingsStore.tunnel.localUrl || settingsStore.system.localUrl || 'http://localhost:3000'" />
        </div>
      </div>

      <!-- Remote URL -->
      <div v-if="settingsStore.tunnel.connected && settingsStore.tunnel.tunnelUrl">
        <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1.5">Remote (Anywhere)</p>
        <div class="flex items-center gap-2">
          <div class="flex-1 px-3 py-2 rounded-md bg-surface-900 border border-surface-600 text-sm font-mono text-success truncate">
            {{ settingsStore.tunnel.tunnelUrl }}
          </div>
          <CopyButton :text="settingsStore.tunnel.tunnelUrl" />
          <button
            @click="showQr = !showQr"
            class="p-1.5 rounded text-muted hover:text-text hover:bg-surface-700 transition-all duration-120"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4H4v8h8V4zm0 8H4v8h8v-8zm8-8h-8v8h8V4zm-4 12h4v4h-4v-4zm-4 0h4v4h-4v-4z" />
            </svg>
          </button>
        </div>
        <Transition name="slide-up">
          <div v-if="showQr" class="mt-3 flex justify-center">
            <QrCode :url="settingsStore.tunnel.tunnelUrl" :size="140" />
          </div>
        </Transition>
      </div>
    </div>

    <!-- CTA -->
    <div class="flex flex-col items-center gap-3">
      <button
        @click="goToDashboard"
        class="flex items-center gap-2 px-8 py-3.5 rounded-xl bg-primary text-white font-semibold text-sm hover:bg-primary-dim transition-all duration-200 shadow-glow"
      >
        Go to Dashboard
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>
      <p class="text-xs text-muted">Redirecting in {{ countdown }}s...</p>
    </div>
  </div>
</template>
