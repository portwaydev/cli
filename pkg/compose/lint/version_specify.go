package lint

import (
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
)

var VersionSpecify = ValidationCheck{
	Code: "SL003",
	Name: "Version is not specified",
	Description: "Version is not specified",
	Severity: SeverityWarning,
	CheckFunc: func(ctx *types.Project) []ValidationIssue {
		var issues []ValidationIssue
		for _, service := range ctx.Services {
			if service.Image != "" && service.Build == nil {
				image := strings.Split(service.Image, ":")
				if len(image) == 1 {
					issues = append(issues, ValidationIssue{
						Service: service.Name,
						Field: "services." + service.Name + ".image",
						Message: "Version is not specified",
						Suggestion: "Please specify the version",
					})
				}
			}
		}
		return issues
	},
}