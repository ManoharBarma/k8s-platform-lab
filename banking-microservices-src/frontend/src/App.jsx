import { Routes, Route, Navigate } from 'react-router-dom'
import { ProtectedRoute } from './auth.jsx'
import Layout from './components/Layout.jsx'
import Login from './pages/Login.jsx'
import Register from './pages/Register.jsx'
import Dashboard from './pages/Dashboard.jsx'
import Accounts from './pages/Accounts.jsx'
import Transfer from './pages/Transfer.jsx'
import Deposit from './pages/Deposit.jsx'
import Cards from './pages/Cards.jsx'
import Loans from './pages/Loans.jsx'

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route path="/" element={<Dashboard />} />
        <Route path="/accounts" element={<Accounts />} />
        <Route path="/transfer" element={<Transfer />} />
        <Route path="/deposit" element={<Deposit />} />
        <Route path="/cards" element={<Cards />} />
        <Route path="/loans" element={<Loans />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
