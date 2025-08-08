package lint

import (
	"regexp"

	"github.com/compose-spec/compose-go/v2/types"
)

var rfc1123 = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

func IsRFC1123(name string) bool {
	return rfc1123.MatchString(name)
}

var ServiceKeyRFC1123 = ValidationCheck{
	Code:        "SL001",
	Name:        "Service key must be RFC1123 compliant",
	Description: "Service key must follow the RFC1123 pattern: lowercase letters, numbers, and hyphens. Must start and end with a letter or number.",
	Severity:    SeverityWarning,
	Category:    "lint",
	CheckFunc:   func(ctx *types.Project) []ValidationIssue {
		var issues []ValidationIssue
		for key, service := range ctx.Services {
			if !IsRFC1123(key) {
				issues = append(issues, ValidationIssue{
					Service:         service.Name,
					Field:           "services." + key,
					Message:         "Service name must be RFC1123 compliant",
					Suggestion:      "Use a name that matches the RFC1123 pattern: lowercase letters, numbers, and hyphens. Must start and end with a letter or number.",
				})
			}
		}
		return issues
	},
}