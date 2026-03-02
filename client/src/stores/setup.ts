import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useFetch } from '@/composables/useFetch'
import type { SystemCheck, TunnelStatus } from '@/types'

export const useSetupStore = defineStore('setup', () => {
  const { get, post } = useFetch()

  const currentStep = ref('welcome')
  const completedSteps = ref<string[]>([])
  const checks = ref<SystemCheck[]>([])
  const tunnelStatus = ref<TunnelStatus>({
    mode: 'none',
    connected: false,
    tunnelUrl: '',
    localUrl: '',
    uptimeS: 0,
  })
  const isLoading = ref(false)
  const loaded = ref(false)

  async function fetchState() {
    isLoading.value = true
    const { data, error } = await get<{
      currentStep: string
      completedSteps: string[]
    }>('/api/setup/state')
    isLoading.value = false

    if (!error && data) {
      currentStep.value = data.currentStep
      completedSteps.value = data.completedSteps || []
      loaded.value = true
    }
    return { data, error }
  }

  async function advanceTo(step: string) {
    const { error } = await post('/api/setup/state', { step })
    if (!error) {
      currentStep.value = step
      if (!completedSteps.value.includes(currentStep.value)) {
        completedSteps.value.push(currentStep.value)
      }
    }
    return { error }
  }

  async function runChecks() {
    const { data, error } = await get<{ checks: SystemCheck[] }>(
      '/api/setup/check/system'
    )
    if (!error && data) {
      checks.value = data.checks
    }
    return { data, error }
  }

  function fixCheck(id: string): EventSource {
    const es = new EventSource(`/api/setup/fix/${id}`)
    return es
  }

  function isStepCompleted(step: string): boolean {
    return completedSteps.value.includes(step)
  }

  return {
    currentStep,
    completedSteps,
    checks,
    tunnelStatus,
    isLoading,
    loaded,
    fetchState,
    advanceTo,
    runChecks,
    fixCheck,
    isStepCompleted,
  }
})
