<script>
  import { api } from '../lib/api.js'

  let events = $state([])
  let error = $state('')

  async function load() {
    try {
      events = (await api.activity(100)) ?? []
    } catch (e) {
      error = e.message
    }
  }

  load()
</script>

<div class="max-w-3xl">
  <div class="flex items-center justify-between mb-3">
    <h2 class="font-medium">Recent Sync Activity</h2>
    <button class="btn-sm" onclick={load}>Refresh</button>
  </div>
  {#if error}<p class="text-red-400 text-sm">{error}</p>{/if}
  <div class="rounded-lg border border-slate-800 divide-y divide-slate-800">
    {#each events as e (e.id)}
      <div class="p-3 text-sm flex items-center justify-between">
        <div>
          <span class="font-mono text-xs px-1.5 py-0.5 rounded bg-slate-800">{e.direction}</span>
          <span class="ml-2 text-slate-300">mapping #{e.mappingId}</span>
          {#if e.message}<div class="text-xs text-slate-500">{e.message}</div>{/if}
        </div>
        <div class="text-xs text-slate-500">{new Date(e.created).toLocaleString()}</div>
      </div>
    {:else}
      <p class="p-3 text-sm text-slate-400">No sync activity yet.</p>
    {/each}
  </div>
</div>
