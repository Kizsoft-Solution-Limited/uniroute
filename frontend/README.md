# UniRoute Frontend

Vue 3 + TypeScript frontend for UniRoute AI Gateway.

## Environment Variables

Environment variables are configured in `.env` files. Vite automatically loads these files:

- `.env` - Default environment file (not committed to git)
- `.env.local` - Local overrides (not committed to git)
- `.env.[mode]` - Mode-specific (e.g., `.env.production`)
- `.env.[mode].local` - Mode-specific local overrides

### Setup

1. Copy the example file:
   ```bash
   cp .env.example .env
   ```

2. Update `.env` with your values:
   ```env
   # API Configuration
   VITE_API_BASE_URL=http://localhost:8084
   
   # Tunnel Server Configuration
   VITE_TUNNEL_SERVER_URL=http://localhost:8080
   ```

### Available Variables

- `VITE_API_BASE_URL` - Backend API base URL (default: `http://localhost:8084` in dev, `https://api.uniroute.co` in prod)
- `VITE_TUNNEL_SERVER_URL` - Tunnel server URL (default: `http://localhost:8080` in dev)
- `VITE_API_URL` - Legacy API URL (optional, kept for compatibility)

**Note:** All Vite environment variables must be prefixed with `VITE_` to be exposed to the client-side code.

### Development vs Production

- **Development** (`npm run dev`): Uses `http://localhost:8084` by default
- **Production** (`npm run build`): Uses `https://api.uniroute.co` by default

You can override these by setting the environment variables in your `.env` file.

## Development

```bash
# Install dependencies
npm install

# Start dev server only
npm run dev

# Start dev server and tunnel (requires UniRoute CLI in PATH)
npm run dev:tunnel

# Build for production
npm run build

# Preview production build
npm run preview
```

## Testing

```bash
# Run unit tests
npm run test:unit

# Run E2E tests
npm run test:e2e

# Run all tests
npm test
```

## Project Structure

```
frontend/
├── src/
│   ├── assets/          # Static assets
│   ├── components/      # Reusable Vue components
│   ├── composables/     # Vue composables
│   ├── layouts/         # Layout components
│   ├── router/          # Vue Router configuration
│   ├── services/        # API services
│   ├── stores/          # Pinia stores
│   ├── types/           # TypeScript types
│   ├── utils/           # Utility functions
│   └── views/           # Page components
├── .env                 # Environment variables (not in git)
├── .env.example         # Example environment file
└── vite.config.ts       # Vite configuration
```

## Tech Stack

- **Vue 3** - Progressive JavaScript framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Tailwind CSS** - Utility-first CSS framework
- **Pinia** - State management
- **Vue Router** - Routing
- **Axios** - HTTP client
- **VeeValidate + Yup** - Form validation
- **Vitest** - Unit testing
- **Playwright** - E2E testing
