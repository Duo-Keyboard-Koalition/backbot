/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_TAILSCALE_AUTH_KEY: string
  readonly VITE_TAILA2A_URL: string
  readonly VITE_TAILFS_URL: string
  readonly VITE_DEBUG: string
  readonly VITE_REFRESH_INTERVAL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
