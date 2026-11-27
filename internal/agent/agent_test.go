package agent_test

import (
	"context"
	"strings"
	"testing"

	"adstudio/internal/agent"
	"adstudio/internal/model"
)

func TestSanitizer(t *testing.T) {
	t.Run("cleans markdown json fences", func(t *testing.T) {
		input := "```json\n{\"key\": \"val\"}\n```"
		expected := "{\"key\": \"val\"}"
		actual := agent.SanitizeJSON(input)
		if actual != expected {
			t.Errorf("expected %q, got %q", expected, actual)
		}
	})

	t.Run("extracts json from surrounding prose", func(t *testing.T) {
		input := "Here is the response:\n{\"status\": \"ok\"}\nHope this helps!"
		expected := "{\"status\": \"ok\"}"
		actual := agent.SanitizeJSON(input)
		if actual != expected {
			t.Errorf("expected %q, got %q", expected, actual)
		}
	})

	t.Run("cleans markdown html fences", func(t *testing.T) {
		input := "```html\n<!DOCTYPE html><html><body><h1>Title</h1></body></html>\n```"
		expected := "<!DOCTYPE html><html><body><h1>Title</h1></body></html>"
		actual := agent.SanitizeHTML(input)
		if actual != expected {
			t.Errorf("expected %q, got %q", expected, actual)
		}
	})
}

func TestStrategist(t *testing.T) {
	ctx := context.Background()

	t.Run("valid json parses into brief", func(t *testing.T) {
		resp := `{"audience":"Busy runners","offer":"Custom socks","tone":"Warm","allowed_claims":["Breathable mesh"],"banned_claims":["Cures pain"]}`
		m := model.NewScriptedModel([]string{resp})

		brief, err := agent.GenerateBrief(ctx, m, "landing page for running socks")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if brief.Audience != "Busy runners" || brief.Offer != "Custom socks" {
			t.Errorf("unexpected brief values: %+v", brief)
		}
	})

	t.Run("markdown wrapped json parses cleanly", func(t *testing.T) {
		resp := "```json\n{\"audience\":\"Tea lovers\",\"offer\":\"Herbal tea\",\"tone\":\"Calm\",\"allowed_claims\":[\"Herbal\"]}\n```"
		m := model.NewScriptedModel([]string{resp})

		brief, err := agent.GenerateBrief(ctx, m, "landing page for herbal tea")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if brief.Audience != "Tea lovers" {
			t.Errorf("unexpected audience: %s", brief.Audience)
		}
	})

	t.Run("malformed json returns error", func(t *testing.T) {
		resp := `{"audience": "incomplete json...`
		m := model.NewScriptedModel([]string{resp})

		_, err := agent.GenerateBrief(ctx, m, "landing page for broken json")
		if err == nil {
			t.Fatal("expected error for malformed json, got nil")
		}
	})

	t.Run("missing required fields returns error", func(t *testing.T) {
		resp := `{"audience":"","offer":""}`
		m := model.NewScriptedModel([]string{resp})

		_, err := agent.GenerateBrief(ctx, m, "landing page with empty fields")
		if err == nil {
			t.Fatal("expected error for empty audience/offer, got nil")
		}
	})
}

func TestCopywriter(t *testing.T) {
	ctx := context.Background()
	brief := &agent.Brief{
		Audience:      "Dog owners",
		Offer:         "Organic pet food subscription",
		Tone:          "Friendly",
		AllowedClaims: []string{"Grain free", "High protein"},
		BannedClaims:  []string{"Cures illness"},
	}

	t.Run("valid json generates copy", func(t *testing.T) {
		resp := `{"headline":"Wholesome Meals For Happy Pups","subhead":"Delivered monthly.","benefits":["Grain free","High protein","Vet approved"],"cta":"Order Now","disclaimer":"Fictional brand.","claims":["Grain free"]}`
		m := model.NewScriptedModel([]string{resp})

		c, err := agent.GenerateCopy(ctx, m, brief, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Headline != "Wholesome Meals For Happy Pups" || len(c.Benefits) != 3 {
			t.Errorf("unexpected copy values: %+v", c)
		}
	})

	t.Run("nil brief returns error", func(t *testing.T) {
		m := model.NewScriptedModel([]string{"{}"})
		_, err := agent.GenerateCopy(ctx, m, nil, nil)
		if err == nil {
			t.Fatal("expected error for nil brief, got nil")
		}
	})

	t.Run("revision prompt includes failed checks", func(t *testing.T) {
		resp := `{"headline":"Revised Headline","subhead":"Better copy.","benefits":["B1","B2","B3"],"cta":"Buy","disclaimer":"Demo","claims":["Grain free"]}`
		m := &inspectingModel{response: resp}

		failedChecks := []agent.CheckFailure{
			{Name: "banned_phrases", Message: "banned phrase detected: 'guaranteed'"},
		}
		_, err := agent.GenerateCopy(ctx, m, brief, failedChecks)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(m.lastPrompt, "REVISION REQUIRED") || !strings.Contains(m.lastPrompt, "banned phrase detected") {
			t.Errorf("prompt missing revision details: %s", m.lastPrompt)
		}
	})

	t.Run("malformed json returns error", func(t *testing.T) {
		m := model.NewScriptedModel([]string{`{not valid json}`})
		_, err := agent.GenerateCopy(ctx, m, brief, nil)
		if err == nil {
			t.Fatal("expected error for malformed json, got nil")
		}
	})
}

func TestDesigner(t *testing.T) {
	ctx := context.Background()
	brief := &agent.Brief{Audience: "Coffee lovers", Offer: "Fresh roast subscription"}
	copyData := &agent.Copy{
		Headline:   "Fresh Coffee Weekly",
		Subhead:    "Direct from micro-lots",
		Benefits:   []string{"Whole bean", "Fair trade", "Weekly roast"},
		CTA:        "Start Tasting",
		Disclaimer: "Demo brand",
	}

	t.Run("valid html output generates properly", func(t *testing.T) {
		html := `<!DOCTYPE html><html><head><meta name="viewport" content="width=device-width"></head><body><h1>Fresh Coffee Weekly</h1></body></html>`
		m := model.NewScriptedModel([]string{html})

		out, err := agent.GenerateHTML(ctx, m, brief, copyData, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "<h1>Fresh Coffee Weekly</h1>") {
			t.Errorf("unexpected html output: %s", out)
		}
	})

	t.Run("nil copy returns error", func(t *testing.T) {
		m := model.NewScriptedModel([]string{"<html></html>"})
		_, err := agent.GenerateHTML(ctx, m, brief, nil, nil)
		if err == nil {
			t.Fatal("expected error for nil copy, got nil")
		}
	})

	t.Run("revision prompt includes failed checks", func(t *testing.T) {
		html := `<html><body>Revised</body></html>`
		m := &inspectingModel{response: html}
		failedChecks := []agent.CheckFailure{
			{Name: "viewport_meta", Message: "HTML missing viewport meta tag"},
		}

		_, err := agent.GenerateHTML(ctx, m, brief, copyData, failedChecks)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(m.lastPrompt, "REVISION REQUIRED") || !strings.Contains(m.lastPrompt, "viewport_meta") {
			t.Errorf("prompt missing revision details: %s", m.lastPrompt)
		}
	})

	t.Run("empty html output returns error", func(t *testing.T) {
		m := model.NewScriptedModel([]string{"   "})
		_, err := agent.GenerateHTML(ctx, m, brief, copyData, nil)
		if err == nil {
			t.Fatal("expected error for empty html, got nil")
		}
	})
}

func TestReviewer(t *testing.T) {
	ctx := context.Background()
	copyData := &agent.Copy{Headline: "Test Headline", Subhead: "Test Subhead", CTA: "Test CTA"}
	html := `<html><body><h1>Test Headline</h1></body></html>`

	t.Run("valid review json parses cleanly", func(t *testing.T) {
		resp := `{"score": 9, "feedback": "Great clarity and compelling CTA."}`
		m := model.NewScriptedModel([]string{resp})

		rev, err := agent.ReviewPage(ctx, m, copyData, html)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rev.Score != 9 || !strings.Contains(rev.Feedback, "Great clarity") {
			t.Errorf("unexpected review: %+v", rev)
		}
	})

	t.Run("malformed reviewer json produces graceful advisory fallback", func(t *testing.T) {
		resp := `Looks great overall but headline could be punchier.`
		m := model.NewScriptedModel([]string{resp})

		rev, err := agent.ReviewPage(ctx, m, copyData, html)
		if err != nil {
			t.Fatalf("reviewer should not hard error on informal text: %v", err)
		}
		if rev.Score != 5 || !strings.Contains(rev.Feedback, "Looks great overall") {
			t.Errorf("expected fallback review score 5 and captured text, got %+v", rev)
		}
	})
}

// inspectingModel records the last prompt received for verification.
type inspectingModel struct {
	response   string
	lastPrompt string
}

func (m *inspectingModel) Generate(ctx context.Context, prompt string) (string, error) {
	m.lastPrompt = prompt
	return m.response, nil
}
