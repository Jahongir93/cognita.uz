<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { openTests, type OpenTest } from '$lib/api/client';

    const CATS = [
        { id: 'fun',         label: 'Qiziqarli testlar', icon: '🎯' },
        { id: 'subjects',    label: 'Fan testlari',       icon: '📚' },
        { id: 'iq',          label: 'IQ testlar',         icon: '🧠' },
        { id: 'attestation', label: 'Attestatsiya',        icon: '📋' },
        { id: 'psychology',  label: 'Psixologik testlar', icon: '🧘' },
    ];

    let activeCat = 'fun';
    let list: OpenTest[] = [];
    let loading = true;

    async function reload() {
        loading = true;
        try { list = await openTests.listAdmin(activeCat); } catch { list = []; }
        loading = false;
    }
    onMount(reload);
    $: activeCat, reload();

    function create() { goto(`/dashboard/open-tests/yaratish?category=${activeCat}`); }
    function edit(t: OpenTest) { goto(`/dashboard/open-tests/yaratish?id=${t.id}`); }
    async function remove(t: OpenTest) {
        if (confirm(`"${t.title}" testini o'chirasizmi?`)) { try { await openTests.remove(t.id); } catch {} reload(); }
    }
    async function togglePublish(t: OpenTest) {
        try {
            const full = await openTests.getForEdit(t.id);
            await openTests.update(t.id, { ...full, is_published: !t.is_published });
        } catch {}
        reload();
    }
    $: catLabel = CATS.find(c => c.id === activeCat)?.label ?? '';
</script>

<svelte:head><title>Ochiq testlar — Cognita.uz Admin</title></svelte:head>

<div class="page">
    <div class="page-header">
        <div>
            <h1>📂 Ochiq testlar</h1>
            <p class="sub">Saytdagi test kategoriyalarini boshqaring. Faqat admin qo'sha oladi. Foydalanuvchilar istalgan vaqt yechadi.</p>
        </div>
        <button class="btn pri" on:click={create}>＋ Yangi test</button>
    </div>

    <div class="cats">
        {#each CATS as c}
            <button class="cat-btn" class:active={activeCat === c.id} on:click={() => activeCat = c.id}>
                <span>{c.icon}</span> {c.label}
            </button>
        {/each}
    </div>

    {#if loading}
        <p class="muted">Yuklanmoqda...</p>
    {:else if list.length === 0}
        <div class="empty">
            <div class="empty-icon">📭</div>
            <p>{catLabel} bo'yicha hali test yo'q.</p>
            <button class="btn pri" on:click={create}>＋ Birinchi testni yarating</button>
        </div>
    {:else}
        <div class="grid">
            {#each list as t (t.id)}
                <div class="card">
                    <div class="card-top">
                        <span class="title">{t.title}</span>
                        {#if t.is_published}<span class="badge pub">Nashr</span>{:else}<span class="badge draft">Qoralama</span>{/if}
                    </div>
                    <p class="desc">{t.description || '—'}</p>
                    <div class="meta">
                        <span>❓ {t.questions} savol</span>
                        <span>{t.scored ? '🏆 Reytingli' : '🧘 Reytingsiz'}</span>
                        <span>▶ {t.play_count}</span>
                    </div>
                    <div class="actions">
                        <button class="a-btn" on:click={() => edit(t)}>✏️ Tahrir</button>
                        <button class="a-btn" on:click={() => togglePublish(t)}>{t.is_published ? '🙈 Yashirish' : '📢 Nashr'}</button>
                        <a class="a-btn" href={`/test/${t.id}`} target="_blank">▶ Ko'rish</a>
                        <button class="a-btn del" on:click={() => remove(t)}>🗑</button>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>

<style>
    .page { max-width: 1000px; }
    .page-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 20px; }
    h1 { font-size: 1.6rem; font-weight: 800; color: var(--text); margin: 0; }
    .sub { font-size: 0.88rem; color: var(--text3); margin-top: 4px; max-width: 560px; }
    .muted { color: var(--text3); }
    .cats { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 20px; }
    .cat-btn { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border: 1.5px solid var(--border); background: var(--white); border-radius: 99px; font-weight: 600; font-size: 0.85rem; color: var(--text2); cursor: pointer; }
    .cat-btn.active { background: var(--primary); color: #fff; border-color: var(--primary); }
    .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(290px, 1fr)); gap: 14px; }
    .card { background: var(--white); border: 1.5px solid var(--border); border-radius: 14px; padding: 16px; display: flex; flex-direction: column; gap: 8px; box-shadow: var(--shadow-sm); }
    .card-top { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
    .title { font-weight: 800; color: var(--text); }
    .badge { font-size: 0.62rem; font-weight: 700; padding: 2px 8px; border-radius: 99px; }
    .badge.pub { background: #dcfce7; color: #16a34a; }
    .badge.draft { background: #f1f5f9; color: #64748b; }
    .desc { font-size: 0.82rem; color: var(--text3); margin: 0; }
    .meta { display: flex; gap: 12px; font-size: 0.76rem; color: var(--text3); flex-wrap: wrap; }
    .actions { display: flex; gap: 6px; flex-wrap: wrap; margin-top: 4px; }
    .a-btn { padding: 6px 10px; border: 1.5px solid var(--border); background: var(--white); border-radius: 8px; font-size: 0.76rem; font-weight: 600; cursor: pointer; color: var(--text2); text-decoration: none; }
    .a-btn:hover { background: var(--bg); }
    .a-btn.del:hover { border-color: var(--danger); color: var(--danger); }
    .btn.pri { padding: 10px 20px; border: none; border-radius: 11px; background: var(--primary); color: #fff; font-weight: 700; cursor: pointer; }
    .btn.pri:hover { filter: brightness(1.08); }
    .empty { text-align: center; padding: 50px 20px; color: var(--text3); }
    .empty-icon { font-size: 3rem; }
    .empty .btn { margin-top: 12px; }
</style>
