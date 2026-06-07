<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  // ── State ──────────────────────────────────────────────────────────────────
  let gamePin = '';
  let examCode = '';
  let olympiadCode = '';

  let pinError = '';
  let examError = '';
  let olympiadError = '';

  interface RecentEntry {
    type: 'game' | 'exam' | 'olympiad';
    title: string;
    pin?: string;
    code?: string;
    date: string;
    percent: number;
  }

  let recentEntries: RecentEntry[] = [];

  // ── Load recent entries from localStorage ─────────────────────────────────
  onMount(() => {
    try {
      const raw = localStorage.getItem('gogame_results');
      if (raw) {
        const all = JSON.parse(raw) as Array<{
          type: 'game' | 'exam' | 'olympiad';
          title: string;
          pin?: string;
          code?: string;
          date: string;
          percent: number;
        }>;
        recentEntries = [...all]
          .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
          .slice(0, 4);
      }
    } catch {
      recentEntries = [];
    }
  });

  // ── Handlers ──────────────────────────────────────────────────────────────
  function handleGame() {
    pinError = '';
    const clean = gamePin.trim();
    if (!clean) { pinError = 'PIN kod kiritilmadi'; return; }
    if (!/^\d{6}$/.test(clean)) { pinError = 'PIN 6 ta raqamdan iborat bo\'lishi kerak'; return; }
    goto(`/join?pin=${clean}`);
  }

  function handleExam() {
    examError = '';
    const clean = examCode.trim().toUpperCase();
    if (!clean) { examError = 'Imtihon kodi kiritilmadi'; return; }
    if (clean.length !== 8) { examError = 'Kod 8 ta belgidan iborat bo\'lishi kerak'; return; }
    goto(`/exam/${clean}`);
  }

  function handleOlympiad() {
    olympiadError = '';
    const clean = olympiadCode.trim().toUpperCase();
    if (!clean) { olympiadError = 'Olimpiada kodi kiritilmadi'; return; }
    if (clean.length !== 8) { olympiadError = 'Kod 8 ta belgidan iborat bo\'lishi kerak'; return; }
    goto(`/olympiad/${clean}`);
  }

  function handlePinInput(e: Event) {
    const input = e.target as HTMLInputElement;
    gamePin = input.value.replace(/\D/g, '').slice(0, 6);
    input.value = gamePin;
    if (pinError) pinError = '';
  }

  function handleCodeInput(field: 'exam' | 'olympiad', e: Event) {
    const input = e.target as HTMLInputElement;
    const val = input.value.toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, 8);
    if (field === 'exam') { examCode = val; if (examError) examError = ''; }
    else { olympiadCode = val; if (olympiadError) olympiadError = ''; }
    input.value = val;
  }

  function handleKeydown(e: KeyboardEvent, action: () => void) {
    if (e.key === 'Enter') action();
  }

  function continueEntry(entry: RecentEntry) {
    if (entry.type === 'game' && entry.pin) goto(`/join?pin=${entry.pin}`);
    else if (entry.type === 'exam' && entry.code) goto(`/exam/${entry.code}`);
    else if (entry.type === 'olympiad' && entry.code) goto(`/olympiad/${entry.code}`);
  }

  function fmtDate(d: string) {
    return new Date(d).toLocaleDateString('uz-UZ', { day: '2-digit', month: 'short' });
  }

  const typeIcon: Record<string, string> = { game: '🎮', exam: '📋', olympiad: '🏆' };
  const typeLabel: Record<string, string> = { game: "O'yin", exam: 'Imtihon', olympiad: 'Olimpiada' };
</script>

<svelte:head><title>O'yinga kirish — Cognita.uz</title></svelte:head>

<!-- ── Page header ── -->
<div class="page-header">
  <div>
    <h1>Kirish</h1>
    <p class="sub">O'yin, imtihon yoki olimpiadaga qo'shiling</p>
  </div>
</div>

<!-- ══════════════════════════════════════════════════
     3 JOIN CARDS
══════════════════════════════════════════════════ -->
<div class="cards-grid">

  <!-- ── Card 1: O'yinga kirish ── -->
  <div class="join-card game-card">
    <div class="card-top game-top">
      <div class="card-dots" aria-hidden="true"></div>
      <div class="card-top-inner">
        <div class="card-emoji">🎮</div>
        <div class="card-top-text">
          <h2>O'yinga kirish</h2>
          <p>Jonli viktorinaga qo'shiling</p>
        </div>
      </div>
    </div>
    <div class="card-body">
      <p class="card-desc">
        O'qituvchi tomonidan berilgan 6 xonali PIN kodni kiriting va darhol o'yinga qo'shiling.
      </p>
      <div class="input-wrap" class:has-error={!!pinError}>
        <input
          type="tel"
          inputmode="numeric"
          pattern="[0-9]*"
          placeholder="000000"
          maxlength="6"
          class="join-input game-input"
          value={gamePin}
          on:input={handlePinInput}
          on:keydown={e => handleKeydown(e, handleGame)}
          aria-label="O'yin PIN kodi"
        />
        <div class="input-hint">6 ta raqam</div>
      </div>
      {#if pinError}
        <div class="error-msg">⚠ {pinError}</div>
      {/if}
      <button class="join-btn game-btn" on:click={handleGame} disabled={gamePin.length !== 6}>
        <span>Kirish</span>
        <span class="btn-arrow">→</span>
      </button>
    </div>
  </div>

  <!-- ── Card 2: Imtihon yechish ── -->
  <div class="join-card exam-card">
    <div class="card-top exam-top">
      <div class="card-dots" aria-hidden="true"></div>
      <div class="card-top-inner">
        <div class="card-emoji">📋</div>
        <div class="card-top-text">
          <h2>Imtihon yechish</h2>
          <p>Mustaqil imtihon topshiring</p>
        </div>
      </div>
    </div>
    <div class="card-body">
      <p class="card-desc">
        Imtihon kodini kiriting va o'z vaqtingizda mustaqil ravishda imtihon topshiring.
      </p>
      <div class="input-wrap" class:has-error={!!examError}>
        <input
          type="text"
          placeholder="ABCD1234"
          maxlength="8"
          class="join-input exam-input"
          value={examCode}
          on:input={e => handleCodeInput('exam', e)}
          on:keydown={e => handleKeydown(e, handleExam)}
          aria-label="Imtihon kodi"
          autocapitalize="characters"
          spellcheck="false"
        />
        <div class="input-hint">8 ta belgi</div>
      </div>
      {#if examError}
        <div class="error-msg">⚠ {examError}</div>
      {/if}
      <button class="join-btn exam-btn" on:click={handleExam} disabled={examCode.length !== 8}>
        <span>Boshlash</span>
        <span class="btn-arrow">→</span>
      </button>
    </div>
  </div>

  <!-- ── Card 3: Olimpiadaga kirish ── -->
  <div class="join-card olympiad-card">
    <div class="card-top olympiad-top">
      <div class="card-dots" aria-hidden="true"></div>
      <div class="card-top-inner">
        <div class="card-emoji">🏆</div>
        <div class="card-top-text">
          <h2>Olimpiadaga kirish</h2>
          <p>Musobaqa va tanlovlarda ishtirok eting</p>
        </div>
      </div>
    </div>
    <div class="card-body">
      <p class="card-desc">
        Olimpiada kodini kiriting va boshqa o'quvchilar bilan musobaqada ishtirok eting.
      </p>
      <div class="input-wrap" class:has-error={!!olympiadError}>
        <input
          type="text"
          placeholder="ABCD1234"
          maxlength="8"
          class="join-input olympiad-input"
          value={olympiadCode}
          on:input={e => handleCodeInput('olympiad', e)}
          on:keydown={e => handleKeydown(e, handleOlympiad)}
          aria-label="Olimpiada kodi"
          autocapitalize="characters"
          spellcheck="false"
        />
        <div class="input-hint">8 ta belgi</div>
      </div>
      {#if olympiadError}
        <div class="error-msg">⚠ {olympiadError}</div>
      {/if}
      <button class="join-btn olympiad-btn" on:click={handleOlympiad} disabled={olympiadCode.length !== 8}>
        <span>Kirish</span>
        <span class="btn-arrow">→</span>
      </button>
    </div>
  </div>

</div>

<!-- ══════════════════════════════════════════════════
     RECENT ENTRIES
══════════════════════════════════════════════════ -->
{#if recentEntries.length > 0}
  <div class="section">
    <div class="section-header">
      <h3 class="section-title">Oxirgi faollik</h3>
      <a href="/dashboard/results" class="section-link">Barchasini ko'rish →</a>
    </div>
    <div class="recent-list">
      {#each recentEntries as entry}
        <div class="recent-item">
          <div class="recent-icon-wrap" class:game-bg={entry.type === 'game'} class:exam-bg={entry.type === 'exam'} class:olympiad-bg={entry.type === 'olympiad'}>
            <span class="recent-icon">{typeIcon[entry.type]}</span>
          </div>
          <div class="recent-info">
            <div class="recent-title">{entry.title}</div>
            <div class="recent-meta">
              <span class="recent-type-badge" class:badge-game={entry.type === 'game'} class:badge-exam={entry.type === 'exam'} class:badge-olympiad={entry.type === 'olympiad'}>
                {typeLabel[entry.type]}
              </span>
              <span class="recent-date">{fmtDate(entry.date)}</span>
              <span class="recent-score">{entry.percent}%</span>
            </div>
          </div>
          <button class="continue-btn" on:click={() => continueEntry(entry)}>
            Davom ettirish →
          </button>
        </div>
      {/each}
    </div>
  </div>
{:else}
  <div class="section">
    <div class="section-header">
      <h3 class="section-title">Faol o'yinlar / testlar</h3>
    </div>
    <div class="empty-recent">
      <span class="empty-emoji">🎯</span>
      <p>Hali hech qanday faollik yo'q</p>
      <span class="empty-hint">Yuqoridagi kartochkalardan birini tanlang va boshlang!</span>
    </div>
  </div>
{/if}

<style>
  /* ── Page header ── */
  .page-header {
    margin-bottom: 28px;
  }
  h1 {
    font-size: 1.75rem;
    font-weight: 800;
    color: var(--text);
    margin: 0;
  }
  .sub {
    font-size: 0.875rem;
    color: var(--text3);
    margin-top: 4px;
  }

  /* ── Cards grid ── */
  .cards-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
    margin-bottom: 36px;
  }

  /* ── Join card base ── */
  .join-card {
    background: var(--white);
    border-radius: var(--radius-lg);
    border: 1.5px solid var(--border);
    box-shadow: var(--shadow-sm);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    transition: transform 0.22s ease, box-shadow 0.22s ease;
    animation: fadeSlideUp 0.35s ease both;
  }
  .join-card:hover {
    transform: translateY(-4px);
    box-shadow: var(--shadow);
  }

  /* ── Card top gradient ── */
  .card-top {
    padding: 28px 24px 24px;
    position: relative;
    overflow: hidden;
  }
  .card-dots {
    position: absolute;
    inset: 0;
    background-image: radial-gradient(circle, rgba(255,255,255,0.18) 1px, transparent 1px);
    background-size: 20px 20px;
    pointer-events: none;
  }
  .card-top-inner {
    position: relative;
    display: flex;
    align-items: center;
    gap: 16px;
  }
  .card-emoji {
    font-size: 2.8rem;
    line-height: 1;
    filter: drop-shadow(0 4px 8px rgba(0,0,0,0.15));
    animation: float 3s ease-in-out infinite;
    flex-shrink: 0;
  }
  @keyframes float {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-6px); }
  }
  .card-top-text h2 {
    font-size: 1.1rem;
    font-weight: 800;
    color: #fff;
    margin: 0 0 4px;
    line-height: 1.2;
  }
  .card-top-text p {
    font-size: 0.78rem;
    color: rgba(255,255,255,0.8);
    margin: 0;
  }

  /* ── Color themes for card tops ── */
  .game-top {
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  }
  .exam-top {
    background: linear-gradient(135deg, #f59e0b 0%, #f97316 100%);
  }
  .olympiad-top {
    background: linear-gradient(135deg, #eab308 0%, #ca8a04 100%);
  }

  /* ── Card body ── */
  .card-body {
    padding: 22px 24px;
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .card-desc {
    font-size: 0.82rem;
    color: var(--text3);
    line-height: 1.6;
    margin: 0;
  }

  /* ── Input ── */
  .input-wrap {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .input-wrap.has-error .join-input {
    border-color: var(--danger);
    box-shadow: 0 0 0 3px rgba(239,68,68,0.1);
  }
  .join-input {
    width: 100%;
    padding: 14px 16px;
    font-size: 1.5rem;
    font-weight: 800;
    text-align: center;
    letter-spacing: 0.2em;
    border: 2px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg);
    color: var(--text);
    outline: none;
    transition: border-color 0.2s, box-shadow 0.2s;
    font-family: 'Courier New', monospace;
    text-transform: uppercase;
  }
  .join-input::placeholder {
    color: var(--text3);
    font-weight: 400;
    letter-spacing: 0.3em;
  }
  .game-input:focus { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,0.15); }
  .exam-input:focus { border-color: #f59e0b; box-shadow: 0 0 0 3px rgba(245,158,11,0.15); }
  .olympiad-input:focus { border-color: #eab308; box-shadow: 0 0 0 3px rgba(234,179,8,0.15); }
  .input-hint {
    font-size: 0.72rem;
    color: var(--text3);
    text-align: center;
    font-weight: 600;
  }

  /* ── Error ── */
  .error-msg {
    font-size: 0.78rem;
    color: var(--danger);
    font-weight: 600;
    text-align: center;
  }

  /* ── Join button ── */
  .join-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    width: 100%;
    padding: 13px 20px;
    border: none;
    border-radius: var(--radius);
    font-size: 0.95rem;
    font-weight: 800;
    color: #fff;
    cursor: pointer;
    transition: var(--transition);
    letter-spacing: 0.3px;
    margin-top: auto;
  }
  .join-btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
    transform: none !important;
    box-shadow: none !important;
  }
  .join-btn:not(:disabled):hover { transform: translateY(-2px); }
  .join-btn:not(:disabled):active { transform: translateY(0); }

  .game-btn {
    background: linear-gradient(135deg, #6366f1, #8b5cf6);
    box-shadow: 0 4px 14px rgba(99,102,241,0.35);
  }
  .game-btn:not(:disabled):hover { box-shadow: 0 6px 20px rgba(99,102,241,0.5); }

  .exam-btn {
    background: linear-gradient(135deg, #f59e0b, #f97316);
    box-shadow: 0 4px 14px rgba(245,158,11,0.35);
  }
  .exam-btn:not(:disabled):hover { box-shadow: 0 6px 20px rgba(245,158,11,0.5); }

  .olympiad-btn {
    background: linear-gradient(135deg, #eab308, #ca8a04);
    box-shadow: 0 4px 14px rgba(234,179,8,0.35);
  }
  .olympiad-btn:not(:disabled):hover { box-shadow: 0 6px 20px rgba(234,179,8,0.5); }

  .btn-arrow {
    font-size: 1.1rem;
    transition: transform 0.2s;
  }
  .join-btn:not(:disabled):hover .btn-arrow { transform: translateX(4px); }

  /* ══════════════════════════════════════════════════
     RECENT ENTRIES SECTION
  ══════════════════════════════════════════════════ */
  .section {
    animation: fadeSlideUp 0.4s ease 0.15s both;
  }
  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
  }
  .section-title {
    font-size: 1rem;
    font-weight: 800;
    color: var(--text);
    margin: 0;
  }
  .section-link {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--primary);
    text-decoration: none;
    transition: var(--transition);
  }
  .section-link:hover { text-decoration: underline; }

  /* ── Recent list ── */
  .recent-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .recent-item {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 14px 18px;
    background: var(--white);
    border-radius: var(--radius);
    border: 1.5px solid var(--border);
    box-shadow: var(--shadow-sm);
    transition: var(--transition);
  }
  .recent-item:hover {
    border-color: var(--primary);
    box-shadow: var(--shadow);
  }
  .recent-icon-wrap {
    width: 44px;
    height: 44px;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.3rem;
    flex-shrink: 0;
  }
  .game-bg     { background: rgba(99,102,241,0.12); }
  .exam-bg     { background: rgba(245,158,11,0.12); }
  .olympiad-bg { background: rgba(234,179,8,0.12); }

  .recent-icon { line-height: 1; }
  .recent-info { flex: 1; min-width: 0; }
  .recent-title {
    font-size: 0.9rem;
    font-weight: 700;
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .recent-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 4px;
    flex-wrap: wrap;
  }
  .recent-type-badge {
    font-size: 0.68rem;
    font-weight: 700;
    padding: 2px 8px;
    border-radius: 99px;
  }
  .badge-game     { background: rgba(99,102,241,0.12); color: #6366f1; }
  .badge-exam     { background: rgba(245,158,11,0.12); color: #d97706; }
  .badge-olympiad { background: rgba(234,179,8,0.12);  color: #ca8a04; }
  .recent-date, .recent-score {
    font-size: 0.75rem;
    color: var(--text3);
    font-weight: 500;
  }
  .recent-score { color: var(--success); font-weight: 700; }
  .continue-btn {
    padding: 8px 16px;
    background: var(--primary-light);
    color: var(--primary);
    border: 1.5px solid var(--primary);
    border-radius: var(--radius-sm);
    font-size: 0.78rem;
    font-weight: 700;
    cursor: pointer;
    white-space: nowrap;
    transition: var(--transition);
    flex-shrink: 0;
  }
  .continue-btn:hover {
    background: var(--primary);
    color: #fff;
    box-shadow: 0 3px 10px rgba(99,102,241,0.3);
  }

  /* ── Empty recent ── */
  .empty-recent {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 40px 24px;
    background: var(--white);
    border-radius: var(--radius-lg);
    border: 1.5px dashed var(--border);
    text-align: center;
  }
  .empty-emoji { font-size: 2.5rem; }
  .empty-recent p { font-size: 0.9rem; color: var(--text2); font-weight: 600; margin: 0; }
  .empty-hint { font-size: 0.8rem; color: var(--text3); }

  /* ── Responsive ── */
  @media (max-width: 1024px) {
    .cards-grid { grid-template-columns: 1fr 1fr; }
  }
  @media (max-width: 640px) {
    h1 { font-size: 1.4rem; }
    .cards-grid { grid-template-columns: 1fr; }
    .card-top { padding: 22px 20px 20px; }
    .card-body { padding: 18px 20px; }
    .join-input { font-size: 1.3rem; padding: 12px 14px; }
    .recent-item { flex-wrap: wrap; }
    .continue-btn { width: 100%; text-align: center; justify-content: center; }
  }

  @keyframes fadeSlideUp {
    from { opacity: 0; transform: translateY(16px); }
    to   { opacity: 1; transform: translateY(0); }
  }
</style>
