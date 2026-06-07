<script lang="ts">
    import type { StudentAnswerOption } from '$lib/api/types';

    export let options: StudentAnswerOption[];
    export let selectedId: string | null = null;
    export let correctIds: string[] = [];
    export let disabled: boolean = false;
    export let showResult: boolean = false;

    // Kahoot-style colors
    const colors = ['#e21b3c', '#1368ce', '#d89e00', '#26890c'];
    const shapes = ['▲', '◆', '●', '■'];

    function handleSelect(id: string) {
        if (!disabled && !selectedId) {
            selectedId = id;
        }
    }

    function getState(opt: StudentAnswerOption): 'default' | 'selected' | 'correct' | 'wrong' | 'dimmed' {
        if (!showResult) {
            return selectedId === opt.id ? 'selected' : 'default';
        }
        if (correctIds.includes(opt.id)) return 'correct';
        if (selectedId === opt.id && !correctIds.includes(opt.id)) return 'wrong';
        return 'dimmed';
    }
</script>

<div class="options-grid" class:two-col={options.length === 2}>
    {#each options as opt, i}
        {@const state = getState(opt)}
        <button
            class="option {state}"
            style="--color: {colors[i % colors.length]}"
            on:click={() => handleSelect(opt.id)}
            disabled={disabled || !!selectedId}
            aria-pressed={selectedId === opt.id}
        >
            <span class="shape">{shapes[i % shapes.length]}</span>
            {#if opt.media_url}
                <img src={opt.media_url} alt={opt.option_text} class="opt-image" />
            {/if}
            <span class="text">{opt.option_text}</span>
            {#if showResult && state === 'correct'}
                <span class="badge">✓</span>
            {:else if showResult && state === 'wrong'}
                <span class="badge">✗</span>
            {/if}
        </button>
    {/each}
</div>

<style>
    .options-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
        width: 100%;
    }
    .two-col {
        grid-template-columns: 1fr 1fr;
    }

    .option {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 16px 20px;
        background: var(--color);
        color: white;
        border: none;
        border-radius: 8px;
        font-size: 1rem;
        font-weight: 600;
        cursor: pointer;
        transition: transform 0.1s, opacity 0.2s;
        min-height: 64px;
        text-align: left;
        position: relative;
        border-bottom: 4px solid color-mix(in srgb, var(--color) 70%, black);
    }
    .option:not(:disabled):hover {
        transform: translateY(-2px);
    }
    .option:not(:disabled):active {
        transform: translateY(2px);
        border-bottom-width: 2px;
    }
    .option.selected {
        outline: 3px solid white;
        outline-offset: 2px;
    }
    .option.correct {
        background: #22c55e !important;
        border-bottom-color: #15803d;
    }
    .option.wrong {
        background: #6b7280 !important;
        opacity: 0.7;
        border-bottom-color: #374151;
    }
    .option.dimmed {
        opacity: 0.45;
    }
    .option:disabled {
        cursor: default;
    }
    .shape {
        font-size: 1.1rem;
        flex-shrink: 0;
    }
    .text {
        flex: 1;
    }
    .badge {
        position: absolute;
        right: 12px;
        font-size: 1.3rem;
    }
    .opt-image {
        width: 48px;
        height: 48px;
        object-fit: cover;
        border-radius: 4px;
    }

    @media (max-width: 480px) {
        .options-grid {
            grid-template-columns: 1fr;
        }
        .option {
            min-height: 52px;
        }
    }
</style>
