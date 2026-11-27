package checker_test

import (
	"testing"

	"adstudio/internal/agent"
	"adstudio/internal/checker"
)

func validTestBrief() *agent.Brief {
	return &agent.Brief{
		Audience: "Active runners",
		Offer:    "Ergonomic running socks",
		Tone:     "Energetic",
		AllowedClaims: []string{
			"Moisture-wicking fabric",
			"Arch support design",
			"Blister resistant",
		},
		BannedClaims: []string{
			"Guaranteed marathon personal record",
		},
		AllowedBannedPhrases: nil,
	}
}

func validTestCopy() *agent.Copy {
	return &agent.Copy{
		Headline: "Run Further With Cloud Comfort Everyday",
		Subhead:  "Ergonomic socks engineered for daily long runs.",
		Benefits: []string{
			"Moisture-wicking fabric",
			"Arch support design",
			"Blister resistant",
		},
		CTA:        "Get Your Pair Today",
		Disclaimer: "Fictional brand. For demo purposes.",
		Claims: []string{
			"Moisture-wicking fabric",
			"Arch support design",
		},
	}
}

func validTestHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SpeedSocks</title>
</head>
<body>
    <h1>Run Further With Cloud Comfort Everyday</h1>
    <p>Ergonomic socks engineered for daily long runs.</p>
    <ul>
        <li>Moisture-wicking fabric</li>
        <li>Arch support design</li>
        <li>Blister resistant</li>
    </ul>
    <a href="#buy" class="cta-btn">Get Your Pair Today</a>
    <p class="disclaimer">Fictional brand. For demo purposes.</p>
</body>
</html>`
}

func TestCheckRequiredCopySections(t *testing.T) {
	t.Run("valid copy passes", func(t *testing.T) {
		res := checker.CheckRequiredCopySections(validTestCopy())
		if !res.Passed {
			t.Errorf("expected pass, got fail: %s", res.Message)
		}
	})

	t.Run("nil copy fails", func(t *testing.T) {
		res := checker.CheckRequiredCopySections(nil)
		if res.Passed {
			t.Error("expected fail for nil copy, got pass")
		}
	})

	t.Run("missing headline fails", func(t *testing.T) {
		c := validTestCopy()
		c.Headline = "   "
		res := checker.CheckRequiredCopySections(c)
		if res.Passed {
			t.Error("expected fail for missing headline, got pass")
		}
	})

	t.Run("fewer than 3 benefits fails", func(t *testing.T) {
		c := validTestCopy()
		c.Benefits = []string{"Benefit 1", "Benefit 2"}
		res := checker.CheckRequiredCopySections(c)
		if res.Passed {
			t.Error("expected fail for < 3 benefits, got pass")
		}
	})

	t.Run("empty benefit in list fails", func(t *testing.T) {
		c := validTestCopy()
		c.Benefits = []string{"Benefit 1", "", "Benefit 3"}
		res := checker.CheckRequiredCopySections(c)
		if res.Passed {
			t.Error("expected fail for empty benefit item, got pass")
		}
	})

	t.Run("missing cta fails", func(t *testing.T) {
		c := validTestCopy()
		c.CTA = ""
		res := checker.CheckRequiredCopySections(c)
		if res.Passed {
			t.Error("expected fail for missing CTA, got pass")
		}
	})

	t.Run("missing disclaimer fails", func(t *testing.T) {
		c := validTestCopy()
		c.Disclaimer = ""
		res := checker.CheckRequiredCopySections(c)
		if res.Passed {
			t.Error("expected fail for missing disclaimer, got pass")
		}
	})
}

func TestCheckRequiredHTMLSections(t *testing.T) {
	c := validTestCopy()

	t.Run("valid HTML passes", func(t *testing.T) {
		res := checker.CheckRequiredHTMLSections(c, validTestHTML())
		if !res.Passed {
			t.Errorf("expected pass, got fail: %s", res.Message)
		}
	})

	t.Run("missing cta in HTML fails", func(t *testing.T) {
		html := `<!DOCTYPE html><html><head><meta name="viewport" content="width=device-width"></head><body><h1>Run Further With Cloud Comfort Everyday</h1><ul><li>Moisture-wicking fabric</li></ul><p>Fictional brand. For demo purposes.</p></body></html>`
		res := checker.CheckRequiredHTMLSections(c, html)
		if res.Passed {
			t.Error("expected fail for missing CTA text, got pass")
		}
	})

	t.Run("missing headline in HTML fails", func(t *testing.T) {
		html := `<!DOCTYPE html><html><head><meta name="viewport" content="width=device-width"></head><body><h1>Different Title</h1><a href="#">Get Your Pair Today</a><ul><li>Moisture-wicking fabric</li></ul><p>Fictional brand. For demo purposes.</p></body></html>`
		res := checker.CheckRequiredHTMLSections(c, html)
		if res.Passed {
			t.Error("expected fail for missing headline text, got pass")
		}
	})

	t.Run("empty HTML fails", func(t *testing.T) {
		res := checker.CheckRequiredHTMLSections(c, "")
		if res.Passed {
			t.Error("expected fail for empty HTML, got pass")
		}
	})
}

func TestCheckBannedPhrases(t *testing.T) {
	b := validTestBrief()
	c := validTestCopy()
	html := validTestHTML()

	t.Run("clean text passes", func(t *testing.T) {
		res := checker.CheckBannedPhrases(b, c, html)
		if !res.Passed {
			t.Errorf("expected pass, got fail: %s", res.Message)
		}
	})

	t.Run("banned phrase 'guaranteed' in headline fails", func(t *testing.T) {
		badCopy := validTestCopy()
		badCopy.Headline = "Guaranteed Best Running Comfort"
		res := checker.CheckBannedPhrases(b, badCopy, html)
		if res.Passed {
			t.Error("expected fail for 'guaranteed' in headline, got pass")
		}
	})

	t.Run("banned phrase 'cure' in benefits fails", func(t *testing.T) {
		badCopy := validTestCopy()
		badCopy.Benefits[0] = "Can cure blister pains immediately"
		res := checker.CheckBannedPhrases(b, badCopy, html)
		if res.Passed {
			t.Error("expected fail for 'cure' in benefits, got pass")
		}
	})

	t.Run("banned phrase 'clinically proven' in HTML fails", func(t *testing.T) {
		badHTML := html + "<p>Clinically proven comfort</p>"
		res := checker.CheckBannedPhrases(b, c, badHTML)
		if res.Passed {
			t.Error("expected fail for 'clinically proven' in HTML, got pass")
		}
	})

	t.Run("banned phrase explicitly allowed in brief passes", func(t *testing.T) {
		briefWithAllowed := validTestBrief()
		briefWithAllowed.AllowedBannedPhrases = []string{"guaranteed"}
		badCopy := validTestCopy()
		badCopy.Headline = "Guaranteed Best Running Comfort"
		res := checker.CheckBannedPhrases(briefWithAllowed, badCopy, html)
		if !res.Passed {
			t.Errorf("expected pass when phrase is allowed, got fail: %s", res.Message)
		}
	})
}

func TestCheckClaims(t *testing.T) {
	b := validTestBrief()

	t.Run("allowed claims pass", func(t *testing.T) {
		c := validTestCopy()
		c.Claims = []string{"Moisture-wicking fabric", "Arch support design"}
		res := checker.CheckClaims(b, c)
		if !res.Passed {
			t.Errorf("expected pass, got fail: %s", res.Message)
		}
	})

	t.Run("unallowed claim fails", func(t *testing.T) {
		c := validTestCopy()
		c.Claims = []string{"Moisture-wicking fabric", "Prevents all ankle sprains"}
		res := checker.CheckClaims(b, c)
		if res.Passed {
			t.Error("expected fail for unallowed claim, got pass")
		}
	})

	t.Run("case-insensitive and trimmed whitespace claim passes", func(t *testing.T) {
		c := validTestCopy()
		c.Claims = []string{"  moisture-wicking fabric  "}
		res := checker.CheckClaims(b, c)
		if !res.Passed {
			t.Errorf("expected pass with normalization, got fail: %s", res.Message)
		}
	})
}

func TestCheckHeadlineLength(t *testing.T) {
	t.Run("headline with 7 words passes", func(t *testing.T) {
		c := validTestCopy()
		c.Headline = "Run Further With Cloud Comfort Everyday"
		res := checker.CheckHeadlineLength(c)
		if !res.Passed {
			t.Errorf("expected pass, got: %s", res.Message)
		}
	})

	t.Run("headline with exactly 12 words passes", func(t *testing.T) {
		c := validTestCopy()
		c.Headline = "One Two Three Four Five Six Seven Eight Nine Ten Eleven Twelve"
		res := checker.CheckHeadlineLength(c)
		if !res.Passed {
			t.Errorf("expected pass for 12 words, got: %s", res.Message)
		}
	})

	t.Run("headline with 13 words fails", func(t *testing.T) {
		c := validTestCopy()
		c.Headline = "One Two Three Four Five Six Seven Eight Nine Ten Eleven Twelve Thirteen"
		res := checker.CheckHeadlineLength(c)
		if res.Passed {
			t.Error("expected fail for 13 words, got pass")
		}
	})
}

func TestCheckViewportMeta(t *testing.T) {
	t.Run("standard viewport passes", func(t *testing.T) {
		res := checker.CheckViewportMeta(`<meta name="viewport" content="width=device-width, initial-scale=1.0">`)
		if !res.Passed {
			t.Errorf("expected pass, got: %s", res.Message)
		}
	})

	t.Run("missing viewport fails", func(t *testing.T) {
		res := checker.CheckViewportMeta(`<html><head><title>No Viewport</title></head></html>`)
		if res.Passed {
			t.Error("expected fail for missing viewport, got pass")
		}
	})
}

func TestCheckSingleH1(t *testing.T) {
	t.Run("exactly one h1 passes", func(t *testing.T) {
		res := checker.CheckSingleH1(`<body><h1>Headline</h1><p>Body</p></body>`)
		if !res.Passed {
			t.Errorf("expected pass, got: %s", res.Message)
		}
	})

	t.Run("missing h1 fails", func(t *testing.T) {
		res := checker.CheckSingleH1(`<body><h2>Only H2</h2></body>`)
		if res.Passed {
			t.Error("expected fail for missing h1, got pass")
		}
	})

	t.Run("multiple h1s fails", func(t *testing.T) {
		res := checker.CheckSingleH1(`<body><h1>Headline 1</h1><h1>Headline 2</h1></body>`)
		if res.Passed {
			t.Error("expected fail for multiple h1s, got pass")
		}
	})
}

func TestCheckWellFormed(t *testing.T) {
	b := validTestBrief()
	c := validTestCopy()
	html := validTestHTML()

	t.Run("valid inputs pass", func(t *testing.T) {
		res := checker.CheckWellFormed(b, c, html)
		if !res.Passed {
			t.Errorf("expected pass, got: %s", res.Message)
		}
	})

	t.Run("nil brief fails", func(t *testing.T) {
		res := checker.CheckWellFormed(nil, c, html)
		if res.Passed {
			t.Error("expected fail for nil brief, got pass")
		}
	})

	t.Run("nil copy fails", func(t *testing.T) {
		res := checker.CheckWellFormed(b, nil, html)
		if res.Passed {
			t.Error("expected fail for nil copy, got pass")
		}
	})

	t.Run("empty html fails", func(t *testing.T) {
		res := checker.CheckWellFormed(b, c, "  ")
		if res.Passed {
			t.Error("expected fail for empty html, got pass")
		}
	})
}

func TestRunAllChecksAndClassification(t *testing.T) {
	b := validTestBrief()
	c := validTestCopy()
	html := validTestHTML()

	t.Run("all valid passes completely", func(t *testing.T) {
		results := checker.RunAllChecks(b, c, html)
		if !checker.AllPassed(results) {
			t.Errorf("expected all passed, got failures: %+v", results)
		}
		if checker.HasCopyFailure(results) {
			t.Error("expected no copy failure")
		}
		if checker.HasHTMLFailure(results) {
			t.Error("expected no HTML failure")
		}
	})

	t.Run("copy failure detected properly", func(t *testing.T) {
		badCopy := validTestCopy()
		badCopy.Claims = []string{"Unapproved Miracle Claim"} // only triggers CheckClaims
		results := checker.RunAllChecks(b, badCopy, html)
		if checker.AllPassed(results) {
			t.Error("expected fail, got pass")
		}
		if !checker.HasCopyFailure(results) {
			t.Error("expected HasCopyFailure to be true")
		}
		if checker.HasHTMLFailure(results) {
			t.Error("expected HasHTMLFailure to be false")
		}
	})

	t.Run("html failure detected properly", func(t *testing.T) {
		badHTML := `<html><body><h1>Only Title</h1></body></html>` // missing viewport, missing sections
		results := checker.RunAllChecks(b, c, badHTML)
		if checker.AllPassed(results) {
			t.Error("expected fail, got pass")
		}
		if !checker.HasHTMLFailure(results) {
			t.Error("expected HasHTMLFailure to be true")
		}
	})

	t.Run("dual failure detected properly", func(t *testing.T) {
		badCopy := validTestCopy()
		badCopy.Headline = "This Headline Exceeds The Maximum Limit Of Twelve Words Per Headline Style Check" // 13 words
		badHTML := `<html><body>No H1 Here</body></html>`                                                     // no h1, no viewport
		results := checker.RunAllChecks(b, badCopy, badHTML)
		if !checker.HasCopyFailure(results) {
			t.Error("expected copy failure in dual failure")
		}
		if !checker.HasHTMLFailure(results) {
			t.Error("expected html failure in dual failure")
		}
	})
}
