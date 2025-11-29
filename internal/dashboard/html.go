package dashboard

import (
	"fmt"
	"html/template"
	"os"
	"time"
	
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
	defer f.Close()
	
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

// writeRunDetailHTML generates a detail page for a single run
func writeRunDetailHTML(path string, run model.RunRecord) error {
	// Prepare data with log previews
	type TaskWithLog struct {
		model.TaskResult
		LogPreview   []string
		ArtifactPath string
		ArtifactSize int64
	}
	
	type DetailData struct {
		model.RunRecord
		TasksWithLogs []TaskWithLog
		Timezone      string
	}
	
	data := DetailData{
		RunRecord:     run,
		TasksWithLogs: make([]TaskWithLog, 0, len(run.Tasks)),
		Timezone:      getLocalTimezone(),
	}
	
	// Load log previews and artifact info for each task
	for _, task := range run.Tasks {
		taskWithLog := TaskWithLog{
			TaskResult: task,
			LogPreview: readLastLines(task.LogPath, 10),
		}
		
		// Check for artifact file (stored in metrics for artifact format)
		if task.Metrics != nil && task.Metrics.SummaryFormat == "artifact" {
			// The artifact path should be in the workdir + metrics path from task definition
			// We need to reconstruct it from the run record
			// For now, check if we can get it from the task's workdir
		}
		
		data.TasksWithLogs = append(data.TasksWithLogs, taskWithLog)
	}
	
	tmpl, err := template.New("rundetail").Funcs(template.FuncMap{
		"formatDuration": formatDuration,
		"formatTime":     formatTime,
		"statusClass":    statusClass,
		"statusSymbol":   statusSymbol,
		"string":         func(s model.TaskStatus) string { return string(s) },
		"deref":          func(i *int) int { if i != nil { return *i }; return 0 },
	}).Parse(runDetailTemplate)
	
	if err != nil {
		return err
	}
	
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	
	return tmpl.Execute(f, data)
}

// readLastLines reads the last N lines from a file
func readLastLines(path string, n int) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{"Error reading log file"}
	}
	
	lines := []string{}
	for _, line := range []byte(string(data)) {
		if line == '\n' {
			lines = append(lines, "")
		}
	}
	
	// Split by newlines properly
	allLines := []string{}
	currentLine := ""
	for _, b := range data {
		if b == '\n' {
			allLines = append(allLines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(b)
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
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        
        header {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }
        
        h1 {
            font-size: 32px;
            margin-bottom: 10px;
            color: #2c3e50;
        }
        
        .subtitle {
            color: #7f8c8d;
            font-size: 14px;
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
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>📊 devpipe Dashboard</h1>
            <div class="subtitle">Last updated: {{formatTime .LastGenerated}}</div>
        </header>
        
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-value">{{.TotalRuns}}</div>
                <div class="stat-label">Total Runs</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{{len .TaskStats}}</div>
                <div class="stat-label">Tasks</div>
            </div>
        </div>
        
        <div class="section">
            <h2>Recent Runs</h2>
            {{if .RecentRuns}}
            <table>
                <thead>
                    <tr>
                        <th>Run ID</th>
                        <th>Timestamp ({{.Timezone}})</th>
                        <th>Status</th>
                        <th>Duration</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .RecentRuns}}
                    <tr>
                        <td class="mono"><a href="runs/{{.RunID}}/report.html">{{.RunID}}</a></td>
                        <td>{{formatTime .Timestamp}}</td>
                        <td>
                            <span class="badge badge-{{.Status | statusClass}}">
                                {{statusSymbol .Status}} {{.Status}}
                            </span>
                        </td>
                        <td>{{formatDuration .Duration}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            {{else}}
            <div class="empty-state">
                <div class="empty-state-icon">📭</div>
                <p>No runs yet. Run devpipe to see results here!</p>
            </div>
            {{end}}
        </div>
        
        <div class="section">
            <h2>Task Statistics</h2>
            {{if .TaskStats}}
            <table>
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
            {{else}}
            <div class="empty-state">
                <div class="empty-state-icon">📊</div>
                <p>No task statistics available yet.</p>
            </div>
            {{end}}
        </div>
    </div>
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
            max-width: 1200px;
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
            align-items: center;
            margin-bottom: 15px;
        }
        
        .task-title {
            font-size: 18px;
            font-weight: 600;
            color: #2c3e50;
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
    </style>
</head>
<body>
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
                    <div class="meta-label">Repo Root</div>
                    <div class="meta-value mono">{{.RepoRoot}}</div>
                </div>
            </div>
        </header>
        
        <div class="section">
            <h2>Run Configuration</h2>
            {{if .Command}}
            <div class="detail-item" style="margin-bottom: 15px;">
                <div class="detail-label">Command</div>
                <div class="detail-value mono" style="font-size: 11px; background: #f8f9fa; padding: 8px; border-radius: 4px;">{{.Command}}</div>
            </div>
            {{end}}
            <div class="task-details">
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
        
        <div class="section">
            <h2>Git Information</h2>
            <div style="display: grid; grid-template-columns: 300px 1fr; gap: 30px;">
                <div>
                    <div class="task-details" style="grid-template-columns: 1fr;">
                        <div class="detail-item">
                            <div class="detail-label">Mode</div>
                            <div class="detail-value">{{.Git.mode}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Reference</div>
                            <div class="detail-value mono">{{.Git.ref}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Changed Files</div>
                            <div class="detail-value">{{len .Git.changedFiles}} files</div>
                        </div>
                    </div>
                </div>
                {{if .Git.changedFiles}}
                <div>
                    <div class="detail-label">Files Changed</div>
                    <div class="detail-value mono" style="font-size: 12px; line-height: 1.8; margin-top: 5px;">
                        {{range .Git.changedFiles}}
                        <div>{{.}}</div>
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        
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
                    {{else if eq .Metrics.SummaryFormat "junit"}}
                    <div class="metrics-title">🧪 Test Results (JUnit)</div>
                    <div class="metrics-grid">
                        <div class="detail-item">
                            <div class="detail-label">Total Tests</div>
                            <div class="detail-value" style="font-size: 18px; font-weight: bold;">{{index .Metrics.Data "tests"}}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Passed</div>
                            <div class="detail-value" style="color: #27ae60; font-weight: bold;">{{index .Metrics.Data "tests"}} ✓</div>
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
                            <div class="detail-value">{{index .Metrics.Data "time"}}s</div>
                        </div>
                    </div>
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
                    <a href="logs/{{.ID}}.log" class="log-link">📄 View Full Log</a>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
    </div>
</body>
</html>
`
