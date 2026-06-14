// Doska o'yinlari (Topshiriqlar) — modul reyestri.
//
// Bu o'yinlar ELEKTRON DOSKADA o'ynaladi: o'qituvchi yaratadi va ishga tushiradi,
// o'quvchilar doskada ishlaydi. Talaba qurilmalarida ko'rinmaydi (WebSocket yo'q).
//
// `kind` — yaratish (create) formasi qaysi kontent shaklini so'rashini belgilaydi.
// `player` — doska o'ynatgichi qaysi komponentdan foydalanishini belgilaydi.
// `implemented: false` modullar ro'yxatda "Tez kunda" sifatida ko'rinadi.

export type ContentKind =
    | 'quiz'       // savol + variantlar (to'g'ri belgilanadi)
    | 'truefalse'  // fikr + to'g'ri/noto'g'ri
    | 'pairs'      // chap–o'ng juftliklar (moslashtirish, xotira, kartochka)
    | 'groups'     // nomlangan guruhlar, har birida elementlar
    | 'words'      // so'zlar ro'yxati (anagramma, so'z qidirish)
    | 'sequence'   // to'g'ri tartibga keltiriladigan elementlar
    | 'prompts'    // ochiq topshiriq/savol matnlari (g'ildirak, quti, nutq)
    | 'fillblank'; // bo'sh joyli gaplar

export type PlayerKind =
    | 'quiz' | 'truefalse' | 'memory' | 'flashcards'
    | 'sort' | 'anagram' | 'wheel' | 'box';

export interface ActivityModule {
    num: number;
    id: string;            // slug — URL va saqlashda ishlatiladi
    name: string;
    desc: string;
    icon: string;
    kind: ContentKind;
    player?: PlayerKind;   // faqat implemented bo'lsa
    implemented: boolean;
    category: 'Savol-javob' | 'Moslashtirish' | 'So\'z' | 'Saralash' | 'Tasodif' | 'Harakatli';
}

export const ACTIVITY_MODULES: ActivityModule[] = [
    { num: 1,  id: 'anagram',        name: 'Harflarni tartiblash',     desc: 'Aralash harflardan so\'zni tiklash',              icon: '🔤', kind: 'words',     player: 'anagram',    implemented: true,  category: 'So\'z' },
    { num: 2,  id: 'wheel',          name: 'G\'ildirakni aylantirish', desc: 'Tasodifiy savol/topshiriqni tanlash',           icon: '🎡', kind: 'prompts',   player: 'wheel',      implemented: true,  category: 'Tasodif' },
    { num: 3,  id: 'open-box',       name: 'Qutini ochish',            desc: 'Qutilarni ochib topshiriqni bajarish',          icon: '🎁', kind: 'prompts',   player: 'box',        implemented: true,  category: 'Tasodif' },
    { num: 4,  id: 'unjumble',       name: 'Gapni tartibga keltirish', desc: 'Gap bo\'laklarini to\'g\'ri tartibga solish',   icon: '🧱', kind: 'sequence',  implemented: false, category: 'So\'z' },
    { num: 5,  id: 'matching-pairs', name: 'Mos juftliklarni topish',  desc: 'Mos kartochkalarni xotiradan topish',           icon: '🃏', kind: 'pairs',     player: 'memory',     implemented: true,  category: 'Moslashtirish' },
    { num: 6,  id: 'quiz',           name: 'Viktorina',                desc: 'Variantlardan to\'g\'ri javobni tanlash',       icon: '❓', kind: 'quiz',      player: 'quiz',       implemented: true,  category: 'Savol-javob' },
    { num: 7,  id: 'group-sort',     name: 'Guruhlarga ajratish',      desc: 'Elementlarni to\'g\'ri guruhga ajratish',       icon: '🗂️', kind: 'groups',    player: 'sort',       implemented: true,  category: 'Saralash' },
    { num: 8,  id: 'match-up',       name: 'Moslashtirish',            desc: 'Atama va ta\'rifni bog\'lash',                  icon: '🔗', kind: 'pairs',     implemented: false, category: 'Moslashtirish' },
    { num: 9,  id: 'flashcards',     name: 'Kartochkalar',             desc: 'Old/orqa tomonli bilim kartochkalari',          icon: '📇', kind: 'pairs',     player: 'flashcards', implemented: true,  category: 'Moslashtirish' },
    { num: 10, id: 'speaking-cards', name: 'Nutq kartochkalari',       desc: 'Tasodifiy kartochka asosida gapirish',          icon: '🗣️', kind: 'prompts',   implemented: false, category: 'Tasodif' },
    { num: 11, id: 'complete-sentence', name: 'Gapni to\'ldirish',     desc: 'Bo\'sh joylarga mos so\'z qo\'yish',            icon: '✏️', kind: 'fillblank', implemented: false, category: 'So\'z' },
    { num: 12, id: 'find-match',     name: 'Mos javobni topish',       desc: 'To\'g\'ri javobni tanlab, noto\'g\'rini o\'chirish', icon: '🎯', kind: 'quiz',  implemented: false, category: 'Savol-javob' },
    { num: 13, id: 'gameshow-quiz',  name: 'Shou-viktorina',           desc: 'Vaqt, jon va bonusli viktorina',                icon: '📺', kind: 'quiz',      implemented: false, category: 'Savol-javob' },
    { num: 14, id: 'type-answer',    name: 'Javobni yozish',           desc: 'Javobni klaviaturada yozish',                   icon: '⌨️', kind: 'quiz',      implemented: false, category: 'Savol-javob' },
    { num: 15, id: 'wordsearch',     name: 'So\'z qidirish',           desc: 'Harflar jadvalidan so\'zni topish',             icon: '🔎', kind: 'words',     implemented: false, category: 'So\'z' },
    { num: 16, id: 'flip-tiles',     name: 'Aylanuvchi plitkalar',     desc: 'Ikki tomonli plitkalarni o\'rganish',           icon: '🔲', kind: 'pairs',     implemented: false, category: 'Moslashtirish' },
    { num: 17, id: 'hangman',        name: 'So\'zni topish',           desc: 'Harf tanlab yashirin so\'zni topish',           icon: '🪢', kind: 'words',     implemented: false, category: 'So\'z' },
    { num: 18, id: 'labelled-diagram', name: 'Belgilangan diagramma',  desc: 'Rasmdagi joylarga nom qo\'yish',                icon: '🖼️', kind: 'pairs',     implemented: false, category: 'Moslashtirish' },
    { num: 19, id: 'spell-word',     name: 'So\'zni yozish',           desc: 'Harflardan so\'z tuzish',                       icon: '📝', kind: 'words',     implemented: false, category: 'So\'z' },
    { num: 20, id: 'crossword',      name: 'Krossvord',                desc: 'Izohlar asosida krossvord to\'ldirish',         icon: '🧩', kind: 'words',     implemented: false, category: 'So\'z' },
    { num: 21, id: 'maze-chase',     name: 'Labirint quvish',          desc: 'To\'g\'ri javob tomon harakatlanish',           icon: '👾', kind: 'quiz',      implemented: false, category: 'Harakatli' },
    { num: 22, id: 'true-false',     name: 'To\'g\'ri yoki noto\'g\'ri', desc: 'Fikr to\'g\'ri yoki noto\'g\'riligini aniqlash', icon: '✅', kind: 'truefalse', player: 'truefalse', implemented: true,  category: 'Savol-javob' },
    { num: 23, id: 'flying-answers', name: 'Uchuvchi javoblar',        desc: 'Harakatlanayotgan to\'g\'ri javobni tanlash',   icon: '🛸', kind: 'quiz',      implemented: false, category: 'Harakatli' },
    { num: 24, id: 'balloon-pop',    name: 'Sharlarni yorish',         desc: 'Sharlarni yorib javobni moslashtirish',         icon: '🎈', kind: 'groups',    implemented: false, category: 'Harakatli' },
    { num: 25, id: 'matching-game',  name: 'Juftmi yoki juft emasmi',  desc: 'Ikki karta mosligini aniqlash',                 icon: '⚖️', kind: 'pairs',     implemented: false, category: 'Moslashtirish' },
    { num: 26, id: 'whack-a-mole',   name: 'To\'g\'ri javobni urish',  desc: 'Faqat to\'g\'ri javoblarni urish',              icon: '🔨', kind: 'quiz',      implemented: false, category: 'Harakatli' },
    { num: 27, id: 'image-quiz',     name: 'Rasmli viktorina',         desc: 'Asta ochiluvchi rasmni aniqlash',               icon: '🏞️', kind: 'quiz',      implemented: false, category: 'Savol-javob' },
    { num: 28, id: 'rank-order',     name: 'Ketma-ketlikka joylashtirish', desc: 'Elementlarni to\'g\'ri tartibga solish',    icon: '📊', kind: 'sequence',  implemented: false, category: 'Saralash' },
    { num: 29, id: 'airplane',       name: 'Samolyot o\'yini',         desc: 'Samolyotni to\'g\'ri javob tomon boshqarish',   icon: '✈️', kind: 'quiz',      implemented: false, category: 'Harakatli' },
    { num: 30, id: 'watch-remember', name: 'Kuzat va eslab qol',       desc: 'Ko\'rganlarni eslab tanlash',                   icon: '👁️', kind: 'prompts',   implemented: false, category: 'Tasodif' },
    { num: 31, id: 'speed-sort',     name: 'Tezkor saralash',          desc: 'Vaqt ichida guruhlarga ajratish',               icon: '⚡', kind: 'groups',    implemented: false, category: 'Saralash' },
    { num: 32, id: 'win-lose-quiz',  name: 'Yutish yoki yutqazish',    desc: 'Ball tikib javob berish',                       icon: '🎰', kind: 'quiz',      implemented: false, category: 'Savol-javob' },
    { num: 33, id: 'maths-generator',name: 'Matematika generatori',    desc: 'Avtomatik matematika savollari',                icon: '🔢', kind: 'quiz',      implemented: false, category: 'Savol-javob' },
    { num: 34, id: 'word-magnets',   name: 'So\'z magnitlari',         desc: 'So\'zlarni sudrab gap tuzish',                  icon: '🧲', kind: 'sequence',  implemented: false, category: 'So\'z' },
];

export function getModule(id: string): ActivityModule | undefined {
    return ACTIVITY_MODULES.find(m => m.id === id);
}
