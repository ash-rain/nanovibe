<script setup lang="ts">
import { useRoute } from 'vue-router'

import {
  HomeIcon,
  CodeBracketIcon,
  ChatBubbleLeftRightIcon,
  FolderOpenIcon,
  CodeBracketSquareIcon,
  Cog6ToothIcon,
} from '@heroicons/vue/24/outline'
import StatusDot from '@/components/ui/StatusDot.vue'
import type { ServiceStatus } from '@/types'

const route = useRoute()

const navItems = [
  { name: 'Dashboard', path: '/app/dashboard', icon: HomeIcon },
  { name: 'IDE', path: '/app/ide', icon: CodeBracketIcon },
  { name: 'Agent', path: '/app/agent', icon: ChatBubbleLeftRightIcon },
  { name: 'Projects', path: '/app/projects', icon: FolderOpenIcon },
  { name: 'GitHub', path: '/app/github', icon: CodeBracketSquareIcon },
  { name: 'Settings', path: '/app/settings', icon: Cog6ToothIcon },
]

function isActive(path: string): boolean {
  if (path === '/app/ide') {
    return route.path.startsWith('/app/ide')
  }
  return route.path === path
}

// Service status placeholders — in production these would come from stores
const services: Array<{ name: string; status: ServiceStatus }> = [
  { name: 'opencode', status: 'stopped' },
  { name: 'NanoClaw', status: 'stopped' },
  { name: 'Cloudflare', status: 'stopped' },
  { name: 'Docker', status: 'stopped' },
]
</script>

<template>
  <aside
    class="flex-col w-[220px] bg-surface-800 border-r border-surface-600 flex-shrink-0"
    style="min-height: 100vh"
  >
    <!-- Logo -->
    <div class="px-5 py-5 border-b border-surface-600">
      <div class="flex items-center gap-2.5">
        <div
          class="w-7 h-7 rounded-md bg-primary flex items-center justify-center flex-shrink-0"
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 14 14"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M2 4L7 1L12 4V10L7 13L2 10V4Z"
              stroke="white"
              stroke-width="1.5"
              stroke-linejoin="round"
            />
            <circle cx="7" cy="7" r="1.5" fill="white" />
          </svg>
        </div>
        <span class="font-semibold text-sm text-text tracking-tight">VibeCodePC</span>
      </div>
    </div>

    <!-- Nav items -->
    <nav class="flex-1 p-3 space-y-0.5">
      <RouterLink
        v-for="item in navItems"
        :key="item.path"
        :to="item.path"
        custom
        v-slot="{ navigate }"
      >
        <button
          @click="navigate"
          :class="[
            'w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-all duration-120 group relative',
            isActive(item.path)
              ? 'bg-primary/15 text-primary-glow font-medium'
              : 'text-muted hover:text-text hover:bg-surface-700',
          ]"
        >
          <!-- Active indicator bar -->
          <div
            v-if="isActive(item.path)"
            class="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-4 bg-primary rounded-r"
          />
          <component
            :is="item.icon"
            :class="[
              'w-4 h-4 flex-shrink-0 transition-colors duration-120',
              isActive(item.path) ? 'text-primary-glow' : 'text-muted group-hover:text-text',
            ]"
          />
          <span>{{ item.name }}</span>
        </button>
      </RouterLink>
    </nav>

    <!-- Service status footer -->
    <div class="p-4 border-t border-surface-600">
      <p class="text-[10px] font-medium uppercase tracking-widest text-muted mb-3">
        Services
      </p>
      <div class="grid grid-cols-2 gap-2">
        <div
          v-for="svc in services"
          :key="svc.name"
          class="flex items-center gap-1.5"
        >
          <StatusDot :status="svc.status" size="xs" />
          <span class="text-xs text-muted truncate">{{ svc.name }}</span>
        </div>
      </div>
    </div>
  </aside>
</template>
