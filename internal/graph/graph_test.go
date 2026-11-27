package graph_test

import (
	"context"
	"testing"

	"adstudio/internal/checker"
	"adstudio/internal/graph"
	"adstudio/internal/model"
)

func TestDecideNextRoute(t *testing.T) {
	t.Run("all passing checks routes to approved", func(t *testing.T) {
		results := []graph.CheckResult{
			{Name: checker.NameRequiredCopySections, Passed: true},
			{Name: checker.NameViewportMeta, Passed: true},
		}
		decision := graph.DecideNextRoute(results, 0)
		if decision.Target != graph.RouteApproved {
			t.Errorf("expected RouteApproved, got %s", decision.Target)
		}
	})

	t.Run("copy failure with revision 0 routes to copywriter", func(t *testing.T) {
		results := []graph.CheckResult{
			{Name: checker.NameBannedPhrases, Passed: false, Message: "banned phrase found"},
			{Name: checker.NameViewportMeta, Passed: true},
		}
		decision := graph.DecideNextRoute(results, 0)
		if decision.Target != graph.RouteCopywriter {
			t.Errorf("expected RouteCopywriter, got %s", decision.Target)
		}
	})

	t.Run("html failure with revision 0 routes to designer", func(t *testing.T) {
		results := []graph.CheckResult{
			{Name: checker.NameBannedPhrases, Passed: true},
			{Name: checker.NameViewportMeta, Passed: false, Message: "missing viewport"},
		}
		decision := graph.DecideNextRoute(results, 0)
		if decision.Target != graph.RouteDesigner {
			t.Errorf("expected RouteDesigner, got %s", decision.Target)
		}
	})

	t.Run("dual copy and html failure with revision 0 routes to copywriter first", func(t *testing.T) {
		results := []graph.CheckResult{
			{Name: checker.NameHeadlineLength, Passed: false, Message: "headline too long"},
			{Name: checker.NameSingleH1, Passed: false, Message: "no h1"},
		}
		decision := graph.DecideNextRoute(results, 0)
		if decision.Target != graph.RouteCopywriter {
			t.Errorf("expected dual failure to prioritize RouteCopywriter, got %s", decision.Target)
		}
	})

	t.Run("any failure with revision 1 routes to needs_human_review", func(t *testing.T) {
		results := []graph.CheckResult{
			{Name: checker.NameHeadlineLength, Passed: false, Message: "headline too long"},
		}
		decision := graph.DecideNextRoute(results, 1)
		if decision.Target != graph.RouteNeedsHumanReview {
			t.Errorf("expected RouteNeedsHumanReview after revision 1, got %s", decision.Target)
		}
	})
}

func TestGraphRun(t *testing.T) {
	ctx := context.Background()

	validBriefJSON := `{"audience":"Athletes","offer":"Electrolyte water","tone":"Crisp","allowed_claims":["Zero sugar","Fast hydration"],"banned_claims":["Cures dehydration"]}`
	validCopyJSON := `{"headline":"Pure Hydration For Peak Athletic Performance","subhead":"Rapid recovery.","benefits":["Zero sugar","Fast hydration","Clean minerals"],"cta":"Order Pack","disclaimer":"Demo brand.","claims":["Zero sugar"]}`
	validHTML := `<!DOCTYPE html><html><head><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Drink</title></head><body><h1>Pure Hydration For Peak Athletic Performance</h1><p>Rapid recovery.</p><ul><li>Zero sugar</li><li>Fast hydration</li><li>Clean minerals</li></ul><a href="#order">Order Pack</a><p>Demo brand.</p></body></html>`
	validReviewJSON := `{"score":9,"feedback":"Strong copy."}`

	t.Run("branch 1: malformed strategist output leads directly to needs_human_review", func(t *testing.T) {
		m := model.NewScriptedModel([]string{`{broken strategist json...`})
		g := graph.New()

		state, err := g.Run(ctx, "idea", m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.Status != graph.StatusNeedsHumanReview {
			t.Errorf("expected StatusNeedsHumanReview, got %s", state.Status)
		}
		if state.RevisionCount != 0 {
			t.Errorf("expected 0 revisions for strategist malformed failure, got %d", state.RevisionCount)
		}
		if len(state.Trace) == 0 {
			t.Error("expected trace entries to be recorded")
		}
	})

	t.Run("branch 2: baseline clean run passes with approved and 0 revisions", func(t *testing.T) {
		m := model.NewScriptedModel([]string{
			validBriefJSON,
			validCopyJSON,
			validHTML,
			validReviewJSON,
		})
		g := graph.New()

		state, err := g.Run(ctx, "idea", m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.Status != graph.StatusApproved {
			t.Errorf("expected StatusApproved, got %s", state.Status)
		}
		if state.RevisionCount != 0 {
			t.Errorf("expected 0 revisions, got %d", state.RevisionCount)
		}
		if state.Brief == nil || state.Copy == nil || state.HTML == "" {
			t.Errorf("state artifacts incomplete: %+v", state)
		}
	})

	t.Run("branch 3: copy failure triggers revision 1 and succeeds", func(t *testing.T) {
		bannedCopyJSON := `{"headline":"Guaranteed Hydration Every Single Day","subhead":"Rapid recovery.","benefits":["Zero sugar","Fast hydration","Clean minerals"],"cta":"Order Pack","disclaimer":"Demo brand.","claims":["Zero sugar"]}`
		bannedHTML := `<!DOCTYPE html><html><head><meta name="viewport" content="width=device-width, initial-scale=1.0"></head><body><h1>Guaranteed Hydration Every Single Day</h1><p>Rapid recovery.</p><ul><li>Zero sugar</li><li>Fast hydration</li><li>Clean minerals</li></ul><a href="#order">Order Pack</a><p>Demo brand.</p></body></html>`

		m := model.NewScriptedModel([]string{
			validBriefJSON,
			bannedCopyJSON,  // pass 0 copy (contains banned word 'guaranteed')
			bannedHTML,      // pass 0 html
			validReviewJSON, // pass 0 review
			validCopyJSON,   // pass 1 copy (revised: clean)
			validHTML,       // pass 1 html (revised: clean)
			validReviewJSON, // pass 1 review
		})
		g := graph.New()

		state, err := g.Run(ctx, "idea", m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.Status != graph.StatusApproved {
			t.Errorf("expected StatusApproved after revision, got %s", state.Status)
		}
		if state.RevisionCount != 1 {
			t.Errorf("expected RevisionCount == 1, got %d", state.RevisionCount)
		}
	})

	t.Run("branch 3: html failure triggers designer revision and succeeds", func(t *testing.T) {
		// Pass 0 HTML missing CTA
		noCTAHTML := `<!DOCTYPE html><html><head><meta name="viewport" content="width=device-width, initial-scale=1.0"></head><body><h1>Pure Hydration For Peak Athletic Performance</h1><p>Rapid recovery.</p><ul><li>Zero sugar</li><li>Fast hydration</li><li>Clean minerals</li></ul><p>Demo brand.</p></body></html>`

		m := model.NewScriptedModel([]string{
			validBriefJSON,
			validCopyJSON,
			noCTAHTML,       // pass 0 html (missing CTA)
			validReviewJSON, // pass 0 review
			validHTML,       // pass 1 html (designer revision adds CTA)
			validReviewJSON, // pass 1 review
		})
		g := graph.New()

		state, err := g.Run(ctx, "idea", m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.Status != graph.StatusApproved {
			t.Errorf("expected StatusApproved after designer revision, got %s", state.Status)
		}
		if state.RevisionCount != 1 {
			t.Errorf("expected RevisionCount == 1, got %d", state.RevisionCount)
		}
	})

	t.Run("branch 4: failure repeated after revision terminates at needs_human_review", func(t *testing.T) {
		longHeadlineCopy := `{"headline":"This Is An Excessively Long Headline That Clearly Exceeds The Maximum Allowed Limit Of Twelve Words","subhead":"Recovery.","benefits":["Zero sugar","Fast hydration","Clean minerals"],"cta":"Order Pack","disclaimer":"Demo brand.","claims":["Zero sugar"]}`

		m := model.NewScriptedModel([]string{
			validBriefJSON,
			longHeadlineCopy, // pass 0 copy (>12 words)
			validHTML,
			validReviewJSON,
			longHeadlineCopy, // pass 1 copy (still >12 words)
			validHTML,
			validReviewJSON,
		})
		g := graph.New()

		state, err := g.Run(ctx, "idea", m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.Status != graph.StatusNeedsHumanReview {
			t.Errorf("expected StatusNeedsHumanReview, got %s", state.Status)
		}
		if state.RevisionCount != 1 {
			t.Errorf("expected RevisionCount == 1, got %d", state.RevisionCount)
		}
	})
}
