import { useEffect, useState } from 'react'
import { api } from '../api.js'
import { useAuth } from '../auth.jsx'

export default function Dashboard() {
  const { user } = useAuth()
  const [accounts, setAccounts] = useState([])
  const [error, setError] = useState('')

  useEffect(() => {
    api.get('/accounts')
      .then((d) => setAccounts(d.items || []))
      .catch((e) => setError(e.message))
  }, [])

  return (
    <div>
      <h1>Welcome back</h1>
      <p className="muted">{user?.email}</p>
      {error && <div className="error">{error}</div>}
      <div className="stat-row">
        <div className="card stat">
          <div className="stat-label">Accounts</div>
          <div className="stat-value">{accounts.length}</div>
        </div>
        <div className="card stat">
          <div className="stat-label">Roles</div>
          <div className="stat-value">{(user?.roles || []).join(', ') || 'customer'}</div>
        </div>
      </div>
      <h2>Your accounts</h2>
      {accounts.length === 0 ? (
        <p className="muted">No accounts yet. Open one from the Accounts page.</p>
      ) : (
        <div className="list">
          {accounts.map((a) => (
            <div key={a.id} className="card row">
              <div>
                <div className="mono">{a.account_number}</div>
                <div className="muted">{a.type} · {a.currency}</div>
              </div>
              <span className={`badge ${a.status === 'ACTIVE' ? 'ok' : ''}`}>{a.status}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
