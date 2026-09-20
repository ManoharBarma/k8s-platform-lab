import { useEffect, useState } from 'react'
import { api } from '../api.js'

// crypto.randomUUID gives us a fresh idempotency key per submission.
function newIdempotencyKey() {
  return (crypto.randomUUID && crypto.randomUUID()) || String(Date.now())
}

export default function Deposit() {
  const [accounts, setAccounts] = useState([])
  const [account, setAccount] = useState('')
  const [amount, setAmount] = useState('')
  const [currency, setCurrency] = useState('USD')
  const [msg, setMsg] = useState(null)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    api.get('/accounts').then((d) => setAccounts(d.items || [])).catch((e) => setError(e.message))
  }, [])

  // Default the currency to the selected account's currency.
  useEffect(() => {
    const a = accounts.find((x) => x.id === account)
    if (a?.currency) setCurrency(a.currency)
  }, [account, accounts])

  async function submit(e) {
    e.preventDefault()
    setError(''); setMsg(null); setBusy(true)
    try {
      const t = await api.post('/transactions/deposit',
        { account_id: account, amount, currency, reference: 'web deposit' },
        { 'Idempotency-Key': newIdempotencyKey() })
      setMsg(`Deposit ${t.status} · ${t.amount} ${t.currency}`)
      setAmount('')
    } catch (e) {
      setError(e.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div>
      <h1>Deposit money</h1>
      <p className="muted">Add funds to one of your accounts.</p>
      {error && <div className="error">{error}</div>}
      {msg && <div className="success">{msg}</div>}
      <form className="card form" onSubmit={submit}>
        <label>Account
          <select value={account} onChange={(e) => setAccount(e.target.value)} required>
            <option value="">Select…</option>
            {accounts.map((a) => (
              <option key={a.id} value={a.id}>{a.account_number} · {a.type} ({a.currency})</option>
            ))}
          </select>
        </label>
        <label>Amount<input value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="1000.00" required /></label>
        <label>Currency
          <select value={currency} onChange={(e) => setCurrency(e.target.value)}>
            <option>USD</option><option>EUR</option><option>GBP</option><option>INR</option>
          </select>
        </label>
        <button disabled={busy} type="submit">{busy ? 'Depositing…' : 'Deposit'}</button>
      </form>
    </div>
  )
}
