<script lang="ts">
    import { page } from '$app/stores';
    import { onMount, onDestroy } from 'svelte';
    import { goto } from '$app/navigation';

    const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

    const OPTION_LABELS = ['A', 'B', 'C', 'D', 'E', 'F'];
    const OPTION_COLORS = ['#f59e0b', '#d97706', '#b45309', '#92400e', '#78350f', '#ef4444'];

    interface AnswerOption {
        id: string;
        option_text: string;
        media_url?: string;
    }

    interface Question {
        id: string;
        question_text: string;
        media_url?: string;
        media_type?: string;
        points: number;
        options: AnswerOption[];
    }

    interface OlympiadInfo {
        id: string;
        title: string;
        description: string;
        code: string;
        time_limit: number;
        start_time: string;
        end_time: string;
        quiz_title: string;
        status: string;
        participant_count?: number;
        max_participants?: number | null;
        total_questions?: number;
    }

    interface Results {
        id: string;
        score: number;
        max_score: number;
        percent: number;
    }

    type Phase = 'enter_name' | 'olympiad_info' | 'loading' | 'taking' | 'submitting' | 'results';

    let phase: Phase = 'enter_name';
    let studentName = '';
    let olympiadInfo: OlympiadInfo | null = null;
    let questions: Question[] = [];
    let currentIndex = 0;
    let answers: { question_id: string; option_id: string }[] = [];
    let secondsLeft = 0;
    let timerInterval: ReturnType<typeof setInterval> | null = null;
    let results: Results | null = null;
    let errorMsg = '';
    let mounted = false;
    let infoLoading = false;
    let showLeaderboard = false;

    const code = $page.params.code;

    $: currentQuestion = questions[currentIndex] ?? null;
    $: selectedOption = currentQuestion
        ? (answers.find(a => a.question_id === currentQuestion.id)?.option_id ?? null)
        : null;
    $: timerColor = secondsLeft > (olympiadInfo ? olympiadInfo.time_limit * 60 * 0.5 : 0)
        ? '#22c55e'
        : secondsLeft > (olympiadInfo ? olympiadInfo.time_limit * 60 * 0.2 : 0)
        ? '#f59e0b'
        : '#ef4444';
    $: timerPct = olympiadInfo ? secondsLeft / (olympiadInfo.time_limit * 60) : 1;

    function formatTime(secs: number): string {
        const m = Math.floor(secs / 60);
        const s = secs % 60;
        return `${m < 10 ? '0' : ''}${m}:${s < 10 ? '0' : ''}${s}`;
    }

    function formatTimeTaken(secs: number): string {
        const m = Math.floor(secs / 60);
        const s = secs % 60;
        if (m === 0) return `${s} soniya`;
        return `${m} daqiqa ${s} soniya`;
    }

    function formatDate(iso: string): string {
        try {
            return new Date(iso).toLocaleString('uz-UZ', {
                day: '2-digit', month: 'long', year: 'numeric',
                hour: '2-digit', minute: '2-digit'
            });
        } catch {
            return iso;
        }
    }

    function perfMessage(pct: number): string {
        if (pct >= 90) return "A'lo natija! Siz eng yaxshilardansiz 🏆";
        if (pct >= 75) return 'Juda yaxshi! Zo\'r harakat 🥇';
        if (pct >= 60) return 'Yaxshi natija! Davom eting 💪';
        return "Ko'proq mashq qiling va qayta urining 📚";
    }

    function startTimer() {
        if (timerInterval) clearInterval(timerInterval);
        timerInterval = setInterval(() => {
            if (secondsLeft <= 0) {
                clearInterval(timerInterval!);
                timerInterval = null;
                submitOlympiad();
            } else {
                secondsLeft -= 1;
            }
        }, 1000);
    }

    onMount(async () => {
        mounted = true;
        infoLoading = true;
        try {
            const res = await fetch(`${BASE_URL}/api/olympiads/join/${code}`);
            if (!res.ok) {
                const body = await res.json().catch(() => ({ error: 'Olimpiada topilmadi' }));
                errorMsg = body.error ?? 'Olimpiada topilmadi yoki faol emas';
                infoLoading = false;
                return;
            }
            olympiadInfo = await res.json();
            if (olympiadInfo && olympiadInfo.status !== 'active') {
                if (olympiadInfo.status === 'upcoming') {
                    errorMsg = "Bu olimpiada hali boshlanmagan";
                } else if (olympiadInfo.status === 'completed') {
                    errorMsg = "Bu olimpiada yakunlangan";
                } else {
                    errorMsg = "Bu olimpiada hozirda faol emas";
                }
                olympiadInfo = null;
            }
        } catch {
            errorMsg = "Server bilan ulanishda xatolik";
        } finally {
            infoLoading = false;
        }
    });

    onDestroy(() => {
        if (timerInterval) clearInterval(timerInterval);
    });

    async function handleStart() {
        if (!studentName.trim()) return;
        phase = 'olympiad_info';
    }

    async function handleBeginOlympiad() {
        phase = 'loading';
        try {
            const res = await fetch(`${BASE_URL}/api/olympiads/take/${code}`);
            if (!res.ok) {
                const body = await res.json().catch(() => ({ error: 'Savollarni yuklashda xatolik' }));
                errorMsg = body.error ?? 'Savollarni yuklashda xatolik';
                phase = 'enter_name';
                return;
            }
            questions = await res.json();
            currentIndex = 0;
            answers = [];
            secondsLeft = (olympiadInfo?.time_limit ?? 60) * 60;
            phase = 'taking';
            startTimer();
        } catch {
            errorMsg = "Server bilan ulanishda xatolik";
            phase = 'enter_name';
        }
    }

    function selectOption(questionId: string, optionId: string) {
        const idx = answers.findIndex(a => a.question_id === questionId);
        if (idx === -1) {
            answers = [...answers, { question_id: questionId, option_id: optionId }];
        } else {
            answers = answers.map(a => a.question_id === questionId ? { ...a, option_id: optionId } : a);
        }
    }

    function goNext() {
        if (currentIndex < questions.length - 1) {
            currentIndex += 1;
        } else {
            submitOlympiad();
        }
    }

    async function submitOlympiad() {
        if (phase === 'submitting' || phase === 'results') return;
        if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
        const totalSecs = (olympiadInfo?.time_limit ?? 60) * 60;
        const timeTaken = totalSecs - secondsLeft;
        phase = 'submitting';
        try {
            const res = await fetch(`${BASE_URL}/api/olympiads/submit/${code}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    student_name: studentName.trim(),
                    answers,
                    time_taken: timeTaken
                })
            });
            if (!res.ok) {
                const body = await res.json().catch(() => ({ error: 'Yuborishda xatolik' }));
                errorMsg = body.error ?? 'Olimpiadani yuborishda xatolik';
                phase = 'taking';
                startTimer();
                return;
            }
            results = await res.json();
            phase = 'results';
            if (results) saveResultToLocalStorage(results);
        } catch {
            errorMsg = "Yuborishda tarmoq xatosi";
            phase = 'taking';
            startTimer();
        }
    }

    function saveResultToLocalStorage(r: { score: number; max_score: number; percent: number }) {
        try {
            const key = 'gogame_results';
            const prev = JSON.parse(localStorage.getItem(key) ?? '[]');
            prev.unshift({
                id: Date.now().toString(),
                type: 'olympiad',
                title: olympiadInfo?.title ?? 'Olimpiada',
                score: r.score,
                maxScore: r.max_score,
                percent: r.percent,
                timeTaken: 0,
                date: new Date().toISOString(),
                code: $page.params.code
            });
            localStorage.setItem(key, JSON.stringify(prev.slice(0, 100)));
        } catch {}
    }

    $: correctCount = results
        ? Math.round((results.percent / 100) * (results.max_score > 0 ? results.max_score : questions.length))
        : 0;
    $: wrongCount = questions.length - correctCount;
    $: timeTaken = results
        ? (olympiadInfo ? olympiadInfo.time_limit * 60 : 0) - secondsLeft
        : 0;
</script>

<svelte:head><title>Olimpiada — Cognita.uz</title></svelte:head>

<div class="page" class:mounted>
    <div class="bg" aria-hidden="true">
        <div class="blob b1"></div>
        <div class="blob b2"></div>
        <div class="blob b3"></div>
    </div>

    <!-- ══════════════════════════════════════════════════════════════
         PHASE: enter_name
    ══════════════════════════════════════════════════════════════ -->
    {#if phase === 'enter_name'}
        <div class="center-wrap">
            {#if infoLoading}
                <div class="card loading-card">
                    <div class="spin-lg"></div>
                    <p>Olimpiada yuklanmoqda...</p>
                </div>
            {:else if errorMsg}
                <div class="card error-card">
                    <div class="err-icon">⚠️</div>
                    <h2>Xatolik</h2>
                    <p>{errorMsg}</p>
                    <a href="/" class="btn-outline">Bosh sahifaga</a>
                </div>
            {:else if olympiadInfo}
                <div class="card enter-card">
                    <a href="/" class="brand">
                        <img src="/sitelogo.png" alt="Cognita.uz" style="height:28px;width:auto" />
                    </a>

                    <div class="card-header">
                        <div class="hicon trophy-spin">🏆</div>
                        <h1>Olimpiadaga kirish</h1>
                        <p>Ro'yxatdan o'tish uchun ismingizni kiriting</p>
                    </div>

                    <div class="olympiad-banner">
                        <div class="ob-title">{olympiadInfo.title}</div>
                        {#if olympiadInfo.description}
                            <div class="ob-desc">{olympiadInfo.description}</div>
                        {/if}
                        <div class="ob-meta">
                            <span class="ob-tag">📚 {olympiadInfo.quiz_title}</span>
                            {#if olympiadInfo.participant_count !== undefined}
                                <span class="ob-tag participants">
                                    👥 {olympiadInfo.participant_count} ishtirokchi
                                    {#if olympiadInfo.max_participants}/ {olympiadInfo.max_participants}{/if}
                                </span>
                            {/if}
                        </div>
                        <div class="ob-times">
                            <div class="ob-time-item">
                                <span class="ob-time-lbl">Boshlanish</span>
                                <span class="ob-time-val">{formatDate(olympiadInfo.start_time)}</span>
                            </div>
                            <div class="ob-time-item">
                                <span class="ob-time-lbl">Tugash</span>
                                <span class="ob-time-val">{formatDate(olympiadInfo.end_time)}</span>
                            </div>
                        </div>
                    </div>

                    <div class="field">
                        <label for="sname">👤 Ismingiz</label>
                        <input
                            id="sname"
                            type="text"
                            placeholder="Ism Familiya"
                            bind:value={studentName}
                            on:keydown={(e) => e.key === 'Enter' && studentName.trim() && handleStart()}
                            autocomplete="name"
                            maxlength="80"
                        />
                    </div>

                    <button
                        class="btn-primary"
                        disabled={!studentName.trim()}
                        on:click={handleStart}
                    >
                        🏆 Olimpiadaga kirish
                        <div class="btn-shine"></div>
                    </button>
                </div>
            {/if}
        </div>

    <!-- ══════════════════════════════════════════════════════════════
         PHASE: olympiad_info
    ══════════════════════════════════════════════════════════════ -->
    {:else if phase === 'olympiad_info'}
        <div class="center-wrap">
            <div class="card info-card">
                <a href="/" class="brand">
                    <img src="/sitelogo.png" alt="Cognita.uz" style="height:28px;width:auto" />
                </a>

                <div class="card-header">
                    <div class="hicon trophy-spin">🏆</div>
                    <h1>{olympiadInfo?.title}</h1>
                    {#if olympiadInfo?.description}
                        <p class="quiz-sub">{olympiadInfo.description}</p>
                    {/if}
                </div>

                <div class="info-grid">
                    <div class="info-item">
                        <span class="info-icon">⏱</span>
                        <div>
                            <div class="info-val">{olympiadInfo?.time_limit} daqiqa</div>
                            <div class="info-lbl">Vaqt chegarasi</div>
                        </div>
                    </div>
                    {#if olympiadInfo?.total_questions}
                        <div class="info-item">
                            <span class="info-icon">📝</span>
                            <div>
                                <div class="info-val">{olympiadInfo.total_questions} ta</div>
                                <div class="info-lbl">Savollar soni</div>
                            </div>
                        </div>
                    {/if}
                    {#if olympiadInfo?.participant_count !== undefined}
                        <div class="info-item">
                            <span class="info-icon">👥</span>
                            <div>
                                <div class="info-val">{olympiadInfo.participant_count}</div>
                                <div class="info-lbl">Ishtirokchilar</div>
                            </div>
                        </div>
                    {/if}
                    <div class="info-item">
                        <span class="info-icon">👤</span>
                        <div>
                            <div class="info-val">{studentName}</div>
                            <div class="info-lbl">Ishtirokchi</div>
                        </div>
                    </div>
                </div>

                <div class="rules-box">
                    <div class="rules-title">🏆 Olimpiada qoidalari</div>
                    <ul class="rules-list">
                        <li>Har bir savolga diqqat bilan javob bering</li>
                        <li>Vaqt tugagach olimpiada avtomatik yakunlanadi</li>
                        <li>Javobni o'zgartirib bo'lmaydi</li>
                        <li>Natijalar reytingda ko'rsatiladi</li>
                    </ul>
                </div>

                <div class="student-badge">
                    <span>👤</span> {studentName} sifatida ishtirok etasiz
                </div>

                <button class="btn-primary" on:click={handleBeginOlympiad}>
                    🏆 Boshlash
                    <div class="btn-shine"></div>
                </button>

                <button class="btn-ghost" on:click={() => { phase = 'enter_name'; }}>
                    ← Orqaga
                </button>
            </div>
        </div>

    <!-- ══════════════════════════════════════════════════════════════
         PHASE: loading
    ══════════════════════════════════════════════════════════════ -->
    {:else if phase === 'loading'}
        <div class="center-wrap">
            <div class="card loading-card">
                <div class="spin-lg"></div>
                <p>Savollar yuklanmoqda...</p>
            </div>
        </div>

    <!-- ══════════════════════════════════════════════════════════════
         PHASE: taking
    ══════════════════════════════════════════════════════════════ -->
    {:else if phase === 'taking' || phase === 'submitting'}
        <div class="taking-screen">
            <!-- Top bar -->
            <div class="top-bar">
                <div class="olympiad-badge">🏆 OLIMPIADA</div>

                <div class="progress-wrap">
                    <div class="progress-label">
                        {currentIndex + 1} / {questions.length}
                    </div>
                    <div class="progress-track">
                        <div
                            class="progress-fill"
                            style="width:{((currentIndex + 1) / questions.length) * 100}%"
                        ></div>
                    </div>
                </div>

                <div class="timer-wrap" style="--tc:{timerColor}">
                    <svg class="timer-ring" viewBox="0 0 36 36">
                        <circle cx="18" cy="18" r="15.9" fill="none" stroke="rgba(255,255,255,0.15)" stroke-width="3"/>
                        <circle
                            cx="18" cy="18" r="15.9" fill="none"
                            stroke={timerColor}
                            stroke-width="3"
                            stroke-dasharray="{timerPct * 100} {(1 - timerPct) * 100}"
                            stroke-dashoffset="25"
                            style="transition:stroke-dasharray 0.9s linear,stroke 0.4s"
                        />
                    </svg>
                    <span class="timer-num" style="color:{timerColor}">{formatTime(secondsLeft)}</span>
                </div>

                <div class="student-name-pill">👤 {studentName}</div>
            </div>

            <!-- Question card -->
            <div class="q-area">
                {#if currentQuestion}
                    <div class="q-card">
                        <div class="q-header-row">
                            <div class="q-num-badge">Savol {currentIndex + 1}</div>
                            <div class="q-olympiad-tag">🏆 Olimpiada</div>
                        </div>

                        {#if currentQuestion.media_url}
                            <div class="q-media-wrap">
                                <img src={currentQuestion.media_url} alt="Savol rasmi" class="q-media" />
                            </div>
                        {/if}

                        <p class="q-text">{currentQuestion.question_text}</p>

                        <div class="options-list">
                            {#each currentQuestion.options as opt, i}
                                {@const label = OPTION_LABELS[i] ?? String(i + 1)}
                                {@const color = OPTION_COLORS[i % OPTION_COLORS.length]}
                                {@const isSelected = selectedOption === opt.id}
                                <button
                                    class="opt-btn"
                                    class:selected={isSelected}
                                    style="--oc:{color}"
                                    disabled={phase === 'submitting'}
                                    on:click={() => selectOption(currentQuestion.id, opt.id)}
                                >
                                    <span class="opt-label" style="background:{isSelected ? color : 'rgba(0,0,0,0.08)'}; color:{isSelected ? 'white' : '#374151'}">{label}</span>
                                    <span class="opt-text">{opt.option_text}</span>
                                    {#if isSelected}
                                        <span class="opt-check">✓</span>
                                    {/if}
                                </button>
                            {/each}
                        </div>

                        <div class="q-footer">
                            {#if phase === 'submitting'}
                                <div class="submitting-pill">
                                    <span class="spin-sm"></span> Yuborilmoqda...
                                </div>
                            {:else}
                                <button
                                    class="btn-next"
                                    disabled={!selectedOption}
                                    on:click={goNext}
                                >
                                    {currentIndex < questions.length - 1 ? 'Keyingi →' : '🏆 Yakunlash'}
                                </button>
                            {/if}
                        </div>
                    </div>
                {/if}
            </div>

            {#if errorMsg}
                <div class="bottom-error">{errorMsg}</div>
            {/if}
        </div>

    <!-- ══════════════════════════════════════════════════════════════
         PHASE: results
    ══════════════════════════════════════════════════════════════ -->
    {:else if phase === 'results' && results}
        <div class="results-screen">
            <!-- Gold confetti particles -->
            <div class="confetti" aria-hidden="true">
                {#each Array(28) as _, i}
                    <div class="conf-particle" style="
                        left:{(i * 3.7) % 100}%;
                        animation-delay:{(i * 0.11) % 2.6}s;
                        background:{['#f59e0b','#fbbf24','#d97706','#fcd34d','#92400e','#fef3c7'][i % 6]};
                        animation-duration:{1.7 + (i % 6) * 0.28}s;
                        width:{8 + (i % 4) * 2}px;
                        height:{8 + (i % 4) * 2}px;
                    "></div>
                {/each}
            </div>

            <div class="results-card">
                <a href="/" class="brand">
                    <img src="/sitelogo.png" alt="Cognita.uz" style="height:28px;width:auto" />
                </a>

                <div class="results-heading">
                    <div class="r-trophy">🏆</div>
                    <h2>Natijangiz</h2>
                </div>

                <div class="score-circle-wrap">
                    <svg class="score-ring" viewBox="0 0 120 120">
                        <circle cx="60" cy="60" r="52" fill="none" stroke="rgba(245,158,11,0.15)" stroke-width="10"/>
                        <circle
                            cx="60" cy="60" r="52" fill="none"
                            stroke={results.percent >= 60 ? '#f59e0b' : '#ef4444'}
                            stroke-width="10"
                            stroke-dasharray="{(results.percent / 100) * 327} {327 - (results.percent / 100) * 327}"
                            stroke-dashoffset="81.75"
                            stroke-linecap="round"
                            style="transition:stroke-dasharray 1.2s cubic-bezier(0.4,0,0.2,1)"
                        />
                    </svg>
                    <div class="score-inner">
                        <div class="score-pct">{Math.round(results.percent)}%</div>
                        <div class="score-frac">{results.score}/{results.max_score}</div>
                    </div>
                </div>

                <div class="perf-message">{perfMessage(results.percent)}</div>

                <div class="results-stats">
                    <div class="stat-item stat-correct">
                        <span class="stat-icon">✅</span>
                        <div class="stat-val">{correctCount}</div>
                        <div class="stat-lbl">To'g'ri</div>
                    </div>
                    <div class="stat-item stat-wrong">
                        <span class="stat-icon">❌</span>
                        <div class="stat-val">{wrongCount}</div>
                        <div class="stat-lbl">Noto'g'ri</div>
                    </div>
                    <div class="stat-item stat-time">
                        <span class="stat-icon">⏱</span>
                        <div class="stat-val time-val">{formatTimeTaken(timeTaken)}</div>
                        <div class="stat-lbl">Sarflangan vaqt</div>
                    </div>
                </div>

                <div class="leaderboard-cta">
                    <p class="lb-hint">Reytingda o'z o'rningizni ko'ring</p>
                    <button
                        class="btn-gold"
                        on:click={() => goto(`/olympiad/${code}/leaderboard`)}
                    >
                        🏆 Reytingni ko'rish
                    </button>
                </div>

                <button class="btn-ghost" on:click={() => goto('/')}>
                    Bosh sahifaga
                </button>
            </div>
        </div>
    {/if}
</div>

<style>
    :global(body) { margin: 0; font-family: 'Segoe UI', system-ui, sans-serif; }
    * { box-sizing: border-box; }

    /* ── Page shell — amber/gold gradient ── */
    .page {
        min-height: 100dvh;
        background: linear-gradient(160deg, #78350f 0%, #92400e 50%, #b45309 100%);
        position: relative;
        overflow-x: hidden;
    }

    .bg { position: fixed; inset: 0; pointer-events: none; z-index: 0; }
    .blob { position: absolute; border-radius: 50%; filter: blur(90px); }
    .b1 { width: 500px; height: 500px; background: radial-gradient(#d97706, #b45309); opacity: .28; top: -160px; right: -100px; animation: drift 13s ease-in-out infinite alternate; }
    .b2 { width: 360px; height: 360px; background: radial-gradient(#92400e, #78350f); opacity: .22; bottom: -110px; left: -80px; animation: drift 17s ease-in-out infinite alternate-reverse; }
    .b3 { width: 280px; height: 280px; background: radial-gradient(#f59e0b, #d97706); opacity: .16; top: 45%; left: 48%; transform: translate(-50%,-50%); animation: drift 20s ease-in-out infinite alternate; }
    @keyframes drift { from { transform: translate(0,0); } to { transform: translate(28px, 18px) scale(1.08); } }

    /* ── Centered layout ── */
    .center-wrap {
        position: relative;
        z-index: 10;
        min-height: 100dvh;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 24px 16px;
    }

    .card {
        width: 100%; max-width: 480px;
        background: rgba(255,255,255,0.97);
        border-radius: 28px;
        padding: 36px 32px;
        box-shadow: 0 40px 100px rgba(0,0,0,0.4), 0 0 0 1px rgba(255,255,255,.1);
        display: flex;
        flex-direction: column;
        gap: 20px;
        opacity: 0;
        transform: translateY(24px) scale(0.97);
        transition: opacity .5s ease, transform .5s cubic-bezier(.34,1.2,.64,1);
    }
    .mounted .card { opacity: 1; transform: none; }

    .loading-card {
        align-items: center;
        text-align: center;
        padding: 48px 32px;
        gap: 20px;
        color: #475569;
        font-size: 1rem;
        font-weight: 600;
    }

    .error-card {
        align-items: center;
        text-align: center;
        gap: 14px;
    }
    .err-icon { font-size: 3rem; }
    .error-card h2 { font-size: 1.4rem; font-weight: 800; color: #0f172a; margin: 0; }
    .error-card p  { color: #64748b; font-size: 0.9rem; margin: 0; }

    /* ── Brand ── */
    .brand {
        display: flex; align-items: center; gap: 6px; justify-content: center;
        text-decoration: none; font-weight: 900; font-size: 1rem; color: #d97706;
    }
    .brand-icon { font-size: 1.3rem; }
    .brand-dot  { color: #f59e0b; }

    /* ── Card header ── */
    .card-header { text-align: center; }
    .hicon { font-size: 2.6rem; }
    .trophy-spin { animation: trophyBounce 2.4s ease-in-out infinite; }
    @keyframes trophyBounce {
        0%,100%{ transform: translateY(0) rotate(0deg); }
        30%    { transform: translateY(-8px) rotate(-5deg); }
        60%    { transform: translateY(-4px) rotate(4deg); }
    }
    .card-header h1 { font-size: 1.85rem; font-weight: 900; color: #0f172a; margin: 8px 0 6px; letter-spacing: -.03em; }
    .card-header p { color: #64748b; font-size: .88rem; margin: 0; }
    .quiz-sub { font-size: .82rem; color: #d97706; font-weight: 600; margin-top: 4px !important; }

    /* ── Olympiad banner ── */
    .olympiad-banner {
        background: linear-gradient(135deg, #fffbeb, #fef3c7);
        border: 1px solid #fde68a;
        border-radius: 16px;
        padding: 16px 18px;
        display: flex;
        flex-direction: column;
        gap: 10px;
    }
    .ob-title { font-size: 1.05rem; font-weight: 800; color: #92400e; }
    .ob-desc  { font-size: 0.85rem; color: #78350f; line-height: 1.4; }
    .ob-meta  { display: flex; gap: 8px; flex-wrap: wrap; }
    .ob-tag {
        font-size: 0.78rem; font-weight: 700;
        background: rgba(217,119,6,0.12);
        color: #b45309;
        padding: 4px 12px; border-radius: 999px;
    }
    .ob-tag.participants { background: rgba(245,158,11,0.15); color: #92400e; }
    .ob-times { display: flex; flex-direction: column; gap: 6px; }
    .ob-time-item { display: flex; gap: 8px; align-items: baseline; }
    .ob-time-lbl { font-size: 0.72rem; font-weight: 700; color: #b45309; text-transform: uppercase; letter-spacing: .04em; min-width: 68px; }
    .ob-time-val { font-size: 0.82rem; color: #78350f; font-weight: 600; }

    /* ── Info grid ── */
    .info-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
        gap: 10px;
    }
    .info-item {
        display: flex;
        align-items: center;
        gap: 10px;
        background: #fffbeb;
        border: 1px solid #fde68a;
        border-radius: 12px;
        padding: 12px;
    }
    .info-icon { font-size: 1.5rem; flex-shrink: 0; }
    .info-val  { font-size: 0.9rem; font-weight: 700; color: #0f172a; }
    .info-lbl  { font-size: 0.72rem; color: #94a3b8; margin-top: 2px; }

    /* ── Rules ── */
    .rules-box {
        background: linear-gradient(135deg, rgba(245,158,11,0.08), rgba(217,119,6,0.05));
        border: 1px solid rgba(245,158,11,0.3);
        border-radius: 14px;
        padding: 14px 16px;
    }
    .rules-title { font-size: 0.82rem; font-weight: 700; color: #b45309; text-transform: uppercase; letter-spacing: .05em; margin-bottom: 8px; }
    .rules-list { margin: 0; padding: 0 0 0 20px; display: flex; flex-direction: column; gap: 5px; }
    .rules-list li { font-size: 0.85rem; color: #475569; line-height: 1.4; }

    /* ── Student badge ── */
    .student-badge {
        text-align: center;
        background: rgba(245,158,11,0.1);
        border: 1px solid rgba(245,158,11,0.3);
        border-radius: 10px;
        padding: 10px;
        color: #92400e;
        font-size: 0.88rem;
        font-weight: 600;
    }

    /* ── Field ── */
    .field { display: flex; flex-direction: column; gap: 6px; }
    .field label {
        font-size: .76rem; font-weight: 700; color: #475569;
        text-transform: uppercase; letter-spacing: .06em;
    }
    .field input {
        width: 100%; padding: 13px 16px;
        border: 2px solid #e2e8f0; border-radius: 12px;
        font-size: 1rem; outline: none; background: white; color: #0f172a;
        transition: border-color .2s, box-shadow .2s;
        font-family: inherit;
    }
    .field input:focus { border-color: #f59e0b; box-shadow: 0 0 0 3px rgba(245,158,11,.15); }

    /* ── Buttons ── */
    .btn-primary {
        position: relative; overflow: hidden;
        padding: 14px; border: none; border-radius: 13px;
        background: linear-gradient(135deg, #d97706, #b45309);
        color: white; font-size: 1rem; font-weight: 800;
        cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px;
        transition: transform .2s, box-shadow .2s, opacity .2s;
        box-shadow: 0 6px 20px rgba(217,119,6,.45);
        font-family: inherit;
    }
    .btn-primary:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 10px 28px rgba(217,119,6,.55); }
    .btn-primary:active:not(:disabled) { transform: scale(.98); }
    .btn-primary:disabled { opacity: .45; cursor: not-allowed; }

    .btn-shine {
        position: absolute; inset: 0;
        background: linear-gradient(105deg, transparent 35%, rgba(255,255,255,.25) 50%, transparent 65%);
        transform: translateX(-100%);
        animation: shine 3s ease-in-out 1s infinite;
    }
    @keyframes shine { 0%,60%{transform:translateX(-100%)} 80%,100%{transform:translateX(200%)} }

    .btn-gold {
        padding: 13px 22px; border: none; border-radius: 12px;
        background: linear-gradient(135deg, #f59e0b, #d97706);
        color: white; font-size: 0.97rem; font-weight: 800;
        cursor: pointer; transition: transform .15s, box-shadow .15s;
        box-shadow: 0 6px 18px rgba(245,158,11,.4);
        font-family: inherit; width: 100%;
    }
    .btn-gold:hover { transform: translateY(-2px); box-shadow: 0 10px 24px rgba(245,158,11,.5); }

    .btn-ghost {
        padding: 11px; border: 2px solid #e2e8f0; border-radius: 12px;
        background: white; color: #64748b; font-size: 0.92rem; font-weight: 600;
        cursor: pointer; transition: border-color .2s, color .2s;
        font-family: inherit;
    }
    .btn-ghost:hover { border-color: #f59e0b; color: #d97706; }

    .btn-outline {
        display: inline-block;
        padding: 10px 22px; border: 2px solid #f59e0b; border-radius: 12px;
        background: white; color: #b45309; font-size: 0.9rem; font-weight: 700;
        text-decoration: none; transition: background .2s, color .2s;
        font-family: inherit; cursor: pointer;
    }
    .btn-outline:hover { background: #f59e0b; color: white; }

    /* ── Spinners ── */
    .spin-lg {
        width: 44px; height: 44px;
        border: 4px solid rgba(217,119,6,.2);
        border-top-color: #d97706;
        border-radius: 50%;
        animation: spinA .8s linear infinite;
    }
    .spin-sm {
        width: 16px; height: 16px;
        border: 2px solid rgba(255,255,255,.3);
        border-top-color: white;
        border-radius: 50%;
        animation: spinA .7s linear infinite;
        flex-shrink: 0;
    }
    @keyframes spinA { to { transform: rotate(360deg); } }

    /* ══════════════════════════════════════════════════════════════
       TAKING SCREEN
    ══════════════════════════════════════════════════════════════ */
    .taking-screen {
        position: relative;
        z-index: 10;
        min-height: 100dvh;
        display: flex;
        flex-direction: column;
    }

    .top-bar {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 12px 20px;
        background: rgba(0,0,0,0.4);
        backdrop-filter: blur(8px);
        border-bottom: 1px solid rgba(255,255,255,0.1);
        flex-shrink: 0;
    }

    .olympiad-badge {
        flex-shrink: 0;
        background: linear-gradient(135deg, #f59e0b, #d97706);
        color: white;
        font-size: 0.7rem;
        font-weight: 800;
        padding: 5px 10px;
        border-radius: 8px;
        letter-spacing: .05em;
        white-space: nowrap;
        box-shadow: 0 2px 8px rgba(245,158,11,.4);
    }

    .progress-wrap { flex: 1; min-width: 0; }
    .progress-label { font-size: 0.78rem; font-weight: 700; color: rgba(255,255,255,0.7); margin-bottom: 5px; }
    .progress-track {
        height: 6px;
        background: rgba(255,255,255,0.15);
        border-radius: 999px;
        overflow: hidden;
    }
    .progress-fill {
        height: 100%;
        background: linear-gradient(90deg, #f59e0b, #fbbf24);
        border-radius: 999px;
        transition: width 0.4s cubic-bezier(0.4,0,0.2,1);
    }

    .timer-wrap {
        position: relative;
        width: 56px; height: 56px;
        flex-shrink: 0;
        display: flex; align-items: center; justify-content: center;
    }
    .timer-ring {
        position: absolute; inset: 0;
        transform: rotate(-90deg);
        width: 100%; height: 100%;
    }
    .timer-num {
        position: relative; z-index: 1;
        font-size: 0.78rem; font-weight: 800;
        font-variant-numeric: tabular-nums;
    }

    .student-name-pill {
        font-size: 0.8rem;
        font-weight: 600;
        color: rgba(255,255,255,0.85);
        background: rgba(255,255,255,0.12);
        padding: 5px 12px;
        border-radius: 999px;
        flex-shrink: 0;
        max-width: 140px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .q-area {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 24px 16px;
        overflow-y: auto;
    }

    .q-card {
        width: 100%; max-width: 680px;
        background: white;
        border-radius: 24px;
        padding: 32px 28px;
        box-shadow: 0 32px 80px rgba(0,0,0,0.45);
        display: flex;
        flex-direction: column;
        gap: 20px;
        animation: cardIn 0.35s cubic-bezier(0.34,1.2,0.64,1);
    }
    @keyframes cardIn {
        from { opacity: 0; transform: translateY(20px) scale(0.97); }
        to   { opacity: 1; transform: none; }
    }

    .q-header-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;
    }

    .q-num-badge {
        display: inline-flex;
        align-items: center;
        background: linear-gradient(135deg, #f59e0b, #d97706);
        color: white;
        font-size: 0.78rem;
        font-weight: 800;
        padding: 4px 14px;
        border-radius: 999px;
        letter-spacing: .03em;
    }

    .q-olympiad-tag {
        font-size: 0.75rem;
        font-weight: 700;
        color: #b45309;
        background: #fef3c7;
        padding: 4px 10px;
        border-radius: 999px;
        border: 1px solid #fde68a;
    }

    .q-media-wrap { text-align: center; }
    .q-media { max-width: 100%; max-height: 200px; border-radius: 12px; object-fit: contain; }

    .q-text {
        font-size: clamp(1.05rem, 2.5vw, 1.5rem);
        font-weight: 700;
        color: #0f172a;
        line-height: 1.45;
        margin: 0;
        text-align: center;
    }

    .options-list { display: flex; flex-direction: column; gap: 10px; }

    .opt-btn {
        display: flex;
        align-items: center;
        gap: 14px;
        padding: 14px 18px;
        border: 2px solid #e2e8f0;
        border-radius: 14px;
        background: white;
        cursor: pointer;
        text-align: left;
        transition: border-color .15s, transform .12s, box-shadow .15s;
        font-family: inherit;
        position: relative;
    }
    .opt-btn:hover:not(:disabled):not(.selected) {
        border-color: var(--oc);
        transform: translateX(3px);
        box-shadow: 0 4px 16px rgba(0,0,0,0.08);
    }
    .opt-btn:active:not(:disabled) { transform: scale(0.98); }
    .opt-btn.selected {
        border-color: var(--oc);
        background: rgba(245,158,11,0.03);
        box-shadow: 0 0 0 3px color-mix(in srgb, var(--oc) 15%, transparent);
    }
    .opt-btn:disabled { opacity: 0.6; cursor: not-allowed; }

    .opt-label {
        width: 34px; height: 34px;
        border-radius: 10px;
        display: flex; align-items: center; justify-content: center;
        font-size: 0.88rem; font-weight: 800;
        flex-shrink: 0;
        transition: background .15s, color .15s;
    }
    .opt-text { flex: 1; font-size: 0.97rem; font-weight: 600; color: #1e293b; line-height: 1.35; }
    .opt-check { color: #f59e0b; font-size: 1.2rem; font-weight: 900; flex-shrink: 0; }

    .q-footer { display: flex; justify-content: flex-end; }

    .btn-next {
        padding: 13px 28px;
        background: linear-gradient(135deg, #f59e0b, #d97706);
        color: white; border: none; border-radius: 12px;
        font-size: 1rem; font-weight: 800; cursor: pointer;
        transition: transform .15s, box-shadow .15s, opacity .2s;
        box-shadow: 0 6px 18px rgba(245,158,11,.4);
        font-family: inherit;
    }
    .btn-next:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 10px 24px rgba(245,158,11,.5); }
    .btn-next:disabled { opacity: .35; cursor: not-allowed; }

    .submitting-pill {
        display: flex; align-items: center; gap: 8px;
        background: #d97706; color: white;
        padding: 11px 22px; border-radius: 12px;
        font-size: 0.92rem; font-weight: 700;
    }

    .bottom-error {
        position: fixed;
        bottom: 20px; left: 50%; transform: translateX(-50%);
        background: #ef4444; color: white;
        padding: 10px 20px; border-radius: 10px;
        font-size: 0.88rem; font-weight: 600;
        z-index: 100;
        box-shadow: 0 8px 24px rgba(239,68,68,0.4);
        animation: fadeInUp .3s ease;
    }
    @keyframes fadeInUp { from { opacity: 0; transform: translateX(-50%) translateY(10px); } to { opacity: 1; transform: translateX(-50%) translateY(0); } }

    /* ══════════════════════════════════════════════════════════════
       RESULTS SCREEN
    ══════════════════════════════════════════════════════════════ */
    .results-screen {
        position: relative;
        z-index: 10;
        min-height: 100dvh;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 24px 16px;
        overflow: hidden;
    }

    .confetti {
        position: fixed; inset: 0;
        pointer-events: none; z-index: 0;
    }
    .conf-particle {
        position: absolute; top: -14px;
        border-radius: 3px;
        animation: confettiFall linear infinite;
    }
    @keyframes confettiFall {
        0%   { transform: translateY(-14px) rotate(0deg);   opacity: 1; }
        80%  { opacity: 1; }
        100% { transform: translateY(110vh) rotate(720deg); opacity: 0; }
    }

    .results-card {
        position: relative; z-index: 1;
        width: 100%; max-width: 460px;
        background: rgba(255,255,255,0.97);
        border-radius: 28px;
        padding: 36px 32px;
        box-shadow: 0 40px 100px rgba(0,0,0,0.45);
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 20px;
        animation: cardIn .5s cubic-bezier(0.34,1.2,0.64,1);
    }

    .results-heading {
        text-align: center;
        display: flex; flex-direction: column; align-items: center; gap: 6px;
    }
    .r-trophy { font-size: 3rem; animation: trophyBounce 2.4s ease-in-out infinite; }
    .results-heading h2 { font-size: 1.6rem; font-weight: 900; color: #0f172a; margin: 0; }

    .score-circle-wrap {
        position: relative;
        width: 150px; height: 150px;
        display: flex; align-items: center; justify-content: center;
    }
    .score-ring {
        position: absolute; inset: 0;
        transform: rotate(-90deg);
        width: 100%; height: 100%;
        overflow: visible;
    }
    .score-inner { position: relative; z-index: 1; text-align: center; }
    .score-pct  { font-size: 2.4rem; font-weight: 900; color: #0f172a; line-height: 1; }
    .score-frac { font-size: 0.88rem; color: #64748b; font-weight: 600; margin-top: 4px; }

    .perf-message {
        font-size: 1.1rem;
        font-weight: 800;
        color: #0f172a;
        text-align: center;
    }

    .results-stats {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 10px;
        width: 100%;
    }
    .stat-item {
        display: flex; flex-direction: column;
        align-items: center; gap: 4px;
        padding: 14px 10px; border-radius: 14px; text-align: center;
    }
    .stat-correct { background: rgba(34,197,94,0.1);  border: 1px solid rgba(34,197,94,0.25); }
    .stat-wrong   { background: rgba(239,68,68,0.1);   border: 1px solid rgba(239,68,68,0.25); }
    .stat-time    {
        grid-column: 1 / -1;
        background: rgba(245,158,11,0.1);
        border: 1px solid rgba(245,158,11,0.25);
        flex-direction: row; gap: 10px; justify-content: center;
    }
    .stat-icon { font-size: 1.4rem; }
    .stat-val  { font-size: 1.4rem; font-weight: 900; color: #0f172a; }
    .stat-lbl  { font-size: 0.72rem; color: #64748b; font-weight: 600; text-transform: uppercase; letter-spacing: .04em; }
    .time-val  { font-size: 1rem !important; }

    .leaderboard-cta {
        width: 100%;
        display: flex; flex-direction: column; gap: 10px; align-items: center;
        background: linear-gradient(135deg, #fffbeb, #fef3c7);
        border: 1px solid #fde68a;
        border-radius: 16px;
        padding: 16px;
        text-align: center;
    }
    .lb-hint { font-size: 0.88rem; color: #92400e; font-weight: 600; margin: 0; }
    .results-card .btn-ghost { width: 100%; }

    @media (max-width: 480px) {
        .card, .results-card { padding: 28px 20px; border-radius: 20px; }
        .card-header h1 { font-size: 1.6rem; }
        .q-card { padding: 22px 16px; border-radius: 18px; }
        .top-bar { padding: 10px 12px; gap: 8px; }
        .olympiad-badge { font-size: 0.62rem; padding: 4px 8px; }
        .student-name-pill { display: none; }
    }
</style>
