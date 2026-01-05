/**
 * Test helper functions
 */

import { mount } from '@vue/test-utils'
import type { Component } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'

/**
 * Create a test Pinia instance
 */
export function createTestPinia() {
  const pinia = createPinia()
  setActivePinia(pinia)
  return pinia
}

/**
 * Create a test router
 */
export function createTestRouter(routes: any[] = []) {
  return createRouter({
    history: createWebHistory(),
    routes: routes.length > 0 ? routes : [
      { path: '/', name: 'home', component: { template: '<div>Home</div>' } },
    ],
  })
}

/**
 * Mount component with default test setup
 */
export function mountComponent<T extends Component>(
  component: T,
  options: any = {}
) {
  const pinia = createTestPinia()
  const router = createTestRouter()

  return mount(component, {
    global: {
      plugins: [pinia, router],
    },
    ...options,
  })
}

/**
 * Wait for next tick
 */
export async function waitForNextTick() {
  return new Promise((resolve) => {
    setTimeout(resolve, 0)
  })
}

/**
 * Wait for a specific condition
 */
export async function waitFor(
  condition: () => boolean,
  timeout = 5000,
  interval = 100
): Promise<void> {
  const startTime = Date.now()
  
  while (Date.now() - startTime < timeout) {
    if (condition()) {
      return
    }
    await new Promise((resolve) => setTimeout(resolve, interval))
  }
  
  throw new Error('Timeout waiting for condition')
}

