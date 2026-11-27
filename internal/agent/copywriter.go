package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"adstudio/internal/graph"
	"adstudio/internal/model"
)

// GenerateCopy invokes the copywriter agent to write marketing copy based on the strategist brief.
func GenerateCopy(ctx context.Context, m model.Model, brief *graph.Brief, failedChecks []graph.CheckResult) (*graph.Copy, error) {
	if brief == nil {
		return nil, fmt.Errorf("copywriter requires a valid brief")
	}

	allowedClaimsJSON, _ := json.Marshal(brief.AllowedClaims)
	bannedClaimsJSON, _ := json.Marshal(brief.BannedClaims)

	var sb strings.Builder
	sb.WriteString("You are a direct-response copywriter for a fictional product landing page.\n")
	sb.WriteString("Create marketing copy strictly adhering to this campaign brief:\n\n")
	sb.WriteString(fmt.Sprintf("Audience: %s\n", brief.Audience))
	sb.WriteString(fmt.Sprintf("Offer: %s\n", brief.Offer))
	sb.WriteString(fmt.Sprintf("Tone: %s\n", brief.Tone))
	sb.WriteString(fmt.Sprintf("Allowed Claims (ONLY use these): %s\n", string(allowedClaimsJSON)))
	sb.WriteString(fmt.Sprintf("Banned Claims (NEVER use these): %s\n\n", string(bannedClaimsJSON)))

	sb.WriteString("Strict Compliance Constraints:\n")
	sb.WriteString("1. Headline MUST be 12 words or fewer.\n")
	sb.WriteString("2. Benefits list MUST have at least 3 distinct benefits.\n")
	sb.WriteString("3. NEVER use any of these banned phrases: 'guaranteed', 'cure', 'no side effects', 'clinically proven'.\n")
	sb.WriteString("4. The 'claims' array must ONLY contain items from Allowed Claims.\n")
	sb.WriteString("5. Return ONLY a valid JSON object matching the schema below.\n\n")

	if len(failedChecks) > 0 {
		sb.WriteString("REVISION REQUIRED: The previous draft failed the following deterministic checks:\n")
		for _, fc := range failedChecks {
			sb.WriteString(fmt.Sprintf("- [%s]: %s\n", fc.Name, fc.Message))
		}
		sb.WriteString("Fix every listed failure in this revised copy.\n\n")
	}

	sb.WriteString(`JSON Schema:
{
  "headline": "punchy headline with 12 words or fewer",
  "subhead": "supporting subhead",
  "benefits": ["benefit 1", "benefit 2", "benefit 3"],
  "cta": "action-oriented call to action button text",
  "disclaimer": "fictional brand disclaimer",
  "claims": ["allowed claim 1", "allowed claim 2"]
}`)

	raw, err := m.Generate(ctx, sb.String())
	if err != nil {
		return nil, fmt.Errorf("copywriter model call failed: %w", err)
	}

	sanitized := SanitizeJSON(raw)
	var copyResult graph.Copy
	if err := json.Unmarshal([]byte(sanitized), &copyResult); err != nil {
		return nil, fmt.Errorf("copywriter output malformed JSON: %w (raw output: %q)", err, raw)
	}

	return &copyResult, nil
}
