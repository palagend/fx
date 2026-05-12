import { ref, provide, inject, type InjectionKey, type Ref } from 'vue'

export interface ThemeContext {
  isDark: Ref<boolean>
  toggleTheme: () => void
}

export const ThemeKey: InjectionKey<ThemeContext> = Symbol('theme')

export function useTheme() {
  const isDark = ref(false)

  const loadTheme = () => {
    const savedTheme = localStorage.getItem('theme')
    if (savedTheme) {
      isDark.value = savedTheme === 'dark'
      document.documentElement.classList.toggle('dark', isDark.value)
    }
  }

  const toggleTheme = () => {
    isDark.value = !isDark.value
    document.documentElement.classList.toggle('dark', isDark.value)
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
  }

  const provideTheme = () => {
    provide(ThemeKey, { isDark, toggleTheme })
  }

  return {
    isDark,
    toggleTheme,
    loadTheme,
    provideTheme
  }
}

export function injectTheme(): ThemeContext {
  const theme = inject(ThemeKey)
  if (!theme) {
    throw new Error('useTheme must be used within a component that provides theme')
  }
  return theme
}
