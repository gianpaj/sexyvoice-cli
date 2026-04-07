# AGENTS.md

This repo is the SexyVoice CLI fork of the original Rime CLI.

## Scope

The active product scope is the public SexyVoice speech API only:
- `POST /api/v1/speech`
- browser-driven API-key login through the SexyVoice dashboard
- local config, curl generation, audio download, and playback

The fork intentionally does not preserve product-specific Rime features like:
- `hello`
- `usage`
- `speedtest`
- `upgrade`
- `changelog`

If legacy files for those commands still exist, treat them as migration leftovers unless they are explicitly brought back into the root command surface.

When in doubt, prefer deleting dead Rime-specific code over keeping unreachable compatibility layers.

## Naming

- Binary name: `sexyvoice`
- Config directory: `~/.sexyvoice`
- Config file: `~/.sexyvoice/sexyvoice.toml`
- Env vars:
  - `SEXYVOICE_API_KEY`
  - `SEXYVOICE_API_URL`
  - `SEXYVOICE_AUTH_HEADER_PREFIX`
- Go module path: `github.com/gianpaj/sexyvoice-cli`

## Login Flow

`sexyvoice login` is not a direct Supabase-token flow.

It works like this:
1. CLI starts a localhost callback server.
2. Browser opens the SexyVoice dashboard CLI login page.
3. User logs in through the site if needed.
4. User selects an existing API key to rotate, or creates a new one if none exists.
5. The dashboard creates a one-time CLI exchange session.
6. The browser redirects back to localhost with an exchange token.
7. CLI redeems the token at `/api/cli-login-sessions/redeem` and stores the returned API key.

Related SexyVoice app files live outside this repo and include:
- `app/[lang]/cli/login/*`
- `app/api/cli-login-sessions/*`
- `lib/api/cli-login.ts`

## Speech Flow

The SexyVoice speech endpoint returns JSON metadata, not raw audio bytes.

CLI request flow:
1. `POST /api/v1/speech`
2. Parse returned JSON `url`
3. Download audio from that URL
4. Save/play/print bytes locally

Do not reintroduce the old Rime assumption that the primary POST response body is audio.

## Tests

Prefer focused tests around the current SexyVoice behavior:
- API client tests should assert JSON response parsing plus audio download
- auth tests should cover exchange-token redemption and localhost callback validation
- command tests should validate `--voice`, `--model-id`, `--format`, `--style`, and `--seed`

Do not preserve old tests that only validate Rime-specific models, flags, or commands.

Before closing out changes, run:
- `gofmt -w` on touched Go files
- `go test ./...`

## Packaging

Release-facing assets should match the SexyVoice fork:
- README examples use `sexyvoice`
- installer installs `sexyvoice`
- Homebrew formula is `sexyvoice-cli`
- GitHub release artifacts are expected as `sexyvoice-<target>.tar.gz`

If the final repo name changes, update:
- `go.mod`
- `README.md`
- `install.sh`
- `Formula/sexyvoice-cli.rb`
- import paths across `*.go`
