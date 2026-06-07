<script lang="ts">
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { rooms } from '$lib/api/client';

    let pin = '';
    let nickname = '';
    let error = '';
    let loading = false;
    let mounted = false;

    const avatars = ['🐶','🐱','🐻','🦊','🐼','🐨','🦁','🐯','🐸','🐺','🦋','🐬','🦄','🐧','🦉','🐲'];
    let selectedAvatar = avatars[Math.floor(Math.random() * avatars.length)];

    let pinInputs: HTMLInputElement[] = [];

    onMount(() => {
        mounted = true;
        pinInputs[0]?.focus();
    });

    function handlePinKey(i: number, e: KeyboardEvent) {
        const input = e.target as HTMLInputElement;
        if (e.key === 'Backspace' && !input.value && i > 0) {
            pinInputs[i - 1].focus();
            pinInputs[i - 1].value = '';
            pin = pin.slice(0, -1);
        }
    }

    function handlePinInput(i: number, e: Event) {
        const val = (e.target as HTMLInputElement).value.replace(/\D/g, '');
        if (!val) { pinInputs[i].value = ''; return; }
        pinInputs[i].value = val.slice(-1);
        pin = pinInputs.map(inp => inp?.value ?? '').join('');
        if (val && i < 5) pinInputs[i + 1].focus();
    }

    function handlePinPaste(e: ClipboardEvent) {
        const text = e.clipboardData?.getData('text').replace(/\D/g, '').slice(0, 6) ?? '';
        if (!text) return;
        e.preventDefault();
        text.split('').forEach((ch, i) => { if (pinInputs[i]) pinInputs[i].value = ch; });
        pin = text;
        pinInputs[Math.min(text.length, 5)]?.focus();
    }

    async function joinGame() {
        if (pin.length < 4 || !nickname.trim()) {
            error = 'PIN va ismingizni kiriting';
            return;
        }
        error = '';
        loading = true;
        try {
            const info = await rooms.info(pin.trim());
            if (info.status === 'completed') { error = "Bu o'yin tugagan"; return; }
            sessionStorage.setItem('nickname', nickname.trim());
            sessionStorage.setItem('avatar', selectedAvatar);
            goto(`/game/play/${pin.trim()}`);
        } catch (e: any) {
            error = e.message ?? 'Xona topilmadi';
        } finally {
            loading = false;
        }
    }
</script>

<svelte:head><title>Cognita.uz — O'yinga kirish</title></svelte:head>

<div class="page" class:mounted>
    <div class="bg" aria-hidden="true">
        <div class="blob b1"></div>
        <div class="blob b2"></div>
        <div class="blob b3"></div>
    </div>

    <div class="card">
        <a href="/" class="brand"><img src="/sitelogo.png" alt="Cognita.uz" style="height:38px;width:auto" /></a>

        <div class="card-header">
            <div class="hicon">🎮</div>
            <h1>O'yinga kirish</h1>
            <p>PIN kodni kiriting va o'yin boshlang!</p>
        </div>

        <form on:submit|preventDefault={joinGame}>
            <!-- PIN boxes -->
            <div class="field">
                <label class="flabel"><span>🔢</span> O'yin PIN kodi</label>
                <div class="pin-row" on:paste={handlePinPaste}>
                    {#each Array(6) as _, i}
                        <input
                            bind:this={pinInputs[i]}
                            type="text"
                            inputmode="numeric"
                            maxlength="1"
                            class="pin-box"
                            class:filled={pinInputs[i]?.value}
                            on:input={(e) => handlePinInput(i, e)}
                            on:keydown={(e) => handlePinKey(i, e)}
                            autocomplete="off"
                        />
                    {/each}
                </div>
            </div>

            <!-- Nickname -->
            <div class="field">
                <label class="flabel" for="nick"><span>✏️</span> Ismingiz</label>
                <div class="iw">
                    <input id="nick" type="text" placeholder="Masalan: Alisher"
                           bind:value={nickname} maxlength="30"
                           class="inp" class:has-val={nickname.length > 0} />
                    {#if nickname.length > 0}<span class="ok">✓</span>{/if}
                </div>
            </div>

            <!-- Avatar -->
            <div class="field">
                <label class="flabel"><span>🐾</span> Avatar tanlang</label>
                <div class="av-grid">
                    {#each avatars as av, i}
                        <button type="button" class="av-btn"
                                class:sel={av === selectedAvatar}
                                on:click={() => selectedAvatar = av}
                                style="animation-delay:{i*25}ms" title={av}>
                            {av}
                        </button>
                    {/each}
                </div>
                <div class="sel-row">
                    <span class="sel-av">{selectedAvatar}</span>
                    <span class="sel-name">{nickname || 'Siz'}</span>
                </div>
            </div>

            {#if error}
                <div class="error-box" role="alert"><span>⚠</span> {error}</div>
            {/if}

            <button type="submit" class="submit-btn"
                    disabled={loading || pin.length < 4 || !nickname.trim()}>
                {#if loading}
                    <span class="spin"></span> Kirmoqda...
                {:else}
                    <span>🚀</span> O'yinga kirish
                    <div class="shine"></div>
                {/if}
            </button>
        </form>

        <p class="footer">
            O'qituvchimisiz? <a href="/auth/login">Hisobingizga kiring</a>
        </p>
    </div>
</div>

<style>
    :global(body) { margin: 0; font-family: 'Segoe UI', system-ui, sans-serif; }
    * { box-sizing: border-box; }

    .page {
        min-height: 100dvh;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 20px 16px;
        background: linear-gradient(160deg, #1e1b4b 0%, #312e81 50%, #4c1d95 100%);
        position: relative;
        overflow-y: auto;
    }

    .bg { position: fixed; inset: 0; pointer-events: none; z-index: 0; }
    .blob { position: absolute; border-radius: 50%; filter: blur(80px); }
    .b1 { width: 500px; height: 500px; background: radial-gradient(#7c3aed,#6366f1); opacity:.25; top:-150px; right:-100px; animation: drift 12s ease-in-out infinite alternate; }
    .b2 { width: 350px; height: 350px; background: radial-gradient(#4f46e5,#312e81); opacity:.2;  bottom:-100px; left:-80px;   animation: drift 16s ease-in-out infinite alternate-reverse; }
    .b3 { width: 220px; height: 220px; background: radial-gradient(#fbbf24,#f59e0b); opacity:.1;  top:40%; left:35%;           animation: drift 10s ease-in-out infinite alternate; }
    @keyframes drift { from{transform:translate(0,0)} to{transform:translate(30px,20px) scale(1.1)} }

    /* ── CARD ── */
    .card {
        position: relative; z-index: 10;
        width: 100%; max-width: 420px;
        background: rgba(255,255,255,0.97);
        border-radius: 24px;
        padding: 28px 24px;
        box-shadow: 0 32px 80px rgba(0,0,0,.4), 0 0 0 1px rgba(255,255,255,.08);
        display: flex; flex-direction: column; gap: 18px;
        opacity: 0; transform: translateY(20px) scale(.97);
        transition: opacity .5s ease, transform .5s cubic-bezier(.34,1.2,.64,1);
    }
    .mounted .card { opacity: 1; transform: none; }

    .brand {
        display: flex; align-items: center; gap: 6px; justify-content: center;
        text-decoration: none; font-weight: 900; font-size: 1rem; color: #6366f1;
    }
    .dot { color: #a78bfa; }

    .card-header { text-align: center; }
    .hicon { font-size: 2.4rem; animation: bounce 2s ease-in-out infinite; }
    @keyframes bounce { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-5px)} }
    h1 { font-size: 1.65rem; font-weight: 900; color: #0f172a; margin: 6px 0 4px; letter-spacing: -.02em; }
    .card-header p { color: #64748b; font-size: .85rem; margin: 0; }

    form { display: flex; flex-direction: column; gap: 16px; }

    .field { display: flex; flex-direction: column; gap: 7px; }
    .flabel {
        font-size: .75rem; font-weight: 700; color: #475569;
        display: flex; align-items: center; gap: 5px;
        text-transform: uppercase; letter-spacing: .06em;
    }

    /* ── PIN boxes: flex so they auto-resize ── */
    .pin-row { display: flex; gap: 6px; }
    .pin-box {
        flex: 1; min-width: 0;
        height: 52px;
        text-align: center;
        font-size: clamp(1.1rem, 4vw, 1.6rem);
        font-weight: 800;
        border: 2px solid #e2e8f0; border-radius: 10px;
        outline: none; background: #fafbff; color: #0f172a;
        transition: border-color .2s, background .2s, transform .15s, box-shadow .2s;
        caret-color: #6366f1;
    }
    .pin-box:focus {
        border-color: #6366f1; background: white;
        transform: scale(1.06);
        box-shadow: 0 0 0 3px rgba(99,102,241,.15);
    }
    .pin-box.filled { border-color: #a78bfa; background: #f5f3ff; color: #6366f1; }

    /* ── Nickname ── */
    .iw { position: relative; }
    .inp {
        width: 100%; padding: 12px 40px 12px 14px;
        border: 2px solid #e2e8f0; border-radius: 11px;
        font-size: .97rem; outline: none; background: #fafbff; color: #0f172a;
        transition: border-color .2s, box-shadow .2s;
    }
    .inp:focus { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,.12); }
    .inp.has-val { border-color: #a78bfa; }
    .ok {
        position: absolute; right: 12px; top: 50%; transform: translateY(-50%);
        color: #22c55e; font-weight: 700;
        animation: okPop .25s cubic-bezier(.34,1.56,.64,1);
    }
    @keyframes okPop { from{transform:translateY(-50%) scale(0)} to{transform:translateY(-50%) scale(1)} }

    /* ── Avatars ── */
    .av-grid {
        display: grid;
        grid-template-columns: repeat(8, 1fr);
        gap: 5px;
    }
    .av-btn {
        aspect-ratio: 1;
        font-size: clamp(1rem, 3.5vw, 1.35rem);
        padding: 0;
        border: 2px solid transparent; border-radius: 8px;
        background: #f8fafc; cursor: pointer; line-height: 1;
        transition: transform .15s, border-color .15s, background .15s;
        animation: avIn .3s ease both;
        display: flex; align-items: center; justify-content: center;
    }
    @keyframes avIn { from{opacity:0;transform:scale(.5)} to{opacity:1;transform:scale(1)} }
    .av-btn:hover:not(.sel) { transform: scale(1.15); background: #ede9fe; }
    .av-btn.sel {
        border-color: #6366f1; background: #ede9fe;
        box-shadow: 0 0 0 2px rgba(99,102,241,.25);
    }

    .sel-row {
        display: flex; align-items: center; gap: 8px;
        background: linear-gradient(135deg,#f5f3ff,#ede9fe);
        border: 1px solid rgba(99,102,241,.2);
        border-radius: 8px; padding: 7px 12px;
    }
    .sel-av  { font-size: 1.3rem; }
    .sel-name { font-weight: 700; color: #4f46e5; font-size: .9rem; }

    /* ── Error ── */
    .error-box {
        display: flex; align-items: center; gap: 8px;
        padding: 10px 14px; border-radius: 10px;
        background: #fef2f2; border: 1px solid #fecaca;
        color: #dc2626; font-size: .87rem; font-weight: 600;
        animation: shake .4s ease;
    }
    @keyframes shake {
        0%,100%{transform:translateX(0)} 20%{transform:translateX(-6px)}
        40%{transform:translateX(6px)} 60%{transform:translateX(-4px)}
    }

    /* ── Submit ── */
    .submit-btn {
        position: relative; overflow: hidden;
        padding: 14px; border: none; border-radius: 13px;
        background: linear-gradient(135deg,#6366f1,#8b5cf6);
        color: white; font-size: 1rem; font-weight: 800;
        cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px;
        transition: transform .2s, box-shadow .2s, opacity .2s;
        box-shadow: 0 6px 24px rgba(99,102,241,.4);
    }
    .submit-btn:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 10px 32px rgba(99,102,241,.5); }
    .submit-btn:active:not(:disabled) { transform: scale(.98); }
    .submit-btn:disabled { opacity: .45; cursor: not-allowed; }
    .shine {
        position: absolute; inset: 0;
        background: linear-gradient(105deg,transparent 30%,rgba(255,255,255,.3) 50%,transparent 70%);
        transform: translateX(-100%);
        animation: shine 3s ease-in-out 1s infinite;
    }
    @keyframes shine { 0%,60%{transform:translateX(-100%)} 80%,100%{transform:translateX(200%)} }
    .spin { width:18px;height:18px;flex-shrink:0;border:2px solid rgba(255,255,255,.35);border-top-color:white;border-radius:50%;animation:spinAnim .7s linear infinite; }
    @keyframes spinAnim { to{transform:rotate(360deg)} }

    .footer { text-align: center; font-size: .84rem; color: #64748b; margin: 0; }
    .footer a { color: #6366f1; font-weight: 700; text-decoration: none; }
    .footer a:hover { text-decoration: underline; }

    @media (max-width: 400px) {
        .card { padding: 22px 16px; border-radius: 18px; }
        .pin-box { height: 44px; }
    }
</style>
