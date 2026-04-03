import { createContext, useState, useEffect } from 'react'
import { getCurrentUser, login as loginAPI, register as registerAPI } from '../api/auth'

export const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    if (token) {
      getCurrentUser()
        .then(setUser)
        .catch(() => {
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
        })
        .finally(() => setLoading(false))
    } else {
      setLoading(false)
    }
  }, [])

  const loginUser = async (credentials) => {
    const data = await loginAPI(credentials)
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    const profile = await getCurrentUser()
    setUser(profile)
    return profile
  }

  const registerUser = async (credentials) => {
    const data = await registerAPI(credentials)
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    const profile = await getCurrentUser()
    setUser(profile)
    return profile
  }

  const logoutUser = () => {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    setUser(null)
  }

  const updateUser = (updated) => {
    setUser((prev) => ({ ...prev, ...updated }))
  }

  return (
    <AuthContext.Provider value={{ user, loading, loginUser, registerUser, logoutUser, updateUser }}>
      {children}
    </AuthContext.Provider>
  )
}
