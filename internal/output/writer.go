package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"adstudio/internal/graph"
)

// TraceFile structures the persistent trace audit log written to trace.json.
type TraceFile struct {
	Idea          string              `json:"idea"`
	Status        graph.FinalStatus   `json:"status"`
	RevisionCount int                 `json:"revision_count"`
	RouteHistory  []string            `json:"route_history"`
	CheckResults  []graph.CheckResult `json:"check_results,omitempty"`
	Review        *graph.Review       `json:"review,omitempty"`
	Trace         []graph.TraceEntry  `json:"trace"`
}

// WriteOutputs saves page.html, brief.json, and trace.json to the target directory.
func WriteOutputs(dir string, state *graph.State) error {
	if state == nil {
		return fmt.Errorf("cannot write outputs for nil state")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// 1. Write page.html
	htmlPath := filepath.Join(dir, "page.html")
	if err := os.WriteFile(htmlPath, []byte(state.HTML), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", htmlPath, err)
	}

	// 2. Write brief.json
	briefPath := filepath.Join(dir, "brief.json")
	var briefBytes []byte
	var err error
	if state.Brief != nil {
		briefBytes, err = json.MarshalIndent(state.Brief, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling brief.json: %w", err)
		}
	} else {
		briefBytes = []byte("{}")
	}
	if err := os.WriteFile(briefPath, briefBytes, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", briefPath, err)
	}

	// 3. Write trace.json
	tracePayload := TraceFile{
		Idea:          state.Idea,
		Status:        state.Status,
		RevisionCount: state.RevisionCount,
		RouteHistory:  state.RouteHistory,
		CheckResults:  state.CheckResults,
		Review:        state.Review,
		Trace:         state.Trace,
	}
	traceBytes, err := json.MarshalIndent(tracePayload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling trace.json: %w", err)
	}
	tracePath := filepath.Join(dir, "trace.json")
	if err := os.WriteFile(tracePath, traceBytes, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", tracePath, err)
	}

	return nil
}
