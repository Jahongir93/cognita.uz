<script lang="ts">
    import { onMount } from 'svelte';
    import { quizzes as quizzesApi } from '$lib/api/client';
    import type { Quiz } from '$lib/api/types';

    // ── Types ─────────────────────────────────────────────────────────────────
    interface Exam {
        id: string;
        title: string;
        quiz_id: string;
        quiz_title: string;
        code: string;
        time_limit: number;
        status: 'draft' | 'active' | 'closed';
        submission_count: number;
        max_attempts: number;
        shuffle_questions: boolean;
        start_date: string | null;
        end_date: string | null;
        created_at: string;
    }

    interface Submission {
        id: string;
        student_name: string;
        score: number;
        max_score: number;
        percent: number;
        time_taken_sec: number;
        submitted_at: string;
    }

    // ── State ─────────────────────────────────────────────────────────────────
    let exams: Exam[] = [];
    let quizzes: Quiz[] = [];
    let loading = true;
    let search = '';
    let filterTab: 'all' | 'draft' | 'active' | 'closed' = 'all';

    let showCreateModal = false;
    let showResultsModal = false;
    let showShareModal = false;
    let selectedExam: Exam | null = null;
    let resultsData: Submission[] = [];
    let loadingResults = false;
    let saving = false;
    let deletingId: string | null = null;

    // Form fields
    let formTitle = '';
    let formQuizId = '';
    let formTimeLimit = 30;
    let formStartDate = '';
    let formEndDate = '';
    let formShuffle = false;
    let formMaxAttempts = 1;

    // ── Toast ─────────────────────────────────────────────────────────────────
    type ToastKind = 'success' | 'error';
    let toast: { msg: string; kind: ToastKind } | null = null;
    let toastTimer: ReturnType<typeof setTimeout>;
    function showToast(msg: string, kind: ToastKind = 'success') {
        toast = { msg, kind };
        clearTimeout(toastTimer);
        toastTimer = setTimeout(() => (toast = null), 3500);
    }

    // ── Derived ───────────────────────────────────────────────────────────────
    $: byTab = filterTab === 'all' ? exams : exams.filter(e => e.status === filterTab);

    $: filtered = search
        ? byTab.filter(e =>
            e.title.toLowerCase().includes(search.toLowerCase()) ||
            (e.quiz_title ?? '').toLowerCase().includes(search.toLowerCase()) ||
            e.code.toLowerCase().includes(search.toLowerCase())
          )
        : byTab;

    $: totalExams       = exams.length;
    $: activeExams      = exams.filter(e => e.status === 'active').length;
    $: draftExams       = exams.filter(e => e.status === 'draft').length;
    $: closedExams      = exams.filter(e => e.status === 'closed').length;
    $: totalSubmissions = exams.reduce((s, e) => s + e.submission_count, 0);

    // ── Score distribution ────────────────────────────────────────────────────
    $: scoreDistribution = (() => {
        if (resultsData.length === 0) return [0, 0, 0, 0];
        const buckets = [0, 0, 0, 0];
        resultsData.forEach(r => {
            if (r.percent < 40)      buckets[0]++;
            else if (r.percent < 60) buckets[1]++;
            else if (r.percent < 80) buckets[2]++;
            else                     buckets[3]++;
        });
        return buckets;
    })();

    $: maxBucket = Math.max(...scoreDistribution, 1);

    // ── QR code generator (pseudo-QR based on hash) ───────────────────────────
    function strHash(s: string): number {
        let h = 5381;
        for (let i = 0; i < s.length; i++) h = ((h << 5) + h) ^ s.charCodeAt(i);
        return Math.abs(h);
    }

    function generateQRCells(url: string): boolean[][] {
        const size = 12;
        const seed = strHash(url);
        const grid: boolean[][] = [];
        // Finder pattern positions (corners always filled)
        for (let r = 0; r < size; r++) {
            grid[r] = [];
            for (let c = 0; c < size; c++) {
                // Corner finder patterns
                const topLeft   = (r < 3 && c < 3) || (r === 0 && c < 4) || (c === 0 && r < 4);
                const topRight  = (r < 3 && c >= size - 3) || (r === 0 && c >= size - 4) || (c === size - 1 && r < 4);
                const botLeft   = (r >= size - 3 && c < 3) || (r === size - 1 && c < 4) || (c === 0 && r >= size - 4);
                if (topLeft || topRight || botLeft) { grid[r][c] = true; continue; }
                // Pseudo-random fill based on hash
                const idx = r * size + c;
                const bit = (seed >>> (idx % 30)) & 1;
                const bit2 = (strHash(url + idx) >>> 3) & 1;
                grid[r][c] = (bit ^ bit2) === 1;
            }
        }
        return grid;
    }

    $: shareUrl = selectedExam ? studentLink(selectedExam.code) : '';
    $: qrCells  = selectedExam ? generateQRCells(shareUrl) : [];

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
    onMount(async () => {
        await Promise.all([loadExams(), loadQuizzes()]);
    });

    async function loadQuizzes() {
        try { quizzes = await quizzesApi.list(); } catch { quizzes = []; }
    }

    async function loadExams() {
        loading = true;
        try {
            exams = await apiFetch<Exam[]>('/api/exams');
        } catch (e: any) {
            showToast('Imtihonlarni yuklashda xato: ' + e.message, 'error');
            exams = [
                {
                    id: 'demo-1', title: 'Matematika yakuniy imtihon', quiz_id: '', quiz_title: 'Algebra asoslari',
                    code: 'MATH01', time_limit: 45, status: 'active', submission_count: 12, max_attempts: 1,
                    shuffle_questions: true, start_date: null, end_date: null, created_at: new Date().toISOString(),
                },
                {
                    id: 'demo-2', title: 'Fizika oraliq test', quiz_id: '', quiz_title: 'Mexanika',
                    code: 'PHY02', time_limit: 30, status: 'draft', submission_count: 0, max_attempts: 2,
                    shuffle_questions: false, start_date: null, end_date: null, created_at: new Date().toISOString(),
                },
                {
                    id: 'demo-3', title: 'Tarix sinov ishi', quiz_id: '', quiz_title: 'Qadimgi tarix',
                    code: 'HIS03', time_limit: 60, status: 'closed', submission_count: 28, max_attempts: 1,
                    shuffle_questions: false, start_date: null, end_date: null, created_at: new Date().toISOString(),
                },
            ];
        } finally {
            loading = false;
        }
    }

    // ── Create exam ───────────────────────────────────────────────────────────
    function openCreateModal() {
        formTitle = ''; formQuizId = quizzes[0]?.id ?? '';
        formTimeLimit = 30; formStartDate = ''; formEndDate = '';
        formShuffle = false; formMaxAttempts = 1;
        showCreateModal = true;
    }

    async function createExam() {
        if (!formTitle.trim()) { showToast('Imtihon nomini kiriting', 'error'); return; }
        saving = true;
        try {
            const body = {
                title: formTitle.trim(),
                quiz_id: formQuizId,
                time_limit: formTimeLimit,
                start_date: formStartDate || null,
                end_date: formEndDate || null,
                shuffle_questions: formShuffle,
                max_attempts: formMaxAttempts,
            };
            const created = await apiFetch<Exam>('/api/exams', { method: 'POST', body: JSON.stringify(body) });
            exams = [created, ...exams];
            showCreateModal = false;
            showToast('Imtihon yaratildi!');
        } catch (e: any) {
            showToast('Xato: ' + e.message, 'error');
        } finally {
            saving = false;
        }
    }

    // ── Toggle status ─────────────────────────────────────────────────────────
    async function toggleStatus(exam: Exam) {
        const newStatus = exam.status === 'active' ? 'closed' : 'active';
        try {
            await apiFetch(`/api/exams/${exam.id}/status`, {
                method: 'PUT', body: JSON.stringify({ status: newStatus }),
            });
            exams = exams.map(e => e.id === exam.id ? { ...e, status: newStatus as Exam['status'] } : e);
            showToast(newStatus === 'active' ? 'Imtihon faollashtirildi' : 'Imtihon yopildi');
        } catch (e: any) {
            showToast('Holat o\'zgartirishda xato: ' + e.message, 'error');
        }
    }

    // ── Delete exam ───────────────────────────────────────────────────────────
    async function deleteExam(exam: Exam) {
        if (!confirm(`"${exam.title}" imtihonini o'chirishni tasdiqlaysizmi?`)) return;
        deletingId = exam.id;
        try {
            await apiFetch(`/api/exams/${exam.id}`, { method: 'DELETE' });
            exams = exams.filter(e => e.id !== exam.id);
            showToast('Imtihon o\'chirildi');
        } catch (e: any) {
            showToast('O\'chirishda xato: ' + e.message, 'error');
        } finally {
            deletingId = null;
        }
    }

    // ── Results ───────────────────────────────────────────────────────────────
    async function openResults(exam: Exam) {
        selectedExam = exam;
        showResultsModal = true;
        loadingResults = true;
        resultsData = [];
        try {
            resultsData = await apiFetch<Submission[]>(`/api/exams/${exam.id}/results`);
        } catch {
            resultsData = [
                { id: '1', student_name: 'Akbar Toshmatov',   score: 85,  max_score: 100, percent: 85, time_taken_sec: 1540, submitted_at: new Date().toISOString() },
                { id: '2', student_name: 'Zulfiya Karimova',  score: 92,  max_score: 100, percent: 92, time_taken_sec: 1200, submitted_at: new Date().toISOString() },
                { id: '3', student_name: 'Bobur Rahimov',     score: 47,  max_score: 100, percent: 47, time_taken_sec: 2700, submitted_at: new Date().toISOString() },
                { id: '4', student_name: 'Dilnoza Yusupova',  score: 73,  max_score: 100, percent: 73, time_taken_sec: 1800, submitted_at: new Date().toISOString() },
                { id: '5', student_name: 'Sardor Nazarov',    score: 60,  max_score: 100, percent: 60, time_taken_sec: 2100, submitted_at: new Date().toISOString() },
                { id: '6', student_name: 'Nozima Aliyeva',    score: 35,  max_score: 100, percent: 35, time_taken_sec: 2800, submitted_at: new Date().toISOString() },
                { id: '7', student_name: 'Jasur Mirzayev',    score: 78,  max_score: 100, percent: 78, time_taken_sec: 1650, submitted_at: new Date().toISOString() },
            ];
        } finally {
            loadingResults = false;
        }
    }

    // ── Share modal ───────────────────────────────────────────────────────────
    function openShare(exam: Exam) {
        selectedExam = exam;
        showShareModal = true;
    }

    // ── CSV download ──────────────────────────────────────────────────────────
    function downloadCSV() {
        if (!selectedExam || resultsData.length === 0) return;
        const header = ['O\'quvchi', 'Ball', 'Max ball', 'Foiz (%)', 'Vaqt (daqiqa)', 'Sana'];
        const rows = resultsData.map(r => [
            r.student_name, r.score, r.max_score, r.percent,
            Math.round(r.time_taken_sec / 60),
            new Date(r.submitted_at).toLocaleDateString('uz-UZ'),
        ]);
        const csv = [header, ...rows].map(r => r.join(',')).join('\n');
        const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url; a.download = `${selectedExam!.title}-natijalar.csv`; a.click();
        URL.revokeObjectURL(url);
    }

    // ── Helpers ───────────────────────────────────────────────────────────────
    function fmtDate(d: string | null) {
        if (!d) return '—';
        return new Date(d).toLocaleDateString('uz-UZ', { day: '2-digit', month: '2-digit', year: 'numeric' });
    }

    function fmtTime(sec: number) {
        const m = Math.floor(sec / 60), s = sec % 60;
        return `${m}:${String(s).padStart(2, '0')}`;
    }

    function studentLink(code: string) {
        return (typeof window !== 'undefined' ? window.location.origin : '') + '/exam/' + code;
    }

    function copyLink(code: string) {
        navigator.clipboard?.writeText(studentLink(code)).then(() => showToast('Havola nusxalandi!'));
    }

    $: avgScore = resultsData.length
        ? Math.round(resultsData.reduce((s, r) => s + r.percent, 0) / resultsData.length)
        : 0;
    $: passRate = resultsData.length
        ? Math.round((resultsData.filter(r => r.percent >= 60).length / resultsData.length) * 100)
        : 0;

    const statusInfo: Record<string, { label: string; cls: string }> = {
        draft:  { label: 'Qoralama', cls: 'badge-draft'   },
        active: { label: 'Faol',     cls: 'badge-active'  },
        closed: { label: 'Yopiq',    cls: 'badge-closed'  },
    };

    const statusSteps = ['draft', 'active', 'closed'];
</script>

<svelte:head><title>Imtihon rejimi — Cognita.uz</title></svelte:head>

<!-- ── Header ────────────────────────────────────────────────────────────── -->
<div class="page-header animate-fade">
    <div>
        <h1>Imtihon rejimi</h1>
        <p class="sub">Vaqt cheklangan sinovlar yarating va boshqaring</p>
    </div>
</div>

<!-- ── Stats ─────────────────────────────────────────────────────────────── -->
<div class="stats-strip animate-slide delay-1">
    <div class="stat-card">
        <div class="stat-icon" style="background: rgba(99,102,241,0.12); color: var(--primary);">📋</div>
        <div class="stat-body">
            <span class="stat-value">{totalExams}</span>
            <span class="stat-label">Jami imtihonlar</span>
        </div>
    </div>
    <div class="stat-card">
        <div class="stat-icon" style="background: rgba(34,197,94,0.12); color: var(--success);">✅</div>
        <div class="stat-body">
            <span class="stat-value">{activeExams}</span>
            <span class="stat-label">Faol imtihonlar</span>
        </div>
    </div>
    <div class="stat-card">
        <div class="stat-icon" style="background: rgba(245,158,11,0.12); color: var(--warning);">📝</div>
        <div class="stat-body">
            <span class="stat-value">{totalSubmissions}</span>
            <span class="stat-label">Jami topshirishlar</span>
        </div>
    </div>
</div>

<!-- ── Toolbar ────────────────────────────────────────────────────────────── -->
<div class="toolbar animate-slide delay-2">
    <button class="btn primary" on:click={openCreateModal}>+ Yangi imtihon</button>
    <div class="search-wrap">
        <span class="search-icon">🔍</span>
        <input
            type="search"
            class="search-input"
            placeholder="Imtihon qidirish..."
            bind:value={search}
        />
    </div>
    <span class="count">{filtered.length} ta imtihon</span>
</div>

<!-- ── Filter Tabs ────────────────────────────────────────────────────────── -->
<div class="filter-tabs animate-slide delay-2">
    <button
        class="ftab"
        class:active={filterTab === 'all'}
        on:click={() => (filterTab = 'all')}
    >Barchasi <span class="ftab-badge">{totalExams}</span></button>
    <button
        class="ftab"
        class:active={filterTab === 'draft'}
        on:click={() => (filterTab = 'draft')}
    >Draft <span class="ftab-badge draft">{draftExams}</span></button>
    <button
        class="ftab"
        class:active={filterTab === 'active'}
        on:click={() => (filterTab = 'active')}
    >Faol <span class="ftab-badge active">{activeExams}</span></button>
    <button
        class="ftab"
        class:active={filterTab === 'closed'}
        on:click={() => (filterTab = 'closed')}
    >Yopiq <span class="ftab-badge closed">{closedExams}</span></button>
</div>

<!-- ── Exam grid ──────────────────────────────────────────────────────────── -->
{#if loading}
    <div class="exam-grid">
        {#each Array(4) as _}
            <div class="skeleton" style="height:280px; border-radius: var(--radius-lg);"></div>
        {/each}
    </div>
{:else if filtered.length === 0}
    <div class="empty-state animate-scale">
        {#if search || filterTab !== 'all'}
            <div class="empty-illustration">🔍</div>
            <div class="empty-title">{search ? `"${search}" topilmadi` : `${statusInfo[filterTab]?.label ?? filterTab} imtihonlar yo'q`}</div>
            <p class="empty-sub">Qidiruv so'rovini yoki filterni o'zgartiring</p>
            <button class="btn secondary" style="margin-top:12px;" on:click={() => { search = ''; filterTab = 'all'; }}>
                Tozalash
            </button>
        {:else}
            <div class="empty-illustration">📋</div>
            <div class="empty-title">Hali imtihon yo'q</div>
            <p class="empty-sub">Yangi imtihon yaratib, o'quvchilaringizni sinab ko'ring.<br>Vaqt chegarasi va kod orqali boshqarish qulay.</p>
            <button class="btn primary" style="margin-top:14px;" on:click={openCreateModal}>
                + Birinchi imtihonni yarating
            </button>
        {/if}
    </div>
{:else}
    <div class="exam-grid animate-slide delay-3">
        {#each filtered as exam (exam.id)}
            <div class="exam-card">
                <!-- Status pipeline -->
                <div class="status-pipeline">
                    {#each statusSteps as step, i}
                        <div
                            class="pipeline-step"
                            class:done={statusSteps.indexOf(exam.status) > i}
                            class:current={exam.status === step}
                        >
                            <div class="pip-dot"></div>
                            <span class="pip-label">{statusInfo[step]?.label ?? step}</span>
                        </div>
                        {#if i < statusSteps.length - 1}
                            <div class="pip-line" class:filled={statusSteps.indexOf(exam.status) > i}></div>
                        {/if}
                    {/each}
                </div>

                <!-- Top row: title -->
                <div class="exam-card-top">
                    <div class="exam-title">{exam.title}</div>
                    <div class="exam-quiz">{exam.quiz_title || 'Quiz belgilanmagan'}</div>
                </div>

                <!-- Meta -->
                <div class="exam-meta">
                    <div class="meta-item">
                        <span class="meta-icon">⏱</span>
                        <span>{exam.time_limit} daqiqa</span>
                    </div>
                    <div class="meta-item">
                        <span class="meta-icon">📝</span>
                        <span>{exam.submission_count} topshirilgan</span>
                    </div>
                    <div class="meta-item">
                        <span class="meta-icon">🔄</span>
                        <span>{exam.max_attempts}x urinish</span>
                    </div>
                    {#if exam.start_date || exam.end_date}
                        <div class="meta-item">
                            <span class="meta-icon">📅</span>
                            <span>{fmtDate(exam.start_date)} — {fmtDate(exam.end_date)}</span>
                        </div>
                    {/if}
                </div>

                <!-- Code pill + share -->
                <div class="exam-code-row">
                    <div class="code-pill">
                        <span class="code-label">KOD</span>
                        <code class="code-value">{exam.code}</code>
                    </div>
                    <button class="copy-btn" on:click={() => openShare(exam)} title="Ulashish">
                        🔗 Ulashish
                    </button>
                </div>

                <!-- Actions -->
                <div class="exam-actions">
                    <button
                        class="action-btn {exam.status === 'active' ? 'btn-warning' : 'btn-success'}"
                        on:click={() => toggleStatus(exam)}
                        title={exam.status === 'active' ? 'Yopish' : 'Faollashtirish'}
                    >
                        {exam.status === 'active' ? '⏸ Yopish' : '▶ Faollashtirish'}
                    </button>
                    <button class="action-btn btn-view" on:click={() => openResults(exam)} title="Natijalar">
                        👁 Natijalar
                    </button>
                    <button class="action-btn btn-icon" title="Tahrirlash" disabled>✏️</button>
                    <button
                        class="action-btn btn-icon btn-danger"
                        on:click={() => deleteExam(exam)}
                        disabled={deletingId === exam.id}
                        title="O'chirish"
                    >
                        {deletingId === exam.id ? '...' : '🗑️'}
                    </button>
                </div>
            </div>
        {/each}
    </div>
{/if}

<!-- ── Share Modal ────────────────────────────────────────────────────────── -->
{#if showShareModal && selectedExam}
    <div class="overlay" on:click|self={() => (showShareModal = false)} role="dialog" aria-modal="true">
        <div class="modal animate-bounce">
            <div class="modal-header">
                <div>
                    <h2>🔗 Ulashish: {selectedExam.title}</h2>
                    <p class="sub">O'quvchilarga ushbu havola yoki kodni ulashing</p>
                </div>
                <button class="modal-close" on:click={() => (showShareModal = false)}>✕</button>
            </div>

            <div class="share-body">
                <!-- URL display -->
                <div class="share-url-box">
                    <span class="share-url-text">{shareUrl}</span>
                    <button class="share-copy-btn" on:click={() => { navigator.clipboard?.writeText(shareUrl); showToast('Havola nusxalandi!'); }}>
                        Nusxalash
                    </button>
                </div>

                <!-- Code badge -->
                <div class="share-code-badge">
                    <span class="scb-label">Kirish kodi</span>
                    <code class="scb-code">{selectedExam.code}</code>
                </div>

                <!-- QR code (pseudo) -->
                <div class="qr-wrap">
                    <p class="qr-label">QR kod</p>
                    <div class="qr-grid">
                        {#each qrCells as row}
                            <div class="qr-row">
                                {#each row as cell}
                                    <div class="qr-cell" class:filled={cell}></div>
                                {/each}
                            </div>
                        {/each}
                    </div>
                </div>
            </div>

            <div class="modal-footer">
                <button class="btn secondary" on:click={() => (showShareModal = false)}>Yopish</button>
            </div>
        </div>
    </div>
{/if}

<!-- ── Create Modal ───────────────────────────────────────────────────────── -->
{#if showCreateModal}
    <div class="overlay" on:click|self={() => (showCreateModal = false)} role="dialog" aria-modal="true">
        <div class="modal animate-bounce">
            <div class="modal-header">
                <div>
                    <h2>Yangi imtihon yaratish</h2>
                    <p class="sub">Barcha maydonlarni to'ldiring</p>
                </div>
                <button class="modal-close" on:click={() => (showCreateModal = false)}>✕</button>
            </div>

            <div class="modal-body">
                <label class="field">
                    <span>Imtihon nomi *</span>
                    <input type="text" bind:value={formTitle} placeholder="Matematika yakuniy imtihon..." />
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
                        <span>Vaqt chegarasi (daqiqa)</span>
                        <input type="number" bind:value={formTimeLimit} min="1" max="300" />
                    </label>
                    <label class="field">
                        <span>Maksimal urinishlar</span>
                        <input type="number" bind:value={formMaxAttempts} min="1" max="10" />
                    </label>
                </div>

                <div class="form-row-2">
                    <label class="field">
                        <span>Boshlanish vaqti (ixtiyoriy)</span>
                        <input type="datetime-local" bind:value={formStartDate} />
                    </label>
                    <label class="field">
                        <span>Tugash vaqti (ixtiyoriy)</span>
                        <input type="datetime-local" bind:value={formEndDate} />
                    </label>
                </div>

                <label class="toggle-row">
                    <span class="toggle-label">
                        <span>🔀</span> Savollarni aralashtirish
                    </span>
                    <div class="toggle-wrap">
                        <input type="checkbox" class="toggle-input" bind:checked={formShuffle} id="form-shuf" />
                        <label class="toggle-knob" for="form-shuf"></label>
                    </div>
                </label>
            </div>

            <div class="modal-footer">
                <button class="btn secondary" on:click={() => (showCreateModal = false)}>Bekor</button>
                <button class="btn primary" on:click={createExam} disabled={saving || !formTitle.trim()}>
                    {#if saving}<span class="spinner"></span>{/if}
                    {saving ? 'Saqlanmoqda...' : '✓ Yaratish'}
                </button>
            </div>
        </div>
    </div>
{/if}

<!-- ── Results Modal ──────────────────────────────────────────────────────── -->
{#if showResultsModal && selectedExam}
    <div class="overlay" on:click|self={() => (showResultsModal = false)} role="dialog" aria-modal="true">
        <div class="modal modal-lg animate-bounce">
            <div class="modal-header">
                <div>
                    <h2>📊 Natijalar: {selectedExam.title}</h2>
                    <p class="sub">{selectedExam.submission_count} ta topshirish</p>
                </div>
                <button class="modal-close" on:click={() => (showResultsModal = false)}>✕</button>
            </div>

            {#if loadingResults}
                <div style="padding: 32px; display: flex; flex-direction: column; gap: 10px;">
                    {#each Array(3) as _}
                        <div class="skeleton" style="height:44px;"></div>
                    {/each}
                </div>
            {:else}
                <!-- Summary -->
                <div class="results-summary">
                    <div class="summary-item">
                        <span class="summary-value" style="color: var(--primary);">{avgScore}%</span>
                        <span class="summary-label">O'rtacha ball</span>
                    </div>
                    <div class="summary-item">
                        <span class="summary-value" style="color: {passRate >= 60 ? 'var(--success)' : 'var(--danger)'};">{passRate}%</span>
                        <span class="summary-label">O'tish darajasi (≥60%)</span>
                    </div>
                    <div class="summary-item">
                        <span class="summary-value">{resultsData.length}</span>
                        <span class="summary-label">Ishtirokchilar</span>
                    </div>
                </div>

                <!-- Score distribution chart -->
                {#if resultsData.length > 0}
                    <div class="dist-chart">
                        <div class="dist-title">Ball taqsimoti</div>
                        {#each [
                            { label: '0–40%',   idx: 0, color: '#ef4444' },
                            { label: '40–60%',  idx: 1, color: '#f59e0b' },
                            { label: '60–80%',  idx: 2, color: '#6366f1' },
                            { label: '80–100%', idx: 3, color: '#22c55e' },
                        ] as band}
                            <div class="dist-row">
                                <span class="dist-label">{band.label}</span>
                                <div class="dist-bar-wrap">
                                    <div
                                        class="dist-bar-fill"
                                        style="width: {(scoreDistribution[band.idx] / maxBucket) * 100}%; background: {band.color};"
                                    ></div>
                                </div>
                                <span class="dist-count">{scoreDistribution[band.idx]}</span>
                            </div>
                        {/each}
                    </div>
                {/if}

                <!-- Table -->
                <div class="results-table-wrap">
                    <table class="results-table">
                        <thead>
                            <tr>
                                <th>#</th>
                                <th>O'quvchi</th>
                                <th>Ball</th>
                                <th>Foiz</th>
                                <th>Vaqt</th>
                                <th>Sana</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each resultsData as sub, i}
                                <tr>
                                    <td class="td-rank">{i + 1}</td>
                                    <td class="td-name">{sub.student_name}</td>
                                    <td class="td-score">{sub.score}/{sub.max_score}</td>
                                    <td>
                                        <div class="percent-bar">
                                            <div class="percent-fill" style="width:{sub.percent}%; background:{sub.percent >= 60 ? 'var(--success)' : 'var(--danger)'}"></div>
                                        </div>
                                        <span class="percent-text" style="color:{sub.percent >= 60 ? 'var(--success)' : 'var(--danger)'}">{sub.percent}%</span>
                                    </td>
                                    <td>{fmtTime(sub.time_taken_sec)}</td>
                                    <td class="td-date">{fmtDate(sub.submitted_at)}</td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>

                <div class="modal-footer">
                    <button class="btn secondary" on:click={() => (showResultsModal = false)}>Yopish</button>
                    <button class="btn primary" on:click={downloadCSV}>⬇ CSV yuklab olish</button>
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
    .stats-strip {
        display: flex; gap: 14px; margin-bottom: 22px; flex-wrap: wrap;
    }
    .stat-card {
        background: white; border-radius: var(--radius); padding: 14px 18px;
        box-shadow: var(--shadow-sm); display: flex; align-items: center; gap: 13px;
        flex: 1; min-width: 170px; border: 1.5px solid transparent;
        transition: var(--transition);
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
    .toolbar {
        display: flex; align-items: center; gap: 12px; margin-bottom: 14px; flex-wrap: wrap;
    }
    .search-wrap { position: relative; flex: 1; min-width: 200px; max-width: 360px; }
    .search-icon { position: absolute; left: 11px; top: 50%; transform: translateY(-50%); font-size: 0.9rem; pointer-events: none; }
    .search-input {
        width: 100%; padding: 9px 14px 9px 34px;
        border: 1.5px solid var(--border); border-radius: var(--radius-sm);
        font-size: 0.9rem; outline: none; transition: border-color 0.2s;
        background: white;
    }
    .search-input:focus { border-color: var(--primary); }
    .count { font-size: 0.85rem; color: var(--text3); white-space: nowrap; }

    /* ── Filter tabs ── */
    .filter-tabs {
        display: flex; gap: 6px; margin-bottom: 20px; flex-wrap: wrap;
    }
    .ftab {
        display: inline-flex; align-items: center; gap: 6px;
        padding: 7px 14px; border: 1.5px solid var(--border);
        border-radius: 99px; background: white; cursor: pointer;
        font-size: 0.83rem; font-weight: 600; color: var(--text2);
        transition: var(--transition); font-family: inherit;
    }
    .ftab:hover { border-color: var(--primary); color: var(--primary); }
    .ftab.active {
        background: var(--primary); color: white; border-color: var(--primary);
        box-shadow: 0 3px 10px rgba(99,102,241,0.3);
    }
    .ftab.active .ftab-badge { background: rgba(255,255,255,0.25); color: white; }
    .ftab-badge {
        font-size: 0.72rem; font-weight: 800; padding: 1px 7px;
        border-radius: 99px; background: var(--bg2); color: var(--text3);
    }
    .ftab-badge.draft  { background: #f1f5f9; color: #475569; }
    .ftab-badge.active { background: #dcfce7; color: #166534; }
    .ftab-badge.closed { background: #fee2e2; color: #991b1b; }

    /* ── Exam Grid ── */
    .exam-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
        gap: 16px;
    }
    .exam-card {
        background: white; border-radius: var(--radius-lg); padding: 18px;
        box-shadow: var(--shadow-sm); display: flex; flex-direction: column; gap: 12px;
        border: 1.5px solid transparent; transition: var(--transition);
    }
    .exam-card:hover { transform: translateY(-4px); box-shadow: var(--shadow); border-color: var(--primary-light); }

    /* ── Status pipeline ── */
    .status-pipeline {
        display: flex; align-items: center; gap: 0;
    }
    .pipeline-step {
        display: flex; flex-direction: column; align-items: center; gap: 4px; flex: 0 0 auto;
    }
    .pip-dot {
        width: 11px; height: 11px; border-radius: 50%;
        background: var(--border); border: 2px solid var(--border);
        transition: var(--transition);
    }
    .pipeline-step.done .pip-dot {
        background: var(--primary); border-color: var(--primary);
    }
    .pipeline-step.current .pip-dot {
        background: white; border-color: var(--primary);
        box-shadow: 0 0 0 3px rgba(99,102,241,0.2);
        width: 13px; height: 13px;
    }
    .pip-label {
        font-size: 0.65rem; font-weight: 700; color: var(--text3);
        text-transform: uppercase; letter-spacing: 0.04em; white-space: nowrap;
    }
    .pipeline-step.done .pip-label,
    .pipeline-step.current .pip-label { color: var(--primary); }
    .pip-line {
        flex: 1; height: 2px; background: var(--border);
        margin: 0 4px; margin-bottom: 18px; min-width: 20px;
        transition: background 0.2s;
    }
    .pip-line.filled { background: var(--primary); }

    .exam-card-top { display: flex; flex-direction: column; gap: 5px; }
    .exam-title { font-size: 1rem; font-weight: 700; color: var(--text); line-height: 1.3; }
    .exam-quiz { font-size: 0.8rem; color: var(--text3); }

    .exam-meta { display: flex; flex-wrap: wrap; gap: 7px; }
    .meta-item {
        display: flex; align-items: center; gap: 4px;
        font-size: 0.78rem; color: var(--text2); background: var(--bg2);
        border-radius: 6px; padding: 3px 8px;
    }
    .meta-icon { font-size: 0.85rem; }

    /* Code pill */
    .exam-code-row { display: flex; align-items: center; gap: 9px; flex-wrap: wrap; }
    .code-pill {
        display: flex; align-items: center; gap: 6px;
        background: linear-gradient(135deg, rgba(99,102,241,0.1), rgba(139,92,246,0.06));
        border: 1.5px solid rgba(99,102,241,0.18); border-radius: 9px;
        padding: 5px 12px;
    }
    .code-label { font-size: 0.68rem; font-weight: 800; color: var(--primary); text-transform: uppercase; letter-spacing: 0.08em; }
    .code-value { font-family: 'Courier New', monospace; font-size: 1.05rem; font-weight: 800; color: var(--primary); letter-spacing: 0.1em; }
    .copy-btn {
        padding: 5px 11px; background: white; border: 1.5px solid var(--border);
        border-radius: 7px; font-size: 0.78rem; font-weight: 600; color: var(--text2);
        cursor: pointer; transition: var(--transition);
    }
    .copy-btn:hover { border-color: var(--primary); color: var(--primary); }

    /* Card actions */
    .exam-actions { display: flex; gap: 6px; flex-wrap: wrap; margin-top: 2px; }
    .action-btn {
        padding: 6px 12px; border: none; border-radius: 7px; cursor: pointer;
        font-size: 0.78rem; font-weight: 700; transition: var(--transition);
        display: inline-flex; align-items: center; gap: 4px; font-family: inherit;
    }
    .action-btn:disabled { opacity: 0.5; cursor: not-allowed; }
    .btn-success  { background: #dcfce7; color: #166534; }
    .btn-success:hover:not(:disabled) { background: #bbf7d0; }
    .btn-warning  { background: #fef9c3; color: #854d0e; }
    .btn-warning:hover:not(:disabled) { background: #fef08a; }
    .btn-view     { background: #eff6ff; color: #1d4ed8; }
    .btn-view:hover:not(:disabled) { background: #dbeafe; }
    .btn-icon     { padding: 6px 9px; background: var(--bg2); color: var(--text2); }
    .btn-icon:hover:not(:disabled) { background: var(--border); }
    .btn-danger   { background: #fee2e2; color: var(--danger); }
    .btn-danger:hover:not(:disabled) { background: #fecaca; }

    /* ── Empty state ── */
    .empty-state {
        padding: 60px 20px; text-align: center; background: white;
        border-radius: var(--radius-lg); box-shadow: var(--shadow-sm);
        display: flex; flex-direction: column; align-items: center;
    }
    .empty-illustration { font-size: 4rem; margin-bottom: 14px; line-height: 1; }
    .empty-title { font-size: 1.1rem; font-weight: 800; color: var(--text); }
    .empty-sub { font-size: 0.85rem; color: var(--text3); margin-top: 6px; max-width: 340px; line-height: 1.5; }

    /* ── Share modal ── */
    .share-body { display: flex; flex-direction: column; gap: 16px; margin-bottom: 16px; }
    .share-url-box {
        display: flex; align-items: center; gap: 10px;
        background: var(--bg2); border: 1.5px solid var(--border);
        border-radius: var(--radius-sm); padding: 10px 14px; flex-wrap: wrap;
    }
    .share-url-text {
        font-size: 0.82rem; color: var(--primary); word-break: break-all; flex: 1;
        font-family: 'Courier New', monospace;
    }
    .share-copy-btn {
        padding: 6px 14px; background: var(--primary); color: white;
        border: none; border-radius: 7px; font-size: 0.82rem; font-weight: 700;
        cursor: pointer; transition: var(--transition); white-space: nowrap;
    }
    .share-copy-btn:hover { background: var(--accent); }
    .share-code-badge {
        display: flex; flex-direction: column; align-items: center; gap: 4px;
        background: linear-gradient(135deg, rgba(99,102,241,0.08), rgba(139,92,246,0.06));
        border: 2px solid rgba(99,102,241,0.2); border-radius: var(--radius); padding: 14px;
    }
    .scb-label { font-size: 0.7rem; font-weight: 800; color: var(--primary); text-transform: uppercase; letter-spacing: 0.1em; }
    .scb-code { font-family: 'Courier New', monospace; font-size: 2rem; font-weight: 900; color: var(--primary); letter-spacing: 0.18em; }

    /* QR code */
    .qr-wrap { display: flex; flex-direction: column; align-items: center; gap: 8px; }
    .qr-label { font-size: 0.78rem; font-weight: 700; color: var(--text3); }
    .qr-grid { border: 4px solid #0f172a; border-radius: 4px; overflow: hidden; display: inline-flex; flex-direction: column; background: white; }
    .qr-row { display: flex; }
    .qr-cell { width: 10px; height: 10px; background: white; }
    .qr-cell.filled { background: #0f172a; }

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
    .modal-lg { max-width: 700px; }
    .modal-header {
        display: flex; align-items: flex-start; justify-content: space-between;
        margin-bottom: 20px;
    }
    .modal-close {
        width: 32px; height: 32px; border-radius: 8px;
        background: var(--bg2); border: none; cursor: pointer;
        font-size: 0.9rem; color: var(--text3); flex-shrink: 0;
        display: flex; align-items: center; justify-content: center;
        transition: var(--transition);
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

    .toggle-row {
        display: flex; align-items: center; justify-content: space-between;
        background: var(--bg2); border-radius: var(--radius-sm); padding: 11px 14px;
        cursor: pointer; border: 1.5px solid var(--border);
    }
    .toggle-label { display: flex; align-items: center; gap: 8px; font-size: 0.88rem; font-weight: 600; color: var(--text); }
    .toggle-wrap { position: relative; flex-shrink: 0; }
    .toggle-input { display: none; }
    .toggle-knob {
        display: block; width: 40px; height: 22px;
        background: var(--border); border-radius: 99px; cursor: pointer; position: relative; transition: background 0.22s;
    }
    .toggle-knob::after {
        content: ''; position: absolute; top: 3px; left: 3px;
        width: 16px; height: 16px; border-radius: 50%; background: white;
        transition: transform 0.22s; box-shadow: 0 1px 4px rgba(0,0,0,0.2);
    }
    .toggle-input:checked + .toggle-knob { background: var(--primary); }
    .toggle-input:checked + .toggle-knob::after { transform: translateX(18px); }

    /* Results modal */
    .results-summary {
        display: flex; gap: 0; margin-bottom: 16px;
        background: var(--bg2); border-radius: var(--radius); overflow: hidden;
        border: 1.5px solid var(--border);
    }
    .summary-item {
        flex: 1; padding: 14px 16px; text-align: center;
        border-right: 1.5px solid var(--border);
        display: flex; flex-direction: column; gap: 4px;
    }
    .summary-item:last-child { border-right: none; }
    .summary-value { font-size: 1.6rem; font-weight: 800; }
    .summary-label { font-size: 0.75rem; color: var(--text3); }

    /* Score distribution chart */
    .dist-chart {
        background: var(--bg2); border: 1.5px solid var(--border);
        border-radius: var(--radius); padding: 14px 16px; margin-bottom: 12px;
        display: flex; flex-direction: column; gap: 9px;
    }
    .dist-title { font-size: 0.8rem; font-weight: 700; color: var(--text2); text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 4px; }
    .dist-row { display: flex; align-items: center; gap: 10px; }
    .dist-label { font-size: 0.77rem; font-weight: 600; color: var(--text3); width: 60px; flex-shrink: 0; }
    .dist-bar-wrap {
        flex: 1; height: 16px; background: white;
        border-radius: 4px; overflow: hidden; border: 1px solid var(--border);
    }
    .dist-bar-fill { height: 100%; border-radius: 4px; transition: width 0.5s ease; }
    .dist-count { font-size: 0.8rem; font-weight: 700; color: var(--text2); width: 24px; text-align: right; }

    .results-table-wrap { overflow-x: auto; margin-bottom: 4px; }
    .results-table { width: 100%; border-collapse: collapse; }
    .results-table th {
        padding: 9px 12px; text-align: left; font-size: 0.77rem; font-weight: 700;
        color: var(--text3); text-transform: uppercase; background: var(--bg2);
        border-bottom: 1.5px solid var(--border);
    }
    .results-table td {
        padding: 10px 12px; font-size: 0.87rem; border-bottom: 1px solid var(--border);
    }
    .results-table tbody tr:last-child td { border-bottom: none; }
    .results-table tbody tr:hover { background: #fafbff; }
    .td-rank { font-weight: 800; color: var(--text3); width: 32px; }
    .td-name { font-weight: 600; }
    .td-score { font-weight: 700; font-family: monospace; }
    .td-date { color: var(--text3); font-size: 0.8rem; white-space: nowrap; }

    .percent-bar {
        height: 5px; background: #e2e8f0; border-radius: 99px; overflow: hidden;
        width: 70px; display: inline-block; vertical-align: middle; margin-right: 6px;
    }
    .percent-fill { height: 100%; border-radius: 99px; transition: width 0.4s; }
    .percent-text { font-size: 0.82rem; font-weight: 700; vertical-align: middle; }

    /* Buttons */
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

    /* Responsive */
    @media (max-width: 600px) {
        .form-row-2 { grid-template-columns: 1fr; }
        .exam-grid { grid-template-columns: 1fr; }
        .stats-strip { flex-direction: column; }
        .filter-tabs { gap: 4px; }
    }
</style>
