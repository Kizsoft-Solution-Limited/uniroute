import { apiClient } from './client'

export interface Tunnel {
  id: string
  subdomain: string
  public_url: string
  local_url: string
  status: 'active' | 'inactive'
  request_count: number
  created_at: string
  last_active?: string | null
  active_since?: string | null
  custom_domain?: string | null
  protocol?: string
  user_id?: string | null
  user_display?: string | null
}

export interface ListTunnelsResponse {
  tunnels: Tunnel[]
}

export interface GetTunnelResponse {
  tunnel: Tunnel
}

export interface TunnelStatsPoint {
  time: string
  active_tunnels: number
  total_tunnels: number
}

export interface TunnelStatsResponse {
  period: {
    start: string
    end: string
  }
  interval_hours: number // Can be fractional (e.g., 0.25 for 15 minutes)
  data: TunnelStatsPoint[]
}

export const tunnelsApi = {
  async list(useJWT: boolean = true): Promise<ListTunnelsResponse> {
    const endpoint = useJWT ? '/auth/tunnels' : '/v1/tunnels'
    const response = await apiClient.get<ListTunnelsResponse>(endpoint)
    return response.data
  },

  async get(id: string, useJWT: boolean = true): Promise<GetTunnelResponse> {
    const endpoint = useJWT ? `/auth/tunnels/${id}` : `/v1/tunnels/${id}`
    const response = await apiClient.get<GetTunnelResponse>(endpoint)
    return response.data
  },

  async disconnect(id: string, useJWT: boolean = true): Promise<void> {
    const endpoint = useJWT ? `/auth/tunnels/${id}/disconnect` : `/v1/tunnels/${id}/disconnect`
    await apiClient.post(endpoint)
  },

  async getStats(hours: number = 24, interval?: number, useJWT: boolean = true): Promise<TunnelStatsResponse> {
    const endpoint = '/admin/tunnels/stats'
    const params: Record<string, any> = { hours }
    if (interval !== undefined) {
      params.interval = interval
    }
    const response = await apiClient.get<TunnelStatsResponse>(endpoint, {
      params
    })
    return response.data
  },

  async setCustomDomain(id: string, domain: string, useJWT: boolean = true): Promise<{ message: string; domain: string }> {
    const endpoint = useJWT ? `/auth/tunnels/${id}/domain` : `/v1/tunnels/${id}/domain`
    const response = await apiClient.post<{ message: string; domain: string }>(endpoint, {
      domain
    })
    return response.data
  },

  async getCountsAdmin(): Promise<{ total: number; active: number }> {
    const response = await apiClient.get<{ total: number; active: number }>('/admin/tunnels/counts')
    return response.data
  },

  async listAdmin(limit = 50, offset = 0): Promise<{ tunnels: Tunnel[]; total: number; count: number }> {
    const response = await apiClient.get<{ tunnels: Tunnel[]; total: number; count: number }>(
      '/admin/tunnels',
      { params: { limit, offset } }
    )
    return response.data
  },

  async deleteAdmin(tunnelId: string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/admin/tunnels/${tunnelId}`)
    return response.data
  },

  async deleteManyAdmin(tunnelIds: string[]): Promise<{ message: string; deleted: number; failed: string[] }> {
    const response = await apiClient.post<{ message: string; deleted: number; failed: string[] }>(
      '/admin/tunnels/delete',
      { tunnel_ids: tunnelIds }
    )
    return response.data
  },
}