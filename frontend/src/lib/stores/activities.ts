// Doska topshiriqlari — localStorage'da saqlanadigan CRUD store.
//
// Eslatma: hozircha localStorage'da saqlanadi (bitta brauzer). O'qituvchi
// yaratgan va o'ynatgan qurilma bir xil bo'lishi kerak. Qurilmalar aro
// sinxronlash uchun keyinchalik backend qo'shilishi mumkin.

import { writable } from 'svelte/store';

export interface QuizContent {
    questions: { text: string; options: string[]; correct: number }[];
}
export interface TrueFalseContent {
    statements: { text: string; answer: boolean }[];
}
export interface PairsContent {
    pairs: { left: string; right: string }[];
}
export interface GroupsContent {
    groups: { name: string; items: string[] }[];
}
export interface WordsContent {
    words: { word: string; hint?: string }[];
}
export interface PromptsContent {
    prompts: string[];
}

export type ActivityContent =
    | QuizContent | TrueFalseContent | PairsContent
    | GroupsContent | WordsContent | PromptsContent
    | Record<string, unknown>;

export interface Activity {
    id: string;
    type: string;   // modul id (activityModules.ts)
    title: string;
    content: ActivityContent;
    createdAt: number;
    updatedAt: number;
}

const KEY = 'cognita_activities';

function read(): Activity[] {
    if (typeof localStorage === 'undefined') return [];
    try {
        const raw = localStorage.getItem(KEY);
        return raw ? (JSON.parse(raw) as Activity[]) : [];
    } catch {
        return [];
    }
}

function write(list: Activity[]) {
    if (typeof localStorage === 'undefined') return;
    localStorage.setItem(KEY, JSON.stringify(list));
}

export const activities = writable<Activity[]>(read());

function refresh() {
    activities.set(read());
}

function genId(): string {
    return 'act_' + Math.random().toString(36).slice(2, 10) + Date.now().toString(36).slice(-4);
}

export const activityStore = {
    subscribe: activities.subscribe,

    list(): Activity[] {
        return read();
    },

    get(id: string): Activity | undefined {
        return read().find(a => a.id === id);
    },

    create(type: string, title: string, content: ActivityContent): Activity {
        const now = Date.now();
        const item: Activity = { id: genId(), type, title, content, createdAt: now, updatedAt: now };
        const list = read();
        list.unshift(item);
        write(list);
        refresh();
        return item;
    },

    update(id: string, patch: Partial<Pick<Activity, 'title' | 'content'>>): void {
        const list = read();
        const i = list.findIndex(a => a.id === id);
        if (i === -1) return;
        list[i] = { ...list[i], ...patch, updatedAt: Date.now() };
        write(list);
        refresh();
    },

    remove(id: string): void {
        write(read().filter(a => a.id !== id));
        refresh();
    },
};
