import { writable, derived } from 'svelte/store';
import type {
    Player, LeaderboardEntry, QuestionPayload,
    AnswerResultPayload, RoomStatePayload, GameOverPayload
} from '$lib/api/types';

export type GamePhase =
    | 'idle'
    | 'lobby'
    | 'question'
    | 'answered'
    | 'question_end'
    | 'leaderboard'
    | 'paused'
    | 'game_over';

interface GameState {
    phase: GamePhase;
    roomInfo: RoomStatePayload | null;
    players: Player[];
    currentQuestion: QuestionPayload | null;
    myAnswer: string | null;
    myResult: AnswerResultPayload | null;
    leaderboard: LeaderboardEntry[];
    finalResult: GameOverPayload | null;
    secondsLeft: number;
    totalQuestions: number;
    currentQuestionIndex: number;
    connectionStatus: 'connecting' | 'connected' | 'disconnected' | 'error';
    myParticipantId: string;
    myNickname: string;
    myAvatar: string;
    answeredCount: number;
    totalCount: number;
}

const initial: GameState = {
    phase: 'idle',
    roomInfo: null,
    players: [],
    currentQuestion: null,
    myAnswer: null,
    myResult: null,
    leaderboard: [],
    finalResult: null,
    secondsLeft: 0,
    totalQuestions: 0,
    currentQuestionIndex: 0,
    connectionStatus: 'disconnected',
    myParticipantId: '',
    myNickname: '',
    myAvatar: '',
    answeredCount: 0,
    totalCount: 0,
};

function createGameStore() {
    const { subscribe, set, update } = writable<GameState>(initial);

    return {
        subscribe,

        setConnectionStatus(status: GameState['connectionStatus']) {
            update(s => ({ ...s, connectionStatus: status }));
        },

        applyRoomState(payload: RoomStatePayload) {
            update(s => ({
                ...s,
                roomInfo: payload,
                players: payload.players,
                totalQuestions: payload.total_questions,
                totalCount: payload.players.filter(p => p.is_active).length,
                phase: payload.status === 'waiting' ? 'lobby' : s.phase
            }));
        },

        setYourInfo(id: string, nickname: string, avatar: string) {
            update(s => ({ ...s, myParticipantId: id, myNickname: nickname, myAvatar: avatar }));
        },

        setAnswerCount(answered: number, total: number) {
            update(s => ({ ...s, answeredCount: answered, totalCount: total }));
        },

        playerJoined(player: Player) {
            update(s => ({
                ...s,
                players: [...s.players.filter(p => p.id !== player.id), player],
                totalCount: s.players.filter(p => p.id !== player.id && p.is_active).length + 1
            }));
        },

        playerLeft(id: string) {
            update(s => ({
                ...s,
                players: s.players.map(p => p.id === id ? { ...p, is_active: false } : p)
            }));
        },

        showQuestion(payload: QuestionPayload) {
            update(s => ({
                ...s,
                phase: 'question',
                currentQuestion: payload,
                currentQuestionIndex: payload.question_index,
                myAnswer: null,
                myResult: null,
                secondsLeft: payload.time_limit,
                answeredCount: 0,
            }));
        },

        setTimer(seconds: number) {
            update(s => ({ ...s, secondsLeft: seconds }));
        },

        submitAnswer(optionId: string) {
            update(s => ({ ...s, myAnswer: optionId, phase: 'answered' }));
        },

        applyAnswerResult(payload: AnswerResultPayload) {
            update(s => ({ ...s, myResult: payload }));
        },

        showQuestionEnd() {
            update(s => ({ ...s, phase: 'question_end' }));
        },

        showLeaderboard(entries: LeaderboardEntry[]) {
            update(s => ({ ...s, phase: 'leaderboard', leaderboard: entries }));
        },

        pause() {
            update(s => ({ ...s, phase: 'paused' }));
        },

        resume() {
            update(s => ({ ...s, phase: 'question' }));
        },

        gameOver(payload: GameOverPayload) {
            update(s => ({
                ...s,
                phase: 'game_over',
                finalResult: payload,
                leaderboard: payload.leaderboard
            }));
        },

        reset() {
            set(initial);
        }
    };
}

export const gameStore = createGameStore();

export const hasAnswered = derived(gameStore, $g => $g.myAnswer !== null);

export const progress = derived(gameStore, $g =>
    $g.totalQuestions > 0
        ? Math.round(($g.currentQuestionIndex / $g.totalQuestions) * 100)
        : 0
);

export const timerPercent = derived(gameStore, $g =>
    $g.currentQuestion
        ? ($g.secondsLeft / $g.currentQuestion.time_limit) * 100
        : 100
);
