<script lang="ts">
    import { sfx } from '$lib/boardFx';
    export let content: any;
    const pairs: { left: string; right: string }[] = content?.pairs ?? [];

    let idx = 0;
    let flipped = false;
    function toggle() { flipped = !flipped; sfx('flip'); }

    $: p = pairs[idx];
    function next() { if (idx < pairs.length - 1) { idx++; flipped = false; } }
    function prev() { if (idx > 0) { idx--; flipped = false; } }
</script>

<div class="wrap">
    <div class="bar">Kartochka {idx + 1} / {pairs.length}</div>

    {#if p}
        <div class="stage">
            <button class="flip-btn prev" on:click={prev} disabled={idx === 0}>‹</button>

            <button class="fcard" class:flipped on:click={toggle}>
                <div class="fc-inner">
                    <div class="fc-face front"><span class="hint">Savol</span>{p.left}</div>
                    <div class="fc-face back"><span class="hint">Javob</span>{p.right}</div>
                </div>
            </button>

            <button class="flip-btn next" on:click={next} disabled={idx === pairs.length - 1}>›</button>
        </div>
        <p class="tap">Aylantirish uchun kartani bosing</p>
    {/if}
</div>

<style>
    .wrap { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; width: 100%; gap: 20px; padding: 24px; box-sizing: border-box; }
    .bar { font-size: 1.2rem; font-weight: 700; color: #94a3b8; }
    .stage { display: flex; align-items: center; gap: clamp(12px, 3vw, 40px); width: 100%; justify-content: center; }
    .fcard { background: none; border: none; cursor: pointer; perspective: 1400px; width: min(70vw, 640px); height: min(48vh, 380px); padding: 0; }
    .fc-inner { position: relative; width: 100%; height: 100%; transition: transform 0.55s; transform-style: preserve-3d; }
    .fcard.flipped .fc-inner { transform: rotateY(180deg); }
    .fc-face {
        position: absolute; inset: 0; backface-visibility: hidden; border-radius: 28px;
        display: flex; align-items: center; justify-content: center; flex-direction: column; gap: 14px;
        font-size: clamp(1.8rem, 5vw, 3.4rem); font-weight: 800; padding: 30px; text-align: center; box-sizing: border-box;
        box-shadow: 0 16px 40px rgba(0,0,0,0.4);
    }
    .front { background: linear-gradient(135deg,#1e293b,#334155); color: #f1f5f9; }
    .back { background: linear-gradient(135deg,#4f46e5,#7c3aed); color: #fff; transform: rotateY(180deg); }
    .hint { font-size: 0.85rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; opacity: 0.6; }
    .flip-btn {
        width: 60px; height: 60px; border-radius: 50%; border: none; flex-shrink: 0;
        background: rgba(255,255,255,0.1); color: #fff; font-size: 2rem; cursor: pointer; line-height: 1;
    }
    .flip-btn:hover:not(:disabled) { background: rgba(255,255,255,0.2); }
    .flip-btn:disabled { opacity: 0.25; cursor: not-allowed; }
    .tap { color: #64748b; font-size: 0.9rem; }
</style>
