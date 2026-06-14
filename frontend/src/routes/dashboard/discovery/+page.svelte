<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { quizzes as quizzesApi, rooms as roomsApi } from '$lib/api/client';
    import type { Quiz } from '$lib/api/types';
    import { allCategories } from '$lib/data/categories';
    import type { Category, Test } from '$lib/data/categories';

    // ── Tabs ───────────────────────────────────────────────────────────────────
    let activeTab: 'categories' | 'teachers' = 'categories';

    // ── Category section state ─────────────────────────────────────────────────
    let selectedCat: Category = allCategories[0];
    let selectedSubcat = 'all';
    let catSearch = '';

    $: catTests = selectedCat.tests.filter(t => {
        const matchSub = selectedSubcat === 'all' || t.subcat === selectedSubcat;
        const matchSearch = !catSearch || t.title.toLowerCase().includes(catSearch.toLowerCase())
            || t.description.toLowerCase().includes(catSearch.toLowerCase());
        return matchSub && matchSearch;
    });

    $: catSortedTests = [...catTests].sort((a, b) => b.plays - a.plays);

    // Total tests count across all categories
    $: totalTests = allCategories.reduce((s, c) => s + c.tests.length, 0);

    // ── Teacher quizzes state ──────────────────────────────────────────────────
    let quizzes: Quiz[] = [];
    let loadingQuizzes = true;
    let search = '';
    let subjectFilter = '';
    let gradeFilter = '';
    let sortBy: 'plays' | 'new' | 'questions' = 'plays';
    let searchTimer: ReturnType<typeof setTimeout>;

    async function loadQuizzes() {
        loadingQuizzes = true;
        try {
            quizzes = await quizzesApi.discover({
                q: search || undefined,
                subject: subjectFilter || undefined,
                grade: gradeFilter || undefined,
            });
        } catch {
            quizzes = [];
        } finally {
            loadingQuizzes = false;
        }
    }

    onMount(() => loadQuizzes());

    $: { search; clearTimeout(searchTimer); searchTimer = setTimeout(loadQuizzes, 300); }
    $: { subjectFilter; gradeFilter; loadQuizzes(); }

    $: sortedQuizzes = [...quizzes].sort((a, b) => {
        if (sortBy === 'questions') return b.total_questions - a.total_questions;
        if (sortBy === 'new') return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
        return b.play_count - a.play_count;
    });

    // ── Game modal ─────────────────────────────────────────────────────────────
    let showGameModal = false;
    let modalQuizTitle = '';
    let modalQuizId = '';
    let gameMode = 'classic';
    let creatingRoom = false;
    let shuffleQuestions = false;
    let showLeaderboard = true;

    function openQuizGameModal(quiz: Quiz) {
        modalQuizId = quiz.id;
        modalQuizTitle = quiz.title;
        gameMode = 'classic';
        shuffleQuestions = false;
        showLeaderboard = true;
        showGameModal = true;
    }

    function openCatGameModal(test: Test) {
        // Katalog (ochiq testlar) sahifasiga yo'naltirish
        goto('/dashboard/open-tests');
    }

    function closeModal() { showGameModal = false; }

    async function startGame() {
        if (!modalQuizId) return;
        creatingRoom = true;
        try {
            const res = await roomsApi.create(modalQuizId, gameMode, {
                shuffle_questions: shuffleQuestions,
                shuffle_answers: true,
                show_leaderboard: showLeaderboard,
                music: true,
                lobby_music: true,
                show_correct_answer: true,
            });
            goto(`/game/host/${res.pin}`);
        } catch (e: any) {
            alert(e.message || 'Xato yuz berdi');
        } finally {
            creatingRoom = false;
            showGameModal = false;
        }
    }

    function handleKey(e: KeyboardEvent) {
        if (e.key === 'Escape') closeModal();
    }

    // ── Helpers ────────────────────────────────────────────────────────────────
    function quizGradient(id: string) {
        const h = (id.charCodeAt(0) * 7 + id.charCodeAt(1) * 3) % 360;
        return `linear-gradient(135deg, hsl(${h},65%,52%), hsl(${(h+45)%360},65%,42%))`;
    }

    function diffColor(d: string) {
        return d === 'easy' ? '#22c55e' : d === 'medium' ? '#f59e0b' : '#ef4444';
    }
    function diffLabel(d: string) {
        return d === 'easy' ? 'Oson' : d === 'medium' ? "O'rta" : 'Qiyin';
    }

    function stars(rating: number) {
        return '★'.repeat(Math.round(rating)) + '☆'.repeat(5 - Math.round(rating));
    }

    function formatPlays(n: number) {
        return n >= 1000 ? `${(n/1000).toFixed(1)}k` : String(n);
    }

    const modes = [
        { value: 'classic',     icon: '⚡', label: 'Classic',  color: '#f59e0b' },
        { value: 'self_paced',  icon: '📲', label: 'Mustaqil', color: '#06b6d4' },
        { value: 'team',        icon: '👥', label: 'Jamoaviy', color: '#3b82f6' },
        { value: 'accuracy',    icon: '🎯', label: 'Aniqlik',  color: '#22c55e' },
        { value: 'confidence',  icon: '💡', label: 'Ishonch',  color: '#8b5cf6' },
        { value: 'zero_stakes', icon: '🌟', label: 'Mashq',    color: '#64748b' },
    ];

    const subjects = ['Matematika','Fizika','Kimyo','Biologiya','Tarix','Geografiya','Ingliz tili','Ona tili','Informatika'];
    const grades = Array.from({ length: 11 }, (_, i) => `${i + 1}`);
</script>

<svelte:head><title>Discovery — Cognita.uz</title></svelte:head>
<svelte:window on:keydown={handleKey} />

<!-- ── Hero Banner ─────────────────────────────────────────────────────────── -->
<div class="hero">
    <div class="hero-content">
        <div class="hero-text">
            <h1>Discovery</h1>
            <p>Admin kategoriyalari va o'qituvchilar yaratgan testlarni toping</p>
        </div>
        <div class="hero-stats">
            <div class="hstat">
                <span class="hstat-num">{totalTests}</span>
                <span class="hstat-label">Kategoriya testi</span>
            </div>
            <div class="hstat-divider"></div>
            <div class="hstat">
                <span class="hstat-num">{quizzes.length}</span>
                <span class="hstat-label">O'qituvchi quizi</span>
            </div>
            <div class="hstat-divider"></div>
            <div class="hstat">
                <span class="hstat-num">{allCategories.length}</span>
                <span class="hstat-label">Kategoriya</span>
            </div>
        </div>
    </div>
    <div class="hero-orbs">
        <div class="orb orb1"></div>
        <div class="orb orb2"></div>
        <div class="orb orb3"></div>
    </div>
</div>

<!-- ── Tab Nav ─────────────────────────────────────────────────────────────── -->
<div class="tab-nav">
    <button
        class="tab-btn" class:active={activeTab === 'categories'}
        on:click={() => activeTab = 'categories'}
    >
        <span class="tab-icon-lg">📂</span>
        <div class="tab-info">
            <span class="tab-title">Admin Kategoriyalar</span>
            <span class="tab-sub">{totalTests} ta test · {allCategories.length} ta bo'lim</span>
        </div>
    </button>
    <button
        class="tab-btn" class:active={activeTab === 'teachers'}
        on:click={() => { activeTab = 'teachers'; }}
    >
        <span class="tab-icon-lg">👩‍🏫</span>
        <div class="tab-info">
            <span class="tab-title">O'qituvchi Quizlari</span>
            <span class="tab-sub">{quizzes.length} ta ochiq quiz</span>
        </div>
        <span class="tab-badge">Yangi</span>
    </button>
</div>

<!-- ═══════════════════════════════════════════════════════════════
     TAB 1 — CATEGORIES
════════════════════════════════════════════════════════════════ -->
{#if activeTab === 'categories'}

    <!-- Category pills row -->
    <div class="cat-pills-row">
        {#each allCategories as cat}
            <button
                class="cat-pill"
                class:active={selectedCat.id === cat.id}
                on:click={() => { selectedCat = cat; selectedSubcat = 'all'; catSearch = ''; }}
                style="--cc1:{cat.g1}; --cc2:{cat.g2}"
            >
                <span class="cp-icon">{cat.icon}</span>
                <span class="cp-label">{cat.title}</span>
                <span class="cp-count">{cat.tests.length}</span>
            </button>
        {/each}
    </div>

    <!-- Selected category header -->
    <div class="cat-header" style="--cc1:{selectedCat.g1}; --cc2:{selectedCat.g2}">
        <div class="cat-header-left">
            <div class="cat-big-icon">{selectedCat.icon}</div>
            <div>
                <h2 class="cat-title">{selectedCat.title}</h2>
                <p class="cat-subtitle">{selectedCat.subtitle}</p>
            </div>
        </div>
        <div class="cat-header-right">
            <div class="cat-header-search">
                <span>🔍</span>
                <input
                    type="text"
                    placeholder="Test qidirish..."
                    bind:value={catSearch}
                    class="cat-search-input"
                />
            </div>
        </div>
    </div>

    <!-- Subcategory filter -->
    <div class="subcat-row">
        {#each selectedCat.subcats as sub}
            <button
                class="subcat-btn"
                class:active={selectedSubcat === sub.id}
                on:click={() => selectedSubcat = sub.id}
            >
                {sub.icon} {sub.label}
            </button>
        {/each}
    </div>

    <!-- Results info -->
    <div class="results-bar">
        <span class="results-txt">{catSortedTests.length} ta test</span>
        {#if catSearch}
            <button class="clear-search-btn" on:click={() => catSearch = ''}>✕ Tozalash</button>
        {/if}
    </div>

    <!-- Tests grid -->
    {#if catSortedTests.length === 0}
        <div class="empty-state">
            <div class="empty-emoji">🔍</div>
            <p class="empty-title">Test topilmadi</p>
            <p class="empty-sub">Boshqa kalit so'z yoki bo'limni sinab ko'ring</p>
            <button class="btn secondary" on:click={() => { catSearch = ''; selectedSubcat = 'all'; }}>
                Filtrlarni tozalash
            </button>
        </div>
    {:else}
        <div class="tests-grid">
            {#each catSortedTests as test (test.id)}
                <div class="test-card">
                    <!-- Difficulty strip -->
                    <div class="test-strip" style="background:{diffColor(test.difficulty)}"></div>

                    <div class="test-body">
                        <div class="test-top">
                            <span class="test-icon">{test.icon}</span>
                            <div class="test-badges">
                                {#if test.isNew}<span class="badge new-badge">Yangi</span>{/if}
                                {#if test.isHot}<span class="badge hot-badge">🔥 Trend</span>{/if}
                                <span class="badge diff-badge" style="color:{diffColor(test.difficulty)};background:{diffColor(test.difficulty)}18">
                                    {diffLabel(test.difficulty)}
                                </span>
                            </div>
                        </div>

                        <h3 class="test-title">{test.title}</h3>
                        <p class="test-desc">{test.description}</p>

                        <div class="test-meta-row">
                            <span class="tmeta">📌 {test.questions} savol</span>
                            <span class="tmeta">⏱ {test.duration} min</span>
                            <span class="tmeta plays">▶ {formatPlays(test.plays)}</span>
                        </div>

                        <div class="test-footer">
                            <div class="rating">
                                <span class="stars">{stars(test.rating)}</span>
                                <span class="rating-num">{test.rating.toFixed(1)}</span>
                            </div>
                            <button class="btn primary sm" on:click={() => openCatGameModal(test)}>
                                O'ynash →
                            </button>
                        </div>
                    </div>
                </div>
            {/each}
        </div>
    {/if}

{/if}

<!-- ═══════════════════════════════════════════════════════════════
     TAB 2 — TEACHER QUIZZES
════════════════════════════════════════════════════════════════ -->
{#if activeTab === 'teachers'}

    <!-- Toolbar -->
    <div class="toolbar">
        <div class="search-wrap">
            <span class="search-icon">🔍</span>
            <input
                type="text"
                placeholder="Quiz qidirish..."
                bind:value={search}
                class="search-input"
            />
            {#if search}
                <button class="clear-btn" on:click={() => search = ''}>✕</button>
            {/if}
        </div>

        <select class="filter-select" bind:value={subjectFilter}>
            <option value="">Fan bo'yicha</option>
            {#each subjects as s}<option value={s}>{s}</option>{/each}
        </select>

        <select class="filter-select" bind:value={gradeFilter}>
            <option value="">Sinf bo'yicha</option>
            {#each grades as g}<option value={g}>{g}-sinf</option>{/each}
        </select>

        <div class="sort-pills">
            <button class="pill" class:active={sortBy === 'plays'}     on:click={() => sortBy = 'plays'}>Ko'p o'ynalgan</button>
            <button class="pill" class:active={sortBy === 'new'}       on:click={() => sortBy = 'new'}>Yangi</button>
            <button class="pill" class:active={sortBy === 'questions'} on:click={() => sortBy = 'questions'}>Ko'p savol</button>
        </div>
    </div>

    {#if !loadingQuizzes}
        <p class="results-txt" style="margin-bottom:16px">
            <strong>{sortedQuizzes.length}</strong> ta quiz topildi
        </p>
    {/if}

    {#if loadingQuizzes}
        <div class="quiz-grid">
            {#each Array(8) as _}
                <div class="skeleton-card">
                    <div class="skel-cover"></div>
                    <div class="skel-body">
                        <div class="skel-line wide"></div>
                        <div class="skel-line medium"></div>
                        <div class="skel-line short"></div>
                    </div>
                </div>
            {/each}
        </div>
    {:else if sortedQuizzes.length === 0}
        <div class="empty-state">
            <div class="empty-emoji">🔭</div>
            <p class="empty-title">
                {#if search || subjectFilter || gradeFilter}Hech narsa topilmadi{:else}Hali ommaviy quiz mavjud emas{/if}
            </p>
            <p class="empty-sub">
                {#if search || subjectFilter || gradeFilter}
                    Boshqa kalit so'z yoki filtrni sinab ko'ring
                {:else}
                    O'qituvchilar quizlarini ommaviy qilib qo'yganda shu yerda ko'rinadi
                {/if}
            </p>
            {#if search || subjectFilter || gradeFilter}
                <button class="btn secondary" on:click={() => { search = ''; subjectFilter = ''; gradeFilter = ''; }}>
                    Filtrlarni tozalash
                </button>
            {/if}
        </div>
    {:else}
        <div class="quiz-grid">
            {#each sortedQuizzes as quiz (quiz.id)}
                <div class="quiz-card">
                    <div class="quiz-cover" style="background:{quizGradient(quiz.id)}">
                        <span class="cover-letter">{quiz.title[0]?.toUpperCase()}</span>
                        <span class="public-badge">Ochiq</span>
                        {#if quiz.play_count > 100}
                            <span class="hot-tag">🔥</span>
                        {/if}
                    </div>
                    <div class="quiz-body">
                        <div class="quiz-title-text">{quiz.title}</div>
                        {#if quiz.description}
                            <p class="quiz-desc">{quiz.description}</p>
                        {/if}
                        <div class="quiz-tags">
                            {#if quiz.subject}<span class="tag subject-tag">{quiz.subject}</span>{/if}
                            {#if quiz.grade_level}<span class="tag grade-tag">{quiz.grade_level}-sinf</span>{/if}
                        </div>
                        <div class="quiz-meta">
                            <span>📌 {quiz.total_questions} savol</span>
                            <span>▶ {formatPlays(quiz.play_count)} o'yin</span>
                        </div>
                        <div class="quiz-actions">
                            <button class="btn primary sm" on:click={() => openQuizGameModal(quiz)}>
                                ▶ O'yin boshlash
                            </button>
                        </div>
                    </div>
                </div>
            {/each}
        </div>
    {/if}

{/if}

<!-- ── Game Modal ──────────────────────────────────────────────────────────── -->
{#if showGameModal}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="modal-overlay" on:click|self={closeModal}>
        <div class="modal">
            <div class="modal-header">
                <div>
                    <p class="modal-pre">O'yinni boshlash</p>
                    <h2 class="modal-title">{modalQuizTitle}</h2>
                </div>
                <button class="modal-close" on:click={closeModal}>✕</button>
            </div>
            <div class="modal-body">
                <p class="section-label">O'yin rejimi</p>
                <div class="modes-grid">
                    {#each modes as mode}
                        <button
                            class="mode-card"
                            class:selected={gameMode === mode.value}
                            on:click={() => gameMode = mode.value}
                            style="--mc:{mode.color}"
                        >
                            <span class="mode-icon">{mode.icon}</span>
                            <span class="mode-name">{mode.label}</span>
                        </button>
                    {/each}
                </div>
                <p class="section-label" style="margin-top:18px">Sozlamalar</p>
                <div class="settings-row">
                    <label class="toggle-label">
                        <span>Savollarni aralashtirish</span>
                        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
                        <div class="toggle" class:on={shuffleQuestions} on:click={() => shuffleQuestions = !shuffleQuestions}>
                            <div class="toggle-thumb"></div>
                        </div>
                    </label>
                    <label class="toggle-label">
                        <span>Liderlar jadvali</span>
                        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
                        <div class="toggle" class:on={showLeaderboard} on:click={() => showLeaderboard = !showLeaderboard}>
                            <div class="toggle-thumb"></div>
                        </div>
                    </label>
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn secondary" on:click={closeModal} disabled={creatingRoom}>Bekor qilish</button>
                <button class="btn primary start-btn" on:click={startGame} disabled={creatingRoom}>
                    {#if creatingRoom}<span class="spinner"></span> Yaratilmoqda...{:else}▶ Boshlash{/if}
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
    /* ── Hero ──────────────────────────────────────────────────────────────── */
    .hero {
        position: relative;
        background: linear-gradient(135deg, #0f172a 0%, #1e1b4b 50%, #0f172a 100%);
        border-radius: 20px;
        padding: 36px 40px;
        margin-bottom: 24px;
        overflow: hidden;
        color: white;
    }
    .hero-content {
        position: relative;
        z-index: 2;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 24px;
        flex-wrap: wrap;
    }
    .hero-text h1 {
        font-size: 2.4rem;
        font-weight: 900;
        margin: 0 0 6px;
        background: linear-gradient(135deg, #c7d2fe, #a5b4fc, #818cf8);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
        letter-spacing: -1px;
    }
    .hero-text p {
        font-size: 0.9rem;
        color: rgba(255,255,255,0.6);
        margin: 0;
    }
    .hero-stats {
        display: flex;
        gap: 0;
        background: rgba(255,255,255,0.07);
        border: 1px solid rgba(255,255,255,0.1);
        border-radius: 16px;
        padding: 16px 24px;
        gap: 24px;
        align-items: center;
        backdrop-filter: blur(10px);
    }
    .hstat { text-align: center; }
    .hstat-num {
        display: block;
        font-size: 1.6rem;
        font-weight: 800;
        color: #a5b4fc;
        line-height: 1;
    }
    .hstat-label {
        display: block;
        font-size: 0.72rem;
        color: rgba(255,255,255,0.5);
        margin-top: 3px;
    }
    .hstat-divider {
        width: 1px;
        height: 32px;
        background: rgba(255,255,255,0.15);
    }
    .hero-orbs {
        position: absolute;
        inset: 0;
        pointer-events: none;
        z-index: 1;
        overflow: hidden;
    }
    .orb {
        position: absolute;
        border-radius: 50%;
        filter: blur(60px);
        opacity: 0.25;
    }
    .orb1 { width: 300px; height: 300px; background: #6366f1; top: -80px; right: 10%; }
    .orb2 { width: 200px; height: 200px; background: #8b5cf6; bottom: -60px; left: 20%; }
    .orb3 { width: 150px; height: 150px; background: #06b6d4; top: 20px; left: 5%; }

    /* ── Tab nav ──────────────────────────────────────────────────────────── */
    .tab-nav {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
        margin-bottom: 24px;
    }
    .tab-btn {
        display: flex;
        align-items: center;
        gap: 14px;
        padding: 18px 22px;
        background: var(--bg);
        border: 2px solid var(--border);
        border-radius: 16px;
        cursor: pointer;
        transition: all 0.2s;
        text-align: left;
        position: relative;
    }
    .tab-btn:hover { border-color: var(--primary); background: var(--bg2); }
    .tab-btn.active {
        border-color: var(--primary);
        background: linear-gradient(135deg, rgba(99,102,241,0.06), rgba(139,92,246,0.06));
        box-shadow: 0 4px 20px rgba(99,102,241,0.12);
    }
    .tab-icon-lg { font-size: 2rem; flex-shrink: 0; }
    .tab-info { display: flex; flex-direction: column; gap: 2px; }
    .tab-title {
        font-size: 1rem;
        font-weight: 700;
        color: var(--text);
    }
    .tab-btn.active .tab-title { color: var(--primary); }
    .tab-sub { font-size: 0.78rem; color: var(--text3); }
    .tab-badge {
        position: absolute;
        top: 12px;
        right: 14px;
        background: linear-gradient(135deg, var(--primary), var(--accent));
        color: white;
        font-size: 0.65rem;
        font-weight: 700;
        padding: 2px 8px;
        border-radius: 99px;
    }

    /* ── Category pills ────────────────────────────────────────────────────── */
    .cat-pills-row {
        display: flex;
        gap: 8px;
        margin-bottom: 20px;
        flex-wrap: wrap;
    }
    .cat-pill {
        display: flex;
        align-items: center;
        gap: 7px;
        padding: 8px 16px;
        border: 2px solid var(--border);
        border-radius: 12px;
        background: var(--bg);
        cursor: pointer;
        transition: all 0.2s;
        font-size: 0.85rem;
        font-weight: 600;
        color: var(--text2);
    }
    .cat-pill:hover {
        border-color: var(--cc1);
        color: var(--cc1);
    }
    .cat-pill.active {
        background: linear-gradient(135deg, var(--cc1), var(--cc2));
        border-color: transparent;
        color: white;
        box-shadow: 0 4px 14px rgba(0,0,0,0.15);
    }
    .cp-icon { font-size: 1.1rem; }
    .cp-label { white-space: nowrap; }
    .cp-count {
        background: rgba(255,255,255,0.25);
        border-radius: 99px;
        padding: 1px 7px;
        font-size: 0.72rem;
    }
    .cat-pill:not(.active) .cp-count {
        background: var(--bg2);
        color: var(--text3);
    }

    /* ── Category header ───────────────────────────────────────────────────── */
    .cat-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 16px;
        padding: 22px 26px;
        background: linear-gradient(135deg, var(--cc1), var(--cc2));
        border-radius: 16px;
        margin-bottom: 16px;
        color: white;
        flex-wrap: wrap;
    }
    .cat-header-left { display: flex; align-items: center; gap: 16px; }
    .cat-big-icon {
        font-size: 2.8rem;
        background: rgba(255,255,255,0.2);
        width: 64px;
        height: 64px;
        border-radius: 16px;
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    .cat-title { font-size: 1.3rem; font-weight: 800; margin: 0 0 4px; }
    .cat-subtitle { font-size: 0.82rem; opacity: 0.85; margin: 0; max-width: 340px; }
    .cat-header-search {
        display: flex;
        align-items: center;
        gap: 8px;
        background: rgba(255,255,255,0.15);
        border: 1px solid rgba(255,255,255,0.25);
        border-radius: 10px;
        padding: 8px 14px;
        backdrop-filter: blur(8px);
    }
    .cat-search-input {
        background: none;
        border: none;
        outline: none;
        color: white;
        font-size: 0.875rem;
        width: 180px;
    }
    .cat-search-input::placeholder { color: rgba(255,255,255,0.6); }

    /* ── Subcategory row ───────────────────────────────────────────────────── */
    .subcat-row {
        display: flex;
        gap: 6px;
        margin-bottom: 14px;
        flex-wrap: wrap;
    }
    .subcat-btn {
        padding: 5px 14px;
        border: 1.5px solid var(--border);
        border-radius: 99px;
        background: var(--bg);
        font-size: 0.8rem;
        font-weight: 600;
        color: var(--text2);
        cursor: pointer;
        transition: all 0.2s;
    }
    .subcat-btn:hover { border-color: var(--primary); color: var(--primary); }
    .subcat-btn.active {
        background: var(--primary);
        border-color: var(--primary);
        color: white;
    }

    /* ── Results bar ───────────────────────────────────────────────────────── */
    .results-bar {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 16px;
    }
    .results-txt { font-size: 0.82rem; color: var(--text3); }
    .results-txt strong { color: var(--text2); }
    .clear-search-btn {
        background: none;
        border: 1px solid var(--border);
        border-radius: 6px;
        font-size: 0.78rem;
        color: var(--text3);
        cursor: pointer;
        padding: 3px 10px;
        transition: all 0.15s;
    }
    .clear-search-btn:hover { border-color: var(--primary); color: var(--primary); }

    /* ── Tests grid ────────────────────────────────────────────────────────── */
    .tests-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 16px;
    }
    .test-card {
        background: var(--bg);
        border-radius: 14px;
        border: 1.5px solid var(--border);
        overflow: hidden;
        display: flex;
        flex-direction: column;
        transition: all 0.2s;
        position: relative;
    }
    .test-card:hover {
        transform: translateY(-3px);
        box-shadow: 0 8px 28px rgba(0,0,0,0.1);
        border-color: var(--primary-light, #e0e7ff);
    }
    .test-strip {
        height: 4px;
        width: 100%;
        flex-shrink: 0;
    }
    .test-body { padding: 16px; display: flex; flex-direction: column; gap: 10px; flex: 1; }
    .test-top {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 8px;
    }
    .test-icon { font-size: 1.8rem; flex-shrink: 0; }
    .test-badges { display: flex; gap: 5px; flex-wrap: wrap; justify-content: flex-end; }
    .badge {
        font-size: 0.65rem;
        font-weight: 700;
        padding: 2px 7px;
        border-radius: 99px;
    }
    .new-badge { background: #dbeafe; color: #1d4ed8; }
    .hot-badge { background: #fef3c7; color: #92400e; }
    .diff-badge { font-weight: 700; }

    .test-title {
        font-size: 0.95rem;
        font-weight: 700;
        color: var(--text);
        margin: 0;
        line-height: 1.35;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }
    .test-desc {
        font-size: 0.78rem;
        color: var(--text3);
        margin: 0;
        line-height: 1.5;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }
    .test-meta-row {
        display: flex;
        gap: 10px;
        font-size: 0.72rem;
        color: var(--text3);
        flex-wrap: wrap;
    }
    .tmeta.plays { color: var(--primary); font-weight: 600; }
    .test-footer {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-top: auto;
        padding-top: 8px;
        border-top: 1px solid var(--border);
    }
    .rating { display: flex; align-items: center; gap: 5px; }
    .stars { font-size: 0.75rem; color: #f59e0b; letter-spacing: -1px; }
    .rating-num { font-size: 0.75rem; font-weight: 700; color: var(--text2); }

    /* ── Teacher quizzes toolbar ───────────────────────────────────────────── */
    .toolbar {
        display: flex;
        align-items: center;
        gap: 10px;
        margin-bottom: 14px;
        flex-wrap: wrap;
    }
    .search-wrap {
        position: relative;
        display: flex;
        align-items: center;
        flex: 1;
        min-width: 200px;
        max-width: 300px;
    }
    .search-icon {
        position: absolute;
        left: 10px;
        font-size: 0.85rem;
        pointer-events: none;
    }
    .search-input {
        width: 100%;
        padding: 8px 32px;
        border: 1.5px solid var(--border);
        border-radius: var(--radius);
        font-size: 0.875rem;
        outline: none;
        transition: border-color 0.2s;
        background: var(--bg);
        color: var(--text);
    }
    .search-input:focus { border-color: var(--primary); }
    .clear-btn {
        position: absolute;
        right: 8px;
        background: none;
        border: none;
        cursor: pointer;
        color: var(--text3);
        font-size: 0.8rem;
    }
    .filter-select {
        padding: 8px 12px;
        border: 1.5px solid var(--border);
        border-radius: var(--radius);
        font-size: 0.875rem;
        outline: none;
        background: var(--bg);
        color: var(--text);
        cursor: pointer;
    }
    .filter-select:focus { border-color: var(--primary); }
    .sort-pills {
        display: flex;
        gap: 4px;
        background: var(--bg2);
        padding: 4px;
        border-radius: 10px;
        margin-left: auto;
    }
    .pill {
        padding: 5px 12px;
        border: none;
        border-radius: 7px;
        font-size: 0.8rem;
        font-weight: 600;
        cursor: pointer;
        background: transparent;
        color: var(--text2);
        transition: all 0.2s;
        white-space: nowrap;
    }
    .pill.active { background: var(--bg); color: var(--primary); box-shadow: 0 1px 4px rgba(0,0,0,0.08); }

    /* ── Quiz grid (teacher tab) ───────────────────────────────────────────── */
    .quiz-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 18px;
    }
    .quiz-card {
        background: var(--bg);
        border-radius: 14px;
        box-shadow: 0 1px 4px rgba(0,0,0,0.07);
        overflow: hidden;
        border: 1.5px solid transparent;
        display: flex;
        flex-direction: column;
        transition: all 0.2s;
    }
    .quiz-card:hover {
        transform: translateY(-4px);
        box-shadow: 0 8px 28px rgba(0,0,0,0.12);
        border-color: var(--primary-light, #e0e7ff);
    }
    .quiz-cover {
        height: 90px;
        position: relative;
        display: flex;
        align-items: center;
        justify-content: center;
    }
    .cover-letter {
        font-size: 2.6rem;
        font-weight: 900;
        color: rgba(255,255,255,0.85);
        text-shadow: 0 2px 8px rgba(0,0,0,0.2);
    }
    .public-badge {
        position: absolute;
        top: 8px;
        right: 8px;
        background: rgba(255,255,255,0.22);
        color: white;
        font-size: 0.68rem;
        font-weight: 700;
        padding: 2px 8px;
        border-radius: 99px;
        border: 1px solid rgba(255,255,255,0.35);
        backdrop-filter: blur(4px);
    }
    .hot-tag {
        position: absolute;
        top: 8px;
        left: 8px;
        font-size: 1rem;
    }
    .quiz-body {
        padding: 14px;
        display: flex;
        flex-direction: column;
        gap: 8px;
        flex: 1;
    }
    .quiz-title-text {
        font-size: 0.95rem;
        font-weight: 700;
        color: var(--text);
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
        line-height: 1.4;
    }
    .quiz-desc {
        font-size: 0.75rem;
        color: var(--text3);
        margin: 0;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
        line-height: 1.45;
    }
    .quiz-tags { display: flex; gap: 6px; flex-wrap: wrap; }
    .tag { font-size: 0.7rem; font-weight: 600; padding: 2px 8px; border-radius: 99px; }
    .subject-tag { background: #ede9fe; color: #5b21b6; }
    .grade-tag   { background: #f0fdf4; color: #15803d; }
    .quiz-meta { display: flex; gap: 10px; font-size: 0.75rem; color: var(--text3); }
    .quiz-actions { margin-top: auto; padding-top: 4px; }

    /* ── Skeleton ──────────────────────────────────────────────────────────── */
    .skeleton-card { background: var(--bg); border-radius: 14px; overflow: hidden; }
    .skel-cover {
        height: 90px;
        background: linear-gradient(90deg, var(--bg2) 25%, var(--border) 50%, var(--bg2) 75%);
        background-size: 200% 100%;
        animation: shimmer 1.4s infinite;
    }
    .skel-body { padding: 14px; display: flex; flex-direction: column; gap: 9px; }
    .skel-line {
        height: 11px;
        border-radius: 6px;
        background: linear-gradient(90deg, var(--bg2) 25%, var(--border) 50%, var(--bg2) 75%);
        background-size: 200% 100%;
        animation: shimmer 1.4s infinite;
    }
    .skel-line.wide { width: 75%; } .skel-line.medium { width: 50%; } .skel-line.short { width: 35%; }
    @keyframes shimmer { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }

    /* ── Empty state ───────────────────────────────────────────────────────── */
    .empty-state {
        background: var(--bg);
        border-radius: 16px;
        padding: 64px 32px;
        text-align: center;
        border: 1.5px solid var(--border);
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 8px;
    }
    .empty-emoji { font-size: 3rem; margin-bottom: 8px; }
    .empty-title { font-size: 1.1rem; font-weight: 700; color: var(--text); margin: 0; }
    .empty-sub   { font-size: 0.875rem; color: var(--text3); margin: 0 0 14px; }

    /* ── Modal ─────────────────────────────────────────────────────────────── */
    .modal-overlay {
        position: fixed;
        inset: 0;
        background: rgba(0,0,0,0.5);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 100;
        padding: 16px;
        animation: fadeIn 0.15s ease;
    }
    @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
    .modal {
        background: var(--bg);
        border-radius: 16px;
        width: 100%;
        max-width: 500px;
        box-shadow: 0 20px 60px rgba(0,0,0,0.25);
        animation: slideUp 0.2s ease;
        overflow: hidden;
        max-height: 90vh;
        overflow-y: auto;
    }
    @keyframes slideUp { from { transform: translateY(20px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
    .modal-header {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        padding: 22px 24px 0;
        gap: 12px;
    }
    .modal-pre { font-size: 0.72rem; font-weight: 600; color: var(--text3); text-transform: uppercase; letter-spacing: 0.07em; margin: 0 0 4px; }
    .modal-title { font-size: 1.1rem; font-weight: 800; color: var(--text); margin: 0; }
    .modal-close {
        background: var(--bg2);
        border: none;
        border-radius: 8px;
        width: 32px;
        height: 32px;
        cursor: pointer;
        color: var(--text3);
        display: flex;
        align-items: center;
        justify-content: center;
        transition: all 0.15s;
        flex-shrink: 0;
    }
    .modal-close:hover { background: var(--border); color: var(--text); }
    .modal-body { padding: 18px 24px; }
    .modal-footer {
        padding: 14px 24px;
        border-top: 1px solid var(--border);
        display: flex;
        justify-content: flex-end;
        gap: 10px;
        background: var(--bg2);
    }
    .section-label {
        font-size: 0.75rem;
        font-weight: 700;
        color: var(--text3);
        text-transform: uppercase;
        letter-spacing: 0.06em;
        margin: 0 0 10px;
    }
    .modes-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 8px; }
    .mode-card {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 5px;
        padding: 12px 4px;
        border: 2px solid var(--border);
        border-radius: var(--radius);
        background: var(--bg);
        cursor: pointer;
        transition: all 0.2s;
    }
    .mode-card:hover { border-color: var(--mc); }
    .mode-card.selected {
        border-color: var(--mc);
        background: color-mix(in srgb, var(--mc) 8%, white);
        box-shadow: 0 0 0 3px color-mix(in srgb, var(--mc) 15%, transparent);
    }
    .mode-icon { font-size: 1.4rem; }
    .mode-name { font-size: 0.7rem; font-weight: 700; color: var(--text); }
    .settings-row { display: flex; flex-direction: column; gap: 12px; }
    .toggle-label {
        display: flex;
        align-items: center;
        justify-content: space-between;
        font-size: 0.875rem;
        color: var(--text);
        cursor: pointer;
    }
    .toggle {
        width: 42px; height: 24px;
        border-radius: 12px;
        background: var(--border);
        position: relative;
        cursor: pointer;
        transition: background 0.2s;
        flex-shrink: 0;
    }
    .toggle.on { background: var(--primary); }
    .toggle-thumb {
        position: absolute;
        top: 3px; left: 3px;
        width: 18px; height: 18px;
        border-radius: 50%;
        background: white;
        box-shadow: 0 1px 4px rgba(0,0,0,0.2);
        transition: transform 0.2s;
    }
    .toggle.on .toggle-thumb { transform: translateX(18px); }
    .start-btn { min-width: 140px; justify-content: center; }

    /* ── Buttons ───────────────────────────────────────────────────────────── */
    .btn {
        padding: 9px 18px;
        border: none;
        border-radius: var(--radius, 8px);
        font-size: 0.875rem;
        font-weight: 600;
        cursor: pointer;
        display: inline-flex;
        align-items: center;
        gap: 6px;
        transition: all 0.2s;
        text-decoration: none;
    }
    .btn.sm { padding: 7px 13px; font-size: 0.8rem; }
    .btn.primary {
        background: linear-gradient(135deg, var(--primary), var(--accent));
        color: white;
        box-shadow: 0 3px 10px rgba(99,102,241,0.3);
    }
    .btn.primary:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 16px rgba(99,102,241,0.4); }
    .btn.secondary { background: var(--bg2); color: var(--text); border: 1px solid var(--border); }
    .btn.secondary:hover:not(:disabled) { background: var(--border); }
    .btn:disabled { opacity: 0.5; cursor: not-allowed; }
    .spinner {
        width: 14px; height: 14px;
        border: 2px solid rgba(255,255,255,0.4);
        border-top-color: white;
        border-radius: 50%;
        animation: spin 0.6s linear infinite;
        display: inline-block;
    }
    @keyframes spin { to { transform: rotate(360deg); } }

    /* ── Responsive ────────────────────────────────────────────────────────── */
    @media (max-width: 1100px) {
        .tests-grid { grid-template-columns: repeat(2, 1fr); }
        .quiz-grid  { grid-template-columns: repeat(2, 1fr); }
        .modes-grid { grid-template-columns: repeat(3, 1fr); }
    }
    @media (max-width: 700px) {
        .hero { padding: 24px 20px; }
        .hero-text h1 { font-size: 1.8rem; }
        .hero-stats { gap: 14px; padding: 12px 16px; }
        .tab-nav { grid-template-columns: 1fr; }
        .tests-grid { grid-template-columns: 1fr; }
        .quiz-grid  { grid-template-columns: 1fr; }
        .toolbar { flex-direction: column; align-items: stretch; }
        .search-wrap { max-width: 100%; }
        .sort-pills { margin-left: 0; }
        .modes-grid { grid-template-columns: repeat(2, 1fr); }
        .cat-header { flex-direction: column; }
        .cat-search-input { width: 140px; }
    }
</style>
