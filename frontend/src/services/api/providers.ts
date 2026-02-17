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
    const response = await apiClient.get<ProvidersResponse>('/v1/providers')
    return response.data
  }
}
