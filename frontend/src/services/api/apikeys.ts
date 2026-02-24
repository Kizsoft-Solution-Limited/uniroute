import { apiClient } from './client'

export interface ApiKey {
  id: string
  name: string
  is_active: boolean
  created_at: string
  expires_at?: string | null
  rate_limit_per_minute: number
  rate_limit_per_day: number
  key_preview?: string
}

export interface CreateApiKeyRequest {
  name: string
  rate_limit_per_minute?: number
  rate_limit_per_day?: number
  expires_at?: string | null
}

export interface CreateApiKeyResponse {
  id: string
  key: string // Only returned once on creation!
  name: string
  created_at: string
  expires_at?: string | null
  message: string
}

export interface ListApiKeysResponse {
  keys: ApiKey[]
}

export const apiKeysApi = {
  async list(): Promise<ListApiKeysResponse> {
    const response = await apiClient.get<ListApiKeysResponse>('/auth/api-keys')
    return response.data
  },

  async create(data: CreateApiKeyRequest): Promise<CreateApiKeyResponse> {
    const response = await apiClient.post<CreateApiKeyResponse>('/auth/api-keys', data)
    return response.data
  },

  async revoke(id: string): Promise<void> {
    await apiClient.delete(`/auth/api-keys/${id}`)
  },

  async deletePermanently(id: string): Promise<void> {
    await apiClient.delete(`/auth/api-keys/${id}/permanent`)
  },
}

