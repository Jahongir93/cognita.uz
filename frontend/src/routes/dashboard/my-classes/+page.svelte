<script lang="ts">
    import { onMount } from 'svelte';
    import { classes, type MyClass } from '$lib/api/client';

    let list: MyClass[] = [];
    let loading = true;
    let code = '';
    let joining = false;
    let msg = '';
    let msgType: 'ok' | 'err' = 'ok';

    async function reload() {
        loading = true;
        try { list = await classes.my(); } catch { list = []; }
        loading = false;
    }
    onMount(reload);

    async function join() {
        msg = '';
        if (!code.trim()) { msg = 'Sinf kodini kiriting'; msgType = 'err'; return; }
        joining = true;
        try {
            const res = await classes.join(code.trim());
            msg = `"${res.name}" sinfiga qo'shildingiz ✓`; msgType = 'ok';
            code = '';
            await reload();
        } catch (e: any) {
            msg = e?.message ?? 'Qo\'shilishda xato'; msgType = 'err';
        } finally {
            joining = false;
        }
    }

    async function leave(cl: MyClass) {
        if (!confirm(`"${cl.name}" sinfidan chiqasizmi?`)) return;
        try { await classes.leave(cl.id); } catch {}
        reload();
    }

    const colors = ['#6366f1','#ec4899','#f59e0b','#22c55e','#06b6d4','#8b5cf6'];
</script>

<svelte:head><title>Sinflarim — Cognita.uz</title></svelte:head>

<div class="page">
    <div class="page-header">
        <div>
            <h1>👨‍🏫 Sinflarim</h1>
            <p class="sub">O'qituvchingiz bergan kod orqali sinfga qo'shiling</p>
        </div>
    </div>

    <!-- Join card -->
    <div class="join-card">
        <div class="jc-icon">🎓</div>
        <div class="jc-body">
            <div class="jc-title">Sinfga qo'shilish</div>
            <div class="jc-row">
                <input class="code-input" bind:value={code} placeholder="SINF KODI (masalan ABC123)"
                       maxlength="10" on:keydown={(e) => e.key === 'Enter' && join()}
                       style="text-transform:uppercase" />
                <button class="join-btn" on:click={join} disabled={joining}>
                    {joining ? '...' : 'Qo\'shilish'}
                </button>
            </div>
            {#if msg}<div class="msg {msgType}">{msg}</div>{/if}
        </div>
    </div>

    <!-- My classes -->
    {#if loading}
        <p class="muted">Yuklanmoqda...</p>
    {:else if list.length === 0}
        <div class="empty">
            <div class="empty-icon">📚</div>
            <p>Siz hali birorta sinfga qo'shilmagansiz.</p>
            <p class="empty-sub">Yuqoridagi maydonga o'qituvchingiz bergan kodni kiriting.</p>
        </div>
    {:else}
        <div class="grid">
            {#each list as cl, i (cl.id)}
                <div class="cl-card" style="--c:{colors[i % colors.length]}">
                    <div class="cl-bar"></div>
                    <div class="cl-top">
                        <span class="cl-name">{cl.name}</span>
                        <button class="leave" on:click={() => leave(cl)} title="Sinfdan chiqish">✕</button>
                    </div>
                    <div class="cl-meta">
                        {#if cl.grade}<span class="tag">{cl.grade}-sinf</span>{/if}
                        {#if cl.subject}<span class="tag">{cl.subject}</span>{/if}
                    </div>
                    <div class="cl-foot">
                        <span>👩‍🏫 {cl.teacher_name || 'O\'qituvchi'}</span>
                        <span>👥 {cl.student_count}</span>
                    </div>
                    <div class="cl-code">Kod: <strong>{cl.class_code}</strong></div>
                </div>
            {/each}
        </div>
    {/if}
</div>

<style>
    .page { max-width: 900px; }
    .page-header { margin-bottom: 20px; }
    h1 { font-size: 1.6rem; font-weight: 800; color: var(--text); margin: 0; }
    .sub { font-size: 0.88rem; color: var(--text3); margin-top: 4px; }
    .muted { color: var(--text3); }

    .join-card {
        display: flex; gap: 16px; align-items: center;
        background: linear-gradient(135deg, var(--primary-light), #eff6ff);
        border: 1.5px solid #c7d2fe; border-radius: 16px; padding: 18px 20px; margin-bottom: 24px;
    }
    .jc-icon { font-size: 2.4rem; flex-shrink: 0; }
    .jc-body { flex: 1; }
    .jc-title { font-weight: 800; color: #4338ca; margin-bottom: 8px; }
    .jc-row { display: flex; gap: 8px; flex-wrap: wrap; }
    .code-input {
        flex: 1; min-width: 180px; padding: 11px 14px; border: 1.5px solid var(--border);
        border-radius: 10px; font-size: 1rem; font-weight: 700; letter-spacing: 0.1em;
        outline: none; background: var(--white); color: var(--text);
    }
    .code-input:focus { border-color: var(--primary); }
    .join-btn { padding: 11px 22px; border: none; border-radius: 10px; background: var(--primary); color: #fff; font-weight: 700; cursor: pointer; white-space: nowrap; }
    .join-btn:hover:not(:disabled) { filter: brightness(1.08); }
    .join-btn:disabled { opacity: 0.6; }
    .msg { margin-top: 10px; font-size: 0.85rem; font-weight: 600; }
    .msg.ok { color: #16a34a; }
    .msg.err { color: var(--danger); }

    .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 14px; }
    .cl-card {
        position: relative; overflow: hidden;
        background: var(--white); border: 1.5px solid var(--border); border-radius: 16px;
        padding: 16px 16px 14px; box-shadow: var(--shadow-sm); display: flex; flex-direction: column; gap: 8px;
    }
    .cl-bar { position: absolute; top: 0; left: 0; right: 0; height: 5px; background: var(--c); }
    .cl-top { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-top: 4px; }
    .cl-name { font-weight: 800; color: var(--text); font-size: 1.05rem; }
    .leave { width: 26px; height: 26px; border: 1.5px solid var(--border); background: var(--white); border-radius: 7px; cursor: pointer; color: var(--text3); flex-shrink: 0; }
    .leave:hover { border-color: var(--danger); color: var(--danger); }
    .cl-meta { display: flex; gap: 6px; flex-wrap: wrap; }
    .tag { font-size: 0.72rem; font-weight: 700; background: var(--bg); color: var(--text2); padding: 3px 10px; border-radius: 99px; }
    .cl-foot { display: flex; justify-content: space-between; font-size: 0.78rem; color: var(--text3); }
    .cl-code { font-size: 0.76rem; color: var(--text3); border-top: 1px solid var(--border); padding-top: 8px; }
    .cl-code strong { color: var(--c); letter-spacing: 0.08em; }

    .empty { text-align: center; padding: 50px 20px; color: var(--text3); }
    .empty-icon { font-size: 3rem; }
    .empty-sub { font-size: 0.85rem; }
</style>
