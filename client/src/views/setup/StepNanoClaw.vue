<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSetupStore } from '@/stores/setup'
import QrCode from '@/components/ui/QrCode.vue'

const router = useRouter()
const setupStore = useSetupStore()

type StageStatus = 'pending' | 'running' | 'done' | 'error'

interface Stage {
  id: string
  label: string
  detail: string
  status: StageStatus
  logs: string[]
}

const stages = ref<Stage[]>([
  { id: 'clone', label: 'Clone repository', detail: 'git clone qwibitai/nanoclaw', status: 'pending', logs: [] },
  { id: 'configure', label: 'Configure environment', detail: 'Writing .env from provider keys', status: 'pending', logs: [] },
  { id: 'build', label: 'Build Docker image', detail: 'docker build (may take a few minutes)', status: 'pending', logs: [] },
  { id: 'start', label: 'Start container', detail: 'docker compose up -d', status: 'pending', logs: [] },
])

const selectedPlatform = ref('web')
const logsRef = ref<HTMLElement | null>(null)
const allDone = ref(false)
const error = ref('')

const platforms = [
  { id: 'web', label: 'Web Only', icon: '🌐' },
  { id: 'whatsapp', label: 'WhatsApp', icon: '📱' },
  { id: 'telegram', label: 'Telegram', icon: '✈️' },
  { id: 'discord', label: 'Discord', icon: '💬' },
]

function setStageStatus(id: string, status: StageStatus) {
  const stage = stages.value.find((s) => s.id === id)
  if (stage) stage.status = status
}

function addStageLog(id: string, log: string) {
  const stage = stages.value.find((s) => s.id === id)
  if (stage) stage.logs.push(log)
}

async function runSetup() {
  error.value = ''

  const es = new EventSource('/api/setup/nanoclaw/setup')
  let currentStage = 'clone'

  es.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.stage) {
        if (currentStage !== data.stage) {
          if (currentStage) setStageStatus(currentStage, 'done')
          currentStage = data.stage
          setStageStatus(data.stage, 'running')
        }
        if (data.log) addStageLog(data.stage, data.log)
      } else if (data.log) {
        addStageLog(currentStage, data.log)
      }

      setTimeout(() => {
        if (logsRef.value) {
          logsRef.value.scrollTop = logsRef.value.scrollHeight
        }
      }, 10)
    } catch {
      // raw log line
    }
  }

  es.addEventListener('done', () => {
    es.close()
    setStageStatus(currentStage, 'done')
    allDone.value = true
  })

  es.onerror = () => {
    es.close()
    setStageStatus(currentStage, 'error')
    error.value = 'Setup failed. You can retry or skip and set up later.'
  }

  setStageStatus('clone', 'running')
}

async function advance() {
  await setupStore.advanceTo('complete')
  router.push('/setup/complete')
}

async function skip() {
  await setupStore.advanceTo('complete')
  router.push('/setup/complete')
}

const stageIcon = (status: StageStatus) => ({
  pending: 'opacity-30',
  running: 'text-warning',
  done: 'text-success',
  error: 'text-danger',
}[status])

onMounted(() => {
  runSetup()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h2 class="text-2xl font-bold text-text mb-2">NanoClaw</h2>
      <p class="text-muted text-sm leading-relaxed">
        Setting up your AI agent with messaging platform support.
        NanoClaw runs in Docker on your Pi.
      </p>
    </div>

    <!-- Stage timeline -->
    <div class="space-y-2">
      <div
        v-for="(stage, i) in stages"
        :key="stage.id"
        class="relative"
      >
        <!-- Connector -->
        <div
          v-if="i < stages.length - 1"
          class="absolute left-[15px] top-[36px] w-0.5 h-5"
          :class="stage.status === 'done' ? 'bg-success/40' : 'bg-surface-600'"
        />

        <div class="flex gap-3 items-start">
          <!-- Status indicator -->
          <div class="flex-shrink-0 mt-1">
            <!-- Done -->
            <div
              v-if="stage.status === 'done'"
              class="w-7 h-7 rounded-full bg-success/15 border border-success flex items-center justify-center"
            >
              <svg class="w-3.5 h-3.5 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
              </svg>
            </div>

            <!-- Running -->
            <div
              v-else-if="stage.status === 'running'"
              class="w-7 h-7 rounded-full bg-warning/15 border border-warning flex items-center justify-center"
            >
              <div class="w-2 h-2 rounded-full bg-warning animate-pulse" />
            </div>

            <!-- Error -->
            <div
              v-else-if="stage.status === 'error'"
              class="w-7 h-7 rounded-full bg-danger/15 border border-danger flex items-center justify-center"
            >
              <span class="text-xs text-danger font-bold">✕</span>
            </div>

            <!-- Pending -->
            <div
              v-else
              class="w-7 h-7 rounded-full bg-surface-700 border border-surface-600 flex items-center justify-center"
            >
              <div class="w-2 h-2 rounded-full bg-surface-500" />
            </div>
          </div>

          <!-- Content -->
          <div class="flex-1 min-w-0 pb-5">
            <div class="flex items-center gap-2">
              <span
                class="text-sm font-medium"
                :class="stage.status === 'pending' ? 'text-muted' : 'text-text'"
              >
                {{ stage.label }}
              </span>
              <span v-if="stage.status === 'running'" class="text-xs text-warning">
                Running...
              </span>
            </div>
            <p class="text-xs text-muted font-mono">{{ stage.detail }}</p>

            <!-- Logs (only for running/error stage) -->
            <div
              v-if="stage.logs.length && (stage.status === 'running' || stage.status === 'error')"
              ref="logsRef"
              class="mt-2 bg-surface-900 rounded p-2 max-h-24 overflow-y-auto font-mono text-xs text-muted"
            >
              <div v-for="(log, j) in stage.logs.slice(-20)" :key="j" class="leading-5">
                {{ log }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Platform selector (show when done) -->
    <Transition name="slide-up">
      <div v-if="allDone" class="space-y-4">
        <h3 class="text-sm font-semibold text-text">Messaging Platform</h3>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="platform in platforms"
            :key="platform.id"
            @click="selectedPlatform = platform.id"
            :class="[
              'flex items-center gap-2 px-3 py-2.5 rounded-md border text-sm font-medium transition-all duration-120',
              selectedPlatform === platform.id
                ? 'bg-primary/15 border-primary/50 text-primary-glow'
                : 'bg-surface-700 border-surface-600 text-muted hover:text-text hover:bg-surface-600',
            ]"
          >
            <span>{{ platform.icon }}</span>
            {{ platform.label }}
          </button>
        </div>

        <!-- WhatsApp QR -->
        <Transition name="fade">
          <div
            v-if="selectedPlatform === 'whatsapp'"
            class="bg-surface-800 border border-surface-600 rounded-lg p-4 flex flex-col items-center gap-3"
          >
            <p class="text-sm text-muted">Scan to connect WhatsApp</p>
            <QrCode url="https://wa.me/" :size="140" />
            <p class="text-xs text-muted">Connect via WhatsApp Web</p>
          </div>
        </Transition>
      </div>
    </Transition>

    <p v-if="error" class="text-sm text-danger">{{ error }}</p>

    <!-- Actions -->
    <div class="flex items-center justify-between pt-2">
      <button
        @click="skip"
        class="text-sm text-muted hover:text-text transition-colors duration-120"
      >
        Skip NanoClaw →
      </button>
      <button
        @click="advance"
        :disabled="!allDone && !error"
        class="px-5 py-2.5 rounded-md text-sm font-medium bg-primary text-white hover:bg-primary-dim disabled:opacity-50 transition-all duration-120"
      >
        Continue →
      </button>
    </div>
  </div>
</template>
