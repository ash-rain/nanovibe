<script setup lang="ts">
import { ref } from 'vue'
import type { TunnelStatus } from '@/types'
import CopyButton from '@/components/ui/CopyButton.vue'
import QrCode from '@/components/ui/QrCode.vue'
import StatusDot from '@/components/ui/StatusDot.vue'

const props = defineProps<{
  tunnel: TunnelStatus
  localUrl?: string
}>()

const showQr = ref<string | null>(null)

const modeLabel = {
  none: 'Local Only',
  quick: 'Quick Tunnel',
  named: 'Named Tunnel',
}
</script>

<template>
  <div class="bg-surface-800 border border-surface-600 rounded-lg p-5 shadow-card">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-sm font-semibold text-text">Access</h3>
      <div class="flex items-center gap-1.5">
        <StatusDot :status="tunnel.connected ? 'running' : 'stopped'" size="xs" />
        <span class="text-xs text-muted">{{ modeLabel[tunnel.mode] }}</span>
      </div>
    </div>

    <div class="space-y-3">
      <!-- Local URL -->
      <div>
        <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1.5">
          Local
        </p>
        <div class="flex items-center gap-2">
          <a
            :href="tunnel.localUrl || localUrl"
            target="_blank"
            class="flex-1 min-w-0 px-3 py-2 rounded-md bg-surface-700 border border-surface-600 text-sm font-mono text-primary-glow truncate hover:border-surface-500 transition-colors"
          >
            {{ tunnel.localUrl || localUrl || 'http://localhost:3000' }}
          </a>
          <CopyButton :text="tunnel.localUrl || localUrl || 'http://localhost:3000'" />
          <button
            @click="showQr = showQr === 'local' ? null : 'local'"
            class="p-1.5 rounded text-muted hover:text-text hover:bg-surface-700 transition-all duration-120"
            title="Show QR code"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4H4v8h8V4zm0 8H4v8h8v-8zm8-8h-8v8h8V4zm-4 12h4v4h-4v-4zm-4 0h4v4h-4v-4z" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Remote URL (if connected) -->
      <div v-if="tunnel.connected && tunnel.tunnelUrl">
        <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1.5">
          Remote
        </p>
        <div class="flex items-center gap-2">
          <a
            :href="tunnel.tunnelUrl"
            target="_blank"
            class="flex-1 min-w-0 px-3 py-2 rounded-md bg-surface-700 border border-surface-600 text-sm font-mono text-success truncate hover:border-surface-500 transition-colors"
          >
            {{ tunnel.tunnelUrl }}
          </a>
          <CopyButton :text="tunnel.tunnelUrl" />
          <button
            @click="showQr = showQr === 'remote' ? null : 'remote'"
            class="p-1.5 rounded text-muted hover:text-text hover:bg-surface-700 transition-all duration-120"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4H4v8h8V4zm0 8H4v8h8v-8zm8-8h-8v8h8V4zm-4 12h4v4h-4v-4zm-4 0h4v4h-4v-4z" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- QR popover -->
    <Transition name="slide-up">
      <div
        v-if="showQr"
        class="mt-4 pt-4 border-t border-surface-600 flex flex-col items-center gap-2"
      >
        <QrCode
          :url="showQr === 'remote' ? tunnel.tunnelUrl : (tunnel.localUrl || localUrl || 'http://localhost:3000')"
          :size="128"
        />
        <p class="text-xs text-muted">Scan to open on mobile</p>
      </div>
    </Transition>
  </div>
</template>
