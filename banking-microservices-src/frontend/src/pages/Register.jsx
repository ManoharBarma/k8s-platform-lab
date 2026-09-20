import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth.jsx'

export default function Register() {
  const { register } = useAuth()
  const navigate = useNavigate()
  const [fullName, setFullName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function onSubmit(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      await register(email, password, fullName)
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
        <h2>Create account</h2>
        {error && <div className="error">{error}</div>}
        <label>Full name<input value={fullName} onChange={(e) => setFullName(e.target.value)} /></label>
        <label>Email<input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required /></label>
        <label>Password<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} /></label>
        <button disabled={busy} type="submit">{busy ? 'Creating…' : 'Create account'}</button>
        <p className="muted">Already have an account? <Link to="/login">Sign in</Link></p>
      </form>
    </div>
  )
}
