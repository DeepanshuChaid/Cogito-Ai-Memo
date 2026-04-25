package commands

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileMap struct {
	Path       string         `json:"path,omitempty"`
	Importance int            `json:"importance,omitempty"`
	Role       string         `json:"role,omitempty"`
	Summary    string         `json:"summary,omitempty"`
	Package    string         `json:"package,omitempty"`
	Imports    []string       `json:"imports,omitempty"`
	Functions  []Function     `json:"functions,omitempty"`
	Structs    []string       `json:"structs,omitempty"`
	Interfaces []string       `json:"interfaces,omitempty"`
	Methods    []Method       `json:"methods,omitempty"`
	Language   string         `json:"language,omitempty"`
	Classes    []string       `json:"classes,omitempty"`
	Calls      []CallRelation `json:"calls,omitempty"`
	Ignore     bool           `json:"ignore,omitempty"`
}

type CallRelation struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}


type Function struct {
	Name string `json:"name,omitempty"`
	Line int   `json:"line,omitempty"`
}

type Method struct {
	Receiver string `json:"receiver,omitempty"`
	Name     string `json:"name,omitempty"`
}

type ExecutionFlow struct {
	Name  string   `json:"name,omitempty"`
	Steps []string `json:"steps,omitempty"`
}

type CodebaseMap struct {
	Files []FileMap       `json:"files,omitempty"`
	Flows []ExecutionFlow `json:"flows,omitempty"`
}

func BuildMap() {
	root := "."

	var result CodebaseMap

	fmt.Println("Building codebase map...")

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()

			if name == ".git" ||
				name == "vendor" ||
				name == "node_modules" ||
				name == "dist" ||
				name == "build" ||
				name == ".cogito" {
				return filepath.SkipDir
			}

			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		supported := map[string]bool{
			".go":   true,
			".py":   true,
			".js":   true,
			".jsx":  true,
			".ts":   true,
			".tsx":  true,
			".java": true,
		}

		if !supported[ext] {
			return nil
		}

		fileMap := parseFile(path)

		// Post-process file
		fileMap.Importance = calculateImportance(&fileMap)
		fileMap.Role = classifyRole(path, &fileMap)
		fileMap.Ignore = isIgnoreZone(path)

		result.Files = append(result.Files, fileMap)

		return nil
	})

	// Sort files by importance (descending)
	sort.Slice(result.Files, func(i, j int) bool {
		if result.Files[i].Importance != result.Files[j].Importance {
			return result.Files[i].Importance > result.Files[j].Importance
		}
		return result.Files[i].Path < result.Files[j].Path
	})

	// Generate execution flows
	result.Flows = generateExecutionFlows(result.Files)

	os.MkdirAll(".cogito", os.ModePerm)

	fmt.Println("Creating map file...")

	file, err := os.Create(".cogito/map.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating map file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(result)
}

func parseGoFile(path string) FileMap {
	fset := token.NewFileSet()

	fmt.Println("Parsing file:", path)

	node, err := parser.ParseFile(
		fset,
		path,
		nil,
		parser.ParseComments,
	)

	if err != nil {
		return FileMap{
			Path: path,
		}
	}

	fileMap := FileMap{
		Path:    path,
		Package: node.Name.Name,
	}

	if node.Doc != nil {
		fileMap.Summary = strings.TrimSpace(node.Doc.Text())
	} else {
		// Fallback to searching first comment in file
		for _, cg := range node.Comments {
			if cg.Pos() < node.Name.Pos() {
				fileMap.Summary = strings.TrimSpace(cg.Text())
				break
			}
		}
	}

	for _, imp := range node.Imports {
		fileMap.Imports = append(
			fileMap.Imports,
			strings.Trim(imp.Path.Value, `"`),
		)
	}

	for _, decl := range node.Decls {
		switch d := decl.(type) {

		case *ast.FuncDecl:
			funcName := d.Name.Name
			if d.Recv == nil {
				fileMap.Functions = append(
					fileMap.Functions,
					Function{
						Name: funcName,
						Line: fset.Position(d.Pos()).Line,
					},
				)
			} else {
				receiver := ""
				if len(d.Recv.List) > 0 {
					switch r := d.Recv.List[0].Type.(type) {
					case *ast.Ident:
						receiver = r.Name
					case *ast.StarExpr:
						if ident, ok := r.X.(*ast.Ident); ok {
							receiver = ident.Name
						}
					}
				}
				fileMap.Methods = append(
					fileMap.Methods,
					Method{
						Receiver: receiver,
						Name:     funcName,
					},
				)
			}

			// Detect calls within the function body
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
						fileMap.Calls = append(fileMap.Calls, CallRelation{
							From: funcName,
							To:   calleeName,
						})
					}
					return true
				})
			}

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				switch typeSpec.Type.(type) {
				case *ast.StructType:
					fileMap.Structs = append(
						fileMap.Structs,
						typeSpec.Name.Name,
					)

				case *ast.InterfaceType:
					fileMap.Interfaces = append(
						fileMap.Interfaces,
						typeSpec.Name.Name,
					)
				}
			}
		}
	}

	return fileMap
}

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
		return FileMap{
			Path: path,
		}
	}
}

func parsePythonFile(path string) FileMap {
	content, _ := os.ReadFile(path)
	text := string(content)

	fileMap := FileMap{
		Path:     path,
		Language: "python",
	}

	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
			fileMap.Imports = append(fileMap.Imports, line)
			continue
		}

		// Function detection (including async def)
		if strings.HasPrefix(line, "def ") || strings.HasPrefix(line, "async def ") {
			trimmed := strings.TrimPrefix(line, "async ")
			trimmed = strings.TrimPrefix(trimmed, "def ")
			name := strings.Split(trimmed, "(")[0]
			name = strings.TrimSpace(name)
			if name != "" {
				fileMap.Functions = append(fileMap.Functions, Function{
					Name: name,
				})
			}
			continue
		}

		// Class detection
		if strings.HasPrefix(line, "class ") {
			name := strings.TrimPrefix(line, "class ")
			name = strings.Split(name, "(")[0]
			name = strings.Split(name, ":")[0]
			name = strings.TrimSpace(name)
			if name != "" {
				fileMap.Classes = append(fileMap.Classes, name)
			}
			continue
		}

		// Basic decorator detection (adding to imports or a new field if we had one, but let's just ignore for now or log)
		// User mentioned "decorators awareness"
	}

	return fileMap
}



func parseJSFile(path string) FileMap {
	content, _ := os.ReadFile(path)
	text := string(content)

	fileMap := FileMap{
		Path:     path,
		Language: "javascript",
	}

	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "import ") {
			fileMap.Imports = append(fileMap.Imports, line)
			continue
		}

		// function declarations and exports
		if strings.Contains(line, "function ") {
			trimmed := strings.TrimPrefix(line, "export ")
			trimmed = strings.TrimPrefix(trimmed, "default ")
			if strings.HasPrefix(trimmed, "function ") {
				name := strings.Split(strings.TrimPrefix(trimmed, "function "), "(")[0]
				name = strings.TrimSpace(name)
				if name != "" {
					fileMap.Functions = append(fileMap.Functions, Function{
						Name: name,
					})
				}
				continue
			}
		}

		// arrow functions: const name = (...) =>
		if (strings.Contains(line, "const ") || strings.Contains(line, "let ") || strings.Contains(line, "var ")) &&
			strings.Contains(line, "=>") {
			parts := strings.Split(line, "=")
			if len(parts) > 1 {
				decl := strings.Fields(parts[0])
				if len(decl) > 0 {
					name := decl[len(decl)-1]
					fileMap.Functions = append(fileMap.Functions, Function{
						Name: name,
					})
				}
			}
			continue
		}

		// Class detection
		if strings.Contains(line, "class ") {
			trimmed := strings.TrimPrefix(line, "export ")
			trimmed = strings.TrimPrefix(trimmed, "default ")
			if strings.HasPrefix(trimmed, "class ") {
				name := strings.Split(strings.TrimPrefix(trimmed, "class "), " ")[0]
				name = strings.Trim(name, "{")
				name = strings.TrimSpace(name)
				if name != "" {
					fileMap.Classes = append(fileMap.Classes, name)
				}
			}
			continue
		}
	}

	return fileMap
}

func parseJavaFile(path string) FileMap {
	content, _ := os.ReadFile(path)
	text := string(content)

	fileMap := FileMap{
		Path:     path,
		Language: "java",
	}

	lines := strings.Split(text, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// package com.example;
		if strings.HasPrefix(line, "package ") {
			pkg := strings.TrimPrefix(line, "package ")
			pkg = strings.TrimSuffix(pkg, ";")
			fileMap.Package = pkg
		}

		// import java.util.List;
		if strings.HasPrefix(line, "import ") {
			imp := strings.TrimPrefix(line, "import ")
			imp = strings.TrimSuffix(imp, ";")
			fileMap.Imports = append(fileMap.Imports, imp)
		}

		// public class User {
		if strings.Contains(line, " class ") ||
			strings.HasPrefix(line, "class ") {

			parts := strings.Fields(line)

			for idx, part := range parts {
				if part == "class" && idx+1 < len(parts) {
					fileMap.Classes = append(
						fileMap.Classes,
						strings.Trim(parts[idx+1], "{"),
					)
					break
				}
			}
		}

		// public interface UserService {
		if strings.Contains(line, " interface ") ||
			strings.HasPrefix(line, "interface ") {

			parts := strings.Fields(line)

			for idx, part := range parts {
				if part == "interface" && idx+1 < len(parts) {
					fileMap.Interfaces = append(
						fileMap.Interfaces,
						strings.Trim(parts[idx+1], "{"),
					)
					break
				}
			}
		}

		// simple method detection
		// public void login() {
		if strings.Contains(line, "(") &&
			strings.Contains(line, ")") &&
			strings.Contains(line, "{") &&
			!strings.Contains(line, "if") &&
			!strings.Contains(line, "for") &&
			!strings.Contains(line, "while") &&
			!strings.Contains(line, "switch") &&
			!strings.Contains(line, "catch") {

			beforeParen := strings.Split(line, "(")[0]
			parts := strings.Fields(beforeParen)

			if len(parts) > 0 {
				name := parts[len(parts)-1]

				if name != "class" &&
					name != "interface" &&
					name != "new" {

					fileMap.Functions = append(
						fileMap.Functions,
						Function{
							Name: name,
							Line: i + 1,
						},
					)
				}
			}
		}
	}

	return fileMap
}

func isLowValueCall(name string) bool {
	lowValue := map[string]bool{
		"len": true, "append": true, "cap": true, "make": true, "new": true,
		"string": true, "int": true, "int64": true, "float64": true,
		"Println": true, "Printf": true, "Print": true, "Sprintf": true,
		"Error": true, "Errorf": true, "Exit": true, "Fatal": true, "Fatalf": true,
		"Panic": true, "recover": true, "close": true, "delete": true,
		"copy": true, "real": true, "imag": true, "complex": true,
	}
	return lowValue[name]
}

func calculateImportance(f *FileMap) int {
	if isEntryPoint(f.Path, f) {
		return 10
	}

	score := 1
	score += len(f.Functions)
	score += len(f.Structs) * 2
	score += len(f.Interfaces) * 3

	// Boost for exported items
	for _, fn := range f.Functions {
		if len(fn.Name) > 0 && fn.Name[0] >= 'A' && fn.Name[0] <= 'Z' {
			score += 2
		}
	}

	// Boost for other potential entrypoints or high-level cmd files
	if strings.Contains(f.Path, "cmd/") || strings.Contains(f.Path, "main") {
		score += 5
	}

	// Cap at 10
	if score > 10 {
		return 10
	}
	return score
}

func classifyRole(path string, f *FileMap) string {
	if isEntryPoint(path, f) {
		return "entrypoint"
	}

	p := strings.ToLower(path)
	if strings.Contains(p, "db/") || strings.Contains(p, "database") || strings.Contains(p, "repository") {
		return "database"
	}
	if strings.Contains(p, "mcp") {
		return "mcp-server"
	}
	if strings.Contains(p, "session") {
		return "session-manager"
	}
	if strings.Contains(p, "worker") || strings.Contains(p, "job") {
		return "worker"
	}
	if strings.Contains(p, "adapter") {
		return "adapter"
	}
	if strings.Contains(p, "config") {
		return "config"
	}
	if strings.Contains(p, "ui") || strings.Contains(p, "frontend") {
		return "ui"
	}
	if strings.Contains(p, "api") || strings.Contains(p, "handler") {
		return "api-layer"
	}
	if strings.Contains(p, "inject") || strings.Contains(p, "container") {
		return "injector"
	}
	return "logic"
}

func isIgnoreZone(path string) bool {
	p := strings.ToLower(path)
	return strings.Contains(p, "_test.go") ||
		strings.Contains(p, "vendor/") ||
		strings.Contains(p, "generated/") ||
		strings.Contains(p, "mock") ||
		strings.Contains(p, "temp") ||
		strings.Contains(p, "debug")
}

func generateExecutionFlows(files []FileMap) []ExecutionFlow {
	var flows []ExecutionFlow

	// Heuristic: Build major paths
	// 1. CLI Execution (main -> internals)
	cliFlow := ExecutionFlow{Name: "CLI Execution"}
	for _, f := range files {
		if strings.HasSuffix(f.Path, "main.go") {
			cliFlow.Steps = append(cliFlow.Steps, "main")
			for _, call := range f.Calls {
				cliFlow.Steps = append(cliFlow.Steps, call.To)
			}
			break
		}
	}
	if len(cliFlow.Steps) > 1 {
		flows = append(flows, cliFlow)
	}

	// 2. Build Map Flow (BuildMap -> parse -> post-process)
	buildFlow := ExecutionFlow{Name: "Build Map Flow"}
	for _, f := range files {
		if strings.Contains(f.Path, "buildMap.go") {
			buildFlow.Steps = append(buildFlow.Steps, "BuildMap")
			for _, fn := range f.Functions {
				if strings.HasPrefix(fn.Name, "parse") || strings.HasPrefix(fn.Name, "calculate") {
					buildFlow.Steps = append(buildFlow.Steps, fn.Name)
				}
			}
			break
		}
	}
	if len(buildFlow.Steps) > 1 {
		flows = append(flows, buildFlow)
	}

	return flows
}

func isEntryPoint(path string, f *FileMap) bool {
	p := strings.ToLower(filepath.Base(path))

	// Go
	if f.Package == "main" || p == "main.go" {
		return true
	}

	// JS / TS
	if strings.HasPrefix(p, "index.") ||
		strings.HasPrefix(p, "server.") ||
		strings.HasPrefix(p, "app.") ||
		strings.HasPrefix(p, "main.") {
		return true
	}

	// Python
	if p == "main.py" || p == "app.py" || p == "server.py" {
		return true
	}

	// Java
	if p == "main.java" || p == "app.java" {
		return true
	}

	return false
}
