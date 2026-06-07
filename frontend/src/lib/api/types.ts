// ─── Enums ───────────────────────────────────────────────────────────────────

export type UserRole = 'admin' | 'teacher' | 'student';
export type QuestionType = 'multiple_choice' | 'true_false' | 'short_answer' | 'poll' | 'image_choice' | 'audio' | 'draw' | 'reorder' | 'fill_blank';
export type ActivityTemplate = 'quiz' | 'match_up' | 'wordsearch' | 'crossword' | 'maze_chase' | 'group_sort' | 'type_answer' | 'labelled_diagram' | 'flashcards';
export type GameMode = 'classic' | 'team' | 'accuracy' | 'confidence' | 'zero_stakes';
export type RoomStatus = 'waiting' | 'in_progress' | 'paused' | 'completed' | 'abandoned';

// ─── Auth ─────────────────────────────────────────────────────────────────────

export interface User {
    id: string;
    email: string;
    username: string;
    full_name: string;
    avatar_url: string | null;
    role: UserRole;
}

export interface AuthResponse {
    token: string;
    refresh_token: string;
    user: User;
}

// ─── Quiz ─────────────────────────────────────────────────────────────────────

export interface AnswerOption {
    id: string;
    option_text: string;
    media_url?: string;
    is_correct: boolean;
    order_index: number;
}

export interface StudentAnswerOption {
    id: string;
    option_text: string;
    media_url?: string;
}

export interface Question {
    id: string;
    quiz_id: string;
    order_index: number;
    type: QuestionType;
    question_text: string;
    media_url?: string;
    media_type?: string;
    time_limit: number;
    points: number;
    explanation?: string;
    options: AnswerOption[];
}

export interface Quiz {
    id: string;
    teacher_id: string;
    title: string;
    description?: string;
    subject?: string;
    grade_level?: string;
    cover_image?: string;
    template: ActivityTemplate;
    is_public: boolean;
    total_questions: number;
    play_count: number;
    tags: string[];
    created_at: string;
    updated_at: string;
    questions?: Question[];
}

// ─── Room / Game ──────────────────────────────────────────────────────────────

export interface RoomSettings {
    shuffle_questions: boolean;
    shuffle_answers: boolean;
    show_leaderboard: boolean;
    music: boolean;
    lobby_music: boolean;
    team_count: number;
    show_correct_answer: boolean;
}

export interface Player {
    id: string;
    nickname: string;
    avatar: string;
    score: number;
    streak: number;
    team_id?: number;
    is_active: boolean;
}

export interface LeaderboardEntry {
    rank: number;
    id: string;
    nickname: string;
    avatar: string;
    score: number;
    streak: number;
    delta: number; // rank change
}

// ─── WebSocket Message Types ──────────────────────────────────────────────────

export interface WSMessage<T = unknown> {
    type: string;
    payload: T;
}

export interface QuestionPayload {
    question_index: number;
    total_questions: number;
    question_id: string;
    type: QuestionType;
    question_text: string;
    media_url?: string;
    media_type?: string;
    time_limit: number;
    points: number;
    options: StudentAnswerOption[];
}

export interface QuestionEndPayload {
    question_id: string;
    correct_options: string[];
    explanation?: string;
    stats: {
        total_answers: number;
        correct_count: number;
        option_counts: Record<string, number>;
        average_time_ms: number;
    };
}

export interface AnswerResultPayload {
    is_correct: boolean;
    points_earned: number;
    total_score: number;
    streak: number;
    streak_bonus: number;
    rank: number;
}

export interface RoomStatePayload {
    room_id: string;
    pin: string;
    status: RoomStatus;
    game_mode: GameMode;
    quiz_title: string;
    host_name: string;
    players: Player[];
    total_questions: number;
}

export interface GameOverPayload {
    leaderboard: LeaderboardEntry[];
    stats: {
        total_players: number;
        total_questions: number;
        average_accuracy: number;
        duration_sec: number;
    };
}
