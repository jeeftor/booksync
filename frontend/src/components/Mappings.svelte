<script>
  import { api } from '../lib/api.js'

  let { profileId = $bindable(null) } = $props()

  let profiles = $state([])
  let suggestions = $state([])
  let mappings = $state([])
  let error = $state('')
  let loadingSuggestions = $state(false)
  let actionStatus = $state({})

  async function loadProfiles() {
    profiles = await api.profiles.list()
    if (!profileId && profiles.length) profileId = profiles[0].id
  }

  async function loadMappings() {
    if (!profileId) return
    error = ''
    try {
      mappings = await api.profiles.mappings(profileId)
    } catch (e) {
      error = e.message
    }
  }

  async function findMatches() {
    if (!profileId) return
    loadingSuggestions = true
    error = ''
    try {
      suggestions = await api.profiles.suggestions(profileId)
    } catch (e) {
      error = e.message
    } finally {
      loadingSuggestions = false
    }
  }

  async function confirmCandidate(candidate) {
    try {
      await api.profiles.confirmMatch(profileId, candidate)
      suggestions = suggestions.filter((s) => s.KindleASIN !== candidate.KindleASIN)
      await loadMappings()
    } catch (e) {
      error = e.message
    }
  }

  async function rejectCandidate(candidate) {
    try {
      await api.profiles.rejectMatch(profileId, candidate)
      suggestions = suggestions.filter(
        (s) => !(s.KindleASIN === candidate.KindleASIN && s.ABSItemID === candidate.ABSItemID)
      )
    } catch (e) {
      error = e.message
    }
  }

  async function syncOne(id) {
    actionStatus = { ...actionStatus, [id]: 'syncing...' }
    try {
      const event = await api.mappings.sync(id)
      actionStatus = { ...actionStatus, [id]: event.direction }
      await loadMappings()
    } catch (e) {
      actionStatus = { ...actionStatus, [id]: `error: ${e.message}` }
    }
  }

  async function removeMapping(id) {
    if (!confirm('Remove this mapping?')) return
    await api.mappings.remove(id)
    await loadMappings()
  }

  $effect(() => {
    if (profileId) loadMappings()
  })

  loadProfiles()
</script>

<div class="space-y-6 max-w-4xl">
  <div class="flex items-center gap-2">
    <label class="text-sm text-slate-400" for="profile-select">Profile</label>
    <select id="profile-select" class="input w-64" bind:value={profileId}>
      {#each profiles as p}<option value={p.id}>{p.label}</option>{/each}
    </select>
    <button class="btn-sm" onclick={findMatches} disabled={!profileId || loadingSuggestions}>
      {loadingSuggestions ? 'Finding matches...' : 'Find New Matches'}
    </button>
  </div>

  {#if error}<p class="text-red-400 text-sm">{error}</p>{/if}

  {#if suggestions.length}
    <div class="rounded-lg border border-slate-800">
      <h2 class="font-medium p-3 border-b border-slate-800">Suggested Matches</h2>
      <div class="divide-y divide-slate-800">
        {#each suggestions as s (s.KindleASIN)}
          <div class="p-3 flex items-center justify-between">
            <div class="text-sm">
              <div><span class="text-slate-400">Kindle:</span> {s.KindleTitle}</div>
              <div><span class="text-slate-400">ABS:</span> {s.ABSTitle}</div>
              <div class="text-xs text-slate-500">Confidence: {(s.Confidence * 100).toFixed(0)}%</div>
            </div>
            <div class="flex gap-2">
              <button class="btn-sm" onclick={() => confirmCandidate(s)}>Confirm</button>
              <button class="btn-sm-danger" onclick={() => rejectCandidate(s)}>Reject</button>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <div class="rounded-lg border border-slate-800">
    <h2 class="font-medium p-3 border-b border-slate-800">Confirmed Mappings</h2>
    <div class="divide-y divide-slate-800">
      {#each mappings as m (m.id)}
        <div class="p-3 flex items-center justify-between">
          <div class="text-sm">
            <div class="font-medium">{m.kindleTitle || m.kindleAsin} &harr; {m.absTitle || m.absItemId}</div>
            <div class="text-xs text-slate-500">
              Kindle {m.lastKindlePct.toFixed(1)}% · ABS {m.lastAbsPct.toFixed(1)}%
              {#if m.lastSynced}· last synced {new Date(m.lastSynced).toLocaleString()}{/if}
            </div>
            {#if actionStatus[m.id]}<div class="text-xs text-slate-400">{actionStatus[m.id]}</div>{/if}
          </div>
          <div class="flex gap-2">
            <button class="btn-sm" onclick={() => syncOne(m.id)}>Sync Now</button>
            <button class="btn-sm-danger" onclick={() => removeMapping(m.id)}>Remove</button>
          </div>
        </div>
      {:else}
        <p class="p-3 text-sm text-slate-400">No confirmed mappings yet — find and confirm matches above.</p>
      {/each}
    </div>
  </div>
</div>
