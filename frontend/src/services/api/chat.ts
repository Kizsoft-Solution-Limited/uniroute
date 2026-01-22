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

export interface StreamChunk {
  id?: string
  content: string
  done: boolean
  usage?: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
  error?: string
  provider?: string
}

export type StreamChunkCallback = (chunk: StreamChunk) => void
export type StreamErrorCallback = (error: Error) => void
export type StreamCompleteCallback = (usage?: StreamChunk['usage']) => void

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

  /**
   * Send a streaming chat completion request via Server-Sent Events (SSE)
   * Uses /auth/chat/stream for frontend (JWT auth) or /v1/chat/stream for API keys
   */
  async chatStream(
    data: ChatRequest,
    onChunk: StreamChunkCallback,
    onError?: StreamErrorCallback,
    onComplete?: StreamCompleteCallback,
    useJWT: boolean = true
  ): Promise<void> {
    const endpoint = useJWT ? '/auth/chat/stream' : '/v1/chat/stream'
    // Get token from localStorage (remember me) or sessionStorage (session only)
    let token = useJWT ? localStorage.getItem('auth_token') : null
    if (!token && useJWT) {
      token = sessionStorage.getItem('auth_token')
    }

    try {
      // Get base URL from apiClient
      const baseURL = apiClient.defaults.baseURL || 'http://localhost:8084'
      const response = await fetch(`${baseURL}${endpoint}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token && { Authorization: `Bearer ${token}` }),
        },
        credentials: 'include', // Include cookies for httpOnly cookies
        body: JSON.stringify(data),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Unknown error' }))
        throw new Error(errorData.error || `HTTP ${response.status}`)
      }

      if (!response.body) {
        throw new Error('Response body is null')
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()

        if (done) {
          break
        }

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || '' // Keep incomplete line in buffer

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6) // Remove "data: " prefix
            if (data.trim() === '') continue

            try {
              const chunk: StreamChunk = JSON.parse(data)
              onChunk(chunk)

              if (chunk.done) {
                if (chunk.error) {
                  onError?.(new Error(chunk.error))
                } else {
                  onComplete?.(chunk.usage)
                }
                return
              }
            } catch (err) {
              console.error('Failed to parse SSE chunk:', err, data)
            }
          }
        }
      }
    } catch (error) {
      onError?.(error instanceof Error ? error : new Error(String(error)))
      throw error
    }
  },

  /**
   * Send a streaming chat completion request via WebSocket
   * Uses /auth/chat/ws for frontend (JWT auth) or /v1/chat/ws for API keys
   */
  async chatStreamWebSocket(
    data: ChatRequest,
    onChunk: StreamChunkCallback,
    onError?: StreamErrorCallback,
    onComplete?: StreamCompleteCallback,
    useJWT: boolean = true
  ): Promise<void> {
    const endpoint = useJWT ? '/auth/chat/ws' : '/v1/chat/ws'
    // Get token from localStorage (remember me) or sessionStorage (session only)
    let token = useJWT ? localStorage.getItem('auth_token') : null
    if (!token && useJWT) {
      token = sessionStorage.getItem('auth_token')
    }

    // Get base URL from apiClient
    const baseURL = apiClient.defaults.baseURL || 'http://localhost:8084'
    
    // Convert http/https to ws/wss
    const wsProtocol = baseURL.startsWith('https') ? 'wss' : 'ws'
    const baseURLWithoutProtocol = baseURL.replace(/^https?:\/\//, '')
    
    // For WebSocket, we need to include token in query or use subprotocol
    // Since middleware should handle auth, we can pass token in query as fallback
    const wsURL = `${wsProtocol}://${baseURLWithoutProtocol}${endpoint}${token ? `?token=${encodeURIComponent(token)}` : ''}`

    return new Promise((resolve, reject) => {
      try {
        const ws = new WebSocket(wsURL)

        ws.onopen = () => {
          // Send chat request
          ws.send(JSON.stringify({
            type: 'request',
            request: data
          }))
        }

        ws.onmessage = (event) => {
          try {
            // Handle binary ping/pong frames
            if (event.data instanceof Blob || event.data instanceof ArrayBuffer) {
              // Binary message (ping/pong) - ignore
              return
            }

            const response: StreamChunk = JSON.parse(event.data)

            // Check for error
            if (response.error) {
              onError?.(new Error(response.error))
              ws.close()
              reject(new Error(response.error))
              return
            }

            // Process chunk
            onChunk({
              id: response.id,
              content: response.content || '',
              done: response.done || false,
              usage: response.usage,
              error: response.error,
              provider: response.provider
            })

            if (response.done) {
              onComplete?.(response.usage)
              ws.close()
              resolve()
            }
          } catch (err) {
            console.error('Failed to parse WebSocket message:', err, event.data)
            onError?.(err instanceof Error ? err : new Error(String(err)))
            ws.close()
            reject(err)
          }
        }

        ws.onerror = (error) => {
          const errorMsg = 'WebSocket connection error'
          onError?.(new Error(errorMsg))
          ws.close()
          reject(new Error(errorMsg))
        }

        ws.onclose = (event) => {
          if (event.code !== 1000 && event.code !== 1001) {
            // Abnormal closure
            const errorMsg = `WebSocket closed unexpectedly: ${event.code} ${event.reason || ''}`
            onError?.(new Error(errorMsg))
            reject(new Error(errorMsg))
          } else {
            // Normal closure
            resolve()
          }
        }
      } catch (error) {
        onError?.(error instanceof Error ? error : new Error(String(error)))
        reject(error)
      }
    })
  },
}

