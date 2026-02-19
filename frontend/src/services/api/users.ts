import { apiClient } from './client'

export interface User {
  id: string
  email: string
  name: string
  email_verified: boolean
  roles: string[]
  created_at: string
}

export interface ListUsersResponse {
  users: User[]
  limit: number
  offset: number
  count: number
  total: number
}

export interface UpdateUserRolesRequest {
  roles: string[]
}

export interface UpdateUserRolesResponse {
  message: string
  user: User
}

export const usersApi = {
  /**
   * List all users (admin only)
   */
  async list(limit = 50, offset = 0): Promise<ListUsersResponse> {
    const response = await apiClient.get<ListUsersResponse>('/admin/users', {
      params: { limit, offset },
    })
    return response.data
  },

  /**
   * Update user roles (admin only)
   */
  async updateRoles(userId: string, roles: string[]): Promise<UpdateUserRolesResponse> {
    const response = await apiClient.put<UpdateUserRolesResponse>(`/admin/users/${userId}/roles`, {
      roles,
    })
    return response.data
  },

  /**
   * Delete a single user and all related data (admin only)
   */
  async delete(userId: string): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`/admin/users/${userId}`)
    return response.data
  },

  /**
   * Delete multiple users and all their related data (admin only)
   */
  async deleteMany(userIds: string[]): Promise<{ message: string; deleted: number; failed: string[]; error?: string }> {
    const response = await apiClient.post<{ message: string; deleted: number; failed: string[]; error?: string }>(
      '/admin/users/delete',
      { user_ids: userIds }
    )
    return response.data
  },
}

