/**
 * API service for routing strategy management (admin only)
 */

import { apiClient } from './client'

export interface RoutingStrategy {
  strategy: string
  available_strategies: string[]
}

export interface SetRoutingStrategyRequest {
  strategy: string
}

export interface SetRoutingStrategyResponse {
  message: string
  strategy: string
}

export const routingApi = {
  /**
   * Get current routing strategy (admin only)
   */
  async getStrategy(): Promise<RoutingStrategy> {
    const response = await apiClient.get<RoutingStrategy>('/admin/routing/strategy')
    return response.data
  },

  /**
   * Set routing strategy (admin only)
   */
  async setStrategy(data: SetRoutingStrategyRequest): Promise<SetRoutingStrategyResponse> {
    const response = await apiClient.post<SetRoutingStrategyResponse>('/admin/routing/strategy', data)
    return response.data
  },
}

