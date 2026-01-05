/**
 * Authentication test fixtures
 */

export const mockUser = {
  id: '123e4567-e89b-12d3-a456-426614174000',
  email: 'test@example.com',
  name: 'Test User',
  roles: ['user'] as string[],
  emailVerified: true,
  createdAt: '2024-01-01T00:00:00Z',
}

export const mockAdminUser = {
  ...mockUser,
  id: '123e4567-e89b-12d3-a456-426614174001',
  email: 'admin@example.com',
  name: 'Admin User',
  roles: ['user', 'admin'] as string[],
}

export const mockAuthResponse = {
  token: 'mock-jwt-token',
  user: mockUser,
}

export const mockLoginRequest = {
  email: 'test@example.com',
  password: 'password123',
  rememberMe: false,
}

export const mockRegisterRequest = {
  email: 'test@example.com',
  password: 'password123',
  name: 'Test User',
}

export const mockPasswordResetRequest = {
  email: 'test@example.com',
}

export const mockPasswordResetConfirm = {
  token: 'reset-token',
  password: 'newpassword123',
}

