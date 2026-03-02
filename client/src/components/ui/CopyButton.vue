<script setup lang="ts">
import { ref } from 'vue'
import { ClipboardIcon, CheckIcon } from '@heroicons/vue/24/outline'

const props = defineProps<{
  text: string
  label?: string
}>()

const copied = ref(false)

async function copy() {
  try {
    await navigator.clipboard.writeText(props.text)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 1500)
  } catch {
    // Fallback
    const el = document.createElement('textarea')
    el.value = props.text
    el.style.position = 'fixed'
    el.style.opacity = '0'
    document.body.appendChild(el)
    el.select()
    document.execCommand('copy')
    document.body.removeChild(el)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 1500)
  }
}
</script>

<template>
  <button
    @click="copy"
    :class="[
      'inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded text-xs transition-all duration-200',
      copied
        ? 'bg-success/15 text-success'
        : 'bg-surface-700 text-muted hover:text-text hover:bg-surface-600',
    ]"
  >
    <Transition name="fade" mode="out-in">
      <CheckIcon v-if="copied" class="w-3.5 h-3.5" />
      <ClipboardIcon v-else class="w-3.5 h-3.5" />
    </Transition>
    <span>{{ copied ? 'Copied!' : (label ?? 'Copy') }}</span>
  </button>
</template>
