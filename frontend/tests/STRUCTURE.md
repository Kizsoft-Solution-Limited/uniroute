# Frontend Test Structure

Quick reference for the frontend test directory structure.

## Directory Layout

```
frontend/tests/
├── unit/                    # Unit tests
│   ├── components/         # Component unit tests
│   ├── stores/             # Pinia store tests
│   ├── composables/        # Composable function tests
│   ├── utils/              # Utility function tests
│   └── services/           # Service unit tests
│
├── integration/            # Integration tests
│   ├── api/                # API integration tests
│   ├── router/             # Router integration tests
│   └── components/         # Component integration tests
│
├── e2e/                    # End-to-end tests (Playwright)
│   ├── auth/               # Authentication flows
│   ├── dashboard/          # Dashboard flows
│   └── settings/           # Settings flows
│
├── fixtures/               # Test data
│   ├── auth.ts             # Auth fixtures
│   ├── api.ts              # API fixtures
│   └── components.ts        # Component fixtures
│
└── testutil/              # Test utilities
    ├── setup.ts            # Vitest setup
    ├── mocks.ts            # Mock factories
    └── helpers.ts          # Helper functions
```

## File Naming

- Unit tests: `*.test.ts` or `*.test.tsx`
- Integration tests: `*.integration.test.ts`
- E2E tests: `*.e2e.test.ts` (Playwright)

## Example Locations

- Component test: `tests/unit/components/Button.test.ts`
- Store test: `tests/unit/stores/auth.test.ts`
- API integration: `tests/integration/api/auth.integration.test.ts`
- E2E flow: `tests/e2e/auth/login.e2e.test.ts`

