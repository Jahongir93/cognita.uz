package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"gogame.uz/backend/internal/config"
)

var DB *pgxpool.Pool

func Connect() {
	pool, err := pgxpool.New(context.Background(), config.App.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	DB = pool
	log.Println("Database connected successfully")
}

func Migrate() {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS app_settings (
			key        VARCHAR(100) PRIMARY KEY,
			value      TEXT NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS exams (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			quiz_id      UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
			teacher_id   UUID NOT NULL REFERENCES users(id),
			title        VARCHAR(200) NOT NULL,
			code         VARCHAR(8) UNIQUE NOT NULL,
			time_limit   INT NOT NULL DEFAULT 30,
			start_date   TIMESTAMPTZ,
			end_date     TIMESTAMPTZ,
			shuffle      BOOLEAN NOT NULL DEFAULT false,
			max_attempts INT NOT NULL DEFAULT 1,
			status       VARCHAR(20) NOT NULL DEFAULT 'draft',
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS exam_submissions (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			exam_id      UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
			student_name VARCHAR(100) NOT NULL,
			answers      JSONB NOT NULL DEFAULT '[]',
			score        INT NOT NULL DEFAULT 0,
			max_score    INT NOT NULL DEFAULT 0,
			time_taken   INT NOT NULL DEFAULT 0,
			submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS olympiads (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			quiz_id           UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
			teacher_id        UUID NOT NULL REFERENCES users(id),
			title             VARCHAR(200) NOT NULL,
			description       TEXT NOT NULL DEFAULT '',
			code              VARCHAR(8) UNIQUE NOT NULL,
			start_time        TIMESTAMPTZ NOT NULL,
			end_time          TIMESTAMPTZ NOT NULL,
			max_participants  INT,
			status            VARCHAR(20) NOT NULL DEFAULT 'upcoming',
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS olympiad_submissions (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			olympiad_id  UUID NOT NULL REFERENCES olympiads(id) ON DELETE CASCADE,
			student_name VARCHAR(100) NOT NULL,
			answers      JSONB NOT NULL DEFAULT '[]',
			score        INT NOT NULL DEFAULT 0,
			max_score    INT NOT NULL DEFAULT 0,
			time_taken   INT NOT NULL DEFAULT 0,
			submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		// Doska topshiriqlari (Topshiriqlar bo'limi)
		`CREATE TABLE IF NOT EXISTS board_activities (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type       VARCHAR(40) NOT NULL,
			title      VARCHAR(200) NOT NULL,
			content    JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_board_activities_teacher ON board_activities(teacher_id)`,
	}

	for _, stmt := range migrations {
		if _, err := DB.Exec(context.Background(), stmt); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}
	log.Println("Database migrations applied successfully")
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
