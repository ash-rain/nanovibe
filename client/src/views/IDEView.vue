<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useProjectsStore } from '@/stores/projects'
import AppTopbar from '@/components/layout/AppTopbar.vue'
import GitPanel from '@/components/projects/GitPanel.vue'
import TerminalPane from '@/components/terminal/TerminalPane.vue'

const route = useRoute()
const projectsStore = useProjectsStore()

const showGitPanel = ref(true)
const sessionConnected = ref(false)

const projectId = computed(
  () => (route.params.projectId as string) || projectsStore.active?.id
)

const project = computed(() =>
  projectId.value
    ? projectsStore.list.find((p) => p.id === projectId.value) ?? null
    : projectsStore.active
)

const gitStatus = computed(() =>
  projectId.value ? projectsStore.gitStatus[projectId.value] : null
)

watch(
  projectId,
  async (id) => {
    if (id) {
      projectsStore.setActive(id)
      await projectsStore.fetchGitStatus(id)
    }
  },
  { immediate: true }
)

async function killSession() {
  // The TerminalPane handles reconnect; we can send a kill signal
  sessionConnected.value = false
}
</script>

<template>
  <div class="flex flex-col h-screen overflow-hidden">
    <!-- Top bar -->
    <AppTopbar
      :project-name="project?.name"
      :branch="gitStatus?.branch"
      :connected="sessionConnected"
      @kill-session="killSession"
      @new-session="() => {}"
    />

    <!-- IDE body -->
    <div class="flex flex-1 min-h-0">
      <!-- Git panel toggle button -->
      <button
        @click="showGitPanel = !showGitPanel"
        class="absolute z-10 top-14 w-5 h-12 bg-surface-700 border-r border-t border-b border-surface-600 flex items-center justify-center text-muted hover:text-text transition-colors duration-120 rounded-r-md"
        :style="{ left: showGitPanel ? '240px' : '0px', transition: 'left 300ms cubic-bezier(0, 0, 0.2, 1)' }"
      >
        <svg
          class="w-3 h-3 transition-transform duration-300"
          :class="showGitPanel ? '' : 'rotate-180'"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>

      <!-- Git panel -->
      <Transition name="slide-right">
        <GitPanel
          v-if="showGitPanel && projectId"
          :project-id="projectId"
        />
      </Transition>

      <!-- Terminal -->
      <div class="flex-1 min-w-0 flex flex-col min-h-0">
        <TerminalPane
          v-if="projectId"
          :project-id="projectId"
        />
        <div
          v-else
          class="flex-1 flex items-center justify-center bg-surface-900"
        >
          <div class="text-center">
            <p class="text-sm text-muted mb-3">No project selected</p>
            <RouterLink
              to="/app/projects"
              class="text-xs text-primary-glow hover:underline"
            >
              Open a project →
            </RouterLink>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
