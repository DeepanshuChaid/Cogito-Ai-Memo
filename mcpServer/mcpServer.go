package mcpServer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
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

	// GOROUNTINE 1: LISTEN FOR STDIN (NON BLOCKING)
	go func() {
		for scanner.Scan() {
			// WE ARE COPYING THE INPUT FROM THE SCANNER TO OUR INPUT CHANNEL
			msg := make([]byte, len(scanner.Bytes()))
			copy(msg, scanner.Bytes())
			inputChan <- msg
		}
		// IF STDIN CLOSES (EOF) TRIGGER A CLEAN SHUTDOWN
		stopChan <- syscall.SIGTERM
	}()

	// MAIN LOOP THE TRAFFIC CONTROLLER
	for {
		select {
			case <- stopChan:
				if currentSession != nil {
					_ = db.CompleteSession(currentSession.SessionID)
				}
				return // EXITS THE FUNCTIONS & STOPS THE PROCESS CLEANLY
			case rawBytes := <- inputChan:
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
					ID: req.ID,
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
