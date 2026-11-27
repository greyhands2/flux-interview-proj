# Ad Studio: Harness and Evals

Evals run on scripted fake model responses, stored inline in each case. No live model calls.

## Checks
All checks are deterministic pass/fail.

| # | Check | Passes when |
|---|---|---|
| 1 | Required sections | Headline, benefits, CTA and disclaimer all exist |
| 2 | Banned phrases | None of "guaranteed", "cure", "no side effects", "clinically proven" appear (case-insensitive), unless listed in the brief's `allowed_banned_phrases` |
| 3 | Claims | Every item in the copywriter's structured `claims` list is in the brief's `allowed_claims` |
| 4 | Headline length | Headline is 12 words or fewer |
| 5 | Viewport meta | HTML has a viewport meta tag |
| 6 | Single h1 | HTML has exactly one h1 |
| 7 | Well-formed output | Every agent's output parses. Malformed output is a fail |

The reviewer score is logged but is not a check.

## Eval cases

| # | Input | Expected route | What it catches |
|---|---|---|---|
| 1 | Clean idea, compliant output | approved | Baseline. Nothing fires, no revision |
| 2 | Idea asks for "guaranteed results" | banned phrase caught → one revision (copywriter) → approved | Banned phrase check, and that a revision can recover |
| 3 | Copywriter invents a claim not in the brief, and the revision still contains it | claim check fails → one revision → fails again → needs_human_review | Claim comparison check, and that a repeated claim failure ends in human review |
| 4 | Designer omits the CTA | missing section caught → one revision (designer) → approved | Required sections check on HTML, and routing to the designer |
| 5 | Strategist returns broken JSON | needs_human_review, no revision | Malformed output is a failure, never approval |
| 6 | A simple failure (headline over 12 words) is still present after the revision | one revision → needs_human_review | The one-revision cap, independent of failure type |

Cases 3 and 6 both end in `needs_human_review`. Case 3 tests the claim check specifically. Case 6 tests the revision cap with a different failure.

## Report format
One row per case:

| Case | Expected | Actual | Pass/Fail |
|---|---|---|---|
| 1 | approved | approved | PASS |

A case passes only if the actual final status and route match the expected ones. The run passes at 6/6.

## What 6 cases can't prove
- **Coverage.** Six scripted scenarios cover six paths. They say nothing about inputs nobody thought of.
- **Real model behavior.** Fake responses are written to trigger the checks. A live model may fail in ways the scripts never imitate.
- **Check quality.** The checks catch exactly what they were written to catch. Passing them does not mean a page is good or safe, only that it is not caught by these rules.
- **Reviewer value.** The reviewer is advisory and is never scored by these cases, so they say nothing about whether its judgments are useful.
- **Interactions.** Cases mostly isolate one failure. Combined failures, such as copy and HTML problems in the same pass, are not tested.
- **Stability.** A pass shows the logic works on these fixed scripts, not that it holds across runs, wording changes or other models.