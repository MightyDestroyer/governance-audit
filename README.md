# Governance Audit CLI

> Teil des [MightyDestroyer governance](https://github.com/MightyDestroyer/governance) Oekosystems.

Prueft ein Repository gegen die MightyDestroyer Governance-Standards und berechnet einen Compliance-Score (0–100).

## Quick Links

| Service | URL |
|---------|-----|
| GitHub | [MightyDestroyer/governance-audit](https://github.com/MightyDestroyer/governance-audit) |
| Governance Wiki | [mightydestroyer.github.io/governance](https://mightydestroyer.github.io/governance/) |
| Governance Repo | [MightyDestroyer/governance](https://github.com/MightyDestroyer/governance) |

## Installation

### Aus Source bauen

```bash
cd tools/governance-audit
go build -o governance-audit .
```

### Via go install

```bash
go install github.com/MightyDestroyer/governance-audit@latest
```

## Usage

```bash
# Aktuelles Verzeichnis pruefen
governance-audit

# Bestimmtes Repository pruefen
governance-audit -repo /path/to/repo

# Mit expliziter governance.yaml
governance-audit -repo /path/to/repo -governance /path/to/governance.yaml

# JSON-Ausgabe (fuer CI/CD Pipelines)
governance-audit -repo /path/to/repo -format json

# Detaillierte Ausgabe
governance-audit -repo /path/to/repo -verbose
```

### Flags

| Flag | Default | Beschreibung |
|------|---------|-------------|
| `-repo` | `.` | Pfad zum Repository |
| `-governance` | (auto) | Pfad zur governance.yaml |
| `-format` | `text` | Ausgabeformat: `text` oder `json` |
| `-verbose` | `false` | Detaillierte Ausgabe mit allen Meldungen |
| `-version` | — | Versionsinformation anzeigen |

## Check-Kategorien

| Kategorie | Gewicht | Pruefungen |
|-----------|---------|-----------|
| **Structure** | 20 | README.md, docs/, CLAUDE.md, .github/, .gitignore |
| **Naming** | 15 | Kebab-case Konvention, keine Leerzeichen |
| **Security** | 20 | Keine Secrets, .gitignore-Patterns, kein .env committed |
| **Documentation** | 15 | README-Inhalt (>10 Zeilen), project-bible.md, ADR-Verzeichnis |
| **Contracts** | 15 | Contract/Schema-Dateien (YAML, JSON, SysML) |
| **Observability** | 15 | Health-Endpoints, Structured Logging, Monitoring-Config |

## Score-Interpretation

| Score | Rating | Bedeutung |
|-------|--------|-----------|
| >= 90 | **Excellent** | Vollstaendig konform |
| >= 70 | **Good** | Konform, kleine Verbesserungen moeglich |
| >= 50 | **Needs Work** | Wesentliche Luecken vorhanden |
| < 50 | **Critical** | Governance-Anforderungen nicht erfuellt |

Das Tool gibt Exit-Code `0` bei Score >= 70, Exit-Code `1` bei Score < 70 zurueck. Damit ist es direkt in CI/CD Pipelines einsetzbar.

## JSON Output

```json
{
  "version": "1.0.0",
  "repo": "/path/to/repo",
  "score": 82,
  "rating": "Excellent",
  "categories": [
    {
      "name": "Structure",
      "weight": 20,
      "checks": [
        {
          "name": "README.md exists",
          "passed": true,
          "message": "README.md found"
        }
      ]
    }
  ]
}
```

## CI/CD Integration

```yaml
# GitHub Actions Beispiel
- name: Governance Audit
  run: |
    governance-audit -repo . -format json > audit-report.json
    governance-audit -repo . # Exit code 1 bei Score < 70
```

## Referenzen

- [governance.yaml](../../governance.yaml) — Governance-Konfiguration
- [standards/tool-development.md](../../standards/tool-development.md) — Tool-Development-Standards
- [templates/repo-structure.md](../../templates/repo-structure.md) — Referenz-Ordnerstruktur

---

*MightyDestroyer Governance · Future First · Keine technischen Schulden*
