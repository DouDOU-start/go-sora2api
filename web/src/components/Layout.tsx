import { NavLink, Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { apiSections } from '../data/apiDocs'

const navItems = [
  { path: '/', label: '概览', icon: BarChartIcon },
  { path: '/accounts', label: '账号', icon: UserIcon },
  { path: '/groups', label: '分组', icon: FolderIcon },
  { path: '/api-keys', label: '密钥', icon: KeyIcon },
  { path: '/tasks', label: '任务', icon: ListIcon },
  { path: '/settings', label: '设置', icon: GearIcon },
  { path: '/docs', label: '文档', icon: BookIcon },
]

// 文档二级导航数据
const docSubItems = apiSections.flatMap((s) =>
  s.endpoints.map((ep) => ({ id: ep.id, label: ep.title, method: ep.testable !== false ? ep.method : '' }))
)

export default function Layout() {
  const { logout, theme, toggleTheme } = useAuthStore()
  const navigate = useNavigate()
  const location = useLocation()
  const [mobileOpen, setMobileOpen] = useState(false)

  // 路由变化时关闭移动端菜单
  useEffect(() => {
    setMobileOpen(false)
  }, [location.pathname])

  // 移动端菜单打开时禁止滚动
  useEffect(() => {
    document.body.style.overflow = mobileOpen ? 'hidden' : ''
    return () => { document.body.style.overflow = '' }
  }, [mobileOpen])

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const currentPage = navItems.find(
    (item) => item.path === '/' ? location.pathname === '/' : location.pathname.startsWith(item.path)
  )

  return (
    <div className="flex h-dvh overflow-hidden" style={{ background: 'var(--bg-root)' }}>
      {/* ── 桌面端侧边栏 ── */}
      <aside
        className="hidden lg:flex flex-col w-[220px] flex-shrink-0"
        style={{ background: 'var(--bg-sidebar)' }}
      >
        {/* Logo */}
        <div className="flex items-center gap-2.5 px-5 h-16 flex-shrink-0">
          <div className="w-8 h-8 rounded-lg flex items-center justify-center"
            style={{ background: 'var(--accent)' }}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round">
              <polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5" />
            </svg>
          </div>
          <span className="text-[15px] font-semibold tracking-tight" style={{ color: 'var(--text-sidebar-active)' }}>
            Sora Console
          </span>
        </div>

        {/* 导航链接 */}
        <nav className="flex-1 px-3 mt-2 space-y-0.5 overflow-y-auto">
          {navItems.map((item) => (
            <div key={item.path}>
              <NavLink
                to={item.path}
                end={item.path === '/'}
                className={({ isActive }) =>
                  `group flex items-center gap-3 px-3 py-2 rounded-lg text-[13px] font-medium transition-all duration-200 ${
                    isActive
                      ? 'text-white'
                      : 'hover:text-gray-300'
                  }`
                }
                style={({ isActive }) => ({
                  background: isActive ? 'var(--bg-sidebar-active)' : 'transparent',
                  color: isActive ? 'var(--text-sidebar-active)' : 'var(--text-sidebar)',
                })}
              >
                {({ isActive }) => (
                  <>
                    <item.icon active={isActive} />
                    {item.label}
                    {isActive && (
                      <div
                        className="ml-auto w-1.5 h-1.5 rounded-full"
                        style={{ background: 'var(--accent)' }}
                      />
                    )}
                  </>
                )}
              </NavLink>
              {/* 文档页二级导航 */}
              {item.path === '/docs' && location.pathname === '/docs' && (
                <DocSubNav />
              )}
            </div>
          ))}
        </nav>

        {/* 底部操作 */}
        <div className="px-3 pb-4 space-y-1">
          <button
            onClick={toggleTheme}
            className="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-[13px] font-medium transition-all duration-200 cursor-pointer"
            style={{ color: 'var(--text-sidebar)', background: 'transparent' }}
            onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-sidebar-hover)' }}
            onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
          >
            {theme === 'light' ? <MoonIcon /> : <SunIcon />}
            {theme === 'light' ? '暗色模式' : '亮色模式'}
          </button>
          <button
            onClick={handleLogout}
            className="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-[13px] font-medium transition-all duration-200 cursor-pointer"
            style={{ color: 'var(--text-sidebar)', background: 'transparent' }}
            onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-sidebar-hover)' }}
            onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
          >
            <LogoutIcon />
            退出登录
          </button>
        </div>
      </aside>

      {/* ── 主内容区 ── */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* 移动端顶栏 */}
        <header
          className="lg:hidden flex items-center justify-between px-4 h-14 flex-shrink-0 border-b"
          style={{
            background: 'var(--bg-surface)',
            borderColor: 'var(--border-default)',
          }}
        >
          <button
            onClick={() => setMobileOpen(true)}
            className="p-2 -ml-2 rounded-lg transition-colors cursor-pointer"
            style={{ color: 'var(--text-secondary)' }}
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="15" y2="12" />
              <line x1="3" y1="18" x2="18" y2="18" />
            </svg>
          </button>
          <span className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>
            {currentPage?.label || 'Sora Console'}
          </span>
          <button
            onClick={toggleTheme}
            className="p-2 -mr-2 rounded-lg transition-colors cursor-pointer"
            style={{ color: 'var(--text-secondary)' }}
          >
            {theme === 'light' ? <MoonIcon /> : <SunIcon />}
          </button>
        </header>

        {/* 页面内容 */}
        <main className="flex-1 overflow-y-auto">
          <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-6 lg:py-8">
            <Outlet />
          </div>
        </main>

        {/* 移动端底部导航 */}
        <nav
          className="lg:hidden flex items-center justify-around h-14 flex-shrink-0 border-t"
          style={{
            background: 'var(--bg-surface)',
            borderColor: 'var(--border-default)',
          }}
        >
          {navItems.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === '/'}
              className="flex flex-col items-center gap-0.5 py-1 px-3"
            >
              {({ isActive }) => (
                <>
                  <item.icon active={isActive} />
                  <span
                    className="text-[10px] font-medium"
                    style={{ color: isActive ? 'var(--accent)' : 'var(--text-tertiary)' }}
                  >
                    {item.label}
                  </span>
                </>
              )}
            </NavLink>
          ))}
        </nav>
      </div>

      {/* ── 移动端侧边栏抽屉 ── */}
      <AnimatePresence>
        {mobileOpen && (
          <>
            <motion.div
              className="fixed inset-0 z-50 lg:hidden"
              style={{ background: 'rgba(0,0,0,0.5)' }}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setMobileOpen(false)}
            />
            <motion.aside
              className="fixed left-0 top-0 bottom-0 z-50 w-[260px] flex flex-col lg:hidden"
              style={{ background: 'var(--bg-sidebar)' }}
              initial={{ x: -260 }}
              animate={{ x: 0 }}
              exit={{ x: -260 }}
              transition={{ type: 'spring', damping: 25, stiffness: 250 }}
            >
              {/* Logo */}
              <div className="flex items-center justify-between px-5 h-16 flex-shrink-0">
                <div className="flex items-center gap-2.5">
                  <div className="w-8 h-8 rounded-lg flex items-center justify-center"
                    style={{ background: 'var(--accent)' }}>
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round">
                      <polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5" />
                    </svg>
                  </div>
                  <span className="text-[15px] font-semibold tracking-tight" style={{ color: 'var(--text-sidebar-active)' }}>
                    Sora Console
                  </span>
                </div>
                <button
                  onClick={() => setMobileOpen(false)}
                  className="p-1.5 rounded-lg cursor-pointer"
                  style={{ color: 'var(--text-sidebar)' }}
                >
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                  </svg>
                </button>
              </div>

              {/* 导航 */}
              <nav className="flex-1 px-3 mt-2 space-y-0.5">
                {navItems.map((item) => (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    end={item.path === '/'}
                    className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-[13px] font-medium transition-all duration-200"
                    style={({ isActive }) => ({
                      background: isActive ? 'var(--bg-sidebar-active)' : 'transparent',
                      color: isActive ? 'var(--text-sidebar-active)' : 'var(--text-sidebar)',
                    })}
                  >
                    {({ isActive }) => (
                      <>
                        <item.icon active={isActive} />
                        {item.label}
                      </>
                    )}
                  </NavLink>
                ))}
              </nav>

              {/* 底部 */}
              <div className="px-3 pb-6 space-y-1">
                <button
                  onClick={handleLogout}
                  className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-[13px] font-medium transition-all duration-200 cursor-pointer"
                  style={{ color: 'var(--text-sidebar)', background: 'transparent' }}
                >
                  <LogoutIcon />
                  退出登录
                </button>
              </div>
            </motion.aside>
          </>
        )}
      </AnimatePresence>
    </div>
  )
}

/* ── 文档二级导航 ── */

function DocSubNav() {
  const [activeId, setActiveId] = useState('')

  // 监听 hash 变化
  useEffect(() => {
    const onHash = () => setActiveId(window.location.hash.slice(1))
    window.addEventListener('hashchange', onHash)
    onHash()
    return () => window.removeEventListener('hashchange', onHash)
  }, [])

  const handleClick = (id: string) => {
    setActiveId(id)
    const el = document.getElementById(id)
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  return (
    <motion.div
      className="ml-5 pl-3 mt-1 mb-1 space-y-0.5"
      style={{ borderLeft: '1px solid var(--bg-sidebar-hover)' }}
      initial={{ opacity: 0, height: 0 }}
      animate={{ opacity: 1, height: 'auto' }}
      transition={{ duration: 0.2 }}
    >
      {docSubItems.map((sub) => (
        <button
          key={sub.id}
          onClick={() => handleClick(sub.id)}
          className="w-full text-left flex items-center gap-2 px-2 py-1.5 rounded-md text-[12px] transition-all duration-150 cursor-pointer"
          style={{
            background: activeId === sub.id ? 'var(--bg-sidebar-active)' : 'transparent',
            color: activeId === sub.id ? 'var(--text-sidebar-active)' : 'var(--text-sidebar)',
            fontWeight: activeId === sub.id ? 500 : 400,
          }}
          onMouseEnter={(e) => { if (activeId !== sub.id) e.currentTarget.style.background = 'var(--bg-sidebar-hover)' }}
          onMouseLeave={(e) => { if (activeId !== sub.id) e.currentTarget.style.background = 'transparent' }}
        >
          {sub.method && <MethodDot method={sub.method} />}
          <span className="truncate">{sub.label}</span>
        </button>
      ))}
    </motion.div>
  )
}

function MethodDot({ method }: { method: string }) {
  const colors: Record<string, string> = {
    GET: 'var(--success, #34d67b)',
    POST: 'var(--accent, #ff6b47)',
    PUT: 'var(--info, #5b9bf5)',
    DELETE: 'var(--danger, #f04858)',
  }
  return (
    <span
      className="w-1.5 h-1.5 rounded-full flex-shrink-0"
      style={{ background: colors[method] || 'var(--text-sidebar)' }}
    />
  )
}

/* ── SVG 图标组件 ── */
function BarChartIcon({ active }: { active?: boolean }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={active ? 'var(--accent)' : 'currentColor'} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="12" width="4" height="9" rx="1" />
      <rect x="10" y="7" width="4" height="14" rx="1" />
      <rect x="17" y="3" width="4" height="18" rx="1" />
    </svg>
  )
}

function UserIcon({ active }: { active?: boolean }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={active ? 'var(--accent)' : 'currentColor'} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="8" r="4" />
      <path d="M6 21v-1a6 6 0 0112 0v1" />
    </svg>
  )
}

function FolderIcon({ active }: { active?: boolean }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={active ? 'var(--accent)' : 'currentColor'} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z" />
    </svg>
  )
}

function KeyIcon({ active }: { active?: boolean }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={active ? 'var(--accent)' : 'currentColor'} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
    </svg>
  )
}

function ListIcon({ active }: { active?: boolean }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={active ? 'var(--accent)' : 'currentColor'} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="4" width="18" height="4" rx="1" />
      <rect x="3" y="10" width="18" height="4" rx="1" />
      <rect x="3" y="16" width="18" height="4" rx="1" />
    </svg>
  )
}

function BookIcon({ active }: { active?: boolean }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={active ? 'var(--accent)' : 'currentColor'} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4 19.5A2.5 2.5 0 016.5 17H20" />
      <path d="M6.5 2H20v20H6.5A2.5 2.5 0 014 19.5v-15A2.5 2.5 0 016.5 2z" />
    </svg>
  )
}

function GearIcon({ active }: { active?: boolean }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={active ? 'var(--accent)' : 'currentColor'} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="3" />
      <path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 01-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z" />
    </svg>
  )
}

function MoonIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
    </svg>
  )
}

function SunIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="5" />
      <line x1="12" y1="1" x2="12" y2="3" />
      <line x1="12" y1="21" x2="12" y2="23" />
      <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
      <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
      <line x1="1" y1="12" x2="3" y2="12" />
      <line x1="21" y1="12" x2="23" y2="12" />
      <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
      <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
    </svg>
  )
}

function LogoutIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4" />
      <polyline points="16 17 21 12 16 7" />
      <line x1="21" y1="12" x2="9" y2="12" />
    </svg>
  )
}
