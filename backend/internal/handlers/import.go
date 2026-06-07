package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
)

// ImportHandler handles AI-powered document question import.
type ImportHandler struct {
	httpClient *http.Client
	loadSetting func(key, def string) string
}

func NewImportHandler(aiH *AIHandler) *ImportHandler {
	return &ImportHandler{
		httpClient:  aiH.httpClient,
		loadSetting: aiH.loadSetting,
	}
}

// ─── Extracted document content ──────────────────────────────────────────────

type extractedDoc struct {
	text   string
	images []extractedImage // base64 PNG/JPEG
}

type extractedImage struct {
	mimeType string
	data     string // base64
}

// ─── DOCX parser ─────────────────────────────────────────────────────────────

// docxBody is a minimal struct to unmarshal word/document.xml
type docxBody struct {
	XMLName xml.Name    `xml:"document"`
	Body    docxContent `xml:"body"`
}
type docxContent struct {
	Paragraphs []docxPara `xml:"p"`
	Tables     []docxTable `xml:"tbl"`
}
type docxPara struct {
	Runs  []docxRun  `xml:"r"`
	Math  []docxMath `xml:"oMath"`
}
type docxRun struct {
	Text string `xml:"t"`
}
type docxMath struct {
	InnerXML string `xml:",innerxml"`
}
type docxTable struct {
	Rows []docxRow `xml:"tr"`
}
type docxRow struct {
	Cells []docxCell `xml:"tc"`
}
type docxCell struct {
	Paragraphs []docxPara `xml:"p"`
}

func parseDocx(data []byte) (*extractedDoc, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("not a valid docx file: %w", err)
	}

	doc := &extractedDoc{}
	var sb strings.Builder

	for _, f := range r.File {
		switch {
		case f.Name == "word/document.xml":
			rc, err := f.Open()
			if err != nil {
				continue
			}
			xmlBytes, _ := io.ReadAll(rc)
			rc.Close()

			var body docxBody
			// Strip namespace prefixes for simpler matching
			cleaned := stripNSPrefixes(xmlBytes)
			if err := xml.Unmarshal(cleaned, &body); err != nil {
				// Fallback: extract text between XML tags
				sb.WriteString(xmlStripTags(string(xmlBytes)))
				continue
			}
			writeDocxContent(&sb, body.Body)

		case strings.HasPrefix(f.Name, "word/media/"):
			ext := strings.ToLower(filepath.Ext(f.Name))
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".gif" {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				continue
			}
			imgBytes, _ := io.ReadAll(rc)
			rc.Close()
			mime := "image/png"
			if ext == ".jpg" || ext == ".jpeg" {
				mime = "image/jpeg"
			}
			doc.images = append(doc.images, extractedImage{
				mimeType: mime,
				data:     base64.StdEncoding.EncodeToString(imgBytes),
			})
		}
	}

	doc.text = sb.String()
	return doc, nil
}

func writeDocxContent(sb *strings.Builder, body docxContent) {
	for _, p := range body.Paragraphs {
		var line strings.Builder
		for _, r := range p.Runs {
			line.WriteString(r.Text)
		}
		for _, m := range p.Math {
			line.WriteString("[formula:" + flattenMathXML(m.InnerXML) + "]")
		}
		if t := strings.TrimSpace(line.String()); t != "" {
			sb.WriteString(t)
			sb.WriteByte('\n')
		}
	}
	for _, tbl := range body.Tables {
		for _, row := range tbl.Rows {
			var cells []string
			for _, cell := range row.Cells {
				var ct strings.Builder
				for _, p := range cell.Paragraphs {
					for _, r := range p.Runs {
						ct.WriteString(r.Text)
					}
				}
				if t := strings.TrimSpace(ct.String()); t != "" {
					cells = append(cells, t)
				}
			}
			if len(cells) > 0 {
				sb.WriteString(strings.Join(cells, " | "))
				sb.WriteByte('\n')
			}
		}
	}
}

// stripNSPrefixes removes XML namespace prefixes (w:, m:, etc.)
func stripNSPrefixes(data []byte) []byte {
	s := string(data)
	// Remove xmlns declarations
	for _, prefix := range []string{"w:", "m:", "r:", "mc:", "wp:", "a:", "pic:", "v:", "o:", "w14:", "w15:"} {
		s = strings.ReplaceAll(s, "<"+prefix, "<")
		s = strings.ReplaceAll(s, "</"+prefix, "</")
	}
	return []byte(s)
}

func xmlStripTags(s string) string {
	var sb strings.Builder
	inTag := false
	for _, ch := range s {
		if ch == '<' {
			inTag = true
		} else if ch == '>' {
			inTag = false
			sb.WriteByte('\n')
		} else if !inTag && !unicode.IsControl(ch) {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}

// flattenMathXML converts OMML XML to a readable text representation
func flattenMathXML(inner string) string {
	stripped := xmlStripTags(inner)
	parts := strings.Fields(stripped)
	return strings.Join(parts, " ")
}

// ─── XLSX parser ──────────────────────────────────────────────────────────────

func parseXlsx(data []byte) (*extractedDoc, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid xlsx: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}
		sb.WriteString("=== " + sheet + " ===\n")
		for _, row := range rows {
			clean := make([]string, 0, len(row))
			for _, cell := range row {
				if t := strings.TrimSpace(cell); t != "" {
					clean = append(clean, t)
				}
			}
			if len(clean) > 0 {
				sb.WriteString(strings.Join(clean, " | "))
				sb.WriteByte('\n')
			}
		}
	}
	return &extractedDoc{text: sb.String()}, nil
}

// ─── PDF parser ───────────────────────────────────────────────────────────────

func parsePDF(data []byte) (*extractedDoc, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("invalid pdf: %w", err)
	}

	var sb strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteByte('\n')
	}
	return &extractedDoc{text: sb.String()}, nil
}

// ─── AI question extraction ───────────────────────────────────────────────────

const importSystemPrompt = `You are an expert educational quiz parser.
Extract quiz questions from the provided document text.

STRICT RULES:
1. Output MAXIMUM 20 questions per response — if more exist, output only the first 20
2. Identify question numbers, letters, or bullets to separate questions
3. For multiple choice: find all options (A/B/C/D or 1/2/3/4 or similar)
4. Detect correct answers from: answer keys, asterisks (*), bold, "Answer:" labels, checkmarks (✓)
5. For true/false: options must be [{"text":"To'g'ri","is_correct":bool},{"text":"Noto'g'ri","is_correct":bool}]
6. For short answer / fill-blank: use type "short_answer", put expected answer in options[0].text
7. Preserve math formulas as LaTeX: $formula$
8. If answer unknown, make best educational guess — never leave is_correct ambiguous
9. Assign time_limit (10-60s) and points (50-200) by difficulty

CRITICAL: Your ENTIRE response must be ONLY a valid complete JSON array.
Start with [ and end with ]. No markdown, no text before or after.
Every object must be fully closed. The array MUST end with ]

Format:
[{"question_text":"...","type":"multiple_choice","options":[{"text":"...","is_correct":true},{"text":"...","is_correct":false}],"explanation":"...","time_limit":20,"points":100}]

Types: multiple_choice | true_false | short_answer | fill_blank`

type visionContent struct {
	Type     string           `json:"type"`
	Text     string           `json:"text,omitempty"`
	ImageURL *visionImageURL  `json:"image_url,omitempty"`
}
type visionImageURL struct {
	URL string `json:"url"`
}
type visionMessage struct {
	Role    string          `json:"role"`
	Content []visionContent `json:"content"`
}
type visionRequest struct {
	Model       string          `json:"model"`
	Messages    []visionMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
}

func (h *ImportHandler) extractWithAI(doc *extractedDoc, fileType string) ([]GeneratedQuestion, error) {
	apiKey := h.loadSetting("groq_api_key", "")
	baseURL := "https://api.groq.com/openai/v1/chat/completions"
	model := h.loadSetting("ai_model", "llama-3.3-70b-versatile")

	maxTokens := 6000
	hasVision := false
	if apiKey == "" || len(doc.images) > 0 {
		apiKey = h.loadSetting("openai_api_key", "")
		baseURL = "https://api.openai.com/v1/chat/completions"
		model = "gpt-4o"
		hasVision = true
		maxTokens = 12000
	}
	if apiKey == "" {
		return nil, fmt.Errorf("AI API kaliti sozlanmagan")
	}

	// Split document text into 3000-char chunks so each AI call stays within limits
	// Each chunk gets max 20 questions → up to 3 chunks = up to 60 questions total
	const chunkSize = 3000
	const maxChunks = 3
	textRunes := []rune(doc.text)
	var chunks []string
	for i := 0; i < len(textRunes) && len(chunks) < maxChunks; i += chunkSize {
		end := i + chunkSize
		if end > len(textRunes) {
			end = len(textRunes)
		}
		chunks = append(chunks, string(textRunes[i:end]))
	}
	if len(chunks) == 0 {
		chunks = []string{""}
	}

	var allQuestions []GeneratedQuestion
	client := &http.Client{Timeout: 90 * time.Second}

	for chunkIdx, chunkText := range chunks {
		userText := fmt.Sprintf("Document type: %s (part %d/%d)\n\nExtracted content:\n%s",
			fileType, chunkIdx+1, len(chunks), chunkText)

		userContent := []visionContent{{Type: "text", Text: userText}}
		// Attach images only to first chunk
		if hasVision && chunkIdx == 0 && len(doc.images) > 0 {
			for i, img := range doc.images {
				if i >= 6 {
					break
				}
				userContent = append(userContent, visionContent{
					Type: "image_url",
					ImageURL: &visionImageURL{
						URL: fmt.Sprintf("data:%s;base64,%s", img.mimeType, img.data),
					},
				})
			}
		}

		reqBody := visionRequest{
			Model: model,
			Messages: []visionMessage{
				{Role: "system", Content: []visionContent{{Type: "text", Text: importSystemPrompt}}},
				{Role: "user", Content: userContent},
			},
			Temperature: 0.1,
			MaxTokens:   maxTokens,
		}

		bodyBytes, _ := json.Marshal(reqBody)
		httpReq, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(bodyBytes))
		if err != nil {
			continue
		}
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(httpReq)
		if err != nil {
			continue
		}
		respBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var aiResp openAIResponse
		if err := json.Unmarshal(respBytes, &aiResp); err != nil {
			continue
		}
		if aiResp.Error != nil || len(aiResp.Choices) == 0 {
			continue
		}

		content := cleanContent(aiResp.Choices[0].Message.Content)
		finishReason := aiResp.Choices[0].FinishReason

		var qs []GeneratedQuestion
		if err := json.Unmarshal([]byte(content), &qs); err != nil {
			// If cut off by token limit, try depth-aware repair
			if finishReason == "length" || strings.Contains(err.Error(), "unexpected end") {
				repaired := repairJSON(content)
				if repaired != content {
					_ = json.Unmarshal([]byte(repaired), &qs)
				}
			}
		}
		allQuestions = append(allQuestions, qs...)

		// Stop if we already have plenty of questions
		if len(allQuestions) >= 40 {
			break
		}
	}

	if len(allQuestions) == 0 {
		return nil, fmt.Errorf("Hujjatda savollar topilmadi yoki AI savollarni tahlil qila olmadi")
	}
	return allQuestions, nil
}

// repairJSON fixes a truncated JSON array by finding the last complete
// top-level object (depth 1→0 transition) and closing the array after it.
func repairJSON(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") {
		return s
	}

	depth := 0
	inString := false
	escape := false
	lastCompletePos := -1 // position of } that closes a top-level question

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if escape {
			escape = false
			continue
		}
		if inString {
			if ch == '\\' {
				escape = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '[', '{':
			depth++
		case ']', '}':
			depth--
			// depth==1 after closing '}' means we just finished a top-level object
			if depth == 1 && ch == '}' {
				lastCompletePos = i
			}
		}
	}

	if lastCompletePos == -1 {
		return s
	}

	trimmed := strings.TrimRight(strings.TrimSpace(s[:lastCompletePos+1]), ", \t\n\r")
	return trimmed + "]"
}

// cleanContent strips markdown code fences and trims whitespace.
func cleanContent(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		start := strings.Index(s, "\n")
		end := strings.LastIndex(s, "```")
		if start != -1 && end > start {
			s = strings.TrimSpace(s[start+1 : end])
		}
	}
	return s
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "\n...[truncated]"
}

// ─── HTTP Handler ─────────────────────────────────────────────────────────────

// POST /api/quizzes/import-file
func (h *ImportHandler) ImportFile(c *fiber.Ctx) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Fayl topilmadi"})
	}
	if fh.Size > 20*1024*1024 {
		return c.Status(400).JSON(fiber.Map{"error": "Fayl hajmi 20MB dan oshmasligi kerak"})
	}

	f, err := fh.Open()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Faylni ochib bo'lmadi"})
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Faylni o'qib bo'lmadi"})
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	var doc *extractedDoc
	var fileType string

	switch ext {
	case ".docx":
		fileType = "Microsoft Word (DOCX)"
		doc, err = parseDocx(data)
	case ".xlsx", ".xls":
		fileType = "Microsoft Excel (XLSX)"
		doc, err = parseXlsx(data)
	case ".pdf":
		fileType = "PDF"
		doc, err = parsePDF(data)
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Qo'llab-quvvatlanmaydigan format. .docx, .xlsx yoki .pdf yuklang"})
	}

	if err != nil {
		return c.Status(422).JSON(fiber.Map{"error": "Faylni tahlil qilib bo'lmadi: " + err.Error()})
	}

	if strings.TrimSpace(doc.text) == "" && len(doc.images) == 0 {
		return c.Status(422).JSON(fiber.Map{"error": "Faylda matn topilmadi"})
	}

	questions, err := h.extractWithAI(doc, fileType)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": err.Error()})
	}
	if len(questions) == 0 {
		return c.Status(422).JSON(fiber.Map{"error": "Hujjatda savollar topilmadi"})
	}

	return c.JSON(fiber.Map{
		"questions":  questions,
		"count":      len(questions),
		"file_type":  fileType,
		"image_count": len(doc.images),
		"text_len":   len([]rune(doc.text)),
	})
}
