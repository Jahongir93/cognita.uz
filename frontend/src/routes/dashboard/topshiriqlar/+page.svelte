<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { ACTIVITY_MODULES, getModule } from '$lib/data/activityModules';
    import { activitiesApi, type BoardActivity } from '$lib/api/client';

    let myActivities: BoardActivity[] = [];
    let loading = true;
    let activeCat: string = 'Hammasi';

    const categories = ['Hammasi', 'Savol-javob', 'Moslashtirish', "So'z", 'Saralash', 'Tasodif', 'Harakatli'];

    $: filteredModules = activeCat === 'Hammasi'
        ? ACTIVITY_MODULES
        : ACTIVITY_MODULES.filter(m => m.category === activeCat);

    async function reload() {
        try { myActivities = await activitiesApi.list(); } catch { myActivities = []; }
        loading = false;
    }
    onMount(reload);

    function createNew(type: string) {
        goto(`/dashboard/topshiriqlar/yaratish/${type}`);
    }
    function play(a: BoardActivity) {
        window.open(`/board/${a.id}`, '_blank');
    }
    function edit(a: BoardActivity) {
        goto(`/dashboard/topshiriqlar/yaratish/${a.type}?id=${a.id}`);
    }
    async function remove(a: BoardActivity) {
        if (confirm(`"${a.title}" topshirig'ini o'chirasizmi?`)) {
            try { await activitiesApi.remove(a.id); } catch {}
            reload();
        }
    }
    function moduleName(type: string) { return getModule(type)?.name ?? type; }
    function moduleIcon(type: string) { return getModule(type)?.icon ?? '🧩'; }
    function moduleImage(type: string) { return getModule(type)?.image; }
</script>

<svelte:head><title>Topshiriqlar — Cognita.uz</title></svelte:head>

<div class="page">
    <div class="page-header">
        <div>
            <h1>🧩 Topshiriqlar</h1>
            <p class="sub">Elektron doskada o'ynaladigan interaktiv o'yinlar. O'qituvchi yaratadi va ishga tushiradi.</p>
        </div>
    </div>

    <!-- Mening topshiriqlarim -->
    {#if myActivities.length}
        <section class="block">
            <h2 class="block-title">Mening topshiriqlarim ({myActivities.length})</h2>
            <div class="my-grid">
                {#each myActivities as a (a.id)}
                    <div class="my-card">
                        {#if moduleImage(a.type)}
                            <img src="/img/board/{moduleImage(a.type)}" alt="" class="my-icon-img" />
                        {:else}
                            <div class="my-icon">{moduleIcon(a.type)}</div>
                        {/if}
                        <div class="my-info">
                            <div class="my-name">{a.title}</div>
                            <div class="my-type">{moduleName(a.type)}</div>
                        </div>
                        <div class="my-actions">
                            <button class="mc-btn play" on:click={() => play(a)} title="Doskada o'ynash">▶</button>
                            <button class="mc-btn" on:click={() => edit(a)} title="Tahrirlash">✏️</button>
                            <button class="mc-btn del" on:click={() => remove(a)} title="O'chirish">🗑</button>
                        </div>
                    </div>
                {/each}
            </div>
        </section>
    {/if}

    <!-- Yangi yaratish -->
    <section class="block">
        <h2 class="block-title">Yangi topshiriq yaratish</h2>

        <div class="cats">
            {#each categories as c}
                <button class="cat-btn" class:active={activeCat === c} on:click={() => activeCat = c}>{c}</button>
            {/each}
        </div>

        <div class="mod-grid">
            {#each filteredModules as m (m.id)}
                <button
                    class="mod-card"
                    class:soon={!m.implemented}
                    disabled={!m.implemented}
                    on:click={() => m.implemented && createNew(m.id)}
                >
                    {#if m.image}
                        <img src="/img/board/{m.image}" alt="" class="mod-icon-img" />
                    {:else}
                        <span class="mod-icon">{m.icon}</span>
                    {/if}
                    <span class="mod-name">{m.name}</span>
                    <span class="mod-desc">{m.desc}</span>
                    {#if !m.implemented}
                        <span class="soon-badge">Tez kunda</span>
                    {:else}
                        <span class="ready-badge">Tayyor</span>
                    {/if}
                </button>
            {/each}
        </div>
    </section>
</div>

<style>
    .page { max-width: 1100px; }
    .page-header { margin-bottom: 24px; }
    h1 { font-size: 1.75rem; font-weight: 800; color: var(--text); margin: 0; }
    .sub { font-size: 0.9rem; color: var(--text3); margin-top: 4px; max-width: 640px; }

    .block { margin-bottom: 32px; }
    .block-title { font-size: 1.05rem; font-weight: 800; color: var(--text); margin: 0 0 14px; }

    /* Mening topshiriqlarim */
    .my-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 12px; }
    .my-card {
        display: flex; align-items: center; gap: 12px;
        background: var(--white); border: 1.5px solid var(--border);
        border-radius: 14px; padding: 14px 16px; box-shadow: var(--shadow-sm);
    }
    .my-icon { font-size: 1.8rem; }
    .my-info { flex: 1; min-width: 0; }
    .my-name { font-weight: 700; color: var(--text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .my-type { font-size: 0.78rem; color: var(--text3); }
    .my-actions { display: flex; gap: 4px; }
    .mc-btn {
        width: 34px; height: 34px; border: 1.5px solid var(--border);
        background: var(--white); border-radius: 9px; cursor: pointer;
        font-size: 0.95rem; transition: var(--transition);
    }
    .mc-btn:hover { background: var(--bg); }
    .mc-btn.play { background: var(--primary); border-color: var(--primary); color: #fff; }
    .mc-btn.play:hover { filter: brightness(1.1); }
    .mc-btn.del:hover { border-color: var(--danger); }

    /* Kategoriyalar */
    .cats { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 16px; }
    .cat-btn {
        padding: 7px 14px; border: 1.5px solid var(--border); background: var(--white);
        border-radius: 99px; font-size: 0.82rem; font-weight: 600; color: var(--text2);
        cursor: pointer; transition: var(--transition);
    }
    .cat-btn:hover { background: var(--bg); }
    .cat-btn.active { background: var(--primary); color: #fff; border-color: var(--primary); }

    /* Modullar gridi */
    .mod-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(190px, 1fr)); gap: 12px; }
    .mod-card {
        position: relative;
        display: flex; flex-direction: column; align-items: flex-start; gap: 6px;
        background: var(--white); border: 1.5px solid var(--border);
        border-radius: 14px; padding: 16px; text-align: left; cursor: pointer;
        transition: transform 0.15s, box-shadow 0.15s, border-color 0.15s;
        min-height: 120px;
    }
    .mod-card:not(.soon):hover {
        transform: translateY(-3px);
        border-color: var(--primary);
        box-shadow: 0 8px 22px rgba(99,102,241,0.18);
    }
    .mod-card.soon { opacity: 0.6; cursor: not-allowed; }
    .mod-icon { font-size: 2rem; }
    .mod-icon-img { width: 52px; height: 52px; object-fit: contain; display: block; }
    .my-icon-img { width: 40px; height: 40px; object-fit: contain; flex-shrink: 0; }
    .mod-name { font-weight: 800; color: var(--text); font-size: 0.95rem; }
    .mod-desc { font-size: 0.76rem; color: var(--text3); line-height: 1.3; }
    .soon-badge, .ready-badge {
        position: absolute; top: 10px; right: 10px;
        font-size: 0.64rem; font-weight: 700; padding: 2px 8px; border-radius: 99px;
    }
    .soon-badge { background: #f1f5f9; color: #64748b; }
    .ready-badge { background: #dcfce7; color: #16a34a; }
</style>
