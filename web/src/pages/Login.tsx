import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import client from '../api/client'
import { motion } from 'framer-motion'

type LoginMode = 'admin' | 'apikey'

export default function Login() {
  const [mode, setMode] = useState<LoginMode>('admin')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { setToken, theme, toggleTheme } = useAuthStore()
  const navigate = useNavigate()

  // 焦点管理
  useEffect(() => {
    if (mode === 'admin') {
      const input = document.getElementById('login-username') as HTMLInputElement
      input?.focus()
    } else {
      const input = document.getElementById('login-apikey') as HTMLInputElement
      input?.focus()
    }
  }, [mode])

  // 切换模式时清空错误
  useEffect(() => {
    setError('')
  }, [mode])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (mode === 'admin' && (!username.trim() || !password.trim())) return
    if (mode === 'apikey' && !apiKey.trim()) return

    setLoading(true)
    setError('')

    try {
      if (mode === 'admin') {
        const res = await client.post('/admin/login', { username, password })
        setToken(res.data.token, res.data.role || 'admin')
      } else {
        const res = await client.post('/admin/login/apikey', { api_key: apiKey })
        setToken(res.data.token, res.data.role || 'viewer')
      }
      navigate('/')
    } catch {
      setError(mode === 'admin' ? '用户名或密码错误' : 'API Key 无效或已禁用')
    }
    setLoading(false)
  }

  const canSubmit = mode === 'admin'
    ? username.trim() && password.trim()
    : apiKey.trim()

  return (
    <div
      className="min-h-dvh flex items-center justify-center px-4 relative overflow-hidden"
      style={{ background: 'var(--bg-root)' }}
    >
      {/* 背景装饰 */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div
          className="absolute -top-[40%] -right-[20%] w-[60vw] h-[60vw] rounded-full opacity-[0.04]"
          style={{ background: 'var(--accent)' }}
        />
        <div
          className="absolute -bottom-[30%] -left-[15%] w-[50vw] h-[50vw] rounded-full opacity-[0.03]"
          style={{ background: 'var(--accent)' }}
        />
      </div>

      {/* 主题切换 */}
      <button
        onClick={toggleTheme}
        className="absolute top-5 right-5 p-2.5 rounded-xl transition-colors cursor-pointer"
        style={{
          color: 'var(--text-tertiary)',
          background: 'var(--bg-surface)',
          border: '1px solid var(--border-default)',
        }}
      >
        {theme === 'light' ? (
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8"><path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" /></svg>
        ) : (
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
        )}
      </button>

      <motion.div
        className="w-full max-w-[380px] relative z-10"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
      >
        {/* Logo & 标题 */}
        <div className="text-center mb-10">
          <motion.div
            className="inline-flex items-center justify-center w-14 h-14 rounded-2xl mb-5"
            style={{ background: 'var(--accent)' }}
            initial={{ scale: 0.5, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ delay: 0.1, type: 'spring', stiffness: 200, damping: 15 }}
          >
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round">
              <polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5" />
            </svg>
          </motion.div>
          <h1
            className="text-[28px] font-semibold tracking-tight"
            style={{ color: 'var(--text-primary)', fontFamily: 'var(--font-sans)' }}
          >
            Sora Console
          </h1>
          <p className="text-sm mt-1.5" style={{ color: 'var(--text-tertiary)' }}>
            账号池管理系统
          </p>
        </div>

        {/* 登录表单 */}
        <motion.form
          onSubmit={handleSubmit}
          className="p-6 rounded-2xl"
          style={{
            background: 'var(--bg-surface)',
            border: '1px solid var(--border-default)',
            boxShadow: 'var(--shadow-lg)',
          }}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2, duration: 0.5 }}
        >
          {/* 登录方式切换 */}
          <div
            className="flex rounded-xl p-1 mb-5"
            style={{ background: 'var(--bg-inset)' }}
          >
            {[
              { key: 'admin' as LoginMode, label: '管理员登录' },
              { key: 'apikey' as LoginMode, label: 'API Key 登录' },
            ].map((tab) => (
              <button
                key={tab.key}
                type="button"
                onClick={() => setMode(tab.key)}
                className="flex-1 py-2 rounded-lg text-[13px] font-medium transition-all duration-200 cursor-pointer"
                style={{
                  background: mode === tab.key ? 'var(--bg-surface)' : 'transparent',
                  color: mode === tab.key ? 'var(--text-primary)' : 'var(--text-tertiary)',
                  boxShadow: mode === tab.key ? 'var(--shadow-sm)' : 'none',
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {mode === 'admin' ? (
            <div className="space-y-4">
              <div>
                <label
                  htmlFor="login-username"
                  className="block text-[13px] font-medium mb-2"
                  style={{ color: 'var(--text-secondary)' }}
                >
                  用户名
                </label>
                <input
                  id="login-username"
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="admin"
                  autoComplete="username"
                  className="w-full px-3.5 py-2.5 rounded-xl text-sm outline-none transition-all duration-200"
                  style={{
                    background: 'var(--bg-inset)',
                    border: '1px solid var(--border-default)',
                    color: 'var(--text-primary)',
                  }}
                  onFocus={(e) => { e.target.style.borderColor = 'var(--accent)'; e.target.style.boxShadow = '0 0 0 3px var(--accent-soft)' }}
                  onBlur={(e) => { e.target.style.borderColor = 'var(--border-default)'; e.target.style.boxShadow = 'none' }}
                />
              </div>
              <div>
                <label
                  htmlFor="login-password"
                  className="block text-[13px] font-medium mb-2"
                  style={{ color: 'var(--text-secondary)' }}
                >
                  密码
                </label>
                <input
                  id="login-password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="******"
                  autoComplete="current-password"
                  className="w-full px-3.5 py-2.5 rounded-xl text-sm outline-none transition-all duration-200"
                  style={{
                    background: 'var(--bg-inset)',
                    border: '1px solid var(--border-default)',
                    color: 'var(--text-primary)',
                  }}
                  onFocus={(e) => { e.target.style.borderColor = 'var(--accent)'; e.target.style.boxShadow = '0 0 0 3px var(--accent-soft)' }}
                  onBlur={(e) => { e.target.style.borderColor = 'var(--border-default)'; e.target.style.boxShadow = 'none' }}
                />
              </div>
            </div>
          ) : (
            <div>
              <label
                htmlFor="login-apikey"
                className="block text-[13px] font-medium mb-2"
                style={{ color: 'var(--text-secondary)' }}
              >
                API Key
              </label>
              <input
                id="login-apikey"
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="sk-..."
                autoComplete="off"
                className="w-full px-3.5 py-2.5 rounded-xl text-sm outline-none transition-all duration-200 font-mono"
                style={{
                  background: 'var(--bg-inset)',
                  border: '1px solid var(--border-default)',
                  color: 'var(--text-primary)',
                }}
                onFocus={(e) => { e.target.style.borderColor = 'var(--accent)'; e.target.style.boxShadow = '0 0 0 3px var(--accent-soft)' }}
                onBlur={(e) => { e.target.style.borderColor = 'var(--border-default)'; e.target.style.boxShadow = 'none' }}
              />
            </div>
          )}

          {/* 错误信息 */}
          {error && (
            <motion.p
              className="mt-3 text-sm px-3 py-2 rounded-lg"
              style={{ background: 'var(--danger-soft)', color: 'var(--danger)' }}
              initial={{ opacity: 0, y: -4 }}
              animate={{ opacity: 1, y: 0 }}
            >
              {error}
            </motion.p>
          )}

          <button
            type="submit"
            disabled={loading || !canSubmit}
            className="w-full mt-5 py-2.5 rounded-xl text-sm font-semibold text-white transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            style={{
              background: 'var(--accent)',
            }}
            onMouseEnter={(e) => { if (!loading) e.currentTarget.style.background = 'var(--accent-hover)' }}
            onMouseLeave={(e) => { e.currentTarget.style.background = 'var(--accent)' }}
          >
            {loading ? (
              <span className="inline-flex items-center gap-2">
                <svg className="w-4 h-4" style={{ animation: 'spin 0.8s linear infinite' }} viewBox="0 0 24 24" fill="none">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
                登录中...
              </span>
            ) : '登录'}
          </button>
        </motion.form>

        <p className="text-center text-xs mt-6" style={{ color: 'var(--text-tertiary)' }}>
          Sora2API
        </p>
      </motion.div>
    </div>
  )
}
