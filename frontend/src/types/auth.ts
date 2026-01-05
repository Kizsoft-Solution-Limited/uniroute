export interface User {
  id: string
  email: string
  name?: string
  roles: ('admin' | 'user' | 'guest')[] // Array of roles: ['user'], ['admin'], or ['user', 'admin']
  email_verified?: boolean
  permissions?: string[]
  created_at: string
  updated_at?: string
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
  token?: string // Optional - registration doesn't return token (email verification required)
  user: User
  message?: string // Optional message (e.g., "Registration successful. Please check your email...")
}

export interface RefreshTokenResponse {
  token: string
}

