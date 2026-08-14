<script>
  import { api } from '../lib/api.js'

  let accounts = $state([])
  let error = $state('')
  let testResults = $state({})

  let form = $state(emptyForm())

  function emptyForm() {
    return {
      label: '',
      ubidMain: '',
      atMain: '',
      sessionId: '',
      xMain: '',
      deviceToken: '',
      tlsProxyUrl: 'http://localhost:8080',
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

  async function create() {
    error = ''
    try {
      await api.kindleAccounts.create(form)
      form = emptyForm()
      await load()
    } catch (e) {
      error = e.message
    }
  }

  async function remove(id) {
    if (!confirm('Delete this Kindle account? Profiles using it will also be removed.')) return
    await api.kindleAccounts.remove(id)
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
  <div class="rounded-lg border border-slate-800 p-4">
    <h2 class="font-medium mb-2">Add Kindle Account</h2>
    <p class="text-xs text-slate-400 mb-3">
      Cookies and device token come from read.amazon.com's devtools (Network tab). See the README for
      how to extract them. The TLS proxy is a separate <code>tls-client-api</code> container required to
      talk to Amazon at all.
    </p>
    <div class="grid grid-cols-2 gap-2">
      <input class="input" placeholder="Label (e.g. Family Kindle)" bind:value={form.label} />
      <input class="input" placeholder="Device token" bind:value={form.deviceToken} />
      <input class="input" placeholder="ubid-main" bind:value={form.ubidMain} />
      <input class="input" placeholder="at-main" bind:value={form.atMain} />
      <input class="input" placeholder="session-id" bind:value={form.sessionId} />
      <input class="input" placeholder="x-main" bind:value={form.xMain} />
      <input class="input" placeholder="TLS proxy URL" bind:value={form.tlsProxyUrl} />
      <input class="input" placeholder="TLS proxy API key" bind:value={form.tlsProxyKey} />
    </div>
    <button class="btn mt-3" onclick={create}>Add Account</button>
    {#if error}<p class="text-red-400 text-sm mt-2">{error}</p>{/if}
  </div>

  <div class="rounded-lg border border-slate-800 divide-y divide-slate-800">
    {#each accounts as acc (acc.id)}
      <div class="p-3 flex items-center justify-between">
        <div>
          <div class="font-medium">{acc.label}</div>
          {#if testResults[acc.id]}<div class="text-xs text-slate-400">{testResults[acc.id]}</div>{/if}
        </div>
        <div class="flex gap-2">
          <button class="btn-sm" onclick={() => test(acc.id)}>Test</button>
          <button class="btn-sm-danger" onclick={() => remove(acc.id)}>Delete</button>
        </div>
      </div>
    {:else}
      <p class="p-3 text-sm text-slate-400">No Kindle accounts yet.</p>
    {/each}
  </div>
</div>
