# SexyVoice CLI

Command-line interface for the [SexyVoice](https://sexyvoice.ai) `/api/v1/speech` API.

It handles:
- browser-based login that rotates or creates an API key through the SexyVoice dashboard
- speech generation from `POST /api/v1/speech`
- curl generation for debugging and integrations
- local audio playback and file saving

## Install

### From Source

```bash
go install github.com/gianpaj/sexyvoice-cli@latest
```

### Shell Script

```bash
curl -fsSL https://raw.githubusercontent.com/gianpaj/sexyvoice-cli/main/install.sh | sh
```

To install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/gianpaj/sexyvoice-cli/main/install.sh | sh -s v0.1.0
```

### Homebrew

If you publish the formula:

```bash
brew install gianpaj/sexyvoice-cli/sexyvoice-cli
```

## Quick Start

```bash
# Open the dashboard login flow and save an API key locally
sexyvoice login

# Generate speech and play it immediately
sexyvoice tts "Hello from SexyVoice" --voice poe --model-id gpro

# Save to a file instead
sexyvoice tts "Hello from SexyVoice" --voice poe --model-id gpro -o hello.wav

# Generate the equivalent curl command
sexyvoice curl "Hello from SexyVoice" --voice poe --model-id gpro --format wav
```

## Commands

### `sexyvoice login`

Opens the SexyVoice dashboard in your browser, lets you reuse an existing API key by rotating it, and saves the replacement key locally.

### `sexyvoice tts TEXT`

Generate speech audio from text.

Common flags:

| Flag | Short | Description |
|------|-------|-------------|
| `--voice` | `-v` | Voice name to use |
| `--model-id` | `-m` | Model ID: `gpro`, `grok`, or `orpheus` |
| `--format` | `-f` | Response format: `wav` or `mp3` |
| `--style` | | Optional style prompt |
| `--seed` | | Optional deterministic seed |
| `--output` | `-o` | Save to file, or use `-` for stdout |
| `--play` | `-p` | Play audio locally |
| `--json` | | Emit CLI result metadata as JSON |

### `sexyvoice curl TEXT`

Print the equivalent `curl` request body for the SexyVoice speech endpoint.

### `sexyvoice key`

Print the resolved API key with no trailing newline.

### `sexyvoice config`

Manage config environments in `~/.sexyvoice/sexyvoice.toml`.

### `sexyvoice play FILE`

Play a local WAV or MP3 file.

### `sexyvoice logout`

Remove the saved local API key.

## Configuration

| Item | Location |
|------|----------|
| Config directory | `~/.sexyvoice/` |
| Config file | `~/.sexyvoice/sexyvoice.toml` |
| Env API key override | `SEXYVOICE_API_KEY` |
| Env API URL override | `SEXYVOICE_API_URL` |
| Env auth prefix override | `SEXYVOICE_AUTH_HEADER_PREFIX` |

Default API URL:

```text
https://sexyvoice.ai/api/v1/speech
```

## Notes

- The CLI validates API keys against `GET /api/v1/models`.
- The speech endpoint returns a JSON payload with a file URL; the CLI follows that URL and downloads the audio for playback or saving.
- The current fork intentionally narrows the surface area to the public speech API rather than preserving Rime-specific commands like `hello`, `speedtest`, `usage`, or `upgrade`.
