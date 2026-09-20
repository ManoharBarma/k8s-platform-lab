import { useEffect, useState } from 'react'
import { api } from '../api.js'

export default function Loans() {
  const [loans, setLoans] = useState([])
  const [accounts, setAccounts] = useState([])
  const [accountId, setAccountId] = useState('')
  const [principal, setPrincipal] = useState('')
  const [rate, setRate] = useState('12')
  const [tenure, setTenure] = useState('12')
  const [error, setError] = useState('')

  async function load() {
    try {
      setLoans((await api.get('/loans')).items || [])
    } catch (e) { setError(e.message) }
  }

  useEffect(() => {
    load()
    api.get('/accounts').then((d) => setAccounts(d.items || [])).catch(() => {})
  }, [])

  async function apply(e) {
    e.preventDefault()
    setError('')
    try {
      await api.post('/loans', {
        account_id: accountId,
        principal,
        currency: 'USD',
        annual_rate_pct: Number(rate),
        tenure_months: Number(tenure),
      })
      await load()
    } catch (e) { setError(e.message) }
  }

  return (
    <div>
      <h1>Loans</h1>
      {error && <div className="error">{error}</div>}
      <form className="card form" onSubmit={apply}>
        <label>Disburse to account
          <select value={accountId} onChange={(e) => setAccountId(e.target.value)} required>
            <option value="">Select…</option>
            {accounts.map((a) => <option key={a.id} value={a.id}>{a.account_number}</option>)}
          </select>
        </label>
        <label>Principal<input value={principal} onChange={(e) => setPrincipal(e.target.value)} placeholder="10000" required /></label>
        <label>Annual rate %<input value={rate} onChange={(e) => setRate(e.target.value)} /></label>
        <label>Tenure (months)<input value={tenure} onChange={(e) => setTenure(e.target.value)} /></label>
        <button type="submit">Apply for loan</button>
      </form>

      <div className="list">
        {loans.map((l) => (
          <div key={l.id} className="card row">
            <div>
              <div className="mono">{l.principal} {l.currency}</div>
              <div className="muted">EMI {l.emi_amount} · {l.tenure_months} mo @ {l.annual_rate_pct}%</div>
            </div>
            <span className="badge">{l.status}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
