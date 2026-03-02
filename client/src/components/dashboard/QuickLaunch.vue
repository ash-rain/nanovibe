<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { PlusIcon, CodeBracketIcon } from '@heroicons/vue/24/outline'
import type { Project } from '@/types'

const props = defineProps<{
  projects: Project[]
}>()

const emit = defineEmits<{
  newProject: []
}>()

const router = useRouter()

const languageColors: Record<string, string> = {
  typescript: '#3178c6',
  javascript: '#f7df1e',
  python: '#3776ab',
  rust: '#ce422b',
  go: '#00add8',
  java: '#b07219',
  cpp: '#f34b7d',
  c: '#555555',
  ruby: '#cc342d',
  php: '#4F5D95',
}

function getLangColor(lang?: string): string {
  if (!lang) return '#64748b'
  return languageColors[lang.toLowerCase()] ?? '#64748b'
}

function openIDE(project: Project) {
  router.push(`/app/ide/${project.id}`)
}
</script>

<template>
  <div class="bg-surface-800 border border-surface-600 rounded-lg p-5 shadow-card">
    <h3 class="text-sm font-semibold text-text mb-4">Quick Launch</h3>

    <div class="grid grid-cols-2 gap-3">
      <!-- Project cards -->
      <div
        v-for="project in projects.slice(0, 4)"
        :key="project.id"
        @click="openIDE(project)"
        class="relative group bg-surface-700 hover:bg-surface-600 border border-surface-600 hover:border-surface-500 rounded-md p-3 cursor-pointer transition-all duration-200"
      >
        <!-- Language dot -->
        <div
          class="w-2 h-2 rounded-full mb-2"
          :style="`background-color: ${getLangColor(project.language)}`"
        />
        <p class="text-sm font-medium text-text truncate">{{ project.name }}</p>
        <p class="text-xs text-muted truncate mt-0.5">{{ project.path }}</p>

        <!-- Hover overlay -->
        <div
          class="absolute inset-0 flex items-center justify-center gap-2 bg-surface-700/90 rounded-md opacity-0 group-hover:opacity-100 transition-opacity duration-200"
        >
          <button
            @click.stop="openIDE(project)"
            class="px-2.5 py-1.5 rounded bg-primary text-white text-xs font-medium hover:bg-primary-dim transition-colors duration-120"
          >
            Open IDE
          </button>
        </div>
      </div>

      <!-- New project card -->
      <button
        @click="emit('newProject')"
        class="flex flex-col items-center justify-center gap-2 bg-surface-700/50 hover:bg-surface-700 border border-dashed border-surface-500 hover:border-surface-400 rounded-md p-3 text-muted hover:text-text transition-all duration-200 cursor-pointer"
        :class="projects.length === 0 ? 'col-span-2' : ''"
      >
        <PlusIcon class="w-5 h-5" />
        <span class="text-xs font-medium">New Project</span>
      </button>
    </div>
  </div>
</template>
