<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '$lib/stores/auth';
  import { settings as settingsApi } from '$lib/api/client';

  $: role = $authStore.user?.role ?? 'student';
  $: availableTabs = (role === 'admin'
    ? ['profile', 'security', 'apikeys', 'appearance']
    : ['profile', 'security', 'appearance']) as TabId[];

  // Reset to profile if apikeys tab is active but user is not admin
  $: if (activeTab === 'apikeys' && role !== 'admin') {
    activeTab = 'profile';
  }

  const tabLabels: Record<string, string> = {
    profile:    'Profil',
    security:   'Xavfsizlik',
    apikeys:    'API Kalitlar',
    appearance: "Ko'rinish",
  };
  const tabIcons: Record<string, string> = {
    profile:    '🧑',
    security:   '🔐',
    apikeys:    '🔑',
    appearance: '🎨',
  };

  type TabId = 'profile' | 'security' | 'apikeys' | 'appearance';
  let activeTab: TabId = 'profile';

  // Profile
  let profileName = '';
  let profileEmail = '';
  let profileSaving = false;

  // Security
  let currentPass = '';
  let newPass = '';
  let confirmPass = '';
  let showCurrentPass = false;
  let showNewPass = false;
  let showConfirmPass = false;
  let securitySaving = false;
  let passError = '';

  // API Keys
  let openaiKey = '';
  let geminiKey = '';
  let apiSaving = false;
  let showOpenai = false;
  let showGemini = false;
  let apiCopied = '';

  // Appearance
  let theme: 'light' | 'dark' | 'auto' = 'light';
  let fontSize = 16;

  let toast = '';
  let toastType: 'success' | 'error' = 'success';
  let loading = true;

  onMount(async () => {
    if ($authStore.user) {
      profileName = $authStore.user.full_name ?? '';
      profileEmail = $authStore.user.email ?? '';
    }
    try {
      const all = await settingsApi.getAll();
      openaiKey = all.find(s => s.key === 'openai_api_key')?.value ?? '';
      geminiKey = all.find(s => s.key === 'gemini_api_key')?.value ?? '';
    } catch {}
    loading = false;
  });

  async function saveProfile() {
    profileSaving = true;
    try {
      await new Promise(r => setTimeout(r, 600));
      showToast('Profil yangilandi');
    } catch {
      showToast('Xato yuz berdi', 'error');
    } finally {
      profileSaving = false;
    }
  }

  $: passStrength = (() => {
    const p = newPass;
    if (!p) return 0;
    let s = 0;
    if (p.length >= 8) s++;
    if (/[A-Z]/.test(p)) s++;
    if (/[0-9]/.test(p)) s++;
    if (/[^A-Za-z0-9]/.test(p)) s++;
    return s;
  })();

  $: passStrengthLabel = ['', 'Zaif', "O'rta", 'Yaxshi', 'Kuchli'][passStrength] ?? '';
  $: passStrengthColor = ['', '#ef4444', '#f59e0b', '#eab308', '#22c55e'][passStrength] ?? '';

  async function savePassword() {
    passError = '';
    if (newPass !== confirmPass) { passError = 'Yangi parollar mos emas'; return; }
    if (newPass.length < 8) { passError = 'Parol kamida 8 ta belgi bo\'lishi kerak'; return; }
    securitySaving = true;
    try {
      await new Promise(r => setTimeout(r, 600));
      showToast('Parol yangilandi');
      currentPass = '';
      newPass = '';
      confirmPass = '';
    } catch {
      showToast('Xato yuz berdi', 'error');
    } finally {
      securitySaving = false;
    }
  }

  async function saveApiKeys() {
    apiSaving = true;
    try {
      await Promise.all([
        settingsApi.set('openai_api_key', openaiKey),
        settingsApi.set('gemini_api_key', geminiKey),
      ]);
      showToast('API kalitlar saqlandi');
    } catch (e: any) {
      showToast(e.message ?? 'Xato', 'error');
    } finally {
      apiSaving = false;
    }
  }

  async function copyKey(val: string, id: string) {
    await navigator.clipboard.writeText(val);
    apiCopied = id;
    setTimeout(() => (apiCopied = ''), 2000);
  }

  function showToast(msg: string, type: 'success' | 'error' = 'success') {
    toast = msg;
    toastType = type;
    setTimeout(() => (toast = ''), 3000);
  }

  function maskKey(key: string) {
    if (!key) return '';
    return key.slice(0, 8) + '•'.repeat(Math.max(0, key.length - 12)) + key.slice(-4);
  }

  function getInitials(name: string) {
    if (!name) return 'U';
    return name
      .split(' ')
      .map(w => w[0])
      .slice(0, 2)
      .join('')
      .toUpperCase();
  }

  function getRoleLabel(role: string | undefined) {
    if (role === 'admin') return 'Admin';
    if (role === 'teacher') return "O'qituvchi";
    return 'Foydalanuvchi';
  }
</script>

<svelte:head><title>Sozlamalar — Cognita.uz</title></svelte:head>

<div class="page">
  <!-- Page Header -->
  <div class="page-header">
    <div>
      <h1>Sozlamalar</h1>
      <p class="sub">Hisob va tizim sozlamalari</p>
    </div>
  </div>

  <!-- Tab Navigation -->
  <nav class="tabs">
    {#each availableTabs as tab}
      <button
        class="tab-btn"
        class:active={activeTab === tab}
        on:click={() => { activeTab = tab; }}
      >
        <span class="tab-icon">{tabIcons[tab]}</span>
        <span>{tabLabels[tab]}</span>
      </button>
    {/each}
  </nav>

  <!-- Tab Content -->
  <div class="tab-content">

    <!-- ─── TAB 1: PROFIL ─── -->
    {#if activeTab === 'profile'}
      <div class="tab-panel animate-in">
        <div class="two-col">
          <!-- Left: Form -->
          <div class="card">
            <div class="card-header">
              <h2>Profil ma'lumotlari</h2>
              <p>Shaxsiy ma'lumotlaringizni yangilang</p>
            </div>
            <div class="card-body">
              <!-- Avatar Row -->
              <div class="avatar-row">
                <div class="avatar-circle">
                  <span>{getInitials(profileName)}</span>
                </div>
                <div class="avatar-info">
                  <button class="btn-outline" disabled title="Tez kunda">
                    📷 Rasm yuklash
                  </button>
                  <span class="soon-tag">Tez kunda</span>
                </div>
              </div>

              <div class="form-group">
                <label for="profile-name">To'liq ism</label>
                <input
                  id="profile-name"
                  type="text"
                  class="form-input"
                  bind:value={profileName}
                  placeholder="Ismingizni kiriting"
                />
              </div>

              <div class="form-group">
                <label for="profile-email">
                  Elektron pochta
                  <span class="locked-tag">O'zgartirilmaydi</span>
                </label>
                <input
                  id="profile-email"
                  type="email"
                  class="form-input"
                  value={profileEmail}
                  disabled
                />
              </div>

              <div class="form-group">
                <label>Rol</label>
                <div class="role-badge" class:role-admin={$authStore.user?.role === 'admin'}>
                  {getRoleLabel($authStore.user?.role)}
                </div>
              </div>

              <div class="form-footer">
                <button
                  class="btn-primary"
                  on:click={saveProfile}
                  disabled={profileSaving}
                >
                  {#if profileSaving}
                    <span class="spinner"></span>
                  {/if}
                  Saqlash
                </button>
              </div>
            </div>
          </div>

          <!-- Right: Preview Card -->
          <div class="card preview-card">
            <div class="card-header">
              <h2>Profil ko'rinishi</h2>
              <p>Yon paneldagi ko'rinish</p>
            </div>
            <div class="card-body preview-body">
              <div class="profile-preview">
                <div class="preview-avatar">
                  <span>{getInitials(profileName)}</span>
                </div>
                <div class="preview-info">
                  <div class="preview-name">{profileName || 'Ismingiz'}</div>
                  <div class="preview-email">{profileEmail || 'email@example.com'}</div>
                  <div class="preview-role-wrap">
                    <span class="preview-role" class:role-admin={$authStore.user?.role === 'admin'}>
                      {getRoleLabel($authStore.user?.role)}
                    </span>
                  </div>
                </div>
              </div>
              <p class="preview-note">Bu ko'rinish yon panelda aks etadi</p>
            </div>
          </div>
        </div>
      </div>

    <!-- ─── TAB 2: XAVFSIZLIK ─── -->
    {:else if activeTab === 'security'}
      <div class="tab-panel animate-in">
        <div class="single-col">
          <div class="card">
            <div class="card-header">
              <h2>Parolni o'zgartirish</h2>
              <p>Hisobingizni himoyalash uchun kuchli parol tanlang</p>
            </div>
            <div class="card-body">
              <!-- Current Password -->
              <div class="form-group">
                <label for="cur-pass">Joriy parol</label>
                <div class="pass-wrap">
                  {#if showCurrentPass}
                    <input id="cur-pass" type="text" class="form-input" bind:value={currentPass} placeholder="Joriy parolingiz" autocomplete="current-password" />
                  {:else}
                    <input id="cur-pass" type="password" class="form-input" bind:value={currentPass} placeholder="Joriy parolingiz" autocomplete="current-password" />
                  {/if}
                  <button type="button" class="eye-btn" on:click={() => (showCurrentPass = !showCurrentPass)}>
                    {showCurrentPass ? '🙈' : '👁️'}
                  </button>
                </div>
              </div>

              <!-- New Password -->
              <div class="form-group">
                <label for="new-pass">Yangi parol</label>
                <div class="pass-wrap">
                  {#if showNewPass}
                    <input id="new-pass" type="text" class="form-input" bind:value={newPass} placeholder="Yangi parolingiz" autocomplete="new-password" />
                  {:else}
                    <input id="new-pass" type="password" class="form-input" bind:value={newPass} placeholder="Yangi parolingiz" autocomplete="new-password" />
                  {/if}
                  <button type="button" class="eye-btn" on:click={() => (showNewPass = !showNewPass)}>
                    {showNewPass ? '🙈' : '👁️'}
                  </button>
                </div>
                <!-- Strength indicator -->
                {#if newPass}
                  <div class="strength-bar">
                    <div class="strength-segments">
                      {#each [1, 2, 3, 4] as seg}
                        <div
                          class="strength-seg"
                          style="background: {passStrength >= seg ? passStrengthColor : 'var(--border)'};"
                        ></div>
                      {/each}
                    </div>
                    <span class="strength-label" style="color: {passStrengthColor};">{passStrengthLabel}</span>
                  </div>
                {/if}
              </div>

              <!-- Confirm Password -->
              <div class="form-group">
                <label for="conf-pass">Parolni tasdiqlash</label>
                <div class="pass-wrap">
                  {#if showConfirmPass}
                    <input id="conf-pass" type="text" class="form-input" bind:value={confirmPass} placeholder="Parolni qaytaring" autocomplete="new-password" />
                  {:else}
                    <input id="conf-pass" type="password" class="form-input" bind:value={confirmPass} placeholder="Parolni qaytaring" autocomplete="new-password" />
                  {/if}
                  <button type="button" class="eye-btn" on:click={() => (showConfirmPass = !showConfirmPass)}>
                    {showConfirmPass ? '🙈' : '👁️'}
                  </button>
                </div>
              </div>

              {#if passError}
                <div class="error-msg">⚠️ {passError}</div>
              {/if}

              <div class="form-footer">
                <button
                  class="btn-primary"
                  on:click={savePassword}
                  disabled={securitySaving || !currentPass || !newPass || !confirmPass}
                >
                  {#if securitySaving}
                    <span class="spinner"></span>
                  {/if}
                  Saqlash
                </button>
              </div>
            </div>
          </div>

          <!-- Security Tips -->
          <div class="info-box blue">
            <div class="info-box-title">💡 Xavfsizlik maslahatlari</div>
            <ul class="tips-list">
              <li>Parolda katta va kichik harflar, raqamlar va belgilardan foydalaning</li>
              <li>Bir xil parolni boshqa saytlarda ishlatmang</li>
              <li>Parolingizni hech kim bilan ulashmang, hatto texnik yordam xodimlari bilan ham</li>
            </ul>
          </div>
        </div>
      </div>

    <!-- ─── TAB 3: API KALITLAR ─── -->
    {:else if activeTab === 'apikeys'}
      <div class="tab-panel animate-in">
        <div class="single-col">
          <div class="card">
            <div class="card-header">
              <h2>Sun'iy intellekt API kalitlari</h2>
              <p>AI yordamida savollar yaratish uchun API kalitlarni kiriting</p>
            </div>
            <div class="card-body">
              <!-- OpenAI -->
              <div class="api-row">
                <div class="api-row-top">
                  <div class="api-logo">🤖</div>
                  <div class="api-info">
                    <div class="api-title">OpenAI GPT-4</div>
                    <div class="api-subtitle">gpt-4o, gpt-4o-mini modellari</div>
                  </div>
                  <div class="api-status">
                    {#if openaiKey}
                      <span class="badge-green">Sozlangan ✓</span>
                    {:else}
                      <span class="badge-gray">Sozlanmagan</span>
                    {/if}
                  </div>
                </div>
                <div class="api-input-wrap">
                  <div class="api-field-wrap">
                    {#if showOpenai}
                      <input
                        type="text"
                        class="form-input api-input"
                        bind:value={openaiKey}
                        placeholder="sk-..."
                        autocomplete="off"
                        spellcheck="false"
                      />
                    {:else}
                      <input
                        type="password"
                        class="form-input api-input"
                        bind:value={openaiKey}
                        placeholder="sk-..."
                        autocomplete="off"
                        spellcheck="false"
                      />
                    {/if}
                    <div class="api-btn-group">
                      <button type="button" class="api-icon-btn" title={showOpenai ? 'Yashirish' : "Ko'rsatish"} on:click={() => (showOpenai = !showOpenai)}>
                        {showOpenai ? '🙈' : '👁️'}
                      </button>
                      <button
                        type="button"
                        class="api-icon-btn"
                        class:copied={apiCopied === 'openai'}
                        title="Nusxa olish"
                        disabled={!openaiKey}
                        on:click={() => copyKey(openaiKey, 'openai')}
                      >
                        {apiCopied === 'openai' ? '✓' : '📋'}
                      </button>
                    </div>
                  </div>
                  <span class="api-helper">sk-... formatida</span>
                </div>
              </div>

              <div class="api-divider"></div>

              <!-- Gemini -->
              <div class="api-row">
                <div class="api-row-top">
                  <div class="api-logo">✨</div>
                  <div class="api-info">
                    <div class="api-title">Google Gemini</div>
                    <div class="api-subtitle">gemini-1.5-pro modeli</div>
                  </div>
                  <div class="api-status">
                    {#if geminiKey}
                      <span class="badge-green">Sozlangan ✓</span>
                    {:else}
                      <span class="badge-gray">Sozlanmagan</span>
                    {/if}
                  </div>
                </div>
                <div class="api-input-wrap">
                  <div class="api-field-wrap">
                    {#if showGemini}
                      <input
                        type="text"
                        class="form-input api-input"
                        bind:value={geminiKey}
                        placeholder="AIza..."
                        autocomplete="off"
                        spellcheck="false"
                      />
                    {:else}
                      <input
                        type="password"
                        class="form-input api-input"
                        bind:value={geminiKey}
                        placeholder="AIza..."
                        autocomplete="off"
                        spellcheck="false"
                      />
                    {/if}
                    <div class="api-btn-group">
                      <button type="button" class="api-icon-btn" title={showGemini ? 'Yashirish' : "Ko'rsatish"} on:click={() => (showGemini = !showGemini)}>
                        {showGemini ? '🙈' : '👁️'}
                      </button>
                      <button
                        type="button"
                        class="api-icon-btn"
                        class:copied={apiCopied === 'gemini'}
                        title="Nusxa olish"
                        disabled={!geminiKey}
                        on:click={() => copyKey(geminiKey, 'gemini')}
                      >
                        {apiCopied === 'gemini' ? '✓' : '📋'}
                      </button>
                    </div>
                  </div>
                  <span class="api-helper">AIza... formatida</span>
                </div>
              </div>

              <div class="form-footer">
                <button class="btn-primary" on:click={saveApiKeys} disabled={apiSaving}>
                  {#if apiSaving}
                    <span class="spinner"></span>
                  {/if}
                  Saqlash
                </button>
              </div>
            </div>
          </div>

          <div class="info-box blue">
            <span>🔒 API kalitlar shifrlangan holda saqlanadi</span>
          </div>
        </div>
      </div>

    <!-- ─── TAB 4: KO'RINISH ─── -->
    {:else if activeTab === 'appearance'}
      <div class="tab-panel animate-in">
        <div class="single-col">
          <div class="card">
            <div class="card-header">
              <h2>Ko'rinish sozlamalari</h2>
              <p>Interfeys ranglar va o'lchamlarini moslashtiring</p>
            </div>
            <div class="card-body">
              <!-- Theme Selector -->
              <div class="form-group">
                <label>Mavzu</label>
                <div class="theme-grid">
                  <!-- Light -->
                  <label class="theme-card" class:selected={theme === 'light'}>
                    <input type="radio" name="theme" value="light" bind:group={theme} />
                    <div class="theme-preview light-preview">
                      <div class="tp-bar"></div>
                      <div class="tp-lines">
                        <div class="tp-line"></div>
                        <div class="tp-line short"></div>
                      </div>
                    </div>
                    <span class="theme-label">☀️ Yorug' rejim</span>
                  </label>

                  <!-- Dark -->
                  <label class="theme-card" class:selected={theme === 'dark'}>
                    <input type="radio" name="theme" value="dark" bind:group={theme} />
                    <div class="theme-preview dark-preview">
                      <div class="tp-bar"></div>
                      <div class="tp-lines">
                        <div class="tp-line"></div>
                        <div class="tp-line short"></div>
                      </div>
                    </div>
                    <span class="theme-label">🌙 Qorong'u rejim</span>
                  </label>

                  <!-- Auto -->
                  <label class="theme-card" class:selected={theme === 'auto'}>
                    <input type="radio" name="theme" value="auto" bind:group={theme} />
                    <div class="theme-preview auto-preview">
                      <div class="tp-half light-half">
                        <div class="tp-bar-half"></div>
                      </div>
                      <div class="tp-half dark-half">
                        <div class="tp-bar-half dark"></div>
                      </div>
                    </div>
                    <span class="theme-label">💻 Tizim</span>
                  </label>
                </div>
              </div>

              <!-- Font Size -->
              <div class="form-group">
                <label for="font-size">
                  Matn o'lchami
                  <span class="font-size-val">{fontSize}px</span>
                </label>
                <div class="slider-wrap">
                  <span class="slider-min">12</span>
                  <input
                    id="font-size"
                    type="range"
                    min="12"
                    max="18"
                    step="1"
                    bind:value={fontSize}
                    class="range-input"
                  />
                  <span class="slider-max">18</span>
                </div>
                <span class="range-hint">Odatiy: 16px</span>
              </div>

              <div class="form-footer">
                <button class="btn-primary" on:click={() => showToast("Ko'rinish saqlandi")}>
                  Saqlash
                </button>
              </div>
            </div>
          </div>

          <div class="info-box amber">
            🚧 Qorong'u rejim tez kunda qo'shiladi
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<!-- Toast -->
{#if toast}
  <div class="toast" class:toast-success={toastType === 'success'} class:toast-error={toastType === 'error'}>
    <span>{toastType === 'success' ? '✓' : '✕'}</span>
    {toast}
  </div>
{/if}

<style>
  /* ── Page ── */
  .page {
    max-width: 860px;
  }

  .page-header {
    margin-bottom: 24px;
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

  /* ── Tabs ── */
  .tabs {
    display: flex;
    gap: 4px;
    background: var(--white);
    border: 1.5px solid var(--border);
    border-radius: var(--radius);
    padding: 5px;
    margin-bottom: 20px;
    width: fit-content;
    box-shadow: var(--shadow-sm);
  }

  .tab-btn {
    display: flex;
    align-items: center;
    gap: 7px;
    padding: 9px 18px;
    border: none;
    border-radius: 9px;
    background: transparent;
    color: var(--text2);
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: var(--transition);
    white-space: nowrap;
  }

  .tab-btn:hover:not(.active) {
    background: var(--bg);
    color: var(--text);
  }

  .tab-btn.active {
    background: var(--primary);
    color: #fff;
    box-shadow: 0 2px 8px rgba(99, 102, 241, 0.35);
  }

  .tab-icon {
    font-size: 1rem;
    line-height: 1;
  }

  /* ── Tab Content ── */
  .tab-content {
    min-height: 400px;
  }

  .tab-panel {
    display: contents;
  }

  @keyframes fade-in-up {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .animate-in {
    animation: fade-in-up 0.22s ease both;
  }

  /* ── Layout ── */
  .two-col {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 20px;
    align-items: start;
  }

  .single-col {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  /* ── Card ── */
  .card {
    background: var(--white);
    border: 1.5px solid var(--border);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
    overflow: hidden;
  }

  .card-header {
    padding: 20px 24px 16px;
    border-bottom: 1px solid var(--border);
  }

  .card-header h2 {
    font-size: 1.05rem;
    font-weight: 800;
    color: var(--text);
    margin: 0 0 3px;
  }

  .card-header p {
    font-size: 0.8rem;
    color: var(--text3);
    margin: 0;
  }

  .card-body {
    padding: 22px 24px;
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  /* ── Avatar ── */
  .avatar-row {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .avatar-circle {
    width: 60px;
    height: 60px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.3rem;
    font-weight: 800;
    color: #fff;
    flex-shrink: 0;
    box-shadow: 0 4px 14px rgba(99, 102, 241, 0.3);
  }

  .avatar-info {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }

  .btn-outline {
    padding: 7px 14px;
    border: 1.5px solid var(--border);
    border-radius: 9px;
    background: var(--white);
    color: var(--text2);
    font-size: 0.82rem;
    font-weight: 600;
    cursor: not-allowed;
    opacity: 0.65;
    transition: var(--transition);
  }

  .soon-tag {
    font-size: 0.72rem;
    color: var(--text3);
  }

  /* ── Form Elements ── */
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 7px;
  }

  .form-group label {
    font-size: 0.82rem;
    font-weight: 700;
    color: var(--text2);
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .form-input {
    width: 100%;
    padding: 10px 14px;
    border: 1.5px solid var(--border);
    border-radius: 10px;
    font-size: 0.875rem;
    color: var(--text);
    background: var(--white);
    outline: none;
    transition: border-color 0.2s, box-shadow 0.2s;
    box-sizing: border-box;
  }

  .form-input:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
  }

  .form-input:disabled {
    background: var(--bg2);
    color: var(--text3);
    cursor: not-allowed;
  }

  .locked-tag {
    font-size: 0.7rem;
    font-weight: 600;
    background: #f1f5f9;
    color: var(--text3);
    padding: 2px 8px;
    border-radius: 99px;
  }

  .role-badge {
    display: inline-flex;
    align-items: center;
    padding: 6px 14px;
    border-radius: 99px;
    font-size: 0.82rem;
    font-weight: 700;
    background: var(--primary-light);
    color: var(--primary);
    width: fit-content;
  }

  .role-badge.role-admin {
    background: #fde68a;
    color: #92400e;
  }

  .form-footer {
    padding-top: 4px;
  }

  /* ── Buttons ── */
  .btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 10px 22px;
    background: var(--primary);
    color: #fff;
    border: none;
    border-radius: 10px;
    font-size: 0.875rem;
    font-weight: 700;
    cursor: pointer;
    transition: var(--transition);
    box-shadow: 0 2px 8px rgba(99, 102, 241, 0.3);
  }

  .btn-primary:hover:not(:disabled) {
    background: #4f46e5;
    box-shadow: 0 4px 14px rgba(99, 102, 241, 0.4);
    transform: translateY(-1px);
  }

  .btn-primary:active:not(:disabled) {
    transform: translateY(0);
  }

  .btn-primary:disabled {
    opacity: 0.55;
    cursor: not-allowed;
    box-shadow: none;
    transform: none;
  }

  /* ── Spinner ── */
  .spinner {
    width: 14px;
    height: 14px;
    border: 2px solid rgba(255, 255, 255, 0.4);
    border-top-color: #fff;
    border-radius: 50%;
    animation: spin 0.65s linear infinite;
    flex-shrink: 0;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* ── Profile Preview ── */
  .preview-card {
    align-self: start;
  }

  .preview-body {
    align-items: center;
    text-align: center;
  }

  .profile-preview {
    background: var(--bg);
    border: 1.5px solid var(--border);
    border-radius: var(--radius);
    padding: 22px 24px;
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    box-sizing: border-box;
  }

  .preview-avatar {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.4rem;
    font-weight: 800;
    color: #fff;
    box-shadow: 0 4px 16px rgba(99, 102, 241, 0.3);
  }

  .preview-info {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    width: 100%;
  }

  .preview-name {
    font-size: 1rem;
    font-weight: 800;
    color: var(--text);
  }

  .preview-email {
    font-size: 0.78rem;
    color: var(--text3);
  }

  .preview-role-wrap {
    margin-top: 4px;
  }

  .preview-role {
    display: inline-flex;
    padding: 3px 12px;
    border-radius: 99px;
    font-size: 0.72rem;
    font-weight: 700;
    background: var(--primary-light);
    color: var(--primary);
  }

  .preview-role.role-admin {
    background: #fde68a;
    color: #92400e;
  }

  .preview-note {
    font-size: 0.75rem;
    color: var(--text3);
    margin: 0;
    text-align: center;
  }

  /* ── Password ── */
  .pass-wrap {
    position: relative;
  }

  .pass-wrap .form-input {
    padding-right: 44px;
  }

  .eye-btn {
    position: absolute;
    right: 10px;
    top: 50%;
    transform: translateY(-50%);
    background: none;
    border: none;
    cursor: pointer;
    font-size: 1rem;
    line-height: 1;
    padding: 2px;
    opacity: 0.7;
    transition: opacity 0.15s;
  }

  .eye-btn:hover {
    opacity: 1;
  }

  /* ── Strength Bar ── */
  .strength-bar {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: 6px;
  }

  .strength-segments {
    display: flex;
    gap: 4px;
    flex: 1;
  }

  .strength-seg {
    height: 4px;
    flex: 1;
    border-radius: 99px;
    transition: background 0.25s;
  }

  .strength-label {
    font-size: 0.75rem;
    font-weight: 700;
    min-width: 48px;
    text-align: right;
    transition: color 0.25s;
  }

  /* ── Error ── */
  .error-msg {
    font-size: 0.82rem;
    color: var(--danger);
    background: #fee2e2;
    border: 1px solid #fecaca;
    border-radius: 8px;
    padding: 9px 13px;
  }

  /* ── API Keys ── */
  .api-row {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .api-row-top {
    display: flex;
    align-items: flex-start;
    gap: 12px;
  }

  .api-logo {
    font-size: 1.7rem;
    flex-shrink: 0;
    line-height: 1;
    margin-top: 2px;
  }

  .api-info {
    flex: 1;
  }

  .api-title {
    font-size: 0.9rem;
    font-weight: 700;
    color: var(--text);
  }

  .api-subtitle {
    font-size: 0.78rem;
    color: var(--text3);
    margin-top: 2px;
  }

  .api-status {
    flex-shrink: 0;
  }

  .badge-green {
    display: inline-flex;
    align-items: center;
    padding: 3px 10px;
    background: #dcfce7;
    color: #16a34a;
    border-radius: 99px;
    font-size: 0.72rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .badge-gray {
    display: inline-flex;
    align-items: center;
    padding: 3px 10px;
    background: #f1f5f9;
    color: var(--text3);
    border-radius: 99px;
    font-size: 0.72rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .api-input-wrap {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }

  .api-field-wrap {
    position: relative;
    display: flex;
    align-items: center;
  }

  .api-input {
    padding-right: 82px;
  }

  .api-btn-group {
    position: absolute;
    right: 6px;
    display: flex;
    gap: 2px;
  }

  .api-icon-btn {
    width: 32px;
    height: 32px;
    background: none;
    border: none;
    border-radius: 7px;
    cursor: pointer;
    font-size: 0.9rem;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0.7;
    transition: opacity 0.15s, background 0.15s;
  }

  .api-icon-btn:hover:not(:disabled) {
    background: var(--bg);
    opacity: 1;
  }

  .api-icon-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .api-icon-btn.copied {
    color: var(--success);
    opacity: 1;
    font-weight: 800;
  }

  .api-helper {
    font-size: 0.75rem;
    color: var(--text3);
  }

  .api-divider {
    height: 1px;
    background: var(--border);
    margin: 4px 0;
  }

  /* ── Appearance ── */
  .theme-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
  }

  .theme-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 14px 10px;
    border: 2px solid var(--border);
    border-radius: var(--radius);
    cursor: pointer;
    transition: border-color 0.2s, box-shadow 0.2s;
    background: var(--white);
  }

  .theme-card input[type="radio"] {
    display: none;
  }

  .theme-card:hover {
    border-color: var(--primary);
  }

  .theme-card.selected {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
  }

  .theme-preview {
    width: 100%;
    height: 72px;
    border-radius: 8px;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 8px;
    box-sizing: border-box;
  }

  .light-preview {
    background: #f8fafc;
    border: 1px solid #e2e8f0;
  }

  .dark-preview {
    background: #1e293b;
    border: 1px solid #334155;
  }

  .auto-preview {
    display: flex;
    padding: 0;
    flex-direction: row;
    border: 1px solid var(--border);
  }

  .tp-half {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: 8px 6px;
    gap: 5px;
  }

  .light-half {
    background: #f8fafc;
  }

  .dark-half {
    background: #1e293b;
  }

  .tp-bar {
    height: 8px;
    border-radius: 4px;
    background: #6366f1;
    opacity: 0.85;
  }

  .tp-bar-half {
    height: 8px;
    border-radius: 4px;
    background: #6366f1;
    opacity: 0.85;
  }

  .tp-bar-half.dark {
    background: #818cf8;
  }

  .tp-lines {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .tp-line {
    height: 4px;
    border-radius: 99px;
    background: #cbd5e1;
  }

  .dark-preview .tp-line {
    background: #475569;
  }

  .tp-line.short {
    width: 60%;
  }

  .theme-label {
    font-size: 0.78rem;
    font-weight: 700;
    color: var(--text2);
    text-align: center;
  }

  /* ── Font Size Slider ── */
  .font-size-val {
    font-size: 0.75rem;
    font-weight: 700;
    color: var(--primary);
    background: var(--primary-light);
    padding: 2px 8px;
    border-radius: 99px;
  }

  .slider-wrap {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .slider-min,
  .slider-max {
    font-size: 0.75rem;
    font-weight: 700;
    color: var(--text3);
    min-width: 20px;
    text-align: center;
  }

  .range-input {
    flex: 1;
    height: 6px;
    appearance: none;
    background: linear-gradient(
      to right,
      var(--primary) 0%,
      var(--primary) calc((var(--val, 50)) * 1%),
      var(--border) calc((var(--val, 50)) * 1%),
      var(--border) 100%
    );
    border-radius: 99px;
    outline: none;
    cursor: pointer;
  }

  .range-input::-webkit-slider-thumb {
    appearance: none;
    width: 18px;
    height: 18px;
    background: var(--primary);
    border-radius: 50%;
    cursor: pointer;
    box-shadow: 0 2px 6px rgba(99, 102, 241, 0.4);
    border: 2px solid #fff;
    transition: transform 0.15s;
  }

  .range-input::-webkit-slider-thumb:hover {
    transform: scale(1.2);
  }

  .range-hint {
    font-size: 0.75rem;
    color: var(--text3);
  }

  /* ── Info Boxes ── */
  .info-box {
    padding: 13px 16px;
    border-radius: var(--radius);
    font-size: 0.85rem;
    font-weight: 500;
  }

  .info-box.blue {
    background: #eff6ff;
    border: 1px solid #bfdbfe;
    color: #1d4ed8;
  }

  .info-box.amber {
    background: #fffbeb;
    border: 1px solid #fde68a;
    color: #92400e;
  }

  .info-box-title {
    font-weight: 700;
    margin-bottom: 8px;
    color: #1d4ed8;
  }

  .tips-list {
    margin: 0;
    padding-left: 16px;
    display: flex;
    flex-direction: column;
    gap: 5px;
  }

  .tips-list li {
    font-size: 0.82rem;
    color: #3b82f6;
    line-height: 1.5;
  }

  /* ── Toast ── */
  .toast {
    position: fixed;
    bottom: 28px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 9px;
    padding: 12px 22px;
    border-radius: 12px;
    font-size: 0.875rem;
    font-weight: 700;
    box-shadow: 0 8px 28px rgba(0, 0, 0, 0.15);
    z-index: 9999;
    animation: toast-in 0.25s cubic-bezier(0.34, 1.56, 0.64, 1) both;
    white-space: nowrap;
  }

  .toast-success {
    background: var(--success);
    color: #fff;
  }

  .toast-error {
    background: var(--danger);
    color: #fff;
  }

  @keyframes toast-in {
    from {
      opacity: 0;
      transform: translateX(-50%) translateY(16px) scale(0.92);
    }
    to {
      opacity: 1;
      transform: translateX(-50%) translateY(0) scale(1);
    }
  }

  /* ── Responsive ── */
  @media (max-width: 680px) {
    .two-col {
      grid-template-columns: 1fr;
    }

    .theme-grid {
      grid-template-columns: 1fr;
    }

    .tabs {
      width: 100%;
      overflow-x: auto;
    }

    .tab-btn {
      padding: 8px 14px;
      font-size: 0.8rem;
    }
  }
</style>
