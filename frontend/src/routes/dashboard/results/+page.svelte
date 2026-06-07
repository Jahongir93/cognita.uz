<script lang="ts">
  import { onMount } from 'svelte';

  // ── Types ─────────────────────────────────────────────────────────────────
  interface StudentResult {
    id: string;
    type: 'game' | 'exam' | 'olympiad';
    title: string;
    score: number;
    maxScore: number;
    percent: number;
    timeTaken: number;
    date: string;
    code?: string;
    pin?: string;
  }

  // ── State ─────────────────────────────────────────────────────────────────
  const filterTabs = [
    { key: 'all'      as const, label: 'Barchasi',    icon: '🗂' },
    { key: 'game'     as const, label: "O'yinlar",    icon: '🎮' },
    { key: 'exam'     as const, label: 'Imtihonlar',  icon: '📋' },
    { key: 'olympiad' as const, label: 'Olimpiadalar', icon: '🏆' },
  ];

  let allResults: StudentResult[] = [];
  let loaded = false;
  let activeTab: 'all' | 'game' | 'exam' | 'olympiad' = 'all';
  let clearConfirm = false;

  // ── Load ──────────────────────────────────────────────────────────────────
  onMount(() => load());

  function load() {
    try {
      const raw = localStorage.getItem('gogame_results');
      allResults = raw ? (JSON.parse(raw) as StudentResult[]) : [];
    } catch {
      allResults = [];
    }
    loaded = true;
  }

  // ── Filtered + sorted ─────────────────────────────────────────────────────
  $: filtered = (activeTab === 'all'
    ? allResults
    : allResults.filter(r => r.type === activeTab)
  ).sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());

  // ── Counts per tab ────────────────────────────────────────────────────────
  $: counts = {
    all:       allResults.length,
    game:      allResults.filter(r => r.type === 'game').length,
    exam:      allResults.filter(r => r.type === 'exam').length,
    olympiad:  allResults.filter(r => r.type === 'olympiad').length,
  };

  // ── Summary stats ─────────────────────────────────────────────────────────
  $: totalPlayed   = allResults.length;
  $: avgPercent    = allResults.length
    ? Math.round(allResults.reduce((s, r) => s + r.percent, 0) / allResults.length)
    : 0;
  $: bestScore     = allResults.length ? Math.max(...allResults.map(r => r.percent)) : 0;
  $: totalTimeSec  = allResults.reduce((s, r) => s + r.timeTaken, 0);

  function fmtTotalTime(sec: number) {
    if (sec < 60) return `${sec}s`;
    const m = Math.floor(sec / 60);
    if (m < 60) return `${m}m`;
    const h = Math.floor(m / 60);
    const rem = m % 60;
    return rem > 0 ? `${h}s ${rem}m` : `${h}s`;
  }

  // ── Clear all ─────────────────────────────────────────────────────────────
  function clearAll() {
    try {
      localStorage.removeItem('gogame_results');
    } catch {}
    allResults = [];
    clearConfirm = false;
  }

  // ── Helpers ───────────────────────────────────────────────────────────────
  function fmtDate(d: string) {
    return new Date(d).toLocaleDateString('uz-UZ', {
      day: '2-digit', month: 'short', year: 'numeric'
    });
  }

  function fmtRelDate(d: string) {
    const now = new Date();
    const then = new Date(d);
    const diff = now.getTime() - then.getTime();
    const days = Math.floor(diff / 86400000);
    if (days === 0) {
      const hrs = Math.floor(diff / 3600000);
      if (hrs === 0) return 'Hozirgina';
      return `${hrs} soat oldin`;
    }
    if (days === 1) return 'Kecha';
    if (days < 7) return `${days} kun oldin`;
    return fmtDate(d);
  }

  function fmtTime(sec: number) {
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    return m > 0 ? `${m}m ${s}s` : `${s}s`;
  }

  const typeConfig = {
    game:      { icon: '🎮', label: "O'yin",     color: '#6366f1', bg: 'rgba(99,102,241,0.12)',  barGrad: 'linear-gradient(90deg,#6366f1,#8b5cf6)' },
    exam:      { icon: '📋', label: 'Imtihon',   color: '#f59e0b', bg: 'rgba(245,158,11,0.12)', barGrad: 'linear-gradient(90deg,#f59e0b,#f97316)' },
    olympiad:  { icon: '🏆', label: 'Olimpiada', color: '#eab308', bg: 'rgba(234,179,8,0.12)',  barGrad: 'linear-gradient(90deg,#eab308,#ca8a04)' },
  } as const;

  function scoreBarGrad(pct: number, type: 'game' | 'exam' | 'olympiad') {
    if (pct >= 80) return 'linear-gradient(90deg,#22c55e,#16a34a)';
    if (pct >= 60) return typeConfig[type].barGrad;
    if (pct >= 40) return 'linear-gradient(90deg,#f59e0b,#f97316)';
    return 'linear-gradient(90deg,#ef4444,#dc2626)';
  }

  function scoreColor(pct: number) {
    if (pct >= 80) return 'var(--success)';
    if (pct >= 60) return 'var(--primary)';
    if (pct >= 40) return '#f59e0b';
    return 'var(--danger)';
  }

  // ── Group results by date for timeline effect ─────────────────────────────
  function groupByDate(results: StudentResult[]) {
    const groups: { dateLabel: string; items: StudentResult[] }[] = [];
    const seen = new Map<string, number>();
    for (const r of results) {
      const label = fmtDate(r.date);
      if (seen.has(label)) {
        groups[seen.get(label)!].items.push(r);
      } else {
        seen.set(label, groups.length);
        groups.push({ dateLabel: label, items: [r] });
      }
    }
    return groups;
  }

  $: grouped = groupByDate(filtered);
</script>

<svelte:head><title>Natijalarim — Cognita.uz</title></svelte:head>

<!-- ══════════════════════════════════════════════════
     PAGE HEADER
══════════════════════════════════════════════════ -->
<div class="page-header">
  <div class="header-left">
    <div>
      <h1>
        Natijalarim
        {#if loaded && allResults.length > 0}
          <span class="count-badge">{allResults.length}</span>
        {/if}
      </h1>
      <p class="sub">Barcha o'yin, imtihon va olimpiada natijalaringiz</p>
    </div>
  </div>
  <a href="/dashboard/join" class="join-btn-link">
    <span>+</span> Yangi kirish
  </a>
</div>

<!-- ══════════════════════════════════════════════════
     SUMMARY STATS ROW
══════════════════════════════════════════════════ -->
<div class="summary-row">
  <div class="summary-card">
    <div class="summary-icon" style="background:rgba(99,102,241,0.12)">🎯</div>
    <div>
      <div class="summary-val">{loaded ? totalPlayed : '…'}</div>
      <div class="summary-lbl">Jami o'ynaldi</div>
    </div>
  </div>
  <div class="summary-card">
    <div class="summary-icon" style="background:rgba(34,197,94,0.12)">📊</div>
    <div>
      <div class="summary-val" style="color:var(--success)">{loaded ? avgPercent : '…'}%</div>
      <div class="summary-lbl">O'rtacha %</div>
    </div>
  </div>
  <div class="summary-card">
    <div class="summary-icon" style="background:rgba(234,179,8,0.12)">🏅</div>
    <div>
      <div class="summary-val" style="color:#eab308">{loaded ? bestScore : '…'}%</div>
      <div class="summary-lbl">Eng yuqori ball</div>
    </div>
  </div>
  <div class="summary-card">
    <div class="summary-icon" style="background:rgba(139,92,246,0.12)">⏱</div>
    <div>
      <div class="summary-val" style="color:var(--accent)">{loaded ? fmtTotalTime(totalTimeSec) : '…'}</div>
      <div class="summary-lbl">Umumiy vaqt</div>
    </div>
  </div>
</div>

<!-- ══════════════════════════════════════════════════
     FILTER TABS
══════════════════════════════════════════════════ -->
<nav class="tabs" aria-label="Filtr">
  {#each filterTabs as tab}
    <button
      class="tab-btn"
      class:active={activeTab === tab.key}
      on:click={() => activeTab = tab.key}
    >
      <span class="tab-icon">{tab.icon}</span>
      <span>{tab.label}</span>
      {#if counts[tab.key] > 0}
        <span class="tab-count" class:active-count={activeTab === tab.key}>
          {counts[tab.key]}
        </span>
      {/if}
    </button>
  {/each}
</nav>

<!-- ══════════════════════════════════════════════════
     TIMELINE / RESULTS LIST
══════════════════════════════════════════════════ -->
{#if !loaded}
  <div class="skeleton-wrap">
    {#each Array(5) as _}
      <div class="skeleton sk-item"></div>
    {/each}
  </div>
{:else if filtered.length === 0}
  <div class="empty-state">
    {#if activeTab === 'all'}
      <div class="empty-emoji">🏅</div>
      <h3>Hali natijalar yo'q</h3>
      <p>O'yin, imtihon yoki olimpiadaga qo'shiling va natijalaringiz shu yerda ko'rinadi.</p>
      <a href="/dashboard/join" class="empty-action-btn">Hozir boshlash →</a>
    {:else if activeTab === 'game'}
      <div class="empty-emoji">🎮</div>
      <h3>Hali o'yin o'ynamagansiz</h3>
      <p>Viktorina o'yiniga kirish uchun PIN kod kiriting.</p>
      <a href="/dashboard/join" class="empty-action-btn">O'yinga kirish →</a>
    {:else if activeTab === 'exam'}
      <div class="empty-emoji">📋</div>
      <h3>Hali imtihon topshirmagansiz</h3>
      <p>Imtihon kodi kiritib, mustaqil test topshiring.</p>
      <a href="/dashboard/join" class="empty-action-btn">Imtihon yechish →</a>
    {:else}
      <div class="empty-emoji">🏆</div>
      <h3>Hali olimpiadada qatnashmadingiz</h3>
      <p>Olimpiada kodini kiritib, musobaqada ishtirok eting.</p>
      <a href="/dashboard/join" class="empty-action-btn">Olimpiadaga kirish →</a>
    {/if}
  </div>
{:else}
  <div class="timeline">
    {#each grouped as group, gi}
      <!-- Date separator -->
      <div class="date-separator" style="animation-delay:{gi * 30}ms">
        <span class="date-line"></span>
        <span class="date-label">{group.dateLabel}</span>
        <span class="date-line"></span>
      </div>

      {#each group.items as result, ri}
        {@const cfg = typeConfig[result.type]}
        {@const grad = scoreBarGrad(result.percent, result.type)}
        {@const sc = scoreColor(result.percent)}
        <div class="result-row" style="animation-delay:{(gi * 5 + ri) * 35}ms">
          <!-- Timeline line + dot -->
          <div class="timeline-track" aria-hidden="true">
            <div class="tl-dot" style="background:{cfg.color};box-shadow:0 0 0 3px {cfg.bg}"></div>
            {#if !(gi === grouped.length - 1 && ri === group.items.length - 1)}
              <div class="tl-line"></div>
            {/if}
          </div>

          <!-- Card -->
          <div class="result-card">
            <!-- Type icon -->
            <div class="type-icon-wrap" style="background:{cfg.bg}">
              <span class="type-icon">{cfg.icon}</span>
            </div>

            <!-- Info -->
            <div class="result-info">
              <div class="result-header-row">
                <span class="result-title">{result.title}</span>
                <span
                  class="type-pill"
                  style="color:{cfg.color};background:{cfg.bg}"
                >{cfg.label}</span>
              </div>
              <div class="result-meta">
                {#if result.pin}
                  <span class="meta-chip">📌 PIN: {result.pin}</span>
                {/if}
                {#if result.code}
                  <span class="meta-chip">🔑 {result.code}</span>
                {/if}
                <span class="meta-chip">⏱ {fmtTime(result.timeTaken)}</span>
                <span class="meta-chip meta-date" title={fmtDate(result.date)}>
                  🕐 {fmtRelDate(result.date)}
                </span>
              </div>
              <!-- Progress bar -->
              <div class="progress-row">
                <div class="progress-track">
                  <div
                    class="progress-fill"
                    style="width:{result.percent}%;background:{grad}"
                  ></div>
                </div>
                <span class="progress-label">{result.score}/{result.maxScore}</span>
              </div>
            </div>

            <!-- Percent -->
            <div class="percent-col">
              <div class="percent-val" style="color:{sc}">{result.percent}%</div>
              <div class="percent-sub">ball</div>
            </div>
          </div>
        </div>
      {/each}
    {/each}
  </div>
{/if}

<!-- ══════════════════════════════════════════════════
     CLEAR ALL BUTTON
══════════════════════════════════════════════════ -->
{#if loaded && allResults.length > 0}
  <div class="clear-section">
    {#if clearConfirm}
      <div class="confirm-modal">
        <span class="confirm-icon">⚠️</span>
        <div class="confirm-text">
          <strong>Barcha natijalar o'chiriladi!</strong>
          <span>Bu amalni qaytarib bo'lmaydi.</span>
        </div>
        <div class="confirm-actions">
          <button class="btn-danger" on:click={clearAll}>Ha, o'chirish</button>
          <button class="btn-cancel" on:click={() => clearConfirm = false}>Bekor qilish</button>
        </div>
      </div>
    {:else}
      <button class="clear-btn" on:click={() => clearConfirm = true}>
        🗑 Barcha natijalarni tozalash
      </button>
    {/if}
  </div>
{/if}

<style>
  /* ── Page header ── */
  .page-header {
    display: flex; align-items: flex-start;
    justify-content: space-between; gap: 16px;
    margin-bottom: 24px; flex-wrap: wrap;
  }
  h1 {
    font-size: 1.75rem; font-weight: 800;
    color: var(--text); margin: 0;
    display: flex; align-items: center; gap: 10px;
  }
  .count-badge {
    display: inline-flex; align-items: center; justify-content: center;
    min-width: 28px; height: 26px; padding: 0 8px;
    background: var(--primary); color: #fff;
    border-radius: 99px; font-size: 0.78rem; font-weight: 800;
    box-shadow: 0 2px 8px rgba(99,102,241,0.35);
  }
  .sub { font-size: 0.875rem; color: var(--text3); margin-top: 4px; }
  .join-btn-link {
    display: inline-flex; align-items: center; gap: 7px;
    padding: 10px 22px;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    color: #fff; border-radius: var(--radius);
    font-size: 0.9rem; font-weight: 700; text-decoration: none;
    box-shadow: 0 4px 14px rgba(99,102,241,0.35);
    transition: var(--transition); white-space: nowrap; flex-shrink: 0;
  }
  .join-btn-link:hover { transform: translateY(-2px); box-shadow: 0 6px 20px rgba(99,102,241,0.45); }

  /* ── Summary row ── */
  .summary-row {
    display: grid; grid-template-columns: repeat(4, 1fr);
    gap: 14px; margin-bottom: 24px;
  }
  .summary-card {
    background: var(--white); border: 1.5px solid var(--border);
    border-radius: var(--radius); padding: 14px 16px;
    display: flex; align-items: center; gap: 12px;
    box-shadow: var(--shadow-sm); transition: var(--transition);
    animation: fadeSlideUp 0.35s ease both;
  }
  .summary-card:hover { transform: translateY(-2px); box-shadow: var(--shadow); }
  .summary-icon {
    width: 42px; height: 42px; border-radius: var(--radius-sm);
    display: flex; align-items: center; justify-content: center;
    font-size: 1.2rem; flex-shrink: 0;
  }
  .summary-val { font-size: 1.35rem; font-weight: 900; color: var(--text); line-height: 1; }
  .summary-lbl { font-size: 0.72rem; color: var(--text3); font-weight: 600; margin-top: 3px; }

  /* ── Filter tabs ── */
  .tabs {
    display: flex; gap: 4px;
    background: var(--white); border: 1.5px solid var(--border);
    border-radius: var(--radius); padding: 5px;
    margin-bottom: 22px; width: fit-content;
    box-shadow: var(--shadow-sm); flex-wrap: wrap;
  }
  .tab-btn {
    display: flex; align-items: center; gap: 6px;
    padding: 8px 16px; border: none; border-radius: 9px;
    background: transparent; color: var(--text2);
    font-size: 0.85rem; font-weight: 600; cursor: pointer;
    transition: var(--transition); white-space: nowrap;
  }
  .tab-btn:hover:not(.active) { background: var(--bg); color: var(--text); }
  .tab-btn.active { background: var(--primary); color: #fff; box-shadow: 0 2px 8px rgba(99,102,241,0.3); }
  .tab-icon { font-size: 0.95rem; }
  .tab-count {
    display: inline-flex; align-items: center; justify-content: center;
    min-width: 20px; height: 18px; padding: 0 5px;
    background: var(--bg); color: var(--text3);
    border-radius: 99px; font-size: 0.65rem; font-weight: 800;
  }
  .tab-count.active-count { background: rgba(255,255,255,0.25); color: rgba(255,255,255,0.9); }

  /* ── Skeleton ── */
  .skeleton-wrap { display: flex; flex-direction: column; gap: 10px; }
  .skeleton {
    background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
    background-size: 200% 100%; animation: shimmer 1.4s infinite; border-radius: var(--radius-sm);
  }
  .sk-item { height: 64px; }
  @keyframes shimmer { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }

  /* ── Empty state ── */
  .empty-state {
    display: flex; flex-direction: column; align-items: center; gap: 12px;
    padding: 64px 24px; text-align: center;
    background: var(--white); border: 1.5px dashed var(--border);
    border-radius: var(--radius-lg);
  }
  .empty-emoji { font-size: 3.5rem; }
  .empty-state h3 { font-size: 1.1rem; font-weight: 800; color: var(--text); margin: 0; }
  .empty-state p { font-size: 0.875rem; color: var(--text3); max-width: 340px; margin: 0; line-height: 1.6; }
  .empty-action-btn {
    margin-top: 8px; padding: 11px 24px;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    color: #fff; border-radius: var(--radius);
    font-size: 0.9rem; font-weight: 700; text-decoration: none;
    box-shadow: 0 4px 14px rgba(99,102,241,0.35); transition: var(--transition);
  }
  .empty-action-btn:hover { transform: translateY(-2px); }

  /* ══════════════════════════════════════════════════
     TIMELINE
  ══════════════════════════════════════════════════ */
  .timeline { display: flex; flex-direction: column; gap: 0; margin-bottom: 28px; }

  /* Date separator */
  .date-separator {
    display: flex; align-items: center; gap: 12px;
    padding: 18px 0 10px; animation: fadeSlideUp 0.3s ease both;
  }
  .date-line { flex: 1; height: 1px; background: var(--border); }
  .date-label {
    font-size: 0.72rem; font-weight: 800; color: var(--text3);
    text-transform: uppercase; letter-spacing: 0.5px;
    white-space: nowrap; padding: 3px 10px;
    background: var(--bg); border-radius: 99px;
    border: 1px solid var(--border);
  }

  /* Result row with timeline track */
  .result-row {
    display: flex; align-items: stretch; gap: 0;
    animation: fadeSlideUp 0.3s ease both;
  }
  .timeline-track {
    display: flex; flex-direction: column; align-items: center;
    width: 32px; flex-shrink: 0; padding-top: 18px;
  }
  .tl-dot {
    width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0;
    transition: transform 0.2s;
  }
  .result-row:hover .tl-dot { transform: scale(1.4); }
  .tl-line { flex: 1; width: 2px; background: var(--border); min-height: 12px; margin-top: 4px; }

  /* Result card */
  .result-card {
    flex: 1; display: flex; align-items: center; gap: 14px;
    padding: 14px 18px; margin-bottom: 8px;
    background: var(--white); border: 1.5px solid var(--border);
    border-radius: var(--radius); box-shadow: var(--shadow-sm);
    transition: var(--transition);
  }
  .result-card:hover {
    border-color: var(--primary);
    box-shadow: var(--shadow);
    transform: translateX(4px);
  }

  /* Type icon */
  .type-icon-wrap {
    width: 42px; height: 42px; border-radius: var(--radius-sm);
    display: flex; align-items: center; justify-content: center;
    font-size: 1.2rem; flex-shrink: 0;
  }
  .type-icon { line-height: 1; }

  /* Info */
  .result-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 5px; }
  .result-header-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
  .result-title {
    font-size: 0.92rem; font-weight: 700; color: var(--text);
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
    flex: 1; min-width: 0;
  }
  .type-pill {
    font-size: 0.65rem; font-weight: 800;
    padding: 2px 8px; border-radius: 99px;
    white-space: nowrap; flex-shrink: 0;
    letter-spacing: 0.3px;
  }
  .result-meta { display: flex; gap: 5px; flex-wrap: wrap; }
  .meta-chip {
    font-size: 0.68rem; font-weight: 600;
    padding: 2px 7px; background: var(--bg);
    color: var(--text3); border-radius: 99px;
    border: 1px solid var(--border); white-space: nowrap;
  }
  .meta-date { font-style: italic; }

  /* Progress bar */
  .progress-row { display: flex; align-items: center; gap: 8px; }
  .progress-track { flex: 1; height: 5px; background: var(--bg); border-radius: 99px; overflow: hidden; }
  .progress-fill { height: 100%; border-radius: 99px; transition: width 0.7s ease; }
  .progress-label { font-size: 0.68rem; font-weight: 700; color: var(--text2); white-space: nowrap; min-width: 36px; text-align: right; }

  /* Percent col */
  .percent-col { display: flex; flex-direction: column; align-items: flex-end; flex-shrink: 0; min-width: 52px; }
  .percent-val { font-size: 1.3rem; font-weight: 900; letter-spacing: -0.5px; line-height: 1; }
  .percent-sub { font-size: 0.65rem; color: var(--text3); font-weight: 600; margin-top: 2px; }

  /* ══════════════════════════════════════════════════
     CLEAR SECTION
  ══════════════════════════════════════════════════ */
  .clear-section { display: flex; justify-content: center; padding: 8px 0 16px; }
  .clear-btn {
    padding: 9px 22px; background: transparent;
    color: var(--danger); border: 1.5px solid rgba(239,68,68,0.4);
    border-radius: var(--radius-sm); font-size: 0.82rem; font-weight: 700;
    cursor: pointer; transition: var(--transition); opacity: 0.7;
  }
  .clear-btn:hover { background: rgba(239,68,68,0.07); opacity: 1; border-color: var(--danger); }

  .confirm-modal {
    display: flex; align-items: center; gap: 16px;
    padding: 16px 22px; background: #fef2f2;
    border: 1.5px solid #fecaca; border-radius: var(--radius);
    flex-wrap: wrap; justify-content: center;
    animation: fadeSlideUp 0.22s ease both;
    max-width: 520px;
  }
  .confirm-icon { font-size: 1.5rem; flex-shrink: 0; }
  .confirm-text { display: flex; flex-direction: column; gap: 2px; flex: 1; min-width: 160px; }
  .confirm-text strong { font-size: 0.875rem; color: #b91c1c; }
  .confirm-text span { font-size: 0.78rem; color: #dc2626; opacity: 0.85; }
  .confirm-actions { display: flex; gap: 8px; flex-shrink: 0; }
  .btn-danger {
    padding: 8px 18px; background: var(--danger); color: #fff;
    border: none; border-radius: var(--radius-sm);
    font-size: 0.82rem; font-weight: 700; cursor: pointer; transition: var(--transition);
  }
  .btn-danger:hover { background: #dc2626; }
  .btn-cancel {
    padding: 8px 18px; background: var(--white); color: var(--text2);
    border: 1.5px solid var(--border); border-radius: var(--radius-sm);
    font-size: 0.82rem; font-weight: 700; cursor: pointer; transition: var(--transition);
  }
  .btn-cancel:hover { background: var(--bg); }

  /* ── Responsive ── */
  @media (max-width: 900px) {
    .summary-row { grid-template-columns: repeat(2, 1fr); }
  }
  @media (max-width: 640px) {
    h1 { font-size: 1.4rem; }
    .summary-row { grid-template-columns: 1fr 1fr; }
    .tabs { width: 100%; overflow-x: auto; }
    .tab-btn { padding: 7px 12px; font-size: 0.78rem; }
    .timeline-track { width: 20px; }
    .result-card { padding: 12px 14px; gap: 10px; }
    .type-icon-wrap { width: 36px; height: 36px; font-size: 1rem; }
    .percent-col { min-width: 44px; }
    .percent-val { font-size: 1.1rem; }
    .page-header { flex-direction: column; }
    .join-btn-link { width: 100%; justify-content: center; }
    .confirm-modal { flex-direction: column; text-align: center; }
  }
  @media (max-width: 400px) {
    .summary-row { grid-template-columns: 1fr; }
  }

  @keyframes fadeSlideUp {
    from { opacity: 0; transform: translateY(14px); }
    to   { opacity: 1; transform: translateY(0); }
  }
</style>
