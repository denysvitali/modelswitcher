package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Padding(0, 0, 1, 0)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Background(lipgloss.Color("236")).
			Padding(0, 1, 0, 0).
			Margin(0, 0)

	activeDotStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 0, 0, 2)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(1, 0, 0, 0)

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Padding(1, 0, 0, 0)

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("235")).
			Padding(0, 1, 0, 1).
			Margin(1, 0, 1, 0)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(1, 0, 0, 0)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("235")).
			Padding(0, 1, 0, 1)

	doneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true).
			Padding(1, 0, 0, 0)
)

func (m *Model) viewProviderSelect() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("modelswitch"))
	b.WriteString(headerStyle.Render("Select a provider or preset:"))

	providers := m.providers()
	providerNames := []string{"openrouter", "anthropic", "openai"}

	// Build ordered list
	ordered := []string{}
	seen := make(map[string]bool)
	for _, name := range providerNames {
		if _, ok := m.cfg.Provider[name]; ok && !seen[name] {
			ordered = append(ordered, name)
			seen[name] = true
		}
	}
	for _, name := range providers {
		if !seen[name] {
			ordered = append(ordered, name)
			seen[name] = true
		}
	}

	// Default providers if none configured
	if len(ordered) == 0 {
		ordered = []string{"openrouter", "anthropic", "openai"}
	}

	hasPresets := false
	for _, pname := range ordered {
		if _, ok := m.cfg.Provider[pname]; ok && len(m.presetsFor(pname)) > 0 {
			hasPresets = true
			break
		}
	}

	// Print providers
	for i, pname := range ordered {
		isSelected := i == m.providerListIndex && !m.presetExpanded
		prefix := "  "
		style := normalStyle
		if isSelected {
			prefix = "▸ "
			style = selectedStyle
		}

		hasConfiguredPresets := false
		presets := m.presetsFor(pname)
		if len(presets) > 0 {
			hasConfiguredPresets = true
		}

		// Show active dot for active preset
		activePrefix := ""
		if hasConfiguredPresets {
			for _, pr := range presets {
				if m.isActive(pname, pr.Name) {
					activePrefix = activeDotStyle.Render("●") + " "
					break
				}
			}
		}

		line := style.Render(prefix + strings.ToUpper(pname[:1]) + pname[1:])
		if hasConfiguredPresets {
			line += dimStyle.Render(" ▶")
		}
		if activePrefix != "" {
			b.WriteString(activePrefix)
		}
		b.WriteString(line + "\n")

		// Expanded presets
		if isSelected && hasConfiguredPresets && m.presetExpanded {
			for j, pr := range presets {
				pIsSelected := j == m.presetListIndex
				ppfx := "    "
				pstyle := dimStyle
				if pIsSelected {
					ppfx = "  ▸ "
					pstyle = selectedStyle
				}
				dot := "  "
				if m.isActive(pname, pr.Name) {
					dot = activeDotStyle.Render("●") + " "
				}
				b.WriteString(dot + pstyle.Render(ppfx+pr.Name) + "\n")
				if pIsSelected && pr.ModelName != "" {
					b.WriteString(descStyle.Render(pr.ModelName))
					if pr.ModelDesc != "" {
						desc := pr.ModelDesc
						if len(desc) > 60 {
							desc = desc[:60] + "…"
						}
						b.WriteString(descStyle.Render("    "+desc))
					}
					b.WriteString("\n")
				}
			}
		}
	}

	if !hasPresets {
		b.WriteString(dimStyle.Render("\n  (no presets saved yet — select a provider first)\n"))
	}

	footer := "↑↓ navigate  → expand  Enter select  a add preset  d delete  q quit"
	if m.presetExpanded {
		footer = "↑↓ navigate  Enter select preset  q back"
	}
	b.WriteString(footerStyle.Render(footer))

	return b.String()
}

func (m *Model) viewOpenRouterBrowse() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("modelswitch — OpenRouter Models"))
	b.WriteString(headerStyle.Render("Live model list from openrouter.ai"))

	if m.fetching {
		frame := spinnerFrames[m.spinnerFrame]
		b.WriteString(searchStyle.Render(frame + " Fetching models..."))
	} else if m.fetchError != "" {
		b.WriteString(errStyle.Render("✘ "+m.fetchError))
		b.WriteString(dimStyle.Render("Press r to retry"))
	} else {
		// Search bar
		placeholder := "/ search models..."
		if m.searchQuery != "" {
			placeholder = m.searchQuery
		}
		searchBar := fmt.Sprintf(" %s ", placeholder)
		b.WriteString(searchStyle.Render(searchBar))

		// Model count
		count := len(m.filteredModels)
		if m.searchQuery != "" {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  %d models matching", count)))
		}

		b.WriteString("\n")

		// Show models (limited to visible area)
		maxVisible := 15
		start := m.modelIndex - maxVisible/2
		if start < 0 {
			start = 0
		}
		end := start + maxVisible
		if end > count {
			end = count
			if end-maxVisible > 0 {
				start = end - maxVisible
			}
		}

		for i := start; i < end; i++ {
			mod := m.filteredModels[i]
			isSelected := i == m.modelIndex
			pfx := "  "
			style := normalStyle
			if isSelected {
				pfx = "▸ "
				style = selectedStyle
			}

			nameDisplay := mod.Name
			if nameDisplay == "" {
				parts := strings.Split(mod.ID, "/")
				nameDisplay = parts[len(parts)-1]
				nameDisplay = strings.ReplaceAll(nameDisplay, "-", " ")
				nameDisplay = strings.ToUpper(nameDisplay[:1]) + nameDisplay[1:]
			}

			b.WriteString(style.Render(pfx + nameDisplay))
			b.WriteString(dimStyle.Render("  " + mod.ID))
			b.WriteString("\n")

			if isSelected && mod.Description != "" {
				desc := mod.Description
				if len(desc) > 80 {
					desc = desc[:80] + "…"
				}
				b.WriteString(descStyle.Render(desc))
				b.WriteString("\n")
			}
		}

		if count == 0 && m.searchQuery != "" {
			b.WriteString(dimStyle.Render("  no models match your search"))
		}
	}

	b.WriteString(footerStyle.Render("↑↓ navigate  Enter select model  / search  q back"))

	return b.String()
}

func (m *Model) viewAddPreset() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("modelswitch — Add New Preset"))
	b.WriteString(headerStyle.Render("Create a new preset"))
	b.WriteString("\n")

	fields := []struct {
		label    string
		value    string
		maxLen   int
		isCursor bool
	}{
		{"Preset Name", m.newPresetName, 30, m.newPresetID == ""},
		{"Model ID", m.newPresetID, 50, m.newPresetID != "" && m.newPresetKey == ""},
		{"API Key (keychain)", m.newPresetKey, 40, m.newPresetKey != ""},
	}

	for i, f := range fields {
		prefix := fmt.Sprintf("  %d. %-18s ", i+1, f.label+":")
		if f.value == "" {
			b.WriteString(dimStyle.Render(prefix + "(required)"))
		} else {
			style := inputStyle
			if f.isCursor {
				style = selectedStyle
			}
			display := f.value
			if i == 2 { // API Key field — mask it
				display = strings.Repeat("*", len(f.value))
			}
			b.WriteString(style.Render(prefix + display + "_"))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(footerStyle.Render("Tab next field  Enter save  Esc cancel"))

	return b.String()
}

func (m *Model) View() string {
	if m.doneMessage != "" {
		return doneStyle.Render(m.doneMessage)
	}

	switch m.mode {
	case ModeProviderSelect:
		return m.viewProviderSelect()
	case ModeOpenRouterBrowse:
		return m.viewOpenRouterBrowse()
	case ModeAddPreset:
		return m.viewAddPreset()
	}
	return ""
}
