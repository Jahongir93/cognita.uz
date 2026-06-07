<script lang="ts">
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { authStore } from '$lib/stores/auth';

    let email = '';
    let password = '';
    let showPass = false;
    let error = '';
    let loading = false;
    let mounted = false;
    let focusedField = '';

    onMount(() => { mounted = true; });

    async function handleLogin() {
        if (!email || !password) { error = 'Email va parolni kiriting'; return; }
        error = '';
        loading = true;
        try {
            await authStore.login(email, password);
            goto('/dashboard');
        } catch (e: any) {
            error = e.message ?? 'Login xatosi';
        } finally {
            loading = false;
        }
    }
</script>

<svelte:head><title>Kirish — Cognita.uz</title></svelte:head>

<div class="page" class:mounted>
    <div class="bg" aria-hidden="true">
        <div class="blob b1"></div>
        <div class="blob b2"></div>
    </div>

    <div class="card">
        <a href="/" class="brand"><img src="/sitelogo.png" alt="Cognita.uz" style="height:38px;width:auto" /></a>

        <div class="card-header">
            <div class="hicon">🔐</div>
            <h1>Kirish</h1>
            <p>Davom etish uchun hisobingizga kiring</p>
        </div>

        <form on:submit|preventDefault={handleLogin}>
            <div class="field" class:focused={focusedField === 'email'}>
                <label for="email"><span class="li">📧</span> Email</label>
                <div class="iw">
                    <input id="email" type="email" placeholder="email@maktab.uz"
                           bind:value={email} autocomplete="email"
                           on:focus={() => focusedField='email'}
                           on:blur={() => focusedField=''} />
                    {#if email.includes('@')}<span class="ok">✓</span>{/if}
                </div>
            </div>

            <div class="field" class:focused={focusedField === 'pass'}>
                <label for="pass"><span class="li">🔒</span> Parol</label>
                <div class="iw">
                    {#if showPass}
                        <input id="pass" type="text"
                               placeholder="••••••••"
                               bind:value={password} autocomplete="current-password"
                               on:focus={() => focusedField='pass'}
                               on:blur={() => focusedField=''} />
                    {:else}
                        <input id="pass" type="password"
                               placeholder="••••••••"
                               bind:value={password} autocomplete="current-password"
                               on:focus={() => focusedField='pass'}
                               on:blur={() => focusedField=''} />
                    {/if}
                    <button type="button" class="eye"
                            on:click={() => showPass = !showPass}
                            tabindex="-1">
                        {showPass ? '🙈' : '👁'}
                    </button>
                </div>
            </div>

            {#if error}
                <div class="error-box" role="alert"><span>⚠</span> {error}</div>
            {/if}

            <button type="submit" class="submit-btn" disabled={loading}>
                {#if loading}
                    <span class="spin"></span> Kirmoqda...
                {:else}
                    <span>🚀</span> Kirish
                    <div class="shine"></div>
                {/if}
            </button>
        </form>

        <div class="divider"><span>yoki</span></div>

        <a href="/join" class="alt-btn">🎮 O'quvchi sifatida PIN bilan kirish</a>

        <p class="footer">
            Hisobingiz yo'qmi? <a href="/auth/register">Ro'yxatdan o'tish</a>
        </p>
    </div>
</div>

<style>
    :global(body) { margin: 0; font-family: 'Segoe UI', system-ui, sans-serif; }
    * { box-sizing: border-box; }

    .page {
        min-height: 100dvh;
        display: flex; align-items: center; justify-content: center;
        padding: 20px 16px;
        background: linear-gradient(160deg, #1e1b4b 0%, #312e81 50%, #4c1d95 100%);
        position: relative; overflow: hidden;
    }

    .bg { position: fixed; inset: 0; pointer-events: none; z-index: 0; }
    .blob { position: absolute; border-radius: 50%; filter: blur(80px); }
    .b1 { width: 500px; height: 500px; background: radial-gradient(#7c3aed, #6366f1); opacity: .25; top: -150px; right: -100px; animation: drift 12s ease-in-out infinite alternate; }
    .b2 { width: 350px; height: 350px; background: radial-gradient(#4f46e5, #312e81); opacity: .2;  bottom: -100px; left: -80px;   animation: drift 16s ease-in-out infinite alternate-reverse; }
    @keyframes drift { from { transform: translate(0,0); } to { transform: translate(30px, 20px) scale(1.1); } }

    .card {
        position: relative; z-index: 10;
        width: 100%; max-width: 420px;
        background: rgba(255,255,255,0.97);
        border-radius: 28px;
        padding: 36px 32px;
        box-shadow: 0 40px 100px rgba(0,0,0,0.35), 0 0 0 1px rgba(255,255,255,.1);
        display: flex; flex-direction: column; gap: 20px;
        opacity: 0; transform: translateY(24px) scale(0.97);
        transition: opacity .5s ease, transform .5s cubic-bezier(.34,1.2,.64,1);
    }
    .mounted .card { opacity: 1; transform: none; }

    .brand {
        display: flex; align-items: center; gap: 6px; justify-content: center;
        text-decoration: none; font-weight: 900; font-size: 1rem; color: #6366f1;
    }
    .dot { color: #a78bfa; }

    .card-header { text-align: center; }
    .hicon { font-size: 2.6rem; animation: bounce 2s ease-in-out infinite; }
    @keyframes bounce { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-6px)} }
    h1 { font-size: 1.9rem; font-weight: 900; color: #0f172a; margin: 8px 0 6px; letter-spacing: -.03em; }
    .card-header p { color: #64748b; font-size: .88rem; margin: 0; }

    form { display: flex; flex-direction: column; gap: 14px; }

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
        width: 100%; padding: 13px 44px 13px 16px;
        border: 2px solid #e2e8f0; border-radius: 12px;
        font-size: 1rem; outline: none; background: white; color: #0f172a;
        transition: border-color .2s, box-shadow .2s;
    }
    .iw input:focus { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,.12); }
    .field.focused .iw input { border-color: #6366f1; }

    .ok {
        position: absolute; right: 14px; top: 50%; transform: translateY(-50%);
        color: #22c55e; font-weight: 700;
        animation: okPop .25s cubic-bezier(.34,1.56,.64,1);
    }
    @keyframes okPop { from{transform:translateY(-50%) scale(0)} to{transform:translateY(-50%) scale(1)} }

    .eye {
        position: absolute; right: 10px; top: 50%; transform: translateY(-50%);
        background: none; border: none; cursor: pointer; font-size: 1rem;
        padding: 4px; border-radius: 6px; transition: transform .15s;
    }
    .eye:hover { transform: translateY(-50%) scale(1.2); }

    .error-box {
        display: flex; align-items: center; gap: 8px;
        padding: 10px 14px; border-radius: 10px;
        background: #fef2f2; border: 1px solid #fecaca;
        color: #dc2626; font-size: .88rem; font-weight: 600;
        animation: shake .4s ease;
    }
    @keyframes shake {
        0%,100%{transform:translateX(0)} 20%{transform:translateX(-5px)}
        40%{transform:translateX(5px)} 60%{transform:translateX(-3px)}
    }

    .submit-btn {
        position: relative; overflow: hidden;
        padding: 14px; border: none; border-radius: 13px;
        background: linear-gradient(135deg,#6366f1,#8b5cf6);
        color: white; font-size: 1rem; font-weight: 800;
        cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px;
        transition: transform .2s, box-shadow .2s, opacity .2s;
        box-shadow: 0 6px 20px rgba(99,102,241,.4);
    }
    .submit-btn:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 10px 28px rgba(99,102,241,.5); }
    .submit-btn:active:not(:disabled) { transform: scale(.98); }
    .submit-btn:disabled { opacity: .5; cursor: not-allowed; }
    .shine {
        position: absolute; inset: 0;
        background: linear-gradient(105deg,transparent 35%,rgba(255,255,255,.28) 50%,transparent 65%);
        transform: translateX(-100%);
        animation: shine 3s ease-in-out 1s infinite;
    }
    @keyframes shine { 0%,60%{transform:translateX(-100%)} 80%,100%{transform:translateX(200%)} }
    .spin { width:18px;height:18px;flex-shrink:0;border:2px solid rgba(255,255,255,.3);border-top-color:white;border-radius:50%;animation:spinAnim .7s linear infinite; }
    @keyframes spinAnim { to{transform:rotate(360deg)} }

    .divider {
        display: flex; align-items: center; gap: 12px;
        color: #94a3b8; font-size: .8rem;
    }
    .divider::before,.divider::after { content:''; flex:1; height:1px; background:#e2e8f0; }

    .alt-btn {
        display: flex; align-items: center; justify-content: center; gap: 8px;
        padding: 13px; border-radius: 13px;
        border: 2px solid #e2e8f0; background: white;
        color: #374151; font-weight: 700; font-size: .92rem;
        text-decoration: none;
        transition: border-color .2s, color .2s, transform .15s;
    }
    .alt-btn:hover { border-color: #6366f1; color: #6366f1; transform: translateY(-1px); }

    .footer { text-align: center; font-size: .85rem; color: #64748b; margin: 0; }
    .footer a { color: #6366f1; font-weight: 700; text-decoration: none; }
    .footer a:hover { text-decoration: underline; }

    @media (max-width: 480px) {
        .card { padding: 28px 20px; border-radius: 20px; }
        h1 { font-size: 1.6rem; }
    }
</style>
