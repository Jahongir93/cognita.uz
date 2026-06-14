<script lang="ts">
    import { onMount } from 'svelte';
    import { classes as classesApi } from '$lib/api/client';
    import type { ClassItem } from '$lib/api/client';

    let classes: ClassItem[] = [];
    let loading = true;
    let showCreateModal = false;
    let editingClass: ClassItem | null = null;
    let saving = false;
    let toast = '';
    let toastErr = false;
    let copiedCode = '';

    let form = { name: '', grade: '', subject: '' };
    const grades = ['1','2','3','4','5','6','7','8','9','10','11'];

    // O'quvchilar ro'yxati modali
    let showStudents = false;
    let studentsClass: ClassItem | null = null;
    let studentsList: { id: string; full_name: string; username: string; email: string; joined_at: string }[] = [];
    let studentsLoading = false;
    async function openStudents(cls: ClassItem) {
        studentsClass = cls; showStudents = true; studentsLoading = true; studentsList = [];
        try { studentsList = await classesApi.students(cls.id); } catch { studentsList = []; }
        studentsLoading = false;
    }
    function closeStudents() { showStudents = false; studentsClass = null; }

    const gradeHues = [210, 240, 270, 300, 330, 0, 30, 60, 120, 160, 180];
    function gradientForGrade(grade: string): string {
        const idx = parseInt(grade) - 1;
        const h = gradeHues[Math.max(0, Math.min(idx, gradeHues.length - 1))] || 240;
        return `linear-gradient(135deg, hsl(${h},65%,55%), hsl(${h + 30},65%,45%))`;
    }
    function gradientForName(name: string): string {
        const h = (name.charCodeAt(0) * 7) % 360;
        return `linear-gradient(135deg, hsl(${h},60%,52%), hsl(${(h+40)%360},60%,42%))`;
    }

    onMount(async () => { await load(); });

    async function load() {
        loading = true;
        try { classes = await classesApi.list(); }
        catch { classes = []; }
        finally { loading = false; }
    }

    async function save() {
        if (!form.name.trim()) return;
        saving = true;
        try {
            if (editingClass) {
                await classesApi.update(editingClass.id, form);
                showToast('Sinf yangilandi');
            } else {
                await classesApi.create(form);
                showToast('Yangi sinf yaratildi');
            }
            showCreateModal = false;
            editingClass = null;
            form = { name: '', grade: '', subject: '' };
            await load();
        } catch (e: any) {
            showToast(e.message || 'Xato yuz berdi', true);
        } finally {
            saving = false;
        }
    }

    async function del(id: string, name: string) {
        if (!confirm(`"${name}" o'chirilsinmi?`)) return;
        try {
            await classesApi.delete(id);
            classes = classes.filter(c => c.id !== id);
            showToast("O'chirildi");
        } catch (e: any) {
            showToast(e.message || 'Xato yuz berdi', true);
        }
    }

    function openEdit(cls: ClassItem) {
        editingClass = cls;
        form = { name: cls.name, grade: cls.grade, subject: cls.subject };
        showCreateModal = true;
    }

    function openCreate() {
        editingClass = null;
        form = { name: '', grade: '', subject: '' };
        showCreateModal = true;
    }

    function closeModal() {
        showCreateModal = false;
        editingClass = null;
        form = { name: '', grade: '', subject: '' };
    }

    async function copyCode(code: string) {
        try {
            await navigator.clipboard.writeText(code);
            copiedCode = code;
            setTimeout(() => copiedCode = '', 2000);
        } catch {
            copiedCode = '';
        }
    }

    function showToast(msg: string, err = false) {
        toast = msg; toastErr = err;
        setTimeout(() => toast = '', 2500);
    }

    function handleModalKey(e: KeyboardEvent) {
        if (e.key === 'Escape') closeModal();
    }

    $: totalStudents = classes.reduce((s, c) => s + c.student_count, 0);
    $: activeClasses = classes.filter(c => c.is_active).length;
</script>

<svelte:head><title>Sinflar — Cognita.uz</title></svelte:head>
<svelte:window on:keydown={handleModalKey} />

<!-- Page Header -->
<div class="page-header">
    <div class="header-left">
        <h1>Sinflar</h1>
        <div class="header-stats">
            <span class="hstat"><strong>{classes.length}</strong> sinf</span>
            <span class="hstat-sep">·</span>
            <span class="hstat"><strong>{totalStudents}</strong> o'quvchi</span>
        </div>
    </div>
    <button class="btn primary" on:click={openCreate}>
        <span>+</span> Yangi sinf
    </button>
</div>

<!-- Stats Strip -->
{#if !loading}
<div class="stats-strip">
    <div class="stat-item">
        <div class="stat-icon" style="background:linear-gradient(135deg,var(--primary),var(--accent))">🏫</div>
        <div>
            <div class="stat-num">{classes.length}</div>
            <div class="stat-lbl">Jami sinflar</div>
        </div>
    </div>
    <div class="stat-divider"></div>
    <div class="stat-item">
        <div class="stat-icon" style="background:linear-gradient(135deg,#22c55e,#16a34a)">✅</div>
        <div>
            <div class="stat-num">{activeClasses}</div>
            <div class="stat-lbl">Faol sinflar</div>
        </div>
    </div>
    <div class="stat-divider"></div>
    <div class="stat-item">
        <div class="stat-icon" style="background:linear-gradient(135deg,#f59e0b,#d97706)">👨‍🎓</div>
        <div>
            <div class="stat-num">{totalStudents}</div>
            <div class="stat-lbl">Jami o'quvchilar</div>
        </div>
    </div>
</div>
{/if}

<!-- Classes Grid -->
{#if loading}
    <div class="classes-grid">
        {#each Array(6) as _}
            <div class="skeleton-card">
                <div class="skel-top"></div>
                <div class="skel-body">
                    <div class="skel-line wide"></div>
                    <div class="skel-line short"></div>
                    <div class="skel-row">
                        <div class="skel-box"></div>
                        <div class="skel-box"></div>
                    </div>
                </div>
            </div>
        {/each}
    </div>
{:else if classes.length === 0}
    <div class="empty-state">
        <div class="empty-emoji">👨‍🏫</div>
        <p class="empty-title">Hali sinf yo'q</p>
        <p class="empty-sub">Birinchi sinfingizni yarating</p>
        <button class="btn primary" on:click={openCreate}>+ Yangi sinf yaratish</button>
    </div>
{:else}
    <div class="classes-grid">
        {#each classes as cls (cls.id)}
            <div class="class-card">
                <!-- Top strip with gradient -->
                <div class="card-strip" style="background: {cls.grade ? gradientForGrade(cls.grade) : gradientForName(cls.name)}">
                    <div class="card-strip-letter">{cls.name[0]?.toUpperCase()}</div>
                    {#if !cls.is_active}
                        <span class="inactive-badge">Nofaol</span>
                    {/if}
                </div>
                <!-- Card body -->
                <div class="card-body">
                    <div class="card-name">{cls.name}</div>
                    <div class="card-tags">
                        {#if cls.subject}<span class="tag subject-tag">{cls.subject}</span>{/if}
                        {#if cls.grade}<span class="tag grade-tag">{cls.grade}-sinf</span>{/if}
                    </div>
                    <!-- Stats row -->
                    <div class="card-stats">
                        <div class="cstat">
                            <div class="cstat-val">{cls.student_count}</div>
                            <div class="cstat-lbl">O'quvchi</div>
                        </div>
                        <div class="cstat-divider"></div>
                        <div class="cstat code-stat">
                            <div class="code-row">
                                <span class="cstat-val code-val">{cls.class_code}</span>
                                <button
                                    class="copy-btn"
                                    on:click={() => copyCode(cls.class_code)}
                                    title="Nusxa olish"
                                >
                                    {copiedCode === cls.class_code ? '✓' : '📋'}
                                </button>
                            </div>
                            <div class="cstat-lbl">Sinf kodi</div>
                        </div>
                    </div>
                    <!-- Actions -->
                    <div class="card-actions">
                        <button class="action-btn students-btn" on:click={() => openStudents(cls)}>
                            👥 O'quvchilar
                        </button>
                        <button class="action-btn edit-btn" on:click={() => openEdit(cls)}>
                            ✏️
                        </button>
                        <button class="action-btn delete-btn" on:click={() => del(cls.id, cls.name)}>
                            🗑️
                        </button>
                    </div>
                </div>
            </div>
        {/each}
    </div>
{/if}

<!-- Info box -->
<div class="info-box">
    <span class="info-icon">💡</span>
    <span>O'quvchilar <strong>sinf kodi</strong> orqali sinfga qo'shiladi. Kodni ulashing!</span>
</div>

<!-- Create / Edit Modal -->
{#if showCreateModal}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="modal-overlay" on:click|self={closeModal}>
        <div class="modal">
            <div class="modal-header">
                <h2>{editingClass ? 'Sinfni tahrirlash' : 'Yangi sinf yaratish'}</h2>
                <button class="modal-close" on:click={closeModal}>✕</button>
            </div>
            <div class="modal-body">
                <label class="field">
                    <span class="field-label">Sinf nomi <span class="required">*</span></span>
                    <input
                        type="text"
                        bind:value={form.name}
                        placeholder="9-A sinfi..."
                        class="field-input"
                        autofocus
                    />
                </label>
                <label class="field">
                    <span class="field-label">Sinf raqami</span>
                    <select bind:value={form.grade} class="field-input">
                        <option value="">— Tanlang —</option>
                        {#each grades as g}<option value={g}>{g}-sinf</option>{/each}
                    </select>
                </label>
                <label class="field">
                    <span class="field-label">Fan</span>
                    <input
                        type="text"
                        bind:value={form.subject}
                        placeholder="Matematika, Fizika..."
                        class="field-input"
                    />
                </label>
            </div>
            <div class="modal-footer">
                <button class="btn secondary" on:click={closeModal} disabled={saving}>Bekor qilish</button>
                <button class="btn primary" on:click={save} disabled={saving || !form.name.trim()}>
                    {#if saving}
                        <span class="spinner"></span> Saqlanmoqda...
                    {:else}
                        {editingClass ? '✓ Saqlash' : '+ Yaratish'}
                    {/if}
                </button>
            </div>
        </div>
    </div>
{/if}

<!-- Students Modal -->
{#if showStudents}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="modal-overlay" on:click|self={closeStudents}>
        <div class="modal students-modal">
            <div class="modal-header">
                <h3>👥 {studentsClass?.name} — o'quvchilar</h3>
                <button class="modal-close" on:click={closeStudents}>✕</button>
            </div>
            <div class="students-body">
                {#if studentsLoading}
                    <p class="muted">Yuklanmoqda...</p>
                {:else if studentsList.length === 0}
                    <div class="st-empty">
                        <div class="st-empty-icon">🪑</div>
                        <p>Hali o'quvchi qo'shilmagan.</p>
                        <p class="st-empty-sub">Sinf kodi: <strong>{studentsClass?.class_code}</strong> — ulashing!</p>
                    </div>
                {:else}
                    <div class="st-count">{studentsList.length} o'quvchi</div>
                    {#each studentsList as s, i}
                        <div class="st-row">
                            <span class="st-n">{i + 1}</span>
                            <div class="st-av">{s.full_name?.[0]?.toUpperCase() ?? '?'}</div>
                            <div class="st-info">
                                <span class="st-name">{s.full_name}</span>
                                <span class="st-meta">{s.username || s.email}</span>
                            </div>
                        </div>
                    {/each}
                {/if}
            </div>
        </div>
    </div>
{/if}

<!-- Toast -->
{#if toast}
    <div class="toast" class:toast-err={toastErr}>
        {toastErr ? '⚠️' : '✅'} {toast}
    </div>
{/if}

<style>
    .students-modal { max-width: 440px; }
    .students-body { padding: 18px 22px 22px; max-height: 60vh; overflow-y: auto; }
    .muted { color: var(--text3); }
    .st-count { font-size: 0.8rem; font-weight: 700; color: var(--text3); margin-bottom: 10px; }
    .st-row { display: flex; align-items: center; gap: 12px; padding: 9px 0; border-bottom: 1px solid var(--border); }
    .st-n { width: 22px; color: var(--text3); font-weight: 700; font-size: 0.85rem; }
    .st-av { width: 38px; height: 38px; border-radius: 50%; background: linear-gradient(135deg,var(--primary),var(--accent,#8b5cf6)); color: #fff; display: flex; align-items: center; justify-content: center; font-weight: 800; flex-shrink: 0; }
    .st-info { display: flex; flex-direction: column; }
    .st-name { font-weight: 700; color: var(--text); }
    .st-meta { font-size: 0.76rem; color: var(--text3); }
    .st-empty { text-align: center; padding: 30px 10px; color: var(--text3); }
    .st-empty-icon { font-size: 2.6rem; }
    .st-empty-sub { font-size: 0.82rem; }
    .students-btn { background: var(--primary-light); color: var(--primary); border-color: transparent; flex: 1; }
    .students-btn:hover { filter: brightness(0.96); }

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

    /* ── Stats Strip ─────────────────────────────────────────────────────────── */
    .stats-strip {
        display: flex; align-items: center; background: var(--white);
        border-radius: var(--radius-lg); padding: 18px 28px;
        box-shadow: var(--shadow-sm); margin-bottom: 24px; gap: 32px;
        overflow-x: auto;
    }
    .stat-item { display: flex; align-items: center; gap: 14px; flex-shrink: 0; }
    .stat-icon {
        width: 48px; height: 48px; border-radius: 12px;
        display: flex; align-items: center; justify-content: center;
        font-size: 1.4rem; box-shadow: 0 3px 10px rgba(0,0,0,0.15);
    }
    .stat-num { font-size: 1.8rem; font-weight: 800; color: var(--text); line-height: 1; }
    .stat-lbl { font-size: 0.78rem; color: var(--text3); margin-top: 3px; }
    .stat-divider { width: 1px; height: 40px; background: var(--border); }

    /* ── Classes Grid ────────────────────────────────────────────────────────── */
    .classes-grid {
        display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
        gap: 18px; margin-bottom: 24px;
    }

    /* ── Class Card ──────────────────────────────────────────────────────────── */
    .class-card {
        background: var(--white); border-radius: var(--radius-lg);
        box-shadow: var(--shadow-sm); overflow: hidden;
        transition: var(--transition); border: 1.5px solid transparent;
    }
    .class-card:hover { transform: translateY(-4px); box-shadow: var(--shadow); border-color: var(--primary-light); }

    .card-strip {
        height: 88px; position: relative; display: flex; align-items: center;
        justify-content: center; overflow: hidden;
    }
    .card-strip-letter {
        font-size: 2.5rem; font-weight: 900; color: rgba(255,255,255,0.85);
        text-shadow: 0 2px 8px rgba(0,0,0,0.2);
    }
    .inactive-badge {
        position: absolute; top: 10px; right: 10px;
        background: rgba(0,0,0,0.35); color: white;
        font-size: 0.7rem; font-weight: 700; padding: 2px 9px; border-radius: 99px;
    }

    .card-body { padding: 16px; display: flex; flex-direction: column; gap: 12px; }
    .card-name { font-size: 1.05rem; font-weight: 800; color: var(--text); }
    .card-tags { display: flex; gap: 6px; flex-wrap: wrap; }
    .tag {
        font-size: 0.72rem; font-weight: 600; padding: 3px 9px; border-radius: 99px;
    }
    .subject-tag { background: var(--primary-light); color: #5b21b6; }
    .grade-tag { background: #f0fdf4; color: #15803d; }

    .card-stats {
        display: flex; align-items: center; background: var(--bg2);
        border-radius: 10px; overflow: hidden;
    }
    .cstat { flex: 1; padding: 10px 12px; text-align: center; }
    .cstat-val { font-size: 1.2rem; font-weight: 800; color: var(--text); }
    .code-val { font-family: monospace; font-size: 1rem; color: var(--primary); letter-spacing: 0.12em; }
    .cstat-lbl { font-size: 0.68rem; color: var(--text3); margin-top: 2px; }
    .cstat-divider { width: 1px; height: 36px; background: var(--border); }
    .code-stat { position: relative; }
    .code-row { display: flex; align-items: center; justify-content: center; gap: 6px; }
    .copy-btn {
        background: none; border: none; cursor: pointer; font-size: 0.85rem;
        padding: 2px 4px; border-radius: 6px; transition: background 0.15s;
        line-height: 1;
    }
    .copy-btn:hover { background: var(--border); }

    .card-actions { display: flex; gap: 8px; padding-top: 4px; border-top: 1px solid var(--border); }
    .action-btn {
        flex: 1; padding: 7px; border: none; border-radius: 8px;
        font-size: 0.8rem; font-weight: 600; cursor: pointer;
        transition: all 0.2s; display: flex; align-items: center; justify-content: center; gap: 5px;
    }
    .edit-btn { background: var(--primary-light); color: var(--primary); }
    .edit-btn:hover { background: #ddd6fe; }
    .delete-btn { background: #fee2e2; color: var(--danger); flex: 0 0 40px; }
    .delete-btn:hover { background: #fecaca; }

    /* ── Skeleton ────────────────────────────────────────────────────────────── */
    .skeleton-card {
        background: var(--white); border-radius: var(--radius-lg);
        overflow: hidden; box-shadow: var(--shadow-sm);
    }
    .skel-top {
        height: 88px;
        background: linear-gradient(90deg, var(--bg2) 25%, var(--border) 50%, var(--bg2) 75%);
        background-size: 200% 100%;
        animation: shimmer 1.4s infinite;
    }
    .skel-body { padding: 16px; display: flex; flex-direction: column; gap: 10px; }
    .skel-line {
        height: 14px; border-radius: 7px;
        background: linear-gradient(90deg, var(--bg2) 25%, var(--border) 50%, var(--bg2) 75%);
        background-size: 200% 100%;
        animation: shimmer 1.4s infinite;
    }
    .skel-line.wide { width: 70%; }
    .skel-line.short { width: 45%; }
    .skel-row { display: flex; gap: 10px; }
    .skel-box {
        flex: 1; height: 52px; border-radius: 10px;
        background: linear-gradient(90deg, var(--bg2) 25%, var(--border) 50%, var(--bg2) 75%);
        background-size: 200% 100%;
        animation: shimmer 1.4s infinite;
    }
    @keyframes shimmer { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }

    /* ── Empty ───────────────────────────────────────────────────────────────── */
    .empty-state {
        background: var(--white); border-radius: var(--radius-lg);
        padding: 64px 32px; text-align: center; box-shadow: var(--shadow-sm);
        margin-bottom: 24px; display: flex; flex-direction: column; align-items: center; gap: 8px;
    }
    .empty-emoji { font-size: 3.5rem; margin-bottom: 8px; }
    .empty-title { font-size: 1.1rem; font-weight: 700; color: var(--text); margin: 0; }
    .empty-sub { font-size: 0.875rem; color: var(--text3); margin: 0 0 16px; }

    /* ── Info Box ────────────────────────────────────────────────────────────── */
    .info-box {
        display: flex; align-items: center; gap: 10px;
        background: #eff6ff; border: 1.5px solid #bfdbfe;
        border-radius: var(--radius); padding: 14px 18px;
        font-size: 0.875rem; color: #1e40af;
    }
    .info-icon { font-size: 1.1rem; flex-shrink: 0; }
    .info-box strong { font-weight: 700; }

    /* ── Modal ───────────────────────────────────────────────────────────────── */
    .modal-overlay {
        position: fixed; inset: 0; background: rgba(0,0,0,0.45);
        display: flex; align-items: center; justify-content: center;
        z-index: 100; padding: 16px;
        animation: fadeIn 0.15s ease;
    }
    @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
    .modal {
        background: var(--white); border-radius: var(--radius-lg);
        width: 100%; max-width: 460px; box-shadow: 0 20px 60px rgba(0,0,0,0.2);
        animation: slideUp 0.2s ease;
        overflow: hidden;
    }
    @keyframes slideUp { from { transform: translateY(20px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
    .modal-header {
        display: flex; align-items: center; justify-content: space-between;
        padding: 20px 24px 0;
    }
    .modal-header h2 { font-size: 1.1rem; font-weight: 800; color: var(--text); margin: 0; }
    .modal-close {
        background: var(--bg2); border: none; border-radius: 8px;
        width: 32px; height: 32px; cursor: pointer; font-size: 0.9rem;
        color: var(--text3); display: flex; align-items: center; justify-content: center;
        transition: all 0.15s;
    }
    .modal-close:hover { background: var(--border); color: var(--text); }
    .modal-body {
        padding: 20px 24px; display: flex; flex-direction: column; gap: 14px;
    }
    .modal-footer {
        padding: 16px 24px; border-top: 1px solid var(--border);
        display: flex; justify-content: flex-end; gap: 10px;
        background: var(--bg2);
    }

    .field { display: flex; flex-direction: column; gap: 5px; }
    .field-label { font-size: 0.82rem; font-weight: 600; color: var(--text2); }
    .required { color: var(--danger); }
    .field-input {
        padding: 9px 13px; border: 1.5px solid var(--border); border-radius: var(--radius);
        font-size: 0.9rem; outline: none; transition: border-color 0.2s;
        background: var(--white); color: var(--text);
        width: 100%; box-sizing: border-box;
    }
    .field-input:focus { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(99,102,241,0.1); }

    /* ── Toast ───────────────────────────────────────────────────────────────── */
    .toast {
        position: fixed; bottom: 28px; left: 50%; transform: translateX(-50%);
        background: var(--text); color: white; padding: 12px 22px;
        border-radius: var(--radius-lg); font-size: 0.875rem; font-weight: 600;
        z-index: 200; box-shadow: 0 8px 24px rgba(0,0,0,0.2);
        animation: toastIn 0.25s ease;
        white-space: nowrap;
    }
    .toast.toast-err { background: var(--danger); }
    @keyframes toastIn { from { transform: translateX(-50%) translateY(16px); opacity: 0; } to { transform: translateX(-50%) translateY(0); opacity: 1; } }

    /* ── Buttons ─────────────────────────────────────────────────────────────── */
    .btn {
        padding: 9px 18px; border: none; border-radius: var(--radius);
        font-size: 0.875rem; font-weight: 600; cursor: pointer;
        display: inline-flex; align-items: center; gap: 7px;
        transition: var(--transition); text-decoration: none;
    }
    .btn.primary {
        background: linear-gradient(135deg, var(--primary), var(--accent));
        color: white; box-shadow: 0 3px 10px rgba(99,102,241,0.3);
    }
    .btn.primary:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 16px rgba(99,102,241,0.4); }
    .btn.secondary { background: var(--bg2); color: var(--text); border: 1px solid var(--border); }
    .btn.secondary:hover:not(:disabled) { background: var(--border); }
    .btn:disabled { opacity: 0.5; cursor: not-allowed; }

    .spinner {
        display: inline-block; width: 14px; height: 14px;
        border: 2px solid rgba(255,255,255,0.4); border-top-color: white;
        border-radius: 50%; animation: spin 0.6s linear infinite;
    }
    @keyframes spin { to { transform: rotate(360deg); } }

    /* ── Responsive ──────────────────────────────────────────────────────────── */
    @media (max-width: 640px) {
        .stats-strip { gap: 20px; padding: 14px 18px; }
        .classes-grid { grid-template-columns: 1fr; }
    }
</style>
