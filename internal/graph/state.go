package graph

import (
	"adstudio/internal/agent"
	"adstudio/internal/checker"
)

// FinalStatus represents the final status of the pipeline run.
type FinalStatus string

const (
	StatusApproved         FinalStatus = "approved"
	StatusNeedsHumanReview FinalStatus = "needs_human_review"
)

// Aliases for domain models.
type Brief = agent.Brief
type Copy = agent.Copy
type Review = agent.Review
type CheckResult = checker.Result

// TraceEntry records a discrete execution step in the pipeline.
type TraceEntry struct {
	Step      string      `json:"step"`
	Timestamp string      `json:"timestamp,omitempty"`
	Details   interface{} `json:"details,omitempty"`
}

// State represents the complete state flowing through the pipeline.
type State struct {
	Idea          string        `json:"idea"`
	Brief         *Brief        `json:"brief,omitempty"`
	Copy          *Copy         `json:"copy,omitempty"`
	HTML          string        `json:"html,omitempty"`
	CheckResults  []CheckResult `json:"check_results,omitempty"`
	Review        *Review       `json:"review,omitempty"`
	RevisionCount int           `json:"revision_count"`
	Status        FinalStatus   `json:"status"`
	RouteHistory  []string      `json:"route_history,omitempty"`
	Trace         []TraceEntry  `json:"trace,omitempty"`
}
