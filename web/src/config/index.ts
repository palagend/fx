export const config = {
  mode: (import.meta.env.VITE_APP_MODE || 'frontend') as 'backend' | 'frontend',

  get isBackend(): boolean {
    return this.mode === 'backend'
  },

  get isFrontend(): boolean {
    return this.mode === 'frontend'
  }
}