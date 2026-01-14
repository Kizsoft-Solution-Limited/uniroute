/**
 * API service for routing strategy management
 */

import { apiClient } from './client'

export interface RoutingStrategy {
  strategy: string
  is_locked?: boolean
  available_strategies: string[]
}

export interface UserRoutingStrategy {
  strategy: string
  user_strategy: string | null // NULL if using default
  default_strategy: string
  is_locked: boolean
  available_strategies: string[]
}

export interface SetRoutingStrategyRequest {
  strategy: string
}

export interface SetRoutingStrategyResponse {
  message: string
  strategy: string
}

export interface RoutingStrategyLockResponse {
  message: string
  locked: boolean
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

  /**
   * Set routing strategy lock (admin only)
   */
  async setStrategyLock(locked: boolean): Promise<RoutingStrategyLockResponse> {
    const response = await apiClient.post<RoutingStrategyLockResponse>('/admin/routing/strategy/lock', { locked })
    return response.data
  },

  /**
   * Get user's routing strategy (user-facing)
   */
  async getUserStrategy(): Promise<UserRoutingStrategy> {
    const response = await apiClient.get<UserRoutingStrategy>('/auth/routing/strategy')
    return response.data
  },

  /**
   * Set user's routing strategy (user-facing)
   */
  async setUserStrategy(data: SetRoutingStrategyRequest): Promise<SetRoutingStrategyResponse> {
    const response = await apiClient.put<SetRoutingStrategyResponse>('/auth/routing/strategy', data)
    return response.data
  },

  /**
   * Clear user's routing strategy (use default)
   */
  async clearUserStrategy(): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>('/auth/routing/strategy')
    return response.data
  },
}

