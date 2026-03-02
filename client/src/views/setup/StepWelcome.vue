<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSetupStore } from '@/stores/setup'

const router = useRouter()
const setupStore = useSetupStore()

// Animation state
const phase = ref<'orb' | 'hostname' | 'tagline' | 'button'>('orb')
const typedHostname = ref('')
const showTagline = ref(false)
const showButton = ref(false)
const hostname = ref('vibecodepc')

// Fetch hostname from API
async function fetchHostname() {
  try {
    const res = await fetch('/api/settings/system')
    if (res.ok) {
      const data = await res.json()
      hostname.value = data.hostname || 'vibecodepc'
    }
  } catch {
    // Use default
  }
}

// Type hostname letter by letter
async function typeHostname(text: string) {
  for (let i = 0; i <= text.length; i++) {
    typedHostname.value = text.slice(0, i)
    await new Promise((r) => setTimeout(r, 60))
  }
}

async function runAnimation() {
  // Phase 1: Orb expands (handled by CSS)
  await new Promise((r) => setTimeout(r, 800))

  // Phase 2: Type hostname
  phase.value = 'hostname'
  await typeHostname(hostname.value)
  await new Promise((r) => setTimeout(r, 200))

  // Phase 3: Tagline
  phase.value = 'tagline'
  showTagline.value = true
  await new Promise((r) => setTimeout(r, 400))

  // Phase 4: Button
  phase.value = 'button'
  showButton.value = true
}

onMounted(async () => {
  await fetchHostname()
  runAnimation()
})

async function startSetup() {
  await setupStore.advanceTo('system-check')
  router.push('/setup/system-check')
}
</script>

<template>
  <div class="flex flex-col items-center justify-center min-h-[60vh] text-center">
    <!-- Violet orb -->
    <div class="relative mb-10">
      <div
        class="w-24 h-24 rounded-full bg-primary/20 flex items-center justify-center transition-all duration-700"
        :class="phase !== 'orb' ? 'scale-100 opacity-100' : 'scale-75 opacity-80'"
      >
        <!-- Inner glow ring -->
        <div class="absolute inset-0 rounded-full bg-primary/10 animate-glow-pulse" />

        <!-- Logo -->
        <div class="relative w-12 h-12 rounded-xl bg-primary flex items-center justify-center shadow-glow">
          <svg width="24" height="24" viewBox="0 0 14 14" fill="none">
            <path
              d="M2 4L7 1L12 4V10L7 13L2 10V4Z"
              stroke="white"
              stroke-width="1.5"
              stroke-linejoin="round"
            />
            <circle cx="7" cy="7" r="1.5" fill="white" />
          </svg>
        </div>

        <!-- Orbit rings -->
        <div class="absolute -inset-4 rounded-full border border-primary/10 animate-spin-slow" style="animation-duration: 8s" />
        <div class="absolute -inset-8 rounded-full border border-primary/5" style="animation: spin-slow 14s linear infinite reverse" />
      </div>
    </div>

    <!-- Hostname (typed) -->
    <Transition name="fade">
      <div v-if="phase !== 'orb'" class="mb-3">
        <h1
          class="text-5xl font-bold text-text tracking-tight"
          style="letter-spacing: -0.04em; line-height: 1.1"
        >
          {{ typedHostname }}<span
            v-if="phase === 'hostname'"
            class="inline-block w-0.5 h-10 bg-primary-glow ml-0.5 animate-pulse"
          />
        </h1>
      </div>
    </Transition>

    <!-- Tagline -->
    <Transition name="fade">
      <p
        v-if="showTagline"
        class="text-base text-muted mb-10 max-w-sm leading-relaxed"
      >
        Your Pi. Your AI. Your code.<br />
        Let's get everything set up in about 5 minutes.
      </p>
    </Transition>

    <!-- CTA Button -->
    <Transition name="slide-up">
      <button
        v-if="showButton"
        @click="startSetup"
        class="group flex items-center gap-3 px-8 py-4 rounded-xl bg-primary hover:bg-primary-dim text-white font-semibold text-base transition-all duration-200 shadow-glow hover:shadow-float"
      >
        Start Setup
        <svg
          class="w-5 h-5 transition-transform duration-200 group-hover:translate-x-1"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </Transition>

    <!-- Version tag -->
    <Transition name="fade">
      <p v-if="showButton" class="mt-6 text-xs text-muted">
        VibeCodePC — Self-hosted AI coding station
      </p>
    </Transition>
  </div>
</template>
