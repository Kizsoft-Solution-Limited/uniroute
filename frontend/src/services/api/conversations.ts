import { apiClient } from './client'

export interface Conversation {
  id: string
  user_id: string
  title: string | null
  model: string | null
  created_at: string
  updated_at: string
}

export interface ConversationMessage {
  id: string
  conversation_id: string
  role: 'system' | 'user' | 'assistant'
  content: string | any[] // string or ContentPart[]
  metadata: Record<string, any> | null
  created_at: string
}

export interface ConversationWithMessages {
  conversation: Conversation
  messages: ConversationMessage[]
}

export interface CreateConversationRequest {
  title?: string
  model?: string
}

export interface UpdateConversationRequest {
  title?: string
  model?: string
}

export const conversationsApi = {
  /**
   * List all conversations for the current user
   */
  async listConversations(limit: number = 50, offset: number = 0): Promise<Conversation[]> {
    const response = await apiClient.get<Conversation[]>('/auth/conversations', {
      params: { limit, offset }
    })
    return response.data
  },

  /**
   * Get a conversation with its messages
   */
  async getConversation(id: string): Promise<ConversationWithMessages> {
    const response = await apiClient.get<ConversationWithMessages>(`/auth/conversations/${id}`)
    return response.data
  },

  /**
   * Create a new conversation
   */
  async createConversation(data: CreateConversationRequest): Promise<Conversation> {
    const response = await apiClient.post<Conversation>('/auth/conversations', data)
    return response.data
  },

  /**
   * Update a conversation (title, model)
   */
  async updateConversation(id: string, data: UpdateConversationRequest): Promise<void> {
    await apiClient.put(`/auth/conversations/${id}`, data)
  },

  /**
   * Delete a conversation
   */
  async deleteConversation(id: string): Promise<void> {
    await apiClient.delete(`/auth/conversations/${id}`)
  },
}
