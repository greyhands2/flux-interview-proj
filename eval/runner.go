package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"adstudio/internal/graph"
	"adstudio/internal/model"
)

// TestCase represents a single regression test scenario.
type TestCase struct {
	ID                int      `json:"id"`
	Name              string   `json:"name"`
	Input             string   `json:"input"`
	ExpectedStatus    string   `json:"expected_status"`
	ExpectedRoute     string   `json:"expected_route"`
	WhatItCatches     string   `json:"what_it_catches"`
	ScriptedResponses []string `json:"scripted_responses"`
}

// CaseResult captures the evaluation outcome for one case.
type CaseResult struct {
	Case         TestCase
	ActualStatus string
	ActualRoute  string
	Passed       bool
	ErrorMessage string
}

// LoadCases loads all test cases from the specified directory.
func LoadCases(dir string) ([]TestCase, error) {
	targetDir := dir
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		altDir := filepath.Join("eval", dir)
		if _, err2 := os.Stat(altDir); err2 == nil {
			targetDir = altDir
		}
	}

	files, err := filepath.Glob(filepath.Join(targetDir, "case*.json"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no test case files found in %s", targetDir)
	}

	var cases []TestCase
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", file, err)
		}
		var tc TestCase
		if err := json.Unmarshal(data, &tc); err != nil {
			return nil, fmt.Errorf("unmarshaling %s: %w", file, err)
		}
		cases = append(cases, tc)
	}

	sort.Slice(cases, func(i, j int) bool {
		return cases[i].ID < cases[j].ID
	})

	return cases, nil
}

// RunCase executes a single test case using the provided graph runner.
func RunCase(ctx context.Context, runner graph.Runner, tc TestCase) CaseResult {
	scriptedModel := model.NewScriptedModel(tc.ScriptedResponses)
	state, err := runner.Run(ctx, tc.Input, scriptedModel)

	res := CaseResult{Case: tc}
	if err != nil {
		res.ActualStatus = fmt.Sprintf("ERROR: %v", err)
		res.Passed = false
		res.ErrorMessage = err.Error()
		return res
	}

	if state == nil {
		res.ActualStatus = "nil state"
		res.Passed = false
		return res
	}

	res.ActualStatus = string(state.Status)
	res.ActualRoute = strings.Join(state.RouteHistory, " -> ")

	// Validate both expected status and expected route/revision behavior
	passed, failureReason := verifyRouteAndStatus(tc, state)
	res.Passed = passed
	if !passed {
		res.ErrorMessage = failureReason
	}

	return res
}

// verifyRouteAndStatus verifies that the state matches expected status and routing constraints from docs/harness-eval.md.
func verifyRouteAndStatus(tc TestCase, state *graph.State) (bool, string) {
	if string(state.Status) != tc.ExpectedStatus {
		return false, fmt.Sprintf("expected status %s, got %s", tc.ExpectedStatus, state.Status)
	}

	switch tc.ID {
	case 1:
		// Baseline: approved, 0 revisions
		if state.RevisionCount != 0 {
			return false, fmt.Sprintf("case 1 baseline expected 0 revisions, got %d", state.RevisionCount)
		}
	case 2:
		// Banned phrase caught -> 1 copywriter revision -> approved
		if state.RevisionCount != 1 {
			return false, fmt.Sprintf("case 2 expected 1 revision, got %d", state.RevisionCount)
		}
	case 3:
		// Claim check fails -> 1 revision -> fails again -> needs_human_review
		if state.RevisionCount != 1 {
			return false, fmt.Sprintf("case 3 expected 1 revision, got %d", state.RevisionCount)
		}
	case 4:
		// Missing section caught -> 1 designer revision -> approved
		if state.RevisionCount != 1 {
			return false, fmt.Sprintf("case 4 expected 1 revision, got %d", state.RevisionCount)
		}
	case 5:
		// Strategist broken JSON -> needs_human_review, no revision
		if state.RevisionCount != 0 {
			return false, fmt.Sprintf("case 5 expected 0 revisions (no revision on strategist failure), got %d", state.RevisionCount)
		}
	case 6:
		// Headline length failure repeated -> 1 revision -> needs_human_review
		if state.RevisionCount != 1 {
			return false, fmt.Sprintf("case 6 expected 1 revision, got %d", state.RevisionCount)
		}
	}

	return true, ""
}

// FormatReport formats the list of CaseResults into a table matching docs/harness-eval.md.
func FormatReport(results []CaseResult) string {
	var sb strings.Builder
	sb.WriteString("\n| Case | Expected | Actual | Pass/Fail |\n")
	sb.WriteString("|---|---|---|---|\n")

	passedCount := 0
	for _, r := range results {
		caseLabel := fmt.Sprintf("%d", r.Case.ID)
		passLabel := "FAIL"
		if r.Passed {
			passLabel = "PASS"
			passedCount++
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			caseLabel, r.Case.ExpectedStatus, r.ActualStatus, passLabel))
	}

	sb.WriteString(fmt.Sprintf("\nScore: %d/%d passed (%.1f%%)\n",
		passedCount, len(results), float64(passedCount)/float64(len(results))*100.0))

	return sb.String()
}
