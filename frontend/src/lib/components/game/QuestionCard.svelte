<script lang="ts">
    import type { QuestionPayload, StudentAnswerOption } from '$lib/api/types';
    import Timer from './Timer.svelte';
    import AnswerOptions from './AnswerOptions.svelte';

    export let question: QuestionPayload;
    export let secondsLeft: number;
    export let selectedAnswer: string | null = null;
    export let correctIds: string[] = [];
    export let showResult: boolean = false;
    export let onAnswer: (optionId: string) => void = () => {};

    let textInput = '';

    function handleSelect(id: string) {
        if (!selectedAnswer) {
            onAnswer(id);
        }
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === 'Enter' && textInput) {
            onAnswer(textInput);
        }
    }

    function submitTextAnswer() {
        if (textInput) onAnswer(textInput);
    }
</script>

<div class="question-card">
    <!-- Header: progress + timer -->
    <div class="header">
        <span class="progress">
            Savol {question.question_index + 1} / {question.total_questions}
        </span>
        <Timer seconds={secondsLeft} maxSeconds={question.time_limit} />
        <span class="points">+{question.points} ball</span>
    </div>

    <!-- Question text -->
    <div class="question-body">
        {#if question.media_url}
            {#if question.media_type === 'video'}
                <video src={question.media_url} controls class="media" />
            {:else if question.media_type === 'audio'}
                <audio src={question.media_url} controls class="media" />
            {:else}
                <img src={question.media_url} alt="Savol rasmi" class="media" />
            {/if}
        {/if}
        <p class="text">{question.question_text}</p>
    </div>

    <!-- Answer options -->
    {#if question.type === 'multiple_choice' || question.type === 'true_false' || question.type === 'image_choice'}
        <AnswerOptions
            options={question.options}
            selectedId={selectedAnswer}
            {correctIds}
            {showResult}
            disabled={!!selectedAnswer}
            on:select={(e) => handleSelect(e.detail)}
        />
    {:else if question.type === 'short_answer' || question.type === 'fill_blank'}
        <div class="text-answer">
            <input
                type="text"
                placeholder="Javobingizni yozing..."
                disabled={!!selectedAnswer}
                bind:value={textInput}
                on:keydown={handleKeydown}
                class="text-input"
            />
            <button
                class="submit-btn"
                disabled={!!selectedAnswer}
                on:click={submitTextAnswer}
            >
                Yuborish
            </button>
        </div>
    {:else if question.type === 'poll'}
        <AnswerOptions
            options={question.options}
            selectedId={selectedAnswer}
            correctIds={[]}
            showResult={false}
            disabled={!!selectedAnswer}
            on:select={(e) => handleSelect(e.detail)}
        />
    {/if}

    {#if selectedAnswer && !showResult}
        <div class="waiting-feedback">
            <span class="spinner" aria-hidden="true"></span>
            Boshqalarni kutmoqda...
        </div>
    {/if}
</div>

<style>
    .question-card {
        display: flex;
        flex-direction: column;
        gap: 20px;
        width: 100%;
        max-width: 800px;
        margin: 0 auto;
        padding: 16px;
    }
    .header {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }
    .progress {
        font-size: 0.85rem;
        color: #64748b;
    }
    .points {
        font-size: 0.9rem;
        font-weight: 600;
        color: #f59e0b;
        background: #fef3c7;
        padding: 4px 10px;
        border-radius: 999px;
    }
    .question-body {
        text-align: center;
    }
    .media {
        max-width: 100%;
        max-height: 220px;
        border-radius: 8px;
        margin-bottom: 12px;
        object-fit: contain;
    }
    .text {
        font-size: clamp(1rem, 3vw, 1.5rem);
        font-weight: 600;
        color: #0f172a;
        line-height: 1.4;
    }
    .text-answer {
        display: flex;
        gap: 8px;
    }
    .text-input {
        flex: 1;
        padding: 12px 16px;
        font-size: 1rem;
        border: 2px solid #e2e8f0;
        border-radius: 8px;
        outline: none;
    }
    .text-input:focus {
        border-color: #3b82f6;
    }
    .submit-btn {
        padding: 12px 20px;
        background: #3b82f6;
        color: white;
        border: none;
        border-radius: 8px;
        font-size: 1rem;
        font-weight: 600;
        cursor: pointer;
    }
    .waiting-feedback {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        color: #64748b;
        font-size: 0.9rem;
    }
    .spinner {
        width: 16px;
        height: 16px;
        border: 2px solid #e2e8f0;
        border-top-color: #3b82f6;
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
    }
    @keyframes spin {
        to { transform: rotate(360deg); }
    }
</style>
