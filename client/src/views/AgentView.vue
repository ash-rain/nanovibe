<script setup lang="ts">
import { ref, nextTick, onMounted, watch } from 'vue'
import { useAgentStore } from '@/stores/agent'
import { useAgentStream } from '@/composables/useAgentStream'
import { useProjectsStore } from '@/stores/projects'
import ChatBubble from '@/components/agent/ChatBubble.vue'
import ChatInput from '@/components/agent/ChatInput.vue'
import TypingIndicator from '@/components/agent/TypingIndicator.vue'

const agentStore = useAgentStore()
const projectsStore = useProjectsStore()
const { connect } = useAgentStream()

const messagesRef = ref<HTMLElement | null>(null)

function scrollToBottom() {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

async function sendMessage(content: string) {
  await agentStore.sendMessage(content)
  scrollToBottom()
}

async function clearMessages() {
  await agentStore.clearMessages()
}

// Watch messages for auto-scroll
watch(
  () => agentStore.messages.length,
  scrollToBottom
)

watch(
  () => agentStore.typing,
  scrollToBottom
)

onMounted(async () => {
  await agentStore.fetchMessages()
  connect()
  scrollToBottom()
})
</script>

<template>
  <div class="flex h-screen overflow-hidden">
    <!-- Sidebar: project selector -->
    <div class="hidden lg:flex w-56 flex-col bg-surface-800 border-r border-surface-600 flex-shrink-0">
      <div class="p-4 border-b border-surface-600">
        <h3 class="text-xs font-semibold text-muted uppercase tracking-wider">Context</h3>
      </div>
      <div class="flex-1 overflow-y-auto p-3">
        <button
          @click="agentStore.setActiveProject(null)"
          :class="[
            'w-full text-left px-3 py-2 rounded-md text-sm transition-all duration-120 mb-1',
            !agentStore.activeProjectId
              ? 'bg-primary/15 text-primary-glow'
              : 'text-muted hover:text-text hover:bg-surface-700',
          ]"
        >
          General
        </button>
        <button
          v-for="project in projectsStore.list"
          :key="project.id"
          @click="agentStore.setActiveProject(project.id)"
          :class="[
            'w-full text-left px-3 py-2 rounded-md text-sm transition-all duration-120',
            agentStore.activeProjectId === project.id
              ? 'bg-primary/15 text-primary-glow'
              : 'text-muted hover:text-text hover:bg-surface-700',
          ]"
        >
          {{ project.name }}
        </button>
      </div>
    </div>

    <!-- Main chat area -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Header -->
      <div class="flex items-center justify-between px-5 py-3 border-b border-surface-600 bg-surface-800">
        <div class="flex items-center gap-3">
          <div class="w-7 h-7 rounded-md bg-primary/20 border border-primary/30 flex items-center justify-center">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <path d="M2 4L7 1L12 4V10L7 13L2 10V4Z" stroke="#a78bfa" stroke-width="1.5" stroke-linejoin="round" />
              <circle cx="7" cy="7" r="1.5" fill="#a78bfa" />
            </svg>
          </div>
          <div>
            <h2 class="text-sm font-semibold text-text">NanoClaw Agent</h2>
            <div class="flex items-center gap-1.5">
              <div
                class="w-1.5 h-1.5 rounded-full"
                :class="agentStore.connected ? 'bg-success animate-pulse' : 'bg-muted'"
              />
              <span class="text-xs text-muted">
                {{ agentStore.connected ? 'Connected' : 'Connecting...' }}
              </span>
            </div>
          </div>
        </div>

        <button
          v-if="agentStore.messages.length"
          @click="clearMessages"
          class="text-xs text-muted hover:text-danger transition-colors duration-120"
        >
          Clear chat
        </button>
      </div>

      <!-- Messages -->
      <div
        ref="messagesRef"
        class="flex-1 overflow-y-auto p-5 space-y-4"
      >
        <!-- Empty state -->
        <div
          v-if="!agentStore.messages.length"
          class="flex flex-col items-center justify-center h-full text-center py-12"
        >
          <div class="w-14 h-14 rounded-xl bg-primary/15 border border-primary/30 flex items-center justify-center mb-4">
            <svg width="24" height="24" viewBox="0 0 14 14" fill="none">
              <path d="M2 4L7 1L12 4V10L7 13L2 10V4Z" stroke="#a78bfa" stroke-width="1.5" stroke-linejoin="round" />
              <circle cx="7" cy="7" r="1.5" fill="#a78bfa" />
            </svg>
          </div>
          <h3 class="text-base font-semibold text-text mb-2">Start a conversation</h3>
          <p class="text-sm text-muted max-w-xs">
            Ask the agent to write code, explain concepts, or help with your project.
          </p>
        </div>

        <!-- Messages -->
        <TransitionGroup name="list" tag="div" class="space-y-4">
          <ChatBubble
            v-for="msg in agentStore.messages"
            :key="msg.id"
            :message="msg"
          />
        </TransitionGroup>

        <!-- Typing indicator -->
        <Transition name="fade">
          <TypingIndicator v-if="agentStore.typing" />
        </Transition>
      </div>

      <!-- Input -->
      <ChatInput
        :disabled="agentStore.typing"
        @send="sendMessage"
      />
    </div>
  </div>
</template>
