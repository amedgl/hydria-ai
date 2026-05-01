// internal/vision/vision.go
// Gemini Vision API integration.
// Reads an image file, sends it to Gemini, and parses personal OSINT clues.
package vision

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// AnalysisResult holds all personal clues extracted from the target image.
type AnalysisResult struct {
	Names       []string `json:"names"`
	Dates       []string `json:"dates"`
	Pets        []string `json:"pets"`
	Locations   []string `json:"locations"`
	Interests   []string `json:"interests"`
	Numbers     []string `json:"numbers"`
	Brands      []string `json:"brands"`
	CustomHints []string `json:"custom_hints"`
}

// CountHints returns the total number of hints across all categories.
func (a AnalysisResult) CountHints() int {
	return len(a.Names) + len(a.Dates) + len(a.Pets) + len(a.Locations) +
		len(a.Interests) + len(a.Numbers) + len(a.Brands) + len(a.CustomHints)
}

// ToDisplayRows converts the result to UI-friendly rows.
func (a AnalysisResult) ToDisplayRows() []struct{ Category, Values string } {
	rows := []struct{ Category, Values string }{
		{"👤 Names", strings.Join(a.Names, ", ")},
		{"📅 Dates", strings.Join(a.Dates, ", ")},
		{"🐾 Pets", strings.Join(a.Pets, ", ")},
		{"🌍 Locations", strings.Join(a.Locations, ", ")},
		{"⚽ Interests", strings.Join(a.Interests, ", ")},
		{"🔢 Numbers", strings.Join(a.Numbers, ", ")},
		{"💼 Brands", strings.Join(a.Brands, ", ")},
		{"💡 Suggestions", strings.Join(a.CustomHints, ", ")},
	}
	var out []struct{ Category, Values string }
	for _, r := range rows {
		if r.Values != "" {
			out = append(out, r)
		}
	}
	return out
}

const analysisPrompt = `
You are a cybersecurity OSINT expert. Analyze the given image and extract personal information
about the target that could be useful for password guessing.

Respond ONLY with valid JSON in this exact format (no markdown, no extra text):

{
  "names": ["first name", "last name", "nickname", "username suggestions"],
  "dates": ["1990", "19900515", "1990-05-15", "0515"],
  "pets": ["pet names"],
  "locations": ["city", "country", "neighborhood"],
  "interests": ["hobbies", "sports teams", "games", "bands"],
  "numbers": ["phone fragments", "postal code", "visible numbers"],
  "brands": ["clothing", "car", "tech brands"],
  "custom_hints": ["top 5-10 most likely password guesses"]
}

Rules:
- All values must be lowercase
- Generate multiple date formats (year, month-day, full date)
- Use [] for empty categories
- If no useful info found, return all empty lists
`

var supportedExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true,
	".webp": true, ".gif": true, ".bmp": true,
}

func mimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	}
	return "image/jpeg"
}

// AnalyzeImage sends the image to Gemini Vision and returns parsed clues.
func AnalyzeImage(ctx context.Context, imagePath, apiKey, modelName string) (AnalysisResult, error) {
	ext := strings.ToLower(filepath.Ext(imagePath))
	if !supportedExts[ext] {
		return AnalysisResult{}, fmt.Errorf("unsupported image format: %s", ext)
	}

	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return AnalysisResult{}, fmt.Errorf("read image: %w", err)
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return AnalysisResult{}, fmt.Errorf("create gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(modelName)
	model.SetTemperature(0.2)
	model.SetMaxOutputTokens(4096)

	resp, err := model.GenerateContent(ctx,
		genai.Text(analysisPrompt),
		genai.ImageData(mimeType(imagePath), imgData),
	)
	if err != nil {
		return AnalysisResult{}, fmt.Errorf("gemini generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return AnalysisResult{}, fmt.Errorf("gemini returned empty response")
	}

	// Collect all text parts
	var rawBuilder strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			rawBuilder.WriteString(string(t))
		}
	}
	raw := strings.TrimSpace(rawBuilder.String())

	// Strip markdown code block if present
	if strings.HasPrefix(raw, "```") {
		lines := strings.Split(raw, "\n")
		if len(lines) > 2 {
			raw = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	var result AnalysisResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return AnalysisResult{}, fmt.Errorf("parse gemini JSON: %w\nRaw: %s", err, raw)
	}

	// Normalize: lowercase and trim all values
	normalize := func(ss []string) []string {
		var out []string
		for _, s := range ss {
			s = strings.ToLower(strings.TrimSpace(s))
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	result.Names = normalize(result.Names)
	result.Dates = normalize(result.Dates)
	result.Pets = normalize(result.Pets)
	result.Locations = normalize(result.Locations)
	result.Interests = normalize(result.Interests)
	result.Numbers = normalize(result.Numbers)
	result.Brands = normalize(result.Brands)
	result.CustomHints = normalize(result.CustomHints)

	return result, nil
}
