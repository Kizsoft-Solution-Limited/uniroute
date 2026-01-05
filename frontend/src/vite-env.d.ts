/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_TUNNEL_SERVER_URL?: string
  readonly VITE_API_URL?: string
  readonly VITE_API_BASE_URL?: string
  readonly DEV?: boolean
  readonly PROD?: boolean
  readonly MODE?: string
  // Add more env variables as needed
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

