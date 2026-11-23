package modal

// ExperimentalOutboundPolicy mirrors modal.ExperimentalOutboundPolicy. It only
// collects the replacements: a sandbox created with one fails with a
// not-implemented error.
type ExperimentalOutboundPolicy struct {
	replacements []ExperimentalOutboundPolicyHeaderReplacement
}

// NewExperimentalOutboundPolicy mirrors the function of the same name.
func NewExperimentalOutboundPolicy(replacements ...ExperimentalOutboundPolicyHeaderReplacement) ExperimentalOutboundPolicy {
	return ExperimentalOutboundPolicy{replacements: append([]ExperimentalOutboundPolicyHeaderReplacement{}, replacements...)}
}

// WithHeaderReplacement mirrors ExperimentalOutboundPolicy.WithHeaderReplacement.
func (p ExperimentalOutboundPolicy) WithHeaderReplacement(replacement ExperimentalOutboundPolicyHeaderReplacement) ExperimentalOutboundPolicy {
	return ExperimentalOutboundPolicy{replacements: append(append([]ExperimentalOutboundPolicyHeaderReplacement{}, p.replacements...), replacement)}
}
