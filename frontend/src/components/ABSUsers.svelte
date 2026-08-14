<script>
  import { api } from '../lib/api.js'

  let users = $state([])
  let error = $state('')
  let testResults = $state({})

  let editingId = $state(null) // null = creating a new user
  let form = $state(emptyForm())

  function emptyForm() {
    return { label: '', baseUrl: '', apiToken: '' }
  }

  async function load() {
    try {
      users = await api.absUsers.list()
    } catch (e) {
      error = e.message
    }
  }

  function resetForm() {
    editingId = null
    form = emptyForm()
    error = ''
  }

  function startEdit(u) {
    editingId = u.id
    form = { label: u.label, baseUrl: u.baseUrl, apiToken: '' }
    error = ''
  }

  // apiToken is required to create; when editing, blank just means "keep the
  // current value" (see service.UpdateABSUser).
  function missingFields() {
    const required = editingId ? ['label', 'baseUrl'] : ['label', 'baseUrl', 'apiToken']
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
        await api.absUsers.update(editingId, form)
      } else {
        await api.absUsers.create(form)
      }
      resetForm()
      await load()
    } catch (e) {
      error = e.message
    }
  }

  async function remove(id) {
    if (!confirm('Delete this Audiobookshelf user? Profiles using it will also be removed.')) return
    await api.absUsers.remove(id)
    if (editingId === id) resetForm()
    await load()
  }

  async function test(id) {
    testResults = { ...testResults, [id]: 'testing...' }
    try {
      const libs = await api.absUsers.test(id)
      testResults = { ...testResults, [id]: `OK - libraries: ${libs.map((l) => l.name).join(', ')}` }
    } catch (e) {
      testResults = { ...testResults, [id]: `Failed: ${e.message}` }
    }
  }

  load()
</script>

<div class="space-y-6 max-w-2xl">
  <div class="rounded-lg border border-slate-800 p-4">
    <h2 class="font-medium mb-2">{editingId ? 'Edit Audiobookshelf User' : 'Add Audiobookshelf User'}</h2>
    {#if !editingId}
      <ol class="text-xs text-slate-400 mb-3 space-y-1 list-decimal list-inside">
        <li>Log into your Audiobookshelf server as the user you want to sync.</li>
        <li>Go to <strong>Settings &rarr; Users &rarr; (your account) &rarr; API Token</strong> and copy it.</li>
        <li>Enter your server's base URL (e.g. <code>https://abs.example.com</code>, or an internal Docker hostname
          like <code>http://audiobookshelf:80</code> if bookSync runs in the same Compose stack) and the token below.</li>
      </ol>
    {:else}
      <p class="text-xs text-slate-400 mb-3">Leave API token blank to keep its current value.</p>
    {/if}
    <div class="grid grid-cols-1 gap-2">
      <input class="input" placeholder="Label (e.g. Jeff)" bind:value={form.label} />
      <input class="input" placeholder="Server URL (https://abs.example.com)" bind:value={form.baseUrl} />
      <input class="input" placeholder={editingId ? 'API token (leave blank to keep)' : 'API token'} bind:value={form.apiToken} />
    </div>
    <div class="flex gap-2 mt-3">
      <button class="btn" onclick={save}>{editingId ? 'Save Changes' : 'Add User'}</button>
      {#if editingId}<button class="btn-sm" onclick={resetForm}>Cancel</button>{/if}
    </div>
    {#if error}<p class="text-red-400 text-sm mt-2">{error}</p>{/if}
  </div>

  <div class="rounded-lg border border-slate-800 divide-y divide-slate-800">
    {#each users as u (u.id)}
      <div class="p-3 flex items-center justify-between">
        <div>
          <div class="font-medium">{u.label}</div>
          <div class="text-xs text-slate-500">{u.baseUrl}</div>
          {#if testResults[u.id]}<div class="text-xs text-slate-400">{testResults[u.id]}</div>{/if}
        </div>
        <div class="flex gap-2">
          <button class="btn-sm" onclick={() => test(u.id)}>Test</button>
          <button class="btn-sm" onclick={() => startEdit(u)}>Edit</button>
          <button class="btn-sm-danger" onclick={() => remove(u.id)}>Delete</button>
        </div>
      </div>
    {:else}
      <p class="p-3 text-sm text-slate-400">No Audiobookshelf users yet.</p>
    {/each}
  </div>
</div>
