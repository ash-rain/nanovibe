<script setup lang="ts">
import AppSidebar from './AppSidebar.vue'
import MobileNav from './MobileNav.vue'
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-surface-900">
    <!-- Sidebar (desktop) -->
    <AppSidebar class="hidden md:flex" />

    <!-- Main content -->
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
      <main class="flex-1 overflow-auto">
        <RouterView v-slot="{ Component, route }">
          <Transition name="page" mode="out-in">
            <component :is="Component" :key="route.path" />
          </Transition>
        </RouterView>
      </main>
    </div>

    <!-- Mobile bottom nav -->
    <MobileNav class="md:hidden" />
  </div>
</template>
