import type { AuthResponse, Quiz, User } from './types';

export interface RoomHistory {
    id: string;
    pin: string;
    quiz_id: string;
    quiz_title: string;
    game_mode: string;
    status: string;
    created_at: string;
    ended_at: string | null;
    player_count: number;
}

export interface ClassItem {
    id: string;
    teacher_id: string;
    name: string;
    grade: string;
    subject: string;
    class_code: string;
    is_active: boolean;
    created_at: string;
    student_count: number;
}

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

class ApiError extends Error {
    constructor(public status: number, message: string) {
        super(message);
    }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
    const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
    const headers: HeadersInit = {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...init?.headers
    };

    const res = await fetch(`${BASE_URL}${path}`, { ...init, headers, credentials: 'include' });

    if (!res.ok) {
        const body = await res.json().catch(() => ({ error: res.statusText }));
        throw new ApiError(res.status, body.error ?? 'Unknown error');
    }

    if (res.status === 204) return undefined as T;
    return res.json();
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

export const auth = {
    register: (data: { email: string; username: string; password: string; full_name: string; role: string }) =>
        request<AuthResponse>('/api/auth/register', { method: 'POST', body: JSON.stringify(data) }),

    login: (email: string, password: string) =>
        request<AuthResponse>('/api/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),

    logout: () =>
        request('/api/auth/logout', { method: 'POST' }),

    me: () =>
        request<User>('/api/auth/me')
};

// ─── Quizzes ──────────────────────────────────────────────────────────────────

export const quizzes = {
    list: () =>
        request<Quiz[]>('/api/quizzes'),

    discover: (params?: { q?: string; subject?: string; grade?: string }) => {
        const qs = new URLSearchParams();
        if (params?.q) qs.set('q', params.q);
        if (params?.subject) qs.set('subject', params.subject);
        if (params?.grade) qs.set('grade', params.grade);
        const suffix = qs.toString() ? `?${qs}` : '';
        return request<Quiz[]>(`/api/quizzes/discover${suffix}`);
    },

    get: (id: string) =>
        request<Quiz>(`/api/quizzes/${id}`),

    create: (data: Partial<Quiz>) =>
        request<{ id: string }>('/api/quizzes', { method: 'POST', body: JSON.stringify(data) }),

    update: (id: string, data: Partial<Quiz>) =>
        request('/api/quizzes/' + id, { method: 'PUT', body: JSON.stringify(data) }),

    delete: (id: string) =>
        request('/api/quizzes/' + id, { method: 'DELETE' })
};

// ─── Rooms ────────────────────────────────────────────────────────────────────

export const rooms = {
    create: (quizId: string, gameMode: string, settings: object) =>
        request<{ room_id: string; pin: string }>('/api/rooms', {
            method: 'POST',
            body: JSON.stringify({ quiz_id: quizId, game_mode: gameMode, settings })
        }),

    info: (pin: string) =>
        request<{ id: string; status: string; quiz_title: string; host_name: string }>(`/api/rooms/${pin}/info`),

    history: () =>
        request<RoomHistory[]>('/api/rooms/history')
};

// ─── Classes ──────────────────────────────────────────────────────────────────

export const classes = {
    list: () =>
        request<ClassItem[]>('/api/classes'),

    create: (data: { name: string; grade: string; subject: string }) =>
        request<{ id: string; class_code: string }>('/api/classes', {
            method: 'POST',
            body: JSON.stringify(data)
        }),

    update: (id: string, data: { name: string; grade: string; subject: string }) =>
        request('/api/classes/' + id, { method: 'PUT', body: JSON.stringify(data) }),

    delete: (id: string) =>
        request('/api/classes/' + id, { method: 'DELETE' })
};

// ─── AI ───────────────────────────────────────────────────────────────────────

export interface AIGenerateRequest {
    topic: string;
    count: number;
    grade_level?: string;
    language?: string;
    question_types?: string[];
}

export interface AIGeneratedQuestion {
    type: string;
    text: string;
    time_limit: number;
    points: number;
    explanation: string;
    options: { text: string; is_correct: boolean }[];
}

export const ai = {
    generate: (data: AIGenerateRequest) =>
        request<AIGeneratedQuestion[]>('/api/ai/generate-questions', {
            method: 'POST',
            body: JSON.stringify(data)
        }),

    // Checks whether a provider key is valid / reachable. Pass `key` to test a
    // freshly typed (unsaved) key, or omit it to test the stored one.
    test: (provider: 'groq' | 'openai' | 'gemini', key?: string) =>
        request<{ ok: boolean; message: string }>('/api/ai/test', {
            method: 'POST',
            body: JSON.stringify({ provider, key: key ?? '' })
        })
};

// ─── Settings ─────────────────────────────────────────────────────────────────

export interface AppSetting {
    key: string;
    value: string;
    updated_at: string;
}

export const settings = {
    // Backend returns an object map: { "groq_api_key": "****abcd", ... }
    // API key values come back masked (only last 4 chars).
    getAll: () =>
        request<Record<string, string>>('/api/settings'),

    set: (key: string, value: string) =>
        request('/api/settings', { method: 'PUT', body: JSON.stringify({ key, value }) })
};

// ─── Exams ────────────────────────────────────────────────────────────────────

export interface ExamItem {
    id: string;
    quiz_id: string;
    teacher_id: string;
    title: string;
    code: string;
    time_limit: number;
    start_date: string | null;
    end_date: string | null;
    shuffle: boolean;
    max_attempts: number;
    status: 'draft' | 'active' | 'closed';
    created_at: string;
    quiz_title: string;
    submission_count: number;
}

export interface ExamSubmission {
    id: string;
    student_name: string;
    score: number;
    max_score: number;
    time_taken: number;
    submitted_at: string;
}

export const exams = {
    list: () => request<ExamItem[]>('/api/exams'),
    create: (data: { quiz_id: string; title: string; time_limit: number; start_date?: string|null; end_date?: string|null; shuffle: boolean; max_attempts: number }) =>
        request<{ id: string; code: string }>('/api/exams', { method: 'POST', body: JSON.stringify(data) }),
    setStatus: (id: string, status: string) =>
        request(`/api/exams/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) }),
    delete: (id: string) => request(`/api/exams/${id}`, { method: 'DELETE' }),
    results: (id: string) => request<ExamSubmission[]>(`/api/exams/${id}/results`),
};

// ─── Olympiads ────────────────────────────────────────────────────────────────

export interface OlympiadItem {
    id: string;
    quiz_id: string;
    title: string;
    description: string;
    code: string;
    start_time: string;
    end_time: string;
    max_participants: number | null;
    status: 'upcoming' | 'active' | 'completed';
    created_at: string;
    quiz_title: string;
    participant_count: number;
}

export interface OlympiadLeaderEntry {
    id: string;
    student_name: string;
    score: number;
    max_score: number;
    time_taken: number;
    rank: number;
    submitted_at: string;
}

export const olympiads = {
    list: () => request<OlympiadItem[]>('/api/olympiads'),
    create: (data: { quiz_id: string; title: string; description: string; start_time: string; end_time: string; max_participants?: number|null }) =>
        request<{ id: string; code: string }>('/api/olympiads', { method: 'POST', body: JSON.stringify(data) }),
    setStatus: (id: string, status: string) =>
        request(`/api/olympiads/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) }),
    delete: (id: string) => request(`/api/olympiads/${id}`, { method: 'DELETE' }),
    leaderboard: (id: string) => request<OlympiadLeaderEntry[]>(`/api/olympiads/${id}/leaderboard`),
};

// ─── WebSocket URL ────────────────────────────────────────────────────────────

export function getWebSocketURL(pin: string, role: 'host' | 'student', token?: string): string {
    const wsBase = BASE_URL.replace(/^http/, 'ws');
    const params = new URLSearchParams({ role });
    if (token) params.set('token', token);
    return `${wsBase}/ws/room/${pin}?${params}`;
}

export { ApiError };
