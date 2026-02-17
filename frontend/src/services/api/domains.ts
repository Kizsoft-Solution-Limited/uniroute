import { apiClient } from './client'

export interface CustomDomain {
  id: string
  domain: string
  verified: boolean
  dns_configured: boolean
  created_at: string
  updated_at: string
}

export interface ListDomainsResponse {
  domains: CustomDomain[]
}

export interface CreateDomainResponse {
  message: string
  domain: CustomDomain
}

export const domainsApi = {
  async list(): Promise<ListDomainsResponse> {
    const response = await apiClient.get<ListDomainsResponse>('/auth/domains')
    return response.data
  },

  async create(domain: string): Promise<CreateDomainResponse> {
    const response = await apiClient.post<CreateDomainResponse>('/auth/domains', {
      domain
    })
    return response.data
  },

  async delete(id: string): Promise<void> {
    await apiClient.delete(`/auth/domains/${id}`)
  },

  async verify(id: string): Promise<{ domain: string; dns_configured: boolean; dns_error?: string; dns_instructions: any }> {
    const response = await apiClient.post<{ domain: string; dns_configured: boolean; dns_error?: string; dns_instructions: any }>(`/auth/domains/${id}/verify`)
    return response.data
  },
}
