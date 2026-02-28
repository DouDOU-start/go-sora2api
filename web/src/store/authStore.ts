import { create } from 'zustand'

type Theme = 'light' | 'dark'

interface AuthState {
  token: string | null
  theme: Theme
  setToken: (token: string) => void
  logout: () => void
  toggleTheme: () => void
  initTheme: () => void
}

function getSystemTheme(): Theme {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function getSavedTheme(): Theme {
  const saved = localStorage.getItem('theme') as Theme | null
  return saved || getSystemTheme()
}

function applyTheme(theme: Theme) {
  document.documentElement.classList.toggle('dark', theme === 'dark')
  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta) meta.setAttribute('content', theme === 'dark' ? '#0a0a12' : '#f4f2ef')
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  theme: getSavedTheme(),

  setToken: (token: string) => {
    localStorage.setItem('token', token)
    set({ token })
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: null })
  },

  toggleTheme: () => {
    set((state) => {
      const next: Theme = state.theme === 'light' ? 'dark' : 'light'
      localStorage.setItem('theme', next)
      applyTheme(next)
      return { theme: next }
    })
  },

  initTheme: () => {
    const theme = getSavedTheme()
    applyTheme(theme)
    set({ theme })
  },
}))
