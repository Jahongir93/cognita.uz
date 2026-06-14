<script lang="ts">
  import type { Category, Test } from '$lib/data/categories';

  export let category: Category;

  let search = '';
  let activeSubcat = 'all';
  let activeDiff: string = 'all';
  let showModal = false;
  let selectedTest: Test | null = null;

  $: filtered = category.tests.filter(t => {
    const matchSearch =
      !search ||
      t.title.toLowerCase().includes(search.toLowerCase()) ||
      t.description.toLowerCase().includes(search.toLowerCase());
    const matchSubcat = activeSubcat === 'all' || t.subcat === activeSubcat;
    const matchDiff = activeDiff === 'all' || t.difficulty === activeDiff;
    return matchSearch && matchSubcat && matchDiff;
  });

  $: totalPlays = category.tests.reduce((s, t) => s + t.plays, 0);
  $: avgRating = (
    category.tests.reduce((s, t) => s + t.rating, 0) / category.tests.length
  ).toFixed(1);

  function openTest(t: Test) {
    selectedTest = t;
    showModal = true;
  }
  function closeModal() {
    showModal = false;
    selectedTest = null;
  }
  function fmtPlays(n: number): string {
    return n >= 1000 ? (n / 1000).toFixed(1) + 'K' : n.toString();
  }
  function diffLabel(d: string): string {
    return d === 'easy' ? 'Oson' : d === 'medium' ? "O'rta" : 'Qiyin';
  }
  function diffColor(d: string): string {
    return d === 'easy' ? '#22c55e' : d === 'medium' ? '#f59e0b' : '#ef4444';
  }
</script>

<svelte:head><title>{category.title} — Cognita.uz</title></svelte:head>

<!-- NAV -->
<nav class="topnav">
  <a href="/" class="brand">
    <img src="/logowhite.png" alt="Cognita.uz" style="height:28px;width:auto" />
  </a>
  <div class="nav-right">
    <a href="/" class="back">← Bosh sahifa</a>
    <a href="/join" class="join-btn">O'yinga kirish</a>
  </div>
</nav>

<!-- HERO -->
<header class="hero" style="--g1:{category.g1};--g2:{category.g2}">
  <div class="hero-bg" aria-hidden="true">
    <div class="shape s1"></div>
    <div class="shape s2"></div>
    <div class="shape s3"></div>
    <div class="shape s4"></div>
    <div class="shape s5"></div>
  </div>
  <div class="hero-inner">
    <div class="hero-icon">{category.icon}</div>
    <h1>{category.title}</h1>
    <p>{category.subtitle}</p>
    <div class="stats-row">
      <div class="stat">
        <span class="stat-num">{category.tests.length}</span>
        <span class="stat-lbl">test</span>
      </div>
      <div class="stat-div"></div>
      <div class="stat">
        <span class="stat-num">{fmtPlays(totalPlays)}</span>
        <span class="stat-lbl">o'ynalgan</span>
      </div>
      <div class="stat-div"></div>
      <div class="stat">
        <span class="stat-num">⭐ {avgRating}</span>
        <span class="stat-lbl">reyting</span>
      </div>
    </div>
  </div>
</header>

<!-- CONTROLS -->
<div class="controls-wrap">
  <div class="controls">
    <div class="search-wrap">
      <span class="search-icon">🔍</span>
      <input
        class="search-input"
        placeholder="Test qidirish..."
        bind:value={search}
      />
    </div>
    <div class="diff-chips">
      {#each [['all','Barchasi'],['easy','Oson'],['medium',"O'rta"],['hard','Qiyin']] as [val, lbl]}
        <button
          class="chip"
          class:active={activeDiff === val}
          on:click={() => activeDiff = val}
        >{lbl}</button>
      {/each}
    </div>
  </div>

  <!-- SUBCATEGORY TABS -->
  <div class="subcats">
    {#each category.subcats as sc}
      <button
        class="subcat-btn"
        class:active={activeSubcat === sc.id}
        on:click={() => activeSubcat = sc.id}
      >
        <span>{sc.icon}</span>
        <span>{sc.label}</span>
      </button>
    {/each}
  </div>
</div>

<!-- GRID -->
<main class="content">
  {#if filtered.length === 0}
    <div class="empty">
      <div class="empty-icon">🔍</div>
      <p>Hech narsa topilmadi</p>
      <button class="reset-btn" on:click={() => { search = ''; activeSubcat = 'all'; activeDiff = 'all'; }}>
        Filterni tozalash
      </button>
    </div>
  {:else}
    <div class="grid">
      {#each filtered as test, i}
        <div
          class="card"
          style="--ci:{i};--g1:{category.g1};--g2:{category.g2}"
          role="button"
          tabindex="0"
          on:click={() => openTest(test)}
          on:keydown={e => e.key === 'Enter' && openTest(test)}
        >
          <div class="card-top">
            <span class="card-icon">{test.icon}</span>
            <div class="badges">
              {#if test.isHot}
                <span class="badge hot">🔥 Mashhur</span>
              {/if}
              {#if test.isNew}
                <span class="badge new-badge">✨ Yangi</span>
              {/if}
            </div>
          </div>
          <h3 class="card-title">{test.title}</h3>
          <p class="card-desc">{test.description}</p>
          <div class="card-meta">
            <span>⏱ {test.duration} min</span>
            <span>👤 {fmtPlays(test.plays)} marta</span>
            <span>⭐ {test.rating}</span>
          </div>
          <div class="card-footer">
            <span class="diff-badge" style="color:{diffColor(test.difficulty)};border-color:{diffColor(test.difficulty)}">
              {diffLabel(test.difficulty)}
            </span>
            <button class="play-btn" on:click|stopPropagation={() => openTest(test)}>
              Boshlash ▶
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</main>

<!-- MODAL -->
{#if showModal && selectedTest}
  <div
    class="overlay"
    role="button"
    tabindex="-1"
    aria-label="Yopish"
    on:click={closeModal}
    on:keydown={e => e.key === 'Escape' && closeModal()}
  >
    <div
      class="modal"
      role="dialog"
      aria-modal="true"
      on:click|stopPropagation
      on:keydown|stopPropagation
    >
      <button class="modal-close" on:click={closeModal} aria-label="Yopish">✕</button>
      <div class="modal-icon">{selectedTest.icon}</div>
      <h2 class="modal-title">{selectedTest.title}</h2>
      <p class="modal-desc">{selectedTest.description}</p>
      <div class="modal-rows">
        <div class="modal-row">
          <span class="modal-lbl">Qiyinlik</span>
          <span class="modal-val" style="color:{diffColor(selectedTest.difficulty)}">
            {diffLabel(selectedTest.difficulty)}
          </span>
        </div>
        <div class="modal-row">
          <span class="modal-lbl">Savollar</span>
          <span class="modal-val">{selectedTest.questions} ta</span>
        </div>
        <div class="modal-row">
          <span class="modal-lbl">Vaqt</span>
          <span class="modal-val">{selectedTest.duration} daqiqa</span>
        </div>
        <div class="modal-row">
          <span class="modal-lbl">O'ynalgan</span>
          <span class="modal-val">{fmtPlays(selectedTest.plays)} marta</span>
        </div>
        <div class="modal-row">
          <span class="modal-lbl">Reyting</span>
          <span class="modal-val">⭐ {selectedTest.rating}</span>
        </div>
      </div>
      <div class="modal-actions">
        <button class="btn-close" on:click={closeModal}>Yopish</button>
        <button class="btn-play" disabled>O'ynash (tez kunda)</button>
      </div>
    </div>
  </div>
{/if}

<style>
  :global(body) {
    margin: 0;
    font-family: 'Segoe UI', system-ui, sans-serif;
    background: #0f0e17;
    color: #e8e6f0;
  }

  /* ── NAV ── */
  .topnav {
    position: sticky;
    top: 0;
    z-index: 200;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 2rem;
    height: 60px;
    background: rgba(20, 18, 60, 0.97);
    backdrop-filter: blur(12px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 1.2rem;
    font-weight: 700;
    color: #fff;
    text-decoration: none;
  }
  .brand-icon { font-size: 1.4rem; }
  .dot { color: #a78bfa; }
  .nav-right {
    display: flex;
    align-items: center;
    gap: 1rem;
  }
  .back {
    color: #c4b5fd;
    text-decoration: none;
    font-size: 0.9rem;
    transition: color 0.2s;
  }
  .back:hover { color: #fff; }
  .join-btn {
    background: linear-gradient(135deg, var(--g1, #a78bfa), var(--g2, #6366f1));
    color: #fff;
    padding: 7px 18px;
    border-radius: 20px;
    text-decoration: none;
    font-size: 0.85rem;
    font-weight: 600;
    transition: opacity 0.2s, transform 0.2s;
  }
  .join-btn:hover { opacity: 0.9; transform: translateY(-1px); }

  /* ── HERO ── */
  .hero {
    position: relative;
    overflow: hidden;
    background: linear-gradient(135deg, var(--g1), var(--g2));
    padding: 60px 24px 50px;
    text-align: center;
  }
  .hero-bg {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }
  .shape {
    position: absolute;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.08);
  }
  .s1 { width: 320px; height: 320px; top: -100px; left: -80px; }
  .s2 { width: 220px; height: 220px; top: 20px; right: 8%; opacity: 0.5; }
  .s3 { width: 160px; height: 160px; bottom: -50px; left: 25%; opacity: 0.4; }
  .s4 { width: 90px;  height: 90px;  top: 45%;  right: 4%; opacity: 0.6; }
  .s5 { width: 120px; height: 120px; bottom: 10px; right: 22%; opacity: 0.3; }

  .hero-inner {
    position: relative;
    z-index: 1;
    max-width: 700px;
    margin: 0 auto;
  }
  .hero-icon {
    font-size: 4rem;
    margin-bottom: 1rem;
    display: inline-block;
    animation: bounce 2s ease infinite;
    filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.3));
  }
  @keyframes bounce {
    0%, 100% { transform: translateY(0); }
    50%       { transform: translateY(-10px); }
  }
  .hero h1 {
    font-size: 2.5rem;
    font-weight: 800;
    margin: 0 0 0.75rem;
    color: #fff;
    text-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
  }
  .hero p {
    font-size: 1.1rem;
    color: rgba(255, 255, 255, 0.9);
    margin: 0 0 2rem;
  }
  .stats-row {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0;
    flex-wrap: wrap;
    background: rgba(255, 255, 255, 0.15);
    backdrop-filter: blur(10px);
    border-radius: 16px;
    padding: 16px 24px;
    display: inline-flex;
  }
  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 0 20px;
  }
  .stat-num {
    font-size: 1.4rem;
    font-weight: 800;
    color: #fff;
  }
  .stat-lbl {
    font-size: 0.78rem;
    color: rgba(255, 255, 255, 0.75);
    margin-top: 2px;
  }
  .stat-div {
    width: 1px;
    height: 36px;
    background: rgba(255, 255, 255, 0.3);
  }

  /* ── CONTROLS ── */
  .controls-wrap {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1.5rem 1.5rem 0;
  }
  .controls {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
    align-items: center;
    margin-bottom: 1rem;
  }
  .search-wrap {
    position: relative;
    flex: 1;
    min-width: 200px;
  }
  .search-icon {
    position: absolute;
    left: 12px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 1rem;
    pointer-events: none;
  }
  .search-input {
    width: 100%;
    box-sizing: border-box;
    padding: 10px 14px 10px 38px;
    border-radius: 10px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    background: rgba(255, 255, 255, 0.07);
    color: #e8e6f0;
    font-size: 0.92rem;
    outline: none;
    transition: border-color 0.2s, background 0.2s;
  }
  .search-input::placeholder { color: rgba(255, 255, 255, 0.35); }
  .search-input:focus {
    border-color: rgba(167, 139, 250, 0.6);
    background: rgba(255, 255, 255, 0.1);
  }
  .diff-chips {
    display: flex;
    gap: 0.4rem;
    flex-wrap: wrap;
  }
  .chip {
    padding: 7px 16px;
    border-radius: 20px;
    border: 1px solid rgba(255, 255, 255, 0.15);
    background: rgba(255, 255, 255, 0.06);
    color: rgba(255, 255, 255, 0.7);
    font-size: 0.85rem;
    cursor: pointer;
    transition: all 0.2s;
  }
  .chip:hover { background: rgba(255, 255, 255, 0.12); color: #fff; }
  .chip.active {
    background: linear-gradient(135deg, var(--g1, #a78bfa), var(--g2, #6366f1));
    border-color: transparent;
    color: #fff;
    font-weight: 600;
  }

  /* ── SUBCATS ── */
  .subcats {
    display: flex;
    gap: 0.5rem;
    overflow-x: auto;
    padding-bottom: 0.5rem;
    scrollbar-width: thin;
    scrollbar-color: rgba(255,255,255,0.15) transparent;
  }
  .subcats::-webkit-scrollbar { height: 4px; }
  .subcats::-webkit-scrollbar-track { background: transparent; }
  .subcats::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.2); border-radius: 4px; }
  .subcat-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
    padding: 8px 18px;
    border-radius: 24px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    background: rgba(255, 255, 255, 0.06);
    color: rgba(255, 255, 255, 0.65);
    font-size: 0.88rem;
    cursor: pointer;
    transition: all 0.2s;
    flex-shrink: 0;
  }
  .subcat-btn:hover { background: rgba(255,255,255,0.12); color: #fff; }
  .subcat-btn.active {
    background: linear-gradient(135deg, var(--g1, #a78bfa), var(--g2, #6366f1));
    border-color: transparent;
    color: #fff;
    font-weight: 600;
  }

  /* ── GRID ── */
  .content {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1.5rem 1.5rem 3rem;
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1.25rem;
  }

  /* ── CARDS ── */
  .card {
    background: #fff;
    border-radius: 16px;
    padding: 1.4rem;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    color: #1e1b4b;
    cursor: pointer;
    animation: cardIn 0.4s ease calc(var(--ci, 0) * 60ms) both;
    transition: transform 0.25s, box-shadow 0.25s;
    border: 2px solid transparent;
    position: relative;
    overflow: hidden;
  }
  .card::before {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: 14px;
    padding: 2px;
    background: linear-gradient(135deg, var(--g1, #a78bfa), var(--g2, #6366f1));
    -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
    mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
    -webkit-mask-composite: xor;
    mask-composite: exclude;
    opacity: 0;
    transition: opacity 0.25s;
    pointer-events: none;
  }
  .card:hover {
    transform: translateY(-5px);
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.18);
  }
  .card:hover::before { opacity: 1; }

  @keyframes cardIn {
    from { opacity: 0; transform: translateY(20px); }
    to   { opacity: 1; transform: none; }
  }

  .card-top {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.5rem;
  }
  .card-icon { font-size: 2.2rem; line-height: 1; }
  .badges { display: flex; flex-direction: column; gap: 4px; align-items: flex-end; }
  .badge {
    font-size: 0.68rem;
    font-weight: 700;
    padding: 3px 8px;
    border-radius: 10px;
    white-space: nowrap;
    letter-spacing: 0.02em;
  }
  .badge.hot {
    background: linear-gradient(135deg, #f59e0b, #ef4444);
    color: #fff;
  }
  .badge.new-badge {
    background: linear-gradient(135deg, #22c55e, #0ea5e9);
    color: #fff;
  }

  .card-title {
    margin: 0;
    font-size: 1rem;
    font-weight: 700;
    color: #1e1b4b;
    line-height: 1.3;
  }
  .card-desc {
    margin: 0;
    font-size: 0.85rem;
    color: #6b7280;
    line-height: 1.5;
    flex: 1;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .card-meta {
    display: flex;
    gap: 0.8rem;
    flex-wrap: wrap;
    font-size: 0.78rem;
    color: #9ca3af;
  }
  .card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    margin-top: auto;
  }
  .diff-badge {
    font-size: 0.78rem;
    font-weight: 700;
    padding: 3px 10px;
    border-radius: 8px;
    border: 1.5px solid;
  }
  .play-btn {
    background: linear-gradient(135deg, var(--g1, #a78bfa), var(--g2, #6366f1));
    color: #fff;
    border: none;
    border-radius: 8px;
    padding: 7px 16px;
    font-size: 0.82rem;
    font-weight: 700;
    cursor: pointer;
    transition: opacity 0.2s, transform 0.2s;
    white-space: nowrap;
  }
  .play-btn:hover { opacity: 0.88; transform: scale(1.04); }

  /* ── EMPTY STATE ── */
  .empty {
    text-align: center;
    padding: 5rem 1rem;
    color: rgba(255, 255, 255, 0.5);
  }
  .empty-icon { font-size: 4rem; margin-bottom: 1rem; }
  .empty p { font-size: 1.1rem; margin: 0 0 1.5rem; }
  .reset-btn {
    background: rgba(255,255,255,0.1);
    color: #fff;
    border: 1px solid rgba(255,255,255,0.2);
    padding: 10px 24px;
    border-radius: 10px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: background 0.2s;
  }
  .reset-btn:hover { background: rgba(255,255,255,0.18); }

  /* ── MODAL ── */
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
  }
  .modal {
    background: #fff;
    border-radius: 20px;
    padding: 2rem;
    max-width: 440px;
    width: 100%;
    position: relative;
    color: #1e1b4b;
    animation: cardIn 0.3s ease both;
  }
  .modal-close {
    position: absolute;
    top: 1rem;
    right: 1rem;
    background: rgba(0, 0, 0, 0.07);
    border: none;
    border-radius: 50%;
    width: 32px;
    height: 32px;
    font-size: 1rem;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.2s;
    color: #374151;
  }
  .modal-close:hover { background: rgba(0,0,0,0.14); }
  .modal-icon { font-size: 3rem; margin-bottom: 0.75rem; }
  .modal-title {
    font-size: 1.3rem;
    font-weight: 800;
    margin: 0 0 0.5rem;
    color: #1e1b4b;
  }
  .modal-desc {
    font-size: 0.9rem;
    color: #6b7280;
    margin: 0 0 1.25rem;
    line-height: 1.6;
  }
  .modal-rows {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
  }
  .modal-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    background: #f9fafb;
    border-radius: 8px;
  }
  .modal-lbl { font-size: 0.85rem; color: #9ca3af; }
  .modal-val { font-size: 0.9rem; font-weight: 600; color: #374151; }
  .modal-actions {
    display: flex;
    gap: 0.75rem;
  }
  .btn-close {
    flex: 1;
    padding: 10px;
    border-radius: 10px;
    border: 1.5px solid #e5e7eb;
    background: #fff;
    color: #374151;
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
  }
  .btn-close:hover { background: #f3f4f6; }
  .btn-play {
    flex: 2;
    padding: 10px;
    border-radius: 10px;
    border: none;
    background: linear-gradient(135deg, #d1d5db, #9ca3af);
    color: #fff;
    font-size: 0.9rem;
    font-weight: 700;
    cursor: not-allowed;
    opacity: 0.75;
  }

  /* ── RESPONSIVE ── */
  @media (max-width: 900px) {
    .grid { grid-template-columns: repeat(2, 1fr); }
    .hero h1 { font-size: 2rem; }
  }
  @media (max-width: 560px) {
    .grid { grid-template-columns: 1fr; }
    .hero { padding: 40px 16px 36px; }
    .hero h1 { font-size: 1.6rem; }
    .topnav { padding: 0 1rem; }
    .controls-wrap { padding: 1rem 1rem 0; }
    .content { padding: 1.25rem 1rem 2.5rem; }
    .stats-row { gap: 0; padding: 12px 16px; }
    .stat { padding: 0 12px; }
    .stat-num { font-size: 1.1rem; }
  }
</style>
