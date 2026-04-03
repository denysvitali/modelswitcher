# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this repository.

## Project

**modelswitch** — a terminal UI for switching LLM providers and models, designed for Claude Code users who want to route requests through different backends (OpenRouter, Anthropic, OpenAI, or custom providers).

## Commands

```bash
go build ./cmd/modelswitch          # Build
go run ./cmd/modelswitch            # Run
```

No test framework or linter is configured.

## Architecture

Bubble Tea TUI with three modes in a single `tea.Model`:

| File | Purpose |
|------|---------|
| `cmd/modelswitch/modelswitch.go` | Entrypoint — loads config, starts Bubble Tea program |
| `cmd/modelswitch/config.go` | TOML config types (`Config`, `Provider`, `Preset`, `ActiveConfig`), load/save, `WriteActiveEnv` generation |
| `cmd/modelswitch/state.go` | Bubble Tea model struct, mode constants, helper methods |
| `cmd/modelswitch/update.go` | Event handlers (`Update`, `handleKey`), key bindings per mode, `activatePreset`, `deleteSelected`, `fetchModelsCmd` |
| `cmd/modelswitch/view.go` | TUI rendering for all three modes (`viewProviderSelect`, `viewOpenRouterBrowse`, `viewAddPreset`) + lipgloss styles |
| `cmd/modelswitch/fetcher.go` | HTTP client for fetching OpenRouter models (JSON endpoint, YAML fallback) |

## Key paths

- Config: `~/.config/modelswitch/config.toml`
- Active env (auto-generated): `~/.claude/active-env.sh`

## Three modes

1. **ModeProviderSelect** — Navigate providers/presets, expand, activate
2. **ModeOpenRouterBrowse** — Live model list from openrouter.ai with search/filter
3. **ModeAddPreset** — Form to create a new preset (name, model ID, API key)

## Dependencies

- `charmbracelet/bubbletea` — TUI framework
- `charmbracelet/lipgloss` — Terminal styling
- `pelletier/go-toml/v2` — TOML parsing/writing
