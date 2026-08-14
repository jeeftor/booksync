const base = '/api'

async function req(path, opts = {}) {
  const res = await fetch(base + path, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  })
  if (!res.ok) {
    let message = res.statusText
    try {
      const body = await res.json()
      message = body.error || message
    } catch {
      /* ignore */
    }
    throw new Error(message)
  }
  if (res.status === 204) return null
  return res.json()
}

const get = (path) => req(path)
const post = (path, body) => req(path, { method: 'POST', body: JSON.stringify(body) })
const put = (path, body) => req(path, { method: 'PUT', body: JSON.stringify(body) })
const del = (path) => req(path, { method: 'DELETE' })

export const api = {
  health: () => get('/health'),

  kindleAccounts: {
    list: () => get('/kindle-accounts'),
    defaults: () => get('/kindle-accounts/defaults'),
    create: (body) => post('/kindle-accounts', body),
    update: (id, body) => put(`/kindle-accounts/${id}`, body),
    remove: (id) => del(`/kindle-accounts/${id}`),
    test: (id) => post(`/kindle-accounts/${id}/test`),
    testDraft: (body) => post('/kindle-accounts/test', body),
  },

  absUsers: {
    list: () => get('/abs-users'),
    create: (body) => post('/abs-users', body),
    update: (id, body) => put(`/abs-users/${id}`, body),
    remove: (id) => del(`/abs-users/${id}`),
    test: (id) => post(`/abs-users/${id}/test`),
  },

  profiles: {
    list: () => get('/profiles'),
    create: (body) => post('/profiles', body),
    update: (id, body) => put(`/profiles/${id}`, body),
    remove: (id) => del(`/profiles/${id}`),
    suggestions: (id) => get(`/profiles/${id}/suggestions`),
    mappings: (id) => get(`/profiles/${id}/mappings`),
    confirmMatch: (id, candidate) => post(`/profiles/${id}/mappings`, candidate),
    rejectMatch: (id, candidate) => post(`/profiles/${id}/suggestions/reject`, candidate),
    sync: (id) => post(`/profiles/${id}/sync`),
  },

  mappings: {
    remove: (id) => del(`/mappings/${id}`),
    sync: (id) => post(`/mappings/${id}/sync`),
    history: (id) => get(`/mappings/${id}/history`),
  },

  activity: (limit = 50) => get(`/activity?limit=${limit}`),
}
