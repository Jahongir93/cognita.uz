<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';
    import { getModule } from '$lib/data/activityModules';
    import { activitiesApi, ai as aiApi } from '$lib/api/client';

    const type = ($page.params as Record<string, string>).type ?? '';
    const mod = getModule(type);
    $: editId = $page.url.searchParams.get('id');

    let title = '';
    let loaded = false;

    // Har xil kontent shakllari
    let questions: { text: string; options: string[]; correct: number }[] = [];
    let statements: { text: string; answer: boolean }[] = [];
    let pairs: { left: string; right: string }[] = [];
    let groups: { name: string; items: string[] }[] = [];
    let words: { word: string; hint: string }[] = [];
    let prompts: string[] = [];

    const kind = mod?.kind ?? 'quiz';

    function seed() {
        if (kind === 'quiz') questions = [{ text: '', options: ['', ''], correct: 0 }];
        else if (kind === 'truefalse') statements = [{ text: '', answer: true }];
        else if (kind === 'pairs') pairs = [{ left: '', right: '' }, { left: '', right: '' }];
        else if (kind === 'groups') groups = [{ name: '', items: [''] }, { name: '', items: [''] }];
        else if (kind === 'words') words = [{ word: '', hint: '' }];
        else if (kind === 'prompts') prompts = [''];
    }

    function applyContent(c: any) {
        if (kind === 'quiz') questions = (c.questions ?? []).map((q: any) => ({ text: q.text ?? '', options: q.options ?? ['', ''], correct: q.correct ?? 0 }));
        else if (kind === 'truefalse') statements = (c.statements ?? []).map((s: any) => ({ text: s.text ?? '', answer: !!s.answer }));
        else if (kind === 'pairs') pairs = (c.pairs ?? []).map((p: any) => ({ left: p.left ?? '', right: p.right ?? '' }));
        else if (kind === 'groups') groups = (c.groups ?? []).map((g: any) => ({ name: g.name ?? '', items: g.items ?? [''] }));
        else if (kind === 'words') words = (c.words ?? []).map((w: any) => ({ word: w.word ?? '', hint: w.hint ?? '' }));
        else if (kind === 'prompts') prompts = c.prompts ?? [''];
    }

    onMount(async () => {
        const id = $page.url.searchParams.get('id');
        if (id) {
            try {
                const a = await activitiesApi.get(id);
                title = a.title;
                applyContent(a.content ?? {});
            } catch { seed(); }
        } else {
            seed();
        }
        loaded = true;
    });

    // ── Quiz helpers ──
    const addQuestion = () => questions = [...questions, { text: '', options: ['', ''], correct: 0 }];
    const delQuestion = (i: number) => questions = questions.filter((_, x) => x !== i);
    const addOption = (qi: number) => { questions[qi].options = [...questions[qi].options, '']; questions = questions; };
    const delOption = (qi: number, oi: number) => {
        questions[qi].options = questions[qi].options.filter((_, x) => x !== oi);
        if (questions[qi].correct >= questions[qi].options.length) questions[qi].correct = 0;
        questions = questions;
    };

    // ── TrueFalse helpers ──
    const addStmt = () => statements = [...statements, { text: '', answer: true }];
    const delStmt = (i: number) => statements = statements.filter((_, x) => x !== i);

    // ── Pairs helpers ──
    const addPair = () => pairs = [...pairs, { left: '', right: '' }];
    const delPair = (i: number) => pairs = pairs.filter((_, x) => x !== i);

    // ── Groups helpers ──
    const addGroup = () => groups = [...groups, { name: '', items: [''] }];
    const delGroup = (i: number) => groups = groups.filter((_, x) => x !== i);
    const addItem = (gi: number) => { groups[gi].items = [...groups[gi].items, '']; groups = groups; };
    const delItem = (gi: number, ii: number) => { groups[gi].items = groups[gi].items.filter((_, x) => x !== ii); groups = groups; };

    // ── Words helpers ──
    const addWord = () => words = [...words, { word: '', hint: '' }];
    const delWord = (i: number) => words = words.filter((_, x) => x !== i);

    // ── Prompts helpers ──
    const addPrompt = () => prompts = [...prompts, ''];
    const delPrompt = (i: number) => prompts = prompts.filter((_, x) => x !== i);

    let error = '';

    function buildContent(): any | null {
        if (kind === 'quiz') {
            const qs = questions.filter(q => q.text.trim() && q.options.filter(o => o.trim()).length >= 2);
            if (!qs.length) { error = "Kamida 1 ta savol va 2 ta variant kiriting"; return null; }
            return { questions: qs.map(q => ({ text: q.text.trim(), options: q.options.map(o => o.trim()).filter(Boolean), correct: q.correct })) };
        }
        if (kind === 'truefalse') {
            const st = statements.filter(s => s.text.trim());
            if (!st.length) { error = "Kamida 1 ta fikr kiriting"; return null; }
            return { statements: st.map(s => ({ text: s.text.trim(), answer: s.answer })) };
        }
        if (kind === 'pairs') {
            const ps = pairs.filter(p => p.left.trim() && p.right.trim());
            if (ps.length < 2) { error = "Kamida 2 ta to'liq juftlik kiriting"; return null; }
            return { pairs: ps.map(p => ({ left: p.left.trim(), right: p.right.trim() })) };
        }
        if (kind === 'groups') {
            const gs = groups.filter(g => g.name.trim() && g.items.filter(i => i.trim()).length);
            if (gs.length < 2) { error = "Kamida 2 ta to'liq guruh kiriting"; return null; }
            return { groups: gs.map(g => ({ name: g.name.trim(), items: g.items.map(i => i.trim()).filter(Boolean) })) };
        }
        if (kind === 'words') {
            const ws = words.filter(w => w.word.trim());
            if (!ws.length) { error = "Kamida 1 ta so'z kiriting"; return null; }
            return { words: ws.map(w => ({ word: w.word.trim(), hint: w.hint.trim() })) };
        }
        if (kind === 'prompts') {
            const ps = prompts.map(p => p.trim()).filter(Boolean);
            if (!ps.length) { error = "Kamida 1 ta topshiriq kiriting"; return null; }
            return { prompts: ps };
        }
        return null;
    }

    let saving = false;
    async function save(thenPlay = false) {
        error = '';
        if (!title.trim()) { error = "Sarlavha kiriting"; return; }
        const content = buildContent();
        if (!content) return;
        saving = true;
        try {
            let id = editId;
            if (id) {
                await activitiesApi.update(id, title.trim(), content);
            } else {
                const a = await activitiesApi.create(type, title.trim(), content);
                id = a.id;
            }
            if (thenPlay && id) window.open(`/board/${id}`, '_blank');
            goto('/dashboard/topshiriqlar');
        } catch (e: any) {
            error = e?.message ?? 'Saqlashda xato';
            saving = false;
        }
    }

    // ── AI bilan yaratish ──
    let aiTopic = '';
    let aiCount = 6;
    let aiLoading = false;
    let aiError = '';
    async function aiGenerate() {
        aiError = '';
        if (!aiTopic.trim()) { aiError = 'Mavzu kiriting'; return; }
        aiLoading = true;
        try {
            const res = await aiApi.generateActivity(kind, aiTopic.trim(), aiCount);
            applyContent(res.content ?? {});
            if (!title.trim()) title = aiTopic.trim();
        } catch (e: any) {
            aiError = e?.message ?? 'AI xatosi';
        } finally {
            aiLoading = false;
        }
    }
</script>

<svelte:head><title>{mod?.name ?? 'Topshiriq'} yaratish — Cognita.uz</title></svelte:head>

<div class="page">
    <a href="/dashboard/topshiriqlar" class="back">← Topshiriqlar</a>

    <div class="hdr">
        <span class="hdr-icon">{mod?.icon ?? '🧩'}</span>
        <div>
            <h1>{editId ? 'Tahrirlash' : 'Yangi'}: {mod?.name ?? type}</h1>
            <p class="sub">{mod?.desc ?? ''}</p>
        </div>
    </div>

    <div class="form-card">
        <label class="fld">
            <span class="lbl">Sarlavha</span>
            <input class="inp" bind:value={title} placeholder="Masalan: Hayvonlar nomi" />
        </label>
    </div>

    <!-- AI bilan to'ldirish -->
    <div class="ai-card">
        <div class="ai-head"><span class="ai-icon">✨</span> AI bilan avtomatik yaratish</div>
        <div class="ai-row">
            <input class="inp" bind:value={aiTopic} placeholder="Mavzu (masalan: O'zbekiston shaharlari)"
                   on:keydown={(e) => e.key === 'Enter' && aiGenerate()} />
            <select class="inp count" bind:value={aiCount}>
                {#each [3,5,6,8,10,12] as n}<option value={n}>{n} ta</option>{/each}
            </select>
            <button class="ai-btn" on:click={aiGenerate} disabled={aiLoading}>
                {#if aiLoading}<span class="spin"></span> Yaratilmoqda...{:else}✨ Yaratish{/if}
            </button>
        </div>
        {#if aiError}<div class="ai-err">⚠️ {aiError}</div>{/if}
        <p class="ai-note">AI yaratgan kontent quyida ko'rinadi — keyin tahrirlashingiz mumkin.</p>
    </div>

    {#if loaded}
    <!-- ── QUIZ ── -->
    {#if kind === 'quiz'}
        {#each questions as q, qi}
            <div class="item-card">
                <div class="item-top">
                    <span class="item-n">Savol {qi + 1}</span>
                    <button class="x" on:click={() => delQuestion(qi)}>✕</button>
                </div>
                <input class="inp" bind:value={q.text} placeholder="Savol matni" />
                <div class="opts">
                    {#each q.options as _, oi}
                        <div class="opt-row">
                            <input type="radio" name="correct-{qi}" checked={q.correct === oi} on:change={() => { q.correct = oi; questions = questions; }} title="To'g'ri javob" />
                            <input class="inp sm" bind:value={q.options[oi]} placeholder="Variant {oi + 1}" />
                            {#if q.options.length > 2}<button class="x sm" on:click={() => delOption(qi, oi)}>✕</button>{/if}
                        </div>
                    {/each}
                    <button class="add-sm" on:click={() => addOption(qi)}>+ variant</button>
                </div>
            </div>
        {/each}
        <button class="add" on:click={addQuestion}>+ Savol qo'shish</button>

    <!-- ── TRUE / FALSE ── -->
    {:else if kind === 'truefalse'}
        {#each statements as s, i}
            <div class="item-card">
                <div class="item-top"><span class="item-n">Fikr {i + 1}</span><button class="x" on:click={() => delStmt(i)}>✕</button></div>
                <input class="inp" bind:value={s.text} placeholder="Fikr matni" />
                <div class="tf-toggle">
                    <button class="tf" class:on={s.answer} on:click={() => { s.answer = true; statements = statements; }}>✅ To'g'ri</button>
                    <button class="tf" class:on={!s.answer} on:click={() => { s.answer = false; statements = statements; }}>❌ Noto'g'ri</button>
                </div>
            </div>
        {/each}
        <button class="add" on:click={addStmt}>+ Fikr qo'shish</button>

    <!-- ── PAIRS ── -->
    {:else if kind === 'pairs'}
        {#each pairs as p, i}
            <div class="pair-row">
                <input class="inp" bind:value={p.left} placeholder="Chap (masalan atama)" />
                <span class="pair-link">↔</span>
                <input class="inp" bind:value={p.right} placeholder="O'ng (masalan ta'rif)" />
                {#if pairs.length > 2}<button class="x" on:click={() => delPair(i)}>✕</button>{/if}
            </div>
        {/each}
        <button class="add" on:click={addPair}>+ Juftlik qo'shish</button>

    <!-- ── GROUPS ── -->
    {:else if kind === 'groups'}
        <div class="groups-grid">
            {#each groups as g, gi}
                <div class="item-card">
                    <div class="item-top">
                        <input class="inp grp-name" bind:value={g.name} placeholder="Guruh nomi" />
                        {#if groups.length > 2}<button class="x" on:click={() => delGroup(gi)}>✕</button>{/if}
                    </div>
                    {#each g.items as _, ii}
                        <div class="opt-row">
                            <input class="inp sm" bind:value={g.items[ii]} placeholder="Element" />
                            {#if g.items.length > 1}<button class="x sm" on:click={() => delItem(gi, ii)}>✕</button>{/if}
                        </div>
                    {/each}
                    <button class="add-sm" on:click={() => addItem(gi)}>+ element</button>
                </div>
            {/each}
        </div>
        <button class="add" on:click={addGroup}>+ Guruh qo'shish</button>

    <!-- ── WORDS ── -->
    {:else if kind === 'words'}
        {#each words as w, i}
            <div class="pair-row">
                <input class="inp" bind:value={w.word} placeholder="So'z (masalan: maktab)" />
                <input class="inp" bind:value={w.hint} placeholder="Izoh (ixtiyoriy)" />
                {#if words.length > 1}<button class="x" on:click={() => delWord(i)}>✕</button>{/if}
            </div>
        {/each}
        <button class="add" on:click={addWord}>+ So'z qo'shish</button>

    <!-- ── PROMPTS ── -->
    {:else if kind === 'prompts'}
        {#each prompts as _, i}
            <div class="pair-row">
                <input class="inp" bind:value={prompts[i]} placeholder="Savol yoki topshiriq" />
                {#if prompts.length > 1}<button class="x" on:click={() => delPrompt(i)}>✕</button>{/if}
            </div>
        {/each}
        <button class="add" on:click={addPrompt}>+ Topshiriq qo'shish</button>
    {/if}
    {/if}

    {#if error}<div class="err">⚠️ {error}</div>{/if}

    <div class="actions">
        <button class="btn ghost" on:click={() => save(false)} disabled={saving}>💾 Saqlash</button>
        <button class="btn pri" on:click={() => save(true)} disabled={saving}>▶ Saqlash va doskada ochish</button>
    </div>
</div>

<style>
    .page { max-width: 800px; }
    .back { color: var(--text3); text-decoration: none; font-size: 0.85rem; font-weight: 600; }
    .back:hover { color: var(--primary); }
    .hdr { display: flex; align-items: center; gap: 14px; margin: 14px 0 20px; }
    .hdr-icon { font-size: 2.4rem; }
    h1 { font-size: 1.4rem; font-weight: 800; color: var(--text); margin: 0; }
    .sub { font-size: 0.85rem; color: var(--text3); margin: 2px 0 0; }

    .form-card, .item-card {
        background: var(--white); border: 1.5px solid var(--border);
        border-radius: 14px; padding: 16px; margin-bottom: 12px;
    }
    .fld { display: flex; flex-direction: column; gap: 6px; }
    .lbl { font-size: 0.8rem; font-weight: 700; color: var(--text2); }
    .inp {
        width: 100%; padding: 10px 13px; border: 1.5px solid var(--border);
        border-radius: 10px; font-size: 0.9rem; color: var(--text); background: var(--white);
        outline: none; box-sizing: border-box; transition: border-color 0.2s;
    }
    .inp:focus { border-color: var(--primary); }
    .inp.sm { padding: 8px 11px; font-size: 0.85rem; }
    .grp-name { font-weight: 700; }

    .item-top { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: 10px; }
    .item-n { font-weight: 800; color: var(--text2); font-size: 0.85rem; }
    .opts { display: flex; flex-direction: column; gap: 7px; margin-top: 10px; }
    .opt-row { display: flex; align-items: center; gap: 8px; }
    .opt-row input[type=radio] { width: 18px; height: 18px; flex-shrink: 0; accent-color: var(--primary); }

    .pair-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
    .pair-link { color: var(--text3); font-weight: 800; }

    .groups-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 12px; }

    .tf-toggle { display: flex; gap: 8px; margin-top: 10px; }
    .tf {
        flex: 1; padding: 9px; border: 1.5px solid var(--border); background: var(--white);
        border-radius: 10px; font-weight: 700; font-size: 0.85rem; cursor: pointer; color: var(--text2);
    }
    .tf.on { border-color: var(--primary); background: var(--primary-light); color: var(--primary); }

    .x {
        width: 30px; height: 30px; border: 1.5px solid var(--border); background: var(--white);
        border-radius: 8px; cursor: pointer; color: var(--text3); flex-shrink: 0;
    }
    .x:hover { border-color: var(--danger); color: var(--danger); }
    .x.sm { width: 26px; height: 26px; font-size: 0.75rem; }

    .add {
        width: 100%; padding: 11px; border: 1.5px dashed var(--border); background: var(--bg);
        border-radius: 12px; font-weight: 700; color: var(--text2); cursor: pointer; margin-bottom: 18px;
    }
    .add:hover { border-color: var(--primary); color: var(--primary); }
    .add-sm {
        align-self: flex-start; padding: 5px 12px; border: 1px dashed var(--border);
        background: none; border-radius: 8px; font-size: 0.78rem; color: var(--text3); cursor: pointer;
    }
    .add-sm:hover { color: var(--primary); border-color: var(--primary); }

    .err { color: var(--danger); background: #fee2e2; border: 1px solid #fecaca; border-radius: 8px; padding: 10px 13px; margin-bottom: 14px; font-size: 0.85rem; }

    .ai-card { background: linear-gradient(135deg, #ede9fe, #eff6ff); border: 1.5px solid #c7d2fe; border-radius: 14px; padding: 16px; margin-bottom: 16px; }
    .ai-head { font-weight: 800; color: #4338ca; margin-bottom: 12px; display: flex; align-items: center; gap: 8px; }
    .ai-icon { font-size: 1.2rem; }
    .ai-row { display: flex; gap: 8px; flex-wrap: wrap; }
    .ai-row .inp { flex: 1; min-width: 180px; }
    .ai-row .count { flex: 0 0 96px; min-width: 0; }
    .ai-btn { padding: 10px 20px; border: none; border-radius: 10px; cursor: pointer; background: linear-gradient(135deg,#7c3aed,#6366f1); color: #fff; font-weight: 700; display: inline-flex; align-items: center; gap: 8px; white-space: nowrap; }
    .ai-btn:disabled { opacity: 0.6; cursor: default; }
    .spin { width: 14px; height: 14px; border: 2px solid rgba(255,255,255,0.4); border-top-color:#fff; border-radius:50%; animation: spin 0.6s linear infinite; }
    @keyframes spin { to { transform: rotate(360deg); } }
    .ai-err { color: var(--danger); font-size: 0.82rem; margin-top: 8px; }
    .ai-note { font-size: 0.76rem; color: #6366f1; margin: 8px 0 0; }

    .actions { display: flex; gap: 10px; flex-wrap: wrap; }
    .btn { padding: 11px 22px; border-radius: 11px; font-weight: 700; font-size: 0.9rem; cursor: pointer; border: 1.5px solid var(--border); }
    .btn.ghost { background: var(--white); color: var(--text2); }
    .btn.ghost:hover { background: var(--bg); }
    .btn.pri { background: var(--primary); color: #fff; border-color: var(--primary); }
    .btn.pri:hover { filter: brightness(1.08); }
</style>
