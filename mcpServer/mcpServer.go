package mcpServer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
)

func ServeMcp() {
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	encoder := json.NewEncoder(os.Stdout)

	stopChan := make(chan os.Signal, 1)
	inputChan := make(chan []byte)

	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stopChan)

	var shutdownOnce sync.Once
	finalizeSession := func() {
		shutdownOnce.Do(func() {
			if currentSession == nil {
				return
			}

			if err := GenerateAutoSummary(currentSession.SessionID, currentSession.Project); err != nil {
				fmt.Fprintf(os.Stderr, "auto-summary failed: %v\n", err)
			}

			if err := db.CompleteSession(currentSession.SessionID); err != nil {
				fmt.Fprintf(os.Stderr, "complete session failed: %v\n", err)
			}
		})
	}
	defer finalizeSession()

	go func() {
		for scanner.Scan() {
			msg := make([]byte, len(scanner.Bytes()))
			copy(msg, scanner.Bytes())
			inputChan <- msg
		}
		stopChan <- syscall.SIGTERM
	}()

	for {
		select {
		case <-stopChan:
			finalizeSession()
			return
		case rawBytes := <-inputChan:
			var req JSONRPCRequest
			err := json.Unmarshal(rawBytes, &req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "JSON decode error: %v\n", err)
				continue
			}

			if req.ID == nil {
				continue
			}

			result := handleRequest(req)

			resp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
			}

			if m, ok := result.(map[string]interface{}); ok {
				if errVal, exists := m["error"]; exists {
					resp.Error = errVal
				} else {
					resp.Result = m
				}
			} else {
				resp.Result = result
			}

			encoder.Encode(resp)
		}
	}
}
