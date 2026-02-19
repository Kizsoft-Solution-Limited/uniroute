import { apiClient } from './client'

export interface ProviderInfo {
  name: string
  healthy: boolean
  models: string[]
}

export interface ProvidersResponse {
  providers: ProviderInfo[]
}

export const providersApi = {
  async list(): Promise<ProvidersResponse> {
    const token = localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token')
    const endpoint = token ? '/auth/providers' : '/v1/providers'
    const response = await apiClient.get<ProvidersResponse>(endpoint)
    return response.data
  }
}
