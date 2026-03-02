import { ref, onUnmounted } from 'vue'
import type { SystemMetrics } from '@/types'

const HISTORY_SIZE = 60

export function useMetrics() {
  const current = ref<SystemMetrics | null>(null)
  const history = ref<SystemMetrics[]>([])
  const connected = ref(false)

  let es: EventSource | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectDelay = 1000
  let destroyed = false

  function connect() {
    if (destroyed) return

    es = new EventSource('/api/metrics/stream')

    es.addEventListener('metrics', (event: MessageEvent) => {
      try {
        const metrics: SystemMetrics = JSON.parse(event.data)
        current.value = metrics
        history.value.push(metrics)
        if (history.value.length > HISTORY_SIZE) {
          history.value.shift()
        }
        connected.value = true
        reconnectDelay = 1000
      } catch {
        // ignore parse errors
      }
    })

    es.onopen = () => {
      connected.value = true
    }

    es.onerror = () => {
      connected.value = false
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
    connected.value = false
  }

  onUnmounted(() => {
    destroyed = true
    disconnect()
  })

  return {
    current,
    history,
    connected,
    connect,
    disconnect,
  }
}
