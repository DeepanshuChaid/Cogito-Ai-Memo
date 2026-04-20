package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      *json.RawMessage       `json:"id,omitempty"` // Pointer lets us check if it's nil
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Result  interface{}      `json:"result,omitempty"`
}

// --- Middleware: Noir Logger ---
func inspector(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		log.Printf("DEBUG: %s | %s", r.URL.Path, string(body))
		next(w, r)
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return
	}

	// 1. NOTIFICATION CHECK (The fix)
	// If ID is nil, it's a notification. We MUST NOT send a response body.
	if req.ID == nil {
		w.WriteHeader(http.StatusNoContent)
		log.Printf("NOTIFICATION: %s handled.", req.Method)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var result interface{}

	// 2. ROUTING
	switch req.Method {
	case "initialize":
		result = map[string]interface{}{
			"protocolVersion": "2025-06-18", // Sync with your client's version
			"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
			"serverInfo":      map[string]string{"name": "cogito", "version": "0.1.0"},
		}

	case "tools/list":
		result = map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "get_codebase_map",
					"description": "Returns full file tree.",
					"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
				},
			},
		}

	case "tools/call":
		name, _ := req.Params["name"].(string)
		var output string
		if name == "get_codebase_map" {
			filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && !filepath.HasPrefix(path, ".") {
					output += path + "\n"
				}
				return nil
			})
		}
		result = map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": output}},
		}

	default:
		result = map[string]string{"status": "unknown_method"}
	}

	// 3. SEND RESPONSE
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", inspector(mainHandler))

	// Some clients look for these specific paths
	mux.HandleFunc("/mcp/initialize", inspector(mainHandler))
	mux.HandleFunc("/mcp/tools", inspector(mainHandler))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("💀 Cogito MCP: Ready for deployment on :8080")
	log.Fatal(server.ListenAndServe())
}
