package annotations

import (
	"fmt"
	"regexp"
	"strings"
)

const prefix = "pgsafemigrate"

// NoLint contains the settings for the ignore syntax:
// Ignore all rules: -- pgsafemigrate:nolint
// Ignore specific rule: -- pgsafemigrate:nolint:rule-alias-1
// Ignore multiple rules: -- pgsafemigrate:nolint:rule-alias-1,rule-alias-2
type NoLint struct {
	Valid     bool
	RuleNames []string
}

func (a NoLint) ExcludesAll() bool {
	return a.Valid && len(a.RuleNames) == 0
}

var pattern = regexp.MustCompile(fmt.Sprintf(`^%s:nolint(:[\-a-z,]+)?$`, prefix))

func Parse(annotation string) NoLint {
	annotation = strings.TrimSpace(annotation)
	r := pattern.FindAllStringSubmatch(annotation, -1)

	if len(r) == 0 {
		return NoLint{}
	}

	var nolintRules []string
	if m := r[0]; len(m) > 1 {
		ruleNamesSetting := strings.TrimPrefix(m[1], ":")
		if ruleNamesSetting != "" {
			nolintRules = strings.Split(ruleNamesSetting, ",")
		}
	}
	return NoLint{
		Valid:     len(r) > 0,
		RuleNames: nolintRules,
	}
}
