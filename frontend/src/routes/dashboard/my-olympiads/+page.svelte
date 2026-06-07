<script lang="ts">
  import { onMount } from 'svelte';

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

  let olympiads: StudentResult[] = [];
  let loaded = false;
  let clearConfirm = false;

  // ── Load from localStorage ────────────────────────────────────────────────
  onMount(() => load());

  function load() {
    try {
      const raw = localStorage.getItem('gogame_results');
      if (raw) {
        const all = JSON.parse(raw) as StudentResult[];
        olympiads = all
          .filter(r => r.type === 'olympiad')
          .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
      } else {
        olympiads = [];
      }
    } catch {
      olympiads = [];
    }
    loaded = true;
  }

  // ── Stats ─────────────────────────────────────────────────────────────────
  $: totalOlympiads = olympiads.length;
  $: avgPercent     = olympiads.length
    ? Math.round(olympiads.reduce((s, o) => s + o.percent, 0) / olympiads.length)
    : 0;
  $: bestPercent    = olympiads.length ? Math.max(...olympiads.map(o => o.percent)) : 0;

  // ── Clear ─────────────────────────────────────────────────────────────────
  function clearAll() {
    try {
      const raw = localStorage.getItem('gogame_results');
      if (raw) {
        const all = JSON.parse(raw) as StudentResult[];
        localStorage.setItem('gogame_results', JSON.stringify(all.filter(r => r.type !== 'olympiad')));
      }
    } catch {}
    olympiads = [];
    clearConfirm = false;
  }

  // ── Helpers ───────────────────────────────────────────────────────────────
  function fmtDate(d: string) {
    return new Date(d).toLocaleDateString('uz-UZ', {
      day: '2-digit', month: 'short', year: 'numeric'
    });
  }

  function fmtTime(sec: number) {
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    return m > 0 ? `${m}m ${s}s` : `${s}s`;
  }

  // Rank medal — based on sorted order (1st, 2nd, 3rd = medals)
  function rankMedal(index: number): { emoji: string; color: string; glow: string } {
    if (index === 0) return { emoji: '🥇', color: '#eab308', glow: 'rgba(234,179,8,0.25)' };
    if (index === 1) return { emoji: '🥈', color: '#94a3b8', glow: 'rgba(148,163,184,0.2)' };
    if (index === 2) return { emoji: '🥉', color: '#f97316', glow: 'rgba(249,115,22,0.2)' };
    return { emoji: `#${index + 1}`, color: 'var(--text3)', glow: 'transparent' };
  }

  function barColor(pct: number) {
    if (pct >= 80) return 'linear-gradient(90deg, #eab308, #ca8a04)';
    if (pct >= 60) return 'linear-gradient(90deg, #f59e0b, #d97706)';
    if (pct >= 40) return 'linear-gradient(90deg, #6366f1, #8b5cf6)';
    return 'linear-gradient(90deg, #ef4444, #dc2626)';
  }
</script>

<svelte:head><title>Olimpiadalarim — Cognita.uz</title></svelte:head>

<!-- ── Page Header ── -->
<div class="page-header">
  <div>
    <h1>Olimpiadalarim</h1>
    <p class="sub">Barcha musobaqalardagi natijalaringiz</p>
  </div>
  <a href="/dashboard/join" class="join-link-btn">
    <span>🏆</span> Olimpiadaga kirish
  </a>
</div>

<!-- ── Stats Strip ── -->
<div class="stats-strip">
  <div class="stat-card">
    <div class="stat-icon" style="background:rgba(234,179,8,0.12);">🏆</div>
    <div class="stat-body">
      <div class="stat-value">{loaded ? totalOlympiads : '…'}</div>
      <div class="stat-label">Jami olimpiadalar</div>
    </div>
  </div>
  <div class="stat-card">
    <div class="stat-icon" style="background:rgba(99,102,241,0.12);">📊</div>
    <div class="stat-body">
      <div class="stat-value" style="color:var(--primary)">{loaded ? avgPercent : '…'}%</div>
      <div class="stat-label">O'rtacha reyting</div>
    </div>
  </div>
  <div class="stat-card">
    <div class="stat-icon" style="background:rgba(234,179,8,0.12);">🥇</div>
    <div class="stat-body">
      <div class="stat-value" style="color:#eab308">{loaded ? bestPercent : '…'}%</div>
      <div class="stat-label">Eng yuqori reyting</div>
    </div>
  </div>
</div>

<!-- ── Content ── -->
{#if !loaded}
  <div class="card">
    {#each Array(4) as _}
      <div class="skeleton sk-row"></div>
    {/each}
  </div>
{:else if olympiads.length === 0}
  <div class="empty-state card">
    <div class="empty-emoji">🏆</div>
    <h3>Hali olimpiadada qatnashmadingiz</h3>
    <p>Olimpiada kodini oling va tanlovda ishtirok eting, sovrinlar yutib oling!</p>
    <a href="/dashboard/join" class="empty-btn">
      🏆 Olimpiadaga kirish
    </a>
  </div>
{:else}
  <!-- Leaderboard-style list sorted by percent desc -->
  <div class="results-list">
    {#each olympiads as item, i}
      {@const medal = rankMedal(i)}
      {@const bc = barColor(item.percent)}
      <div
        class="result-card"
        class:top-card={i < 3}
        style="animation-delay:{i * 40}ms;{i < 3 ? `--glow:${medal.glow}` : ''}"
      >
        <!-- Rank -->
        <div class="rank-col" style="color:{medal.color}">
          {#if i < 3}
            <span class="medal">{medal.emoji}</span>
          {:else}
            <span class="rank-num">{medal.emoji}</span>
          {/if}
        </div>

        <!-- Info -->
        <div class="result-info">
          <div class="result-title">{item.title}</div>
          <div class="result-meta">
            {#if item.code}
              <span class="meta-pill">🔑 {item.code}</span>
            {/if}
            <span class="meta-pill">⏱ {fmtTime(item.timeTaken)}</span>
            <span class="meta-pill">📅 {fmtDate(item.date)}</span>
          </div>
          <!-- Score bar -->
          <div class="score-bar-wrap">
            <div class="score-bar-track">
              <div class="score-bar-fill" style="width:{item.percent}%;background:{bc}"></div>
            </div>
            <span class="score-text">{item.score}/{item.maxScore}</span>
          </div>
        </div>

        <!-- Percent -->
        <div
          class="result-percent"
          style="color:{item.percent >= 80 ? '#eab308' : item.percent >= 60 ? 'var(--primary)' : 'var(--text2)'}"
        >
          {item.percent}%
        </div>
      </div>
    {/each}
  </div>

  <!-- Clear section -->
  <div class="clear-wrap">
    {#if clearConfirm}
      <div class="confirm-box">
        <span>Barcha olimpiada natijalarini o'chirmoqchimisiz?</span>
        <div class="confirm-btns">
          <button class="confirm-yes" on:click={clearAll}>Ha, o'chirish</button>
          <button class="confirm-no" on:click={() => clearConfirm = false}>Bekor qilish</button>
        </div>
      </div>
    {:else}
      <button class="clear-btn" on:click={() => clearConfirm = true}>
        🗑 Barcha natijalarni o'chirish
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
  h1 { font-size: 1.75rem; font-weight: 800; color: var(--text); margin: 0; }
  .sub { font-size: 0.875rem; color: var(--text3); margin-top: 4px; }
  .join-link-btn {
    display: inline-flex; align-items: center; gap: 7px;
    padding: 10px 20px;
    background: linear-gradient(135deg, #eab308, #ca8a04);
    color: #fff; border-radius: var(--radius);
    font-size: 0.875rem; font-weight: 700; text-decoration: none;
    box-shadow: 0 4px 14px rgba(234,179,8,0.35);
    transition: var(--transition); white-space: nowrap; flex-shrink: 0;
  }
  .join-link-btn:hover { transform: translateY(-2px); box-shadow: 0 6px 20px rgba(234,179,8,0.45); }

  /* ── Stats strip ── */
  .stats-strip {
    display: grid; grid-template-columns: repeat(3, 1fr);
    gap: 14px; margin-bottom: 24px;
  }
  .stat-card {
    background: var(--white); border: 1.5px solid var(--border);
    border-radius: var(--radius); padding: 16px 18px;
    display: flex; align-items: center; gap: 14px;
    box-shadow: var(--shadow-sm); transition: var(--transition);
    animation: fadeSlideUp 0.35s ease both;
  }
  .stat-card:hover { transform: translateY(-2px); box-shadow: var(--shadow); }
  .stat-icon {
    width: 46px; height: 46px; border-radius: var(--radius-sm);
    display: flex; align-items: center; justify-content: center;
    font-size: 1.3rem; flex-shrink: 0;
  }
  .stat-value { font-size: 1.5rem; font-weight: 900; color: var(--text); line-height: 1; }
  .stat-label { font-size: 0.75rem; color: var(--text3); font-weight: 600; margin-top: 3px; }

  /* ── Skeleton ── */
  .card {
    background: var(--white); border: 1.5px solid var(--border);
    border-radius: var(--radius-lg); padding: 20px; box-shadow: var(--shadow-sm);
  }
  .skeleton {
    background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
    background-size: 200% 100%; animation: shimmer 1.4s infinite;
    border-radius: var(--radius-sm);
  }
  .sk-row { height: 52px; margin-bottom: 10px; }
  @keyframes shimmer { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }

  /* ── Empty state ── */
  .empty-state {
    display: flex; flex-direction: column;
    align-items: center; gap: 12px;
    padding: 56px 24px; text-align: center;
  }
  .empty-emoji { font-size: 3.5rem; }
  .empty-state h3 { font-size: 1.1rem; font-weight: 800; color: var(--text); margin: 0; }
  .empty-state p { font-size: 0.875rem; color: var(--text3); max-width: 360px; margin: 0; }
  .empty-btn {
    display: inline-flex; align-items: center; gap: 7px;
    margin-top: 8px; padding: 11px 24px;
    background: linear-gradient(135deg, #eab308, #ca8a04);
    color: #fff; border-radius: var(--radius);
    font-size: 0.9rem; font-weight: 700; text-decoration: none;
    box-shadow: 0 4px 14px rgba(234,179,8,0.35); transition: var(--transition);
  }
  .empty-btn:hover { transform: translateY(-2px); }

  /* ── Results list ── */
  .results-list { display: flex; flex-direction: column; gap: 10px; margin-bottom: 28px; }
  .result-card {
    display: flex; align-items: center; gap: 16px;
    padding: 16px 20px;
    background: var(--white);
    border: 1.5px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow-sm);
    transition: var(--transition);
    animation: fadeSlideUp 0.35s ease both;
  }
  .result-card:hover { border-color: #eab308; transform: translateX(3px); }
  .top-card {
    border-color: rgba(234,179,8,0.4);
    box-shadow: 0 2px 12px var(--glow, transparent);
  }
  .top-card:hover { box-shadow: 0 4px 20px var(--glow, transparent); }

  /* Rank */
  .rank-col { width: 44px; text-align: center; flex-shrink: 0; }
  .medal { font-size: 1.8rem; line-height: 1; }
  .rank-num { font-size: 0.85rem; font-weight: 800; color: var(--text3); }

  .result-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6px; }
  .result-title {
    font-size: 0.95rem; font-weight: 700; color: var(--text);
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  }
  .result-meta { display: flex; gap: 6px; flex-wrap: wrap; }
  .meta-pill {
    font-size: 0.7rem; font-weight: 600;
    padding: 2px 8px; background: var(--bg);
    color: var(--text3); border-radius: 99px;
    border: 1px solid var(--border); white-space: nowrap;
  }
  .score-bar-wrap { display: flex; align-items: center; gap: 10px; }
  .score-bar-track { flex: 1; height: 6px; background: var(--bg); border-radius: 99px; overflow: hidden; }
  .score-bar-fill { height: 100%; border-radius: 99px; transition: width 0.6s ease; }
  .score-text { font-size: 0.72rem; font-weight: 700; color: var(--text2); white-space: nowrap; min-width: 40px; text-align: right; }

  .result-percent { font-size: 1.4rem; font-weight: 900; flex-shrink: 0; min-width: 56px; text-align: right; letter-spacing: -0.5px; }

  /* ── Clear section ── */
  .clear-wrap { display: flex; justify-content: center; padding-bottom: 8px; }
  .clear-btn {
    padding: 9px 20px; background: transparent; color: var(--danger);
    border: 1.5px solid var(--danger); border-radius: var(--radius-sm);
    font-size: 0.82rem; font-weight: 700; cursor: pointer;
    transition: var(--transition); opacity: 0.7;
  }
  .clear-btn:hover { background: rgba(239,68,68,0.08); opacity: 1; }
  .confirm-box {
    display: flex; flex-direction: column; align-items: center; gap: 12px;
    padding: 16px 24px; background: #fee2e2;
    border: 1.5px solid #fecaca; border-radius: var(--radius);
    font-size: 0.875rem; color: #b91c1c; font-weight: 600; text-align: center;
    animation: fadeSlideUp 0.2s ease both;
  }
  .confirm-btns { display: flex; gap: 10px; }
  .confirm-yes { padding: 8px 18px; background: var(--danger); color: #fff; border: none; border-radius: var(--radius-sm); font-size: 0.82rem; font-weight: 700; cursor: pointer; transition: var(--transition); }
  .confirm-yes:hover { background: #dc2626; }
  .confirm-no { padding: 8px 18px; background: var(--white); color: var(--text2); border: 1.5px solid var(--border); border-radius: var(--radius-sm); font-size: 0.82rem; font-weight: 700; cursor: pointer; transition: var(--transition); }
  .confirm-no:hover { background: var(--bg); }

  /* ── Responsive ── */
  @media (max-width: 640px) {
    h1 { font-size: 1.4rem; }
    .stats-strip { grid-template-columns: 1fr 1fr; }
    .stats-strip .stat-card:last-child { grid-column: 1 / -1; }
    .result-card { flex-wrap: wrap; }
    .result-percent { width: 100%; text-align: left; }
    .page-header { flex-direction: column; }
    .join-link-btn { width: 100%; justify-content: center; }
  }
  @media (max-width: 400px) {
    .stats-strip { grid-template-columns: 1fr; }
    .stats-strip .stat-card:last-child { grid-column: auto; }
  }

  @keyframes fadeSlideUp {
    from { opacity: 0; transform: translateY(14px); }
    to   { opacity: 1; transform: translateY(0); }
  }
</style>
