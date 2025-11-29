package config

// BuiltInTasks returns the hardcoded task definitions from Iteration 1
// These are used as fallback when no config.toml is present
func BuiltInTasks(repoRoot string) map[string]TaskConfig {
	return map[string]TaskConfig{
		"lint": {
			Name:             "Lint",
			Type:             "quality",
			Command:          repoRoot + "/hello-world.sh lint",
			Workdir:          repoRoot,
			EstimatedSeconds: 5,
			Enabled:          boolPtr(true),
		},
		"format": {
			Name:             "Format",
			Type:             "quality",
			Command:          repoRoot + "/hello-world.sh format",
			Workdir:          repoRoot,
			EstimatedSeconds: 5,
			Enabled:          boolPtr(true),
		},
		"type-check": {
			Name:             "Type Check",
			Type:             "correctness",
			Command:          repoRoot + "/hello-world.sh type-check",
			Workdir:          repoRoot,
			EstimatedSeconds: 10,
			Enabled:          boolPtr(true),
		},
		"build": {
			Name:             "Build",
			Type:             "release",
			Command:          repoRoot + "/hello-world.sh build",
			Workdir:          repoRoot,
			EstimatedSeconds: 15,
			Enabled:          boolPtr(true),
			MetricsFormat:    "artifact",
			MetricsPath:      "artifacts/build/app.txt",
		},
		"unit-tests": {
			Name:             "Unit Tests",
			Type:             "correctness",
			Command:          repoRoot + "/hello-world.sh unit-tests",
			Workdir:          repoRoot,
			EstimatedSeconds: 20,
			Enabled:          boolPtr(true),
			MetricsFormat:    "junit",
			MetricsPath:      "artifacts/test/junit.xml",
		},
		"e2e-tests": {
			Name:             "E2E Tests",
			Type:             "correctness",
			Command:          repoRoot + "/hello-world.sh e2e-tests",
			Workdir:          repoRoot,
			EstimatedSeconds: 600,
			Enabled:          boolPtr(true),
		},
	}
}

// GetTaskOrder returns the default order for built-in tasks
func GetTaskOrder() []string {
	return []string{
		"lint",
		"format",
		"type-check",
		"build",
		"unit-tests",
		"e2e-tests",
	}
}
