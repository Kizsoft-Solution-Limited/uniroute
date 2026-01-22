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
  custom_domain?: string | null
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
  /**
   * List all tunnels for the authenticated user
   * Uses /auth/tunnels for frontend (JWT auth) or /v1/tunnels for API keys
   */
  async list(useJWT: boolean = true): Promise<ListTunnelsResponse> {
    const endpoint = useJWT ? '/auth/tunnels' : '/v1/tunnels'
    const response = await apiClient.get<ListTunnelsResponse>(endpoint)
    return response.data
  },

  /**
   * Get tunnel details by ID
   */
  async get(id: string, useJWT: boolean = true): Promise<GetTunnelResponse> {
    const endpoint = useJWT ? `/auth/tunnels/${id}` : `/v1/tunnels/${id}`
    const response = await apiClient.get<GetTunnelResponse>(endpoint)
    return response.data
  },

  /**
   * Disconnect a tunnel
   */
  async disconnect(id: string, useJWT: boolean = true): Promise<void> {
    const endpoint = useJWT ? `/auth/tunnels/${id}/disconnect` : `/v1/tunnels/${id}/disconnect`
    await apiClient.post(endpoint)
  },

  /**
   * Get tunnel statistics over time for charting (admin only - shows all tunnels)
   * @param hours - Time range in hours (6, 24, or 168 for 7d)
   * @param interval - Optional interval in hours (only needed for 7d view, otherwise backend aggregates)
   */
  async getStats(hours: number = 24, interval?: number, useJWT: boolean = true): Promise<TunnelStatsResponse> {
    // Admin endpoint - always use /admin/tunnels/stats
    // For 6h/24h: backend returns single aggregated point (no interval needed)
    // For 7d: backend returns one point per day (interval=24)
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

  /**
   * Set custom domain for a tunnel
   * @param id - Tunnel ID
   * @param domain - Custom domain to set
   * @param useJWT - Whether to use JWT authentication (default: true)
   */
  async setCustomDomain(id: string, domain: string, useJWT: boolean = true): Promise<{ message: string; domain: string }> {
    const endpoint = useJWT ? `/auth/tunnels/${id}/domain` : `/v1/tunnels/${id}/domain`
    const response = await apiClient.post<{ message: string; domain: string }>(endpoint, {
      domain
    })
    return response.data
  },
}