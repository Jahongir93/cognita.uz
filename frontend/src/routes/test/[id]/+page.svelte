<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { openTests } from '$lib/api/client';
    import { authStore } from '$lib/stores/auth';

    const id = $page.params.id ?? '';

    const THEMES: Record<string, { icon: string; g1: string; g2: string; label: string }> = {
        fun:         { icon: '🎯', g1: '#f59e0b', g2: '#ef4444', label: 'Qiziqarli test' },
        subjects:    { icon: '📚', g1: '#3b82f6', g2: '#6366f1', label: 'Fan testi' },
        iq:          { icon: '🧠', g1: '#8b5cf6', g2: '#6366f1', label: 'IQ test' },
        attestation: { icon: '📋', g1: '#22c55e', g2: '#0ea5e9', label: 'Attestatsiya' },
        psychology:  { icon: '🧘', g1: '#ec4899', g2: '#8b5cf6', label: 'Psixologik test' },
    };

    let phase: 'loading' | 'lobby' | 'play' | 'done' | 'error' = 'loading';
    let test: { id: string; title: string; category: string; scored: boolean; questions: { text: string; options: string[] }[] } | null = null;
    let nickname = '';
    let answers: number[] = [];
    let idx = 0;
    let selected: number | null = null;
    let startTime = 0;
    let result: { score: number; total: number; percent: number; scored: boolean; rank: number } | null = null;
    let leaderboard: { rank: number; nickname: string; score: number; total: number; time_ms: number }[] = [];

    $: theme = test ? (THEMES[test.category] ?? THEMES.fun) : THEMES.fun;
    $: q = test?.questions[idx];

    const optColors = ['#e21b3c', '#1368ce', '#d89e00', '#26890c', '#9333ea', '#0891b2'];
    const optShapes = ['▲', '◆', '●', '■', '★', '⬢'];

    onMount(async () => {
        if ($authStore.user?.full_name) nickname = $authStore.user.full_name;
        try {
            test = await openTests.take(id);
            answers = new Array(test.questions.length).fill(-1);
            phase = 'lobby';
        } catch {
            phase = 'error';
        }
    });

    function start() {
        if (!nickname.trim()) nickname = 'Anonim';
        startTime = Date.now();
        idx = 0; selected = null;
        phase = 'play';
    }

    function pick(i: number) {
        if (selected !== null) return;
        selected = i;
        answers[idx] = i;
        setTimeout(next, 650);
    }
    async function next() {
        if (!test) return;
        if (idx < test.questions.length - 1) {
            idx++; selected = null;
        } else {
            await finish();
        }
    }
    async function finish() {
        const timeMs = Date.now() - startTime;
        try {
            result = await openTests.submit(id, { nickname: nickname.trim() || 'Anonim', answers, time_ms: timeMs });
            if (result.scored) {
                try { leaderboard = await openTests.leaderboard(id); } catch {}
            }
        } catch {}
        phase = 'done';
    }

    function fmtTime(ms: number) {
        const s = Math.round(ms / 1000);
        return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
    }
</script>

<svelte:head><title>{test?.title ?? 'Test'} — Cognita.uz</title></svelte:head>

<div class="screen" style="--g1:{theme.g1};--g2:{theme.g2}">
    <div class="bg" aria-hidden="true"><div class="orb o1"></div><div class="orb o2"></div></div>

    <a href="/" class="logo"><img src="/logowhite.png" alt="Cognita.uz" /></a>

    {#if phase === 'loading'}
        <div class="center"><div class="spin"></div></div>

    {:else if phase === 'error'}
        <div class="card center-card">
            <div class="big-emoji">😕</div>
            <h2>Test topilmadi</h2>
            <a href="/" class="btn-ghost">Bosh sahifa</a>
        </div>

    <!-- ══ NEW LOBBY ══ -->
    {:else if phase === 'lobby' && test}
        <div class="lobby">
            <div class="badge">{theme.icon} {theme.label}</div>
            <h1 class="l-title">{test.title}</h1>
            <div class="l-meta">
                <span class="chip">❓ {test.questions.length} savol</span>
                <span class="chip">{test.scored ? '🏆 Reytingli' : '🧘 Reytingsiz'}</span>
            </div>
            <div class="name-field">
                <label>Ismingiz</label>
                <input bind:value={nickname} placeholder="Ismingizni kiriting" maxlength="40"
                       on:keydown={(e) => e.key === 'Enter' && start()} />
            </div>
            <button class="start-btn" on:click={start}>
                <span class="sb-icon">🚀</span> Boshlash
            </button>
            <p class="hint">Istalgan vaqt, istalgan kishi yecha oladi</p>
        </div>

    <!-- ══ PLAY ══ -->
    {:else if phase === 'play' && q}
        <div class="play">
            <div class="p-top">
                <span class="p-prog">{idx + 1} / {test?.questions.length}</span>
                <div class="p-bar"><div class="p-fill" style="width:{((idx) / (test?.questions.length ?? 1)) * 100}%"></div></div>
            </div>
            <h2 class="p-q">{q.text}</h2>
            <div class="p-opts">
                {#each q.options as opt, i}
                    <button class="p-opt" style="--c:{optColors[i % optColors.length]}"
                            class:sel={selected === i} class:dim={selected !== null && selected !== i}
                            disabled={selected !== null} on:click={() => pick(i)}>
                        <span class="p-sh">{optShapes[i % optShapes.length]}</span>
                        <span class="p-lbl">{opt}</span>
                    </button>
                {/each}
            </div>
        </div>

    <!-- ══ DONE ══ -->
    {:else if phase === 'done'}
        <div class="done">
            {#if result?.scored}
                <div class="score-ring">
                    <div class="sr-pct">{result.percent}%</div>
                    <div class="sr-sub">{result.score}/{result.total} to'g'ri</div>
                </div>
                <div class="rank-badge">🏅 {result.rank}-o'rin</div>
                <h2 class="d-title">Ajoyib!</h2>

                {#if leaderboard.length}
                    <div class="lb">
                        <div class="lb-head">🏆 Reyting (Top {Math.min(leaderboard.length, 50)})</div>
                        {#each leaderboard.slice(0, 12) as e}
                            <div class="lb-row" class:me={e.nickname === nickname.trim()}>
                                <span class="lb-rank">{e.rank === 1 ? '🥇' : e.rank === 2 ? '🥈' : e.rank === 3 ? '🥉' : `#${e.rank}`}</span>
                                <span class="lb-name">{e.nickname}</span>
                                <span class="lb-score">{e.score}/{e.total}</span>
                                <span class="lb-time">{fmtTime(e.time_ms)}</span>
                            </div>
                        {/each}
                    </div>
                {/if}
            {:else}
                <div class="big-emoji">🧘</div>
                <h2 class="d-title">Test yakunlandi</h2>
                <p class="d-sub">Javoblaringiz qabul qilindi. Bu test reytingsiz.</p>
            {/if}
            <div class="d-actions">
                <a href={`/tests/${test?.category}`} class="btn-ghost">← Boshqa testlar</a>
                <button class="btn-main" on:click={() => { phase = 'lobby'; result = null; selected = null; answers = new Array(test?.questions.length ?? 0).fill(-1); }}>↻ Qayta yechish</button>
            </div>
        </div>
    {/if}
</div>

<style>
    :global(body) { margin: 0; }
    .screen {
        min-height: 100dvh; position: relative; overflow: hidden;
        display: flex; align-items: center; justify-content: center; padding: 24px 16px;
        font-family: 'Segoe UI', system-ui, sans-serif; color: #fff;
        background: linear-gradient(160deg, #0f172a 0%, #1e1b4b 55%, #2e1065 100%);
    }
    .bg { position: absolute; inset: 0; pointer-events: none; }
    .orb { position: absolute; border-radius: 50%; filter: blur(90px); opacity: 0.5; }
    .o1 { width: 460px; height: 460px; background: var(--g1); top: -160px; right: -120px; animation: drift 14s ease-in-out infinite alternate; }
    .o2 { width: 380px; height: 380px; background: var(--g2); bottom: -140px; left: -100px; animation: drift 18s ease-in-out infinite alternate-reverse; }
    @keyframes drift { from { transform: translate(0,0) scale(1); } to { transform: translate(40px,30px) scale(1.15); } }

    .logo { position: absolute; top: 20px; left: 24px; z-index: 5; }
    .logo img { height: 30px; width: auto; opacity: 0.9; }

    .center { position: relative; z-index: 1; }
    .spin { width: 48px; height: 48px; border: 4px solid rgba(255,255,255,0.2); border-top-color: #fff; border-radius: 50%; animation: rot 0.8s linear infinite; }
    @keyframes rot { to { transform: rotate(360deg); } }

    /* ── Lobby ── */
    .lobby {
        position: relative; z-index: 1; width: 100%; max-width: 460px;
        background: rgba(255,255,255,0.07); backdrop-filter: blur(14px);
        border: 1px solid rgba(255,255,255,0.14); border-radius: 28px;
        padding: 36px 30px; text-align: center;
        box-shadow: 0 30px 80px rgba(0,0,0,0.45);
        animation: pop 0.5s cubic-bezier(0.34,1.4,0.64,1) both;
    }
    @keyframes pop { from { opacity: 0; transform: translateY(24px) scale(0.96); } to { opacity: 1; transform: none; } }
    .badge {
        display: inline-block; padding: 7px 18px; border-radius: 99px; font-weight: 800; font-size: 0.85rem;
        background: linear-gradient(135deg, var(--g1), var(--g2)); box-shadow: 0 8px 20px rgba(0,0,0,0.3); margin-bottom: 18px;
    }
    .l-title { font-size: clamp(1.5rem, 4vw, 2.1rem); font-weight: 900; margin: 0 0 16px; line-height: 1.2; }
    .l-meta { display: flex; gap: 10px; justify-content: center; margin-bottom: 24px; flex-wrap: wrap; }
    .chip { background: rgba(255,255,255,0.12); padding: 6px 14px; border-radius: 99px; font-size: 0.82rem; font-weight: 600; }
    .name-field { text-align: left; margin-bottom: 20px; }
    .name-field label { display: block; font-size: 0.78rem; font-weight: 700; color: rgba(255,255,255,0.7); margin-bottom: 6px; text-transform: uppercase; letter-spacing: 0.06em; }
    .name-field input {
        width: 100%; padding: 13px 16px; border-radius: 12px; border: 2px solid rgba(255,255,255,0.2);
        background: rgba(255,255,255,0.1); color: #fff; font-size: 1rem; outline: none; box-sizing: border-box;
    }
    .name-field input:focus { border-color: #fff; background: rgba(255,255,255,0.16); }
    .name-field input::placeholder { color: rgba(255,255,255,0.45); }
    .start-btn {
        width: 100%; padding: 16px; border: none; border-radius: 16px; cursor: pointer;
        background: linear-gradient(135deg, var(--g1), var(--g2)); color: #fff; font-size: 1.15rem; font-weight: 800;
        display: flex; align-items: center; justify-content: center; gap: 10px;
        box-shadow: 0 12px 30px rgba(0,0,0,0.35); transition: transform 0.18s;
        animation: glow 2.4s ease-in-out infinite;
    }
    @keyframes glow { 0%,100% { box-shadow: 0 12px 30px rgba(0,0,0,0.35); } 50% { box-shadow: 0 12px 40px var(--g1); } }
    .start-btn:hover { transform: translateY(-3px); }
    .sb-icon { font-size: 1.3rem; }
    .hint { color: rgba(255,255,255,0.5); font-size: 0.8rem; margin: 16px 0 0; }

    /* ── Play ── */
    .play { position: relative; z-index: 1; width: 100%; max-width: 760px; }
    .p-top { display: flex; align-items: center; gap: 14px; margin-bottom: 26px; }
    .p-prog { font-weight: 800; font-size: 1rem; flex-shrink: 0; }
    .p-bar { flex: 1; height: 8px; background: rgba(255,255,255,0.12); border-radius: 99px; overflow: hidden; }
    .p-fill { height: 100%; background: linear-gradient(90deg, var(--g1), var(--g2)); border-radius: 99px; transition: width 0.4s ease; }
    .p-q { font-size: clamp(1.4rem, 4vw, 2.4rem); font-weight: 800; text-align: center; margin: 0 0 30px; line-height: 1.3; }
    .p-opts { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
    .p-opt {
        display: flex; align-items: center; gap: 14px; padding: 20px 22px; border: none; border-radius: 18px;
        background: var(--c); color: #fff; cursor: pointer; font-size: clamp(1rem, 2.4vw, 1.4rem); font-weight: 800;
        text-align: left; box-shadow: 0 8px 0 rgba(0,0,0,0.25); transition: transform 0.14s, opacity 0.25s;
    }
    .p-opt:hover:not(:disabled) { transform: translateY(-3px); filter: brightness(1.08); }
    .p-opt:active:not(:disabled) { transform: translateY(3px); }
    .p-opt.sel { outline: 4px solid #fff; outline-offset: -4px; }
    .p-opt.dim { opacity: 0.4; }
    .p-sh { font-size: 1.4em; }
    @media (max-width: 560px) { .p-opts { grid-template-columns: 1fr; } }

    /* ── Done ── */
    .done {
        position: relative; z-index: 1; width: 100%; max-width: 480px; text-align: center;
        animation: pop 0.5s cubic-bezier(0.34,1.4,0.64,1) both;
    }
    .score-ring {
        width: 180px; height: 180px; border-radius: 50%; margin: 0 auto 16px;
        display: flex; flex-direction: column; align-items: center; justify-content: center;
        background: conic-gradient(var(--g1) calc(var(--p,0) * 1%), rgba(255,255,255,0.1) 0);
        background: linear-gradient(135deg, var(--g1), var(--g2));
        box-shadow: 0 16px 40px rgba(0,0,0,0.4);
    }
    .sr-pct { font-size: 3rem; font-weight: 900; }
    .sr-sub { font-size: 0.9rem; opacity: 0.85; }
    .rank-badge { display: inline-block; background: rgba(255,255,255,0.14); padding: 6px 18px; border-radius: 99px; font-weight: 800; margin-bottom: 8px; }
    .d-title { font-size: 1.8rem; font-weight: 900; margin: 4px 0 18px; }
    .d-sub { color: rgba(255,255,255,0.7); margin: 0 0 18px; }
    .big-emoji { font-size: 4.5rem; }

    .lb { background: rgba(255,255,255,0.07); border: 1px solid rgba(255,255,255,0.12); border-radius: 18px; padding: 14px; margin-bottom: 20px; text-align: left; }
    .lb-head { font-weight: 800; margin-bottom: 10px; text-align: center; }
    .lb-row { display: flex; align-items: center; gap: 10px; padding: 8px 10px; border-radius: 10px; font-size: 0.88rem; }
    .lb-row.me { background: rgba(255,255,255,0.14); font-weight: 800; }
    .lb-rank { width: 34px; flex-shrink: 0; }
    .lb-name { flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .lb-score { font-weight: 800; color: #fbbf24; }
    .lb-time { font-size: 0.78rem; opacity: 0.6; min-width: 44px; text-align: right; }

    .d-actions { display: flex; gap: 10px; justify-content: center; flex-wrap: wrap; }
    .card.center-card, .center-card { position: relative; z-index: 1; background: rgba(255,255,255,0.08); border-radius: 24px; padding: 36px; text-align: center; }
    .btn-main { padding: 13px 26px; border: none; border-radius: 13px; background: linear-gradient(135deg, var(--g1), var(--g2)); color: #fff; font-weight: 800; cursor: pointer; font-size: 0.95rem; }
    .btn-ghost { padding: 13px 22px; border-radius: 13px; border: 2px solid rgba(255,255,255,0.25); background: rgba(255,255,255,0.08); color: #fff; font-weight: 700; text-decoration: none; display: inline-flex; align-items: center; }
    .btn-ghost:hover { background: rgba(255,255,255,0.16); }
</style>
