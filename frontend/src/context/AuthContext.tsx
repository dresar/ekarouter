import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { UserProfile, AuthResponse } from '../types/api.ts'
import { api } from '../api/client.ts'
import {
  getStoredToken,
  setStoredToken,
  removeStoredToken,
  getStoredUser,
  setStoredUser,
  removeStoredUser,
} from '../utils/storage.ts'

interface AuthContextType {
  user: UserProfile | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<UserProfile | null>(getStoredUser())
  const [token, setToken] = useState<string | null>(getStoredToken())
  const [isLoading, setIsLoading] = useState<boolean>(true)

  useEffect(() => {
    async function verifySession() {
      const activeToken = getStoredToken()
      if (!activeToken) {
        setUser(null)
        setToken(null)
        setIsLoading(false)
        return
      }

      try {
        interface MeResponse {
          user?: string
          username?: string
          role?: string
        }
        const res = await api.get<MeResponse>('/api/auth/me')
        const stored = getStoredUser()
        const resolvedUsername = res.username || res.user || stored?.username || 'admin'
        const profile: UserProfile = {
          username: resolvedUsername,
          role: res.role || stored?.role || 'admin',
          is_default_password: stored ? stored.is_default_password : true,
        }
        setUser(profile)
        setStoredUser(profile)
      } catch {
        const stored = getStoredUser()
        if (stored) {
          setUser(stored)
        } else {
          removeStoredToken()
          removeStoredUser()
          setUser(null)
          setToken(null)
        }
      } finally {
        setIsLoading(false)
      }
    }

    verifySession()
  }, [])

  const login = async (username: string, password: string): Promise<void> => {
    const isDefault = username === 'admin' && (password === 'admin1234' || password === 'admin12345')
    const data = await api.post<AuthResponse>('/api/auth/login', { username, password })
    if (data.token) {
      setToken(data.token)
      setStoredToken(data.token)
    }
    const resolvedUser: UserProfile = {
      username: data.user?.username || username,
      role: data.user?.role || 'admin',
      is_default_password: isDefault,
    }
    setUser(resolvedUser)
    setStoredUser(resolvedUser)
  }

  const logout = async (): Promise<void> => {
    try {
      await api.post('/api/auth/logout')
    } catch {}
    removeStoredToken()
    removeStoredUser()
    setUser(null)
    setToken(null)
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated: !!token && !!user,
        isLoading,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}
