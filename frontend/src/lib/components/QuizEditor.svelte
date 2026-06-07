<script context="module" lang="ts">
    import type { QuestionType } from '$lib/api/types';

    export interface EditOption {
        _id: string;
        text: string;
        is_correct: boolean;
    }
    export interface EditQuestion {
        _id: string;
        type: QuestionType;
        text: string;
        time_limit: number;
        points: number;
        explanation: string;
        options: EditOption[];
    }
    export interface QuizMeta {
        title: string;
        description: string;
        subject: string;
        grade_level: string;
        is_public: boolean;
        tags: string[];
    }
</script>

<script lang="ts">
    import { createEventDispatcher, onMount } from 'svelte';
    import { ai as aiApi } from '$lib/api/client';
    import ImportModal from './ImportModal.svelte';

    export let initialTitle = '';
    export let initialDescription = '';
    export let initialSubject = '';
    export let initialGrade = '';
    export let initialPublic = false;
    export let initialTags: string[] = [];
    export let initialQuestions: EditQuestion[] = [];
    export let saving = false;
    export let pageTitle = 'Yangi quiz';
    export let backHref = '/dashboard/quizzes';
    export let autoOpen: 'import' | 'ai' | null = null;

    const dispatch = createEventDispatcher<{
        save: { meta: QuizMeta; questions: EditQuestion[] }
    }>();

    let title       = initialTitle;
    let description = initialDescription;
    let subject     = initialSubject;
    let gradeLevel  = initialGrade;
    let isPublic    = initialPublic;
    let tags        = [...initialTags];
    let tagInput    = '';
    let questions   = initialQuestions.length ? [...initialQuestions] : [newQuestion()];
    let activeIdx   = 0;
    let error       = '';
    let metaOpen    = true;

    let showImport = false;
    function onImported(e: CustomEvent<{ questions: EditQuestion[] }>) {
        const incoming = e.detail.questions;
        questions = [...questions.filter(q => q.text || q.options.some(o => o.text)), ...incoming];
        activeIdx = questions.length - incoming.length;
        showImport = false;
    }

    let showAI    = false;
    let aiTopic   = '';
    let aiCount   = 5;
    let aiGrade   = gradeLevel || '';
    let aiLang    = 'uz';
    let aiTypes   = ['multiple_choice'];
    let aiLoading = false;
    let aiError   = '';
    let aiResult: EditQuestion[] = [];
    let aiSelected: Set<number> = new Set();

    $: activeQ = questions[activeIdx] ?? null;
    $: totalTime = questions.reduce((s, q) => s + q.time_limit, 0);

    onMount(() => {
        if (autoOpen === 'import') showImport = true;
        else if (autoOpen === 'ai') showAI = true;
    });

    function uid() { return Math.random().toString(36).slice(2, 9); }

    function newQuestion(): EditQuestion {
        return {
            _id: uid(), type: 'multiple_choice', text: '',
            time_limit: 20, points: 100, explanation: '',
            options: [
                { _id: uid(), text: '', is_correct: false },
                { _id: uid(), text: '', is_correct: false },
                { _id: uid(), text: '', is_correct: false },
                { _id: uid(), text: '', is_correct: false },
            ]
        };
    }

    function addQuestion() {
        questions = [...questions, newQuestion()];
        activeIdx = questions.length - 1;
    }

    function removeQuestion(idx: number) {
        if (questions.length === 1) { questions = [newQuestion()]; activeIdx = 0; return; }
        questions = questions.filter((_, i) => i !== idx);
        if (activeIdx >= questions.length) activeIdx = questions.length - 1;
    }

    function changeType(val: string) {
        questions[activeIdx].type = val as QuestionType;
        onTypeChange(questions[activeIdx]);
    }

    function onTypeChange(q: EditQuestion) {
        if (q.type === 'true_false') {
            q.options = [
                { _id: uid(), text: "To'g'ri", is_correct: true },
                { _id: uid(), text: "Noto'g'ri", is_correct: false },
            ];
        } else if (q.type === 'short_answer' || q.type === 'fill_blank') {
            q.options = [{ _id: uid(), text: '', is_correct: true }];
        } else if (q.type === 'poll') {
            q.options = [
                { _id: uid(), text: '', is_correct: false },
                { _id: uid(), text: '', is_correct: false },
            ];
        } else {
            q.options = [
                { _id: uid(), text: '', is_correct: false },
                { _id: uid(), text: '', is_correct: false },
                { _id: uid(), text: '', is_correct: false },
                { _id: uid(), text: '', is_correct: false },
            ];
        }
        questions = [...questions];
    }

    function setCorrect(q: EditQuestion, optId: string) {
        q.options = q.options.map(o => ({ ...o, is_correct: o._id === optId }));
        questions = [...questions];
    }

    function addOption(q: EditQuestion) {
        if (q.options.length >= 6) return;
        q.options = [...q.options, { _id: uid(), text: '', is_correct: false }];
        questions = [...questions];
    }

    function removeOption(q: EditQuestion, optId: string) {
        if (q.options.length <= 2) return;
        q.options = q.options.filter(o => o._id !== optId);
        questions = [...questions];
    }

    function addTag() {
        const t = tagInput.trim();
        if (t && !tags.includes(t)) tags = [...tags, t];
        tagInput = '';
    }

    function save() {
        if (!title.trim()) { error = 'Quiz nomini kiriting'; return; }
        if (questions.length === 0) { error = 'Kamida 1 ta savol kerak'; return; }
        const emptyQ = questions.findIndex(q => !q.text.trim());
        if (emptyQ !== -1) { error = `${emptyQ+1}-savol matni bo'sh`; activeIdx = emptyQ; return; }
        error = '';
        dispatch('save', {
            meta: { title: title.trim(), description: description.trim(), subject: subject.trim(), grade_level: gradeLevel, is_public: isPublic, tags },
            questions
        });
    }

    async function generateAI() {
        if (!aiTopic.trim()) { aiError = 'Mavzu kiriting'; return; }
        aiLoading = true; aiError = ''; aiResult = []; aiSelected = new Set();
        try {
            const raw = await aiApi.generate({
                topic: aiTopic.trim(),
                count: aiCount,
                grade_level: aiGrade || gradeLevel,
                language: aiLang,
                question_types: aiTypes,
            });
            aiResult = raw.map((q: any): EditQuestion => ({
                _id: uid(),
                type: q.type ?? 'multiple_choice',
                text: q.question_text,
                time_limit: q.time_limit ?? 20,
                points: q.points ?? 100,
                explanation: q.explanation ?? '',
                options: (q.options ?? []).map((o: any) => ({
                    _id: uid(), text: o.text, is_correct: o.is_correct ?? false,
                })),
            }));
            aiSelected = new Set(aiResult.map((_, i) => i));
        } catch (e: any) {
            aiError = e.message ?? 'Xato yuz berdi';
        } finally {
            aiLoading = false;
        }
    }

    function toggleAISelect(i: number) {
        if (aiSelected.has(i)) { aiSelected.delete(i); } else { aiSelected.add(i); }
        aiSelected = aiSelected;
    }

    function addAIQuestions() {
        const toAdd = aiResult.filter((_, i) => aiSelected.has(i));
        questions = [...questions, ...toAdd];
        activeIdx = questions.length - toAdd.length;
        showAI = false;
        aiResult = [];
    }

    const questionTypes = [
        { value: 'multiple_choice', label: 'Ko\'p tanlov',       icon: '⊙' },
        { value: 'true_false',      label: 'To\'g\'ri/Noto\'g\'ri', icon: '⊛' },
        { value: 'short_answer',    label: 'Qisqa javob',         icon: '✎' },
        { value: 'fill_blank',      label: 'Bo\'sh to\'ldirish',   icon: '▭' },
        { value: 'poll',            label: 'So\'rovnoma',          icon: '◉' },
    ];
    const grades    = ['1','2','3','4','5','6','7','8','9','10','11'];
    const timeOpts  = [5,10,15,20,30,45,60,90,120];
    const pointOpts = [50,100,200,500,1000];
    const aiTypeOpts = [
        { value: 'multiple_choice', label: 'Ko\'p tanlov' },
        { value: 'true_false',      label: 'To\'g\'ri/Noto\'g\'ri' },
        { value: 'short_answer',    label: 'Qisqa javob' },
    ];

    // Option colors: A=red, B=blue, C=yellow/gold, D=green, E=orange, F=purple
    const optColors = [
        { bg: '#ef4444', light: '#fef2f2', label: 'A' },
        { bg: '#3b82f6', light: '#eff6ff', label: 'B' },
        { bg: '#f59e0b', light: '#fffbeb', label: 'C' },
        { bg: '#10b981', light: '#ecfdf5', label: 'D' },
        { bg: '#f97316', light: '#fff7ed', label: 'E' },
        { bg: '#8b5cf6', light: '#f5f3ff', label: 'F' },
    ];

    function typeLabel(type: string) {
        return questionTypes.find(t => t.value === type)?.label ?? type;
    }
    function typeIcon(type: string) {
        return questionTypes.find(t => t.value === type)?.icon ?? '?';
    }

    function formatTime(s: number) {
        if (s < 60) return `${s}s`;
        return `${Math.floor(s/60)}m ${s%60>0?s%60+'s':''}`;
    }
</script>

<div class="editor">
    <!-- ══ Top Bar ══ -->
    <header class="topbar">
        <div class="topbar-left">
            <a href={backHref} class="back-btn">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
            </a>
            <div class="title-group">
                <input
                    class="title-input"
                    bind:value={title}
                    placeholder="Quiz nomini kiriting..."
                    spellcheck="false"
                />
                {#if error}<span class="title-error">⚠ {error}</span>{/if}
            </div>
        </div>
        <div class="topbar-right">
            <div class="stats-pill">
                <span>📋 {questions.length} savol</span>
                <span class="divider">·</span>
                <span>⏱ {formatTime(totalTime)}</span>
            </div>
            <button class="btn-ghost" on:click={() => showImport = true} type="button">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                Import
            </button>
            <button class="btn-ai" on:click={() => showAI = true} type="button">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
                AI yaratish
            </button>
            <button class="btn-save" on:click={save} disabled={saving} type="button">
                {#if saving}
                    <span class="spin"></span> Saqlanmoqda...
                {:else}
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M19 21H5a2 2 0 01-2-2V5a2 2 0 012-2h11l5 5v11a2 2 0 01-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
                    Saqlash
                {/if}
            </button>
        </div>
    </header>

    <div class="workspace">
        <!-- ══ Left Sidebar ══ -->
        <aside class="sidebar">
            <!-- Meta block -->
            <div class="sidebar-section">
                <button class="section-toggle" type="button" on:click={() => metaOpen = !metaOpen}>
                    <span class="section-title">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                        Ma'lumotlar
                    </span>
                    <svg class="chevron" class:open={metaOpen} width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="6 9 12 15 18 9"/></svg>
                </button>

                {#if metaOpen}
                    <div class="meta-fields">
                        <label class="meta-field">
                            <span>Tavsif</span>
                            <textarea bind:value={description} placeholder="Quiz haqida qisqa ma'lumot..." rows="2"></textarea>
                        </label>
                        <div class="meta-row2">
                            <label class="meta-field">
                                <span>Fan</span>
                                <input bind:value={subject} placeholder="Matematika" />
                            </label>
                            <label class="meta-field">
                                <span>Sinf</span>
                                <select bind:value={gradeLevel}>
                                    <option value="">—</option>
                                    {#each grades as g}<option value={g}>{g}</option>{/each}
                                </select>
                            </label>
                        </div>
                        <label class="meta-toggle">
                            <div class="toggle-track" class:on={isPublic} on:click={() => isPublic = !isPublic} role="switch" aria-checked={isPublic}>
                                <div class="toggle-thumb"></div>
                            </div>
                            <span>Ommaviy quiz</span>
                        </label>
                        <div class="tags-area">
                            {#each tags as t}
                                <span class="tag">
                                    {t}
                                    <button type="button" on:click={() => tags = tags.filter(x => x !== t)}>×</button>
                                </span>
                            {/each}
                            <input
                                class="tag-input"
                                bind:value={tagInput}
                                placeholder="+ Teg qo'shish"
                                on:keydown={e => e.key === 'Enter' && (e.preventDefault(), addTag())}
                                on:blur={addTag}
                            />
                        </div>
                    </div>
                {/if}
            </div>

            <!-- Question list -->
            <div class="qlist-section">
                <div class="qlist-header">
                    <span class="section-title">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg>
                        Savollar
                        <span class="q-count">{questions.length}</span>
                    </span>
                    <button class="add-q-icon" type="button" on:click={addQuestion} title="Savol qo'shish">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                    </button>
                </div>

                <div class="qlist">
                    {#each questions as q, i}
                        <button
                            type="button"
                            class="qlist-item"
                            class:active={activeIdx === i}
                            on:click={() => activeIdx = i}
                        >
                            <span class="q-num" class:active-num={activeIdx === i}>{i + 1}</span>
                            <div class="q-info">
                                <span class="q-preview">{q.text || 'Savol matni...'}</span>
                                <span class="q-type-badge">{typeIcon(q.type)} {typeLabel(q.type)}</span>
                            </div>
                            <button
                                type="button"
                                class="q-del"
                                on:click|stopPropagation={() => removeQuestion(i)}
                                title="O'chirish"
                            >
                                <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                            </button>
                        </button>
                    {/each}
                </div>

                <button type="button" class="add-q-btn" on:click={addQuestion}>
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                    Savol qo'shish
                </button>
            </div>
        </aside>

        <!-- ══ Main Editor ══ -->
        <main class="main">
            {#if activeQ}
                <div class="q-editor">
                    <!-- Header bar -->
                    <div class="q-header">
                        <div class="q-nav">
                            <button
                                type="button"
                                class="nav-btn"
                                disabled={activeIdx === 0}
                                on:click={() => activeIdx--}
                            >
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="15 18 9 12 15 6"/></svg>
                            </button>
                            <span class="q-counter">{activeIdx + 1} / {questions.length}</span>
                            <button
                                type="button"
                                class="nav-btn"
                                disabled={activeIdx === questions.length - 1}
                                on:click={() => activeIdx++}
                            >
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="9 18 15 12 9 6"/></svg>
                            </button>
                        </div>

                        <!-- Type selector -->
                        <div class="type-tabs">
                            {#each questionTypes as t}
                                <button
                                    type="button"
                                    class="type-tab"
                                    class:active={activeQ.type === t.value}
                                    on:click={() => changeType(t.value)}
                                >
                                    {t.icon} {t.label}
                                </button>
                            {/each}
                        </div>

                        <!-- Time & Points -->
                        <div class="q-settings">
                            <label class="setting-pill">
                                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                                <select bind:value={questions[activeIdx].time_limit}>
                                    {#each timeOpts as t}<option value={t}>{t}s</option>{/each}
                                </select>
                            </label>
                            <label class="setting-pill gold">
                                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                                <select bind:value={questions[activeIdx].points}>
                                    {#each pointOpts as p}<option value={p}>{p}</option>{/each}
                                </select>
                            </label>
                        </div>
                    </div>

                    <!-- Question text area -->
                    <div class="q-text-card">
                        <div class="q-text-label">Savol matni</div>
                        <textarea
                            bind:value={questions[activeIdx].text}
                            placeholder="Savolingizni shu yerga yozing..."
                            rows="3"
                            class="q-textarea"
                        ></textarea>
                    </div>

                    <!-- Answers -->
                    {#if activeQ.type === 'short_answer' || activeQ.type === 'fill_blank'}
                        <div class="answer-card">
                            <div class="answer-card-label">
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="2.5"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
                                To'g'ri javobni yozing
                            </div>
                            <input
                                class="short-answer-input"
                                bind:value={questions[activeIdx].options[0].text}
                                placeholder="To'g'ri javob..."
                            />
                        </div>

                    {:else if activeQ.type === 'true_false'}
                        <div class="tf-section">
                            <div class="tf-label">Javobni tanlang</div>
                            <div class="tf-options">
                                {#each activeQ.options as opt, oi}
                                    <button
                                        type="button"
                                        class="tf-card"
                                        class:correct={opt.is_correct}
                                        class:tf-true={oi === 0}
                                        class:tf-false={oi !== 0}
                                        on:click={() => setCorrect(activeQ, opt._id)}
                                    >
                                        <span class="tf-icon">{oi === 0 ? '✓' : '✗'}</span>
                                        <span class="tf-text">{opt.text}</span>
                                        {#if opt.is_correct}
                                            <span class="correct-badge">To'g'ri javob</span>
                                        {/if}
                                    </button>
                                {/each}
                            </div>
                        </div>

                    {:else}
                        <div class="options-section">
                            <div class="options-label">
                                {activeQ.type === 'poll' ? 'Variantlar' : "Variantlar — to'g'ri javobni bosing"}
                            </div>
                            <div class="options-grid" class:two-col={activeQ.options.length <= 2}>
                                {#each activeQ.options as opt, oi}
                                    {@const col = optColors[oi] ?? optColors[0]}
                                    <div
                                        class="opt-card"
                                        class:correct={opt.is_correct}
                                        style="--opt-bg:{col.bg}; --opt-light:{col.light}"
                                    >
                                        <button
                                            type="button"
                                            class="opt-letter"
                                            on:click={() => activeQ.type !== 'poll' && setCorrect(activeQ, opt._id)}
                                            title={activeQ.type !== 'poll' ? "To'g'ri javob sifatida belgilash" : ''}
                                        >
                                            {#if opt.is_correct}
                                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
                                            {:else}
                                                {col.label}
                                            {/if}
                                        </button>
                                        <input
                                            class="opt-input"
                                            bind:value={opt.text}
                                            placeholder="Variant {col.label}..."
                                        />
                                        <button
                                            type="button"
                                            class="opt-remove"
                                            on:click={() => removeOption(activeQ, opt._id)}
                                            disabled={activeQ.options.length <= 2}
                                            title="O'chirish"
                                        >
                                            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                                        </button>
                                    </div>
                                {/each}

                                {#if activeQ.options.length < 6}
                                    <button type="button" class="add-opt-card" on:click={() => addOption(activeQ)}>
                                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                                        Variant qo'shish
                                    </button>
                                {/if}
                            </div>
                        </div>
                    {/if}

                    <!-- Explanation -->
                    <details class="explanation-details">
                        <summary>
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
                            Izoh qo'shish (ixtiyoriy)
                        </summary>
                        <input
                            class="explanation-input"
                            bind:value={questions[activeIdx].explanation}
                            placeholder="Javob nima uchun to'g'ri ekanligini tushuntiring..."
                        />
                    </details>
                </div>
            {:else}
                <div class="empty-state">
                    <div class="empty-icon">📝</div>
                    <p>Chap paneldan savol tanlang yoki yangi savol qo'shing</p>
                    <button type="button" class="btn-primary" on:click={addQuestion}>+ Savol qo'shish</button>
                </div>
            {/if}
        </main>
    </div>
</div>

<!-- ══ Import Modal ══ -->
<ImportModal bind:show={showImport} on:import={onImported} on:close={() => showImport = false} />

<!-- ══ AI Modal ══ -->
{#if showAI}
    <div class="overlay" on:click|self={() => showAI = false} role="dialog" aria-modal="true">
        <div class="ai-modal">
            <div class="ai-head">
                <div class="ai-head-left">
                    <div class="ai-icon">✨</div>
                    <div>
                        <h2>AI bilan savollar yaratish</h2>
                        <p>Mavzu kiriting — AI savollarni avtomatik tayyorlaydi</p>
                    </div>
                </div>
                <button type="button" class="modal-close" on:click={() => showAI = false}>
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
            </div>

            {#if aiResult.length === 0}
                <div class="ai-form">
                    <label class="modal-field">
                        <span>Mavzu *</span>
                        <input
                            bind:value={aiTopic}
                            placeholder="Masalan: Fotosintez, Kvadrat tenglamalar..."
                            on:keydown={e => e.key === 'Enter' && generateAI()}
                            autofocus
                        />
                    </label>
                    <div class="ai-grid3">
                        <label class="modal-field">
                            <span>Savollar soni</span>
                            <select bind:value={aiCount}>
                                {#each [3,5,8,10,15] as n}<option value={n}>{n} ta</option>{/each}
                            </select>
                        </label>
                        <label class="modal-field">
                            <span>Sinf</span>
                            <select bind:value={aiGrade}>
                                <option value="">—</option>
                                {#each grades as g}<option value={g}>{g}</option>{/each}
                            </select>
                        </label>
                        <label class="modal-field">
                            <span>Til</span>
                            <select bind:value={aiLang}>
                                <option value="uz">O'zbek</option>
                                <option value="ru">Русский</option>
                                <option value="en">English</option>
                            </select>
                        </label>
                    </div>
                    <div>
                        <p class="modal-label">Savol turlari</p>
                        <div class="type-chips">
                            {#each aiTypeOpts as t}
                                <button
                                    type="button"
                                    class="type-chip"
                                    class:active={aiTypes.includes(t.value)}
                                    on:click={() => {
                                        if (aiTypes.includes(t.value)) {
                                            if (aiTypes.length > 1) aiTypes = aiTypes.filter(x => x !== t.value);
                                        } else {
                                            aiTypes = [...aiTypes, t.value];
                                        }
                                    }}
                                >{t.label}</button>
                            {/each}
                        </div>
                    </div>

                    {#if aiError}
                        <div class="ai-err">{aiError}</div>
                    {/if}

                    <button type="button" class="btn-ai-gen" on:click={generateAI} disabled={aiLoading || !aiTopic.trim()}>
                        {#if aiLoading}
                            <span class="spin"></span> Yaratilmoqda...
                        {:else}
                            ✨ Savollar yaratish
                        {/if}
                    </button>
                </div>

            {:else}
                <div class="ai-results">
                    <div class="ai-results-bar">
                        <p><strong>{aiResult.length}</strong> ta savol yaratildi</p>
                        <div class="sel-actions">
                            <button type="button" on:click={() => aiSelected = new Set(aiResult.map((_, i) => i))}>Barchasi</button>
                            <button type="button" on:click={() => aiSelected = new Set()}>Hech biri</button>
                        </div>
                    </div>
                    <div class="ai-q-list">
                        {#each aiResult as q, i}
                            <button type="button" class="ai-q-card" class:selected={aiSelected.has(i)} on:click={() => toggleAISelect(i)}>
                                <div class="ai-q-check" class:checked={aiSelected.has(i)}>
                                    {#if aiSelected.has(i)}<svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>{/if}
                                </div>
                                <div class="ai-q-content">
                                    <p class="ai-q-text">{q.text}</p>
                                    <div class="ai-q-opts">
                                        {#each q.options.slice(0, 4) as opt}
                                            <span class:ai-correct={opt.is_correct}>{opt.text}</span>
                                        {/each}
                                    </div>
                                </div>
                            </button>
                        {/each}
                    </div>
                    <div class="ai-result-footer">
                        <button type="button" class="btn-back" on:click={() => aiResult = []}>← Qayta</button>
                        <button type="button" class="btn-ai-add" on:click={addAIQuestions} disabled={aiSelected.size === 0}>
                            {aiSelected.size} ta savolni qo'shish →
                        </button>
                    </div>
                </div>
            {/if}
        </div>
    </div>
{/if}

<style>
/* ════ Layout ════ */
.editor {
    display: flex;
    flex-direction: column;
    height: 100%;
    margin: -28px -32px;
    background: #f8fafc;
    font-family: inherit;
}

/* ════ Top Bar ════ */
.topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 0 20px;
    height: 58px;
    background: white;
    border-bottom: 1px solid #e8ecf0;
    flex-shrink: 0;
    z-index: 10;
}
.topbar-left {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
    flex: 1;
}
.back-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    border-radius: 9px;
    background: #f1f5f9;
    color: #64748b;
    text-decoration: none;
    flex-shrink: 0;
    transition: all 0.15s;
}
.back-btn:hover { background: #e2e8f0; color: #334155; }

.title-group { display: flex; align-items: center; gap: 10px; min-width: 0; flex: 1; }
.title-input {
    font-size: 1.05rem;
    font-weight: 700;
    color: #0f172a;
    border: none;
    outline: none;
    background: transparent;
    min-width: 0;
    flex: 1;
    font-family: inherit;
    max-width: 400px;
}
.title-input::placeholder { color: #94a3b8; font-weight: 500; }
.title-input:focus { color: #6366f1; }
.title-error { font-size: 0.78rem; color: #ef4444; font-weight: 600; white-space: nowrap; }

.topbar-right { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }

.stats-pill {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    background: #f1f5f9;
    border-radius: 99px;
    font-size: 0.78rem;
    color: #64748b;
    font-weight: 600;
}
.stats-pill .divider { color: #cbd5e1; }

.btn-ghost {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 14px;
    border: 1.5px solid #e2e8f0;
    border-radius: 9px;
    background: white;
    color: #475569;
    font-size: 0.83rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s;
    font-family: inherit;
}
.btn-ghost:hover { border-color: #6366f1; color: #6366f1; background: #eef2ff; }

.btn-ai {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 16px;
    background: linear-gradient(135deg, #7c3aed, #a855f7);
    color: white;
    border: none;
    border-radius: 9px;
    font-size: 0.83rem;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.15s;
    box-shadow: 0 2px 8px rgba(124,58,237,0.3);
    font-family: inherit;
}
.btn-ai:hover { transform: translateY(-1px); box-shadow: 0 4px 14px rgba(124,58,237,0.4); }

.btn-save {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 18px;
    background: linear-gradient(135deg, #6366f1, #818cf8);
    color: white;
    border: none;
    border-radius: 9px;
    font-size: 0.85rem;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.15s;
    box-shadow: 0 2px 8px rgba(99,102,241,0.3);
    font-family: inherit;
}
.btn-save:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 14px rgba(99,102,241,0.4); }
.btn-save:disabled { opacity: 0.55; cursor: not-allowed; transform: none; }

.spin {
    width: 14px; height: 14px;
    border: 2px solid rgba(255,255,255,0.4);
    border-top-color: white;
    border-radius: 50%;
    display: inline-block;
    animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ════ Workspace ════ */
.workspace { display: flex; flex: 1; overflow: hidden; }

/* ════ Sidebar ════ */
.sidebar {
    width: 280px;
    flex-shrink: 0;
    background: white;
    border-right: 1px solid #e8ecf0;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
}

.sidebar-section {
    border-bottom: 1px solid #f1f5f9;
}
.section-toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 12px 16px;
    background: none;
    border: none;
    cursor: pointer;
    font-family: inherit;
    transition: background 0.15s;
}
.section-toggle:hover { background: #f8fafc; }
.section-title {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.72rem;
    font-weight: 800;
    color: #64748b;
    letter-spacing: 0.06em;
    text-transform: uppercase;
}
.chevron { color: #94a3b8; transition: transform 0.2s; }
.chevron.open { transform: rotate(180deg); }

.meta-fields {
    padding: 4px 16px 16px;
    display: flex;
    flex-direction: column;
    gap: 10px;
}
.meta-field { display: flex; flex-direction: column; gap: 4px; }
.meta-field span { font-size: 0.74rem; font-weight: 700; color: #64748b; }
.meta-field input,
.meta-field textarea,
.meta-field select {
    padding: 7px 10px;
    border: 1.5px solid #e2e8f0;
    border-radius: 8px;
    font-size: 0.84rem;
    font-family: inherit;
    color: #0f172a;
    background: #f8fafc;
    outline: none;
    transition: border-color 0.15s, background 0.15s;
}
.meta-field input:focus,
.meta-field textarea:focus,
.meta-field select:focus { border-color: #6366f1; background: white; }
.meta-field textarea { resize: vertical; }
.meta-row2 { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }

.meta-toggle { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.toggle-track {
    width: 36px; height: 20px;
    border-radius: 99px;
    background: #e2e8f0;
    position: relative;
    transition: background 0.2s;
    cursor: pointer;
    flex-shrink: 0;
}
.toggle-track.on { background: #6366f1; }
.toggle-thumb {
    width: 14px; height: 14px;
    background: white;
    border-radius: 50%;
    position: absolute;
    top: 3px; left: 3px;
    transition: transform 0.2s;
    box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}
.toggle-track.on .toggle-thumb { transform: translateX(16px); }
.meta-toggle span { font-size: 0.83rem; color: #475569; }

.tags-area { display: flex; flex-wrap: wrap; gap: 5px; align-items: center; min-height: 28px; }
.tag {
    display: flex; align-items: center; gap: 4px;
    background: #eef2ff; color: #6366f1;
    font-size: 0.73rem; font-weight: 600;
    padding: 3px 8px; border-radius: 99px;
}
.tag button { background: none; border: none; cursor: pointer; color: inherit; font-size: 0.9rem; padding: 0; line-height: 1; }
.tag-input { border: none; outline: none; font-size: 0.8rem; background: transparent; color: #64748b; min-width: 60px; font-family: inherit; }

/* Question list */
.qlist-section { flex: 1; display: flex; flex-direction: column; padding: 0; }
.qlist-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid #f1f5f9;
}
.q-count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: #6366f1;
    color: white;
    font-size: 0.65rem;
    font-weight: 800;
    padding: 1px 6px;
    border-radius: 99px;
    min-width: 18px;
}
.add-q-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 26px; height: 26px;
    background: #6366f1;
    color: white;
    border: none;
    border-radius: 7px;
    cursor: pointer;
    transition: all 0.15s;
}
.add-q-icon:hover { background: #4f46e5; transform: scale(1.08); }

.qlist { flex: 1; overflow-y: auto; padding: 8px; display: flex; flex-direction: column; gap: 3px; }

.qlist-item {
    display: flex;
    align-items: center;
    gap: 9px;
    padding: 9px 10px;
    border-radius: 10px;
    border: 1.5px solid transparent;
    background: transparent;
    cursor: pointer;
    text-align: left;
    width: 100%;
    transition: all 0.15s;
    position: relative;
}
.qlist-item:hover { background: #f8fafc; border-color: #e2e8f0; }
.qlist-item.active { background: #eef2ff; border-color: #c7d2fe; }

.q-num {
    width: 24px; height: 24px;
    background: #e2e8f0;
    color: #64748b;
    border-radius: 6px;
    font-size: 0.7rem;
    font-weight: 800;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
}
.q-num.active-num { background: #6366f1; color: white; }

.q-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.q-preview {
    font-size: 0.79rem;
    color: #334155;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-weight: 500;
}
.q-type-badge { font-size: 0.67rem; color: #94a3b8; font-weight: 500; }

.q-del {
    opacity: 0;
    display: flex; align-items: center; justify-content: center;
    width: 20px; height: 20px;
    background: none;
    border: none;
    border-radius: 5px;
    color: #94a3b8;
    cursor: pointer;
    transition: all 0.15s;
    flex-shrink: 0;
}
.qlist-item:hover .q-del { opacity: 1; }
.q-del:hover { background: #fee2e2; color: #ef4444; }

.add-q-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    margin: 8px;
    padding: 10px;
    border: 1.5px dashed #c7d2fe;
    border-radius: 10px;
    background: none;
    cursor: pointer;
    font-size: 0.82rem;
    font-weight: 600;
    color: #94a3b8;
    transition: all 0.15s;
    font-family: inherit;
}
.add-q-btn:hover { border-color: #6366f1; color: #6366f1; background: #eef2ff; }

/* ════ Main Editor ════ */
.main {
    flex: 1;
    overflow-y: auto;
    padding: 28px 32px;
    background: #f8fafc;
}

.q-editor {
    max-width: 700px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.q-header {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
}

.q-nav { display: flex; align-items: center; gap: 6px; }
.nav-btn {
    display: flex; align-items: center; justify-content: center;
    width: 30px; height: 30px;
    border: 1.5px solid #e2e8f0;
    border-radius: 8px;
    background: white;
    color: #64748b;
    cursor: pointer;
    transition: all 0.15s;
}
.nav-btn:hover:not(:disabled) { border-color: #6366f1; color: #6366f1; }
.nav-btn:disabled { opacity: 0.35; cursor: not-allowed; }
.q-counter { font-size: 0.78rem; font-weight: 700; color: #64748b; min-width: 40px; text-align: center; }

.type-tabs {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
    flex: 1;
}
.type-tab {
    padding: 5px 10px;
    border: 1.5px solid #e2e8f0;
    border-radius: 99px;
    background: white;
    color: #64748b;
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s;
    font-family: inherit;
    white-space: nowrap;
}
.type-tab:hover { border-color: #6366f1; color: #6366f1; }
.type-tab.active { background: #6366f1; border-color: #6366f1; color: white; }

.q-settings { display: flex; gap: 6px; }
.setting-pill {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 5px 10px;
    background: white;
    border: 1.5px solid #e2e8f0;
    border-radius: 99px;
    color: #64748b;
    cursor: pointer;
    transition: border-color 0.15s;
}
.setting-pill:hover { border-color: #6366f1; }
.setting-pill select {
    border: none;
    background: transparent;
    outline: none;
    font-size: 0.78rem;
    font-weight: 700;
    color: #334155;
    cursor: pointer;
    font-family: inherit;
}
.setting-pill svg { flex-shrink: 0; color: #94a3b8; }
.setting-pill.gold svg { color: #f59e0b; }
.setting-pill.gold select { color: #92400e; }

/* Question text */
.q-text-card {
    background: white;
    border-radius: 14px;
    border: 2px solid #e8ecf0;
    overflow: hidden;
    transition: border-color 0.2s;
}
.q-text-card:focus-within { border-color: #6366f1; }
.q-text-label {
    padding: 10px 16px 6px;
    font-size: 0.72rem;
    font-weight: 800;
    color: #94a3b8;
    letter-spacing: 0.06em;
    text-transform: uppercase;
}
.q-textarea {
    display: block;
    width: 100%;
    border: none;
    outline: none;
    font-size: 1.05rem;
    font-weight: 500;
    color: #0f172a;
    padding: 6px 16px 16px;
    resize: vertical;
    min-height: 90px;
    font-family: inherit;
    background: transparent;
    line-height: 1.5;
}
.q-textarea::placeholder { color: #c8d0dc; }

/* Short answer */
.answer-card {
    background: white;
    border-radius: 14px;
    border: 2px solid #e8ecf0;
    padding: 16px;
}
.answer-card-label {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.8rem;
    font-weight: 700;
    color: #10b981;
    margin-bottom: 10px;
}
.short-answer-input {
    width: 100%;
    padding: 11px 14px;
    border: 2px solid #d1fae5;
    border-radius: 10px;
    font-size: 0.95rem;
    background: #f0fdf4;
    outline: none;
    font-family: inherit;
    transition: border-color 0.2s;
    color: #065f46;
    font-weight: 600;
}
.short-answer-input:focus { border-color: #10b981; background: white; }

/* True/False */
.tf-section { display: flex; flex-direction: column; gap: 8px; }
.tf-label {
    font-size: 0.78rem;
    font-weight: 700;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}
.tf-options { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.tf-card {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 24px 16px;
    border-radius: 16px;
    border: 2.5px solid #e2e8f0;
    background: white;
    cursor: pointer;
    transition: all 0.2s;
    font-family: inherit;
}
.tf-card:hover { transform: translateY(-2px); box-shadow: 0 6px 20px rgba(0,0,0,0.08); }
.tf-card.tf-true:hover { border-color: #10b981; }
.tf-card.tf-false:hover { border-color: #ef4444; }
.tf-card.correct.tf-true { border-color: #10b981; background: #f0fdf4; }
.tf-card.correct.tf-false { border-color: #ef4444; background: #fef2f2; }
.tf-icon {
    font-size: 2rem;
    font-weight: 900;
}
.tf-card.tf-true .tf-icon { color: #10b981; }
.tf-card.tf-false .tf-icon { color: #ef4444; }
.tf-text { font-size: 1rem; font-weight: 700; color: #334155; }
.correct-badge {
    position: absolute;
    top: -1px; right: -1px;
    background: #10b981;
    color: white;
    font-size: 0.65rem;
    font-weight: 700;
    padding: 3px 8px;
    border-radius: 0 14px 0 10px;
}
.tf-card.tf-false .correct-badge { background: #ef4444; }

/* Multiple choice options */
.options-section { display: flex; flex-direction: column; gap: 10px; }
.options-label {
    font-size: 0.78rem;
    font-weight: 700;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}
.options-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
}
.options-grid.two-col { grid-template-columns: 1fr 1fr; }

.opt-card {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    background: white;
    border-radius: 13px;
    border: 2px solid #e8ecf0;
    transition: all 0.2s;
    position: relative;
}
.opt-card:hover { border-color: var(--opt-bg); background: var(--opt-light); }
.opt-card.correct { border-color: var(--opt-bg); background: var(--opt-light); }

.opt-letter {
    width: 34px; height: 34px;
    background: var(--opt-bg);
    color: white;
    border: none;
    border-radius: 9px;
    font-size: 0.85rem;
    font-weight: 900;
    cursor: pointer;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
    opacity: 0.65;
}
.opt-card:hover .opt-letter { opacity: 1; transform: scale(1.05); }
.opt-card.correct .opt-letter { opacity: 1; }

.opt-input {
    flex: 1;
    border: none;
    background: transparent;
    outline: none;
    font-size: 0.9rem;
    font-family: inherit;
    color: #334155;
    font-weight: 500;
    min-width: 0;
}
.opt-input::placeholder { color: #c8d0dc; }

.opt-remove {
    opacity: 0;
    display: flex; align-items: center; justify-content: center;
    width: 20px; height: 20px;
    background: none;
    border: none;
    border-radius: 5px;
    color: #94a3b8;
    cursor: pointer;
    transition: all 0.15s;
    flex-shrink: 0;
}
.opt-card:hover .opt-remove { opacity: 1; }
.opt-remove:hover { background: #fee2e2; color: #ef4444; }
.opt-remove:disabled { opacity: 0 !important; cursor: not-allowed; }

.add-opt-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 16px;
    border: 2px dashed #c7d2fe;
    border-radius: 13px;
    background: none;
    cursor: pointer;
    font-size: 0.82rem;
    font-weight: 600;
    color: #94a3b8;
    transition: all 0.15s;
    font-family: inherit;
}
.add-opt-card:hover { border-color: #6366f1; color: #6366f1; background: #eef2ff; }

/* Explanation */
.explanation-details {
    background: white;
    border-radius: 12px;
    border: 1.5px solid #e8ecf0;
    overflow: hidden;
}
.explanation-details summary {
    display: flex;
    align-items: center;
    gap: 7px;
    padding: 11px 14px;
    cursor: pointer;
    font-size: 0.82rem;
    font-weight: 600;
    color: #64748b;
    list-style: none;
    user-select: none;
    transition: background 0.15s;
}
.explanation-details summary:hover { background: #f8fafc; }
.explanation-details[open] summary { color: #6366f1; border-bottom: 1px solid #e8ecf0; }
.explanation-input {
    display: block;
    width: 100%;
    padding: 12px 14px;
    border: none;
    outline: none;
    font-size: 0.88rem;
    font-family: inherit;
    color: #334155;
    background: transparent;
    box-sizing: border-box;
}
.explanation-input::placeholder { color: #c8d0dc; }

/* Empty state */
.empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    height: 300px;
    color: #94a3b8;
}
.empty-icon { font-size: 3rem; opacity: 0.5; }
.empty-state p { font-size: 0.9rem; }
.btn-primary {
    padding: 10px 20px;
    background: #6366f1;
    color: white;
    border: none;
    border-radius: 10px;
    font-size: 0.88rem;
    font-weight: 700;
    cursor: pointer;
    font-family: inherit;
    transition: all 0.15s;
}
.btn-primary:hover { background: #4f46e5; transform: translateY(-1px); }

/* ════ Overlay / Modal ════ */
.overlay {
    position: fixed;
    inset: 0;
    background: rgba(15,23,42,0.6);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 500;
    padding: 20px;
}

.ai-modal {
    background: white;
    border-radius: 20px;
    padding: 28px;
    width: 100%;
    max-width: 580px;
    max-height: 88vh;
    overflow-y: auto;
    box-shadow: 0 24px 60px rgba(0,0,0,0.2);
}

.ai-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 24px;
}
.ai-head-left { display: flex; align-items: center; gap: 14px; }
.ai-icon {
    width: 42px; height: 42px;
    background: linear-gradient(135deg, #7c3aed, #a855f7);
    border-radius: 12px;
    display: flex; align-items: center; justify-content: center;
    font-size: 1.3rem;
    flex-shrink: 0;
}
.ai-head h2 { font-size: 1.1rem; font-weight: 800; color: #0f172a; }
.ai-head p { font-size: 0.8rem; color: #94a3b8; margin-top: 2px; }
.modal-close {
    display: flex; align-items: center; justify-content: center;
    width: 32px; height: 32px;
    background: #f1f5f9;
    border: none;
    border-radius: 8px;
    color: #64748b;
    cursor: pointer;
    transition: all 0.15s;
}
.modal-close:hover { background: #fee2e2; color: #ef4444; }

.ai-form { display: flex; flex-direction: column; gap: 16px; }
.modal-field { display: flex; flex-direction: column; gap: 5px; }
.modal-field span { font-size: 0.78rem; font-weight: 700; color: #64748b; }
.modal-field input,
.modal-field select {
    padding: 9px 12px;
    border: 1.5px solid #e2e8f0;
    border-radius: 9px;
    font-size: 0.88rem;
    font-family: inherit;
    color: #0f172a;
    outline: none;
    transition: border-color 0.15s;
    background: #f8fafc;
}
.modal-field input:focus,
.modal-field select:focus { border-color: #7c3aed; background: white; }
.ai-grid3 { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 10px; }
.modal-label { font-size: 0.78rem; font-weight: 700; color: #64748b; margin-bottom: 8px; }
.type-chips { display: flex; flex-wrap: wrap; gap: 7px; }
.type-chip {
    padding: 6px 14px;
    border: 2px solid #e2e8f0;
    border-radius: 99px;
    background: white;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s;
    font-family: inherit;
    color: #475569;
}
.type-chip:hover { border-color: #7c3aed; color: #7c3aed; }
.type-chip.active { border-color: #7c3aed; background: #f5f3ff; color: #7c3aed; }
.ai-err {
    background: #fef2f2;
    color: #ef4444;
    padding: 10px 14px;
    border-radius: 9px;
    font-size: 0.84rem;
    font-weight: 500;
}
.btn-ai-gen {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 14px;
    background: linear-gradient(135deg, #7c3aed, #a855f7);
    color: white;
    border: none;
    border-radius: 12px;
    font-size: 0.95rem;
    font-weight: 800;
    cursor: pointer;
    transition: all 0.15s;
    box-shadow: 0 4px 14px rgba(124,58,237,0.35);
    font-family: inherit;
}
.btn-ai-gen:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 20px rgba(124,58,237,0.45); }
.btn-ai-gen:disabled { opacity: 0.5; cursor: not-allowed; }

.ai-results { display: flex; flex-direction: column; gap: 12px; }
.ai-results-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 8px;
    padding: 10px 14px;
    background: #f8fafc;
    border-radius: 10px;
    font-size: 0.85rem;
    color: #475569;
}
.sel-actions { display: flex; gap: 6px; }
.sel-actions button {
    padding: 4px 10px;
    border: 1.5px solid #e2e8f0;
    border-radius: 6px;
    background: white;
    font-size: 0.76rem;
    font-weight: 600;
    cursor: pointer;
    font-family: inherit;
    transition: all 0.15s;
}
.sel-actions button:hover { border-color: #7c3aed; color: #7c3aed; }
.ai-q-list { display: flex; flex-direction: column; gap: 6px; max-height: 340px; overflow-y: auto; }
.ai-q-card {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 12px 14px;
    border: 2px solid #e2e8f0;
    border-radius: 12px;
    background: white;
    cursor: pointer;
    text-align: left;
    width: 100%;
    transition: all 0.15s;
}
.ai-q-card:hover { border-color: #c4b5fd; }
.ai-q-card.selected { border-color: #7c3aed; background: #f5f3ff; }
.ai-q-check {
    width: 22px; height: 22px;
    border: 2px solid #e2e8f0;
    border-radius: 6px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
    margin-top: 1px;
}
.ai-q-check.checked { background: #7c3aed; border-color: #7c3aed; }
.ai-q-content { flex: 1; }
.ai-q-text { font-size: 0.87rem; font-weight: 600; color: #0f172a; margin-bottom: 5px; line-height: 1.4; }
.ai-q-opts { display: flex; flex-wrap: wrap; gap: 4px; }
.ai-q-opts span { font-size: 0.73rem; color: #64748b; padding: 2px 8px; background: #f1f5f9; border-radius: 5px; }
.ai-q-opts .ai-correct { background: #dcfce7; color: #15803d; font-weight: 600; }
.ai-result-footer { display: flex; gap: 10px; }
.btn-back {
    padding: 11px 18px;
    background: #f1f5f9;
    border: none;
    border-radius: 10px;
    cursor: pointer;
    font-size: 0.87rem;
    font-weight: 600;
    font-family: inherit;
    color: #475569;
    transition: background 0.15s;
}
.btn-back:hover { background: #e2e8f0; }
.btn-ai-add {
    flex: 1;
    padding: 12px;
    background: linear-gradient(135deg, #7c3aed, #a855f7);
    color: white;
    border: none;
    border-radius: 10px;
    font-size: 0.9rem;
    font-weight: 800;
    cursor: pointer;
    font-family: inherit;
    box-shadow: 0 3px 12px rgba(124,58,237,0.3);
    transition: all 0.15s;
}
.btn-ai-add:hover:not(:disabled) { transform: translateY(-1px); }
.btn-ai-add:disabled { opacity: 0.45; cursor: not-allowed; }
</style>
