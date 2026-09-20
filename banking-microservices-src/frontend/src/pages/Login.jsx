import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth.jsx'

export default function Login() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function onSubmit(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      await login(email, password)
      navigate('/')
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="auth-page">
      <form className="card auth-card" onSubmit={onSubmit}>
        <h1>🏦 Banking Platform</h1>
        <h2>Sign in</h2>
        {error && <div className="error">{error}</div>}
        <label>Email<input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required /></label>
        <label>Password<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required /></label>
        <button disabled={busy} type="submit">{busy ? 'Signing in…' : 'Sign in'}</button>
        <p className="muted">No account? <Link to="/register">Create one</Link></p>
      </form>
    </div>
  )
}
