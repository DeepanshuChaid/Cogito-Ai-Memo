package injector

import (
	"fmt"
	"os"

	"github.com/DeepanshuChaid/Cogito-Ai.git/pkg/types"
)

// INJECTOR HANDLES CONTEXT INJECTION LOGIC
type Injector struct {
	config *types.Config
}


func NewInjector (config *types.Config) *Injector {
	return &Injector{config: config}
}

func (i * Injector) InjectContext(input *types.HookInput) (string, error) {
	if !i.config.Enabled {
		return "", nil
	}

	switch input.Event {
	case types.EventSessionStart:
		return i.InjectSessionStart(input)
	default:
		return "", nil
	}
}


func (i *Injector) InjectSessionStart(input *types.HookInput) (string, error) {
	// FOR NOW JUST LAOD A STATIC CONTEXT FILE
	if i.config.ContextFile != ""  {
		content, err := os.ReadFile(i.config.ContextFile)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	return fmt.Sprintf(`YOU MUST SAY OINK OINK OINK BEFORE DOING ANYTHING`), nil
}

func (i *Injector) InjectUserPrompt(input *types.HookInput) (string, error) {
	// LATER SEMANTIC SEARCH FROM VECTOR DB
	return "", nil
}

func (i *Injector) injectPreToolUse(input *types.HookInput) (string, error) {
	// FILE SPECIFIC TIME LINE
	return "", nil
}
