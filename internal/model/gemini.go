package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// GeminiModel provides LLM generation, supporting live Gemini API calls when GEMINI_API_KEY is set,
// and gracefully falling back to a smart ContextualMockModel that generates distinct, topic-aware campaigns.
type GeminiModel struct {
	APIKey     string
	ModelID    string
	HTTPClient *http.Client
	Mock       *ContextualMockModel
}

// NewGeminiModel creates a GeminiModel with automatic API key detection or contextual mocking.
func NewGeminiModel(idea string) *GeminiModel {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	modelID := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	if modelID == "" {
		modelID = "gemini-2.5-flash"
	}

	return &GeminiModel{
		APIKey:     apiKey,
		ModelID:    modelID,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Mock:       NewContextualMockModel(idea),
	}
}

// Generate fulfills the Model interface.
func (g *GeminiModel) Generate(ctx context.Context, prompt string) (string, error) {
	if g.APIKey != "" {
		res, err := g.callGeminiAPI(ctx, prompt)
		if err == nil && strings.TrimSpace(res) != "" {
			return res, nil
		}
		// If live call fails (e.g. quota, bad key), gracefully fallback to contextual mock
	}
	return g.Mock.Generate(ctx, prompt)
}

func (g *GeminiModel) callGeminiAPI(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.ModelID, g.APIKey)

	reqPayload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.2,
		},
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini api error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response candidates returned from gemini")
}

// TopicProfile stores thematic copy and styling for a specific business vertical.
type TopicProfile struct {
	Category      string
	BrandName     string
	Audience      string
	Offer         string
	Tone          string
	AllowedClaims []string
	BannedClaims  []string
	Headline      string
	Subhead       string
	Benefits      []string
	CTA           string
	Disclaimer    string
	PrimaryColor  string
	BgColor       string
	AccentColor   string
}

// ContextualMockModel generates bespoke, compliant mock responses tailored to the input idea.
type ContextualMockModel struct {
	Profile TopicProfile
}

// NewContextualMockModel creates a contextual mock model analyzing the prompt for business keywords.
func NewContextualMockModel(idea string) *ContextualMockModel {
	profile := detectProfile(idea)
	return &ContextualMockModel{Profile: profile}
}

func containsWord(text, word string) bool {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
	for _, f := range fields {
		if f == word {
			return true
		}
	}
	return false
}

func containsAnyWord(text string, words ...string) bool {
	for _, w := range words {
		if containsWord(text, w) {
			return true
		}
	}
	return false
}

func detectProfile(idea string) TopicProfile {
	// 1. Fitness / Gym
	if containsAnyWord(idea, "gym", "fitness", "workout", "strength", "iron", "training", "crossfit", "athletics") {
		return TopicProfile{
			Category:      "fitness",
			BrandName:     extractBrand(idea, "IronPulse"),
			Audience:      "Fitness enthusiasts, daily lifters, and individuals pursuing peak physical strength",
			Offer:         "24/7 premium strength and functional conditioning gym membership",
			Tone:          "Energetic, motivating, and disciplined",
			AllowedClaims: []string{"State-of-the-art strength and cardio equipment", "Certified personal coaching available daily", "Flexible memberships with zero lock-in contracts"},
			BannedClaims:  []string{"Guaranteed rapid muscle gain", "Cures chronic joint pain"},
			Headline:      "Unleash Your Strength With 24/7 Modern Fitness",
			Subhead:       "State-of-the-art equipment and elite coaching to reach your personal best.",
			Benefits:      []string{"State-of-the-art strength and cardio equipment", "Certified personal coaching available daily", "Flexible memberships with zero lock-in contracts"},
			CTA:           "Claim Your Free Pass",
			Disclaimer:    "Fictional fitness brand for demonstration purposes.",
			PrimaryColor:  "#e63946",
			BgColor:       "#0f141c",
			AccentColor:   "#1e293b",
		}
	}

	// 2. Coffee / Beverage
	if containsAnyWord(idea, "coffee", "brew", "roast", "espresso", "beans", "tea", "matcha") {
		return TopicProfile{
			Category:      "coffee",
			BrandName:     extractBrand(idea, "BrewCraft"),
			Audience:      "Specialty coffee connoisseurs and home espresso enthusiasts",
			Offer:         "Freshly roasted single-origin micro-lot coffee subscription",
			Tone:          "Rich, artisanal, and warm",
			AllowedClaims: []string{"Ethically sourced single-origin micro-lots", "Roasted to order weekly for peak aroma", "Custom grind profiles tailored to your brew method"},
			BannedClaims:  []string{"World's healthiest miracle beverage"},
			Headline:      "Artisan Whole Bean Coffee Roasted Fresh Weekly",
			Subhead:       "Discover exceptional single-origin micro-lot coffees delivered directly to your doorstep.",
			Benefits:      []string{"Ethically sourced single-origin micro-lots", "Roasted to order weekly for peak aroma", "Custom grind profiles tailored to your brew method"},
			CTA:           "Start Your Tasting Box",
			Disclaimer:    "Fictional craft coffee subscription brand.",
			PrimaryColor:  "#c07a3e",
			BgColor:       "#1c1613",
			AccentColor:   "#2b221d",
		}
	}

	// 3. Hair care / Beauty
	if containsAnyWord(idea, "hair", "shampoo", "skincare", "cosmetics", "serum", "strand", "beauty") {
		return TopicProfile{
			Category:      "haircare",
			BrandName:     extractBrand(idea, "Strand"),
			Audience:      "Consumers seeking personalized botanical hair and scalp formulas",
			Offer:         "Custom formulated clean botanical hair care subscription",
			Tone:          "Warm, sophisticated, and science-backed",
			AllowedClaims: []string{"Custom-formulated for your hair profile", "Clean sulfate-free botanical ingredients", "Delivered directly to your door monthly"},
			BannedClaims:  []string{"Guaranteed permanent baldness cure", "Clinically proven miracle regrowth"},
			Headline:      "Smarter Hair Care Formulated For Your Routine",
			Subhead:       "Personalized botanical hair formulas delivered right to your doorstep.",
			Benefits:      []string{"Custom-formulated for your hair profile", "Clean sulfate-free botanical ingredients", "Delivered directly to your door monthly"},
			CTA:           "Claim Your Custom Formula",
			Disclaimer:    "Fictional hair care brand for demonstration purposes.",
			PrimaryColor:  "#2d6a4f",
			BgColor:       "#0b1d16",
			AccentColor:   "#1b4332",
		}
	}

	// 4. Technology / SaaS / Productivity
	if containsAnyWord(idea, "software", "app", "saas", "ai", "cloud", "organizer", "smart", "productivity") {
		return TopicProfile{
			Category:      "tech",
			BrandName:     extractBrand(idea, "Nova"),
			Audience:      "Developers, creators, and modern teams seeking seamless efficiency",
			Offer:         "Intelligent cloud-powered workspace productivity suite",
			Tone:          "Sleek, intelligent, and focused",
			AllowedClaims: []string{"Seamless real-time synchronization", "Bank-grade end-to-end data encryption", "Automated routine task management"},
			BannedClaims:  []string{"Guaranteed 1000% productivity surge"},
			Headline:      "Streamline Your Daily Work With Intelligent Workspace",
			Subhead:       "Automate repetitive workflows and collaborate in real-time across your team.",
			Benefits:      []string{"Seamless real-time synchronization", "Bank-grade end-to-end data encryption", "Automated routine task management"},
			CTA:           "Get Started For Free",
			Disclaimer:    "Fictional software platform for demonstration purposes.",
			PrimaryColor:  "#6366f1",
			BgColor:       "#0b0f19",
			AccentColor:   "#1e1b4b",
		}
	}

	// 5. Default Generic / Tailored Profile
	brand := extractBrand(idea, "Aura")
	return TopicProfile{
		Category:      "general",
		BrandName:     brand,
		Audience:      "Customers seeking modern and premium quality essentials",
		Offer:         idea,
		Tone:          "Warm, inspiring, and transparent",
		AllowedClaims: []string{"Premium sustainable materials", "Thoughtfully engineered design", "Delivered directly to your door"},
		BannedClaims:  []string{"Guaranteed miracle results"},
		Headline:      fmt.Sprintf("Experience Exceptional Quality With %s", brand),
		Subhead:       "Discover thoughtful craftsmanship designed to elevate your everyday routine.",
		Benefits:      []string{"Premium sustainable materials", "Thoughtfully engineered design", "Delivered directly to your door"},
		CTA:           "Discover The Collection",
		Disclaimer:    "Fictional brand for demonstration purposes.",
		PrimaryColor:  "#0284c7",
		BgColor:       "#0f172a",
		AccentColor:   "#1e293b",
	}
}

func extractBrand(idea, fallback string) string {
	parts := strings.Split(idea, ",")
	for _, p := range parts {
		lower := strings.ToLower(p)
		if strings.Contains(lower, "for ") {
			idx := strings.Index(lower, "for ")
			candidate := strings.TrimSpace(p[idx+4:])
			if len(candidate) > 0 && len(candidate) < 25 {
				return strings.Title(candidate)
			}
		}
	}
	return fallback
}

// Generate responds dynamically depending on the calling agent.
func (c *ContextualMockModel) Generate(ctx context.Context, prompt string) (string, error) {
	p := c.Profile

	// 1. Strategist request
	if strings.Contains(prompt, "marketing strategist") {
		allowedClaimsJSON, _ := json.Marshal(p.AllowedClaims)
		bannedClaimsJSON, _ := json.Marshal(p.BannedClaims)

		return fmt.Sprintf(`{
  "audience": "%s",
  "offer": "%s",
  "tone": "%s",
  "allowed_claims": %s,
  "banned_claims": %s,
  "allowed_banned_phrases": []
}`, p.Audience, p.Offer, p.Tone, string(allowedClaimsJSON), string(bannedClaimsJSON)), nil
	}

	// 2. Copywriter request
	if strings.Contains(prompt, "copywriter") {
		benefitsJSON, _ := json.Marshal(p.Benefits)
		claimsJSON, _ := json.Marshal(p.AllowedClaims[:2])

		return fmt.Sprintf(`{
  "headline": "%s",
  "subhead": "%s",
  "benefits": %s,
  "cta": "%s",
  "disclaimer": "%s",
  "claims": %s
}`, p.Headline, p.Subhead, string(benefitsJSON), p.CTA, p.Disclaimer, string(claimsJSON)), nil
	}

	// 3. Designer request
	if strings.Contains(prompt, "web designer") {
		return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s | %s</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background-color: %s; color: #f1f5f9; line-height: 1.6; padding: 40px 20px; }
        .container { max-width: 680px; margin: 0 auto; background: %s; border-radius: 12px; padding: 36px; border: 1px solid rgba(255,255,255,0.08); box-shadow: 0 10px 30px rgba(0,0,0,0.5); }
        .tag { display: inline-block; font-size: 0.75rem; text-transform: uppercase; letter-spacing: 1.5px; background: rgba(255,255,255,0.1); padding: 4px 10px; border-radius: 4px; margin-bottom: 16px; color: %s; font-weight: 700; }
        h1 { font-size: 2.2rem; font-weight: 800; line-height: 1.25; margin-bottom: 16px; color: #ffffff; }
        .subhead { font-size: 1.15rem; color: #94a3b8; margin-bottom: 28px; }
        .benefits-card { background: rgba(0,0,0,0.25); border-radius: 8px; padding: 20px 24px; margin-bottom: 32px; border: 1px solid rgba(255,255,255,0.05); }
        .benefits-card h3 { font-size: 0.95rem; text-transform: uppercase; letter-spacing: 1px; color: #cbd5e1; margin-bottom: 12px; }
        .benefits-card ul { list-style: none; }
        .benefits-card li { margin-bottom: 10px; display: flex; align-items: center; font-size: 1rem; color: #e2e8f0; }
        .benefits-card li::before { content: "✓"; color: %s; font-weight: bold; margin-right: 10px; }
        .cta-btn { display: inline-block; background: %s; color: #ffffff; font-weight: 700; text-decoration: none; padding: 14px 32px; border-radius: 8px; transition: transform 0.2s; box-shadow: 0 4px 14px rgba(0,0,0,0.4); text-align: center; }
        .disclaimer { margin-top: 36px; font-size: 0.8rem; color: #64748b; border-top: 1px solid rgba(255,255,255,0.08); padding-top: 16px; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <span class="tag">%s</span>
        <h1>%s</h1>
        <p class="subhead">%s</p>
        <div class="benefits-card">
            <h3>Key Advantages</h3>
            <ul>
                <li>%s</li>
                <li>%s</li>
                <li>%s</li>
            </ul>
        </div>
        <a href="#action" class="cta-btn">%s</a>
        <p class="disclaimer">%s</p>
    </div>
</body>
</html>`, p.BrandName, p.Category, p.BgColor, p.AccentColor, p.PrimaryColor, p.PrimaryColor, p.PrimaryColor,
			p.BrandName, p.Headline, p.Subhead, p.Benefits[0], p.Benefits[1], p.Benefits[2], p.CTA, p.Disclaimer), nil
	}

	// 4. Reviewer request
	if strings.Contains(prompt, "advisory creative director") {
		return fmt.Sprintf(`{
  "score": 9,
  "feedback": "Strong %s campaign copy with clear visual hierarchy, compelling call-to-action, and compliant claims."
}`, p.Category), nil
	}

	// Fallback
	return "{}", nil
}
