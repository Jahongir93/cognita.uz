export interface Test {
  id: string;
  icon: string;
  title: string;
  description: string;
  questions: number;
  difficulty: 'easy' | 'medium' | 'hard';
  duration: number;   // minutes
  plays: number;
  rating: number;     // 1-5
  subcat: string;     // matches SubCat.id
  isNew?: boolean;
  isHot?: boolean;
}

export interface SubCat {
  id: string;
  label: string;
  icon: string;
}

export interface Category {
  id: string;
  slug: string;
  icon: string;
  title: string;
  subtitle: string;
  g1: string;
  g2: string;
  subcats: SubCat[];
  tests: Test[];
}

// ─── 1. Qiziqarli testlar ────────────────────────────────────────────────────

export const funCategory: Category = {
  id: 'fun',
  slug: 'tests/fun',
  icon: '🎯',
  title: 'Qiziqarli testlar',
  subtitle: 'Shaxsiyat, fan-afsona, madaniyat va tabiat haqida qiziqarli testlar',
  g1: '#f59e0b',
  g2: '#ef4444',
  subcats: [
    { id: 'all',         label: 'Barchasi',   icon: '🌟' },
    { id: 'personality', label: 'Shaxsiyat',  icon: '😄' },
    { id: 'science',     label: 'Fan-afsona', icon: '🧪' },
    { id: 'culture',     label: 'Madaniyat',  icon: '🎬' },
    { id: 'nature',      label: 'Tabiat',     icon: '🌿' },
  ],
  tests: [
    {
      id: 'f1', icon: '😄',
      title: 'Shaxsiyat testi',
      description: "Qaysi shaxsiyat turiga kirasiz? 16 tip ichidan birini aniqlang",
      questions: 20, difficulty: 'easy', duration: 8,
      plays: 15420, rating: 4.8, subcat: 'personality', isHot: true,
    },
    {
      id: 'f2', icon: '🦁',
      title: 'Qaysi hayvonsiz?',
      description: 'Javoblaringiz orqali ichki totemingizni aniqlang',
      questions: 12, difficulty: 'easy', duration: 5,
      plays: 8320, rating: 4.6, subcat: 'personality', isNew: true,
    },
    {
      id: 'f3', icon: '🧪',
      title: 'Fan yoki afsona?',
      description: "Hayotdagi faktlarni rost yoki yolg'on deb aniqlang",
      questions: 20, difficulty: 'medium', duration: 10,
      plays: 6840, rating: 4.5, subcat: 'science',
    },
    {
      id: 'f4', icon: '🚀',
      title: 'Koinot siri',
      description: 'Kosmik kashfiyotlar va koinot faktlari',
      questions: 22, difficulty: 'hard', duration: 15,
      plays: 4210, rating: 4.7, subcat: 'science',
    },
    {
      id: 'f5', icon: '🎬',
      title: 'Kino bilimdonligi',
      description: 'Mashxur filmlar va aktyorlar haqida test',
      questions: 25, difficulty: 'easy', duration: 12,
      plays: 11200, rating: 4.4, subcat: 'culture', isHot: true,
    },
    {
      id: 'f6', icon: '🎵',
      title: 'Musiqa ziyraklik',
      description: "Qo'shiqlar va san'at asarlari haqida savollar",
      questions: 15, difficulty: 'easy', duration: 7,
      plays: 5630, rating: 4.3, subcat: 'culture',
    },
    {
      id: 'f7', icon: '🍕',
      title: 'Dunyo oshxonasi',
      description: 'Milliy taomlar va pishiriq sirlari',
      questions: 18, difficulty: 'easy', duration: 8,
      plays: 7890, rating: 4.5, subcat: 'culture',
    },
    {
      id: 'f8', icon: '🌍',
      title: 'Geografiya bilimingiz',
      description: "Dunyo davlatlari, poytaxtlar va tabiiy mo'jizalar",
      questions: 15, difficulty: 'medium', duration: 8,
      plays: 9100, rating: 4.6, subcat: 'nature',
    },
    {
      id: 'f9', icon: '🌊',
      title: 'Okean va dengizlar',
      description: 'Suv olami, baliqlar va dengiz hayoti',
      questions: 16, difficulty: 'medium', duration: 9,
      plays: 3200, rating: 4.4, subcat: 'nature',
    },
    {
      id: 'f10', icon: '🦋',
      title: 'Hayvonot dunyosi',
      description: 'Hayvonlar haqida ajoyib faktlar va testlar',
      questions: 20, difficulty: 'easy', duration: 10,
      plays: 5400, rating: 4.5, subcat: 'nature',
    },
    {
      id: 'f11', icon: '🏛️',
      title: 'Tarix siri',
      description: "Jahon tarixi bo'yicha qiziqarli savollar",
      questions: 18, difficulty: 'medium', duration: 10,
      plays: 4800, rating: 4.4, subcat: 'culture',
    },
    {
      id: 'f12', icon: '🔬',
      title: 'Ilmiy qiziqarlilar',
      description: 'Zamonaviy fan va texnologiya haqida faktlar',
      questions: 20, difficulty: 'medium', duration: 12,
      plays: 3900, rating: 4.6, subcat: 'science', isNew: true,
    },
  ],
};

// ─── 2. Fan testlari ─────────────────────────────────────────────────────────

export const subjectsCategory: Category = {
  id: 'subjects',
  slug: 'tests/subjects',
  icon: '📚',
  title: 'Fan testlari',
  subtitle: "Maktab fanlari bo'yicha chuqur bilimlarni tekshiring",
  g1: '#3b82f6',
  g2: '#6366f1',
  subcats: [
    { id: 'all',        label: 'Barchasi',       icon: '🌟' },
    { id: 'math',       label: 'Matematika',      icon: '🔢' },
    { id: 'science',    label: 'Tabiiy fanlar',   icon: '⚗️' },
    { id: 'humanities', label: 'Ijtimoiy fanlar', icon: '📜' },
    { id: 'languages',  label: 'Tillar',          icon: '🌐' },
    { id: 'it',         label: 'Informatika',     icon: '💻' },
  ],
  tests: [
    {
      id: 's1', icon: '🔢',
      title: 'Algebra asoslari',
      description: 'Tenglamalar, tengsizliklar va algebraik ifodalar',
      questions: 30, difficulty: 'medium', duration: 20,
      plays: 22400, rating: 4.7, subcat: 'math', isHot: true,
    },
    {
      id: 's2', icon: '📐',
      title: 'Geometriya',
      description: "Planimetriya va stereometriya bo'yicha testlar",
      questions: 25, difficulty: 'hard', duration: 18,
      plays: 14300, rating: 4.5, subcat: 'math',
    },
    {
      id: 's3', icon: '📊',
      title: 'Statistika va ehtimollik',
      description: "Ma'lumotlarni tahlil qilish va ehtimollik nazariyasi",
      questions: 20, difficulty: 'hard', duration: 15,
      plays: 8900, rating: 4.4, subcat: 'math', isNew: true,
    },
    {
      id: 's4', icon: '⚗️',
      title: 'Kimyo - elementlar',
      description: 'Davriy jadval va kimyoviy elementlar xossalari',
      questions: 35, difficulty: 'hard', duration: 22,
      plays: 12100, rating: 4.6, subcat: 'science',
    },
    {
      id: 's5', icon: '🔭',
      title: 'Fizika - mexanika',
      description: 'Harakat, kuch va energiya qonunlari',
      questions: 30, difficulty: 'hard', duration: 20,
      plays: 15600, rating: 4.5, subcat: 'science', isHot: true,
    },
    {
      id: 's6', icon: '🌿',
      title: 'Biologiya - hujayralar',
      description: 'Hujayra tuzilishi, fotosintez va biologik jarayonlar',
      questions: 28, difficulty: 'medium', duration: 16,
      plays: 11200, rating: 4.6, subcat: 'science',
    },
    {
      id: 's7', icon: '📜',
      title: "O'zbekiston tarixi",
      description: "Qadimgi davrdan mustaqillikkacha bo'lgan tarix",
      questions: 40, difficulty: 'medium', duration: 25,
      plays: 18900, rating: 4.8, subcat: 'humanities', isHot: true,
    },
    {
      id: 's8', icon: '🌍',
      title: 'Jahon geografiyasi',
      description: "Qit'alar, davlatlar va tabiiy geografiya",
      questions: 30, difficulty: 'medium', duration: 18,
      plays: 9800, rating: 4.5, subcat: 'humanities',
    },
    {
      id: 's9', icon: '🇬🇧',
      title: 'Ingliz tili - Grammar',
      description: 'Ingliz tili grammatika qoidalari va testlar',
      questions: 40, difficulty: 'medium', duration: 20,
      plays: 28400, rating: 4.7, subcat: 'languages', isHot: true,
    },
    {
      id: 's10', icon: '📝',
      title: "O'zbek tili imlo",
      description: "Imlo qoidalari, so'z yasalishi va sintaksis",
      questions: 35, difficulty: 'medium', duration: 20,
      plays: 16200, rating: 4.6, subcat: 'languages',
    },
    {
      id: 's11', icon: '💻',
      title: 'Algoritm asoslari',
      description: 'Dasturlash mantiqiy tafakkur va algoritmlar',
      questions: 25, difficulty: 'hard', duration: 18,
      plays: 7400, rating: 4.5, subcat: 'it',
    },
    {
      id: 's12', icon: '🖥️',
      title: 'Kompyuter bilim',
      description: 'Hardware, software va tarmoq asoslari',
      questions: 30, difficulty: 'medium', duration: 15,
      plays: 9200, rating: 4.4, subcat: 'it', isNew: true,
    },
  ],
};

// ─── 3. Attestatsiya savollari ───────────────────────────────────────────────

export const attestationCategory: Category = {
  id: 'attestation',
  slug: 'tests/attestation',
  icon: '📋',
  title: 'Attestatsiya savollari',
  subtitle: "O'qituvchilar attestatsiyasiga tayyorgarlik ko'ring",
  g1: '#22c55e',
  g2: '#0ea5e9',
  subcats: [
    { id: 'all',        label: 'Barchasi',       icon: '🌟' },
    { id: 'pedagogy',   label: 'Pedagogika',      icon: '👩‍🏫' },
    { id: 'math',       label: 'Matematika',      icon: '🔢' },
    { id: 'science',    label: 'Tabiiy fanlar',   icon: '⚗️' },
    { id: 'humanities', label: 'Ijtimoiy fanlar', icon: '📜' },
    { id: 'management', label: 'Boshqaruv',       icon: '💼' },
  ],
  tests: [
    {
      id: 'a1', icon: '👩‍🏫',
      title: 'Pedagogik mahorat',
      description: "Zamonaviy ta'lim metodlari va pedagogika asoslari",
      questions: 50, difficulty: 'medium', duration: 30,
      plays: 6800, rating: 4.7, subcat: 'pedagogy', isHot: true,
    },
    {
      id: 'a2', icon: '🧠',
      title: 'Pedagogik psixologiya',
      description: "O'quvchi psixologiyasi va motivatsiya usullari",
      questions: 45, difficulty: 'medium', duration: 28,
      plays: 4200, rating: 4.6, subcat: 'pedagogy',
    },
    {
      id: 'a3', icon: '🔢',
      title: "Matematika o'qituvchisi",
      description: "Maktab matematikasi bo'yicha attestatsiya savollari",
      questions: 60, difficulty: 'hard', duration: 40,
      plays: 5100, rating: 4.5, subcat: 'math', isHot: true,
    },
    {
      id: 'a4', icon: '⚗️',
      title: "Kimyo o'qituvchisi",
      description: "Kimyo fani bo'yicha to'liq attestatsiya kursi",
      questions: 55, difficulty: 'hard', duration: 35,
      plays: 3800, rating: 4.4, subcat: 'science',
    },
    {
      id: 'a5', icon: '🔭',
      title: "Fizika o'qituvchisi",
      description: "Fizika fani bo'yicha attestatsiya savol banki",
      questions: 58, difficulty: 'hard', duration: 38,
      plays: 4100, rating: 4.5, subcat: 'science',
    },
    {
      id: 'a6', icon: '📜',
      title: "Tarix o'qituvchisi",
      description: "O'zbekiston va jahon tarixi metodikasi",
      questions: 52, difficulty: 'medium', duration: 32,
      plays: 4900, rating: 4.6, subcat: 'humanities', isNew: true,
    },
    {
      id: 'a7', icon: '📖',
      title: "O'zbek tili va adabiyoti",
      description: 'Grammatika, imlo va adabiyot nazariyasi',
      questions: 48, difficulty: 'medium', duration: 30,
      plays: 5600, rating: 4.7, subcat: 'humanities',
    },
    {
      id: 'a8', icon: '🇬🇧',
      title: 'Chet til metodikasi',
      description: "Ingliz tili o'qitish metodikasi va grammatika",
      questions: 45, difficulty: 'medium', duration: 28,
      plays: 4300, rating: 4.5, subcat: 'humanities',
    },
    {
      id: 'a9', icon: '💼',
      title: 'Maktab rahbariyati',
      description: "Ta'lim menejmenti va huquqiy normalar",
      questions: 40, difficulty: 'medium', duration: 25,
      plays: 2800, rating: 4.3, subcat: 'management', isNew: true,
    },
    {
      id: 'a10', icon: '💻',
      title: "Informatika o'qituvchisi",
      description: 'Dasturlash asoslari va axborot texnologiyalari',
      questions: 38, difficulty: 'hard', duration: 24,
      plays: 3200, rating: 4.4, subcat: 'math',
    },
  ],
};

// ─── 4. IQ testlar ───────────────────────────────────────────────────────────

export const iqCategory: Category = {
  id: 'iq',
  slug: 'tests/iq',
  icon: '🧠',
  title: 'IQ testlar',
  subtitle: "Mantiqiy tafakkur, xotira va aqliy qobiliyatlarni o'lchang",
  g1: '#8b5cf6',
  g2: '#6366f1',
  subcats: [
    { id: 'all',     label: 'Barchasi', icon: '🌟' },
    { id: 'numeric', label: 'Raqamlar', icon: '🔢' },
    { id: 'spatial', label: 'Fazoviy',  icon: '🔷' },
    { id: 'verbal',  label: 'Verbal',   icon: '🔤' },
    { id: 'memory',  label: 'Xotira',   icon: '💾' },
    { id: 'logic',   label: 'Mantiq',   icon: '⚙️' },
  ],
  tests: [
    {
      id: 'i1', icon: '🔢',
      title: 'Raqamli ketma-ketlik',
      description: 'Sonlar orasidagi qonuniyatni toping',
      questions: 20, difficulty: 'medium', duration: 15,
      plays: 9800, rating: 4.7, subcat: 'numeric', isHot: true,
    },
    {
      id: 'i2', icon: '➗',
      title: 'Arifmetik ziyraklik',
      description: 'Tezkor arifmetik hisob-kitob va raqamli mantiq',
      questions: 25, difficulty: 'hard', duration: 18,
      plays: 7200, rating: 4.5, subcat: 'numeric',
    },
    {
      id: 'i3', icon: '🔷',
      title: 'Geometrik figuralar',
      description: "Shakllar orasidagi bog'liqlikni aniqlang",
      questions: 20, difficulty: 'hard', duration: 20,
      plays: 8100, rating: 4.6, subcat: 'spatial',
    },
    {
      id: 'i4', icon: '🧩',
      title: 'Matritsa analizi',
      description: 'Raven progressiv matritsalari uslubidagi test',
      questions: 30, difficulty: 'hard', duration: 25,
      plays: 11400, rating: 4.8, subcat: 'spatial', isHot: true,
    },
    {
      id: 'i5', icon: '🌀',
      title: 'Fazoviy tasavvur',
      description: '3D shakllarni mental burib-aylantirish',
      questions: 18, difficulty: 'hard', duration: 15,
      plays: 5600, rating: 4.4, subcat: 'spatial', isNew: true,
    },
    {
      id: 'i6', icon: '🔤',
      title: "So'z analogiyasi",
      description: "Tushunchalar orasidagi munosabatni toping",
      questions: 25, difficulty: 'medium', duration: 15,
      plays: 6800, rating: 4.5, subcat: 'verbal',
    },
    {
      id: 'i7', icon: '📚',
      title: "So'z ma'nosi",
      description: "So'zlar ma'nosi va ularning munosabatini aniqlang",
      questions: 20, difficulty: 'medium', duration: 12,
      plays: 5100, rating: 4.4, subcat: 'verbal',
    },
    {
      id: 'i8', icon: '💾',
      title: 'Qisqa xotira testi',
      description: 'Qisqa muddatli xotirangizni tekshiring',
      questions: 15, difficulty: 'easy', duration: 10,
      plays: 12300, rating: 4.6, subcat: 'memory',
    },
    {
      id: 'i9', icon: '🗂️',
      title: 'Uzoq xotira',
      description: 'Diqqat va uzoq muddatli xotira sinovlari',
      questions: 18, difficulty: 'medium', duration: 12,
      plays: 7800, rating: 4.5, subcat: 'memory', isNew: true,
    },
    {
      id: 'i10', icon: '⚙️',
      title: 'Murakkab mantiq',
      description: "Ko'p bosqichli muammolarni hal qiling",
      questions: 20, difficulty: 'hard', duration: 20,
      plays: 6400, rating: 4.7, subcat: 'logic', isHot: true,
    },
  ],
};

// ─── 5. Psixologik testlar ───────────────────────────────────────────────────

export const psychologyCategory: Category = {
  id: 'psychology',
  slug: 'tests/psychology',
  icon: '🧘',
  title: 'Psixologik testlar',
  subtitle: "O'zingizni yaxshiroq tushunish uchun ilmiy testlar",
  g1: '#ec4899',
  g2: '#8b5cf6',
  subcats: [
    { id: 'all',         label: 'Barchasi',   icon: '🌟' },
    { id: 'personality', label: 'Shaxsiyat',  icon: '🎭' },
    { id: 'emotion',     label: "His-tuyg'u", icon: '💝' },
    { id: 'social',      label: 'Ijtimoiy',   icon: '🤝' },
    { id: 'career',      label: 'Kasb',       icon: '🎯' },
  ],
  tests: [
    {
      id: 'p1', icon: '🎭',
      title: 'MBTI shaxsiyat testi',
      description: '16 shaxsiyat tipidan qaysiligingizni bilib oling',
      questions: 60, difficulty: 'medium', duration: 25,
      plays: 28400, rating: 4.9, subcat: 'personality', isHot: true,
    },
    {
      id: 'p2', icon: '😊',
      title: 'Baxtlilik darajasi',
      description: 'Hayotingizdan qanchalik qoniqayotganingizni aniqlang',
      questions: 15, difficulty: 'easy', duration: 8,
      plays: 14200, rating: 4.7, subcat: 'emotion',
    },
    {
      id: 'p3', icon: '😰',
      title: 'Tashvish darajasi',
      description: 'Vaziyatlarga emosional munosabatingizni baholang',
      questions: 16, difficulty: 'easy', duration: 8,
      plays: 9800, rating: 4.6, subcat: 'emotion',
    },
    {
      id: 'p4', icon: '💪',
      title: 'Stressga chidamlilik',
      description: "Qiyin vaziyatlarda qanday munosabatda bo'lishingizni tekshiring",
      questions: 20, difficulty: 'medium', duration: 12,
      plays: 11600, rating: 4.7, subcat: 'emotion', isHot: true,
    },
    {
      id: 'p5', icon: '🤝',
      title: 'Muloqot uslubi',
      description: "Odamlar bilan qanday munosabat o'rnatishingizni bilib oling",
      questions: 18, difficulty: 'easy', duration: 10,
      plays: 8400, rating: 4.5, subcat: 'social',
    },
    {
      id: 'p6', icon: '👥',
      title: 'Liderlik qobiliyati',
      description: 'Jamoadagi rolingiz va liderlik salohiyatingizni aniqlang',
      questions: 22, difficulty: 'medium', duration: 14,
      plays: 7200, rating: 4.6, subcat: 'social', isNew: true,
    },
    {
      id: 'p7', icon: '🧠',
      title: 'Hissiy intellekt (EQ)',
      description: "His-tuyg'ularni boshqarish va tushunish qobiliyati",
      questions: 30, difficulty: 'medium', duration: 18,
      plays: 12800, rating: 4.8, subcat: 'emotion', isHot: true,
    },
    {
      id: 'p8', icon: '🎯',
      title: 'Maqsadga intilish',
      description: "Motivatsiya va maqsad qo'yish qobiliyatingizni baholang",
      questions: 22, difficulty: 'medium', duration: 12,
      plays: 6600, rating: 4.5, subcat: 'career',
    },
    {
      id: 'p9', icon: '💼',
      title: 'Kasb tanlash',
      description: "Qobiliyat va qiziqishlaringizga mos kasb yo'nalishini toping",
      questions: 35, difficulty: 'medium', duration: 20,
      plays: 9400, rating: 4.7, subcat: 'career', isNew: true,
    },
    {
      id: 'p10', icon: '🌟',
      title: "O'zini-o'zi baholash",
      description: "O'z-o'ziga ishonch darajangizni aniqlang",
      questions: 18, difficulty: 'easy', duration: 10,
      plays: 8100, rating: 4.5, subcat: 'personality',
    },
  ],
};

// ─── 6. O'yinlar ─────────────────────────────────────────────────────────────

export const gamesCategory: Category = {
  id: 'games',
  slug: 'games',
  icon: '🎮',
  title: "O'yinlar",
  subtitle: "So'z o'yinlari, mantiq, tezlik va xotira o'yinlari",
  g1: '#f59e0b',
  g2: '#22c55e',
  subcats: [
    { id: 'all',    label: 'Barchasi',       icon: '🌟' },
    { id: 'word',   label: "So'z o'yinlari", icon: '🔤' },
    { id: 'logic',  label: 'Mantiq',         icon: '🧩' },
    { id: 'speed',  label: 'Tezlik',         icon: '⚡' },
    { id: 'memory', label: 'Xotira',         icon: '🃏' },
  ],
  tests: [
    {
      id: 'g1', icon: '🔤',
      title: "So'z topish (Wordle)",
      description: "Yashirin so'zni 6 urinishda toping — Wordle uslubi",
      questions: 1, difficulty: 'medium', duration: 5,
      plays: 38400, rating: 4.9, subcat: 'word', isHot: true,
    },
    {
      id: 'g2', icon: '📝',
      title: 'Krossvord',
      description: "O'zbekcha krossvord — bilim va izlanish",
      questions: 15, difficulty: 'medium', duration: 20,
      plays: 14200, rating: 4.6, subcat: 'word',
    },
    {
      id: 'g3', icon: '🔡',
      title: 'Anagram',
      description: "Aralashtirilgan harflardan so'z hosil qiling",
      questions: 20, difficulty: 'easy', duration: 8,
      plays: 9800, rating: 4.5, subcat: 'word', isNew: true,
    },
    {
      id: 'g4', icon: '⚡',
      title: 'Tezkor matematika',
      description: 'Vaqt chegarasida hisob-kitob musobaqasi',
      questions: 30, difficulty: 'medium', duration: 5,
      plays: 16800, rating: 4.7, subcat: 'speed', isHot: true,
    },
    {
      id: 'g5', icon: '🗺️',
      title: 'Xarita topish',
      description: 'Davlatlar va shaharlarni xaritada toping',
      questions: 20, difficulty: 'medium', duration: 10,
      plays: 11400, rating: 4.6, subcat: 'speed',
    },
    {
      id: 'g6', icon: '🎯',
      title: "Nishon o'yini",
      description: "Ko'rinadigan ob'ektni tanlang — reflex va diqqat",
      questions: 30, difficulty: 'easy', duration: 3,
      plays: 8200, rating: 4.4, subcat: 'speed',
    },
    {
      id: 'g7', icon: '🔢',
      title: 'Sudoku',
      description: "Raqamlarni to'g'ri joylashtiring — mantiq va sabr",
      questions: 9, difficulty: 'hard', duration: 25,
      plays: 12600, rating: 4.8, subcat: 'logic', isHot: true,
    },
    {
      id: 'g8', icon: '🧩',
      title: '15-puzzle',
      description: "Raqamli qovoq o'yini — minimal harakatda tartibga sol",
      questions: 1, difficulty: 'hard', duration: 10,
      plays: 6400, rating: 4.5, subcat: 'logic',
    },
    {
      id: 'g9', icon: '🃏',
      title: 'Memory karta',
      description: 'Juftlikdagi kartochkalarni toping — xotira mashqi',
      questions: 16, difficulty: 'easy', duration: 5,
      plays: 14800, rating: 4.6, subcat: 'memory',
    },
    {
      id: 'g10', icon: '🏆',
      title: 'Bilim musobaqasi',
      description: "Turli fanlar bo'yicha 60 soniyalik yakkama-yakka",
      questions: 60, difficulty: 'hard', duration: 60,
      plays: 5600, rating: 4.7, subcat: 'memory', isNew: true,
    },
  ],
};

// ─── Exported collection ─────────────────────────────────────────────────────

export const allCategories: Category[] = [
  funCategory,
  subjectsCategory,
  attestationCategory,
  iqCategory,
  psychologyCategory,
  gamesCategory,
];
