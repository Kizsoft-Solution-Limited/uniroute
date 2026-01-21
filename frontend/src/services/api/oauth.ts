/**
 * API service for OAuth authentication
 */

import { apiClient } from './client'

export interface OAuthAuthResponse {
  auth_url: string
  state: string
}

export const oauthApi = {
  /**
   * Get Google OAuth authorization URL
   */
  async getGoogleAuthURL(): Promise<OAuthAuthResponse> {
    const response = await apiClient.get<OAuthAuthResponse>('/auth/google')
    return response.data
  },

  /**
   * Get X (Twitter) OAuth authorization URL
   */
  async getXAuthURL(): Promise<OAuthAuthResponse> {
    const response = await apiClient.get<OAuthAuthResponse>('/auth/x')
    return response.data
  },
}
