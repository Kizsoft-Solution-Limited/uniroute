import { apiClient } from './client'

export interface ProviderKey {
  id: string
  provider: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface ProviderKeysResponse {
  keys: ProviderKey[]
}

export interface TestKeyResponse {
  message: string
  provider: string
  status: string
}

export const providerKeysApi = {
  /**
   * List all provider keys for the current user
   */
  async list(): Promise<ProviderKeysResponse> {
    const response = await apiClient.get<ProviderKeysResponse>('/admin/provider-keys')
    return response.data
  },

  /**
   * Add a new provider key
   */
  async add(provider: string, apiKey: string): Promise<void> {
    await apiClient.post('/admin/provider-keys', {
      provider,
      api_key: apiKey
    })
  },

  /**
   * Update an existing provider key
   */
  async update(provider: string, apiKey: string): Promise<void> {
    await apiClient.put(`/admin/provider-keys/${provider}`, {
      api_key: apiKey
    })
  },

  /**
   * Delete a provider key
   */
  async delete(provider: string): Promise<void> {
    await apiClient.delete(`/admin/provider-keys/${provider}`)
  },

  /**
   * Test a provider key connection
   */
  async test(provider: string): Promise<TestKeyResponse> {
    const response = await apiClient.post<TestKeyResponse>(`/admin/provider-keys/${provider}/test`)
    return response.data
  }
}

