import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import type { User } from '$lib/api/types';
import { auth as authApi } from '$lib/api/client';

interface AuthState {
    user: User | null;
    token: string | null;
    loading: boolean;
}

function createAuthStore() {
    const { subscribe, set, update } = writable<AuthState>({
        user: null,
        token: browser ? localStorage.getItem('token') : null,
        loading: false
    });

    return {
        subscribe,

        async login(email: string, password: string) {
            update(s => ({ ...s, loading: true }));
            try {
                const res = await authApi.login(email, password);
                localStorage.setItem('token', res.token);
                set({ user: res.user, token: res.token, loading: false });
                return res;
            } catch (e) {
                update(s => ({ ...s, loading: false }));
                throw e;
            }
        },

        async register(data: Parameters<typeof authApi.register>[0]) {
            update(s => ({ ...s, loading: true }));
            try {
                const res = await authApi.register(data);
                localStorage.setItem('token', res.token);
                set({ user: res.user, token: res.token, loading: false });
                return res;
            } catch (e) {
                update(s => ({ ...s, loading: false }));
                throw e;
            }
        },

        async logout() {
            await authApi.logout().catch(() => {});
            localStorage.removeItem('token');
            set({ user: null, token: null, loading: false });
        },

        async loadUser() {
            if (!browser) return;
            const token = localStorage.getItem('token');
            if (!token) return;
            try {
                const user = await authApi.me();
                update(s => ({ ...s, user }));
            } catch {
                localStorage.removeItem('token');
                set({ user: null, token: null, loading: false });
            }
        },

        setUser(user: User) {
            update(s => ({ ...s, user }));
        }
    };
}

export const authStore = createAuthStore();
export const isLoggedIn = derived(authStore, $s => $s.user !== null);
export const isTeacher = derived(authStore, $s => $s.user?.role === 'teacher' || $s.user?.role === 'admin');
