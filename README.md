# modelswitch

A terminal UI for switching LLM providers and models, designed primarily for **Claude Code** users who want to route requests through different backends (OpenRouter, Anthropic, OpenAI, or custom providers) without changing environment variables manually.

## Why this exists

Claude Code uses `ANTHROPIC_BASE_URL`, `ANTHROPIC_API_KEY`, and model ID environment variables to determine which LLM backend to talk to. By default it targets Anthropic's own API, but — thanks to anthropic-compatible endpoints — you can point it at _any_ provider.

Manually editing `.env` files or exporting variables every time you want to try a different model is tedious. `modelswitch` solves this with a Bubble Tea–based TUI where you:

1. **Browse** available providers and presets
2. **Pick** a model (including live browsing of OpenRouter's catalog)
3. **Activate** it — which writes a ready-to-source `active-env.sh` file

One `source` command and your next `claude` invocation uses the new provider/model.

## How it works

```
┌─────────────────────────────────┐
│  modelswitch (Bubble Tea TUI)   │
│                                 │
│  Config loaded from:            │  ─┐
│  ~/.config/modelswitch/         │   │
│    config.toml                  │   │
│                                 │   │ Read & write
│  User selects a preset ────────>│   │
│                                 │   │
│  Writes to:                     │  ─┘
│  ~/.claude/active-env.sh        │
│    (export ANTHROPIC_* vars)    │
└─────────────────────────────────┘
              │
              │ source
              v
┌─────────────────────────────────┐
│  Claude Code (or any tool that  │
│  reads ANTHROPIC_BASE_URL)      │
│                                 │
│  → routes to chosen provider    │
└─────────────────────────────────┘
```

### Two data sources

When activating a preset, `modelswitch` writes a shell script with `export` statements:

```bash
export ANTHROPIC_BASE_URL="https://openrouter.ai/api"
export ANTHROPIC_API_KEY="sk-or-..."
export ANTHROPIC_DEFAULT_OPUS_MODEL="google/gemini-2.5-pro"
export ANTHROPIC_DEFAULT_SONNET_MODEL="google/gemini-2.5-pro"
export ANTHROPIC_DEFAULT_HAIKU_MODEL="google/gemini-2.5-pro"
```

All three model-role variables (`OPUS`, `SONNET`, `HAIKU`) are set to the same model ID because Claude Code picks among them based on task complexity. When proxying through a non-Anthropic provider, you typically want them all to resolve to the same backend model.

The **Base URL** is determined by the provider:

| Provider    | Default Base URL                  |
|-------------|------------------------------------|
| openrouter  | `https://openrouter.ai/api`        |
| anthropic   | `https://api.anthropic.com`        |
| openai      | `https://api.openai.com/v1`        |

Custom providers can override the base URL in config.

## Installation

```bash
go install github.com/denysvitali/modelswitcher@latest
```

Or build from source:

```bash
git clone https://github.com/denysvitali/modelswitcher.git
cd modelswitcher
go build -o modelswitch ./cmd/modelswitch
```

## Usage

Launch the TUI:

```bash
modelswitch
```

### Main screen — Provider Select

- **↑/↓** — Navigate providers and presets
- **→** — Expand a provider to show its presets
- **Enter** — Activate the selected preset (writes env file and exits)
- **a** — Add a new preset
- **d** — Delete the currently highlighted preset or provider
- **r** — Refresh (re-fetch model list from OpenRouter)
- **q** — Quit

### OpenRouter Browse mode

If a provider has no presets, pressing **→** or **Enter** opens the built-in model browser, which fetches the live list from `openrouter.ai/api/v1/models` (falling back to parsing `openrouter.ai/openapi.yaml` if the JSON endpoint fails).

- **↑/↓ / PgUp/PgDn** — Navigate the model list
- **/** — Enter search mode (filters by model ID, name, or description)
- **Esc** — Exit search mode / go back
- **Enter** — Select the highlighted model, save as a preset, activate, and exit
- **r** — Re-fetch the model list

### Add Preset mode

A simple form with three fields (filled in order):

1. **Preset Name** — A human-friendly label for the preset
2. **Model ID** — The full model identifier (e.g. `google/gemini-2.5-pro`)
3. **API Key** — Optional override key for this provider

- **Tab / ↑/↓** — Move between fields
- **Enter** — Save and go back
- **Esc** — Cancel

After saving, `modelswitch` writes both the TOML config and the env file, returns to the main screen, and prompts you to source the file.

### Activating a model

When you select a preset or model and press **Enter**:

1. The preset is saved (or updated) in `~/.config/modelswitch/config.toml`
2. `~/.claude/active-env.sh` is written with the appropriate `ANTHROPIC_*` environment variables
3. The TUI displays: `source ~/.claude/active-env.sh`

Source the file in your shell, then run `claude` — it will now use the selected provider and model.

```bash
source ~/.claude/active-env.sh
claude "Hello from a different backend"
```

## Configuration

### File: `~/.config/modelswitch/config.toml`

TOML configuration with provider credentials and saved presets:

```toml
[provider.openrouter]
base_url = "https://openrouter.ai/api"
api_key = "sk-or-..."

[[provider.openrouter.preset]]
name = "Gemini 2.5 Pro"
model_id = "google/gemini-2.5-pro"
model_name = "Gemini 2.5 Pro"
model_desc = "Google's most capable model"

[provider.anthropic]
api_key = "sk-ant-..."

[[provider.anthropic.preset]]
name = "Opus"
model_id = "claude-opus-4-6"
model_name = "Claude Opus 4.6"

[active]
name = "Gemini 2.5 Pro"
provider = "openrouter"
model_id = "google/gemini-2.5-pro"
model_name = "Gemini 2.5 Pro"
```

### File: `~/.claude/active-env.sh` (auto-generated)

Written automatically by `modelswitch` when a preset is activated. Do not edit manually — your changes will be overwritten the next time you use the TUI.

## License

MIT
