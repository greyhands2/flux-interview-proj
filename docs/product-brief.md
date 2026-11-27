# Ad Studio: Product Brief

## Problem
Turning a one-line idea into a landing page is slow, and AI-generated marketing copy tends to make claims nobody approved. Ad Studio is a command-line tool where a few AI agents build the page and a deterministic checker decides whether it is safe to approve.

Everything is fictional: no real brands, no medical claims.

## Input
One line of text, for example:

`landing page for Strand, a fictional hair care subscription`

## Output
Each run writes three files within the project root and ends with one final status.

| File | Contents |
|---|---|
| `page.html` | The one-page HTML landing page |
| `brief.json` | The strategist's brief (audience, offer, tone, allowed claims, banned claims) |
| `trace.json` | Every step, check result, reviewer score and routing decision |

Final status is either `approved` or `needs_human_review`.

## Scope
**In**
- The agent steps: strategist, copywriter, designer, checker, reviewer
- A router with exactly one revision
- The deterministic checks
- The 6 eval cases
- The trace file
- A fake model for tests
- An option to use a real model (Gemini 3.8 Flash High)

**Out (non-goals)**
- No database
- No web server
- No UI
- No extra frameworks

## Principles
- The router goes only by the deterministic checks.
- The LLM reviewer's score is advisory. It is logged and never overrides a failed check.
- Malformed or unparseable agent output is a failure, never an approval.
- Evals run on scripted fake model responses. No live model calls in tests.

## Success criteria
- 6/6 eval cases pass.
- Every run produces a readable trace.

## Decisions
- Language: Go
- Real model option: Gemini 3.8 Flash High (exact API model ID: TBD)
- CLI command name: TBD