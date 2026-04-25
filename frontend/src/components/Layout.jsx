import { useEffect, useState } from 'react'
import { Outlet, NavLink, useNavigate, useLocation, Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { useTheme } from '../hooks/useTheme'
import Avatar from './Avatar'

const navItems = [
  { to: '/dashboard', label: 'Dashboard' },
  { to: '/collection', label: 'Collection' },
  { to: '/feed', label: 'Feed' },
  { to: '/friends', label: 'Friends' },
]

function HamburgerIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="w-6 h-6" aria-hidden="true">
      <line x1="3" y1="6" x2="21" y2="6" />
      <line x1="3" y1="12" x2="21" y2="12" />
      <line x1="3" y1="18" x2="21" y2="18" />
    </svg>
  )
}

function CloseIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="w-6 h-6" aria-hidden="true">
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
  )
}

export default function Layout() {
  const { user, logoutUser } = useAuth()
  const { theme, toggleTheme } = useTheme()
  const navigate = useNavigate()
  const location = useLocation()
  const [drawerOpen, setDrawerOpen] = useState(false)

  useEffect(() => {
    setDrawerOpen(false)
  }, [location.pathname])

  useEffect(() => {
    if (!drawerOpen) return
    const onKey = (e) => { if (e.key === 'Escape') setDrawerOpen(false) }
    document.addEventListener('keydown', onKey)
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.style.overflow = ''
    }
  }, [drawerOpen])

  const handleLogout = () => {
    setDrawerOpen(false)
    logoutUser()
    navigate('/')
  }

  const desktopLinkClass = ({ isActive }) =>
    `px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
      isActive
        ? 'bg-indigo-600 text-white'
        : 'text-slate-600 hover:bg-slate-200 hover:text-slate-900 dark:text-slate-300 dark:hover:bg-slate-700 dark:hover:text-white'
    }`

  const drawerLinkClass = ({ isActive }) =>
    `block px-4 py-3 rounded-lg text-base font-medium transition-colors ${
      isActive
        ? 'bg-indigo-600 text-white'
        : 'text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700'
    }`

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <nav className="bg-white dark:bg-slate-800 shadow-lg relative z-30">
        <div className="max-w-7xl mx-auto px-4">
          <div className="flex items-center justify-between h-16">
            <Link to="/dashboard" className="text-xl font-bold text-indigo-600 dark:text-indigo-400">
              Alaya Archive
            </Link>

            <div className="hidden md:flex items-center gap-1 ml-8 flex-1">
              {navItems.map((item) => (
                <NavLink key={item.to} to={item.to} className={desktopLinkClass}>{item.label}</NavLink>
              ))}
            </div>

            <div className="hidden md:flex items-center gap-3">
              <button
                type="button"
                onClick={toggleTheme}
                className="p-2 rounded-lg text-slate-600 hover:bg-slate-200 dark:text-slate-300 dark:hover:bg-slate-700"
                title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
              >
                {theme === 'dark' ? '☀️' : '🌙'}
              </button>
              <NavLink to="/profile" className="flex items-center gap-2 text-sm text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white">
                <Avatar username={user?.username} displayName={user?.display_name} size="sm" version={user?.updated_at} />
                <span>{user?.display_name || user?.username}</span>
              </NavLink>
              <button
                type="button"
                onClick={handleLogout}
                className="text-sm text-slate-500 hover:text-red-600 dark:text-slate-400 dark:hover:text-red-400 transition-colors"
              >
                Logout
              </button>
            </div>

            <button
              type="button"
              onClick={() => setDrawerOpen(true)}
              aria-label="Open menu"
              aria-expanded={drawerOpen}
              aria-controls="mobile-drawer"
              className="md:hidden p-2 -mr-2 rounded-lg text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
            >
              <HamburgerIcon />
            </button>
          </div>
        </div>
      </nav>

      <div
        className={`fixed inset-0 z-40 bg-black/50 transition-opacity duration-200 md:hidden ${
          drawerOpen ? 'opacity-100' : 'opacity-0 pointer-events-none'
        }`}
        onClick={() => setDrawerOpen(false)}
        aria-hidden="true"
      />

      <aside
        id="mobile-drawer"
        role="dialog"
        aria-modal="true"
        aria-label="Main menu"
        className={`fixed inset-y-0 right-0 z-50 w-72 max-w-[85vw] bg-white dark:bg-slate-800 shadow-2xl transform transition-transform duration-200 md:hidden flex flex-col ${
          drawerOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        <div className="flex items-center justify-between p-4 border-b border-slate-200 dark:border-slate-700">
          <span className="text-lg font-semibold text-indigo-600 dark:text-indigo-400">Menu</span>
          <button
            type="button"
            onClick={() => setDrawerOpen(false)}
            aria-label="Close menu"
            className="p-2 -mr-2 rounded-lg text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
          >
            <CloseIcon />
          </button>
        </div>

        <NavLink
          to="/profile"
          className={({ isActive }) =>
            `flex items-center gap-3 p-4 border-b border-slate-200 dark:border-slate-700 ${
              isActive ? 'bg-indigo-50 dark:bg-slate-900' : 'hover:bg-slate-50 dark:hover:bg-slate-700'
            }`
          }
        >
          <Avatar username={user?.username} displayName={user?.display_name} size="lg" version={user?.updated_at} />
          <div className="min-w-0 flex-1">
            <div className="font-semibold text-slate-900 dark:text-white truncate">
              {user?.display_name || user?.username}
            </div>
            {user?.username && (
              <div className="text-sm text-slate-500 dark:text-slate-400 truncate">@{user.username}</div>
            )}
            <div className="text-xs text-indigo-600 dark:text-indigo-400 mt-1">View profile</div>
          </div>
        </NavLink>

        <nav className="flex-1 overflow-y-auto p-3">
          <div className="space-y-1">
            {navItems.map((item) => (
              <NavLink key={item.to} to={item.to} className={drawerLinkClass}>{item.label}</NavLink>
            ))}
          </div>
        </nav>

        <div className="p-3 border-t border-slate-200 dark:border-slate-700 space-y-1">
          <button
            type="button"
            onClick={toggleTheme}
            className="w-full flex items-center justify-between px-4 py-3 rounded-lg text-base font-medium text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
          >
            <span>Switch to {theme === 'dark' ? 'light' : 'dark'} mode</span>
            <span aria-hidden="true">{theme === 'dark' ? '☀️' : '🌙'}</span>
          </button>
          <button
            type="button"
            onClick={handleLogout}
            className="w-full text-left px-4 py-3 rounded-lg text-base font-medium text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
          >
            Log out
          </button>
        </div>
      </aside>

      <main className="max-w-7xl mx-auto px-4 py-6">
        <Outlet />
      </main>
    </div>
  )
}
