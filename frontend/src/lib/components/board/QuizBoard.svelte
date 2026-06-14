<script lang="ts">
    import { sfx, confetti } from '$lib/boardFx';
    export let content: any;
    const questions: { text: string; options: string[]; correct: number }[] = content?.questions ?? [];

    let idx = 0;
    let selected: number | null = null;
    let revealed = false;
    let score = 0;

    $: q = questions[idx];
    const colors = ['#e21b3c', '#1368ce', '#d89e00', '#26890c', '#9333ea', '#0891b2'];
    const shapes = ['▲', '◆', '●', '■', '★', '⬢'];

    function pick(i: number) {
        if (revealed) return;
        selected = i;
        revealed = true;
        if (i === q.correct) { score++; sfx('correct'); } else { sfx('wrong'); }
    }
    function next() {
        if (idx < questions.length - 1) { idx++; selected = null; revealed = false; }
        else { finished = true; sfx('win'); confetti(); }
    }
    let finished = false;
    function restart() { idx = 0; selected = null; revealed = false; score = 0; finished = false; }
</script>

{#if finished}
    <div class="done">
        <div class="done-art">
            <img src="/img/board/star-burst.png" alt="" class="done-burst" />
            <img src="/img/board/cogni-trophy.png" alt="" class="done-cogni" />
        </div>
        <h1>Yakunlandi!</h1>
        <p class="done-score">{score} / {questions.length} to'g'ri</p>
        <button class="big-btn" on:click={restart}>↻ Qaytadan</button>
    </div>
{:else if q}
    <div class="quiz">
        <div class="qbar">
            <span class="qprog">Savol {idx + 1} / {questions.length}</span>
            <span class="qscore">⭐ {score}</span>
        </div>
        <h1 class="qtext">{q.text}</h1>
        <div class="grid">
            {#each q.options as opt, i}
                <button
                    class="opt"
                    style="--c:{colors[i % colors.length]}"
                    class:correct={revealed && i === q.correct}
                    class:wrong={revealed && i === selected && i !== q.correct}
                    class:fade={revealed && i !== q.correct && i !== selected}
                    on:click={() => pick(i)}
                >
                    <span class="sh">{shapes[i % shapes.length]}</span>
                    <span class="lbl">{opt}</span>
                    {#if revealed && i === q.correct}<span class="mark">✓</span>{/if}
                    {#if revealed && i === selected && i !== q.correct}<span class="mark">✕</span>{/if}
                </button>
            {/each}
        </div>
        <div class="footer">
            {#if revealed}
                <button class="big-btn" on:click={next}>{idx < questions.length - 1 ? 'Keyingi →' : 'Yakunlash'}</button>
            {/if}
        </div>

        {#if revealed}
            <img class="reaction" alt="" src="/img/board/{selected === q.correct ? 'cogni-happy' : 'cogni-oops'}.png" />
        {/if}
    </div>
{/if}

<style>
    .quiz { flex: 1; display: flex; flex-direction: column; width: 100%; padding: 24px clamp(16px, 4vw, 60px); box-sizing: border-box; position: relative; overflow: hidden; }
    .reaction {
        position: absolute; right: 18px; bottom: 18px;
        width: clamp(90px, 14vw, 160px); height: auto;
        filter: drop-shadow(0 8px 20px rgba(0,0,0,0.4));
        animation: reactIn 0.4s cubic-bezier(0.34,1.56,0.64,1) both;
        pointer-events: none;
    }
    @keyframes reactIn { from { opacity: 0; transform: translateY(40px) scale(0.7); } to { opacity: 1; transform: none; } }
    .qbar { display: flex; justify-content: space-between; font-size: 1.1rem; font-weight: 700; color: #94a3b8; }
    .qscore { color: #fbbf24; }
    .qtext { font-size: clamp(1.6rem, 4vw, 3rem); font-weight: 800; text-align: center; margin: clamp(12px, 4vh, 40px) 0; line-height: 1.25; }
    .grid { flex: 1; display: grid; grid-template-columns: 1fr 1fr; gap: clamp(12px, 2vw, 22px); }
    .opt {
        display: flex; align-items: center; gap: 18px; padding: clamp(16px, 3vw, 34px);
        border: none; border-radius: 20px; background: var(--c); color: #fff; cursor: pointer;
        font-size: clamp(1.2rem, 3vw, 2.2rem); font-weight: 800; text-align: left;
        box-shadow: 0 8px 0 rgba(0,0,0,0.25), 0 12px 26px rgba(0,0,0,0.3);
        transition: transform 0.15s, filter 0.2s, opacity 0.3s; position: relative; overflow: hidden;
    }
    .opt:hover { filter: brightness(1.1); transform: translateY(-3px); }
    .opt:active { transform: translateY(4px); }
    .sh { font-size: 1.4em; }
    .lbl { flex: 1; }
    .mark { font-size: 1.6em; margin-left: auto; }
    .opt.correct { animation: pop 0.4s ease; box-shadow: 0 0 0 6px rgba(255,255,255,0.7), 0 12px 30px rgba(34,197,94,0.5); }
    .opt.wrong { filter: grayscale(0.3) brightness(0.85); }
    .opt.fade { opacity: 0.35; }
    @keyframes pop { 0%{transform:scale(1);} 45%{transform:scale(1.05);} 100%{transform:scale(1);} }
    .footer { display: flex; justify-content: center; min-height: 70px; align-items: center; margin-top: 18px; }
    .big-btn {
        padding: 16px 48px; border: none; border-radius: 16px; cursor: pointer;
        background: linear-gradient(135deg,#6366f1,#8b5cf6); color: #fff;
        font-size: 1.4rem; font-weight: 800; box-shadow: 0 8px 24px rgba(99,102,241,0.5);
        transition: transform 0.15s;
    }
    .big-btn:hover { transform: translateY(-2px); }
    .done { margin: auto; text-align: center; }
    .done-art { position: relative; width: 280px; height: 280px; margin: 0 auto 8px; }
    .done-burst { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: contain; animation: spinSlow 12s linear infinite; opacity: 0.9; }
    .done-cogni { position: absolute; inset: 0; margin: auto; width: 70%; height: 70%; object-fit: contain; animation: bob 1.8s ease-in-out infinite; }
    @keyframes bob { 0%,100%{transform:translateY(0);} 50%{transform:translateY(-12px);} }
    @keyframes spinSlow { to { transform: rotate(360deg); } }
    .done h1 { font-size: 2.6rem; margin: 8px 0; }
    .done-score { font-size: 1.6rem; color: #fbbf24; font-weight: 800; margin-bottom: 24px; }
</style>
