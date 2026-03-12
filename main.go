package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const version = "1.0.0"

func main() {
	repoPath := flag.String("repo", ".", "Path to the repository to audit")
	govPath := flag.String("governance", "", "Path to governance.yaml (auto-detected if empty)")
	format := flag.String("format", "text", "Output format: text or json")
	verbose := flag.Bool("verbose", false, "Show detailed output")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel(*verbose),
	}))
	slog.SetDefault(logger)

	if *showVersion {
		fmt.Printf("governance-audit v%s\n", version)
		os.Exit(0)
	}

	absRepo, err := filepath.Abs(*repoPath)
	if err != nil {
		slog.Error("failed to resolve repo path", "path", *repoPath, "error", err)
		os.Exit(2)
	}

	if _, err := os.Stat(absRepo); os.IsNotExist(err) {
		slog.Error("repo path does not exist", "path", absRepo)
		os.Exit(2)
	}

	governancePath := resolveGovernancePath(*govPath, absRepo)
	slog.Info("starting audit", "repo", absRepo, "governance", governancePath, "version", version)

	var gov *GovernanceConfig
	if governancePath != "" {
		gov, err = loadGovernanceConfig(governancePath)
		if err != nil {
			slog.Warn("failed to load governance.yaml, running with defaults", "error", err)
		}
	}

	results := runAllChecks(absRepo, gov)
	score := calculateScore(results)

	slog.Info("audit complete", "score", score, "total_checks", countChecks(results))

	switch *format {
	case "json":
		printJSONReport(results, score, absRepo)
	default:
		printTextReport(results, score, absRepo, *verbose)
	}

	if score < 70 {
		os.Exit(1)
	}
}

func logLevel(verbose bool) slog.Level {
	if verbose {
		return slog.LevelDebug
	}
	return slog.LevelWarn
}

func resolveGovernancePath(explicit, repoPath string) string {
	if explicit != "" {
		return explicit
	}

	candidates := []string{
		filepath.Join(repoPath, "governance.yaml"),
		filepath.Join(repoPath, "..", "Governance", "governance.yaml"),
		filepath.Join(repoPath, "..", "governance", "governance.yaml"),
	}

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, "Governance", "governance.yaml"),
			filepath.Join(home, "dev", "Governance", "governance.yaml"),
		)
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func loadGovernanceConfig(path string) (*GovernanceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading governance.yaml: %w", err)
	}

	var gov GovernanceConfig
	if err := parseYAML(data, &gov); err != nil {
		return nil, fmt.Errorf("parsing governance.yaml: %w", err)
	}
	return &gov, nil
}

func parseYAML(data []byte, v any) error {
	return yamlUnmarshal(data, v)
}

func countChecks(categories []CategoryResult) int {
	total := 0
	for _, c := range categories {
		total += len(c.Checks)
	}
	return total
}

func printJSONReport(categories []CategoryResult, score int, repoPath string) {
	report := JSONReport{
		Version: version,
		Repo:    repoPath,
		Score:   score,
		Rating:  scoreRating(score),
		Categories: make([]JSONCategory, 0, len(categories)),
	}

	for _, cat := range categories {
		jc := JSONCategory{
			Name:   cat.Name,
			Weight: cat.Weight,
			Checks: make([]JSONCheck, 0, len(cat.Checks)),
		}
		for _, ch := range cat.Checks {
			jc.Checks = append(jc.Checks, JSONCheck{
				Name:    ch.Name,
				Passed:  ch.Passed,
				Message: ch.Message,
			})
		}
		report.Categories = append(report.Categories, jc)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		slog.Error("failed to encode JSON report", "error", err)
	}
}
