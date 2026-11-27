package graph

// FinalStatus represents the final status of the pipeline run.
type FinalStatus string

const (
	StatusApproved         FinalStatus = "approved"
	StatusNeedsHumanReview FinalStatus = "needs_human_review"
)

// Brief holds the campaign strategy created by the strategist.
type Brief struct {
	Audience             string   `json:"audience"`
	Offer                string   `json:"offer"`
	Tone                 string   `json:"tone"`
	AllowedClaims        []string `json:"allowed_claims"`
	BannedClaims         []string `json:"banned_claims"`
	AllowedBannedPhrases []string `json:"allowed_banned_phrases,omitempty"`
}

// Copy holds the marketing content produced by the copywriter.
type Copy struct {
	Headline   string   `json:"headline"`
	Subhead    string   `json:"subhead"`
	Benefits   []string `json:"benefits"`
	CTA        string   `json:"cta"`
	Disclaimer string   `json:"disclaimer"`
	Claims     []string `json:"claims"`
}

// CheckResult records a deterministic check outcome.
type CheckResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
}

// Review holds advisory feedback from the reviewer agent.
type Review struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

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
