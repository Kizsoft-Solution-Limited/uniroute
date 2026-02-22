import { apiClient } from './client'

export interface UsageStats {
  period: {
    start: string
    end: string
  }
  total_requests: number
  total_tokens: number
  total_cost: number
  average_latency_ms: number
  requests_by_provider: Record<string, number>
  requests_by_model: Record<string, number>
  cost_by_provider: Record<string, number>
}

export interface Request {
  id: string
  provider: string
  model: string
  input_tokens: number
  output_tokens: number
  total_tokens: number
  cost: number
  latency_ms: number
  status_code: number
  created_at: string
}

export interface RequestsResponse {
  requests: Request[]
  limit: number
  offset: number
  count: number
}

export interface CostEstimateRequest {
  model: string
  messages: Array<{
    role: 'system' | 'user' | 'assistant'
    content: string
  }>
}

export interface CostEstimateResponse {
  model: string
  estimates: Record<string, number> // provider -> cost
}

export interface LatencyStats {
  latency_stats: Record<string, {
    average_ms: number
    min_ms: number
    max_ms: number
    samples: number
  }>
}

export const analyticsApi = {
  /**
   * Get usage statistics
   * Uses /auth/analytics/usage for frontend (JWT auth) or /v1/analytics/usage for API keys
   */
  async getUsageStats(startTime?: string, endTime?: string, useJWT: boolean = true): Promise<UsageStats> {
    const endpoint = useJWT ? '/auth/analytics/usage' : '/v1/analytics/usage'
    const params: Record<string, string> = {}
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime

    const response = await apiClient.get<UsageStats>(endpoint, { params })
    return response.data
  },

  async getUsageStatsAdmin(startTime?: string, endTime?: string): Promise<UsageStats> {
    const params: Record<string, string> = {}
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime
    const response = await apiClient.get<UsageStats>('/admin/analytics/usage', { params })
    return response.data
  },

  /**
   * Get request history
   * Uses /auth/analytics/requests for frontend (JWT auth) or /v1/analytics/requests for API keys
   */
  async getRequests(limit = 50, offset = 0, useJWT: boolean = true): Promise<RequestsResponse> {
    const endpoint = useJWT ? '/auth/analytics/requests' : '/v1/analytics/requests'
    const response = await apiClient.get<RequestsResponse>(endpoint, {
      params: { limit, offset },
    })
    return response.data
  },

  /**
   * Estimate cost for a chat request.
   * Uses /auth/routing/estimate-cost when logged in (includes BYOK providers).
   */
  async estimateCost(data: CostEstimateRequest, useJWT: boolean = true): Promise<CostEstimateResponse> {
    const endpoint = useJWT ? '/auth/routing/estimate-cost' : '/v1/routing/estimate-cost'
    const response = await apiClient.post<CostEstimateResponse>(endpoint, data)
    return response.data
  },

  /**
   * Get latency statistics for all providers.
   * Uses /auth/routing/latency when logged in.
   */
  async getLatencyStats(useJWT: boolean = true): Promise<LatencyStats> {
    const endpoint = useJWT ? '/auth/routing/latency' : '/v1/routing/latency'
    const response = await apiClient.get<LatencyStats>(endpoint)
    return response.data
  },
}

