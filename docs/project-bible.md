# Project Bible — governance-audit

## Uebersicht

- **Projekt:** governance-audit
- **Ziel:** CLI-Tool das Repositories gegen MightyDestroyer Governance-Standards prueft und einen Compliance-Score berechnet
- **Stack:** Go
- **Start:** Maerz 2026
- **Governance:** [MightyDestroyer/governance](https://github.com/MightyDestroyer/governance) v3.5.0

---

## Sprint-Protokoll

### Sprint 1 — 2026-03-12

**Ziel:** Initiale Implementierung des Governance Audit CLI.

**Erreicht:**
- 6 Check-Kategorien (Structure, Naming, Security, Documentation, Contracts, Observability)
- 17 Einzelpruefungen mit gewichtetem Score (0-100)
- Text- und JSON-Output, ANSI-Farben, Verbose-Modus
- governance.yaml als Input (SSoT)

### Sprint 2 — 2026-03-13

**Ziel:** False Positives fixen, Checks smarter machen.

**Erreicht:**
- Go/Python/Rust/Ruby snake_case als valide Naming-Konvention akzeptiert
- `_`-Prefix-Directories (Build-Output) ausgeschlossen
- Health-Endpoint-Check erkennt automatisch ob Repo ein Service ist
- Observability-Check akzeptiert OTel/Prometheus SDK im Source-Code
- Contract-Check akzeptiert Root-Level Specs (governance.yaml, openapi.yaml)
- Konventions-Dateinamen erweitert (SKILL.md, AGENTS.md, CONTRIBUTING.md, etc.)

**Entscheidungen:** Keine False Positives tolerieren — CLI muss Sprachkonventionen respektieren

---

## Architektur-Entscheidungen (ADRs)

| ADR | Titel | Status | Datum |
|-----|-------|--------|-------|
| ADR-001 | Go CLI Architecture | Akzeptiert | 2026-03 |

Vollstaendige ADRs unter `docs/adrs/`.

---

## Zentrale Abhaengigkeiten

| Abhaengigkeit | Version | Zweck |
|--------------|---------|-------|
| Go | 1.24 | Sprache |
| gopkg.in/yaml.v3 | latest | governance.yaml Parsing |

---

## Offene Punkte / Backlog

Kein Backlog — alle Findings werden sofort behoben.

---

*governance-audit · Governance v3.5.0 · Compliance: vollstaendig*
