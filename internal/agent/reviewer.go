package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"adstudio/internal/model"
)

// ReviewPage invokes the reviewer agent to provide advisory tone and clarity feedback.
// Note: This score is advisory and logged only; it never gates or overrides checker outcomes.
func ReviewPage(ctx context.Context, m model.Model, c *Copy, html string) (*Review, error) {
	var sb strings.Builder
	sb.WriteString("You are an advisory creative director reviewing a fictional product landing page.\n")
	sb.WriteString("Evaluate the copy and HTML for tone, clarity, and overall marketing effectiveness.\n\n")

	if c != nil {
		sb.WriteString(fmt.Sprintf("Headline: %s\n", c.Headline))
		sb.WriteString(fmt.Sprintf("Subhead: %s\n", c.Subhead))
		sb.WriteString(fmt.Sprintf("CTA: %s\n\n", c.CTA))
	}

	sb.WriteString("Return ONLY a JSON object with this schema:\n")
	sb.WriteString(`{
  "score": 8,
  "feedback": "constructive advisory critique of clarity and tone"
}`)

	raw, err := m.Generate(ctx, sb.String())
	if err != nil {
		return nil, fmt.Errorf("reviewer model call failed: %w", err)
	}

	sanitized := SanitizeJSON(raw)
	var review Review
	if err := json.Unmarshal([]byte(sanitized), &review); err != nil {
		// Provide graceful advisory fallback if reviewer output had unparseable formatting
		return &Review{
			Score:    5,
			Feedback: fmt.Sprintf("Advisory reviewer note: %s", strings.TrimSpace(raw)),
		}, nil
	}

	return &review, nil
}
