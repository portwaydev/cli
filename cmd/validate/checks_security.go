package validate

import "strings"

var checkPrivilegedMode = ValidationCheck{
	Code:        "PW001",
	Name:        "Privileged Mode",
	Description: "Check for service privileged mode",
	Severity:    "WARNING",
	Category:    "security",
	CheckFunc: func(ctx ValidationContext) []ValidationIssue {
		var issues []ValidationIssue
		for _, service := range ctx.Project.Services {
			if service.Privileged {
				issues = append(issues, ValidationIssue{
					Service:    service.Name,
					Field:      "services." + service.Name + ".privileged",
					Message:    "Privileged mode is not supported",
					Suggestion: "Consider using a non-privileged mode",
				})
			}
		}
		return issues
	},
}

var checkCapabilities = ValidationCheck{
	Code:        "PW002",
	Name:        "Capabilities",
	Description: "Check for dangerous Linux capabilities",
	Severity:    "WARNING",
	Category:    "security",
	CheckFunc: func(ctx ValidationContext) []ValidationIssue {
		var issues []ValidationIssue
		dangerousCaps := map[string]bool{
			"SYS_ADMIN":  true,
			"NET_ADMIN":  true,
			"ALL":        true,
			"SYS_MODULE": true,
			"SYS_RAWIO":  true,
			"SYS_PTRACE": true,
		}

		for _, service := range ctx.Project.Services {
			var dangerousCapsFound []string
			for _, cap := range service.CapAdd {
				if dangerousCaps[cap] {
					dangerousCapsFound = append(dangerousCapsFound, cap)
				}
			}
			if len(dangerousCapsFound) > 0 {
				issues = append(issues, ValidationIssue{
					Service:    service.Name,
					Field:      "services." + service.Name + ".cap_add",
					Message:    "These capabilities are not supported by Portway: " + strings.Join(dangerousCapsFound, ", "),
					Suggestion: "Please remove them or use a more restrictive alternative.",
				})
			}
		}
		return issues
	},
}

func init() {
	checkRegistry.Register(checkPrivilegedMode)
	checkRegistry.Register(checkCapabilities)
}
