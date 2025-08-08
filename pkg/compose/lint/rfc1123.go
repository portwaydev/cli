package lint

import (
	"regexp"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
)

var rfc1123 = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

func IsRFC1123(name string) bool {
	return rfc1123.MatchString(name)
}

func ToRFC1123(name string) string {
	// Convert to lowercase
	result := strings.ToLower(name)

	// Replace any non-alphanumeric characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	result = reg.ReplaceAllString(result, "-")

	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")

	return result
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
					Message:         "Service name must be RFC1123 compliant. If not this could cause issues with DNS resolution.",
					Suggestion:      "Recommend: " + ToRFC1123(key) + ". Lowercase letters, numbers, and hyphens. Must start and end with a letter or number.",
				})
			}
		}
		return issues
	},
}