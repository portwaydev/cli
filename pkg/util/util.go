package util

import "os"

func IsCI() bool {
	return os.Getenv("CI") == "true" ||
		os.Getenv("CONTINUOUS_INTEGRATION") == "true" ||
		os.Getenv("BUILD_NUMBER") != "" ||
		os.Getenv("GITHUB_ACTIONS") == "true" ||
		os.Getenv("GITLAB_CI") == "true" ||
		os.Getenv("TRAVIS") == "true" ||
		os.Getenv("CIRCLECI") == "true"
}
