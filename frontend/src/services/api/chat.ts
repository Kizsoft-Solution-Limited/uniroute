import { apiClient } from './client'

export interface ContentPart {
  type: 'text' | 'image_url' | 'audio_url'
  text?: string
  image_url?: {
    url: string // Can be data URL (base64) or HTTP URL
  }
  audio_url?: {
    url: string // Can be data URL (base64) or HTTP URL
  }
}

export interface Message {
  role: 'system' | 'user' | 'assistant'
  content: string | ContentPart[] // string for text-only, ContentPart[] for multimodal
}

export interface ChatRequest {
  model: string
  messages: Message[]
  temperature?: number
  max_tokens?: number
  conversation_id?: string // Optional: save to conversation
}

export interface ChatResponse {
  id: string
  model: string
  provider: string
  choices: Array<{
    message: Message
  }>
  usage: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
  cost?: number
  latency_ms?: number
}

export const chatApi = {
  /**
   * Send a chat completion request
   * Uses /auth/chat for frontend (JWT auth) or /v1/chat for API keys
   */
  async chat(data: ChatRequest, useJWT: boolean = true): Promise<ChatResponse> {
    // Frontend users use JWT auth endpoint, API users use API key endpoint
    const endpoint = useJWT ? '/auth/chat' : '/v1/chat'
    const response = await apiClient.post<ChatResponse>(endpoint, data)
    return response.data
  },
}

