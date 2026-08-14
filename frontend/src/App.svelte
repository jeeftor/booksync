<script>
  import { api } from './lib/api.js'
  import KindleAccounts from './components/KindleAccounts.svelte'
  import ABSUsers from './components/ABSUsers.svelte'
  import Profiles from './components/Profiles.svelte'
  import Mappings from './components/Mappings.svelte'
  import Activity from './components/Activity.svelte'

  const tabs = [
    { id: 'profiles', label: 'Profiles' },
    { id: 'mappings', label: 'Mappings' },
    { id: 'kindle', label: 'Kindle Accounts' },
    { id: 'abs', label: 'Audiobookshelf Users' },
    { id: 'activity', label: 'Activity' },
  ]

  let active = $state('profiles')
  let selectedProfileId = $state(null)
  let health = $state(null)

  function openMappings(profileId) {
    selectedProfileId = profileId
    active = 'mappings'
  }

  // First-run UX: if nothing is configured yet, land on whichever setup step
  // comes first instead of the (empty) Profiles tab.
  async function pickInitialTab() {
    try {
      const [kindleAccounts, absUsers] = await Promise.all([api.kindleAccounts.list(), api.absUsers.list()])
      if (!kindleAccounts.length) {
        active = 'kindle'
      } else if (!absUsers.length) {
        active = 'abs'
      }
    } catch {
      /* ignore - default tab stands */
    }
  }

  async function loadHealth() {
    try {
      health = await api.health()
    } catch {
      /* ignore - version badge just won't show */
    }
  }

  pickInitialTab()
  loadHealth()
</script>

<div class="min-h-screen">
  <header class="border-b border-slate-800 px-6 py-4 flex items-center justify-between">
    <div>
      <h1 class="text-xl font-semibold">bookSync</h1>
      <p class="text-sm text-slate-400">Kindle &lt;-&gt; Audiobookshelf progress sync</p>
    </div>
    {#if health?.version}
      <span
        class="text-xs font-mono text-slate-500 border border-slate-800 rounded px-2 py-1"
        title={`commit ${health.commit}${health.date && health.date !== 'unknown' ? ` · built ${new Date(health.date).toLocaleString()}` : ''}`}
      >
        {health.version}
      </span>
    {/if}
  </header>

  <nav class="flex gap-1 border-b border-slate-800 px-6">
    {#each tabs as tab}
      <button
        class="px-4 py-2 text-sm rounded-t-md {active === tab.id
          ? 'bg-slate-800 text-white'
          : 'text-slate-400 hover:text-slate-200'}"
        onclick={() => (active = tab.id)}
      >
        {tab.label}
      </button>
    {/each}
  </nav>

  <main class="p-6">
    {#if active === 'profiles'}
      <Profiles onOpenMappings={openMappings} />
    {:else if active === 'mappings'}
      <Mappings bind:profileId={selectedProfileId} />
    {:else if active === 'kindle'}
      <KindleAccounts />
    {:else if active === 'abs'}
      <ABSUsers />
    {:else if active === 'activity'}
      <Activity />
    {/if}
  </main>
</div>
