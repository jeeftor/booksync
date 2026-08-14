<script>
  import { api } from '../lib/api.js'

  let { onOpenMappings } = $props()

  let profiles = $state([])
  let kindleAccounts = $state([])
  let absUsers = $state([])
  let error = $state('')
  let syncStatus = $state({})

  let form = $state({ label: '', kindleAccountId: '', absUserId: '', absLibraryId: '', pollMinutes: 15 })

  async function load() {
    try {
      ;[profiles, kindleAccounts, absUsers] = await Promise.all([
        api.profiles.list(),
        api.kindleAccounts.list(),
        api.absUsers.list(),
      ])
    } catch (e) {
      error = e.message
    }
  }

  function labelFor(list, id) {
    return list.find((x) => x.id === id)?.label ?? `#${id}`
  }

  async function create() {
    error = ''
    try {
      await api.profiles.create({
        label: form.label,
        kindleAccountId: Number(form.kindleAccountId),
        absUserId: Number(form.absUserId),
        absLibraryId: form.absLibraryId,
        pollMinutes: Number(form.pollMinutes) || 15,
      })
      form = { label: '', kindleAccountId: '', absUserId: '', absLibraryId: '', pollMinutes: 15 }
      await load()
    } catch (e) {
      error = e.message
    }
  }

  async function remove(id) {
    if (!confirm('Delete this profile and its book mappings?')) return
    await api.profiles.remove(id)
    await load()
  }

  async function syncNow(id) {
    syncStatus = { ...syncStatus, [id]: 'syncing...' }
    try {
      await api.profiles.sync(id)
      syncStatus = { ...syncStatus, [id]: 'synced' }
    } catch (e) {
      syncStatus = { ...syncStatus, [id]: `error: ${e.message}` }
    }
  }

  load()
</script>

<div class="space-y-6 max-w-3xl">
  <div class="rounded-lg border border-slate-800 p-4">
    <h2 class="font-medium mb-2">New Profile</h2>
    <p class="text-xs text-slate-400 mb-3">
      A profile pairs one Kindle account with one Audiobookshelf user and owns its own confirmed book
      mappings. Use one profile per person. Add a Kindle account and an Audiobookshelf user first (see
      the other two tabs), then create a profile here, then open <strong>Mappings</strong> to find and
      confirm book pairings and sync.
    </p>
    <div class="grid grid-cols-2 gap-2">
      <input class="input" placeholder="Label (e.g. Jeff)" bind:value={form.label} />
      <input class="input" placeholder="Poll interval (minutes)" type="number" bind:value={form.pollMinutes} />
      <select class="input" bind:value={form.kindleAccountId}>
        <option value="">Select Kindle account</option>
        {#each kindleAccounts as a}<option value={a.id}>{a.label}</option>{/each}
      </select>
      <select class="input" bind:value={form.absUserId}>
        <option value="">Select Audiobookshelf user</option>
        {#each absUsers as u}<option value={u.id}>{u.label}</option>{/each}
      </select>
      <input class="input col-span-2" placeholder="ABS library ID (optional, defaults to first library)" bind:value={form.absLibraryId} />
    </div>
    <button class="btn mt-3" onclick={create}>Create Profile</button>
    {#if error}<p class="text-red-400 text-sm mt-2">{error}</p>{/if}
  </div>

  <div class="rounded-lg border border-slate-800 divide-y divide-slate-800">
    {#each profiles as p (p.id)}
      <div class="p-3 flex items-center justify-between">
        <div>
          <div class="font-medium">{p.label}</div>
          <div class="text-xs text-slate-500">
            Kindle: {labelFor(kindleAccounts, p.kindleAccountId)} · ABS: {labelFor(absUsers, p.absUserId)} · every {p.pollMinutes}m
          </div>
          {#if syncStatus[p.id]}<div class="text-xs text-slate-400">{syncStatus[p.id]}</div>{/if}
        </div>
        <div class="flex gap-2">
          <button class="btn-sm" onclick={() => onOpenMappings(p.id)}>Mappings</button>
          <button class="btn-sm" onclick={() => syncNow(p.id)}>Sync Now</button>
          <button class="btn-sm-danger" onclick={() => remove(p.id)}>Delete</button>
        </div>
      </div>
    {:else}
      <p class="p-3 text-sm text-slate-400">No profiles yet.</p>
    {/each}
  </div>
</div>
