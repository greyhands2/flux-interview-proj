package eval

import (
	"context"
	"fmt"
	"testing"

	"adstudio/internal/graph"
)

func TestEval(t *testing.T) {
	cases, err := LoadCases("testdata")
	if err != nil {
		t.Fatalf("failed to load eval cases: %v", err)
	}

	runner := graph.New()
	results := make([]CaseResult, 0, len(cases))
	allPassed := true

	for _, tc := range cases {
		res := RunCase(context.Background(), runner, tc)
		results = append(results, res)
		if !res.Passed {
			allPassed = false
		}
	}

	report := FormatReport(results)
	fmt.Print(report)

	if !allPassed {
		t.Errorf("Eval suite: 6/6 passing required by DoD (got %d/%d)", 0, len(results))
	}
}
