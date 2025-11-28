package config

// BuiltInStages returns the hardcoded stage definitions from Iteration 1
// These are used as fallback when no config.toml is present
func BuiltInStages(repoRoot string) map[string]StageConfig {
	return map[string]StageConfig{
		"lint": {
			Name:             "Lint",
			Group:            "quality",
			Command:          repoRoot + "/hello-world.sh lint",
			Workdir:          repoRoot,
			EstimatedSeconds: 5,
			Enabled:          boolPtr(true),
		},
		"format": {
			Name:             "Format",
			Group:            "quality",
			Command:          repoRoot + "/hello-world.sh format",
			Workdir:          repoRoot,
			EstimatedSeconds: 5,
			Enabled:          boolPtr(true),
		},
		"type-check": {
			Name:             "Type Check",
			Group:            "correctness",
			Command:          repoRoot + "/hello-world.sh type-check",
			Workdir:          repoRoot,
			EstimatedSeconds: 10,
			Enabled:          boolPtr(true),
		},
		"build": {
			Name:             "Build",
			Group:            "release",
			Command:          repoRoot + "/hello-world.sh build",
			Workdir:          repoRoot,
			EstimatedSeconds: 15,
			Enabled:          boolPtr(true),
		},
		"unit-tests": {
			Name:             "Unit Tests",
			Group:            "correctness",
			Command:          repoRoot + "/hello-world.sh unit-tests",
			Workdir:          repoRoot,
			EstimatedSeconds: 20,
			Enabled:          boolPtr(true),
		},
		"e2e-tests": {
			Name:             "E2E Tests",
			Group:            "correctness",
			Command:          repoRoot + "/hello-world.sh e2e-tests",
			Workdir:          repoRoot,
			EstimatedSeconds: 600,
			Enabled:          boolPtr(true),
		},
	}
}

// GetStageOrder returns the default order for built-in stages
func GetStageOrder() []string {
	return []string{
		"lint",
		"format",
		"type-check",
		"build",
		"unit-tests",
		"e2e-tests",
	}
}
