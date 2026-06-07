package handlers

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// knownSettingKeys are the recognised app_settings keys.
var knownSettingKeys = []string{
	"openai_api_key",
	"ai_provider",
	"ai_model",
}

// SettingsHandler manages application settings with an in-memory cache backed
// by the app_settings table.
type SettingsHandler struct {
	db    *pgxpool.Pool
	mu    sync.RWMutex
	cache map[string]string
}

func NewSettingsHandler(db *pgxpool.Pool) *SettingsHandler {
	h := &SettingsHandler{
		db:    db,
		cache: make(map[string]string),
	}
	h.loadCache()
	return h
}

// loadCache pre-warms the in-memory cache from the database.
func (h *SettingsHandler) loadCache() {
	rows, err := h.db.Query(context.Background(),
		`SELECT key, value FROM app_settings`)
	if err != nil {
		return
	}
	defer rows.Close()

	h.mu.Lock()
	defer h.mu.Unlock()
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err == nil {
			h.cache[k] = v
		}
	}
}

// Get reads a setting from cache (falls back to DB on miss).
func (h *SettingsHandler) Get(key string) (string, bool) {
	h.mu.RLock()
	v, ok := h.cache[key]
	h.mu.RUnlock()
	if ok {
		return v, true
	}

	// DB fallback
	var val string
	err := h.db.QueryRow(context.Background(),
		`SELECT value FROM app_settings WHERE key = $1`, key).Scan(&val)
	if err != nil {
		return "", false
	}

	h.mu.Lock()
	h.cache[key] = val
	h.mu.Unlock()
	return val, true
}

// maskAPIKey shows only the last 4 characters of an API key.
func maskAPIKey(v string) string {
	if len(v) <= 4 {
		return "****"
	}
	return "****" + v[len(v)-4:]
}

// isAPIKeyField returns true for setting keys that hold secret credentials.
func isAPIKeyField(key string) bool {
	return key == "openai_api_key" || key == "gemini_api_key" || key == "groq_api_key"
}

// GET /api/settings
func (h *SettingsHandler) GetAll(c *fiber.Ctx) error {
	h.mu.RLock()
	snapshot := make(map[string]string, len(h.cache))
	for k, v := range h.cache {
		snapshot[k] = v
	}
	h.mu.RUnlock()

	result := make(fiber.Map, len(snapshot))
	for k, v := range snapshot {
		if isAPIKeyField(k) {
			result[k] = maskAPIKey(v)
		} else {
			result[k] = v
		}
	}

	return c.JSON(result)
}

// SetRequest is the body for PUT /api/settings.
type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PUT /api/settings
func (h *SettingsHandler) Set(c *fiber.Ctx) error {
	var req SetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "key is required"})
	}

	_, err := h.db.Exec(context.Background(),
		`INSERT INTO app_settings (key, value, updated_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (key) DO UPDATE
		   SET value = EXCLUDED.value,
		       updated_at = EXCLUDED.updated_at`,
		req.Key, req.Value, time.Now(),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save setting"})
	}

	// Update cache
	h.mu.Lock()
	h.cache[req.Key] = req.Value
	h.mu.Unlock()

	return c.JSON(fiber.Map{"message": "Setting saved"})
}
