package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"adstudio/internal/graph"
	"adstudio/internal/model"
	"adstudio/internal/output"
)

func main() {
	outDir := flag.String("out", ".", "Output directory for page.html, brief.json, and trace.json")
	liveMode := flag.Bool("live", false, "Use live model adapter (mocked/live)")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: adstudio [flags] \"<idea>\"\n\nFlags:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	idea := strings.TrimSpace(strings.Join(args, " "))
	if idea == "" {
		fmt.Fprintln(os.Stderr, "Error: idea argument cannot be empty")
		os.Exit(1)
	}

	fmt.Println("==================================================")
	fmt.Println("               Ad Studio CLI                      ")
	fmt.Println("==================================================")
	fmt.Printf("Idea: %s\n", idea)
	if *liveMode {
		fmt.Println("Model: Live/Mocked Gemini Client")
	} else {
		fmt.Println("Model: Scripted Model")
	}
	fmt.Printf("Output Directory: %s\n\n", *outDir)

	m := createModel(*liveMode, idea)
	g := graph.New()

	ctx := context.Background()
	state, err := g.Run(ctx, idea, m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error executing pipeline: %v\n", err)
		os.Exit(1)
	}

	// Print execution summary
	fmt.Println("Execution Trace:")
	for _, entry := range state.Trace {
		fmt.Printf("  • %-26s %s\n", entry.Step, entry.Timestamp)
	}

	// Persist generated files
	if err := output.WriteOutputs(*outDir, state); err != nil {
		fmt.Fprintf(os.Stderr, "\nError writing output artifacts: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--------------------------------------------------")
	fmt.Printf("Final Status:    %s\n", strings.ToUpper(string(state.Status)))
	fmt.Printf("Revision Count:  %d\n", state.RevisionCount)
	fmt.Printf("Route Traversed: %s\n", strings.Join(state.RouteHistory, " -> "))
	fmt.Println("--------------------------------------------------")
	fmt.Println("Generated Files:")
	fmt.Printf("  ✓ %s/page.html\n", *outDir)
	fmt.Printf("  ✓ %s/brief.json\n", *outDir)
	fmt.Printf("  ✓ %s/trace.json\n", *outDir)
	fmt.Println("==================================================")

	if state.Status == graph.StatusNeedsHumanReview {
		// Exit code 0 as pipeline successfully concluded with needs_human_review status
		os.Exit(0)
	}
}

func createModel(live bool, idea string) model.Model {
	// Standard scripted baseline responses for clean runs
	brief := fmt.Sprintf(`{
  "audience": "Customers seeking innovative solutions",
  "offer": "%s",
  "tone": "Warm, confident, and approachable",
  "allowed_claims": ["Custom formulated", "Ethically sourced ingredients", "Delivered directly to your door"],
  "banned_claims": ["Guaranteed medical cure", "Permanent miracle fix"],
  "allowed_banned_phrases": []
}`, idea)

	copyData := `{
  "headline": "Transform Your Routine With Smarter Daily Care",
  "subhead": "Personalized formulations created exclusively for your needs.",
  "benefits": [
    "Custom formulated for your specific profile",
    "Ethically sourced ingredients with complete transparency",
    "Delivered directly to your door on your schedule"
  ],
  "cta": "Start Your Subscription Today",
  "disclaimer": "Fictional brand for demonstration purposes.",
  "claims": [
    "Custom formulated",
    "Ethically sourced ingredients",
    "Delivered directly to your door"
  ]
}`

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Ad Studio Landing Page</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; color: #222; max-width: 720px; margin: 40px auto; padding: 0 20px; }
        h1 { color: #111; font-size: 2.2rem; }
        .subhead { font-size: 1.2rem; color: #555; }
        .benefits { margin: 24px 0; padding-left: 20px; }
        .benefits li { margin-bottom: 8px; }
        .cta-btn { display: inline-block; background: #0066cc; color: #fff; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .disclaimer { font-size: 0.85rem; color: #888; border-top: 1px solid #eee; padding-top: 12px; margin-top: 40px; }
    </style>
</head>
<body>
    <h1>Transform Your Routine With Smarter Daily Care</h1>
    <p class="subhead">Personalized formulations created exclusively for your needs.</p>
    <ul class="benefits">
        <li>Custom formulated for your specific profile</li>
        <li>Ethically sourced ingredients with complete transparency</li>
        <li>Delivered directly to your door on your schedule</li>
    </ul>
    <a href="#subscribe" class="cta-btn">Start Your Subscription Today</a>
    <p class="disclaimer">Fictional brand for demonstration purposes.</p>
</body>
</html>`

	review := `{"score": 9, "feedback": "Well-structured landing page with clear value proposition and compliant claims."}`

	return model.NewScriptedModel([]string{brief, copyData, html, review})
}
