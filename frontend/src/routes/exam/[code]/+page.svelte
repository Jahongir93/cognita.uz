<script lang="ts">
    import { page } from '$app/stores';
    import { onMount, onDestroy } from 'svelte';
    import { goto } from '$app/navigation';

    const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

    const OPTION_LABELS = ['A', 'B', 'C', 'D', 'E', 'F'];
    const OPTION_COLORS = ['#6366f1', '#8b5cf6', '#ec4899', '#f59e0b', '#22c55e', '#ef4444'];

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

    interface ExamInfo {
        id: string;
        title: string;
        code: string;
        time_limit: number;
        quiz_title: string;
        status: string;
        total_questions?: number;
    }

    interface Results {
        id: string;
        score: number;
        max_score: number;
        percent: number;
    }

    type Phase = 'enter_name' | 'exam_info' | 'loading' | 'taking' | 'submitting' | 'results';

    let phase: Phase = 'enter_name';
    let studentName = '';
    let examInfo: ExamInfo | null = null;
    let questions: Question[] = [];
    let currentIndex = 0;
    let answers: { question_id: string; option_id: string }[] = [];
    let secondsLeft = 0;
    let timerInterval: ReturnType<typeof setInterval> | null = null;
    let results: Results | null = null;
    let errorMsg = '';
    let mounted = false;
    let infoLoading = false;

    const code = $page.params.code;

    $: currentQuestion = questions[currentIndex] ?? null;
    $: selectedOption = currentQuestion
        ? (answers.find(a => a.question_id === currentQuestion.id)?.option_id ?? null)
        : null;
    $: timerColor = secondsLeft > (examInfo ? examInfo.time_limit * 60 * 0.5 : 0)
        ? '#22c55e'
        : secondsLeft > (examInfo ? examInfo.time_limit * 60 * 0.2 : 0)
        ? '#f59e0b'
        : '#ef4444';
    $: timerPct = examInfo ? secondsLeft / (examInfo.time_limit * 60) : 1;

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

    function perfMessage(pct: number): string {
        if (pct >= 90) return "A'lo! Barakalla 🎉";
        if (pct >= 75) return 'Yaxshi natija 👍';
        if (pct >= 60) return 'Qoniqarli ✅';
        return "Ko'proq mashq qiling 📚";
    }

    function startTimer() {
        if (timerInterval) clearInterval(timerInterval);
        timerInterval = setInterval(() => {
            if (secondsLeft <= 0) {
                clearInterval(timerInterval!);
                timerInterval = null;
                submitExam();
            } else {
                secondsLeft -= 1;
            }
        }, 1000);
    }

    onMount(async () => {
        mounted = true;
        infoLoading = true;
        try {
            const res = await fetch(`${BASE_URL}/api/exams/join/${code}`);
            if (!res.ok) {
                const body = await res.json().catch(() => ({ error: "Imtihon topilmadi" }));
                errorMsg = body.error ?? "Imtihon topilmadi yoki faol emas";
                infoLoading = false;
                return;
            }
            examInfo = await res.json();
            if (examInfo && examInfo.status !== 'active') {
                errorMsg = "Bu imtihon hozirda faol emas";
                examInfo = null;
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
        phase = 'exam_info';
    }

    async function handleBeginExam() {
        phase = 'loading';
        try {
            const res = await fetch(`${BASE_URL}/api/exams/take/${code}`);
            if (!res.ok) {
                const body = await res.json().catch(() => ({ error: "Savollarni yuklashda xatolik" }));
                errorMsg = body.error ?? "Savollarni yuklashda xatolik";
                phase = 'enter_name';
                return;
            }
            questions = await res.json();
            currentIndex = 0;
            answers = [];
            secondsLeft = (examInfo?.time_limit ?? 30) * 60;
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
            submitExam();
        }
    }

    async function submitExam() {
        if (phase === 'submitting' || phase === 'results') return;
        if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
        const totalSecs = (examInfo?.time_limit ?? 30) * 60;
        const timeTaken = totalSecs - secondsLeft;
        phase = 'submitting';
        try {
            const res = await fetch(`${BASE_URL}/api/exams/submit/${code}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    student_name: studentName.trim(),
                    answers,
                    time_taken: timeTaken
                })
            });
            if (!res.ok) {
                const body = await res.json().catch(() => ({ error: "Yuborishda xatolik" }));
                errorMsg = body.error ?? "Imtihonni yuborishda xatolik";
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

    function saveResultToLocalStorage(r: Results) {
        try {
            const key = 'gogame_results';
            const prev = JSON.parse(localStorage.getItem(key) ?? '[]');
            prev.unshift({
                id: Date.now().toString(),
                type: 'exam',
                title: examInfo?.title ?? 'Imtihon',
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
        ? (examInfo ? examInfo.time_limit * 60 : 0) - secondsLeft
        : 0;
</script>

<svelte:head><title>Imtihon — Cognita.uz</title></svelte:head>

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
                    <p>Imtihon yuklanmoqda...</p>
                </div>
            {:else if errorMsg}
                <div class="card error-card">
                    <div class="err-icon">⚠️</div>
                    <h2>Xatolik</h2>
                    <p>{errorMsg}</p>
                    <a href="/" class="btn-outline">Bosh sahifaga</a>
                </div>
            {:else if examInfo}
                <div class="card enter-card">
                    <a href="/" class="brand">
                        <img src="/sitelogo.png" alt="Cognita.uz" style="height:28px;width:auto" />
                    </a>

                    <div class="card-header">
                        <div class="hicon">📋</div>
                        <h1>Imtihonga kirish</h1>
                        <p>Ismingizni kiriting va imtihonni boshlang</p>
                    </div>

                    <div class="exam-info-badge">
                        <div class="eib-item">
                            <span class="eib-icon">📚</span>
                            <div>
                                <div class="eib-title">{examInfo.title}</div>
                                <div class="eib-sub">{examInfo.quiz_title}</div>
                            </div>
                        </div>
                        <div class="eib-stats">
                            <span>⏱ {examInfo.time_limit} daqiqa</span>
                            {#if examInfo.total_questions}
                                <span>📝 {examInfo.total_questions} ta savol</span>
                            {/if}
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
                        Imtihonni boshlash →
                        <div class="btn-shine"></div>
                    </button>
                </div>
            {/if}
        </div>

    <!-- ══════════════════════════════════════════════════════════════
         PHASE: exam_info
    ══════════════════════════════════════════════════════════════ -->
    {:else if phase === 'exam_info'}
        <div class="center-wrap">
            <div class="card info-card">
                <a href="/" class="brand">
                    <img src="/sitelogo.png" alt="Cognita.uz" style="height:28px;width:auto" />
                </a>

                <div class="card-header">
                    <div class="hicon">📋</div>
                    <h1>{examInfo?.title}</h1>
                    <p class="quiz-sub">{examInfo?.quiz_title}</p>
                </div>

                <div class="info-grid">
                    <div class="info-item">
                        <span class="info-icon">⏱</span>
                        <div>
                            <div class="info-val">{examInfo?.time_limit} daqiqa</div>
                            <div class="info-lbl">Vaqt chegarasi</div>
                        </div>
                    </div>
                    {#if examInfo?.total_questions}
                        <div class="info-item">
                            <span class="info-icon">📝</span>
                            <div>
                                <div class="info-val">{examInfo.total_questions} ta</div>
                                <div class="info-lbl">Savollar soni</div>
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
                    <div class="rules-title">📋 Qoidalar</div>
                    <ul class="rules-list">
                        <li>Har bir savolga diqqat bilan javob bering</li>
                        <li>Vaqt tugagach imtihon avtomatik yakunlanadi</li>
                        <li>Javobni o'zgartirib bo'lmaydi</li>
                    </ul>
                </div>

                <div class="student-badge">
                    <span>👤</span> {studentName} sifatida kirasiz
                </div>

                <button class="btn-primary" on:click={handleBeginExam}>
                    Boshlash
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
                        <div class="q-num-badge">Savol {currentIndex + 1}</div>

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
                                    {currentIndex < questions.length - 1 ? 'Keyingi →' : 'Yakunlash ✓'}
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
            <!-- Confetti particles -->
            <div class="confetti" aria-hidden="true">
                {#each Array(24) as _, i}
                    <div class="conf-particle" style="
                        left:{(i * 4.2) % 100}%;
                        animation-delay:{(i * 0.13) % 2.4}s;
                        background:{OPTION_COLORS[i % OPTION_COLORS.length]};
                        animation-duration:{1.8 + (i % 5) * 0.3}s;
                    "></div>
                {/each}
            </div>

            <div class="results-card">
                <a href="/" class="brand">
                    <img src="/sitelogo.png" alt="Cognita.uz" style="height:28px;width:auto" />
                </a>

                <div class="score-circle-wrap">
                    <svg class="score-ring" viewBox="0 0 120 120">
                        <circle cx="60" cy="60" r="52" fill="none" stroke="rgba(99,102,241,0.15)" stroke-width="10"/>
                        <circle
                            cx="60" cy="60" r="52" fill="none"
                            stroke={results.percent >= 60 ? '#22c55e' : '#ef4444'}
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

                <button class="btn-primary" on:click={() => goto('/')}>
                    Bosh sahifaga
                </button>
            </div>
        </div>
    {/if}
</div>

<style>
    :global(body) { margin: 0; font-family: 'Segoe UI', system-ui, sans-serif; }
    * { box-sizing: border-box; }

    /* ── Page shell ── */
    .page {
        min-height: 100dvh;
        background: linear-gradient(160deg, #1e1b4b 0%, #312e81 50%, #4c1d95 100%);
        position: relative;
        overflow-x: hidden;
    }

    .bg { position: fixed; inset: 0; pointer-events: none; z-index: 0; }
    .blob { position: absolute; border-radius: 50%; filter: blur(90px); }
    .b1 { width: 500px; height: 500px; background: radial-gradient(#7c3aed, #6366f1); opacity: .22; top: -160px; right: -100px; animation: drift 13s ease-in-out infinite alternate; }
    .b2 { width: 360px; height: 360px; background: radial-gradient(#4f46e5, #312e81); opacity: .18; bottom: -110px; left: -80px; animation: drift 17s ease-in-out infinite alternate-reverse; }
    .b3 { width: 280px; height: 280px; background: radial-gradient(#9333ea, #7c3aed); opacity: .14; top: 50%; left: 50%; transform: translate(-50%,-50%); animation: drift 20s ease-in-out infinite alternate; }
    @keyframes drift { from { transform: translate(0,0); } to { transform: translate(28px, 18px) scale(1.08); } }

    /* ── Centered layout for enter_name / exam_info / loading / error ── */
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
        width: 100%; max-width: 460px;
        background: rgba(255,255,255,0.97);
        border-radius: 28px;
        padding: 36px 32px;
        box-shadow: 0 40px 100px rgba(0,0,0,0.35), 0 0 0 1px rgba(255,255,255,.1);
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
    .error-card p { color: #64748b; font-size: 0.9rem; margin: 0; }

    /* ── Brand ── */
    .brand {
        display: flex; align-items: center; gap: 6px; justify-content: center;
        text-decoration: none; font-weight: 900; font-size: 1rem; color: #6366f1;
    }
    .brand-icon { font-size: 1.3rem; }
    .brand-dot { color: #a78bfa; }

    /* ── Card header ── */
    .card-header { text-align: center; }
    .hicon { font-size: 2.6rem; animation: bounce 2.2s ease-in-out infinite; }
    @keyframes bounce { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-6px)} }
    .card-header h1 { font-size: 1.85rem; font-weight: 900; color: #0f172a; margin: 8px 0 6px; letter-spacing: -.03em; }
    .card-header p { color: #64748b; font-size: .88rem; margin: 0; }
    .quiz-sub { font-size: .82rem; color: #6366f1; font-weight: 600; margin-top: 4px !important; }

    /* ── Exam info badge ── */
    .exam-info-badge {
        background: #f1f5f9;
        border: 1px solid #e2e8f0;
        border-radius: 14px;
        padding: 14px 16px;
        display: flex;
        flex-direction: column;
        gap: 10px;
    }
    .eib-item { display: flex; align-items: center; gap: 10px; }
    .eib-icon { font-size: 1.5rem; flex-shrink: 0; }
    .eib-title { font-size: 0.95rem; font-weight: 700; color: #0f172a; }
    .eib-sub { font-size: 0.78rem; color: #64748b; margin-top: 2px; }
    .eib-stats { display: flex; gap: 12px; flex-wrap: wrap; }
    .eib-stats span {
        background: rgba(99,102,241,0.1);
        color: #4f46e5;
        font-size: 0.8rem;
        font-weight: 700;
        padding: 4px 12px;
        border-radius: 999px;
    }

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
        background: #f8fafc;
        border: 1px solid #e2e8f0;
        border-radius: 12px;
        padding: 12px;
    }
    .info-icon { font-size: 1.5rem; flex-shrink: 0; }
    .info-val { font-size: 0.9rem; font-weight: 700; color: #0f172a; }
    .info-lbl { font-size: 0.72rem; color: #94a3b8; margin-top: 2px; }

    /* ── Rules ── */
    .rules-box {
        background: linear-gradient(135deg, rgba(99,102,241,0.06), rgba(139,92,246,0.04));
        border: 1px solid rgba(99,102,241,0.2);
        border-radius: 14px;
        padding: 14px 16px;
    }
    .rules-title { font-size: 0.82rem; font-weight: 700; color: #4f46e5; text-transform: uppercase; letter-spacing: .05em; margin-bottom: 8px; }
    .rules-list { margin: 0; padding: 0 0 0 20px; display: flex; flex-direction: column; gap: 5px; }
    .rules-list li { font-size: 0.85rem; color: #475569; line-height: 1.4; }

    /* ── Student badge ── */
    .student-badge {
        text-align: center;
        background: rgba(99,102,241,0.08);
        border: 1px solid rgba(99,102,241,0.2);
        border-radius: 10px;
        padding: 10px;
        color: #4f46e5;
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
    .field input:focus { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,.12); }

    /* ── Buttons ── */
    .btn-primary {
        position: relative; overflow: hidden;
        padding: 14px; border: none; border-radius: 13px;
        background: linear-gradient(135deg, #6366f1, #8b5cf6);
        color: white; font-size: 1rem; font-weight: 800;
        cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px;
        transition: transform .2s, box-shadow .2s, opacity .2s;
        box-shadow: 0 6px 20px rgba(99,102,241,.4);
        font-family: inherit;
    }
    .btn-primary:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 10px 28px rgba(99,102,241,.5); }
    .btn-primary:active:not(:disabled) { transform: scale(.98); }
    .btn-primary:disabled { opacity: .45; cursor: not-allowed; }

    .btn-shine {
        position: absolute; inset: 0;
        background: linear-gradient(105deg, transparent 35%, rgba(255,255,255,.25) 50%, transparent 65%);
        transform: translateX(-100%);
        animation: shine 3s ease-in-out 1s infinite;
    }
    @keyframes shine { 0%,60%{transform:translateX(-100%)} 80%,100%{transform:translateX(200%)} }

    .btn-ghost {
        padding: 11px; border: 2px solid #e2e8f0; border-radius: 12px;
        background: white; color: #64748b; font-size: 0.92rem; font-weight: 600;
        cursor: pointer; transition: border-color .2s, color .2s;
        font-family: inherit;
    }
    .btn-ghost:hover { border-color: #6366f1; color: #6366f1; }

    .btn-outline {
        display: inline-block;
        padding: 10px 22px; border: 2px solid #6366f1; border-radius: 12px;
        background: white; color: #6366f1; font-size: 0.9rem; font-weight: 700;
        text-decoration: none; transition: background .2s, color .2s;
        font-family: inherit; cursor: pointer;
    }
    .btn-outline:hover { background: #6366f1; color: white; }

    /* ── Spinners ── */
    .spin-lg {
        width: 44px; height: 44px;
        border: 4px solid rgba(99,102,241,.2);
        border-top-color: #6366f1;
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

    /* Top bar */
    .top-bar {
        display: flex;
        align-items: center;
        gap: 14px;
        padding: 12px 20px;
        background: rgba(0,0,0,0.35);
        backdrop-filter: blur(8px);
        border-bottom: 1px solid rgba(255,255,255,0.1);
        flex-shrink: 0;
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
        background: linear-gradient(90deg, #6366f1, #8b5cf6);
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
        color: rgba(255,255,255,0.8);
        background: rgba(255,255,255,0.1);
        padding: 5px 12px;
        border-radius: 999px;
        flex-shrink: 0;
        max-width: 140px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    /* Question area */
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
        box-shadow: 0 32px 80px rgba(0,0,0,0.4);
        display: flex;
        flex-direction: column;
        gap: 20px;
        animation: cardIn 0.35s cubic-bezier(0.34,1.2,0.64,1);
    }
    @keyframes cardIn {
        from { opacity: 0; transform: translateY(20px) scale(0.97); }
        to   { opacity: 1; transform: none; }
    }

    .q-num-badge {
        display: inline-flex;
        align-items: center;
        background: linear-gradient(135deg, #6366f1, #8b5cf6);
        color: white;
        font-size: 0.78rem;
        font-weight: 800;
        padding: 4px 14px;
        border-radius: 999px;
        align-self: flex-start;
        letter-spacing: .03em;
    }

    .q-media-wrap { text-align: center; }
    .q-media {
        max-width: 100%; max-height: 200px;
        border-radius: 12px; object-fit: contain;
    }

    .q-text {
        font-size: clamp(1.05rem, 2.5vw, 1.5rem);
        font-weight: 700;
        color: #0f172a;
        line-height: 1.45;
        margin: 0;
        text-align: center;
    }

    .options-list {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

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
        background: rgba(0,0,0,0.02);
        box-shadow: 0 0 0 3px color-mix(in srgb, var(--oc) 18%, transparent);
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
    .opt-text {
        flex: 1;
        font-size: 0.97rem;
        font-weight: 600;
        color: #1e293b;
        line-height: 1.35;
    }
    .opt-check {
        color: #22c55e;
        font-size: 1.2rem;
        font-weight: 900;
        flex-shrink: 0;
    }

    /* Q footer */
    .q-footer { display: flex; justify-content: flex-end; }

    .btn-next {
        padding: 13px 28px;
        background: linear-gradient(135deg, #6366f1, #8b5cf6);
        color: white; border: none; border-radius: 12px;
        font-size: 1rem; font-weight: 800; cursor: pointer;
        transition: transform .15s, box-shadow .15s, opacity .2s;
        box-shadow: 0 6px 18px rgba(99,102,241,.35);
        font-family: inherit;
    }
    .btn-next:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 10px 24px rgba(99,102,241,.45); }
    .btn-next:disabled { opacity: .35; cursor: not-allowed; }

    .submitting-pill {
        display: flex; align-items: center; gap: 8px;
        background: #6366f1; color: white;
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

    /* Confetti */
    .confetti {
        position: fixed;
        inset: 0;
        pointer-events: none;
        z-index: 0;
    }
    .conf-particle {
        position: absolute;
        top: -12px;
        width: 10px;
        height: 10px;
        border-radius: 2px;
        animation: confettiFall linear infinite;
    }
    @keyframes confettiFall {
        0%   { transform: translateY(-12px) rotate(0deg);   opacity: 1; }
        80%  { opacity: 1; }
        100% { transform: translateY(110vh) rotate(720deg); opacity: 0; }
    }

    .results-card {
        position: relative; z-index: 1;
        width: 100%; max-width: 460px;
        background: rgba(255,255,255,0.97);
        border-radius: 28px;
        padding: 36px 32px;
        box-shadow: 0 40px 100px rgba(0,0,0,0.4);
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 20px;
        animation: cardIn .5s cubic-bezier(0.34,1.2,0.64,1);
    }

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
    .score-inner {
        position: relative; z-index: 1;
        text-align: center;
    }
    .score-pct { font-size: 2.4rem; font-weight: 900; color: #0f172a; line-height: 1; }
    .score-frac { font-size: 0.88rem; color: #64748b; font-weight: 600; margin-top: 4px; }

    .perf-message {
        font-size: 1.2rem;
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
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 4px;
        padding: 14px 10px;
        border-radius: 14px;
        text-align: center;
    }
    .stat-correct { background: rgba(34,197,94,0.1); border: 1px solid rgba(34,197,94,0.25); }
    .stat-wrong   { background: rgba(239,68,68,0.1);  border: 1px solid rgba(239,68,68,0.25); }
    .stat-time    {
        grid-column: 1 / -1;
        background: rgba(99,102,241,0.08);
        border: 1px solid rgba(99,102,241,0.2);
        flex-direction: row; gap: 10px;
        justify-content: center;
    }
    .stat-icon { font-size: 1.4rem; }
    .stat-val { font-size: 1.4rem; font-weight: 900; color: #0f172a; }
    .stat-lbl { font-size: 0.72rem; color: #64748b; font-weight: 600; text-transform: uppercase; letter-spacing: .04em; }
    .time-val { font-size: 1rem !important; }

    .results-card .btn-primary { width: 100%; }

    @media (max-width: 480px) {
        .card, .results-card { padding: 28px 20px; border-radius: 20px; }
        .card-header h1 { font-size: 1.6rem; }
        .q-card { padding: 22px 16px; border-radius: 18px; }
        .top-bar { padding: 10px 14px; gap: 10px; }
        .student-name-pill { display: none; }
    }
</style>
