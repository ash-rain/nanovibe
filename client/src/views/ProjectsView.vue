<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { FolderOpenIcon, MagnifyingGlassIcon } from '@heroicons/vue/24/outline'
import { useProjectsStore } from '@/stores/projects'
import ProjectCard from '@/components/projects/ProjectCard.vue'
import NewProjectModal from '@/components/projects/NewProjectModal.vue'
import EmptyState from '@/components/ui/EmptyState.vue'

const router = useRouter()
const projectsStore = useProjectsStore()

const showNewProject = ref(false)
const search = ref('')

const filtered = computed(() => {
  if (!search.value.trim()) return projectsStore.list
  const q = search.value.toLowerCase()
  return projectsStore.list.filter(
    (p) =>
      p.name.toLowerCase().includes(q) ||
      p.path.toLowerCase().includes(q) ||
      p.language?.toLowerCase().includes(q)
  )
})

function openIDE(id: string) {
  projectsStore.setActive(id)
  router.push(`/app/ide/${id}`)
}

function openAgent(id: string) {
  projectsStore.setActive(id)
  router.push('/app/agent')
}

async function removeProject(id: string) {
  if (confirm('Remove this project? The files on disk are not affected.')) {
    await projectsStore.remove(id)
  }
}

onMounted(async () => {
  await projectsStore.fetchAll()
  // Fetch git status for visible projects
  for (const p of projectsStore.list) {
    projectsStore.fetchGitStatus(p.id)
  }
})
</script>

<template>
  <div class="p-6 lg:p-8 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-text">Projects</h1>
        <p class="text-sm text-muted mt-1">{{ projectsStore.list.length }} projects</p>
      </div>
      <button
        @click="showNewProject = true"
        class="flex items-center gap-2 px-4 py-2.5 rounded-md bg-primary text-white text-sm font-medium hover:bg-primary-dim transition-all duration-120 shadow-glow"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        New Project
      </button>
    </div>

    <!-- Search -->
    <div v-if="projectsStore.list.length" class="relative mb-6">
      <MagnifyingGlassIcon class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted" />
      <input
        v-model="search"
        type="search"
        placeholder="Search projects..."
        class="w-full pl-9 pr-4 py-2.5 rounded-md bg-surface-800 border border-surface-600 focus:border-primary text-sm text-text placeholder-muted outline-none transition-colors duration-120"
      />
    </div>

    <!-- Projects grid -->
    <div v-if="projectsStore.isLoading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="i in 6"
        :key="i"
        class="h-36 rounded-lg bg-surface-800 border border-surface-600 animate-pulse"
      />
    </div>

    <div v-else-if="!projectsStore.list.length">
      <EmptyState
        title="No projects yet"
        subtitle="Add a local folder or import from GitHub to get started."
        :icon="FolderOpenIcon"
        action-label="New Project"
        @action="showNewProject = true"
      />
    </div>

    <div v-else-if="!filtered.length" class="py-12 text-center">
      <p class="text-sm text-muted">No projects match your search</p>
    </div>

    <TransitionGroup
      v-else
      name="list"
      tag="div"
      class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
    >
      <ProjectCard
        v-for="project in filtered"
        :key="project.id"
        :project="project"
        :git-status="projectsStore.gitStatus[project.id]"
        @open-i-d-e="openIDE"
        @open-agent="openAgent"
      />
    </TransitionGroup>
  </div>

  <!-- New project modal -->
  <NewProjectModal
    v-if="showNewProject"
    @close="showNewProject = false"
    @created="(id) => { showNewProject = false; router.push(`/app/ide/${id}`) }"
  />
</template>
