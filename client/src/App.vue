<script setup lang="ts">
import { useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()

const isSetupRoute = computed(() =>
  route.path.startsWith('/setup')
)
</script>

<template>
  <RouterView v-slot="{ Component, route: currentRoute }">
    <Transition name="page" mode="out-in">
      <component :is="Component" :key="currentRoute.path" />
    </Transition>
  </RouterView>
</template>
