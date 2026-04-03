package main

import (
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
)

const (
	ModeProviderSelect = iota
	ModeOpenRouterBrowse
	ModeAddPreset
)

var defaultProviders = []string{"openrouter", "anthropic", "openai"}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type spinnerTickMsg struct{}

func spinnerCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return spinnerTickMsg{} })
}

type Model struct {
	cfg             *Config
	configPath      string
	activeEnvPath   string
	mode            int
	selectedProvider string
	expandedProvider string

	// Provider list
	providerListIndex int
	// Preset list
	presetListIndex int
	presetExpanded  bool

	// OpenRouter browse
	models         []OpenRouterModel
	filteredModels []OpenRouterModel
	modelIndex     int
	fetching       bool
	fetchError     string
	searchQuery    string
	searchMode     bool
	spinnerFrame   int
	lastSpinnerUpdate time.Time

	// Add preset form
	newPresetName string
	newPresetID   string
	newPresetKey  string

	// General
	quitting    bool
	doneMessage string
}

func NewModel(cfg *Config, configPath, activeEnvPath string) *Model {
	return &Model{
		cfg:           cfg,
		configPath:    configPath,
		activeEnvPath: activeEnvPath,
		mode:          ModeProviderSelect,
		providerListIndex: 0,
		presetListIndex: 0,
	}
}

func (m *Model) setMode(mode int) {
	m.mode = mode
	m.modelIndex = 0
	m.searchQuery = ""
	m.searchMode = false
	m.filteredModels = nil
	m.fetchError = ""
	m.fetching = false
	m.presetExpanded = false
	m.presetListIndex = 0
}

func (m *Model) providers() []string {
	providers := make([]string, 0, len(m.cfg.Provider))
	for name := range m.cfg.Provider {
		providers = append(providers, name)
	}
	sort.Strings(providers)
	return providers
}

func (m *Model) presetsFor(provider string) []Preset {
	p, ok := m.cfg.Provider[provider]
	if !ok {
		return nil
	}
	return p.Presets
}

func (m *Model) activePreset() *Preset {
	if m.cfg.Active.Name == "" || m.cfg.Active.Provider == "" {
		return nil
	}
	presets := m.presetsFor(m.cfg.Active.Provider)
	for i := range presets {
		if presets[i].Name == m.cfg.Active.Name {
			return &presets[i]
		}
	}
	return nil
}

func (m *Model) isActive(provider, presetName string) bool {
	return m.cfg.Active.Provider == provider && m.cfg.Active.Name == presetName
}

func (m *Model) updateFiltered() {
	if m.models == nil {
		m.filteredModels = nil
		return
	}
	if m.searchQuery == "" {
		m.filteredModels = m.models
	} else {
		q := strings.ToLower(m.searchQuery)
		filtered := make([]OpenRouterModel, 0, len(m.models))
		for _, mod := range m.models {
			if strings.Contains(strings.ToLower(mod.ID), q) ||
				strings.Contains(strings.ToLower(mod.Name), q) ||
				strings.Contains(strings.ToLower(mod.Description), q) {
				filtered = append(filtered, mod)
			}
		}
		m.filteredModels = filtered
	}
}

func (m *Model) Init() tea.Cmd { return nil }
