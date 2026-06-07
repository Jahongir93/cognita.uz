package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gogame.uz/backend/internal/config"
	"gogame.uz/backend/internal/models"
)

type Claims struct {
	UserID uuid.UUID        `json:"user_id"`
	Role   models.UserRole  `json:"role"`
	jwt.RegisteredClaims
}

func GenerateTokens(userID uuid.UUID, role models.UserRole) (token, refresh string, err error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.App.JWTExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString([]byte(config.App.JWTSecret))
	if err != nil {
		return
	}

	refreshClaims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.App.RefreshExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refresh, err = rt.SignedString([]byte(config.App.JWTSecret))
	return
}

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := extractClaims(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

func RequireRole(roles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(models.UserRole)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}
}

func extractClaims(c *fiber.Ctx) (*Claims, error) {
	authHeader := c.Get("Authorization")
	var tokenStr string

	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		// Also check cookie for browser clients
		tokenStr = c.Cookies("token")
	}

	if tokenStr == "" {
		return nil, fiber.ErrUnauthorized
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(config.App.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fiber.ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fiber.ErrUnauthorized
	}
	return claims, nil
}

// GetUserID extracts user ID from context (must be after Protected middleware)
func GetUserID(c *fiber.Ctx) uuid.UUID {
	return c.Locals("userID").(uuid.UUID)
}

func GetRole(c *fiber.Ctx) models.UserRole {
	return c.Locals("role").(models.UserRole)
}
