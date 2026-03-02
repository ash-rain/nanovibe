<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSetupStore } from '@/stores/setup'
import { useSettingsStore } from '@/stores/settings'
import StatusDot from '@/components/ui/StatusDot.vue'
import CopyButton from '@/components/ui/CopyButton.vue'
import QrCode from '@/components/ui/QrCode.vue'

const router = useRouter()
const setupStore = useSetupStore()
const settingsStore = useSettingsStore()

const upgradeToken = ref('')
const upgrading = ref(false)
const upgradeError = ref('')
const showUpgrade = ref(false)
const showQr = ref(false)

let pollTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  await settingsStore.fetchTunnel()
  // Poll every 3s until the tunnel URL appears (cloudflared takes ~10s to assign a URL)
  if (!settingsStore.tunnel.connected) {
    let attempts = 0
    pollTimer = setInterval(async () => {
      attempts++
      await settingsStore.fetchTunnel()
      if (settingsStore.tunnel.connected || attempts >= 20) {
        clearInterval(pollTimer!)
        pollTimer = null
      }
    }, 3000)
  }
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
})

async function restart() {
  await settingsStore.restartTunnel()
}

async function upgrade() {
  if (!upgradeToken.value.trim()) return
  upgrading.value = true
  upgradeError.value = ''
  const { error } = await settingsStore.upgradeTunnel(upgradeToken.value.trim())
  upgrading.value = false
  if (error) {
    upgradeError.value = error
  } else {
    showUpgrade.value = false
    await settingsStore.fetchTunnel()
  }
}

async function advance() {
  await setupStore.advanceTo('github')
  router.push('/setup/github')
}

async function skip() {
  await setupStore.advanceTo('github')
  router.push('/setup/github')
}

const modeLabel = {
  none: 'Not running',
  quick: 'Quick Tunnel',
  named: 'Named Tunnel',
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h2 class="text-2xl font-bold text-text mb-2">Cloudflare Tunnel</h2>
      <p class="text-muted text-sm leading-relaxed">
        Access your Pi securely from anywhere with a Cloudflare tunnel.
        No port forwarding required.
      </p>
    </div>

    <!-- Status card -->
    <div class="bg-surface-800 border border-surface-600 rounded-lg p-5">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-md bg-orange-500/15 border border-orange-500/30 flex items-center justify-center">
            <svg class="w-5 h-5 text-orange-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418" />
            </svg>
          </div>
          <div>
            <div class="text-sm font-semibold text-text">Cloudflare Tunnel</div>
            <div class="flex items-center gap-1.5 mt-0.5">
              <StatusDot
                :status="settingsStore.tunnel.connected ? 'running' : 'stopped'"
                size="xs"
              />
              <span class="text-xs text-muted">
                {{ modeLabel[settingsStore.tunnel.mode] }}
              </span>
            </div>
          </div>
        </div>
        <button
          @click="restart"
          class="px-3 py-1.5 text-xs text-muted hover:text-text bg-surface-700 hover:bg-surface-600 rounded border border-surface-600 transition-all duration-120"
        >
          Restart
        </button>
      </div>

      <!-- Data flow diagram -->
      <div class="flex items-center gap-2 py-3 px-2 mb-4">
        <div class="px-3 py-1.5 rounded bg-surface-700 text-xs font-mono text-muted">Your Device</div>
        <div class="flex-1 flex items-center gap-1">
          <div class="flex-1 h-px bg-surface-600 relative">
            <div
              class="absolute inset-y-0 left-0 w-1/2 bg-gradient-to-r from-transparent to-orange-500/50"
              :class="settingsStore.tunnel.connected ? 'animate-pulse' : ''"
            />
          </div>
          <div class="text-[10px] text-orange-400/70 px-1">CF</div>
          <div class="flex-1 h-px bg-surface-600 relative">
            <div
              class="absolute inset-y-0 right-0 w-1/2 bg-gradient-to-l from-transparent to-orange-500/50"
              :class="settingsStore.tunnel.connected ? 'animate-pulse' : ''"
            />
          </div>
        </div>
        <div class="px-3 py-1.5 rounded bg-surface-700 text-xs font-mono text-muted">Internet</div>
      </div>

      <!-- URL pills -->
      <div class="space-y-2">
        <!-- Local URL -->
        <div>
          <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1.5">Local URL</p>
          <div class="flex items-center gap-2">
            <div class="flex-1 px-3 py-2 rounded-md bg-surface-900 border border-surface-600 text-sm font-mono text-text truncate">
              {{ settingsStore.tunnel.localUrl || 'http://localhost:3000' }}
            </div>
            <CopyButton :text="settingsStore.tunnel.localUrl || 'http://localhost:3000'" />
          </div>
        </div>

        <!-- Tunnel URL (if connected) -->
        <div v-if="settingsStore.tunnel.connected && settingsStore.tunnel.tunnelUrl">
          <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1.5">Remote URL</p>
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
    </div>

    <!-- Upgrade section (collapsible) -->
    <div class="bg-surface-800 border border-surface-600 rounded-lg overflow-hidden">
      <button
        @click="showUpgrade = !showUpgrade"
        class="w-full flex items-center justify-between px-5 py-4 text-sm text-text hover:bg-surface-700 transition-colors duration-120"
      >
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
          </svg>
          <span>Upgrade to Named Tunnel (optional)</span>
        </div>
        <svg
          class="w-4 h-4 text-muted transition-transform duration-200"
          :class="showUpgrade ? 'rotate-180' : ''"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      <Transition name="slide-up">
        <div v-if="showUpgrade" class="px-5 pb-5 space-y-3 border-t border-surface-600">
          <p class="text-xs text-muted pt-3">
            A named tunnel gives you a stable, permanent URL. Get your token from the Cloudflare Zero Trust dashboard.
          </p>
          <input
            v-model="upgradeToken"
            type="password"
            placeholder="Tunnel token..."
            class="w-full px-3 py-2.5 rounded-md bg-surface-700 border border-surface-600 focus:border-primary text-sm text-text font-mono outline-none transition-colors duration-120"
          />
          <p v-if="upgradeError" class="text-xs text-danger">{{ upgradeError }}</p>
          <button
            @click="upgrade"
            :disabled="upgrading || !upgradeToken.trim()"
            class="px-4 py-2 rounded-md text-sm font-medium bg-primary text-white hover:bg-primary-dim disabled:opacity-50 transition-all duration-120"
          >
            {{ upgrading ? 'Upgrading...' : 'Upgrade' }}
          </button>
        </div>
      </Transition>
    </div>

    <!-- Actions -->
    <div class="flex items-center justify-between pt-2">
      <button
        @click="skip"
        class="text-sm text-muted hover:text-text transition-colors duration-120"
      >
        Skip for now →
      </button>
      <button
        @click="advance"
        class="px-5 py-2.5 rounded-md text-sm font-medium bg-primary text-white hover:bg-primary-dim transition-all duration-120"
      >
        Continue →
      </button>
    </div>
  </div>
</template>
