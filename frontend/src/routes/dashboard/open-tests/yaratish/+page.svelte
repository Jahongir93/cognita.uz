<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';
    import { openTests, ai as aiApi } from '$lib/api/client';

    const CATS = [
        { id: 'fun', label: 'Qiziqarli testlar' },
        { id: 'subjects', label: 'Fan testlari' },
        { id: 'iq', label: 'IQ testlar' },
        { id: 'attestation', label: 'Attestatsiya' },
        { id: 'psychology', label: 'Psixologik testlar' },
    ];

    type Q = { text: string; options: { text: string; is_correct: boolean }[]; explanation: string };

    let editId = '';
    let category = 'fun';
    let title = '';
    let description = '';
    let scored = true;
    let isPublished = true;
    let questions: Q[] = [];
    let loaded = false;
    let error = '';
    let saving = false;

    // psixologik → reytingsiz
    $: if (category === 'psychology') scored = false;

    function newQ(): Q { return { text: '', options: [{ text: '', is_correct: true }, { text: '', is_correct: false }], explanation: '' }; }

    onMount(async () => {
        editId = $page.url.searchParams.get('id') ?? '';
        const qcat = $page.url.searchParams.get('category');
        if (editId) {
            try {
                const t = await openTests.getForEdit(editId);
                category = t.category; title = t.title; description = t.description ?? '';
                scored = t.scored; isPublished = t.is_published;
                questions = (t.questions ?? []).map((q: any) => ({
                    text: q.text ?? '',
                    options: (q.options ?? []).map((o: any) => ({ text: o.text ?? '', is_correct: !!o.is_correct })),
                    explanation: q.explanation ?? '',
                }));
                if (!questions.length) questions = [newQ()];
            } catch { error = 'Test yuklanmadi'; }
        } else {
            if (qcat) category = qcat;
            questions = [newQ()];
        }
        loaded = true;
    });

    const addQ = () => questions = [...questions, newQ()];
    const delQ = (i: number) => questions = questions.filter((_, x) => x !== i);
    const addOpt = (qi: number) => { questions[qi].options = [...questions[qi].options, { text: '', is_correct: false }]; questions = questions; };
    const delOpt = (qi: number, oi: number) => { questions[qi].options = questions[qi].options.filter((_, x) => x !== oi); questions = questions; };
    function setCorrect(qi: number, oi: number) {
        questions[qi].options = questions[qi].options.map((o, x) => ({ ...o, is_correct: x === oi }));
        questions = questions;
    }

    // ── AI ──
    let aiTopic = ''; let aiCount = 6; let aiLoading = false; let aiError = '';
    async function aiGenerate() {
        aiError = '';
        if (!aiTopic.trim()) { aiError = 'Mavzu kiriting'; return; }
        aiLoading = true;
        try {
            const res = await aiApi.generateActivity('quiz', aiTopic.trim(), aiCount);
            const qs = (res.content?.questions ?? []) as any[];
            questions = qs.map(q => ({
                text: q.text ?? '',
                options: (q.options ?? []).map((opt: string, i: number) => ({ text: opt, is_correct: i === (q.correct ?? 0) })),
                explanation: '',
            }));
            if (!title.trim()) title = aiTopic.trim();
        } catch (e: any) { aiError = e?.message ?? 'AI xatosi'; }
        finally { aiLoading = false; }
    }

    // ── Import JSON ──
    let showImport = false; let importText = ''; let importError = '';
    function doImport() {
        importError = '';
        try {
            const data = JSON.parse(importText);
            const qs = Array.isArray(data) ? data : (data.questions ?? []);
            questions = qs.map((q: any) => ({
                text: q.text ?? q.question_text ?? '',
                options: (q.options ?? []).map((o: any, i: number) =>
                    typeof o === 'string'
                        ? { text: o, is_correct: i === (q.correct ?? -1) }
                        : { text: o.text ?? o.option_text ?? '', is_correct: !!(o.is_correct) }),
                explanation: q.explanation ?? '',
            })).filter((q: Q) => q.text);
            if (!questions.length) { importError = 'Savol topilmadi'; return; }
            showImport = false; importText = '';
        } catch { importError = "Noto'g'ri JSON"; }
    }

    async function save() {
        error = '';
        if (!title.trim()) { error = 'Sarlavha kiriting'; return; }
        const valid = questions.filter(q => q.text.trim() && q.options.filter(o => o.text.trim()).length >= 2);
        if (!valid.length) { error = "Kamida 1 ta to'liq savol kerak"; return; }
        const payload = {
            category, title: title.trim(), description: description.trim(),
            scored, is_published: isPublished,
            questions: valid.map(q => ({
                text: q.text.trim(),
                options: q.options.filter(o => o.text.trim()).map(o => ({ text: o.text.trim(), is_correct: o.is_correct })),
                explanation: q.explanation.trim(),
            })),
        };
        saving = true;
        try {
            if (editId) await openTests.update(editId, payload);
            else await openTests.create(payload);
            goto('/dashboard/open-tests');
        } catch (e: any) { error = e?.message ?? 'Saqlashda xato'; saving = false; }
    }
</script>

<svelte:head><title>{editId ? 'Test tahrirlash' : 'Yangi test'} — Cognita.uz</title></svelte:head>

<div class="page">
    <a href="/dashboard/open-tests" class="back">← Katalog</a>
    <h1>{editId ? 'Testni tahrirlash' : 'Yangi test'}</h1>

    {#if loaded}
    <div class="card">
        <div class="row2">
            <label class="fld">
                <span class="lbl">Kategoriya</span>
                <select class="inp" bind:value={category}>
                    {#each CATS as c}<option value={c.id}>{c.label}</option>{/each}
                </select>
            </label>
            <label class="fld">
                <span class="lbl">Reyting (leaderboard)</span>
                <select class="inp" bind:value={scored} disabled={category === 'psychology'}>
                    <option value={true}>Ha — ballanadi</option>
                    <option value={false}>Yo'q</option>
                </select>
            </label>
        </div>
        <label class="fld"><span class="lbl">Sarlavha</span><input class="inp" bind:value={title} placeholder="Masalan: Mantiqiy IQ test #1" /></label>
        <label class="fld"><span class="lbl">Tavsif (ixtiyoriy)</span><input class="inp" bind:value={description} placeholder="Qisqa tavsif" /></label>
        <label class="chk"><input type="checkbox" bind:checked={isPublished} /> Nashr etilsin (foydalanuvchilarga ko'rinadi)</label>
    </div>

    <!-- AI -->
    <div class="ai-card">
        <div class="ai-head">✨ AI bilan savollar yaratish</div>
        <div class="ai-row">
            <input class="inp" bind:value={aiTopic} placeholder="Mavzu (masalan: Mantiqiy ketma-ketliklar)" on:keydown={(e)=> e.key==='Enter' && aiGenerate()} />
            <select class="inp count" bind:value={aiCount}>{#each [5,6,8,10,12,15] as n}<option value={n}>{n} ta</option>{/each}</select>
            <button class="ai-btn" on:click={aiGenerate} disabled={aiLoading}>{aiLoading ? 'Yaratilmoqda...' : '✨ Yaratish'}</button>
            <button class="ai-btn ghost" on:click={() => showImport = !showImport}>📥 Import (JSON)</button>
        </div>
        {#if aiError}<div class="err sm">⚠️ {aiError}</div>{/if}
        {#if showImport}
            <textarea class="imp" bind:value={importText} placeholder='JSON joylang: [&#123;"text":"Savol","options":["A","B","C"],"correct":0&#125;]'></textarea>
            {#if importError}<div class="err sm">⚠️ {importError}</div>{/if}
            <button class="ai-btn" on:click={doImport}>Import qilish</button>
        {/if}
    </div>

    <!-- Savollar -->
    {#each questions as q, qi}
        <div class="card q-card">
            <div class="q-top"><span class="q-n">Savol {qi + 1}</span><button class="x" on:click={() => delQ(qi)}>✕</button></div>
            <input class="inp" bind:value={q.text} placeholder="Savol matni" />
            <div class="opts">
                {#each q.options as o, oi}
                    <div class="opt-row">
                        {#if scored}
                            <input type="radio" name="c-{qi}" checked={o.is_correct} on:change={() => setCorrect(qi, oi)} title="To'g'ri javob" />
                        {/if}
                        <input class="inp sm" bind:value={q.options[oi].text} placeholder="Variant {oi + 1}" />
                        {#if q.options.length > 2}<button class="x sm" on:click={() => delOpt(qi, oi)}>✕</button>{/if}
                    </div>
                {/each}
                <button class="add-sm" on:click={() => addOpt(qi)}>+ variant</button>
            </div>
        </div>
    {/each}
    <button class="add" on:click={addQ}>+ Savol qo'shish</button>

    {#if error}<div class="err">⚠️ {error}</div>{/if}
    <div class="actions">
        <button class="btn pri" on:click={save} disabled={saving}>💾 {saving ? 'Saqlanmoqda...' : 'Saqlash'}</button>
        <a class="btn ghost" href="/dashboard/open-tests">Bekor qilish</a>
    </div>
    {/if}
</div>

<style>
    .page { max-width: 800px; }
    .back { color: var(--text3); text-decoration: none; font-size: 0.85rem; font-weight: 600; }
    .back:hover { color: var(--primary); }
    h1 { font-size: 1.4rem; font-weight: 800; color: var(--text); margin: 10px 0 18px; }
    .card { background: var(--white); border: 1.5px solid var(--border); border-radius: 14px; padding: 16px; margin-bottom: 12px; display: flex; flex-direction: column; gap: 12px; }
    .row2 { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
    .fld { display: flex; flex-direction: column; gap: 6px; }
    .lbl { font-size: 0.8rem; font-weight: 700; color: var(--text2); }
    .inp { width: 100%; padding: 10px 13px; border: 1.5px solid var(--border); border-radius: 10px; font-size: 0.9rem; color: var(--text); background: var(--white); outline: none; box-sizing: border-box; }
    .inp:focus { border-color: var(--primary); }
    .inp.sm { padding: 8px 11px; font-size: 0.85rem; }
    .chk { display: flex; align-items: center; gap: 8px; font-size: 0.85rem; color: var(--text2); }
    .ai-card { background: linear-gradient(135deg,#ede9fe,#eff6ff); border: 1.5px solid #c7d2fe; border-radius: 14px; padding: 16px; margin-bottom: 14px; }
    .ai-head { font-weight: 800; color: #4338ca; margin-bottom: 10px; }
    .ai-row { display: flex; gap: 8px; flex-wrap: wrap; }
    .ai-row .inp { flex: 1; min-width: 160px; }
    .ai-row .count { flex: 0 0 90px; min-width: 0; }
    .ai-btn { padding: 10px 16px; border: none; border-radius: 10px; background: linear-gradient(135deg,#7c3aed,#6366f1); color: #fff; font-weight: 700; cursor: pointer; white-space: nowrap; }
    .ai-btn.ghost { background: var(--white); color: var(--primary); border: 1.5px solid #c7d2fe; }
    .ai-btn:disabled { opacity: 0.6; }
    .imp { width: 100%; min-height: 90px; margin-top: 10px; padding: 10px; border: 1.5px solid #c7d2fe; border-radius: 10px; font-family: monospace; font-size: 0.8rem; box-sizing: border-box; }
    .q-card { gap: 10px; }
    .q-top { display: flex; align-items: center; justify-content: space-between; }
    .q-n { font-weight: 800; color: var(--text2); font-size: 0.85rem; }
    .opts { display: flex; flex-direction: column; gap: 7px; }
    .opt-row { display: flex; align-items: center; gap: 8px; }
    .opt-row input[type=radio] { width: 18px; height: 18px; accent-color: var(--primary); flex-shrink: 0; }
    .x { width: 30px; height: 30px; border: 1.5px solid var(--border); background: var(--white); border-radius: 8px; cursor: pointer; color: var(--text3); flex-shrink: 0; }
    .x.sm { width: 26px; height: 26px; font-size: 0.75rem; }
    .x:hover { border-color: var(--danger); color: var(--danger); }
    .add-sm { align-self: flex-start; padding: 5px 12px; border: 1px dashed var(--border); background: none; border-radius: 8px; font-size: 0.78rem; color: var(--text3); cursor: pointer; }
    .add { width: 100%; padding: 11px; border: 1.5px dashed var(--border); background: var(--bg); border-radius: 12px; font-weight: 700; color: var(--text2); cursor: pointer; margin-bottom: 16px; }
    .add:hover { border-color: var(--primary); color: var(--primary); }
    .err { color: var(--danger); background: #fee2e2; border: 1px solid #fecaca; border-radius: 8px; padding: 10px 13px; margin-bottom: 12px; font-size: 0.85rem; }
    .err.sm { padding: 6px 10px; margin: 8px 0 0; }
    .actions { display: flex; gap: 10px; }
    .btn { padding: 11px 22px; border-radius: 11px; font-weight: 700; font-size: 0.9rem; cursor: pointer; border: 1.5px solid var(--border); text-decoration: none; display: inline-flex; align-items: center; }
    .btn.pri { background: var(--primary); color: #fff; border-color: var(--primary); }
    .btn.ghost { background: var(--white); color: var(--text2); }
</style>
