package agent

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

// Review holds advisory feedback from the reviewer agent.
type Review struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

// CheckFailure records a check that failed and needs revision.
type CheckFailure struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}
