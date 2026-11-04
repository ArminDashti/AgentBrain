function envOr(key: string, fallback: string): string {
  const v = process.env[key]
  return v && v.trim() !== '' ? v.trim() : fallback
}

export type PromptWrite = {
  prompt: string
  technically_revised?: string
  english_revised?: string
  harness?: string
}

export class BrainClient {
  readonly baseUrl: string
  readonly username: string
  readonly password: string
  private token: string | null = null

  constructor(opts?: { baseUrl?: string; username?: string; password?: string }) {
    this.baseUrl = (opts?.baseUrl ?? envOr('AGENT_BRAIN_API_URL', 'http://127.0.0.1:5090')).replace(
      /\/$/,
      '',
    )
    this.username = opts?.username ?? envOr('AGENT_BRAIN_USERNAME', 'armin')
    this.password = opts?.password ?? envOr('AGENT_BRAIN_PASSWORD', 'dopadopa123')
  }

  async login(username?: string, password?: string) {
    const res = await fetch(`${this.baseUrl}/api/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: username ?? this.username,
        password: password ?? this.password,
      }),
    })
    if (!res.ok) throw new Error(`login ${res.status}: ${await res.text()}`)
    const data = (await res.json()) as { token: string }
    this.token = data.token
    return data
  }

  private async authed(path: string, init: RequestInit = {}) {
    if (!this.token) await this.login()
    const headers = new Headers(init.headers)
    headers.set('Authorization', `Bearer ${this.token}`)
    if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
    const res = await fetch(`${this.baseUrl}${path}`, { ...init, headers })
    if (res.status === 401) {
      await this.login()
      headers.set('Authorization', `Bearer ${this.token}`)
      const retry = await fetch(`${this.baseUrl}${path}`, { ...init, headers })
      if (!retry.ok) throw new Error(`${path} ${retry.status}: ${await retry.text()}`)
      return retry.json()
    }
    if (!res.ok) throw new Error(`${path} ${res.status}: ${await res.text()}`)
    return res.json()
  }

  createPrompt(body: PromptWrite) {
    return this.authed('/api/v1/prompts', {
      method: 'POST',
      body: JSON.stringify({
        prompt: body.prompt,
        technically_revised: body.technically_revised ?? '',
        english_revised: body.english_revised ?? '',
        harness: body.harness ?? '',
      }),
    })
  }
}
