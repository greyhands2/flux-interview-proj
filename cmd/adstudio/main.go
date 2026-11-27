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
		fmt.Println("Model: Contextual Mock Model")
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

	// Print execution trace
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
		os.Exit(0)
	}
}

func createModel(live bool, idea string) model.Model {
	if live {
		return model.NewGeminiModel(idea)
	}
	return model.NewContextualMockModel(idea)
}
