// Authentication context: holds the current session and exposes login/register/
// logout. Tokens are persisted in localStorage via api.js.
import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { api, clearTokens, getToken, setTokens } from './api.js'

const AuthCtx = createContext(null)

function decodeSubject(token) {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return { subject: payload.sub, email: payload.email, roles: payload.roles || [] }
  } catch {
    return null
  }
}

export function AuthProvider({ children }) {
  const [user, setUser] = useState(() => {
    const t = getToken()
    return t ? decodeSubject(t) : null
  })

  useEffect(() => {}, [])

  async function login(email, password) {
    const tokens = await api.post('/auth/login', { email, password })
    setTokens(tokens)
    setUser(decodeSubject(tokens.access_token))
  }

  async function register(email, password, fullName) {
    const res = await api.post('/auth/register', { email, password, full_name: fullName })
    setTokens(res.tokens)
    setUser(decodeSubject(res.tokens.access_token))
  }

  function logout() {
    clearTokens()
    setUser(null)
  }

  const value = useMemo(() => ({ user, login, register, logout }), [user])
  return <AuthCtx.Provider value={value}>{children}</AuthCtx.Provider>
}

export function useAuth() {
  return useContext(AuthCtx)
}

export function ProtectedRoute({ children }) {
  const { user } = useAuth()
  const location = useLocation()
  if (!user) return <Navigate to="/login" state={{ from: location }} replace />
  return children
}
