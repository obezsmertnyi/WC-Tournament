/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Build version, set from VITE_APP_VERSION at build time ("dev" locally). */
  readonly VITE_APP_VERSION?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
