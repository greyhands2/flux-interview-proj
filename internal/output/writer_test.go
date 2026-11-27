package output_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"adstudio/internal/graph"
	"adstudio/internal/output"
)

func TestWriteOutputs(t *testing.T) {
	t.Run("successfully writes all three root artifacts", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "adstudio-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		state := &graph.State{
			Idea: "landing page for EcoSip, reusable bamboo straws",
			Brief: &graph.Brief{
				Audience:      "Eco-conscious consumers",
				Offer:         "Reusable bamboo straw set",
				Tone:          "Green, refreshing",
				AllowedClaims: []string{"100% biodegradable"},
			},
			Copy: &graph.Copy{
				Headline: "Sip Sustainably With Natural Bamboo Straws",
			},
			HTML:          "<!DOCTYPE html><html><body><h1>Sip Sustainably</h1></body></html>",
			RevisionCount: 0,
			Status:        graph.StatusApproved,
			RouteHistory:  []string{"strategist", "copywriter", "designer", "checker", "reviewer", "router", "approved"},
			Trace: []graph.TraceEntry{
				{Step: "start", Details: "init"},
				{Step: "node:strategist:success"},
			},
		}

		if err := output.WriteOutputs(tempDir, state); err != nil {
			t.Fatalf("unexpected error writing outputs: %v", err)
		}

		// Verify page.html
		htmlContent, err := os.ReadFile(filepath.Join(tempDir, "page.html"))
		if err != nil {
			t.Errorf("error reading page.html: %v", err)
		}
		if string(htmlContent) != state.HTML {
			t.Errorf("page.html content mismatch, got: %s", string(htmlContent))
		}

		// Verify brief.json
		briefContent, err := os.ReadFile(filepath.Join(tempDir, "brief.json"))
		if err != nil {
			t.Errorf("error reading brief.json: %v", err)
		}
		var briefData graph.Brief
		if err := json.Unmarshal(briefContent, &briefData); err != nil {
			t.Errorf("brief.json invalid json: %v", err)
		}
		if briefData.Audience != "Eco-conscious consumers" {
			t.Errorf("unexpected brief audience: %s", briefData.Audience)
		}

		// Verify trace.json
		traceContent, err := os.ReadFile(filepath.Join(tempDir, "trace.json"))
		if err != nil {
			t.Errorf("error reading trace.json: %v", err)
		}
		var traceData output.TraceFile
		if err := json.Unmarshal(traceContent, &traceData); err != nil {
			t.Errorf("trace.json invalid json: %v", err)
		}
		if traceData.Status != graph.StatusApproved || len(traceData.RouteHistory) != 7 {
			t.Errorf("unexpected trace data: %+v", traceData)
		}
	})

	t.Run("handles nil brief gracefully", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "adstudio-test-nil-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		state := &graph.State{
			Idea:   "idea with failed strategist",
			Brief:  nil,
			Status: graph.StatusNeedsHumanReview,
		}

		if err := output.WriteOutputs(tempDir, state); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		briefContent, err := os.ReadFile(filepath.Join(tempDir, "brief.json"))
		if err != nil {
			t.Fatalf("reading brief.json: %v", err)
		}
		if string(briefContent) != "{}" {
			t.Errorf("expected empty json object, got: %s", string(briefContent))
		}
	})

	t.Run("fails on nil state", func(t *testing.T) {
		if err := output.WriteOutputs(".", nil); err == nil {
			t.Fatal("expected error for nil state, got nil")
		}
	})
}
