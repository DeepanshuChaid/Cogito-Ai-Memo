package injector

import (
	"cogito/internals/config"
	"strings"
)

var IntensityPrompts = map[string]string{
	"lite": `[MODE: LITE]
Goal: Professional brevity.
Rules:
- Drop filler words (just, really, basically, essentially).
- Drop pleasantries ("I'd be happy to", "Sure thing").
- Keep basic grammar and sentence structure.
- Technical accuracy > politeness.`,

	"full": `[MODE: FULL CAVEMAN]
Goal: Maximum density.
Rules:
- Pattern: [thing] [action] [reason]. [next step].
- Drop ALL articles (a, an, the).
- Drop fillers, pleasantries, and hedging ("it might be", "possibly").
- Use fragments. No full sentences.
- Short synonyms only (e.g., "fix" instead of "implement a solution for").
- PRESERVE EXACTLY: Code blocks, inline backticks, URLs, file paths, and commands.`,

	"ultra": `[MODE: ULTRA]
Goal: Telegraphic compression.
Rules:
- Extreme brevity. Abbreviate everything.
- Use symbols for logic: (→ for leads to, ↑ for increase, ↓ for decrease, = for equals).
- Remove all non-essential words.
- Only technical substance.
- PRESERVE EXACTLY: Code blocks, inline backticks, URLs, and file paths.`,

	"wenyan": `[MODE: WENYAN]
Goal: Classical Chinese compression.
Rules:
- Use Classical Chinese (文言文) for natural language.
- Keep all technical terms, code, and English identifiers in original English.
- Maximum token efficiency.`,
}

func BuildFinalPrompt(userQuery string, memories []string, cfg *config.Config) string {
	if !cfg.Enabled {
		return userQuery
	}

	var sb strings.Builder

	// 1. Selection of the "Mouth" (Intensity)
	prompt, ok := IntensityPrompts[cfg.Intensity]
	if !ok {
		prompt = IntensityPrompts["full"]
	}

	// 2. The System Instructions
	sb.WriteString("### SYSTEM DIRECTIVE\n")
	sb.WriteString(prompt)
	sb.WriteString("\n\nCRITICAL: Do not revert to polite mode. No filler drift. Code blocks must remain UNCHANGED.\n")

	// 3. The "Brain" (Project Knowledge)
	if len(memories) > 0 {
		sb.WriteString("\n\n### PROJECT KNOWLEDGE\n")
		for _, mem := range memories {
			sb.WriteString("- " + mem + "\n")
		}
	}

	// 4. The Target (User Query)
	sb.WriteString("\n\n### USER QUERY\n")
	sb.WriteString(userQuery)

	return sb.String()
}
