package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"adstudio/internal/model"
)

// GenerateBrief invokes the strategist agent to develop a marketing strategy brief.
func GenerateBrief(ctx context.Context, m model.Model, idea string) (*Brief, error) {
	prompt := fmt.Sprintf(`You are an expert marketing strategist for a fictional brand campaign.
Analyze the following idea and produce a comprehensive campaign brief:

Idea: %s

Strict Requirements:
1. Everything must be fictional. No real brands, no medical or disease claims.
2. Return ONLY a valid JSON object with no commentary or markdown wrappers.

JSON Schema:
{
  "audience": "string describing target demographic",
  "offer": "string describing the core product offer",
  "tone": "string describing the messaging tone",
  "allowed_claims": ["array", "of", "safe", "fictional", "claims"],
  "banned_claims": ["array", "of", "strictly", "prohibited", "claims"],
  "allowed_banned_phrases": []
}`, idea)

	raw, err := m.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("strategist model call failed: %w", err)
	}

	sanitized := SanitizeJSON(raw)
	var brief Brief
	if err := json.Unmarshal([]byte(sanitized), &brief); err != nil {
		return nil, fmt.Errorf("strategist output malformed JSON: %w (raw output: %q)", err, raw)
	}

	if strings.TrimSpace(brief.Audience) == "" || strings.TrimSpace(brief.Offer) == "" {
		return nil, fmt.Errorf("strategist output missing required audience or offer fields")
	}

	return &brief, nil
}
