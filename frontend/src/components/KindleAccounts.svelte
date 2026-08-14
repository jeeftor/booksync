<script>
  import { api } from '../lib/api.js'

  let accounts = $state([])
  let error = $state('')
  let testResults = $state({})

  let editingId = $state(null) // null = creating a new account
  let form = $state(emptyForm())
  let rawCookieHeader = $state('')
  let cookieParseStatus = $state(null) // { ok: bool, message: string } | null
  let rawDeviceTokenInput = $state('')
  let deviceTokenParseStatus = $state(null)

  function emptyForm() {
    return {
      label: '',
      ubidMain: '',
      atMain: '',
      sessionId: '',
      xMain: '',
      deviceToken: '',
      tlsProxyUrl: 'http://booksync-tls-proxy:8080',
      tlsProxyKey: '',
    }
  }

  async function load() {
    try {
      accounts = await api.kindleAccounts.list()
    } catch (e) {
      error = e.message
    }
  }

  // Cookie-Editor (and similar extensions) can export in 3 different shapes;
  // detect which one was pasted and normalize to a { cookieName: value } map.
  function detectAndParseCookies(raw) {
    const trimmed = raw.trim()
    if (!trimmed) return { map: {}, format: '' }

    // JSON export: an array (or single object) of {name, value, ...}.
    if (trimmed.startsWith('[') || trimmed.startsWith('{')) {
      try {
        const parsed = JSON.parse(trimmed)
        const arr = Array.isArray(parsed) ? parsed : [parsed]
        const map = {}
        for (const c of arr) {
          if (c && typeof c.name === 'string' && typeof c.value === 'string') {
            map[c.name.toLowerCase()] = c.value
          }
        }
        if (Object.keys(map).length) return { map, format: 'JSON export' }
      } catch {
        /* not valid JSON - fall through to other formats */
      }
    }

    // Netscape cookies.txt format: tab-separated fields, 7 per line, one
    // cookie per line, comment lines start with `#`.
    const lines = trimmed.split('\n').map((l) => l.trim()).filter(Boolean)
    const netscapeLines = lines.filter((l) => !l.startsWith('#') && l.split('\t').length >= 7)
    if (netscapeLines.length) {
      const map = {}
      for (const line of netscapeLines) {
        const parts = line.split('\t')
        const name = parts[5]
        const value = parts[6]
        if (name) map[name.toLowerCase()] = (value ?? '').trim()
      }
      if (Object.keys(map).length) return { map, format: 'Netscape cookies.txt' }
    }

    // Header string: "name=value; name2=value2" (optionally prefixed with
    // "Cookie: "), the most common export/DevTools-copy format.
    const cleaned = trimmed.replace(/^cookie:\s*/i, '')
    const map = {}
    for (const part of cleaned.split(';')) {
      const idx = part.indexOf('=')
      if (idx === -1) continue
      map[part.slice(0, idx).trim().toLowerCase()] = part.slice(idx + 1).trim()
    }
    return { map, format: 'Header string' }
  }

  function parseCookieHeader() {
    const { map, format } = detectAndParseCookies(rawCookieHeader)
    const wanted = { 'ubid-main': 'ubidMain', 'at-main': 'atMain', 'session-id': 'sessionId', 'x-main': 'xMain' }
    let found = 0
    for (const [cookieName, field] of Object.entries(wanted)) {
      if (map[cookieName]) {
        form[field] = map[cookieName]
        found++
      }
    }
    if (!format) {
      cookieParseStatus = { ok: false, message: 'Paste something first' }
    } else if (found === 4) {
      cookieParseStatus = { ok: true, message: `Found all 4 cookies (detected ${format})` }
    } else {
      cookieParseStatus = {
        ok: false,
        message: `Found ${found}/4 (detected ${format}) — need ubid-main, at-main, session-id, x-main`,
      }
    }
  }

  // Accepts either the bare token value, or the full getDeviceToken request
  // URL/query string copied from DevTools, and extracts serialNumber from it.
  function parseDeviceToken() {
    const raw = rawDeviceTokenInput.trim()
    if (!raw) {
      deviceTokenParseStatus = { ok: false, message: 'Paste something first' }
      return
    }
    const qIndex = raw.indexOf('?')
    if (qIndex !== -1 || raw.includes('serialNumber=')) {
      const params = new URLSearchParams(qIndex !== -1 ? raw.slice(qIndex + 1) : raw)
      const serial = params.get('serialNumber')
      if (serial) {
        form.deviceToken = serial
        deviceTokenParseStatus = { ok: true, message: 'Extracted serialNumber from the request URL' }
        return
      }
    }
    form.deviceToken = raw
    deviceTokenParseStatus = { ok: true, message: 'Using pasted value as-is' }
  }

  function resetForm() {
    editingId = null
    form = emptyForm()
    rawCookieHeader = ''
    cookieParseStatus = null
    rawDeviceTokenInput = ''
    deviceTokenParseStatus = null
    error = ''
  }

  function startEdit(acc) {
    editingId = acc.id
    form = {
      label: acc.label,
      ubidMain: '',
      atMain: '',
      sessionId: '',
      xMain: '',
      deviceToken: '',
      tlsProxyUrl: acc.tlsProxyUrl,
      tlsProxyKey: '',
    }
    rawCookieHeader = ''
    cookieParseStatus = null
    rawDeviceTokenInput = ''
    deviceTokenParseStatus = null
    error = ''
  }

  // Required for a brand-new account; when editing, secret fields left blank
  // just mean "keep the current value" (see service.UpdateKindleAccount).
  function missingFields() {
    const always = ['label', 'tlsProxyUrl']
    const onlyOnCreate = ['ubidMain', 'atMain', 'sessionId', 'xMain', 'deviceToken', 'tlsProxyKey']
    const required = editingId ? always : [...always, ...onlyOnCreate]
    return required.filter((f) => !form[f]?.trim())
  }

  async function save() {
    error = ''
    const missing = missingFields()
    if (missing.length) {
      error = `Missing required field(s): ${missing.join(', ')}`
      return
    }
    try {
      if (editingId) {
        await api.kindleAccounts.update(editingId, form)
      } else {
        await api.kindleAccounts.create(form)
      }
      resetForm()
      await load()
    } catch (e) {
      error = e.message
    }
  }

  async function remove(id) {
    if (!confirm('Delete this Kindle account? Profiles using it will also be removed.')) return
    await api.kindleAccounts.remove(id)
    if (editingId === id) resetForm()
    await load()
  }

  async function test(id) {
    testResults = { ...testResults, [id]: 'testing...' }
    try {
      const res = await api.kindleAccounts.test(id)
      testResults = { ...testResults, [id]: `OK - ${res.bookCount} books found` }
    } catch (e) {
      testResults = { ...testResults, [id]: `Failed: ${e.message}` }
    }
  }

  load()
</script>

<div class="space-y-6 max-w-3xl">
  <div class="rounded-xl border border-slate-800 bg-gradient-to-b from-slate-900/60 to-slate-900/20 p-5 space-y-5 shadow-lg shadow-black/20">
    <div class="flex items-center gap-2">
      <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-amber-500/10 text-amber-400 ring-1 ring-amber-500/30">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
      </span>
      <h2 class="font-medium text-lg">{editingId ? 'Edit Kindle Account' : 'Add Kindle Account'}</h2>
    </div>

    {#if !editingId}
      {@const steps = [
        {
          title: 'Install a cookie-export extension',
          body: 'e.g. Cookie-Editor (Chrome/Firefox, open source).',
          link: { href: 'https://cookie-editor.com/', label: 'cookie-editor.com' },
        },
        {
          title: 'Log into read.amazon.com',
          body: "Open your library in that browser so you're fully signed in.",
          link: { href: 'https://read.amazon.com', label: 'read.amazon.com' },
        },
        {
          title: 'Export cookies and parse them below',
          body: 'Cookie-Editor icon → Export → any format (Header String, JSON, or Netscape all work — auto-detected).',
        },
        {
          title: 'Grab the device token from DevTools',
          body: 'F12 → Network tab → reload → find getDeviceToken?serialNumber=...&deviceType=.... Paste the full URL below.',
        },
        {
          title: 'Confirm the TLS proxy settings',
          body: "Defaults are correct for this homelab's booksync-tls-proxy sidecar — just fill in its API key.",
          link: { href: 'https://github.com/bogdanfinn/tls-client-api', label: 'tls-client-api' },
        },
      ]}
      <ol class="space-y-3">
        {#each steps as step, i}
          <li class="flex gap-3">
            <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-slate-800 text-xs font-semibold text-slate-300 ring-1 ring-slate-700">
              {i + 1}
            </span>
            <div class="text-sm">
              <div class="font-medium text-slate-200">{step.title}</div>
              <div class="text-slate-400">
                {step.body}
                {#if step.link}
                  &mdash; <a class="text-sky-400 underline decoration-sky-800 hover:decoration-sky-400" href={step.link.href} target="_blank" rel="noreferrer">{step.link.label}</a>
                {/if}
              </div>
            </div>
          </li>
        {/each}
      </ol>

      <div class="grid gap-3 sm:grid-cols-2">
        <div class="rounded-lg border border-slate-800 bg-slate-950/40 p-3 space-y-2">
          <div class="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-slate-400">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="8" y="2" width="8" height="4" rx="1"/><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/></svg>
            Cookie export
          </div>
          <textarea
            class="input w-full font-mono text-xs resize-y min-h-[4.5rem]"
            placeholder="ubid-main=...; at-main=...; session-id=...; x-main=...&#10;(or JSON / Netscape cookies.txt — auto-detected)"
            bind:value={rawCookieHeader}
          ></textarea>
          <div class="flex items-center justify-between gap-2">
            <button class="btn-sm" onclick={parseCookieHeader}>Parse</button>
            {#if cookieParseStatus}
              <span class="text-xs {cookieParseStatus.ok ? 'text-emerald-400' : 'text-amber-400'}">{cookieParseStatus.message}</span>
            {/if}
          </div>
        </div>

        <div class="rounded-lg border border-slate-800 bg-slate-950/40 p-3 space-y-2">
          <div class="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-slate-400">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="2"/><path d="M16.24 7.76a6 6 0 0 1 0 8.49m-8.48-8.49a6 6 0 0 0 0 8.49M19.07 4.93a10 10 0 0 1 0 14.14M4.93 4.93a10 10 0 0 0 0 14.14"/></svg>
            Device token
          </div>
          <input
            class="input w-full font-mono text-xs"
            placeholder="https://read.amazon.com/.../getDeviceToken?serialNumber=...&deviceType=..."
            bind:value={rawDeviceTokenInput}
          />
          <div class="flex items-center justify-between gap-2">
            <button class="btn-sm" onclick={parseDeviceToken}>Parse</button>
            {#if deviceTokenParseStatus}
              <span class="text-xs {deviceTokenParseStatus.ok ? 'text-emerald-400' : 'text-amber-400'}">{deviceTokenParseStatus.message}</span>
            {/if}
          </div>
        </div>
      </div>
    {:else}
      <p class="text-xs text-slate-400">
        Leave any cookie/device-token/TLS-key field blank to keep its current value — only fill in what you want to change.
      </p>
    {/if}

    <div class="grid grid-cols-2 gap-2 border-t border-slate-800 pt-4">
      <input class="input col-span-2" placeholder="Label (e.g. Family Kindle)" bind:value={form.label} />
      <input class="input" placeholder={editingId ? 'ubid-main (leave blank to keep)' : 'ubid-main'} bind:value={form.ubidMain} />
      <input class="input" placeholder={editingId ? 'at-main (leave blank to keep)' : 'at-main'} bind:value={form.atMain} />
      <input class="input" placeholder={editingId ? 'session-id (leave blank to keep)' : 'session-id'} bind:value={form.sessionId} />
      <input class="input" placeholder={editingId ? 'x-main (leave blank to keep)' : 'x-main'} bind:value={form.xMain} />
      <input class="input col-span-2" placeholder={editingId ? 'Device token (leave blank to keep)' : 'Device token (serialNumber)'} bind:value={form.deviceToken} />
      <input class="input" placeholder="TLS proxy URL" bind:value={form.tlsProxyUrl} />
      <input class="input" placeholder={editingId ? 'TLS proxy API key (leave blank to keep)' : 'TLS proxy API key'} bind:value={form.tlsProxyKey} />
    </div>
    <div class="flex gap-2">
      <button class="btn" onclick={save}>{editingId ? 'Save Changes' : 'Add Account'}</button>
      {#if editingId}<button class="btn-sm" onclick={resetForm}>Cancel</button>{/if}
    </div>
    {#if error}<p class="text-red-400 text-sm">{error}</p>{/if}
  </div>

  <div class="rounded-lg border border-slate-800 divide-y divide-slate-800">
    {#each accounts as acc (acc.id)}
      <div class="p-3 flex items-center justify-between">
        <div>
          <div class="font-medium">{acc.label || `Account #${acc.id}`}</div>
          {#if testResults[acc.id]}<div class="text-xs text-slate-400">{testResults[acc.id]}</div>{/if}
        </div>
        <div class="flex gap-2">
          <button class="btn-sm" onclick={() => test(acc.id)}>Test</button>
          <button class="btn-sm" onclick={() => startEdit(acc)}>Edit</button>
          <button class="btn-sm-danger" onclick={() => remove(acc.id)}>Delete</button>
        </div>
      </div>
    {:else}
      <p class="p-3 text-sm text-slate-400">No Kindle accounts yet.</p>
    {/each}
  </div>
</div>
