-- ============================================================
-- GoGame.uz Database Schema
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- full-text search

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE user_role AS ENUM ('admin', 'teacher', 'student');
CREATE TYPE question_type AS ENUM (
    'multiple_choice', 'true_false', 'short_answer',
    'poll', 'image_choice', 'audio', 'draw', 'reorder', 'fill_blank'
);
CREATE TYPE activity_template AS ENUM (
    'quiz', 'match_up', 'wordsearch', 'crossword',
    'maze_chase', 'group_sort', 'type_answer',
    'labelled_diagram', 'flashcards', 'spin_wheel'
);
CREATE TYPE game_mode AS ENUM ('classic', 'self_paced', 'team', 'accuracy', 'confidence', 'zero_stakes');
CREATE TYPE room_status AS ENUM ('waiting', 'in_progress', 'paused', 'completed', 'abandoned');
CREATE TYPE assignment_status AS ENUM ('draft', 'active', 'closed');

-- ============================================================
-- USERS & AUTH
-- ============================================================

CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email       VARCHAR(255) UNIQUE NOT NULL,
    username    VARCHAR(100) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    full_name   VARCHAR(255) NOT NULL,
    avatar_url  VARCHAR(500),
    role        user_role NOT NULL DEFAULT 'student',
    school_id   UUID,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE schools (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    region      VARCHAR(100),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE users ADD CONSTRAINT fk_users_school
    FOREIGN KEY (school_id) REFERENCES schools(id) ON DELETE SET NULL;

-- ============================================================
-- CLASSES & GROUPS
-- ============================================================

CREATE TABLE classes (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    teacher_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    grade       VARCHAR(20),
    subject     VARCHAR(100),
    class_code  VARCHAR(10) UNIQUE NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE class_students (
    class_id    UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    student_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (class_id, student_id)
);

-- ============================================================
-- QUIZ & QUESTIONS
-- ============================================================

CREATE TABLE quizzes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    teacher_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           VARCHAR(500) NOT NULL,
    description     TEXT,
    subject         VARCHAR(100),
    grade_level     VARCHAR(20),
    cover_image     VARCHAR(500),
    template        activity_template NOT NULL DEFAULT 'quiz',
    is_public       BOOLEAN NOT NULL DEFAULT false,
    total_questions INT NOT NULL DEFAULT 0,
    play_count      INT NOT NULL DEFAULT 0,
    tags            TEXT[],
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE questions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id         UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    order_index     INT NOT NULL DEFAULT 0,
    type            question_type NOT NULL DEFAULT 'multiple_choice',
    question_text   TEXT NOT NULL,
    media_url       VARCHAR(500),
    media_type      VARCHAR(20),          -- image, video, audio
    time_limit      INT NOT NULL DEFAULT 20,  -- seconds
    points          INT NOT NULL DEFAULT 100,
    explanation     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE answer_options (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    question_id     UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    option_text     TEXT NOT NULL,
    media_url       VARCHAR(500),
    is_correct      BOOLEAN NOT NULL DEFAULT false,
    order_index     INT NOT NULL DEFAULT 0
);

-- ============================================================
-- ROOMS (LIVE GAME SESSIONS)
-- ============================================================

CREATE TABLE rooms (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id         UUID NOT NULL REFERENCES quizzes(id),
    host_id         UUID NOT NULL REFERENCES users(id),
    pin             VARCHAR(10) UNIQUE NOT NULL,
    status          room_status NOT NULL DEFAULT 'waiting',
    game_mode       game_mode NOT NULL DEFAULT 'classic',
    settings        JSONB NOT NULL DEFAULT '{}',
    current_question_index INT NOT NULL DEFAULT -1,
    started_at      TIMESTAMPTZ,
    ended_at        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- settings JSONB fields:
-- { "shuffle_questions": bool, "shuffle_answers": bool,
--   "show_leaderboard": bool, "music": bool,
--   "team_count": int, "lobby_music": bool }

CREATE TABLE room_participants (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id     UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    nickname    VARCHAR(100) NOT NULL,
    avatar      VARCHAR(10),             -- emoji avatar
    team_id     INT,
    score       INT NOT NULL DEFAULT 0,
    streak      INT NOT NULL DEFAULT 0,
    rank        INT,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- GAME ANSWERS (LIVE SESSION)
-- ============================================================

CREATE TABLE game_answers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id         UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    participant_id  UUID NOT NULL REFERENCES room_participants(id) ON DELETE CASCADE,
    question_id     UUID NOT NULL REFERENCES questions(id),
    selected_option_id UUID REFERENCES answer_options(id),
    text_answer     TEXT,
    is_correct      BOOLEAN,
    points_earned   INT NOT NULL DEFAULT 0,
    response_time_ms INT,                -- milliseconds
    answered_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- ASSIGNMENTS (SELF-PACED)
-- ============================================================

CREATE TABLE assignments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id         UUID NOT NULL REFERENCES quizzes(id),
    teacher_id      UUID NOT NULL REFERENCES users(id),
    class_id        UUID REFERENCES classes(id),
    title           VARCHAR(500) NOT NULL,
    status          assignment_status NOT NULL DEFAULT 'draft',
    due_date        TIMESTAMPTZ,
    settings        JSONB NOT NULL DEFAULT '{}',
    -- { "allow_retries": bool, "max_retries": int, "show_answers": bool,
    --   "timer_enabled": bool, "read_aloud": bool, "focus_mode": bool }
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE assignment_submissions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    assignment_id   UUID NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    student_id      UUID NOT NULL REFERENCES users(id),
    attempt         INT NOT NULL DEFAULT 1,
    score           INT NOT NULL DEFAULT 0,
    max_score       INT NOT NULL DEFAULT 0,
    accuracy        DECIMAL(5,2),
    time_spent_sec  INT,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE submission_answers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    submission_id   UUID NOT NULL REFERENCES assignment_submissions(id) ON DELETE CASCADE,
    question_id     UUID NOT NULL REFERENCES questions(id),
    selected_option_id UUID REFERENCES answer_options(id),
    text_answer     TEXT,
    is_correct      BOOLEAN,
    points_earned   INT NOT NULL DEFAULT 0,
    response_time_ms INT
);

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX idx_quizzes_teacher ON quizzes(teacher_id);
CREATE INDEX idx_quizzes_public ON quizzes(is_public) WHERE is_public = true;
CREATE INDEX idx_questions_quiz ON questions(quiz_id, order_index);
CREATE INDEX idx_rooms_pin ON rooms(pin);
CREATE INDEX idx_rooms_host ON rooms(host_id);
CREATE INDEX idx_game_answers_room ON game_answers(room_id, question_id);
CREATE INDEX idx_participants_room ON room_participants(room_id);
CREATE INDEX idx_assignments_class ON assignments(class_id);
CREATE INDEX idx_submissions_assignment ON assignment_submissions(assignment_id, student_id);

-- Full-text search on quizzes
CREATE INDEX idx_quizzes_search ON quizzes USING gin(to_tsvector('simple', title || ' ' || COALESCE(description, '')));
