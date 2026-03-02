<script setup lang="ts">
import { ref, computed } from 'vue'
import { XMarkIcon } from '@heroicons/vue/24/outline'
import { useProjectsStore } from '@/stores/projects'
import { useGithubStore } from '@/stores/github'

const emit = defineEmits<{
  close: []
  created: [projectId: string]
}>()

const projectsStore = useProjectsStore()
const githubStore = useGithubStore()

const activeTab = ref<'local' | 'github'>('local')

// Local tab
const localName = ref('')
const localPath = ref('')
const localLang = ref('')
const localError = ref('')
const localLoading = ref(false)

// GitHub tab
const githubSearch = ref('')
const githubPath = ref('')
const githubName = ref('')
const githubError = ref('')
const importLogs = ref<string[]>([])
const importing = ref(false)

const languages = [
  'TypeScript', 'JavaScript', 'Python', 'Go', 'Rust',
  'Java', 'C++', 'C', 'Ruby', 'PHP', 'Other',
]

const filteredRepos = computed(() => {
  if (!githubSearch.value) return githubStore.repos.slice(0, 10)
  const q = githubSearch.value.toLowerCase()
  return githubStore.repos
    .filter((r) => r.fullName.toLowerCase().includes(q))
    .slice(0, 10)
})

async function createLocal() {
  localError.value = ''
  if (!localName.value.trim() || !localPath.value.trim()) {
    localError.value = 'Name and path are required'
    return
  }
  localLoading.value = true
  const { data, error } = await projectsStore.create({
    name: localName.value.trim(),
    path: localPath.value.trim(),
    language: localLang.value || undefined,
  })
  localLoading.value = false
  if (error) {
    localError.value = error
  } else if (data) {
    emit('created', data.id)
    emit('close')
  }
}

function startImport(repoUrl: string, repoName: string) {
  if (!githubPath.value.trim()) {
    githubError.value = 'Destination path is required'
    return
  }
  githubError.value = ''
  importing.value = true
  importLogs.value = []

  const name = githubName.value.trim() || repoName
  const params = new URLSearchParams({
    repoUrl,
    name,
    path: githubPath.value.trim(),
  })
  const es = new EventSource(`/api/github/import?${params}`)

  es.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.log) importLogs.value.push(data.log)
      if (data.projectId) {
        es.close()
        importing.value = false
        emit('created', data.projectId)
        emit('close')
      }
    } catch {
      importLogs.value.push(event.data)
    }
  }

  es.onerror = () => {
    es.close()
    importing.value = false
    githubError.value = 'Import failed'
  }
}

// Load repos if github tab selected
async function onTabGithub() {
  activeTab.value = 'github'
  if (!githubStore.repos.length) {
    await githubStore.fetchRepos()
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
      @click.self="emit('close')"
    >
      <div
        class="w-full max-w-lg bg-surface-800 border border-surface-600 rounded-xl shadow-float"
      >
        <!-- Header -->
        <div class="flex items-center justify-between p-5 border-b border-surface-600">
          <h2 class="text-base font-semibold text-text">New Project</h2>
          <button
            @click="emit('close')"
            class="p-1.5 rounded text-muted hover:text-text hover:bg-surface-700 transition-all duration-120"
          >
            <XMarkIcon class="w-5 h-5" />
          </button>
        </div>

        <!-- Tabs -->
        <div class="flex border-b border-surface-600">
          <button
            @click="activeTab = 'local'"
            :class="[
              'flex-1 py-3 text-sm font-medium transition-colors duration-120 border-b-2',
              activeTab === 'local'
                ? 'text-primary-glow border-primary'
                : 'text-muted border-transparent hover:text-text',
            ]"
          >
            Local Path
          </button>
          <button
            @click="onTabGithub"
            :class="[
              'flex-1 py-3 text-sm font-medium transition-colors duration-120 border-b-2',
              activeTab === 'github'
                ? 'text-primary-glow border-primary'
                : 'text-muted border-transparent hover:text-text',
            ]"
          >
            From GitHub
          </button>
        </div>

        <!-- Local tab -->
        <div v-if="activeTab === 'local'" class="p-5 space-y-4">
          <div>
            <label class="block text-xs font-medium text-muted uppercase tracking-wider mb-1.5">Project Name</label>
            <input
              v-model="localName"
              placeholder="my-project"
              class="w-full px-3 py-2.5 rounded-md bg-surface-700 border border-surface-600 focus:border-primary text-sm text-text outline-none transition-colors duration-120"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-muted uppercase tracking-wider mb-1.5">Path</label>
            <input
              v-model="localPath"
              placeholder="/home/pi/projects/my-project"
              class="w-full px-3 py-2.5 rounded-md bg-surface-700 border border-surface-600 focus:border-primary text-sm text-text font-mono outline-none transition-colors duration-120"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-muted uppercase tracking-wider mb-1.5">Language (optional)</label>
            <select
              v-model="localLang"
              class="w-full px-3 py-2.5 rounded-md bg-surface-700 border border-surface-600 focus:border-primary text-sm text-text outline-none transition-colors duration-120"
            >
              <option value="">Auto-detect</option>
              <option v-for="lang in languages" :key="lang" :value="lang.toLowerCase()">
                {{ lang }}
              </option>
            </select>
          </div>

          <p v-if="localError" class="text-sm text-danger">{{ localError }}</p>

          <button
            @click="createLocal"
            :disabled="localLoading"
            class="w-full py-2.5 rounded-md bg-primary text-white text-sm font-medium hover:bg-primary-dim disabled:opacity-50 transition-all duration-120"
          >
            {{ localLoading ? 'Creating...' : 'Create Project' }}
          </button>
        </div>

        <!-- GitHub tab -->
        <div v-else class="p-5 space-y-4">
          <div v-if="!githubStore.connected" class="py-6 text-center">
            <p class="text-sm text-muted">Connect GitHub to import repositories</p>
            <RouterLink
              to="/app/settings"
              @click="emit('close')"
              class="mt-3 inline-block text-sm text-primary-glow hover:underline"
            >
              Go to Settings →
            </RouterLink>
          </div>

          <template v-else>
            <input
              v-model="githubSearch"
              placeholder="Search repositories..."
              class="w-full px-3 py-2.5 rounded-md bg-surface-700 border border-surface-600 focus:border-primary text-sm text-text outline-none transition-colors duration-120"
            />

            <div v-if="!importing" class="max-h-48 overflow-y-auto space-y-1.5">
              <button
                v-for="repo in filteredRepos"
                :key="repo.fullName"
                @click="startImport(repo.cloneUrl, repo.fullName.split('/')[1])"
                class="w-full text-left px-3 py-2.5 rounded-md bg-surface-700 hover:bg-surface-600 border border-surface-600 transition-all duration-120"
              >
                <div class="flex items-center justify-between">
                  <span class="text-sm font-medium text-text">{{ repo.fullName }}</span>
                  <span v-if="repo.language" class="text-xs text-muted">{{ repo.language }}</span>
                </div>
                <p v-if="repo.description" class="text-xs text-muted mt-0.5 truncate">
                  {{ repo.description }}
                </p>
              </button>
            </div>

            <div>
              <label class="block text-xs font-medium text-muted uppercase tracking-wider mb-1.5">Destination Path</label>
              <input
                v-model="githubPath"
                placeholder="/home/pi/projects"
                class="w-full px-3 py-2.5 rounded-md bg-surface-700 border border-surface-600 focus:border-primary text-sm text-text font-mono outline-none transition-colors duration-120"
              />
            </div>

            <!-- Import progress -->
            <div
              v-if="importing || importLogs.length"
              class="bg-surface-900 rounded-md p-3 max-h-32 overflow-y-auto font-mono text-xs text-muted space-y-0.5"
            >
              <div v-for="(line, i) in importLogs" :key="i" class="leading-5">{{ line }}</div>
              <div v-if="importing" class="flex items-center gap-2 text-primary-glow">
                <span class="w-3 h-3 rounded-full border border-primary-glow border-t-transparent animate-spin" />
                Importing...
              </div>
            </div>

            <p v-if="githubError" class="text-sm text-danger">{{ githubError }}</p>
          </template>
        </div>
      </div>
    </div>
  </Teleport>
</template>
