<script>
  import { api } from '../lib/api.js'

  let users = $state([])
  let error = $state('')
  let testResults = $state({})
  let form = $state({ label: '', baseUrl: '', apiToken: '' })

  async function load() {
    try {
      users = await api.absUsers.list()
    } catch (e) {
      error = e.message
    }
  }

  async function create() {
    error = ''
    try {
      await api.absUsers.create(form)
      form = { label: '', baseUrl: '', apiToken: '' }
      await load()
    } catch (e) {
      error = e.message
    }
  }

  async function remove(id) {
    if (!confirm('Delete this Audiobookshelf user? Profiles using it will also be removed.')) return
    await api.absUsers.remove(id)
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
    <h2 class="font-medium mb-2">Add Audiobookshelf User</h2>
    <ol class="text-xs text-slate-400 mb-3 space-y-1 list-decimal list-inside">
      <li>Log into your Audiobookshelf server as the user you want to sync.</li>
      <li>Go to <strong>Settings &rarr; Users &rarr; (your account) &rarr; API Token</strong> and copy it.</li>
      <li>Enter your server's base URL (e.g. <code>https://abs.example.com</code>, or an internal Docker hostname
        like <code>http://audiobookshelf:80</code> if bookSync runs in the same Compose stack) and the token below.</li>
    </ol>
    <div class="grid grid-cols-1 gap-2">
      <input class="input" placeholder="Label (e.g. Jeff)" bind:value={form.label} />
      <input class="input" placeholder="Server URL (https://abs.example.com)" bind:value={form.baseUrl} />
      <input class="input" placeholder="API token" bind:value={form.apiToken} />
    </div>
    <button class="btn mt-3" onclick={create}>Add User</button>
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
          <button class="btn-sm-danger" onclick={() => remove(u.id)}>Delete</button>
        </div>
      </div>
    {:else}
      <p class="p-3 text-sm text-slate-400">No Audiobookshelf users yet.</p>
    {/each}
  </div>
</div>
