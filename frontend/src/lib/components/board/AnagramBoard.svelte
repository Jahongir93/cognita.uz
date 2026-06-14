<script lang="ts">
    export let content: any;
    const words: { word: string; hint?: string }[] = content?.words ?? [];

    let idx = 0;
    type Tile = { ch: string; id: number; used: boolean };
    let tiles: Tile[] = [];
    let built: { ch: string; id: number }[] = [];
    let solved = false;
    let finished = false;

    function shuffle<T>(a: T[]): T[] { const r=[...a]; for(let i=r.length-1;i>0;i--){const j=Math.floor(Math.random()*(i+1));[r[i],r[j]]=[r[j],r[i]];} return r; }

    $: target = (words[idx]?.word ?? '').toUpperCase();

    function load() {
        const chars = target.split('').map((ch, id) => ({ ch, id, used: false }));
        // qayta aralashtirish, agar tasodifan to'g'ri chiqsa
        let s = shuffle(chars);
        if (s.map(t => t.ch).join('') === target && target.length > 1) s = shuffle(chars);
        tiles = s; built = []; solved = false;
    }
    $: if (target) load();

    function pick(t: Tile) {
        if (t.used || solved) return;
        t.used = true; tiles = tiles;
        built = [...built, { ch: t.ch, id: t.id }];
        check();
    }
    function unpick(i: number) {
        if (solved) return;
        const b = built[i];
        const t = tiles.find(x => x.id === b.id); if (t) t.used = false;
        tiles = tiles;
        built = built.filter((_, x) => x !== i);
    }
    function check() {
        if (built.length === target.length) {
            solved = built.map(b => b.ch).join('') === target;
        }
    }
    function next() {
        if (idx < words.length - 1) idx++;
        else finished = true;
    }
    function restart() { idx = 0; finished = false; load(); }
</script>

{#if finished}
    <div class="win"><div class="we">🎉</div><h1>Barcha so'zlar tiklandi!</h1><button class="big-btn" on:click={restart}>↻ Qaytadan</button></div>
{:else if words[idx]}
    <div class="wrap">
        <div class="bar">So'z {idx + 1} / {words.length}</div>
        {#if words[idx].hint}<p class="hint">💡 {words[idx].hint}</p>{/if}

        <div class="slots" class:solved>
            {#each built as b, i}
                <button class="slot filled" on:click={() => unpick(i)}>{b.ch}</button>
            {/each}
            {#each Array(Math.max(0, target.length - built.length)) as _}
                <span class="slot empty"></span>
            {/each}
        </div>

        {#if solved}
            <div class="ok">✅ To'g'ri!</div>
            <button class="big-btn" on:click={next}>{idx < words.length - 1 ? 'Keyingi so\'z →' : 'Yakunlash'}</button>
        {:else}
            <div class="tiles">
                {#each tiles as t (t.id)}
                    <button class="tile" class:used={t.used} on:click={() => pick(t)} disabled={t.used}>{t.ch}</button>
                {/each}
            </div>
        {/if}
    </div>
{/if}

<style>
    .wrap { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 26px; width: 100%; padding: 24px; box-sizing: border-box; }
    .bar { font-size: 1.2rem; font-weight: 700; color: #94a3b8; }
    .hint { font-size: clamp(1.1rem,2.5vw,1.6rem); color: #fbbf24; margin: 0; text-align: center; }
    .slots { display: flex; flex-wrap: wrap; gap: 10px; justify-content: center; }
    .slot {
        width: clamp(48px, 8vw, 84px); height: clamp(56px, 9vw, 96px); border-radius: 14px;
        display: flex; align-items: center; justify-content: center;
        font-size: clamp(1.6rem, 4vw, 3rem); font-weight: 800;
    }
    .slot.empty { background: rgba(255,255,255,0.05); border: 2px dashed rgba(255,255,255,0.2); }
    .slot.filled { background: #4f46e5; color: #fff; border: none; cursor: pointer; box-shadow: 0 5px 0 rgba(0,0,0,0.3); }
    .slots.solved .slot.filled { background: #16a34a; box-shadow: 0 5px 0 rgba(0,0,0,0.3), 0 0 0 4px #22c55e; }
    .tiles { display: flex; flex-wrap: wrap; gap: 12px; justify-content: center; max-width: 900px; }
    .tile {
        width: clamp(48px, 8vw, 84px); height: clamp(56px, 9vw, 96px); border: none; border-radius: 14px; cursor: pointer;
        background: #f8fafc; color: #1e293b; font-size: clamp(1.6rem, 4vw, 3rem); font-weight: 800;
        box-shadow: 0 6px 0 rgba(0,0,0,0.25); transition: transform 0.12s;
    }
    .tile:hover:not(:disabled) { transform: translateY(-3px); }
    .tile.used { opacity: 0; pointer-events: none; }
    .ok { font-size: 2rem; font-weight: 800; color: #22c55e; animation: pop 0.4s; }
    @keyframes pop { 0%{transform:scale(0.6);} 100%{transform:scale(1);} }
    .win { margin: auto; text-align: center; }
    .we { font-size: 5rem; animation: bob 1.6s ease-in-out infinite; }
    @keyframes bob { 0%,100%{transform:translateY(0);} 50%{transform:translateY(-12px);} }
    .win h1 { font-size: 2.4rem; margin: 10px 0 24px; }
    .big-btn { padding: 16px 48px; border: none; border-radius: 16px; cursor: pointer; background: linear-gradient(135deg,#6366f1,#8b5cf6); color:#fff; font-size: 1.4rem; font-weight: 800; box-shadow: 0 8px 24px rgba(99,102,241,0.5); }
</style>
