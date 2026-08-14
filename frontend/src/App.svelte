<script>
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

  function openMappings(profileId) {
    selectedProfileId = profileId
    active = 'mappings'
  }
</script>

<div class="min-h-screen">
  <header class="border-b border-slate-800 px-6 py-4">
    <h1 class="text-xl font-semibold">bookSync</h1>
    <p class="text-sm text-slate-400">Kindle &lt;-&gt; Audiobookshelf progress sync</p>
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
