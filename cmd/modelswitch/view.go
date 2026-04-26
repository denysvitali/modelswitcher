package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("63"))

	activeDotStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("247"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("238"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("238"))

	doneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)
)

func (m *Model) viewProviderSelect() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("modelswitch"))
	b.WriteString("\n\n")
	b.WriteString("Select a provider or preset:\n")

	providers := m.allProviderList()

	hasPresets := false
	for _, pname := range providers {
		if _, ok := m.cfg.Provider[pname]; ok && len(m.presetsFor(pname)) > 0 {
			hasPresets = true
			break
		}
	}

	for i, pname := range providers {
		isCurrent := i == m.providerListIndex
		isSelected := isCurrent && !m.presetExpanded
		isExpanded := isCurrent && m.presetExpanded
		_, isConfigured := m.cfg.Provider[pname]

		presets := m.presetsFor(pname)
		hasPresetsForProvider := len(presets) > 0

		displayName := ""
		if info, ok := ProviderInfoFor(pname); ok {
			displayName = info.DisplayName
		} else {
			displayName = strings.ToUpper(pname[:1]) + pname[1:]
		}

		base := "  " + displayName
		line := ""
		if isSelected {
			base = "› " + displayName
			line = selectedStyle.Render(base)
		} else if isExpanded {
			line = dimStyle.Render(base)
		} else if !isConfigured {
			line = dimStyle.Render(base)
		} else {
			line = normalStyle.Render(base)
		}

		if hasPresetsForProvider {
			line += dimStyle.Render(" ▾")
		} else if isSelected && !isConfigured {
			line += dimStyle.Render(" [press a]")
		}

		if hasPresetsForProvider {
			for _, pr := range presets {
				if m.isActive(pname, pr.Name) {
					b.WriteString(activeDotStyle.Render("●") + " ")
					break
				}
			}
		}

		b.WriteString(line + "\n")

		if isExpanded && hasPresetsForProvider {
			for j, pr := range presets {
				pIsSelected := j == m.presetListIndex
				pBase := "    "
				if pIsSelected {
					pBase = "  › "
				}

				var pLine string
				if pIsSelected {
					pLine = selectedStyle.Render(pBase + pr.Name)
				} else {
					pLine = dimStyle.Render(pBase + pr.Name)
				}

				if m.isActive(pname, pr.Name) {
					b.WriteString(activeDotStyle.Render("●") + " ")
				} else {
					b.WriteString("  ")
				}

				b.WriteString(pLine + "\n")

				if pIsSelected && (pr.ModelDesc != "" || (pr.ModelName != "" && pr.ModelName != pr.Name)) {
					if pr.ModelName != "" && pr.ModelName != pr.Name {
						b.WriteString("     " + descStyle.Render(pr.ModelName))
					}
					if pr.ModelDesc != "" {
						descText := pr.ModelDesc
						if len(descText) > 60 {
							descText = descText[:60] + "…"
						}
						if pr.ModelName != "" && pr.ModelName != pr.Name {
							b.WriteString("    ")
						} else {
							b.WriteString("     ")
						}
						b.WriteString(descStyle.Render(descText))
					}
					b.WriteString("\n")
				}
			}
		}
	}

	if !hasPresets {
		b.WriteString(dimStyle.Render("  (no presets saved yet)\n"))
	}

	footer := "↑↓ navigate  → expand  Enter select  a add preset  d delete  q quit"
	if m.presetExpanded {
		footer = "↑↓ navigate  Enter select preset  q back"
	}
	b.WriteString("\n")
	b.WriteString(footerStyle.Render(footer))

	return b.String()
}

func (m *Model) viewOpenRouterBrowse() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("modelswitch — OpenRouter Models"))
	b.WriteString("\n\n")
	b.WriteString("Live model list from openrouter.ai\n")

	if m.fetching {
		frame := spinnerFrames[m.spinnerFrame]
		b.WriteString("  " + frame + " Fetching models...\n")
	} else if m.fetchError != "" {
		b.WriteString(errStyle.Render("✘ Error: "+m.fetchError) + "\n")
		b.WriteString(dimStyle.Render("  press r to retry\n"))
	} else {
		placeholder := "/ search models..."
		if m.searchQuery != "" {
			placeholder = m.searchQuery
		}
		b.WriteString("  " + searchStyle.Render(" "+placeholder+" ") + "\n")

		count := len(m.filteredModels)
		if m.searchQuery != "" {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  %d models matching", count)) + "\n")
		}

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
			if isSelected {
				pfx = "› "
			}

			nameDisplay := mod.Name
			if nameDisplay == "" {
				parts := strings.Split(mod.ID, "/")
				nameDisplay = parts[len(parts)-1]
				nameDisplay = strings.ReplaceAll(nameDisplay, "-", " ")
				nameDisplay = strings.ToUpper(nameDisplay[:1]) + nameDisplay[1:]
			}

			line := pfx + nameDisplay
			if isSelected {
				b.WriteString(selectedStyle.Render(line))
			} else {
				b.WriteString(normalStyle.Render(line))
			}
			b.WriteString(dimStyle.Render("  " + mod.ID))
			b.WriteString("\n")

			if isSelected && mod.Description != "" {
				desc := mod.Description
				if len(desc) > 80 {
					desc = desc[:80] + "…"
				}
				b.WriteString("   " + descStyle.Render(desc) + "\n")
			}
		}

		if count == 0 && m.searchQuery != "" {
			b.WriteString(dimStyle.Render("  no models match your search\n"))
		}
	}

	b.WriteString("\n")
	b.WriteString(footerStyle.Render("↑↓ navigate  Enter select model  / search  r refresh  Esc back"))

	return b.String()
}

func (m *Model) viewAddPreset() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("modelswitch — Add New Preset"))
	b.WriteString("\n\n")
	b.WriteString("Create a new preset\n")
	b.WriteString("\n")

	fields := []struct {
		label    string
		value    string
		isCursor bool
	}{
		{"Preset Name", m.newPresetName, m.focusedField == 0},
		{"Model ID", m.newPresetID, m.focusedField == 1},
		{"API Key", m.newPresetKey, m.focusedField == 2},
	}

		for i, f := range fields {
			label := fmt.Sprintf("  %d. %s: ", i+1, f.label)
			display := f.value
			if i == 2 && display != "" {
				display = strings.Repeat("*", len(f.value))
			}
			if display == "" {
				display = "(required)"
			}
			cursor := "_"
			rendered := label + display + cursor
			if f.isCursor {
				b.WriteString(selectedStyle.Render(rendered) + "\n")
			} else {
				b.WriteString(normalStyle.Render(rendered) + "\n")
			}
		}

	b.WriteString("\n")
	b.WriteString(footerStyle.Render("Tab next field  Enter save  Esc cancel"))

	return b.String()
}

func (m *Model) View() string {
	if m.doneMessage != "" {
		return "\n" + doneStyle.Render(m.doneMessage) + "\n"
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
