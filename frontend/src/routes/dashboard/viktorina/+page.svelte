<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { quizzes as quizzesApi, rooms as roomsApi } from '$lib/api/client';
    import type { Quiz } from '$lib/api/types';
    import type { RoomHistory } from '$lib/api/client';

    // ── State ─────────────────────────────────────────────────────────────────
    let quizzes: Quiz[] = [];
    let history: RoomHistory[] = [];
    let loadingQuizzes = true;
    let loadingHistory = true;
    let launching = false;

    let selectedQuizId = '';
    let gameMode = 'classic';
    let shuffleQuestions = false;
    let shuffleAnswers = true;
    let showLeaderboard = true;
    let showCorrectAnswer = true;
    let music = true;
    let timePerQuestion = 20;

    // Accordion state
    let accordionOpen = false;

    // Active session after creation
    let activeSession: { pin: string; quizTitle: string; playerCount: number } | null = null;

    // ── Toast ─────────────────────────────────────────────────────────────────
    type ToastKind = 'success' | 'error';
    let toast: { msg: string; kind: ToastKind } | null = null;
    let toastTimer: ReturnType<typeof setTimeout>;

    function showToast(msg: string, kind: ToastKind = 'success') {
        toast = { msg, kind };
        clearTimeout(toastTimer);
        toastTimer = setTimeout(() => (toast = null), 3500);
    }

    // ── Derived stats ─────────────────────────────────────────────────────────
    $: activeSessions  = history.filter(r => r.status === 'in_progress').length;
    $: totalSessions   = history.length;
    $: totalPlayers    = history.reduce((s, r) => s + r.player_count, 0);

    // Most played quiz from history
    $: mostPlayedQuiz = (() => {
        if (history.length === 0) return '—';
        const counts: Record<string, number> = {};
        history.forEach(r => {
            if (r.quiz_title) counts[r.quiz_title] = (counts[r.quiz_title] ?? 0) + 1;
        });
        const top = Object.entries(counts).sort((a, b) => b[1] - a[1])[0];
        return top ? top[0] : '—';
    })();

    // Today's sessions
    $: todaySessions = (() => {
        const today = new Date().toDateString();
        return history.filter(r => new Date(r.created_at).toDateString() === today).length;
    })();

    // ── Game modes ────────────────────────────────────────────────────────────
    const modes = [
        { value: 'classic',    icon: '🎯',   label: 'Klassik',  desc: 'Savol ekranda, tezlik' },
        { value: 'self_paced', icon: '📲', label: 'Mustaqil', desc: 'Har kim o\'z qurilmasida' },
        { value: 'team',       icon: '👥',   label: 'Jamoaviy', desc: 'Jamoalarga bo\'linib' },
        { value: 'accuracy', icon: '💯', label: 'Aniqlik',  desc: 'Faqat to\'g\'ri javob' },
        { value: 'speed',    icon: '⚡',   label: 'Tezlik',   desc: 'Eng tez javob g\'alaba' },
    ];

    const modeColors: Record<string, string> = {
        classic:  '#6366f1',
        team:     '#22c55e',
        accuracy: '#f59e0b',
        speed:    '#ef4444',
    };

    // ── Load data ─────────────────────────────────────────────────────────────
    onMount(async () => {
        await Promise.all([loadQuizzes(), loadHistory()]);
    });

    async function loadQuizzes() {
        loadingQuizzes = true;
        try {
            quizzes = await quizzesApi.list();
            if (quizzes.length > 0) selectedQuizId = quizzes[0].id;
        } catch (e: any) {
            showToast('Quizlarni yuklashda xato: ' + (e.message ?? 'Noma\'lum xato'), 'error');
        } finally {
            loadingQuizzes = false;
        }
    }

    async function loadHistory() {
        loadingHistory = true;
        try {
            history = await roomsApi.history();
        } catch (e: any) {
            showToast('Tarixni yuklashda xato: ' + (e.message ?? 'Noma\'lum xato'), 'error');
            history = [];
        } finally {
            loadingHistory = false;
        }
    }

    // ── Launch ────────────────────────────────────────────────────────────────
    async function launchGame() {
        if (!selectedQuizId) { showToast('Quiz tanlang', 'error'); return; }
        launching = true;
        try {
            const res = await roomsApi.create(selectedQuizId, gameMode, {
                shuffle_questions: shuffleQuestions,
                shuffle_answers: shuffleAnswers,
                show_leaderboard: showLeaderboard,
                music,
                lobby_music: music,
                show_correct_answer: showCorrectAnswer,
                time_per_question: timePerQuestion,
            });
            const quiz = quizzes.find(q => q.id === selectedQuizId);
            activeSession = { pin: res.pin, quizTitle: quiz?.title ?? 'Quiz', playerCount: 0 };
            showToast('O\'yin yaratildi! PIN: ' + res.pin);
            await loadHistory();
        } catch (e: any) {
            showToast('O\'yin yaratishda xato: ' + (e.message ?? 'Noma\'lum xato'), 'error');
        } finally {
            launching = false;
        }
    }

    // Restart with same quiz
    async function restartSession(row: RoomHistory) {
        selectedQuizId = row.quiz_id ?? selectedQuizId;
        gameMode = row.game_mode ?? 'classic';
        await launchGame();
    }

    // ── Duration helper ───────────────────────────────────────────────────────
    function calcDuration(row: RoomHistory): string {
        if (!row.created_at || !row.ended_at) return '—';
        const diff = new Date(row.ended_at).getTime() - new Date(row.created_at).getTime();
        if (diff <= 0) return '—';
        const mins = Math.floor(diff / 60000);
        const secs = Math.floor((diff % 60000) / 1000);
        return mins > 0 ? `${mins}m ${secs}s` : `${secs}s`;
    }

    // ── Helpers ───────────────────────────────────────────────────────────────
    function fmtDate(d: string) {
        if (!d) return '—';
        return new Date(d).toLocaleDateString('uz-UZ', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' });
    }

    function modeLabel(m: string) {
        return modes.find(x => x.value === m)?.label ?? m;
    }

    $: selectedQuizObj = quizzes.find(q => q.id === selectedQuizId);

    const statusMap: Record<string, { label: string; cls: string }> = {
        in_progress: { label: 'Jonli',         cls: 'badge-active'    },
        waiting:     { label: 'Kutilmoqda',    cls: 'badge-waiting'   },
        completed:   { label: 'Tugagan',       cls: 'badge-completed' },
        abandoned:   { label: 'Bekor',         cls: 'badge-danger'    },
        paused:      { label: 'To\'xtatilgan', cls: 'badge-waiting'   },
    };
</script>

<svelte:head><title>Viktorina — Cognita.uz</title></svelte:head>

<!-- ── Page Header ─────────────────────────────────────────────────────────── -->
<div class="page-header animate-fade">
    <div>
        <h1>Viktorina — Jonli musobaqa</h1>
        <p class="sub">Sinf bilan real vaqtda o'ynang va natijalarni kuzating</p>
    </div>
</div>

<!-- ── Quick Stats ────────────────────────────────────────────────────────── -->
<div class="stats-strip animate-slide delay-1">
    <div class="stat-card">
        <div class="stat-icon" style="background: rgba(34,197,94,0.12); color: var(--success);">📅</div>
        <div class="stat-body">
            <span class="stat-value">{todaySessions}</span>
            <span class="stat-label">Bugungi sessiyalar</span>
        </div>
    </div>
    <div class="stat-card">
        <div class="stat-icon" style="background: rgba(99,102,241,0.12); color: var(--primary);">📊</div>
        <div class="stat-body">
            <span class="stat-value">{totalSessions}</span>
            <span class="stat-label">Jami sessiyalar</span>
        </div>
    </div>
    <div class="stat-card">
        <div class="stat-icon" style="background: rgba(139,92,246,0.12); color: var(--accent);">👥</div>
        <div class="stat-body">
            <span class="stat-value">{totalPlayers}</span>
            <span class="stat-label">Jami o'yinchilar</span>
        </div>
    </div>
    <div class="stat-card stat-card-wide">
        <div class="stat-icon" style="background: rgba(245,158,11,0.12); color: var(--warning);">🏅</div>
        <div class="stat-body">
            <span class="stat-value stat-value-sm">{mostPlayedQuiz}</span>
            <span class="stat-label">Eng ko'p o'ynalgan</span>
        </div>
    </div>
</div>

<!-- ── Active Session Banner ──────────────────────────────────────────────── -->
{#if activeSession}
    <div class="active-session-banner animate-slide">
        <div class="asb-left">
            <div class="asb-live-dot"></div>
            <div>
                <div class="asb-label">Faol sessiya</div>
                <div class="asb-quiz">{activeSession.quizTitle}</div>
            </div>
        </div>
        <div class="asb-pin-block">
            <span class="asb-pin-label">PIN</span>
            <span class="asb-pin">{activeSession.pin}</span>
        </div>
        <div class="asb-right">
            <a href="/game/host/{activeSession.pin}" class="asb-host-btn">
                📺 Host ekranini oching
            </a>
            <button class="asb-close" on:click={() => (activeSession = null)}>✕</button>
        </div>
    </div>
{/if}

<!-- ── Launch Card ────────────────────────────────────────────────────────── -->
<div class="launch-card animate-slide delay-2">
    <div class="launch-card-header">
        <div class="launch-icon">▶</div>
        <div>
            <h2>Yangi viktorina boshlash</h2>
            <p class="sub">Quiz tanlang, rejimni belgilang va o'yinni ishga tushiring</p>
        </div>
    </div>

    <div class="launch-body">
        <!-- Quiz selector -->
        <div class="field-group">
            <label class="field-label" for="quiz-select">Quiz tanlang</label>
            {#if loadingQuizzes}
                <div class="skeleton" style="height:56px; border-radius: var(--radius-sm);"></div>
            {:else if quizzes.length === 0}
                <div class="empty-hint">
                    Quizlar topilmadi. <a href="/dashboard/quizzes/new">Yangi quiz yarating</a>
                </div>
            {:else}
                <div class="quiz-selector">
                    <select id="quiz-select" class="select-hidden" bind:value={selectedQuizId}>
                        {#each quizzes as q}
                            <option value={q.id}>{q.title}</option>
                        {/each}
                    </select>
                    <div class="quiz-display">
                        {#if selectedQuizObj}
                            <div class="quiz-dot" style="background: {modeColors[gameMode] ?? 'var(--primary)'}"></div>
                            <div class="quiz-info">
                                <span class="quiz-title-display">{selectedQuizObj.title}</span>
                                {#if selectedQuizObj.total_questions}
                                    <span class="quiz-q-chip">{selectedQuizObj.total_questions} savol</span>
                                {/if}
                            </div>
                            <span class="quiz-chevron">▾</span>
                        {/if}
                    </div>
                    <select class="quiz-overlay-select" bind:value={selectedQuizId}>
                        {#each quizzes as q}
                            <option value={q.id}>{q.title}{q.total_questions ? ` (${q.total_questions} savol)` : ''}</option>
                        {/each}
                    </select>
                </div>
            {/if}
        </div>

        <!-- Game mode cards -->
        <div class="field-group">
            <span class="field-label">O'yin rejimi</span>
            <div class="mode-grid">
                {#each modes as m}
                    <button
                        class="mode-card"
                        class:active={gameMode === m.value}
                        on:click={() => (gameMode = m.value)}
                        type="button"
                        style="--mode-color: {modeColors[m.value]}"
                    >
                        <span class="mode-icon">{m.icon}</span>
                        <span class="mode-name">{m.label}</span>
                        <span class="mode-desc">{m.desc}</span>
                        {#if gameMode === m.value}
                            <span class="mode-check">✓</span>
                        {/if}
                    </button>
                {/each}
            </div>
        </div>

        <!-- Settings accordion -->
        <div class="accordion">
            <button
                class="accordion-trigger"
                on:click={() => (accordionOpen = !accordionOpen)}
                type="button"
            >
                <span class="accordion-icon">⚙️</span>
                <span>Qo'shimcha sozlamalar</span>
                <span class="accordion-arrow" class:open={accordionOpen}>▾</span>
            </button>
            {#if accordionOpen}
                <div class="accordion-body">
                    <div class="toggles">
                        <label class="toggle-row">
                            <span class="toggle-label">
                                <span class="toggle-icon">🔀</span>
                                Savollarni aralashtirish
                            </span>
                            <div class="toggle-wrap">
                                <input type="checkbox" class="toggle-input" bind:checked={shuffleQuestions} id="tog-q" />
                                <label class="toggle-knob" for="tog-q"></label>
                            </div>
                        </label>
                        <label class="toggle-row">
                            <span class="toggle-label">
                                <span class="toggle-icon">🔄</span>
                                Javoblarni aralashtirish
                            </span>
                            <div class="toggle-wrap">
                                <input type="checkbox" class="toggle-input" bind:checked={shuffleAnswers} id="tog-a" />
                                <label class="toggle-knob" for="tog-a"></label>
                            </div>
                        </label>
                        <label class="toggle-row">
                            <span class="toggle-label">
                                <span class="toggle-icon">🏆</span>
                                Reyting jadvalini ko'rsatish
                            </span>
                            <div class="toggle-wrap">
                                <input type="checkbox" class="toggle-input" bind:checked={showLeaderboard} id="tog-l" />
                                <label class="toggle-knob" for="tog-l"></label>
                            </div>
                        </label>
                        <label class="toggle-row">
                            <span class="toggle-label">
                                <span class="toggle-icon">✅</span>
                                To'g'ri javobni ko'rsatish
                            </span>
                            <div class="toggle-wrap">
                                <input type="checkbox" class="toggle-input" bind:checked={showCorrectAnswer} id="tog-c" />
                                <label class="toggle-knob" for="tog-c"></label>
                            </div>
                        </label>
                        <label class="toggle-row">
                            <span class="toggle-label">
                                <span class="toggle-icon">🎵</span>
                                Musiqa
                            </span>
                            <div class="toggle-wrap">
                                <input type="checkbox" class="toggle-input" bind:checked={music} id="tog-m" />
                                <label class="toggle-knob" for="tog-m"></label>
                            </div>
                        </label>
                    </div>

                    <!-- Time per question -->
                    <div class="time-select-row">
                        <span class="toggle-label">
                            <span class="toggle-icon">⏱</span>
                            Savol uchun vaqt
                        </span>
                        <select class="time-select" bind:value={timePerQuestion}>
                            <option value={10}>10 soniya</option>
                            <option value={20}>20 soniya</option>
                            <option value={30}>30 soniya</option>
                            <option value={60}>60 soniya</option>
                            <option value={120}>120 soniya</option>
                        </select>
                    </div>
                </div>
            {/if}
        </div>

        <!-- Launch button -->
        <button
            class="launch-btn"
            on:click={launchGame}
            disabled={launching || loadingQuizzes || quizzes.length === 0}
        >
            {#if launching}
                <span class="spinner"></span> Yaratilmoqda...
            {:else}
                ▶ Jonli o'yin boshlash
            {/if}
        </button>
    </div>
</div>

<!-- ── Recent Sessions ────────────────────────────────────────────────────── -->
<div class="section animate-slide delay-3">
    <div class="section-header">
        <h2 class="section-title">So'nggi sessiyalar</h2>
        <button class="btn-refresh" on:click={loadHistory} disabled={loadingHistory}>
            {loadingHistory ? '...' : '↻ Yangilash'}
        </button>
    </div>

    {#if loadingHistory}
        <div class="table-skeleton">
            {#each Array(4) as _}
                <div class="skeleton" style="height:52px; border-radius: var(--radius-sm);"></div>
            {/each}
        </div>
    {:else if history.length === 0}
        <div class="empty-state">
            <div class="empty-icon">🕹️</div>
            <p>Hali hech qanday o'yin o'tkazilmagan</p>
            <span>Yuqoridagi tugma orqali birinchi o'yiningizni boshlang!</span>
        </div>
    {:else}
        <div class="table-wrap">
            <table class="history-table">
                <thead>
                    <tr>
                        <th>Quiz nomi</th>
                        <th>Rejim</th>
                        <th>PIN</th>
                        <th>O'yinchilar</th>
                        <th>Davomiyligi</th>
                        <th>Holat</th>
                        <th>Sana</th>
                        <th>Amallar</th>
                    </tr>
                </thead>
                <tbody>
                    {#each history as row (row.id)}
                        <tr>
                            <td class="td-title">{row.quiz_title || '—'}</td>
                            <td>
                                <span class="mode-pill" style="background: {modeColors[row.game_mode] ?? '#6366f1'}20; color: {modeColors[row.game_mode] ?? '#6366f1'}">
                                    {modeLabel(row.game_mode)}
                                </span>
                            </td>
                            <td>
                                <code class="pin-code">{row.pin}</code>
                            </td>
                            <td class="td-center">{row.player_count}</td>
                            <td class="td-dur">{calcDuration(row)}</td>
                            <td>
                                <span class="badge {statusMap[row.status]?.cls ?? 'badge-completed'}" class:badge-pulse={row.status === 'in_progress'}>
                                    {#if row.status === 'in_progress'}<span class="pulse-dot"></span>{/if}
                                    {statusMap[row.status]?.label ?? row.status}
                                </span>
                            </td>
                            <td class="td-date">{fmtDate(row.created_at)}</td>
                            <td class="td-actions">
                                {#if row.status === 'in_progress'}
                                    <a href="/game/host/{row.pin}" class="action-btn btn-view">
                                        📺 Host ekran
                                    </a>
                                {:else if row.status === 'completed'}
                                    <div style="display:flex; gap:5px; flex-wrap:wrap;">
                                        <a href="/dashboard/reports?room={row.id}" class="action-btn btn-results">
                                            📊 Natijalar
                                        </a>
                                        <button
                                            class="action-btn btn-restart"
                                            on:click={() => restartSession(row)}
                                            title="Qayta boshlash"
                                        >
                                            🔁 Qayta
                                        </button>
                                    </div>
                                {/if}
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
    {/if}
</div>

<!-- ── Toast ──────────────────────────────────────────────────────────────── -->
{#if toast}
    <div class="toast {toast.kind === 'error' ? 'error' : 'success'} animate-slide">
        {toast.kind === 'error' ? '⚠️' : '✅'} {toast.msg}
    </div>
{/if}

<style>
    /* ── Header ── */
    .page-header {
        display: flex; align-items: flex-start; justify-content: space-between;
        margin-bottom: 22px; gap: 16px;
    }
    h1 { font-size: 1.75rem; font-weight: 800; color: var(--text); line-height: 1.2; }
    h2 { font-size: 1.15rem; font-weight: 800; color: var(--text); }
    .sub { font-size: 0.88rem; color: var(--text3); margin-top: 3px; }

    /* ── Stats ── */
    .stats-strip {
        display: flex; gap: 14px; margin-bottom: 22px; flex-wrap: wrap;
    }
    .stat-card {
        background: white; border-radius: var(--radius); padding: 14px 18px;
        box-shadow: var(--shadow-sm); display: flex; align-items: center; gap: 13px;
        flex: 1; min-width: 150px; border: 1.5px solid transparent;
        transition: var(--transition);
    }
    .stat-card-wide { min-width: 200px; flex: 1.5; }
    .stat-card:hover { border-color: var(--primary-light); transform: translateY(-2px); box-shadow: var(--shadow); }
    .stat-icon {
        width: 44px; height: 44px; border-radius: 11px;
        display: flex; align-items: center; justify-content: center;
        font-size: 1.2rem; flex-shrink: 0;
    }
    .stat-body { display: flex; flex-direction: column; min-width: 0; }
    .stat-value { font-size: 1.6rem; font-weight: 800; color: var(--text); line-height: 1; }
    .stat-value-sm { font-size: 0.9rem; font-weight: 700; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .stat-label { font-size: 0.75rem; color: var(--text3); margin-top: 3px; }

    /* ── Active Session Banner ── */
    .active-session-banner {
        background: linear-gradient(135deg, #f0fdf4, #dcfce7);
        border: 2px solid #22c55e;
        border-radius: var(--radius-lg);
        padding: 16px 20px;
        margin-bottom: 20px;
        display: flex;
        align-items: center;
        gap: 16px;
        flex-wrap: wrap;
        box-shadow: 0 4px 16px rgba(34,197,94,0.18);
    }
    .asb-left { display: flex; align-items: center; gap: 12px; flex: 1; min-width: 160px; }
    .asb-live-dot {
        width: 14px; height: 14px; border-radius: 50%;
        background: #22c55e; flex-shrink: 0;
        box-shadow: 0 0 0 4px rgba(34,197,94,0.25);
        animation: livepulse 1.5s ease infinite;
    }
    @keyframes livepulse {
        0%, 100% { box-shadow: 0 0 0 4px rgba(34,197,94,0.25); }
        50% { box-shadow: 0 0 0 8px rgba(34,197,94,0.1); }
    }
    .asb-label { font-size: 0.72rem; font-weight: 700; color: #166534; text-transform: uppercase; letter-spacing: 0.06em; }
    .asb-quiz { font-size: 0.92rem; font-weight: 700; color: var(--text); }
    .asb-pin-block {
        display: flex; flex-direction: column; align-items: center;
        background: white; border-radius: 12px; padding: 8px 20px;
        border: 1.5px solid rgba(34,197,94,0.3); gap: 2px;
    }
    .asb-pin-label { font-size: 0.65rem; font-weight: 800; color: #166534; text-transform: uppercase; letter-spacing: 0.1em; }
    .asb-pin {
        font-family: 'Courier New', monospace; font-size: 1.8rem; font-weight: 900;
        color: #166534; letter-spacing: 0.14em; line-height: 1;
    }
    .asb-right { display: flex; align-items: center; gap: 10px; }
    .asb-host-btn {
        padding: 9px 18px; background: #22c55e; color: white;
        border-radius: 9px; font-size: 0.85rem; font-weight: 700;
        text-decoration: none; transition: var(--transition);
        box-shadow: 0 3px 10px rgba(34,197,94,0.3);
    }
    .asb-host-btn:hover { background: #16a34a; transform: translateY(-1px); }
    .asb-close {
        width: 30px; height: 30px; border-radius: 8px;
        background: rgba(34,197,94,0.15); border: none; cursor: pointer;
        font-size: 0.85rem; color: #166534; transition: var(--transition);
        display: flex; align-items: center; justify-content: center;
    }
    .asb-close:hover { background: rgba(34,197,94,0.3); }

    /* ── Launch Card ── */
    .launch-card {
        background: white; border-radius: var(--radius-lg); margin-bottom: 26px;
        box-shadow: var(--shadow-sm);
        border: 2px solid transparent;
        background-clip: padding-box;
        position: relative;
        overflow: hidden;
    }
    .launch-card::before {
        content: '';
        position: absolute; inset: 0;
        background: linear-gradient(135deg, rgba(99,102,241,0.06), rgba(139,92,246,0.04));
        pointer-events: none;
    }
    .launch-card-header {
        display: flex; align-items: center; gap: 14px;
        padding: 20px 24px 0;
    }
    .launch-icon {
        width: 46px; height: 46px; flex-shrink: 0;
        background: linear-gradient(135deg, var(--primary), var(--accent));
        border-radius: 13px; display: flex; align-items: center; justify-content: center;
        font-size: 1.3rem; color: white;
        box-shadow: 0 4px 14px rgba(99,102,241,0.35);
    }
    .launch-body { padding: 20px 24px 24px; display: flex; flex-direction: column; gap: 20px; }

    /* Field groups */
    .field-group { display: flex; flex-direction: column; gap: 8px; }
    .field-label { font-size: 0.83rem; font-weight: 700; color: var(--text2); letter-spacing: 0.03em; text-transform: uppercase; }

    /* Quiz selector */
    .quiz-selector { position: relative; max-width: 500px; }
    .select-hidden { display: none; }
    .quiz-display {
        display: flex; align-items: center; gap: 10px;
        padding: 10px 14px; border: 1.5px solid var(--border);
        border-radius: var(--radius-sm); background: white;
        pointer-events: none;
    }
    .quiz-dot { width: 11px; height: 11px; border-radius: 50%; flex-shrink: 0; transition: background 0.2s; }
    .quiz-info { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0; }
    .quiz-title-display { font-size: 0.92rem; font-weight: 600; color: var(--text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .quiz-q-chip {
        font-size: 0.72rem; font-weight: 700; padding: 2px 8px;
        background: var(--primary-light); color: var(--primary); border-radius: 99px; white-space: nowrap; flex-shrink: 0;
    }
    .quiz-chevron { color: var(--text3); font-size: 1rem; flex-shrink: 0; }
    .quiz-overlay-select {
        position: absolute; inset: 0; opacity: 0; cursor: pointer; width: 100%;
        font-size: 0.92rem;
    }

    .empty-hint { font-size: 0.88rem; color: var(--text3); }
    .empty-hint a { color: var(--primary); font-weight: 600; }

    /* Mode grid */
    .mode-grid {
        display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px;
    }
    .mode-card {
        position: relative; padding: 14px 12px;
        border: 2px solid var(--border); border-radius: 13px;
        background: white; cursor: pointer; text-align: left;
        display: flex; flex-direction: column; gap: 4px;
        transition: var(--transition); font-family: inherit;
    }
    .mode-card:hover { border-color: var(--mode-color, var(--primary)); background: color-mix(in srgb, var(--mode-color, var(--primary)) 6%, white); }
    .mode-card.active {
        border-color: var(--mode-color, var(--primary));
        background: color-mix(in srgb, var(--mode-color, var(--primary)) 8%, white);
        box-shadow: 0 3px 12px color-mix(in srgb, var(--mode-color, var(--primary)) 30%, transparent);
    }
    .mode-icon { font-size: 1.4rem; }
    .mode-name { font-size: 0.87rem; font-weight: 700; color: var(--text); }
    .mode-desc { font-size: 0.73rem; color: var(--text3); line-height: 1.3; }
    .mode-check {
        position: absolute; top: 8px; right: 9px;
        width: 20px; height: 20px;
        background: var(--mode-color, var(--primary)); color: white; border-radius: 50%;
        font-size: 0.68rem; font-weight: 900;
        display: flex; align-items: center; justify-content: center;
    }

    /* Accordion */
    .accordion {
        border: 1.5px solid var(--border); border-radius: var(--radius); overflow: hidden;
        max-width: 520px;
    }
    .accordion-trigger {
        width: 100%; display: flex; align-items: center; gap: 9px;
        padding: 12px 16px; background: var(--bg2); border: none; cursor: pointer;
        font-size: 0.9rem; font-weight: 700; color: var(--text); text-align: left;
        transition: background 0.15s; font-family: inherit;
    }
    .accordion-trigger:hover { background: var(--primary-light); }
    .accordion-icon { font-size: 1rem; }
    .accordion-arrow {
        margin-left: auto; transition: transform 0.22s; display: inline-block;
        color: var(--text3); font-size: 1.1rem;
    }
    .accordion-arrow.open { transform: rotate(180deg); }
    .accordion-body { padding: 14px 16px 16px; display: flex; flex-direction: column; gap: 10px; border-top: 1px solid var(--border); }

    /* Toggles */
    .toggles { display: flex; flex-direction: column; gap: 9px; }
    .toggle-row {
        display: flex; align-items: center; justify-content: space-between;
        background: var(--bg2); border-radius: var(--radius-sm); padding: 10px 12px;
        cursor: pointer; border: 1.5px solid var(--border); transition: var(--transition);
    }
    .toggle-row:hover { border-color: var(--primary-light); }
    .toggle-label { display: flex; align-items: center; gap: 8px; font-size: 0.88rem; font-weight: 600; color: var(--text); }
    .toggle-icon { font-size: 1rem; }
    .toggle-wrap { position: relative; flex-shrink: 0; }
    .toggle-input { display: none; }
    .toggle-knob {
        display: block; width: 40px; height: 22px;
        background: var(--border); border-radius: 99px; cursor: pointer;
        position: relative; transition: background 0.22s;
    }
    .toggle-knob::after {
        content: ''; position: absolute; top: 3px; left: 3px;
        width: 16px; height: 16px; border-radius: 50%; background: white;
        transition: transform 0.22s; box-shadow: 0 1px 4px rgba(0,0,0,0.2);
    }
    .toggle-input:checked + .toggle-knob { background: var(--primary); }
    .toggle-input:checked + .toggle-knob::after { transform: translateX(18px); }

    /* Time select row */
    .time-select-row {
        display: flex; align-items: center; justify-content: space-between;
        background: var(--bg2); border-radius: var(--radius-sm); padding: 10px 12px;
        border: 1.5px solid var(--border);
    }
    .time-select {
        padding: 5px 10px; border: 1.5px solid var(--border);
        border-radius: 7px; font-size: 0.85rem; background: white;
        color: var(--text); outline: none; cursor: pointer;
        transition: border-color 0.2s;
    }
    .time-select:focus { border-color: var(--primary); }

    /* Launch button */
    .launch-btn {
        width: 100%; padding: 15px;
        background: linear-gradient(135deg, var(--primary), var(--accent));
        color: white; border: none; border-radius: var(--radius);
        font-size: 1.05rem; font-weight: 800; cursor: pointer;
        box-shadow: 0 6px 20px rgba(99,102,241,0.38);
        transition: var(--transition);
        display: flex; align-items: center; justify-content: center; gap: 9px;
        letter-spacing: 0.02em;
    }
    .launch-btn:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 10px 28px rgba(99,102,241,0.48); }
    .launch-btn:active:not(:disabled) { transform: scale(0.98); }
    .launch-btn:disabled { opacity: 0.55; cursor: not-allowed; }

    /* ── Section ── */
    .section { margin-top: 4px; }
    .section-header {
        display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px;
    }
    .section-title { font-size: 1.15rem; font-weight: 800; color: var(--text); }
    .btn-refresh {
        padding: 7px 14px; background: white; border: 1.5px solid var(--border);
        border-radius: var(--radius-sm); font-size: 0.83rem; font-weight: 600;
        color: var(--text2); cursor: pointer; transition: var(--transition);
    }
    .btn-refresh:hover:not(:disabled) { border-color: var(--primary); color: var(--primary); }
    .btn-refresh:disabled { opacity: 0.5; cursor: not-allowed; }

    /* Table */
    .table-skeleton { display: flex; flex-direction: column; gap: 8px; }
    .table-wrap {
        background: white; border-radius: var(--radius-lg); overflow: hidden;
        box-shadow: var(--shadow-sm); border: 1.5px solid var(--border);
    }
    .history-table { width: 100%; border-collapse: collapse; }
    .history-table thead tr { background: var(--bg2); border-bottom: 1.5px solid var(--border); }
    .history-table th {
        padding: 11px 16px; text-align: left;
        font-size: 0.77rem; font-weight: 700; color: var(--text3);
        text-transform: uppercase; letter-spacing: 0.05em; white-space: nowrap;
    }
    .history-table td {
        padding: 12px 16px; font-size: 0.88rem; color: var(--text);
        border-bottom: 1px solid var(--border);
    }
    .history-table tbody tr:last-child td { border-bottom: none; }
    .history-table tbody tr:hover { background: #fafbff; }

    .td-title { font-weight: 600; max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .td-center { text-align: center; font-weight: 700; }
    .td-date { color: var(--text3); font-size: 0.8rem; white-space: nowrap; }
    .td-dur { font-size: 0.82rem; color: var(--text2); font-weight: 600; white-space: nowrap; }
    .td-actions { white-space: nowrap; }

    .pin-code {
        font-family: 'Courier New', monospace; font-size: 1rem; font-weight: 800;
        color: var(--primary); background: var(--primary-light);
        padding: 2px 8px; border-radius: 6px; letter-spacing: 0.12em;
    }
    .mode-pill {
        font-size: 0.75rem; font-weight: 700; padding: 3px 9px; border-radius: 99px;
    }

    /* Badges */
    .badge {
        display: inline-flex; align-items: center; gap: 5px;
        padding: 3px 10px; border-radius: 99px;
        font-size: 0.73rem; font-weight: 700;
    }
    .badge-active    { background: #dcfce7; color: #166534; }
    .badge-waiting   { background: #fef9c3; color: #854d0e; }
    .badge-completed { background: #f1f5f9; color: #475569; }
    .badge-danger    { background: #fee2e2; color: #991b1b; }

    .pulse-dot {
        width: 7px; height: 7px; border-radius: 50%; background: #22c55e;
        animation: dotpulse 1.2s ease infinite; flex-shrink: 0;
    }
    @keyframes dotpulse {
        0%, 100% { opacity: 1; transform: scale(1); }
        50% { opacity: 0.5; transform: scale(0.7); }
    }

    /* Action buttons */
    .action-btn {
        display: inline-flex; align-items: center; gap: 4px;
        padding: 5px 11px; border-radius: 7px;
        font-size: 0.78rem; font-weight: 700;
        text-decoration: none; transition: var(--transition); cursor: pointer; border: none;
        font-family: inherit;
    }
    .btn-view    { background: #eff6ff; color: #1d4ed8; }
    .btn-view:hover { background: #dbeafe; }
    .btn-results { background: #f0fdf4; color: #166534; }
    .btn-results:hover { background: #dcfce7; }
    .btn-restart { background: #fef9c3; color: #854d0e; }
    .btn-restart:hover { background: #fef08a; }

    /* Empty state */
    .empty-state {
        padding: 56px 20px; text-align: center;
        background: white; border-radius: var(--radius-lg);
        box-shadow: var(--shadow-sm);
    }
    .empty-icon { font-size: 3.5rem; margin-bottom: 12px; }
    .empty-state p { font-size: 1rem; font-weight: 700; color: var(--text); margin-bottom: 6px; }
    .empty-state span { font-size: 0.85rem; color: var(--text3); }

    /* Responsive */
    @media (max-width: 768px) {
        .mode-grid { grid-template-columns: repeat(2, 1fr); }
        .stats-strip { flex-direction: column; }
        .table-wrap { overflow-x: auto; }
        .history-table { min-width: 700px; }
        .active-session-banner { flex-wrap: wrap; }
    }
    @media (max-width: 480px) {
        .mode-grid { grid-template-columns: 1fr 1fr; }
        .launch-body { padding: 16px; }
        .launch-card-header { padding: 16px 16px 0; }
        .asb-pin { font-size: 1.4rem; }
    }
</style>
