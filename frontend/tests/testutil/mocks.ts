/**
 * Test utilities for mocking services and components
 */

import { vi } from 'vitest'

/**
 * Mock Axios instance
 */
export function createMockAxios() {
  return {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: {
        use: vi.fn(),
        eject: vi.fn(),
      },
      response: {
        use: vi.fn(),
        eject: vi.fn(),
      },
    },
  }
}

/**
 * Mock Pinia store
 */
export function createMockStore() {
  return {
    state: {},
    getters: {},
    actions: {},
    $patch: vi.fn(),
    $reset: vi.fn(),
    $subscribe: vi.fn(),
  }
}

/**
 * Mock Vue Router
 */
export function createMockRouter() {
  return {
    push: vi.fn(),
    replace: vi.fn(),
    go: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    currentRoute: {
      value: {
        path: '/',
        name: 'home',
        params: {},
        query: {},
        meta: {},
      },
    },
  }
}

/**
 * Mock API response
 */
export function createMockApiResponse<T>(data: T, status = 200) {
  return {
    data,
    status,
    statusText: 'OK',
    headers: {},
    config: {} as any,
  }
}

/**
 * Mock API error response
 */
export function createMockApiError(message: string, status = 400) {
  return {
    response: {
      data: { error: message },
      status,
      statusText: 'Bad Request',
      headers: {},
    },
    message,
    isAxiosError: true,
  }
}

