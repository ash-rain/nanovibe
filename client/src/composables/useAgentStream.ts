import { onUnmounted } from 'vue'
import { useAgentStore } from '@/stores/agent'

export function useAgentStream() {
  const agentStore = useAgentStore()

  let es: EventSource | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectDelay = 1000
  let destroyed = false

  function connect() {
    if (destroyed) return

    es = new EventSource('/api/agent/stream')

    es.addEventListener('message', (event: MessageEvent) => {
      try {
        const msg = JSON.parse(event.data)
        agentStore.addMessage(msg)
        agentStore.typing = false
        reconnectDelay = 1000
      } catch {
        // ignore
      }
    })

    es.addEventListener('ping', () => {
      // Keep-alive ping — no action needed
    })

    es.addEventListener('typing', (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data)
        agentStore.typing = data.typing === true
      } catch {
        // ignore
      }
    })

    es.onopen = () => {
      agentStore.connected = true
    }

    es.onerror = () => {
      agentStore.connected = false
      es?.close()
      es = null
      if (!destroyed) {
        scheduleReconnect()
      }
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    reconnectTimer = setTimeout(() => {
      if (!destroyed) {
        reconnectDelay = Math.min(reconnectDelay * 2, 30000)
        connect()
      }
    }, reconnectDelay)
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (es) {
      es.close()
      es = null
    }
    agentStore.connected = false
  }

  onUnmounted(() => {
    destroyed = true
    disconnect()
  })

  return { connect, disconnect }
}
