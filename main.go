package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const version = "1.0.0"

func main() {
	repoPath := flag.String("repo", ".", "Path to the repository to audit")
	govPath := flag.String("governance", "", "Path to governance.yaml (auto-detected if empty)")
	format := flag.String("format", "text", "Output format: text or json")
	verbose := flag.Bool("verbose", false, "Show detailed output")
	saveMetrics := flag.Bool("save-metrics", false, "Save JSON report to metrics/ in the audited repo")
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

	totalChecks := countChecks(results)
	slog.Info("audit complete", "score", score, "total_checks", totalChecks)

	govVersion := ""
	if gov != nil {
		govVersion = gov.Version
	}

	switch *format {
	case "json":
		printJSONReport(results, score, absRepo, govVersion)
	default:
		printTextReport(results, score, absRepo, *verbose)
	}

	if *saveMetrics {
		saveMetricsJSON(results, score, absRepo, govVersion)
	}

	if score < 100 {
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

func buildJSONReport(categories []CategoryResult, score int, repoPath, govVersion string) JSONReport {
	passed := 0
	total := 0
	for _, cat := range categories {
		for _, ch := range cat.Checks {
			total++
			if ch.Passed {
				passed++
			}
		}
	}

	report := JSONReport{
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		Version:           version,
		GovernanceVersion: govVersion,
		Repo:              repoPath,
		RepoName:          filepath.Base(repoPath),
		Score:             score,
		MaxScore:          100,
		Rating:            scoreRating(score),
		TotalChecks:       total,
		PassedChecks:      passed,
		FailedChecks:      total - passed,
		Categories:        make([]JSONCategory, 0, len(categories)),
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

	return report
}

func printJSONReport(categories []CategoryResult, score int, repoPath, govVersion string) {
	report := buildJSONReport(categories, score, repoPath, govVersion)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		slog.Error("failed to encode JSON report", "error", err)
	}
}

func saveMetricsJSON(categories []CategoryResult, score int, repoPath, govVersion string) {
	metricsDir := filepath.Join(repoPath, "metrics")
	if err := os.MkdirAll(metricsDir, 0o755); err != nil {
		slog.Error("failed to create metrics directory", "path", metricsDir, "error", err)
		return
	}

	now := time.Now()
	filename := fmt.Sprintf("audit-%s.json", now.Format("2006-01-02T150405"))
	outPath := filepath.Join(metricsDir, filename)

	report := buildJSONReport(categories, score, repoPath, govVersion)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		slog.Error("failed to marshal metrics JSON", "error", err)
		return
	}

	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		slog.Error("failed to write metrics file", "path", outPath, "error", err)
		return
	}

	slog.Info("metrics saved", "path", outPath)
	fmt.Fprintf(os.Stderr, "Metrics saved to %s\n", outPath)
}
