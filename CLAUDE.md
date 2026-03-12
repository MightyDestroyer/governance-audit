# CLAUDE.md — Governance Audit CLI

## Purpose

CLI tool that audits a project repository against MightyDestroyer governance standards and calculates a compliance score (0–100).

## Build

```bash
go build -o governance-audit .
```

## Run

```bash
./governance-audit -repo /path/to/repo
./governance-audit -repo /path/to/repo -format json
./governance-audit -repo /path/to/repo -verbose
```

## Architecture

| File | Responsibility |
|------|---------------|
| `main.go` | CLI entry point, flag parsing, orchestration |
| `checker.go` | Check definitions and execution logic |
| `report.go` | Text and JSON report generation |

## Check Categories

- **Structure** (weight 20): Required files and directories
- **Naming** (weight 15): Kebab-case convention enforcement
- **Security** (weight 20): Secret scanning, .gitignore patterns
- **Documentation** (weight 15): README content, project bible, ADRs
- **Contracts** (weight 15): API specs, schemas, SysML models
- **Observability** (weight 15): Health endpoints, logging, monitoring config

## Conventions

- Go 1.22, no external dependencies beyond `gopkg.in/yaml.v3`
- Structured logging via `log/slog` (JSON to stderr)
- Exit code 0 if score >= 70, exit code 1 if below
- Follows `standards/tool-development.md`

## Scoring

`score = sum(category_weight * passed / total) / sum(weights) * 100`
