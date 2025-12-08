package dashboard

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/acarl005/stripansi"

	"github.com/drew/devpipe/internal/model"
)

// writeHTMLDashboard generates the HTML report
func writeHTMLDashboard(path string, summary Summary) error {
	type DashboardData struct {
		Summary
		Timezone string
	}

	data := DashboardData{
		Summary:  summary,
		Timezone: getLocalTimezone(),
	}

	tmpl, err := template.New("dashboard").Funcs(template.FuncMap{
		"formatDuration": formatDuration,
		"formatTime":     formatTime,
		"statusClass":    statusClass,
		"statusSymbol":   statusSymbol,
		"shortRunID":     shortRunID,
		"truncate":       truncateString,
		"phaseEmoji":     phaseEmoji,
		"float64":        func(i int) float64 { return float64(i) },
		"mul":            func(a, b float64) float64 { return a * b },
		"div":            func(a, b float64) float64 { return a / b },
		"int64":          func(f float64) int64 { return int64(f) },
	}).Parse(dashboardTemplate)

	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", cerr)
		}
	}()

	return tmpl.Execute(f, data)
}

func formatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	seconds := float64(ms) / 1000.0
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", seconds)
	}
	minutes := int(seconds / 60)
	secs := int(seconds) % 60
	return fmt.Sprintf("%dm %ds", minutes, secs)
}

func formatTime(timestamp string) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	// Convert to local timezone
	local := t.Local()
	return local.Format("2006-01-02 15:04:05")
}

func getLocalTimezone() string {
	t := time.Now()
	zone, _ := t.Zone()
	return zone
}

func statusClass(status string) string {
	switch status {
	case "PASS":
		return "pass"
	case "FAIL":
		return "fail"
	case "SKIPPED":
		return "skip"
	default:
		return ""
	}
}

func statusSymbol(status string) string {
	switch status {
	case "PASS":
		return "✓"
	case "FAIL":
		return "✗"
	case "SKIPPED":
		return "⊘"
	default:
		return "•"
	}
}

// shortRunID extracts the short ID from a full run ID
// Example: "2025-11-30T08-15-34Z_003617" -> "003617"
func shortRunID(fullID string) string {
	// Find the last underscore and return everything after it
	for i := len(fullID) - 1; i >= 0; i-- {
		if fullID[i] == '_' {
			return fullID[i+1:]
		}
	}
	// If no underscore found, return the full ID
	return fullID
}

// truncateString truncates a string to maxLen characters and adds "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// phaseEmoji returns an emoji for a phase based on its name or ID
func phaseEmoji(phaseName string) string {
	// Normalize to lowercase for matching
	name := strings.ToLower(phaseName)

	// Map phase names/keywords to emojis
	emojiMap := map[string]string{
		"validation":  "🧪",  // Test tube for validation/testing
		"test":        "🧪",  // Test tube
		"testing":     "🧪",  // Test tube
		"build":       "📦",  // Package for build
		"package":     "📦",  // Package
		"compile":     "🔨",  // Hammer for compilation
		"deploy":      "🚀",  // Rocket for deployment
		"release":     "🚀",  // Rocket for release
		"lint":        "🔍",  // Magnifying glass for linting
		"security":    "🔒",  // Lock for security
		"e2e":         "🎯",  // Target for end-to-end tests
		"end-to-end":  "🎯",  // Target
		"integration": "🔗",  // Link for integration
		"setup":       "⚙️", // Gear for setup
		"cleanup":     "🧹",  // Broom for cleanup
		"docs":        "📚",  // Books for documentation
		"publish":     "📤",  // Outbox for publishing
	}

	// Check for exact match first
	if emoji, ok := emojiMap[name]; ok {
		return emoji
	}

	// Check if any keyword is contained in the phase name
	for keyword, emoji := range emojiMap {
		if strings.Contains(name, keyword) {
			return emoji
		}
	}

	// Default emoji
	return "📋" // Clipboard as default
}

// writeRunDetailHTML generates a detail page for a single run
func writeRunDetailHTML(path string, run model.RunRecord) error {
	// Prepare data with log previews
	type TaskWithLog struct {
		model.TaskResult
		LogPreview []string
		OutputPath string
		OutputSize int64
	}

	type DetailData struct {
		model.RunRecord
		TasksWithLogs    []TaskWithLog
		Timezone         string
		RawConfigContent string
		Phases           []PhaseGroup
	}

	data := DetailData{
		RunRecord:     run,
		TasksWithLogs: make([]TaskWithLog, 0, len(run.Tasks)),
		Timezone:      getLocalTimezone(),
	}

	// Load raw config file if it exists
	configPath := filepath.Join(filepath.Dir(path), "config.toml")
	if run.ConfigPath != "" {
		if configData, err := os.ReadFile(configPath); err == nil {
			data.RawConfigContent = string(configData)
		}
	}

	// Parse phases from config
	if phases, err := ParsePhasesFromConfig(configPath, run.Tasks); err == nil {
		data.Phases = phases
	}

	// Load log previews and artifact info for each task
	for _, task := range run.Tasks {
		taskWithLog := TaskWithLog{
			TaskResult: task,
			LogPreview: readLastLines(task.LogPath, 10),
		}

		// Check for artifact file (stored in metrics for artifact format)
		// TODO: The artifact path should be in the workdir + metrics path from task definition
		// We need to reconstruct it from the run record
		// For now, check if we can get it from the task's workdir
		_ = task.Metrics // Placeholder for future artifact handling

		data.TasksWithLogs = append(data.TasksWithLogs, taskWithLog)
	}

	tmpl, err := template.New("rundetail").Funcs(template.FuncMap{
		"formatDuration": formatDuration,
		"formatTime":     formatTime,
		"statusClass":    statusClass,
		"statusSymbol":   statusSymbol,
		"phaseEmoji":     phaseEmoji,
		"string":         func(s model.TaskStatus) string { return string(s) },
		"add": func(a, b interface{}) int {
			return int(toFloat64(a)) + int(toFloat64(b))
		},
		"gt": func(a, b interface{}) bool {
			return toFloat64(a) > toFloat64(b)
		},
		"sub": func(a, b interface{}) float64 {
			return toFloat64(a) - toFloat64(b)
		},
		"mul": func(a, b interface{}) float64 {
			return toFloat64(a) * toFloat64(b)
		},
		"div": func(a, b interface{}) float64 {
			aVal := toFloat64(a)
			bVal := toFloat64(b)
			if bVal == 0 {
				return 0
			}
			return aVal / bVal
		},
		"deref": func(i *int) int {
			if i != nil {
				return *i
			}
			return 0
		},
		"hasPrefix": func(s, prefix string) bool { return len(s) >= len(prefix) && s[:len(prefix)] == prefix },
		"trimPrefix": func(s, prefix string) string {
			if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
				return s[len(prefix):]
			}
			return s
		},
		"slice": func() []model.ConfigValue { return []model.ConfigValue{} },
		"append": func(slice []model.ConfigValue, item model.ConfigValue) []model.ConfigValue {
			return append(slice, item)
		},
		"truncate": func(s string, maxLen int) string {
			if len(s) <= maxLen {
				return s
			}
			return s[:maxLen-3] + "..."
		},
	}).Parse(runDetailTemplate)

	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", cerr)
		}
	}()

	return tmpl.Execute(f, data)
}

// toFloat64 converts various numeric types to float64 for template arithmetic
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float64:
		return val
	case float32:
		return float64(val)
	default:
		return 0
	}
}

// readLastLines reads the last N lines from a file and strips ANSI codes
func readLastLines(path string, n int) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{"Error reading log file"}
	}

	// Strip ANSI codes from entire content first
	cleanData := stripansi.Strip(string(data))

	// Split by newlines
	allLines := []string{}
	currentLine := ""
	for _, ch := range cleanData {
		if ch == '\n' {
			allLines = append(allLines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(ch)
		}
	}
	if currentLine != "" {
		allLines = append(allLines, currentLine)
	}

	// Return last N lines
	if len(allLines) <= n {
		return allLines
	}
	return allLines[len(allLines)-n:]
}

const dashboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>devpipe Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: #f5f5f5;
            color: #333;
            line-height: 1.6;
        }
        
        .container {
            max-width: 1100px;
            margin: 0 auto;
            padding: 20px;
        }
        
        header {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 30px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .header-content {
            flex: 1;
        }
        
        h1 {
            font-size: 32px;
            margin-bottom: 8px;
            color: #2c3e50;
        }
        
        .tagline {
            color: #7f8c8d;
            font-size: 16px;
            margin-bottom: 12px;
        }
        
        .header-stats {
            color: #7f8c8d;
            font-size: 14px;
            margin-bottom: 8px;
        }
        
        .header-stats strong {
            color: #2c3e50;
            font-weight: 600;
        }
        
        .subtitle {
            color: #95a5a6;
            font-size: 13px;
        }
        
        .mascot-label {
            text-align: center;
            color: #95a5a6;
            font-size: 12px;
        }
        
        .mascot-wrapper {
            display: flex;
            align-items: flex-start;
        }
        
        .speech-bubble {
            position: relative;
            background: #f8f9fa;
            border: 2px solid #dee2e6;
            border-radius: 12px;
            padding: 10px 14px;
            font-size: 13px;
            color: #2c3e50;
            white-space: nowrap;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
            margin-top: 50px;
        }
        
        .speech-bubble:after {
            content: '';
            position: absolute;
            right: -10px;
            top: 50%;
            transform: translateY(-50%);
            width: 0;
            height: 0;
            border-left: 10px solid #dee2e6;
            border-top: 8px solid transparent;
            border-bottom: 8px solid transparent;
        }
        
        .speech-bubble:before {
            content: '';
            position: absolute;
            right: -7px;
            top: 50%;
            transform: translateY(-50%);
            width: 0;
            height: 0;
            border-left: 8px solid #f8f9fa;
            border-top: 6px solid transparent;
            border-bottom: 6px solid transparent;
            z-index: 1;
        }
        
        .version-info {
            color: #95a5a6;
            font-size: 13px;
            font-family: monospace;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        
        .stat-value {
            font-size: 36px;
            font-weight: bold;
            color: #2c3e50;
        }
        
        .stat-label {
            color: #7f8c8d;
            font-size: 14px;
            margin-top: 5px;
        }
        
        .section {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }
        
        h2 {
            font-size: 24px;
            margin-bottom: 20px;
            color: #2c3e50;
        }
        
        table {
            width: 100%;
            border-collapse: collapse;
            table-layout: fixed;
        }
        
        th {
            text-align: left;
            padding: 12px;
            background: #f8f9fa;
            font-weight: 600;
            color: #2c3e50;
            border-bottom: 2px solid #dee2e6;
        }
        
        td {
            padding: 12px;
            border-bottom: 1px solid #dee2e6;
        }
        
        /* Fixed column widths for Recent Runs table */
        #runsTable th:nth-child(1),
        #runsTable td:nth-child(1) { width: 90px; white-space: nowrap; } /* Run ID (short) */
        #runsTable th:nth-child(2),
        #runsTable td:nth-child(2) { width: 170px; white-space: nowrap; } /* Timestamp */
        #runsTable th:nth-child(3),
        #runsTable td:nth-child(3) { width: 110px; } /* Status */
        #runsTable th:nth-child(4),
        #runsTable td:nth-child(4) { width: 80px; white-space: nowrap; } /* Duration */
        #runsTable th:nth-child(5),
        #runsTable td:nth-child(5) { width: 60px; text-align: center; } /* Tasks */
        #runsTable th:nth-child(6),
        #runsTable td:nth-child(6) { width: auto; min-width: 300px; } /* Command - takes remaining space */
        
        tr:hover {
            background: #f8f9fa;
        }
        
        .status-pass {
            color: #27ae60;
            font-weight: bold;
        }
        
        .status-fail {
            color: #e74c3c;
            font-weight: bold;
        }
        
        .status-skip {
            color: #f39c12;
            font-weight: bold;
        }
        
        .badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 600;
        }
        
        .badge-pass {
            background: #d4edda;
            color: #155724;
        }
        
        .badge-fail {
            background: #f8d7da;
            color: #721c24;
        }
        
        .badge-skip {
            background: #fff3cd;
            color: #856404;
        }
        
        .mono {
            font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
            font-size: 13px;
        }
        
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #7f8c8d;
        }
        
        .empty-state-icon {
            font-size: 48px;
            margin-bottom: 20px;
        }
        
        .load-more-btn {
            padding: 12px 24px;
            background: #3498db;
            color: white;
            border: none;
            border-radius: 6px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            transition: background 0.2s;
        }
        
        .load-more-btn:hover {
            background: #2980b9;
        }
        
        .load-more-btn:disabled {
            background: #95a5a6;
            cursor: not-allowed;
        }
        
        /* Mascot Styles */
        .mascot {
            width: 200px;
            height: 200px;
            flex-shrink: 0;
            pointer-events: none;
            margin-left: 30px;
        }
        
        .mascot-container {
            position: relative;
            width: 100%;
            height: 100%;
        }
        
        .mascot-image {
            width: 100%;
            height: 100%;
            object-fit: contain;
        }
        
        .mascot-eyes-overlay {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            pointer-events: none;
        }
        
        .mascot-eye {
            position: absolute;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        
        .mascot-eye-left {
            top: 29.5%;
            left: 13.1%;
            width: 13.08px;
            height: 19.91px;
        }
        
        .mascot-eye-right {
            top: 29.7%;
            left: 37.4%;
            width: 16.50px;
            height: 21.61px;
        }
        
        .mascot-pupil {
            position: relative;
            width: 5.69px;
            height: 6.83px;
            background: #ffffff;
            border-radius: 50%;
            transition: transform 0.1s ease-out;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="header-content">
                <h1>📊 DevPipe Dashboard</h1>
                <div class="tagline">Fast, local pipeline runner for development workflows.</div>
                <div class="header-stats"><strong>{{.TotalRuns}}</strong> Total Runs, <strong>{{len .TaskStats}}</strong> Tasks.</div>
                <br />
                <div class="version-info">{{.Version}}</div>
                <div class="subtitle">Last updated: {{formatTime .LastGenerated}}</div>
            </div>
            <!-- Mascot -->
            <div class="mascot-wrapper">
                <div class="speech-bubble">{{.Greeting}} {{.Username}}!</div>
                <div>
                    <div class="mascot">
                        <div class="mascot-container">
                            <img src="mascot/squirrel-blank-eyes-transparent-cropped.png" alt="DevPipe Squirrel" class="mascot-image" onerror="this.style.display='none'">
                            <div class="mascot-eyes-overlay">
                                <div class="mascot-eye mascot-eye-left">
                                    <div class="mascot-pupil"></div>
                                </div>
                                <div class="mascot-eye mascot-eye-right">
                                    <div class="mascot-pupil"></div>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="mascot-label">(flowmunk)</div>
                </div>
            </div>
        </header>
        
        <div class="section">
            <h2>Recent Runs</h2>
            {{if .RecentRuns}}
            <table id="runsTable">
                <thead>
                    <tr>
                        <th>Run ID</th>
                        <th>Timestamp ({{.Timezone}})</th>
                        <th>Status</th>
                        <th>Duration</th>
                        <th>Tasks</th>
                        <th>Version</th>
                        <th>Command</th>
                    </tr>
                </thead>
                <tbody id="runsTableBody">
                    {{range .RecentRuns}}
                    <tr class="run-row" data-index="{{$.RecentRuns | len}}">
                        <td class="mono"><a href="runs/{{.RunID}}/report.html" title="{{.RunID}}">{{shortRunID .RunID}}</a></td>
                        <td>{{formatTime .Timestamp}}</td>
                        <td>
                            <span class="badge badge-{{.Status | statusClass}}">
                                {{statusSymbol .Status}} {{.Status}}
                            </span>
                        </td>
                        <td>{{formatDuration .Duration}}</td>
                        <td>{{.TotalTasks}}</td>
                        <td class="mono" style="font-size: 11px;">{{.PipelineVersion}}</td>
                        <td class="mono" style="font-size: 11px; max-width: 400px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" title="{{.Command}}">{{.Command}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            <div id="loadMoreContainer" style="text-align: center; margin-top: 20px;">
                <button id="loadMoreBtn" class="load-more-btn" onclick="loadMoreRuns()" style="display: none;">
                    Load More (25)
                </button>
            </div>
            {{else}}
            <div class="empty-state">
                <div class="empty-state-icon">📭</div>
                <p>No runs yet. Run devpipe to see results here!</p>
            </div>
            {{end}}
        </div>
        
        <div class="section">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                <h2 style="margin: 0;">Task Statistics</h2>
                <div style="display: flex; align-items: center; gap: 10px;">
                    <label for="statsFilter" style="font-size: 14px; color: #7f8c8d;">Show stats for:</label>
                    <select id="statsFilter" onchange="filterTaskStats()" style="padding: 8px 12px; border: 1px solid #dee2e6; border-radius: 4px; font-size: 14px; background: white; cursor: pointer;">
                        <option value="recent">Most Recent Run</option>
                        <option value="last25" selected>Last 25 Runs</option>
                        <option value="all">All Runs</option>
                    </select>
                </div>
            </div>
            {{if .TaskStats}}
            
            <!-- All Runs Stats -->
            <table class="task-stats-table" data-filter="all" style="display: none;">
                <thead>
                    <tr>
                        <th>Task</th>
                        <th>Total Runs</th>
                        <th>Pass Rate</th>
                        <th>Avg Duration</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .TaskStats}}
                    <tr>
                        <td><strong>{{.Name}}</strong> <span class="mono" style="color: #7f8c8d;">({{.ID}})</span></td>
                        <td>{{.TotalRuns}}</td>
                        <td>
                            {{if gt .TotalRuns 0}}
                            {{printf "%.0f%%" (div (mul (float64 .PassCount) 100.0) (float64 .TotalRuns))}}
                            ({{.PassCount}}/{{.TotalRuns}})
                            {{else}}
                            N/A
                            {{end}}
                        </td>
                        <td>{{formatDuration (int64 .AvgDuration)}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            
            <!-- Last 25 Runs Stats -->
            <table class="task-stats-table" data-filter="last25" style="display: table;">
                <thead>
                    <tr>
                        <th>Task</th>
                        <th>Total Runs</th>
                        <th>Pass Rate</th>
                        <th>Avg Duration</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .TaskStatsLast25}}
                    <tr>
                        <td><strong>{{.Name}}</strong> <span class="mono" style="color: #7f8c8d;">({{.ID}})</span></td>
                        <td>{{.TotalRuns}}</td>
                        <td>
                            {{if gt .TotalRuns 0}}
                            {{printf "%.0f%%" (div (mul (float64 .PassCount) 100.0) (float64 .TotalRuns))}}
                            ({{.PassCount}}/{{.TotalRuns}})
                            {{else}}
                            N/A
                            {{end}}
                        </td>
                        <td>{{formatDuration (int64 .AvgDuration)}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            
            <!-- Most Recent Run Stats -->
            <table class="task-stats-table" data-filter="recent" style="display: none;">
                <thead>
                    <tr>
                        <th>Task</th>
                        <th>Total Runs</th>
                        <th>Pass Rate</th>
                        <th>Avg Duration</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .TaskStatsRecent}}
                    <tr>
                        <td><strong>{{.Name}}</strong> <span class="mono" style="color: #7f8c8d;">({{.ID}})</span></td>
                        <td>{{.TotalRuns}}</td>
                        <td>
                            {{if gt .TotalRuns 0}}
                            {{printf "%.0f%%" (div (mul (float64 .PassCount) 100.0) (float64 .TotalRuns))}}
                            ({{.PassCount}}/{{.TotalRuns}})
                            {{else}}
                            N/A
                            {{end}}
                        </td>
                        <td>{{formatDuration (int64 .AvgDuration)}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            
            {{else}}
            <div class="empty-state">
                <div class="empty-state-icon">📊</div>
                <p>No task statistics available yet.</p>
            </div>
            {{end}}
        </div>
    </div>
    
    <script>
        // Pagination for Recent Runs
        let visibleRunCount = 25;
        const runsPerLoad = 25;
        const maxRuns = 100;
        
        function initializePagination() {
            const allRows = document.querySelectorAll('.run-row');
            const totalRuns = allRows.length;
            
            // Hide rows beyond the initial count
            allRows.forEach((row, index) => {
                if (index >= visibleRunCount) {
                    row.style.display = 'none';
                }
            });
            
            // Show "Load More" button if there are more runs to display
            const loadMoreBtn = document.getElementById('loadMoreBtn');
            if (totalRuns > visibleRunCount) {
                loadMoreBtn.style.display = 'inline-block';
                updateLoadMoreButton(totalRuns);
            }
        }
        
        function loadMoreRuns() {
            const allRows = document.querySelectorAll('.run-row');
            const totalRuns = allRows.length;
            
            // Show next batch of runs
            const newVisibleCount = Math.min(visibleRunCount + runsPerLoad, totalRuns);
            
            for (let i = visibleRunCount; i < newVisibleCount; i++) {
                allRows[i].style.display = '';
            }
            
            visibleRunCount = newVisibleCount;
            
            // Update or hide the button
            if (visibleRunCount >= totalRuns) {
                document.getElementById('loadMoreBtn').style.display = 'none';
            } else {
                updateLoadMoreButton(totalRuns);
            }
        }
        
        function updateLoadMoreButton(totalRuns) {
            const remaining = totalRuns - visibleRunCount;
            const nextBatch = Math.min(runsPerLoad, remaining);
            document.getElementById('loadMoreBtn').textContent = 'Load More (' + nextBatch + ')';
        }
        
        // Task Statistics Filtering - Simple table toggling
        function filterTaskStats() {
            const filter = document.getElementById('statsFilter').value;
            const tables = document.querySelectorAll('.task-stats-table');
            
            // Hide all tables
            tables.forEach(table => {
                table.style.display = 'none';
            });
            
            // Show the selected table
            const selectedTable = document.querySelector('.task-stats-table[data-filter="' + filter + '"]');
            if (selectedTable) {
                selectedTable.style.display = 'table';
            }
        }
        
        // Initialize on page load
        document.addEventListener('DOMContentLoaded', function() {
            initializePagination();
        });
        
        // Mascot eye tracking
        (function() {
            const pupils = document.querySelectorAll('.mascot-pupil');
            const eyes = document.querySelectorAll('.mascot-eye');
            const mascot = document.querySelector('.mascot');
            
            // Base movement bounds for 200px mascot
            const baseLeftMaxDistanceX = 1.3;
            const baseLeftMaxDistanceY = 2.7;
            const baseRightMaxDistanceX = 3.0;
            const baseRightMaxDistanceY = 2.7;
            
            document.addEventListener('mousemove', function(e) {
                pupils.forEach(function(pupil, index) {
                    const eye = eyes[index];
                    const eyeRect = eye.getBoundingClientRect();
                    
                    // Get eye center position
                    const eyeCenterX = eyeRect.left + eyeRect.width / 2;
                    const eyeCenterY = eyeRect.top + eyeRect.height / 2;
                    
                    // Calculate angle to mouse
                    const angleRad = Math.atan2(e.clientY - eyeCenterY, e.clientX - eyeCenterX);
                    
                    // Get current mascot size and calculate scale factor
                    const currentSize = parseFloat(mascot.style.width) || 200;
                    const scaleFactor = currentSize / 200;
                    
                    // Scale movement bounds based on current size
                    const leftMaxDistanceX = baseLeftMaxDistanceX * scaleFactor;
                    const leftMaxDistanceY = baseLeftMaxDistanceY * scaleFactor;
                    const rightMaxDistanceX = baseRightMaxDistanceX * scaleFactor;
                    const rightMaxDistanceY = baseRightMaxDistanceY * scaleFactor;
                    
                    // Get max distance for this eye
                    const maxDistanceX = index === 0 ? leftMaxDistanceX : rightMaxDistanceX;
                    const maxDistanceY = index === 0 ? leftMaxDistanceY : rightMaxDistanceY;
                    
                    // Calculate pupil position
                    let pupilX = Math.cos(angleRad) * maxDistanceX;
                    let pupilY = Math.sin(angleRad) * maxDistanceY;
                    
                    // Constrain to ellipse
                    const normalizedX = pupilX / maxDistanceX;
                    const normalizedY = pupilY / maxDistanceY;
                    const distanceFromCenter = Math.sqrt(normalizedX * normalizedX + normalizedY * normalizedY);
                    
                    if (distanceFromCenter > 1) {
                        pupilX = (normalizedX / distanceFromCenter) * maxDistanceX;
                        pupilY = (normalizedY / distanceFromCenter) * maxDistanceY;
                    }
                    
                    // Apply transform
                    pupil.style.transform = 'translate(' + pupilX + 'px, ' + pupilY + 'px)';
                });
            });
        })();
    </script>
</body>
</html>
`

const runDetailTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Run {{.RunID}} - devpipe</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: #f5f5f5;
            color: #333;
            line-height: 1.6;
        }
        
        .container {
            max-width: 1100px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .breadcrumb {
            margin-bottom: 20px;
            color: #7f8c8d;
        }
        
        .breadcrumb a {
            color: #3498db;
            text-decoration: none;
        }
        
        .breadcrumb a:hover {
            text-decoration: underline;
        }
        
        header {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }
        
        h1 {
            font-size: 28px;
            margin-bottom: 10px;
            color: #2c3e50;
        }
        
        .run-meta {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-top: 20px;
        }
        
        .meta-item {
            display: flex;
            flex-direction: column;
        }
        
        .meta-label {
            font-size: 12px;
            color: #7f8c8d;
            text-transform: uppercase;
            margin-bottom: 5px;
        }
        
        .meta-value {
            font-size: 16px;
            color: #2c3e50;
        }
        
        .section {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }
        
        h2 {
            font-size: 20px;
            margin-bottom: 20px;
            color: #2c3e50;
        }
        
        .task-card {
            border: 1px solid #dee2e6;
            border-radius: 6px;
            padding: 20px;
            margin-bottom: 15px;
        }
        
        .task-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 15px;
            gap: 15px;
        }
        
        .task-header > div:first-child {
            flex: 1;
            min-width: 0;
        }
        
        .task-title {
            font-size: 18px;
            font-weight: 600;
            color: #2c3e50;
            word-break: break-word;
        }
        
        .task-id {
            font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
            font-size: 13px;
            color: #7f8c8d;
            margin-left: 10px;
        }
        
        .task-details {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
            margin-top: 15px;
        }
        
        .detail-item {
            display: flex;
            flex-direction: column;
        }
        
        .detail-label {
            font-size: 11px;
            color: #7f8c8d;
            text-transform: uppercase;
            margin-bottom: 3px;
        }
        
        .detail-value {
            font-size: 14px;
            color: #2c3e50;
        }
        
        .exit-code-success {
            color: #27ae60;
            font-weight: bold;
        }
        
        .exit-code-error {
            color: #e74c3c;
            font-weight: bold;
        }
        
        .status-pass {
            color: #27ae60;
            font-weight: bold;
        }
        
        .status-fail {
            color: #e74c3c;
            font-weight: bold;
        }
        
        .status-skip {
            color: #f39c12;
            font-weight: bold;
        }
        
        .badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 600;
        }
        
        .badge-pass {
            background: #d4edda;
            color: #155724;
        }
        
        .badge-fail {
            background: #f8d7da;
            color: #721c24;
        }
        
        .badge-skip {
            background: #fff3cd;
            color: #856404;
        }
        
        .mono {
            font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
            font-size: 13px;
        }
        
        .metrics-box {
            background: #f8f9fa;
            border-left: 4px solid #3498db;
            padding: 15px;
            margin-top: 15px;
            border-radius: 4px;
        }
        
        /* Phase Flow Styles */
        .phase-flow-container {
            position: relative;
        }
        
        .phase-flow-container.fullscreen {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            z-index: 9999;
            background: white;
            padding: 20px;
            overflow-y: auto;
            margin: 0;
            display: flex;
            flex-direction: column;
        }
        
        .phase-flow-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            flex-shrink: 0;
        }
        
        .fullscreen-btn {
            padding: 8px 16px;
            background: #3498db;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
        }
        
        .fullscreen-btn:hover {
            background: #2980b9;
        }
        
        .fullscreen-btn.active {
            background: #27ae60;
        }
        
        .fullscreen-btn.active:hover {
            background: #229954;
        }
        
        /* Column slider styles */
        .slider-container {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        .slider-label {
            font-size: 14px;
            color: #2c3e50;
            font-weight: 500;
            white-space: nowrap;
        }
        
        .column-slider {
            width: 120px;
            height: 6px;
            border-radius: 3px;
            background: #dee2e6;
            outline: none;
            -webkit-appearance: none;
        }
        
        .column-slider::-webkit-slider-thumb {
            -webkit-appearance: none;
            appearance: none;
            width: 18px;
            height: 18px;
            border-radius: 50%;
            background: #3498db;
            cursor: pointer;
        }
        
        .column-slider::-moz-range-thumb {
            width: 18px;
            height: 18px;
            border-radius: 50%;
            background: #3498db;
            cursor: pointer;
            border: none;
        }
        
        .phase-flow-scroll {
            overflow-x: auto;
            overflow-y: hidden;
            padding-bottom: 10px;
        }
        
        .phase-flow-container.fullscreen .phase-flow-scroll {
            flex: 1;
            overflow-x: auto;
            overflow-y: auto;
            min-height: 0;
        }
        
        .phase-flow {
            display: flex;
            gap: 40px;
            min-width: min-content;
        }
        
        .phase-container {
            flex-shrink: 0;
            width: fit-content;
            background: #f8f9fa;
            border: 2px solid #dee2e6;
            border-radius: 8px;
            overflow: visible;
        }
        
        .phase-header {
            background: #f8f9fa;
            border-bottom: 2px solid #dee2e6;
            padding: 12px;
        }
        
        .phase-header-top {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
        }
        
        .phase-header h3 {
            font-size: 18px;
            color: #2c3e50;
            margin: 0;
        }
        
        .phase-status {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        
        .phase-status-icon {
            font-size: 20px;
        }
        
        .phase-status-icon.success { color: #27ae60; }
        .phase-status-icon.fail { color: #e74c3c; }
        
        .phase-desc {
            font-size: 12px;
            color: #7f8c8d;
            margin-top: 6px;
            margin-bottom: 4px;
            line-height: 1.5;
            overflow: hidden;
            text-overflow: ellipsis;
            display: -webkit-box;
            -webkit-line-clamp: 2;
            -webkit-box-orient: vertical;
        }
        
        .phase-meta {
            font-size: 13px;
            color: #7f8c8d;
            display: flex;
            justify-content: space-between;
        }
        
        .phase-tasks {
            padding: 12px;
            display: flex;
            flex-direction: column;
            gap: 8px;
        }
        
        .phase-tasks.compact {
            display: grid;
            grid-auto-flow: column;
        }
        
        .phase-task-card {
            background: white;
            border: 1px solid #dee2e6;
            border-radius: 6px;
            padding: 12px;
            cursor: pointer;
            transition: all 0.2s;
            box-sizing: border-box;
        }
        
        .phase-task-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            border-color: #3498db;
        }
        
        .phase-task-card-header {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 6px;
        }
        
        .phase-task-icon {
            font-size: 18px;
        }
        
        .phase-task-icon.success { color: #27ae60; }
        .phase-task-icon.fail { color: #e74c3c; }
        .phase-task-icon.skip { color: #f39c12; }
        
        .phase-task-name {
            font-weight: 600;
            font-size: 14px;
            color: #2c3e50;
            flex: 1;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            min-width: 0;
        }
        
        .phase-task-duration {
            font-size: 12px;
            color: #7f8c8d;
            font-weight: 500;
        }
        
        .phase-task-desc {
            font-size: 11px;
            color: #7f8c8d;
            margin-top: 4px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        
        .phase-task-type {
            display: inline-block;
            font-size: 10px;
            padding: 2px 6px;
            background: #ecf0f1;
            color: #7f8c8d;
            border-radius: 3px;
            margin-top: 4px;
        }
        
        .phase-arrow {
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 32px;
            color: #95a5a6;
            flex-shrink: 0;
        }
        
        .metrics-title {
            font-weight: 600;
            margin-bottom: 10px;
            color: #2c3e50;
        }
        
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            gap: 10px;
        }
        
        .log-link {
            display: inline-block;
            margin-top: 10px;
            color: #3498db;
            text-decoration: none;
            font-size: 14px;
        }
        
        .log-link:hover {
            text-decoration: underline;
        }
        
        /* Mascot Styles */
        .mascot {
            position: fixed;
            top: 20px;
            right: 20px;
            width: 200px;
            height: 200px;
            z-index: 1000;
            pointer-events: none;
            transition: opacity 0.3s ease-out;
        }
        
        .mascot.hidden {
            opacity: 0;
            pointer-events: none;
        }
        
        @media (max-width: 1500px) {
            .mascot {
                display: none;
            }
        }
        
        .mascot-container {
            position: relative;
            width: 100%;
            height: 100%;
        }
        
        .mascot-image {
            width: 100%;
            height: 100%;
            object-fit: contain;
        }
        
        .mascot-eyes-overlay {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            pointer-events: none;
        }
        
        .mascot-eye {
            position: absolute;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        
        .mascot-eye-left {
            top: 29.5%;
            left: 13.1%;
            width: 13.08px;
            height: 19.91px;
        }
        
        .mascot-eye-right {
            top: 29.7%;
            left: 37.4%;
            width: 16.50px;
            height: 21.61px;
        }
        
        .mascot-pupil {
            position: relative;
            width: 5.69px;
            height: 6.83px;
            background: #ffffff;
            border-radius: 50%;
            transition: transform 0.1s ease-out;
        }
    </style>
</head>
<body>
    <!-- Mascot -->
    <div class="mascot">
        <div class="mascot-container">
            <img src="../../mascot/squirrel-blank-eyes-transparent.png" alt="DevPipe Squirrel" class="mascot-image" onerror="this.style.display='none'">
            <div class="mascot-eyes-overlay">
                <div class="mascot-eye mascot-eye-left">
                    <div class="mascot-pupil"></div>
                </div>
                <div class="mascot-eye mascot-eye-right">
                    <div class="mascot-pupil"></div>
                </div>
            </div>
        </div>
    </div>

    <div class="container">
        <div class="breadcrumb">
            <a href="../../report.html">← Back to Dashboard</a>
        </div>
        
        <header>
            <h1>Run Details</h1>
            <div class="run-meta">
                <div class="meta-item">
                    <div class="meta-label">Run ID</div>
                    <div class="meta-value mono">{{.RunID}}</div>
                </div>
                <div class="meta-item">
                    <div class="meta-label">Timestamp ({{.Timezone}})</div>
                    <div class="meta-value">{{formatTime .Timestamp}}</div>
                </div>
                <div class="meta-item">
                    <div class="meta-label">Project Root</div>
                    <div class="meta-value mono">{{.ProjectRoot}}</div>
                </div>
                {{if .PipelineVersion}}
                <div class="meta-item">
                    <div class="meta-label">Pipeline Version</div>
                    <div class="meta-value mono">{{.PipelineVersion}}</div>
                </div>
                {{end}}
                {{if .ReportVersion}}
                <div class="meta-item">
                    <div class="meta-label">Report Version</div>
                    <div class="meta-value mono">{{.ReportVersion}}</div>
                </div>
                {{end}}
            </div>
        </header>
        
        <div class="section">
            <h2>Run</h2>
            
            {{if .Command}}
            <div class="detail-item" style="margin-bottom: 20px;">
                <div class="detail-label">Command</div>
                <div class="detail-value mono" style="font-size: 11px; background: #f8f9fa; padding: 8px; border-radius: 4px;">{{.Command}}</div>
            </div>
            {{end}}
            
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 30px; margin-bottom: 20px;">
                <div>
                    <h3 style="font-size: 16px; color: #2c3e50; margin-bottom: 15px; padding-bottom: 8px; border-bottom: 2px solid #dee2e6;">Configuration</h3>
                    <div class="task-details" style="grid-template-columns: 1fr;">
                        <div class="detail-item">
                            <div class="detail-label">Config File</div>
                            <div class="detail-value">
                                {{if .ConfigPath}}
                                <a href="config.toml" class="log-link">{{.ConfigPath}}</a>
                                {{else}}
                                Built-in tasks
                                {{end}}
                            </div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Fast Mode</div>
                            <div class="detail-value">{{if .Flags.Fast}}Yes{{else}}No{{end}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Fail Fast</div>
                            <div class="detail-value">{{if .Flags.FailFast}}Yes{{else}}No{{end}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Dry Run</div>
                            <div class="detail-value">{{if .Flags.DryRun}}Yes{{else}}No{{end}}</div>
                        </div>
                        {{if .Flags.Only}}
                        <div class="detail-item">
                            <div class="detail-label">Only</div>
                            <div class="detail-value mono">{{.Flags.Only}}</div>
                        </div>
                        {{end}}
                        {{if .Flags.Skip}}
                        <div class="detail-item">
                            <div class="detail-label">Skip</div>
                            <div class="detail-value mono">{{range .Flags.Skip}}{{.}} {{end}}</div>
                        </div>
                        {{end}}
                    </div>
                </div>
                
                {{if .Git}}
                <div>
                    <h3 style="font-size: 16px; color: #2c3e50; margin-bottom: 15px; padding-bottom: 8px; border-bottom: 2px solid #dee2e6;">Git Information</h3>
                    <div class="task-details" style="grid-template-columns: 1fr;">
                        <div class="detail-item">
                            <div class="detail-label">Mode</div>
                            <div class="detail-value">{{.Git.mode}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Reference</div>
                            <div class="detail-value mono">{{.Git.ref}}</div>
                        </div>
                        {{if .Git.changedFiles}}
                        <div class="detail-item">
                            <div class="detail-label">Changed Files ({{len .Git.changedFiles}})</div>
                            <div class="detail-value mono" style="font-size: 12px; line-height: 1.8;">
                                {{range .Git.changedFiles}}
                                <div>{{.}}</div>
                                {{end}}
                            </div>
                        </div>
                        {{else}}
                        <div class="detail-item">
                            <div class="detail-label">Changed Files</div>
                            <div class="detail-value">0 files</div>
                        </div>
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        
        {{if or .EffectiveConfig .ConfigPath}}
        <div class="section">
            <details>
                <summary style="cursor: pointer; font-weight: 600; color: #2c3e50; font-size: 20px; margin-bottom: 20px;">
                    <h2 style="display: inline;">⚙️ Configuration</h2>
                </summary>
                <div>
                    {{if .EffectiveConfig}}
                    <div>
                        <div style="border: 1px solid #dee2e6; border-radius: 6px; padding: 15px; background: #f8f9fa;">
                            <h3 style="font-weight: 600; color: #2c3e50; font-size: 16px; margin: 0 0 10px 0;">⚙️ Effective Configuration</h3>
                            <p style="color: #7f8c8d; margin: 15px 0; font-size: 13px;">
                                Final configuration values used for this run, including their sources and overrides.
                            </p>
                        
                        {{$defaults := slice}}
                        {{$defaultsGit := slice}}
                        {{$taskDefaults := slice}}
                        {{$tasks := slice}}
                        {{range .EffectiveConfig.Values}}
                            {{if hasPrefix .Key "defaults.git."}}
                                {{$defaultsGit = append $defaultsGit .}}
                            {{else if hasPrefix .Key "defaults."}}
                                {{$defaults = append $defaults .}}
                            {{else if hasPrefix .Key "task_defaults."}}
                                {{$taskDefaults = append $taskDefaults .}}
                            {{else if hasPrefix .Key "tasks."}}
                                {{$tasks = append $tasks .}}
                            {{end}}
                        {{end}}
                        
                        {{if $defaults}}
                        <div style="margin-top: 20px;">
                            <h4 style="color: #2c3e50; font-size: 14px; margin-bottom: 10px; padding-bottom: 5px; border-bottom: 2px solid #dee2e6;">Defaults</h4>
                            <table style="width: 100%; font-size: 13px;">
                                {{range $defaults}}
                                <tr>
                                    <td class="mono" style="color: #495057; padding: 6px 0; width: 30%;">{{trimPrefix .Key "defaults."}}</td>
                                    <td class="mono" style="font-weight: bold; padding: 6px 0;">{{.Value}}</td>
                                    <td style="padding: 6px 0; text-align: right;">
                                        {{if eq .Source "config-file"}}
                                        <span class="badge" style="background: #d4edda; color: #155724; font-size: 10px;">📄 Config</span>
                                        {{else if eq .Source "cli-flag"}}
                                        <span class="badge" style="background: #cce5ff; color: #004085; font-size: 10px;">🚩 CLI</span>
                                        {{else if eq .Source "default"}}
                                        <span class="badge" style="background: #e2e3e5; color: #383d41; font-size: 10px;">⚙️ Default</span>
                                        {{end}}
                                        {{if .Overrode}}
                                        <br><span style="color: #7f8c8d; font-size: 11px;">(was: {{.Overrode}})</span>
                                        {{end}}
                                    </td>
                                </tr>
                                {{end}}
                            </table>
                        </div>
                        {{end}}
                        
                        {{if $defaultsGit}}
                        <div style="margin-top: 20px;">
                            <h4 style="color: #2c3e50; font-size: 14px; margin-bottom: 10px; padding-bottom: 5px; border-bottom: 2px solid #dee2e6;">Git Settings</h4>
                            <table style="width: 100%; font-size: 13px;">
                                {{range $defaultsGit}}
                                <tr>
                                    <td class="mono" style="color: #495057; padding: 6px 0; width: 30%;">{{trimPrefix .Key "defaults.git."}}</td>
                                    <td class="mono" style="font-weight: bold; padding: 6px 0;">{{.Value}}</td>
                                    <td style="padding: 6px 0; text-align: right;">
                                        {{if eq .Source "config-file"}}
                                        <span class="badge" style="background: #d4edda; color: #155724; font-size: 10px;">📄 Config</span>
                                        {{else if eq .Source "cli-flag"}}
                                        <span class="badge" style="background: #cce5ff; color: #004085; font-size: 10px;">🚩 CLI</span>
                                        {{else if eq .Source "default"}}
                                        <span class="badge" style="background: #e2e3e5; color: #383d41; font-size: 10px;">⚙️ Default</span>
                                        {{end}}
                                        {{if .Overrode}}
                                        <br><span style="color: #7f8c8d; font-size: 11px;">(was: {{.Overrode}})</span>
                                        {{end}}
                                    </td>
                                </tr>
                                {{end}}
                            </table>
                        </div>
                        {{end}}
                        
                        {{if $taskDefaults}}
                        <div style="margin-top: 20px;">
                            <h4 style="color: #2c3e50; font-size: 14px; margin-bottom: 10px; padding-bottom: 5px; border-bottom: 2px solid #dee2e6;">Task Defaults</h4>
                            <table style="width: 100%; font-size: 13px;">
                                {{range $taskDefaults}}
                                <tr>
                                    <td class="mono" style="color: #495057; padding: 6px 0; width: 30%;">{{trimPrefix .Key "task_defaults."}}</td>
                                    <td class="mono" style="font-weight: bold; padding: 6px 0;">{{.Value}}</td>
                                    <td style="padding: 6px 0; text-align: right;">
                                        {{if eq .Source "config-file"}}
                                        <span class="badge" style="background: #d4edda; color: #155724; font-size: 10px;">📄 Config</span>
                                        {{else if eq .Source "default"}}
                                        <span class="badge" style="background: #e2e3e5; color: #383d41; font-size: 10px;">⚙️ Default</span>
                                        {{end}}
                                    </td>
                                </tr>
                                {{end}}
                            </table>
                        </div>
                        {{end}}
                        
                        {{if $tasks}}
                        <div style="margin-top: 20px;">
                            <h4 style="color: #2c3e50; font-size: 14px; margin-bottom: 10px; padding-bottom: 5px; border-bottom: 2px solid #dee2e6;">Task Overrides</h4>
                            <table style="width: 100%; font-size: 13px;">
                                {{range $tasks}}
                                <tr>
                                    <td class="mono" style="color: #495057; padding: 6px 0; width: 30%;">{{.Key}}</td>
                                    <td class="mono" style="font-weight: bold; padding: 6px 0;">{{.Value}}</td>
                                    <td style="padding: 6px 0; text-align: right;">
                                        {{if eq .Source "historical"}}
                                        <span class="badge" style="background: #fff3cd; color: #856404; font-size: 10px;">📊 Historical</span>
                                        {{else if eq .Source "config-file"}}
                                        <span class="badge" style="background: #d4edda; color: #155724; font-size: 10px;">📄 Config</span>
                                        {{end}}
                                        {{if .Overrode}}
                                        <br><span style="color: #7f8c8d; font-size: 11px;">(was: {{.Overrode}})</span>
                                        {{end}}
                                    </td>
                                </tr>
                                {{end}}
                            </table>
                        </div>
                        {{end}}
                        </div>
                    </div>
                    {{end}}
                    
                    {{if .ConfigPath}}
                    <div style="margin-top: 20px;">
                        <div style="border: 1px solid #dee2e6; border-radius: 6px; padding: 15px; background: #f8f9fa;">
                            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">
                                <h3 style="font-weight: 600; color: #2c3e50; font-size: 16px; margin: 0;">📄 Raw Configuration File</h3>
                                <button onclick="toggleConfigFullscreen()" style="background: #3498db; color: white; border: none; padding: 6px 12px; border-radius: 4px; cursor: pointer; font-size: 12px; font-weight: 600;">
                                    ⛶ Fullscreen
                                </button>
                            </div>
                            <p style="color: #7f8c8d; margin: 15px 0 10px 0; font-size: 13px;">
                                Configuration file as it was on disk at the time of the run.
                            </p>
                            <div id="rawConfigContainer" style="position: relative; overflow-x: auto;">
                                <pre class="line-numbers" style="background: #2c3e50; border-radius: 4px; margin: 0;"><code id="rawConfigContent" class="language-toml" style="font-size: 12px;">{{.RawConfigContent}}</code></pre>
                            </div>
                        </div>
                    </div>
                    <link href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism-tomorrow.min.css" rel="stylesheet" />
                    <link href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.css" rel="stylesheet" />
                    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/prism.min.js"></script>
                    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-toml.min.js"></script>
                    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.js"></script>
                    <script>
                    (function() {
                        // Trigger Prism highlighting
                        Prism.highlightAll();
                    })();
                    
                    function toggleConfigFullscreen() {
                        const elem = document.getElementById('rawConfigContainer');
                        if (!document.fullscreenElement) {
                            elem.requestFullscreen().catch(err => {
                                alert('Error attempting to enable fullscreen: ' + err.message);
                            });
                        } else {
                            document.exitFullscreen();
                        }
                    }
                    </script>
                    <style>
                    /* Override Prism theme colors */
                    .line-numbers .line-numbers-rows {
                        border-right-color: #34495e !important;
                    }
                    .line-numbers-rows > span:before {
                        color: #7f8c8d !important;
                    }
                    /* Wrap indicator for continuation lines */
                    .line-numbers .line-numbers-rows > span {
                        position: relative;
                    }
                    #rawConfigContainer:fullscreen {
                        background: #2c3e50;
                        padding: 20px;
                        overflow: auto;
                    }
                    #rawConfigContainer:fullscreen pre {
                        font-size: 14px !important;
                        max-width: 100%;
                        height: 100%;
                    }
                    #rawConfigContainer:fullscreen code {
                        font-size: 14px !important;
                    }
                    </style>
                    {{end}}
                </div>
            </details>
        </div>
        {{end}}
        
        {{if .Phases}}
        <div class="section phase-flow-container" id="phaseFlow">
            <div class="phase-flow-header">
                <h2>🔄 Pipeline Flow</h2>
                <div style="display: flex; gap: 15px; align-items: center;">
                    <div class="slider-container">
                        <span class="slider-label">Density:</span>
                        <input type="range" min="1" max="4" value="4" class="column-slider" id="columnSlider" oninput="updateColumns()">
                    </div>
                    <button class="fullscreen-btn" onclick="toggleFullscreen()">⛶ Fullscreen</button>
                </div>
            </div>
            
            <div class="phase-flow-scroll">
                <div class="phase-flow">
                    {{range $phaseIndex, $phase := .Phases}}
                    {{if $phaseIndex}}<div class="phase-arrow">→</div>{{end}}
                    
                    <div class="phase-container">
                        <div class="phase-header">
                            <div class="phase-header-top">
                                <h3>{{phaseEmoji $phase.ID}} {{$phase.Name}}</h3>
                                <div class="phase-status">
                                    {{if eq $phase.Status "PASS"}}
                                    <span class="phase-status-icon success">✓</span>
                                    {{else}}
                                    <span class="phase-status-icon fail">✗</span>
                                    {{end}}
                                </div>
                            </div>
                            {{if $phase.Desc}}
                            <div class="phase-desc" title="{{$phase.Desc}}">{{truncate $phase.Desc 50}}</div>
                            {{end}}
                            <div class="phase-meta">
                                <span>{{$phase.TaskCount}} tasks in parallel</span>
                                <span><strong>{{formatDuration $phase.TotalMs}}</strong></span>
                            </div>
                        </div>
                        <div class="phase-tasks compact">
                            {{range $phase.Tasks}}
                            <div class="phase-task-card" data-task-id="{{.ID}}" onclick="scrollToTask('{{.ID}}')">
                                <div class="phase-task-card-header">
                                    {{if eq (string .Status) "PASS"}}
                                    <span class="phase-task-icon success">✓</span>
                                    {{else if eq (string .Status) "FAIL"}}
                                    <span class="phase-task-icon fail">✗</span>
                                    {{else}}
                                    <span class="phase-task-icon skip">⊘</span>
                                    {{end}}
                                    <span class="phase-task-name" title="{{.Name}}{{if .Metrics}}{{if eq .Metrics.SummaryFormat "junit"}} 🧪{{else if eq .Metrics.SummaryFormat "sarif"}} 🔒{{else if eq .Metrics.SummaryFormat "artifact"}} 📦{{end}}{{end}}">{{truncate .Name 25}}{{if .Metrics}}{{if eq .Metrics.SummaryFormat "junit"}} 🧪{{else if eq .Metrics.SummaryFormat "sarif"}} 🔒{{else if eq .Metrics.SummaryFormat "artifact"}} 📦{{end}}{{end}}</span>
                                    <span class="phase-task-duration">{{formatDuration .DurationMs}}</span>
                                </div>
                                {{if .Desc}}
                                <div class="phase-task-desc" title="{{.Desc}}">{{truncate .Desc 80}}</div>
                                {{else}}
                                <div class="phase-task-desc">{{.ID}}</div>
                                {{end}}
                                {{if .Type}}
                                <span class="phase-task-type">{{.Type}}</span>
                                {{end}}
                            </div>
                            {{end}}
                        </div>
                    </div>
                    {{end}}
                </div>
            </div>
        </div>
        {{end}}
        
        <div class="section">
            <h2>Tasks ({{len .TasksWithLogs}})</h2>
            {{range .TasksWithLogs}}
            <div class="task-card">
                <div class="task-header">
                    <div>
                        <span class="task-title">{{.Name}}</span>
                        <span class="task-id">({{.ID}})</span>
                    </div>
                    <span class="badge badge-{{.Status | string | statusClass}}">
                        {{.Status | string | statusSymbol}} {{.Status}}
                    </span>
                </div>
                
                <div class="task-details">
                    <div class="detail-item">
                        <div class="detail-label">Type</div>
                        <div class="detail-value">{{.Type}}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Duration</div>
                        <div class="detail-value">{{formatDuration .DurationMs}}</div>
                    </div>
                    {{if .ExitCode}}
                    <div class="detail-item">
                        <div class="detail-label">Exit Code</div>
                        <div class="detail-value">
                            {{$exitCode := deref .ExitCode}}
                            {{if eq $exitCode 0}}
                            <span class="exit-code-success">0</span>
                            {{else}}
                            <span class="exit-code-error">{{$exitCode}}</span>
                            {{end}}
                        </div>
                    </div>
                    {{end}}
                    {{if .AutoFixed}}
                    <div class="detail-item">
                        <div class="detail-label">Auto-Fixed</div>
                        <div class="detail-value">
                            <span class="badge" style="background: #d4edda; color: #155724;">🔧 Yes</span>
                        </div>
                    </div>
                    {{if .InitialExitCode}}
                    <div class="detail-item">
                        <div class="detail-label">Initial Exit Code</div>
                        <div class="detail-value">
                            <span class="exit-code-error">{{deref .InitialExitCode}}</span>
                        </div>
                    </div>
                    {{end}}
                    <div class="detail-item">
                        <div class="detail-label">Fix Duration</div>
                        <div class="detail-value">{{formatDuration .FixDurationMs}}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Recheck Duration</div>
                        <div class="detail-value">{{formatDuration .RecheckDurationMs}}</div>
                    </div>
                    {{end}}
                    <div class="detail-item">
                        <div class="detail-label">Start Time ({{$.Timezone}})</div>
                        <div class="detail-value">{{formatTime .StartTime}}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">End Time ({{$.Timezone}})</div>
                        <div class="detail-value">{{formatTime .EndTime}}</div>
                    </div>
                </div>
                
                <div class="detail-item" style="margin-top: 15px;">
                    <div class="detail-label">Command</div>
                    <div class="detail-value mono">{{.Command}}</div>
                </div>
                
                {{if .FixCommand}}
                <div class="detail-item" style="margin-top: 10px;">
                    <div class="detail-label">Fix Command</div>
                    <div class="detail-value mono">{{.FixCommand}}</div>
                </div>
                {{end}}
                
                {{if .Metrics}}
                <div class="metrics-box">
                    {{if eq .Metrics.SummaryFormat "artifact"}}
                    <div class="metrics-title">📦 Build Artifact</div>
                    <div class="metrics-grid">
                        <div class="detail-item">
                            <div class="detail-label">File Path</div>
                            <div class="detail-value mono" style="font-size: 11px;">{{index .Metrics.Data "path"}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">File Size</div>
                            <div class="detail-value" style="font-weight: bold;">{{index .Metrics.Data "size"}} bytes</div>
                        </div>
                    </div>
                    {{else if eq .Metrics.SummaryFormat "sarif"}}
                    <div class="metrics-title">🔒 Security Scan Results (SARIF)</div>
                    <div class="metrics-grid">
                        <div class="detail-item">
                            <div class="detail-label">Total Issues</div>
                            <div class="detail-value" style="font-size: 18px; font-weight: bold; color: {{if gt (index .Metrics.Data "total") 0}}#e74c3c{{else}}#27ae60{{end}};">{{index .Metrics.Data "total"}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Errors</div>
                            <div class="detail-value" style="color: {{if gt (index .Metrics.Data "errors") 0}}#e74c3c{{else}}#95a5a6{{end}}; font-weight: bold;">{{index .Metrics.Data "errors"}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Warnings</div>
                            <div class="detail-value" style="color: {{if gt (index .Metrics.Data "warnings") 0}}#f39c12{{else}}#95a5a6{{end}}; font-weight: bold;">{{index .Metrics.Data "warnings"}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Notes</div>
                            <div class="detail-value" style="color: #3498db;">{{index .Metrics.Data "notes"}}</div>
                        </div>
                    </div>
                    {{$findings := index .Metrics.Data "findings"}}
                    {{if $findings}}
                    <details style="margin-top: 15px;" {{if gt (index .Metrics.Data "total") 0}}open{{end}}>
                        <summary style="cursor: pointer; color: #3498db; font-weight: 600; user-select: none;">
                            🔍 View Security Findings
                        </summary>
                        <div style="margin-top: 10px; padding: 10px; background: white; border-radius: 4px; font-size: 13px;">
                            {{$rules := index .Metrics.Data "rules"}}
                            {{if $rules}}
                            <div style="margin-bottom: 15px; padding-bottom: 15px; border-bottom: 1px solid #dee2e6;">
                                <strong style="display: block; margin-bottom: 10px;">Issues by Rule:</strong>
                                <div style="display: flex; flex-wrap: wrap; gap: 8px;">
                                    {{range $rules}}
                                    <span style="padding: 4px 8px; background: #f8f9fa; border-radius: 3px; font-size: 11px; border-left: 3px solid #e74c3c;">
                                        <strong>{{.id}}</strong>: {{.count}} issue{{if gt .count 1}}s{{end}}
                                        {{if .severity}}<span style="color: #95a5a6; margin-left: 4px;">(severity: {{.severity}})</span>{{end}}
                                    </span>
                                    {{end}}
                                </div>
                            </div>
                            {{end}}
                            <div style="max-height: 500px; overflow-y: auto;">
                                {{range $findings}}
                                <div style="padding: 10px; margin-bottom: 8px; background: #f8f9fa; border-left: 4px solid {{if eq .level "error"}}#e74c3c{{else if eq .level "warning"}}#f39c12{{else}}#3498db{{end}}; border-radius: 3px; font-size: 12px;">
                                    <div style="display: flex; justify-content: space-between; align-items: start; margin-bottom: 6px;">
                                        <div style="flex: 1;">
                                            <span style="font-weight: 600; color: #2c3e50;">{{.ruleId}}</span>
                                            {{if .ruleName}}
                                            <span style="color: #7f8c8d; font-size: 11px; margin-left: 5px;">({{.ruleName}})</span>
                                            {{end}}
                                        </div>
                                        <span style="color: {{if eq .level "error"}}#e74c3c{{else if eq .level "warning"}}#f39c12{{else}}#3498db{{end}}; font-weight: bold; font-size: 11px; white-space: nowrap; margin-left: 10px;">
                                            {{if eq .level "error"}}❌{{else if eq .level "warning"}}⚠️{{else}}ℹ️{{end}} {{.level}}
                                        </span>
                                    </div>
                                    <div style="color: #2c3e50; margin-bottom: 4px;">{{.message}}</div>
                                    <div style="color: #7f8c8d; font-size: 11px; font-family: 'Monaco', 'Menlo', monospace;">
                                        📄 {{.file}}:{{.line}}{{if .column}}:{{.column}}{{end}}
                                    </div>
                                    {{if .shortDesc}}
                                    <div style="color: #7f8c8d; font-size: 11px; margin-top: 4px; font-style: italic;">{{.shortDesc}}</div>
                                    {{end}}
                                    {{if .severity}}
                                    <div style="margin-top: 4px;">
                                        <span style="background: #e74c3c; color: white; padding: 2px 6px; border-radius: 3px; font-size: 10px; font-weight: bold;">
                                            CVSS: {{.severity}}
                                        </span>
                                    </div>
                                    {{end}}
                                    {{if .tags}}
                                    <div style="margin-top: 6px; display: flex; flex-wrap: wrap; gap: 4px;">
                                        {{range .tags}}
                                        <span style="background: #ecf0f1; color: #7f8c8d; padding: 2px 6px; border-radius: 3px; font-size: 10px;">{{.}}</span>
                                        {{end}}
                                    </div>
                                    {{end}}
                                    {{if .dataFlow}}
                                    <details style="margin-top: 8px;">
                                        <summary style="cursor: pointer; color: #3498db; font-size: 11px; user-select: none;">
                                            🔄 Data Flow ({{len .dataFlow}} steps)
                                        </summary>
                                        <div style="margin-top: 6px; padding-left: 10px; border-left: 2px solid #dee2e6;">
                                            {{range $index, $step := .dataFlow}}
                                            <div style="margin: 4px 0; font-size: 11px; color: #7f8c8d;">
                                                {{add $index 1}}. {{$step.file}}:{{$step.line}}{{if $step.column}}:{{$step.column}}{{end}}
                                                {{if $step.message}}<span style="color: #95a5a6;"> - {{$step.message}}</span>{{end}}
                                            </div>
                                            {{end}}
                                        </div>
                                    </details>
                                    {{end}}
                                </div>
                                {{end}}
                            </div>
                        </div>
                    </details>
                    {{end}}
                    {{else if eq .Metrics.SummaryFormat "junit"}}
                    <div class="metrics-title">🧪 Test Results (JUnit)</div>
                    <div class="metrics-grid">
                        <div class="detail-item">
                            <div class="detail-label">Total Tests</div>
                            <div class="detail-value" style="font-size: 18px; font-weight: bold;">{{index .Metrics.Data "tests"}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Passed</div>
                            {{$total := index .Metrics.Data "tests"}}
                            {{$failures := index .Metrics.Data "failures"}}
                            {{$errors := index .Metrics.Data "errors"}}
                            {{$skipped := index .Metrics.Data "skipped"}}
                            {{$passed := sub (sub (sub $total $failures) $errors) $skipped}}
                            <div class="detail-value" style="color: #27ae60; font-weight: bold;">{{$passed}} ✓</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Failed</div>
                            <div class="detail-value" style="color: {{if gt (index .Metrics.Data "failures") 0.0}}#e74c3c{{else}}#95a5a6{{end}}; font-weight: bold;">{{index .Metrics.Data "failures"}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Errors</div>
                            <div class="detail-value" style="color: {{if gt (index .Metrics.Data "errors") 0.0}}#e74c3c{{else}}#95a5a6{{end}}; font-weight: bold;">{{index .Metrics.Data "errors"}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Skipped</div>
                            <div class="detail-value" style="color: #f39c12;">{{index .Metrics.Data "skipped"}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Duration</div>
                            <div class="detail-value">{{printf "%.2f" (index .Metrics.Data "time")}}s</div>
                        </div>
                    </div>
                    <details style="margin-top: 15px;" {{if or (gt $failures 0.0) (gt $errors 0.0)}}open{{end}}>
                        <summary style="cursor: pointer; color: #3498db; font-weight: 600; user-select: none;">
                            📋 View Detailed Breakdown
                        </summary>
                        <div style="margin-top: 10px; padding: 10px; background: white; border-radius: 4px; font-size: 13px;">
                            <div style="margin-bottom: 8px;">
                                <strong>Pass Rate:</strong> 
                                {{$passRate := mul (div (sub (sub (sub $total $failures) $errors) $skipped) $total) 100.0}}
                                <span style="color: {{if ge $passRate 80.0}}#27ae60{{else if ge $passRate 50.0}}#f39c12{{else}}#e74c3c{{end}}; font-weight: bold;">
                                    {{printf "%.1f" $passRate}}%
                                </span>
                            </div>
                            <div style="margin-bottom: 8px;">
                                <strong>Status Distribution:</strong>
                                <div style="display: flex; gap: 10px; margin-top: 5px;">
                                    <span style="color: #27ae60;">✓ Passed: {{$passed}}</span>
                                    <span style="color: #e74c3c;">✗ Failed: {{$failures}}</span>
                                    <span style="color: #e74c3c;">⚠ Errors: {{$errors}}</span>
                                    <span style="color: #f39c12;">⊘ Skipped: {{$skipped}}</span>
                                </div>
                            </div>
                            <div style="margin-bottom: 15px;">
                                <strong>Average Test Duration:</strong> 
                                {{$avgDuration := div (index .Metrics.Data "time") $total}}
                                {{printf "%.3f" $avgDuration}}s per test
                            </div>
                            {{$testcases := index .Metrics.Data "testcases"}}
                            {{if $testcases}}
                            <div style="border-top: 1px solid #dee2e6; padding-top: 15px;">
                                <strong style="display: block; margin-bottom: 10px;">Individual Test Cases:</strong>
                                <div style="max-height: 400px; overflow-y: auto;">
                                    {{range $testcases}}
                                    <div style="padding: 8px; margin-bottom: 6px; background: #f8f9fa; border-left: 3px solid {{if eq .status "passed"}}#27ae60{{else if eq .status "failed"}}#e74c3c{{else if eq .status "error"}}#e74c3c{{else}}#f39c12{{end}}; border-radius: 3px; font-size: 12px;">
                                        <div style="display: flex; justify-content: space-between; align-items: start;">
                                            <div style="flex: 1;">
                                                <span style="font-weight: 600; color: #2c3e50;">{{.name}}</span>
                                                {{if .classname}}
                                                <span style="color: #7f8c8d; font-size: 11px; margin-left: 5px;">({{.classname}})</span>
                                                {{end}}
                                                {{if .message}}
                                                <div style="color: #e74c3c; margin-top: 4px; font-size: 11px;">{{.message}}</div>
                                                {{end}}
                                            </div>
                                            <div style="display: flex; align-items: center; gap: 8px; margin-left: 10px;">
                                                <span style="color: {{if eq .status "passed"}}#27ae60{{else if eq .status "failed"}}#e74c3c{{else if eq .status "error"}}#e74c3c{{else}}#f39c12{{end}}; font-weight: bold; white-space: nowrap;">
                                                    {{if eq .status "passed"}}✓{{else if eq .status "failed"}}✗{{else if eq .status "error"}}⚠{{else}}⊘{{end}} {{.status}}
                                                </span>
                                                <span style="color: #95a5a6; font-size: 11px; white-space: nowrap;">{{printf "%.3f" .time}}s</span>
                                            </div>
                                        </div>
                                    </div>
                                    {{end}}
                                </div>
                            </div>
                            {{end}}
                        </div>
                    </details>
                    {{else}}
                    <div class="metrics-title">📊 Metrics ({{.Metrics.SummaryFormat}})</div>
                    <div class="metrics-grid">
                        {{range $key, $value := .Metrics.Data}}
                        <div class="detail-item">
                            <div class="detail-label">{{$key}}</div>
                            <div class="detail-value">{{$value}}</div>
                        </div>
                        {{end}}
                    </div>
                    {{end}}
                </div>
                {{end}}
                
                {{if .LogPath}}
                <div class="detail-item" style="margin-top: 15px;">
                    <div class="detail-label">Output (last 10 lines)</div>
                    <pre style="background: #2c3e50; color: #ecf0f1; padding: 15px; border-radius: 4px; overflow-x: auto; font-size: 12px; line-height: 1.5;">{{range .LogPreview}}{{.}}
{{end}}</pre>
                    <div style="display: flex; gap: 15px; margin-top: 10px;">
                        <a href="logs/{{.ID}}.log" class="log-link">📄 View raw log</a>
                        <a href="ide.html?file=logs/{{.ID}}.log" class="log-link">🖥️ View in web IDE</a>
                    </div>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
    </div>
    
    <script>
        function updateColumns() {
            const slider = document.getElementById('columnSlider');
            const phaseTasks = document.querySelectorAll('.phase-tasks');
            
            const maxCols = parseInt(slider.value);
            
            // Update each phase based on its task count
            phaseTasks.forEach(tasks => {
                const taskCount = tasks.querySelectorAll('.phase-task-card').length;
                
                // Calculate actual columns needed
                // Use min(maxCols, ceil(taskCount / 5)) to determine columns
                const neededCols = Math.min(maxCols, Math.ceil(taskCount / 5));
                
                // Calculate rows per column (distribute evenly)
                const rowsPerCol = Math.ceil(taskCount / neededCols);
                
                // Set grid properties
                const colWidth = 280;
                const gap = 8;
                const padding = 12;
                
                // Width = padding-left + (colWidth * cols) + (gap * (cols - 1)) + padding-right
                // Simplifies to: (colWidth * cols) + (gap * (cols - 1)) + (padding * 2)
                const gridContentWidth = (colWidth * neededCols) + (gap * (neededCols - 1));
                const totalWidth = gridContentWidth + (padding * 2);
                
                tasks.style.gridTemplateColumns = 'repeat(' + neededCols + ', ' + colWidth + 'px)';
                tasks.style.gridTemplateRows = 'repeat(' + rowsPerCol + ', auto)';
                tasks.style.width = totalWidth + 'px';
                tasks.style.gap = gap + 'px';
                tasks.style.padding = padding + 'px';
            });
        }
        
        // Initialize on page load
        document.addEventListener('DOMContentLoaded', updateColumns);
        
        function toggleFullscreen() {
            const container = document.getElementById('phaseFlow');
            const fullscreenBtns = container.querySelectorAll('.fullscreen-btn');
            
            container.classList.toggle('fullscreen');
            
            // Find the fullscreen button (not the compact button)
            fullscreenBtns.forEach(btn => {
                if (btn.id !== 'compactBtn') {
                    if (container.classList.contains('fullscreen')) {
                        btn.textContent = '✕ Exit Fullscreen';
                    } else {
                        btn.textContent = '⛶ Fullscreen';
                    }
                }
            });
        }
        
        function scrollToTask(taskId) {
            // Exit fullscreen if active
            const container = document.getElementById('phaseFlow');
            const wasFullscreen = container && container.classList.contains('fullscreen');
            
            if (wasFullscreen) {
                container.classList.remove('fullscreen');
                // Update the fullscreen button (not the slider)
                const fullscreenBtns = container.querySelectorAll('.fullscreen-btn');
                fullscreenBtns.forEach(btn => {
                    if (btn.id !== 'compactBtn') {
                        btn.textContent = '⛶ Fullscreen';
                    }
                });
            }
            
            // Function to find and scroll to task
            const doScroll = () => {
                const taskCards = document.querySelectorAll('.task-card');
                for (const card of taskCards) {
                    const titleEl = card.querySelector('.task-id');
                    if (titleEl && titleEl.textContent.includes(taskId)) {
                        card.scrollIntoView({ behavior: 'smooth', block: 'start' });
                        // Highlight the card briefly
                        card.style.boxShadow = '0 0 0 3px #3498db';
                        setTimeout(() => {
                            card.style.boxShadow = '';
                        }, 2000);
                        break;
                    }
                }
            };
            
            // If we exited fullscreen, wait for DOM to reflow before scrolling
            if (wasFullscreen) {
                setTimeout(doScroll, 100);
            } else {
                doScroll();
            }
        }
        
        // Mascot keyboard toggle (h = hide, s = show)
        (function() {
            const mascot = document.querySelector('.mascot');
            
            document.addEventListener('keydown', function(e) {
                // Ignore if user is typing in an input field
                if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
                    return;
                }
                
                if (e.key === 'h' || e.key === 'H') {
                    mascot.classList.add('hidden');
                } else if (e.key === 's' || e.key === 'S') {
                    mascot.classList.remove('hidden');
                }
            });
        })();
        
        // Mascot eye tracking
        (function() {
            const pupils = document.querySelectorAll('.mascot-pupil');
            const eyes = document.querySelectorAll('.mascot-eye');
            
            // Movement bounds for 200px mascot (scaled from 777px original, accounting for aspect ratio)
            const leftMaxDistanceX = 1.14;
            const leftMaxDistanceY = 4.55;
            const rightMaxDistanceX = 2.84;
            const rightMaxDistanceY = 4.55;
            
            document.addEventListener('mousemove', function(e) {
                pupils.forEach(function(pupil, index) {
                    const eye = eyes[index];
                    const eyeRect = eye.getBoundingClientRect();
                    
                    // Get eye center position
                    const eyeCenterX = eyeRect.left + eyeRect.width / 2;
                    const eyeCenterY = eyeRect.top + eyeRect.height / 2;
                    
                    // Calculate angle to mouse
                    const angleRad = Math.atan2(e.clientY - eyeCenterY, e.clientX - eyeCenterX);
                    
                    // Get max distance for this eye
                    const maxDistanceX = index === 0 ? leftMaxDistanceX : rightMaxDistanceX;
                    const maxDistanceY = index === 0 ? leftMaxDistanceY : rightMaxDistanceY;
                    
                    // Calculate pupil position
                    let pupilX = Math.cos(angleRad) * maxDistanceX;
                    let pupilY = Math.sin(angleRad) * maxDistanceY;
                    
                    // Constrain to ellipse
                    const normalizedX = pupilX / maxDistanceX;
                    const normalizedY = pupilY / maxDistanceY;
                    const distanceFromCenter = Math.sqrt(normalizedX * normalizedX + normalizedY * normalizedY);
                    
                    if (distanceFromCenter > 1) {
                        pupilX = (normalizedX / distanceFromCenter) * maxDistanceX;
                        pupilY = (normalizedY / distanceFromCenter) * maxDistanceY;
                    }
                    
                    // Apply transform
                    pupil.style.transform = 'translate(' + pupilX + 'px, ' + pupilY + 'px)';
                });
            });
        })();
    </script>
</body>
</html>
`
