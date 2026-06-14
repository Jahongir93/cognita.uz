<script lang="ts">
    import { onMount } from 'svelte';

    let pin = '';
    let scrolled = false;
    let heroVisible = false;
    let wordIdx = 0;
    let statsVisible = false;
    let featVisible = false;

    // Cycling hero words
    const heroWords = ["qiziqarli!", "zavqli!", "ajoyib!", "o'yin!"];
    let currentWord = heroWords[0];
    let wordFading = false;

    // Animated counters
    let count1 = 0, count2 = 0, count3 = 0;
    const targets = [10000, 500, 25000];

    // Floating particles config
    const particles = Array.from({ length: 18 }, (_, i) => ({
        id: i,
        x: Math.random() * 100,
        y: Math.random() * 100,
        size: 10 + Math.random() * 24,
        dur: 6 + Math.random() * 10,
        delay: Math.random() * 8,
        color: ['#6366f1','#a78bfa','#f59e0b','#e21b3c','#1368ce','#26890c','#ec4899'][i % 7],
        shape: ['▲','◆','●','■','★'][i % 5],
        opacity: 0.12 + Math.random() * 0.15,
    }));

    function easeOutCubic(t: number) { return 1 - Math.pow(1 - t, 3); }

    function animateCounter(setter: (v: number) => void, target: number, duration = 2000) {
        const start = performance.now();
        function tick(now: number) {
            const t = Math.min((now - start) / duration, 1);
            setter(Math.round(easeOutCubic(t) * target));
            if (t < 1) requestAnimationFrame(tick);
        }
        requestAnimationFrame(tick);
    }

    onMount(() => {
        setTimeout(() => { heroVisible = true; }, 80);

        // Word cycler
        const wordTimer = setInterval(() => {
            wordFading = true;
            setTimeout(() => {
                wordIdx = (wordIdx + 1) % heroWords.length;
                currentWord = heroWords[wordIdx];
                wordFading = false;
            }, 350);
        }, 2600);

        // Scroll
        const onScroll = () => { scrolled = window.scrollY > 20; };
        window.addEventListener('scroll', onScroll, { passive: true });

        // IntersectionObserver for stats & features
        const io = new IntersectionObserver((entries) => {
            entries.forEach(e => {
                if (e.isIntersecting) {
                    if (e.target.classList.contains('stats-section')) {
                        statsVisible = true;
                        animateCounter(v => count1 = v, targets[0]);
                        animateCounter(v => count2 = v, targets[1]);
                        animateCounter(v => count3 = v, targets[2]);
                    }
                    if (e.target.classList.contains('features-section')) featVisible = true;
                    io.unobserve(e.target);
                }
            });
        }, { threshold: 0.2 });

        document.querySelectorAll('.stats-section, .features-section').forEach(el => io.observe(el));

        return () => {
            clearInterval(wordTimer);
            window.removeEventListener('scroll', onScroll);
            io.disconnect();
        };
    });

    function joinGame() {
        if (pin.trim().length >= 4) window.location.href = `/game/play/${pin.trim()}`;
    }
    function fmt(n: number) { return n.toLocaleString('ru-RU'); }
</script>

<svelte:head>
    <title>Cognita.uz — Interaktiv o'quv platformasi</title>
    <meta name="description" content="O'zbekiston maktablari uchun interaktiv quiz va o'yinli o'qitish platformasi" />
</svelte:head>

<!-- ═══ NAV ══════════════════════════════════════════════════════════════════ -->
<nav class="nav" class:scrolled>
    <a href="/" class="brand">
        <img src="/sitelogo.png" alt="Cognita.uz" style="height:36px;width:auto;display:block" />
    </a>
    <div class="nav-links">
        <a href="/join" class="nav-link">O'yinga kirish</a>
        <a href="/auth/login" class="nav-link">Kirish</a>
        <a href="/auth/register" class="nav-cta">
            <span>Bepul boshlash</span>
            <span class="cta-shine"></span>
        </a>
    </div>
</nav>

<!-- ═══ HERO ══════════════════════════════════════════════════════════════════ -->
<section class="hero">
    <!-- Mesh bg -->
    <div class="mesh" aria-hidden="true"></div>

    <!-- Floating particles -->
    <div class="particles" aria-hidden="true">
        {#each particles as p}
            <div class="particle"
                 style="left:{p.x}%;top:{p.y}%;font-size:{p.size}px;color:{p.color};opacity:{p.opacity};
                        animation:partFloat {p.dur}s ease-in-out {p.delay}s infinite alternate,
                                   partSpin {p.dur * 1.5}s linear {p.delay}s infinite">
                {p.shape}
            </div>
        {/each}
    </div>

    <!-- Blobs -->
    <div class="blob bl1" aria-hidden="true"></div>
    <div class="blob bl2" aria-hidden="true"></div>
    <div class="blob bl3" aria-hidden="true"></div>

    <div class="hero-inner" class:visible={heroVisible}>
        <!-- Left -->
        <div class="hero-left">
            <div class="hero-badge">
                <span class="badge-dot"></span>
                🇺🇿 O'zbekiston uchun yaratilgan
            </div>

            <h1 class="hero-title">
                O'rganish endi<br>
                <span class="word-wrap">
                    <span class="grad-word" class:fading={wordFading}>{currentWord}</span>
                </span>
            </h1>

            <p class="hero-sub">
                Kahoot uslubidagi live quiz, AI savol generatori, sinf boshqaruvi —
                hammasi bir joyda. O'qituvchilar uchun <strong>bepul</strong>.
            </p>

            <div class="hero-actions">
                <a href="/auth/register?role=teacher" class="btn-hero-primary">
                    <span class="btn-glow"></span>
                    🚀 O'qituvchi sifatida boshlash
                </a>
                <a href="/join" class="btn-hero-ghost">O'yinga kirish →</a>
            </div>

            <div class="pin-join">
                <p class="pin-label">Yoki PIN orqali kiring:</p>
                <div class="pin-row">
                    <div class="pin-input-wrap" class:has-val={pin.length > 0}>
                        <input
                            type="text"
                            inputmode="numeric"
                            maxlength="8"
                            placeholder="PIN kodni kiriting"
                            bind:value={pin}
                            on:keydown={e => e.key === 'Enter' && joinGame()}
                            class="pin-input"
                        />
                        <div class="pin-glow"></div>
                    </div>
                    <button class="pin-btn" on:click={joinGame} disabled={pin.trim().length < 4}>
                        Kirish ↗
                    </button>
                </div>
            </div>
        </div>

        <!-- Right: live game mockup -->
        <div class="hero-right" aria-hidden="true">
            <div class="mockup-wrap">
                <!-- Glow ring behind illustration -->
                <div class="card-glow"></div>

                <img src="/img/hero-main.png" alt="O'qituvchi va o'quvchilar interaktiv darsда" class="hero-illu" />

                <!-- Floating result pop-ups -->
                <div class="pop p1">✅ +450 ball</div>
                <div class="pop p2">🔥 ×4 streak!</div>
                <div class="pop p3">🥇 1-o'rin</div>

                <!-- Emoji reactions floating up -->
                <div class="emoji-float ef1">👏</div>
                <div class="emoji-float ef2">🔥</div>
                <div class="emoji-float ef3">😮</div>
                <div class="emoji-float ef4">❤️</div>
            </div>
        </div>
    </div>

    <!-- Scroll hint -->
    <div class="scroll-hint" aria-hidden="true">
        <div class="scroll-line"></div>
    </div>
</section>

<!-- ═══ STATS ════════════════════════════════════════════════════════════════ -->
<section class="stats-section">
    <div class="stats-inner">
        <div class="stat-card" class:visible={statsVisible} style="--d:0ms">
            <span class="stat-n">{fmt(count1)}+</span>
            <span class="stat-l">O'quvchilar</span>
            <div class="stat-bar" style="--w:85%"></div>
        </div>
        <div class="stat-card" class:visible={statsVisible} style="--d:100ms">
            <span class="stat-n">{fmt(count2)}+</span>
            <span class="stat-l">O'qituvchilar</span>
            <div class="stat-bar" style="--w:60%"></div>
        </div>
        <div class="stat-card" class:visible={statsVisible} style="--d:200ms">
            <span class="stat-n">{fmt(count3)}+</span>
            <span class="stat-l">Quizlar o'ynalgan</span>
            <div class="stat-bar" style="--w:100%"></div>
        </div>
        <div class="stat-card" class:visible={statsVisible} style="--d:300ms">
            <span class="stat-n stat-free">Bepul</span>
            <span class="stat-l">O'qituvchilar uchun</span>
            <div class="stat-bar" style="--w:100%;--bc:linear-gradient(90deg,#22c55e,#16a34a)"></div>
        </div>
    </div>
</section>

<!-- ═══ FEATURES ════════════════════════════════════════════════════════════ -->
<section class="features-section">
    <div class="sec-inner">
        <div class="sec-label">Imkoniyatlar</div>
        <h2 class="sec-title">Maktab uchun yaratilgan platforma</h2>
        <p class="sec-sub">Zamonaviy o'qitish usullari bilan o'quvchilarning qiziqishini oshiring</p>

        <div class="feat-grid">
            {#each [
                { img:'feat-live.png',     grad:'#6366f1,#8b5cf6', title:'Live Multiplayer',     desc:'Real-time o\'yin rejimi. Barcha o\'quvchilar bir vaqtda qatnashadi, leaderboard jonli yangilanadi.', tag:'Asosiy' },
                { img:'feat-ai.png',       grad:'#f59e0b,#ef4444', title:'AI Savol Generatori', desc:'Mavzu va sinf darajasini kiriting — Groq AI bir zumda 10-20 savol tayyorlaydi.',          tag:'Yangi' },
                { img:'feat-reports.png',  grad:'#22c55e,#16a34a', title:'Batafsil Hisobotlar',  desc:'Har bir o\'quvchi, sinf va savol bo\'yicha to\'liq analitika. Zaif tomonlarni aniqlang.',  tag:'' },
                { img:'feat-modes.png',    grad:'#06b6d4,#3b82f6', title:'Ko\'p O\'yin Rejimlari',desc:'Classic, Mustaqil, Jamoaviy — har darsga mos rejim tanlang.',                           tag:'' },
                { img:'feat-plickers.png', grad:'#ec4899,#f43f5e', title:'Doska o\'yinlari',     desc:'34 ta interaktiv doska o\'yini — o\'quvchilar elektron doskada ishlaydi.',                tag:'' },
                { img:'feat-homework.png', grad:'#a855f7,#6366f1', title:'Uy Vazifalari',         desc:'Vaqtni belgilang, o\'quvchilar o\'z temp\'larida ishlaydi. Natijalar avtomatik yig\'iladi.',tag:'' },
            ] as f, i}
                <div class="feat-card" class:visible={featVisible} style="--fd:{i * 80}ms">
                    <div class="feat-icon">
                        <img src="/img/{f.img}" alt="" class="feat-img" loading="lazy" />
                        {#if f.tag}<span class="feat-tag">{f.tag}</span>{/if}
                    </div>
                    <h3>{f.title}</h3>
                    <p>{f.desc}</p>
                    <div class="feat-line" style="--g:linear-gradient(90deg,{f.grad})"></div>
                </div>
            {/each}
        </div>
    </div>
</section>

<!-- ═══ HOW IT WORKS ═════════════════════════════════════════════════════════ -->
<section class="how-section">
    <div class="sec-inner">
        <div class="sec-label">Qanday ishlaydi</div>
        <h2 class="sec-title">3 qadamda tayyor</h2>

        <div class="steps">
            <div class="step-card">
                <div class="step-num-wrap">
                    <span class="step-num">01</span>
                    <div class="step-ring"></div>
                </div>
                <img src="/img/step-create.png" alt="" class="step-img" loading="lazy" />
                <h3>Quiz yarating</h3>
                <p>Savollarni qo'lda kiriting yoki AI yordamida avtomatik yarating. Rasm, video qo'shish mumkin.</p>
            </div>
            <div class="step-connector">
                <div class="conn-line"></div>
                <div class="conn-arrow">›</div>
            </div>
            <div class="step-card">
                <div class="step-num-wrap">
                    <span class="step-num">02</span>
                    <div class="step-ring" style="animation-delay:0.5s"></div>
                </div>
                <img src="/img/step-share.png" alt="" class="step-img" loading="lazy" />
                <h3>PIN ulashing</h3>
                <p>O'quvchilar cognita.uz/join ga kirib PIN kodni tераdilar. Qurilma farqi yo'q.</p>
            </div>
            <div class="step-connector">
                <div class="conn-line"></div>
                <div class="conn-arrow">›</div>
            </div>
            <div class="step-card">
                <div class="step-num-wrap">
                    <span class="step-num">03</span>
                    <div class="step-ring" style="animation-delay:1s"></div>
                </div>
                <img src="/img/step-play.png" alt="" class="step-img" loading="lazy" />
                <h3>O'ynang!</h3>
                <p>Jonli leaderboard, ball tizimi, streak bonuslari. O'rganish hech qachon bu qadar qiziqarli bo'lmagan.</p>
            </div>
        </div>
    </div>
</section>

<!-- ═══ CATEGORIES ════════════════════════════════════════════════════════════ -->
<section class="cat-section">
    <div class="sec-inner">
        <div class="sec-label">Kategoriyalar</div>
        <h2 class="sec-title">Nima qiziqtiradi?</h2>
        <p class="sec-sub">Test, quiz va o'yinlarning katta to'plamidan o'zingizga mosini tanlang</p>

        <div class="cat-grid">
            {#each [
                { href:'/tests/fun',         img:'cat-fun.png',         g:'#f59e0b,#ef4444', title:'Qiziqarli testlar',   desc:"Ko'ngil ochish, shaxsiyat va qiziqarli faktlar",      cnt:'120+ test'  },
                { href:'/tests/subjects',    img:'cat-subjects.png',    g:'#3b82f6,#6366f1', title:'Fan testlari',         desc:'Matematika, Fizika, Kimyo, Tarix va boshqa fanlar',   cnt:'500+ test'  },
                { href:'/tests/attestation', img:'cat-attestation.png', g:'#22c55e,#0ea5e9', title:'Attestatsiya',          desc:"O'qituvchilar uchun attestatsiya tayyorgarligi",       cnt:'200+ savol' },
                { href:'/tests/iq',          img:'cat-iq.png',          g:'#8b5cf6,#6366f1', title:'IQ testlar',            desc:'Mantiq, xotira va aqliy qobiliyat testlari',          cnt:'50+ test'   },
                { href:'/tests/psychology',  img:'cat-psychology.png',  g:'#ec4899,#8b5cf6', title:'Psixologik testlar',   desc:"Shaxsiyat, his-tuyg'u va xarakter tahlili",           cnt:'80+ test'   },
                { href:'/games',             img:'cat-games.png',       g:'#f59e0b,#22c55e', title:"O'yinlar",              desc:"So'z topish, krossvord, sudoku va boshqa o'yinlar",    cnt:"30+ o'yin"  },
            ] as c, i}
                <a href={c.href} class="cat-card" style="--i:{i};--g:{c.g}">
                    <div class="cat-icon-wrap">
                        <img src="/img/{c.img}" alt="" class="cat-img" loading="lazy" />
                        <div class="cat-icon-bg"></div>
                    </div>
                    <div class="cat-body">
                        <h3>{c.title}</h3>
                        <p>{c.desc}</p>
                    </div>
                    <div class="cat-foot">
                        <span class="cat-cnt">{c.cnt}</span>
                        <span class="cat-arr">→</span>
                    </div>
                    <div class="cat-glow"></div>
                </a>
            {/each}
        </div>
    </div>
</section>

<!-- ═══ MARQUEE TICKER ════════════════════════════════════════════════════════ -->
<div class="ticker" aria-hidden="true">
    <div class="ticker-track">
        {#each Array(3) as _}
            <span>⚡ Live Quiz</span>
            <span>🤖 AI Generator</span>
            <span>🏆 Leaderboard</span>
            <span>🎯 Real-time</span>
            <span>📊 Analytics</span>
            <span>🔥 Streak Bonus</span>
            <span>🇺🇿 O'zbekiston</span>
            <span>🎮 Cognita.uz</span>
        {/each}
    </div>
</div>

<!-- ═══ CTA ═══════════════════════════════════════════════════════════════════ -->
<section class="cta-section">
    <div class="cta-mesh" aria-hidden="true"></div>
    <div class="cta-particles" aria-hidden="true">
        {#each Array(12) as _, i}
            <div class="cta-spark"
                 style="left:{8 + i * 8}%;animation-delay:{i * 0.3}s;animation-duration:{3 + (i % 3)}s">
            </div>
        {/each}
    </div>
    <div class="cta-inner">
        <img src="/img/mascot.png" alt="" class="cta-mascot" />
        <div class="cta-badge">🎉 Bugun 500+ o'qituvchi foydalanmoqda</div>
        <h2 class="cta-title">Darsni qiziqarli o'yinga aylantiring</h2>
        <p class="cta-sub">Ro'yxatdan o'tish bir daqiqa, birinchi quiz yaratish esa uch daqiqa oladi.</p>
        <div class="cta-btns">
            <a href="/auth/register?role=teacher" class="btn-cta-main">
                <span class="sparkle s1">✦</span>
                <span class="sparkle s2">✦</span>
                <span class="sparkle s3">✦</span>
                🚀 Hoziroq boshlash — bepul!
            </a>
            <a href="/join" class="btn-cta-ghost">PIN orqali kirish</a>
        </div>
        <p class="cta-note">Kredit karta talab etilmaydi</p>
    </div>
</section>

<!-- ═══ FOOTER ════════════════════════════════════════════════════════════════ -->
<footer class="footer">
    <div class="footer-inner">
        <div class="footer-left">
            <div class="footer-brand"><img src="/logowhite.png" alt="Cognita.uz" style="height:30px;width:auto" /></div>
            <p>© 2025 Cognita.uz — O'zbekiston maktablari uchun</p>
        </div>
        <div class="footer-links">
            <a href="/auth/login">Kirish</a>
            <a href="/auth/register">Ro'yxatdan o'tish</a>
            <a href="/join">O'yinga kirish</a>
            <a href="/dashboard">Dashboard</a>
        </div>
    </div>
</footer>

<style>
    :global(html) { scroll-behavior: smooth; }
    :global(body) { margin: 0; font-family: 'Segoe UI', system-ui, -apple-system, sans-serif; overflow-x: hidden; }
    * { box-sizing: border-box; }

    /* ══ NAV ══ */
    .nav {
        position: fixed; top: 0; left: 0; right: 0; z-index: 200;
        display: flex; align-items: center; justify-content: space-between;
        padding: 18px 52px;
        transition: all 0.35s cubic-bezier(0.4,0,0.2,1);
    }
    .nav.scrolled {
        background: rgba(255,255,255,0.88);
        backdrop-filter: blur(16px) saturate(180%);
        box-shadow: 0 1px 0 rgba(0,0,0,0.06), 0 4px 24px rgba(99,102,241,0.08);
        padding: 10px 52px;
    }
    .brand { display: flex; align-items: center; gap: 8px; text-decoration: none; }
    .brand-icon { font-size: 1.6rem; animation: brandWiggle 3s ease-in-out infinite; }
    @keyframes brandWiggle {
        0%,100% { transform: rotate(0deg); }
        20%     { transform: rotate(-8deg); }
        40%     { transform: rotate(8deg); }
        60%     { transform: rotate(-4deg); }
        80%     { transform: rotate(4deg); }
    }
    .brand-name { font-size: 1.3rem; font-weight: 900; color: #0f172a; letter-spacing: -0.02em; }
    .dot { color: #6366f1; }
    .nav-links { display: flex; align-items: center; gap: 6px; }
    .nav-link { padding: 8px 14px; color: #475569; font-weight: 500; font-size: 0.9rem; text-decoration: none; border-radius: 8px; transition: all 0.2s; }
    .nav-link:hover { color: #6366f1; background: rgba(99,102,241,0.08); }
    .nav-cta {
        position: relative; overflow: hidden;
        padding: 9px 22px;
        background: linear-gradient(135deg, #6366f1, #8b5cf6);
        color: white; text-decoration: none; font-weight: 700; font-size: 0.88rem;
        border-radius: 10px;
        transition: transform 0.2s, box-shadow 0.2s;
        box-shadow: 0 4px 14px rgba(99,102,241,0.4);
    }
    .nav-cta:hover { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(99,102,241,0.5); }
    .cta-shine {
        position: absolute; inset: 0;
        background: linear-gradient(105deg, transparent 40%, rgba(255,255,255,0.35) 50%, transparent 60%);
        transform: translateX(-100%);
        animation: shine 3s ease-in-out infinite;
    }
    @keyframes shine { 0%,70% { transform: translateX(-100%); } 85%,100% { transform: translateX(200%); } }

    /* ══ HERO ══ */
    .hero {
        position: relative; min-height: 100dvh; overflow: hidden;
        display: flex; flex-direction: column; align-items: center; justify-content: center;
        background: #fafbff; padding-top: 80px;
    }
    .mesh {
        position: absolute; inset: 0; pointer-events: none;
        background:
            radial-gradient(ellipse 80% 50% at 70% -10%, rgba(139,92,246,0.18) 0%, transparent 60%),
            radial-gradient(ellipse 60% 60% at -10% 60%, rgba(99,102,241,0.14) 0%, transparent 60%),
            radial-gradient(ellipse 40% 40% at 90% 90%, rgba(251,191,36,0.1) 0%, transparent 60%),
            radial-gradient(ellipse 50% 30% at 50% 50%, rgba(167,139,250,0.07) 0%, transparent 70%);
        animation: meshPulse 8s ease-in-out infinite alternate;
    }
    @keyframes meshPulse {
        from { opacity: 0.7; }
        to   { opacity: 1; }
    }

    .particles { position: absolute; inset: 0; pointer-events: none; overflow: hidden; }
    .particle {
        position: absolute;
        font-weight: 900;
        user-select: none;
        will-change: transform;
    }
    @keyframes partFloat {
        from { transform: translateY(0px) translateX(0px); }
        to   { transform: translateY(-40px) translateX(20px); }
    }
    @keyframes partSpin {
        from { transform: rotate(0deg); }
        to   { transform: rotate(360deg); }
    }

    .blob {
        position: absolute; border-radius: 50%; filter: blur(90px);
        pointer-events: none; will-change: transform;
    }
    .bl1 { width: 700px; height: 700px; background: radial-gradient(circle, #a78bfa, #6366f1); opacity: 0.25; top: -250px; right: -180px; animation: blobDrift 14s ease-in-out infinite alternate; }
    .bl2 { width: 500px; height: 500px; background: radial-gradient(circle, #6366f1, #4f46e5); opacity: 0.18; bottom: -150px; left: -120px; animation: blobDrift 18s ease-in-out infinite alternate-reverse; }
    .bl3 { width: 350px; height: 350px; background: radial-gradient(circle, #fbbf24, #f59e0b); opacity: 0.12; top: 50%; left: 38%; animation: blobDrift 11s ease-in-out infinite alternate; }
    @keyframes blobDrift {
        from { transform: translate(0, 0) scale(1); }
        to   { transform: translate(40px, 30px) scale(1.15); }
    }

    .hero-inner {
        position: relative; z-index: 10;
        width: 100%; max-width: 1240px; margin: 0 auto;
        padding: 40px 52px 60px;
        display: grid; grid-template-columns: 1fr 1fr;
        align-items: center; gap: 60px;
        opacity: 0; transform: translateY(32px);
        transition: opacity 0.8s cubic-bezier(0.4,0,0.2,1), transform 0.8s cubic-bezier(0.4,0,0.2,1);
    }
    .hero-inner.visible { opacity: 1; transform: none; }

    .hero-badge {
        display: inline-flex; align-items: center; gap: 8px;
        padding: 6px 16px; border-radius: 999px; margin-bottom: 22px;
        background: rgba(99,102,241,0.08); color: #6366f1;
        border: 1px solid rgba(99,102,241,0.2); font-size: 0.82rem; font-weight: 600;
        animation: badgePulse 3s ease-in-out infinite;
    }
    @keyframes badgePulse {
        0%,100% { box-shadow: 0 0 0 0 rgba(99,102,241,0.15); }
        50%     { box-shadow: 0 0 0 6px rgba(99,102,241,0); }
    }
    .badge-dot {
        width: 6px; height: 6px; border-radius: 50%;
        background: #22c55e;
        animation: dotBlink 1.5s ease-in-out infinite;
    }
    @keyframes dotBlink { 0%,100% { opacity: 1; } 50% { opacity: 0.3; } }

    .hero-title {
        font-size: clamp(2.4rem, 5vw, 3.8rem);
        font-weight: 900; line-height: 1.1; color: #0f172a;
        margin: 0 0 20px; letter-spacing: -0.03em;
    }
    .word-wrap {
        display: inline-block; position: relative;
        background: linear-gradient(135deg, #6366f1, #a78bfa 40%, #f59e0b 80%);
        background-size: 200% auto;
        -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text;
        animation: gradMove 4s linear infinite;
    }
    @keyframes gradMove { 0% { background-position: 0% center; } 100% { background-position: 200% center; } }
    .grad-word { display: inline-block; transition: opacity 0.3s ease, transform 0.3s ease; }
    .grad-word.fading { opacity: 0; transform: translateY(-8px); }

    .hero-sub { font-size: 1.05rem; color: #475569; line-height: 1.7; margin: 0 0 30px; max-width: 460px; }
    .hero-sub strong { color: #6366f1; }

    .hero-actions { display: flex; gap: 12px; flex-wrap: wrap; margin-bottom: 32px; }

    .btn-hero-primary {
        position: relative; overflow: hidden;
        display: inline-flex; align-items: center; gap: 8px;
        padding: 14px 28px; border-radius: 14px;
        background: linear-gradient(135deg, #6366f1, #8b5cf6);
        color: white; font-weight: 800; font-size: 0.97rem;
        text-decoration: none; transition: transform 0.2s, box-shadow 0.2s;
        box-shadow: 0 6px 24px rgba(99,102,241,0.45), 0 0 0 0 rgba(99,102,241,0.3);
        animation: heroBtnPulse 2.5s ease-in-out infinite;
    }
    @keyframes heroBtnPulse {
        0%,100% { box-shadow: 0 6px 24px rgba(99,102,241,0.45), 0 0 0 0 rgba(99,102,241,0.3); }
        50%     { box-shadow: 0 6px 24px rgba(99,102,241,0.45), 0 0 0 10px rgba(99,102,241,0); }
    }
    .btn-hero-primary:hover { transform: translateY(-3px); }
    .btn-glow {
        position: absolute; inset: 0;
        background: linear-gradient(105deg, transparent 30%, rgba(255,255,255,0.3) 50%, transparent 70%);
        transform: translateX(-100%);
        animation: btnShine 3.5s ease-in-out 1s infinite;
    }
    @keyframes btnShine { 0%,65% { transform: translateX(-100%); } 80%,100% { transform: translateX(200%); } }

    .btn-hero-ghost {
        display: inline-flex; align-items: center;
        padding: 14px 24px; border-radius: 14px;
        border: 2px solid #e2e8f0; background: white;
        color: #374151; font-weight: 700; font-size: 0.95rem;
        text-decoration: none; transition: all 0.2s;
    }
    .btn-hero-ghost:hover { border-color: #6366f1; color: #6366f1; transform: translateY(-2px); box-shadow: 0 4px 16px rgba(99,102,241,0.15); }

    /* PIN input */
    .pin-join { border-top: 1px solid #e2e8f0; padding-top: 24px; }
    .pin-label { color: #94a3b8; font-size: 0.78rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.08em; margin: 0 0 10px; }
    .pin-row { display: flex; gap: 8px; max-width: 330px; }
    .pin-input-wrap { flex: 1; position: relative; }
    .pin-glow {
        position: absolute; inset: -2px; border-radius: 12px;
        background: linear-gradient(135deg, #6366f1, #a78bfa);
        opacity: 0; z-index: -1; transition: opacity 0.3s;
        filter: blur(8px);
    }
    .pin-input-wrap.has-val .pin-glow { opacity: 0.5; }
    .pin-input {
        width: 100%; padding: 13px 16px;
        border: 2px solid #e2e8f0; border-radius: 11px;
        font-size: 1.05rem; font-weight: 700; letter-spacing: 0.12em;
        outline: none; background: white; position: relative; z-index: 1;
        transition: border-color 0.25s, box-shadow 0.25s;
    }
    .pin-input:focus { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,0.15); }
    .pin-btn {
        padding: 13px 22px; border: none; border-radius: 11px;
        background: linear-gradient(135deg, #6366f1, #8b5cf6);
        color: white; font-weight: 800; font-size: 0.95rem; cursor: pointer;
        transition: opacity 0.2s, transform 0.15s;
        white-space: nowrap;
    }
    .pin-btn:disabled { opacity: 0.4; cursor: not-allowed; }
    .pin-btn:not(:disabled):hover { opacity: 0.88; transform: translateY(-1px); }

    /* ── Game card mockup ── */
    .hero-right { display: flex; align-items: center; justify-content: center; }
    .mockup-wrap { position: relative; }
    .card-glow {
        position: absolute; inset: -30px;
        background: radial-gradient(ellipse at center, rgba(99,102,241,0.3), transparent 70%);
        border-radius: 50%; filter: blur(20px);
        animation: glowPulse 3s ease-in-out infinite;
    }
    @keyframes glowPulse {
        0%,100% { opacity: 0.6; transform: scale(1); }
        50%      { opacity: 1; transform: scale(1.08); }
    }
    .hero-illu {
        position: relative; z-index: 1;
        width: 100%; height: auto; display: block;
        filter: drop-shadow(0 24px 50px rgba(0,0,0,0.35));
        animation: heroFloat 6s ease-in-out infinite;
    }
    @keyframes heroFloat {
        0%,100% { transform: translateY(0) rotate(-0.5deg); }
        50%     { transform: translateY(-16px) rotate(0.5deg); }
    }
    .game-card {
        position: relative; z-index: 1;
        background: linear-gradient(160deg, #1e293b, #0f172a);
        border: 1px solid rgba(255,255,255,0.08);
        border-radius: 22px; padding: 22px;
        width: 340px;
        box-shadow: 0 40px 100px rgba(0,0,0,0.35), 0 0 0 1px rgba(255,255,255,0.05);
        animation: cardHover 5s ease-in-out infinite;
    }
    @keyframes cardHover {
        0%,100% { transform: translateY(0px) rotateY(-2deg) rotateX(1deg); }
        50%     { transform: translateY(-14px) rotateY(-2deg) rotateX(1deg); }
    }
    .gc-bar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px; }
    .gc-qn  { font-size: 0.72rem; color: #64748b; font-weight: 600; }
    .gc-pts { font-size: 0.72rem; color: #fbbf24; background: rgba(251,191,36,0.12); padding: 3px 10px; border-radius: 999px; font-weight: 700; }
    .gc-timer { position: relative; width: 44px; height: 44px; display: flex; align-items: center; justify-content: center; }
    .gc-timer svg { position: absolute; inset: 0; transform: rotate(-90deg); width: 100%; height: 100%; }
    .gc-timer span { font-size: 0.9rem; font-weight: 800; color: #fbbf24; }
    @keyframes timerShrink {
        from { stroke-dasharray: 100 0; }
        to   { stroke-dasharray: 0 100; }
    }
    .gc-q { color: #f1f5f9; font-size: 0.92rem; font-weight: 700; text-align: center; margin: 0 0 16px; line-height: 1.4; }
    .gc-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-bottom: 14px; }
    .gc-opt {
        display: flex; align-items: center; gap: 8px;
        background: var(--c); padding: 11px 12px; border-radius: 9px;
        font-size: 0.82rem; font-weight: 700; color: white;
        border: 2px solid transparent; cursor: default;
        transition: all 0.3s;
    }
    .gc-opt.gc-sel { border-color: white; transform: scale(1.03); }
    .gc-opt.gc-correct { box-shadow: 0 0 0 3px white, 0 0 20px rgba(255,255,255,0.3); }
    .gc-opt.gc-wrong   { opacity: 0.5; }
    .gc-sh { font-size: 0.95rem; flex-shrink: 0; }
    .gc-lbl { flex: 1; }
    .gc-ck { margin-left: auto; font-size: 1rem; animation: popIn 0.3s cubic-bezier(0.34,1.56,0.64,1); }
    @keyframes popIn { from { transform: scale(0); } to { transform: scale(1); } }
    .gc-progress { display: flex; flex-direction: column; gap: 5px; }
    .gc-track { height: 6px; background: rgba(255,255,255,0.08); border-radius: 999px; overflow: hidden; }
    .gc-fill { height: 100%; background: linear-gradient(90deg,#22c55e,#16a34a); border-radius: 999px; transition: width 0.5s cubic-bezier(0.4,0,0.2,1); }
    .gc-progress span { font-size: 0.7rem; color: #64748b; text-align: center; }

    /* Floating pops */
    .pop {
        position: absolute; background: white; border-radius: 999px;
        padding: 7px 14px; font-size: 0.8rem; font-weight: 700; color: #0f172a;
        box-shadow: 0 8px 24px rgba(0,0,0,0.15);
        animation: popFloat var(--dur,5s) ease-in-out var(--del,0s) infinite;
        white-space: nowrap;
    }
    .p1 { top: 5%;  right: -30px; --dur:5s;   --del:0s;   color: #22c55e; }
    .p2 { top: 40%; right: -45px; --dur:6.5s; --del:0.8s; color: #f59e0b; }
    .p3 { bottom: 10%; left: -35px; --dur:4.5s; --del:1.6s; color: #6366f1; }
    @keyframes popFloat {
        0%,100% { transform: translateY(0) scale(1); }
        40%     { transform: translateY(-10px) scale(1.04); }
        70%     { transform: translateY(-6px) scale(1.02); }
    }

    /* Emoji reactions */
    .emoji-float {
        position: absolute; font-size: 1.4rem;
        animation: emojiRise var(--ed,4s) ease-out var(--edd,0s) infinite;
        opacity: 0;
    }
    .ef1 { left: -50px; bottom: 20%; --ed:4s;   --edd:0s; }
    .ef2 { left: -60px; bottom: 35%; --ed:5s;   --edd:1.2s; }
    .ef3 { right: -55px; bottom: 25%; --ed:4.5s; --edd:0.6s; }
    .ef4 { right: -50px; bottom: 40%; --ed:6s;   --edd:1.8s; }
    @keyframes emojiRise {
        0%   { opacity: 0;   transform: translateY(0) scale(0.5); }
        20%  { opacity: 1;   transform: translateY(-20px) scale(1); }
        80%  { opacity: 0.8; transform: translateY(-60px) scale(1); }
        100% { opacity: 0;   transform: translateY(-90px) scale(0.8); }
    }

    /* Scroll hint */
    .scroll-hint { position: absolute; bottom: 28px; left: 50%; transform: translateX(-50%); z-index: 10; }
    .scroll-line {
        width: 2px; height: 40px; margin: 0 auto;
        background: linear-gradient(to bottom, rgba(99,102,241,0.6), transparent);
        border-radius: 999px; animation: scrollPulse 2s ease-in-out infinite;
    }
    @keyframes scrollPulse {
        0%,100% { opacity: 0.3; transform: scaleY(0.6); transform-origin: top; }
        50%     { opacity: 1;   transform: scaleY(1);   transform-origin: top; }
    }

    /* ══ STATS ══ */
    .stats-section { background: white; border-top: 1px solid #f1f5f9; border-bottom: 1px solid #f1f5f9; padding: 40px 52px; }
    .stats-inner { max-width: 1000px; margin: 0 auto; display: grid; grid-template-columns: repeat(4,1fr); gap: 2px; }
    .stat-card {
        padding: 24px 20px; text-align: center;
        opacity: 0; transform: translateY(16px);
        transition: opacity 0.5s ease var(--d,0ms), transform 0.5s ease var(--d,0ms);
    }
    .stat-card.visible { opacity: 1; transform: none; }
    .stat-n { display: block; font-size: 2.2rem; font-weight: 900; color: #6366f1; font-variant-numeric: tabular-nums; }
    .stat-free { background: linear-gradient(135deg,#22c55e,#16a34a); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
    .stat-l { display: block; font-size: 0.82rem; color: #94a3b8; font-weight: 500; margin-top: 4px; margin-bottom: 10px; }
    .stat-bar { height: 3px; border-radius: 999px; background: var(--bc, linear-gradient(90deg,#6366f1,#a78bfa)); width: 0; transition: width 1.5s ease 0.3s; max-width: 80px; margin: 0 auto; }
    .stat-card.visible .stat-bar { width: var(--w,60%); }

    /* ══ FEATURES ══ */
    .features-section { padding: 110px 52px; background: #fafbff; }
    .sec-inner { max-width: 1120px; margin: 0 auto; }
    .sec-label {
        display: inline-block; font-size: 0.75rem; font-weight: 700;
        letter-spacing: 0.14em; text-transform: uppercase; color: #6366f1;
        background: rgba(99,102,241,0.08); padding: 4px 14px;
        border-radius: 999px; margin-bottom: 14px;
    }
    .sec-title { font-size: clamp(1.8rem, 4vw, 2.6rem); font-weight: 900; color: #0f172a; margin: 0 0 12px; letter-spacing: -0.02em; }
    .sec-sub   { color: #64748b; font-size: 1.05rem; margin: 0 0 60px; max-width: 480px; line-height: 1.6; }

    .feat-grid { display: grid; grid-template-columns: repeat(3,1fr); gap: 20px; }
    .feat-card {
        position: relative; overflow: hidden;
        background: white; border: 1px solid #f1f5f9; border-radius: 20px; padding: 30px;
        opacity: 0; transform: translateY(24px);
        transition: opacity 0.5s ease var(--fd,0ms), transform 0.5s ease var(--fd,0ms), box-shadow 0.3s, border-color 0.3s;
    }
    .feat-card.visible { opacity: 1; transform: none; }
    .feat-card::before {
        content: ''; position: absolute; inset: 0;
        background: linear-gradient(135deg, rgba(99,102,241,0.04), transparent);
        opacity: 0; transition: opacity 0.3s;
    }
    .feat-card:hover { box-shadow: 0 24px 56px rgba(99,102,241,0.13); border-color: rgba(99,102,241,0.25); transform: translateY(-8px) !important; }
    .feat-card:hover::before { opacity: 1; }
    .feat-icon {
        width: 88px; height: 88px; margin-bottom: 16px; position: relative;
        transition: transform 0.3s;
    }
    .feat-img { width: 100%; height: 100%; object-fit: contain; display: block; }
    .feat-card:hover .feat-icon { transform: scale(1.08) rotate(-4deg); }
    .feat-tag {
        position: absolute; top: -6px; right: -10px;
        background: #ef4444; color: white; font-size: 0.6rem; font-weight: 800;
        padding: 2px 7px; border-radius: 999px; letter-spacing: 0.05em; text-transform: uppercase;
        animation: tagPulse 2s ease-in-out infinite;
    }
    @keyframes tagPulse {
        0%,100% { box-shadow: 0 0 0 0 rgba(239,68,68,0.4); }
        50%     { box-shadow: 0 0 0 5px rgba(239,68,68,0); }
    }
    .feat-card h3 { font-size: 1.05rem; font-weight: 700; color: #0f172a; margin: 0 0 8px; }
    .feat-card p  { font-size: 0.87rem; color: #64748b; line-height: 1.65; margin: 0 0 16px; }
    .feat-line {
        height: 2px; border-radius: 999px; width: 0;
        background: var(--g); transition: width 0.5s ease;
    }
    .feat-card:hover .feat-line { width: 100%; }

    /* ══ HOW IT WORKS ══ */
    .how-section { padding: 110px 52px; background: white; }
    .steps { display: flex; align-items: center; gap: 0; margin-top: 16px; }
    .step-card {
        flex: 1; padding: 36px 28px; background: #fafbff;
        border: 1px solid #f1f5f9; border-radius: 20px; text-align: center;
        transition: transform 0.3s, box-shadow 0.3s;
    }
    .step-card:hover { transform: translateY(-6px); box-shadow: 0 20px 48px rgba(99,102,241,0.1); }
    .step-num-wrap { position: relative; display: inline-block; margin-bottom: 12px; }
    .step-num {
        font-size: 3.5rem; font-weight: 900; line-height: 1;
        background: linear-gradient(135deg, #6366f1, #a78bfa);
        -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text;
    }
    .step-ring {
        position: absolute; inset: -8px; border-radius: 50%;
        border: 2px solid rgba(99,102,241,0.25);
        animation: ringExpand 2.5s ease-out infinite;
    }
    @keyframes ringExpand {
        0%   { transform: scale(0.8); opacity: 0.8; }
        100% { transform: scale(1.8); opacity: 0; }
    }
    .step-icon { font-size: 2.2rem; margin-bottom: 14px; display: block; }
    .step-img { width: 130px; height: 130px; object-fit: contain; margin: 0 auto 14px; display: block; transition: transform .3s cubic-bezier(.34,1.56,.64,1); }
    .step-card:hover .step-img { transform: scale(1.08) translateY(-4px); }
    .step-card h3 { font-size: 1.1rem; font-weight: 700; color: #0f172a; margin: 0 0 10px; }
    .step-card p  { font-size: 0.86rem; color: #64748b; line-height: 1.65; margin: 0; }
    .step-connector { flex-shrink: 0; width: 60px; display: flex; flex-direction: column; align-items: center; gap: 6px; }
    .conn-line {
        height: 2px; width: 100%;
        background: linear-gradient(90deg, #6366f1, #a78bfa);
        animation: connPulse 2s ease-in-out infinite;
    }
    @keyframes connPulse {
        0%,100% { opacity: 0.4; } 50% { opacity: 1; }
    }
    .conn-arrow {
        font-size: 1.8rem; color: #a78bfa; line-height: 1;
        animation: arrowBounce 1.5s ease-in-out infinite;
    }
    @keyframes arrowBounce {
        0%,100% { transform: translateX(0); } 50% { transform: translateX(5px); }
    }

    /* ══ TICKER ══ */
    .ticker {
        background: linear-gradient(135deg, #6366f1, #8b5cf6);
        padding: 14px 0; overflow: hidden; white-space: nowrap;
    }
    .ticker-track {
        display: inline-block;
        animation: tickerRoll 30s linear infinite;
    }
    .ticker-track span {
        display: inline-block; margin: 0 24px;
        color: rgba(255,255,255,0.85); font-size: 0.88rem; font-weight: 700;
        letter-spacing: 0.05em;
    }
    @keyframes tickerRoll { from { transform: translateX(0); } to { transform: translateX(-33.33%); } }

    /* ══ CTA ══ */
    .cta-section {
        position: relative; overflow: hidden;
        background: linear-gradient(160deg, #1e1b4b 0%, #312e81 45%, #4c1d95 100%);
        padding: 120px 52px; text-align: center;
    }
    .cta-mesh {
        position: absolute; inset: 0; pointer-events: none;
        background:
            radial-gradient(ellipse 60% 60% at 20% 50%, rgba(99,102,241,0.35), transparent),
            radial-gradient(ellipse 50% 50% at 80% 50%, rgba(167,139,250,0.25), transparent);
        animation: meshPulse 6s ease-in-out infinite alternate;
    }
    .cta-particles { position: absolute; inset: 0; pointer-events: none; }
    .cta-spark {
        position: absolute; bottom: -10px;
        width: 2px; background: linear-gradient(to top, #a78bfa, transparent);
        border-radius: 999px;
        animation: sparkRise var(--sd,4s) ease-out var(--dd,0s) infinite;
    }
    @keyframes sparkRise {
        0%   { height: 0; opacity: 0; bottom: 0; }
        10%  { opacity: 0.8; }
        100% { height: 120px; opacity: 0; bottom: 100%; }
    }
    .cta-inner { position: relative; z-index: 1; max-width: 640px; margin: 0 auto; }
    .cta-mascot {
        width: 130px; height: 130px; object-fit: contain; display: block; margin: 0 auto 8px;
        filter: drop-shadow(0 12px 24px rgba(0,0,0,0.3));
        animation: heroFloat 5s ease-in-out infinite;
    }
    .cta-badge {
        display: inline-block; margin-bottom: 20px;
        background: rgba(255,255,255,0.1); color: rgba(255,255,255,0.85);
        border: 1px solid rgba(255,255,255,0.2); padding: 6px 18px;
        border-radius: 999px; font-size: 0.82rem; font-weight: 600;
        backdrop-filter: blur(4px);
    }
    .cta-title {
        font-size: clamp(2rem, 5vw, 3.2rem); font-weight: 900; color: white;
        margin: 0 0 16px; letter-spacing: -0.03em; line-height: 1.1;
    }
    .cta-sub { color: rgba(255,255,255,0.65); font-size: 1.05rem; margin: 0 0 40px; line-height: 1.65; }
    .cta-btns { display: flex; gap: 14px; justify-content: center; flex-wrap: wrap; margin-bottom: 20px; }
    .btn-cta-main {
        position: relative; overflow: hidden;
        display: inline-flex; align-items: center; gap: 8px;
        padding: 16px 34px; border-radius: 16px;
        background: white; color: #4f46e5; font-weight: 800; font-size: 1rem;
        text-decoration: none; transition: transform 0.2s, box-shadow 0.2s;
        box-shadow: 0 8px 32px rgba(0,0,0,0.3);
    }
    .btn-cta-main:hover { transform: translateY(-3px) scale(1.02); box-shadow: 0 16px 48px rgba(0,0,0,0.4); }
    .sparkle {
        position: absolute; font-size: 0.7rem; color: #fbbf24;
        animation: sparklePop 2s ease-in-out infinite;
        pointer-events: none;
    }
    .s1 { top: 4px; left: 10px; animation-delay: 0s; }
    .s2 { top: 4px; right: 14px; animation-delay: 0.7s; }
    .s3 { bottom: 4px; left: 50%; animation-delay: 1.4s; }
    @keyframes sparklePop {
        0%,100% { opacity: 0; transform: scale(0) rotate(0deg); }
        50%     { opacity: 1; transform: scale(1.4) rotate(20deg); }
    }
    .btn-cta-ghost {
        display: inline-flex; align-items: center;
        padding: 16px 28px; border-radius: 16px;
        border: 2px solid rgba(255,255,255,0.3);
        background: rgba(255,255,255,0.08);
        color: rgba(255,255,255,0.9); font-weight: 700; font-size: 1rem;
        text-decoration: none; backdrop-filter: blur(4px);
        transition: all 0.25s;
    }
    .btn-cta-ghost:hover { background: rgba(255,255,255,0.16); border-color: rgba(255,255,255,0.5); transform: translateY(-2px); }
    .cta-note { color: rgba(255,255,255,0.35); font-size: 0.78rem; margin: 0; }

    /* ══ FOOTER ══ */
    .footer { background: #0a0f1e; padding: 32px 52px; border-top: 1px solid rgba(255,255,255,0.04); }
    .footer-inner { max-width: 1120px; margin: 0 auto; display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 16px; }
    .footer-left {}
    .footer-brand { display: flex; align-items: center; gap: 8px; font-size: 1.15rem; font-weight: 900; color: white; margin-bottom: 4px; }
    .footer-left p { color: #475569; font-size: 0.8rem; margin: 0; }
    .footer-links { display: flex; gap: 20px; }
    .footer-links a { color: #475569; font-size: 0.82rem; text-decoration: none; transition: color 0.2s; }
    .footer-links a:hover { color: #a78bfa; }

    /* ══ RESPONSIVE ══ */
    @media (max-width: 1024px) {
        .feat-grid { grid-template-columns: repeat(2,1fr); }
    }
    @media (max-width: 900px) {
        .nav, .nav.scrolled { padding: 14px 24px; }
        .nav-link { display: none; }
        .hero-inner { grid-template-columns: 1fr; padding: 36px 24px 56px; gap: 48px; }
        .hero-right { display: none; }
        .stats-section { padding: 32px 24px; }
        .stats-inner { grid-template-columns: repeat(2,1fr); }
        .stat-div { display: none; }
        .features-section, .how-section, .cta-section { padding: 72px 24px; }
        .steps { flex-direction: column; gap: 8px; }
        .step-connector { flex-direction: row; width: 100%; height: 40px; }
        .conn-line { height: 100%; width: 2px; }
        .conn-arrow { font-size: 1.4rem; transform: rotate(90deg); }
        @keyframes arrowBounce {
            0%,100% { transform: rotate(90deg) translateX(0); }
            50%     { transform: rotate(90deg) translateX(5px); }
        }
        .footer { padding: 24px; }
        .footer-inner { flex-direction: column; align-items: center; text-align: center; }
        .cta-btns { flex-direction: column; align-items: center; }
    }
    @media (max-width: 600px) {
        .hero-title { font-size: 2.2rem; }
        .feat-grid { grid-template-columns: 1fr; }
        .stats-inner { grid-template-columns: 1fr 1fr; }
    }

    /* ══ CATEGORIES ══════════════════════════════════════════════════════════ */
    .cat-section { padding: 80px 24px; background: #f8faff; }
    .cat-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 20px;
        margin-top: 48px;
    }
    .cat-card {
        position: relative; overflow: hidden;
        background: white;
        border-radius: 20px;
        padding: 24px;
        text-decoration: none; color: inherit;
        border: 1.5px solid #e8eaf6;
        display: flex; flex-direction: column; gap: 14px;
        box-shadow: 0 2px 12px rgba(99,102,241,.06);
        transition: transform .25s cubic-bezier(.34,1.3,.64,1),
                    box-shadow .25s ease,
                    border-color .25s;
        opacity: 0; transform: translateY(24px);
        animation: catIn .5s cubic-bezier(.34,1.2,.64,1) calc(var(--i) * 80ms) forwards;
    }
    @keyframes catIn {
        to { opacity: 1; transform: translateY(0); }
    }
    .cat-card:hover {
        transform: translateY(-6px) scale(1.02);
        box-shadow: 0 16px 40px rgba(99,102,241,.15);
        border-color: transparent;
    }

    .cat-icon-wrap {
        position: relative; width: 104px; height: 80px;
    }
    .cat-img {
        position: relative; z-index: 1;
        width: 100%; height: 100%; object-fit: contain; display: block;
        padding: 6px; box-sizing: border-box;
        transition: transform .3s cubic-bezier(.34,1.56,.64,1);
    }
    .cat-icon-bg {
        position: absolute; inset: 0; border-radius: 16px;
        background: linear-gradient(135deg, var(--g));
        opacity: .12;
        transition: opacity .25s;
    }
    .cat-card:hover .cat-img { transform: scale(1.1) rotate(-3deg); }
    .cat-card:hover .cat-icon-bg { opacity: .24; }

    .cat-body h3 {
        font-size: 1.05rem; font-weight: 800; color: #0f172a;
        margin: 0 0 6px; letter-spacing: -.02em;
    }
    .cat-body p {
        font-size: .83rem; color: #64748b; line-height: 1.5; margin: 0;
    }

    .cat-foot {
        display: flex; align-items: center; justify-content: space-between;
        margin-top: auto;
    }
    .cat-cnt {
        font-size: .78rem; font-weight: 700;
        background: linear-gradient(135deg, var(--g));
        -webkit-background-clip: text; -webkit-text-fill-color: transparent;
        background-clip: text;
    }
    .cat-arr {
        font-size: 1.1rem; color: #94a3b8;
        transition: transform .2s, color .2s;
    }
    .cat-card:hover .cat-arr { transform: translateX(5px); color: #6366f1; }

    .cat-glow {
        position: absolute; inset: 0; border-radius: 20px;
        background: linear-gradient(135deg, var(--g));
        opacity: 0;
        transition: opacity .3s;
        pointer-events: none;
        z-index: 0;
    }
    .cat-card:hover .cat-glow { opacity: .04; }

    @media (max-width: 900px) { .cat-grid { grid-template-columns: 1fr 1fr; } }
    @media (max-width: 560px) { .cat-grid { grid-template-columns: 1fr; } .cat-section { padding: 60px 16px; } }
</style>
