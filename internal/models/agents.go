package models

import (
	"maps"
	"slices"
)

// AIAssistantDisplayNames maps agent identifiers to their display names.
// This is the single source of truth for supported agents.
var AIAssistantDisplayNames = map[string]string{
	"claude":  "Claude Code",
	"gemini":  "Gemini CLI",
	"copilot": "GitHub Copilot",
	"codex":   "OpenAI Codex",
}

// ListAgents returns the supported AI assistant identifiers (sorted).
func ListAgents() []string {
	keys := slices.Collect(maps.Keys(AIAssistantDisplayNames))
	slices.Sort(keys)
	return keys
}

// IsValidAgent reports whether the identifier is a supported agent.
func IsValidAgent(agent string) bool {
	_, ok := AIAssistantDisplayNames[agent]
	return ok
}
