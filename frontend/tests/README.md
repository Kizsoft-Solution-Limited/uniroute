# Frontend Tests

This directory contains all frontend tests organized by type.

## Structure

```
tests/
├── unit/              # Unit tests (components, stores, utils)
├── integration/       # Integration tests (API calls, router)
├── e2e/              # End-to-end tests (Playwright)
├── fixtures/         # Test data and fixtures
└── testutil/         # Test utilities and helpers
```

## Running Tests

### Unit Tests
```bash
npm run test:unit
```

### Integration Tests
```bash
npm run test:integration
```

### E2E Tests
```bash
npm run test:e2e
```

### All Tests
```bash
npm run test
```

## Test Types

### Unit Tests
- Test individual components in isolation
- Test stores (Pinia) with mocked dependencies
- Test utility functions
- Fast execution, no external dependencies

### Integration Tests
- Test component interactions
- Test API service integration
- Test router navigation
- May require mocked API responses

### E2E Tests
- Test full user flows
- Test in real browser environment
- Test against running application
- Slower but most realistic

## Writing Tests

### Component Tests
```typescript
import { describe, it, expect } from 'vitest'
import { mountComponent } from '../testutil/helpers'
import MyComponent from '@/components/MyComponent.vue'

describe('MyComponent', () => {
  it('renders correctly', () => {
    const wrapper = mountComponent(MyComponent)
    expect(wrapper.text()).toContain('Hello')
  })
})
```

### Store Tests
```typescript
import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'

describe('AuthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('initializes with empty user', () => {
    const store = useAuthStore()
    expect(store.user).toBeNull()
  })
})
```

## Test Utilities

See `testutil/` directory for:
- `setup.ts` - Vitest configuration
- `mocks.ts` - Mock factories
- `helpers.ts` - Test helper functions

## Fixtures

See `fixtures/` directory for:
- `auth.ts` - Authentication test data
- `api.ts` - API response fixtures

## Best Practices

1. **Keep tests isolated** - Each test should be independent
2. **Use fixtures** - Reuse test data from `fixtures/`
3. **Mock external dependencies** - Use mocks for API calls, router, etc.
4. **Test behavior, not implementation** - Focus on what the component does, not how
5. **Use descriptive test names** - Clear test names help debugging
6. **Keep tests fast** - Unit tests should run quickly
7. **Test edge cases** - Don't just test happy paths

