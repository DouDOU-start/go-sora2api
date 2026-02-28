import { create } from 'zustand'

type Theme = 'light' | 'dark'
export type UserRole = 'admin' | 'viewer'

interface AuthState {
  token: string | null
  role: UserRole | null
  theme: Theme
  setToken: (token: string, role?: UserRole) => void
  logout: () => void
  isAdmin: () => boolean
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

export const useAuthStore = create<AuthState>((set, get) => ({
  token: localStorage.getItem('token'),
  role: (localStorage.getItem('role') as UserRole) || null,
  theme: getSavedTheme(),

  setToken: (token: string, role?: UserRole) => {
    localStorage.setItem('token', token)
    const r = role || 'admin'
    localStorage.setItem('role', r)
    set({ token, role: r })
  },

  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('role')
    set({ token: null, role: null })
  },

  isAdmin: () => get().role === 'admin',

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
