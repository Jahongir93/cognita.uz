<script lang="ts">
    import { onMount } from 'svelte';
    export let content: any;
    const groups: { name: string; items: string[] }[] = content?.groups ?? [];

    type Tile = { text: string; group: number; placed: number | null };
    let tiles: Tile[] = [];
    let selected: number | null = null;
    let wrongGroup: number | null = null;

    function shuffle<T>(a: T[]): T[] { const r=[...a]; for(let i=r.length-1;i>0;i--){const j=Math.floor(Math.random()*(i+1));[r[i],r[j]]=[r[j],r[i]];} return r; }

    function build() {
        const raw: Tile[] = [];
        groups.forEach((g, gi) => g.items.forEach(it => raw.push({ text: it, group: gi, placed: null })));
        tiles = shuffle(raw);
        selected = null;
    }
    onMount(build);

    $: pool = tiles.map((t, i) => ({ t, i })).filter(x => x.t.placed === null);
    $: done = tiles.length > 0 && pool.length === 0;

    function selectTile(i: number) { selected = selected === i ? null : i; }
    function dropTo(gi: number) {
        if (selected === null) return;
        const t = tiles[selected];
        if (t.group === gi) {
            tiles[selected].placed = gi; tiles = tiles; selected = null;
        } else {
            wrongGroup = gi;
            setTimeout(() => wrongGroup = null, 500);
        }
    }
</script>

<div class="wrap">
    {#if done}
        <div class="win"><div class="we">🎉</div><h1>Hammasi to'g'ri ajratildi!</h1><button class="big-btn" on:click={build}>↻ Qaytadan</button></div>
    {:else}
        <div class="pool">
            {#each pool as x (x.i)}
                <button class="tile" class:sel={selected === x.i} on:click={() => selectTile(x.i)}>{x.t.text}</button>
            {/each}
            {#if pool.length === 0}<span class="pool-empty">Tugadi!</span>{/if}
        </div>
        <p class="hint">Elementni tanlang, so'ng guruhni bosing</p>
        <div class="groups">
            {#each groups as g, gi}
                <button class="group" class:wrong={wrongGroup === gi} on:click={() => dropTo(gi)}>
                    <div class="g-name">{g.name}</div>
                    <div class="g-items">
                        {#each tiles.filter(t => t.placed === gi) as t}
                            <span class="placed">{t.text}</span>
                        {/each}
                    </div>
                </button>
            {/each}
        </div>
    {/if}
</div>

<style>
    .wrap { flex: 1; display: flex; flex-direction: column; width: 100%; padding: 20px clamp(16px,3vw,50px); box-sizing: border-box; gap: 14px; }
    .pool { display: flex; flex-wrap: wrap; gap: 10px; justify-content: center; min-height: 60px; align-content: center; }
    .tile {
        padding: 14px 22px; border: none; border-radius: 14px; cursor: pointer;
        background: #f8fafc; color: #1e293b; font-size: clamp(1rem,2vw,1.5rem); font-weight: 800;
        box-shadow: 0 5px 0 rgba(0,0,0,0.25); transition: transform 0.15s;
    }
    .tile:hover { transform: translateY(-2px); }
    .tile.sel { background: #fbbf24; box-shadow: 0 0 0 4px #f59e0b, 0 5px 0 rgba(0,0,0,0.25); transform: translateY(-2px); }
    .pool-empty { color: #22c55e; font-weight: 800; font-size: 1.4rem; }
    .hint { text-align: center; color: #64748b; font-size: 0.9rem; margin: 0; }
    .groups { flex: 1; display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 14px; }
    .group {
        display: flex; flex-direction: column; gap: 10px; padding: 16px; border-radius: 18px; cursor: pointer;
        background: rgba(255,255,255,0.06); border: 2px dashed rgba(255,255,255,0.18); color: inherit; text-align: center;
        transition: background 0.2s, border-color 0.2s, transform 0.1s;
    }
    .group:hover { background: rgba(255,255,255,0.1); border-color: #6366f1; }
    .group.wrong { animation: shake 0.45s; border-color: #ef4444; }
    @keyframes shake { 0%,100%{transform:translateX(0);} 25%{transform:translateX(-10px);} 75%{transform:translateX(10px);} }
    .g-name { font-size: clamp(1.1rem,2vw,1.6rem); font-weight: 800; color: #f1f5f9; }
    .g-items { display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; }
    .placed { background: #dcfce7; color: #16a34a; padding: 8px 14px; border-radius: 10px; font-weight: 700; font-size: clamp(0.85rem,1.6vw,1.2rem); animation: drop 0.3s ease; }
    @keyframes drop { from{transform:scale(0.5);opacity:0;} to{transform:scale(1);opacity:1;} }
    .win { margin: auto; text-align: center; }
    .we { font-size: 5rem; animation: bob 1.6s ease-in-out infinite; }
    @keyframes bob { 0%,100%{transform:translateY(0);} 50%{transform:translateY(-12px);} }
    .win h1 { font-size: 2.4rem; margin: 10px 0 24px; }
    .big-btn { padding: 16px 48px; border: none; border-radius: 16px; cursor: pointer; background: linear-gradient(135deg,#6366f1,#8b5cf6); color:#fff; font-size: 1.4rem; font-weight: 800; box-shadow: 0 8px 24px rgba(99,102,241,0.5); }
</style>
