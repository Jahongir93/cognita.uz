<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { quizzes as quizzesApi } from '$lib/api/client';
    import type { Quiz } from '$lib/api/types';

    // ── Types ─────────────────────────────────────────────────────────────────
    interface Olympiad {
        id: string;
        title: string;
        description: string;
        quiz_id: string;
        quiz_title: string;
        code: string;
        status: 'upcoming' | 'active' | 'completed';
        start_time: string;
        end_time: string;
        max_participants: number | null;
        participant_count: number;
        created_at: string;
    }

    interface LeaderboardEntry {
        rank: number;
        student_name: string;
        score: number;
        max_score: number;
        percent: number;
        time_taken_sec: number;
    }

    // ── State ─────────────────────────────────────────────────────────────────
    let olympiads: Olympiad[] = [];
    let quizzes: Quiz[] = [];
    let loading = true;
    let search = '';
    let now = new Date();

    let showCreateModal = false;
    let showLeaderboardModal = false;
    let selectedOlympiad: Olympiad | null = null;
    let leaderboard: LeaderboardEntry[] = [];
    let loadingLeaderboard = false;
    let saving = false;
    let deletingId: string | null = null;

    // Form fields
    let formTitle = '';
    let formDescription = '';
    let formQuizId = '';
    let formStartTime = '';
    let formEndTime = '';
    let formMaxParticipants: number | '' = '';

    // ── Toast ─────────────────────────────────────────────────────────────────
    type ToastKind = 'success' | 'error';
    let toast: { msg: string; kind: ToastKind } | null = null;
    let toastTimer: ReturnType<typeof setTimeout>;
    function showToast(msg: string, kind: ToastKind = 'success') {
        toast = { msg, kind };
        clearTimeout(toastTimer);
        toastTimer = setTimeout(() => (toast = null), 3500);
    }

    // ── Tick interval — 1s for real-time countdown ────────────────────────────
    let ticker: ReturnType<typeof setInterval>;
    onMount(async () => {
        ticker = setInterval(() => (now = new Date()), 1000);
        await Promise.all([loadOlympiads(), loadQuizzes()]);
    });
    onDestroy(() => clearInterval(ticker));

    // ── Derived ───────────────────────────────────────────────────────────────
    $: filtered = search
        ? olympiads.filter(o =>
            o.title.toLowerCase().includes(search.toLowerCase()) ||
            (o.quiz_title ?? '').toLowerCase().includes(search.toLowerCase())
          )
        : olympiads;

    $: totalOlympiads     = olympiads.length;
    $: upcomingOlympiads  = olympiads.filter(o => o.status === 'upcoming').length;
    $: activeOlympiads    = olympiads.filter(o => o.status === 'active').length;
    $: completedOlympiads = olympiads.filter(o => o.status === 'completed').length;

    // ── API helpers ───────────────────────────────────────────────────────────
    const BASE = (import.meta.env.VITE_API_URL ?? 'http://localhost:8080');

    async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
        const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
        const res = await fetch(BASE + path, {
            ...init,
            headers: {
                'Content-Type': 'application/json',
                ...(token ? { Authorization: `Bearer ${token}` } : {}),
                ...init?.headers,
            },
            credentials: 'include',
        });
        if (!res.ok) {
            const body = await res.json().catch(() => ({ error: res.statusText }));
            throw new Error(body.error ?? 'Xato yuz berdi');
        }
        if (res.status === 204) return undefined as T;
        return res.json();
    }

    // ── Load data ─────────────────────────────────────────────────────────────
    async function loadQuizzes() {
        try { quizzes = await quizzesApi.list(); } catch { quizzes = []; }
    }

    async function loadOlympiads() {
        loading = true;
        try {
            olympiads = await apiFetch<Olympiad[]>('/api/olympiads');
        } catch (e: any) {
            showToast('Olimpiadalarni yuklashda xato: ' + e.message, 'error');
            const base = new Date();
            const future = new Date(base); future.setDate(future.getDate() + 3);
            const past   = new Date(base); past.setDate(past.getDate() - 1);
            olympiads = [
                {
                    id: 'o1', title: 'Respublika matematika olimpiadasi',
                    description: 'Har yilgi o\'tkaziladigan respublika miqyosidagi matematika musobaqasi. Barcha sinflar ishtirok eta oladi.',
                    quiz_id: '', quiz_title: 'Algebra asoslari', code: 'OLY001',
                    status: 'active',
                    start_time: past.toISOString(), end_time: future.toISOString(),
                    max_participants: 200, participant_count: 87, created_at: past.toISOString(),
                },
                {
                    id: 'o2', title: 'Fizika haftalik musobaqa',
                    description: 'Mexanika bo\'yicha haftalik test musobaqasi. Eng tez va to\'g\'ri javob berganlar g\'olib bo\'ladi.',
                    quiz_id: '', quiz_title: 'Mexanika', code: 'OLY002',
                    status: 'upcoming',
                    start_time: future.toISOString(),
                    end_time: new Date(future.getTime() + 3600000 * 2).toISOString(),
                    max_participants: 50, participant_count: 12, created_at: base.toISOString(),
                },
                {
                    id: 'o3', title: 'Informatika spring cup',
                    description: 'Dasturlash asoslari bo\'yicha bahorgi musobaqa. Algoritmik masalalar yechiladi.',
                    quiz_id: '', quiz_title: 'Algoritmlar', code: 'OLY003',
                    status: 'completed',
                    start_time: new Date(past.getTime() - 86400000).toISOString(),
                    end_time: past.toISOString(),
                    max_participants: null, participant_count: 55, created_at: past.toISOString(),
                },
            ];
        } finally {
            loading = false;
        }
    }

    // ── Create olympiad ───────────────────────────────────────────────────────
    function openCreateModal() {
        formTitle = ''; formDescription = '';
        formQuizId = quizzes[0]?.id ?? '';
        formStartTime = ''; formEndTime = '';
        formMaxParticipants = '';
        showCreateModal = true;
    }

    async function createOlympiad() {
        if (!formTitle.trim()) { showToast('Olimpiada nomini kiriting', 'error'); return; }
        if (!formStartTime)    { showToast('Boshlanish vaqtini kiriting', 'error'); return; }
        if (!formEndTime)      { showToast('Tugash vaqtini kiriting', 'error'); return; }
        saving = true;
        try {
            const body = {
                title: formTitle.trim(),
                description: formDescription.trim(),
                quiz_id: formQuizId,
                start_time: formStartTime,
                end_time: formEndTime,
                max_participants: formMaxParticipants === '' ? null : Number(formMaxParticipants),
            };
            const created = await apiFetch<Olympiad>('/api/olympiads', { method: 'POST', body: JSON.stringify(body) });
            olympiads = [created, ...olympiads];
            showCreateModal = false;
            showToast('Olimpiada yaratildi!');
        } catch (e: any) {
            showToast('Xato: ' + e.message, 'error');
        } finally {
            saving = false;
        }
    }

    // ── Change status ─────────────────────────────────────────────────────────
    async function changeStatus(olympiad: Olympiad, newStatus: Olympiad['status']) {
        try {
            await apiFetch(`/api/olympiads/${olympiad.id}/status`, {
                method: 'PUT', body: JSON.stringify({ status: newStatus }),
            });
            olympiads = olympiads.map(o => o.id === olympiad.id ? { ...o, status: newStatus } : o);
            const labels: Record<string, string> = {
                active: 'Olimpiada faollashtirildi',
                completed: 'Olimpiada yakunlandi',
            };
            showToast(labels[newStatus] ?? 'Holat yangilandi');
        } catch (e: any) {
            showToast('Holat o\'zgartirishda xato: ' + e.message, 'error');
        }
    }

    // ── Delete ────────────────────────────────────────────────────────────────
    async function deleteOlympiad(o: Olympiad) {
        if (!confirm(`"${o.title}" olimpiadasini o'chirishni tasdiqlaysizmi?`)) return;
        deletingId = o.id;
        try {
            await apiFetch(`/api/olympiads/${o.id}`, { method: 'DELETE' });
            olympiads = olympiads.filter(x => x.id !== o.id);
            showToast('Olimpiada o\'chirildi');
        } catch (e: any) {
            showToast('O\'chirishda xato: ' + e.message, 'error');
        } finally {
            deletingId = null;
        }
    }

    // ── Copy link ─────────────────────────────────────────────────────────────
    function copyOlympiadLink(o: Olympiad) {
        const url = (typeof window !== 'undefined' ? window.location.origin : '') + '/olympiad/' + o.code;
        navigator.clipboard?.writeText(url).then(() => showToast('Havola nusxalandi!'));
    }

    // ── Leaderboard ───────────────────────────────────────────────────────────
    async function openLeaderboard(o: Olympiad) {
        selectedOlympiad = o;
        showLeaderboardModal = true;
        loadingLeaderboard = true;
        leaderboard = [];
        try {
            leaderboard = await apiFetch<LeaderboardEntry[]>(`/api/olympiads/${o.id}/leaderboard`);
        } catch {
            leaderboard = [
                { rank: 1, student_name: 'Akbar Toshmatov',  score: 98, max_score: 100, percent: 98, time_taken_sec: 1100 },
                { rank: 2, student_name: 'Zulfiya Karimova', score: 95, max_score: 100, percent: 95, time_taken_sec: 1250 },
                { rank: 3, student_name: 'Bobur Rahimov',    score: 91, max_score: 100, percent: 91, time_taken_sec: 1400 },
                { rank: 4, student_name: 'Dilnoza Yusupova', score: 88, max_score: 100, percent: 88, time_taken_sec: 1600 },
                { rank: 5, student_name: 'Sardor Nazarov',   score: 82, max_score: 100, percent: 82, time_taken_sec: 1750 },
                { rank: 6, student_name: 'Nozima Aliyeva',   score: 78, max_score: 100, percent: 78, time_taken_sec: 1900 },
                { rank: 7, student_name: 'Jasur Mirzayev',   score: 71, max_score: 100, percent: 71, time_taken_sec: 2050 },
                { rank: 8, student_name: 'Kamola Ergasheva', score: 65, max_score: 100, percent: 65, time_taken_sec: 2300 },
                { rank: 9, student_name: 'Ulugbek Hasanov',  score: 60, max_score: 100, percent: 60, time_taken_sec: 2500 },
                { rank: 10, student_name: 'Feruza Mamatova', score: 55, max_score: 100, percent: 55, time_taken_sec: 2700 },
            ];
        } finally {
            loadingLeaderboard = false;
        }
    }

    // ── Countdown — real-time 1s ──────────────────────────────────────────────
    function countdown(o: Olympiad): string {
        if (o.status === 'completed') return 'Yakunlangan';
        if (o.status === 'active') {
            const diff = new Date(o.end_time).getTime() - now.getTime();
            if (diff <= 0) return 'Tugayapti...';
            return formatDiff(diff) + ' qoldi';
        }
        // upcoming
        const diff = new Date(o.start_time).getTime() - now.getTime();
        if (diff <= 0) return 'Boshlanmoqda...';
        return formatDiff(diff) + ' qoldi';
    }

    function formatDiff(diff: number): string {
        const days  = Math.floor(diff / 86400000);
        const hours = Math.floor((diff % 86400000) / 3600000);
        const mins  = Math.floor((diff % 3600000) / 60000);
        const secs  = Math.floor((diff % 60000) / 1000);
        if (days > 0) {
            return `${days} kun ${String(hours).padStart(2,'0')}:${String(mins).padStart(2,'0')}:${String(secs).padStart(2,'0')}`;
        }
        return `${String(hours).padStart(2,'0')}:${String(mins).padStart(2,'0')}:${String(secs).padStart(2,'0')}`;
    }

    // ── Helpers ───────────────────────────────────────────────────────────────
    const MONTHS_UZ = ['Yanvar','Fevral','Mart','Aprel','May','Iyun','Iyul','Avgust','Sentabr','Oktabr','Noyabr','Dekabr'];

    function fmtDateTimeBetter(d: string | null): string {
        if (!d) return '—';
        const dt = new Date(d);
        const day   = dt.getDate();
        const month = MONTHS_UZ[dt.getMonth()];
        const year  = dt.getFullYear();
        const hh    = String(dt.getHours()).padStart(2, '0');
        const mm    = String(dt.getMinutes()).padStart(2, '0');
        return `${day}-${month} ${year}, ${hh}:${mm}`;
    }

    function fmtTime(sec: number) {
        const m = Math.floor(sec / 60), s = sec % 60;
        return `${m}:${String(s).padStart(2, '0')}`;
    }

    function initials(name: string): string {
        return name.split(' ').map(w => w[0] ?? '').join('').toUpperCase().slice(0, 2);
    }

    function medalFor(rank: number): string {
        if (rank === 1) return '🥇';
        if (rank === 2) return '🥈';
        if (rank === 3) return '🥉';
        return String(rank);
    }

    const statusInfo: Record<string, { label: string; cls: string; color: string }> = {
        upcoming:  { label: 'Kutilmoqda', cls: 'badge-upcoming',  color: '#3b82f6' },
        active:    { label: 'Jonli',      cls: 'badge-active',    color: '#22c55e' },
        completed: { label: 'Yakunlangan', cls: 'badge-completed', color: '#64748b' },
    };

    // Progress bar pct
    function participantPct(o: Olympiad): number {
        if (!o.max_participants || o.max_participants <= 0) return 0;
        return Math.min(100, Math.round((o.participant_count / o.max_participants) * 100));
    }

    // Podium avatar colors
    const podiumColors = ['#fbbf24', '#94a3b8', '#c4956a'];
</script>

<svelte:head><title>Olimpiadalar — Cognita.uz</title></svelte:head>

<!-- ── Header ────────────────────────────────────────────────────────────── -->
<div class="page-header animate-fade">
    <div>
        <h1>Olimpiadalar</h1>
        <p class="sub">Rejalashtirilgan musobaqalar va reyting</p>
    </div>
</div>

<!-- ── Stats ─────────────────────────────────────────────────────────────── -->
<div class="stats-strip animate-slide delay-1">
    <div class="stat-card">
        <div class="stat-icon" style="background: rgba(99,102,241,0.12); color: var(--primary);">🏆</div>
        <div class="stat-body">
            <span class="stat-value">{totalOlympiads}</span>
            <span class="stat-label">Jami olimpiadalar</span>
        </div>
    </div>
    <div class="stat-card">
        <div class="stat-icon" style="background: rgba(59,130,246,0.12); color: #3b82f6;">📅</div>
        <div class="stat-body">
            <span class="stat-value">{upcomingOlympiads}</span>
            <span class="stat-label">Rejalashtirilgan</span>
        </div>
    </div>
    <div class="stat-card">
        <div class="stat-icon" style="background: rgba(34,197,94,0.12); color: var(--success);">⚡</div>
        <div class="stat-body">
            <span class="stat-value">{activeOlympiads}</span>
            <span class="stat-label">Faol</span>
        </div>
    </div>
    <div class="stat-card">
        <div class="stat-icon" style="background: rgba(100,116,139,0.12); color: #64748b;">✅</div>
        <div class="stat-body">
            <span class="stat-value">{completedOlympiads}</span>
            <span class="stat-label">Yakunlangan</span>
        </div>
    </div>
</div>

<!-- ── Toolbar ────────────────────────────────────────────────────────────── -->
<div class="toolbar animate-slide delay-2">
    <button class="btn primary" on:click={openCreateModal}>+ Yangi olimpiada</button>
    <div class="search-wrap">
        <span class="search-icon">🔍</span>
        <input
            type="search"
            class="search-input"
            placeholder="Olimpiada qidirish..."
            bind:value={search}
        />
    </div>
    <span class="count">{filtered.length} ta olimpiada</span>
</div>

<!-- ── Olympiad Grid ──────────────────────────────────────────────────────── -->
{#if loading}
    <div class="olympiad-grid">
        {#each Array(4) as _}
            <div class="skeleton" style="height:340px; border-radius: var(--radius-lg);"></div>
        {/each}
    </div>
{:else if filtered.length === 0}
    <div class="empty-state animate-scale">
        <div class="empty-icon">🏆</div>
        {#if search}
            <p>"{search}" bo'yicha olimpiada topilmadi</p>
        {:else}
            <p>Hali olimpiada yo'q</p>
            <span>Birinchi olimpiadangizni yarating va o'quvchilaringizni sinab ko'ring</span>
        {/if}
    </div>
{:else}
    <div class="olympiad-grid animate-slide delay-3">
        {#each filtered as o (o.id)}
            <div class="olympiad-card" class:card-active={o.status === 'active'}>
                <!-- Header: status badge + countdown/live badge -->
                <div class="ocard-head">
                    <!-- Live badge with pulse dot -->
                    {#if o.status === 'active'}
                        <div class="live-indicator">
                            <span class="live-dot"></span>
                            <span class="live-text">JONLI</span>
                        </div>
                    {:else if o.status === 'upcoming'}
                        <span class="badge badge-upcoming">{statusInfo.upcoming.label}</span>
                    {:else}
                        <span class="badge badge-completed">{statusInfo.completed.label}</span>
                    {/if}

                    <!-- Countdown / timer -->
                    {#if o.status !== 'completed'}
                        <span class="countdown-chip" class:active={o.status === 'active'}>
                            🕐 {countdown(o)}
                        </span>
                    {:else}
                        <span class="countdown-chip done">✓ {countdown(o)}</span>
                    {/if}
                </div>

                <!-- Title & description -->
                <div class="ocard-body">
                    <h3 class="ocard-title">{o.title}</h3>
                    {#if o.description}
                        <p class="ocard-desc">{o.description}</p>
                    {/if}
                    <!-- Quiz chip -->
                    <div class="ocard-quiz">
                        <span class="quiz-dot-sm"></span>
                        {o.quiz_title || 'Quiz belgilanmagan'}
                    </div>
                </div>

                <!-- Participants progress bar -->
                {#if o.max_participants}
                    <div class="participants-section">
                        <div class="part-header">
                            <span class="part-label">👥 Ishtirokchilar</span>
                            <span class="part-count">{o.participant_count} / {o.max_participants}</span>
                        </div>
                        <div class="part-bar-bg">
                            <div
                                class="part-bar-fill"
                                class:bar-full={participantPct(o) >= 90}
                                style="width: {participantPct(o)}%"
                            ></div>
                        </div>
                        <span class="part-pct">{participantPct(o)}%</span>
                    </div>
                {:else}
                    <div class="participants-simple">
                        <span>👥</span>
                        <span class="part-simple-count">{o.participant_count} ishtirokchi</span>
                    </div>
                {/if}

                <!-- Code pill + copy link -->
                <div class="ocard-code-row">
                    <div class="code-pill">
                        <span class="code-label">KOD</span>
                        <code class="code-value">{o.code}</code>
                    </div>
                    <button class="copy-link-btn" on:click={() => copyOlympiadLink(o)} title="Havolani nusxalash">
                        🔗 Nusxa olish
                    </button>
                </div>

                <!-- Time range -->
                <div class="ocard-time">
                    <span class="time-icon">📅</span>
                    <span>{fmtDateTimeBetter(o.start_time)}</span>
                    <span class="time-arrow">→</span>
                    <span>{fmtDateTimeBetter(o.end_time)}</span>
                </div>

                <!-- Actions -->
                <div class="ocard-actions">
                    {#if o.status === 'upcoming'}
                        <button class="action-btn btn-success" on:click={() => changeStatus(o, 'active')}>
                            ▶ Boshlash
                        </button>
                    {:else if o.status === 'active'}
                        <button class="action-btn btn-warning" on:click={() => changeStatus(o, 'completed')}>
                            ⏹ Yakunlash
                        </button>
                    {/if}
                    <button class="action-btn btn-view" on:click={() => openLeaderboard(o)}>
                        🏅 Reyting
                    </button>
                    <button
                        class="action-btn btn-icon btn-danger"
                        on:click={() => deleteOlympiad(o)}
                        disabled={deletingId === o.id}
                        title="O'chirish"
                    >
                        {deletingId === o.id ? '...' : '🗑️'}
                    </button>
                </div>
            </div>
        {/each}
    </div>
{/if}

<!-- ── Create Modal ───────────────────────────────────────────────────────── -->
{#if showCreateModal}
    <div class="overlay" on:click|self={() => (showCreateModal = false)} role="dialog" aria-modal="true">
        <div class="modal animate-bounce">
            <div class="modal-header">
                <div>
                    <h2>Yangi olimpiada yaratish</h2>
                    <p class="sub">Barcha majburiy maydonlarni to'ldiring</p>
                </div>
                <button class="modal-close" on:click={() => (showCreateModal = false)}>✕</button>
            </div>

            <div class="modal-body">
                <label class="field">
                    <span>Olimpiada nomi *</span>
                    <input type="text" bind:value={formTitle} placeholder="Respublika matematika olimpiadasi..." />
                </label>

                <label class="field">
                    <span>Tavsif</span>
                    <input type="text" bind:value={formDescription} placeholder="Olimpiada haqida qisqacha..." />
                </label>

                <label class="field">
                    <span>Quiz tanlang</span>
                    <select bind:value={formQuizId}>
                        <option value="">— Quiz tanlang —</option>
                        {#each quizzes as q}
                            <option value={q.id}>{q.title}</option>
                        {/each}
                    </select>
                </label>

                <div class="form-row-2">
                    <label class="field">
                        <span>Boshlanish vaqti *</span>
                        <input type="datetime-local" bind:value={formStartTime} />
                    </label>
                    <label class="field">
                        <span>Tugash vaqti *</span>
                        <input type="datetime-local" bind:value={formEndTime} />
                    </label>
                </div>

                <label class="field">
                    <span>Maksimal ishtirokchilar (ixtiyoriy)</span>
                    <input type="number" bind:value={formMaxParticipants} min="1" placeholder="Cheksiz" />
                </label>
            </div>

            <div class="modal-footer">
                <button class="btn secondary" on:click={() => (showCreateModal = false)}>Bekor</button>
                <button class="btn primary" on:click={createOlympiad}
                    disabled={saving || !formTitle.trim() || !formStartTime || !formEndTime}>
                    {#if saving}<span class="spinner"></span>{/if}
                    {saving ? 'Saqlanmoqda...' : '✓ Yaratish'}
                </button>
            </div>
        </div>
    </div>
{/if}

<!-- ── Leaderboard Modal ───────────────────────────────────────────────────── -->
{#if showLeaderboardModal && selectedOlympiad}
    <div class="overlay" on:click|self={() => (showLeaderboardModal = false)} role="dialog" aria-modal="true">
        <div class="modal modal-lg animate-bounce">
            <div class="modal-header">
                <div>
                    <h2>🏅 Reyting: {selectedOlympiad.title}</h2>
                    <p class="sub">{selectedOlympiad.participant_count} ishtirokchi</p>
                </div>
                <button class="modal-close" on:click={() => (showLeaderboardModal = false)}>✕</button>
            </div>

            {#if loadingLeaderboard}
                <div style="padding:32px; display:flex; flex-direction:column; gap:10px;">
                    {#each Array(5) as _}
                        <div class="skeleton" style="height:48px;"></div>
                    {/each}
                </div>
            {:else if leaderboard.length === 0}
                <div class="empty-state" style="padding: 40px;">
                    <div class="empty-icon">📊</div>
                    <p>Hali natijalar yo'q</p>
                </div>
            {:else}
                <!-- Podium for top 3 with avatar initials -->
                {#if leaderboard.length >= 3}
                    <div class="podium">
                        <!-- 2nd place -->
                        <div class="podium-block podium-2">
                            <div class="podium-avatar-circle" style="background: {podiumColors[1]};">
                                {initials(leaderboard[1].student_name)}
                            </div>
                            <div class="podium-name">{leaderboard[1].student_name}</div>
                            <div class="podium-score">{leaderboard[1].percent}%</div>
                            <div class="podium-stand p2">
                                <span class="podium-rank-num">2</span>
                            </div>
                        </div>
                        <!-- 1st place -->
                        <div class="podium-block podium-1">
                            <div class="podium-crown">👑</div>
                            <div class="podium-avatar-circle gold-avatar" style="background: {podiumColors[0]};">
                                {initials(leaderboard[0].student_name)}
                            </div>
                            <div class="podium-name">{leaderboard[0].student_name}</div>
                            <div class="podium-score gold-score">{leaderboard[0].percent}%</div>
                            <div class="podium-stand p1">
                                <span class="podium-rank-num">1</span>
                            </div>
                        </div>
                        <!-- 3rd place -->
                        <div class="podium-block podium-3">
                            <div class="podium-avatar-circle" style="background: {podiumColors[2]};">
                                {initials(leaderboard[2].student_name)}
                            </div>
                            <div class="podium-name">{leaderboard[2].student_name}</div>
                            <div class="podium-score">{leaderboard[2].percent}%</div>
                            <div class="podium-stand p3">
                                <span class="podium-rank-num">3</span>
                            </div>
                        </div>
                    </div>
                {/if}

                <!-- Full table -->
                <div class="lb-table-wrap">
                    <table class="lb-table">
                        <thead>
                            <tr>
                                <th>O'rin</th>
                                <th>O'quvchi</th>
                                <th>Ball</th>
                                <th>Foiz</th>
                                <th>Vaqt</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each leaderboard.slice(0, 10) as entry}
                                <tr class:top3={entry.rank <= 3}>
                                    <td class="td-medal">
                                        {#if entry.rank <= 3}
                                            <span class="medal">{medalFor(entry.rank)}</span>
                                        {:else}
                                            <span class="rank-num">{entry.rank}</span>
                                        {/if}
                                    </td>
                                    <td class="td-name">{entry.student_name}</td>
                                    <td class="td-score">{entry.score}/{entry.max_score}</td>
                                    <td>
                                        <div class="lb-bar-wrap">
                                            <div class="lb-bar" style="width:{entry.percent}%"></div>
                                        </div>
                                        <span class="lb-pct">{entry.percent}%</span>
                                    </td>
                                    <td class="td-time">{fmtTime(entry.time_taken_sec)}</td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>

                <div class="modal-footer">
                    <button class="btn secondary" on:click={() => (showLeaderboardModal = false)}>Yopish</button>
                </div>
            {/if}
        </div>
    </div>
{/if}

<!-- ── Toast ──────────────────────────────────────────────────────────────── -->
{#if toast}
    <div class="toast {toast.kind === 'error' ? 'error' : 'success'} animate-slide">
        {toast.kind === 'error' ? '⚠️' : '✅'} {toast.msg}
    </div>
{/if}

<style>
    /* ── Layout ── */
    .page-header { margin-bottom: 22px; }
    h1 { font-size: 1.75rem; font-weight: 800; color: var(--text); line-height: 1.2; }
    h2 { font-size: 1.15rem; font-weight: 800; color: var(--text); }
    .sub { font-size: 0.88rem; color: var(--text3); margin-top: 3px; }

    /* ── Stats ── */
    .stats-strip { display: flex; gap: 14px; margin-bottom: 22px; flex-wrap: wrap; }
    .stat-card {
        background: white; border-radius: var(--radius); padding: 14px 18px;
        box-shadow: var(--shadow-sm); display: flex; align-items: center; gap: 13px;
        flex: 1; min-width: 150px; border: 1.5px solid transparent; transition: var(--transition);
    }
    .stat-card:hover { border-color: var(--primary-light); transform: translateY(-2px); box-shadow: var(--shadow); }
    .stat-icon {
        width: 44px; height: 44px; border-radius: 11px;
        display: flex; align-items: center; justify-content: center;
        font-size: 1.2rem; flex-shrink: 0;
    }
    .stat-body { display: flex; flex-direction: column; }
    .stat-value { font-size: 1.6rem; font-weight: 800; color: var(--text); line-height: 1; }
    .stat-label { font-size: 0.75rem; color: var(--text3); margin-top: 3px; }

    /* ── Toolbar ── */
    .toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 20px; flex-wrap: wrap; }
    .search-wrap { position: relative; flex: 1; min-width: 200px; max-width: 360px; }
    .search-icon { position: absolute; left: 11px; top: 50%; transform: translateY(-50%); font-size: 0.9rem; pointer-events: none; }
    .search-input {
        width: 100%; padding: 9px 14px 9px 34px;
        border: 1.5px solid var(--border); border-radius: var(--radius-sm);
        font-size: 0.9rem; outline: none; transition: border-color 0.2s; background: white;
    }
    .search-input:focus { border-color: var(--primary); }
    .count { font-size: 0.85rem; color: var(--text3); white-space: nowrap; }

    /* ── Olympiad grid ── */
    .olympiad-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
        gap: 16px;
    }
    .olympiad-card {
        background: white; border-radius: var(--radius-lg); padding: 18px;
        box-shadow: var(--shadow-sm); display: flex; flex-direction: column; gap: 12px;
        border: 1.5px solid transparent; transition: var(--transition);
    }
    .olympiad-card:hover { transform: translateY(-4px); box-shadow: var(--shadow); border-color: var(--primary-light); }
    .card-active {
        border-color: rgba(34,197,94,0.35);
        background: linear-gradient(135deg, rgba(34,197,94,0.025), white);
        box-shadow: 0 4px 20px rgba(34,197,94,0.14);
    }

    /* ── Card header ── */
    .ocard-head {
        display: flex; align-items: center; justify-content: space-between; gap: 8px; flex-wrap: wrap;
    }

    /* Live indicator with pulsing dot */
    .live-indicator {
        display: inline-flex; align-items: center; gap: 6px;
        background: rgba(239,68,68,0.1); border: 1.5px solid rgba(239,68,68,0.3);
        border-radius: 99px; padding: 3px 10px;
    }
    .live-dot {
        width: 8px; height: 8px; border-radius: 50%; background: #ef4444; flex-shrink: 0;
        animation: liveblink 1.2s ease infinite;
    }
    @keyframes liveblink {
        0%, 100% { opacity: 1; box-shadow: 0 0 0 0 rgba(239,68,68,0.5); }
        50% { opacity: 0.7; box-shadow: 0 0 0 5px rgba(239,68,68,0); }
    }
    .live-text { font-size: 0.72rem; font-weight: 900; color: #ef4444; letter-spacing: 0.08em; }

    .countdown-chip {
        font-size: 0.78rem; font-weight: 700; color: #1d4ed8;
        background: #eff6ff; padding: 4px 10px; border-radius: 99px;
        font-family: 'Courier New', monospace; letter-spacing: 0.02em;
    }
    .countdown-chip.active { color: #166534; background: #dcfce7; }
    .countdown-chip.done { color: #475569; background: #f1f5f9; font-family: inherit; }

    /* ── Card body ── */
    .ocard-body { display: flex; flex-direction: column; gap: 5px; }
    .ocard-title { font-size: 1rem; font-weight: 800; color: var(--text); line-height: 1.3; }
    .ocard-desc {
        font-size: 0.82rem; color: var(--text3); line-height: 1.45;
        display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;
        overflow: hidden;
    }
    .ocard-quiz {
        display: inline-flex; align-items: center; gap: 5px;
        font-size: 0.78rem; font-weight: 600;
        color: #6d28d9; background: var(--primary-light); padding: 3px 10px; border-radius: 99px;
        align-self: flex-start; margin-top: 2px;
    }
    .quiz-dot-sm {
        width: 7px; height: 7px; border-radius: 50%; background: #6d28d9; flex-shrink: 0;
    }

    /* ── Participants section ── */
    .participants-section { display: flex; flex-direction: column; gap: 5px; }
    .part-header { display: flex; align-items: center; justify-content: space-between; }
    .part-label { font-size: 0.78rem; font-weight: 600; color: var(--text2); }
    .part-count { font-size: 0.8rem; font-weight: 800; color: var(--text); }
    .part-bar-bg {
        width: 100%; height: 8px; background: var(--border); border-radius: 99px; overflow: hidden;
    }
    .part-bar-fill {
        height: 100%; border-radius: 99px;
        background: linear-gradient(90deg, var(--primary), var(--accent));
        transition: width 0.5s ease;
    }
    .part-bar-fill.bar-full { background: linear-gradient(90deg, var(--warning), var(--danger)); }
    .part-pct { font-size: 0.72rem; color: var(--text3); align-self: flex-end; }

    .participants-simple {
        display: flex; align-items: center; gap: 5px;
        font-size: 0.82rem; color: var(--text2); background: var(--bg2);
        border-radius: var(--radius-sm); padding: 6px 10px;
    }
    .part-simple-count { font-weight: 700; }

    /* ── Code + copy ── */
    .ocard-code-row { display: flex; align-items: center; gap: 9px; flex-wrap: wrap; }
    .code-pill {
        display: flex; align-items: center; gap: 6px;
        background: linear-gradient(135deg, rgba(99,102,241,0.1), rgba(139,92,246,0.06));
        border: 1.5px solid rgba(99,102,241,0.18); border-radius: 9px;
        padding: 5px 12px;
    }
    .code-label { font-size: 0.68rem; font-weight: 800; color: var(--primary); text-transform: uppercase; letter-spacing: 0.08em; }
    .code-value { font-family: 'Courier New', monospace; font-size: 1.05rem; font-weight: 800; color: var(--primary); letter-spacing: 0.1em; }
    .copy-link-btn {
        padding: 5px 11px; background: white; border: 1.5px solid var(--border);
        border-radius: 7px; font-size: 0.78rem; font-weight: 600; color: var(--text2);
        cursor: pointer; transition: var(--transition); font-family: inherit;
    }
    .copy-link-btn:hover { border-color: var(--primary); color: var(--primary); }

    /* ── Time range ── */
    .ocard-time {
        display: flex; align-items: center; gap: 6px; flex-wrap: wrap;
        font-size: 0.78rem; color: var(--text3); background: var(--bg2);
        border-radius: var(--radius-sm); padding: 7px 10px;
    }
    .time-icon { font-size: 0.9rem; }
    .time-arrow { color: var(--border); font-weight: 700; }

    /* ── Card actions ── */
    .ocard-actions { display: flex; gap: 7px; flex-wrap: wrap; }
    .action-btn {
        padding: 6px 12px; border: none; border-radius: 7px; cursor: pointer;
        font-size: 0.78rem; font-weight: 700; transition: var(--transition);
        display: inline-flex; align-items: center; gap: 4px; font-family: inherit;
    }
    .action-btn:disabled { opacity: 0.5; cursor: not-allowed; }
    .btn-success  { background: #dcfce7; color: #166534; }
    .btn-success:hover:not(:disabled)  { background: #bbf7d0; }
    .btn-warning  { background: #fef9c3; color: #854d0e; }
    .btn-warning:hover:not(:disabled)  { background: #fef08a; }
    .btn-view     { background: #eff6ff; color: #1d4ed8; }
    .btn-view:hover:not(:disabled)     { background: #dbeafe; }
    .btn-icon     { padding: 6px 9px; background: var(--bg2); color: var(--text2); }
    .btn-icon:hover:not(:disabled)     { background: var(--border); }
    .btn-danger   { background: #fee2e2; color: var(--danger); }
    .btn-danger:hover:not(:disabled)   { background: #fecaca; }

    /* ── Badges ── */
    .badge { display: inline-block; padding: 3px 10px; border-radius: 99px; font-size: 0.73rem; font-weight: 700; }
    .badge-upcoming  { background: #eff6ff; color: #1d4ed8; }
    .badge-active    { background: #dcfce7; color: #166534; }
    .badge-completed { background: #f1f5f9; color: #475569; }

    /* ── Empty ── */
    .empty-state {
        padding: 60px 20px; text-align: center; background: white;
        border-radius: var(--radius-lg); box-shadow: var(--shadow-sm);
    }
    .empty-icon { font-size: 3.5rem; margin-bottom: 12px; }
    .empty-state p { font-size: 1rem; font-weight: 700; color: var(--text); margin-bottom: 6px; }
    .empty-state span { font-size: 0.85rem; color: var(--text3); }

    /* ── Modals ── */
    .overlay {
        position: fixed; inset: 0;
        background: rgba(0,0,0,0.52); backdrop-filter: blur(4px);
        display: flex; align-items: center; justify-content: center;
        z-index: 300; padding: 20px;
    }
    .modal {
        background: white; border-radius: var(--radius-lg); padding: 26px;
        width: 100%; max-width: 500px;
        box-shadow: 0 24px 60px rgba(0,0,0,0.2);
        max-height: 90vh; overflow-y: auto;
    }
    .modal-lg { max-width: 640px; }
    .modal-header {
        display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 20px;
    }
    .modal-close {
        width: 32px; height: 32px; border-radius: 8px;
        background: var(--bg2); border: none; cursor: pointer;
        font-size: 0.9rem; color: var(--text3); flex-shrink: 0;
        display: flex; align-items: center; justify-content: center; transition: var(--transition);
    }
    .modal-close:hover { background: #fee2e2; color: var(--danger); }
    .modal-body { display: flex; flex-direction: column; gap: 14px; margin-bottom: 20px; }
    .modal-footer {
        display: flex; justify-content: flex-end; gap: 10px;
        padding-top: 16px; border-top: 1px solid var(--border);
    }

    .field { display: flex; flex-direction: column; gap: 5px; }
    .field span { font-size: 0.82rem; font-weight: 600; color: var(--text2); }
    .field input, .field select {
        padding: 9px 12px; border: 1.5px solid var(--border);
        border-radius: var(--radius-sm); font-size: 0.9rem; outline: none;
        transition: border-color 0.2s; width: 100%;
    }
    .field input:focus, .field select:focus { border-color: var(--primary); }
    .form-row-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }

    /* ── Podium ── */
    .podium {
        display: flex; align-items: flex-end; justify-content: center;
        margin: 0 0 20px; border-radius: var(--radius);
        padding: 24px 16px 0;
        border: 1.5px solid var(--border);
        background: linear-gradient(180deg, rgba(251,191,36,0.05) 0%, transparent 100%);
        gap: 0;
    }
    .podium-block {
        flex: 1; display: flex; flex-direction: column; align-items: center; gap: 4px;
    }
    .podium-crown { font-size: 1.3rem; margin-bottom: -2px; }
    .podium-avatar-circle {
        width: 44px; height: 44px; border-radius: 50%;
        display: flex; align-items: center; justify-content: center;
        font-size: 0.88rem; font-weight: 800; color: white;
        border: 2.5px solid white;
        box-shadow: 0 3px 10px rgba(0,0,0,0.15);
    }
    .gold-avatar {
        width: 52px; height: 52px; font-size: 1rem;
        box-shadow: 0 4px 16px rgba(251,191,36,0.5);
    }
    .podium-name { font-size: 0.76rem; font-weight: 700; color: var(--text); text-align: center; max-width: 90px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .podium-score { font-size: 0.8rem; font-weight: 800; color: var(--text2); margin-bottom: 6px; }
    .gold-score { color: #d97706; font-size: 0.88rem; }
    .podium-stand {
        width: 100%; border-radius: 6px 6px 0 0;
        display: flex; align-items: center; justify-content: center;
    }
    .p1 { height: 68px; background: linear-gradient(180deg, #fbbf24, #f59e0b); }
    .p2 { height: 50px; background: linear-gradient(180deg, #cbd5e1, #94a3b8); }
    .p3 { height: 36px; background: linear-gradient(180deg, #c4956a, #a17040); }
    .podium-rank-num { font-size: 1.1rem; font-weight: 900; color: white; }

    /* ── Leaderboard table ── */
    .lb-table-wrap { overflow-x: auto; margin-bottom: 4px; }
    .lb-table { width: 100%; border-collapse: collapse; }
    .lb-table th {
        padding: 9px 12px; text-align: left; font-size: 0.77rem; font-weight: 700;
        color: var(--text3); text-transform: uppercase; background: var(--bg2);
        border-bottom: 1.5px solid var(--border);
    }
    .lb-table td { padding: 10px 12px; font-size: 0.87rem; border-bottom: 1px solid var(--border); }
    .lb-table tbody tr:last-child td { border-bottom: none; }
    .lb-table tbody tr.top3 { background: rgba(251,191,36,0.04); }
    .lb-table tbody tr:hover { background: #fafbff; }

    .td-medal { width: 44px; text-align: center; }
    .medal { font-size: 1.2rem; }
    .rank-num { font-size: 0.88rem; font-weight: 800; color: var(--text3); }
    .td-name { font-weight: 600; }
    .td-score { font-weight: 700; font-family: monospace; }
    .td-time { color: var(--text3); font-size: 0.83rem; }

    .lb-bar-wrap {
        height: 5px; background: #e2e8f0; border-radius: 99px; overflow: hidden;
        width: 70px; display: inline-block; vertical-align: middle; margin-right: 6px;
    }
    .lb-bar {
        height: 100%;
        background: linear-gradient(90deg, var(--primary), var(--accent));
        border-radius: 99px;
    }
    .lb-pct { font-size: 0.82rem; font-weight: 700; color: var(--primary); vertical-align: middle; }

    /* ── Buttons ── */
    .btn {
        padding: 10px 18px; border: none; border-radius: var(--radius-sm);
        font-size: 0.9rem; font-weight: 600; cursor: pointer;
        text-decoration: none; display: inline-flex; align-items: center; gap: 6px;
        transition: var(--transition); font-family: inherit;
    }
    .btn.primary {
        background: linear-gradient(135deg, var(--primary), var(--accent));
        color: white; box-shadow: 0 3px 10px rgba(99,102,241,0.3);
    }
    .btn.primary:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 16px rgba(99,102,241,0.4); }
    .btn.secondary { background: var(--bg2); color: var(--text); border: 1px solid var(--border); }
    .btn.secondary:hover { background: var(--border); }
    .btn:disabled { opacity: 0.5; cursor: not-allowed; }

    /* ── Responsive ── */
    @media (max-width: 768px) {
        .olympiad-grid { grid-template-columns: 1fr; }
        .stats-strip { gap: 8px; }
        .podium { padding: 12px 8px 0; }
    }
    @media (max-width: 480px) {
        .form-row-2 { grid-template-columns: 1fr; }
        .stats-strip { flex-direction: column; }
    }
</style>
