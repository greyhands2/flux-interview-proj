package graph

import (
	"adstudio/internal/checker"
)

// RouteTarget represents the next destination decided by the router.
type RouteTarget string

const (
	RouteApproved         RouteTarget = "approved"
	RouteNeedsHumanReview RouteTarget = "needs_human_review"
	RouteCopywriter       RouteTarget = "copywriter"
	RouteDesigner         RouteTarget = "designer"
)

// RouterDecision contains the route decision and rationale.
type RouterDecision struct {
	Target       RouteTarget   `json:"target"`
	Reason       string        `json:"reason"`
	FailedChecks []CheckResult `json:"failed_checks,omitempty"`
}

// DecideNextRoute inspects check results and current revision count to determine the next step.
// Hard Boundary: The LLM reviewer score is advisory only and NEVER alters routing decisions.
func DecideNextRoute(results []CheckResult, revisionCount int) RouterDecision {
	// If all deterministic checks pass -> approved
	if checker.AllPassed(results) {
		return RouterDecision{
			Target: RouteApproved,
			Reason: "all deterministic checks passed",
		}
	}

	var failed []CheckResult
	for _, r := range results {
		if !r.Passed {
			failed = append(failed, r)
		}
	}

	// If checks failed and already used the 1 revision -> needs_human_review
	if revisionCount >= 1 {
		return RouterDecision{
			Target:       RouteNeedsHumanReview,
			Reason:       "deterministic checks failed and maximum revision count (1) reached",
			FailedChecks: failed,
		}
	}

	// revisionCount == 0: Decide revision destination
	hasCopy := checker.HasCopyFailure(results)
	hasHTML := checker.HasHTMLFailure(results)

	// Dual failure priority: route to copywriter first (designer & checker re-run after)
	if hasCopy || (hasCopy && hasHTML) {
		return RouterDecision{
			Target:       RouteCopywriter,
			Reason:       "copy failure detected (or dual copy+HTML failure); routing to copywriter first",
			FailedChecks: failed,
		}
	}

	if hasHTML {
		return RouterDecision{
			Target:       RouteDesigner,
			Reason:       "HTML-specific failure detected; routing to designer",
			FailedChecks: failed,
		}
	}

	// General well-formedness or fallback
	return RouterDecision{
		Target:       RouteCopywriter,
		Reason:       "unspecified check failure; routing to copywriter for revision",
		FailedChecks: failed,
	}
}
