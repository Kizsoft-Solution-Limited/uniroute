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
}

export interface ListTunnelsResponse {
  tunnels: Tunnel[]
}

export interface GetTunnelResponse {
  tunnel: Tunnel
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
  async get(id: string, useJWT: boolean = true): Promise<Tunnel> {
    const endpoint = useJWT ? `/auth/tunnels/${id}` : `/v1/tunnels/${id}`
    const response = await apiClient.get<{ tunnel: Tunnel }>(endpoint)
    return response.data.tunnel
  },

  /**
   * Disconnect a tunnel
   */
  async disconnect(id: string, useJWT: boolean = true): Promise<void> {
    const endpoint = useJWT ? `/auth/tunnels/${id}/disconnect` : `/v1/tunnels/${id}/disconnect`
    await apiClient.post(endpoint)
  },
}

