const TOKEN_KEY = 'agent_brain_token'

export type StatsResponse = {
  session_count: number
}

function apiBase(): string {
  return (import.meta.env.VITE_API_BASE as string | undefined)?.replace(/\/$/, '') || ''
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (!headers.has('Content-Type') && init.body) headers.set('Content-Type', 'application/json')
  const token = getToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const res = await fetch(`${apiBase()}${path}`, { ...init, headers })
  if (!res.ok) {
    if (res.status === 401 && !path.includes('/auth/login')) {
      clearToken()
      if (!window.location.pathname.startsWith('/login')) {
        const redirect = encodeURIComponent(window.location.pathname + window.location.search)
        window.location.assign(`/login?redirect=${redirect}`)
      }
    }
    const text = await res.text()
    throw new Error(text || res.statusText)
  }
  return res.json() as Promise<T>
}

export async function login(username: string, password: string) {
  const data = await request<{ token: string }>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
  setToken(data.token)
  return data
}

export async function getStats() {
  const raw = await request<{
    session_count: number
    total_knowledge?: number
    by_kind?: Record<string, number>
    recent_knowledge?: unknown[]
  }>('/api/v1/stats')
  return { session_count: raw.session_count } satisfies StatsResponse
}

export type SessionSummary = {
  uuid: string
  name: string
  subtitle: string
  status: string
  unified_mode: string
  workspace_id: string
  workspace_path: string
  context_usage_percent: number | null
  input_tokens: number | null
  output_tokens: number | null
  cache_read_tokens: number | null
  cache_write_tokens: number | null
  model_config: unknown
  total_lines_added: number
  total_lines_removed: number
  files_changed_count: number
  turn_count?: number
  thinking_count?: number
}

export function listSessions(q = '') {
  const qs = q ? `?q=${encodeURIComponent(q)}` : ''
  return request<{ sessions: SessionSummary[] }>(`/api/v1/sessions${qs}`)
}

export function getSession(uuid: string) {
  return request<SessionSummary>(`/api/v1/sessions/${uuid}`)
}

export function getSessionThinking(uuid: string) {
  return request<{ thinking: Array<Record<string, unknown>> }>(`/api/v1/sessions/${uuid}/thinking`)
}

export function getSessionTurns(uuid: string) {
  return request<{ turns: Array<Record<string, unknown>> }>(`/api/v1/sessions/${uuid}/turns`)
}

export type PromptRow = {
  id: string
  prompt: string
  technically_revised: string
  english_revised: string
  harness: string
  created_at: string
  updated_at: string
}

export type PromptWrite = {
  prompt: string
  technically_revised: string
  english_revised: string
  harness?: string
}

export function listPrompts() {
  return request<{ prompts: PromptRow[] }>('/api/v1/prompts')
}

export function getPrompt(id: string) {
  return request<PromptRow>(`/api/v1/prompts/${id}`)
}

export function createPrompt(body: PromptWrite) {
  return request<{ id: string }>('/api/v1/prompts', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function updatePrompt(id: string, body: PromptWrite) {
  return request<{ ok: boolean }>(`/api/v1/prompts/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
}

export function deletePrompt(id: string) {
  return request<{ ok: boolean }>(`/api/v1/prompts/${id}`, { method: 'DELETE' })
}

const DATE_TIME_RE = /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/

export function fmtDateTime(iso: string | undefined): string {
  if (!iso) return '—'
  const trimmed = iso.trim()
  if (DATE_TIME_RE.test(trimmed)) return trimmed
  const d = new Date(trimmed.includes('T') ? trimmed : `${trimmed.replace(' ', 'T')}Z`)
  if (Number.isNaN(d.getTime())) return trimmed
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getUTCFullYear()}-${pad(d.getUTCMonth() + 1)}-${pad(d.getUTCDate())} ${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}:${pad(d.getUTCSeconds())}`
}

function clip(s: string, max: number): string {
  const t = s.trim()
  if (t.length <= max) return t || '—'
  return `${t.slice(0, max)}…`
}

export function clipPromptCell(s: string, max = 80): string {
  return clip(s, max)
}

export function fmtNum(v: number | null | undefined): string {
  if (v == null) return '—'
  return new Intl.NumberFormat().format(v)
}

export function fmtPct(v: number | null | undefined): string {
  if (v == null) return '—'
  return `${v.toFixed(1)}%`
}

