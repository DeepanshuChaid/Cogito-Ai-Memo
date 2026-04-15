package injector

import (
	"fmt"
	"strings"
)

const CavemanRules = `[SYSTEM RULES:
Use Caveman Style.
Pattern: [thing] [action] [reason].
Drop articles (a, an, the).
Drop fillers, pleasantries.
Be terse. Technical fragments only.]`

func BuildFinalPrompt(userQuery string, memories []string) string {
	// Combine rules + any compressed project data + user question
	var sb strings.Builder
	sb.WriteString(CavemanRules)
	sb.WriteString("\n\nPROJECT KNOWLEDGE:\n")

	for _, mem := range memories {
		sb.WriteString("- " + mem + "\n")
		fmt.Println("INJECTOR FUNCTIONS RUNNING")
	}

	sb.WriteString("\nUSER QUERY: " + userQuery)
	fmt.Println("INJECTOR COMPLETED")
	return sb.String()
}
