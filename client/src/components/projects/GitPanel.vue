<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ChevronDownIcon, ArrowUpIcon, ArrowDownIcon } from '@heroicons/vue/24/outline'
import { useProjectsStore } from '@/stores/projects'
import type { Branch } from '@/types'
import { useFetch } from '@/composables/useFetch'

const props = defineProps<{
  projectId: string
}>()

const projectsStore = useProjectsStore()
const { get } = useFetch()

const commitMessage = ref('')
const showBranchDropdown = ref(false)
const branches = ref<Branch[]>([])
const showDiff = ref(false)
const diff = ref('')
const selectedFile = ref<string | null>(null)
const isCommitting = ref(false)
const isPushing = ref(false)
const commitError = ref('')

const status = computed(() => projectsStore.gitStatus[props.projectId])

const allChangedFiles = computed(() => {
  if (!status.value) return []
  return [
    ...status.value.staged.map((f) => ({ file: f, state: 'M' as const, staged: true })),
    ...status.value.unstaged.map((f) => ({ file: f, state: 'M' as const, staged: false })),
    ...status.value.untracked.map((f) => ({ file: f, state: '?' as const, staged: false })),
  ]
})

const fileBadgeColor = (state: string, staged: boolean) => {
  if (staged) return 'text-success bg-success/10'
  if (state === '?') return 'text-muted bg-surface-600'
  return 'text-warning bg-warning/10'
}

async function loadBranches() {
  const { data } = await get<{ branches: Branch[] }>(
    `/api/projects/${props.projectId}/git/branches`
  )
  if (data) branches.value = data.branches
}

async function loadDiff() {
  const { data } = await get<{ diff: string }>(
    `/api/projects/${props.projectId}/git/diff`
  )
  if (data) diff.value = data.diff
}

async function checkout(branch: string) {
  await projectsStore.gitCheckout(props.projectId, branch)
  showBranchDropdown.value = false
}

async function commitAll() {
  if (!commitMessage.value.trim()) return
  isCommitting.value = true
  commitError.value = ''
  const { error } = await projectsStore.gitCommit(props.projectId, commitMessage.value.trim())
  isCommitting.value = false
  if (error) {
    commitError.value = error
  } else {
    commitMessage.value = ''
  }
}

async function push() {
  isPushing.value = true
  await projectsStore.gitPush(props.projectId)
  isPushing.value = false
}

async function pull() {
  await projectsStore.gitPull(props.projectId)
}

onMounted(async () => {
  await projectsStore.fetchGitStatus(props.projectId)
  await loadBranches()
})

watch(showDiff, async (val) => {
  if (val && !diff.value) {
    await loadDiff()
  }
})
</script>

<template>
  <div class="flex flex-col h-full bg-surface-800 border-r border-surface-600 w-60 flex-shrink-0">
    <!-- Header -->
    <div class="px-4 py-3 border-b border-surface-600">
      <h3 class="text-xs font-semibold text-muted uppercase tracking-wider">Git</h3>
    </div>

    <!-- Branch selector -->
    <div class="px-3 py-2 border-b border-surface-600 relative">
      <button
        @click="showBranchDropdown = !showBranchDropdown"
        class="w-full flex items-center gap-2 px-2.5 py-2 rounded-md bg-surface-700 border border-surface-600 hover:border-surface-500 text-xs font-mono text-text transition-all duration-120"
      >
        <svg class="w-3 h-3 text-muted flex-shrink-0" fill="currentColor" viewBox="0 0 16 16">
          <path d="M5 3.25a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0zm0 2.122a2.25 2.25 0 1 0-1.5 0v.878A2.25 2.25 0 0 0 5.75 8.5h1.5v2.128a2.251 2.251 0 1 0 1.5 0V8.5h1.5a2.25 2.25 0 0 0 2.25-2.25v-.878a2.25 2.25 0 1 0-1.5 0v.878a.75.75 0 0 1-.75.75h-4.5A.75.75 0 0 1 5 6.25v-.878z"/>
        </svg>
        <span class="flex-1 text-left truncate">
          {{ status?.branch ?? 'loading...' }}
        </span>
        <div v-if="status && (status.ahead > 0 || status.behind > 0)" class="flex items-center gap-1">
          <span v-if="status.ahead > 0" class="text-success text-[10px]">↑{{ status.ahead }}</span>
          <span v-if="status.behind > 0" class="text-warning text-[10px]">↓{{ status.behind }}</span>
        </div>
        <ChevronDownIcon class="w-3 h-3 text-muted flex-shrink-0" />
      </button>

      <!-- Dropdown -->
      <Transition name="slide-up">
        <div
          v-if="showBranchDropdown"
          class="absolute left-3 right-3 top-full mt-1 bg-surface-700 border border-surface-600 rounded-md shadow-float z-10 max-h-40 overflow-y-auto"
        >
          <button
            v-for="branch in branches"
            :key="branch.name"
            @click="checkout(branch.name)"
            :class="[
              'w-full text-left px-3 py-2 text-xs font-mono hover:bg-surface-600 transition-colors duration-120',
              branch.current ? 'text-primary-glow' : 'text-text',
            ]"
          >
            <div class="flex items-center gap-2">
              <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="branch.current ? 'bg-primary' : 'bg-transparent'" />
              {{ branch.name }}
              <span v-if="branch.remote" class="ml-auto text-muted text-[10px]">remote</span>
            </div>
          </button>
        </div>
      </Transition>
    </div>

    <!-- File list -->
    <div class="flex-1 overflow-y-auto">
      <div v-if="!status || allChangedFiles.length === 0" class="px-4 py-4 text-xs text-muted">
        No changes
      </div>
      <div v-else class="py-1">
        <div
          v-for="file in allChangedFiles"
          :key="file.file"
          @click="selectedFile = selectedFile === file.file ? null : file.file; showDiff = true"
          :class="[
            'flex items-center gap-2 px-3 py-1.5 cursor-pointer hover:bg-surface-700 transition-colors duration-120',
            selectedFile === file.file ? 'bg-surface-700' : '',
          ]"
        >
          <span
            :class="[
              'text-[10px] font-mono font-bold w-4 text-center flex-shrink-0',
              fileBadgeColor(file.state, file.staged),
            ]"
          >
            {{ file.state }}
          </span>
          <span class="text-xs text-text truncate font-mono">
            {{ file.file.split('/').pop() }}
          </span>
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div class="p-3 border-t border-surface-600 space-y-2">
      <!-- Quick actions row -->
      <div class="flex gap-2">
        <button
          @click="pull"
          class="flex-1 flex items-center justify-center gap-1 py-1.5 rounded text-xs text-muted hover:text-text hover:bg-surface-700 transition-all duration-120"
        >
          <ArrowDownIcon class="w-3 h-3" />
          Pull
        </button>
        <button
          @click="push"
          :disabled="isPushing"
          class="flex-1 flex items-center justify-center gap-1 py-1.5 rounded text-xs text-muted hover:text-text hover:bg-surface-700 transition-all duration-120 disabled:opacity-50"
        >
          <ArrowUpIcon class="w-3 h-3" />
          {{ isPushing ? '...' : 'Push' }}
        </button>
      </div>

      <!-- Commit message -->
      <textarea
        v-model="commitMessage"
        placeholder="Commit message..."
        rows="2"
        class="w-full px-2.5 py-2 rounded-md bg-surface-700 border border-surface-600 focus:border-primary text-xs text-text placeholder-muted outline-none resize-none transition-colors duration-120"
      />

      <p v-if="commitError" class="text-xs text-danger">{{ commitError }}</p>

      <button
        @click="commitAll"
        :disabled="isCommitting || !commitMessage.trim()"
        class="w-full py-2 rounded-md bg-primary text-white text-xs font-medium hover:bg-primary-dim disabled:opacity-50 transition-all duration-120"
      >
        {{ isCommitting ? 'Committing...' : 'Commit' }}
      </button>
    </div>
  </div>
</template>
