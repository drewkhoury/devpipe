package config

// BuiltInTasks returns the hardcoded task definitions from Iteration 1
// These are used as fallback when no config.toml is present
func BuiltInTasks(projectRoot string) map[string]TaskConfig {
	return map[string]TaskConfig{
		"lint": {
			Name:    "Lint",
			Type:    "quality",
			Command: projectRoot + "/hello-world.sh lint",
			Workdir: projectRoot,
			Enabled: boolPtr(true),
		},
		"format": {
			Name:    "Format",
			Type:    "quality",
			Command: projectRoot + "/hello-world.sh format",
			Workdir: projectRoot,
			Enabled: boolPtr(true),
		},
		"type-check": {
			Name:    "Type Check",
			Type:    "correctness",
			Command: projectRoot + "/hello-world.sh type-check",
			Workdir: projectRoot,
			Enabled: boolPtr(true),
		},
		"build": {
			Name:       "Build",
			Type:       "release",
			Command:    projectRoot + "/hello-world.sh build",
			Workdir:    projectRoot,
			Enabled:    boolPtr(true),
			OutputType: "artifact",
			OutputPath: "artifacts/build/app.txt",
		},
		"unit-tests": {
			Name:       "Unit Tests",
			Type:       "correctness",
			Command:    projectRoot + "/hello-world.sh unit-tests",
			Workdir:    projectRoot,
			Enabled:    boolPtr(true),
			OutputType: "junit",
			OutputPath: "artifacts/test/junit.xml",
		},
		"e2e-tests": {
			Name:    "E2E Tests",
			Type:    "correctness",
			Command: projectRoot + "/hello-world.sh e2e-tests",
			Workdir: projectRoot,
			Enabled: boolPtr(true),
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
