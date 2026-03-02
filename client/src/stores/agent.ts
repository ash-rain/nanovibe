import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useFetch } from '@/composables/useFetch'
import type { AgentMessage } from '@/types'

export const useAgentStore = defineStore('agent', () => {
  const { get, post, del } = useFetch()

  const messages = ref<AgentMessage[]>([])
  const connected = ref(false)
  const typing = ref(false)
  const activeProjectId = ref<string | null>(null)

  async function fetchMessages() {
    const { data, error } = await get<{ data: AgentMessage[] }>(
      '/api/agent/messages'
    )
    if (!error && data) {
      messages.value = data.data
    }
    return { error }
  }

  async function sendMessage(content: string) {
    const payload: { content: string; projectId?: string } = { content }
    if (activeProjectId.value) {
      payload.projectId = activeProjectId.value
    }
    const { error } = await post('/api/agent/message', payload)
    if (!error) {
      typing.value = true
      messages.value.push({
        id: Date.now(),
        role: 'user',
        content,
        createdAt: Date.now(),
        projectId: activeProjectId.value ?? undefined,
      })
    }
    return { error }
  }

  async function clearMessages() {
    const { error } = await del('/api/agent/messages')
    if (!error) {
      messages.value = []
    }
    return { error }
  }

  function setActiveProject(id: string | null) {
    activeProjectId.value = id
  }

  function addMessage(msg: AgentMessage) {
    const existing = messages.value.find((m) => m.id === msg.id)
    if (!existing) {
      messages.value.push(msg)
    }
  }

  return {
    messages,
    connected,
    typing,
    activeProjectId,
    fetchMessages,
    sendMessage,
    clearMessages,
    setActiveProject,
    addMessage,
  }
})
