<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useSetupStore } from '@/stores/setup'
import { useGithubStore } from '@/stores/github'

const router = useRouter()
const setupStore = useSetupStore()
const githubStore = useGithubStore()

const connectError = ref('')

onMounted(async () => {
  await githubStore.fetchStatus()
})

function connectGitHub() {
  // Open GitHub OAuth in a popup
  const w = 600
  const h = 700
  const left = window.screenX + (window.outerWidth - w) / 2
  const top = window.screenY + (window.outerHeight - h) / 2
  const popup = window.open(
    '/auth/github/start',
    'github-auth',
    `width=${w},height=${h},left=${left},top=${top}`
  )

  // Poll for popup close or message
  const pollTimer = setInterval(async () => {
    if (!popup || popup.closed) {
      clearInterval(pollTimer)
      await githubStore.fetchStatus()
    }
  }, 500)
}

async function advance() {
  await setupStore.advanceTo('providers')
  router.push('/setup/providers')
}

async function skip() {
  await setupStore.advanceTo('providers')
  router.push('/setup/providers')
}

async function disconnect() {
  await githubStore.disconnect()
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h2 class="text-2xl font-bold text-text mb-2">GitHub</h2>
      <p class="text-muted text-sm leading-relaxed">
        Connect your GitHub account to import repositories and create pull requests directly from the app.
      </p>
    </div>

    <!-- Not connected state -->
    <div
      v-if="!githubStore.connected"
      class="bg-surface-800 border border-surface-600 rounded-xl p-8 flex flex-col items-center gap-5 text-center"
    >
      <div class="w-16 h-16 rounded-full bg-surface-700 border border-surface-600 flex items-center justify-center">
        <svg class="w-8 h-8 text-muted" fill="currentColor" viewBox="0 0 24 24">
          <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
        </svg>
      </div>
      <div>
        <h3 class="text-base font-semibold text-text mb-1">Connect GitHub</h3>
        <p class="text-sm text-muted">Requires read:user and repo scopes</p>
      </div>
      <button
        @click="connectGitHub"
        class="flex items-center gap-2.5 px-5 py-3 rounded-lg bg-white text-gray-900 font-semibold text-sm hover:bg-gray-100 transition-all duration-200"
      >
        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
          <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
        </svg>
        Continue with GitHub
      </button>
      <p v-if="connectError" class="text-sm text-danger">{{ connectError }}</p>
    </div>

    <!-- Connected state: identity card -->
    <div
      v-else
      class="bg-surface-800 border border-success/30 rounded-xl p-5"
    >
      <div class="flex items-center gap-4">
        <img
          v-if="githubStore.user?.avatarUrl"
          :src="githubStore.user.avatarUrl"
          :alt="githubStore.user.login"
          class="w-14 h-14 rounded-full border-2 border-success/30"
        />
        <div v-else class="w-14 h-14 rounded-full bg-surface-700 border-2 border-surface-600" />

        <div class="flex-1">
          <div class="flex items-center gap-2 mb-1">
            <div class="w-4 h-4 text-success">
              <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <span class="text-sm font-semibold text-text">{{ githubStore.user?.login }}</span>
          </div>
          <p class="text-xs text-muted">
            {{ githubStore.user?.publicRepos ?? 0 }} public repositories
          </p>
        </div>

        <button
          @click="disconnect"
          class="text-xs text-muted hover:text-danger transition-colors duration-120"
        >
          Disconnect
        </button>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex items-center justify-between pt-2">
      <button
        v-if="!githubStore.connected"
        @click="skip"
        class="text-sm text-muted hover:text-text transition-colors duration-120"
      >
        Skip for now →
      </button>
      <div v-else />
      <button
        @click="advance"
        class="px-5 py-2.5 rounded-md text-sm font-medium bg-primary text-white hover:bg-primary-dim transition-all duration-120 ml-auto"
      >
        Continue →
      </button>
    </div>
  </div>
</template>
