package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerTickMsg:
		m.lastSpinnerUpdate = time.Now()
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
		return m, spinnerCmd()

	case tea.KeyMsg:
		return m.handleKey(msg)

	case fetchModelsMsg:
		m.fetching = false
		if msg.err != nil {
			m.fetchError = msg.err.Error()
		} else {
			m.models = msg.models
			m.updateFiltered()
			m.fetchError = ""
		}
		return m, nil

	case tea.WindowSizeMsg:
		return m, nil
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	switch m.mode {

	case ModeProviderSelect:
		return m.handleProviderSelectKey(msg)

	case ModeOpenRouterBrowse:
		return m.handleOpenRouterBrowseKey(msg)

	case ModeAddPreset:
		return m.handleAddPresetKey(msg)
	}

	return m, nil
}

// ─── Provider Select ─────────────────────────────────────────────────────────

func (m *Model) handleProviderSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {

	case tea.KeyUp:
		if m.presetExpanded {
			if m.presetListIndex > 0 {
				m.presetListIndex--
			}
		} else {
			if m.providerListIndex > 0 {
				m.providerListIndex--
			}
		}

	case tea.KeyDown:
		if m.presetExpanded {
			presets := m.presetsFor(m.expandedProvider)
			if m.presetListIndex < len(presets)-1 {
				m.presetListIndex++
			}
		} else {
			allProviders := m.allProviderList()
			if m.providerListIndex < len(allProviders)-1 {
				m.providerListIndex++
			}
		}

	case tea.KeyRight, tea.KeyEnter:
		if m.presetExpanded {
			// Select preset
			presets := m.presetsFor(m.expandedProvider)
			if len(presets) > 0 && m.presetListIndex < len(presets) {
				pr := presets[m.presetListIndex]
				return m.activatePreset(m.expandedProvider, &pr)
			}
			m.presetExpanded = false
			return m, nil
		}

		// Expand provider or enter browser
		allProviders := m.allProviderList()
		if m.providerListIndex >= len(allProviders) {
			return m, nil
		}
		pname := allProviders[m.providerListIndex]
		presets := m.presetsFor(pname)

		if len(presets) > 0 {
			m.presetExpanded = true
			m.expandedProvider = pname
			m.presetListIndex = 0
		} else if info, ok := ProviderInfoFor(pname); ok && info.SupportsBrowse {
			return m.enterOpenRouterBrowse(pname)
		} else {
			m.selectedProvider = pname
			m.newPresetName = ""
			m.newPresetID = ""
			m.newPresetKey = ""
			m.setMode(ModeAddPreset)
		}

	case tea.KeyLeft:
		if m.presetExpanded {
			m.presetExpanded = false
		}

	case tea.KeyEsc, tea.KeyRunes:
		if msg.Type == tea.KeyRunes && msg.Runes[0] == 'q' {
			m.quitting = true
			return m, tea.Quit
		}
		if msg.Type == tea.KeyRunes && msg.Runes[0] == 'a' {
			// Add preset — go to add mode for selected provider
			allProviders := m.allProviderList()
			if m.providerListIndex < len(allProviders) {
				m.selectedProvider = allProviders[m.providerListIndex]
			}
			m.newPresetName = ""
			m.newPresetID = ""
			m.newPresetKey = ""
			m.setMode(ModeAddPreset)
			return m, nil
		}
		if msg.Type == tea.KeyRunes && msg.Runes[0] == 'd' {
			return m.deleteSelected()
		}
		if msg.Type == tea.KeyRunes && msg.Runes[0] == 'r' {
			// Refresh browser
			allProviders := m.allProviderList()
			if m.providerListIndex < len(allProviders) {
				if info, ok := ProviderInfoFor(allProviders[m.providerListIndex]); ok && info.SupportsBrowse { return m.enterOpenRouterBrowse(allProviders[m.providerListIndex]) }
			}
		}
	}

	return m, nil
}

func (m *Model) enterOpenRouterBrowse(provider string) (tea.Model, tea.Cmd) {
	m.selectedProvider = provider
	m.setMode(ModeOpenRouterBrowse)
	m.fetching = true
	m.fetchError = ""

	fetcher := NewFetcher()
	apiKey, _ := ResolveAPIKey(provider, m.cfg.Provider[provider])

	return m, tea.Batch(spinnerCmd(), fetchModelsCmd(fetcher, apiKey))
}

// ─── OpenRouter Browse ───────────────────────────────────────────────────────

func (m *Model) handleOpenRouterBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In search mode, only Esc/Backspace/Enter/arrows work; everything else types
	if m.searchMode && msg.Type == tea.KeyRunes {
		m.searchQuery += string(msg.Runes)
		m.updateFiltered()
		return m, nil
	}

	switch msg.Type {

	case tea.KeyUp:
		if m.modelIndex > 0 {
			m.modelIndex--
		}

	case tea.KeyDown:
		if m.modelIndex < len(m.filteredModels)-1 {
			m.modelIndex++
		}

	case tea.KeyPgUp:
		if m.modelIndex >= 10 {
			m.modelIndex -= 10
		} else {
			m.modelIndex = 0
		}

	case tea.KeyPgDown:
		if m.modelIndex+10 < len(m.filteredModels) {
			m.modelIndex += 10
		} else {
			m.modelIndex = len(m.filteredModels) - 1
		}

	case tea.KeyEnter:
		if len(m.filteredModels) > 0 && m.modelIndex < len(m.filteredModels) {
			mod := m.filteredModels[m.modelIndex]
			preset := Preset{
				Name:      mod.ID,
				ModelID:   mod.ID,
				ModelName: mod.Name,
				ModelDesc: mod.Description,
			}
			return m.activatePreset(m.selectedProvider, &preset)
		}

	case tea.KeyEsc:
		if m.searchMode || m.searchQuery != "" {
			m.searchMode = false
			m.searchQuery = ""
			m.updateFiltered()
		} else {
			m.setMode(ModeProviderSelect)
		}
		return m, nil

	case tea.KeyRunes:
		switch msg.Runes[0] {
		case 'r':
			m.fetching = true
			m.fetchError = ""
			fetcher := NewFetcher()
			apiKey, _ := ResolveAPIKey(m.selectedProvider, m.cfg.Provider[m.selectedProvider])
			return m, tea.Batch(spinnerCmd(), fetchModelsCmd(fetcher, apiKey))
		case '/':
			m.searchMode = true
			return m, nil
		default:
			m.searchQuery += string(msg.Runes)
			m.updateFiltered()
		}

	case tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.updateFiltered()
		} else if m.searchMode {
			m.searchMode = false
			m.updateFiltered()
		}
	}

	return m, nil
}

// ─── Add Preset ─────────────────────────────────────────────────────────────

func (m *Model) handleAddPresetKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {

	case tea.KeyTab:
		m.focusedField = (m.focusedField + 1) % 3

	case tea.KeyUp:
		if m.focusedField > 0 {
			m.focusedField--
		}

	case tea.KeyDown:
		if m.focusedField < 2 {
			m.focusedField++
		}

	case tea.KeyEnter:
		if m.newPresetName != "" && m.newPresetID != "" {
			provider := m.selectedProvider
			if provider == "" {
				provider = "openrouter"
			}
			preset := Preset{
				Name:      m.newPresetName,
				ModelID:   m.newPresetID,
				ModelName: m.newPresetName,
			}
			p := m.cfg.Provider[provider]
			if m.newPresetKey != "" {
				if err := KeyringSet(provider, m.newPresetKey); err != nil {
					m.fetchError = "failed to store key in keychain: " + err.Error()
					return m, nil
				}
				p.UseKeyring = true
			}
			p.Presets = append(p.Presets, preset)
			m.cfg.Provider[provider] = p
			if err := SaveConfig(m.configPath, m.cfg); err != nil {
				m.fetchError = "failed to save: " + err.Error()
			}
			m.setMode(ModeProviderSelect)
		}

	case tea.KeyEsc:
		m.setMode(ModeProviderSelect)

		case tea.KeyRunes:
			text := string(msg.Runes)
			switch m.focusedField {
			case 0:
				m.newPresetName += text
			case 1:
				m.newPresetID += text
			case 2:
				m.newPresetKey += text
			}

	case tea.KeyBackspace:
		switch m.focusedField {
		case 0:
			if len(m.newPresetName) > 0 {
				m.newPresetName = m.newPresetName[:len(m.newPresetName)-1]
			}
		case 1:
			if len(m.newPresetID) > 0 {
				m.newPresetID = m.newPresetID[:len(m.newPresetID)-1]
			}
		case 2:
			if len(m.newPresetKey) > 0 {
				m.newPresetKey = m.newPresetKey[:len(m.newPresetKey)-1]
			}
		}
	}

	return m, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (m *Model) activatePreset(provider string, preset *Preset) (tea.Model, tea.Cmd) {
	m.cfg.Active.Provider = provider
	m.cfg.Active.Name = preset.Name
	m.cfg.Active.ModelID = preset.ModelID
	m.cfg.Active.ModelName = preset.ModelName

	// Add preset to provider list if not already there
	p := m.cfg.Provider[provider]
	found := false
	for i := range p.Presets {
		if p.Presets[i].Name == preset.Name {
			p.Presets[i] = *preset
			found = true
			break
		}
	}
	if !found {
		p.Presets = append(p.Presets, *preset)
	}
	m.cfg.Provider[provider] = p

	if err := SaveConfig(m.configPath, m.cfg); err != nil {
		m.fetchError = "failed to save config: " + err.Error()
		return m, nil
	}

	if err := WriteActiveEnv(m.activeEnvPath, m.cfg, preset, provider); err != nil {
		m.fetchError = "failed to write env file: " + err.Error()
		return m, nil
	}

	m.doneMessage = fmt.Sprintf(
		"✓ Done! Preset '%s' activated.\n\nRun: source %s",
		preset.Name, m.activeEnvPath,
	)
	return m, tea.Quit
}

func (m *Model) deleteSelected() (tea.Model, tea.Cmd) {
	allProviders := m.allProviderList()
	if m.providerListIndex >= len(allProviders) {
		return m, nil
	}

	if m.presetExpanded {
		pname := m.expandedProvider
		presets := m.presetsFor(pname)
		if len(presets) > 0 && m.presetListIndex < len(presets) {
			pr := presets[m.presetListIndex]
			// Remove preset
			updated := append([]Preset{}, presets[:m.presetListIndex]...)
			updated = append(updated, presets[m.presetListIndex+1:]...)
			p := m.cfg.Provider[pname]
			p.Presets = updated
			m.cfg.Provider[pname] = p
			if m.isActive(pname, pr.Name) {
				m.cfg.Active = ActiveConfig{}
			}
			_ = SaveConfig(m.configPath, m.cfg)
			if len(updated) == 0 {
				m.presetExpanded = false
			} else if m.presetListIndex >= len(updated) {
				m.presetListIndex = len(updated) - 1
			}
		}
	} else {
		pname := allProviders[m.providerListIndex]
		_ = KeyringDelete(pname)
		delete(m.cfg.Provider, pname)
		if m.cfg.Active.Provider == pname {
			m.cfg.Active = ActiveConfig{}
		}
		_ = SaveConfig(m.configPath, m.cfg)
		if len(allProviders) > 0 {
			m.providerListIndex = 0
		}
	}

	return m, nil
}

// ─── Commands ────────────────────────────────────────────────────────────────

type fetchModelsMsg struct {
	models []OpenRouterModel
	err    error
}

func fetchModelsCmd(fetcher *Fetcher, apiKey string) tea.Cmd {
	return func() tea.Msg {
		models, err := fetcher.FetchModels(apiKey)
		if err != nil {
			// Fall back to YAML
			models, err = fetcher.FetchOpenAPIYAML(apiKey)
		}
		if err != nil {
			return fetchModelsMsg{err: err}
		}
		// Sort alphabetically
		sort.Slice(models, func(i, j int) bool {
			return models[i].ID < models[j].ID
		})
		return fetchModelsMsg{models: models}
	}
}
