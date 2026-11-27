package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"adstudio/internal/graph"
	"adstudio/internal/model"
)

// GenerateHTML invokes the designer agent to generate a clean, responsive HTML landing page.
func GenerateHTML(ctx context.Context, m model.Model, brief *graph.Brief, c *graph.Copy, failedChecks []graph.CheckResult) (string, error) {
	if c == nil {
		return "", fmt.Errorf("designer requires valid copy")
	}

	benefitsJSON, _ := json.Marshal(c.Benefits)

	var sb strings.Builder
	sb.WriteString("You are a web designer creating a modern HTML landing page for a fictional brand.\n")
	sb.WriteString("Build a complete, standalone HTML page using the approved marketing copy:\n\n")
	sb.WriteString(fmt.Sprintf("Headline: %s\n", c.Headline))
	sb.WriteString(fmt.Sprintf("Subhead: %s\n", c.Subhead))
	sb.WriteString(fmt.Sprintf("Benefits: %s\n", string(benefitsJSON)))
	sb.WriteString(fmt.Sprintf("CTA: %s\n", c.CTA))
	sb.WriteString(fmt.Sprintf("Disclaimer: %s\n\n", c.Disclaimer))

	sb.WriteString("Strict HTML Constraints:\n")
	sb.WriteString("1. Standalone valid HTML5 document (<!DOCTYPE html><html lang=\"en\">...</html>).\n")
	sb.WriteString("2. MUST include <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"> in the <head>.\n")
	sb.WriteString("3. MUST have EXACTLY ONE <h1> tag containing the headline.\n")
	sb.WriteString("4. MUST explicitly include sections for Headline, Subhead, Benefits (bulleted list), CTA button or link, and Disclaimer text.\n")
	sb.WriteString("5. Use modern inline CSS in <style> for elegant typography, padding, and mobile responsiveness.\n")
	sb.WriteString("6. Return ONLY the raw HTML code without markdown code blocks.\n\n")

	if len(failedChecks) > 0 {
		sb.WriteString("REVISION REQUIRED: The previous HTML draft failed the following deterministic checks:\n")
		for _, fc := range failedChecks {
			sb.WriteString(fmt.Sprintf("- [%s]: %s\n", fc.Name, fc.Message))
		}
		sb.WriteString("Fix every listed failure in this revised HTML.\n\n")
	}

	raw, err := m.Generate(ctx, sb.String())
	if err != nil {
		return "", fmt.Errorf("designer model call failed: %w", err)
	}

	sanitized := SanitizeHTML(raw)
	if strings.TrimSpace(sanitized) == "" {
		return "", fmt.Errorf("designer produced empty HTML output")
	}

	return sanitized, nil
}
