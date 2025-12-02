package dashboard

import (
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"

	"github.com/acarl005/stripansi"
)

// FileInfo represents a file in the IDE file tree
type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Content string `json:"content"`
}

// writeIDEViewer generates the IDE viewer HTML page for a run
func writeIDEViewer(path, runID, runDir string) error {
	// Collect files list
	files := collectFiles(runDir)

	// Convert to JSON string
	filesJSON, err := json.Marshal(files)
	if err != nil {
		return err
	}

	type IDEData struct {
		RunID     string
		FilesJSON template.JS
	}

	data := IDEData{
		RunID:     runID,
		FilesJSON: template.JS(filesJSON),
	}

	tmpl, err := template.New("ide").Parse(ideTemplate)
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

// collectFiles gathers all files from the run directory
func collectFiles(runDir string) []FileInfo {
	var files []FileInfo

	// Add pipeline.log if exists
	pipelineLog := filepath.Join(runDir, "pipeline.log")
	if info, err := os.Stat(pipelineLog); err == nil {
		content, _ := os.ReadFile(pipelineLog)
		files = append(files, FileInfo{
			Name:    "pipeline.log",
			Path:    "pipeline.log",
			Size:    info.Size(),
			Content: stripansi.Strip(string(content)),
		})
	}

	// Add config.toml if exists
	configFile := filepath.Join(runDir, "config.toml")
	if info, err := os.Stat(configFile); err == nil {
		content, _ := os.ReadFile(configFile)
		files = append(files, FileInfo{
			Name:    "config.toml",
			Path:    "config.toml",
			Size:    info.Size(),
			Content: stripansi.Strip(string(content)),
		})
	}

	// Add run.json if exists
	runJSON := filepath.Join(runDir, "run.json")
	if info, err := os.Stat(runJSON); err == nil {
		content, _ := os.ReadFile(runJSON)
		files = append(files, FileInfo{
			Name:    "run.json",
			Path:    "run.json",
			Size:    info.Size(),
			Content: string(content), // JSON doesn't need ANSI stripping
		})
	}

	// Add all log files
	logsDir := filepath.Join(runDir, "logs")
	if entries, err := os.ReadDir(logsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				info, _ := entry.Info()
				logPath := filepath.Join(logsDir, entry.Name())
				content, _ := os.ReadFile(logPath)
				files = append(files, FileInfo{
					Name:    entry.Name(),
					Path:    "logs/" + entry.Name(),
					Size:    info.Size(),
					Content: stripansi.Strip(string(content)),
				})
			}
		}
	}

	// Add all artifact files
	artifactsDir := filepath.Join(runDir, "artifacts")
	if entries, err := os.ReadDir(artifactsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				info, _ := entry.Info()
				artifactPath := filepath.Join(artifactsDir, entry.Name())
				content, _ := os.ReadFile(artifactPath)
				files = append(files, FileInfo{
					Name:    entry.Name(),
					Path:    "artifacts/" + entry.Name(),
					Size:    info.Size(),
					Content: stripansi.Strip(string(content)),
				})
			}
		}
	}

	return files
}

const ideTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DevPipe IDE - {{.RunID}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            height: 100vh;
            overflow: hidden;
            background: #1e1e1e;
            color: #d4d4d4;
        }
        .header {
            background: #2d2d30;
            padding: 10px 20px;
            border-bottom: 1px solid #3e3e42;
            display: flex;
            align-items: center;
            gap: 20px;
        }
        .back-btn {
            background: #0e639c;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            text-decoration: none;
            font-size: 14px;
        }
        .back-btn:hover {
            background: #1177bb;
        }
        .title {
            font-size: 14px;
            color: #cccccc;
        }
        .container {
            display: flex;
            height: calc(100vh - 45px);
        }
        .sidebar {
            width: 250px;
            background: #252526;
            border-right: 1px solid #3e3e42;
            overflow-y: auto;
            display: flex;
            flex-direction: column;
        }
        .sidebar-header {
            padding: 10px 15px;
            font-size: 11px;
            text-transform: uppercase;
            color: #888;
            font-weight: 600;
        }
        .search-box {
            padding: 10px 15px;
            border-bottom: 1px solid #3e3e42;
        }
        .search-input {
            width: 100%;
            background: #3c3c3c;
            border: 1px solid #3e3e42;
            color: #cccccc;
            padding: 6px 10px;
            border-radius: 4px;
            font-size: 13px;
        }
        .search-input:focus {
            outline: none;
            border-color: #007acc;
        }
        .search-results {
            padding: 10px 15px;
            font-size: 12px;
            color: #888;
        }
        .search-result-item {
            padding: 6px 10px;
            margin: 4px 0;
            background: #2d2d30;
            border-radius: 3px;
            cursor: pointer;
            font-size: 12px;
        }
        .search-result-item:hover {
            background: #3e3e42;
        }
        .search-result-file {
            color: #4ec9b0;
            font-weight: 500;
        }
        .search-result-line {
            color: #888;
            margin-left: 5px;
        }
        .search-result-preview {
            color: #d4d4d4;
            margin-top: 3px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        .search-match {
            background: #515c6a;
            color: #f9f9f9;
        }
        .file-tree {
            list-style: none;
        }
        .file-item, .folder-item {
            padding: 4px 15px;
            cursor: pointer;
            font-size: 13px;
            display: flex;
            align-items: center;
            gap: 6px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
        .file-item:hover, .folder-item:hover {
            background: #2a2d2e;
        }
        .file-item.active {
            background: #094771;
        }
        .folder-item {
            font-weight: 500;
        }
        .folder-children {
            padding-left: 20px;
        }
        .folder-icon {
            font-size: 12px;
        }
        .file-icon {
            font-size: 12px;
        }
        .editor-container {
            flex: 1;
            display: flex;
            flex-direction: column;
        }
        .editor-tabs {
            background: #2d2d30;
            border-bottom: 1px solid #3e3e42;
            padding: 0 10px;
            display: flex;
            align-items: center;
            min-height: 35px;
        }
        .editor-tab {
            padding: 8px 12px;
            font-size: 13px;
            color: #969696;
            background: transparent;
            border: none;
            cursor: pointer;
        }
        .editor-tab.active {
            color: #ffffff;
            background: #1e1e1e;
        }
        #editor {
            flex: 1;
        }
        .loading {
            display: flex;
            align-items: center;
            justify-content: center;
            height: 100%;
            color: #888;
        }
    </style>
</head>
<body>
    <div class="header">
        <a href="report.html" class="back-btn">← Back to Dashboard</a>
        <span class="title">DevPipe IDE - Run {{.RunID}}</span>
    </div>
    
    <div class="container">
        <div class="sidebar">
            <div class="search-box">
                <input type="text" class="search-input" id="searchInput" placeholder="🔍 Search in files...">
            </div>
            <div id="searchResults" class="search-results" style="display: none;"></div>
            <div id="fileTreeContainer">
                <div class="sidebar-header">Files</div>
                <ul class="file-tree" id="fileTree">
                    <li class="loading">Loading files...</li>
                </ul>
            </div>
        </div>
        
        <div class="editor-container">
            <div class="editor-tabs" id="editorTabs">
                <span style="color: #888; font-size: 13px;">No file open</span>
            </div>
            <div id="editor"></div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs/loader.js"></script>
    <script>
        let editor;
        let currentFile = null;
        let files = {{.FilesJSON}};  // Embedded file contents
        
        require.config({ paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs' }});
        
        require(['vs/editor/editor.main'], function() {
            // Initialize Monaco Editor
            editor = monaco.editor.create(document.getElementById('editor'), {
                value: '',
                theme: 'vs-dark',
                readOnly: true,
                automaticLayout: true,
                minimap: { enabled: true },
                fontSize: 13,
                lineNumbers: 'on',
                wordWrap: 'off',
                scrollBeyondLastLine: false,
            });
            
            // Load embedded file tree
            renderFileTree(files);
            
            // Check URL for initial file to open
            const params = new URLSearchParams(window.location.search);
            const fileToOpen = params.get('file');
            if (fileToOpen) {
                setTimeout(() => openFile(fileToOpen), 500);
            }
            
            // Setup search
            setupSearch();
        });
        
        
        function renderFileTree(files) {
            const tree = document.getElementById('fileTree');
            tree.innerHTML = '';
            
            // Group files by directory
            const structure = {
                'root': [],
                'logs': [],
                'artifacts': []
            };
            
            files.forEach(file => {
                if (file.path.startsWith('logs/')) {
                    structure.logs.push(file);
                } else if (file.path.startsWith('artifacts/')) {
                    structure.artifacts.push(file);
                } else {
                    structure.root.push(file);
                }
            });
            
            // Render root files
            structure.root.forEach(file => {
                tree.appendChild(createFileItem(file));
            });
            
            // Render logs folder
            if (structure.logs.length > 0) {
                const logsFolder = createFolderItem('📁 logs', structure.logs);
                tree.appendChild(logsFolder);
            }
            
            // Render artifacts folder
            if (structure.artifacts.length > 0) {
                const artifactsFolder = createFolderItem('📦 artifacts', structure.artifacts);
                tree.appendChild(artifactsFolder);
            }
        }
        
        function createFileItem(file) {
            const li = document.createElement('li');
            li.className = 'file-item';
            li.innerHTML = getFileIcon(file.path) + ' ' + file.name;
            li.onclick = () => openFile(file.path);
            li.dataset.path = file.path;
            return li;
        }
        
        function createFolderItem(label, files) {
            const container = document.createElement('div');
            
            const folderHeader = document.createElement('li');
            folderHeader.className = 'folder-item';
            folderHeader.innerHTML = label;
            
            const children = document.createElement('ul');
            children.className = 'folder-children';
            children.style.listStyle = 'none';
            files.forEach(file => {
                children.appendChild(createFileItem(file));
            });
            
            folderHeader.onclick = (e) => {
                e.stopPropagation();
                const isHidden = children.style.display === 'none';
                children.style.display = isHidden ? 'block' : 'none';
                // Update folder icon
                folderHeader.innerHTML = (isHidden ? '📂 ' : '📁 ') + label.split(' ').slice(1).join(' ');
            };
            
            const wrapper = document.createElement('li');
            wrapper.style.listStyle = 'none';
            wrapper.appendChild(folderHeader);
            wrapper.appendChild(children);
            
            return wrapper;
        }
        
        function getFileIcon(path) {
            if (path.endsWith('.log')) return '📄';
            if (path.endsWith('.toml')) return '⚙️';
            if (path.endsWith('.xml')) return '📋';
            if (path.endsWith('.json')) return '📊';
            return '📄';
        }
        
        function getLanguage(path) {
            if (path.endsWith('.log')) return 'plaintext';
            if (path.endsWith('.toml')) return 'ini';
            if (path.endsWith('.xml')) return 'xml';
            if (path.endsWith('.json')) return 'json';
            if (path.endsWith('.sh')) return 'shell';
            if (path.endsWith('.go')) return 'go';
            return 'plaintext';
        }
        
        function openFile(path) {
            try {
                // Find file in embedded files array
                const file = files.find(f => f.path === path);
                if (!file) {
                    throw new Error('File not found: ' + path);
                }
                
                // Update editor
                const language = getLanguage(path);
                monaco.editor.setModelLanguage(editor.getModel(), language);
                editor.setValue(file.content);
                
                // Update active file in tree
                document.querySelectorAll('.file-item').forEach(item => {
                    item.classList.remove('active');
                    if (item.dataset.path === path) {
                        item.classList.add('active');
                    }
                });
                
                // Update tab
                const tabs = document.getElementById('editorTabs');
                tabs.innerHTML = '<div class="editor-tab active">' + getFileIcon(path) + ' ' + path.split('/').pop() + '</div>';
                
                currentFile = path;
            } catch (error) {
                console.error('Error opening file:', error);
                editor.setValue('❌ Error loading file: ' + path + '\n\n' + error.message);
            }
        }
        
        function setupSearch() {
            const searchInput = document.getElementById('searchInput');
            const searchResults = document.getElementById('searchResults');
            const fileTreeContainer = document.getElementById('fileTreeContainer');
            
            let searchTimeout;
            searchInput.addEventListener('input', (e) => {
                clearTimeout(searchTimeout);
                const query = e.target.value.trim();
                
                if (query.length < 2) {
                    searchResults.style.display = 'none';
                    fileTreeContainer.style.display = 'block';
                    return;
                }
                
                searchTimeout = setTimeout(() => {
                    performSearch(query);
                }, 300);
            });
        }
        
        function performSearch(query) {
            const searchResults = document.getElementById('searchResults');
            const fileTreeContainer = document.getElementById('fileTreeContainer');
            const results = [];
            
            // Search through all files
            files.forEach(file => {
                const lines = file.content.split('\n');
                lines.forEach((line, lineNum) => {
                    if (line.toLowerCase().includes(query.toLowerCase())) {
                        results.push({
                            file: file,
                            lineNum: lineNum + 1,
                            line: line.trim()
                        });
                    }
                });
            });
            
            // Display results
            if (results.length === 0) {
                searchResults.innerHTML = '<div style="padding: 10px;">No results found</div>';
            } else {
                const maxResults = 50;
                const displayResults = results.slice(0, maxResults);
                
                searchResults.innerHTML = 
                    '<div style="padding: 10px; border-bottom: 1px solid #3e3e42;">' +
                    results.length + ' result' + (results.length !== 1 ? 's' : '') +
                    (results.length > maxResults ? ' (showing first ' + maxResults + ')' : '') +
                    '</div>';
                
                displayResults.forEach(result => {
                    const item = document.createElement('div');
                    item.className = 'search-result-item';
                    
                    const highlighted = result.line.replace(
                        new RegExp(query, 'gi'),
                        match => '<span class="search-match">' + match + '</span>'
                    );
                    
                    item.innerHTML = 
                        '<div><span class="search-result-file">' + result.file.name + '</span>' +
                        '<span class="search-result-line">:' + result.lineNum + '</span></div>' +
                        '<div class="search-result-preview">' + highlighted + '</div>';
                    
                    item.onclick = () => {
                        openFile(result.file.path);
                        // Jump to line after a short delay
                        setTimeout(() => {
                            editor.revealLineInCenter(result.lineNum);
                            editor.setPosition({ lineNumber: result.lineNum, column: 1 });
                        }, 100);
                    };
                    
                    searchResults.appendChild(item);
                });
            }
            
            searchResults.style.display = 'block';
            fileTreeContainer.style.display = 'none';
        }
    </script>
</body>
</html>
`
