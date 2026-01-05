/**
 * API service for error logs (admin only)
 */

import { apiClient } from './client'

export interface ErrorLog {
  id: string
  user_id?: string
  error_type: 'exception' | 'message' | 'network' | 'server'
  message: string
  stack_trace?: string
  url?: string
  user_agent?: string
  ip_address?: string
  context?: Record<string, any>
  severity: 'error' | 'warning' | 'info'
  resolved: boolean
  created_at: string
}

export interface ErrorLogsResponse {
  errors: ErrorLog[]
  count: number
  limit: number
}

export interface ErrorLogFilters {
  user_id?: string
  error_type?: string
  severity?: string
  resolved?: boolean
  limit?: number
  offset?: number
}

export const errorsApi = {
  /**
   * Get error logs (admin only)
   */
  async getErrorLogs(filters?: ErrorLogFilters): Promise<ErrorLogsResponse> {
    const params = new URLSearchParams()
    
    if (filters?.user_id) params.append('user_id', filters.user_id)
    if (filters?.error_type) params.append('error_type', filters.error_type)
    if (filters?.severity) params.append('severity', filters.severity)
    if (filters?.resolved !== undefined) params.append('resolved', String(filters.resolved))
    if (filters?.limit) params.append('limit', String(filters.limit))
    if (filters?.offset) params.append('offset', String(filters.offset))

    const queryString = params.toString()
    const url = `/admin/errors${queryString ? `?${queryString}` : ''}`
    
    const response = await apiClient.get<ErrorLogsResponse>(url)
    return response.data
  },

  /**
   * Mark error as resolved (admin only)
   */
  async markResolved(errorId: string): Promise<{ message: string; id: string }> {
    const response = await apiClient.patch<{ message: string; id: string }>(`/admin/errors/${errorId}/resolve`)
    return response.data
  },
}


