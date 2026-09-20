import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth.jsx'

export default function Layout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    navigate('/login')
  }

  const links = [
    ['/', 'Dashboard'],
    ['/accounts', 'Accounts'],
    ['/deposit', 'Deposit'],
    ['/transfer', 'Transfer'],
    ['/cards', 'Cards'],
    ['/loans', 'Loans'],
  ]

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">🏦 Banking</div>
        <nav>
          {links.map(([to, label]) => (
            <NavLink key={to} to={to} end={to === '/'} className={({ isActive }) => (isActive ? 'active' : '')}>
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="user">
          <div className="email">{user?.email || user?.subject}</div>
          <button onClick={handleLogout}>Log out</button>
        </div>
      </aside>
      <main className="content">
        <Outlet />
      </main>
    </div>
  )
}
