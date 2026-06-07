<script lang="ts">
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { quizzes as quizzesApi } from '$lib/api/client';
    import QuizEditor from '$lib/components/QuizEditor.svelte';
    import type { QuizMeta, EditQuestion, EditOption } from '$lib/components/QuizEditor.svelte';

    $: quizId = $page.params.id ?? '';

    let loading = true;
    let loadError = '';
    let saving = false;

    let initialTitle = '';
    let initialDescription = '';
    let initialSubject = '';
    let initialGrade = '';
    let initialPublic = false;
    let initialTags: string[] = [];
    let initialQuestions: EditQuestion[] = [];

    function uid() { return Math.random().toString(36).slice(2, 9); }

    onMount(async () => {
        try {
            const quiz = await quizzesApi.get(quizId);
            initialTitle       = quiz.title;
            initialDescription = quiz.description ?? '';
            initialSubject     = quiz.subject ?? '';
            initialGrade       = quiz.grade_level ?? '';
            initialPublic      = quiz.is_public;
            initialTags        = quiz.tags ?? [];
            initialQuestions   = (quiz.questions ?? []).map(q => ({
                _id: uid(),
                type: q.type,
                text: q.question_text,
                time_limit: q.time_limit,
                points: q.points,
                explanation: q.explanation ?? '',
                options: (q.options ?? []).map(o => ({
                    _id: uid(),
                    text: o.option_text,
                    is_correct: o.is_correct,
                }))
            }));
        } catch (e: any) {
            loadError = e.message ?? 'Quiz yuklanmadi';
        } finally {
            loading = false;
        }
    });

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
            await quizzesApi.update(quizId, payload);
            goto('/dashboard/quizzes');
        } catch (err: any) {
            saving = false;
            throw err;
        }
    }
</script>

<svelte:head><title>Quiz tahrirlash — Cognita.uz</title></svelte:head>

{#if loading}
    <div class="loading-wrap">
        <div class="spinner dark"></div>
        <span>Yuklanmoqda...</span>
    </div>
{:else if loadError}
    <div class="err-wrap">
        <p>{loadError}</p>
        <a href="/dashboard/quizzes">← Quizlarga qaytish</a>
    </div>
{:else}
    <QuizEditor
        pageTitle="Quiz tahrirlash"
        {saving}
        {initialTitle}
        {initialDescription}
        {initialSubject}
        initialGrade={initialGrade}
        {initialPublic}
        {initialTags}
        {initialQuestions}
        on:save={handleSave}
    />
{/if}

<style>
    .loading-wrap {
        display: flex; align-items: center; justify-content: center;
        gap: 12px; height: 300px; color: var(--text3); font-size: 0.95rem;
    }
    .err-wrap {
        display: flex; flex-direction: column; align-items: center;
        gap: 14px; padding: 60px 20px; text-align: center;
        color: var(--danger);
    }
    .err-wrap a { color: var(--primary); text-decoration: none; font-weight: 600; }
</style>
