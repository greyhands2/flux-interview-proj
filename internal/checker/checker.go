package checker

import (
	"fmt"
	"regexp"
	"strings"

	"adstudio/internal/graph"
)

// Standard check identifiers.
const (
	NameRequiredCopySections = "required_copy_sections"
	NameRequiredHTMLSections = "required_html_sections"
	NameBannedPhrases        = "banned_phrases"
	NameClaims               = "claims"
	NameHeadlineLength       = "headline_length"
	NameViewportMeta         = "viewport_meta"
	NameSingleH1             = "single_h1"
	NameWellFormed           = "well_formed"
)

// Default banned phrases (case-insensitive).
var DefaultBannedPhrases = []string{
	"guaranteed",
	"cure",
	"no side effects",
	"clinically proven",
}

var (
	viewportRegex = regexp.MustCompile(`(?i)<meta\s+[^>]*name=["']viewport["'][^>]*>`)
	h1TagRegex    = regexp.MustCompile(`(?i)<h1\b[^>]*>`)
)

// CheckRequiredCopySections ensures Headline, Benefits (at least 3), CTA, and Disclaimer exist in the copy.
func CheckRequiredCopySections(c *graph.Copy) graph.CheckResult {
	if c == nil {
		return graph.CheckResult{
			Name:    NameRequiredCopySections,
			Passed:  false,
			Message: "copy is nil",
		}
	}

	var missing []string
	if strings.TrimSpace(c.Headline) == "" {
		missing = append(missing, "headline")
	}
	if len(c.Benefits) < 3 {
		missing = append(missing, fmt.Sprintf("benefits (need at least 3, got %d)", len(c.Benefits)))
	} else {
		for i, b := range c.Benefits {
			if strings.TrimSpace(b) == "" {
				missing = append(missing, fmt.Sprintf("benefit[%d] is empty", i))
			}
		}
	}
	if strings.TrimSpace(c.CTA) == "" {
		missing = append(missing, "cta")
	}
	if strings.TrimSpace(c.Disclaimer) == "" {
		missing = append(missing, "disclaimer")
	}

	if len(missing) > 0 {
		return graph.CheckResult{
			Name:    NameRequiredCopySections,
			Passed:  false,
			Message: "missing required copy section(s): " + strings.Join(missing, ", "),
		}
	}

	return graph.CheckResult{
		Name:   NameRequiredCopySections,
		Passed: true,
	}
}

// CheckRequiredHTMLSections ensures the rendered HTML includes Headline, Benefits, CTA, and Disclaimer.
func CheckRequiredHTMLSections(c *graph.Copy, html string) graph.CheckResult {
	if strings.TrimSpace(html) == "" {
		return graph.CheckResult{
			Name:    NameRequiredHTMLSections,
			Passed:  false,
			Message: "HTML is empty",
		}
	}

	lowerHTML := strings.ToLower(html)
	var missing []string

	if c != nil {
		if strings.TrimSpace(c.Headline) != "" && !strings.Contains(lowerHTML, strings.ToLower(strings.TrimSpace(c.Headline))) {
			missing = append(missing, "headline text")
		}
		if strings.TrimSpace(c.CTA) != "" && !strings.Contains(lowerHTML, strings.ToLower(strings.TrimSpace(c.CTA))) {
			missing = append(missing, "cta text")
		}
		if strings.TrimSpace(c.Disclaimer) != "" && !strings.Contains(lowerHTML, strings.ToLower(strings.TrimSpace(c.Disclaimer))) {
			missing = append(missing, "disclaimer text")
		}
		hasBenefit := false
		for _, b := range c.Benefits {
			if strings.TrimSpace(b) != "" && strings.Contains(lowerHTML, strings.ToLower(strings.TrimSpace(b))) {
				hasBenefit = true
				break
			}
		}
		if len(c.Benefits) > 0 && !hasBenefit {
			missing = append(missing, "benefits text")
		}
	}

	if len(missing) > 0 {
		return graph.CheckResult{
			Name:    NameRequiredHTMLSections,
			Passed:  false,
			Message: "HTML missing required section(s): " + strings.Join(missing, ", "),
		}
	}

	return graph.CheckResult{
		Name:   NameRequiredHTMLSections,
		Passed: true,
	}
}

// CheckBannedPhrases scans copy and HTML for banned phrases unless explicitly allowed in brief.
func CheckBannedPhrases(b *graph.Brief, c *graph.Copy, html string) graph.CheckResult {
	allowedSet := make(map[string]bool)
	if b != nil {
		for _, phrase := range b.AllowedBannedPhrases {
			allowedSet[strings.ToLower(strings.TrimSpace(phrase))] = true
		}
	}

	// Gather text to scan
	var corpus []string
	if c != nil {
		corpus = append(corpus, c.Headline, c.Subhead, c.CTA, c.Disclaimer)
		corpus = append(corpus, c.Benefits...)
		corpus = append(corpus, c.Claims...)
	}
	corpus = append(corpus, html)
	combinedText := strings.ToLower(strings.Join(corpus, " "))

	for _, phrase := range DefaultBannedPhrases {
		lowerPhrase := strings.ToLower(phrase)
		if allowedSet[lowerPhrase] {
			continue
		}
		if strings.Contains(combinedText, lowerPhrase) {
			return graph.CheckResult{
				Name:    NameBannedPhrases,
				Passed:  false,
				Message: fmt.Sprintf("banned phrase detected: '%s'", phrase),
			}
		}
	}

	return graph.CheckResult{
		Name:   NameBannedPhrases,
		Passed: true,
	}
}

// CheckClaims verifies that all copy claims are strictly present in the brief's allowed claims.
func CheckClaims(b *graph.Brief, c *graph.Copy) graph.CheckResult {
	if b == nil {
		return graph.CheckResult{
			Name:    NameClaims,
			Passed:  false,
			Message: "brief is nil",
		}
	}
	if c == nil {
		return graph.CheckResult{
			Name:    NameClaims,
			Passed:  false,
			Message: "copy is nil",
		}
	}

	allowed := make(map[string]bool)
	for _, ac := range b.AllowedClaims {
		allowed[strings.ToLower(strings.TrimSpace(ac))] = true
	}

	for _, claim := range c.Claims {
		norm := strings.ToLower(strings.TrimSpace(claim))
		if norm == "" {
			continue
		}
		if !allowed[norm] {
			return graph.CheckResult{
				Name:    NameClaims,
				Passed:  false,
				Message: fmt.Sprintf("claim not in allowed claims: '%s'", claim),
			}
		}
	}

	return graph.CheckResult{
		Name:   NameClaims,
		Passed: true,
	}
}

// CheckHeadlineLength ensures headline is 12 words or fewer.
func CheckHeadlineLength(c *graph.Copy) graph.CheckResult {
	if c == nil {
		return graph.CheckResult{
			Name:    NameHeadlineLength,
			Passed:  false,
			Message: "copy is nil",
		}
	}

	words := strings.Fields(c.Headline)
	if len(words) > 12 {
		return graph.CheckResult{
			Name:    NameHeadlineLength,
			Passed:  false,
			Message: fmt.Sprintf("headline exceeds 12 words (got %d words)", len(words)),
		}
	}

	return graph.CheckResult{
		Name:   NameHeadlineLength,
		Passed: true,
	}
}

// CheckViewportMeta ensures the HTML contains a viewport meta tag.
func CheckViewportMeta(html string) graph.CheckResult {
	if !viewportRegex.MatchString(html) {
		return graph.CheckResult{
			Name:    NameViewportMeta,
			Passed:  false,
			Message: "HTML missing viewport meta tag",
		}
	}

	return graph.CheckResult{
		Name:   NameViewportMeta,
		Passed: true,
	}
}

// CheckSingleH1 ensures the HTML has exactly one <h1> tag.
func CheckSingleH1(html string) graph.CheckResult {
	matches := h1TagRegex.FindAllString(html, -1)
	count := len(matches)

	if count == 0 {
		return graph.CheckResult{
			Name:    NameSingleH1,
			Passed:  false,
			Message: "HTML has no <h1> tag",
		}
	}
	if count > 1 {
		return graph.CheckResult{
			Name:    NameSingleH1,
			Passed:  false,
			Message: fmt.Sprintf("HTML has %d <h1> tags (expected exactly 1)", count),
		}
	}

	return graph.CheckResult{
		Name:   NameSingleH1,
		Passed: true,
	}
}

// CheckWellFormed ensures all required elements exist and parse cleanly.
func CheckWellFormed(b *graph.Brief, c *graph.Copy, html string) graph.CheckResult {
	if b == nil {
		return graph.CheckResult{
			Name:    NameWellFormed,
			Passed:  false,
			Message: "strategist brief is missing or malformed",
		}
	}
	if c == nil {
		return graph.CheckResult{
			Name:    NameWellFormed,
			Passed:  false,
			Message: "copywriter output is missing or malformed",
		}
	}
	if strings.TrimSpace(html) == "" {
		return graph.CheckResult{
			Name:    NameWellFormed,
			Passed:  false,
			Message: "designer HTML is missing or empty",
		}
	}

	return graph.CheckResult{
		Name:   NameWellFormed,
		Passed: true,
	}
}

// RunAllChecks runs all 7 deterministic checks on the generated artifacts.
func RunAllChecks(b *graph.Brief, c *graph.Copy, html string) []graph.CheckResult {
	return []graph.CheckResult{
		CheckWellFormed(b, c, html),
		CheckRequiredCopySections(c),
		CheckRequiredHTMLSections(c, html),
		CheckBannedPhrases(b, c, html),
		CheckClaims(b, c),
		CheckHeadlineLength(c),
		CheckViewportMeta(html),
		CheckSingleH1(html),
	}
}

// AllPassed returns true if every check passed.
func AllPassed(results []graph.CheckResult) bool {
	for _, r := range results {
		if !r.Passed {
			return false
		}
	}
	return true
}

// HasCopyFailure checks if any copy-specific rule failed.
func HasCopyFailure(results []graph.CheckResult) bool {
	for _, r := range results {
		if !r.Passed {
			switch r.Name {
			case NameRequiredCopySections, NameBannedPhrases, NameClaims, NameHeadlineLength:
				return true
			}
		}
	}
	return false
}

// HasHTMLFailure checks if any HTML/designer-specific rule failed.
func HasHTMLFailure(results []graph.CheckResult) bool {
	for _, r := range results {
		if !r.Passed {
			switch r.Name {
			case NameRequiredHTMLSections, NameViewportMeta, NameSingleH1:
				return true
			}
		}
	}
	return false
}
