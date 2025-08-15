# Repository Guidelines

## Project Structure & Module Organization
- Root module: `simplified-cce`; CLI entry at `main.go`.
- Core files in repo root: `launcher.go`, `config.go`, `ui.go`.
- Tests colocated as `*_test.go` (e.g., `main_test.go`).
- CI/templates in `.github/`; lint config in `.golangci.yml`.
- Runtime config: `~/.codex-env/config.json` with backups under `~/.codex-env/backups/`.

## Build, Test, and Development Commands
- `make build`: Compile the CLI to `./cde`.
- `make test`: Run all unit tests verbosely.
- `make test-coverage`: Generate `coverage.out` and `coverage.html`.
- `make bench`: Execute benchmarks.
- `make fmt` / `make vet`: Format code and run vet checks.
- Optional: `golangci-lint run` for the full lint suite.

## Coding Style & Naming Conventions
- Formatting: `gofmt`/`goimports` via `make fmt`.
- Import groups: stdlib, third‑party, then local (`simplified-cce/...`).
- Naming: exported `CamelCase`, unexported `camelCase`.
- Keep functions small; avoid naked returns in long functions.

## Testing Guidelines
- Use table‑driven tests where practical.
- Names: tests `TestXxx`, benchmarks `BenchmarkXxx`.
- Run subset: `go test -run TestName ./...`.
- Maintain or improve coverage; verify with `make test-coverage`.

## Commit & Pull Request Guidelines
- Commits: concise conventional prefixes — `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`, `style:`.
- Keep commits small and focused; explain “why” when non‑obvious.
- PRs: clear description, linked issues (e.g., `Closes #123`), tests for changes, and CLI output/screenshots if behavior changes.
- Ensure `make quality` (and lints) pass before requesting review.

## Security & Configuration Tips
- Do not commit secrets. Configuration lives under `~/.codex-env/` with `0700/0600` permissions enforced.
- Backups are created before edits; use secure file handling patterns in `config.go`.
- Run security checks with `make test-security` and consider `golangci-lint run` (includes `gosec`).
