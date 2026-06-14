<script lang="ts">
    import { page } from '$app/stores';
    import { onMount, onDestroy } from 'svelte';
    import { goto } from '$app/navigation';
    import { GameSocket } from '$lib/websocket/connection';
    import { gameStore } from '$lib/stores/game';
    import { getWebSocketURL } from '$lib/api/client';
    import type { QuestionEndPayload } from '$lib/api/types';
    import QRCode from 'qrcode';

    const pin = $page.params.roomCode;

    // QR kod + kirish havolasi (har sessiyada avtomatik)
    let joinUrl = '';
    let qrDataUrl = '';
    let urlCopied = false;
    function copyUrl() {
        navigator.clipboard.writeText(joinUrl).then(() => {
            urlCopied = true;
            setTimeout(() => (urlCopied = false), 2000);
        }).catch(() => {});
    }
    let socket: GameSocket;
    let correctIds: string[] = [];
    let questionStats: QuestionEndPayload['stats'] | null = null;

    $: phase = $gameStore.phase;
    $: question = $gameStore.currentQuestion;
    $: activePlayers = $gameStore.players.filter(p => p.is_active);
    $: isSelfPaced = $gameStore.gameMode === 'self_paced' || $gameStore.gameMode === 'team';
    $: spProgress = $gameStore.selfPacedProgress;
    $: teamStandings = $gameStore.teamStandings;
    let monitoring = false;
    $: answerPct = $gameStore.totalCount > 0
        ? Math.round(($gameStore.answeredCount / $gameStore.totalCount) * 100)
        : 0;

    $: timerRatio = question ? $gameStore.secondsLeft / question.time_limit : 1;
    $: timerColor = timerRatio > 0.5 ? '#22c55e' : timerRatio > 0.25 ? '#f59e0b' : '#ef4444';
    $: dashArr = `${timerRatio * 100} ${(1 - timerRatio) * 100}`;

    const optColors = ['#e21b3c', '#1368ce', '#d89e00', '#26890c'];
    const optShapes = ['▲', '◆', '●', '■'];

    onMount(() => {
        const token = localStorage.getItem('token');
        if (!token) { goto('/auth/login'); return; }

        // Kirish havolasi va QR kodni generatsiya qilish
        joinUrl = `${window.location.origin}/join?pin=${pin}`;
        QRCode.toDataURL(joinUrl, { width: 360, margin: 1, color: { dark: '#0f172a', light: '#ffffff' } })
            .then(url => qrDataUrl = url)
            .catch(() => {});

        socket = new GameSocket(getWebSocketURL(pin, 'host', token));
        socket.onStatus(s => gameStore.setConnectionStatus(s));

        socket.on('room_state',   (m) => gameStore.applyRoomState(m.payload));
        socket.on('player_joined',(m) => gameStore.playerJoined(m.payload.player));
        socket.on('player_left',  (m) => gameStore.playerLeft(m.payload.id));
        socket.on('game_started', () => { if (isSelfPaced) monitoring = true; });
        socket.on('self_paced_progress', (m) => gameStore.applySelfPacedProgress(m.payload));
        socket.on('question', (m) => {
            correctIds = [];
            questionStats = null;
            gameStore.showQuestion(m.payload);
        });
        socket.on('timer',        (m) => gameStore.setTimer(m.payload.seconds_left));
        socket.on('answer_count', (m) => gameStore.setAnswerCount(m.payload.answered, m.payload.total));
        socket.on('question_end', (m: { payload: QuestionEndPayload }) => {
            correctIds = m.payload.correct_options;
            questionStats = m.payload.stats;
            gameStore.showQuestionEnd();
        });
        socket.on('leaderboard',  (m) => gameStore.showLeaderboard(m.payload.players));
        socket.on('game_paused',  () => gameStore.pause());
        socket.on('game_resumed', () => gameStore.resume());
        socket.on('game_over',    (m) => gameStore.gameOver(m.payload));

        socket.connect();
    });

    onDestroy(() => { socket?.disconnect(); gameStore.reset(); });

    function startGame()   { socket.send('start_game'); }
    function nextQuestion(){ socket.send('next_question'); }
    function pauseGame()   { socket.send('pause_game'); }
    function resumeGame()  { socket.send('resume_game'); }
    function endGame() {
        if (confirm("O'yinni yakunlashni xohlaysizmi?")) socket.send('end_game');
    }
    function kickPlayer(id: string) { socket.send('kick_player', { participant_id: id }); }

    function optionAnswerCount(optId: string): number {
        return questionStats?.option_counts?.[optId] ?? 0;
    }
    function maxOptionCount(): number {
        if (!questionStats?.option_counts) return 1;
        return Math.max(1, ...Object.values(questionStats.option_counts));
    }

    function rankEmoji(r: number) {
        return r === 1 ? '🥇' : r === 2 ? '🥈' : r === 3 ? '🥉' : `#${r}`;
    }
</script>

<svelte:head><title>Host — PIN: {pin}</title></svelte:head>

<div class="host">

    <!-- Top bar -->
    <header class="bar">
        <div class="pin-wrap">
            <span class="pin-label">PIN</span>
            <span class="pin-val">{pin}</span>
        </div>
        <div class="conn-dot" class:live={$gameStore.connectionStatus === 'connected'}>
            {$gameStore.connectionStatus === 'connected' ? '● Live' : '● Uzilgan'}
        </div>
        <div class="bar-actions">
            {#if phase === 'question' || phase === 'answered'}
                <button class="btn sec" on:click={pauseGame}>⏸</button>
            {:else if phase === 'paused'}
                <button class="btn pri" on:click={resumeGame}>▶ Davom</button>
            {/if}
            <button class="btn dan" on:click={endGame}>■ Yakunla</button>
        </div>
    </header>

    <main class="main">

        <!-- ═══ SELF-PACED / TEAM MONITORING ════════════════════════════════ -->
        {#if monitoring && phase !== 'game_over'}
            <div class="monitor">
                <div class="mon-head">
                    <h2 class="mon-title">{$gameStore.gameMode === 'team' ? '👥 Jamoaviy rejim' : '🎯 Mustaqil rejim'} — jonli natijalar</h2>
                    <p class="mon-sub">O'quvchilar o'z tezligida yechmoqda</p>
                </div>

                {#if $gameStore.gameMode === 'team' && teamStandings.length}
                    <div class="team-board">
                        {#each teamStandings as t (t.team_id)}
                            <div class="team-card">
                                <span class="team-rank">{rankEmoji(t.rank)}</span>
                                <span class="team-name">{t.name}</span>
                                <span class="team-members">{t.members} a'zo</span>
                                <span class="team-score">{t.score.toLocaleString()}</span>
                            </div>
                        {/each}
                    </div>
                {/if}

                <div class="mon-list">
                    {#each spProgress as p (p.id)}
                        <div class="mon-row" class:done={p.finished}>
                            <span class="mon-av">{p.avatar}</span>
                            <span class="mon-nick">
                                {p.nickname}
                                {#if p.team_id}<span class="mon-team">J{p.team_id}</span>{/if}
                            </span>
                            <div class="mon-bar">
                                <div class="mon-bar-fill" style="width:{p.total ? (p.answered / p.total) * 100 : 0}%"></div>
                            </div>
                            <span class="mon-count">{p.answered}/{p.total}</span>
                            <span class="mon-score">{p.score.toLocaleString()}</span>
                            {#if p.finished}<span class="mon-done">✓</span>{/if}
                        </div>
                    {:else}
                        <p class="mon-empty">Hali natija yo'q...</p>
                    {/each}
                </div>
            </div>

        <!-- ═══ LOBBY ══════════════════════════════════════════════════════ -->
        {:else if phase === 'lobby' || phase === 'idle'}
            <div class="lobby">
                <div class="lob-quiz">
                    <p class="lob-title">{$gameStore.roomInfo?.quiz_title ?? 'Yuklanmoqda...'}</p>
                    <p class="lob-sub">{$gameStore.totalQuestions} ta savol</p>
                </div>
                <div class="pin-card">
                    <div class="join-cols">
                        <div class="join-left">
                            <p class="pin-instr">O'quvchilar <strong>cognita.uz/join</strong> ga kirib PIN ni yozadi:</p>
                            <div class="big-pin">{pin}</div>
                            {#if joinUrl}
                                <div class="url-row">
                                    <span class="url-text" title={joinUrl}>{joinUrl}</span>
                                    <button class="url-copy" on:click={copyUrl}>
                                        {urlCopied ? '✓ Nusxalandi' : '📋 Nusxa'}
                                    </button>
                                </div>
                            {/if}
                        </div>
                        {#if qrDataUrl}
                            <div class="qr-box">
                                <img src={qrDataUrl} alt="QR kod orqali kirish" class="qr-img" />
                                <span class="qr-cap">📱 QR kodni skanerlang</span>
                            </div>
                        {/if}
                    </div>
                </div>
                <div class="players-wrap">
                    <p class="players-heading">👥 {activePlayers.length} o'yinchi qo'shildi</p>
                    <div class="players-grid">
                        {#each activePlayers as p (p.id)}
                            <div class="p-chip">
                                <span class="p-av">{p.avatar}</span>
                                <span class="p-nick">{p.nickname}</span>
                                <button class="kick" on:click={() => kickPlayer(p.id)} title="Chiqarish">✕</button>
                            </div>
                        {/each}
                    </div>
                </div>
                <button
                    class="btn pri large"
                    disabled={activePlayers.length === 0}
                    on:click={startGame}
                >
                    🚀 Boshlash ({activePlayers.length} o'yinchi)
                </button>
            </div>

        <!-- ═══ QUESTION ════════════════════════════════════════════════════ -->
        {:else if phase === 'question' || phase === 'answered'}
            <div class="q-view">
                <div class="q-header">
                    <div class="q-progress-wrap">
                        <span class="q-idx">{(question?.question_index ?? 0) + 1} / {$gameStore.totalQuestions}</span>
                        <div class="q-prog-bar">
                            <div class="q-prog-fill" style="width:{((( question?.question_index ?? 0) + 1) / ($gameStore.totalQuestions || 1)) * 100}%"></div>
                        </div>
                    </div>
                    <div class="timer-wrap">
                        <svg viewBox="0 0 36 36" class="timer-svg">
                            <circle cx="18" cy="18" r="15.9" fill="none" stroke="#334155" stroke-width="3.2"/>
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
                    <span class="pts-badge">⭐ {question?.points ?? 0}</span>
                </div>

                <div class="q-body">
                    {#if question?.media_url}
                        <img src={question.media_url} alt="" class="q-media" />
                    {/if}
                    <h2 class="q-text">{question?.question_text ?? ''}</h2>
                </div>

                {#if question?.options}
                    <div class="opts-grid">
                        {#each question.options as opt, i}
                            <div class="opt-chip" style="background:{optColors[i % 4]}">
                                <span class="opt-sh">{optShapes[i % 4]}</span>
                                <span class="opt-lbl">{opt.option_text}</span>
                            </div>
                        {/each}
                    </div>
                {/if}

                <!-- Live answer progress -->
                <div class="ans-bar-wrap">
                    <div class="ans-bar-track">
                        <div class="ans-bar-fill" style="width:{answerPct}%;transition:width 0.4s ease"></div>
                    </div>
                    <span class="ans-count">
                        {$gameStore.answeredCount} / {$gameStore.totalCount} javob berdi
                        {#if answerPct === 100}
                            <span class="all-done">✓ Hammasi!</span>
                        {/if}
                    </span>
                </div>
            </div>

        <!-- ═══ QUESTION END ════════════════════════════════════════════════ -->
        {:else if phase === 'question_end'}
            <div class="end-view">
                <h2 class="end-heading">✅ To'g'ri javob</h2>

                {#if question}
                    <div class="correct-opts">
                        {#each question.options.filter(o => correctIds.includes(o.id)) as opt, i}
                            <div class="correct-opt" style="background:{optColors[i % 4]}">{opt.option_text}</div>
                        {/each}
                    </div>
                {/if}

                <!-- Answer distribution chart -->
                {#if question && questionStats}
                    {@const mx = maxOptionCount()}
                    <div class="dist-chart">
                        {#each question.options as opt, i}
                            {@const cnt = optionAnswerCount(opt.id)}
                            {@const isCorrect = correctIds.includes(opt.id)}
                            <div class="dist-col">
                                <span class="dist-count">{cnt}</span>
                                <div class="dist-bar-wrap">
                                    <div class="dist-bar"
                                         class:correct-bar={isCorrect}
                                         style="height:{mx > 0 ? Math.round((cnt / mx) * 100) : 0}%;background:{optColors[i % 4]};transition:height 0.6s cubic-bezier(0.34,1.56,0.64,1) {i * 80}ms">
                                    </div>
                                </div>
                                <span class="dist-shape" style="color:{optColors[i % 4]}">{optShapes[i % 4]}</span>
                            </div>
                        {/each}
                    </div>
                {/if}

                {#if questionStats}
                    <div class="stats-row">
                        <div class="stat-box">
                            <span class="stat-n">{questionStats.total_answers}</span>
                            <span class="stat-l">Javob berdi</span>
                        </div>
                        <div class="stat-box">
                            <span class="stat-n">
                                {questionStats.total_answers > 0
                                    ? Math.round(questionStats.correct_count / questionStats.total_answers * 100)
                                    : 0}%
                            </span>
                            <span class="stat-l">To'g'ri</span>
                        </div>
                        <div class="stat-box">
                            <span class="stat-n">{Math.round((questionStats.average_time_ms ?? 0) / 1000)}s</span>
                            <span class="stat-l">O'rtacha vaqt</span>
                        </div>
                    </div>
                {/if}

                <button class="btn pri large" on:click={nextQuestion}>
                    Keyingi savol →
                </button>
            </div>

        <!-- ═══ LEADERBOARD ═════════════════════════════════════════════════ -->
        {:else if phase === 'leaderboard'}
            <div class="lb-view">
                <h2 class="lb-heading">🏆 Liderlar jadvali</h2>

                <!-- Top 3 podium -->
                {#if $gameStore.leaderboard.length >= 3}
                    {@const top = $gameStore.leaderboard.slice(0, 3)}
                    <div class="podium">
                        <div class="pod-col pod-2">
                            <div class="pod-av">{top[1].avatar}</div>
                            <div class="pod-name">{top[1].nickname}</div>
                            <div class="pod-score">{top[1].score.toLocaleString()}</div>
                            <div class="pod-block h2">🥈</div>
                        </div>
                        <div class="pod-col pod-1">
                            <div class="pod-av">{top[0].avatar}</div>
                            <div class="pod-name">{top[0].nickname}</div>
                            <div class="pod-score">{top[0].score.toLocaleString()}</div>
                            <div class="pod-block h1">🥇</div>
                        </div>
                        <div class="pod-col pod-3">
                            <div class="pod-av">{top[2].avatar}</div>
                            <div class="pod-name">{top[2].nickname}</div>
                            <div class="pod-score">{top[2].score.toLocaleString()}</div>
                            <div class="pod-block h3">🥉</div>
                        </div>
                    </div>
                {/if}

                <div class="lb-list">
                    {#each $gameStore.leaderboard.slice(0, 15) as e, i}
                        <div class="lb-row" style="animation-delay:{i * 50}ms">
                            <span class="lb-rank">{rankEmoji(e.rank)}</span>
                            <span class="lb-av">{e.avatar}</span>
                            <span class="lb-name">{e.nickname}</span>
                            {#if e.streak >= 2}<span class="lb-streak">🔥{e.streak}</span>{/if}
                            <span class="lb-score">{e.score.toLocaleString()}</span>
                            {#if e.delta !== 0}
                                <span class="lb-delta" class:up={e.delta > 0} class:dn={e.delta < 0}>
                                    {e.delta > 0 ? `▲${e.delta}` : `▼${Math.abs(e.delta)}`}
                                </span>
                            {/if}
                        </div>
                    {/each}
                </div>

                <button class="btn pri large" on:click={nextQuestion}>
                    Keyingi savol →
                </button>
            </div>

        <!-- ═══ PAUSED ══════════════════════════════════════════════════════ -->
        {:else if phase === 'paused'}
            <div class="centered">
                <div class="pause-card">
                    <div class="pause-icon">⏸</div>
                    <h2>Pauza</h2>
                    <button class="btn pri" on:click={resumeGame}>▶ Davom ettirish</button>
                </div>
            </div>

        <!-- ═══ GAME OVER ════════════════════════════════════════════════════ -->
        {:else if phase === 'game_over'}
            <div class="lb-view">
                <h1 class="go-title">🏆 O'yin tugadi!</h1>

                {#if $gameStore.leaderboard.length >= 3}
                    {@const top = $gameStore.leaderboard.slice(0, 3)}
                    <div class="podium">
                        <div class="pod-col pod-2">
                            <div class="pod-av">{top[1].avatar}</div>
                            <div class="pod-name">{top[1].nickname}</div>
                            <div class="pod-score">{top[1].score.toLocaleString()}</div>
                            <div class="pod-block h2">🥈</div>
                        </div>
                        <div class="pod-col pod-1">
                            <div class="pod-av">{top[0].avatar}</div>
                            <div class="pod-name">{top[0].nickname}</div>
                            <div class="pod-score">{top[0].score.toLocaleString()}</div>
                            <div class="pod-block h1">🥇</div>
                        </div>
                        <div class="pod-col pod-3">
                            <div class="pod-av">{top[2].avatar}</div>
                            <div class="pod-name">{top[2].nickname}</div>
                            <div class="pod-score">{top[2].score.toLocaleString()}</div>
                            <div class="pod-block h3">🥉</div>
                        </div>
                    </div>
                {/if}

                {#if teamStandings.length}
                    <div class="team-board final">
                        {#each teamStandings as t (t.team_id)}
                            <div class="team-card">
                                <span class="team-rank">{rankEmoji(t.rank)}</span>
                                <span class="team-name">{t.name}</span>
                                <span class="team-members">{t.members} a'zo</span>
                                <span class="team-score">{t.score.toLocaleString()}</span>
                            </div>
                        {/each}
                    </div>
                {/if}

                <div class="lb-list">
                    {#each $gameStore.leaderboard.slice(0, 30) as e, i}
                        <div class="lb-row" style="animation-delay:{i * 40}ms">
                            <span class="lb-rank">{rankEmoji(e.rank)}</span>
                            <span class="lb-av">{e.avatar}</span>
                            <span class="lb-name">{e.nickname}</span>
                            <span class="lb-score">{e.score.toLocaleString()}</span>
                        </div>
                    {/each}
                </div>

                <a href="/dashboard" class="btn pri large dash-link">Dashboardga qaytish</a>
            </div>
        {/if}

    </main>
</div>

<style>
    :global(body) { margin: 0; }

    .host {
        min-height: 100dvh;
        display: flex;
        flex-direction: column;
        background: #0f172a;
        color: #e2e8f0;
        font-family: 'Segoe UI', system-ui, sans-serif;
    }

    /* ── TOP BAR ── */
    .bar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 10px 20px;
        background: #1e293b;
        border-bottom: 1px solid #334155;
        flex-wrap: wrap;
        gap: 8px;
        position: sticky;
        top: 0;
        z-index: 50;
    }
    .pin-wrap { display: flex; align-items: center; gap: 8px; }
    .pin-label { font-size: 0.75rem; color: #64748b; text-transform: uppercase; letter-spacing: 0.1em; }
    .pin-val   { font-size: 1.5rem; font-weight: 900; letter-spacing: 0.18em; color: #f1f5f9; }
    .conn-dot  { font-size: 0.82rem; color: #ef4444; }
    .conn-dot.live { color: #22c55e; }
    .bar-actions { display: flex; gap: 6px; }

    .main { flex: 1; padding: 20px; overflow-y: auto; }

    /* ── LOBBY ── */
    .lobby { max-width: 900px; margin: 0 auto; display: flex; flex-direction: column; align-items: center; gap: 24px; }
    .lob-quiz { text-align: center; }
    .lob-title { font-size: 1.6rem; font-weight: 800; margin: 0 0 4px; }
    .lob-sub   { color: #64748b; margin: 0; }
    .pin-card {
        background: #1e293b;
        border: 1px solid #334155;
        border-radius: 20px;
        padding: 26px 32px;
        text-align: center;
        width: 100%;
        max-width: 640px;
    }
    .pin-instr { color: #94a3b8; margin: 0 0 12px; font-size: 0.95rem; }
    .pin-instr strong { color: #6366f1; }
    .big-pin { font-size: clamp(2.6rem, 9vw, 4.5rem); font-weight: 900; letter-spacing: 0.28em; color: #f1f5f9; line-height: 1; }

    /* ── QR + havola ── */
    .join-cols { display: flex; align-items: center; gap: 28px; justify-content: center; flex-wrap: wrap; }
    .join-left { flex: 1; min-width: 220px; }
    .url-row {
        display: flex; align-items: center; gap: 8px; margin-top: 16px;
        background: #0f172a; border: 1px solid #334155; border-radius: 10px; padding: 6px 6px 6px 14px;
    }
    .url-text { flex: 1; color: #cbd5e1; font-size: 0.82rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; text-align: left; }
    .url-copy {
        flex-shrink: 0; border: none; border-radius: 8px; cursor: pointer;
        background: #6366f1; color: #fff; font-weight: 700; font-size: 0.78rem; padding: 8px 12px; white-space: nowrap;
    }
    .url-copy:hover { filter: brightness(1.1); }
    .qr-box { display: flex; flex-direction: column; align-items: center; gap: 8px; }
    .qr-img {
        width: 180px; height: 180px; border-radius: 14px; background: #fff; padding: 8px;
        box-shadow: 0 8px 24px rgba(0,0,0,0.35);
    }
    .qr-cap { color: #94a3b8; font-size: 0.8rem; font-weight: 600; }
    @media (max-width: 560px) { .qr-img { width: 150px; height: 150px; } }
    .players-wrap { width: 100%; max-width: 700px; }
    .players-heading { color: #94a3b8; font-size: 0.85rem; margin: 0 0 10px; text-align: center; }
    .players-grid { display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; }
    .p-chip {
        display: flex; align-items: center; gap: 6px;
        background: #1e293b; border: 1px solid #334155; border-radius: 999px;
        padding: 6px 14px; font-size: 0.9rem;
        animation: popIn 0.25s cubic-bezier(0.34,1.56,0.64,1);
    }
    @keyframes popIn { from { transform: scale(0.5); opacity: 0; } to { transform: scale(1); opacity: 1; } }
    .p-av   { font-size: 1.2rem; }
    .p-nick { font-weight: 600; color: #e2e8f0; }
    .kick   { background: none; border: none; color: #475569; cursor: pointer; font-size: 0.75rem; padding: 0 2px; }
    .kick:hover { color: #ef4444; }

    /* ── QUESTION ── */
    .q-view { max-width: 900px; margin: 0 auto; display: flex; flex-direction: column; gap: 20px; }
    .q-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
    .q-progress-wrap { flex: 1; display: flex; flex-direction: column; gap: 6px; }
    .q-idx { font-size: 0.85rem; color: #94a3b8; }
    .q-prog-bar { height: 4px; background: #334155; border-radius: 999px; overflow: hidden; }
    .q-prog-fill { height: 100%; background: linear-gradient(90deg, #6366f1, #a78bfa); border-radius: 999px; transition: width 0.4s ease; }
    .pts-badge { background: rgba(251,191,36,0.15); color: #fbbf24; padding: 4px 12px; border-radius: 999px; font-size: 0.85rem; font-weight: 700; white-space: nowrap; }

    .timer-wrap { position: relative; width: 60px; height: 60px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
    .timer-svg { position: absolute; inset: 0; transform: rotate(-90deg); width: 100%; height: 100%; }
    .timer-num { font-size: 1.2rem; font-weight: 800; position: relative; z-index: 1; }

    .q-body { text-align: center; }
    .q-media { max-width: 100%; max-height: 220px; border-radius: 12px; margin-bottom: 12px; }
    .q-text { font-size: clamp(1.2rem, 3vw, 2rem); font-weight: 700; color: #f1f5f9; line-height: 1.3; margin: 0; }

    .opts-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
    .opt-chip {
        display: flex; align-items: center; gap: 10px;
        padding: 14px 18px; border-radius: 10px;
        font-weight: 700; font-size: clamp(0.85rem, 2vw, 1rem);
        color: white;
    }
    .opt-sh  { font-size: 1.1rem; flex-shrink: 0; }
    .opt-lbl { line-height: 1.2; }

    .ans-bar-wrap { display: flex; flex-direction: column; gap: 6px; }
    .ans-bar-track {
        height: 14px;
        background: #1e293b;
        border-radius: 999px;
        overflow: hidden;
        border: 1px solid #334155;
    }
    .ans-bar-fill {
        height: 100%;
        background: linear-gradient(90deg, #22c55e, #16a34a);
        border-radius: 999px;
    }
    .ans-count { font-size: 0.9rem; color: #94a3b8; text-align: center; }
    .all-done  { color: #22c55e; font-weight: 700; margin-left: 6px; }

    /* ── QUESTION END ── */
    .end-view { max-width: 700px; margin: 0 auto; display: flex; flex-direction: column; align-items: center; gap: 20px; }
    .end-heading { font-size: 1.4rem; font-weight: 700; margin: 0; color: #f1f5f9; }
    .correct-opts { display: flex; flex-direction: column; gap: 8px; width: 100%; }
    .correct-opt {
        padding: 14px 20px; border-radius: 10px;
        font-size: 1rem; font-weight: 700; color: white; text-align: center;
        animation: slideIn 0.35s ease;
    }
    @keyframes slideIn { from { opacity: 0; transform: translateY(-8px); } to { opacity: 1; transform: none; } }

    .dist-chart { display: flex; gap: 12px; align-items: flex-end; height: 140px; width: 100%; padding: 0 4px; }
    .dist-col { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 4px; height: 100%; }
    .dist-count { font-size: 0.8rem; font-weight: 700; color: #94a3b8; }
    .dist-bar-wrap { flex: 1; display: flex; align-items: flex-end; width: 100%; }
    .dist-bar { width: 100%; border-radius: 6px 6px 0 0; min-height: 4px; }
    .dist-bar.correct-bar { box-shadow: 0 0 12px rgba(34,197,94,0.5); }
    .dist-shape { font-size: 1rem; font-weight: 700; }

    .stats-row { display: flex; gap: 16px; flex-wrap: wrap; justify-content: center; }
    .stat-box { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 14px 20px; text-align: center; min-width: 90px; }
    .stat-n { display: block; font-size: 1.6rem; font-weight: 800; color: #f1f5f9; }
    .stat-l { font-size: 0.75rem; color: #64748b; }

    /* ── LEADERBOARD ── */
    .lb-view { max-width: 700px; margin: 0 auto; display: flex; flex-direction: column; align-items: center; gap: 20px; }
    .lb-heading { font-size: 1.5rem; font-weight: 800; margin: 0; }
    .go-title   { font-size: 2rem; font-weight: 800; margin: 0; }

    .podium { display: flex; align-items: flex-end; justify-content: center; gap: 8px; width: 100%; max-width: 420px; }
    .pod-col { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 4px; }
    .pod-av   { font-size: 2rem; }
    .pod-name { font-size: 0.75rem; font-weight: 600; text-align: center; color: #e2e8f0; max-width: 80px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .pod-score{ font-size: 0.7rem; color: #fbbf24; font-weight: 700; }
    .pod-block { width: 100%; border-radius: 8px 8px 0 0; display: flex; align-items: center; justify-content: center; font-size: 1.5rem; padding: 10px 0; }
    .h1 { background: linear-gradient(180deg, #fbbf24, #d97706); height: 80px; }
    .h2 { background: linear-gradient(180deg, #94a3b8, #64748b); height: 60px; }
    .h3 { background: linear-gradient(180deg, #cd7c2f, #92400e); height: 44px; }

    .lb-list { display: flex; flex-direction: column; gap: 6px; width: 100%; }
    .lb-row {
        display: flex; align-items: center; gap: 10px;
        background: #1e293b; border: 1px solid #334155; border-radius: 10px;
        padding: 10px 14px; color: #e2e8f0;
        animation: rowIn 0.35s ease both;
    }
    @keyframes rowIn { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: none; } }
    .lb-rank  { width: 36px; text-align: center; font-size: 1rem; flex-shrink: 0; }
    .lb-av    { font-size: 1.4rem; flex-shrink: 0; }
    .lb-name  { flex: 1; font-size: 0.9rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .lb-streak{ color: #fbbf24; font-size: 0.8rem; flex-shrink: 0; }
    .lb-score { font-size: 0.95rem; font-weight: 700; color: #fbbf24; flex-shrink: 0; }
    .lb-delta { font-size: 0.72rem; flex-shrink: 0; }
    .lb-delta.up { color: #22c55e; }
    .lb-delta.dn { color: #ef4444; }

    /* ── PAUSED ── */
    .centered { display: flex; align-items: center; justify-content: center; padding: 60px 20px; }
    .pause-card { text-align: center; background: #1e293b; border: 1px solid #334155; border-radius: 20px; padding: 40px 48px; display: flex; flex-direction: column; align-items: center; gap: 16px; }
    .pause-icon { font-size: 3rem; }
    .pause-card h2 { margin: 0; font-size: 1.4rem; }

    /* ── BUTTONS ── */
    .btn {
        padding: 10px 18px; border: none; border-radius: 8px;
        font-size: 0.88rem; font-weight: 700; cursor: pointer; transition: opacity 0.18s, transform 0.1s;
        white-space: nowrap;
    }
    .btn.pri { background: linear-gradient(135deg, #4f46e5, #7c3aed); color: white; }
    .btn.sec { background: #334155; color: #e2e8f0; }
    .btn.dan { background: #ef4444; color: white; }
    .btn.large { padding: 14px 28px; font-size: 1rem; border-radius: 12px; }
    .btn:disabled { opacity: 0.45; cursor: not-allowed; }
    .btn:not(:disabled):hover { opacity: 0.87; }
    .btn:not(:disabled):active { transform: scale(0.97); }
    .dash-link { text-decoration: none; display: inline-block; }

    /* ── Self-paced / team monitoring ── */
    .monitor { width: 100%; max-width: 760px; margin: 0 auto; }
    .mon-head { text-align: center; margin-bottom: 20px; }
    .mon-title { font-size: 1.4rem; font-weight: 800; color: #f1f5f9; margin: 0; }
    .mon-sub { color: #94a3b8; font-size: 0.9rem; margin-top: 4px; }
    .mon-list { display: flex; flex-direction: column; gap: 8px; }
    .mon-row {
        display: flex; align-items: center; gap: 12px;
        background: #1e293b; border: 1px solid #334155;
        border-radius: 12px; padding: 10px 14px;
        animation: fadeIn 0.3s ease;
    }
    .mon-row.done { border-color: #22c55e; background: rgba(34,197,94,0.08); }
    .mon-av { font-size: 1.5rem; }
    .mon-nick { font-weight: 700; color: #e2e8f0; min-width: 120px; display: flex; align-items: center; gap: 6px; }
    .mon-team { font-size: 0.65rem; background: #6366f1; color: #fff; padding: 1px 6px; border-radius: 99px; }
    .mon-bar { flex: 1; height: 8px; background: #334155; border-radius: 99px; overflow: hidden; }
    .mon-bar-fill { height: 100%; background: linear-gradient(90deg,#6366f1,#22c55e); transition: width 0.4s ease; }
    .mon-count { font-size: 0.82rem; color: #94a3b8; min-width: 42px; text-align: right; }
    .mon-score { font-weight: 800; color: #fbbf24; min-width: 64px; text-align: right; }
    .mon-done { color: #22c55e; font-weight: 800; }
    .mon-empty { text-align: center; color: #64748b; padding: 30px; }
    @keyframes fadeIn { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: none; } }

    /* ── Team standings ── */
    .team-board { display: flex; flex-direction: column; gap: 8px; margin-bottom: 20px; }
    .team-board.final { max-width: 520px; margin: 0 auto 20px; }
    .team-card {
        display: flex; align-items: center; gap: 12px;
        background: #1e293b; border: 1px solid #334155;
        border-radius: 12px; padding: 12px 16px;
    }
    .team-rank { font-size: 1.2rem; min-width: 34px; }
    .team-name { font-weight: 800; color: #f1f5f9; flex: 1; }
    .team-members { font-size: 0.78rem; color: #94a3b8; }
    .team-score { font-weight: 800; color: #fbbf24; min-width: 70px; text-align: right; }

    /* ── Dizayn boyitish: animatsiyalar ── */
    .big-pin { animation: pinPulse 2.6s ease-in-out infinite; }
    @keyframes pinPulse {
        0%,100% { transform: scale(1); text-shadow: 0 0 0 rgba(99,102,241,0); }
        50%     { transform: scale(1.04); text-shadow: 0 0 30px rgba(99,102,241,0.65); }
    }
    .opt-chip {
        animation: chipIn 0.45s cubic-bezier(0.22,1,0.36,1) both;
        box-shadow: 0 5px 0 rgba(0,0,0,0.22), 0 8px 18px rgba(0,0,0,0.25);
        position: relative; overflow: hidden;
    }
    .opt-chip::before {
        content: ''; position: absolute; inset: 0;
        background: linear-gradient(180deg, rgba(255,255,255,0.2), transparent 50%);
        pointer-events: none;
    }
    @keyframes chipIn { from { opacity: 0; transform: translateY(16px) scale(0.94); } to { opacity: 1; transform: none; } }
    .ans-bar-fill { transition: width 0.5s cubic-bezier(0.22,1,0.36,1); }

    .pod-col { animation: podRise 0.65s cubic-bezier(0.34,1.56,0.64,1) both; }
    .pod-1 { animation-delay: 0.1s; }
    .pod-2 { animation-delay: 0.28s; }
    .pod-3 { animation-delay: 0.44s; }
    @keyframes podRise { from { opacity: 0; transform: translateY(46px); } to { opacity: 1; transform: none; } }

    .btn.pri.large { animation: btnGlow 2.2s ease-in-out infinite; }
    @keyframes btnGlow {
        0%,100% { box-shadow: 0 0 0 rgba(99,102,241,0.0); }
        50%     { box-shadow: 0 0 26px rgba(99,102,241,0.55); }
    }
</style>
