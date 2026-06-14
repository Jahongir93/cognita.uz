<script lang="ts">
    export let content: any;
    const prompts: string[] = content?.prompts ?? [];

    const palette = ['#e21b3c', '#1368ce', '#d89e00', '#26890c', '#9333ea', '#0891b2', '#db2777', '#f59e0b'];
    const colors = prompts.map((_, i) => palette[i % palette.length]);

    const n = prompts.length;
    const seg = n ? 360 / n : 360;

    // conic-gradient segmentlari
    $: gradient = `conic-gradient(${prompts.map((_, i) => `${colors[i]} ${i*seg}deg ${(i+1)*seg}deg`).join(', ')})`;

    let rotation = 0;
    let spinning = false;
    let result: string | null = null;

    function spin() {
        if (spinning || n === 0) return;
        result = null;
        spinning = true;
        const target = Math.floor(Math.random() * n);
        const desiredMod = ((360 - (target * seg + seg / 2)) % 360 + 360) % 360;
        const currentMod = ((rotation % 360) + 360) % 360;
        let delta = desiredMod - currentMod;
        if (delta < 0) delta += 360;
        rotation += 5 * 360 + delta;
        setTimeout(() => { result = prompts[target]; spinning = false; }, 4200);
    }
</script>

<div class="wrap">
    <div class="wheel-area">
        <div class="pointer">▼</div>
        <div class="wheel" style="transform: rotate({rotation}deg); background: {gradient}">
            {#each prompts as _, i}
                <span class="seg-num" style="transform: rotate({i*seg + seg/2}deg) translateY(-42%)">{i + 1}</span>
            {/each}
        </div>
        <button class="hub" on:click={spin} disabled={spinning}>{spinning ? '...' : 'AYLANTIR'}</button>
    </div>

    {#if result}
        <div class="result">
            <p class="r-text">{result}</p>
            <button class="big-btn" on:click={spin}>🎡 Yana aylantirish</button>
        </div>
    {:else}
        <p class="tip">G'ildirakni aylantirish uchun markazni bosing</p>
    {/if}
</div>

<style>
    .wrap { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 24px; width: 100%; padding: 20px; box-sizing: border-box; }
    .wheel-area { position: relative; width: min(58vh, 460px); height: min(58vh, 460px); }
    .pointer { position: absolute; top: -6px; left: 50%; transform: translateX(-50%); font-size: 2.4rem; color: #fbbf24; z-index: 4; filter: drop-shadow(0 2px 3px rgba(0,0,0,0.5)); }
    .wheel {
        width: 100%; height: 100%; border-radius: 50%;
        transition: transform 4.1s cubic-bezier(0.17, 0.67, 0.12, 0.99);
        box-shadow: 0 0 0 10px rgba(255,255,255,0.12), 0 20px 50px rgba(0,0,0,0.5);
        position: relative;
    }
    .seg-num { position: absolute; top: 50%; left: 50%; transform-origin: 0 0; color: #fff; font-weight: 800; font-size: 1.3rem; text-shadow: 0 1px 2px rgba(0,0,0,0.5); }
    .hub {
        position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%);
        width: 28%; height: 28%; border-radius: 50%; border: 5px solid #fff; cursor: pointer;
        background: #1e293b; color: #fff; font-weight: 800; font-size: clamp(0.8rem,1.6vw,1.1rem); z-index: 3;
        box-shadow: 0 6px 16px rgba(0,0,0,0.5);
    }
    .hub:disabled { cursor: default; }
    .hub:hover:not(:disabled) { background: #334155; }
    .result { text-align: center; display: flex; flex-direction: column; gap: 16px; align-items: center; }
    .r-text {
        font-size: clamp(1.6rem, 4vw, 3rem); font-weight: 800; margin: 0; max-width: 80vw;
        background: rgba(255,255,255,0.08); padding: 20px 36px; border-radius: 20px; animation: pop 0.4s;
    }
    @keyframes pop { 0%{transform:scale(0.7);opacity:0;} 100%{transform:scale(1);opacity:1;} }
    .tip { color: #64748b; }
    .big-btn { padding: 14px 36px; border: none; border-radius: 14px; cursor: pointer; background: linear-gradient(135deg,#6366f1,#8b5cf6); color:#fff; font-size: 1.2rem; font-weight: 800; box-shadow: 0 8px 24px rgba(99,102,241,0.5); }
</style>
