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

export interface CustomRoutingRule {
  id?: number
  name: string
  condition_type: 'model' | 'cost_threshold' | 'latency_threshold'
  condition_value: {
    model?: string
    max_cost?: number
    max_latency_ms?: number
  }
  provider_name: string
  priority: number
  enabled: boolean
  description?: string
  user_id?: string | null
  created_at?: string
  updated_at?: string
}

export interface CustomRulesResponse {
  rules: CustomRoutingRule[]
  count: number
  user_specific?: boolean
}

export interface SetCustomRulesRequest {
  rules: CustomRoutingRule[]
}

export interface SetCustomRulesResponse {
  message: string
  count: number
  user_specific?: boolean
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

  /**
   * Get custom routing rules (user-facing)
   */
  async getUserCustomRules(): Promise<CustomRulesResponse> {
    const response = await apiClient.get<CustomRulesResponse>('/auth/routing/custom-rules')
    return response.data
  },

  /**
   * Set custom routing rules (user-facing)
   */
  async setUserCustomRules(data: SetCustomRulesRequest): Promise<SetCustomRulesResponse> {
    const response = await apiClient.post<SetCustomRulesResponse>('/auth/routing/custom-rules', data)
    return response.data
  },

  /**
   * Get custom routing rules (admin - global rules)
   */
  async getAdminCustomRules(): Promise<CustomRulesResponse> {
    const response = await apiClient.get<CustomRulesResponse>('/admin/routing/custom-rules')
    return response.data
  },

  /**
   * Set custom routing rules (admin - global rules)
   */
  async setAdminCustomRules(data: SetCustomRulesRequest): Promise<SetCustomRulesResponse> {
    const response = await apiClient.post<SetCustomRulesResponse>('/admin/routing/custom-rules', data)
    return response.data
  },
}

