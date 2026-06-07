<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import type { EditQuestion, EditOption } from './QuizEditor.svelte';

    const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

    export let show = false;

    const dispatch = createEventDispatcher<{
        close: void;
        import: { questions: EditQuestion[] };
    }>();

    // ── State ─────────────────────────────────────────────────────────────────
    type Step = 'upload' | 'analyzing' | 'preview' | 'error';
    let step: Step = 'upload';
    let dragOver = false;
    let file: File | null = null;
    let errorMsg = '';
    let analyzeProgress = 0;
    let progressTimer: ReturnType<typeof setInterval>;

    interface PreviewQuestion {
        question_text: string;
        type: string;
        options: { text: string; is_correct: boolean }[];
        explanation: string;
        time_limit: number;
        points: number;
    }
    let parsed: PreviewQuestion[] = [];
    let selected: Set<number> = new Set();
    let fileInfo = { file_type: '', image_count: 0, text_len: 0, count: 0 };

    function uid() { return Math.random().toString(36).slice(2, 9); }

    function close() {
        reset();
        dispatch('close');
    }

    function reset() {
        step = 'upload';
        file = null;
        errorMsg = '';
        parsed = [];
        selected = new Set();
        analyzeProgress = 0;
        clearInterval(progressTimer);
    }

    // ── File handling ─────────────────────────────────────────────────────────
    function onDrop(e: DragEvent) {
        e.preventDefault();
        dragOver = false;
        const f = e.dataTransfer?.files[0];
        if (f) pickFile(f);
    }
    function onDragOver(e: DragEvent) { e.preventDefault(); dragOver = true; }
    function onDragLeave() { dragOver = false; }
    function onInput(e: Event) {
        const input = e.target as HTMLInputElement;
        if (input.files?.[0]) pickFile(input.files[0]);
    }

    function pickFile(f: File) {
        const ext = f.name.toLowerCase().split('.').pop() ?? '';
        if (!['docx','xlsx','xls','pdf'].includes(ext)) {
            errorMsg = 'Faqat .docx, .xlsx yoki .pdf fayl yuklay olasiz';
            step = 'error';
            return;
        }
        file = f;
        uploadAndAnalyze();
    }

    async function uploadAndAnalyze() {
        if (!file) return;
        step = 'analyzing';
        analyzeProgress = 0;

        // Fake progress animation
        progressTimer = setInterval(() => {
            analyzeProgress = Math.min(analyzeProgress + Math.random() * 8, 88);
        }, 400);

        try {
            const token = localStorage.getItem('token') ?? '';
            const form = new FormData();
            form.append('file', file);

            const res = await fetch(`${BASE_URL}/api/quizzes/import-file`, {
                method: 'POST',
                headers: { Authorization: `Bearer ${token}` },
                body: form,
            });

            clearInterval(progressTimer);
            analyzeProgress = 100;

            const json = await res.json();
            if (!res.ok) {
                errorMsg = json.error ?? 'Server xatosi';
                step = 'error';
                return;
            }

            parsed = json.questions ?? [];
            fileInfo = {
                file_type: json.file_type ?? '',
                image_count: json.image_count ?? 0,
                text_len: json.text_len ?? 0,
                count: json.count ?? parsed.length,
            };
            selected = new Set(parsed.map((_, i) => i));
            step = 'preview';
        } catch (e: any) {
            clearInterval(progressTimer);
            errorMsg = e.message ?? 'Tarmoq xatosi';
            step = 'error';
        }
    }

    function toggleAll() {
        if (selected.size === parsed.length) {
            selected = new Set();
        } else {
            selected = new Set(parsed.map((_, i) => i));
        }
    }
    function toggleOne(i: number) {
        if (selected.has(i)) { selected.delete(i); } else { selected.add(i); }
        selected = selected; // trigger reactivity
    }

    function doImport() {
        const questions: EditQuestion[] = [...selected].sort((a,b)=>a-b).map(i => {
            const q = parsed[i];
            return {
                _id: uid(),
                type: (q.type as any) || 'multiple_choice',
                text: q.question_text,
                time_limit: q.time_limit || 20,
                points: q.points || 100,
                explanation: q.explanation || '',
                options: q.type === 'short_answer' || q.type === 'fill_blank'
                    ? [{ _id: uid(), text: q.options?.[0]?.text ?? '', is_correct: true }]
                    : (q.options || []).map((o): EditOption => ({
                        _id: uid(),
                        text: o.text,
                        is_correct: o.is_correct,
                    })),
            };
        });
        dispatch('import', { questions });
        close();
    }

    function typeLabel(t: string) {
        const m: Record<string,string> = {
            multiple_choice: 'Ko\'p tanlov',
            true_false: 'To\'g\'ri/Noto\'g\'ri',
            short_answer: 'Qisqa javob',
            fill_blank: 'Bo\'sh to\'ldirish',
        };
        return m[t] ?? t;
    }
    function typeColor(t: string) {
        const m: Record<string,string> = {
            multiple_choice: '#6366f1',
            true_false: '#22c55e',
            short_answer: '#f59e0b',
            fill_blank: '#8b5cf6',
        };
        return m[t] ?? '#94a3b8';
    }

    $: correctLabel = (q: PreviewQuestion) => {
        const c = q.options?.filter(o => o.is_correct);
        if (!c?.length) return '—';
        return c.map(o => o.text).join(', ');
    };
    $: fileExt = file?.name.toLowerCase().split('.').pop() ?? '';
    $: fileIcon = fileExt === 'pdf' ? '📄' : fileExt === 'docx' ? '📝' : '📊';
</script>

{#if show}
<div class="backdrop" on:click|self={close} role="dialog" aria-modal="true">
<div class="modal">

    <!-- ══ HEADER ══════════════════════════════════════════════════════════════ -->
    <div class="mhd">
        <div class="mhd-left">
            <div class="mhd-icon">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2.2">
                    <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/>
                    <polyline points="17 8 12 3 7 8"/>
                    <line x1="12" y1="3" x2="12" y2="15"/>
                </svg>
            </div>
            <div>
                <h2>Hujjatdan import qilish</h2>
                <p>Word, Excel yoki PDF fayldan savollarni AI yordamida aniqlash</p>
            </div>
        </div>
        <button class="close-btn" on:click={close}>✕</button>
    </div>

    <!-- ══ STEP: UPLOAD ═════════════════════════════════════════════════════════ -->
    {#if step === 'upload'}
    <div class="body">
        <!-- Format info pills -->
        <div class="fmt-pills">
            <div class="fmt-pill">
                <span>📝</span>
                <div>
                    <strong>Word (.docx)</strong>
                    <small>Matn, rasm, formulalar</small>
                </div>
            </div>
            <div class="fmt-pill">
                <span>📊</span>
                <div>
                    <strong>Excel (.xlsx)</strong>
                    <small>Jadval formatidagi savollar</small>
                </div>
            </div>
            <div class="fmt-pill">
                <span>📄</span>
                <div>
                    <strong>PDF</strong>
                    <small>Har qanday PDF hujjat</small>
                </div>
            </div>
        </div>

        <!-- Drop zone -->
        <div
            class="dropzone"
            class:drag-over={dragOver}
            on:drop={onDrop}
            on:dragover={onDragOver}
            on:dragleave={onDragLeave}
            role="button"
            tabindex="0"
            on:click={() => document.getElementById('file-input')?.click()}
            on:keydown={e => e.key === 'Enter' && document.getElementById('file-input')?.click()}
        >
            <div class="dz-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                    <polyline points="14 2 14 8 20 8"/>
                    <line x1="12" y1="18" x2="12" y2="12"/>
                    <polyline points="9 15 12 12 15 15"/>
                </svg>
            </div>
            <p class="dz-title">Faylni shu yerga tashlang</p>
            <p class="dz-sub">yoki bosib tanlang</p>
            <div class="dz-formats">.docx &nbsp;·&nbsp; .xlsx &nbsp;·&nbsp; .pdf &nbsp;·&nbsp; max 20MB</div>
        </div>
        <input type="file" id="file-input" accept=".docx,.xlsx,.xls,.pdf" on:change={onInput} style="display:none" />

        <!-- Tips -->
        <div class="tips">
            <div class="tip-title">💡 Maslahatlar</div>
            <ul class="tip-list">
                <li>Savollar raqam (1. 2. 3.) yoki harf (A. B. C.) bilan ajratilgan bo'lsa yaxshi aniqlanadi</li>
                <li>To'g'ri javob * yulduzi, kurpivka, yoki "Javob:" qatori bilan belgilangan bo'lishi mumkin</li>
                <li>Formulalar LaTeX ($x^2+y^2=r^2$) yoki rasmda bo'lsa ham AI aniqlaydi</li>
                <li>Excel'da har qatorda bir savol bo'lsa eng to'g'ri natija beradi</li>
            </ul>
        </div>
    </div>

    <!-- ══ STEP: ANALYZING ══════════════════════════════════════════════════════ -->
    {:else if step === 'analyzing'}
    <div class="body analyzing">
        <div class="analyzing-ani">
            <div class="brain-wrap">
                <div class="brain">🧠</div>
                <div class="pulse r1"></div>
                <div class="pulse r2"></div>
                <div class="pulse r3"></div>
            </div>
        </div>
        <h3 class="ana-title">AI tahlil qilmoqda...</h3>
        <p class="ana-sub">
            {fileIcon} <strong>{file?.name}</strong> faylini o'qib, savollarni aniqlayapti
        </p>
        <div class="progress-wrap">
            <div class="progress-bar">
                <div class="progress-fill" style="width:{analyzeProgress}%"></div>
            </div>
            <span class="progress-pct">{Math.round(analyzeProgress)}%</span>
        </div>
        <div class="ana-steps">
            <div class="ana-step" class:done={analyzeProgress > 20}>
                <span class="as-dot"></span> Hujjat o'qilmoqda
            </div>
            <div class="ana-step" class:done={analyzeProgress > 50}>
                <span class="as-dot"></span> Matn va rasmlar ajratilmoqda
            </div>
            <div class="ana-step" class:done={analyzeProgress > 80}>
                <span class="as-dot"></span> AI savollarni aniqlayapti
            </div>
        </div>
    </div>

    <!-- ══ STEP: PREVIEW ════════════════════════════════════════════════════════ -->
    {:else if step === 'preview'}
    <div class="body preview">
        <!-- Summary bar -->
        <div class="summary-bar">
            <div class="sum-chip">
                <span>{fileIcon}</span>
                <span>{fileInfo.file_type}</span>
            </div>
            <div class="sum-chip green">
                <span>✅</span>
                <span>{parsed.length} ta savol topildi</span>
            </div>
            {#if fileInfo.image_count > 0}
            <div class="sum-chip blue">
                <span>🖼️</span>
                <span>{fileInfo.image_count} ta rasm</span>
            </div>
            {/if}
            <div class="sum-sep"></div>
            <button class="sel-toggle" on:click={toggleAll}>
                {selected.size === parsed.length ? 'Hammasini bekor' : 'Hammasini tanlash'}
            </button>
            <span class="sel-count">{selected.size}/{parsed.length} tanlangan</span>
        </div>

        <!-- Questions list -->
        <div class="q-list">
            {#each parsed as q, i}
            <div class="q-card" class:selected={selected.has(i)} on:click={() => toggleOne(i)} role="checkbox" aria-checked={selected.has(i)} tabindex="0" on:keydown={e => e.key === 'Enter' && toggleOne(i)}>
                <div class="q-check">
                    <div class="checkbox" class:checked={selected.has(i)}>
                        {#if selected.has(i)}<svg width="12" height="12" viewBox="0 0 12 12" fill="none"><polyline points="2,6 5,9 10,3" stroke="white" stroke-width="2" stroke-linecap="round"/></svg>{/if}
                    </div>
                </div>
                <div class="q-body">
                    <div class="q-meta">
                        <span class="q-num">{i + 1}</span>
                        <span class="q-type-badge" style="background:{typeColor(q.type)}20;color:{typeColor(q.type)}">
                            {typeLabel(q.type)}
                        </span>
                        <span class="q-pts">⏱ {q.time_limit}s</span>
                        <span class="q-pts">⭐ {q.points}</span>
                    </div>
                    <p class="q-text">{q.question_text}</p>
                    {#if q.options?.length}
                    <div class="q-options">
                        {#each q.options as opt}
                        <div class="q-opt" class:correct={opt.is_correct}>
                            <span class="opt-dot" class:dot-correct={opt.is_correct}></span>
                            <span class="opt-text">{opt.text}</span>
                            {#if opt.is_correct}<span class="opt-correct-tag">✓ To'g'ri</span>{/if}
                        </div>
                        {/each}
                    </div>
                    {/if}
                    {#if q.explanation}
                    <div class="q-explanation">
                        <span>💬</span> {q.explanation}
                    </div>
                    {/if}
                </div>
            </div>
            {/each}
        </div>
    </div>

    <!-- ══ STEP: ERROR ══════════════════════════════════════════════════════════ -->
    {:else if step === 'error'}
    <div class="body error-body">
        <div class="err-icon">❌</div>
        <h3>Xato yuz berdi</h3>
        <p class="err-msg">{errorMsg}</p>
        <button class="btn primary" on:click={reset}>Qayta urinish</button>
    </div>
    {/if}

    <!-- ══ FOOTER ══════════════════════════════════════════════════════════════ -->
    {#if step === 'preview'}
    <div class="mfoot">
        <button class="btn ghost" on:click={reset}>
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
            Orqaga
        </button>
        <div class="foot-right">
            <span class="foot-info">{selected.size} ta savol qo'shiladi</span>
            <button class="btn primary" on:click={doImport} disabled={selected.size === 0}>
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><polyline points="20 6 9 17 4 12"/></svg>
                Quizga qo'shish
            </button>
        </div>
    </div>
    {/if}

</div>
</div>
{/if}

<style>
    /* ── Backdrop ────────────────────────────────────────────────────────────── */
    .backdrop {
        position:fixed; inset:0; background:rgba(0,0,0,0.6); backdrop-filter:blur(5px);
        z-index:1100; display:flex; align-items:center; justify-content:center;
        padding:20px; animation:fadein 0.2s ease;
    }
    @keyframes fadein { from{opacity:0} to{opacity:1} }
    .modal {
        background:var(--white); border-radius:20px; width:100%; max-width:720px;
        max-height:90vh; display:flex; flex-direction:column;
        box-shadow:0 30px 90px rgba(0,0,0,0.35); animation:slideup 0.25s ease;
        overflow:hidden;
    }
    @keyframes slideup { from{opacity:0;transform:translateY(20px)} to{opacity:1;transform:translateY(0)} }

    /* ── Header ──────────────────────────────────────────────────────────────── */
    .mhd {
        display:flex; align-items:center; justify-content:space-between;
        padding:22px 28px; border-bottom:1.5px solid var(--bg2);
        background:linear-gradient(135deg, rgba(99,102,241,0.07), rgba(139,92,246,0.04));
        flex-shrink:0;
    }
    .mhd-left { display:flex; align-items:center; gap:14px; }
    .mhd-icon {
        width:46px; height:46px; border-radius:13px; flex-shrink:0;
        background:linear-gradient(135deg,#6366f1,#8b5cf6);
        display:flex; align-items:center; justify-content:center;
        box-shadow:0 4px 14px rgba(99,102,241,0.4);
    }
    .mhd h2 { font-size:1.1rem; font-weight:800; color:var(--text); margin:0 0 3px; }
    .mhd p { font-size:0.78rem; color:var(--text3); margin:0; }
    .close-btn { background:var(--bg2); border:none; border-radius:50%; width:34px; height:34px; display:flex; align-items:center; justify-content:center; cursor:pointer; color:var(--text2); font-size:0.85rem; transition:all 0.2s; flex-shrink:0; }
    .close-btn:hover { background:var(--danger); color:white; }

    /* ── Body ────────────────────────────────────────────────────────────────── */
    .body { padding:24px 28px; overflow-y:auto; flex:1; min-height:0; }

    /* ── Format pills ────────────────────────────────────────────────────────── */
    .fmt-pills { display:grid; grid-template-columns:repeat(3,1fr); gap:10px; margin-bottom:20px; }
    .fmt-pill { display:flex; align-items:center; gap:10px; background:var(--bg2); border-radius:12px; padding:12px 14px; border:1.5px solid var(--border); }
    .fmt-pill span { font-size:1.6rem; flex-shrink:0; }
    .fmt-pill strong { display:block; font-size:0.82rem; font-weight:700; color:var(--text); }
    .fmt-pill small { font-size:0.72rem; color:var(--text3); }

    /* ── Drop zone ───────────────────────────────────────────────────────────── */
    .dropzone {
        border:2.5px dashed var(--border); border-radius:16px;
        padding:48px 32px; text-align:center; cursor:pointer;
        transition:all 0.2s; background:var(--bg2);
    }
    .dropzone:hover, .dropzone.drag-over {
        border-color:var(--primary); background:rgba(99,102,241,0.05);
    }
    .dropzone.drag-over { transform:scale(1.01); }
    .dz-icon { color:var(--text3); margin-bottom:12px; }
    .dz-title { font-size:1rem; font-weight:700; color:var(--text); margin:0 0 4px; }
    .dz-sub { font-size:0.85rem; color:var(--text3); margin:0 0 12px; }
    .dz-formats { font-size:0.75rem; color:var(--text3); background:var(--white); display:inline-block; padding:4px 12px; border-radius:20px; border:1px solid var(--border); }

    /* ── Tips ────────────────────────────────────────────────────────────────── */
    .tips { margin-top:20px; background:rgba(99,102,241,0.05); border:1.5px solid rgba(99,102,241,0.15); border-radius:12px; padding:14px 18px; }
    .tip-title { font-size:0.82rem; font-weight:700; color:var(--primary); margin-bottom:8px; }
    .tip-list { margin:0; padding-left:18px; }
    .tip-list li { font-size:0.78rem; color:var(--text2); margin-bottom:5px; line-height:1.5; }

    /* ── Analyzing ───────────────────────────────────────────────────────────── */
    .analyzing { text-align:center; padding:40px 28px; }
    .analyzing-ani { margin-bottom:24px; }
    .brain-wrap { position:relative; display:inline-block; width:90px; height:90px; }
    .brain { font-size:3rem; position:absolute; top:50%; left:50%; transform:translate(-50%,-50%); z-index:2; animation:bob 1.5s ease-in-out infinite; }
    @keyframes bob { 0%,100%{transform:translate(-50%,-50%) scale(1)} 50%{transform:translate(-50%,-52%) scale(1.05)} }
    .pulse { position:absolute; top:50%; left:50%; transform:translate(-50%,-50%); border-radius:50%; border:2px solid rgba(99,102,241,0.35); animation:pulse-out 2s ease-out infinite; }
    .pulse.r1 { width:70px; height:70px; animation-delay:0s; }
    .pulse.r2 { width:90px; height:90px; animation-delay:0.5s; }
    .pulse.r3 { width:110px; height:110px; animation-delay:1s; }
    @keyframes pulse-out { 0%{opacity:1;transform:translate(-50%,-50%) scale(0.8)} 100%{opacity:0;transform:translate(-50%,-50%) scale(1.4)} }
    .ana-title { font-size:1.2rem; font-weight:800; color:var(--text); margin:0 0 8px; }
    .ana-sub { font-size:0.85rem; color:var(--text3); margin:0 0 24px; }
    .progress-wrap { display:flex; align-items:center; gap:12px; margin-bottom:20px; }
    .progress-bar { flex:1; height:8px; background:var(--bg2); border-radius:99px; overflow:hidden; }
    .progress-fill { height:100%; background:linear-gradient(90deg,#6366f1,#8b5cf6); border-radius:99px; transition:width 0.4s ease; }
    .progress-pct { font-size:0.8rem; font-weight:700; color:var(--primary); width:36px; text-align:right; }
    .ana-steps { display:flex; flex-direction:column; gap:8px; text-align:left; max-width:280px; margin:0 auto; }
    .ana-step { display:flex; align-items:center; gap:10px; font-size:0.8rem; color:var(--text3); transition:color 0.3s; }
    .ana-step.done { color:var(--text); }
    .as-dot { width:8px; height:8px; border-radius:50%; background:var(--border); flex-shrink:0; transition:background 0.3s; }
    .ana-step.done .as-dot { background:#22c55e; }

    /* ── Preview ─────────────────────────────────────────────────────────────── */
    .preview { padding:0; }
    .summary-bar {
        display:flex; align-items:center; gap:8px; flex-wrap:wrap;
        padding:14px 24px; background:var(--bg2); border-bottom:1px solid var(--border);
        flex-shrink:0; position:sticky; top:0; z-index:1;
    }
    .sum-chip { display:flex; align-items:center; gap:5px; background:var(--white); border:1.5px solid var(--border); border-radius:8px; padding:5px 10px; font-size:0.78rem; font-weight:600; color:var(--text2); }
    .sum-chip.green { border-color:#bbf7d0; color:#15803d; background:#f0fdf4; }
    .sum-chip.blue { border-color:#bfdbfe; color:#1d4ed8; background:#eff6ff; }
    .sum-sep { flex:1; }
    .sel-toggle { background:transparent; border:none; font-size:0.78rem; color:var(--primary); cursor:pointer; font-weight:600; text-decoration:underline; padding:4px; }
    .sel-count { font-size:0.78rem; color:var(--text3); font-weight:600; }
    .q-list { padding:16px 24px; display:flex; flex-direction:column; gap:10px; }
    .q-card {
        display:flex; gap:12px; background:var(--bg2); border-radius:14px; padding:14px;
        cursor:pointer; transition:all 0.15s; border:2px solid transparent;
        user-select:none;
    }
    .q-card:hover { border-color:var(--border); background:var(--white); }
    .q-card.selected { border-color:var(--primary); background:rgba(99,102,241,0.04); }
    .q-check { flex-shrink:0; padding-top:2px; }
    .checkbox {
        width:20px; height:20px; border-radius:6px; border:2px solid var(--border);
        background:var(--white); display:flex; align-items:center; justify-content:center;
        transition:all 0.15s;
    }
    .checkbox.checked { background:var(--primary); border-color:var(--primary); }
    .q-body { flex:1; min-width:0; }
    .q-meta { display:flex; align-items:center; gap:6px; margin-bottom:6px; flex-wrap:wrap; }
    .q-num { font-size:0.72rem; font-weight:800; color:var(--text3); background:var(--border); border-radius:5px; padding:1px 6px; }
    .q-type-badge { font-size:0.7rem; font-weight:700; padding:2px 8px; border-radius:6px; }
    .q-pts { font-size:0.72rem; color:var(--text3); background:var(--white); border:1px solid var(--border); border-radius:5px; padding:1px 6px; }
    .q-text { font-size:0.875rem; font-weight:600; color:var(--text); margin:0 0 10px; line-height:1.5; }
    .q-options { display:flex; flex-direction:column; gap:4px; margin-bottom:8px; }
    .q-opt { display:flex; align-items:center; gap:8px; padding:5px 8px; border-radius:8px; background:var(--white); border:1.5px solid var(--border); font-size:0.8rem; }
    .q-opt.correct { background:rgba(34,197,94,0.08); border-color:rgba(34,197,94,0.4); }
    .opt-dot { width:7px; height:7px; border-radius:50%; background:var(--border); flex-shrink:0; }
    .opt-dot.dot-correct { background:#22c55e; }
    .opt-text { flex:1; color:var(--text2); }
    .opt-correct-tag { font-size:0.68rem; font-weight:700; color:#16a34a; background:#dcfce7; padding:1px 6px; border-radius:4px; white-space:nowrap; }
    .q-explanation { font-size:0.75rem; color:var(--text3); background:rgba(99,102,241,0.06); border-radius:6px; padding:6px 10px; display:flex; gap:5px; line-height:1.4; }

    /* ── Error ───────────────────────────────────────────────────────────────── */
    .error-body { text-align:center; padding:40px 28px; }
    .err-icon { font-size:3rem; margin-bottom:12px; }
    .error-body h3 { font-size:1.1rem; font-weight:700; color:var(--text); margin:0 0 8px; }
    .err-msg { font-size:0.875rem; color:var(--text3); margin:0 0 24px; background:var(--bg2); border-radius:10px; padding:12px 16px; }

    /* ── Footer ──────────────────────────────────────────────────────────────── */
    .mfoot { display:flex; align-items:center; justify-content:space-between; padding:16px 28px; border-top:1.5px solid var(--bg2); background:var(--white); flex-shrink:0; }
    .foot-right { display:flex; align-items:center; gap:12px; }
    .foot-info { font-size:0.8rem; color:var(--text3); font-weight:600; }

    /* ── Buttons ─────────────────────────────────────────────────────────────── */
    .btn { padding:9px 18px; border:none; border-radius:var(--radius); font-size:0.875rem; font-weight:600; cursor:pointer; display:inline-flex; align-items:center; gap:7px; transition:var(--transition); }
    .btn.primary { background:linear-gradient(135deg,var(--primary),var(--accent)); color:white; box-shadow:0 3px 10px rgba(99,102,241,0.3); }
    .btn.primary:hover:not(:disabled) { transform:translateY(-1px); box-shadow:0 6px 16px rgba(99,102,241,0.4); }
    .btn.primary:disabled { opacity:0.5; cursor:not-allowed; }
    .btn.ghost { background:transparent; border:1.5px solid var(--border); color:var(--text2); }
    .btn.ghost:hover { border-color:var(--primary); color:var(--primary); }

    @media(max-width:640px) {
        .mhd, .body, .mfoot { padding-left:16px; padding-right:16px; }
        .fmt-pills { grid-template-columns:1fr; }
        .summary-bar { padding:10px 16px; }
        .q-list { padding:12px 16px; }
    }
</style>
