/**
 * API test fixtures
 */

export const mockApiKey = {
  id: '123e4567-e89b-12d3-a456-426614174000',
  name: 'Test API Key',
  key: 'ur_testkey123456789012345678901234567890',
  rateLimitPerMinute: 60,
  rateLimitPerDay: 10000,
  createdAt: '2024-01-01T00:00:00Z',
  expiresAt: null,
  isActive: true,
}

export const mockProviderKey = {
  id: '123e4567-e89b-12d3-a456-426614174001',
  provider: 'openai',
  name: 'OpenAI Key',
  createdAt: '2024-01-01T00:00:00Z',
  lastUsedAt: null,
}

export const mockErrorLog = {
  id: '123e4567-e89b-12d3-a456-426614174002',
  message: 'Test error message',
  stack: 'Error: Test error\n    at test.js:1:1',
  url: 'http://localhost:3000/test',
  userAgent: 'Mozilla/5.0',
  userId: '123e4567-e89b-12d3-a456-426614174000',
  resolved: false,
  createdAt: '2024-01-01T00:00:00Z',
}

export const mockApiResponse = {
  success: true,
  data: {},
}

export const mockApiError = {
  error: 'Test error',
  message: 'Something went wrong',
  code: 'TEST_ERROR',
}

