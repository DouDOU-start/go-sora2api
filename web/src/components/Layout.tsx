import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'

const navItems = [
  { path: '/', label: '概览', icon: '📊' },
  { path: '/accounts', label: '账号管理', icon: '👤' },
  { path: '/groups', label: '账号组', icon: '📁' },
  { path: '/tasks', label: '任务列表', icon: '📋' },
  { path: '/settings', label: '系统设置', icon: '⚙️' },
]

export default function Layout() {
  const { logout, theme, toggleTheme } = useAuthStore()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className={`min-h-screen bg-gray-50 dark:bg-gray-950 ${theme === 'dark' ? 'dark' : ''}`}>
      {/* 顶栏 */}
      <header className="sticky top-0 z-50 bg-white/80 dark:bg-gray-900/80 backdrop-blur-xl border-b border-gray-200 dark:border-gray-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 flex items-center justify-between h-14">
          <div className="flex items-center gap-6">
            <h1 className="text-lg font-bold bg-gradient-to-r from-indigo-500 to-purple-500 bg-clip-text text-transparent">
              Sora2API
            </h1>
            <nav className="hidden md:flex items-center gap-1">
              {navItems.map((item) => (
                <NavLink
                  key={item.path}
                  to={item.path}
                  end={item.path === '/'}
                  className={({ isActive }) =>
                    `px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                      isActive
                        ? 'bg-indigo-50 dark:bg-indigo-950 text-indigo-600 dark:text-indigo-400'
                        : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'
                    }`
                  }
                >
                  <span className="mr-1">{item.icon}</span>
                  {item.label}
                </NavLink>
              ))}
            </nav>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={toggleTheme}
              className="p-2 rounded-lg text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            >
              {theme === 'light' ? '🌙' : '☀️'}
            </button>
            <button
              onClick={handleLogout}
              className="px-3 py-1.5 rounded-lg text-sm text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            >
              退出
            </button>
          </div>
        </div>
      </header>

      {/* 移动端导航 */}
      <nav className="md:hidden flex items-center gap-1 px-4 py-2 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 overflow-x-auto">
        {navItems.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            end={item.path === '/'}
            className={({ isActive }) =>
              `flex-shrink-0 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-indigo-50 dark:bg-indigo-950 text-indigo-600 dark:text-indigo-400'
                  : 'text-gray-600 dark:text-gray-400'
              }`
            }
          >
            <span className="mr-1">{item.icon}</span>
            {item.label}
          </NavLink>
        ))}
      </nav>

      {/* 内容 */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 py-6">
        <Outlet />
      </main>
    </div>
  )
}
