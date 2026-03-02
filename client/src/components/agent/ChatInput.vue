<script setup lang="ts">
import { ref, nextTick } from 'vue'

const props = defineProps<{
  disabled?: boolean
  placeholder?: string
}>()

const emit = defineEmits<{
  send: [content: string]
}>()

const value = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)

function autoResize() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 200) + 'px'
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    submit()
  }
}

function submit() {
  const text = value.value.trim()
  if (!text || props.disabled) return
  emit('send', text)
  value.value = ''
  nextTick(() => {
    if (textareaRef.value) {
      textareaRef.value.style.height = 'auto'
    }
  })
}
</script>

<template>
  <div class="relative flex items-end gap-3 p-4 border-t border-surface-600 bg-surface-800">
    <div class="flex-1 bg-surface-700 border border-surface-600 rounded-md focus-within:border-primary transition-colors duration-120">
      <textarea
        ref="textareaRef"
        v-model="value"
        :placeholder="placeholder ?? 'Message the agent... (Shift+Enter for newline)'"
        :disabled="disabled"
        @input="autoResize"
        @keydown="handleKeydown"
        rows="1"
        class="w-full resize-none bg-transparent px-3 py-2.5 text-sm text-text placeholder-muted outline-none disabled:opacity-50 leading-relaxed"
        style="max-height: 200px; min-height: 40px"
      />
    </div>

    <button
      @click="submit"
      :disabled="disabled || !value.trim()"
      class="flex-shrink-0 w-9 h-9 rounded-md bg-primary hover:bg-primary-dim disabled:opacity-40 disabled:cursor-not-allowed transition-all duration-120 flex items-center justify-center shadow-glow"
    >
      <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18" />
      </svg>
    </button>
  </div>
</template>
