package injector

import (
	"strings"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
)

var IntensityPrompts = map[config.Intensity]string{
	config.IntensityLite: `[MODE: LITE]
Goal: Professional brevity.
Rules:
- Drop filler words (just, really, basically, essentially).
- Drop pleasantries ("I'd be happy to", "Sure thing").
- Keep basic grammar and sentence structure.
- Technical accuracy > politeness.`,

	config.IntensityNormal: `[MODE: FULL CAVEMAN]
Goal: Maximum density.
Rules:
- Pattern: [thing] [action] [reason]. [next step].
- Drop ALL articles (a, an, the).
- Drop fillers, pleasantries, and hedging ("it might be", "possibly").
- Use fragments. No full sentences.
- Short synonyms only.
- PRESERVE EXACTLY: Code blocks, inline backticks, URLs, file paths, and commands.`,

	config.IntensityUltra: `[MODE: ULTRA]
Goal: Telegraphic compression.
Rules:
- Extreme brevity. Abbreviate everything.
- Use symbols for logic: (→ for leads to, ↑ for increase, ↓ for decrease).
- Remove all non-essential words.
- Only technical substance.
- PRESERVE EXACTLY: Code blocks, inline backticks, URLs, and file paths.`,
}

func BuildFinalPrompt(userQuery string, memories []string, cfg *config.Config) string {
	if !cfg.Enabled {
		return userQuery
	}

	var sb strings.Builder

	prompt, ok := IntensityPrompts[cfg.Intensity]
	if !ok {
		prompt = IntensityPrompts[config.IntensityNormal]
	}

	sb.WriteString("### SYSTEM DIRECTIVE\n")
	sb.WriteString(prompt)
	sb.WriteString("\n\nCRITICAL: Do not revert to polite mode. No filler drift. Code blocks must remain UNCHANGED.\n")

	if len(memories) > 0 {
		sb.WriteString("\n\n### PROJECT KNOWLEDGE\n")
		for _, mem := range memories {
			sb.WriteString("- " + mem + "\n")
		}
	}

	sb.WriteString("\n\n### USER QUERY\n")
	sb.WriteString(userQuery)

	return sb.String()
}
