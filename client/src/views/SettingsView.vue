<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import { useGithubStore } from '@/stores/github'
import KeyInput from '@/components/ui/KeyInput.vue'
import StatusDot from '@/components/ui/StatusDot.vue'

const settingsStore = useSettingsStore()
const githubStore = useGithubStore()

const activeTab = ref<'providers' | 'system'>('providers')

interface ProviderEdit {
  key: string
  state: 'idle' | 'testing' | 'valid' | 'invalid'
  error: string
  saving: boolean
}

const providerEdits = ref<Record<string, ProviderEdit>>({})

const providerDefs = [
  { id: 'anthropic', name: 'Anthropic (Claude)', color: '#d4a27a' },
  { id: 'openai', name: 'OpenAI (GPT-4)', color: '#74aa9c' },
  { id: 'google', name: 'Google AI (Gemini)', color: '#4285f4' },
  { id: 'ollama', name: 'Ollama (Local)', color: '#888' },
]

const updateLogs = ref<string[]>([])
const updating = ref(false)
const updateError = ref('')
const logsRef = ref<HTMLElement | null>(null)

function getProviderEdit(id: string): ProviderEdit {
  if (!providerEdits.value[id]) {
    const existing = settingsStore.providers.find((p) => p.name === id)
    providerEdits.value[id] = {
      key: existing?.key ?? '',
      state: existing?.valid ? 'valid' : 'idle',
      error: '',
      saving: false,
    }
  }
  return providerEdits.value[id]
}

async function saveProvider(id: string) {
  const edit = getProviderEdit(id)
  if (!edit.key) return

  edit.saving = true
  edit.state = 'testing'
  edit.error = ''

  const { error } = await settingsStore.saveProvider(id, edit.key)
  edit.saving = false

  if (error) {
    edit.state = 'invalid'
    edit.error = error
  } else {
    edit.state = 'valid'
  }
}

function startUpdate() {
  updating.value = true
  updateLogs.value = []
  updateError.value = ''

  const es = settingsStore.update()

  es.onmessage = (event) => {
    updateLogs.value.push(event.data)
    setTimeout(() => {
      if (logsRef.value) logsRef.value.scrollTop = logsRef.value.scrollHeight
    }, 10)
  }

  es.addEventListener('done', () => {
    es.close()
    updating.value = false
  })

  es.onerror = () => {
    es.close()
    updating.value = false
    updateError.value = 'Update failed'
  }
}

onMounted(async () => {
  await Promise.all([
    settingsStore.fetchProviders(),
    settingsStore.fetchTunnel(),
    settingsStore.fetchSystem(),
    githubStore.fetchStatus(),
  ])

  // Init edits from loaded providers
  for (const p of settingsStore.providers) {
    providerEdits.value[p.name] = {
      key: p.key,
      state: p.valid ? 'valid' : 'idle',
      error: '',
      saving: false,
    }
  }
})
</script>

<template>
  <div class="max-w-3xl mx-auto p-6 lg:p-8">
    <!-- Header -->
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-text">Settings</h1>
      <p class="text-sm text-muted mt-1">Manage your VibeCodePC configuration</p>
    </div>

    <!-- Tabs -->
    <div class="flex border-b border-surface-600 mb-6">
      <button
        v-for="tab in [{ id: 'providers', label: 'AI Providers' }, { id: 'system', label: 'System' }]"
        :key="tab.id"
        @click="activeTab = tab.id as 'providers' | 'system'"
        :class="[
          'px-4 py-3 text-sm font-medium border-b-2 transition-colors duration-120',
          activeTab === tab.id
            ? 'text-primary-glow border-primary'
            : 'text-muted border-transparent hover:text-text',
        ]"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Providers tab -->
    <div v-if="activeTab === 'providers'" class="space-y-4">
      <div
        v-for="prov in providerDefs"
        :key="prov.id"
        class="bg-surface-800 border border-surface-600 rounded-lg p-5 transition-all duration-200"
        :class="getProviderEdit(prov.id).state === 'valid' ? 'border-success/30' : ''"
      >
        <div class="flex items-center gap-3 mb-4">
          <div
            class="w-8 h-8 rounded-md flex items-center justify-center text-sm font-bold"
            :style="`background-color: ${prov.color}20; border: 1px solid ${prov.color}30; color: ${prov.color}`"
          >
            {{ prov.name.charAt(0) }}
          </div>
          <div class="flex-1">
            <div class="flex items-center gap-2">
              <span class="text-sm font-semibold text-text">{{ prov.name }}</span>
              <span
                v-if="getProviderEdit(prov.id).state === 'valid'"
                class="text-[10px] px-1.5 py-0.5 rounded bg-success/15 text-success font-medium"
              >
                Active
              </span>
            </div>
          </div>
        </div>

        <div class="space-y-3">
          <KeyInput
            :model-value="getProviderEdit(prov.id).key"
            :state="getProviderEdit(prov.id).state"
            :placeholder="prov.id === 'ollama' ? 'No key needed' : 'Enter API key...'"
            @update:model-value="(val) => { getProviderEdit(prov.id).key = val; getProviderEdit(prov.id).state = 'idle' }"
          />

          <p v-if="getProviderEdit(prov.id).error" class="text-xs text-danger">
            {{ getProviderEdit(prov.id).error }}
          </p>

          <div class="flex items-center gap-2">
            <button
              @click="saveProvider(prov.id)"
              :disabled="getProviderEdit(prov.id).saving || !getProviderEdit(prov.id).key"
              class="px-3 py-1.5 rounded text-xs font-medium bg-primary text-white hover:bg-primary-dim disabled:opacity-50 transition-all duration-120"
            >
              {{ getProviderEdit(prov.id).saving ? 'Saving...' : 'Save & Test' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- System tab -->
    <div v-else-if="activeTab === 'system'" class="space-y-5">
      <!-- System info -->
      <div class="bg-surface-800 border border-surface-600 rounded-lg p-5">
        <h3 class="text-sm font-semibold text-text mb-4">System Information</h3>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1">Hostname</p>
            <p class="text-sm font-mono text-text">{{ settingsStore.system.hostname || '—' }}</p>
          </div>
          <div>
            <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1">IP Address</p>
            <p class="text-sm font-mono text-text">{{ settingsStore.system.ip || '—' }}</p>
          </div>
          <div>
            <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1">Go Version</p>
            <p class="text-sm font-mono text-text">{{ settingsStore.system.goVersion || '—' }}</p>
          </div>
          <div>
            <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1">App Version</p>
            <p class="text-sm font-mono text-text">{{ settingsStore.system.appVersion || '—' }}</p>
          </div>
          <div>
            <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1">Docker</p>
            <p class="text-sm font-mono text-text">{{ settingsStore.system.dockerVersion || '—' }}</p>
          </div>
          <div>
            <p class="text-[10px] font-medium uppercase tracking-wider text-muted mb-1">Local URL</p>
            <a
              :href="settingsStore.system.localUrl"
              target="_blank"
              class="text-sm font-mono text-primary-glow hover:underline truncate block"
            >
              {{ settingsStore.system.localUrl || '—' }}
            </a>
          </div>
        </div>
      </div>

      <!-- Tunnel info -->
      <div class="bg-surface-800 border border-surface-600 rounded-lg p-5">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-semibold text-text">Cloudflare Tunnel</h3>
          <div class="flex items-center gap-1.5">
            <StatusDot :status="settingsStore.tunnel.connected ? 'running' : 'stopped'" size="xs" />
            <span class="text-xs text-muted capitalize">{{ settingsStore.tunnel.mode }}</span>
          </div>
        </div>
        <div class="space-y-2">
          <div v-if="settingsStore.tunnel.tunnelUrl">
            <p class="text-xs text-muted mb-1">Remote URL</p>
            <p class="text-sm font-mono text-success truncate">{{ settingsStore.tunnel.tunnelUrl }}</p>
          </div>
          <button
            @click="settingsStore.restartTunnel()"
            class="mt-2 px-3 py-1.5 rounded text-xs text-muted hover:text-text bg-surface-700 hover:bg-surface-600 border border-surface-600 transition-all duration-120"
          >
            Restart Tunnel
          </button>
        </div>
      </div>

      <!-- GitHub -->
      <div class="bg-surface-800 border border-surface-600 rounded-lg p-5">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-semibold text-text">GitHub</h3>
          <div class="flex items-center gap-1.5">
            <StatusDot :status="githubStore.connected ? 'running' : 'stopped'" size="xs" />
            <span class="text-xs text-muted">{{ githubStore.connected ? 'Connected' : 'Not connected' }}</span>
          </div>
        </div>
        <div v-if="githubStore.user" class="flex items-center gap-3 mb-3">
          <img :src="githubStore.user.avatarUrl" class="w-8 h-8 rounded-full" />
          <div>
            <p class="text-sm text-text font-medium">{{ githubStore.user.login }}</p>
            <p class="text-xs text-muted">{{ githubStore.user.publicRepos }} public repos</p>
          </div>
        </div>
        <button
          v-if="githubStore.connected"
          @click="githubStore.disconnect()"
          class="text-xs text-danger/70 hover:text-danger transition-colors duration-120"
        >
          Disconnect GitHub
        </button>
      </div>

      <!-- System update -->
      <div class="bg-surface-800 border border-surface-600 rounded-lg p-5">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-semibold text-text">App Update</h3>
        </div>
        <p class="text-xs text-muted mb-3">
          Updates VibeCodePC to the latest version. Your data is preserved.
        </p>

        <button
          @click="startUpdate"
          :disabled="updating"
          class="px-4 py-2 rounded-md text-sm font-medium bg-primary text-white hover:bg-primary-dim disabled:opacity-50 transition-all duration-120"
        >
          {{ updating ? 'Updating...' : 'Check for Updates' }}
        </button>

        <div
          v-if="updateLogs.length || updating"
          ref="logsRef"
          class="mt-3 bg-surface-900 rounded-md p-3 max-h-32 overflow-y-auto font-mono text-xs text-muted"
        >
          <div v-for="(log, i) in updateLogs" :key="i" class="leading-5">{{ log }}</div>
        </div>
        <p v-if="updateError" class="mt-2 text-xs text-danger">{{ updateError }}</p>
      </div>
    </div>
  </div>
</template>
