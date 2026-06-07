package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AIHandler handles AI-powered question generation.
type AIHandler struct {
	db         *pgxpool.Pool
	httpClient *http.Client
}

func NewAIHandler(db *pgxpool.Pool) *AIHandler {
	return &AIHandler{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// loadSetting reads a value from app_settings, returning the defaultVal when
// the key is not found.
func (h *AIHandler) loadSetting(key, defaultVal string) string {
	var val string
	err := h.db.QueryRow(context.Background(),
		`SELECT value FROM app_settings WHERE key = $1`, key).Scan(&val)
	if err != nil || val == "" {
		return defaultVal
	}
	return val
}

// GenerateQuestionsRequest is the body for POST /api/ai/generate-questions.
type GenerateQuestionsRequest struct {
	Topic         string   `json:"topic"`
	Count         int      `json:"count"`
	GradeLevel    string   `json:"grade_level"`
	Language      string   `json:"language"`
	QuestionTypes []string `json:"question_types"`
}

// GeneratedOption mirrors the JSON structure returned by the AI.
type GeneratedOption struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
}

// GeneratedQuestion mirrors the JSON structure returned by the AI.
type GeneratedQuestion struct {
	QuestionText string            `json:"question_text"`
	Type         string            `json:"type"`
	Options      []GeneratedOption `json:"options"`
	Explanation  string            `json:"explanation"`
	TimeLimit    int               `json:"time_limit"`
	Points       int               `json:"points"`
}

// openAIRequest is the payload sent to the OpenAI Chat Completions endpoint.
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIResponse is the subset of the Chat Completions response we care about.
type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// TestConnectionRequest is the body for POST /api/ai/test.
type TestConnectionRequest struct {
	Provider string `json:"provider"` // "groq" | "openai" | "gemini"
	Key      string `json:"key"`      // optional: test this key; if empty, test the saved one
}

// providerInfo maps a provider id to its settings key and "list models" URL,
// which we use as a lightweight, side-effect-free connectivity check.
func providerInfo(provider string) (settingKey, modelsURL string, ok bool) {
	switch provider {
	case "groq":
		return "groq_api_key", "https://api.groq.com/openai/v1/models", true
	case "openai":
		return "openai_api_key", "https://api.openai.com/v1/models", true
	case "gemini":
		return "gemini_api_key", "https://generativelanguage.googleapis.com/v1beta/models", true
	default:
		return "", "", false
	}
}

// POST /api/ai/test — checks whether an AI provider key is valid / reachable.
func (h *AIHandler) TestConnection(c *fiber.Ctx) error {
	var req TestConnectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	settingKey, modelsURL, ok := providerInfo(req.Provider)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Noma'lum provider"})
	}

	// Use the supplied key, unless it's empty or a masked placeholder — then
	// fall back to the stored key.
	key := strings.TrimSpace(req.Key)
	if key == "" || strings.HasPrefix(key, "****") {
		key = h.loadSetting(settingKey, "")
	}
	if key == "" {
		return c.JSON(fiber.Map{"ok": false, "message": "Kalit kiritilmagan"})
	}

	// Build the connectivity-check request.
	var httpReq *http.Request
	var err error
	if req.Provider == "gemini" {
		httpReq, err = http.NewRequest(http.MethodGet, modelsURL+"?key="+key, nil)
	} else {
		httpReq, err = http.NewRequest(http.MethodGet, modelsURL, nil)
		if err == nil {
			httpReq.Header.Set("Authorization", "Bearer "+key)
		}
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to build request"})
	}

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return c.JSON(fiber.Map{"ok": false, "message": "Ulanib bo'lmadi: " + err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return c.JSON(fiber.Map{"ok": true, "message": "Ulanish muvaffaqiyatli ✓"})
	}

	// Surface a short reason for common failures.
	msg := fmt.Sprintf("Kalit ishlamadi (HTTP %d)", resp.StatusCode)
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		msg = "Kalit noto'g'ri yoki ruxsat yo'q"
	}
	return c.JSON(fiber.Map{"ok": false, "message": msg})
}

// POST /api/ai/generate-questions
func (h *AIHandler) GenerateQuestions(c *fiber.Ctx) error {
	var req GenerateQuestionsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Topic == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "topic is required"})
	}
	if req.Count <= 0 {
		req.Count = 5
	}
	if req.Language == "" {
		req.Language = "uz"
	}

	// Load settings from DB — prefer Groq (faster + free), fall back to OpenAI
	apiKey := h.loadSetting("groq_api_key", "")
	baseURL := "https://api.groq.com/openai/v1/chat/completions"
	model := h.loadSetting("ai_model", "llama-3.3-70b-versatile")

	if apiKey == "" {
		apiKey = h.loadSetting("openai_api_key", "")
		baseURL = "https://api.openai.com/v1/chat/completions"
		model = h.loadSetting("ai_model", "gpt-4o-mini")
	}
	if apiKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "AI API kaliti sozlanmagan. Sozlamalar sahifasiga o'ting.",
		})
	}

	// Build the system prompt
	qtypes := "multiple_choice"
	if len(req.QuestionTypes) > 0 {
		qtypes = strings.Join(req.QuestionTypes, ", ")
	}

	systemPrompt := fmt.Sprintf(
		`You are an educational quiz creator. Generate %d quiz questions about "%s" for grade %s students in Uzbek language.
Return ONLY valid JSON array, no markdown, no explanation:
[{"question_text":"...","type":"multiple_choice","options":[{"text":"...","is_correct":true},{"text":"...","is_correct":false},...],"explanation":"...","time_limit":20,"points":100}]
For true_false type: options must be [{"text":"To'g'ri","is_correct":bool},{"text":"Noto'g'ri","is_correct":bool}]
Question types to use: %s`,
		req.Count, req.Topic, req.GradeLevel, qtypes,
	)

	userContent := fmt.Sprintf("Generate %d questions about %s", req.Count, req.Topic)

	payload := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		Temperature: 0.7,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to build AI request"})
	}

	httpReq, err := http.NewRequest(http.MethodPost,
		baseURL,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create HTTP request"})
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "AI service unavailable: " + err.Error()})
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Failed to read AI response"})
	}

	var aiResp openAIResponse
	if err := json.Unmarshal(respBytes, &aiResp); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Failed to parse AI response"})
	}

	if aiResp.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "OpenAI error: " + aiResp.Error.Message})
	}

	if len(aiResp.Choices) == 0 {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "AI returned no choices"})
	}

	content := strings.TrimSpace(aiResp.Choices[0].Message.Content)

	// Strip markdown code fences if the model wrapped the JSON anyway
	if strings.HasPrefix(content, "```") {
		start := strings.Index(content, "\n")
		end := strings.LastIndex(content, "```")
		if start != -1 && end > start {
			content = strings.TrimSpace(content[start+1 : end])
		}
	}

	var questions []GeneratedQuestion
	if err := json.Unmarshal([]byte(content), &questions); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":   "Failed to parse questions from AI response",
			"raw":     content,
		})
	}

	return c.Status(fiber.StatusOK).JSON(questions)
}
