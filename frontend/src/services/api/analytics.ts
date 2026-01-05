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

export const analyticsApi = {
  /**
   * Get usage statistics
   */
  async getUsageStats(startTime?: string, endTime?: string): Promise<UsageStats> {
    const params: Record<string, string> = {}
    if (startTime) params.start_time = startTime
    if (endTime) params.end_time = endTime

    const response = await apiClient.get<UsageStats>('/v1/analytics/usage', { params })
    return response.data
  },

  /**
   * Get request history
   */
  async getRequests(limit = 50, offset = 0): Promise<RequestsResponse> {
    const response = await apiClient.get<RequestsResponse>('/v1/analytics/requests', {
      params: { limit, offset },
    })
    return response.data
  },
}

