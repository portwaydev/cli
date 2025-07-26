package validate

var checkPorts = ValidationCheck{
	Code:        "PW003",
	Name:        "Missing Port Configuration",
	Description: "Check for services without port or expose configuration",
	Severity:    "WARNING",
	Category:    "networking",
	CheckFunc: func(ctx ValidationContext) []ValidationIssue {
		var issues []ValidationIssue
		for _, service := range ctx.Project.Services {
			// Skip check if restart is set to "no"
			if service.Restart == "no" {
				continue
			}

			if len(service.Ports) == 0 && len(service.Expose) == 0 {
				issues = append(issues, ValidationIssue{
					Service:    service.Name,
					Field:      "services." + service.Name + ".ports",
					Message:    "Service does not expose any ports",
					Suggestion: "Add ports or expose configuration to make the service accessible. If this service is not expected to expose ports, you can safely ignore this warning.",
				})
			}
		}
		return issues
	},
}

func init() {
	checkRegistry.Register(checkPorts)
}
