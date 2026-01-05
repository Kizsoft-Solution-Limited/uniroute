import { apiClient } from './client'

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
   * List all tunnels
   * @param tunnelServerUrl - URL of the tunnel server (e.g., http://localhost:8080)
   */
  async listTunnels(tunnelServerUrl: string): Promise<Tunnel[]> {
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
  ): Promise<{ requests: RequestSummary[]; count: number; limit: number; offset: number }> {
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

  /**
   * Replay a request
   * @param tunnelServerUrl - URL of the tunnel server
   * @param tunnelId - Tunnel ID
   * @param requestId - Request ID to replay
   */
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
      throw new Error(`Failed to replay request: ${response.statusText}`)
    }
    return await response.json()
  }
}

