<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSetupStore } from '@/stores/setup'
import { useFetch } from '@/composables/useFetch'
import KeyInput from '@/components/ui/KeyInput.vue'

const router = useRouter()
const setupStore = useSetupStore()
const { post } = useFetch()

interface ProviderState {
  key: string
  state: 'idle' | 'testing' | 'valid' | 'invalid'
  error: string
}

const providers = reactive<Record<string, ProviderState>>({
  anthropic: { key: '', state: 'idle', error: '' },
  openai: { key: '', state: 'idle', error: '' },
  google: { key: '', state: 'idle', error: '' },
  ollama: { key: 'local', state: 'idle', error: '' },
})

const providerInfo = [
  {
    id: 'anthropic',
    name: 'Anthropic',
    description: 'Claude — best for complex reasoning and code',
    keyPlaceholder: 'sk-ant-...',
    color: '#d4a27a',
    icon: 'A',
    keyPrefix: 'sk-ant-',
  },
  {
    id: 'openai',
    name: 'OpenAI',
    description: 'GPT-4 and o-series models',
    keyPlaceholder: 'sk-...',
    color: '#74aa9c',
    icon: 'O',
    keyPrefix: 'sk-',
  },
  {
    id: 'google',
    name: 'Google AI',
    description: 'Gemini models',
    keyPlaceholder: 'AIza...',
    color: '#4285f4',
    icon: 'G',
    keyPrefix: '',
  },
  {
    id: 'ollama',
    name: 'Ollama',
    description: 'Local models running on your Pi',
    keyPlaceholder: 'No key needed',
    color: '#888888',
    icon: '⚙',
    keyPrefix: '',
    local: true,
  },
]

const anyValid = computed(() =>
  Object.values(providers).some((p) => p.state === 'valid')
)

async function detectProviders() {
  const { data } = await post<{ detected: string[] }>('/api/setup/providers/detect')
  if (data?.detected) {
    for (const name of data.detected) {
      if (providers[name]) {
        providers[name].state = 'valid'
      }
    }
  }
}

async function testProvider(id: string) {
  const p = providers[id]
  if (!p.key && id !== 'ollama') return

  p.state = 'testing'
  p.error = ''

  const { data, error } = await post<{ valid: boolean; models?: string[] }>(
    '/api/setup/providers/test',
    { provider: id, key: p.key }
  )

  if (error || !data?.valid) {
    p.state = 'invalid'
    p.error = error ?? 'Invalid key'
  } else {
    p.state = 'valid'
  }
}

function handleKeyChange(id: string, val: string) {
  providers[id].key = val
  providers[id].state = 'idle'
  providers[id].error = ''
}

async function saveAndContinue() {
  // Save valid providers
  const validProviders = Object.entries(providers)
    .filter(([_, v]) => v.state === 'valid' && v.key)
    .map(([name, v]) => ({ provider: name, key: v.key }))

  for (const p of validProviders) {
    await post('/api/settings/providers', p)
  }

  await setupStore.advanceTo('opencode')
  router.push('/setup/opencode')
}

onMounted(async () => {
  await detectProviders()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h2 class="text-2xl font-bold text-text mb-2">AI Providers</h2>
      <p class="text-muted text-sm leading-relaxed">
        Add at least one AI provider key to power your coding assistant.
        Keys are stored encrypted on your device — never sent to our servers.
      </p>
    </div>

    <!-- Provider cards -->
    <div class="space-y-3">
      <div
        v-for="prov in providerInfo"
        :key="prov.id"
        class="bg-surface-800 border border-surface-600 rounded-lg p-4 transition-all duration-200"
        :class="providers[prov.id].state === 'valid' ? 'border-success/30' : ''"
      >
        <!-- Provider header -->
        <div class="flex items-center gap-3 mb-3">
          <div
            class="w-8 h-8 rounded-md flex items-center justify-center text-sm font-bold text-white"
            :style="`background-color: ${prov.color}30; border: 1px solid ${prov.color}40`"
          >
            <span :style="`color: ${prov.color}`">{{ prov.icon }}</span>
          </div>
          <div class="flex-1">
            <div class="flex items-center gap-2">
              <span class="text-sm font-semibold text-text">{{ prov.name }}</span>
              <!-- Valid badge -->
              <span
                v-if="providers[prov.id].state === 'valid'"
                class="text-[10px] font-medium px-1.5 py-0.5 rounded bg-success/15 text-success"
              >
                Connected
              </span>
            </div>
            <p class="text-xs text-muted">{{ prov.description }}</p>
          </div>
        </div>

        <!-- Key input -->
        <div v-if="!prov.local" class="space-y-2">
          <KeyInput
            :model-value="providers[prov.id].key"
            :placeholder="prov.keyPlaceholder"
            :state="providers[prov.id].state"
            @update:model-value="handleKeyChange(prov.id, $event)"
          />
          <p v-if="providers[prov.id].error" class="text-xs text-danger">
            {{ providers[prov.id].error }}
          </p>
          <button
            v-if="providers[prov.id].key && providers[prov.id].state !== 'valid'"
            @click="testProvider(prov.id)"
            :disabled="providers[prov.id].state === 'testing'"
            class="text-xs px-3 py-1.5 rounded bg-surface-700 text-muted hover:text-text hover:bg-surface-600 border border-surface-600 transition-all duration-120 disabled:opacity-50"
          >
            {{ providers[prov.id].state === 'testing' ? 'Testing...' : 'Test connection' }}
          </button>
        </div>

        <!-- Ollama: local -->
        <div v-else>
          <button
            @click="testProvider(prov.id)"
            :disabled="providers[prov.id].state === 'testing'"
            class="text-xs px-3 py-1.5 rounded bg-surface-700 text-muted hover:text-text hover:bg-surface-600 border border-surface-600 transition-all duration-120 disabled:opacity-50"
          >
            {{ providers[prov.id].state === 'testing' ? 'Checking...' : 'Check local instance' }}
          </button>
          <p v-if="providers[prov.id].state === 'invalid'" class="text-xs text-warning mt-1">
            Ollama not found at http://localhost:11434
          </p>
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex items-center justify-between pt-2">
      <p v-if="!anyValid" class="text-xs text-muted">
        Add at least one provider to continue
      </p>
      <div v-else />
      <button
        @click="saveAndContinue"
        :disabled="!anyValid"
        class="px-5 py-2.5 rounded-md text-sm font-medium bg-primary text-white hover:bg-primary-dim disabled:opacity-50 transition-all duration-120 ml-auto"
      >
        Continue →
      </button>
    </div>
  </div>
</template>
