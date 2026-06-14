<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { activitiesApi, type BoardActivity } from '$lib/api/client';
    import { getModule } from '$lib/data/activityModules';

    import QuizBoard from '$lib/components/board/QuizBoard.svelte';
    import TrueFalseBoard from '$lib/components/board/TrueFalseBoard.svelte';
    import MemoryBoard from '$lib/components/board/MemoryBoard.svelte';
    import FlashcardsBoard from '$lib/components/board/FlashcardsBoard.svelte';
    import SortBoard from '$lib/components/board/SortBoard.svelte';
    import AnagramBoard from '$lib/components/board/AnagramBoard.svelte';
    import WheelBoard from '$lib/components/board/WheelBoard.svelte';
    import BoxBoard from '$lib/components/board/BoxBoard.svelte';

    const id = $page.params.id ?? '';
    let activity: BoardActivity | undefined;
    let player = '';
    let notFound = false;

    onMount(async () => {
        try {
            activity = await activitiesApi.get(id);
            player = getModule(activity.type)?.player ?? '';
        } catch {
            notFound = true;
        }
    });

    function exit() { window.close(); if (!window.closed) history.back(); }
</script>

<svelte:head>
    <title>{activity?.title ?? 'Doska'} — Cognita</title>
</svelte:head>

<div class="board">
    <header class="bhead">
        <span class="bicon">{getModule(activity?.type ?? '')?.icon ?? '🧩'}</span>
        <span class="btitle">{activity?.title ?? ''}</span>
        <button class="bexit" on:click={exit} title="Chiqish">✕</button>
    </header>

    <main class="bmain">
        {#if notFound}
            <div class="msg">
                <p class="msg-icon">😕</p>
                <p>Topshiriq topilmadi.</p>
                <p class="msg-sub">Topshiriq o'chirilgan bo'lishi yoki sizga tegishli bo'lmasligi mumkin. Tizimga kirganingizni tekshiring.</p>
            </div>
        {:else if activity}
            {#if player === 'quiz'}
                <QuizBoard content={activity.content} />
            {:else if player === 'truefalse'}
                <TrueFalseBoard content={activity.content} />
            {:else if player === 'memory'}
                <MemoryBoard content={activity.content} />
            {:else if player === 'flashcards'}
                <FlashcardsBoard content={activity.content} />
            {:else if player === 'sort'}
                <SortBoard content={activity.content} />
            {:else if player === 'anagram'}
                <AnagramBoard content={activity.content} />
            {:else if player === 'wheel'}
                <WheelBoard content={activity.content} />
            {:else if player === 'box'}
                <BoxBoard content={activity.content} />
            {:else}
                <div class="msg"><p>Bu o'yin turi hali tayyor emas.</p></div>
            {/if}
        {/if}
    </main>
</div>

<style>
    :global(body) { margin: 0; }
    .board {
        min-height: 100dvh;
        display: flex;
        flex-direction: column;
        background:
            linear-gradient(rgba(11,18,32,0.85), rgba(11,18,32,0.92)),
            url('/img/board/board-bg-default.png') center / cover no-repeat fixed,
            #0b1220;
        font-family: 'Segoe UI', system-ui, sans-serif;
        color: #f1f5f9;
    }
    .bhead {
        display: flex; align-items: center; gap: 12px;
        padding: 12px 20px; background: rgba(255,255,255,0.05);
        border-bottom: 1px solid rgba(255,255,255,0.08);
    }
    .bicon { font-size: 1.5rem; }
    .btitle { font-weight: 800; font-size: 1.15rem; flex: 1; }
    .bexit {
        width: 42px; height: 42px; border-radius: 10px; border: none;
        background: rgba(255,255,255,0.1); color: #fff; font-size: 1.1rem; cursor: pointer;
    }
    .bexit:hover { background: rgba(239,68,68,0.8); }
    .bmain { flex: 1; display: flex; min-height: 0; }
    .msg { margin: auto; text-align: center; color: #cbd5e1; }
    .msg-icon { font-size: 3rem; margin: 0; }
    .msg-sub { font-size: 0.85rem; color: #64748b; max-width: 380px; }
</style>
