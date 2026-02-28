import { useEffect, useState } from 'react'
import { getSettings, updateSettings } from '../api/settings'

export default function Settings() {
  const [apiKeys, setApiKeys] = useState('')
  const [proxyUrl, setProxyUrl] = useState('')
  const [tokenRefreshInterval, setTokenRefreshInterval] = useState('')
  const [creditSyncInterval, setCreditSyncInterval] = useState('')
  const [subscriptionSyncInterval, setSubscriptionSyncInterval] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  useEffect(() => {
    loadSettings()
  }, [])

  const loadSettings = async () => {
    try {
      const res = await getSettings()
      const data = res.data
      // api_keys 是 JSON 字符串，显示时转为每行一个
      try {
        const keys = JSON.parse(data.api_keys || '[]') as string[]
        setApiKeys(keys.join('\n'))
      } catch {
        setApiKeys(data.api_keys || '')
      }
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
      // 将多行文本转为 JSON 数组
      const keys = apiKeys
        .split('\n')
        .map((k) => k.trim())
        .filter(Boolean)
      const keysJSON = JSON.stringify(keys)

      await updateSettings({
        api_keys: keysJSON,
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

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="text-gray-400">加载中...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">系统设置</h2>

      <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-800 divide-y divide-gray-200 dark:divide-gray-800">
        {/* API Keys */}
        <div className="p-6">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-1">API Keys</h3>
          <p className="text-xs text-gray-500 dark:text-gray-400 mb-3">
            用于 /v1/ API 接口认证，每行一个 Key。留空则不需要认证。
          </p>
          <textarea
            value={apiKeys}
            onChange={(e) => setApiKeys(e.target.value)}
            rows={4}
            placeholder="sk-your-api-key-1&#10;sk-your-api-key-2"
            className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm font-mono focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none resize-none"
          />
        </div>

        {/* Proxy URL */}
        <div className="p-6">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-1">代理地址</h3>
          <p className="text-xs text-gray-500 dark:text-gray-400 mb-3">
            全局 HTTP/SOCKS5 代理，所有 Sora API 请求会通过此代理。留空则直连。
          </p>
          <input
            type="text"
            value={proxyUrl}
            onChange={(e) => setProxyUrl(e.target.value)}
            placeholder="http://127.0.0.1:7890 或 socks5://127.0.0.1:1080"
            className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
          />
        </div>

        {/* 同步设置 */}
        <div className="p-6">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-1">后台同步间隔</h3>
          <p className="text-xs text-gray-500 dark:text-gray-400 mb-4">
            Go duration 格式，如 30m（30分钟）、1h（1小时）、6h（6小时）。修改后下个周期生效。
          </p>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                Token 刷新间隔
              </label>
              <input
                type="text"
                value={tokenRefreshInterval}
                onChange={(e) => setTokenRefreshInterval(e.target.value)}
                placeholder="30m"
                className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                配额同步间隔
              </label>
              <input
                type="text"
                value={creditSyncInterval}
                onChange={(e) => setCreditSyncInterval(e.target.value)}
                placeholder="10m"
                className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                订阅同步间隔
              </label>
              <input
                type="text"
                value={subscriptionSyncInterval}
                onChange={(e) => setSubscriptionSyncInterval(e.target.value)}
                placeholder="6h"
                className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
              />
            </div>
          </div>
        </div>
      </div>

      {/* 保存按钮和消息 */}
      <div className="flex items-center gap-4">
        <button
          onClick={handleSave}
          disabled={saving}
          className="px-6 py-2.5 rounded-lg bg-gradient-to-r from-indigo-500 to-purple-500 text-white font-medium text-sm hover:from-indigo-600 hover:to-purple-600 disabled:opacity-50 transition-all"
        >
          {saving ? '保存中...' : '保存设置'}
        </button>

        {message && (
          <span
            className={`text-sm ${
              message.type === 'success' ? 'text-green-600 dark:text-green-400' : 'text-red-500'
            }`}
          >
            {message.text}
          </span>
        )}
      </div>
    </div>
  )
}
