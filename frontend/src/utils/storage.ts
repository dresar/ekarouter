import { UserProfile } from '../types/api.ts'

const TOKEN_KEY = 'ekarouter_token'
const USER_KEY = 'ekarouter_user'
const THEME_KEY = 'ekarouter_theme'

export function getStoredToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export function setStoredToken(token: string): void {
  try {
    localStorage.setItem(TOKEN_KEY, token)
  } catch {}
}

export function removeStoredToken(): void {
  try {
    localStorage.removeItem(TOKEN_KEY)
  } catch {}
}

export function getStoredUser(): UserProfile | null {
  try {
    const data = localStorage.getItem(USER_KEY)
    return data ? JSON.parse(data) : null
  } catch {
    return null
  }
}

export function setStoredUser(user: UserProfile): void {
  try {
    localStorage.setItem(USER_KEY, JSON.stringify(user))
  } catch {}
}

export function removeStoredUser(): void {
  try {
    localStorage.removeItem(USER_KEY)
  } catch {}
}

export function getStoredTheme(): 'dark' | 'light' | 'system' {
  try {
    const theme = localStorage.getItem(THEME_KEY)
    if (theme === 'dark' || theme === 'light' || theme === 'system') {
      return theme
    }
  } catch {}
  return 'dark'
}

export function setStoredTheme(theme: 'dark' | 'light' | 'system'): void {
  try {
    localStorage.setItem(THEME_KEY, theme)
  } catch {}
}
