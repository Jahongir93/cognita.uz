<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { quizzes as quizzesApi, rooms as roomsApi } from '$lib/api/client';
  import { authStore } from '$lib/stores/auth';
  import type { Quiz } from '$lib/api/types';
  import type { RoomHistory } from '$lib/api/client';

  $: role = $authStore.user?.role ?? 'student';

  let quizzes: Quiz[] = [];
  let history: RoomHistory[] = [];
  let loading = true;

  // ── Student-specific state ────────────────────────────────────────────────
  let pinInput = '';
  let codeInput = '';

  interface StudentResult {
    type: 'game' | 'exam';
    title: string;
    score: number;
    date: string;
  }
  let studentResults: StudentResult[] = [];
  let studentStats = { games: 0, exams: 0, avg: 0 };

  onMount(async () => {
    // Load student localStorage data
    try {
      const raw = localStorage.getItem('gogame_results');
      if (raw) {
        const all: StudentResult[] = JSON.parse(raw);
        studentResults = all.slice(-5).reverse();
        studentStats.games = all.filter(r => r.type === 'game').length;
        studentStats.exams = all.filter(r => r.type === 'exam').length;
        studentStats.avg = all.length
          ? Math.round(all.reduce((s, r) => s + r.score, 0) / all.length)
          : 0;
      }
    } catch {}

    if (role === 'student') {
      loading = false;
      return;
    }
    try {
      [quizzes, history] = await Promise.all([
        quizzesApi.list().catch(() => [] as Quiz[]),
        roomsApi.history().catch(() => [] as RoomHistory[])
      ]);
    } finally {
      loading = false;
    }
  });

  function joinGame() {
    const p = pinInput.trim();
    if (p) goto(`/join?pin=${encodeURIComponent(p)}`);
  }

  function startExam() {
    const c = codeInput.trim();
    if (c) goto(`/exam/${encodeURIComponent(c)}`);
  }

  // ── Computed stats ─────────────────────────────────────────────────────────
  $: totalPlays      = quizzes.reduce((s, q) => s + q.play_count, 0);
  $: totalQuestions  = quizzes.reduce((s, q) => s + q.total_questions, 0);
  $: recentHistory   = history.slice(0, 5);
  $: completedSessions = history.filter(h => h.status === 'completed').length;
  $: topQuizzes      = [...quizzes].sort((a, b) => b.play_count - a.play_count).slice(0, 5);

  // ── Greeting ───────────────────────────────────────────────────────────────
  function greeting(): string {
    const h = new Date().getHours();
    if (h < 12) return 'Xayrli tong';
    if (h < 17) return 'Xayrli kun';
    return 'Xayrli kech';
  }

  // ── Date helpers ───────────────────────────────────────────────────────────
  function fmtDate(d: string): string {
    if (!d) return '—';
    return new Date(d).toLocaleDateString('uz-UZ');
  }

  // ── Mode badge colours ─────────────────────────────────────────────────────
  const modeMeta: Record<string, { label: string; color: string; bg: string }> = {
    classic:     { label: 'Classic',  color: '#92400e', bg: '#fef3c7' },
    team:        { label: 'Jamoa',    color: '#1e3a8a', bg: '#dbeafe' },
    accuracy:    { label: 'Aniqlik', color: '#14532d', bg: '#dcfce7' },
    confidence:  { label: 'Ishonch', color: '#4c1d95', bg: '#ede9fe' },
    zero_stakes: { label: 'Mashq',   color: '#374151', bg: '#f3f4f6' },
  };
  function modeBadge(mode: string) {
    return modeMeta[mode] ?? { label: mode, color: '#374151', bg: '#f3f4f6' };
  }

  // ── Status badge ──────────────────────────────────────────────────────────
  const statusMeta: Record<string, { label: string; color: string; bg: string }> = {
    completed:   { label: 'Tugadi',   color: '#14532d', bg: '#dcfce7' },
    in_progress: { label: 'Davom etmoqda', color: '#1e3a8a', bg: '#dbeafe' },
    abandoned:   { label: 'To\'xtatildi', color: '#7f1d1d', bg: '#fee2e2' },
    waiting:     { label: 'Kutmoqda', color: '#78350f', bg: '#fef3c7' },
  };
  function statusBadge(status: string) {
    return statusMeta[status] ?? { label: status, color: '#374151', bg: '#f3f4f6' };
  }

  // ── Quick actions (admin) ─────────────────────────────────────────────────
  const quickActions = [
    {
      icon: '📝',
      label: 'Yangi quiz',
      sub: 'Savol yaratish',
      href: '/dashboard/quizzes/new',
      grad: 'linear-gradient(135deg, #6366f1, #8b5cf6)',
      shadow: 'rgba(99,102,241,0.35)',
    },
    {
      icon: '🎮',
      label: 'Viktorina boshlash',
      sub: "Jonli o'yin",
      href: '/dashboard/viktorina',
      grad: 'linear-gradient(135deg, #8b5cf6, #a855f7)',
      shadow: 'rgba(139,92,246,0.35)',
    },
    {
      icon: '📋',
      label: 'Imtihon yaratish',
      sub: 'Mustaqil topshiriq',
      href: '/dashboard/exam',
      grad: 'linear-gradient(135deg, #f59e0b, #f97316)',
      shadow: 'rgba(245,158,11,0.35)',
    },
    {
      icon: '🏆',
      label: 'Olimpiada',
      sub: 'Tanlov musobaqasi',
      href: '/dashboard/olympiad',
      grad: 'linear-gradient(135deg, #eab308, #f59e0b)',
      shadow: 'rgba(234,179,8,0.35)',
    },
  ];

  // ── Quick actions (teacher) ───────────────────────────────────────────────
  const teacherQuickActions = [
    {
      icon: '📝',
      label: 'Yangi quiz yaratish',
      sub: 'Savol yaratish',
      href: '/dashboard/quizzes/new',
      grad: 'linear-gradient(135deg, #6366f1, #8b5cf6)',
      shadow: 'rgba(99,102,241,0.35)',
    },
    {
      icon: '🎮',
      label: 'Viktorina boshlash',
      sub: "Jonli o'yin",
      href: '/dashboard/viktorina',
      grad: 'linear-gradient(135deg, #8b5cf6, #a855f7)',
      shadow: 'rgba(139,92,246,0.35)',
    },
    {
      icon: '📋',
      label: 'Imtihon yaratish',
      sub: 'Mustaqil topshiriq',
      href: '/dashboard/exam',
      grad: 'linear-gradient(135deg, #f59e0b, #f97316)',
      shadow: 'rgba(245,158,11,0.35)',
    },
    {
      icon: '🔍',
      label: 'Discovery',
      sub: "Ommaviy quizlar",
      href: '/dashboard/discovery',
      grad: 'linear-gradient(135deg, #22c55e, #16a34a)',
      shadow: 'rgba(34,197,94,0.35)',
    },
  ];

  // ── Activity chart (last 7 days) ──────────────────────────────────────────
  function buildWeeklyChart(hist: RoomHistory[]) {
    const days: { label: string; count: number; date: string }[] = [];
    const dayNames = ['Yak', 'Du', 'Se', 'Cho', 'Pay', 'Ju', 'Sha'];
    for (let i = 6; i >= 0; i--) {
      const d = new Date();
      d.setDate(d.getDate() - i);
      const dateStr = d.toISOString().slice(0, 10);
      const count = hist.filter(h => h.created_at && h.created_at.startsWith(dateStr)).length;
      days.push({ label: dayNames[d.getDay()], count, date: dateStr });
    }
    return days;
  }

  $: chartDays = buildWeeklyChart(history);
  $: chartMax  = Math.max(...chartDays.map(d => d.count), 1);

  // ── Top-quiz colour dots ───────────────────────────────────────────────────
  const dotColors = ['#6366f1', '#8b5cf6', '#f59e0b', '#22c55e', '#ef4444'];
</script>

<svelte:head><title>Dashboard — Cognita.uz</title></svelte:head>

{#if role === 'student'}
<!-- ══════════════════════════════════════════════════════
     STUDENT VIEW
══════════════════════════════════════════════════════ -->

  <!-- Welcome card -->
  <div class="student-welcome">
    <div class="banner-dots" aria-hidden="true"></div>
    <div class="banner-body">
      <div class="banner-text">
        <p class="banner-greeting">
          Xush kelibsiz, <strong>{$authStore.user?.full_name ?? "O'quvchi"}</strong>! 🎓
        </p>
        <p class="banner-sub">Bugun ham yangi bilimlar o'rganamiz!</p>
      </div>
      <div class="banner-deco" aria-hidden="true">🎓</div>
    </div>
  </div>

  <!-- Student stats row -->
  <div class="student-stats-row">
    <div class="sstat-card">
      <span class="sstat-icon">🎮</span>
      <span class="sstat-v">{studentStats.games}</span>
      <span class="sstat-l">O'ynalgan o'yinlar</span>
    </div>
    <div class="sstat-card">
      <span class="sstat-icon">📋</span>
      <span class="sstat-v">{studentStats.exams}</span>
      <span class="sstat-l">Topshirilgan testlar</span>
    </div>
    <div class="sstat-card">
      <span class="sstat-icon">⭐</span>
      <span class="sstat-v">{studentStats.avg}%</span>
      <span class="sstat-l">O'rtacha ball</span>
    </div>
  </div>

  <!-- Two primary action cards -->
  <div class="student-actions">

    <!-- PIN: join game -->
    <div class="student-action-card">
      <div class="sa-icon-wrap" style="background:linear-gradient(135deg,#6366f1,#8b5cf6)">
        🎮
      </div>
      <h3 class="sa-title">O'yinga kirish</h3>
      <p class="sa-desc">6-xonali PIN kodni kiriting</p>
      <form class="sa-form" on:submit|preventDefault={joinGame}>
        <input
          class="sa-input"
          type="text"
          inputmode="numeric"
          maxlength="6"
          placeholder="123456"
          bind:value={pinInput}
        />
        <button class="sa-btn sa-btn-primary" type="submit" disabled={!pinInput.trim()}>
          Kirish →
        </button>
      </form>
    </div>

    <!-- Code: start exam -->
    <div class="student-action-card">
      <div class="sa-icon-wrap" style="background:linear-gradient(135deg,#f59e0b,#f97316)">
        📋
      </div>
      <h3 class="sa-title">Test yechish</h3>
      <p class="sa-desc">O'qituvchingiz bergan kodni kiriting</p>
      <form class="sa-form" on:submit|preventDefault={startExam}>
        <input
          class="sa-input"
          type="text"
          placeholder="Kodni kiriting"
          bind:value={codeInput}
        />
        <button class="sa-btn sa-btn-amber" type="submit" disabled={!codeInput.trim()}>
          Boshlash →
        </button>
      </form>
    </div>

  </div>

  <!-- Recent activity -->
  <div class="card" style="margin-top:0">
    <div class="card-head">
      <h2 class="card-title-text">So'nggi faollik</h2>
    </div>
    {#if studentResults.length === 0}
      <div class="empty-state">
        <span class="empty-icon">📊</span>
        <p>Hali hech qanday faollik yo'q. O'yin yoki test boshlang!</p>
      </div>
    {:else}
      <ul class="activity-list">
        {#each studentResults as r}
          <li class="activity-row">
            <span class="act-icon">{r.type === 'game' ? '🎮' : '📋'}</span>
            <span class="act-title">{r.title}</span>
            <span class="act-score">{r.score}%</span>
            <span class="act-date">{new Date(r.date).toLocaleDateString('uz-UZ')}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </div>

{:else}
<!-- ══════════════════════════════════════════════════════
     ADMIN / TEACHER VIEW
══════════════════════════════════════════════════════ -->

  <!-- ── 1. WELCOME BANNER ── -->
  <div class="welcome-banner">
    <div class="banner-dots" aria-hidden="true"></div>
    <div class="banner-body">
      <div class="banner-text">
        {#if role === 'teacher'}
          <p class="banner-greeting">
            Xush kelibsiz, <strong>{$authStore.user?.full_name ?? "O'qituvchi"}</strong>! 📚
          </p>
        {:else}
          <p class="banner-greeting">
            {greeting()}, <strong>{$authStore.user?.full_name ?? "O'qituvchi"}</strong>! 👋
          </p>
        {/if}
        <p class="banner-sub">Bugun nima o'rganamiz?</p>

        <!-- Mini stats row -->
        <div class="banner-stats">
          <div class="bstat">
            <span class="bstat-v">{loading ? '…' : quizzes.length}</span>
            <span class="bstat-l">Quizlar</span>
          </div>
          <div class="bstat-div"></div>
          <div class="bstat">
            <span class="bstat-v">{loading ? '…' : totalQuestions}</span>
            <span class="bstat-l">Savollar</span>
          </div>
          <div class="bstat-div"></div>
          <div class="bstat">
            <span class="bstat-v">{loading ? '…' : totalPlays}</span>
            <span class="bstat-l">O'yinlar</span>
          </div>
          <div class="bstat-div"></div>
          <div class="bstat">
            <span class="bstat-v">{loading ? '…' : completedSessions}</span>
            <span class="bstat-l">Sessiyalar</span>
          </div>
        </div>
      </div>

      <div class="banner-deco" aria-hidden="true">🎮</div>
    </div>
  </div>

  <!-- ── 2. QUICK ACTIONS ── -->
  <div class="section-title">Tezkor amallar</div>
  <div class="quick-actions">
    {#each (role === 'teacher' ? teacherQuickActions : quickActions) as qa}
      <a href={qa.href} class="qa-card" style="--qa-grad:{qa.grad};--qa-shadow:{qa.shadow}">
        <div class="qa-icon-wrap">
          <span class="qa-icon">{qa.icon}</span>
        </div>
        <div class="qa-text">
          <span class="qa-label">{qa.label}</span>
          <span class="qa-sub">{qa.sub}</span>
        </div>
        <span class="qa-arrow">→</span>
      </a>
    {/each}
  </div>

  {#if role === 'teacher'}
    <div class="teacher-discovery-link">
      <a href="/dashboard/discovery" class="discovery-link-card">
        🔍 Ommaviy quizlarni ko'rish — Discovery →
      </a>
    </div>
  {/if}

  <!-- ── 3. TWO-COLUMN SECTION ── -->
  <div class="two-col">

    <!-- LEFT: Recent sessions -->
    <div class="col-main card">
      <div class="card-head">
        <h2 class="card-title-text">So'nggi o'yinlar</h2>
        <a href="/dashboard/reports" class="card-link">Barchasi →</a>
      </div>

      {#if loading}
        <div class="table-wrap">
          {#each Array(4) as _}
            <div class="skeleton sk-row"></div>
          {/each}
        </div>
      {:else if recentHistory.length === 0}
        <div class="empty-state">
          <span class="empty-icon">🎮</span>
          <p>Hali o'yin o'tkazilmagan</p>
          <a href="/dashboard/viktorina" class="empty-link">Birinchi o'yinni boshlash →</a>
        </div>
      {:else}
        <div class="table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>Quiz nomi</th>
                <th>Rejim</th>
                <th>O'yinchilar</th>
                <th>Sana</th>
                <th>Holat</th>
              </tr>
            </thead>
            <tbody>
              {#each recentHistory as h}
                {@const mb = modeBadge(h.game_mode)}
                {@const sb = statusBadge(h.status)}
                <tr class="tr-hover">
                  <td class="td-name">{h.quiz_title}</td>
                  <td>
                    <span class="badge" style="color:{mb.color};background:{mb.bg}">{mb.label}</span>
                  </td>
                  <td class="td-num">{h.player_count}</td>
                  <td class="td-date">{fmtDate(h.created_at)}</td>
                  <td>
                    <span class="badge" style="color:{sb.color};background:{sb.bg}">{sb.label}</span>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>

    <!-- RIGHT: My quizzes mini-list -->
    <div class="col-side card">
      <div class="card-head">
        <h2 class="card-title-text">Mening quizlarim</h2>
        <a href="/dashboard/quizzes" class="card-link">Barchasi →</a>
      </div>

      {#if loading}
        {#each Array(4) as _}
          <div class="skeleton sk-qrow"></div>
        {/each}
      {:else if topQuizzes.length === 0}
        <div class="empty-state">
          <span class="empty-icon">📝</span>
          <p>Hali quiz yo'q</p>
        </div>
      {:else}
        <ul class="quiz-list">
          {#each topQuizzes as q, i}
            <li class="quiz-row">
              <span class="qdot" style="background:{dotColors[i % dotColors.length]}"></span>
              <span class="qname">{q.title}</span>
              <span class="qplays">{q.play_count} <small>o'yin</small></span>
            </li>
          {/each}
        </ul>
      {/if}

      <a href="/dashboard/quizzes/new" class="new-quiz-btn">
        <span>+</span> Yangi quiz
      </a>
    </div>

  </div>

  <!-- ── 4. ACTIVITY CHART ── -->
  <div class="card chart-card">
    <div class="card-head">
      <h2 class="card-title-text">Haftalik faollik</h2>
      <span class="chart-sub">Oxirgi 7 kun</span>
    </div>

    <div class="chart-area">
      {#each chartDays as day, i}
        {@const pct = chartMax > 0 ? (day.count / chartMax) * 100 : 0}
        <div class="chart-col">
          <div class="bar-wrap">
            <div
              class="bar"
              style="height:{Math.max(pct, day.count > 0 ? 6 : 2)}%;animation-delay:{i * 60}ms"
              title="{day.date}: {day.count} o'yin"
            >
              {#if day.count > 0}
                <span class="bar-val">{day.count}</span>
              {/if}
            </div>
          </div>
          <span class="bar-label">{day.label}</span>
        </div>
      {/each}
    </div>
  </div>

{/if}

<style>
  /* ── Skeletons ── */
  .skeleton {
    background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
    background-size: 200% 100%;
    animation: shimmer 1.4s infinite;
    border-radius: 8px;
  }
  @keyframes shimmer {
    0%   { background-position: 200% 0; }
    100% { background-position: -200% 0; }
  }
  .sk-row  { height: 42px; margin-bottom: 8px; }
  .sk-qrow { height: 36px; margin-bottom: 8px; border-radius: 8px; }

  /* ── Section title ── */
  .section-title {
    font-size: 0.78rem;
    font-weight: 700;
    letter-spacing: 0.5px;
    color: var(--text3);
    text-transform: uppercase;
    margin: 0 0 10px;
  }

  /* ════════════════════════════════════════
     STUDENT VIEW
  ════════════════════════════════════════ */
  .student-welcome {
    position: relative;
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 60%, #a855f7 100%);
    border-radius: var(--radius-lg);
    overflow: hidden;
    margin-bottom: 20px;
    box-shadow: 0 8px 30px rgba(99, 102, 241, 0.3);
  }

  .student-stats-row {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 14px;
    margin-bottom: 20px;
  }

  .sstat-card {
    background: var(--white);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow-sm);
    padding: 18px 20px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    text-align: center;
  }

  .sstat-icon { font-size: 1.6rem; line-height: 1; }
  .sstat-v {
    font-size: 1.6rem;
    font-weight: 900;
    color: var(--text);
    line-height: 1;
  }
  .sstat-l {
    font-size: 0.72rem;
    font-weight: 600;
    color: var(--text3);
  }

  .student-actions {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 18px;
    margin-bottom: 20px;
  }

  .student-action-card {
    background: var(--white);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
    padding: 24px 22px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .sa-icon-wrap {
    width: 52px;
    height: 52px;
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.6rem;
    box-shadow: 0 4px 14px rgba(0,0,0,0.15);
  }

  .sa-title {
    font-size: 1.1rem;
    font-weight: 800;
    color: var(--text);
    margin: 0;
  }

  .sa-desc {
    font-size: 0.82rem;
    color: var(--text3);
    margin: 0;
  }

  .sa-form {
    display: flex;
    gap: 8px;
    margin-top: 4px;
  }

  .sa-input {
    flex: 1;
    padding: 10px 14px;
    border: 1.5px solid var(--border);
    border-radius: 10px;
    font-size: 1rem;
    font-weight: 700;
    color: var(--text);
    background: var(--white);
    outline: none;
    transition: border-color 0.2s, box-shadow 0.2s;
    min-width: 0;
    letter-spacing: 0.5px;
  }

  .sa-input:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
  }

  .sa-btn {
    padding: 10px 16px;
    border: none;
    border-radius: 10px;
    font-size: 0.875rem;
    font-weight: 700;
    cursor: pointer;
    white-space: nowrap;
    transition: var(--transition);
  }

  .sa-btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .sa-btn-primary {
    background: var(--primary);
    color: #fff;
    box-shadow: 0 2px 8px rgba(99,102,241,0.3);
  }
  .sa-btn-primary:hover:not(:disabled) {
    background: #4f46e5;
    box-shadow: 0 4px 14px rgba(99,102,241,0.4);
  }

  .sa-btn-amber {
    background: #f59e0b;
    color: #fff;
    box-shadow: 0 2px 8px rgba(245,158,11,0.3);
  }
  .sa-btn-amber:hover:not(:disabled) {
    background: #d97706;
    box-shadow: 0 4px 14px rgba(245,158,11,0.4);
  }

  .activity-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .activity-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 10px;
    border-radius: 8px;
    transition: background 0.15s;
  }

  .activity-row:hover { background: var(--bg); }

  .act-icon { font-size: 1.1rem; flex-shrink: 0; }
  .act-title {
    flex: 1;
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .act-score {
    font-size: 0.82rem;
    font-weight: 800;
    color: var(--primary);
    white-space: nowrap;
  }
  .act-date {
    font-size: 0.72rem;
    color: var(--text3);
    white-space: nowrap;
  }

  /* teacher discovery quick link */
  .teacher-discovery-link {
    margin-bottom: 20px;
  }

  .discovery-link-card {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 10px 18px;
    background: var(--white);
    border: 1.5px solid #22c55e;
    border-radius: var(--radius);
    color: #16a34a;
    font-size: 0.875rem;
    font-weight: 700;
    text-decoration: none;
    transition: var(--transition);
    box-shadow: var(--shadow-sm);
  }

  .discovery-link-card:hover {
    background: #f0fdf4;
    box-shadow: 0 4px 14px rgba(34,197,94,0.2);
  }

  /* ════════════════════════════════════════
     WELCOME BANNER
  ════════════════════════════════════════ */
  .welcome-banner {
    position: relative;
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
    border-radius: var(--radius-lg);
    overflow: hidden;
    margin-bottom: 28px;
    box-shadow: 0 8px 30px rgba(99, 102, 241, 0.3);
  }
  .banner-dots {
    position: absolute;
    inset: 0;
    background-image:
      radial-gradient(circle, rgba(255,255,255,0.15) 1px, transparent 1px);
    background-size: 22px 22px;
    pointer-events: none;
  }
  .banner-body {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 28px 32px;
    gap: 20px;
  }
  .banner-text { flex: 1; min-width: 0; }
  .banner-greeting {
    font-size: 1.45rem;
    font-weight: 700;
    color: #fff;
    margin: 0 0 4px;
    line-height: 1.3;
  }
  .banner-greeting strong { font-weight: 900; }
  .banner-sub {
    font-size: 0.95rem;
    color: rgba(255,255,255,0.75);
    margin: 0 0 20px;
  }

  /* mini stats */
  .banner-stats {
    display: flex;
    align-items: center;
    gap: 0;
    background: rgba(0,0,0,0.18);
    border-radius: 14px;
    padding: 12px 20px;
    width: fit-content;
    flex-wrap: wrap;
    gap: 0;
  }
  .bstat {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 0 18px;
  }
  .bstat:first-child { padding-left: 0; }
  .bstat:last-child  { padding-right: 0; }
  .bstat-v {
    font-size: 1.4rem;
    font-weight: 900;
    color: #fff;
    line-height: 1;
  }
  .bstat-l {
    font-size: 0.68rem;
    color: rgba(255,255,255,0.65);
    margin-top: 3px;
    font-weight: 600;
    white-space: nowrap;
  }
  .bstat-div {
    width: 1px;
    height: 32px;
    background: rgba(255,255,255,0.2);
    flex-shrink: 0;
  }

  /* decorative emoji */
  .banner-deco {
    font-size: 5rem;
    line-height: 1;
    opacity: 0.85;
    animation: float 3.5s ease-in-out infinite;
    flex-shrink: 0;
    user-select: none;
  }
  @keyframes float {
    0%, 100% { transform: translateY(0) rotate(-4deg); }
    50%       { transform: translateY(-10px) rotate(4deg); }
  }

  /* ════════════════════════════════════════
     QUICK ACTIONS
  ════════════════════════════════════════ */
  .quick-actions {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 14px;
    margin-bottom: 28px;
  }
  .qa-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px 18px;
    background: var(--white);
    border-radius: var(--radius);
    text-decoration: none;
    box-shadow: var(--shadow-sm);
    border: 1px solid var(--border);
    transition: transform 0.22s ease, box-shadow 0.22s ease;
    overflow: hidden;
    position: relative;
  }
  .qa-card::before {
    content: '';
    position: absolute;
    inset: 0;
    background: var(--qa-grad);
    opacity: 0;
    transition: opacity 0.22s;
  }
  .qa-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 12px 28px var(--qa-shadow);
    border-color: transparent;
  }
  .qa-card:hover::before { opacity: 1; }
  .qa-card:hover .qa-label { color: #fff; }
  .qa-card:hover .qa-sub   { color: rgba(255,255,255,0.75); }
  .qa-card:hover .qa-arrow { color: rgba(255,255,255,0.9); }
  .qa-card:active { transform: scale(0.97) translateY(-2px); }

  .qa-icon-wrap {
    width: 42px;
    height: 42px;
    background: var(--qa-grad);
    border-radius: 11px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.3rem;
    flex-shrink: 0;
    position: relative;
    box-shadow: 0 4px 12px var(--qa-shadow);
    transition: transform 0.22s;
  }
  .qa-card:hover .qa-icon-wrap { transform: scale(1.1) rotate(-5deg); }
  .qa-icon { position: relative; }
  .qa-text {
    flex: 1;
    min-width: 0;
    position: relative;
  }
  .qa-label {
    display: block;
    font-size: 0.875rem;
    font-weight: 700;
    color: var(--text);
    line-height: 1.2;
    transition: color 0.22s;
  }
  .qa-sub {
    display: block;
    font-size: 0.72rem;
    color: var(--text3);
    margin-top: 2px;
    transition: color 0.22s;
  }
  .qa-arrow {
    font-size: 1rem;
    color: var(--text3);
    position: relative;
    transition: color 0.22s, transform 0.22s;
    flex-shrink: 0;
  }
  .qa-card:hover .qa-arrow { transform: translateX(3px); }

  /* ════════════════════════════════════════
     TWO-COLUMN LAYOUT
  ════════════════════════════════════════ */
  .two-col {
    display: grid;
    grid-template-columns: 60fr 40fr;
    gap: 18px;
    margin-bottom: 22px;
    align-items: start;
  }

  /* ── Card base ── */
  .card {
    background: var(--white);
    border-radius: var(--radius);
    box-shadow: var(--shadow-sm);
    border: 1px solid var(--border);
    padding: 20px 22px;
  }
  .card-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    gap: 10px;
  }
  .card-title-text {
    font-size: 1rem;
    font-weight: 700;
    color: var(--text);
    margin: 0;
  }
  .card-link {
    font-size: 0.78rem;
    font-weight: 600;
    color: var(--primary);
    text-decoration: none;
    white-space: nowrap;
    transition: var(--transition);
  }
  .card-link:hover { text-decoration: underline; }

  /* ── Table ── */
  .table-wrap { overflow-x: auto; }
  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.83rem;
  }
  .data-table th {
    text-align: left;
    font-size: 0.7rem;
    font-weight: 700;
    letter-spacing: 0.4px;
    color: var(--text3);
    text-transform: uppercase;
    padding: 0 10px 10px;
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
  }
  .data-table th:first-child { padding-left: 0; }
  .data-table td {
    padding: 10px 10px;
    color: var(--text2);
    border-bottom: 1px solid #f8fafc;
    vertical-align: middle;
  }
  .data-table td:first-child { padding-left: 0; }
  .tr-hover { transition: background 0.15s; }
  .tr-hover:hover td { background: var(--bg); }
  .tr-hover:last-child td { border-bottom: none; }
  .td-name {
    font-weight: 600;
    color: var(--text);
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .td-num {
    text-align: center;
    font-weight: 700;
    color: var(--text);
  }
  .td-date { white-space: nowrap; font-size: 0.78rem; }

  .badge {
    display: inline-block;
    font-size: 0.7rem;
    font-weight: 700;
    padding: 3px 9px;
    border-radius: 99px;
    white-space: nowrap;
  }

  /* ── Empty state ── */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 32px 16px;
    text-align: center;
  }
  .empty-icon { font-size: 2.5rem; }
  .empty-state p { font-size: 0.88rem; color: var(--text3); margin: 0; }
  .empty-link {
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--primary);
    text-decoration: none;
    margin-top: 4px;
  }
  .empty-link:hover { text-decoration: underline; }

  /* ── Quiz list (right col) ── */
  .quiz-list {
    list-style: none;
    margin: 0 0 16px;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .quiz-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 10px;
    border-radius: 8px;
    transition: background 0.15s;
    cursor: default;
  }
  .quiz-row:hover { background: var(--bg); }
  .qdot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .qname {
    flex: 1;
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .qplays {
    font-size: 0.78rem;
    font-weight: 700;
    color: var(--primary);
    white-space: nowrap;
  }
  .qplays small { font-weight: 500; color: var(--text3); }

  .new-quiz-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    width: 100%;
    padding: 10px;
    background: var(--primary-light);
    color: var(--primary-dark);
    border: 1.5px dashed var(--primary);
    border-radius: 10px;
    font-size: 0.85rem;
    font-weight: 700;
    text-decoration: none;
    transition: var(--transition);
  }
  .new-quiz-btn:hover {
    background: var(--primary);
    color: #fff;
    border-style: solid;
    box-shadow: 0 4px 14px rgba(99,102,241,0.3);
  }

  /* ════════════════════════════════════════
     ACTIVITY CHART
  ════════════════════════════════════════ */
  .chart-card { margin-bottom: 8px; }
  .chart-sub {
    font-size: 0.75rem;
    color: var(--text3);
  }
  .chart-area {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 8px;
    height: 140px;
    padding-top: 16px;
  }
  .chart-col {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    height: 100%;
    gap: 6px;
  }
  .bar-wrap {
    flex: 1;
    width: 100%;
    display: flex;
    align-items: flex-end;
    justify-content: center;
  }
  .bar {
    width: 100%;
    max-width: 36px;
    background: linear-gradient(180deg, var(--primary) 0%, var(--accent) 100%);
    border-radius: 6px 6px 4px 4px;
    min-height: 4px;
    position: relative;
    animation: growUp 0.5s ease forwards;
    transform-origin: bottom;
    box-shadow: 0 2px 8px rgba(99,102,241,0.25);
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding-top: 4px;
  }
  @keyframes growUp {
    from { transform: scaleY(0); opacity: 0; }
    to   { transform: scaleY(1); opacity: 1; }
  }
  .bar-val {
    font-size: 0.6rem;
    font-weight: 800;
    color: rgba(255,255,255,0.9);
    line-height: 1;
  }
  .bar-label {
    font-size: 0.7rem;
    font-weight: 600;
    color: var(--text3);
    white-space: nowrap;
  }

  /* ════════════════════════════════════════
     RESPONSIVE
  ════════════════════════════════════════ */
  @media (max-width: 1100px) {
    .quick-actions { grid-template-columns: repeat(2, 1fr); }
  }

  @media (max-width: 900px) {
    .two-col { grid-template-columns: 1fr; }
  }

  @media (max-width: 640px) {
    .banner-body { padding: 22px 20px; }
    .banner-greeting { font-size: 1.15rem; }
    .banner-deco { font-size: 3.5rem; }
    .banner-stats { padding: 10px 14px; }
    .bstat { padding: 0 12px; }
    .bstat-v { font-size: 1.1rem; }
    .quick-actions { grid-template-columns: repeat(2, 1fr); gap: 10px; }
    .qa-card { padding: 13px 14px; gap: 10px; }
    .qa-icon-wrap { width: 36px; height: 36px; font-size: 1.1rem; }
    .chart-area { height: 110px; }
  }

  @media (max-width: 400px) {
    .banner-stats { flex-wrap: wrap; gap: 8px; }
    .bstat-div { display: none; }
    .bstat { padding: 0 8px; }
    .quick-actions { grid-template-columns: 1fr 1fr; }
  }

  @media (max-width: 640px) {
    .student-actions { grid-template-columns: 1fr; }
    .student-stats-row { grid-template-columns: 1fr 1fr 1fr; gap: 8px; }
    .sstat-card { padding: 14px 10px; }
    .sstat-v { font-size: 1.2rem; }
    .sa-form { flex-direction: column; }
    .sa-btn { width: 100%; }
  }

  @media (max-width: 400px) {
    .student-stats-row { grid-template-columns: 1fr; }
  }
</style>
