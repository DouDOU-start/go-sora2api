import { useEffect, useState } from 'react'
import { getSettings, updateSettings, testProxy, type ProxyTestResult } from '../api/settings'
import GlassCard from '../components/ui/GlassCard'
import LoadingState from '../components/ui/LoadingState'
import { motion, AnimatePresence } from 'framer-motion'

const inputStyle = {
  background: 'var(--bg-inset)',
  border: '1px solid var(--border-default)',
  color: 'var(--text-primary)',
  borderRadius: 'var(--radius-md)',
}

const inputFocus = (e: React.FocusEvent<HTMLInputElement>) => {
  e.target.style.borderColor = 'var(--accent)'
  e.target.style.boxShadow = '0 0 0 3px var(--accent-soft)'
}
const inputBlur = (e: React.FocusEvent<HTMLInputElement>) => {
  e.target.style.borderColor = 'var(--border-default)'
  e.target.style.boxShadow = 'none'
}

export default function Settings() {
  const [proxyUrl, setProxyUrl] = useState('')
  const [tokenRefreshInterval, setTokenRefreshInterval] = useState('')
  const [creditSyncInterval, setCreditSyncInterval] = useState('')
  const [subscriptionSyncInterval, setSubscriptionSyncInterval] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<ProxyTestResult | null>(null)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  useEffect(() => {
    loadSettings()
  }, [])

  // 消息自动消失
  useEffect(() => {
    if (message) {
      const t = setTimeout(() => setMessage(null), 3000)
      return () => clearTimeout(t)
    }
  }, [message])

  const loadSettings = async () => {
    try {
      const res = await getSettings()
      const data = res.data
      setProxyUrl(data.proxy_url || '')
      setTokenRefreshInterval(data.token_refresh_interval || '30m')
      setCreditSyncInterval(data.credit_sync_interval || '10m')
      setSubscriptionSyncInterval(data.subscription_sync_interval || '6h')
    } catch {
      setMessage({ type: 'error', text: '加载设置失败' })
    }
    setLoading(false)
  }

  const handleSave = async () => {
    setSaving(true)
    setMessage(null)
    try {
      await updateSettings({
        proxy_url: proxyUrl,
        token_refresh_interval: tokenRefreshInterval,
        credit_sync_interval: creditSyncInterval,
        subscription_sync_interval: subscriptionSyncInterval,
      })
      setMessage({ type: 'success', text: '设置已保存' })
    } catch {
      setMessage({ type: 'error', text: '保存失败' })
    }
    setSaving(false)
  }

  const handleTestProxy = async () => {
    setTesting(true)
    setTestResult(null)
    try {
      const res = await testProxy(proxyUrl)
      setTestResult(res.data)
    } catch {
      setTestResult({ success: false, error: '请求失败，请检查网络' })
    }
    setTesting(false)
  }

  if (loading) return <LoadingState />

  return (
    <div>
      {/* 页头 */}
      <motion.div
        className="mb-6"
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <h1 className="text-2xl font-semibold tracking-tight" style={{ color: 'var(--text-primary)' }}>
          系统设置
        </h1>
        <p className="text-sm mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
          配置代理和同步策略
        </p>
      </motion.div>

      <div className="space-y-4">
        {/* 代理地址 */}
        <GlassCard delay={0} className="overflow-hidden">
          <div className="p-5 sm:p-6">
            <div className="flex items-start gap-3 mb-4">
              <div
                className="w-9 h-9 rounded-xl flex items-center justify-center flex-shrink-0 mt-0.5"
                style={{ background: 'var(--info-soft)' }}
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--info)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <circle cx="12" cy="12" r="10" />
                  <line x1="2" y1="12" x2="22" y2="12" />
                  <path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z" />
                </svg>
              </div>
              <div>
                <h3 className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>代理地址</h3>
                <p className="text-xs mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
                  全局 HTTP/SOCKS5 代理，所有 Sora API 请求通过此代理。留空直连。
                </p>
              </div>
            </div>
            <input
              type="text"
              value={proxyUrl}
              onChange={(e) => { setProxyUrl(e.target.value); setTestResult(null) }}
              placeholder="http://127.0.0.1:7890 或 socks5://127.0.0.1:1080 或 ip:port:user:pass"
              className="w-full px-3.5 py-2.5 text-sm outline-none transition-all"
              style={inputStyle}
              onFocus={inputFocus}
              onBlur={inputBlur}
            />
            <div className="flex items-center gap-3 mt-3">
              <button
                onClick={handleTestProxy}
                disabled={testing}
                className="px-4 py-1.5 rounded-lg text-xs font-medium transition-all cursor-pointer"
                style={{
                  background: 'var(--bg-inset)',
                  border: '1px solid var(--border-default)',
                  color: 'var(--text-secondary)',
                }}
                onMouseEnter={(e) => { if (!testing) { e.currentTarget.style.borderColor = 'var(--accent)'; e.currentTarget.style.color = 'var(--accent)' } }}
                onMouseLeave={(e) => { e.currentTarget.style.borderColor = 'var(--border-default)'; e.currentTarget.style.color = 'var(--text-secondary)' }}
              >
                {testing ? (
                  <span className="inline-flex items-center gap-1.5">
                    <svg className="w-3 h-3" style={{ animation: 'spin 0.8s linear infinite' }} viewBox="0 0 24 24" fill="none">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                    </svg>
                    测试中...
                  </span>
                ) : '测试连通性'}
              </button>
              <AnimatePresence>
                {testResult && (
                  <motion.span
                    className="text-xs font-medium"
                    style={{ color: testResult.success ? 'var(--success)' : 'var(--danger)' }}
                    initial={{ opacity: 0, x: -8 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0 }}
                  >
                    {testResult.success
                      ? `连接成功 (${testResult.latency}ms)`
                      : testResult.error || '连接失败'}
                  </motion.span>
                )}
              </AnimatePresence>
            </div>
          </div>
        </GlassCard>

        {/* 同步间隔 */}
        <GlassCard delay={1} className="overflow-hidden">
          <div className="p-5 sm:p-6">
            <div className="flex items-start gap-3 mb-4">
              <div
                className="w-9 h-9 rounded-xl flex items-center justify-center flex-shrink-0 mt-0.5"
                style={{ background: 'var(--success-soft)' }}
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--success)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="23 4 23 10 17 10" />
                  <polyline points="1 20 1 14 7 14" />
                  <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15" />
                </svg>
              </div>
              <div>
                <h3 className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>后台同步间隔</h3>
                <p className="text-xs mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
                  Go duration 格式，如 30m、1h、6h。修改后下个周期生效。
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>
                  Token 刷新
                </label>
                <input
                  type="text"
                  value={tokenRefreshInterval}
                  onChange={(e) => setTokenRefreshInterval(e.target.value)}
                  placeholder="30m"
                  className="w-full px-3.5 py-2.5 text-sm outline-none transition-all"
                  style={inputStyle}
                  onFocus={inputFocus}
                  onBlur={inputBlur}
                />
              </div>
              <div>
                <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>
                  配额同步
                </label>
                <input
                  type="text"
                  value={creditSyncInterval}
                  onChange={(e) => setCreditSyncInterval(e.target.value)}
                  placeholder="10m"
                  className="w-full px-3.5 py-2.5 text-sm outline-none transition-all"
                  style={inputStyle}
                  onFocus={inputFocus}
                  onBlur={inputBlur}
                />
              </div>
              <div>
                <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>
                  订阅同步
                </label>
                <input
                  type="text"
                  value={subscriptionSyncInterval}
                  onChange={(e) => setSubscriptionSyncInterval(e.target.value)}
                  placeholder="6h"
                  className="w-full px-3.5 py-2.5 text-sm outline-none transition-all"
                  style={inputStyle}
                  onFocus={inputFocus}
                  onBlur={inputBlur}
                />
              </div>
            </div>
          </div>
        </GlassCard>
      </div>

      {/* 保存 & 消息 */}
      <div className="flex items-center gap-4 mt-6">
        <button
          onClick={handleSave}
          disabled={saving}
          className="px-6 py-2.5 rounded-xl text-sm font-semibold text-white disabled:opacity-50 transition-all cursor-pointer"
          style={{ background: 'var(--accent)' }}
          onMouseEnter={(e) => { if (!saving) e.currentTarget.style.background = 'var(--accent-hover)' }}
          onMouseLeave={(e) => { e.currentTarget.style.background = 'var(--accent)' }}
        >
          {saving ? (
            <span className="inline-flex items-center gap-2">
              <svg className="w-4 h-4" style={{ animation: 'spin 0.8s linear infinite' }} viewBox="0 0 24 24" fill="none">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              保存中...
            </span>
          ) : '保存设置'}
        </button>

        <AnimatePresence>
          {message && (
            <motion.span
              className="text-sm font-medium"
              style={{ color: message.type === 'success' ? 'var(--success)' : 'var(--danger)' }}
              initial={{ opacity: 0, x: -8 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0 }}
            >
              {message.type === 'success' ? '✓ ' : ''}{message.text}
            </motion.span>
          )}
        </AnimatePresence>
      </div>
    </div>
  )
}
