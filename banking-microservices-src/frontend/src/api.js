// Thin API client for the gateway. All calls go to /api/v1/* (the gateway
// proxies to backend services). The access token is attached automatically.

const BASE = '/api/v1'

export function getToken() {
  return localStorage.getItem('access_token') || ''
}

export function setTokens(tokens) {
  if (tokens?.access_token) localStorage.setItem('access_token', tokens.access_token)
  if (tokens?.refresh_token) localStorage.setItem('refresh_token', tokens.refresh_token)
}

export function clearTokens() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
}

// request performs a JSON API call, throwing an Error with the API message on
// non-2xx responses.
export async function request(method, path, body, extraHeaders = {}) {
  const headers = { 'Content-Type': 'application/json', ...extraHeaders }
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })

  const text = await res.text()
  const data = text ? JSON.parse(text) : null
  if (!res.ok) {
    const msg = data?.message || `request failed (${res.status})`
    const err = new Error(msg)
    err.status = res.status
    throw err
  }
  return data
}

export const api = {
  get: (p, h) => request('GET', p, null, h),
  post: (p, b, h) => request('POST', p, b, h),
  put: (p, b, h) => request('PUT', p, b, h),
  del: (p, h) => request('DELETE', p, null, h),
}
