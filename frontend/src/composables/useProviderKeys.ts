import { ref } from 'vue'
import { providerKeysApi } from '@/services/api/providerKeys'

export interface ProviderKey {
  id: string
  provider: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export function useProviderKeys() {
  const providerKeys = ref<ProviderKey[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const fetchKeys = async () => {
    loading.value = true
    error.value = null
    try {
      const response = await providerKeysApi.list()
      providerKeys.value = response.keys || []
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch provider keys'
      throw err
    } finally {
      loading.value = false
    }
  }

  const addKey = async (provider: string, apiKey: string) => {
    loading.value = true
    error.value = null
    try {
      await providerKeysApi.add(provider, apiKey)
      await fetchKeys()
    } catch (err: any) {
      error.value = err.message || 'Failed to add provider key'
      throw err
    } finally {
      loading.value = false
    }
  }

  const updateKey = async (provider: string, apiKey: string) => {
    loading.value = true
    error.value = null
    try {
      await providerKeysApi.update(provider, apiKey)
      await fetchKeys()
    } catch (err: any) {
      error.value = err.message || 'Failed to update provider key'
      throw err
    } finally {
      loading.value = false
    }
  }

  const deleteKey = async (provider: string) => {
    loading.value = true
    error.value = null
    try {
      await providerKeysApi.delete(provider)
      await fetchKeys()
    } catch (err: any) {
      error.value = err.message || 'Failed to delete provider key'
      throw err
    } finally {
      loading.value = false
    }
  }

  const testKey = async (provider: string) => {
    loading.value = true
    error.value = null
    try {
      const response = await providerKeysApi.test(provider)
      return response
    } catch (err: any) {
      error.value = err.message || 'Failed to test provider key'
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    providerKeys,
    loading,
    error,
    fetchKeys,
    addKey,
    updateKey,
    deleteKey,
    testKey
  }
}

