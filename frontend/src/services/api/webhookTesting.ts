import { apiClient } from './client'
import { tunnelsApi } from './tunnels'

export interface Tunnel {
  id: string
  subdomain: string
  public_url: string
  local_url: string
  request_count: number
  created_at: string
  last_active: string
}

export interface RequestSummary {
  id: string
  request_id: string
  method: string
  path: string
  query_string: string
  status_code: number
  latency_ms: number
  request_size: number
  response_size: number
  remote_addr: string
  user_agent: string
  created_at: string
}

export interface RequestDetail extends RequestSummary {
  request_headers: Record<string, string>
  request_body: string
  response_headers: Record<string, string>
  response_body: string
}

export interface ReplayResponse {
  success: boolean
  request_id: string
  original_id: string
  status_code: number
  response_size: number
  replayed_at: string
}

export interface ListRequestsParams {
  method?: string
  path?: string
  limit?: number
  offset?: number
}

export const webhookTestingApi = {
  /**
   * List all tunnels for the authenticated user
   * Uses the gateway's authenticated endpoint to get user's tunnels
   */
  async listTunnels(tunnelServerUrl?: string): Promise<Tunnel[]> {
    try {
      // Use the authenticated tunnels API endpoint (filters by user)
      const response = await tunnelsApi.list(true) // Use JWT auth
      return response.tunnels.map(t => ({
        id: t.id,
        subdomain: t.subdomain,
        public_url: t.public_url,
        local_url: t.local_url,
        request_count: t.request_count || 0,
        created_at: t.created_at,
        last_active: t.last_active || ''
      }))
    } catch (error: any) {
      // Fallback to tunnel server API if gateway endpoint fails
      // This is for backward compatibility or if user is not authenticated
      if (tunnelServerUrl) {
        const response = await fetch(`${tunnelServerUrl}/api/tunnels`, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json'
          }
        })
        if (!response.ok) {
          throw new Error(`Failed to list tunnels: ${response.statusText}`)
        }
        const data = await response.json()
        return data.tunnels || []
      }
      throw error
    }
  },

  /**
   * List requests for a tunnel
   * @param tunnelServerUrl - URL of the tunnel server
   * @param tunnelId - Tunnel ID
   * @param params - Query parameters
   */
  async listRequests(
    tunnelServerUrl: string,
    tunnelId: string,
    params: ListRequestsParams = {}
  ): Promise<{ requests: RequestSummary[]; count: number; total: number; limit: number; offset: number }> {
    const queryParams = new URLSearchParams()
    if (params.method) queryParams.append('method', params.method)
    if (params.path) queryParams.append('path', params.path)
    if (params.limit) queryParams.append('limit', params.limit.toString())
    if (params.offset) queryParams.append('offset', params.offset.toString())

    const response = await fetch(
      `${tunnelServerUrl}/api/tunnels/${tunnelId}/requests?${queryParams.toString()}`,
      {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      }
    )
    if (!response.ok) {
      throw new Error(`Failed to list requests: ${response.statusText}`)
    }
    return await response.json()
  },

  /**
   * Get detailed request information
   * @param tunnelServerUrl - URL of the tunnel server
   * @param tunnelId - Tunnel ID
   * @param requestId - Request ID
   */
  async getRequest(
    tunnelServerUrl: string,
    tunnelId: string,
    requestId: string
  ): Promise<RequestDetail> {
    const response = await fetch(
      `${tunnelServerUrl}/api/tunnels/${tunnelId}/requests/${requestId}`,
      {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      }
    )
    if (!response.ok) {
      throw new Error(`Failed to get request: ${response.statusText}`)
    }
    return await response.json()
  },

  async replayRequest(
    tunnelServerUrl: string,
    tunnelId: string,
    requestId: string
  ): Promise<ReplayResponse> {
    const response = await fetch(
      `${tunnelServerUrl}/api/tunnels/${tunnelId}/requests/${requestId}/replay`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      }
    )
    if (!response.ok) {
      let message = response.statusText
      try {
        const body = await response.text()
        if (body) {
          const parsed = JSON.parse(body)
          if (typeof parsed === 'object' && parsed.message) message = parsed.message
          else if (typeof body === 'string' && body.length < 200) message = body
        }
      } catch (_) {}
      if (response.status === 404) {
        message = 'Request not found or tunnel disconnected. Ensure the tunnel is running and request logging is enabled.'
      }
      throw new Error(message)
    }
    return await response.json()
  }
}

