import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { useTheme } from '../hooks/useTheme'
import Avatar from './Avatar'

export default function Layout() {
  const { user, logoutUser } = useAuth()
  const { theme, toggleTheme } = useTheme()
  const navigate = useNavigate()

  const handleLogout = () => {
    logoutUser()
    navigate('/')
  }

  const linkClass = ({ isActive }) =>
    `px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
      isActive
        ? 'bg-indigo-600 text-white'
        : 'text-slate-600 hover:bg-slate-200 hover:text-slate-900 dark:text-slate-300 dark:hover:bg-slate-700 dark:hover:text-white'
    }`

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <nav className="bg-white dark:bg-slate-800 shadow-lg">
        <div className="max-w-7xl mx-auto px-4">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-2">
              <span className="text-xl font-bold text-indigo-600 dark:text-indigo-400">Alaya Archive</span>
              <div className="hidden md:flex items-center gap-1 ml-8">
                <NavLink to="/dashboard" className={linkClass}>Dashboard</NavLink>
                <NavLink to="/collection" className={linkClass}>Collection</NavLink>
                <NavLink to="/feed" className={linkClass}>Feed</NavLink>
                <NavLink to="/friends" className={linkClass}>Friends</NavLink>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <button
                onClick={toggleTheme}
                className="p-2 rounded-lg text-slate-600 hover:bg-slate-200 dark:text-slate-300 dark:hover:bg-slate-700"
                title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
              >
                {theme === 'dark' ? '\u2600\uFE0F' : '\uD83C\uDF19'}
              </button>
              <NavLink to="/profile" className="flex items-center gap-2 text-sm text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white">
                <Avatar username={user?.username} displayName={user?.display_name} size="sm" version={user?.updated_at} />
                <span className="hidden sm:inline">{user?.display_name || user?.username}</span>
              </NavLink>
              <button
                onClick={handleLogout}
                className="text-sm text-slate-500 hover:text-red-600 dark:text-slate-400 dark:hover:text-red-400 transition-colors"
              >
                Logout
              </button>
            </div>
          </div>

          {/* Mobile nav */}
          <div className="flex md:hidden gap-1 pb-3 overflow-x-auto">
            <NavLink to="/dashboard" className={linkClass}>Dashboard</NavLink>
            <NavLink to="/collection" className={linkClass}>Collection</NavLink>
            <NavLink to="/feed" className={linkClass}>Feed</NavLink>
            <NavLink to="/friends" className={linkClass}>Friends</NavLink>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto px-4 py-6">
        <Outlet />
      </main>
    </div>
  )
}
