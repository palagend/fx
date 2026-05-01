interface ImportMetaEnv {
  VITE_APP_MODE?: string
}

declare module 'vue' {
  interface ImportMeta {
    readonly env: ImportMetaEnv
  }
}

export const config = {
  mode: ((import.meta.env.VITE_APP_MODE as string | undefined) || 'frontend') as 'backend' | 'frontend',

  get isBackend(): boolean {
    return this.mode === 'backend'
  },

  get isFrontend(): boolean {
    return this.mode === 'frontend'
  }
}