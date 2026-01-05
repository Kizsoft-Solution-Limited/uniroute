/**
 * API service for email configuration and testing (admin only)
 */

import { apiClient } from './client'

export interface EmailConfig {
  configured: boolean
  status: 'configured' | 'not configured'
  smtp?: {
    host?: string
    port?: number
    configured?: boolean
  }
  note?: string
  troubleshooting?: {
    check_mailtrap?: string
    check_credentials?: string
    check_port?: string
    check_host?: string
  }
  required_env_vars?: string[]
}

export interface TestEmailRequest {
  to: string
  subject?: string
  message?: string
}

export interface TestEmailResponse {
  message: string
  to: string
  note?: string
}

export const emailApi = {
  /**
   * Get email configuration status (admin only)
   */
  async getConfig(): Promise<EmailConfig> {
    const response = await apiClient.get<EmailConfig>('/admin/email/config')
    return response.data
  },

  /**
   * Send a test email (admin only)
   */
  async testEmail(data: TestEmailRequest): Promise<TestEmailResponse> {
    const response = await apiClient.post<TestEmailResponse>('/admin/email/test', data)
    return response.data
  },
}

