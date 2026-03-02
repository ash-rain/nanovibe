<script setup lang="ts">
import { useRoute } from 'vue-router'
import {
  HomeIcon,
  CodeBracketIcon,
  ChatBubbleLeftRightIcon,
  FolderOpenIcon,
  Cog6ToothIcon,
} from '@heroicons/vue/24/outline'
import {
  HomeIcon as HomeIconSolid,
  CodeBracketIcon as CodeBracketIconSolid,
  ChatBubbleLeftRightIcon as ChatBubbleLeftRightIconSolid,
  FolderOpenIcon as FolderOpenIconSolid,
  Cog6ToothIcon as Cog6ToothIconSolid,
} from '@heroicons/vue/24/solid'

const route = useRoute()

const navItems = [
  { name: 'Home', path: '/app/dashboard', icon: HomeIcon, iconSolid: HomeIconSolid },
  { name: 'IDE', path: '/app/ide', icon: CodeBracketIcon, iconSolid: CodeBracketIconSolid },
  { name: 'Agent', path: '/app/agent', icon: ChatBubbleLeftRightIcon, iconSolid: ChatBubbleLeftRightIconSolid },
  { name: 'Projects', path: '/app/projects', icon: FolderOpenIcon, iconSolid: FolderOpenIconSolid },
  { name: 'Settings', path: '/app/settings', icon: Cog6ToothIcon, iconSolid: Cog6ToothIconSolid },
]

function isActive(path: string): boolean {
  if (path === '/app/ide') return route.path.startsWith('/app/ide')
  return route.path === path
}
</script>

<template>
  <nav
    class="fixed bottom-0 left-0 right-0 flex items-center bg-surface-800 border-t border-surface-600 pb-safe z-50"
  >
    <RouterLink
      v-for="item in navItems"
      :key="item.path"
      :to="item.path"
      class="flex-1 flex flex-col items-center gap-0.5 py-2 transition-colors duration-120"
      :class="isActive(item.path) ? 'text-primary-glow' : 'text-muted'"
    >
      <component
        :is="isActive(item.path) ? item.iconSolid : item.icon"
        class="w-5 h-5"
      />
      <span class="text-[10px] font-medium">{{ item.name }}</span>
    </RouterLink>
  </nav>
</template>
