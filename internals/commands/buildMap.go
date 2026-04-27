package commands

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Schema v4.0 - CRG Substrate (LLM Reasoning Surface)

func BuildMap() {
	root := "."
	_ = LoadCache()
	newCache := Cache{Files: make(map[string]string)}
	var allFiles []FileMap

	fmt.Println("Building CRG Substrate v4.0...")

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil { return nil }
		if info.IsDir() {
			if ShouldSkipDir(info.Name()) { return filepath.SkipDir }
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !isSupported(ext) { return nil }

		hash := GetFileHash(path)
		newCache.Files[path] = hash
		fileMap := parseFile(path)
		allFiles = append(allFiles, fileMap)
		return nil
	})
	SaveCache(newCache)

	// Pass 1: Global Symbol Indexing
	funcToPath := make(map[string]string)
	for _, f := range allFiles {
		for _, fn := range f.Functions {
			funcToPath[fn.Name] = f.Path
		}
	}

	// Pass 2: Graph construction & Weighting
	edgeMap := make(map[string]int)
	for _, f := range allFiles {
		for _, call := range f.Calls {
			targetPath, ok := funcToPath[call.To]
			if ok && targetPath != f.Path {
				edgeMap[f.Path+"|"+targetPath]++
			}
		}
	}

	// Pass 2.5: Centrality Calculation
	pathToIdx := make(map[string]int)
	for i, f := range allFiles {
		pathToIdx[f.Path] = i
	}
	for edge, weight := range edgeMap {
		parts := strings.Split(edge, "|")
		if idx, ok := pathToIdx[parts[0]]; ok {
			allFiles[idx].FanOut += weight
		}
		if idx, ok := pathToIdx[parts[1]]; ok {
			allFiles[idx].FanIn += weight
		}
	}

	// Pass 3: Iterative Importance Propagation
	importance := make(map[string]int)
	for _, f := range allFiles {
		importance[f.Path] = scoreFile(f)
	}
	for i := 0; i < 2; i++ {
		nextImportance := make(map[string]int)
		for p, s := range importance { nextImportance[p] = s }
		for edge, weight := range edgeMap {
			parts := strings.Split(edge, "|")
			from, to := parts[0], parts[1]
			nextImportance[to] += (importance[from] * weight) / 10
		}
		importance = nextImportance
	}

	// Pass 4: Final Assembly (v4.0 Substrate)
	var sb strings.Builder
	sb.WriteString("# CRG v4.0 Substrate\n\n@idx\n")

	idMap := make(map[string]string)
	nodePaths := []string{}

	count := 0
	for _, f := range allFiles {
		score := importance[f.Path]
		if score < 15 && !isEntryPoint(f.Path, &f) { continue }
		id := fmt.Sprintf("%d", count)
		idMap[f.Path] = id
		nodePaths = append(nodePaths, f.Path)
		sb.WriteString(fmt.Sprintf("%s:%s\n", id, filepath.ToSlash(f.Path)))
		count++
	}

	sb.WriteString("\n@nodes\n")
	for _, path := range nodePaths {
		id := idMap[path]
		f := findFile(allFiles, path)
		score := importance[path]
		if score > 99 { score = 99 }

		typeCode := "u" // util
		p := strings.ToLower(filepath.ToSlash(path))
		if isEntryPoint(path, f) {
			typeCode = "e"
		} else if strings.Contains(p, "/db/") || strings.Contains(p, "/model") || strings.Contains(p, "/repository") {
			typeCode = "p" // persistence
		} else if strings.Contains(p, "config") || strings.Contains(p, "session") || strings.Contains(p, "cache") {
			typeCode = "s" // state/config
		} else if strings.Contains(p, "handler") || strings.Contains(p, "serve") || strings.Contains(p, "controller") {
			typeCode = "o" // orchestrator
		} else if strings.Contains(p, "adapter") || strings.Contains(p, "api/") || strings.Contains(p, "client/") {
			typeCode = "a" // adapter
		} else if score >= 80 {
			typeCode = "c" // core (generic high value)
		}

		keys := ""
		if kf := extractKeyFunctions(f); len(kf) > 0 {
			keys = " !" + strings.Join(kf, ",")
		}

		sb.WriteString(fmt.Sprintf("%s %s%02d%s\n", id, typeCode, score, keys))
	}

	sb.WriteString("\n@edges\n")
	for edge, weight := range edgeMap {
		parts := strings.Split(edge, "|")
		fromID, ok1 := idMap[parts[0]]
		toID, ok2 := idMap[parts[1]]
		if ok1 && ok2 {
			rel := "c" // call
			target := findFile(allFiles, parts[1])
			if target != nil {
				if strings.Contains(strings.ToLower(parts[1]), "db/") { rel = "w" } else if strings.Contains(strings.ToLower(parts[1]), "config") { rel = "r" }
			}
			sb.WriteString(fmt.Sprintf("%s>%s:%d%s\n", fromID, toID, weight, rel))
		}
	}

	sb.WriteString("\n@why\n")
	for _, path := range nodePaths {
		id := idMap[path]
		f := findFile(allFiles, path)
		if f != nil && f.Summary != "" {
			sb.WriteString(fmt.Sprintf("%s +%s\n", id, strings.ReplaceAll(f.Summary, "\n", " ")))
		}
	}

	sb.WriteString("\n@flows\n")
	for path, id := range idMap {
		if isEntryPoint(path, findFile(allFiles, path)) {
			chain := []string{id}
			curr := path
			seen := map[string]bool{id: true}
			for i := 0; i < 4; i++ {
				next := ""
				maxW := -1
				for edge, w := range edgeMap {
					pts := strings.Split(edge, "|")
					if pts[0] == curr {
						if nid := idMap[pts[1]]; nid != "" && !seen[nid] && w > maxW {
							maxW = w
							next = pts[1]
						}
					}
				}
				if next != "" {
					nid := idMap[next]
					chain = append(chain, nid)
					seen[nid] = true
					curr = next
				} else { break }
			}
			if len(chain) > 1 {
				name := "EXEC"
				if strings.Contains(path, "main") { name = "CLI" } else if strings.Contains(path, "mcp") { name = "MCP" }
				sb.WriteString(fmt.Sprintf("%s: %s\n", name, strings.Join(chain, ">")))
			}
		}
	}

	os.MkdirAll(".cogito", os.ModePerm)
	os.WriteFile(".cogito/substrate.txt", []byte(sb.String()), 0644)
	fmt.Println("CRG Substrate v4.0 successfully created at .cogito/substrate.txt")
}

func findFile(files []FileMap, path string) *FileMap {
	for i := range files {
		if files[i].Path == path { return &files[i] }
	}
	return nil
}

func extractKeyFunctions(f *FileMap) []string {
	var keys []string
	if f == nil { return keys }
	
	base := strings.TrimSuffix(filepath.Base(f.Path), filepath.Ext(f.Path))
	
	for _, fn := range f.Functions {
		// Priority 1: File Ownership (e.g. detect.py -> detect())
		if fn.Name == base {
			keys = append(keys, fn.Name)
			continue
		}
		
		// Priority 2: High Signal Filtering
		if isHighSignalFunc(fn.Name) {
			keys = append(keys, fn.Name)
		}
	}
	return keys
}

// Helpers

func isSupported(ext string) bool {
	exts := map[string]bool{".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true, ".java": true}
	return exts[ext]
}

func scoreFile(f FileMap) int {
	score := 20 // Base importance

	// 1. Entrypoint & Strategic Locations
	if isEntryPoint(f.Path, &f) {
		score += 50
	}
	if strings.Contains(f.Path, "cmd/") || strings.Contains(f.Path, "server") {
		score += 30
	}

	// 2. Core Roles (Keywords in path/package)
	p := strings.ToLower(f.Path)
	pkg := strings.ToLower(f.Package)
	roles := []string{"service", "controller", "db", "repository", "config", "handler", "manager", "client", "provider"}
	for _, role := range roles {
		if strings.Contains(p, role) || strings.Contains(pkg, role) {
			score += 20
		}
	}

	// 3. Structural Significance
	exportedCount := 0
	constructorCount := 0
	asyncCount := 0
	for _, fn := range f.Functions {
		if fn.IsExported {
			exportedCount++
		} else if f.Language == "go" && len(fn.Name) > 0 && fn.Name[0] >= 'A' && fn.Name[0] <= 'Z' {
			exportedCount++
		}
		
		if strings.HasPrefix(fn.Name, "New") || fn.Name == "constructor" {
			constructorCount++
		}
		if fn.IsAsync {
			asyncCount++
		}
	}
	
	score += exportedCount * 4
	score += len(f.Interfaces) * 10
	score += len(f.Structs) * 5
	score += len(f.Classes) * 10
	score += constructorCount * 10
	score += asyncCount * 3

	// Barrel file reduction (Many exports, nothing else)
	if exportedCount > 5 && len(f.Functions) == exportedCount && len(f.Structs) == 0 {
		score -= 30
	}

	// 4. Centrality (Fan-in > Fan-out is higher value)
	score += f.FanIn * 5
	if f.FanIn > f.FanOut {
		score += 15
	}

	// 5. Language-specific signals
	switch f.Language {
	case "python":
		// Entrypoint block
		if f.HasMainBlock {
			score += 40
		}
		// Framework decorators: web routes, events, DI
		webDecorators := 0
		for _, ann := range f.Annotations {
			a := strings.ToLower(ann)
			if strings.Contains(a, "route") || strings.Contains(a, "get") || strings.Contains(a, "post") ||
				strings.Contains(a, "put") || strings.Contains(a, "delete") || strings.Contains(a, "on_event") ||
				strings.Contains(a, "task") || strings.Contains(a, "celery") {
				webDecorators++
			}
		}
		score += webDecorators * 8
		// Test files
		base := strings.ToLower(filepath.Base(p))
		if strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py") {
			score -= 40
		}
		// Simple utility scripts (no classes, few public functions)
		publicFns := 0
		for _, fn := range f.Functions {
			if fn.IsExported { publicFns++ }
		}
		if len(f.Classes) == 0 && publicFns < 3 && !f.HasMainBlock {
			score -= 15
		}

	case "java":
		// Spring entrypoint
		springAnnotations := map[string]int{
			"SpringBootApplication": 50,
			"RestController":        30,
			"Controller":            25,
			"Service":               20,
			"Repository":            20,
			"Component":             15,
			"Configuration":         20,
			"Entity":                15,
		}
		for _, ann := range f.Annotations {
			if bonus, ok := springAnnotations[ann]; ok {
				score += bonus
			}
		}
		// Public method density
		score += f.PublicMethods * 4
		// Test classes
		base := strings.ToLower(filepath.Base(p))
		if strings.HasSuffix(base, "test.java") || strings.HasSuffix(base, "tests.java") {
			score -= 40
		}
		// DTO / pure data classes (no public methods beyond getters)
		if len(f.Classes) > 0 && f.PublicMethods < 4 && len(f.Annotations) == 0 {
			score -= 20
		}
	}

	// 6. Reductions (Noise suppression — universal)
	if strings.HasSuffix(p, "_test.go") || strings.Contains(p, "/test/") || strings.Contains(p, "/mocks/") || filepath.Base(p) == "test.go" {
		score -= 40
	}
	if strings.Contains(p, "generated") || strings.HasSuffix(p, ".gen.go") || strings.HasPrefix(filepath.Base(p), "mock_") {
		score -= 60
	}

	// Utility wrapper reduction (Exempt entrypoints)
	if !isEntryPoint(f.Path, &f) && len(f.Functions) < 4 && len(f.Structs) == 0 && len(f.Interfaces) == 0 {
		score -= 15
	}

	if score < 5 { score = 5 }
	if score > 100 { score = 100 }
	return score
}

func classifyTags(path string, f *FileMap) []string {
	var tags []string
	if isEntryPoint(path, f) {
		tags = append(tags, "entrypoint")
	}
	p := strings.ToLower(path)
	if strings.Contains(p, "db/") || strings.Contains(p, "database") {
		tags = append(tags, "persistence-layer")
	}
	if strings.Contains(p, "session") || strings.Contains(p, "config") {
		tags = append(tags, "state-owner")
	}
	if strings.Contains(p, "adapter") || strings.Contains(p, "api") {
		tags = append(tags, "boundary", "adapter")
	}
	if strings.Contains(p, "handler") || strings.Contains(p, "serve") {
		tags = append(tags, "orchestrator")
	}
	if len(tags) == 0 {
		tags = append(tags, "utility")
	}
	return tags
}

// Parsers

func parseFile(path string) FileMap {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return parseGoFile(path)
	case ".py":
		return parsePythonFile(path)
	case ".js", ".jsx", ".ts", ".tsx":
		return parseJSFile(path)
	case ".java":
		return parseJavaFile(path)
	default:
		return FileMap{Path: path}
	}
}

func parseGoFile(path string) FileMap {
	fset := token.NewFileSet()
	fmt.Println("Parsing file:", path)
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return FileMap{Path: path}
	}
	fileMap := FileMap{Path: path, Package: node.Name.Name}
	if node.Doc != nil {
		fileMap.Summary = strings.TrimSpace(node.Doc.Text())
	}
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			funcName := d.Name.Name
			function := Function{
				Name:       funcName,
				Line:       fset.Position(d.Pos()).Line,
				IsExported: d.Name.IsExported(),
			}
			fileMap.Functions = append(fileMap.Functions, function)
			if d.Body != nil {
				ast.Inspect(d.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					var calleeName string
					switch fun := call.Fun.(type) {
					case *ast.Ident:
						calleeName = fun.Name
					case *ast.SelectorExpr:
						calleeName = fun.Sel.Name
					}
					if calleeName != "" && !isLowValueCall(calleeName) {
						fileMap.Calls = append(fileMap.Calls, CallRelation{From: funcName, To: calleeName})
					}
					return true
				})
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						fileMap.Structs = append(fileMap.Structs, typeSpec.Name.Name)
					}
					if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
						fileMap.Interfaces = append(fileMap.Interfaces, typeSpec.Name.Name)
					}
				}
			}
		}
	}
	return fileMap
}

var pyFuncRegex = regexp.MustCompile(`^\s*(?:async\s+)?def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
var pyClassRegex = regexp.MustCompile(`^\s*class\s+([a-zA-Z_][a-zA-Z0-9_]*)\b`)

func parsePythonFile(path string) FileMap {
	fmt.Println("Parsing file:", path)
	content, _ := os.ReadFile(path)
	text := string(content)
	fileMap := FileMap{Path: path, Language: "python"}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") { continue }

		// Entrypoint block
		if strings.Contains(trimmed, "__name__") && strings.Contains(trimmed, "__main__") {
			fileMap.HasMainBlock = true
		}

		// Framework decorators (@app.route, @router.get, @app.on_event)
		if strings.HasPrefix(trimmed, "@") {
			dec := strings.Split(strings.TrimPrefix(trimmed, "@"), "(")[0]
			dec = strings.TrimSpace(dec)
			fileMap.Annotations = append(fileMap.Annotations, dec)
		}

		// Function detection
		if match := pyFuncRegex.FindStringSubmatch(line); match != nil {
			name := match[1]
			isAsync := strings.Contains(trimmed, "async ")
			isPublic := !strings.HasPrefix(name, "_") || (strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__"))
			fileMap.Functions = append(fileMap.Functions, Function{
				Name: name, IsAsync: isAsync, IsExported: isPublic,
			})
		}

		// Class detection
		if match := pyClassRegex.FindStringSubmatch(line); match != nil {
			fileMap.Classes = append(fileMap.Classes, match[1])
		}
	}
	return fileMap
}

var jsFuncRegexDecl = regexp.MustCompile(`(?:^|\s)(?:export\s+)?(?:default\s+)?(?:async\s+)?function(?:\s*\*\s*|\s+)([a-zA-Z_$][a-zA-Z0-9_$]*)?\s*\(`)
var jsFuncRegexExpr = regexp.MustCompile(`(?:^|\s)(?:export\s+)?(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(?:async\s+)?(?:function\s*(?:\*)?\s*(?:[a-zA-Z_$][a-zA-Z0-9_$]*)?\s*\(|\([^)]*\)\s*=>|[a-zA-Z_$][a-zA-Z0-9_$]*\s*=>)`)
var jsMethodRegex = regexp.MustCompile(`^\s*(?:async\s+)?(?:get\s+|set\s+|static\s+)?\*?\s*([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*(?::\s*[^\{]+)?\s*\{`)

func isJSKeyword(name string) bool {
	kw := map[string]bool{"if": true, "for": true, "while": true, "switch": true, "catch": true, "constructor": true, "function": true, "return": true, "await": true, "yield": true}
	return kw[name]
}

func parseJSFile(path string) FileMap {
	fmt.Println("Parsing file:", path)
	content, _ := os.ReadFile(path)
	text := string(content)
	ext := strings.ToLower(filepath.Ext(path))
	fileMap := FileMap{Path: path, Language: "javascript"}
	if ext == ".ts" || ext == ".tsx" { fileMap.Language = "typescript" }
	
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lineStr := line
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") { continue }

		isExported := strings.Contains(line, "export ") || strings.Contains(line, "exports.") || strings.Contains(line, "module.exports")
		isAsync := strings.Contains(line, "async ")

		// Function detection
		name := ""
		if match := jsFuncRegexDecl.FindStringSubmatch(lineStr); match != nil {
			name = match[1]
			if name == "" && strings.Contains(line, "export default") {
				name = "default"
			}
		} else if match := jsFuncRegexExpr.FindStringSubmatch(lineStr); match != nil && match[1] != "" {
			name = match[1]
		} else if match := jsMethodRegex.FindStringSubmatch(lineStr); match != nil && match[1] != "" {
			name = match[1]
		}

		if name != "" && !isJSKeyword(name) {
			fileMap.Functions = append(fileMap.Functions, Function{
				Name: name, Line: i + 1, IsAsync: isAsync, IsExported: isExported,
			})
		}

		// Class detection 
		if strings.Contains(line, "class ") {
			parts := strings.Split(line, "class ")
			if len(parts) > 1 {
				cname := strings.TrimSpace(strings.Split(parts[1], " ")[0])
				cname = strings.Trim(cname, "{")
				if cname != "" {
					fileMap.Classes = append(fileMap.Classes, cname)
				}
			}
		}
		
		// React detection heuristic
		if (ext == ".tsx" || ext == ".jsx") && (strings.Contains(line, "return (") || strings.Contains(line, "=> (")) {
			if !strings.Contains(fileMap.Summary, "React Component") {
				fileMap.Summary += " [React Component]"
			}
		}
	}
	return fileMap
}

var javaMethodRegex = regexp.MustCompile(`^\s*(?:(?:public|private|protected|static|final|native|synchronized|abstract|default)\s+)*\s*(?:[\w<>\[\]\?]+\s+)*([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(`)

func isJavaKeyword(name string) bool {
	kw := map[string]bool{"if": true, "for": true, "while": true, "switch": true, "catch": true, "synchronized": true, "return": true, "new": true, "super": true, "this": true}
	return kw[name]
}

func parseJavaFile(path string) FileMap {
	fmt.Println("Parsing file:", path)
	content, _ := os.ReadFile(path)
	text := string(content)
	fileMap := FileMap{Path: path, Language: "java"}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") { continue }

		// Package declaration
		if strings.HasPrefix(trimmed, "package ") {
			fileMap.Package = strings.TrimSuffix(strings.TrimPrefix(trimmed, "package "), ";")
		}

		// Spring/Java annotations
		if strings.HasPrefix(trimmed, "@") {
			ann := strings.Split(strings.TrimPrefix(trimmed, "@"), "(")[0]
			fileMap.Annotations = append(fileMap.Annotations, strings.TrimSpace(ann))
		}

		// Class / interface detection
		if strings.Contains(trimmed, "{") {
			parts := strings.Fields(trimmed)
			for j, p := range parts {
				if p == "class" && j+1 < len(parts) {
					fileMap.Classes = append(fileMap.Classes, strings.Trim(parts[j+1], "{"))
				}
				if p == "interface" && j+1 < len(parts) {
					fileMap.Interfaces = append(fileMap.Interfaces, strings.Trim(parts[j+1], "{"))
				}
			}
		}

		// Method detection
		if !strings.Contains(trimmed, "class ") && !strings.Contains(trimmed, "interface ") && !strings.Contains(trimmed, "new ") {
			if match := javaMethodRegex.FindStringSubmatch(line); match != nil && match[1] != "" {
				name := match[1]
				if !isJavaKeyword(name) {
					isPublic := strings.Contains(trimmed, "public ") || (!strings.Contains(trimmed, "private ") && !strings.Contains(trimmed, "protected "))
					isAsync := strings.Contains(trimmed, "CompletableFuture") || strings.Contains(trimmed, "Mono<") || strings.Contains(trimmed, "Flux<") || strings.Contains(trimmed, "@Async")
					if isPublic {
						fileMap.PublicMethods++
					}
					fileMap.Functions = append(fileMap.Functions, Function{
						Name: name, Line: i + 1, IsAsync: isAsync, IsExported: isPublic,
					})
				}
			}
		}
	}
	return fileMap
}

func isEntryPoint(path string, f *FileMap) bool {
	p := strings.ToLower(filepath.Base(path))
	// Go
	if f.Package == "main" || p == "main.go" { return true }
	// JS/TS
	if strings.HasPrefix(p, "index.") || strings.HasPrefix(p, "server.") ||
		strings.HasPrefix(p, "app.") || strings.HasPrefix(p, "main.") ||
		strings.Contains(p, "mcpserver") || strings.Contains(p, "handlerequest") { return true }
	// Python: if __name__ == "__main__" block found
	if f.Language == "python" && f.HasMainBlock { return true }
	// Java: SpringBootApplication annotation present
	if f.Language == "java" {
		for _, ann := range f.Annotations {
			if ann == "SpringBootApplication" { return true }
		}
	}
	return false
}

func isHighSignalFunc(name string) bool {
	if len(name) == 0 { return false }
	
	// Universal Noise
	noise := map[string]bool{
		"String": true, "Error": true, "Len": true, "temp": true, "helper": true,
		"test": true, "random": true, "wrapper": true, "callback": true,
	}
	if noise[name] { return false }
	
	low := strings.ToLower(name)
	if strings.Contains(low, "temp") || strings.Contains(low, "helper") || strings.Contains(low, "test") {
		return false
	}

	// 1. Exported Naming (Go-style)
	if name[0] >= 'A' && name[0] <= 'Z' {
		return true
	}

	// 2. Significant Internal/Architectural patterns (Python/JS/Java)
	// Allow meaningful underscore-prefixed if they aren't noise
	trimmed := strings.TrimPrefix(name, "_")
	if len(trimmed) == 0 { return false }

	prefixes := []string{
		"get", "set", "create", "update", "build", "handle", "start", "stop",
		"serve", "process", "load", "save", "init", "initialize", "run", "watch",
		"inject", "complete", "mark", "classify", "count", "convert", "extract",
		"detect", "parse", "is", "has", "should", "validate", "check",
	}

	lowTrimmed := strings.ToLower(trimmed)
	for _, p := range prefixes {
		if strings.HasPrefix(lowTrimmed, p) {
			return true
		}
	}

	// 3. Snake_case heuristic for architectural functions
	if strings.Contains(name, "_") && len(name) > 8 {
		return true
	}

	return false
}

func isLowValueCall(name string) bool {
	low := map[string]bool{"len": true, "append": true, "make": true, "new": true, "Println": true, "Printf": true}
	return low[name]
}

type FileMap struct {
	Path        string
	Summary     string
	Package     string
	Functions   []Function
	Structs     []string
	Interfaces  []string
	Classes     []string
	Calls       []CallRelation
	Language    string
	FanIn       int
	FanOut      int
	Annotations []string // Java Spring / Python decorators
	HasMainBlock bool    // Python __name__=="__main__"
	PublicMethods int   // Java public method count
}

type Function struct {
	Name       string
	Line       int
	IsAsync    bool
	IsExported bool
}

type CallRelation struct {
	From string
	To   string
}
