import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { SystemMetrics } from '@/types'

export const useMetricsStore = defineStore('metrics', () => {
  const current = ref<SystemMetrics | null>(null)
  const history = ref<SystemMetrics[]>([])
  const connected = ref(false)

  function addMetrics(metrics: SystemMetrics) {
    current.value = metrics
    history.value.push(metrics)
    if (history.value.length > 60) {
      history.value.shift()
    }
    connected.value = true
  }

  function setConnected(val: boolean) {
    connected.value = val
  }

  return {
    current,
    history,
    connected,
    addMetrics,
    setConnected,
  }
})
