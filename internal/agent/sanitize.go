package agent

import (
	"strings"
)

// SanitizeJSON cleans a model output by stripping markdown code block wrappers
// and isolating the JSON payload.
func SanitizeJSON(raw string) string {
	trimmed := strings.TrimSpace(raw)

	// Strip leading ```json or ```
	if strings.HasPrefix(trimmed, "```") {
		lines := strings.Split(trimmed, "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "```") {
			lines = lines[1:]
		}
		if len(lines) > 0 && strings.HasSuffix(strings.TrimSpace(lines[len(lines)-1]), "```") {
			lines = lines[:len(lines)-1]
		}
		trimmed = strings.TrimSpace(strings.Join(lines, "\n"))
	}

	// Extract outermost JSON block if surrounding prose exists
	startObj := strings.Index(trimmed, "{")
	endObj := strings.LastIndex(trimmed, "}")
	if startObj != -1 && endObj != -1 && endObj > startObj {
		return trimmed[startObj : endObj+1]
	}

	startArr := strings.Index(trimmed, "[")
	endArr := strings.LastIndex(trimmed, "]")
	if startArr != -1 && endArr != -1 && endArr > startArr {
		return trimmed[startArr : endArr+1]
	}

	return trimmed
}

// SanitizeHTML strips code fences from model HTML output.
func SanitizeHTML(raw string) string {
	trimmed := strings.TrimSpace(raw)

	// Strip leading ```html or ```
	if strings.HasPrefix(trimmed, "```") {
		lines := strings.Split(trimmed, "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "```") {
			lines = lines[1:]
		}
		if len(lines) > 0 && strings.HasSuffix(strings.TrimSpace(lines[len(lines)-1]), "```") {
			lines = lines[:len(lines)-1]
		}
		trimmed = strings.TrimSpace(strings.Join(lines, "\n"))
	}

	return trimmed
}
