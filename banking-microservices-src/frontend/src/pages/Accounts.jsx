import { useEffect, useState } from 'react'
import { api } from '../api.js'

export default function Accounts() {
  const [accounts, setAccounts] = useState([])
  const [type, setType] = useState('SAVINGS')
  const [currency, setCurrency] = useState('USD')
  const [balances, setBalances] = useState({})
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function load() {
    try {
      const d = await api.get('/accounts')
      setAccounts(d.items || [])
    } catch (e) {
      setError(e.message)
    }
  }

  useEffect(() => { load() }, [])

  async function open(e) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      await api.post('/accounts', { type, currency })
      await load()
    } catch (e) {
      setError(e.message)
    } finally {
      setBusy(false)
    }
  }

  async function checkBalance(id) {
    try {
      const b = await api.get(`/accounts/${id}/balance`)
      setBalances((prev) => ({ ...prev, [id]: `${b.balance} ${b.currency}` }))
    } catch (e) {
      setBalances((prev) => ({ ...prev, [id]: e.message }))
    }
  }

  return (
    <div>
      <h1>Accounts</h1>
      {error && <div className="error">{error}</div>}
      <form className="card inline-form" onSubmit={open}>
        <label>Type
          <select value={type} onChange={(e) => setType(e.target.value)}>
            <option>SAVINGS</option>
            <option>CURRENT</option>
          </select>
        </label>
        <label>Currency
          <select value={currency} onChange={(e) => setCurrency(e.target.value)}>
            <option>USD</option><option>EUR</option><option>GBP</option><option>INR</option>
          </select>
        </label>
        <button disabled={busy} type="submit">{busy ? 'Opening…' : 'Open account'}</button>
      </form>

      <div className="list">
        {accounts.map((a) => (
          <div key={a.id} className="card row">
            <div>
              <div className="mono">{a.account_number}</div>
              <div className="muted">{a.type} · {a.currency}</div>
            </div>
            <div className="row-actions">
              {balances[a.id] && <span className="badge">{balances[a.id]}</span>}
              <button onClick={() => checkBalance(a.id)}>Balance</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
