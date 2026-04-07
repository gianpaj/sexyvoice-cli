# SexyVoice CLI

Command-line interface for the [SexyVoice](https://sexyvoice.ai) `/api/v1/speech` API.

It handles:
- browser-based login that rotates or creates an API key through the SexyVoice dashboard
- listing available voices from `GET /api/v1/voices`
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

# List available voices
sexyvoice voices

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

### `sexyvoice voices`

List voices available from the SexyVoice voices endpoint.

Examples:

```text
❯ go run . voices
NAME        MODEL    LANGUAGE  FORMATS  STYLE  ID
achernar    gpro     multiple  wav      yes    390f5864-111b-4795-81ea-7026a1e64cfc
aoede       gpro     multiple  wav      yes    5797178d-c047-44ea-aef4-94e97fb48663
autonoe     gpro     multiple  wav      yes    cd71a399-9c9a-4384-b76e-14e96a45fc8a
callirrhoe  gpro     multiple  wav      yes    c38d0c52-e8c1-404b-9cb1-56d4cf252b9f
despina     gpro     multiple  wav      yes    f586bd2c-909f-4bf3-a859-4e45e45a22d6
erinome     gpro     multiple  wav      yes    530c8354-bb32-4231-a9c9-3c05fcdd220b
gacrux      gpro     multiple  wav      yes    3da76670-6fca-435b-b097-eed8cc6f37a5
kore        gpro     multiple  wav      yes    5a2bffef-5e60-4fe6-b989-9ed3e68b6c48
puck        gpro     multiple  wav      yes    0111e4dd-a38a-48e0-a8bb-a0a057f4cceb
sulafat     gpro     multiple  wav      yes    a68482ea-c14e-43cf-bf8c-16074bff9f8d
zephyr      gpro     multiple  wav      yes    03925a20-4e15-4745-b86c-2c2e9d7de870
ara         grok     multiple  mp3,wav  no     9d028af7-3c3e-466a-b53f-9a08dc66bf30
eve         grok     multiple  mp3,wav  no     9aff7755-72c9-4d90-bcf1-c06e18e769b6
leo         grok     multiple  mp3,wav  no     fd965707-5367-4b89-a05a-c8cd086ba5ab
rex         grok     multiple  mp3,wav  no     1391aaf8-d3b5-41a5-8554-2c0c9b6d099d
sal         grok     multiple  mp3,wav  no     57085a6a-2396-445c-a6c1-fd0ce8368d72
dan         orpheus  en-GB 🇬🇧  mp3      no     218f9750-e7bc-4ae5-9e1d-c6f9d4ffaa74
emma        orpheus  en-US 🇺🇸  mp3      no     698e2e46-7c37-4ff4-ad15-558b30514dea
javi        orpheus  es-ES 🇪🇸  mp3      no     969a0811-7c67-403e-a3c0-666462e72cb7
josh        orpheus  en-US 🇺🇸  mp3      no     a5bd6e81-0d1e-490d-b669-bc383ee4ec7c
pietro      orpheus  it-IT 🇮🇹  mp3      no     3b4edf58-37da-440a-96ed-353b52c080e9
tara        orpheus  en-US 🇺🇸  mp3      no     e55e4fc7-c140-46eb-82da-4d8c97f03d95
```

Use `--json` for raw machine-readable output.

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
