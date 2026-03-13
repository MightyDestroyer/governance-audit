# ADR-001: Go CLI Architecture

**Status:** Akzeptiert
**Datum:** 2026-03-12
**Kontext:** Architektur-Entscheidung fuer das governance-audit Tool

---

## Kontext

Es wird ein CLI-Tool benoetigt das Repositories gegen die MightyDestroyer Governance-Standards prueft und einen Compliance-Score berechnet.

## Entscheidung

- **Sprache:** Go — konsistent mit dem Oekosystem, single binary, cross-platform
- **Architektur:** Single-Package CLI mit Category-basiertem Checker-Pattern
- **Output:** Text (ANSI) und JSON fuer maschinelle Weiterverarbeitung
- **Config:** governance.yaml als Input-Schema (SSoT)
- **Scoring:** Gewichtete Kategorien (Structure, Naming, Security, Documentation, Contracts, Observability)

## Alternativen

| Alternative | Grund fuer Ablehnung |
|------------|---------------------|
| Python Script | Dependency-Management komplexer, keine Single Binary |
| Shell Script | Zu fragil fuer komplexe Checks, schlechte Testbarkeit |
| Node.js | Unnoetige Runtime-Abhaengigkeit |

## Konsequenzen

- Einfaches Deployment (single binary, kein Runtime noetig)
- Cross-Platform (Windows, Linux, macOS)
- Erweiterbar durch neue Check-Funktionen im Category-Pattern
- JSON-Output ermoeglicht Integration in CI/CD und Dashboards

---

*governance-audit · ADR-001*
