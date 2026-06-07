<script lang="ts">
    import { goto } from '$app/navigation';
    import { quizzes as quizzesApi } from '$lib/api/client';
    import QuizEditor from '$lib/components/QuizEditor.svelte';
    import type { QuizMeta, EditQuestion, EditOption } from '$lib/components/QuizEditor.svelte';

    type Mode = 'manual' | 'import' | 'ai';
    let mode: Mode | null = null;
    let saving = false;

    function choose(m: Mode) {
        mode = m;
    }

    async function handleSave(e: CustomEvent<{ meta: QuizMeta; questions: EditQuestion[] }>) {
        saving = true;
        try {
            const { meta, questions } = e.detail;
            const payload: any = {
                ...meta,
                template: 'quiz',
                questions: questions.map((q, i) => ({
                    order_index: i,
                    type: q.type,
                    question_text: q.text,
                    time_limit: q.time_limit,
                    points: q.points,
                    explanation: q.explanation,
                    options: q.type === 'short_answer' || q.type === 'fill_blank'
                        ? [{ option_text: q.options[0]?.text ?? '', is_correct: true, order_index: 0 }]
                        : q.options.map((o: EditOption, j: number) => ({
                            option_text: o.text,
                            is_correct: o.is_correct,
                            order_index: j,
                          }))
                }))
            };
            const res = await quizzesApi.create(payload);
            goto(`/dashboard/quizzes/${res.id}`);
        } catch (err: any) {
            saving = false;
            throw err;
        }
    }
</script>

<svelte:head><title>Yangi quiz — Cognita.uz</title></svelte:head>

{#if mode === null}
    <!-- ── Selection Screen ── -->
    <div class="select-wrap">
        <div class="select-header">
            <a href="/dashboard/quizzes" class="back-link">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
                Quizlarga qaytish
            </a>
            <h1>Yangi quiz yaratish</h1>
            <p>Quizni qanday yaratmoqchisiz?</p>
        </div>

        <div class="cards">
            <!-- Manual -->
            <button class="card card-manual" type="button" on:click={() => choose('manual')}>
                <div class="card-icon">
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                        <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
                        <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
                    </svg>
                </div>
                <div class="card-body">
                    <h2>O'zim tuzaman</h2>
                    <p>Savollarni qo'lda yozib, variantlarini belgilab chiqasiz</p>
                </div>
                <div class="card-tags">
                    <span>Ko'p tanlov</span>
                    <span>To'g'ri/Noto'g'ri</span>
                    <span>Qisqa javob</span>
                </div>
                <div class="card-arrow">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="9 18 15 12 9 6"/></svg>
                </div>
            </button>

            <!-- Import -->
            <button class="card card-import" type="button" on:click={() => choose('import')}>
                <div class="card-icon">
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                        <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                        <polyline points="14 2 14 8 20 8"/>
                        <line x1="12" y1="18" x2="12" y2="12"/>
                        <line x1="9" y1="15" x2="15" y2="15"/>
                    </svg>
                </div>
                <div class="card-body">
                    <h2>Fayldan import</h2>
                    <p>Word, Excel yoki PDF hujjatdan savollarni avtomatik oladi</p>
                </div>
                <div class="card-tags">
                    <span>.docx</span>
                    <span>.xlsx</span>
                    <span>.pdf</span>
                </div>
                <div class="card-arrow">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="9 18 15 12 9 6"/></svg>
                </div>
            </button>

            <!-- AI -->
            <button class="card card-ai" type="button" on:click={() => choose('ai')}>
                <div class="card-icon">
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                        <path d="M12 2a2 2 0 012 2c0 .74-.4 1.39-1 1.73V7h1a7 7 0 017 7h1a1 1 0 010 2h-1v1a2 2 0 01-2 2H5a2 2 0 01-2-2v-1H2a1 1 0 010-2h1a7 7 0 017-7h1V5.73c-.6-.34-1-.99-1-1.73a2 2 0 012-2z"/>
                        <circle cx="7.5" cy="13.5" r="1.5"/>
                        <circle cx="16.5" cy="13.5" r="1.5"/>
                    </svg>
                </div>
                <div class="card-body">
                    <h2>AI bilan yaratish</h2>
                    <p>Mavzuni kiriting — sun'iy intellekt savollarni tayyorlaydi</p>
                </div>
                <div class="card-tags">
                    <span>✨ Avtomatik</span>
                    <span>O'zbek tilida</span>
                    <span>Tez va qulay</span>
                </div>
                <div class="card-arrow">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="9 18 15 12 9 6"/></svg>
                </div>
            </button>
        </div>
    </div>

{:else}
    <!-- ── Editor ── -->
    <QuizEditor
        pageTitle="Yangi quiz"
        autoOpen={mode === 'manual' ? null : mode}
        {saving}
        on:save={handleSave}
    />
{/if}

<style>
.select-wrap {
    min-height: calc(100vh - 80px);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    background: var(--bg2, #f8fafc);
    margin: -28px -32px;
}

.select-header {
    text-align: center;
    margin-bottom: 40px;
}
.back-link {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: var(--primary, #6366f1);
    text-decoration: none;
    font-size: 0.85rem;
    font-weight: 600;
    margin-bottom: 24px;
    opacity: 0.8;
    transition: opacity 0.15s;
}
.back-link:hover { opacity: 1; }
.select-header h1 {
    font-size: 2rem;
    font-weight: 900;
    color: #0f172a;
    margin-bottom: 8px;
}
.select-header p {
    font-size: 1rem;
    color: #64748b;
}

.cards {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
    width: 100%;
    max-width: 860px;
}

.card {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 28px 24px;
    border-radius: 20px;
    border: 2px solid transparent;
    cursor: pointer;
    text-align: left;
    font-family: inherit;
    transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
    background: white;
    box-shadow: 0 2px 12px rgba(0,0,0,0.06);
}
.card:hover {
    transform: translateY(-6px);
    box-shadow: 0 12px 36px rgba(0,0,0,0.12);
}
.card:active { transform: translateY(-2px) scale(0.99); }

.card-manual:hover { border-color: #6366f1; }
.card-import:hover { border-color: #0ea5e9; }
.card-ai:hover { border-color: #a855f7; }

.card-icon {
    width: 64px;
    height: 64px;
    border-radius: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
}
.card-manual .card-icon { background: #eef2ff; color: #6366f1; }
.card-import .card-icon { background: #e0f2fe; color: #0284c7; }
.card-ai .card-icon {
    background: linear-gradient(135deg, #f5f3ff, #fdf4ff);
    color: #9333ea;
}

.card-body h2 {
    font-size: 1.15rem;
    font-weight: 800;
    color: #0f172a;
    margin-bottom: 6px;
}
.card-body p {
    font-size: 0.84rem;
    color: #64748b;
    line-height: 1.5;
}

.card-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
    margin-top: auto;
}
.card-tags span {
    font-size: 0.72rem;
    font-weight: 600;
    padding: 3px 9px;
    border-radius: 99px;
    background: #f1f5f9;
    color: #64748b;
}
.card-manual .card-tags span { background: #eef2ff; color: #6366f1; }
.card-import .card-tags span { background: #e0f2fe; color: #0284c7; }
.card-ai .card-tags span { background: #f5f3ff; color: #9333ea; }

.card-arrow {
    position: absolute;
    top: 24px;
    right: 20px;
    color: #cbd5e1;
    transition: all 0.2s;
}
.card:hover .card-arrow { color: inherit; transform: translateX(3px); }
.card-manual:hover .card-arrow { color: #6366f1; }
.card-import:hover .card-arrow { color: #0284c7; }
.card-ai:hover .card-arrow { color: #9333ea; }

/* AI card special glow */
.card-ai {
    background: linear-gradient(135deg, #ffffff 60%, #fdf4ff);
}
.card-ai::before {
    content: '';
    position: absolute;
    inset: -2px;
    border-radius: 22px;
    background: linear-gradient(135deg, #a855f7, #ec4899);
    z-index: -1;
    opacity: 0;
    transition: opacity 0.2s;
}
.card-ai:hover::before { opacity: 1; }
.card-ai:hover { border-color: transparent; }

@media (max-width: 700px) {
    .cards { grid-template-columns: 1fr; max-width: 420px; }
    .select-header h1 { font-size: 1.5rem; }
}
</style>
