<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMetrics } from '@/composables/useMetrics'

import { useProjectsStore } from '@/stores/projects'
import { useGithubStore } from '@/stores/github'
import { useSettingsStore } from '@/stores/settings'
import ServiceCard from '@/components/dashboard/ServiceCard.vue'
import VitalsPanel from '@/components/dashboard/VitalsPanel.vue'
import ActivityFeed from '@/components/dashboard/ActivityFeed.vue'
import QuickLaunch from '@/components/dashboard/QuickLaunch.vue'
import AccessPanel from '@/components/dashboard/AccessPanel.vue'
import NewProjectModal from '@/components/projects/NewProjectModal.vue'

const router = useRouter()
const projectsStore = useProjectsStore()
const githubStore = useGithubStore()
const settingsStore = useSettingsStore()
const { current, history, connect: connectMetrics } = useMetrics()

const showNewProject = ref(false)
let activityTimer: ReturnType<typeof setInterval> | null = null

const serviceCards = [
  {
    id: 'opencode',
    title: 'OpenCode',
    description: 'AI coding agent',
  },
  {
    id: 'nanoclaw',
    title: 'NanoClaw',
    description: 'AI chat agent',
  },
  {
    id: 'cloudflare',
    title: 'Cloudflare',
    description: 'Secure tunnel',
  },
  {
    id: 'docker',
    title: 'Docker',
    description: 'Container runtime',
  },
]

async function fetchActivity() {
  if (githubStore.connected) {
    await githubStore.fetchActivity()
  }
}

onMounted(async () => {
  // Start metrics SSE
  connectMetrics()

  // Load data
  await Promise.all([
    projectsStore.fetchAll(),
    settingsStore.fetchTunnel(),
    settingsStore.fetchSystem(),
    githubStore.fetchStatus().then(() => fetchActivity()),
  ])

  // Poll activity every 60s
  activityTimer = setInterval(fetchActivity, 60_000)
})

onUnmounted(() => {
  if (activityTimer) clearInterval(activityTimer)
})

function getRecentProjects() {
  return projectsStore.getRecentProjects(4)
}
</script>

<template>
  <div class="p-6 lg:p-8 max-w-7xl mx-auto space-y-6">
    <!-- Page title -->
    <div>
      <h1 class="text-2xl font-bold text-text">Dashboard</h1>
      <p class="text-sm text-muted mt-1">Your AI coding station</p>
    </div>

    <!-- Service cards row -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <ServiceCard
        v-for="svc in serviceCards"
        :key="svc.id"
        :title="svc.title"
        status="stopped"
      />
    </div>

    <!-- Middle row: Quick Launch + Vitals -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <QuickLaunch
        :projects="getRecentProjects()"
        @new-project="showNewProject = true"
      />
      <VitalsPanel :current="current" :history="history" />
    </div>

    <!-- Bottom row: Activity + Access -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <ActivityFeed
        :events="githubStore.activity"
        :loading="githubStore.isLoading"
      />
      <AccessPanel
        :tunnel="settingsStore.tunnel"
        :local-url="settingsStore.system.localUrl"
      />
    </div>
  </div>

  <!-- New project modal -->
  <NewProjectModal
    v-if="showNewProject"
    @close="showNewProject = false"
    @created="(id) => { showNewProject = false; router.push(`/app/ide/${id}`) }"
  />
</template>
