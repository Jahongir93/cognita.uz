<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { quizzes as quizzesApi, rooms as roomsApi } from '$lib/api/client';
    import type { Quiz } from '$lib/api/types';

    let quizzes: Quiz[] = [];
    let loading = true;
    let search = '';
    let viewMode: 'grid' | 'list' = 'grid';
    let sortBy: 'updated' | 'plays' | 'questions' | 'name' = 'updated';
    let filterPublic: 'all' | 'public' | 'private' = 'all';
    let showGameModal = false;
    let selectedQuiz: Quiz | null = null;
    let gameMode = 'classic';
    let creatingRoom = false;
    let shuffleQuestions = false;
    let showLeaderboard = true;

    onMount(async () => {
        try { quizzes = await quizzesApi.list(); }
        catch { quizzes = []; }
        finally { loading = false; }
    });

    $: filtered = quizzes.filter(q => {
        if (search && !q.title.toLowerCase().includes(search.toLowerCase())) return false;
        if (filterPublic === 'public' && !q.is_public) return false;
        if (filterPublic === 'private' && q.is_public) return false;
        return true;
    }).sort((a, b) => {
        if (sortBy === 'plays') return b.play_count - a.play_count;
        if (sortBy === 'questions') return b.total_questions - a.total_questions;
        if (sortBy === 'name') return a.title.localeCompare(b.title);
        return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
    });

    $: totalPlays = quizzes.reduce((s, q) => s + q.play_count, 0);
    $: totalQuestions = quizzes.reduce((s, q) => s + q.total_questions, 0);

    async function startGame() {
        if (!selectedQuiz) return;
        creatingRoom = true;
        try {
            const res = await roomsApi.create(selectedQuiz.id, gameMode, {
                shuffle_questions: shuffleQuestions,
                shuffle_answers: true,
                show_leaderboard: showLeaderboard,
                music: true,
                lobby_music: true,
                show_correct_answer: true
            });
            goto(`/game/host/${res.pin}`);
        } catch (e: any) {
            alert(e.message || 'Xato yuz berdi');
        } finally {
            creatingRoom = false;
            showGameModal = false;
        }
    }

    async function del(id: string, title: string) {
        if (!confirm(`"${title}" o'chirilsinmi?`)) return;
        try {
            await quizzesApi.delete(id);
            quizzes = quizzes.filter(q => q.id !== id);
        } catch (e: any) {
            alert(e.message || 'Xato yuz berdi');
        }
    }

    function openGameModal(quiz: Quiz) {
        selectedQuiz = quiz;
        gameMode = 'classic';
        shuffleQuestions = false;
        showLeaderboard = true;
        showGameModal = true;
    }

    function closeModal() {
        showGameModal = false;
        selectedQuiz = null;
    }

    function handleKey(e: KeyboardEvent) {
        if (e.key === 'Escape') closeModal();
    }

    function hue(id: string) {
        return (id.charCodeAt(0) * 5) % 360;
    }

    function coverGradient(id: string) {
        const h = hue(id);
        return `linear-gradient(135deg, hsl(${h},60%,52%), hsl(${(h+40)%360},60%,42%))`;
    }

    function formatDate(d: string) {
        return new Date(d).toLocaleDateString('uz-UZ', { day: '2-digit', month: 'short', year: 'numeric' });
    }

    const modes = [
        { value: 'classic',    icon: '⚡', label: 'Classic',  color: '#f59e0b', desc: "Savol ekranda, tezlik" },
        { value: 'self_paced', icon: '📲', label: 'Mustaqil', color: '#06b6d4', desc: "Har kim o'z qurilmasida" },
        { value: 'team',       icon: '👥', label: 'Jamoaviy', color: '#3b82f6', desc: "Jamoalarga bo'linib" },
        { value: 'accuracy',   icon: '🎯', label: 'Aniqlik',  color: '#22c55e', desc: "Faqat to'g'ri javob" },
        { value: 'confidence', icon: '💡', label: 'Ishonch',  color: '#8b5cf6', desc: 'Ishonch multiplikatori' },
        { value: 'zero_stakes',icon: '🌟', label: 'Mashq',    color: '#64748b', desc: "Ball yo'q, o'rganish" },
    ];
</script>

<svelte:head><title>Quizlar — Cognita.uz</title></svelte:head>
<svelte:window on:keydown={handleKey} />

<!-- Page Header -->
<div class="page-header">
    <div class="header-left">
        <h1>Quizlar</h1>
        <div class="header-stats">
            <span class="hstat"><strong>{quizzes.length}</strong> quiz</span>
            <span class="hstat-sep">·</span>
            <span class="hstat"><strong>{totalPlays}</strong> o'yin</span>
            <span class="hstat-sep">·</span>
            <span class="hstat"><strong>{totalQuestions}</strong> savol</span>
        </div>
    </div>
    <a href="/dashboard/quizzes/new" class="btn primary">
        <span>+</span> Yangi quiz
    </a>
</div>

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

    <div class="filter-pills">
        <button class="pill" class:active={filterPublic === 'all'} on:click={() => filterPublic = 'all'}>Barchasi</button>
        <button class="pill" class:active={filterPublic === 'public'} on:click={() => filterPublic = 'public'}>Ochiq</button>
        <button class="pill" class:active={filterPublic === 'private'} on:click={() => filterPublic = 'private'}>Yopiq</button>
    </div>

    <div class="toolbar-right">
        <select class="sort-select" bind:value={sortBy}>
            <option value="updated">Yangi</option>
            <option value="plays">Ko'p o'ynalgan</option>
            <option value="questions">Savollar</option>
            <option value="name">Nomi</option>
        </select>
        <div class="view-toggle">
            <button class="view-btn" class:active={viewMode === 'grid'} on:click={() => viewMode = 'grid'} title="Grid ko'rinish">
                ⊞
            </button>
            <button class="view-btn" class:active={viewMode === 'list'} on:click={() => viewMode = 'list'} title="Ro'yxat ko'rinish">
                ☰
            </button>
        </div>
    </div>
</div>

<!-- Content -->
{#if loading}
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
{:else if filtered.length === 0}
    <div class="empty-state">
        <div class="empty-emoji">📝</div>
        <p class="empty-title">
            {#if search || filterPublic !== 'all'}Hech narsa topilmadi{:else}Hali quiz yo'q{/if}
        </p>
        <p class="empty-sub">
            {#if search || filterPublic !== 'all'}
                Boshqa qidiruv mezonini sinab ko'ring
            {:else}
                Birinchi quizingizni yarating!
            {/if}
        </p>
        {#if !search && filterPublic === 'all'}
            <a href="/dashboard/quizzes/new" class="btn primary">+ Yangi quiz yaratish</a>
        {/if}
    </div>
{:else if viewMode === 'grid'}
    <!-- Grid Mode -->
    <div class="quiz-grid">
        {#each filtered as quiz (quiz.id)}
            <div class="quiz-card">
                <!-- Cover -->
                <div class="quiz-cover" style="background:{coverGradient(quiz.id)}">
                    <span class="cover-letter">{quiz.title[0]?.toUpperCase()}</span>
                    {#if quiz.is_public}
                        <span class="public-badge">Ochiq</span>
                    {/if}
                </div>
                <!-- Body -->
                <div class="quiz-body">
                    <div class="quiz-title-text">{quiz.title}</div>
                    <div class="quiz-tags">
                        {#if quiz.subject}<span class="tag subject-tag">{quiz.subject}</span>{/if}
                        {#if quiz.grade_level}<span class="tag grade-tag">{quiz.grade_level}-sinf</span>{/if}
                    </div>
                    <div class="quiz-meta">
                        <span>📌 {quiz.total_questions} savol</span>
                        <span>▶ {quiz.play_count} o'yin</span>
                        <span>📅 {formatDate(quiz.updated_at)}</span>
                    </div>
                    <!-- Actions -->
                    <div class="quiz-actions">
                        <button class="btn primary sm" on:click={() => openGameModal(quiz)}>
                            ▶ Boshlash
                        </button>
                        <a href="/dashboard/quizzes/{quiz.id}" class="btn secondary sm">
                            ✏️ Tahrir
                        </a>
                        <button class="btn danger-sm sm" on:click={() => del(quiz.id, quiz.title)}>
                            🗑️
                        </button>
                    </div>
                </div>
            </div>
        {/each}
    </div>
{:else}
    <!-- List Mode -->
    <div class="list-card">
        <div class="list-head">
            <span>Quiz nomi</span>
            <span>Fan</span>
            <span class="col-center">Savollar</span>
            <span class="col-center">O'yinlar</span>
            <span>Sana</span>
            <span>Amallar</span>
        </div>
        {#each filtered as quiz (quiz.id)}
            <div class="list-row">
                <div class="list-title-col">
                    <span class="color-dot" style="background:{coverGradient(quiz.id)}">
                        {quiz.title[0]?.toUpperCase()}
                    </span>
                    <div>
                        <div class="list-title">{quiz.title}</div>
                        {#if quiz.is_public}<span class="public-dot">Ochiq</span>{/if}
                    </div>
                </div>
                <span class="list-tag">
                    {#if quiz.subject}{quiz.subject}{:else}—{/if}
                </span>
                <span class="col-center list-num">{quiz.total_questions}</span>
                <span class="col-center list-num">{quiz.play_count}</span>
                <span class="list-date">{formatDate(quiz.updated_at)}</span>
                <div class="list-actions">
                    <button class="btn primary sm" on:click={() => openGameModal(quiz)}>▶</button>
                    <a href="/dashboard/quizzes/{quiz.id}" class="btn secondary sm">✏️</a>
                    <button class="btn danger-sm sm" on:click={() => del(quiz.id, quiz.title)}>🗑️</button>
                </div>
            </div>
        {/each}
    </div>
{/if}

<!-- Game Mode Modal -->
{#if showGameModal && selectedQuiz}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="modal-overlay" on:click|self={closeModal}>
        <div class="modal">
            <div class="modal-header">
                <div>
                    <p class="modal-pre">O'yinni boshlash</p>
                    <h2 class="modal-title">{selectedQuiz.title}</h2>
                </div>
                <button class="modal-close" on:click={closeModal}>✕</button>
            </div>

            <div class="modal-body">
                <p class="section-label">O'yin rejimini tanlang</p>
                <div class="modes-grid">
                    {#each modes as mode}
                        <button
                            class="mode-card"
                            class:selected={gameMode === mode.value}
                            on:click={() => gameMode = mode.value}
                            style="--mc: {mode.color}"
                        >
                            <span class="mode-icon">{mode.icon}</span>
                            <span class="mode-name">{mode.label}</span>
                            <span class="mode-desc">{mode.desc}</span>
                        </button>
                    {/each}
                </div>

                <p class="section-label" style="margin-top:20px">Sozlamalar</p>
                <div class="settings-row">
                    <label class="toggle-label">
                        <span>Savollarni aralashtirish</span>
                        <div class="toggle" class:on={shuffleQuestions} on:click={() => shuffleQuestions = !shuffleQuestions}>
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <div class="toggle-thumb"></div>
                        </div>
                    </label>
                    <label class="toggle-label">
                        <span>Liderlar jadvali</span>
                        <div class="toggle" class:on={showLeaderboard} on:click={() => showLeaderboard = !showLeaderboard}>
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <div class="toggle-thumb"></div>
                        </div>
                    </label>
                </div>
            </div>

            <div class="modal-footer">
                <button class="btn secondary" on:click={closeModal} disabled={creatingRoom}>
                    Bekor qilish
                </button>
                <button class="btn primary start-btn" on:click={startGame} disabled={creatingRoom}>
                    {#if creatingRoom}
                        <span class="spinner"></span> Yaratilmoqda...
                    {:else}
                        ▶ O'yinni boshlash
                    {/if}
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
    /* ── Page Header ─────────────────────────────────────────────────────────── */
    .page-header {
        display: flex; align-items: flex-start; justify-content: space-between;
        margin-bottom: 24px; gap: 16px;
    }
    h1 { font-size: 1.85rem; font-weight: 800; color: var(--text); margin: 0 0 6px; }
    .header-stats { display: flex; align-items: center; gap: 8px; }
    .hstat { font-size: 0.875rem; color: var(--text2); }
    .hstat strong { color: var(--text); }
    .hstat-sep { color: var(--text3); }

    /* ── Toolbar ─────────────────────────────────────────────────────────────── */
    .toolbar {
        display: flex; align-items: center; gap: 12px; margin-bottom: 20px; flex-wrap: wrap;
    }
    .search-wrap {
        position: relative; display: flex; align-items: center; flex: 1; min-width: 200px; max-width: 280px;
    }
    .search-icon {
        position: absolute; left: 10px; font-size: 0.85rem; pointer-events: none; z-index: 1;
    }
    .search-input {
        width: 100%; padding: 8px 32px 8px 32px; border: 1.5px solid var(--border);
        border-radius: var(--radius); font-size: 0.875rem; outline: none;
        transition: border-color 0.2s; background: var(--white);
    }
    .search-input:focus { border-color: var(--primary); }
    .clear-btn {
        position: absolute; right: 8px; background: none; border: none;
        cursor: pointer; color: var(--text3); font-size: 0.8rem; padding: 2px; line-height: 1;
    }

    .filter-pills { display: flex; gap: 6px; background: var(--bg2); padding: 4px; border-radius: 10px; }
    .pill {
        padding: 6px 14px; border: none; border-radius: 7px; font-size: 0.83rem;
        font-weight: 600; cursor: pointer; background: transparent; color: var(--text2);
        transition: all 0.2s;
    }
    .pill.active { background: var(--white); color: var(--primary); box-shadow: var(--shadow-sm); }
    .pill:hover:not(.active) { background: rgba(255,255,255,0.6); }

    .toolbar-right { display: flex; gap: 8px; align-items: center; margin-left: auto; }
    .sort-select {
        padding: 8px 12px; border: 1.5px solid var(--border); border-radius: var(--radius);
        font-size: 0.875rem; outline: none; background: var(--white); color: var(--text);
        cursor: pointer; transition: border-color 0.2s;
    }
    .sort-select:focus { border-color: var(--primary); }

    .view-toggle { display: flex; background: var(--bg2); border-radius: 8px; padding: 3px; }
    .view-btn {
        width: 32px; height: 32px; border: none; border-radius: 6px;
        background: transparent; cursor: pointer; font-size: 1.05rem; color: var(--text3);
        display: flex; align-items: center; justify-content: center;
        transition: all 0.15s;
    }
    .view-btn.active { background: var(--white); color: var(--primary); box-shadow: var(--shadow-sm); }

    /* ── Grid Mode ───────────────────────────────────────────────────────────── */
    .quiz-grid {
        display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
        gap: 18px; margin-bottom: 8px;
    }
    .quiz-card {
        background: var(--white); border-radius: var(--radius-lg);
        box-shadow: var(--shadow-sm); overflow: hidden;
        transition: var(--transition); border: 1.5px solid transparent;
        display: flex; flex-direction: column;
    }
    .quiz-card:hover { transform: translateY(-4px); box-shadow: var(--shadow); border-color: var(--primary-light); }

    .quiz-cover {
        height: 100px; position: relative; display: flex;
        align-items: center; justify-content: center; overflow: hidden;
    }
    .cover-letter {
        font-size: 2.8rem; font-weight: 900; color: rgba(255,255,255,0.85);
        text-shadow: 0 2px 8px rgba(0,0,0,0.2);
    }
    .public-badge {
        position: absolute; top: 10px; right: 10px;
        background: rgba(255,255,255,0.25); color: white; backdrop-filter: blur(4px);
        font-size: 0.68rem; font-weight: 700; padding: 3px 9px; border-radius: 99px;
        border: 1px solid rgba(255,255,255,0.4);
    }

    .quiz-body { padding: 14px; display: flex; flex-direction: column; gap: 10px; flex: 1; }
    .quiz-title-text {
        font-size: 0.95rem; font-weight: 700; color: var(--text);
        display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;
        overflow: hidden;
    }
    .quiz-tags { display: flex; gap: 6px; flex-wrap: wrap; }
    .tag { font-size: 0.7rem; font-weight: 600; padding: 2px 8px; border-radius: 99px; }
    .subject-tag { background: var(--primary-light); color: #5b21b6; }
    .grade-tag { background: #f0fdf4; color: #15803d; }
    .quiz-meta { display: flex; gap: 10px; font-size: 0.75rem; color: var(--text3); flex-wrap: wrap; }

    .quiz-actions { display: flex; gap: 7px; margin-top: auto; }

    /* ── List Mode ───────────────────────────────────────────────────────────── */
    .list-card {
        background: var(--white); border-radius: var(--radius-lg);
        box-shadow: var(--shadow-sm); overflow: hidden; margin-bottom: 8px;
    }
    .list-head {
        display: grid; grid-template-columns: 2fr 1fr 90px 90px 1.1fr 140px;
        gap: 12px; padding: 12px 20px;
        font-size: 0.7rem; font-weight: 700; color: var(--text3);
        text-transform: uppercase; letter-spacing: 0.07em;
        background: var(--bg2); border-bottom: 1.5px solid var(--border);
    }
    .list-row {
        display: grid; grid-template-columns: 2fr 1fr 90px 90px 1.1fr 140px;
        gap: 12px; align-items: center; padding: 12px 20px;
        border-bottom: 1px solid var(--bg2); transition: background 0.15s;
    }
    .list-row:last-child { border-bottom: none; }
    .list-row:hover { background: var(--bg2); }
    .col-center { text-align: center; justify-self: center; }

    .list-title-col { display: flex; align-items: center; gap: 10px; min-width: 0; }
    .color-dot {
        width: 36px; height: 36px; border-radius: 9px; flex-shrink: 0;
        display: flex; align-items: center; justify-content: center;
        font-size: 0.95rem; font-weight: 800; color: rgba(255,255,255,0.9);
    }
    .list-title {
        font-size: 0.875rem; font-weight: 700; color: var(--text);
        white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
    }
    .public-dot {
        font-size: 0.68rem; font-weight: 600; color: #16a34a; background: #dcfce7;
        padding: 1px 7px; border-radius: 99px;
    }
    .list-tag {
        font-size: 0.78rem; color: var(--text2);
        white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
    }
    .list-num { font-size: 0.875rem; font-weight: 600; color: var(--text2); }
    .list-date { font-size: 0.78rem; color: var(--text3); }
    .list-actions { display: flex; gap: 6px; }

    /* ── Skeleton ────────────────────────────────────────────────────────────── */
    .skeleton-card {
        background: var(--white); border-radius: var(--radius-lg);
        overflow: hidden; box-shadow: var(--shadow-sm);
    }
    .skel-cover {
        height: 100px;
        background: linear-gradient(90deg, var(--bg2) 25%, var(--border) 50%, var(--bg2) 75%);
        background-size: 200% 100%;
        animation: shimmer 1.4s infinite;
    }
    .skel-body { padding: 14px; display: flex; flex-direction: column; gap: 9px; }
    .skel-line {
        height: 12px; border-radius: 6px;
        background: linear-gradient(90deg, var(--bg2) 25%, var(--border) 50%, var(--bg2) 75%);
        background-size: 200% 100%;
        animation: shimmer 1.4s infinite;
    }
    .skel-line.wide { width: 75%; }
    .skel-line.medium { width: 50%; }
    .skel-line.short { width: 35%; }
    @keyframes shimmer { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }

    /* ── Empty ───────────────────────────────────────────────────────────────── */
    .empty-state {
        background: var(--white); border-radius: var(--radius-lg);
        padding: 64px 32px; text-align: center; box-shadow: var(--shadow-sm);
        display: flex; flex-direction: column; align-items: center; gap: 8px;
    }
    .empty-emoji { font-size: 3.5rem; margin-bottom: 8px; }
    .empty-title { font-size: 1.1rem; font-weight: 700; color: var(--text); margin: 0; }
    .empty-sub { font-size: 0.875rem; color: var(--text3); margin: 0 0 16px; }

    /* ── Modal ───────────────────────────────────────────────────────────────── */
    .modal-overlay {
        position: fixed; inset: 0; background: rgba(0,0,0,0.5);
        display: flex; align-items: center; justify-content: center;
        z-index: 100; padding: 16px;
        animation: fadeIn 0.15s ease;
    }
    @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
    .modal {
        background: var(--white); border-radius: var(--radius-lg);
        width: 100%; max-width: 540px; box-shadow: 0 20px 60px rgba(0,0,0,0.25);
        animation: slideUp 0.2s ease; overflow: hidden;
        max-height: 90vh; overflow-y: auto;
    }
    @keyframes slideUp { from { transform: translateY(20px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
    .modal-header {
        display: flex; align-items: flex-start; justify-content: space-between;
        padding: 22px 24px 0; gap: 12px;
    }
    .modal-pre { font-size: 0.75rem; font-weight: 600; color: var(--text3); text-transform: uppercase; letter-spacing: 0.06em; margin: 0 0 4px; }
    .modal-title { font-size: 1.15rem; font-weight: 800; color: var(--text); margin: 0; }
    .modal-close {
        background: var(--bg2); border: none; border-radius: 8px;
        width: 32px; height: 32px; cursor: pointer; font-size: 0.9rem;
        color: var(--text3); display: flex; align-items: center; justify-content: center;
        transition: all 0.15s; flex-shrink: 0;
    }
    .modal-close:hover { background: var(--border); color: var(--text); }
    .modal-body { padding: 20px 24px; }
    .modal-footer {
        padding: 16px 24px; border-top: 1px solid var(--border);
        display: flex; justify-content: flex-end; gap: 10px;
        background: var(--bg2);
    }

    .section-label { font-size: 0.78rem; font-weight: 700; color: var(--text3); text-transform: uppercase; letter-spacing: 0.06em; margin: 0 0 10px; }

    /* Mode cards */
    .modes-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 8px; }
    .mode-card {
        display: flex; flex-direction: column; align-items: center; gap: 4px;
        padding: 12px 6px; border: 2px solid var(--border); border-radius: var(--radius);
        background: var(--white); cursor: pointer; transition: all 0.2s;
        text-align: center;
    }
    .mode-card:hover { border-color: var(--mc, var(--primary)); background: var(--bg2); }
    .mode-card.selected {
        border-color: var(--mc, var(--primary));
        background: color-mix(in srgb, var(--mc, var(--primary)) 8%, white);
        box-shadow: 0 0 0 3px color-mix(in srgb, var(--mc, var(--primary)) 15%, transparent);
    }
    .mode-icon { font-size: 1.4rem; }
    .mode-name { font-size: 0.72rem; font-weight: 700; color: var(--text); }
    .mode-desc { font-size: 0.65rem; color: var(--text3); line-height: 1.3; }

    /* Settings toggles */
    .settings-row { display: flex; flex-direction: column; gap: 12px; }
    .toggle-label {
        display: flex; align-items: center; justify-content: space-between;
        font-size: 0.875rem; color: var(--text); cursor: pointer; padding: 4px 0;
    }
    .toggle {
        width: 42px; height: 24px; border-radius: 12px; background: var(--border);
        position: relative; cursor: pointer; transition: background 0.2s;
        flex-shrink: 0;
    }
    .toggle.on { background: var(--primary); }
    .toggle-thumb {
        position: absolute; top: 3px; left: 3px;
        width: 18px; height: 18px; border-radius: 50%;
        background: white; box-shadow: 0 1px 4px rgba(0,0,0,0.2);
        transition: transform 0.2s;
    }
    .toggle.on .toggle-thumb { transform: translateX(18px); }

    /* Start button */
    .start-btn { min-width: 160px; justify-content: center; }

    /* ── Buttons ─────────────────────────────────────────────────────────────── */
    .btn {
        padding: 9px 18px; border: none; border-radius: var(--radius);
        font-size: 0.875rem; font-weight: 600; cursor: pointer;
        display: inline-flex; align-items: center; gap: 7px;
        transition: var(--transition); text-decoration: none;
    }
    .btn.sm { padding: 6px 12px; font-size: 0.8rem; }
    .btn.primary {
        background: linear-gradient(135deg, var(--primary), var(--accent));
        color: white; box-shadow: 0 3px 10px rgba(99,102,241,0.3);
    }
    .btn.primary:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 16px rgba(99,102,241,0.4); }
    .btn.secondary { background: var(--bg2); color: var(--text); border: 1px solid var(--border); }
    .btn.secondary:hover:not(:disabled) { background: var(--border); }
    .btn.danger-sm { background: #fee2e2; color: var(--danger); }
    .btn.danger-sm:hover:not(:disabled) { background: #fecaca; }
    .btn:disabled { opacity: 0.5; cursor: not-allowed; }

    .spinner {
        display: inline-block; width: 14px; height: 14px;
        border: 2px solid rgba(255,255,255,0.4); border-top-color: white;
        border-radius: 50%; animation: spin 0.6s linear infinite;
    }
    @keyframes spin { to { transform: rotate(360deg); } }

    /* ── Responsive ──────────────────────────────────────────────────────────── */
    @media (max-width: 1024px) {
        .list-head, .list-row { grid-template-columns: 2fr 80px 80px 1fr 130px; }
        .list-head span:nth-child(2), .list-tag { display: none; }
        .modes-grid { grid-template-columns: repeat(3, 1fr); }
    }
    @media (max-width: 640px) {
        .quiz-grid { grid-template-columns: 1fr; }
        .toolbar { flex-direction: column; align-items: stretch; }
        .search-wrap { max-width: 100%; }
        .toolbar-right { margin-left: 0; }
        .modes-grid { grid-template-columns: repeat(2, 1fr); }
        .list-head, .list-row { grid-template-columns: 1fr 80px 120px; }
        .list-head span:nth-child(3),
        .list-head span:nth-child(4),
        .list-num { display: none; }
    }
</style>
