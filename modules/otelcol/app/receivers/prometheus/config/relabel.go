package config

import (
	"strings"

	"github.com/caddyserver/caddy/v2"
)

// Config is the configuration for relabeling of target label sets.
type RelabelConfig struct {
	// A list of labels from which values are taken and concatenated
	// with the configured separator in order.
	SourceLabels string `json:"source_labels,omitempty"`
	// Separator is the string between concatenated values from the source labels.
	Separator string `json:"separator,omitempty"`
	// Regex against which the concatenation is matched.
	Regex string `json:"regex,omitempty"`
	// Modulus to take of the hash of concatenated values from the source labels.
	Modulus uint64 `json:"modulus,omitempty"`
	// TargetLabel is the label to which the resulting string is written in a replacement.
	// Regexp interpolation is allowed for the replace action.
	TargetLabel string `json:"target_label,omitempty"`
	// Replacement is the regex replacement pattern to be used.
	Replacement string `json:"replacement,omitempty"`
	// Action is the action to be performed for the relabeling.
	Action string `json:"action,omitempty"`
}

func (rc *RelabelConfig) ReplaceAll(repl *caddy.Replacer) error {
	if rc.SourceLabels != "" && rc.Separator != "" {
		labels := strings.Split(rc.SourceLabels, rc.Separator)
		for i, sourceLabel := range labels {
			sourceLabel, err := repl.ReplaceOrErr(sourceLabel, true, true)
			if err != nil {
				return err
			}
			labels[i] = sourceLabel
		}
		rc.SourceLabels = strings.Join(labels, rc.Separator)
	}
	if rc.Separator != "" {
		separator, err := repl.ReplaceOrErr(rc.Separator, true, true)
		if err != nil {
			return err
		}
		rc.Separator = separator
	}
	if rc.TargetLabel != "" {
		targetLabel, err := repl.ReplaceOrErr(rc.TargetLabel, true, true)
		if err != nil {
			return err
		}
		rc.TargetLabel = targetLabel
	}
	if rc.Replacement != "" {
		replacement, err := repl.ReplaceOrErr(rc.Replacement, true, true)
		if err != nil {
			return err
		}
		rc.Replacement = replacement
	}
	return nil
}
