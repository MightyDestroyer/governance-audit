package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

type JSONReport struct {
	Timestamp         string         `json:"timestamp"`
	Version           string         `json:"version"`
	GovernanceVersion string         `json:"governance_version"`
	Repo              string         `json:"repo"`
	RepoName          string         `json:"repo_name"`
	Score             int            `json:"score"`
	MaxScore          int            `json:"max_score"`
	Rating            string         `json:"rating"`
	TotalChecks       int            `json:"total_checks"`
	PassedChecks      int            `json:"passed_checks"`
	FailedChecks      int            `json:"failed_checks"`
	Categories        []JSONCategory `json:"categories"`
}

type JSONCategory struct {
	Name   string      `json:"name"`
	Weight int         `json:"weight"`
	Checks []JSONCheck `json:"checks"`
}

type JSONCheck struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

func printTextReport(categories []CategoryResult, score int, repoPath string, verbose bool) {
	useColor := isTerminal()

	printHeader(useColor, repoPath)

	for _, cat := range categories {
		printCategoryHeader(useColor, cat)
		for _, ch := range cat.Checks {
			printCheckLine(useColor, ch, verbose)
		}
		fmt.Println()
	}

	printScoreSummary(useColor, score)
}

func printHeader(color bool, repoPath string) {
	fmt.Println()
	if color {
		fmt.Printf("%s%s Governance Audit Report %s\n", colorBold, colorCyan, colorReset)
	} else {
		fmt.Println("=== Governance Audit Report ===")
	}
	fmt.Printf("Repository: %s\n", repoPath)
	fmt.Println(strings.Repeat("─", 60))
}

func printCategoryHeader(color bool, cat CategoryResult) {
	passed := 0
	for _, ch := range cat.Checks {
		if ch.Passed {
			passed++
		}
	}

	if color {
		fmt.Printf("\n%s%s [weight: %d]%s  %d/%d passed\n",
			colorBold, cat.Name, cat.Weight, colorReset, passed, len(cat.Checks))
	} else {
		fmt.Printf("\n%s [weight: %d]  %d/%d passed\n",
			cat.Name, cat.Weight, passed, len(cat.Checks))
	}
}

func printCheckLine(color bool, ch CheckResult, verbose bool) {
	var status string
	if ch.Passed {
		if color {
			status = colorGreen + "  PASS" + colorReset
		} else {
			status = "  PASS"
		}
	} else {
		if color {
			status = colorRed + "  FAIL" + colorReset
		} else {
			status = "  FAIL"
		}
	}

	fmt.Printf("  %s  %s", status, ch.Name)

	if verbose || !ch.Passed {
		if color {
			fmt.Printf("  %s%s%s", colorDim, ch.Message, colorReset)
		} else {
			fmt.Printf("  (%s)", ch.Message)
		}
	}
	fmt.Println()
}

func printScoreSummary(color bool, score int) {
	fmt.Println(strings.Repeat("─", 60))

	bar := renderScoreBar(score)
	rating := scoreRating(score)

	if color {
		scoreColor := colorForScore(score)
		fmt.Printf("\n  Score: %s%s%d/100%s  %s\n", colorBold, scoreColor, score, colorReset, bar)
		fmt.Printf("  Rating: %s%s%s%s\n\n", colorBold, scoreColor, rating, colorReset)
	} else {
		fmt.Printf("\n  Score: %d/100  %s\n", score, bar)
		fmt.Printf("  Rating: %s\n\n", rating)
	}

	if score < 70 {
		msg := "Score below 70 — governance requirements not met."
		if color {
			fmt.Printf("  %s%s%s\n\n", colorRed, msg, colorReset)
		} else {
			fmt.Printf("  %s\n\n", msg)
		}
	}
}

func renderScoreBar(score int) string {
	const barWidth = 20
	filled := score * barWidth / 100
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "] " +
		fmt.Sprintf("%d/100", score)
}

func scoreRating(score int) string {
	switch {
	case score >= 90:
		return "Excellent"
	case score >= 70:
		return "Good"
	case score >= 50:
		return "Needs Work"
	default:
		return "Critical"
	}
}

func colorForScore(score int) string {
	switch {
	case score >= 90:
		return colorGreen
	case score >= 70:
		return colorGreen
	case score >= 50:
		return colorYellow
	default:
		return colorRed
	}
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
