<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { authStore } from '$lib/stores/auth';
  import { onMount } from 'svelte';

  interface NavItem {
    href: string;
    icon: string;
    label: string;
    exact?: boolean;
    badge?: string;
  }

  let sidebarOpen = false;

  onMount(async () => {
    await authStore.loadUser();
    if (!$authStore.user) goto('/auth/login');
  });

  async function logout() {
    await authStore.logout();
    goto('/auth/login');
  }

  $: path = $page.url.pathname;

  const adminNavGroups: { label: string; items: NavItem[] }[] = [
    {
      label: 'ASOSIY',
      items: [
        { href: '/dashboard',            icon: '📊', label: 'Dashboard',    exact: true },
        { href: '/dashboard/quizzes',    icon: '📝', label: 'Quizlar' },
      ]
    },
    {
      label: "O'YINLAR",
      items: [
        { href: '/dashboard/viktorina',  icon: '🎮', label: 'Viktorina' },
        { href: '/dashboard/exam',       icon: '📋', label: 'Imtihon' },
        { href: '/dashboard/olympiad',   icon: '🏆', label: 'Olimpiada' },
      ]
    },
    {
      label: 'DOSKA',
      items: [
        { href: '/dashboard/topshiriqlar', icon: '🧩', label: 'Topshiriqlar', badge: 'Yangi' },
      ]
    },
    {
      label: 'BOSHQARUV',
      items: [
        { href: '/dashboard/open-tests', icon: '📂', label: 'Ochiq testlar', badge: 'Yangi' },
        { href: '/dashboard/categories', icon: '🗂️', label: 'Kategoriyalar' },
        { href: '/dashboard/classes',    icon: '👨‍🏫', label: 'Sinflar' },
        { href: '/dashboard/reports',    icon: '📈', label: 'Hisobotlar' },
      ]
    },
    {
      label: 'TIZIM',
      items: [
        { href: '/dashboard/settings',   icon: '⚙️', label: 'Sozlamalar' },
      ]
    },
  ];

  const teacherNavGroups: { label: string; items: NavItem[] }[] = [
    {
      label: 'ASOSIY',
      items: [
        { href: '/dashboard',            icon: '📊', label: 'Dashboard',    exact: true },
        { href: '/dashboard/quizzes',    icon: '📝', label: 'Quizlarim' },
      ]
    },
    {
      label: "O'YINLAR",
      items: [
        { href: '/dashboard/viktorina',  icon: '🎮', label: 'Viktorina' },
        { href: '/dashboard/exam',       icon: '📋', label: 'Imtihon' },
        { href: '/dashboard/olympiad',   icon: '🏆', label: 'Olimpiada' },
      ]
    },
    {
      label: 'DOSKA',
      items: [
        { href: '/dashboard/topshiriqlar', icon: '🧩', label: 'Topshiriqlar', badge: 'Yangi' },
      ]
    },
    {
      label: 'KASHFIYOT',
      items: [
        { href: '/dashboard/discovery',  icon: '🔍', label: 'Discovery',   badge: 'Yangi' },
      ]
    },
    {
      label: 'BOSHQARUV',
      items: [
        { href: '/dashboard/classes',    icon: '👨‍🏫', label: 'Sinflar' },
        { href: '/dashboard/reports',    icon: '📈', label: 'Hisobotlar' },
      ]
    },
    {
      label: 'TIZIM',
      items: [
        { href: '/dashboard/settings',   icon: '⚙️', label: 'Sozlamalar' },
      ]
    },
  ];

  const studentNavGroups: { label: string; items: NavItem[] }[] = [
    {
      label: 'ASOSIY',
      items: [
        { href: '/dashboard',            icon: '🏠', label: 'Bosh sahifa',  exact: true },
      ]
    },
    {
      label: "O'YINLAR",
      items: [
        { href: '/dashboard/join',       icon: '🎮', label: "O'yinga kirish" },
        { href: '/dashboard/games',      icon: '📊', label: "O'yinlarim" },
      ]
    },
    {
      label: 'TESTLAR',
      items: [
        { href: '/dashboard/my-exams',   icon: '📋', label: 'Imtihonlar' },
        { href: '/dashboard/my-olympiads', icon: '🏆', label: 'Olimpiadalar' },
      ]
    },
    {
      label: 'MENING',
      items: [
        { href: '/dashboard/results',    icon: '🏅', label: 'Natijalarim' },
        { href: '/dashboard/settings',   icon: '⚙️', label: 'Sozlamalar' },
      ]
    },
  ];

  $: navGroups = $authStore.user?.role === 'admin'
    ? adminNavGroups
    : $authStore.user?.role === 'teacher'
      ? teacherNavGroups
      : studentNavGroups;

  const allPageTitles: Record<string, string> = {
    '/dashboard':                'Dashboard',
    '/dashboard/quizzes':        'Quizlar',
    '/dashboard/viktorina':      'Viktorina',
    '/dashboard/exam':           'Imtihon rejimi',
    '/dashboard/olympiad':       'Olimpiada',
    '/dashboard/categories':     'Kategoriyalar',
    '/dashboard/classes':        'Sinflar',
    '/dashboard/reports':        'Hisobotlar',
    '/dashboard/settings':       'Sozlamalar',
    '/dashboard/discovery':      'Discovery',
    '/dashboard/join':           "O'yinga kirish",
    '/dashboard/games':          "O'yinlarim",
    '/dashboard/my-exams':       'Imtihonlarim',
    '/dashboard/my-olympiads':   'Olimpiadalarim',
    '/dashboard/results':        'Natijalarim',
  };

  $: pageTitle = allPageTitles[path]
    ?? allPageTitles[Object.keys(allPageTitles).find(k => path.startsWith(k) && k !== '/dashboard') ?? '']
    ?? 'Dashboard';

  function isActive(item: { href: string; exact?: boolean }) {
    return item.exact ? path === item.href : path.startsWith(item.href);
  }

  $: bottomNav = $authStore.user?.role === 'student'
    ? [
        { href: '/dashboard',          icon: '🏠', label: 'Bosh'     },
        { href: '/dashboard/join',     icon: '🎮', label: "Kirish"   },
        { href: '/dashboard/my-exams', icon: '📋', label: 'Test'     },
        { href: '/dashboard/results',  icon: '🏅', label: 'Natija'   },
        { href: '/dashboard/settings', icon: '⚙️', label: 'Sozlama'  },
      ]
    : $authStore.user?.role === 'teacher'
      ? [
          { href: '/dashboard',            icon: '📊', label: 'Bosh'    },
          { href: '/dashboard/quizzes',    icon: '📝', label: 'Quiz'    },
          { href: '/dashboard/viktorina',  icon: '🎮', label: "O'yin"   },
          { href: '/dashboard/discovery',  icon: '🔍', label: 'Kashf'   },
          { href: '/dashboard/reports',    icon: '📈', label: 'Hisobot' },
        ]
      : [
          { href: '/dashboard',            icon: '📊', label: 'Bosh'    },
          { href: '/dashboard/quizzes',    icon: '📝', label: 'Quiz'    },
          { href: '/dashboard/viktorina',  icon: '🎮', label: "O'yin"   },
          { href: '/dashboard/exam',       icon: '📋', label: 'Imtihon' },
          { href: '/dashboard/reports',    icon: '📈', label: 'Hisobot' },
        ];

  function closeSidebar() {
    sidebarOpen = false;
  }
</script>

<!-- ── Sidebar overlay (mobile) ── -->
{#if sidebarOpen}
  <div
    class="sidebar-overlay"
    role="button"
    tabindex="-1"
    aria-label="Yopish"
    on:click={closeSidebar}
    on:keydown={e => e.key === 'Escape' && closeSidebar()}
  ></div>
{/if}

<div class="layout">

  <!-- ════════════════ SIDEBAR ════════════════ -->
  <aside class="sidebar" class:open={sidebarOpen}>

    <!-- Logo -->
    <div class="sidebar-top">
      <a href="/" class="brand">
        <img src="/logowhite.png" alt="Cognita.uz" style="height:32px;width:auto;display:block" />
      </a>
    </div>

    <!-- Nav -->
    <nav class="nav">
      {#each navGroups as group}
        <div class="nav-group">
          <div class="nav-group-label">
            <span class="label-line"></span>
            <span class="label-text">{group.label}</span>
            <span class="label-line"></span>
          </div>
          {#each group.items as item}
            <a
              href={item.href}
              class="nav-item"
              class:active={isActive(item)}
              on:click={closeSidebar}
            >
              <span class="nav-icon">{item.icon}</span>
              <span class="nav-label">{item.label}</span>
              {#if item.badge}
                <span class="nav-badge">{item.badge}</span>
              {/if}
            </a>
          {/each}
        </div>
      {/each}
    </nav>

    <!-- User card + logout -->
    <div class="sidebar-bottom">
      {#if $authStore.user}
        <div class="user-card">
          <div class="user-avatar">
            {$authStore.user.full_name?.[0]?.toUpperCase() ?? '?'}
          </div>
          <div class="user-info">
            <div class="user-name">{$authStore.user.full_name}</div>
            <div class="user-role role-{$authStore.user.role}">
              {$authStore.user.role === 'admin' ? '⚡ Admin' : $authStore.user.role === 'teacher' ? '📚 O\'qituvchi' : '🎓 Talaba'}
            </div>
          </div>
        </div>
        <button class="logout-btn" on:click={logout}>
          <span class="logout-icon">↪</span> Chiqish
        </button>
      {/if}
    </div>

  </aside>
  <!-- ════════════════ /SIDEBAR ════════════════ -->

  <!-- ════════════════ MAIN AREA ════════════════ -->
  <div class="main-area">

    <!-- Top header -->
    <header class="top-header">
      <div class="header-left">
        <button
          class="hamburger"
          aria-label="Menyuni ochish"
          on:click={() => sidebarOpen = !sidebarOpen}
        >
          <span class="ham-line"></span>
          <span class="ham-line"></span>
          <span class="ham-line"></span>
        </button>
        <div class="header-title-wrap">
          <div class="header-breadcrumb">
            <span class="bc-root">Cognita.uz</span>
            <span class="bc-sep">›</span>
            <span class="bc-current">{pageTitle}</span>
          </div>
          <h1 class="header-title">{pageTitle}</h1>
        </div>
      </div>

      <div class="header-right">
        <button class="header-logo-center" aria-label="Cognita logo">
          <img src="/sitelogo.png" alt="Cognita.uz" style="height:28px;width:auto" />
        </button>
        <button class="notif-btn" aria-label="Bildirishnomalar">
          🔔
          <span class="notif-badge">3</span>
        </button>
        {#if $authStore.user}
          <button class="header-avatar" aria-label="Profil">
            {$authStore.user.full_name?.[0]?.toUpperCase() ?? '?'}
          </button>
        {/if}
      </div>
    </header>

    <!-- Page content -->
    <main class="content">
      <slot />
    </main>

  </div>
  <!-- ════════════════ /MAIN AREA ════════════════ -->

</div>

<!-- ════════════════ BOTTOM NAV (mobile) ════════════════ -->
<nav class="bottom-nav" aria-label="Alt navigatsiya">
  {#each bottomNav as item}
    <a
      href={item.href}
      class="bottom-nav-item"
      class:active={item.href === '/dashboard' ? path === '/dashboard' : path.startsWith(item.href)}
    >
      <span class="bnav-icon">{item.icon}</span>
      <span class="bnav-label">{item.label}</span>
    </a>
  {/each}
</nav>

<style>
  /* ── Reset / base ── */
  * { box-sizing: border-box; }

  /* ── Root layout ── */
  .layout {
    display: flex;
    height: 100vh;
    overflow: hidden;
    background: var(--bg);
  }

  /* ════════════════════════════════════════
     SIDEBAR
  ════════════════════════════════════════ */
  .sidebar {
    width: 240px;
    flex-shrink: 0;
    background: linear-gradient(180deg, #0a0f1e 0%, #0d1424 60%, #111827 100%);
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
    box-shadow: 4px 0 28px rgba(0, 0, 0, 0.3);
    position: relative;
    z-index: 50;
    transition: transform 0.28s cubic-bezier(0.4, 0, 0.2, 1);
  }

  /* ── Logo / brand ── */
  .sidebar-top {
    padding: 20px 18px 18px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
    flex-shrink: 0;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 11px;
    text-decoration: none;
  }
  .brand-icon {
    width: 38px;
    height: 38px;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.2rem;
    box-shadow: 0 4px 14px rgba(99, 102, 241, 0.45);
    flex-shrink: 0;
  }
  .brand-text { line-height: 1.2; }
  .brand-name {
    font-size: 1.1rem;
    font-weight: 800;
    color: #fff;
    display: block;
    letter-spacing: -0.3px;
  }
  .brand-dot {
    font-size: 0.73rem;
    color: var(--accent);
    font-weight: 600;
  }

  /* ── Nav scroll area ── */
  .nav {
    flex: 1;
    padding: 10px 10px 6px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    overflow-y: auto;
    scrollbar-width: thin;
    scrollbar-color: rgba(255,255,255,0.08) transparent;
  }
  .nav::-webkit-scrollbar { width: 4px; }
  .nav::-webkit-scrollbar-thumb {
    background: rgba(255,255,255,0.08);
    border-radius: 4px;
  }

  /* ── Group label ── */
  .nav-group { margin-bottom: 4px; }
  .nav-group-label {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 10px 6px;
  }
  .label-line {
    flex: 1;
    height: 1px;
    background: rgba(255, 255, 255, 0.07);
  }
  .label-text {
    font-size: 0.63rem;
    font-weight: 700;
    letter-spacing: 0.8px;
    color: #4a5568;
    white-space: nowrap;
    flex-shrink: 0;
  }

  /* ── Nav items ── */
  .nav-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 12px;
    border-radius: 10px;
    text-decoration: none;
    color: #7c8fa6;
    font-size: 0.875rem;
    font-weight: 500;
    transition: var(--transition);
    position: relative;
    border-left: 3px solid transparent;
  }
  .nav-item:hover {
    color: #c4cdd8;
    background: rgba(255, 255, 255, 0.05);
  }
  .nav-item:hover .nav-icon { transform: scale(1.15); }
  .nav-item.active {
    color: #fff;
    background: linear-gradient(90deg, rgba(99,102,241,0.22) 0%, rgba(139,92,246,0.08) 100%);
    border-left-color: var(--primary);
    font-weight: 600;
  }
  .nav-item.active .nav-icon {
    filter: drop-shadow(0 0 6px rgba(99, 102, 241, 0.7));
  }

  .nav-icon {
    font-size: 1rem;
    flex-shrink: 0;
    width: 20px;
    text-align: center;
    transition: transform 0.2s ease;
  }
  .nav-label { flex: 1; }

  .nav-badge {
    font-size: 0.6rem;
    font-weight: 800;
    letter-spacing: 0.3px;
    padding: 2px 6px;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    color: #fff;
    border-radius: 99px;
    flex-shrink: 0;
    box-shadow: 0 2px 8px rgba(99,102,241,0.35);
  }

  /* ── User card + logout ── */
  .sidebar-bottom {
    padding: 12px 10px;
    border-top: 1px solid rgba(255, 255, 255, 0.07);
    display: flex;
    flex-direction: column;
    gap: 8px;
    flex-shrink: 0;
  }
  .user-card {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 10px;
    border: 1px solid rgba(255, 255, 255, 0.07);
  }
  .user-avatar {
    width: 34px;
    height: 34px;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    border-radius: 9px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.88rem;
    font-weight: 800;
    color: #fff;
    flex-shrink: 0;
    box-shadow: 0 3px 10px rgba(99, 102, 241, 0.4);
  }
  .user-info { min-width: 0; }
  .user-name {
    font-size: 0.82rem;
    font-weight: 700;
    color: #e2e8f0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .user-role {
    font-size: 0.65rem;
    font-weight: 600;
    margin-top: 2px;
    padding: 1px 6px;
    border-radius: 99px;
    display: inline-block;
  }
  .role-admin    { background: rgba(239,68,68,0.15);  color: #f87171; }
  .role-teacher  { background: rgba(99,102,241,0.15); color: #a5b4fc; }
  .role-student  { background: rgba(34,197,94,0.15);  color: #4ade80; }
  .logout-btn {
    width: 100%;
    padding: 8px 12px;
    background: rgba(239, 68, 68, 0.08);
    border: 1px solid rgba(239, 68, 68, 0.15);
    border-radius: 8px;
    color: #f87171;
    font-size: 0.82rem;
    font-weight: 600;
    cursor: pointer;
    transition: var(--transition);
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    font-family: inherit;
  }
  .logout-btn:hover {
    background: rgba(239, 68, 68, 0.18);
    color: #fca5a5;
    border-color: rgba(239, 68, 68, 0.3);
  }
  .logout-icon { font-style: normal; }

  /* ════════════════════════════════════════
     MAIN AREA
  ════════════════════════════════════════ */
  .main-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
  }

  /* ── Top header ── */
  .top-header {
    height: 56px;
    background: var(--white);
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 24px;
    flex-shrink: 0;
    gap: 16px;
    box-shadow: 0 1px 4px rgba(0,0,0,0.04);
  }
  .header-left {
    display: flex;
    align-items: center;
    gap: 14px;
    min-width: 0;
  }
  .hamburger {
    display: none;
    flex-direction: column;
    justify-content: center;
    gap: 4.5px;
    width: 36px;
    height: 36px;
    padding: 8px;
    background: transparent;
    border: 1px solid var(--border);
    border-radius: 8px;
    cursor: pointer;
    transition: var(--transition);
    flex-shrink: 0;
  }
  .hamburger:hover { background: var(--bg2); }
  .ham-line {
    width: 100%;
    height: 1.8px;
    background: var(--text2);
    border-radius: 2px;
    display: block;
  }
  .header-title-wrap { min-width: 0; }
  .header-breadcrumb {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 0.72rem;
    color: var(--text3);
    margin-bottom: 1px;
  }
  .bc-root { color: var(--text3); }
  .bc-sep { color: var(--border); }
  .bc-current { color: var(--primary); font-weight: 600; }
  .header-title {
    font-size: 1rem;
    font-weight: 700;
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    line-height: 1.2;
    margin: 0;
  }

  /* Center logo — only visible on mobile */
  .header-logo-center {
    display: none;
    align-items: center;
    gap: 7px;
    background: transparent;
    border: none;
    cursor: pointer;
    text-decoration: none;
    padding: 0;
  }
  .header-logo-icon {
    width: 28px;
    height: 28px;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    border-radius: 7px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.9rem;
  }
  .header-logo-text {
    font-size: 0.95rem;
    font-weight: 800;
    color: var(--text);
  }
  .header-logo-text span { color: var(--accent); }

  .header-right {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-shrink: 0;
  }
  .notif-btn {
    position: relative;
    width: 36px;
    height: 36px;
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
    cursor: pointer;
    transition: var(--transition);
  }
  .notif-btn:hover { background: var(--primary-light); border-color: var(--primary); }
  .notif-badge {
    position: absolute;
    top: -3px;
    right: -3px;
    width: 16px;
    height: 16px;
    background: var(--danger);
    color: #fff;
    font-size: 0.6rem;
    font-weight: 800;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--white);
    line-height: 1;
  }
  .header-avatar {
    width: 34px;
    height: 34px;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    border-radius: 9px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.82rem;
    font-weight: 800;
    color: #fff;
    cursor: pointer;
    border: none;
    box-shadow: 0 3px 10px rgba(99, 102, 241, 0.35);
    transition: var(--transition);
  }
  .header-avatar:hover { transform: scale(1.07); }

  /* ── Page content ── */
  .content {
    flex: 1;
    overflow-y: auto;
    padding: 24px 28px;
    scrollbar-width: thin;
    scrollbar-color: var(--border) transparent;
  }
  .content::-webkit-scrollbar { width: 6px; }
  .content::-webkit-scrollbar-thumb { background: var(--border); border-radius: 6px; }

  /* ════════════════════════════════════════
     SIDEBAR OVERLAY (mobile)
  ════════════════════════════════════════ */
  .sidebar-overlay {
    display: none;
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(2px);
    z-index: 40;
    animation: fadeIn 0.2s ease;
  }

  /* ════════════════════════════════════════
     BOTTOM NAV (mobile)
  ════════════════════════════════════════ */
  .bottom-nav {
    display: none;
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: 56px;
    background: var(--white);
    border-top: 1px solid var(--border);
    z-index: 60;
    align-items: stretch;
    box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.08);
  }
  .bottom-nav-item {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 2px;
    text-decoration: none;
    color: var(--text3);
    transition: var(--transition);
    padding: 4px 0;
  }
  .bottom-nav-item.active { color: var(--primary); }
  .bottom-nav-item.active .bnav-icon { transform: scale(1.18); }
  .bnav-icon { font-size: 1.1rem; line-height: 1; }
  .bnav-label { font-size: 0.6rem; font-weight: 600; line-height: 1; }

  /* ════════════════════════════════════════
     ANIMATIONS
  ════════════════════════════════════════ */
  @keyframes fadeIn {
    from { opacity: 0; }
    to   { opacity: 1; }
  }

  /* ════════════════════════════════════════
     RESPONSIVE — MOBILE (≤768px)
  ════════════════════════════════════════ */
  @media (max-width: 768px) {
    /* Sidebar: slide off-screen by default */
    .sidebar {
      position: fixed;
      top: 0;
      left: 0;
      height: 100vh;
      z-index: 50;
      transform: translateX(-100%);
    }
    .sidebar.open { transform: translateX(0); }

    /* Show overlay */
    .sidebar-overlay { display: block; }

    /* Show hamburger */
    .hamburger { display: flex; }

    /* Hide desktop title/breadcrumb, show center logo */
    .header-title-wrap { display: none; }
    .header-logo-center { display: flex; }

    /* Shrink header padding */
    .top-header { padding: 0 14px; }

    /* Content padding reduction + bottom nav spacing */
    .content {
      padding: 16px;
      padding-bottom: 72px; /* space for bottom nav */
    }

    /* Show bottom nav */
    .bottom-nav { display: flex; }
  }

  /* ════════════════════════════════════════
     TABLET (769px – 1024px)
  ════════════════════════════════════════ */
  @media (min-width: 769px) and (max-width: 1024px) {
    .sidebar { width: 200px; }
    .content { padding: 20px; }
  }
</style>
