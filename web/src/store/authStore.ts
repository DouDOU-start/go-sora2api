import { create } from 'zustand'

interface AuthState {
  token: string | null
  theme: 'light' | 'dark'
  setToken: (token: string) => void
  logout: () => void
  toggleTheme: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  theme: (localStorage.getItem('theme') as 'light' | 'dark') || 'light',

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
      const next = state.theme === 'light' ? 'dark' : 'light'
      localStorage.setItem('theme', next)
      document.documentElement.classList.toggle('dark', next === 'dark')
      return { theme: next }
    })
  },
}))
