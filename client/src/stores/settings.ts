import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useFetch } from '@/composables/useFetch'
import type { ProviderConfig, TunnelStatus, SystemInfo } from '@/types'

export const useSettingsStore = defineStore('settings', () => {
  const { get, post } = useFetch()

  const providers = ref<ProviderConfig[]>([])
  const tunnel = ref<TunnelStatus>({
    mode: 'none',
    connected: false,
    tunnelUrl: '',
    localUrl: '',
    uptimeS: 0,
  })
  const system = ref<SystemInfo>({
    hostname: '',
    ip: '',
    localUrl: '',
    goVersion: '',
    dockerVersion: '',
    appVersion: '',
  })
  const isLoading = ref(false)

  async function fetchProviders() {
    const { data, error } = await get<{ providers: ProviderConfig[] }>(
      '/api/settings/providers'
    )
    if (!error && data) {
      providers.value = data.providers
    }
    return { error }
  }

  async function saveProvider(provider: string, key: string) {
    const { error } = await post('/api/settings/providers', { provider, key })
    if (!error) {
      await fetchProviders()
    }
    return { error }
  }

  async function fetchTunnel() {
    const { data, error } = await get<TunnelStatus>('/api/settings/tunnel')
    if (!error && data) {
      tunnel.value = data
    }
    return { error }
  }

  async function restartTunnel() {
    const { error } = await post('/api/settings/tunnel/restart')
    if (!error) {
      await fetchTunnel()
    }
    return { error }
  }

  async function upgradeTunnel(token: string) {
    const { error } = await post('/api/settings/tunnel/upgrade', { token })
    if (!error) {
      await fetchTunnel()
    }
    return { error }
  }

  async function fetchSystem() {
    const { data, error } = await get<SystemInfo>('/api/settings/system')
    if (!error && data) {
      system.value = data
    }
    return { error }
  }

  function update(): EventSource {
    return new EventSource('/api/settings/system/update')
  }

  return {
    providers,
    tunnel,
    system,
    isLoading,
    fetchProviders,
    saveProvider,
    fetchTunnel,
    restartTunnel,
    upgradeTunnel,
    fetchSystem,
    update,
  }
})
