<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSetupStore } from '@/stores/setup'

const router = useRouter()
const setupStore = useSetupStore()

const installLogs = ref<string[]>([])
const installing = ref(false)
const installed = ref(false)
const installError = ref('')
const version = ref('')
const logsRef = ref<HTMLElement | null>(null)

async function checkInstalled() {
  try {
    const res = await fetch('/api/setup/opencode/status')
    if (res.ok) {
      const data = await res.json()
      installed.value = data.installed
      version.value = data.version || ''
    }
  } catch {
    // not installed
  }
}

async function install() {
  installing.value = true
  installLogs.value = []
  installError.value = ''

  const es = new EventSource('/api/setup/opencode/install')

  es.onmessage = (event) => {
    installLogs.value.push(event.data)
    // Auto-scroll
    setTimeout(() => {
      if (logsRef.value) {
        logsRef.value.scrollTop = logsRef.value.scrollHeight
      }
    }, 10)
  }

  es.addEventListener('done', (event) => {
    es.close()
    installing.value = false
    installed.value = true
    try {
      const data = JSON.parse((event as MessageEvent).data)
      version.value = data.version || ''
    } catch {
      // ignore
    }
  })

  es.onerror = () => {
    es.close()
    installing.value = false
    installError.value = 'Installation failed. Check the logs above.'
  }
}

async function advance() {
  await setupStore.advanceTo('nanoclaw')
  router.push('/setup/nanoclaw')
}

onMounted(async () => {
  await checkInstalled()
  if (!installed.value) {
    install()
  }
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h2 class="text-2xl font-bold text-text mb-2">OpenCode</h2>
      <p class="text-muted text-sm leading-relaxed">
        Installing the AI coding agent that powers your in-browser terminal.
        This runs locally on your Pi.
      </p>
    </div>

    <!-- Status card -->
    <div class="bg-surface-800 border border-surface-600 rounded-lg p-5">
      <div class="flex items-center gap-4 mb-4">
        <div class="w-10 h-10 rounded-md bg-violet-500/15 border border-violet-500/30 flex items-center justify-center">
          <svg class="w-5 h-5 text-primary-glow" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </div>
        <div class="flex-1">
          <div class="flex items-center gap-2">
            <span class="text-sm font-semibold text-text">opencode</span>
            <span
              v-if="installed && version"
              class="text-[10px] px-1.5 py-0.5 rounded bg-success/15 text-success font-mono"
            >
              v{{ version }}
            </span>
            <span
              v-else-if="installed"
              class="text-[10px] px-1.5 py-0.5 rounded bg-success/15 text-success"
            >
              Installed
            </span>
          </div>
          <div class="flex items-center gap-2 mt-0.5">
            <div
              v-if="installing"
              class="w-2 h-2 rounded-full bg-warning animate-pulse"
            />
            <div
              v-else-if="installed"
              class="w-2 h-2 rounded-full bg-success"
            />
            <div v-else class="w-2 h-2 rounded-full bg-muted" />
            <span class="text-xs text-muted">
              {{ installing ? 'Installing...' : installed ? 'Ready' : 'Not installed' }}
            </span>
          </div>
        </div>
      </div>

      <!-- Installation log terminal -->
      <div
        v-if="installLogs.length || installing"
        ref="logsRef"
        class="bg-surface-900 rounded-md p-3 h-40 overflow-y-auto font-mono text-xs space-y-0.5"
      >
        <div
          v-for="(line, i) in installLogs"
          :key="i"
          class="text-muted leading-5"
          :class="line.includes('error') || line.includes('Error') ? 'text-danger' : ''"
        >
          {{ line }}
        </div>
        <div v-if="installing" class="flex items-center gap-2 text-primary-glow">
          <span class="w-2 h-2 rounded-full border border-primary-glow border-t-transparent animate-spin" />
          <span>Running npm install -g opencode...</span>
        </div>
      </div>

      <!-- Error -->
      <div v-if="installError" class="mt-3">
        <p class="text-sm text-danger mb-2">{{ installError }}</p>
        <button
          @click="install"
          class="text-xs px-3 py-1.5 rounded bg-surface-700 text-muted hover:text-text transition-all duration-120"
        >
          Retry installation
        </button>
      </div>
    </div>

    <!-- Already installed state: provider selector -->
    <div v-if="installed" class="bg-surface-800 border border-success/20 rounded-lg p-4">
      <p class="text-sm text-success flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        OpenCode is ready to use
      </p>
    </div>

    <!-- Actions -->
    <div class="flex items-center justify-end pt-2">
      <button
        @click="advance"
        :disabled="installing && !installed"
        class="px-5 py-2.5 rounded-md text-sm font-medium bg-primary text-white hover:bg-primary-dim disabled:opacity-50 transition-all duration-120"
      >
        Continue →
      </button>
    </div>
  </div>
</template>
