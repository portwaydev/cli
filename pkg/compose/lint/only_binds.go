package lint

import (
	"github.com/compose-spec/compose-go/v2/types"
)

var OnlyBinds = ValidationCheck{
	Code: "SL002",
	Name: "Only volume types are allowed",
	Description: "Only volume types are allowed",
	Severity: SeverityError,
	CheckFunc: func(ctx *types.Project) []ValidationIssue {
		var issues []ValidationIssue
		for _, service := range ctx.Services {
			for _, volume := range service.Volumes {
				if volume.Type != "volume" {
					issues = append(issues, ValidationIssue{
						Service: service.Name,
						Field: "services." + service.Name + ".volumes",
						Message: "Only volume type is allowed, found " + volume.Type + " volume type.",
						Suggestion: "Please use a volume type, or include the content in the image.",
					})
				}
			}
		}
		return issues
	},
}