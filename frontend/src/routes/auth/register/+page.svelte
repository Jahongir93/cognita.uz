<script lang="ts">
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { authStore } from '$lib/stores/auth';
    import type { UserRole } from '$lib/api/types';

    $: defaultRole = ($page.url.searchParams.get('role') ?? 'teacher') as UserRole;

    let fullName = '';
    let email    = '';
    let username = '';
    let password = '';
    let role: UserRole = defaultRole;
    let showPass = false;
    let error    = '';
    let loading  = false;
    let mounted  = false;
    let focusedField = '';

    onMount(() => { mounted = true; });

    $: passStrength = password.length === 0 ? 0
        : password.length < 6  ? 1
        : password.length < 8  ? 2
        : password.length < 12 ? 3 : 4;

    const strengthLabel = ['','Juda zaif','Zaif','Yaxshi','Kuchli'];
    const strengthColor = ['','#ef4444','#f59e0b','#22c55e','#6366f1'];

    async function handleRegister() {
        if (!fullName || !email || !username || !password) {
            error = "Barcha maydonlarni to'ldiring"; return;
        }
        if (password.length < 8) {
            error = "Parol kamida 8 ta belgidan iborat bo'lishi kerak"; return;
        }
        error = ''; loading = true;
        try {
            await authStore.register({ email, username, password, full_name: fullName, role });
            goto('/dashboard');
        } catch (e: any) {
            error = e.message ?? "Ro'yxatdan o'tishda xato";
        } finally {
            loading = false;
        }
    }
</script>

<svelte:head><title>Ro'yxatdan o'tish — Cognita.uz</title></svelte:head>

<div class="page" class:mounted>
    <!-- Left panel -->
    <div class="left-panel">
        <div class="left-inner">
            <a href="/" class="brand"><img src="/logowhite.png" alt="Cognita.uz" style="height:38px;width:auto" /></a>
            <div class="lc">
                <div class="lc-icon">🚀</div>
                <h2>Boshlang!</h2>
                <p>O'qituvchi yoki o'quvchi — ikkalasi uchun ham ideal platforma.</p>
                <div class="steps-list">
                    {#each [
                        { n:'01', t:"Hisob yarating",    d:"1 daqiqa vaqt oladi" },
                        { n:'02', t:"Quiz yarating",     d:"AI yoki qo'lda, tezkor" },
                        { n:'03', t:"O'yinni boshlang",  d:"PIN ulashing, o'ynang!" },
                    ] as s, i}
                        <div class="step-row" style="animation-delay:{i*120+400}ms">
                            <div class="step-n">{s.n}</div>
                            <div>
                                <div class="step-t">{s.t}</div>
                                <div class="step-d">{s.d}</div>
                            </div>
                        </div>
                    {/each}
                </div>
            </div>
            <div class="left-ornament" aria-hidden="true">
                {#each ['▲','◆','●','■','★'] as sh, i}
                    <div class="orn" style="left:{10+i*18}%;bottom:{5+((i*31)%35)}%;
                         font-size:{12+i*4}px;opacity:.15;animation-delay:{i*0.7}s;
                         color:{'#a78bfa,#fbbf24,#a78bfa,#fbbf24,#a78bfa'.split(',')[i]}">
                        {sh}
                    </div>
                {/each}
            </div>
        </div>
    </div>

    <!-- Right panel -->
    <div class="right-panel">
        <div class="form-card">
            <div class="form-header">
                <h1>Hisob yaratish</h1>
                <p>Bir daqiqada boshlang — bepul!</p>
            </div>

            <!-- Role selector -->
            <div class="role-wrap">
                <button
                    type="button"
                    class="role-card"
                    class:active={role === 'teacher'}
                    on:click={() => role = 'teacher'}
                >
                    <span class="role-ic">👨‍🏫</span>
                    <span class="role-t">O'qituvchi</span>
                    <span class="role-d">Quiz yarat, boshqar</span>
                    {#if role === 'teacher'}<span class="role-check">✓</span>{/if}
                </button>
                <button
                    type="button"
                    class="role-card"
                    class:active={role === 'student'}
                    on:click={() => role = 'student'}
                >
                    <span class="role-ic">🎓</span>
                    <span class="role-t">O'quvchi</span>
                    <span class="role-d">O'yin va testlarda qatnash</span>
                    {#if role === 'student'}<span class="role-check">✓</span>{/if}
                </button>
            </div>

            <form on:submit|preventDefault={handleRegister}>
                <!-- Full name -->
                <div class="field" class:focused={focusedField === 'name'}>
                    <label for="fname"><span class="li">👤</span> To'liq ism</label>
                    <div class="iw">
                        <input id="fname" type="text" placeholder="Sardor Aliyev"
                               bind:value={fullName} autocomplete="name"
                               on:focus={() => focusedField='name'} on:blur={() => focusedField=''} />
                        {#if fullName.trim().length >= 2}<span class="fok">✓</span>{/if}
                    </div>
                </div>

                <!-- Email -->
                <div class="field" class:focused={focusedField === 'email'}>
                    <label for="reg-email"><span class="li">📧</span> Email</label>
                    <div class="iw">
                        <input id="reg-email" type="email" placeholder="email@maktab.uz"
                               bind:value={email} autocomplete="email"
                               on:focus={() => focusedField='email'} on:blur={() => focusedField=''} />
                        {#if email.includes('@') && email.includes('.')}<span class="fok">✓</span>{/if}
                    </div>
                </div>

                <!-- Username -->
                <div class="field" class:focused={focusedField === 'uname'}>
                    <label for="uname"><span class="li">🏷</span> Foydalanuvchi nomi</label>
                    <div class="iw">
                        <input id="uname" type="text" placeholder="sardor2025"
                               bind:value={username} autocomplete="username"
                               on:focus={() => focusedField='uname'} on:blur={() => focusedField=''} />
                        {#if username.length >= 3}<span class="fok">✓</span>{/if}
                    </div>
                </div>

                <!-- Password -->
                <div class="field" class:focused={focusedField === 'pass'}>
                    <label for="reg-pass"><span class="li">🔒</span> Parol</label>
                    <div class="iw">
                        {#if showPass}
                            <input id="reg-pass" type="text"
                                   placeholder="Kamida 8 belgi"
                                   bind:value={password} autocomplete="new-password"
                                   on:focus={() => focusedField='pass'} on:blur={() => focusedField=''} />
                        {:else}
                            <input id="reg-pass" type="password"
                                   placeholder="Kamida 8 belgi"
                                   bind:value={password} autocomplete="new-password"
                                   on:focus={() => focusedField='pass'} on:blur={() => focusedField=''} />
                        {/if}
                        <button type="button" class="eye" on:click={() => showPass = !showPass} tabindex="-1">
                            {showPass ? '🙈' : '👁'}
                        </button>
                    </div>
                    <!-- Strength meter -->
                    {#if password.length > 0}
                        <div class="strength-wrap">
                            <div class="strength-bars">
                                {#each Array(4) as _, i}
                                    <div class="sbar" class:active={i < passStrength}
                                         style="background:{i < passStrength ? strengthColor[passStrength] : '#e2e8f0'}"></div>
                                {/each}
                            </div>
                            <span class="strength-label" style="color:{strengthColor[passStrength]}">
                                {strengthLabel[passStrength]}
                            </span>
                        </div>
                    {/if}
                </div>

                {#if error}
                    <div class="error-box" role="alert"><span>⚠</span> {error}</div>
                {/if}

                <button type="submit" class="submit-btn" disabled={loading}>
                    {#if loading}
                        <span class="spin"></span> Yaratilmoqda...
                    {:else}
                        🎉 Hisob yaratish
                        <div class="btn-shine"></div>
                    {/if}
                </button>
            </form>

            <p class="footer-text">
                Hisobingiz bormi? <a href="/auth/login">Kirish</a>
            </p>
        </div>
    </div>
</div>

<style>
    :global(body) { margin: 0; font-family: 'Segoe UI', system-ui, sans-serif; }
    * { box-sizing: border-box; }

    .page {
        min-height: 100dvh;
        display: grid;
        grid-template-columns: 1fr 1fr;
    }

    /* ── LEFT ── */
    .left-panel {
        position: relative; overflow: hidden;
        background: linear-gradient(160deg, #0f172a 0%, #1e1b4b 40%, #312e81 100%);
        display: flex; align-items: center; justify-content: center;
    }
    .left-inner {
        position: relative; z-index: 2;
        padding: 48px 40px;
        display: flex; flex-direction: column;
        width: 100%; max-width: 420px;
    }
    .brand {
        display: flex; align-items: center; gap: 6px;
        text-decoration: none; font-weight: 900; font-size: 1.2rem; color: white;
        margin-bottom: 56px;
        opacity: 0; animation: fadeUp .6s ease .1s forwards;
    }
    .dot { color: #a78bfa; }
    .lc { flex: 1; display: flex; flex-direction: column; gap: 14px; }
    .lc-icon { font-size: 3.5rem; animation: rocketAnim 2s ease-in-out infinite; opacity:0; animation: rocketAnim 2s ease-in-out 1s infinite, fadeUp .6s ease .2s forwards; }
    @keyframes rocketAnim { 0%,100%{transform:translateY(0) rotate(-5deg)} 50%{transform:translateY(-12px) rotate(5deg)} }
    h2 { font-size: 2rem; font-weight: 900; color: white; margin: 0; opacity:0; animation: fadeUp .6s ease .3s forwards; }
    .lc > p { color: rgba(255,255,255,.6); font-size: .92rem; line-height: 1.6; margin: 0; opacity:0; animation: fadeUp .6s ease .35s forwards; }
    @keyframes fadeUp { from{opacity:0;transform:translateY(14px)} to{opacity:1;transform:none} }

    .steps-list { display: flex; flex-direction: column; gap: 12px; margin-top: 8px; }
    .step-row {
        display: flex; align-items: center; gap: 14px;
        background: rgba(255,255,255,.07); border: 1px solid rgba(255,255,255,.1);
        border-radius: 12px; padding: 12px 16px;
        opacity: 0; animation: fadeUp .5s ease both;
    }
    .step-n {
        font-size: 1.3rem; font-weight: 900; min-width: 36px;
        background: linear-gradient(135deg,#6366f1,#a78bfa);
        -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text;
    }
    .step-t { color: white; font-weight: 700; font-size: .9rem; }
    .step-d { color: rgba(255,255,255,.45); font-size: .78rem; margin-top: 2px; }

    .left-ornament { position: absolute; inset: 0; pointer-events: none; }
    .orn {
        position: absolute; font-weight: 900;
        animation: ornDrift 6s ease-in-out var(--od,0s) infinite alternate;
    }
    @keyframes ornDrift { from{transform:translateY(0) rotate(0deg)} to{transform:translateY(-30px) rotate(180deg)} }

    /* ── RIGHT ── */
    .right-panel {
        background: #fafbff;
        display: flex; align-items: center; justify-content: center;
        padding: 32px 24px; overflow-y: auto;
    }
    .form-card {
        width: 100%; max-width: 420px;
        display: flex; flex-direction: column; gap: 18px;
        opacity: 0; transform: translateX(20px);
        transition: opacity .6s ease .1s, transform .6s cubic-bezier(.34,1.1,.64,1) .1s;
    }
    .mounted .form-card { opacity: 1; transform: none; }

    .form-header { margin-bottom: 2px; }
    h1 { font-size: 1.9rem; font-weight: 900; color: #0f172a; margin: 0 0 6px; letter-spacing: -.03em; }
    .form-header p { color: #64748b; font-size: .88rem; margin: 0; }

    /* Role cards */
    .role-wrap { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
    .role-card {
        position: relative; display: flex; flex-direction: column; align-items: center;
        gap: 4px; padding: 16px 12px;
        border: 2px solid #e2e8f0; border-radius: 14px;
        background: white; cursor: pointer;
        transition: all .2s; text-align: center;
    }
    .role-card:hover:not(.active) { border-color: #c4b5fd; background: #faf5ff; transform: translateY(-2px); }
    .role-card.active {
        border-color: #6366f1; background: linear-gradient(135deg,#f5f3ff,#ede9fe);
        box-shadow: 0 4px 16px rgba(99,102,241,.2);
        transform: translateY(-2px);
    }
    .role-ic { font-size: 1.8rem; }
    .role-t { font-weight: 700; font-size: .88rem; color: #0f172a; }
    .role-d { font-size: .73rem; color: #94a3b8; }
    .role-check {
        position: absolute; top: -8px; right: -8px;
        background: #6366f1; color: white;
        width: 22px; height: 22px; border-radius: 50%;
        display: flex; align-items: center; justify-content: center;
        font-size: .75rem; font-weight: 800;
        animation: checkPop .25s cubic-bezier(.34,1.56,.64,1);
        box-shadow: 0 2px 8px rgba(99,102,241,.4);
    }
    @keyframes checkPop { from{transform:scale(0)} to{transform:scale(1)} }

    form { display: flex; flex-direction: column; gap: 12px; }
    .field { display: flex; flex-direction: column; gap: 6px; }
    label {
        display: flex; align-items: center; gap: 5px;
        font-size: .76rem; font-weight: 700; color: #475569;
        text-transform: uppercase; letter-spacing: .06em;
        transition: color .2s;
    }
    .li { font-size: .88rem; }
    .field.focused label { color: #6366f1; }
    .iw { position: relative; }
    .iw input {
        width: 100%; padding: 12px 42px 12px 15px;
        border: 2px solid #e2e8f0; border-radius: 11px;
        font-size: .97rem; outline: none; background: white; color: #0f172a;
        transition: border-color .2s, box-shadow .2s;
    }
    .iw input:focus { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,.1); }
    .field.focused .iw input { border-color: #6366f1; }
    .fok {
        position: absolute; right: 12px; top: 50%; transform: translateY(-50%);
        color: #22c55e; font-weight: 800;
        animation: okPop .25s cubic-bezier(.34,1.56,.64,1);
    }
    @keyframes okPop { from{transform:translateY(-50%) scale(0)} to{transform:translateY(-50%) scale(1)} }
    .eye {
        position: absolute; right: 10px; top: 50%; transform: translateY(-50%);
        background: none; border: none; cursor: pointer; font-size: 1rem;
        padding: 4px; transition: transform .15s;
    }
    .eye:hover { transform: translateY(-50%) scale(1.2); }

    /* Password strength */
    .strength-wrap { display: flex; align-items: center; gap: 8px; margin-top: 4px; }
    .strength-bars { display: flex; gap: 4px; flex: 1; }
    .sbar { flex: 1; height: 4px; border-radius: 999px; transition: background .3s; }
    .strength-label { font-size: .72rem; font-weight: 700; min-width: 60px; text-align: right; transition: color .3s; }

    .error-box {
        display: flex; align-items: center; gap: 8px;
        padding: 10px 14px; border-radius: 10px;
        background: #fef2f2; border: 1px solid #fecaca;
        color: #dc2626; font-size: .86rem; font-weight: 600;
        animation: shake .4s ease;
    }
    @keyframes shake {
        0%,100%{transform:translateX(0)} 20%{transform:translateX(-5px)}
        40%{transform:translateX(5px)}   60%{transform:translateX(-3px)}
    }

    .submit-btn {
        position: relative; overflow: hidden;
        padding: 14px; border: none; border-radius: 13px;
        background: linear-gradient(135deg,#6366f1,#8b5cf6);
        color: white; font-size: 1rem; font-weight: 800;
        cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px;
        transition: transform .2s, box-shadow .2s, opacity .2s;
        box-shadow: 0 6px 20px rgba(99,102,241,.4);
        margin-top: 4px;
    }
    .submit-btn:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 10px 28px rgba(99,102,241,.5); }
    .submit-btn:active:not(:disabled){ transform: scale(.98); }
    .submit-btn:disabled { opacity: .5; cursor: not-allowed; }
    .btn-shine {
        position: absolute; inset: 0;
        background: linear-gradient(105deg,transparent 35%,rgba(255,255,255,.28) 50%,transparent 65%);
        transform: translateX(-100%);
        animation: shine 3s ease-in-out 1s infinite;
    }
    @keyframes shine { 0%,60%{transform:translateX(-100%)} 80%,100%{transform:translateX(200%)} }
    .spin { width:18px;height:18px;flex-shrink:0;border:2px solid rgba(255,255,255,.3);border-top-color:white;border-radius:50%;animation:spinAnim .7s linear infinite; }
    @keyframes spinAnim { to{transform:rotate(360deg)} }

    .footer-text { text-align: center; font-size: .84rem; color: #64748b; margin: 0; }
    .footer-text a { color: #6366f1; font-weight: 700; text-decoration: none; }
    .footer-text a:hover { text-decoration: underline; }

    @media (max-width: 768px) {
        .page { grid-template-columns: 1fr; }
        .left-panel { display: none; }
        .right-panel {
            min-height: 100dvh;
            background: linear-gradient(160deg,#0f172a,#1e1b4b 40%,#312e81);
            padding: 24px 20px;
            align-items: flex-start;
            padding-top: 48px;
        }
        .form-card {
            background: rgba(255,255,255,.97);
            border-radius: 24px; padding: 28px 22px;
            box-shadow: 0 32px 80px rgba(0,0,0,.35);
            opacity: 1 !important; transform: none !important;
        }
    }
</style>
