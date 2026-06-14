<script lang="ts">
    import { sfx } from '$lib/boardFx';
    export let content: any;
    const prompts: string[] = content?.prompts ?? [];

    let opened: boolean[] = prompts.map(() => false);
    let active: number | null = null;

    function open(i: number) {
        active = i;
        opened[i] = true; opened = opened;
        sfx('open');
    }
    function close() { active = null; }
    function reset() { opened = prompts.map(() => false); active = null; }

    $: allOpen = prompts.length > 0 && opened.every(Boolean);
</script>

<div class="wrap">
    <div class="bar">
        <span>Ochilgan: {opened.filter(Boolean).length} / {prompts.length}</span>
        {#if opened.some(Boolean)}<button class="reset" on:click={reset}>↻ Boshidan</button>{/if}
    </div>

    <div class="grid">
        {#each prompts as _, i}
            <button class="box" class:opened={opened[i]} on:click={() => open(i)}>
                <span class="box-emoji">{opened[i] ? '📭' : '🎁'}</span>
                <span class="box-num">{i + 1}</span>
            </button>
        {/each}
    </div>

    {#if allOpen}<p class="done-msg">🎉 Barcha qutilar ochildi!</p>{/if}
</div>

{#if active !== null}
    <div class="overlay" on:click={close} role="presentation">
        <div class="modal" on:click|stopPropagation role="presentation">
            <div class="m-num">Quti {active + 1}</div>
            <p class="m-text">{prompts[active]}</p>
            <button class="big-btn" on:click={close}>Yopish</button>
        </div>
    </div>
{/if}

<style>
    .wrap { flex: 1; display: flex; flex-direction: column; width: 100%; padding: 20px clamp(16px,3vw,50px); box-sizing: border-box; }
    .bar { display: flex; justify-content: space-between; align-items: center; font-size: 1.1rem; font-weight: 700; color: #94a3b8; margin-bottom: 18px; }
    .reset { background: rgba(255,255,255,0.1); border: none; color: #fff; padding: 7px 16px; border-radius: 10px; cursor: pointer; font-weight: 700; }
    .grid { flex: 1; display: grid; grid-template-columns: repeat(auto-fill, minmax(130px, 1fr)); gap: 16px; align-content: start; }
    .box {
        aspect-ratio: 1; border: none; border-radius: 20px; cursor: pointer;
        background: linear-gradient(135deg,#4f46e5,#7c3aed); color: #fff;
        display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 6px;
        box-shadow: 0 8px 0 rgba(0,0,0,0.3); transition: transform 0.15s;
    }
    .box:hover { transform: translateY(-4px) scale(1.02); }
    .box:active { transform: translateY(4px); }
    .box.opened { background: rgba(255,255,255,0.06); box-shadow: 0 8px 0 rgba(0,0,0,0.15); opacity: 0.6; }
    .box-emoji { font-size: clamp(2.4rem, 6vw, 4rem); }
    .box-num { font-size: 1.4rem; font-weight: 800; }
    .done-msg { text-align: center; font-size: 1.6rem; font-weight: 800; color: #22c55e; margin-top: 16px; }

    .overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.7); display: flex; align-items: center; justify-content: center; z-index: 50; animation: fade 0.2s; }
    @keyframes fade { from{opacity:0;} to{opacity:1;} }
    .modal {
        background: linear-gradient(135deg,#1e293b,#334155); border-radius: 28px; padding: 40px;
        text-align: center; max-width: 80vw; display: flex; flex-direction: column; gap: 22px; align-items: center;
        animation: zoom 0.3s cubic-bezier(0.34,1.56,0.64,1); box-shadow: 0 20px 60px rgba(0,0,0,0.6);
    }
    @keyframes zoom { from{transform:scale(0.6);opacity:0;} to{transform:scale(1);opacity:1;} }
    .m-num { font-size: 1rem; font-weight: 700; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.1em; }
    .m-text { font-size: clamp(1.6rem, 4vw, 3rem); font-weight: 800; margin: 0; color: #f1f5f9; }
    .big-btn { padding: 14px 40px; border: none; border-radius: 14px; cursor: pointer; background: linear-gradient(135deg,#6366f1,#8b5cf6); color:#fff; font-size: 1.2rem; font-weight: 800; }
</style>
