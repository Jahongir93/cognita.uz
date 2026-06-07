<script lang="ts">
    import { page } from '$app/stores';
    import { onMount, onDestroy } from 'svelte';
    import { goto } from '$app/navigation';
    import { GameSocket } from '$lib/websocket/connection';
    import { gameStore } from '$lib/stores/game';
    import { getWebSocketURL } from '$lib/api/client';
    import type { QuestionEndPayload } from '$lib/api/types';

    const pin = $page.params.roomCode;
    let socket: GameSocket;
    let correctIds: string[] = [];
    let textInput = '';
    let resultVisible = false;

    $: phase = $gameStore.phase;
    $: question = $gameStore.currentQuestion;
    $: result = $gameStore.myResult;

    const optConfig = [
        { shape: '▲', color: '#e21b3c' },
        { shape: '◆', color: '#1368ce' },
        { shape: '●', color: '#d89e00' },
        { shape: '■', color: '#26890c' },
    ];

    $: timerRatio = question ? $gameStore.secondsLeft / question.time_limit : 1;
    $: timerColor = timerRatio > 0.5 ? '#22c55e' : timerRatio > 0.25 ? '#f59e0b' : '#ef4444';
    $: dashArr = `${timerRatio * 100} ${(1 - timerRatio) * 100}`;

    onMount(() => {
        const nickname = sessionStorage.getItem('nickname');
        const avatar   = sessionStorage.getItem('avatar');
        if (!nickname) { goto('/join'); return; }

        socket = new GameSocket(getWebSocketURL(pin, 'student'));
        socket.onStatus(s => gameStore.setConnectionStatus(s));

        socket.on('room_state',   (m) => gameStore.applyRoomState(m.payload));
        socket.on('your_info',    (m) => gameStore.setYourInfo(m.payload.participant_id, m.payload.nickname, m.payload.avatar));
        socket.on('player_joined',(m) => gameStore.playerJoined(m.payload.player));
        socket.on('player_left',  (m) => gameStore.playerLeft(m.payload.id));
        socket.on('game_started', () => {});
        socket.on('question', (m) => {
            correctIds = [];
            resultVisible = false;
            textInput = '';
            gameStore.showQuestion(m.payload);
        });
        socket.on('timer',        (m) => gameStore.setTimer(m.payload.seconds_left));
        socket.on('answer_result',(m) => gameStore.applyAnswerResult(m.payload));
        socket.on('question_end', (m: { payload: QuestionEndPayload }) => {
            correctIds = m.payload.correct_options;
            gameStore.showQuestionEnd();
            setTimeout(() => { resultVisible = true; }, 50);
        });
        socket.on('leaderboard',  (m) => gameStore.showLeaderboard(m.payload.players));
        socket.on('game_paused',  () => gameStore.pause());
        socket.on('game_resumed', () => gameStore.resume());
        socket.on('game_over',    (m) => gameStore.gameOver(m.payload));
        socket.on('error', (m) => {
            if (m.payload?.code === 'KICKED') {
                alert("Siz o'yindan chiqarildingiz");
                goto('/join');
            }
        });

        socket.connect();
        setTimeout(() => {
            socket.send('join_room', { pin, nickname, avatar: avatar ?? '🐶' });
        }, 500);
    });

    onDestroy(() => { socket?.disconnect(); gameStore.reset(); });

    function submitOption(id: string) {
        if (!question || $gameStore.myAnswer) return;
        socket.send('submit_answer', {
            question_id: question.question_id,
            option_id: id,
            response_time_ms: (question.time_limit - $gameStore.secondsLeft) * 1000
        });
        gameStore.submitAnswer(id);
    }

    function submitText() {
        if (!question || $gameStore.myAnswer || !textInput.trim()) return;
        socket.send('submit_answer', {
            question_id: question.question_id,
            text_answer: textInput.trim(),
            response_time_ms: (question.time_limit - $gameStore.secondsLeft) * 1000
        });
        gameStore.submitAnswer(textInput.trim());
    }

    function sendEmoji(e: string) { socket.send('send_emoji', { emoji: e }); }

    const emojis = ['👏', '🔥', '😂', '😮', '❤️', '👍'];

    function rankEmoji(r: number) {
        return r === 1 ? '🥇' : r === 2 ? '🥈' : r === 3 ? '🥉' : `#${r}`;
    }
</script>

<svelte:head><title>O'yin — {pin}</title></svelte:head>

<div class="screen">

    {#if $gameStore.connectionStatus !== 'connected'}
        <div class="conn-bar" class:err={$gameStore.connectionStatus === 'error'}>
            {$gameStore.connectionStatus === 'connecting' ? '⟳ Ulanmoqda...' : '⚠ Uzildi. Qayta ulanmoqda...'}
        </div>
    {/if}

    <!-- ═══ LOBBY ══════════════════════════════════════════════════════════ -->
    {#if phase === 'lobby'}
        <div class="lobby">
            <p class="quiz-name">{$gameStore.roomInfo?.quiz_title ?? ''}</p>
            <div class="av-wrap">
                <div class="ring r1"></div>
                <div class="ring r2"></div>
                <div class="av-circle">
                    {$gameStore.myAvatar || sessionStorage.getItem('avatar') || '🐶'}
                </div>
            </div>
            <p class="my-name">{$gameStore.myNickname || sessionStorage.getItem('nickname') || ''}</p>
            <p class="wait-msg">O'qituvchi boshlashini kuting...</p>
            <div class="pc-badge">👥 {$gameStore.players.filter(p => p.is_active).length} o'yinchi</div>
        </div>

    <!-- ═══ QUESTION / ANSWERED ════════════════════════════════════════════ -->
    {:else if phase === 'question' || phase === 'answered'}
        <div class="q-screen">
            <div class="q-topbar">
                <span class="q-num">{(question?.question_index ?? 0) + 1} / {$gameStore.totalQuestions}</span>
                <div class="timer-wrap">
                    <svg viewBox="0 0 36 36" class="timer-svg">
                        <circle cx="18" cy="18" r="15.9" fill="none" stroke="#e2e8f0" stroke-width="3.2"/>
                        <circle cx="18" cy="18" r="15.9" fill="none"
                            stroke={timerColor}
                            stroke-width="3.2"
                            stroke-dasharray={dashArr}
                            stroke-dashoffset="25"
                            style="transition:stroke-dasharray 0.85s linear,stroke 0.4s"
                        />
                    </svg>
                    <span class="timer-num" style="color:{timerColor}">{$gameStore.secondsLeft}</span>
                </div>
                <span class="q-pts">⭐ {question?.points ?? 0}</span>
            </div>

            <div class="q-body">
                {#if question?.media_url}
                    <img src={question.media_url} alt="" class="q-media" />
                {/if}
                <p class="q-text">{question?.question_text ?? ''}</p>
            </div>

            {#if question?.type === 'multiple_choice' || question?.type === 'true_false' || question?.type === 'image_choice'}
                <div class="opts-grid">
                    {#each (question?.options ?? []) as opt, i}
                        {@const cfg = optConfig[i % 4]}
                        <button
                            class="opt"
                            class:selected={$gameStore.myAnswer === opt.id}
                            class:dimmed={phase === 'answered' && $gameStore.myAnswer !== opt.id}
                            style="--c:{cfg.color}"
                            disabled={phase === 'answered'}
                            on:click={() => submitOption(opt.id)}
                        >
                            <span class="opt-shape">{cfg.shape}</span>
                            <span class="opt-label">{opt.option_text}</span>
                            {#if phase === 'answered' && $gameStore.myAnswer === opt.id}
                                <span class="opt-check">✓</span>
                            {/if}
                        </button>
                    {/each}
                </div>
            {:else if question?.type === 'short_answer' || question?.type === 'fill_blank'}
                <div class="text-answer">
                    <input
                        type="text"
                        placeholder="Javobingizni yozing..."
                        bind:value={textInput}
                        disabled={!!$gameStore.myAnswer}
                        on:keydown={(e) => e.key === 'Enter' && submitText()}
                        class="text-input"
                    />
                    <button class="text-submit" disabled={!!$gameStore.myAnswer} on:click={submitText}>
                        ➤
                    </button>
                </div>
            {:else if question?.type === 'poll'}
                <div class="opts-grid">
                    {#each (question?.options ?? []) as opt, i}
                        {@const cfg = optConfig[i % 4]}
                        <button
                            class="opt"
                            class:selected={$gameStore.myAnswer === opt.id}
                            class:dimmed={phase === 'answered' && $gameStore.myAnswer !== opt.id}
                            style="--c:{cfg.color}"
                            disabled={phase === 'answered'}
                            on:click={() => submitOption(opt.id)}
                        >
                            <span class="opt-shape">{cfg.shape}</span>
                            <span class="opt-label">{opt.option_text}</span>
                        </button>
                    {/each}
                </div>
            {/if}

            {#if phase === 'answered'}
                <div class="waiting-pill">
                    <span class="dot-spin"></span>
                    Boshqalarni kutmoqda...
                </div>
            {/if}

            <div class="emoji-bar">
                {#each emojis as e}
                    <button class="emoji-btn" on:click={() => sendEmoji(e)}>{e}</button>
                {/each}
            </div>
        </div>

    <!-- ═══ QUESTION END ════════════════════════════════════════════════════ -->
    {:else if phase === 'question_end'}
        <div class="end-screen">
            {#if result}
                <div class="result-card" class:correct={result.is_correct} class:wrong={!result.is_correct}
                     class:visible={resultVisible}>
                    <div class="result-icon">{result.is_correct ? '✅' : '❌'}</div>
                    {#if result.is_correct}
                        <div class="pts-earned">+{result.points_earned} ball</div>
                        {#if result.streak >= 2}
                            <div class="streak-fire">🔥 {result.streak}ta ketma-ket!</div>
                        {/if}
                    {:else}
                        <div class="wrong-label">Noto'g'ri</div>
                    {/if}
                    <div class="total-score">Jami: {result.total_score.toLocaleString()} ball</div>
                </div>
            {/if}

            <div class="correct-section">
                <p class="correct-heading">To'g'ri javob</p>
                <div class="correct-opts">
                    {#each (question?.options ?? []).filter(o => correctIds.includes(o.id)) as opt, i}
                        {@const cfg = optConfig[i % 4]}
                        <div class="correct-opt" style="background:{cfg.color}">{opt.option_text}</div>
                    {/each}
                </div>
            </div>
        </div>

    <!-- ═══ LEADERBOARD ══════════════════════════════════════════════════════ -->
    {:else if phase === 'leaderboard'}
        <div class="lb-screen">
            <h2 class="lb-heading">🏆 Natijalar</h2>
            {#if $gameStore.myParticipantId}
                {@const me = $gameStore.leaderboard.find(e => e.id === $gameStore.myParticipantId)}
                {#if me}
                    <div class="my-rank">
                        <span class="mr-rank">{rankEmoji(me.rank)}</span>
                        <span class="mr-av">{me.avatar}</span>
                        <span class="mr-name">{me.nickname} (Siz)</span>
                        <span class="mr-score">{me.score.toLocaleString()}</span>
                    </div>
                {/if}
            {/if}
            <div class="lb-list">
                {#each $gameStore.leaderboard.slice(0, 10) as e, i}
                    <div class="lb-row" class:lb-me={e.id === $gameStore.myParticipantId}
                         style="animation-delay:{i * 60}ms">
                        <span class="lb-rank">{rankEmoji(e.rank)}</span>
                        <span class="lb-av">{e.avatar}</span>
                        <span class="lb-name">{e.nickname}</span>
                        {#if e.streak >= 2}<span class="lb-streak">🔥{e.streak}</span>{/if}
                        <span class="lb-score">{e.score.toLocaleString()}</span>
                    </div>
                {/each}
            </div>
        </div>

    <!-- ═══ PAUSED ═══════════════════════════════════════════════════════════ -->
    {:else if phase === 'paused'}
        <div class="centered">
            <div class="pause-card">
                <div class="pause-icon">⏸</div>
                <h2>Pauza</h2>
                <p>O'qituvchi o'yinni to'xtatdi</p>
            </div>
        </div>

    <!-- ═══ GAME OVER ═════════════════════════════════════════════════════════ -->
    {:else if phase === 'game_over'}
        <div class="gameover-screen">
            <h1 class="go-title">🏆 O'yin tugadi!</h1>
            {#if $gameStore.myParticipantId}
                {@const me = $gameStore.leaderboard.find(e => e.id === $gameStore.myParticipantId)}
                {#if me}
                    <div class="final-card">
                        <div class="fc-medal">{me.rank <= 3 ? rankEmoji(me.rank) : '🎯'}</div>
                        <div class="fc-pos">#{me.rank} o'rin</div>
                        <div class="fc-score">{me.score.toLocaleString()} ball</div>
                        {#if me.streak >= 3}<div class="fc-streak">🔥 {me.streak}ta streak</div>{/if}
                    </div>
                {/if}
            {/if}
            <div class="lb-list">
                {#each $gameStore.leaderboard.slice(0, 15) as e, i}
                    <div class="lb-row" class:lb-me={e.id === $gameStore.myParticipantId}
                         style="animation-delay:{i * 50}ms">
                        <span class="lb-rank">{rankEmoji(e.rank)}</span>
                        <span class="lb-av">{e.avatar}</span>
                        <span class="lb-name">{e.nickname}</span>
                        <span class="lb-score">{e.score.toLocaleString()}</span>
                    </div>
                {/each}
            </div>
            <a href="/join" class="play-again">Yana o'ynash →</a>
        </div>
    {/if}
</div>

<style>
    :global(body) { margin: 0; }

    .screen {
        min-height: 100dvh;
        display: flex;
        flex-direction: column;
        font-family: 'Segoe UI', system-ui, sans-serif;
    }

    .conn-bar {
        background: #f59e0b;
        color: white;
        text-align: center;
        padding: 7px;
        font-size: 0.82rem;
        font-weight: 600;
        z-index: 100;
    }
    .conn-bar.err { background: #ef4444; }

    /* ── LOBBY ── */
    .lobby {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 16px;
        background: linear-gradient(160deg, #4f46e5 0%, #7c3aed 60%, #9333ea 100%);
        padding: 32px 20px;
        color: white;
        text-align: center;
    }
    .quiz-name {
        font-size: 1rem;
        opacity: 0.75;
        margin: 0;
        background: rgba(255,255,255,0.15);
        padding: 4px 14px;
        border-radius: 999px;
    }
    .av-wrap {
        position: relative;
        width: 120px;
        height: 120px;
        display: flex;
        align-items: center;
        justify-content: center;
        margin: 8px 0;
    }
    .ring {
        position: absolute;
        border-radius: 50%;
        border: 2px solid rgba(255,255,255,0.3);
        animation: expand 2s ease-out infinite;
    }
    .ring.r1 { width: 120px; height: 120px; }
    .ring.r2 { width: 120px; height: 120px; animation-delay: 0.7s; }
    @keyframes expand {
        0%   { transform: scale(0.85); opacity: 0.6; }
        100% { transform: scale(1.5);  opacity: 0; }
    }
    .av-circle {
        width: 90px;
        height: 90px;
        border-radius: 50%;
        background: rgba(255,255,255,0.2);
        backdrop-filter: blur(4px);
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 3.2rem;
        position: relative;
        z-index: 1;
        border: 3px solid rgba(255,255,255,0.4);
    }
    .my-name { font-size: 1.4rem; font-weight: 700; margin: 0; }
    .wait-msg { font-size: 0.95rem; opacity: 0.8; margin: 0; }
    .pc-badge {
        background: rgba(255,255,255,0.2);
        border-radius: 999px;
        padding: 6px 18px;
        font-size: 0.9rem;
        font-weight: 600;
    }

    /* ── QUESTION ── */
    .q-screen {
        flex: 1;
        display: flex;
        flex-direction: column;
        background: #0f172a;
        overflow: hidden;
    }
    .q-topbar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 10px 16px;
        background: #1e293b;
        border-bottom: 1px solid #334155;
    }
    .q-num { color: #94a3b8; font-size: 0.85rem; font-weight: 600; }
    .q-pts {
        color: #fbbf24;
        font-size: 0.85rem;
        font-weight: 700;
        background: rgba(251,191,36,0.15);
        padding: 3px 10px;
        border-radius: 999px;
    }
    .timer-wrap {
        position: relative;
        width: 48px;
        height: 48px;
        display: flex;
        align-items: center;
        justify-content: center;
    }
    .timer-svg {
        position: absolute;
        inset: 0;
        transform: rotate(-90deg);
        width: 100%;
        height: 100%;
    }
    .timer-num {
        font-size: 1rem;
        font-weight: 800;
        position: relative;
        z-index: 1;
    }
    .q-body {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 20px 16px 12px;
        text-align: center;
        min-height: 0;
    }
    .q-media {
        max-width: 100%;
        max-height: 180px;
        border-radius: 10px;
        margin-bottom: 12px;
        object-fit: contain;
    }
    .q-text {
        font-size: clamp(1rem, 3.5vw, 1.6rem);
        font-weight: 700;
        color: #f1f5f9;
        line-height: 1.35;
        margin: 0;
    }

    .opts-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 8px;
        padding: 0 12px 12px;
    }
    .opt {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 14px 16px;
        border: none;
        border-radius: 10px;
        background: var(--c);
        color: white;
        font-size: clamp(0.85rem, 2.5vw, 1rem);
        font-weight: 700;
        cursor: pointer;
        transition: transform 0.12s, opacity 0.2s, filter 0.2s;
        text-align: left;
        min-height: 58px;
        position: relative;
    }
    .opt:hover:not(:disabled) { transform: scale(1.03); filter: brightness(1.1); }
    .opt:active:not(:disabled) { transform: scale(0.97); }
    .opt.selected { outline: 4px solid white; }
    .opt.dimmed { opacity: 0.35; filter: grayscale(0.4); }
    .opt-shape { font-size: 1.1rem; flex-shrink: 0; }
    .opt-label { flex: 1; line-height: 1.2; }
    .opt-check { font-size: 1.2rem; margin-left: auto; }

    .text-answer {
        display: flex;
        gap: 8px;
        padding: 0 12px 12px;
    }
    .text-input {
        flex: 1;
        padding: 12px 16px;
        border: 2px solid #334155;
        border-radius: 10px;
        background: #1e293b;
        color: #f1f5f9;
        font-size: 1rem;
        outline: none;
    }
    .text-input:focus { border-color: #6366f1; }
    .text-submit {
        padding: 12px 18px;
        background: #6366f1;
        color: white;
        border: none;
        border-radius: 10px;
        font-size: 1.2rem;
        cursor: pointer;
    }

    .waiting-pill {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        color: #94a3b8;
        font-size: 0.88rem;
        padding: 8px;
        animation: fadeIn 0.3s ease;
    }
    .dot-spin {
        width: 14px;
        height: 14px;
        border: 2px solid #334155;
        border-top-color: #6366f1;
        border-radius: 50%;
        animation: spin 0.7s linear infinite;
    }
    @keyframes spin { to { transform: rotate(360deg); } }
    @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }

    .emoji-bar {
        display: flex;
        justify-content: center;
        gap: 6px;
        padding: 8px 12px 14px;
        background: #1e293b;
        border-top: 1px solid #334155;
    }
    .emoji-btn {
        font-size: 1.4rem;
        background: #334155;
        border: none;
        border-radius: 50%;
        width: 40px;
        height: 40px;
        cursor: pointer;
        transition: transform 0.1s;
    }
    .emoji-btn:active { transform: scale(0.88); }

    /* ── QUESTION END ── */
    .end-screen {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 20px;
        padding: 24px 16px;
        background: linear-gradient(160deg, #0f172a, #1e293b);
        overflow-y: auto;
    }
    .result-card {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 8px;
        padding: 28px 36px;
        border-radius: 20px;
        border: 3px solid rgba(255,255,255,0.1);
        color: white;
        text-align: center;
        transform: translateY(30px);
        opacity: 0;
        transition: transform 0.4s cubic-bezier(0.34,1.56,0.64,1), opacity 0.35s ease;
        min-width: 220px;
    }
    .result-card.visible { transform: translateY(0); opacity: 1; }
    .result-card.correct { background: linear-gradient(135deg, #14532d, #15803d); }
    .result-card.wrong   { background: linear-gradient(135deg, #7f1d1d, #b91c1c); }
    .result-icon { font-size: 3rem; }
    .pts-earned  { font-size: 2rem; font-weight: 800; color: #86efac; }
    .streak-fire { font-size: 1.1rem; color: #fbbf24; font-weight: 700; }
    .wrong-label { font-size: 1.3rem; font-weight: 700; color: #fca5a5; }
    .total-score { font-size: 0.95rem; color: rgba(255,255,255,0.7); margin-top: 4px; }

    .correct-section { width: 100%; max-width: 420px; }
    .correct-heading { color: #94a3b8; font-size: 0.85rem; text-align: center; margin: 0 0 8px; text-transform: uppercase; letter-spacing: 0.05em; }
    .correct-opts { display: flex; flex-direction: column; gap: 8px; }
    .correct-opt {
        padding: 14px 20px;
        border-radius: 10px;
        font-weight: 700;
        font-size: 1rem;
        color: white;
        text-align: center;
        animation: slideIn 0.3s ease;
    }
    @keyframes slideIn { from { transform: translateX(-16px); opacity: 0; } to { transform: none; opacity: 1; } }

    /* ── LEADERBOARD ── */
    .lb-screen {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 12px;
        padding: 20px 16px;
        background: linear-gradient(160deg, #0f172a, #1e293b);
        overflow-y: auto;
    }
    .lb-heading { color: #f1f5f9; font-size: 1.4rem; font-weight: 800; margin: 0; }

    .my-rank {
        display: flex;
        align-items: center;
        gap: 10px;
        background: linear-gradient(135deg, #4f46e5, #7c3aed);
        border-radius: 12px;
        padding: 12px 20px;
        width: 100%;
        max-width: 420px;
        color: white;
        font-weight: 700;
    }
    .mr-rank { font-size: 1.3rem; }
    .mr-av   { font-size: 1.5rem; }
    .mr-name { flex: 1; font-size: 0.95rem; }
    .mr-score { font-size: 1.1rem; }

    .lb-list { display: flex; flex-direction: column; gap: 6px; width: 100%; max-width: 420px; }
    .lb-row {
        display: flex;
        align-items: center;
        gap: 10px;
        background: #1e293b;
        border: 1px solid #334155;
        border-radius: 10px;
        padding: 10px 14px;
        color: #e2e8f0;
        animation: rowIn 0.35s ease both;
    }
    .lb-row.lb-me {
        background: linear-gradient(135deg, #1e1b4b, #2e1065);
        border-color: #6366f1;
        font-weight: 700;
    }
    @keyframes rowIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: none; } }
    .lb-rank { width: 36px; text-align: center; font-size: 1rem; flex-shrink: 0; }
    .lb-av   { font-size: 1.4rem; flex-shrink: 0; }
    .lb-name { flex: 1; font-size: 0.9rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .lb-streak { color: #fbbf24; font-size: 0.8rem; flex-shrink: 0; }
    .lb-score { font-size: 0.95rem; font-weight: 700; color: #fbbf24; flex-shrink: 0; }

    /* ── PAUSED ── */
    .centered { flex: 1; display: flex; align-items: center; justify-content: center; background: #0f172a; }
    .pause-card {
        text-align: center;
        background: #1e293b;
        border-radius: 20px;
        padding: 40px 48px;
        color: white;
        border: 1px solid #334155;
    }
    .pause-icon { font-size: 3rem; margin-bottom: 12px; }
    .pause-card h2 { margin: 0 0 8px; font-size: 1.4rem; }
    .pause-card p  { color: #94a3b8; margin: 0; font-size: 0.9rem; }

    /* ── GAME OVER ── */
    .gameover-screen {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 16px;
        padding: 24px 16px;
        background: linear-gradient(160deg, #0f172a, #1e293b);
        overflow-y: auto;
    }
    .go-title { color: #f1f5f9; font-size: 1.8rem; font-weight: 800; margin: 0; }
    .final-card {
        text-align: center;
        background: linear-gradient(135deg, #1e1b4b, #4f46e5);
        border-radius: 20px;
        padding: 24px 40px;
        color: white;
        min-width: 200px;
    }
    .fc-medal { font-size: 3rem; }
    .fc-pos   { font-size: 1.1rem; color: rgba(255,255,255,0.7); margin-top: 4px; }
    .fc-score { font-size: 2rem; font-weight: 800; color: #fbbf24; }
    .fc-streak{ color: #fbbf24; font-size: 0.95rem; margin-top: 4px; }

    .play-again {
        display: inline-block;
        padding: 14px 28px;
        background: linear-gradient(135deg, #4f46e5, #7c3aed);
        color: white;
        border-radius: 12px;
        text-decoration: none;
        font-weight: 700;
        font-size: 1rem;
        transition: opacity 0.2s;
        margin-top: 8px;
    }
    .play-again:hover { opacity: 0.85; }
</style>
