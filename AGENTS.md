# Ad Studio: Agent Guidelines

## 1. Project Overview
Ad Studio is a CLI tool that turns a one-line idea into a compliant one-page HTML landing page. A pipeline of specialized AI agents drafts the page, while a deterministic checker gates approvals. Everything is fictional: no real brands, no medical claims.
- **Workflow Pipeline**: Strategist → Copywriter → Designer → Checker → Reviewer (advisory) → Router.
- **Root Outputs**: `page.html` (landing page), `brief.json` (strategy brief), `trace.json` (audit log).
- **Final Status**: `approved` or `needs_human_review`.

## 2. Tech Stack & Project Structure
- **Stack**: Go (standard library preferred; no external web server, UI framework, or DB).
- **Models**: Scripted fake models for evals/unit tests; optional Gemini 3.8 Flash High for live runs.
- **Structure**:
  - `cmd/adstudio/` - CLI entrypoint.
  - `internal/agent/` - Agent implementations (strategist, copywriter, designer, reviewer).
  - `internal/checker/` - Deterministic rule checks (sections, banned phrases, claims, HTML validation).
  - `internal/graph/` - State machine, router, and trace logger.
  - `internal/model/` - Model adapter (fake scripted responses and live Gemini client).
  - `internal/output/` - Root artifact persistence (page.html, brief.json, trace.json).
  - `eval/` - 6 scripted regression eval test cases.
  - `docs/` - System specifications (`product-brief.md`, `graph.md`, `harness-eval.md`).

## 3. Build, Run, and Test Commands
- **Check (Format, Lint, Tests)**: `make check` (runs `gofmt`, `go vet`, and unit tests).
- **Eval (Regression Suite)**: `make eval` (runs the 6 deterministic eval cases with fake models).
- **Run via Make**: `make "landing page for Strand, a fictional hair care subscription"`
- **Run CLI**: `go run ./cmd/adstudio "landing page for Strand, a fictional hair care subscription"`
- **Build Binary**: `make build` (compiles binary to `bin/adstudio`).

## 4. Coding Conventions
- **Standard Go Idioms**: Format strictly with `gofmt`; vet with `go vet`.
- **Error Handling**: Return explicit errors; never panic. Malformed JSON returns an error.
- **Traceability**: Every node execution, checker verdict, reviewer score, and routing step must append to `trace.json`.
- **Separation of Concerns**: Checker logic must be 100% deterministic code independent of LLM prompts.

## 5. Hard Boundaries
- **Max 1 Revision**: The router allows at most one revision pass (`revision_count <= 1`). Never loop indefinitely.
- **Deterministic Gates Only**: The router decides solely based on checker pass/fail. The LLM reviewer's score is advisory and logged; it never overrides a failed check.
- **No Live Calls in Tests/Evals**: Evals and unit tests MUST use scripted fake responses. Zero network/live model calls.
- **Dual-Failure Priority**: If both copy and HTML fail on pass 1, route to copywriter first (designer & checker re-run after).
- **Malformed Output is a Failure**: Any unparseable or broken agent output is treated as a failed check, never an approval.
- **Scope Limits**: No database, no web server, no GUI. Writes only `page.html`, `brief.json`, and `trace.json` to repo root.

## 6. Definition of Done
- **Evals Pass**: 6/6 eval cases pass in `make eval` matching expected routes and outcomes.
- **Checks Pass**: `make check` exits 0 cleanly with zero format issues, lint warnings, or test failures.
- **Artifacts Generated**: Every run produces valid `page.html`, `brief.json`, and `trace.json`.
- **Status Integrity**: Final status is strictly `approved` or `needs_human_review`.
