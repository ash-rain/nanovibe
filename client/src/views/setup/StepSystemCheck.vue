<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSetupStore } from '@/stores/setup'
import CheckRow from '@/components/ui/CheckRow.vue'

const router = useRouter()
const setupStore = useSetupStore()

const isRunning = ref(false)
const countdown = ref(0)
const countdownTimer = ref<ReturnType<typeof setInterval> | null>(null)
const fixLogs = ref<Record<string, string[]>>({})
const fixingId = ref<string | null>(null)

const allCriticalPass = computed(() =>
  setupStore.checks
    .filter((c) => c.critical)
    .every((c) => c.status === 'pass')
)

const allChecksComplete = computed(() =>
  setupStore.checks.length > 0 &&
  setupStore.checks.every(
    (c) => c.status === 'pass' || c.status === 'fail' || c.status === 'warning'
  )
)

async function runChecks() {
  isRunning.value = true
  await setupStore.runChecks()
  isRunning.value = false

  if (allCriticalPass.value) {
    startCountdown()
  }
}

function startCountdown() {
  countdown.value = 3
  countdownTimer.value = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(countdownTimer.value!)
      advance()
    }
  }, 1000)
}

async function advance() {
  await setupStore.advanceTo('cloudflare')
  router.push('/setup/cloudflare')
}

async function handleFix(checkId: string) {
  fixingId.value = checkId
  fixLogs.value[checkId] = []

  const es = setupStore.fixCheck(checkId)

  es.onmessage = (event) => {
    if (!fixLogs.value[checkId]) fixLogs.value[checkId] = []
    fixLogs.value[checkId].push(event.data)
  }

  es.onerror = () => {
    es.close()
    fixingId.value = null
    // Re-run checks after fix attempt
    runChecks()
  }

  es.addEventListener('done', () => {
    es.close()
    fixingId.value = null
    runChecks()
  })
}

onMounted(() => {
  runChecks()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h2 class="text-2xl font-bold text-text mb-2">System Check</h2>
      <p class="text-muted text-sm leading-relaxed">
        Checking that your system has everything needed to run VibeCodePC.
        Critical items must pass before continuing.
      </p>
    </div>

    <!-- Checks list -->
    <div class="space-y-2">
      <div v-if="isRunning && !setupStore.checks.length" class="space-y-2">
        <div
          v-for="i in 6"
          :key="i"
          class="h-14 rounded-md bg-surface-800 border border-surface-600 animate-pulse"
        />
      </div>

      <TransitionGroup name="list" tag="div" class="space-y-2">
        <CheckRow
          v-for="check in setupStore.checks"
          :key="check.id"
          :check="check"
          @fix="handleFix"
        />
      </TransitionGroup>
    </div>

    <!-- Fix logs -->
    <Transition name="slide-up">
      <div
        v-if="fixingId && fixLogs[fixingId]?.length"
        class="bg-surface-900 rounded-md p-4 font-mono text-xs text-muted max-h-40 overflow-y-auto space-y-0.5"
      >
        <div v-for="(line, i) in fixLogs[fixingId]" :key="i">{{ line }}</div>
      </div>
    </Transition>

    <!-- Status / actions -->
    <div class="flex items-center justify-between pt-2">
      <div>
        <div v-if="isRunning" class="flex items-center gap-2 text-sm text-muted">
          <div class="w-3 h-3 rounded-full border border-warning border-t-transparent animate-spin" />
          Running checks...
        </div>
        <div
          v-else-if="allCriticalPass"
          class="flex items-center gap-2 text-sm text-success"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
          All critical checks pass
          <span v-if="countdown > 0" class="text-muted"> — continuing in {{ countdown }}s</span>
        </div>
        <div v-else-if="allChecksComplete" class="text-sm text-danger">
          Some required checks failed. Please fix them to continue.
        </div>
      </div>

      <div class="flex gap-3">
        <button
          v-if="!isRunning"
          @click="runChecks"
          class="px-4 py-2 rounded-md text-sm text-muted hover:text-text bg-surface-700 hover:bg-surface-600 border border-surface-600 transition-all duration-120"
        >
          Re-check
        </button>
        <button
          v-if="allCriticalPass"
          @click="advance"
          class="px-4 py-2 rounded-md text-sm font-medium bg-primary text-white hover:bg-primary-dim transition-all duration-120"
        >
          Continue →
        </button>
      </div>
    </div>
  </div>
</template>
