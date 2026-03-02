<script setup lang="ts">
import { computed } from 'vue'
import type { AgentMessage } from '@/types'
import MarkdownIt from 'markdown-it'

const props = defineProps<{
  message: AgentMessage
}>()

const md = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
  breaks: true,
})

const isUser = computed(() => props.message.role === 'user')

const renderedContent = computed(() => {
  if (isUser.value) return null
  return md.render(props.message.content)
})

function timeStr(ts: number): string {
  return new Date(ts).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div
    :class="[
      'flex gap-3',
      isUser ? 'justify-end' : 'justify-start',
    ]"
  >
    <!-- Agent avatar -->
    <div
      v-if="!isUser"
      class="w-7 h-7 rounded-md bg-primary/20 border border-primary/30 flex items-center justify-center flex-shrink-0 mt-0.5"
    >
      <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
        <path d="M2 4L7 1L12 4V10L7 13L2 10V4Z" stroke="#a78bfa" stroke-width="1.5" stroke-linejoin="round" />
        <circle cx="7" cy="7" r="1.5" fill="#a78bfa" />
      </svg>
    </div>

    <!-- Bubble -->
    <div :class="['max-w-[80%] flex flex-col', isUser ? 'items-end' : 'items-start']">
      <div
        :class="[
          'px-4 py-3 rounded-xl text-sm',
          isUser
            ? 'bg-primary text-white rounded-br-sm'
            : 'bg-surface-700 border border-surface-600 text-text rounded-bl-sm',
        ]"
      >
        <!-- User message: plain text -->
        <p v-if="isUser" class="whitespace-pre-wrap leading-relaxed">
          {{ message.content }}
        </p>

        <!-- Agent message: markdown rendered -->
        <div
          v-else
          class="prose prose-sm max-w-none prose-invert"
          v-html="renderedContent"
        />
      </div>

      <!-- Timestamp -->
      <span class="text-[10px] text-muted mt-1 px-1">
        {{ timeStr(message.createdAt) }}
      </span>
    </div>

    <!-- User avatar -->
    <div
      v-if="isUser"
      class="w-7 h-7 rounded-md bg-surface-600 flex items-center justify-center flex-shrink-0 mt-0.5 text-xs text-muted font-medium"
    >
      U
    </div>
  </div>
</template>

<style>
/* Prose overrides for dark theme */
.prose-invert code {
  background-color: var(--color-surface-800);
  border: 1px solid var(--color-surface-500);
  border-radius: 4px;
  padding: 0.125rem 0.375rem;
  font-family: var(--font-mono);
  font-size: 0.8em;
  color: var(--color-primary-glow);
}

.prose-invert pre {
  background-color: var(--color-surface-900);
  border: 1px solid var(--color-surface-600);
  border-radius: 8px;
  padding: 1rem;
  overflow-x: auto;
}

.prose-invert pre code {
  background: none;
  border: none;
  padding: 0;
  color: var(--color-text);
}

.prose-invert a {
  color: var(--color-primary-glow);
}

.prose-invert p {
  margin: 0.5rem 0;
  line-height: 1.6;
}

.prose-invert p:first-child {
  margin-top: 0;
}

.prose-invert p:last-child {
  margin-bottom: 0;
}

.prose-invert ul, .prose-invert ol {
  padding-left: 1.5rem;
  margin: 0.5rem 0;
}

.prose-invert li {
  margin: 0.25rem 0;
}

.prose-invert h1, .prose-invert h2, .prose-invert h3 {
  font-weight: 600;
  margin: 1rem 0 0.5rem;
  color: var(--color-text);
}
</style>
