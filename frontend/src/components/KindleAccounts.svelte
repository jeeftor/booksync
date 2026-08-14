<script>
  import { api } from '../lib/api.js'

  let accounts = $state([])
  let error = $state('')
  let testResults = $state({})

  let editingId = $state(null) // null = creating a new account
  let form = $state(emptyForm())
  let rawCookieHeader = $state('')
  let cookieParseStatus = $state('')
  let rawDeviceTokenInput = $state('')
  let deviceTokenParseStatus = $state('')

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

  // Accepts whatever a cookie-export extension (e.g. Cookie-Editor's
  // "Export > Header String") or a pasted `Cookie:` request header gives you,
  // and pulls out just the 4 fields bookSync needs.
  function parseCookieHeader() {
    const cleaned = rawCookieHeader.trim().replace(/^cookie:\s*/i, '')
    if (!cleaned) return
    const map = {}
    for (const part of cleaned.split(';')) {
      const idx = part.indexOf('=')
      if (idx === -1) continue
      map[part.slice(0, idx).trim().toLowerCase()] = part.slice(idx + 1).trim()
    }
    const wanted = { 'ubid-main': 'ubidMain', 'at-main': 'atMain', 'session-id': 'sessionId', 'x-main': 'xMain' }
    let found = 0
    for (const [cookieName, field] of Object.entries(wanted)) {
      if (map[cookieName]) {
        form[field] = map[cookieName]
        found++
      }
    }
    cookieParseStatus =
      found === 4
        ? '✓ Found all 4 cookies'
        : `Found ${found}/4 — make sure you copied the whole Cookie header from read.amazon.com (need ubid-main, at-main, session-id, x-main)`
  }

  // Accepts either the bare token value, or the full getDeviceToken request
  // URL/query string copied from DevTools, and extracts serialNumber from it.
  function parseDeviceToken() {
    const raw = rawDeviceTokenInput.trim()
    if (!raw) return
    const qIndex = raw.indexOf('?')
    if (qIndex !== -1 || raw.includes('serialNumber=')) {
      const params = new URLSearchParams(qIndex !== -1 ? raw.slice(qIndex + 1) : raw)
      const serial = params.get('serialNumber')
      if (serial) {
        form.deviceToken = serial
        deviceTokenParseStatus = '✓ Extracted serialNumber'
        return
      }
    }
    form.deviceToken = raw
    deviceTokenParseStatus = '✓ Using pasted value as-is'
  }

  function resetForm() {
    editingId = null
    form = emptyForm()
    rawCookieHeader = ''
    cookieParseStatus = ''
    rawDeviceTokenInput = ''
    deviceTokenParseStatus = ''
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
    cookieParseStatus = ''
    rawDeviceTokenInput = ''
    deviceTokenParseStatus = ''
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
  <div class="rounded-lg border border-slate-800 p-4 space-y-4">
    <h2 class="font-medium">{editingId ? 'Edit Kindle Account' : 'Add Kindle Account'}</h2>

    {#if !editingId}
      <ol class="text-sm text-slate-400 space-y-3 list-decimal list-inside">
        <li>
          Install a cookie-export browser extension — e.g.
          <a class="text-sky-400 underline" href="https://cookie-editor.com/" target="_blank" rel="noreferrer">Cookie-Editor</a>
          (Chrome/Firefox, open source).
        </li>
        <li>
          Log into <a class="text-sky-400 underline" href="https://read.amazon.com" target="_blank" rel="noreferrer">read.amazon.com</a>
          in that browser (open your library so you're fully signed in).
        </li>
        <li>
          Click the Cookie-Editor icon &rarr; <strong>Export</strong> &rarr; <strong>Header String</strong> (copies a single
          <code>ubid-main=...; at-main=...; session-id=...; x-main=...</code> string). Paste it below and click Parse — it'll
          fill in the 4 cookie fields for you.
        </li>
        <li>
          Open DevTools (F12) &rarr; <strong>Network</strong> tab, reload the page, and find a request to
          <code>getDeviceToken?serialNumber=...&amp;deviceType=...</code>. Paste that full request URL (or just the
          <code>serialNumber</code> value) into the device token box below and click Parse.
        </li>
        <li>
          Leave the TLS proxy fields as the defaults if you're running bookSync in this homelab's Docker Compose stack
          (<code>booksync-tls-proxy</code> sidecar) — but you still need to fill in its <strong>TLS proxy API key</strong>
          yourself (the <code>AUTH_KEYS</code>/<code>api_auth_keys</code> value it was configured with). Otherwise point
          the URL at your own
          <a class="text-sky-400 underline" href="https://github.com/bogdanfinn/tls-client-api" target="_blank" rel="noreferrer">tls-client-api</a>
          instance.
        </li>
      </ol>

      <div class="space-y-2 border-t border-slate-800 pt-3">
        <label class="text-xs text-slate-400" for="raw-cookie">Paste Cookie header (from Cookie-Editor or DevTools)</label>
        <div class="flex gap-2">
          <input id="raw-cookie" class="input flex-1" placeholder="ubid-main=...; at-main=...; session-id=...; x-main=..." bind:value={rawCookieHeader} />
          <button class="btn-sm" onclick={parseCookieHeader}>Parse</button>
        </div>
        {#if cookieParseStatus}<p class="text-xs text-slate-400">{cookieParseStatus}</p>{/if}
      </div>

      <div class="space-y-2">
        <label class="text-xs text-slate-400" for="raw-device-token">Paste getDeviceToken request URL (or bare serialNumber)</label>
        <div class="flex gap-2">
          <input id="raw-device-token" class="input flex-1" placeholder="https://read.amazon.com/service/web/register/getDeviceToken?serialNumber=...&deviceType=..." bind:value={rawDeviceTokenInput} />
          <button class="btn-sm" onclick={parseDeviceToken}>Parse</button>
        </div>
        {#if deviceTokenParseStatus}<p class="text-xs text-slate-400">{deviceTokenParseStatus}</p>{/if}
      </div>
    {:else}
      <p class="text-xs text-slate-400">
        Leave any cookie/device-token/TLS-key field blank to keep its current value — only fill in what you want to change.
      </p>
    {/if}

    <div class="grid grid-cols-2 gap-2 border-t border-slate-800 pt-3">
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
