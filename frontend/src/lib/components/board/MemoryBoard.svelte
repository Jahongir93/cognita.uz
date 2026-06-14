<script lang="ts">
    import { onMount } from 'svelte';
    import { sfx, confetti } from '$lib/boardFx';
    export let content: any;
    const pairs: { left: string; right: string }[] = content?.pairs ?? [];

    type Card = { uid: number; pid: number; text: string; flipped: boolean; matched: boolean };
    let cards: Card[] = [];
    let first: number | null = null;
    let busy = false;
    let matchedCount = 0;
    let moves = 0;

    function shuffle<T>(arr: T[]): T[] {
        const a = [...arr];
        for (let i = a.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [a[i], a[j]] = [a[j], a[i]];
        }
        return a;
    }

    function build() {
        const raw: Card[] = [];
        let uid = 0;
        pairs.forEach((p, pid) => {
            raw.push({ uid: uid++, pid, text: p.left, flipped: false, matched: false });
            raw.push({ uid: uid++, pid, text: p.right, flipped: false, matched: false });
        });
        cards = shuffle(raw);
        first = null; busy = false; matchedCount = 0; moves = 0;
    }
    onMount(build);

    function flip(i: number) {
        if (busy || cards[i].flipped || cards[i].matched) return;
        cards[i].flipped = true; cards = cards; sfx('flip');
        if (first === null) { first = i; return; }
        moves++;
        const a = cards[first], b = cards[i];
        if (a.pid === b.pid) {
            cards[first].matched = true; cards[i].matched = true; cards = cards;
            matchedCount++; first = null;
            sfx('correct');
            if (matchedCount === pairs.length) { sfx('win'); confetti(); }
        } else {
            busy = true;
            const f = first; first = null;
            setTimeout(() => { cards[f].flipped = false; cards[i].flipped = false; cards = cards; busy = false; }, 900);
        }
    }
    $: won = pairs.length > 0 && matchedCount === pairs.length;
    $: cols = Math.min(6, Math.ceil(Math.sqrt(cards.length || 1)));
</script>

<div class="wrap">
    <div class="bar"><span>Topildi: {matchedCount} / {pairs.length}</span><span>Urinishlar: {moves}</span></div>

    {#if won}
        <div class="win">
            <div class="win-art">
                <img src="/img/board/star-burst.png" alt="" class="win-burst" />
                <img src="/img/board/cogni-trophy.png" alt="" class="win-cogni" />
            </div>
            <h1>Barakalla! Hammasi topildi</h1>
            <button class="big-btn" on:click={build}>↻ Qaytadan</button>
        </div>
    {:else}
        <div class="grid" style="grid-template-columns: repeat({cols}, 1fr)">
            {#each cards as c, i (c.uid)}
                <button class="card" class:flipped={c.flipped || c.matched} class:matched={c.matched} on:click={() => flip(i)}>
                    <span class="face back">?</span>
                    <span class="face front">{c.text}</span>
                </button>
            {/each}
        </div>
    {/if}
</div>

<style>
    .wrap { flex: 1; display: flex; flex-direction: column; width: 100%; padding: 20px clamp(16px,3vw,50px); box-sizing: border-box; }
    .bar { display: flex; justify-content: space-between; font-size: 1.1rem; font-weight: 700; color: #94a3b8; margin-bottom: 16px; }
    .grid { flex: 1; display: grid; gap: clamp(8px,1.5vw,16px); align-content: stretch; }
    .card {
        position: relative; border: none; background: none; cursor: pointer;
        min-height: 90px; border-radius: 16px; perspective: 800px; padding: 0;
    }
    .face {
        position: absolute; inset: 0; display: flex; align-items: center; justify-content: center;
        border-radius: 16px; backface-visibility: hidden; transition: transform 0.4s;
        font-weight: 800; padding: 10px; text-align: center; box-sizing: border-box;
    }
    .back {
        background: linear-gradient(135deg,#4f46e5,#7c3aed); color: #fff; font-size: 2rem;
        box-shadow: 0 6px 0 rgba(0,0,0,0.25);
    }
    .front {
        background: #f8fafc; color: #1e293b; font-size: clamp(0.9rem,1.8vw,1.4rem);
        transform: rotateY(180deg); box-shadow: 0 6px 0 rgba(0,0,0,0.2);
    }
    .card.flipped .back { transform: rotateY(180deg); }
    .card.flipped .front { transform: rotateY(360deg); }
    .card.matched .front { background: #dcfce7; color: #16a34a; box-shadow: 0 0 0 4px #22c55e; animation: pop 0.4s; }
    @keyframes pop { 0%{transform:rotateY(360deg) scale(1);} 50%{transform:rotateY(360deg) scale(1.08);} 100%{transform:rotateY(360deg) scale(1);} }
    .win { margin: auto; text-align: center; }
    .win-art { position: relative; width: 260px; height: 260px; margin: 0 auto 6px; }
    .win-burst { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: contain; animation: spinSlow 12s linear infinite; opacity: 0.9; }
    .win-cogni { position: absolute; inset: 0; margin: auto; width: 70%; height: 70%; object-fit: contain; animation: bob 1.8s ease-in-out infinite; }
    @keyframes bob { 0%,100%{transform:translateY(0);} 50%{transform:translateY(-12px);} }
    @keyframes spinSlow { to { transform: rotate(360deg); } }
    .win h1 { font-size: 2.4rem; margin: 10px 0 24px; }
    .big-btn { padding: 16px 48px; border: none; border-radius: 16px; cursor: pointer; background: linear-gradient(135deg,#6366f1,#8b5cf6); color:#fff; font-size: 1.4rem; font-weight: 800; box-shadow: 0 8px 24px rgba(99,102,241,0.5); }
</style>
