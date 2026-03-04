package policy

import (
	"fmt"
	"strings"
)

type Request struct {
	Kind      string
	Tool      string
	Operation string
	Path      string // used by demo rules for filesystem writes
}

type MatchedRule struct {
	RuleID      string
	Effect      string
	Explanation string
}

type Decision struct {
	PolicyID     string
	PolicyHash   string
	PolicySource string

	Decision string // "allow" or "deny"
	Reason   string
	Matched  []MatchedRule

	// optional: raw context strings for hashing upstream
	ContextKV map[string]string
}

func (p *Policy) Evaluate(req Request) Decision {
	req.Kind = strings.ToLower(strings.TrimSpace(req.Kind))
	req.Tool = strings.ToLower(strings.TrimSpace(req.Tool))
	req.Operation = strings.ToLower(strings.TrimSpace(req.Operation))
	req.Path = strings.TrimSpace(req.Path)

	for _, r := range p.Rules {
		if !ruleMatches(r, req) {
			continue
		}

		eff := r.Effect
		reason := r.Explanation
		if strings.TrimSpace(reason) == "" {
			reason = fmt.Sprintf("Matched rule %s.", r.RuleID)
		}

		return Decision{
			PolicyID:     p.PolicyID,
			PolicyHash:   p.PolicyHash,
			PolicySource: p.SourcePath,

			Decision: eff,
			Reason:   reason,
			Matched: []MatchedRule{
				{RuleID: r.RuleID, Effect: eff, Explanation: reason},
			},
			ContextKV: map[string]string{
				"requested_path": req.Path,
			},
		}
	}

	// No matches -> default
	return Decision{
		PolicyID:     p.PolicyID,
		PolicyHash:   p.PolicyHash,
		PolicySource: p.SourcePath,

		Decision: p.DefaultEffect,
		Reason:   fmt.Sprintf("No rule matched; default_effect=%s.", p.DefaultEffect),
		Matched:  []MatchedRule{},
		ContextKV: map[string]string{
			"requested_path": req.Path,
		},
	}
}

func ruleMatches(r Rule, req Request) bool {
	if r.Kind != "" && strings.ToLower(strings.TrimSpace(r.Kind)) != req.Kind {
		return false
	}
	if r.Tool != "" && strings.ToLower(strings.TrimSpace(r.Tool)) != req.Tool {
		return false
	}
	if r.Operation != "" && strings.ToLower(strings.TrimSpace(r.Operation)) != req.Operation {
		return false
	}

	// Path constraints (demo)
	if r.PathExact != "" && req.Path != strings.TrimSpace(r.PathExact) {
		return false
	}
	if r.PathPrefix != "" && !strings.HasPrefix(req.Path, strings.TrimSpace(r.PathPrefix)) {
		return false
	}

	return true
}
