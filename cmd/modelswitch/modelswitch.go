package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
)

func main() {
	configPath := DefaultConfigPath()
	activeEnvPath := DefaultActiveEnvPath()

	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Migrate any plaintext API keys to system keyring
	if err := MigratePlaintextKeys(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "warning: keyring migration skipped: %v\n", err)
	}

	model := NewModel(cfg, configPath, activeEnvPath)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running program: %v\n", err)
		os.Exit(1)
	}
}
