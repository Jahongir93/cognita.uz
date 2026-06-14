<script lang="ts">
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { authStore } from '$lib/stores/auth';
    import { allCategories } from '$lib/data/categories';
    import { quizzes as quizzesApi, rooms as roomsApi } from '$lib/api/client';
    import type { Quiz } from '$lib/api/types';

    // ── Types ─────────────────────────────────────────────────────────────────
    interface Test {
        id: string;
        icon: string;
        title: string;
        description: string;
        questions: number;
        difficulty: 'easy' | 'medium' | 'hard';
        duration: number;
        plays: number;
        rating: number;
        subcat: string;
        isNew?: boolean;
        isHot?: boolean;
        quizId?: string;
        quizTitle?: string;
        tags?: string[];
    }
    interface SubCat { id: string; label: string; icon: string; }
    interface Category {
        id: string;
        slug: string;
        icon: string;
        title: string;
        subtitle: string;
        g1: string;
        g2: string;
        subcats: SubCat[];
        tests: Test[];
    }

    // ── Init from localStorage ─────────────────────────────────────────────────
    const STORAGE_KEY = 'gogame_resource_center';

    function loadCategories(): Category[] {
        try {
            const raw = localStorage.getItem(STORAGE_KEY);
            if (raw) return JSON.parse(raw) as Category[];
        } catch {}
        return allCategories.map(c => ({ ...c, tests: c.tests.map(t => ({ ...t, tags: [] })) })) as Category[];
    }

    function saveCategories() {
        try { localStorage.setItem(STORAGE_KEY, JSON.stringify(categories)); } catch {}
    }

    // ── State ─────────────────────────────────────────────────────────────────
    let categories: Category[] = [];
    let selectedCatId = '';
    let selectedSubcat = 'all';
    let searchQuery = '';
    let sortBy: 'default' | 'plays' | 'rating' | 'name' | 'difficulty' = 'default';
    let selectedTests = new Set<string>();
    let detailTest: Test | null = null;
    let quizzesList: Quiz[] = [];

    // Panel/modal state
    let showCatModal = false;
    let editingCat: Partial<Category> | null = null;
    let showTestModal = false;
    let editingTest: Partial<Test> | null = null;
    let isNewTest = false;
    let showImportModal = false;
    let importSelected = new Set<string>();
    let showSubcatModal = false;
    let editingSubcat: Partial<SubCat> | null = null;
    let isNewSubcat = false;
    let showDeleteCatConfirm = false;
    let globalSearch = '';
    let showGlobalResults = false;
    let toast = '';
    let toastType: 'success' | 'error' = 'success';
    let toastTimer: ReturnType<typeof setTimeout> | null = null;
    let launchingGame = false;

    // ── Derived ───────────────────────────────────────────────────────────────
    $: selectedCat = categories.find(c => c.id === selectedCatId) ?? categories[0];
    $: filteredTests = (selectedCat?.tests ?? []).filter((t: Test) => {
        const ms = !searchQuery || t.title.toLowerCase().includes(searchQuery.toLowerCase())
            || t.description.toLowerCase().includes(searchQuery.toLowerCase());
        const mc = selectedSubcat === 'all' || t.subcat === selectedSubcat;
        return ms && mc;
    });
    $: visibleTests = [...filteredTests].sort((a: Test, b: Test) => {
        if (sortBy === 'plays') return b.plays - a.plays;
        if (sortBy === 'rating') return b.rating - a.rating;
        if (sortBy === 'name') return a.title.localeCompare(b.title);
        if (sortBy === 'difficulty') {
            const o = { easy: 0, medium: 1, hard: 2 } as Record<string, number>;
            return (o[a.difficulty] ?? 1) - (o[b.difficulty] ?? 1);
        }
        return 0;
    });
    $: globalResults = globalSearch.length > 1
        ? categories.flatMap(c => c.tests.filter(t =>
            t.title.toLowerCase().includes(globalSearch.toLowerCase()) ||
            t.description.toLowerCase().includes(globalSearch.toLowerCase())
          ).map(t => ({ ...t, catTitle: c.title, catId: c.id, catIcon: c.icon })))
        : [];
    $: totalTests = categories.reduce((s, c) => s + c.tests.length, 0);
    $: totalPlays = categories.reduce((s, c) => s + c.tests.reduce((ss, t) => ss + t.plays, 0), 0);
    $: linkedTests = categories.reduce((s, c) => s + c.tests.filter(t => t.quizId).length, 0);

    // ── Lifecycle ─────────────────────────────────────────────────────────────
    onMount(async () => {
        if ($authStore.user && $authStore.user.role !== 'admin') {
            goto('/dashboard');
            return;
        }
        categories = loadCategories();
        if (categories.length > 0) selectedCatId = categories[0].id;
        try { quizzesList = await quizzesApi.list(); } catch { quizzesList = []; }
    });

    // ── Helpers ───────────────────────────────────────────────────────────────
    function showToast(msg: string, type: 'success' | 'error' = 'success') {
        toast = msg; toastType = type;
        if (toastTimer) clearTimeout(toastTimer);
        toastTimer = setTimeout(() => { toast = ''; }, 2800);
    }

    function formatPlays(n: number) {
        if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
        if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
        return n.toString();
    }

    function diffLabel(d: string) {
        return d === 'easy' ? 'Oson' : d === 'medium' ? "O'rta" : 'Qiyin';
    }

    function uid() { return Date.now().toString(36) + Math.random().toString(36).slice(2, 6); }

    const PRESET_ICONS = ['🎯','📚','🔬','🧠','🎮','🏆','💡','🌍','🎨','🔢','⚗️','📖','🎵','🌿','💻','🏛️','🎬','🚀','🔭','🌊'];
    const PRESET_COLORS = [
        ['#f59e0b','#ef4444'], ['#3b82f6','#6366f1'], ['#10b981','#059669'],
        ['#8b5cf6','#ec4899'], ['#f97316','#f59e0b'], ['#06b6d4','#0ea5e9'],
        ['#84cc16','#22c55e'], ['#e11d48','#f43f5e'],
    ];

    // ── Category CRUD ──────────────────────────────────────────────────────────
    function openNewCat() {
        editingCat = {
            id: uid(), slug: '', icon: '📂', title: '',
            subtitle: '', g1: '#6366f1', g2: '#8b5cf6',
            subcats: [{ id: 'all', label: 'Barchasi', icon: '🌟' }],
            tests: [],
        };
        showCatModal = true;
    }

    function openEditCat() {
        if (!selectedCat) return;
        editingCat = JSON.parse(JSON.stringify(selectedCat));
        showCatModal = true;
    }

    function saveCat() {
        if (!editingCat?.title?.trim()) return;
        const idx = categories.findIndex(c => c.id === editingCat!.id);
        const cat = {
            ...editingCat,
            slug: editingCat.slug || editingCat.title!.toLowerCase().replace(/\s+/g, '-'),
            subcats: editingCat.subcats && editingCat.subcats.length > 0
                ? editingCat.subcats
                : [{ id: 'all', label: 'Barchasi', icon: '🌟' }],
        } as Category;
        if (idx >= 0) {
            categories[idx] = cat;
        } else {
            categories = [...categories, cat];
            selectedCatId = cat.id;
        }
        categories = [...categories];
        saveCategories();
        showCatModal = false;
        showToast(idx >= 0 ? 'Kategoriya yangilandi' : 'Kategoriya qo\'shildi');
    }

    function deleteCat() {
        if (!selectedCat) return;
        categories = categories.filter(c => c.id !== selectedCatId);
        selectedCatId = categories[0]?.id ?? '';
        saveCategories();
        showDeleteCatConfirm = false;
        showToast('Kategoriya o\'chirildi');
    }

    // ── Subcategory CRUD ───────────────────────────────────────────────────────
    function openNewSubcat() {
        editingSubcat = { id: uid(), label: '', icon: '📌' };
        isNewSubcat = true;
        showSubcatModal = true;
    }

    function openEditSubcat(sc: SubCat) {
        editingSubcat = { ...sc };
        isNewSubcat = false;
        showSubcatModal = true;
    }

    function saveSubcat() {
        if (!editingSubcat?.label?.trim() || !selectedCat) return;
        const cat = categories.find(c => c.id === selectedCatId)!;
        if (isNewSubcat) {
            cat.subcats = [...cat.subcats, editingSubcat as SubCat];
        } else {
            cat.subcats = cat.subcats.map(s => s.id === editingSubcat!.id ? { ...editingSubcat } as SubCat : s);
        }
        categories = [...categories];
        saveCategories();
        showSubcatModal = false;
        showToast(isNewSubcat ? 'Bo\'lim qo\'shildi' : 'Bo\'lim yangilandi');
    }

    function deleteSubcat(scId: string) {
        if (!selectedCat || scId === 'all') return;
        const cat = categories.find(c => c.id === selectedCatId)!;
        cat.subcats = cat.subcats.filter(s => s.id !== scId);
        cat.tests = cat.tests.map(t => t.subcat === scId ? { ...t, subcat: 'all' } : t);
        categories = [...categories];
        saveCategories();
        if (selectedSubcat === scId) selectedSubcat = 'all';
        showToast('Bo\'lim o\'chirildi');
    }

    // ── Test CRUD ──────────────────────────────────────────────────────────────
    function openNewTest() {
        editingTest = {
            id: uid(), icon: '📝', title: '', description: '',
            questions: 10, difficulty: 'medium', duration: 10,
            plays: 0, rating: 4.0, isNew: true, isHot: false,
            quizId: '', tags: [],
            subcat: selectedSubcat === 'all'
                ? (selectedCat?.subcats[1]?.id ?? 'all')
                : selectedSubcat,
        };
        isNewTest = true;
        showTestModal = true;
    }

    function openEditTest(test: Test) {
        editingTest = { ...test, tags: [...(test.tags ?? [])] };
        isNewTest = false;
        showTestModal = true;
        detailTest = null;
    }

    function saveTest() {
        if (!editingTest?.title?.trim() || !selectedCat) return;
        const cat = categories.find(c => c.id === selectedCatId)!;
        if (isNewTest) {
            cat.tests = [...cat.tests, editingTest as Test];
        } else {
            cat.tests = cat.tests.map(t => t.id === editingTest!.id ? { ...editingTest } as Test : t);
        }
        categories = [...categories];
        saveCategories();
        showTestModal = false;
        showToast(isNewTest ? 'Test qo\'shildi' : 'Test yangilandi');
        if (!isNewTest) detailTest = editingTest as Test;
    }

    function deleteTest(testId: string) {
        const cat = categories.find(c => c.id === selectedCatId)!;
        const t = cat.tests.find(t => t.id === testId);
        cat.tests = cat.tests.filter(t => t.id !== testId);
        categories = [...categories];
        saveCategories();
        if (detailTest?.id === testId) detailTest = null;
        showToast(`"${t?.title}" o'chirildi`);
    }

    function bulkDelete() {
        if (!selectedTests.size) return;
        if (!confirm(`${selectedTests.size} ta test o'chirilsinmi?`)) return;
        const cat = categories.find(c => c.id === selectedCatId)!;
        cat.tests = cat.tests.filter(t => !selectedTests.has(t.id));
        categories = [...categories];
        saveCategories();
        showToast(`${selectedTests.size} ta test o'chirildi`);
        selectedTests = new Set();
    }

    function toggleSelect(id: string) {
        if (selectedTests.has(id)) selectedTests.delete(id);
        else selectedTests.add(id);
        selectedTests = new Set(selectedTests);
    }

    function selectAll() {
        selectedTests = selectedTests.size === visibleTests.length && visibleTests.length > 0
            ? new Set() : new Set(visibleTests.map(t => t.id));
    }

    // ── Import from quizzes ────────────────────────────────────────────────────
    function openImport() { importSelected = new Set(); showImportModal = true; }

    function confirmImport() {
        if (!importSelected.size || !selectedCat) return;
        const cat = categories.find(c => c.id === selectedCatId)!;
        const existing = new Set(cat.tests.map(t => t.quizId));
        let added = 0;
        quizzesList.forEach(q => {
            if (importSelected.has(q.id) && !existing.has(q.id)) {
                cat.tests = [...cat.tests, {
                    id: uid(),
                    icon: '📝',
                    title: q.title,
                    description: q.description ?? '',
                    questions: q.total_questions,
                    difficulty: 'medium',
                    duration: Math.ceil(q.total_questions * 1.5),
                    plays: q.play_count,
                    rating: 4.0,
                    subcat: selectedSubcat === 'all' ? (selectedCat.subcats[1]?.id ?? 'all') : selectedSubcat,
                    quizId: q.id,
                    quizTitle: q.title,
                    tags: [q.subject ?? '', q.grade_level ?? ''].filter(Boolean),
                } as Test];
                added++;
            }
        });
        categories = [...categories];
        saveCategories();
        showImportModal = false;
        showToast(`${added} ta quiz import qilindi`);
    }

    // ── Quiz launch ────────────────────────────────────────────────────────────
    async function launchGame(test: Test) {
        if (!test.quizId) return;
        launchingGame = true;
        try {
            const res = await roomsApi.create(test.quizId, 'classic', {
                shuffle_questions: false, shuffle_answers: true,
                show_leaderboard: true, music: true, lobby_music: true, show_correct_answer: true,
            });
            goto(`/game/host/${res.pin}`);
        } catch (e: any) {
            showToast(e.message || 'Xato', 'error');
        } finally { launchingGame = false; }
    }

    // ── Category link + subcats shortcuts ─────────────────────────────────────
    function selectCat(id: string) {
        selectedCatId = id;
        selectedSubcat = 'all';
        searchQuery = '';
        selectedTests = new Set();
        detailTest = null;
    }

    function jumpGlobal(catId: string) {
        selectedCatId = catId;
        selectedSubcat = 'all';
        searchQuery = globalSearch;
        globalSearch = '';
        showGlobalResults = false;
    }

    // ── Export ─────────────────────────────────────────────────────────────────
    function exportJSON() {
        const blob = new Blob([JSON.stringify(categories, null, 2)], { type: 'application/json' });
        const a = document.createElement('a');
        a.href = URL.createObjectURL(blob);
        a.download = 'resurs-markazi.json'; a.click();
        URL.revokeObjectURL(a.href);
        showToast('Export qilindi');
    }

    function importJSON() {
        const input = document.createElement('input');
        input.type = 'file'; input.accept = '.json';
        input.onchange = (e) => {
            const file = (e.target as HTMLInputElement).files?.[0];
            if (!file) return;
            const reader = new FileReader();
            reader.onload = (ev) => {
                try {
                    const data = JSON.parse(ev.target?.result as string) as Category[];
                    categories = data;
                    selectedCatId = data[0]?.id ?? '';
                    saveCategories();
                    showToast('Import muvaffaqiyatli');
                } catch { showToast('Fayl noto\'g\'ri', 'error'); }
            };
            reader.readAsText(file);
        };
        input.click();
    }
</script>

<svelte:head><title>Resurs Markazi — Cognita.uz Admin</title></svelte:head>

<!-- ── Page Header ──────────────────────────────────────────────────────────── -->
<div class="page-header">
    <div class="header-left">
        <div class="page-icon">🗂️</div>
        <div>
            <h1 class="page-title">Resurs Markazi</h1>
            <p class="page-sub">Kategoriyalar, testlar va o'quv resurslarini boshqaring</p>
        </div>
    </div>

    <div class="header-center">
        <div class="global-search-wrap">
            <span class="gs-icon">🔍</span>
            <input
                class="global-search-input"
                placeholder="Barcha resurslarda qidirish..."
                bind:value={globalSearch}
                on:focus={() => showGlobalResults = true}
                on:blur={() => setTimeout(() => showGlobalResults = false, 200)}
            />
            {#if showGlobalResults && globalSearch.length > 1}
                <div class="global-results">
                    {#if globalResults.length > 0}
                        {#each globalResults.slice(0, 8) as r}
                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <div class="gr-item" role="button" tabindex="0" on:click={() => jumpGlobal(r.catId)}>
                                <span class="gr-icon">{r.icon}</span>
                                <div class="gr-info">
                                    <span class="gr-title">{r.title}</span>
                                    <span class="gr-meta">{r.catIcon} {r.catTitle} · {r.questions} savol</span>
                                </div>
                                {#if r.quizId}<span class="gr-linked">🔗</span>{/if}
                            </div>
                        {/each}
                        {#if globalResults.length > 8}
                            <div class="gr-more">+{globalResults.length - 8} ta natija</div>
                        {/if}
                    {:else}
                        <div class="gr-empty">Natija topilmadi</div>
                    {/if}
                </div>
            {/if}
        </div>
    </div>

    <div class="header-right">
        <button class="hdr-btn" on:click={importJSON} title="JSON import">📥</button>
        <button class="hdr-btn" on:click={exportJSON} title="JSON export">📤 Export</button>
        <button class="hdr-btn primary" on:click={openNewCat}>＋ Kategoriya</button>
    </div>
</div>

<!-- ── Stats strip ───────────────────────────────────────────────────────────── -->
<div class="stats-strip">
    <div class="stat-card">
        <span class="stat-icon">📂</span>
        <div class="stat-info">
            <span class="stat-val">{categories.length}</span>
            <span class="stat-lbl">Kategoriya</span>
        </div>
    </div>
    <div class="stat-card">
        <span class="stat-icon">📋</span>
        <div class="stat-info">
            <span class="stat-val">{totalTests}</span>
            <span class="stat-lbl">Jami resurslar</span>
        </div>
    </div>
    <div class="stat-card">
        <span class="stat-icon">🔗</span>
        <div class="stat-info">
            <span class="stat-val">{linkedTests}</span>
            <span class="stat-lbl">Quiz bilan bog'liq</span>
        </div>
    </div>
    <div class="stat-card">
        <span class="stat-icon">▶</span>
        <div class="stat-info">
            <span class="stat-val">{formatPlays(totalPlays)}</span>
            <span class="stat-lbl">Jami o'yinlar</span>
        </div>
    </div>
</div>

<!-- ── Main workspace ────────────────────────────────────────────────────────── -->
<div class="workspace" class:has-detail={!!detailTest}>

    <!-- ════ SIDEBAR ════ -->
    <aside class="sidebar">
        <div class="sidebar-top">
            <span class="sb-label">KATEGORIYALAR</span>
            <button class="sb-add-btn" on:click={openNewCat} title="Yangi kategoriya">＋</button>
        </div>

        <div class="cat-list">
            {#each categories as cat (cat.id)}
                <button
                    class="cat-item"
                    class:active={cat.id === selectedCatId}
                    on:click={() => selectCat(cat.id)}
                    style="--g1:{cat.g1};--g2:{cat.g2}"
                >
                    <span class="ci-icon">{cat.icon}</span>
                    <div class="ci-info">
                        <span class="ci-name">{cat.title}</span>
                        <span class="ci-count">{cat.tests.length} resurs</span>
                    </div>
                    {#if cat.id === selectedCatId}
                        <span class="ci-dot"></span>
                    {/if}
                </button>
            {/each}
        </div>
    </aside>

    <!-- ════ CENTER PANEL ════ -->
    <section class="center-panel" style={selectedCat ? `--g1:${selectedCat.g1};--g2:${selectedCat.g2}` : ''}>
        {#if !selectedCat}
            <div class="empty-state big">
                <div class="empty-icon">📂</div>
                <p class="empty-title">Kategoriya yo'q</p>
                <p class="empty-sub">Birinchi kategoriyangizni yarating</p>
                <button class="btn primary" on:click={openNewCat}>＋ Kategoriya yaratish</button>
            </div>
        {:else}

            <!-- Category hero -->
            <div class="cat-hero">
                <div class="cat-hero-bg" aria-hidden="true"></div>
                <div class="cat-hero-content">
                    <div class="cat-hero-left">
                        <div class="cat-hero-icon">{selectedCat.icon}</div>
                        <div>
                            <h2 class="cat-hero-title">{selectedCat.title}</h2>
                            <p class="cat-hero-sub">{selectedCat.subtitle}</p>
                        </div>
                    </div>
                    <div class="cat-hero-meta">
                        <div class="cm">
                            <strong>{selectedCat.tests.length}</strong>
                            <span>Resurs</span>
                        </div>
                        <div class="cm">
                            <strong>{selectedCat.subcats.filter(s => s.id !== 'all').length}</strong>
                            <span>Bo'lim</span>
                        </div>
                        <div class="cm">
                            <strong>{selectedCat.tests.filter(t => t.quizId).length}</strong>
                            <span>Bog'liq</span>
                        </div>
                    </div>
                    <div class="cat-hero-actions">
                        <button class="hero-btn" on:click={openEditCat} title="Tahrirlash">✏️ Tahrirlash</button>
                        <button class="hero-btn danger" on:click={() => showDeleteCatConfirm = true} title="O'chirish">🗑️</button>
                    </div>
                </div>
            </div>

            <!-- Subcategory tabs -->
            <div class="subcat-bar">
                <div class="subcat-tabs">
                    {#each selectedCat.subcats as sc (sc.id)}
                        <button
                            class="sc-tab"
                            class:active={selectedSubcat === sc.id}
                            on:click={() => selectedSubcat = sc.id}
                        >
                            <span>{sc.icon}</span>
                            <span>{sc.label}</span>
                            <span class="sc-count">
                                {sc.id === 'all'
                                    ? selectedCat.tests.length
                                    : selectedCat.tests.filter(t => t.subcat === sc.id).length}
                            </span>
                            {#if sc.id !== 'all'}
                                <!-- svelte-ignore a11y-click-events-have-key-events -->
                                <span
                                    class="sc-del"
                                    role="button"
                                    tabindex="-1"
                                    on:click|stopPropagation={() => deleteSubcat(sc.id)}
                                    title="Bo'limni o'chirish"
                                >×</span>
                            {/if}
                        </button>
                    {/each}
                    <button class="sc-add-btn" on:click={openNewSubcat} title="Bo'lim qo'shish">＋ Bo'lim</button>
                </div>
            </div>

            <!-- Toolbar -->
            <div class="toolbar">
                <div class="search-wrap">
                    <span class="si">🔍</span>
                    <input class="search-input" placeholder="Qidirish..." bind:value={searchQuery} />
                    {#if searchQuery}<button class="sc" on:click={() => searchQuery = ''}>✕</button>{/if}
                </div>

                <div class="toolbar-right">
                    <span class="result-count">{visibleTests.length} ta</span>
                    <select class="sort-select" bind:value={sortBy}>
                        <option value="default">Tartib</option>
                        <option value="plays">Ko'p o'ynalgan</option>
                        <option value="rating">Reyting</option>
                        <option value="name">Nomi</option>
                        <option value="difficulty">Qiyinlik</option>
                    </select>
                    <button class="btn ghost" on:click={openImport} title="Quizdan import">
                        📥 Import
                    </button>
                    <button class="btn primary sm" on:click={openNewTest}>＋ Resurs qo'shish</button>
                </div>
            </div>

            <!-- Bulk bar -->
            {#if visibleTests.length > 0}
                <div class="bulk-bar">
                    <label class="bulk-check">
                        <input type="checkbox"
                            checked={selectedTests.size === visibleTests.length && visibleTests.length > 0}
                            on:change={selectAll} />
                        <span>Barchasi</span>
                    </label>
                    {#if selectedTests.size > 0}
                        <span class="bulk-count">{selectedTests.size} tanlandi</span>
                        <button class="bulk-del" on:click={bulkDelete}>🗑️ O'chirish</button>
                    {/if}
                </div>
            {/if}

            <!-- Tests grid -->
            {#if visibleTests.length === 0}
                <div class="empty-state">
                    <div class="empty-icon">
                        {#if searchQuery}🔍{:else}📋{/if}
                    </div>
                    <p class="empty-title">
                        {#if searchQuery}"{searchQuery}" topilmadi{:else}Hali resurs yo'q{/if}
                    </p>
                    {#if !searchQuery}
                        <div class="empty-actions">
                            <button class="btn primary sm" on:click={openNewTest}>＋ Resurs yaratish</button>
                            <button class="btn ghost sm" on:click={openImport}>📥 Quizdan import</button>
                        </div>
                    {:else}
                        <button class="btn ghost sm" on:click={() => searchQuery = ''}>Tozalash</button>
                    {/if}
                </div>
            {:else}
                <div class="tests-grid">
                    {#each visibleTests as test (test.id)}
                        <div
                            class="test-card"
                            class:selected={selectedTests.has(test.id)}
                            class:detail-open={detailTest?.id === test.id}
                        >
                            <div class="tc-header">
                                <label class="tc-check" on:click|stopPropagation>
                                    <input type="checkbox"
                                        checked={selectedTests.has(test.id)}
                                        on:change={() => toggleSelect(test.id)} />
                                </label>
                                <span class="tc-icon">{test.icon}</span>
                                <div class="tc-badges">
                                    {#if test.isNew}<span class="badge new">Yangi</span>{/if}
                                    {#if test.isHot}<span class="badge hot">🔥</span>{/if}
                                    {#if test.quizId}<span class="badge linked" title="Quiz bilan bog'liq">🔗</span>{/if}
                                </div>
                            </div>

                            <!-- svelte-ignore a11y-click-events-have-key-events -->
                            <div class="tc-body" role="button" tabindex="0" on:click={() => detailTest = detailTest?.id === test.id ? null : test}>
                                <h3 class="tc-title">{test.title}</h3>
                                <p class="tc-desc">{test.description}</p>

                                <div class="tc-meta">
                                    <span class="chip">📌 {test.questions}</span>
                                    <span class="chip diff-{test.difficulty}">{diffLabel(test.difficulty)}</span>
                                    <span class="chip">⏱ {test.duration}m</span>
                                </div>
                            </div>

                            <div class="tc-footer">
                                <span class="tc-plays">▶ {formatPlays(test.plays)}</span>
                                <div class="tc-actions">
                                    {#if test.quizId}
                                        <button class="act play" on:click|stopPropagation={() => launchGame(test)} disabled={launchingGame}>
                                            ▶ O'ynash
                                        </button>
                                    {/if}
                                    <button class="act edit" on:click|stopPropagation={() => openEditTest(test)}>✏️</button>
                                    <button class="act del" on:click|stopPropagation={() => deleteTest(test.id)}>🗑️</button>
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}

        {/if}
    </section>

    <!-- ════ DETAIL PANEL ════ -->
    {#if detailTest}
        {@const dt = detailTest}
        <aside class="detail-panel">
            <div class="dp-header">
                <h3>Resurs tafsiloti</h3>
                <button class="dp-close" on:click={() => detailTest = null}>✕</button>
            </div>

            <div class="dp-cover">
                <span class="dp-big-icon">{dt.icon}</span>
                <div class="dp-badges">
                    {#if dt.isNew}<span class="badge new">Yangi</span>{/if}
                    {#if dt.isHot}<span class="badge hot">🔥 Trend</span>{/if}
                    <span class="badge diff-{dt.difficulty}">{diffLabel(dt.difficulty)}</span>
                </div>
            </div>

            <div class="dp-body">
                <h2 class="dp-title">{dt.title}</h2>
                <p class="dp-desc">{dt.description}</p>

                <div class="dp-stats">
                    <div class="dps"><span class="dps-val">{dt.questions}</span><span class="dps-lbl">Savol</span></div>
                    <div class="dps"><span class="dps-val">{dt.duration}m</span><span class="dps-lbl">Vaqt</span></div>
                    <div class="dps"><span class="dps-val">{formatPlays(dt.plays)}</span><span class="dps-lbl">O'yin</span></div>
                    <div class="dps"><span class="dps-val">⭐{dt.rating.toFixed(1)}</span><span class="dps-lbl">Reyting</span></div>
                </div>

                {#if dt.tags?.length}
                    <div class="dp-tags">
                        {#each dt.tags as tag}
                            <span class="dp-tag">{tag}</span>
                        {/each}
                    </div>
                {/if}

                {#if dt.quizId}
                    <div class="dp-quiz-link">
                        <div class="dql-icon">🔗</div>
                        <div class="dql-info">
                            <span class="dql-label">Bog'liq quiz</span>
                            <span class="dql-title">{dt.quizTitle || dt.quizId}</span>
                        </div>
                    </div>
                {/if}
            </div>

            <div class="dp-footer">
                {#if dt.quizId}
                    <button class="btn primary" on:click={() => launchGame(dt)} disabled={launchingGame}>
                        {#if launchingGame}<span class="spin"></span>{:else}▶{/if}
                        Viktorina boshlash
                    </button>
                    <a href="/dashboard/quizzes/{dt.quizId}" class="btn ghost">
                        Quiz tahrirlash →
                    </a>
                {:else}
                    <div class="dp-no-quiz">
                        <p>Bu resurs hali quiz bilan bog'lanmagan</p>
                        <button class="btn ghost sm" on:click={() => openEditTest(dt)}>
                            🔗 Quiz bog'lash
                        </button>
                    </div>
                {/if}
                <div class="dp-edit-del">
                    <button class="btn outline sm" on:click={() => openEditTest(dt)}>✏️ Tahrirlash</button>
                    <button class="btn danger sm" on:click={() => { deleteTest(dt.id); }}>🗑️ O'chirish</button>
                </div>
            </div>
        </aside>
    {/if}

</div>

<!-- ════════════════════════════════════════════
     MODALS
════════════════════════════════════════════ -->

<!-- Category modal -->
{#if showCatModal && editingCat}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="overlay" on:click|self={() => showCatModal = false}>
        <div class="modal lg">
            <div class="modal-hd">
                <h3>{editingCat.id && categories.find(c => c.id === editingCat.id) ? 'Kategoriyani tahrirlash' : 'Yangi kategoriya'}</h3>
                <button class="modal-x" on:click={() => showCatModal = false}>✕</button>
            </div>
            <div class="modal-bd">
                <div class="form-row-2">
                    <div class="frow">
                        <label class="flabel">Ikonka</label>
                        <div class="icon-picker">
                            <span class="icon-preview">{editingCat.icon}</span>
                            <div class="icon-grid">
                                {#each PRESET_ICONS as ic}
                                    <button
                                        class="icon-opt"
                                        class:sel={editingCat.icon === ic}
                                        on:click={() => { editingCat = { ...editingCat, icon: ic }; }}
                                    >{ic}</button>
                                {/each}
                            </div>
                        </div>
                    </div>
                    <div class="frow" style="flex:2">
                        <div class="frow">
                            <label class="flabel">Sarlavha *</label>
                            <input class="finput" bind:value={editingCat.title} placeholder="Kategoriya nomi" />
                        </div>
                        <div class="frow" style="margin-top:10px">
                            <label class="flabel">Tavsif</label>
                            <input class="finput" bind:value={editingCat.subtitle} placeholder="Qisqacha tavsif" />
                        </div>
                    </div>
                </div>

                <div class="frow">
                    <label class="flabel">Rang sxemasi</label>
                    <div class="color-presets">
                        {#each PRESET_COLORS as [c1, c2]}
                            <button
                                class="color-preset"
                                class:sel={editingCat.g1 === c1}
                                style="background:linear-gradient(135deg,{c1},{c2})"
                                on:click={() => { editingCat = { ...editingCat, g1: c1, g2: c2 }; }}
                            ></button>
                        {/each}
                        <div class="color-custom">
                            <input type="color" bind:value={editingCat.g1} title="1-rang" />
                            <input type="color" bind:value={editingCat.g2} title="2-rang" />
                        </div>
                    </div>
                    <div class="color-preview" style="background:linear-gradient(135deg,{editingCat.g1},{editingCat.g2})">
                        <span>{editingCat.icon}</span> {editingCat.title || 'Kategoriya nomi'}
                    </div>
                </div>

                <div class="frow">
                    <div class="flabel-row">
                        <label class="flabel">Bo'limlar</label>
                        <button class="link-btn" on:click={() => {
                            editingCat = { ...editingCat, subcats: [...(editingCat?.subcats ?? []), { id: uid(), label: 'Yangi bo\'lim', icon: '📌' }] };
                        }}>＋ Qo'shish</button>
                    </div>
                    <div class="subcat-list">
                        {#each (editingCat.subcats ?? []) as sc, i}
                            <div class="sc-row">
                                <input class="finput sm" bind:value={sc.icon} style="width:52px;text-align:center" placeholder="📌" />
                                <input class="finput sm" bind:value={sc.label} placeholder="Bo'lim nomi" style="flex:1" />
                                {#if sc.id !== 'all'}
                                    <button class="sc-rm" on:click={() => {
                                        const s = [...(editingCat?.subcats ?? [])];
                                        s.splice(i, 1);
                                        editingCat = { ...editingCat, subcats: s };
                                    }}>×</button>
                                {/if}
                            </div>
                        {/each}
                    </div>
                </div>
            </div>
            <div class="modal-ft">
                <button class="btn ghost" on:click={() => showCatModal = false}>Bekor qilish</button>
                <button class="btn primary" on:click={saveCat} disabled={!editingCat.title?.trim()}>
                    Saqlash
                </button>
            </div>
        </div>
    </div>
{/if}

<!-- Subcategory modal -->
{#if showSubcatModal && editingSubcat}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="overlay" on:click|self={() => showSubcatModal = false}>
        <div class="modal sm">
            <div class="modal-hd">
                <h3>{isNewSubcat ? "Bo'lim qo'shish" : "Bo'limni tahrirlash"}</h3>
                <button class="modal-x" on:click={() => showSubcatModal = false}>✕</button>
            </div>
            <div class="modal-bd">
                <div class="frow"><label class="flabel">Ikonka</label>
                    <input class="finput" bind:value={editingSubcat.icon} style="width:70px;font-size:1.2rem;text-align:center" /></div>
                <div class="frow"><label class="flabel">Nomi *</label>
                    <input class="finput" bind:value={editingSubcat.label} placeholder="Bo'lim nomi" /></div>
            </div>
            <div class="modal-ft">
                <button class="btn ghost" on:click={() => showSubcatModal = false}>Bekor</button>
                <button class="btn primary" on:click={saveSubcat} disabled={!editingSubcat.label?.trim()}>Saqlash</button>
            </div>
        </div>
    </div>
{/if}

<!-- Test/Resource modal -->
{#if showTestModal && editingTest}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="overlay" on:click|self={() => showTestModal = false}>
        <div class="modal lg">
            <div class="modal-hd">
                <h3>{isNewTest ? 'Yangi resurs qo\'shish' : 'Resursni tahrirlash'}</h3>
                <button class="modal-x" on:click={() => showTestModal = false}>✕</button>
            </div>
            <div class="modal-bd">
                <div class="form-row-2">
                    <div class="frow">
                        <label class="flabel">Ikonka</label>
                        <input class="finput icon-inp" bind:value={editingTest.icon} placeholder="📝" maxlength="4" />
                    </div>
                    <div class="frow" style="flex:3">
                        <label class="flabel">Sarlavha *</label>
                        <input class="finput" bind:value={editingTest.title} placeholder="Resurs nomi" />
                    </div>
                </div>
                <div class="frow">
                    <label class="flabel">Tavsif</label>
                    <textarea class="finput" rows="2" bind:value={editingTest.description} placeholder="Qisqacha tavsif"></textarea>
                </div>
                <div class="form-row-3">
                    <div class="frow">
                        <label class="flabel">Savollar</label>
                        <input class="finput" type="number" bind:value={editingTest.questions} min="1" />
                    </div>
                    <div class="frow">
                        <label class="flabel">Vaqt (min)</label>
                        <input class="finput" type="number" bind:value={editingTest.duration} min="1" />
                    </div>
                    <div class="frow">
                        <label class="flabel">Qiyinlik</label>
                        <select class="finput" bind:value={editingTest.difficulty}>
                            <option value="easy">Oson</option>
                            <option value="medium">O'rta</option>
                            <option value="hard">Qiyin</option>
                        </select>
                    </div>
                </div>
                <div class="form-row-2">
                    <div class="frow">
                        <label class="flabel">Bo'lim</label>
                        <select class="finput" bind:value={editingTest.subcat}>
                            {#each (selectedCat?.subcats ?? []) as sc}
                                <option value={sc.id}>{sc.icon} {sc.label}</option>
                            {/each}
                        </select>
                    </div>
                    <div class="frow">
                        <label class="flabel">Reyting (1–5)</label>
                        <input class="finput" type="number" bind:value={editingTest.rating} min="1" max="5" step="0.1" />
                    </div>
                </div>

                <!-- Quiz linking -->
                <div class="frow quiz-link-section">
                    <div class="flabel-row">
                        <label class="flabel">🔗 Quiz bilan bog'lash</label>
                        <span class="flabel-hint">O'ynash tugmasini faollashtiradi</span>
                    </div>
                    {#if quizzesList.length > 0}
                        <select class="finput" bind:value={editingTest.quizId}>
                            <option value="">— Quiz tanlang (ixtiyoriy) —</option>
                            {#each quizzesList as q}
                                <option value={q.id}>{q.title} ({q.total_questions} savol)</option>
                            {/each}
                        </select>
                    {:else}
                        <p class="no-quiz-hint">Hali quiz yaratilmagan. <a href="/dashboard/quizzes/new">Yangi quiz yarating →</a></p>
                    {/if}
                </div>

                <div class="form-row-checks">
                    <label class="check-lbl">
                        <input type="checkbox" bind:checked={editingTest.isNew} />
                        <span>Yangi belgisi</span>
                    </label>
                    <label class="check-lbl">
                        <input type="checkbox" bind:checked={editingTest.isHot} />
                        <span>🔥 Trend belgisi</span>
                    </label>
                </div>
            </div>
            <div class="modal-ft">
                <button class="btn ghost" on:click={() => showTestModal = false}>Bekor qilish</button>
                <button class="btn primary" on:click={saveTest} disabled={!editingTest.title?.trim()}>
                    {isNewTest ? 'Qo\'shish' : 'Saqlash'}
                </button>
            </div>
        </div>
    </div>
{/if}

<!-- Import from quizzes modal -->
{#if showImportModal}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="overlay" on:click|self={() => showImportModal = false}>
        <div class="modal lg">
            <div class="modal-hd">
                <h3>Quizdan resurs import qilish</h3>
                <button class="modal-x" on:click={() => showImportModal = false}>✕</button>
            </div>
            <div class="modal-bd">
                <p class="import-hint">
                    Quizlaringizni "<strong>{selectedCat?.title}</strong>" kategoriyasiga resurs sifatida qo'shing.
                    Bog'liq quizlar "O'ynash" tugmasini faollashtiradi.
                </p>
                {#if quizzesList.length === 0}
                    <div class="empty-state">
                        <div class="empty-icon">📝</div>
                        <p class="empty-title">Quiz yo'q</p>
                        <a href="/dashboard/quizzes/new" class="btn primary sm">Yangi quiz yaratish</a>
                    </div>
                {:else}
                    <div class="import-list">
                        {#each quizzesList as q (q.id)}
                            {@const alreadyAdded = (selectedCat?.tests ?? []).some(t => t.quizId === q.id)}
                            <label class="import-item" class:disabled={alreadyAdded}>
                                <input
                                    type="checkbox"
                                    checked={importSelected.has(q.id) || alreadyAdded}
                                    disabled={alreadyAdded}
                                    on:change={() => {
                                        if (alreadyAdded) return;
                                        if (importSelected.has(q.id)) importSelected.delete(q.id);
                                        else importSelected.add(q.id);
                                        importSelected = new Set(importSelected);
                                    }}
                                />
                                <div class="ii-info">
                                    <span class="ii-title">{q.title}</span>
                                    <span class="ii-meta">
                                        {q.total_questions} savol
                                        {#if q.subject} · {q.subject}{/if}
                                        {#if q.grade_level} · {q.grade_level}-sinf{/if}
                                    </span>
                                </div>
                                {#if alreadyAdded}
                                    <span class="ii-added">✓ Qo'shilgan</span>
                                {/if}
                            </label>
                        {/each}
                    </div>
                {/if}
            </div>
            <div class="modal-ft">
                <button class="btn ghost" on:click={() => showImportModal = false}>Bekor qilish</button>
                <button class="btn primary" on:click={confirmImport} disabled={importSelected.size === 0}>
                    📥 {importSelected.size} ta import qilish
                </button>
            </div>
        </div>
    </div>
{/if}

<!-- Delete category confirm -->
{#if showDeleteCatConfirm}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="overlay" on:click|self={() => showDeleteCatConfirm = false}>
        <div class="modal sm confirm-modal">
            <div class="confirm-icon">⚠️</div>
            <h3 class="confirm-title">"{selectedCat?.title}" o'chirilsinmi?</h3>
            <p class="confirm-sub">Bu kategoriya ichidagi barcha {selectedCat?.tests.length} ta resurs ham o'chadi.</p>
            <div class="modal-ft">
                <button class="btn ghost" on:click={() => showDeleteCatConfirm = false}>Bekor qilish</button>
                <button class="btn danger" on:click={deleteCat}>Ha, o'chirish</button>
            </div>
        </div>
    </div>
{/if}

<!-- Toast -->
{#if toast}
    <div class="toast" class:toast-err={toastType === 'error'}>{toast}</div>
{/if}

<style>
    /* ── Page header ── */
    .page-header {
        display: flex;
        align-items: center;
        gap: 16px;
        margin-bottom: 16px;
        flex-wrap: wrap;
    }
    .header-left { display: flex; align-items: center; gap: 14px; flex: 0 0 auto; }
    .page-icon { font-size: 2.2rem; }
    .page-title { font-size: 1.5rem; font-weight: 800; color: var(--text); margin: 0 0 2px; }
    .page-sub { font-size: 0.82rem; color: var(--text3); margin: 0; }
    .header-center { flex: 1; min-width: 200px; }
    .header-right { display: flex; gap: 8px; align-items: center; flex-shrink: 0; }

    .global-search-wrap { position: relative; }
    .gs-icon { position: absolute; left: 10px; top: 50%; transform: translateY(-50%); font-size: 0.85rem; pointer-events: none; z-index:1; }
    .global-search-input {
        width: 100%; padding: 9px 14px 9px 34px;
        border: 1.5px solid var(--border); border-radius: 10px;
        font-size: 0.875rem; outline: none; background: var(--bg); color: var(--text);
        transition: border-color 0.2s; font-family: inherit;
    }
    .global-search-input:focus { border-color: var(--primary); }
    .global-results {
        position: absolute; top: calc(100% + 6px); left: 0;
        width: 360px; background: var(--bg); border-radius: 12px;
        box-shadow: 0 12px 40px rgba(99,102,241,0.15); border: 1.5px solid var(--border);
        z-index: 200; overflow: hidden;
    }
    .gr-item {
        display: flex; align-items: center; gap: 10px;
        padding: 10px 14px; cursor: pointer; transition: background 0.15s;
    }
    .gr-item:hover { background: var(--bg2); }
    .gr-icon { font-size: 1.4rem; flex-shrink: 0; }
    .gr-info { flex: 1; min-width: 0; }
    .gr-title { display: block; font-size: 0.85rem; font-weight: 600; color: var(--text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .gr-meta { display: block; font-size: 0.72rem; color: var(--text3); }
    .gr-linked { font-size: 0.8rem; }
    .gr-empty, .gr-more { padding: 12px 14px; font-size: 0.83rem; color: var(--text3); text-align: center; }

    .hdr-btn {
        padding: 8px 14px; border: 1.5px solid var(--border);
        border-radius: 9px; background: var(--bg); color: var(--text2);
        font-size: 0.83rem; font-weight: 600; cursor: pointer;
        transition: all 0.2s; white-space: nowrap; font-family: inherit;
    }
    .hdr-btn:hover { border-color: var(--primary); color: var(--primary); }
    .hdr-btn.primary {
        background: linear-gradient(135deg, var(--primary), var(--accent));
        color: white; border-color: transparent;
        box-shadow: 0 3px 10px rgba(99,102,241,0.3);
    }
    .hdr-btn.primary:hover { transform: translateY(-1px); }

    /* ── Stats strip ── */
    .stats-strip {
        display: flex; gap: 12px; margin-bottom: 20px; flex-wrap: wrap;
    }
    .stat-card {
        display: flex; align-items: center; gap: 12px;
        background: var(--bg); border: 1.5px solid var(--border);
        border-radius: 12px; padding: 12px 18px; flex: 1; min-width: 140px;
        box-shadow: 0 1px 4px rgba(0,0,0,0.05);
    }
    .stat-icon { font-size: 1.6rem; }
    .stat-info { display: flex; flex-direction: column; }
    .stat-val {
        font-size: 1.3rem; font-weight: 800;
        background: linear-gradient(135deg, var(--primary), var(--accent));
        -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text;
        line-height: 1;
    }
    .stat-lbl { font-size: 0.72rem; color: var(--text3); margin-top: 2px; }

    /* ── Workspace ── */
    .workspace {
        display: grid;
        grid-template-columns: 220px 1fr;
        gap: 0;
        background: var(--bg);
        border: 1.5px solid var(--border);
        border-radius: 16px;
        overflow: hidden;
        box-shadow: 0 4px 20px rgba(0,0,0,0.07);
        min-height: 600px;
    }
    .workspace.has-detail {
        grid-template-columns: 220px 1fr 320px;
    }

    /* ── Sidebar (sahifa ichidagi kategoriya ro'yxati — yorug' panel) ── */
    .sidebar {
        background: var(--white);
        display: flex; flex-direction: column;
        border-right: 1px solid var(--border);
    }
    .sidebar-top {
        display: flex; align-items: center; justify-content: space-between;
        padding: 14px 14px 10px;
        border-bottom: 1px solid var(--border);
    }
    .sb-label { font-size: 0.65rem; font-weight: 700; color: var(--text3); letter-spacing: 0.08em; text-transform: uppercase; }
    .sb-add-btn {
        width: 24px; height: 24px; border-radius: 6px;
        background: var(--primary-light); border: 1px solid rgba(99,102,241,0.25);
        color: var(--primary); font-size: 1rem; line-height: 1;
        cursor: pointer; display: flex; align-items: center; justify-content: center;
        transition: all 0.15s;
    }
    .sb-add-btn:hover { background: rgba(99,102,241,0.2); }

    .cat-list { flex: 1; overflow-y: auto; padding: 8px 10px; display: flex; flex-direction: column; gap: 2px; }
    .cat-item {
        display: flex; align-items: center; gap: 9px;
        padding: 9px 10px; border-radius: 9px;
        border: none; background: transparent; cursor: pointer;
        width: 100%; text-align: left; color: var(--text2);
        transition: all 0.18s; position: relative;
    }
    .cat-item:hover { color: var(--text); background: var(--bg); }
    .cat-item.active {
        background: linear-gradient(135deg, color-mix(in srgb,var(--g1) 16%,transparent), color-mix(in srgb,var(--g2) 9%,transparent));
        box-shadow: inset 3px 0 0 var(--g1); color: var(--text);
    }
    .ci-icon { font-size: 1.15rem; flex-shrink: 0; width: 24px; text-align: center; }
    .ci-info { flex: 1; min-width: 0; }
    .ci-name { display: block; font-size: 0.8rem; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .ci-count { font-size: 0.66rem; color: var(--text3); }
    .cat-item.active .ci-count { color: var(--text3); }
    .ci-dot { width: 6px; height: 6px; border-radius: 50%; background: var(--g1); box-shadow: 0 0 8px var(--g1); flex-shrink: 0; }

    /* ── Center panel ── */
    .center-panel {
        display: flex; flex-direction: column;
        background: #f8faff; min-width: 0; overflow-y: auto;
    }

    /* Category hero */
    .cat-hero { position: relative; flex-shrink: 0; overflow: hidden; }
    .cat-hero-bg { position: absolute; inset: 0; background: linear-gradient(135deg,var(--g1),var(--g2)); opacity: 0.9; }
    .cat-hero-content {
        position: relative; z-index: 1;
        display: flex; align-items: center; gap: 14px; padding: 18px 22px;
        flex-wrap: wrap;
    }
    .cat-hero-left { display: flex; align-items: center; gap: 12px; flex: 1; min-width: 200px; }
    .cat-hero-icon { font-size: 2.5rem; filter: drop-shadow(0 3px 8px rgba(0,0,0,0.2)); flex-shrink: 0; }
    .cat-hero-title { font-size: 1.2rem; font-weight: 800; color: white; margin: 0 0 3px; }
    .cat-hero-sub { font-size: 0.78rem; color: rgba(255,255,255,0.8); margin: 0; }
    .cat-hero-meta { display: flex; gap: 8px; flex-shrink: 0; }
    .cm {
        display: flex; flex-direction: column; align-items: center;
        background: rgba(255,255,255,0.18); backdrop-filter: blur(8px);
        border-radius: 8px; padding: 6px 12px;
    }
    .cm strong { font-size: 1rem; font-weight: 800; color: white; line-height: 1; }
    .cm span { font-size: 0.65rem; color: rgba(255,255,255,0.75); margin-top: 2px; }
    .cat-hero-actions { display: flex; gap: 7px; flex-shrink: 0; }
    .hero-btn {
        padding: 6px 14px; border-radius: 8px;
        background: rgba(255,255,255,0.2); border: 1px solid rgba(255,255,255,0.35);
        color: white; font-size: 0.8rem; font-weight: 600; cursor: pointer;
        transition: all 0.18s; backdrop-filter: blur(4px); font-family: inherit;
    }
    .hero-btn:hover { background: rgba(255,255,255,0.3); }
    .hero-btn.danger { background: rgba(239,68,68,0.25); border-color: rgba(239,68,68,0.4); }
    .hero-btn.danger:hover { background: rgba(239,68,68,0.4); }

    /* Subcategory bar */
    .subcat-bar { background: var(--bg); border-bottom: 1px solid var(--border); flex-shrink: 0; }
    .subcat-tabs {
        display: flex; gap: 4px; padding: 10px 18px;
        overflow-x: auto; align-items: center;
    }
    .sc-tab {
        display: flex; align-items: center; gap: 5px;
        padding: 5px 12px; border: 1.5px solid var(--border);
        border-radius: 20px; background: var(--bg);
        font-size: 0.78rem; font-weight: 500; color: var(--text2);
        cursor: pointer; white-space: nowrap; transition: all 0.18s;
        position: relative; flex-shrink: 0;
    }
    .sc-tab:hover { border-color: var(--g1); color: var(--text); }
    .sc-tab.active {
        background: linear-gradient(135deg,var(--g1),var(--g2));
        border-color: transparent; color: white; font-weight: 600;
        box-shadow: 0 2px 8px rgba(0,0,0,0.15);
    }
    .sc-count {
        background: rgba(0,0,0,0.1); border-radius: 99px;
        padding: 0 6px; font-size: 0.68rem; font-weight: 700;
    }
    .sc-tab.active .sc-count { background: rgba(255,255,255,0.25); }
    .sc-del {
        font-size: 0.75rem; opacity: 0; margin-left: 2px;
        color: rgba(255,255,255,0.7); transition: opacity 0.15s; cursor: pointer; font-weight: 700;
    }
    .sc-tab:hover .sc-del { opacity: 1; }
    .sc-tab:not(.active) .sc-del { color: var(--text3); }
    .sc-add-btn {
        padding: 5px 12px; border: 1.5px dashed var(--border);
        border-radius: 20px; background: transparent;
        font-size: 0.78rem; color: var(--text3); cursor: pointer;
        transition: all 0.18s; white-space: nowrap; font-family: inherit;
    }
    .sc-add-btn:hover { border-color: var(--primary); color: var(--primary); }

    /* Toolbar */
    .toolbar {
        display: flex; align-items: center; justify-content: space-between;
        gap: 10px; padding: 12px 18px; flex-shrink: 0; flex-wrap: wrap;
        background: var(--bg); border-bottom: 1px solid var(--border);
    }
    .search-wrap { position: relative; display: flex; align-items: center; flex: 1; max-width: 280px; }
    .si { position: absolute; left: 10px; font-size: 0.85rem; pointer-events: none; }
    .search-input {
        width: 100%; padding: 7px 30px 7px 30px;
        border: 1.5px solid var(--border); border-radius: 8px;
        font-size: 0.82rem; background: var(--bg2); color: var(--text);
        outline: none; font-family: inherit; transition: border-color 0.2s;
    }
    .search-input:focus { border-color: var(--primary); background: var(--bg); }
    .sc { position: absolute; right: 8px; background: none; border: none; cursor: pointer; color: var(--text3); font-size: 0.78rem; }
    .toolbar-right { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
    .result-count { font-size: 0.78rem; color: var(--text3); white-space: nowrap; }
    .sort-select {
        padding: 6px 10px; border: 1.5px solid var(--border); border-radius: 8px;
        font-size: 0.78rem; color: var(--text2); background: var(--bg);
        cursor: pointer; outline: none; font-family: inherit;
    }

    /* Bulk bar */
    .bulk-bar {
        display: flex; align-items: center; gap: 10px;
        padding: 7px 18px; background: rgba(99,102,241,0.05);
        border-bottom: 1px solid var(--border); flex-shrink: 0;
    }
    .bulk-check { display: flex; align-items: center; gap: 6px; font-size: 0.8rem; font-weight: 600; color: var(--text2); cursor: pointer; }
    .bulk-count { font-size: 0.8rem; color: var(--primary); font-weight: 700; }
    .bulk-del {
        padding: 4px 12px; background: #fee2e2; color: #b91c1c;
        border: none; border-radius: 7px; font-size: 0.78rem; font-weight: 700;
        cursor: pointer; transition: all 0.15s; font-family: inherit;
    }
    .bulk-del:hover { background: #fecaca; }

    /* Empty state */
    .empty-state {
        display: flex; flex-direction: column; align-items: center; justify-content: center;
        gap: 10px; padding: 60px 24px; text-align: center; color: var(--text3); flex: 1;
    }
    .empty-state.big { font-size: 1.1rem; }
    .empty-icon { font-size: 3rem; }
    .empty-title { font-size: 1rem; font-weight: 700; color: var(--text2); margin: 0; }
    .empty-sub { font-size: 0.84rem; color: var(--text3); margin: 0; }
    .empty-actions { display: flex; gap: 8px; flex-wrap: wrap; justify-content: center; }

    /* Tests grid */
    .tests-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
        gap: 14px; padding: 18px; align-content: start;
    }

    /* Test card */
    .test-card {
        background: var(--bg); border-radius: 12px;
        border: 1.5px solid var(--border); overflow: hidden;
        display: flex; flex-direction: column;
        transition: all 0.2s; cursor: pointer;
        box-shadow: 0 1px 4px rgba(0,0,0,0.05);
    }
    .test-card:hover { border-color: var(--g1, var(--primary)); box-shadow: 0 6px 22px rgba(0,0,0,0.1); transform: translateY(-2px); }
    .test-card.selected { border-color: var(--primary); background: rgba(99,102,241,0.04); }
    .test-card.detail-open { border-color: var(--primary); box-shadow: 0 0 0 2px rgba(99,102,241,0.2); }

    .tc-header { display: flex; align-items: flex-start; justify-content: space-between; padding: 12px 12px 4px; }
    .tc-check input { width: 15px; height: 15px; accent-color: var(--primary); cursor: pointer; }
    .tc-icon { font-size: 1.6rem; }
    .tc-badges { display: flex; gap: 4px; flex-wrap: wrap; justify-content: flex-end; }

    .tc-body { padding: 0 12px 8px; flex: 1; }
    .tc-title { font-size: 0.88rem; font-weight: 700; color: var(--text); margin: 4px 0 6px; line-height: 1.3; }
    .tc-desc {
        font-size: 0.76rem; color: var(--text3); margin: 0 0 8px; line-height: 1.45;
        display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
    }
    .tc-meta { display: flex; gap: 5px; flex-wrap: wrap; }
    .chip {
        font-size: 0.68rem; padding: 2px 7px; border-radius: 99px;
        background: var(--bg2); color: var(--text2); font-weight: 500; white-space: nowrap;
    }
    .diff-easy { background: #dcfce7; color: #16a34a; }
    .diff-medium { background: #fef9c3; color: #ca8a04; }
    .diff-hard { background: #fee2e2; color: #dc2626; }

    .tc-footer {
        display: flex; align-items: center; justify-content: space-between;
        padding: 8px 12px; border-top: 1px solid var(--border);
        background: var(--bg2);
    }
    .tc-plays { font-size: 0.72rem; color: var(--text3); }
    .tc-actions { display: flex; gap: 5px; }
    .act {
        padding: 4px 9px; border: none; border-radius: 6px;
        font-size: 0.72rem; font-weight: 600; cursor: pointer;
        transition: all 0.15s; font-family: inherit;
    }
    .act.play { background: linear-gradient(135deg,var(--primary),var(--accent)); color: white; padding: 4px 10px; }
    .act.play:hover:not(:disabled) { transform: translateY(-1px); }
    .act.play:disabled { opacity: 0.5; cursor: not-allowed; }
    .act.edit { background: #ede9fe; color: #5b21b6; }
    .act.edit:hover { background: #ddd6fe; }
    .act.del { background: #fee2e2; color: #b91c1c; }
    .act.del:hover { background: #fecaca; }

    /* Badges */
    .badge {
        font-size: 0.62rem; font-weight: 700; padding: 2px 7px;
        border-radius: 99px; text-transform: uppercase;
    }
    .badge.new { background: #dcfce7; color: #16a34a; }
    .badge.hot { background: #fef3c7; color: #d97706; }
    .badge.linked { background: #ede9fe; color: #5b21b6; font-size: 0.75rem; }
    .badge.diff-easy { background: #dcfce7; color: #16a34a; text-transform: none; }
    .badge.diff-medium { background: #fef9c3; color: #ca8a04; text-transform: none; }
    .badge.diff-hard { background: #fee2e2; color: #dc2626; text-transform: none; }

    /* ── Detail panel ── */
    .detail-panel {
        background: var(--bg); border-left: 1.5px solid var(--border);
        display: flex; flex-direction: column; overflow-y: auto;
    }
    .dp-header {
        display: flex; align-items: center; justify-content: space-between;
        padding: 14px 18px; border-bottom: 1px solid var(--border); flex-shrink: 0;
    }
    .dp-header h3 { font-size: 0.9rem; font-weight: 700; color: var(--text); margin: 0; }
    .dp-close {
        width: 28px; height: 28px; border-radius: 7px; border: none;
        background: var(--bg2); color: var(--text3); cursor: pointer; font-size: 0.9rem;
        display: flex; align-items: center; justify-content: center; transition: all 0.15s;
    }
    .dp-close:hover { background: var(--border); color: var(--text); }

    .dp-cover {
        padding: 20px 18px 12px;
        background: linear-gradient(135deg, #f8faff, #eef2ff);
        border-bottom: 1px solid var(--border);
        display: flex; align-items: center; gap: 12px;
    }
    .dp-big-icon { font-size: 3rem; }
    .dp-badges { display: flex; flex-direction: column; gap: 5px; }

    .dp-body { padding: 16px 18px; flex: 1; display: flex; flex-direction: column; gap: 12px; }
    .dp-title { font-size: 1.05rem; font-weight: 800; color: var(--text); margin: 0; }
    .dp-desc { font-size: 0.82rem; color: var(--text2); margin: 0; line-height: 1.55; }
    .dp-stats {
        display: grid; grid-template-columns: 1fr 1fr; gap: 8px;
    }
    .dps {
        display: flex; flex-direction: column; align-items: center;
        background: var(--bg2); border-radius: 10px; padding: 8px;
        border: 1px solid var(--border);
    }
    .dps-val { font-size: 1rem; font-weight: 800; color: var(--primary); }
    .dps-lbl { font-size: 0.68rem; color: var(--text3); margin-top: 2px; }

    .dp-tags { display: flex; gap: 5px; flex-wrap: wrap; }
    .dp-tag { font-size: 0.72rem; padding: 3px 9px; background: #ede9fe; color: #5b21b6; border-radius: 99px; font-weight: 600; }

    .dp-quiz-link {
        display: flex; align-items: center; gap: 10px;
        background: #f0fdf4; border: 1.5px solid #bbf7d0;
        border-radius: 10px; padding: 10px 14px;
    }
    .dql-icon { font-size: 1.3rem; }
    .dql-label { display: block; font-size: 0.68rem; color: #16a34a; font-weight: 700; text-transform: uppercase; }
    .dql-title { display: block; font-size: 0.84rem; font-weight: 600; color: var(--text); }

    .dp-footer {
        padding: 14px 18px; border-top: 1px solid var(--border);
        display: flex; flex-direction: column; gap: 8px; flex-shrink: 0;
        background: var(--bg2);
    }
    .dp-no-quiz { text-align: center; }
    .dp-no-quiz p { font-size: 0.78rem; color: var(--text3); margin: 0 0 8px; }
    .dp-edit-del { display: flex; gap: 8px; }

    /* ── Modals ── */
    .overlay {
        position: fixed; inset: 0;
        background: rgba(15,23,42,0.55); backdrop-filter: blur(4px);
        display: flex; align-items: center; justify-content: center;
        z-index: 1000; padding: 20px;
    }
    .modal {
        background: var(--bg); border-radius: 16px;
        width: 100%; max-height: 90vh;
        display: flex; flex-direction: column;
        box-shadow: 0 20px 60px rgba(0,0,0,0.25); overflow: hidden;
    }
    .modal.sm { max-width: 400px; }
    .modal.lg { max-width: 560px; }
    .modal-hd {
        display: flex; align-items: center; justify-content: space-between;
        padding: 16px 22px; border-bottom: 1px solid var(--border); flex-shrink: 0;
    }
    .modal-hd h3 { font-size: 1rem; font-weight: 700; color: var(--text); margin: 0; }
    .modal-x {
        width: 28px; height: 28px; border-radius: 7px; border: none;
        background: var(--bg2); color: var(--text3); cursor: pointer; font-size: 0.9rem;
        display: flex; align-items: center; justify-content: center;
    }
    .modal-x:hover { background: var(--border); color: var(--text); }
    .modal-bd {
        padding: 18px 22px; overflow-y: auto; flex: 1;
        display: flex; flex-direction: column; gap: 14px;
    }
    .modal-ft {
        display: flex; gap: 10px; justify-content: flex-end;
        padding: 14px 22px; border-top: 1px solid var(--border);
        background: var(--bg2); flex-shrink: 0;
    }

    /* Forms */
    .frow { display: flex; flex-direction: column; gap: 5px; }
    .flabel { font-size: 0.76rem; font-weight: 600; color: var(--text2); }
    .flabel-row { display: flex; align-items: center; justify-content: space-between; }
    .flabel-hint { font-size: 0.72rem; color: var(--text3); }
    .link-btn { font-size: 0.75rem; color: var(--primary); background: none; border: none; cursor: pointer; font-family: inherit; font-weight: 600; }
    .finput {
        padding: 8px 11px; border: 1.5px solid var(--border); border-radius: 8px;
        font-size: 0.85rem; color: var(--text); background: var(--bg);
        outline: none; transition: border-color 0.2s; font-family: inherit;
        width: 100%; box-sizing: border-box;
    }
    .finput:focus { border-color: var(--primary); }
    .finput.sm { padding: 6px 9px; font-size: 0.8rem; }
    .icon-inp { width: 70px; text-align: center; font-size: 1.3rem; }
    .form-row-2 { display: grid; grid-template-columns: auto 1fr; gap: 12px; align-items: start; }
    .form-row-3 { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 10px; }
    .form-row-checks { display: flex; gap: 20px; }
    .check-lbl { display: flex; align-items: center; gap: 7px; cursor: pointer; font-size: 0.83rem; color: var(--text); }
    .check-lbl input { width: 15px; height: 15px; accent-color: var(--primary); cursor: pointer; }

    /* Icon picker */
    .icon-picker { display: flex; align-items: flex-start; gap: 10px; }
    .icon-preview { font-size: 2rem; width: 44px; height: 44px; background: var(--bg2); border-radius: 10px; display: flex; align-items: center; justify-content: center; border: 1.5px solid var(--border); flex-shrink: 0; }
    .icon-grid { display: flex; flex-wrap: wrap; gap: 4px; }
    .icon-opt {
        width: 32px; height: 32px; border: 1.5px solid var(--border); border-radius: 7px;
        background: var(--bg); font-size: 1rem; cursor: pointer; transition: all 0.15s;
        display: flex; align-items: center; justify-content: center;
    }
    .icon-opt:hover { border-color: var(--primary); background: var(--bg2); }
    .icon-opt.sel { border-color: var(--primary); background: rgba(99,102,241,0.1); }

    /* Color presets */
    .color-presets { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; }
    .color-preset {
        width: 32px; height: 32px; border-radius: 8px; border: 2px solid transparent;
        cursor: pointer; transition: all 0.15s;
    }
    .color-preset:hover, .color-preset.sel { border-color: var(--text); transform: scale(1.1); }
    .color-custom { display: flex; gap: 4px; }
    .color-custom input[type="color"] { width: 32px; height: 32px; border: 1.5px solid var(--border); border-radius: 8px; cursor: pointer; padding: 2px; }
    .color-preview {
        margin-top: 8px; padding: 12px 16px; border-radius: 10px;
        color: white; font-weight: 700; font-size: 0.9rem;
        display: flex; align-items: center; gap: 8px;
    }
    .color-preview span { font-size: 1.3rem; }

    /* Subcats list in cat modal */
    .subcat-list { display: flex; flex-direction: column; gap: 6px; }
    .sc-row { display: flex; align-items: center; gap: 7px; }
    .sc-rm { width: 24px; height: 24px; border-radius: 6px; border: none; background: #fee2e2; color: #b91c1c; cursor: pointer; font-size: 1rem; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }

    /* Quiz link section */
    .quiz-link-section { background: #f0fdf4; border: 1.5px solid #bbf7d0; border-radius: 10px; padding: 12px 14px; }
    .no-quiz-hint { font-size: 0.82rem; color: var(--text3); margin: 0; }
    .no-quiz-hint a { color: var(--primary); text-decoration: none; font-weight: 600; }

    /* Import modal */
    .import-hint { font-size: 0.84rem; color: var(--text2); margin: 0 0 6px; }
    .import-list { display: flex; flex-direction: column; gap: 6px; max-height: 320px; overflow-y: auto; }
    .import-item {
        display: flex; align-items: center; gap: 10px;
        padding: 10px 14px; border: 1.5px solid var(--border);
        border-radius: 10px; cursor: pointer; transition: all 0.15s;
        background: var(--bg);
    }
    .import-item:hover:not(.disabled) { border-color: var(--primary); background: #fafbff; }
    .import-item.disabled { opacity: 0.6; cursor: default; }
    .import-item input[type="checkbox"] { width: 16px; height: 16px; accent-color: var(--primary); flex-shrink: 0; }
    .ii-info { flex: 1; min-width: 0; }
    .ii-title { display: block; font-size: 0.875rem; font-weight: 600; color: var(--text); }
    .ii-meta { font-size: 0.72rem; color: var(--text3); }
    .ii-added { font-size: 0.72rem; color: #16a34a; font-weight: 700; flex-shrink: 0; }

    /* Confirm modal */
    .confirm-modal { padding: 0; }
    .confirm-modal .modal-ft { justify-content: center; }
    .confirm-icon { font-size: 3rem; text-align: center; padding-top: 24px; }
    .confirm-title { font-size: 1rem; font-weight: 700; color: var(--text); text-align: center; margin: 8px 22px 4px; }
    .confirm-sub { font-size: 0.84rem; color: var(--text3); text-align: center; margin: 0 22px 16px; }

    /* ── Buttons ── */
    .btn {
        padding: 9px 18px; border: none; border-radius: 8px;
        font-size: 0.875rem; font-weight: 600; cursor: pointer;
        display: inline-flex; align-items: center; gap: 6px;
        transition: all 0.2s; text-decoration: none; font-family: inherit;
    }
    .btn.sm { padding: 6px 13px; font-size: 0.8rem; }
    .btn.primary {
        background: linear-gradient(135deg, var(--primary), var(--accent));
        color: white; box-shadow: 0 3px 10px rgba(99,102,241,0.3);
    }
    .btn.primary:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 16px rgba(99,102,241,0.4); }
    .btn.primary:disabled { opacity: 0.5; cursor: not-allowed; }
    .btn.ghost { background: var(--bg2); color: var(--text); border: 1.5px solid var(--border); }
    .btn.ghost:hover { background: var(--border); }
    .btn.outline { background: transparent; color: var(--primary); border: 1.5px solid var(--primary); }
    .btn.outline:hover { background: rgba(99,102,241,0.07); }
    .btn.danger { background: #fee2e2; color: #b91c1c; border: 1.5px solid #fecaca; }
    .btn.danger:hover { background: #fecaca; }
    .spin {
        width: 14px; height: 14px; border: 2px solid rgba(255,255,255,0.4);
        border-top-color: white; border-radius: 50%;
        animation: spin 0.6s linear infinite; display: inline-block;
    }
    @keyframes spin { to { transform: rotate(360deg); } }

    /* ── Toast ── */
    .toast {
        position: fixed; bottom: 28px; left: 50%; transform: translateX(-50%);
        background: #15803d; color: white; padding: 11px 22px;
        border-radius: 10px; font-size: 0.875rem; font-weight: 600;
        box-shadow: 0 8px 28px rgba(0,0,0,0.2); z-index: 9999; white-space: nowrap;
        animation: toastUp 0.3s cubic-bezier(0.34,1.56,0.64,1) both;
    }
    .toast.toast-err { background: #b91c1c; }
    @keyframes toastUp {
        from { opacity: 0; transform: translateX(-50%) translateY(14px); }
        to   { opacity: 1; transform: translateX(-50%) translateY(0); }
    }

    /* ── Responsive ── */
    @media (max-width: 960px) {
        .workspace { grid-template-columns: 1fr; }
        .workspace.has-detail { grid-template-columns: 1fr; }
        .sidebar { flex-direction: row; overflow-x: auto; max-height: 80px; }
        .sidebar-top { display: none; }
        .cat-list { flex-direction: row; padding: 8px; gap: 6px; }
        .cat-item { flex-shrink: 0; width: auto; min-width: 100px; }
        .ci-count { display: none; }
        .detail-panel { border-left: none; border-top: 1.5px solid var(--border); }
        .page-header { flex-direction: column; align-items: stretch; }
        .header-center { order: -1; }
    }
    @media (max-width: 600px) {
        .tests-grid { grid-template-columns: 1fr; padding: 12px; }
        .cat-hero-meta { display: none; }
        .form-row-3 { grid-template-columns: 1fr 1fr; }
    }
</style>
