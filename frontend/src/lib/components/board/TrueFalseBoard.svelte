<script lang="ts">
    import { sfx, confetti } from '$lib/boardFx';
    export let content: any;
    const statements: { text: string; answer: boolean }[] = content?.statements ?? [];

    let idx = 0;
    let picked: boolean | null = null;
    let revealed = false;
    let score = 0;
    let finished = false;

    $: s = statements[idx];

    function pick(val: boolean) {
        if (revealed) return;
        picked = val;
        revealed = true;
        if (val === s.answer) { score++; sfx('correct'); } else { sfx('wrong'); }
    }
    function next() {
        if (idx < statements.length - 1) { idx++; picked = null; revealed = false; }
        else { finished = true; sfx('win'); confetti(); }
    }
    function restart() { idx = 0; picked = null; revealed = false; score = 0; finished = false; }
</script>

{#if finished}
    <div class="done">
        <div class="done-emoji">🎉</div>
        <h1>Yakunlandi!</h1>
        <p class="done-score">{score} / {statements.length} to'g'ri</p>
        <button class="big-btn" on:click={restart}>↻ Qaytadan</button>
    </div>
{:else if s}
    <div class="tf">
        <div class="qbar"><span>{idx + 1} / {statements.length}</span><span class="sc">⭐ {score}</span></div>
        <div class="card" class:ok={revealed && s.answer} class:no={revealed && !s.answer}>
            <p class="stmt">{s.text}</p>
            {#if revealed}
                <div class="verdict">{s.answer ? '✅ To\'g\'ri' : '❌ Noto\'g\'ri'}</div>
            {/if}
        </div>
        <div class="btns">
            <button class="tfbtn yes" class:dim={revealed && !s.answer} class:hit={revealed && picked && !s.answer} on:click={() => pick(true)}>✅ To'g'ri</button>
            <button class="tfbtn no"  class:dim={revealed && s.answer}  class:hit={revealed && picked===false && s.answer} on:click={() => pick(false)}>❌ Noto'g'ri</button>
        </div>
        <div class="footer">
            {#if revealed}<button class="big-btn" on:click={next}>{idx < statements.length - 1 ? 'Keyingi →' : 'Yakunlash'}</button>{/if}
        </div>
    </div>
{/if}

<style>
    .tf { flex: 1; display: flex; flex-direction: column; width: 100%; padding: 24px clamp(16px,5vw,80px); box-sizing: border-box; }
    .qbar { display: flex; justify-content: space-between; font-size: 1.1rem; font-weight: 700; color: #94a3b8; }
    .sc { color: #fbbf24; }
    .card {
        flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 24px;
        margin: 24px 0; border-radius: 28px; background: rgba(255,255,255,0.06);
        border: 2px solid rgba(255,255,255,0.1); padding: 40px; text-align: center;
        transition: background 0.3s, border-color 0.3s;
    }
    .card.ok { background: rgba(34,197,94,0.15); border-color: #22c55e; }
    .card.no { background: rgba(239,68,68,0.15); border-color: #ef4444; }
    .stmt { font-size: clamp(1.8rem, 4.5vw, 3.4rem); font-weight: 800; line-height: 1.3; margin: 0; }
    .verdict { font-size: 1.8rem; font-weight: 800; animation: pop 0.4s ease; }
    @keyframes pop { 0%{transform:scale(0.6);opacity:0;} 100%{transform:scale(1);opacity:1;} }
    .btns { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
    .tfbtn {
        padding: clamp(18px,3vw,32px); border: none; border-radius: 20px; cursor: pointer;
        font-size: clamp(1.3rem,3vw,2.2rem); font-weight: 800; color: #fff;
        box-shadow: 0 8px 0 rgba(0,0,0,0.25); transition: transform 0.15s, opacity 0.3s;
    }
    .tfbtn.yes { background: #16a34a; }
    .tfbtn.no  { background: #dc2626; }
    .tfbtn:active { transform: translateY(4px); }
    .tfbtn.dim { opacity: 0.3; }
    .tfbtn.hit { animation: shake 0.4s; }
    @keyframes shake { 0%,100%{transform:translateX(0);} 25%{transform:translateX(-8px);} 75%{transform:translateX(8px);} }
    .footer { display: flex; justify-content: center; min-height: 70px; align-items: center; margin-top: 18px; }
    .big-btn { padding: 16px 48px; border: none; border-radius: 16px; cursor: pointer; background: linear-gradient(135deg,#6366f1,#8b5cf6); color:#fff; font-size: 1.4rem; font-weight: 800; box-shadow: 0 8px 24px rgba(99,102,241,0.5); }
    .done { margin: auto; text-align: center; }
    .done-emoji { font-size: 5rem; animation: bob 1.6s ease-in-out infinite; }
    @keyframes bob { 0%,100%{transform:translateY(0);} 50%{transform:translateY(-12px);} }
    .done h1 { font-size: 2.6rem; margin: 8px 0; }
    .done-score { font-size: 1.6rem; color: #fbbf24; font-weight: 800; margin-bottom: 24px; }
</style>
