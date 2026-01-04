export interface User {
  id: string
  email: string
  name?: string
  role: 'admin' | 'user' | 'guest'
  permissions?: string[]
  created_at: string
  updated_at: string
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterData {
  email: string
  password: string
  name?: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface RefreshTokenResponse {
  token: string
}

