package mcpServer

import (
	"os"
	// "path/filepath"
	// "strings"

	// "github.com/DeepanshuChaid/Cogito-Ai.git/internals/commands"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/commands"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

var currentSession *schemaModels.Session
var observationCreatedThisSession bool

func handleRequest(req JSONRPCRequest) interface{} {

	switch req.Method {

	//==============================================
	case "initialize":
		cwd, _ := os.Getwd()

		uniqueID := newSessionID()

		session, err := db.InitializeProjectSession(uniqueID, cwd)
		if err == nil {
			currentSession = session
			observationCreatedThisSession = false
		}

		// go func() {
		// 	commands.BuildMap()
		// }()

		// go commands.BuildMap()

		return map[string]interface{}{
			"protocolVersion": "2025-06-18",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": false,
				},
				"prompts": map[string]interface{}{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "cogito",
				"version": "0.1.0",
			},
		}

	//==============================================
	case "initialized":
		return map[string]interface{}{}

	//==============================================
	case "tools/list":
		return map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "create_observation",
					"description": "Store one durable engineering memory from an important activity, decision, discovery, or bugfix",

					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"memory": map[string]interface{}{
								"type":        "string",
								"description": "Compressed durable summary of what happened and why it matters",
							},
							"facts": map[string]interface{}{
								"type":        "string",
								"description": "JSON array of pure factual points for retrieval",
							},
						},
						"required": []string{
							"memory",
						},
					},
				},
				{
					"name":        "get_codebase_map",
					"description": "Get A full Map of the Codebase with Details like importance and functions flow.",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
				{
					"name":        "create_summary",
					"description": "Store a high-level summary of the session: what the user asked, what was learned, and next steps.",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"request":   map[string]interface{}{"type": "string", "description": "The user's original request"},
							"learned":   map[string]interface{}{"type": "string", "description": "Key things learned/changed"},
							"nextSteps": map[string]interface{}{"type": "string", "description": "Recommended next steps"},
						},
						"required": []string{"request", "learned", "nextSteps"},
					},
				},
			},
		}

	//==============================================
	case "prompts/list":
		return map[string]interface{}{
			"prompts": []map[string]interface{}{
				{
					"name":        "caveman-review",
					"description": "Ultra-compressed code review",
				},
			},
		}

	//==============================================
	case "prompts/get":
		name, _ := req.Params["name"].(string)

		if name == "caveman-review" {

			lore := ""
			if currentSession != nil {
				// future: fetch observations
			}

			return map[string]interface{}{
				"messages": []map[string]interface{}{
					{
						"role": "system",
						"content": map[string]interface{}{
							"type": "text",
							"text": PROMPT + "\n\n" + lore,
						},
					},
				},
			}
		}

		return errorResponse(-32601, "prompt not found")

	//==============================================
	case "tools/call":
		name, _ := req.Params["name"].(string)

		if name == "create_observation" {
			arg, ok := req.Params["arguments"].(map[string]interface{})
			if !ok {
				return errorResponse(-32602, "arguments missing")
			}

			memoryText, ok := arg["memory"].(string)
			if !ok || memoryText == "" {
				return errorResponse(-32602, "memoryText missing")
			}

			var fact string
			fact, _ = arg["facts"].(string)

			cwd, _ := os.Getwd()

			if currentSession == nil {
				return errorResponse(-32602, "no active session")
			}

			err := db.CreateObservation(currentSession.SessionID, cwd, memoryText, fact)
			if err != nil {
				return errorResponse(-32603, err.Error())
			}
			observationCreatedThisSession = true

			// importance := 5
			// if val, ok := arg["importance"].(float64); ok {
			// 	importance = int(val)
			// }

			return map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": "Observations Saved Successfully!"},
				},
			}
		}

		if name == "get_codebase_map" {
			commands.BuildMap(false)

			data, _ := os.ReadFile(".cogito/substrate.txt")
			output := string(data)

			return map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": output},
				},
			}
		}

		// Add this to the tools/call switch (around line 188)
		if name == "create_summary" {
			arg, _ := req.Params["arguments"].(map[string]interface{})
			request, _ := arg["request"].(string)
			learned, _ := arg["learned"].(string)
			nextSteps, _ := arg["nextSteps"].(string)
			if request == "" || learned == "" || nextSteps == "" {
				return errorResponse(-32602, "request, learned, nextSteps are required")
			}
			cwd, _ := os.Getwd()
			if currentSession == nil {
				return errorResponse(-32602, "no active session")
			}
			err := db.CreateSessionSummary(currentSession.SessionID, cwd, request, learned, nextSteps)
			if err != nil {
				return errorResponse(-32603, err.Error())
			}
			return map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": "Session Summary Saved Successfully!"},
				},
			}
		}

	}

	return errorResponse(-32601, "method not found")
}
