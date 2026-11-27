package graph

import (
	"context"
	"fmt"

	"adstudio/internal/agent"
	"adstudio/internal/checker"
	"adstudio/internal/model"
)

// Runner executes the Ad Studio agent pipeline graph.
type Runner interface {
	Run(ctx context.Context, idea string, m model.Model) (*State, error)
}

// Graph implements the state machine pipeline.
type Graph struct{}

// New creates a new Graph runner.
func New() *Graph {
	return &Graph{}
}

func toCheckFailures(results []CheckResult) []agent.CheckFailure {
	var failures []agent.CheckFailure
	for _, r := range results {
		if !r.Passed {
			failures = append(failures, agent.CheckFailure{
				Name:    r.Name,
				Message: r.Message,
			})
		}
	}
	return failures
}

// Run executes the graph state machine from initial idea to approved or needs_human_review.
func (g *Graph) Run(ctx context.Context, idea string, m model.Model) (*State, error) {
	if m == nil {
		return nil, fmt.Errorf("graph runner requires a non-nil model")
	}

	state := &State{
		Idea:          idea,
		RevisionCount: 0,
		RouteHistory:  make([]string, 0),
		Trace:         make([]TraceEntry, 0),
	}
	state.LogTrace("start", map[string]string{"idea": idea})

	// 1. Strategist Node
	state.RecordRoute("strategist")
	brief, err := agent.GenerateBrief(ctx, m, state.Idea)
	if err != nil {
		// Branch 1: Strategist output malformed -> needs_human_review. No revision.
		state.LogTrace("node:strategist:malformed", map[string]string{"error": err.Error()})
		state.Status = StatusNeedsHumanReview
		state.RecordRoute(string(StatusNeedsHumanReview))
		state.LogTrace("terminal", map[string]interface{}{
			"status": state.Status,
			"reason": "strategist output was malformed or unparseable",
		})
		return state, nil
	}
	state.Brief = brief
	state.LogTrace("node:strategist:success", brief)

	// 2. Initial Draft: Copywriter -> Designer -> Checker -> Reviewer
	state.RecordRoute("copywriter")
	c, err := agent.GenerateCopy(ctx, m, state.Brief, nil)
	if err != nil {
		state.LogTrace("node:copywriter:error", map[string]string{"error": err.Error()})
		state.Copy = nil
	} else {
		state.Copy = c
		state.LogTrace("node:copywriter:success", c)
	}

	state.RecordRoute("designer")
	html, err := agent.GenerateHTML(ctx, m, state.Brief, state.Copy, nil)
	if err != nil {
		state.LogTrace("node:designer:error", map[string]string{"error": err.Error()})
		state.HTML = ""
	} else {
		state.HTML = html
		state.LogTrace("node:designer:success", map[string]interface{}{"length": len(html)})
	}

	// Checker evaluation (deterministic)
	state.RecordRoute("checker")
	state.CheckResults = checker.RunAllChecks(state.Brief, state.Copy, state.HTML)
	state.LogTrace("node:checker:results", state.CheckResults)

	// Reviewer evaluation (advisory only)
	state.RecordRoute("reviewer")
	state.Review, _ = agent.ReviewPage(ctx, m, state.Copy, state.HTML)
	state.LogTrace("node:reviewer:advisory", state.Review)

	// Router evaluation
	state.RecordRoute("router")
	decision := DecideNextRoute(state.CheckResults, state.RevisionCount)
	state.LogTrace("node:router:decision", decision)

	// Branch 2: All checks pass -> approved
	if decision.Target == RouteApproved {
		state.Status = StatusApproved
		state.RecordRoute(string(StatusApproved))
		state.LogTrace("terminal", map[string]interface{}{
			"status": state.Status,
			"reason": decision.Reason,
		})
		return state, nil
	}

	// Branch 3: Check fails and revisionCount is 0 -> exactly one revision pass allowed
	if state.RevisionCount == 0 && (decision.Target == RouteCopywriter || decision.Target == RouteDesigner) {
		state.RevisionCount = 1
		failures := toCheckFailures(decision.FailedChecks)

		if decision.Target == RouteCopywriter {
			// Copywriter revision
			state.RecordRoute("copywriter")
			revCopy, err := agent.GenerateCopy(ctx, m, state.Brief, failures)
			if err != nil {
				state.LogTrace("node:copywriter_revision:error", map[string]string{"error": err.Error()})
				state.Copy = nil
			} else {
				state.Copy = revCopy
				state.LogTrace("node:copywriter_revision:success", revCopy)
			}

			// Designer re-runs after copy revision
			state.RecordRoute("designer")
			revHTML, err := agent.GenerateHTML(ctx, m, state.Brief, state.Copy, failures)
			if err != nil {
				state.LogTrace("node:designer_revision:error", map[string]string{"error": err.Error()})
				state.HTML = ""
			} else {
				state.HTML = revHTML
				state.LogTrace("node:designer_revision:success", map[string]interface{}{"length": len(revHTML)})
			}
		} else if decision.Target == RouteDesigner {
			// Designer only revision
			state.RecordRoute("designer")
			revHTML, err := agent.GenerateHTML(ctx, m, state.Brief, state.Copy, failures)
			if err != nil {
				state.LogTrace("node:designer_revision:error", map[string]string{"error": err.Error()})
				state.HTML = ""
			} else {
				state.HTML = revHTML
				state.LogTrace("node:designer_revision:success", map[string]interface{}{"length": len(revHTML)})
			}
		}

		// Re-run checker
		state.RecordRoute("checker")
		state.CheckResults = checker.RunAllChecks(state.Brief, state.Copy, state.HTML)
		state.LogTrace("node:checker_revision:results", state.CheckResults)

		// Re-run reviewer (advisory)
		state.RecordRoute("reviewer")
		state.Review, _ = agent.ReviewPage(ctx, m, state.Copy, state.HTML)
		state.LogTrace("node:reviewer_revision:advisory", state.Review)

		// Re-run router
		state.RecordRoute("router")
		postRevDecision := DecideNextRoute(state.CheckResults, state.RevisionCount)
		state.LogTrace("node:router_revision:decision", postRevDecision)

		if postRevDecision.Target == RouteApproved {
			state.Status = StatusApproved
			state.RecordRoute(string(StatusApproved))
			state.LogTrace("terminal", map[string]interface{}{
				"status": state.Status,
				"reason": postRevDecision.Reason,
			})
			return state, nil
		}

		// Branch 4: A check fails and revision_count is 1 -> needs_human_review
		state.Status = StatusNeedsHumanReview
		state.RecordRoute(string(StatusNeedsHumanReview))
		state.LogTrace("terminal", map[string]interface{}{
			"status": state.Status,
			"reason": postRevDecision.Reason,
		})
		return state, nil
	}

	// Any other failure path terminates at needs_human_review
	state.Status = StatusNeedsHumanReview
	state.RecordRoute(string(StatusNeedsHumanReview))
	state.LogTrace("terminal", map[string]interface{}{
		"status": state.Status,
		"reason": decision.Reason,
	})
	return state, nil
}
