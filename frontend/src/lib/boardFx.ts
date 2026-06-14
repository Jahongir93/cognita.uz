// Doska o'yinlari uchun ovoz effektlari va konfetti.
// Web Audio API — fayl talab qilmaydi, foydalanuvchi bosgandan keyin ishlaydi.

let ctx: AudioContext | null = null;
function ac(): AudioContext | null {
    if (typeof window === 'undefined') return null;
    if (!ctx) {
        const AC = window.AudioContext || (window as any).webkitAudioContext;
        if (!AC) return null;
        ctx = new AC();
    }
    return ctx;
}

function tone(freq: number, dur: number, type: OscillatorType = 'sine', delay = 0, gain = 0.18) {
    const c = ac();
    if (!c) return;
    const t0 = c.currentTime + delay;
    const osc = c.createOscillator();
    const g = c.createGain();
    osc.type = type;
    osc.frequency.value = freq;
    g.gain.setValueAtTime(0.0001, t0);
    g.gain.exponentialRampToValueAtTime(gain, t0 + 0.02);
    g.gain.exponentialRampToValueAtTime(0.0001, t0 + dur);
    osc.connect(g).connect(c.destination);
    osc.start(t0);
    osc.stop(t0 + dur + 0.02);
}

export type Sfx = 'correct' | 'wrong' | 'win' | 'flip' | 'pick' | 'reveal' | 'open';

export function sfx(kind: Sfx) {
    try {
        switch (kind) {
            case 'correct': tone(660, 0.12); tone(990, 0.16, 'sine', 0.09); break;
            case 'wrong':   tone(200, 0.28, 'sawtooth', 0, 0.14); break;
            case 'win':     [523, 659, 784, 1046].forEach((f, i) => tone(f, 0.2, 'triangle', i * 0.12, 0.2)); break;
            case 'flip':    tone(440, 0.06, 'square', 0, 0.07); break;
            case 'pick':    tone(560, 0.05, 'sine', 0, 0.08); break;
            case 'reveal':  tone(392, 0.1, 'triangle'); tone(587, 0.16, 'triangle', 0.1); break;
            case 'open':    tone(330, 0.08, 'square'); tone(660, 0.14, 'sine', 0.08); break;
        }
    } catch { /* ignore */ }
}

export function confetti(count = 70) {
    if (typeof document === 'undefined') return;
    const colors = ['#e21b3c', '#1368ce', '#d89e00', '#26890c', '#9333ea', '#fbbf24', '#ec4899'];
    for (let i = 0; i < count; i++) {
        const el = document.createElement('div');
        const c = colors[i % colors.length];
        const size = 7 + Math.random() * 8;
        el.style.cssText =
            `position:fixed;top:-16px;left:${Math.random() * 100}vw;` +
            `width:${size}px;height:${size * 1.5}px;background:${c};` +
            `z-index:99999;border-radius:2px;pointer-events:none;opacity:0.95;`;
        document.body.appendChild(el);
        const dx = (Math.random() * 2 - 1) * 200;
        const dur = 2200 + Math.random() * 1400;
        const anim = el.animate(
            [
                { transform: 'translate(0,0) rotate(0deg)', opacity: 1 },
                { transform: `translate(${dx}px, ${window.innerHeight + 60}px) rotate(${600 + Math.random() * 540}deg)`, opacity: 0.9 },
            ],
            { duration: dur, easing: 'cubic-bezier(0.2, 0.6, 0.35, 1)' },
        );
        anim.onfinish = () => el.remove();
    }
}
