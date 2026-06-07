package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"gogame.uz/backend/internal/middleware"
	"gogame.uz/backend/internal/models"
)

type AuthHandler struct {
	db *pgxpool.Pool
}

func NewAuthHandler(db *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{db: db}
}

// POST /api/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	user := &models.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Username: req.Username,
		Password: string(hash),
		FullName: req.FullName,
		Role:     req.Role,
	}

	_, err = h.db.Exec(context.Background(),
		`INSERT INTO users (id, email, username, password, full_name, role)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Email, user.Username, user.Password, user.FullName, user.Role,
	)
	if err != nil {
		return c.Status(409).JSON(fiber.Map{"error": "Email or username already exists"})
	}

	token, refresh, err := middleware.GenerateTokens(user.ID, user.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Token generation failed"})
	}

	return c.Status(201).JSON(models.AuthResponse{
		Token:        token,
		RefreshToken: refresh,
		User:         user.ToPublic(),
	})
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var user models.User
	err := h.db.QueryRow(context.Background(),
		`SELECT id, email, username, password, full_name, avatar_url, role, is_active
		 FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Username, &user.Password,
		&user.FullName, &user.AvatarURL, &user.Role, &user.IsActive)

	if err != nil || !user.IsActive {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, refresh, err := middleware.GenerateTokens(user.ID, user.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Token generation failed"})
	}

	// Set HTTP-only cookie for browser clients
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   false, // true in production with HTTPS
		SameSite: "Lax",
	})

	return c.JSON(models.AuthResponse{
		Token:        token,
		RefreshToken: refresh,
		User:         user.ToPublic(),
	})
}

// GET /api/auth/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var user models.User
	err := h.db.QueryRow(context.Background(),
		`SELECT id, email, username, full_name, avatar_url, role, created_at
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Email, &user.Username, &user.FullName,
		&user.AvatarURL, &user.Role, &user.CreatedAt)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user.ToPublic())
}

// POST /api/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:    "token",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
	})
	return c.JSON(fiber.Map{"message": "Logged out"})
}
