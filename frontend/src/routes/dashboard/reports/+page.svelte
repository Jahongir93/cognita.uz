<script lang="ts">
    import { onMount } from 'svelte';
    import { rooms as roomsApi } from '$lib/api/client';
    import type { RoomHistory } from '$lib/api/client';
    import * as XLSX from 'xlsx';

    // ── Types ─────────────────────────────────────────────────────────────────
    interface PlayerResult {
        rank: number;
        nickname: string;
        avatar: string;
        score: number;
        streak: number;
        correct_count: number;
        total_answered: number;
        avg_time_ms: number;
    }

    // ── State ─────────────────────────────────────────────────────────────────
    let history: RoomHistory[] = [];
    let loading = true;
    let filter: 'all' | 'week' | 'today' = 'all';
    let sortBy: 'date' | 'players' = 'date';
    let search = '';

    // detail modal
    let selectedRoom: RoomHistory | null = null;
    let playerResults: PlayerResult[] = [];
    let resultsLoading = false;
    let showModal = false;

    // ── Load data ─────────────────────────────────────────────────────────────
    onMount(async () => {
        try { history = await roomsApi.history(); }
        catch { history = []; }
        finally { loading = false; }
    });

    async function openRoom(room: RoomHistory) {
        selectedRoom = room;
        showModal = true;
        resultsLoading = true;
        playerResults = [];
        try {
            const token = localStorage.getItem('token') ?? '';
            const res = await fetch(`http://localhost:8080/api/rooms/${room.id}/results`, {
                headers: { Authorization: `Bearer ${token}` }
            });
            if (res.ok) playerResults = await res.json();
        } catch { playerResults = []; }
        finally { resultsLoading = false; }
    }

    function closeModal() { showModal = false; selectedRoom = null; playerResults = []; }

    // ── Filters ───────────────────────────────────────────────────────────────
    $: filtered = history.filter(h => {
        if (search && !h.quiz_title.toLowerCase().includes(search.toLowerCase())) return false;
        if (filter === 'today') return new Date(h.created_at).toDateString() === new Date().toDateString();
        if (filter === 'week') {
            const weekAgo = new Date(); weekAgo.setDate(weekAgo.getDate() - 7);
            return new Date(h.created_at) >= weekAgo;
        }
        return true;
    });

    $: sorted = [...filtered].sort((a, b) =>
        sortBy === 'players'
            ? b.player_count - a.player_count
            : new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    );

    $: stats = {
        total: filtered.length,
        completed: filtered.filter(h => h.status === 'completed').length,
        players: filtered.reduce((s, h) => s + h.player_count, 0),
        avgPlayers: filtered.length
            ? Math.round(filtered.reduce((s, h) => s + h.player_count, 0) / filtered.length)
            : 0,
    };

    $: chartData = (() => {
        const days = Array.from({ length: 7 }, (_, i) => {
            const d = new Date(); d.setDate(d.getDate() - (6 - i));
            return { label: d.toLocaleDateString('uz-UZ', { weekday: 'short' }), date: d.toDateString(), count: 0 };
        });
        history.forEach(h => {
            const ds = new Date(h.created_at).toDateString();
            const day = days.find(d => d.date === ds);
            if (day) day.count++;
        });
        return days;
    })();
    $: maxCount = Math.max(...chartData.map(d => d.count), 1);

    // ── Helpers ───────────────────────────────────────────────────────────────
    function formatDate(d: string) {
        return new Date(d).toLocaleDateString('uz-UZ', { day: '2-digit', month: 'short', year: 'numeric' });
    }
    function formatDateFull(d: string) {
        return new Date(d).toLocaleDateString('uz-UZ', { day: '2-digit', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' });
    }
    function duration(start: string, end: string | null) {
        if (!end) return '—';
        const ms = new Date(end).getTime() - new Date(start).getTime();
        const min = Math.floor(ms / 60000);
        const sec = Math.floor((ms % 60000) / 1000);
        return `${min}:${sec.toString().padStart(2, '0')}`;
    }
    function statusLabel(s: string) {
        const map: Record<string, string> = { waiting: 'Kutmoqda', in_progress: 'Davom etmoqda', completed: 'Tugadi', abandoned: 'Bekor' };
        return map[s] ?? s;
    }
    function modeLabel(m: string) {
        const map: Record<string, string> = { classic: '⚡ Classic', self_paced: '📲 Mustaqil', team: '👥 Jamoaviy', accuracy: '🎯 Aniqlik', confidence: '💡 Ishonch', zero_stakes: '🌟 Mashq' };
        return map[m] ?? m;
    }
    function accuracy(p: PlayerResult) {
        return p.total_answered > 0 ? Math.round(p.correct_count / p.total_answered * 100) : 0;
    }
    function avgTimeSec(ms: number) { return (ms / 1000).toFixed(1) + 's'; }
    function medalEmoji(rank: number) { return rank === 1 ? '🥇' : rank === 2 ? '🥈' : rank === 3 ? '🥉' : `#${rank}`; }
    function scorePercent(score: number) {
        if (!playerResults.length) return 0;
        const max = playerResults[0].score || 1;
        return Math.round(score / max * 100);
    }

    // ── XLSX Export ───────────────────────────────────────────────────────────
    function downloadXLSX() {
        const wb = XLSX.utils.book_new();

        // ── Sheet 1: Summary ───────────────────────────────────────────────
        const summaryData = [
            ['Cognita.uz — O\'yin Hisoboti'],
            ['Hisobot sanasi', new Date().toLocaleDateString('uz-UZ')],
            [],
            ['Ko\'rsatkich', 'Qiymat'],
            ['Jami o\'yinlar', stats.total],
            ['Tugatilgan', stats.completed],
            ['Jami o\'yinchilar', stats.players],
            ["O'rtacha ishtirokchilar", stats.avgPlayers],
        ];
        const ws1 = XLSX.utils.aoa_to_sheet(summaryData);
        ws1['!cols'] = [{ wch: 28 }, { wch: 20 }];
        ws1['!merges'] = [{ s: { r: 0, c: 0 }, e: { r: 0, c: 1 } }];
        styleCell(ws1, 'A1', { font: { bold: true, sz: 16, color: { rgb: '6366F1' } } });
        styleCell(ws1, 'A4', { font: { bold: true, color: { rgb: 'FFFFFF' } }, fill: { fgColor: { rgb: '6366F1' } } });
        styleCell(ws1, 'B4', { font: { bold: true, color: { rgb: 'FFFFFF' } }, fill: { fgColor: { rgb: '6366F1' } } });
        XLSX.utils.book_append_sheet(wb, ws1, 'Umumiy');

        // ── Sheet 2: Sessions ──────────────────────────────────────────────
        const headers = ['#', 'Quiz nomi', 'PIN', 'Rejim', "O'yinchilar", 'Davomiyligi', 'Sana', 'Holat'];
        const rows = sorted.map((h, i) => [
            i + 1,
            h.quiz_title,
            h.pin,
            modeLabel(h.game_mode).replace(/[⚡👥🎯💡🌟]/g, '').trim(),
            h.player_count,
            duration(h.created_at, h.ended_at),
            formatDate(h.created_at),
            statusLabel(h.status),
        ]);
        const ws2 = XLSX.utils.aoa_to_sheet([headers, ...rows]);
        ws2['!cols'] = [{ wch: 5 }, { wch: 32 }, { wch: 10 }, { wch: 14 }, { wch: 12 }, { wch: 12 }, { wch: 18 }, { wch: 14 }];
        styleHeaderRow(ws2, headers.length);
        XLSX.utils.book_append_sheet(wb, ws2, 'Sessiyalar');

        // ── Sheet 3: Selected room players (if modal open) ─────────────────
        if (selectedRoom && playerResults.length) {
            const ph = ['#', 'O\'yinchi', 'Avatar', 'Ball', 'Streak', "To'g'ri", 'Jami', "Aniqlik %", "O'rt. vaqt"];
            const pr = playerResults.map(p => [
                p.rank,
                p.nickname,
                p.avatar,
                p.score,
                p.streak,
                p.correct_count,
                p.total_answered,
                accuracy(p),
                avgTimeSec(p.avg_time_ms),
            ]);
            const ws3 = XLSX.utils.aoa_to_sheet([ph, ...pr]);
            ws3['!cols'] = [{ wch: 5 }, { wch: 20 }, { wch: 8 }, { wch: 10 }, { wch: 10 }, { wch: 10 }, { wch: 10 }, { wch: 12 }, { wch: 12 }];
            styleHeaderRow(ws3, ph.length);
            // Color top 3 rows
            ['A2','B2','C2','D2','E2','F2','G2','H2','I2'].forEach(c => styleCell(ws3, c, { fill: { fgColor: { rgb: 'FFF8DC' } } }));
            ['A3','B3','C3','D3','E3','F3','G3','H3','I3'].forEach(c => styleCell(ws3, c, { fill: { fgColor: { rgb: 'F0F0F0' } } }));
            ['A4','B4','C4','D4','E4','F4','G4','H4','I4'].forEach(c => styleCell(ws3, c, { fill: { fgColor: { rgb: 'FDF5E6' } } }));
            XLSX.utils.book_append_sheet(wb, ws3, selectedRoom.quiz_title.slice(0, 28));
        }

        const filename = `hisobot-${new Date().toISOString().slice(0, 10)}.xlsx`;
        XLSX.writeFile(wb, filename);
    }

    function styleCell(ws: XLSX.WorkSheet, addr: string, style: Record<string, unknown>) {
        if (!ws[addr]) ws[addr] = { t: 's', v: '' };
        ws[addr].s = style;
    }
    function styleHeaderRow(ws: XLSX.WorkSheet, cols: number) {
        for (let c = 0; c < cols; c++) {
            const addr = XLSX.utils.encode_cell({ r: 0, c });
            if (!ws[addr]) ws[addr] = { t: 's', v: '' };
            ws[addr].s = {
                font: { bold: true, color: { rgb: 'FFFFFF' }, sz: 11 },
                fill: { fgColor: { rgb: '6366F1' } },
                alignment: { horizontal: 'center', vertical: 'center' },
                border: { bottom: { style: 'thin', color: { rgb: '4F46E5' } } },
            };
        }
    }
</script>

<svelte:head><title>Hisobotlar — Cognita.uz</title></svelte:head>

<!-- ═══════════════════════════════════════════════════════ PAGE HEADER -->
<div class="page-header">
    <div class="header-left">
        <h1>Hisobotlar</h1>
        <p class="subtitle">O'yin tarixingiz va statistikangiz</p>
    </div>
    <button class="btn primary" on:click={downloadXLSX} disabled={sorted.length === 0}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
        Excel yuklab olish
    </button>
</div>

<!-- ═══════════════════════════════════════════════════════ KPI CARDS -->
{#if !loading}
<div class="kpi-grid">
    <div class="kpi-card">
        <div class="kpi-icon" style="background:linear-gradient(135deg,#6366f1,#8b5cf6)">🎮</div>
        <div class="kpi-body">
            <div class="kpi-value">{stats.total}</div>
            <div class="kpi-label">Jami o'yinlar</div>
            <div class="kpi-sub"><span class="trend-up">↑</span> Barcha davr</div>
        </div>
    </div>
    <div class="kpi-card">
        <div class="kpi-icon" style="background:linear-gradient(135deg,#22c55e,#16a34a)">✅</div>
        <div class="kpi-body">
            <div class="kpi-value">{stats.completed}</div>
            <div class="kpi-label">Tugatilgan</div>
            <div class="kpi-sub">{stats.total > 0 ? Math.round(stats.completed/stats.total*100) : 0}% jami o'yinlardan</div>
        </div>
    </div>
    <div class="kpi-card">
        <div class="kpi-icon" style="background:linear-gradient(135deg,#f59e0b,#d97706)">👥</div>
        <div class="kpi-body">
            <div class="kpi-value">{stats.players}</div>
            <div class="kpi-label">Jami o'yinchilar</div>
            <div class="kpi-sub">Barcha sessiyalar</div>
        </div>
    </div>
    <div class="kpi-card">
        <div class="kpi-icon" style="background:linear-gradient(135deg,#8b5cf6,#7c3aed)">📊</div>
        <div class="kpi-body">
            <div class="kpi-value">{stats.avgPlayers}</div>
            <div class="kpi-label">O'rtacha ishtirok</div>
            <div class="kpi-sub">Har sessiyada</div>
        </div>
    </div>
</div>
{/if}

<!-- ═══════════════════════════════════════════════════════ TOOLBAR -->
<div class="toolbar">
    <div class="filter-pills">
        <button class="pill" class:active={filter==='all'} on:click={() => filter='all'}>Barchasi</button>
        <button class="pill" class:active={filter==='today'} on:click={() => filter='today'}>Bugun</button>
        <button class="pill" class:active={filter==='week'} on:click={() => filter='week'}>Bu hafta</button>
    </div>
    <div class="toolbar-right">
        <div class="search-wrap">
            <span class="search-icon">🔍</span>
            <input type="text" placeholder="Quiz qidirish..." bind:value={search} class="search-input" />
            {#if search}<button class="clear-btn" on:click={() => search=''}>✕</button>{/if}
        </div>
        <select class="sort-select" bind:value={sortBy}>
            <option value="date">Sana bo'yicha</option>
            <option value="players">O'yinchilar bo'yicha</option>
        </select>
    </div>
</div>

<!-- ═══════════════════════════════════════════════════════ BAR CHART -->
{#if !loading && history.length > 0}
<div class="chart-card">
    <div class="chart-header">
        <h3>Haftalik faollik</h3>
        <span class="chart-subtitle">Oxirgi 7 kun</span>
    </div>
    <div class="chart">
        {#each chartData as day, i}
            <div class="bar-wrap">
                <div class="bar-label">{day.count > 0 ? day.count : ''}</div>
                <div class="bar-track">
                    <div class="bar" style="height:{day.count/maxCount*100}%;transition:height 0.5s ease calc({i}*0.07s)"></div>
                </div>
                <div class="bar-day">{day.label}</div>
            </div>
        {/each}
    </div>
</div>
{/if}

<!-- ═══════════════════════════════════════════════════════ SESSION TABLE -->
{#if loading}
    <div class="table-card">
        {#each Array(5) as _}<div class="skeleton-row"></div>{/each}
    </div>
{:else if sorted.length === 0}
    <div class="empty-state">
        <div class="empty-emoji">📈</div>
        <p class="empty-title">{#if search || filter!=='all'}Hech narsa topilmadi{:else}Hali o'yin o'tkazilmagan{/if}</p>
        <p class="empty-sub">{#if search || filter!=='all'}Boshqa qidiruv mezonini sinab ko'ring{:else}Dashboard'dan quiz boshlang!{/if}</p>
    </div>
{:else}
    <div class="table-card">
        <div class="table-head">
            <span>Quiz nomi</span>
            <span>Rejim</span>
            <span class="col-center">O'yinchilar</span>
            <span class="col-center">Davomiyligi</span>
            <span>Sana</span>
            <span class="col-center">Holat</span>
            <span class="col-center">Natija</span>
        </div>
        {#each sorted as item (item.id)}
            <div class="table-row" on:click={() => openRoom(item)} role="button" tabindex="0"
                on:keydown={e => e.key==='Enter' && openRoom(item)}>
                <div class="quiz-col">
                    <div class="quiz-pin">PIN: {item.pin}</div>
                    <div class="quiz-title-text">{item.quiz_title}</div>
                </div>
                <span class="mode-chip">{modeLabel(item.game_mode)}</span>
                <span class="col-center players-val"><span class="player-icon">👤</span>{item.player_count}</span>
                <span class="col-center dur-val">{duration(item.created_at, item.ended_at)}</span>
                <span class="date-val">{formatDate(item.created_at)}</span>
                <span class="col-center">
                    <span class="status-badge status-{item.status}">{statusLabel(item.status)}</span>
                </span>
                <span class="col-center">
                    <button class="view-btn" on:click|stopPropagation={() => openRoom(item)}>
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                        Ko'rish
                    </button>
                </span>
            </div>
        {/each}
    </div>
{/if}

<!-- ═══════════════════════════════════════════════════════ DETAIL MODAL -->
{#if showModal && selectedRoom}
    {@const room = selectedRoom}
    <div class="modal-backdrop" on:click|self={closeModal} role="dialog" aria-modal="true">
        <div class="modal-box">

            <!-- Modal Header -->
            <div class="modal-hd">
                <div class="modal-hd-left">
                    <div class="modal-icon">🎮</div>
                    <div>
                        <h2 class="modal-title">{room.quiz_title}</h2>
                        <p class="modal-sub">{formatDateFull(room.created_at)}</p>
                    </div>
                </div>
                <button class="modal-close" on:click={closeModal}>✕</button>
            </div>

            <!-- Session meta chips -->
            <div class="meta-chips">
                <div class="meta-chip">
                    <span class="mc-icon">🔢</span>
                    <span class="mc-label">PIN</span>
                    <span class="mc-val">{room.pin}</span>
                </div>
                <div class="meta-chip">
                    <span class="mc-icon">⚡</span>
                    <span class="mc-label">Rejim</span>
                    <span class="mc-val">{modeLabel(room.game_mode)}</span>
                </div>
                <div class="meta-chip">
                    <span class="mc-icon">👥</span>
                    <span class="mc-label">Ishtirokchilar</span>
                    <span class="mc-val">{room.player_count} ta</span>
                </div>
                <div class="meta-chip">
                    <span class="mc-icon">⏱️</span>
                    <span class="mc-label">Davomiyligi</span>
                    <span class="mc-val">{duration(room.created_at, room.ended_at)}</span>
                </div>
                <div class="meta-chip">
                    <span class="mc-icon"
                        class:mc-completed={room.status==='completed'}
                        class:mc-progress={room.status==='in_progress'}
                    ></span>
                    <span class="mc-label">Holat</span>
                    <span class="mc-val"><span class="status-badge status-{room.status}">{statusLabel(room.status)}</span></span>
                </div>
            </div>

            <!-- Leaderboard -->
            <div class="lb-header">
                <h3>Natijalar jadvali</h3>
                <button class="btn ghost sm" on:click={downloadXLSX}>
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                    Excel
                </button>
            </div>

            {#if resultsLoading}
                <div class="lb-loading">
                    {#each Array(4) as _}<div class="skeleton-row sm"></div>{/each}
                </div>
            {:else if playerResults.length === 0}
                <div class="lb-empty">
                    <span>📭</span>
                    <p>Bu sessiya uchun natijalar mavjud emas</p>
                </div>
            {:else}
                <!-- Podium top-3 -->
                {#if playerResults.length >= 3}
                <div class="podium">
                    <!-- 2nd place -->
                    <div class="podium-col second">
                        <div class="pod-avatar">{playerResults[1].avatar}</div>
                        <div class="pod-name">{playerResults[1].nickname}</div>
                        <div class="pod-score">{playerResults[1].score.toLocaleString()}</div>
                        <div class="pod-bar b2">🥈</div>
                    </div>
                    <!-- 1st place -->
                    <div class="podium-col first">
                        <div class="pod-crown">👑</div>
                        <div class="pod-avatar lg">{playerResults[0].avatar}</div>
                        <div class="pod-name">{playerResults[0].nickname}</div>
                        <div class="pod-score">{playerResults[0].score.toLocaleString()}</div>
                        <div class="pod-bar b1">🥇</div>
                    </div>
                    <!-- 3rd place -->
                    <div class="podium-col third">
                        <div class="pod-avatar">{playerResults[2].avatar}</div>
                        <div class="pod-name">{playerResults[2].nickname}</div>
                        <div class="pod-score">{playerResults[2].score.toLocaleString()}</div>
                        <div class="pod-bar b3">🥉</div>
                    </div>
                </div>
                {/if}

                <!-- Full table -->
                <div class="lb-table">
                    <div class="lb-thead">
                        <span>#</span>
                        <span>O'yinchi</span>
                        <span class="tc">Ball</span>
                        <span class="tc">To'g'ri</span>
                        <span class="tc">Aniqlik</span>
                        <span class="tc">Vaqt</span>
                    </div>
                    {#each playerResults as p}
                        <div class="lb-row" class:lb-gold={p.rank===1} class:lb-silver={p.rank===2} class:lb-bronze={p.rank===3}>
                            <span class="lb-rank">{medalEmoji(p.rank)}</span>
                            <div class="lb-player">
                                <span class="lb-av">{p.avatar}</span>
                                <span class="lb-nick">{p.nickname}</span>
                                {#if p.streak >= 3}
                                    <span class="streak-badge">🔥{p.streak}</span>
                                {/if}
                            </div>
                            <div class="lb-score-cell tc">
                                <div class="score-val">{p.score.toLocaleString()}</div>
                                <div class="score-bar">
                                    <div class="score-fill" style="width:{scorePercent(p.score)}%"></div>
                                </div>
                            </div>
                            <span class="tc correct-val">{p.correct_count}/{p.total_answered}</span>
                            <span class="tc acc-val" class:acc-high={accuracy(p)>=80} class:acc-mid={accuracy(p)>=50&&accuracy(p)<80}>{accuracy(p)}%</span>
                            <span class="tc time-val">{avgTimeSec(p.avg_time_ms)}</span>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    </div>
{/if}

<style>
    /* ── Page Header ─────────────────────────────────────────────────────────── */
    .page-header { display:flex; align-items:flex-start; justify-content:space-between; margin-bottom:28px; gap:16px; }
    h1 { font-size:1.85rem; font-weight:800; color:var(--text); margin:0 0 4px; }
    .subtitle { font-size:0.875rem; color:var(--text3); margin:0; }

    /* ── KPI Cards ───────────────────────────────────────────────────────────── */
    .kpi-grid { display:grid; grid-template-columns:repeat(4,1fr); gap:16px; margin-bottom:24px; }
    .kpi-card {
        background:var(--white); border-radius:var(--radius-lg); padding:20px;
        box-shadow:var(--shadow-sm); display:flex; align-items:center; gap:14px;
        transition:var(--transition); border:1.5px solid transparent;
    }
    .kpi-card:hover { transform:translateY(-3px); box-shadow:var(--shadow); border-color:var(--primary-light); }
    .kpi-icon { width:52px; height:52px; border-radius:14px; display:flex; align-items:center; justify-content:center; font-size:1.4rem; flex-shrink:0; box-shadow:0 4px 12px rgba(0,0,0,0.15); }
    .kpi-body { flex:1; min-width:0; }
    .kpi-value { font-size:2rem; font-weight:800; color:var(--text); line-height:1; }
    .kpi-label { font-size:0.8rem; font-weight:600; color:var(--text2); margin:4px 0 3px; }
    .kpi-sub { font-size:0.75rem; color:var(--text3); }
    .trend-up { color:var(--success); font-weight:700; }

    /* ── Toolbar ─────────────────────────────────────────────────────────────── */
    .toolbar { display:flex; align-items:center; justify-content:space-between; gap:16px; margin-bottom:20px; flex-wrap:wrap; }
    .filter-pills { display:flex; gap:6px; background:var(--bg2); padding:4px; border-radius:10px; }
    .pill { padding:6px 16px; border:none; border-radius:7px; font-size:0.85rem; font-weight:600; cursor:pointer; background:transparent; color:var(--text2); transition:all 0.2s; }
    .pill.active { background:var(--white); color:var(--primary); box-shadow:var(--shadow-sm); }
    .pill:hover:not(.active) { background:rgba(255,255,255,0.6); color:var(--text); }
    .toolbar-right { display:flex; gap:10px; align-items:center; }
    .search-wrap { position:relative; display:flex; align-items:center; }
    .search-icon { position:absolute; left:10px; font-size:0.85rem; pointer-events:none; z-index:1; }
    .search-input { padding:8px 32px; border:1.5px solid var(--border); border-radius:var(--radius); font-size:0.875rem; outline:none; transition:border-color 0.2s; background:var(--white); width:220px; }
    .search-input:focus { border-color:var(--primary); }
    .clear-btn { position:absolute; right:8px; background:none; border:none; cursor:pointer; color:var(--text3); font-size:0.8rem; padding:2px; line-height:1; }
    .sort-select { padding:8px 12px; border:1.5px solid var(--border); border-radius:var(--radius); font-size:0.875rem; outline:none; background:var(--white); color:var(--text); cursor:pointer; }
    .sort-select:focus { border-color:var(--primary); }

    /* ── Chart ───────────────────────────────────────────────────────────────── */
    .chart-card { background:var(--white); border-radius:var(--radius-lg); padding:24px; box-shadow:var(--shadow-sm); margin-bottom:20px; }
    .chart-header { display:flex; align-items:center; justify-content:space-between; margin-bottom:20px; }
    .chart-header h3 { font-size:1rem; font-weight:700; color:var(--text); margin:0; }
    .chart-subtitle { font-size:0.8rem; color:var(--text3); }
    .chart { display:flex; align-items:flex-end; gap:12px; height:120px; }
    .bar-wrap { flex:1; display:flex; flex-direction:column; align-items:center; gap:6px; height:100%; }
    .bar-label { font-size:0.72rem; font-weight:700; color:var(--text2); min-height:16px; }
    .bar-track { flex:1; width:100%; display:flex; align-items:flex-end; background:var(--bg2); border-radius:6px; overflow:hidden; min-height:4px; }
    .bar { width:100%; min-height:4px; background:linear-gradient(180deg,var(--primary) 0%,var(--accent) 100%); border-radius:6px; }
    .bar-day { font-size:0.72rem; color:var(--text3); font-weight:500; }

    /* ── Table ───────────────────────────────────────────────────────────────── */
    .table-card { background:var(--white); border-radius:var(--radius-lg); box-shadow:var(--shadow-sm); overflow:hidden; margin-bottom:8px; }
    .table-head {
        display:grid; grid-template-columns:2fr 1.1fr 90px 100px 1.1fr 120px 100px;
        gap:12px; padding:12px 20px;
        font-size:0.7rem; font-weight:700; color:var(--text3); text-transform:uppercase; letter-spacing:0.07em;
        background:var(--bg2); border-bottom:1.5px solid var(--border);
    }
    .table-row {
        display:grid; grid-template-columns:2fr 1.1fr 90px 100px 1.1fr 120px 100px;
        gap:12px; align-items:center; padding:13px 20px;
        border-bottom:1px solid var(--bg2); transition:background 0.15s; cursor:pointer;
    }
    .table-row:last-child { border-bottom:none; }
    .table-row:hover { background:rgba(99,102,241,0.04); }
    .col-center { text-align:center; justify-self:center; }
    .quiz-col { display:flex; flex-direction:column; gap:2px; min-width:0; }
    .quiz-pin { font-size:0.7rem; color:var(--text3); font-family:monospace; }
    .quiz-title-text { font-size:0.88rem; font-weight:600; color:var(--text); white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
    .mode-chip { font-size:0.8rem; color:var(--text2); }
    .players-val { font-size:0.85rem; font-weight:600; color:var(--text2); display:flex; align-items:center; gap:4px; }
    .player-icon { font-size:0.75rem; }
    .dur-val { font-size:0.85rem; color:var(--text2); font-variant-numeric:tabular-nums; }
    .date-val { font-size:0.8rem; color:var(--text3); }
    .status-badge { font-size:0.72rem; font-weight:700; padding:4px 10px; border-radius:99px; display:inline-block; white-space:nowrap; }
    .status-completed { background:#dcfce7; color:#16a34a; }
    .status-in_progress { background:#fee2e2; color:var(--danger); }
    .status-waiting { background:#fef3c7; color:#d97706; }
    .status-abandoned { background:var(--bg2); color:var(--text3); }
    .view-btn {
        display:inline-flex; align-items:center; gap:5px; padding:5px 11px;
        border:1.5px solid var(--border); border-radius:8px; background:transparent;
        font-size:0.75rem; font-weight:600; color:var(--text2); cursor:pointer;
        transition:all 0.2s;
    }
    .view-btn:hover { border-color:var(--primary); color:var(--primary); background:rgba(99,102,241,0.06); }

    /* ── Skeleton ────────────────────────────────────────────────────────────── */
    .skeleton-row { height:64px; margin:6px 16px; background:linear-gradient(90deg,var(--bg2) 25%,var(--border) 50%,var(--bg2) 75%); background-size:200% 100%; animation:shimmer 1.4s infinite; border-radius:8px; }
    .skeleton-row.sm { height:48px; }
    @keyframes shimmer { 0%{background-position:200% 0} 100%{background-position:-200% 0} }

    /* ── Empty ───────────────────────────────────────────────────────────────── */
    .empty-state { background:var(--white); border-radius:var(--radius-lg); padding:64px 32px; text-align:center; box-shadow:var(--shadow-sm); }
    .empty-emoji { font-size:3.5rem; margin-bottom:16px; }
    .empty-title { font-size:1.1rem; font-weight:700; color:var(--text); margin:0 0 8px; }
    .empty-sub { font-size:0.875rem; color:var(--text3); margin:0; }

    /* ── Buttons ─────────────────────────────────────────────────────────────── */
    .btn { padding:9px 18px; border:none; border-radius:var(--radius); font-size:0.875rem; font-weight:600; cursor:pointer; display:inline-flex; align-items:center; gap:7px; transition:var(--transition); text-decoration:none; }
    .btn.primary { background:linear-gradient(135deg,var(--primary),var(--accent)); color:white; box-shadow:0 3px 10px rgba(99,102,241,0.3); }
    .btn.primary:hover:not(:disabled) { transform:translateY(-1px); box-shadow:0 6px 16px rgba(99,102,241,0.4); }
    .btn.ghost { background:transparent; border:1.5px solid var(--border); color:var(--text2); }
    .btn.ghost:hover { border-color:var(--primary); color:var(--primary); background:rgba(99,102,241,0.05); }
    .btn.sm { padding:6px 12px; font-size:0.8rem; }
    .btn:disabled { opacity:0.5; cursor:not-allowed; }

    /* ── Modal Backdrop ──────────────────────────────────────────────────────── */
    .modal-backdrop {
        position:fixed; inset:0; background:rgba(0,0,0,0.55); backdrop-filter:blur(4px);
        z-index:1000; display:flex; align-items:center; justify-content:center;
        padding:20px; animation:fadein 0.2s ease;
    }
    @keyframes fadein { from{opacity:0} to{opacity:1} }
    .modal-box {
        background:var(--white); border-radius:20px; width:100%; max-width:760px;
        max-height:92vh; overflow-y:auto; box-shadow:0 25px 80px rgba(0,0,0,0.3);
        animation:slideup 0.25s ease;
    }
    @keyframes slideup { from{opacity:0;transform:translateY(24px)} to{opacity:1;transform:translateY(0)} }

    /* ── Modal Header ────────────────────────────────────────────────────────── */
    .modal-hd {
        display:flex; align-items:center; justify-content:space-between;
        padding:24px 28px 20px; border-bottom:1.5px solid var(--bg2);
        background:linear-gradient(135deg, rgba(99,102,241,0.06) 0%, rgba(139,92,246,0.04) 100%);
        border-radius:20px 20px 0 0;
    }
    .modal-hd-left { display:flex; align-items:center; gap:14px; }
    .modal-icon { width:48px; height:48px; border-radius:14px; background:linear-gradient(135deg,#6366f1,#8b5cf6); display:flex; align-items:center; justify-content:center; font-size:1.5rem; flex-shrink:0; box-shadow:0 4px 14px rgba(99,102,241,0.35); }
    .modal-title { font-size:1.2rem; font-weight:800; color:var(--text); margin:0 0 3px; }
    .modal-sub { font-size:0.8rem; color:var(--text3); margin:0; }
    .modal-close { background:var(--bg2); border:none; border-radius:50%; width:34px; height:34px; display:flex; align-items:center; justify-content:center; cursor:pointer; font-size:0.85rem; color:var(--text2); transition:all 0.2s; flex-shrink:0; }
    .modal-close:hover { background:var(--danger); color:white; }

    /* ── Meta Chips ──────────────────────────────────────────────────────────── */
    .meta-chips { display:flex; gap:10px; padding:18px 28px; flex-wrap:wrap; border-bottom:1px solid var(--bg2); }
    .meta-chip { display:flex; align-items:center; gap:7px; background:var(--bg2); border-radius:10px; padding:8px 14px; }
    .mc-icon { font-size:1rem; }
    .mc-label { font-size:0.72rem; color:var(--text3); font-weight:600; text-transform:uppercase; letter-spacing:0.05em; }
    .mc-val { font-size:0.85rem; font-weight:700; color:var(--text); }
    .mc-completed { color: #16a34a; }
    .mc-progress { color: var(--danger); }

    /* ── Leaderboard Header ──────────────────────────────────────────────────── */
    .lb-header { display:flex; align-items:center; justify-content:space-between; padding:18px 28px 12px; }
    .lb-header h3 { font-size:1rem; font-weight:700; color:var(--text); margin:0; }
    .lb-loading { padding:0 28px 20px; }
    .lb-empty { text-align:center; padding:40px 28px; color:var(--text3); }
    .lb-empty span { font-size:2rem; display:block; margin-bottom:8px; }
    .lb-empty p { font-size:0.875rem; margin:0; }

    /* ── Podium ──────────────────────────────────────────────────────────────── */
    .podium {
        display:flex; align-items:flex-end; justify-content:center;
        gap:8px; padding:0 28px 20px; margin-bottom:4px;
    }
    .podium-col { display:flex; flex-direction:column; align-items:center; gap:4px; flex:1; max-width:160px; }
    .pod-crown { font-size:1.4rem; margin-bottom:2px; }
    .pod-avatar { font-size:2rem; line-height:1; }
    .pod-avatar.lg { font-size:2.5rem; }
    .pod-name { font-size:0.78rem; font-weight:700; color:var(--text); text-align:center; max-width:110px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
    .pod-score { font-size:0.72rem; font-weight:600; color:var(--text2); }
    .pod-bar { width:100%; border-radius:8px 8px 0 0; font-size:1.2rem; display:flex; align-items:center; justify-content:center; padding:8px 0; margin-top:4px; }
    .b1 { background:linear-gradient(135deg,#fbbf24,#f59e0b); min-height:72px; }
    .b2 { background:linear-gradient(135deg,#94a3b8,#cbd5e1); min-height:56px; }
    .b3 { background:linear-gradient(135deg,#d97706,#b45309); min-height:44px; }

    /* ── Leaderboard Table ───────────────────────────────────────────────────── */
    .lb-table { padding:0 28px 28px; display:flex; flex-direction:column; gap:6px; }
    .lb-thead {
        display:grid; grid-template-columns:44px 1fr 100px 80px 80px 70px;
        gap:8px; padding:8px 14px;
        font-size:0.7rem; font-weight:700; color:var(--text3); text-transform:uppercase; letter-spacing:0.06em;
        background:var(--bg2); border-radius:10px;
    }
    .lb-row {
        display:grid; grid-template-columns:44px 1fr 100px 80px 80px 70px;
        gap:8px; align-items:center; padding:10px 14px;
        background:var(--bg2); border-radius:12px; transition:transform 0.15s, box-shadow 0.15s;
        border:1.5px solid transparent;
    }
    .lb-row:hover { transform:translateX(3px); box-shadow:var(--shadow-sm); border-color:var(--border); }
    .lb-gold { background:linear-gradient(135deg, rgba(251,191,36,0.15), rgba(245,158,11,0.08)); border-color:rgba(251,191,36,0.4); }
    .lb-silver { background:linear-gradient(135deg, rgba(148,163,184,0.15), rgba(203,213,225,0.08)); border-color:rgba(148,163,184,0.3); }
    .lb-bronze { background:linear-gradient(135deg, rgba(217,119,6,0.12), rgba(180,83,9,0.06)); border-color:rgba(217,119,6,0.25); }
    .lb-rank { font-size:1.1rem; text-align:center; }
    .lb-player { display:flex; align-items:center; gap:8px; min-width:0; }
    .lb-av { font-size:1.3rem; flex-shrink:0; }
    .lb-nick { font-size:0.875rem; font-weight:600; color:var(--text); overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
    .streak-badge { font-size:0.7rem; font-weight:700; background:#fee2e2; color:#dc2626; padding:2px 6px; border-radius:6px; flex-shrink:0; }
    .tc { text-align:center; justify-self:center; }
    .lb-score-cell { display:flex; flex-direction:column; gap:4px; align-items:center; width:100%; }
    .score-val { font-size:0.88rem; font-weight:800; color:var(--primary); }
    .score-bar { width:100%; height:4px; background:var(--border); border-radius:99px; overflow:hidden; }
    .score-fill { height:100%; background:linear-gradient(90deg, var(--primary), var(--accent)); border-radius:99px; transition:width 0.6s ease; }
    .correct-val { font-size:0.82rem; font-weight:600; color:var(--text2); }
    .acc-val { font-size:0.82rem; font-weight:700; color:var(--text3); }
    .acc-high { color:#16a34a; }
    .acc-mid { color:#d97706; }
    .time-val { font-size:0.78rem; color:var(--text3); font-variant-numeric:tabular-nums; }

    /* ── Responsive ──────────────────────────────────────────────────────────── */
    @media(max-width:1024px) {
        .kpi-grid { grid-template-columns:repeat(2,1fr); }
        .table-head,.table-row { grid-template-columns:2fr 1fr 80px 1fr 100px; }
        .table-head span:nth-child(4),.dur-val { display:none; }
    }
    @media(max-width:640px) {
        .kpi-grid { grid-template-columns:repeat(2,1fr); }
        .toolbar { flex-direction:column; align-items:stretch; }
        .toolbar-right { flex-direction:column; }
        .search-input { width:100%; }
        .table-head,.table-row { grid-template-columns:1fr 80px 100px; }
        .table-head span:nth-child(2),.table-head span:nth-child(3),.mode-chip,.players-val { display:none; }
        .lb-thead,.lb-row { grid-template-columns:36px 1fr 80px 60px; }
        .lb-thead span:nth-child(5),.lb-thead span:nth-child(6),.acc-val,.time-val { display:none; }
        .podium { padding:0 16px 16px; }
        .modal-hd,.meta-chips,.lb-header,.lb-table { padding-left:16px; padding-right:16px; }
    }
</style>
