import { useState, type ReactElement } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import axios, { type AxiosRequestConfig } from 'axios'
import { apiSections } from '../data/apiDocs'
import type { ApiEndpoint } from '../data/apiDocs'
import GlassCard from '../components/ui/GlassCard'

const BASE_URL = window.location.origin

const inputStyle = {
  background: 'var(--bg-inset)',
  border: '1px solid var(--border-default)',
  color: 'var(--text-primary)',
  borderRadius: 'var(--radius-md)',
}

const inputFocus = (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
  e.target.style.borderColor = 'var(--accent)'
  e.target.style.boxShadow = '0 0 0 3px var(--accent-soft)'
}
const inputBlur = (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
  e.target.style.borderColor = 'var(--border-default)'
  e.target.style.boxShadow = 'none'
}

// ── 主页面 ──

export default function Docs() {
  const [apiKey, setApiKey] = useState(() => localStorage.getItem('docs_api_key') || '')

  const setAndPersistKey = (v: string) => {
    setApiKey(v)
    if (v) localStorage.setItem('docs_api_key', v)
    else localStorage.removeItem('docs_api_key')
  }

  return (
    <div>
      {/* 页头 */}
      <motion.div className="mb-6" initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }}>
        <h1 className="text-2xl font-semibold tracking-tight" style={{ color: 'var(--text-primary)' }}>
          API 文档
        </h1>
        <p className="text-sm mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
          完整接口参考与在线调试
        </p>
      </motion.div>

      {/* API Key 输入 */}
      <GlassCard delay={0} className="overflow-hidden mb-6">
        <div className="p-4 flex items-center gap-3">
          <div
            className="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
            style={{ background: 'var(--accent-soft)' }}
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--accent)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
            </svg>
          </div>
          <div className="flex-1 min-w-0">
            <input
              type="password"
              value={apiKey}
              onChange={(e) => setAndPersistKey(e.target.value)}
              placeholder="输入 API Key 以启用在线测试（sk-...）"
              className="w-full px-3.5 py-2 text-sm outline-none transition-all font-mono"
              style={{ ...inputStyle, fontSize: '13px' }}
              onFocus={inputFocus}
              onBlur={inputBlur}
            />
          </div>
          {apiKey && (
            <button
              onClick={() => setAndPersistKey('')}
              className="text-xs px-3 py-1.5 rounded-lg cursor-pointer transition-colors flex-shrink-0"
              style={{ color: 'var(--text-tertiary)', background: 'var(--bg-inset)' }}
              onMouseEnter={(e) => { e.currentTarget.style.color = 'var(--text-secondary)' }}
              onMouseLeave={(e) => { e.currentTarget.style.color = 'var(--text-tertiary)' }}
            >
              清除
            </button>
          )}
        </div>
      </GlassCard>

      {/* 文档内容 */}
      {apiSections.map((section, si) => (
        <div key={section.id} className="mb-10">
          <motion.div
            className="flex items-center gap-3 mb-4"
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: si * 0.05 }}
          >
            <span
              className="text-[11px] font-bold px-2 py-0.5 rounded-md font-mono flex-shrink-0"
              style={{ background: 'var(--accent-soft)', color: 'var(--accent)' }}
            >
              {String(si + 1).padStart(2, '0')}
            </span>
            <div>
              <h2 className="text-lg font-semibold" style={{ color: 'var(--text-primary)' }}>{section.title}</h2>
              {section.description && (
                <p className="text-xs mt-0.5" style={{ color: 'var(--text-tertiary)' }}>{section.description}</p>
              )}
            </div>
          </motion.div>

          <div className="space-y-4">
            {section.endpoints.map((ep, i) => (
              <EndpointCard key={ep.id} endpoint={ep} delay={i} apiKey={apiKey} />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

// ── 接口卡片 ──

function EndpointCard({ endpoint: ep, delay, apiKey }: { endpoint: ApiEndpoint; delay: number; apiKey: string }) {
  const isSpecial = ep.testable === false
  const showTester = !isSpecial
  const [expanded, setExpanded] = useState(true)

  const [pathValues, setPathValues] = useState<Record<string, string>>({})
  const [bodyValues, setBodyValues] = useState<Record<string, string>>({})

  const curl = showTester ? buildCurl(ep, pathValues, bodyValues) : ''

  return (
    <GlassCard delay={delay} className="overflow-hidden">
      <div id={ep.id} data-endpoint>
        {/* 标题栏 — 可折叠 */}
        <button
          onClick={() => setExpanded(!expanded)}
          className="w-full flex items-center justify-between px-5 py-3.5 cursor-pointer transition-colors"
          style={{ borderBottom: expanded ? '1px solid var(--border-default)' : '1px solid transparent' }}
          onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-surface-hover)' }}
          onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
        >
          <div className="flex items-center gap-3 min-w-0">
            {!isSpecial && <MethodBadge method={ep.method} />}
            {!isSpecial && (
              <code className="text-xs font-mono hidden sm:inline-block px-2 py-0.5 rounded-md" style={{ color: 'var(--text-secondary)', background: 'var(--bg-inset)' }}>
                {ep.path}
              </code>
            )}
            <span className="text-sm font-semibold truncate" style={{ color: 'var(--text-primary)' }}>{ep.title}</span>
          </div>
          <motion.div
            animate={{ rotate: expanded ? 180 : 0 }}
            transition={{ duration: 0.2 }}
            className="flex-shrink-0 ml-2"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round">
              <polyline points="6 9 12 15 18 9" />
            </svg>
          </motion.div>
        </button>

        <AnimatePresence initial={false}>
          {expanded && (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: 'auto', opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
              style={{ overflow: 'hidden' }}
            >
              <div className="px-5 py-4 space-y-4">
                {/* 描述 */}
                <div className="text-sm leading-relaxed" style={{ color: 'var(--text-secondary)' }}>
                  <MarkdownContent text={ep.description} />
                </div>

                {/* 参数表格 */}
                {ep.params && <ParamTable title="路径参数" params={ep.params} />}
                {ep.queryParams && <ParamTable title="查询参数" params={ep.queryParams} />}
                {ep.bodyParams && <ParamTable title="请求体参数" params={ep.bodyParams} />}

                {/* 测试面板 + 示例 */}
                {(showTester || ep.responseExample) && (
                  <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                    {/* 左：在线测试 */}
                    {showTester && (
                      <div className="min-w-0">
                        <ApiTester
                          endpoint={ep}
                          apiKey={apiKey}
                          pathValues={pathValues}
                          setPathValues={setPathValues}
                          bodyValues={bodyValues}
                          setBodyValues={setBodyValues}
                        />
                      </div>
                    )}
                    {/* 右：示例 */}
                    <div className="min-w-0 space-y-3">
                      {showTester && (
                        <div>
                          <div className="flex items-center gap-1.5 mb-2">
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                              <polyline points="4 17 10 11 4 5" /><line x1="12" y1="19" x2="20" y2="19" />
                            </svg>
                            <h4 className="text-xs font-semibold" style={{ color: 'var(--text-tertiary)' }}>请求示例</h4>
                          </div>
                          <CodeBlock code={curl} />
                        </div>
                      )}
                      {ep.responseExample && (
                        <div>
                          <div className="flex items-center gap-1.5 mb-2">
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                              <path d="M8 3H7a2 2 0 00-2 2v5a2 2 0 01-2 2 2 2 0 012 2v5c0 1.1.9 2 2 2h1" />
                              <path d="M16 21h1a2 2 0 002-2v-5c0-1.1.9-2 2-2a2 2 0 01-2-2V5a2 2 0 00-2-2h-1" />
                            </svg>
                            <h4 className="text-xs font-semibold" style={{ color: 'var(--text-tertiary)' }}>响应示例</h4>
                          </div>
                          <CodeBlock code={ep.responseExample} />
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </GlassCard>
  )
}

// ── 在线测试面板 ──

function ApiTester({
  endpoint: ep, apiKey,
  pathValues, setPathValues,
  bodyValues, setBodyValues,
}: {
  endpoint: ApiEndpoint; apiKey: string
  pathValues: Record<string, string>; setPathValues: (v: Record<string, string>) => void
  bodyValues: Record<string, string>; setBodyValues: (v: Record<string, string>) => void
}) {
  const [sending, setSending] = useState(false)
  const [result, setResult] = useState<{ status: number; time: number; body: string } | null>(null)
  const [confirmOpen, setConfirmOpen] = useState(false)

  const buildUrl = () => {
    let url = ep.path
    if (ep.params) for (const p of ep.params) url = url.replace(`:${p.name}`, pathValues[p.name] || '')
    return url
  }

  const doSend = async () => {
    if (!apiKey) {
      setResult({ status: 0, time: 0, body: '请先在页面顶部输入 API Key' })
      return
    }
    setSending(true)
    setResult(null)
    const start = performance.now()
    const isDownload = ep.id === 'download-video'

    try {
      const config: AxiosRequestConfig = {
        method: ep.method.toLowerCase(),
        url: buildUrl(),
        headers: { Authorization: `Bearer ${apiKey}` },
      }
      if (isDownload) config.responseType = 'blob'

      if (ep.method === 'POST' || ep.method === 'PUT') {
        const data: Record<string, string | number> = {}
        for (const p of ep.bodyParams || []) {
          const v = bodyValues[p.name]
          if (v) {
            data[p.name] = p.type === 'number' || p.type === 'integer' ? Number(v) : v
          }
        }
        config.data = data
        config.headers!['Content-Type'] = 'application/json'
      }

      const res = await axios.request(config)
      const elapsed = Math.round(performance.now() - start)

      if (isDownload) {
        const blob = res.data as Blob
        const a = document.createElement('a')
        a.href = URL.createObjectURL(blob)
        a.download = `video_${pathValues['id'] || 'download'}.mp4`
        a.click()
        URL.revokeObjectURL(a.href)
        setResult({ status: res.status, time: elapsed, body: `视频已下载 (${(blob.size / 1024 / 1024).toFixed(2)} MB)` })
      } else {
        setResult({ status: res.status, time: elapsed, body: JSON.stringify(res.data, null, 2) })
      }
    } catch (err) {
      const elapsed = Math.round(performance.now() - start)
      let status = 0, body = '请求失败'
      if (axios.isAxiosError(err)) {
        status = err.response?.status || 0
        const data = err.response?.data
        body = typeof data === 'string' ? data : data ? JSON.stringify(data, null, 2) : err.message
      }
      setResult({ status, time: elapsed, body })
    } finally {
      setSending(false)
    }
  }

  const handleSend = () => {
    if (ep.dangerWarning) setConfirmOpen(true)
    else doSend()
  }

  return (
    <div>
      <div className="flex items-center gap-1.5 mb-3">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--text-tertiary)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <polygon points="5 3 19 12 5 21 5 3" />
        </svg>
        <h4 className="text-xs font-semibold" style={{ color: 'var(--text-tertiary)' }}>在线测试</h4>
      </div>
      <div className="space-y-3">
        {/* 路径参数 */}
        {ep.params?.map((p) => (
          <div key={p.name}>
            <label className="block text-[11px] font-medium font-mono mb-1" style={{ color: 'var(--accent)' }}>{p.name}</label>
            <input
              value={pathValues[p.name] || ''}
              onChange={(e) => setPathValues({ ...pathValues, [p.name]: e.target.value })}
              placeholder={p.description}
              className="w-full px-3 py-2 text-sm outline-none transition-all font-mono"
              style={inputStyle}
              onFocus={inputFocus}
              onBlur={inputBlur}
            />
          </div>
        ))}

        {/* 请求体参数 */}
        {ep.bodyParams?.map((p) => (
          <div key={p.name}>
            <label className="block text-[11px] font-medium font-mono mb-1" style={{ color: 'var(--text-secondary)' }}>
              {p.name}
              {p.required && <span style={{ color: 'var(--danger)' }} className="ml-0.5">*</span>}
            </label>
            {p.name === 'model' ? (
              <select
                value={bodyValues[p.name] || ''}
                onChange={(e) => setBodyValues({ ...bodyValues, [p.name]: e.target.value })}
                className="w-full px-3 py-2 text-sm outline-none transition-all font-mono cursor-pointer"
                style={inputStyle}
                onFocus={inputFocus}
                onBlur={inputBlur}
              >
                <option value="">选择模型...</option>
                <optgroup label="标准画质 720p">
                  <option value="sora-2-landscape-10s">横屏 10s (1280x720)</option>
                  <option value="sora-2-landscape-15s">横屏 15s (1280x720)</option>
                  <option value="sora-2-landscape-25s">横屏 25s (1280x720)</option>
                  <option value="sora-2-portrait-10s">竖屏 10s (720x1280)</option>
                  <option value="sora-2-portrait-15s">竖屏 15s (720x1280)</option>
                  <option value="sora-2-portrait-25s">竖屏 25s (720x1280)</option>
                </optgroup>
                <optgroup label="高清画质 1080p (Pro)">
                  <option value="sora-2-pro-landscape-hd-10s">横屏 10s (1920x1080)</option>
                  <option value="sora-2-pro-landscape-hd-15s">横屏 15s (1920x1080)</option>
                  <option value="sora-2-pro-landscape-hd-25s">横屏 25s (1920x1080)</option>
                  <option value="sora-2-pro-portrait-hd-10s">竖屏 10s (1080x1920)</option>
                  <option value="sora-2-pro-portrait-hd-15s">竖屏 15s (1080x1920)</option>
                  <option value="sora-2-pro-portrait-hd-25s">竖屏 25s (1080x1920)</option>
                </optgroup>
              </select>
            ) : p.name === 'prompt' ? (
              <textarea
                value={bodyValues[p.name] || ''}
                onChange={(e) => setBodyValues({ ...bodyValues, [p.name]: e.target.value })}
                placeholder={p.description}
                rows={3}
                className="w-full px-3 py-2 text-sm outline-none transition-all resize-none"
                style={inputStyle}
                onFocus={inputFocus as unknown as React.FocusEventHandler<HTMLTextAreaElement>}
                onBlur={inputBlur as unknown as React.FocusEventHandler<HTMLTextAreaElement>}
              />
            ) : (
              <input
                value={bodyValues[p.name] || ''}
                onChange={(e) => setBodyValues({ ...bodyValues, [p.name]: e.target.value })}
                placeholder={p.description}
                className="w-full px-3 py-2 text-sm outline-none transition-all font-mono"
                style={inputStyle}
                onFocus={inputFocus}
                onBlur={inputBlur}
              />
            )}
          </div>
        ))}

        {/* URL 预览 + 发送 */}
        <div
          className="flex items-center gap-3 px-3 py-2 rounded-lg"
          style={{ background: 'var(--bg-inset)', border: '1px solid var(--border-subtle)' }}
        >
          <code className="flex-1 text-[11px] font-mono truncate" style={{ color: 'var(--text-tertiary)' }}>
            <span className="font-bold" style={{ color: 'var(--accent)' }}>{ep.method}</span>{' '}{buildUrl()}
          </code>
          <button
            onClick={handleSend}
            disabled={sending}
            className="px-4 py-1.5 rounded-lg text-sm font-medium text-white disabled:opacity-50 transition-all cursor-pointer flex items-center gap-1.5 flex-shrink-0"
            style={{ background: 'var(--accent)' }}
            onMouseEnter={(e) => { if (!sending) e.currentTarget.style.background = 'var(--accent-hover)' }}
            onMouseLeave={(e) => { e.currentTarget.style.background = 'var(--accent)' }}
          >
            {sending ? (
              <>
                <svg className="w-3.5 h-3.5" style={{ animation: 'spin 0.6s linear infinite' }} viewBox="0 0 24 24" fill="none">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
                发送中...
              </>
            ) : '发送请求'}
          </button>
        </div>

        {/* 危险确认 */}
        <AnimatePresence>
          {confirmOpen && (
            <motion.div
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              className="p-3 rounded-xl"
              style={{ background: 'var(--danger-soft)', border: '1px solid var(--danger)' }}
            >
              <p className="text-sm mb-2" style={{ color: 'var(--danger)' }}>{ep.dangerWarning}</p>
              <div className="flex gap-2">
                <button
                  onClick={() => { setConfirmOpen(false); doSend() }}
                  disabled={sending}
                  className="px-3 py-1.5 text-sm rounded-lg font-medium text-white cursor-pointer"
                  style={{ background: 'var(--danger)' }}
                >
                  确认发送
                </button>
                <button
                  onClick={() => setConfirmOpen(false)}
                  className="px-3 py-1.5 text-sm rounded-lg cursor-pointer"
                  style={{ color: 'var(--text-secondary)', background: 'var(--bg-inset)' }}
                >
                  取消
                </button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* 响应结果 */}
        {result && (
          <motion.div initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }}>
            <div className="flex items-center gap-3 mb-2">
              <h4 className="text-xs font-semibold" style={{ color: 'var(--text-tertiary)' }}>响应</h4>
              <span
                className="text-[11px] font-mono font-bold px-1.5 py-0.5 rounded"
                style={{
                  color: result.status >= 200 && result.status < 300 ? 'var(--success)' : result.status >= 400 ? 'var(--danger)' : 'var(--warning, var(--text-tertiary))',
                  background: result.status >= 200 && result.status < 300 ? 'var(--success-soft)' : result.status >= 400 ? 'var(--danger-soft)' : 'var(--warning-soft)',
                }}
              >
                {result.status || 'ERR'}
              </span>
              <span className="text-[11px] font-mono" style={{ color: 'var(--text-tertiary)' }}>{result.time}ms</span>
            </div>
            <CodeBlock code={result.body} />
          </motion.div>
        )}
      </div>
    </div>
  )
}

// ── 工具组件 ──

function MethodBadge({ method }: { method: string }) {
  const colors: Record<string, { fg: string; bg: string }> = {
    GET: { fg: 'var(--success)', bg: 'var(--success-soft)' },
    POST: { fg: 'var(--accent)', bg: 'var(--accent-soft)' },
    PUT: { fg: 'var(--info, var(--accent))', bg: 'var(--info-soft, var(--accent-soft))' },
    DELETE: { fg: 'var(--danger)', bg: 'var(--danger-soft)' },
  }
  const c = colors[method] || { fg: 'var(--text-secondary)', bg: 'var(--bg-inset)' }
  return (
    <span
      className="text-[10px] font-bold font-mono px-1.5 py-0.5 rounded flex-shrink-0 uppercase tracking-wide"
      style={{ color: c.fg, background: c.bg }}
    >
      {method}
    </span>
  )
}

function ParamTable({ title, params }: { title: string; params: { name: string; type: string; required: boolean; description: string }[] }) {
  return (
    <div>
      <h4 className="text-[10px] font-semibold mb-2 uppercase tracking-wider" style={{ color: 'var(--text-tertiary)' }}>{title}</h4>
      <div className="overflow-x-auto rounded-xl" style={{ border: '1px solid var(--border-default)' }}>
        <table className="w-full text-sm">
          <thead>
            <tr style={{ background: 'var(--bg-inset)' }}>
              <th className="text-left px-3 py-2 text-[10px] font-semibold uppercase tracking-wider" style={{ color: 'var(--text-tertiary)' }}>参数</th>
              <th className="text-left px-3 py-2 text-[10px] font-semibold uppercase tracking-wider" style={{ color: 'var(--text-tertiary)' }}>类型</th>
              <th className="text-left px-3 py-2 text-[10px] font-semibold uppercase tracking-wider" style={{ color: 'var(--text-tertiary)' }}>必填</th>
              <th className="text-left px-3 py-2 text-[10px] font-semibold uppercase tracking-wider" style={{ color: 'var(--text-tertiary)' }}>说明</th>
            </tr>
          </thead>
          <tbody>
            {params.map((p) => (
              <tr key={p.name} style={{ borderTop: '1px solid var(--border-subtle)' }}>
                <td className="px-3 py-2">
                  <code className="text-xs font-mono px-1.5 py-0.5 rounded" style={{ color: 'var(--accent)', background: 'var(--accent-soft)' }}>{p.name}</code>
                </td>
                <td className="px-3 py-2 text-xs font-mono" style={{ color: 'var(--text-tertiary)' }}>{p.type}</td>
                <td className="px-3 py-2">
                  <span
                    className="text-[10px] font-medium px-1.5 py-0.5 rounded-full"
                    style={{
                      background: p.required ? 'var(--danger-soft)' : 'var(--bg-inset)',
                      color: p.required ? 'var(--danger)' : 'var(--text-tertiary)',
                    }}
                  >
                    {p.required ? '必填' : '可选'}
                  </span>
                </td>
                <td className="px-3 py-2 text-xs" style={{ color: 'var(--text-secondary)' }}>{p.description}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function CodeBlock({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)
  const copy = () => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }
  return (
    <div className="relative group rounded-xl overflow-hidden" style={{ background: 'var(--bg-inset)', border: '1px solid var(--border-default)' }}>
      <button
        onClick={copy}
        className="absolute top-2 right-2 px-2 py-1 rounded-lg text-[10px] font-medium opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer flex items-center gap-1"
        style={{ background: 'var(--bg-surface)', color: 'var(--text-tertiary)', border: '1px solid var(--border-subtle)' }}
      >
        {copied ? (
          <>
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="var(--success)" strokeWidth="2.5" strokeLinecap="round"><polyline points="20 6 9 17 4 12" /></svg>
            已复制
          </>
        ) : (
          <>
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" /><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" /></svg>
            复制
          </>
        )}
      </button>
      <pre className="p-3 text-xs font-mono overflow-x-auto whitespace-pre-wrap leading-relaxed" style={{ color: 'var(--text-secondary)' }}>
        {code}
      </pre>
    </div>
  )
}

function buildCurl(ep: ApiEndpoint, pathValues: Record<string, string>, bodyValues: Record<string, string>): string {
  const method = ep.method.toUpperCase()
  let url = ep.path
  if (ep.params) for (const p of ep.params) url = url.replace(`:${p.name}`, pathValues[p.name] || `<${p.name}>`)
  const fullUrl = `${BASE_URL}${url}`
  const parts: string[] = []

  if (ep.id === 'download-video') {
    parts.push(`curl -o video.mp4 \\`)
    parts.push(`  -H "Authorization: Bearer <API_KEY>" \\`)
    parts.push(`  ${fullUrl}`)
    return parts.join('\n')
  }

  if (method === 'GET') {
    parts.push(`curl "${fullUrl}" \\`)
    parts.push(`  -H "Authorization: Bearer <API_KEY>"`)
  } else {
    parts.push(`curl -X ${method} ${fullUrl} \\`)
    parts.push(`  -H "Authorization: Bearer <API_KEY>" \\`)
    parts.push(`  -H "Content-Type: application/json" \\`)
    const data: Record<string, string | number> = {}
    for (const p of ep.bodyParams || []) {
      const v = bodyValues[p.name]
      if (v) data[p.name] = p.type === 'number' ? Number(v) : v
      else if (p.required) data[p.name] = `<${p.name}>`
    }
    parts.push(`  -d '${JSON.stringify(data, null, 2).split('\n').join('\n  ')}'`)
  }
  return parts.join('\n')
}

// ── Markdown 渲染 ──

function renderInlineMarkdown(text: string) {
  const parts: (string | ReactElement)[] = []
  const regex = /(\*\*(.+?)\*\*|`(.+?)`)/g
  let lastIndex = 0
  let match: RegExpExecArray | null
  let key = 0
  while ((match = regex.exec(text)) !== null) {
    if (match.index > lastIndex) parts.push(text.slice(lastIndex, match.index))
    if (match[2]) parts.push(<strong key={key++} style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{match[2]}</strong>)
    else if (match[3]) parts.push(<code key={key++} className="text-[11px] px-1 py-0.5 rounded font-mono" style={{ background: 'var(--bg-inset)', color: 'var(--accent)' }}>{match[3]}</code>)
    lastIndex = match.index + match[0].length
  }
  if (lastIndex < text.length) parts.push(text.slice(lastIndex))
  return parts
}

function MdTable({ lines }: { lines: string[] }) {
  const rows = lines.filter((l) => !l.match(/^\|[\s-|]+\|$/)).map((l) => l.split('|').filter(Boolean).map((c) => c.trim()))
  const header = rows[0] || []
  const body = rows.slice(1)
  if (!header.length) return null
  return (
    <div className="overflow-x-auto rounded-xl my-2" style={{ border: '1px solid var(--border-default)' }}>
      <table className="w-full text-sm">
        <thead><tr style={{ background: 'var(--bg-inset)' }}>
          {header.map((h, i) => <th key={i} className="text-left px-3 py-2 text-[10px] font-semibold uppercase tracking-wider" style={{ color: 'var(--text-tertiary)' }}>{h}</th>)}
        </tr></thead>
        <tbody>
          {body.map((row, i) => (
            <tr key={i} style={{ borderTop: '1px solid var(--border-subtle)' }}>
              {row.map((cell, j) => <td key={j} className="px-3 py-1.5 text-xs font-mono" style={{ color: 'var(--text-secondary)' }}>{cell}</td>)}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function MarkdownContent({ text }: { text: string }) {
  const lines = text.split('\n')
  type Segment = { type: 'table' | 'text'; lines: string[] }
  const segments: Segment[] = []
  for (const line of lines) {
    const isTable = line.trim().startsWith('|')
    const last = segments[segments.length - 1]
    if (last && last.type === (isTable ? 'table' : 'text')) last.lines.push(line)
    else segments.push({ type: isTable ? 'table' : 'text', lines: [line] })
  }
  return (
    <div>
      {segments.map((seg, i) => {
        if (seg.type === 'table') return <MdTable key={i} lines={seg.lines} />
        return seg.lines.map((l, j) => {
          const trimmed = l.trim()
          if (!trimmed) return <br key={`${i}-${j}`} />
          return <p key={`${i}-${j}`}>{renderInlineMarkdown(trimmed)}</p>
        })
      })}
    </div>
  )
}
