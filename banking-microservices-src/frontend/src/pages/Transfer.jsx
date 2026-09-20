import { useEffect, useState } from 'react'
import { api } from '../api.js'

// crypto.randomUUID gives us a fresh idempotency key per submission.
function newIdempotencyKey() {
  return (crypto.randomUUID && crypto.randomUUID()) || String(Date.now())
}

export default function Transfer() {
  const [accounts, setAccounts] = useState([])
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [amount, setAmount] = useState('')
  const [currency, setCurrency] = useState('USD')
  const [msg, setMsg] = useState(null)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    api.get('/accounts').then((d) => setAccounts(d.items || [])).catch((e) => setError(e.message))
  }, [])

  async function submit(e) {
    e.preventDefault()
    setError(''); setMsg(null); setBusy(true)
    try {
      const t = await api.post('/transactions/transfer',
        { from_account_id: from, to_account_id: to, amount, currency, reference: 'web transfer' },
        { 'Idempotency-Key': newIdempotencyKey() })
      setMsg(`Transfer ${t.status} · ${t.amount} ${t.currency}`)
    } catch (e) {
      setError(e.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div>
      <h1>Transfer money</h1>
      {error && <div className="error">{error}</div>}
      {msg && <div className="success">{msg}</div>}
      <form className="card form" onSubmit={submit}>
        <label>From account
          <select value={from} onChange={(e) => setFrom(e.target.value)} required>
            <option value="">Select…</option>
            {accounts.map((a) => <option key={a.id} value={a.id}>{a.account_number} ({a.currency})</option>)}
          </select>
        </label>
        <label>To account ID
          <input value={to} onChange={(e) => setTo(e.target.value)} placeholder="destination account id" required />
        </label>
        <label>Amount<input value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="100.00" required /></label>
        <label>Currency
          <select value={currency} onChange={(e) => setCurrency(e.target.value)}>
            <option>USD</option><option>EUR</option><option>GBP</option><option>INR</option>
          </select>
        </label>
        <button disabled={busy} type="submit">{busy ? 'Sending…' : 'Send transfer'}</button>
      </form>
    </div>
  )
}
