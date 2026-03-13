package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

func yamlUnmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}

type GovernanceConfig struct {
	Version    string `yaml:"version"`
	Name       string `yaml:"name"`
	Principles []struct {
		ID   string `yaml:"id"`
		Name string `yaml:"name"`
	} `yaml:"principles"`
	Standards []struct {
		ID   string `yaml:"id"`
		Name string `yaml:"name"`
		File string `yaml:"file"`
	} `yaml:"standards"`
	Templates []struct {
		ID       string `yaml:"id"`
		File     string `yaml:"file"`
		Required bool   `yaml:"required"`
	} `yaml:"templates"`
}

type CheckResult struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
	Category string `json:"category"`
}

type CategoryResult struct {
	Name   string        `json:"name"`
	Weight int           `json:"weight"`
	Checks []CheckResult `json:"checks"`
}

func runAllChecks(repoPath string, gov *GovernanceConfig) []CategoryResult {
	return []CategoryResult{
		runStructureChecks(repoPath),
		runNamingChecks(repoPath),
		runSecurityChecks(repoPath),
		runDocumentationChecks(repoPath),
		runContractChecks(repoPath),
		runObservabilityChecks(repoPath),
	}
}

func calculateScore(categories []CategoryResult) int {
	totalWeightedScore := 0.0
	totalWeight := 0

	for _, cat := range categories {
		if len(cat.Checks) == 0 {
			continue
		}
		passed := 0
		for _, ch := range cat.Checks {
			if ch.Passed {
				passed++
			}
		}
		ratio := float64(passed) / float64(len(cat.Checks))
		totalWeightedScore += float64(cat.Weight) * ratio
		totalWeight += cat.Weight
	}

	if totalWeight == 0 {
		return 0
	}
	return int((totalWeightedScore / float64(totalWeight)) * 100)
}

// --- Structure Checks (weight 20) ---

func runStructureChecks(repoPath string) CategoryResult {
	cat := CategoryResult{Name: "Structure", Weight: 20}

	cat.Checks = append(cat.Checks, checkFileExists(repoPath, "README.md", "Structure"))
	cat.Checks = append(cat.Checks, checkDirExists(repoPath, "docs", "Structure"))
	cat.Checks = append(cat.Checks, checkFileExists(repoPath, "CLAUDE.md", "Structure"))
	cat.Checks = append(cat.Checks, checkDirExists(repoPath, ".github", "Structure"))
	cat.Checks = append(cat.Checks, checkFileExists(repoPath, ".gitignore", "Structure"))

	return cat
}

// --- Naming Checks (weight 15) ---

func runNamingChecks(repoPath string) CategoryResult {
	cat := CategoryResult{Name: "Naming", Weight: 15}

	cat.Checks = append(cat.Checks, checkKebabCase(repoPath))
	cat.Checks = append(cat.Checks, checkNoSpaces(repoPath))

	return cat
}

var namingExcludeDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true,
	"__pycache__": true, "dist": true, "build": true,
	".next": true, ".nuxt": true, "target": true,
}

// Uppercase filenames allowed by convention
var namingExcludeFiles = map[string]bool{
	"README.md": true, "CLAUDE.md": true, "MEMORY.md": true,
	"TASKS.md": true, "CHANGELOG.md": true, "LICENSE": true,
	"LICENSE.md": true, "Makefile": true, "Taskfile.yml": true,
	"Dockerfile": true, "Procfile": true, "Gemfile": true,
	"Rakefile": true, "Vagrantfile": true, "Brewfile": true,
	"SKILL.md": true, "AGENTS.md": true, "CONTRIBUTING.md": true,
	"SECURITY.md": true, "CODE_OF_CONDUCT.md": true,
}

var kebabCaseRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*(\.[a-z0-9]+)*$`)
var goSnakeCaseRegex = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*\.go$`)

var snakeCaseLanguages = map[string]bool{
	".go": true,
	".py": true,
	".rs": true,
	".rb": true,
}

func checkKebabCase(repoPath string) CheckResult {
	violations := 0
	var examples []string

	_ = filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		name := d.Name()

		if d.IsDir() {
			if namingExcludeDirs[name] || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
				return filepath.SkipDir
			}
			if !isKebabCase(name) {
				violations++
				if len(examples) < 3 {
					rel, _ := filepath.Rel(repoPath, path)
					examples = append(examples, rel)
				}
			}
			return nil
		}

		if strings.HasPrefix(name, ".") || namingExcludeFiles[name] {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(name))
		if snakeCaseLanguages[ext] {
			if !isLanguageConventionCompliant(name, ext) {
				violations++
				if len(examples) < 3 {
					rel, _ := filepath.Rel(repoPath, path)
					examples = append(examples, rel)
				}
			}
			return nil
		}

		if !isKebabCase(name) {
			violations++
			if len(examples) < 3 {
				rel, _ := filepath.Rel(repoPath, path)
				examples = append(examples, rel)
			}
		}
		return nil
	})

	if violations == 0 {
		return CheckResult{
			Name: "Kebab-case naming", Passed: true,
			Message: "All files/dirs follow naming conventions", Category: "Naming",
		}
	}
	msg := pluralize(violations, "violation")
	if len(examples) > 0 {
		msg += ": " + strings.Join(examples, ", ")
	}
	return CheckResult{
		Name: "Kebab-case naming", Passed: false,
		Message: msg, Category: "Naming",
	}
}

func isKebabCase(name string) bool {
	return kebabCaseRegex.MatchString(name)
}

func isLanguageConventionCompliant(name, ext string) bool {
	if isKebabCase(name) {
		return true
	}
	switch ext {
	case ".go":
		return goSnakeCaseRegex.MatchString(name)
	case ".py", ".rs", ".rb":
		base := strings.TrimSuffix(name, ext)
		snakeRegex := regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*$`)
		return snakeRegex.MatchString(base)
	}
	return false
}

func checkNoSpaces(repoPath string) CheckResult {
	violations := 0
	var examples []string

	_ = filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() && (namingExcludeDirs[name] || strings.HasPrefix(name, ".")) {
			return filepath.SkipDir
		}
		if strings.ContainsRune(name, ' ') {
			violations++
			if len(examples) < 3 {
				rel, _ := filepath.Rel(repoPath, path)
				examples = append(examples, rel)
			}
		}
		return nil
	})

	if violations == 0 {
		return CheckResult{
			Name: "No spaces in filenames", Passed: true,
			Message: "No spaces found", Category: "Naming",
		}
	}
	msg := pluralize(violations, "file with spaces")
	if len(examples) > 0 {
		msg += ": " + strings.Join(examples, ", ")
	}
	return CheckResult{
		Name: "No spaces in filenames", Passed: false,
		Message: msg, Category: "Naming",
	}
}

// --- Security Checks (weight 20) ---

func runSecurityChecks(repoPath string) CategoryResult {
	cat := CategoryResult{Name: "Security", Weight: 20}

	cat.Checks = append(cat.Checks, checkNoSecrets(repoPath))
	cat.Checks = append(cat.Checks, checkGitignorePatterns(repoPath))
	cat.Checks = append(cat.Checks, checkNoCommittedEnv(repoPath))

	return cat
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)API_KEY\s*=\s*["']?[A-Za-z0-9_\-]{16,}`),
	regexp.MustCompile(`(?i)API_SECRET\s*=\s*["']?[A-Za-z0-9_\-]{16,}`),
	regexp.MustCompile(`(?i)SECRET_KEY\s*=\s*["']?[A-Za-z0-9_\-]{16,}`),
	regexp.MustCompile(`(?i)PASSWORD\s*=\s*["']?[^\s"']{8,}`),
	regexp.MustCompile(`(?i)PRIVATE_KEY\s*=\s*["']?[A-Za-z0-9_\-/+=]{16,}`),
	regexp.MustCompile(`(?i)AWS_ACCESS_KEY_ID\s*=\s*["']?AK[A-Z0-9]{14,}`),
	regexp.MustCompile(`(?i)TOKEN\s*=\s*["']?[A-Za-z0-9_\-]{20,}`),
}

var secretScanExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true,
	".yaml": true, ".yml": true, ".json": true, ".toml": true,
	".env": true, ".cfg": true, ".ini": true, ".conf": true,
	".sh": true, ".bash": true, ".rb": true, ".java": true,
	".cs": true, ".php": true, ".rs": true,
}

func checkNoSecrets(repoPath string) CheckResult {
	findings := 0
	var examples []string

	_ = filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if namingExcludeDirs[name] || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(name))
		if !secretScanExtensions[ext] {
			return nil
		}

		if containsSecret(path) {
			findings++
			if len(examples) < 3 {
				rel, _ := filepath.Rel(repoPath, path)
				examples = append(examples, rel)
			}
		}
		return nil
	})

	if findings == 0 {
		return CheckResult{
			Name: "No secrets in code", Passed: true,
			Message: "No hardcoded secrets detected", Category: "Security",
		}
	}
	msg := pluralize(findings, "file with potential secrets")
	if len(examples) > 0 {
		msg += ": " + strings.Join(examples, ", ")
	}
	return CheckResult{
		Name: "No secrets in code", Passed: false,
		Message: msg, Category: "Security",
	}
}

func containsSecret(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if isExampleLine(line) {
			continue
		}
		for _, pat := range secretPatterns {
			if pat.MatchString(line) {
				slog.Debug("potential secret found", "file", path, "pattern", pat.String())
				return true
			}
		}
	}
	return false
}

func isExampleLine(line string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(line))
	placeholders := []string{
		"your-", "your_", "xxx", "changeme", "placeholder",
		"example", "<your", "${", "todo", "fixme",
	}
	for _, p := range placeholders {
		if strings.Contains(trimmed, p) {
			return true
		}
	}
	return false
}

func checkGitignorePatterns(repoPath string) CheckResult {
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		return CheckResult{
			Name: ".gitignore security patterns", Passed: false,
			Message: ".gitignore not found", Category: "Security",
		}
	}

	content := string(data)
	required := []string{".env", "*.key"}
	missing := []string{}

	for _, pattern := range required {
		if !strings.Contains(content, pattern) {
			missing = append(missing, pattern)
		}
	}

	if len(missing) == 0 {
		return CheckResult{
			Name: ".gitignore security patterns", Passed: true,
			Message: "Required patterns present (.env, *.key)", Category: "Security",
		}
	}
	return CheckResult{
		Name: ".gitignore security patterns", Passed: false,
		Message: "Missing patterns: " + strings.Join(missing, ", "), Category: "Security",
	}
}

func checkNoCommittedEnv(repoPath string) CheckResult {
	envPath := filepath.Join(repoPath, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return CheckResult{
			Name: "No .env committed", Passed: true,
			Message: "No .env file found (good)", Category: "Security",
		}
	}

	gitignorePath := filepath.Join(repoPath, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		return CheckResult{
			Name: "No .env committed", Passed: false,
			Message: ".env exists but no .gitignore found", Category: "Security",
		}
	}

	if strings.Contains(string(data), ".env") {
		return CheckResult{
			Name: "No .env committed", Passed: true,
			Message: ".env exists but is in .gitignore", Category: "Security",
		}
	}

	return CheckResult{
		Name: "No .env committed", Passed: false,
		Message: ".env exists and is NOT in .gitignore", Category: "Security",
	}
}

// --- Documentation Checks (weight 15) ---

func runDocumentationChecks(repoPath string) CategoryResult {
	cat := CategoryResult{Name: "Documentation", Weight: 15}

	cat.Checks = append(cat.Checks, checkReadmeContent(repoPath))
	cat.Checks = append(cat.Checks, checkProjectBible(repoPath))
	cat.Checks = append(cat.Checks, checkADRDirectory(repoPath))

	return cat
}

func checkReadmeContent(repoPath string) CheckResult {
	readmePath := filepath.Join(repoPath, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return CheckResult{
			Name: "README.md has content", Passed: false,
			Message: "README.md not found", Category: "Documentation",
		}
	}

	lines := strings.Split(string(data), "\n")
	nonEmpty := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
	}

	if nonEmpty > 10 {
		return CheckResult{
			Name: "README.md has content", Passed: true,
			Message: pluralize(nonEmpty, "non-empty line"), Category: "Documentation",
		}
	}
	return CheckResult{
		Name: "README.md has content", Passed: false,
		Message: pluralize(nonEmpty, "non-empty line") + " (minimum 10)", Category: "Documentation",
	}
}

func checkProjectBible(repoPath string) CheckResult {
	candidates := []string{
		filepath.Join(repoPath, "docs", "project-bible.md"),
		filepath.Join(repoPath, "docs", "PROJECT-BIBLE.md"),
		filepath.Join(repoPath, "PROJECT-BIBLE.md"),
		filepath.Join(repoPath, "docs", "sprint-log.md"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			rel, _ := filepath.Rel(repoPath, c)
			return CheckResult{
				Name: "Project bible exists", Passed: true,
				Message: "Found: " + rel, Category: "Documentation",
			}
		}
	}
	return CheckResult{
		Name: "Project bible exists", Passed: false,
		Message: "No project-bible.md or equivalent found in docs/", Category: "Documentation",
	}
}

func checkADRDirectory(repoPath string) CheckResult {
	candidates := []string{
		filepath.Join(repoPath, "docs", "adrs"),
		filepath.Join(repoPath, "docs", "adr"),
		filepath.Join(repoPath, "adrs"),
		filepath.Join(repoPath, "adr"),
	}

	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && info.IsDir() {
			rel, _ := filepath.Rel(repoPath, c)
			return CheckResult{
				Name: "ADR directory exists", Passed: true,
				Message: "Found: " + rel, Category: "Documentation",
			}
		}
	}
	return CheckResult{
		Name: "ADR directory exists", Passed: false,
		Message: "No ADR directory found (expected docs/adrs/)", Category: "Documentation",
	}
}

// --- Contract Checks (weight 15) ---

func runContractChecks(repoPath string) CategoryResult {
	cat := CategoryResult{Name: "Contracts", Weight: 15}

	cat.Checks = append(cat.Checks, checkContractFiles(repoPath))

	return cat
}

func checkContractFiles(repoPath string) CheckResult {
	patterns := []struct {
		dir  string
		glob string
		desc string
	}{
		{"contracts", "*.yaml", "contract YAML"},
		{"contracts", "*.yml", "contract YML"},
		{"api", "*.yaml", "API spec YAML"},
		{"api", "*.yml", "API spec YML"},
		{"api", "*.json", "API spec JSON"},
		{"schemas", "*.json", "JSON schema"},
		{"schemas", "*.yaml", "YAML schema"},
		{"models", "*.sysml", "SysML model"},
	}

	for _, p := range patterns {
		dir := filepath.Join(repoPath, p.dir)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		matches, err := filepath.Glob(filepath.Join(dir, p.glob))
		if err == nil && len(matches) > 0 {
			return CheckResult{
				Name: "Contract/schema files exist", Passed: true,
				Message: pluralize(len(matches), p.desc) + " in " + p.dir + "/",
				Category: "Contracts",
			}
		}
	}

	rootContracts := []string{
		"governance.yaml", "openapi.yaml", "openapi.yml", "openapi.json",
		"asyncapi.yaml", "asyncapi.yml", "swagger.yaml", "swagger.json",
	}
	for _, rc := range rootContracts {
		path := filepath.Join(repoPath, rc)
		if _, err := os.Stat(path); err == nil {
			return CheckResult{
				Name: "Contract/schema files exist", Passed: true,
				Message: "Root-level contract found: " + rc,
				Category: "Contracts",
			}
		}
	}

	return CheckResult{
		Name: "Contract/schema files exist", Passed: false,
		Message: "No contracts, API specs, schemas, or SysML models found",
		Category: "Contracts",
	}
}

// --- Observability Checks (weight 15) ---

func runObservabilityChecks(repoPath string) CategoryResult {
	cat := CategoryResult{Name: "Observability", Weight: 15}

	cat.Checks = append(cat.Checks, checkHealthEndpoint(repoPath))
	cat.Checks = append(cat.Checks, checkLoggingSetup(repoPath))
	cat.Checks = append(cat.Checks, checkObservabilityConfig(repoPath))

	return cat
}

func checkHealthEndpoint(repoPath string) CheckResult {
	if !hasServiceEntrypoint(repoPath) {
		return CheckResult{
			Name: "Health endpoint reference", Passed: true,
			Message: "N/A — no service entrypoint detected (library/standards/CLI repo)",
			Category: "Observability",
		}
	}

	patterns := []string{`/health`, `/ready`, `/healthz`, `/readyz`, `/livez`}
	found := searchInSourceFiles(repoPath, patterns)

	if found {
		return CheckResult{
			Name: "Health endpoint reference", Passed: true,
			Message: "Health/readiness endpoint pattern found in source", Category: "Observability",
		}
	}
	return CheckResult{
		Name: "Health endpoint reference", Passed: false,
		Message: "No /health or /ready endpoint pattern found", Category: "Observability",
	}
}

func hasServiceEntrypoint(repoPath string) bool {
	serviceIndicators := []string{
		"http.ListenAndServe", "http.Serve", "net.Listen",
		"gin.Default", "echo.New", "fiber.New", "mux.NewRouter",
		"app.listen", "express()", "fastify(", "createServer",
		"uvicorn.run", "flask.run", "app.run",
	}

	cmdDir := filepath.Join(repoPath, "cmd")
	if info, err := os.Stat(cmdDir); err == nil && info.IsDir() {
		if searchInSourceFiles(repoPath, serviceIndicators) {
			return true
		}
	}

	servicesDir := filepath.Join(repoPath, "services")
	if info, err := os.Stat(servicesDir); err == nil && info.IsDir() {
		return true
	}

	return searchInSourceFiles(repoPath, serviceIndicators)
}

func checkLoggingSetup(repoPath string) CheckResult {
	if !hasSourceFiles(repoPath) {
		return CheckResult{
			Name: "Structured logging setup", Passed: true,
			Message: "N/A — no source files detected (template/docs repo)",
			Category: "Observability",
		}
	}

	patterns := []string{
		`slog`, `structlog`, `createLogger`, `pino`, `winston`, `log/slog`,
		`logging.getLogger`, `logging.basicConfig`, `JSONFormatter`, `logging.Logger`,
		`console.error`, `console.warn`,
	}
	found := searchInSourceFiles(repoPath, patterns)

	if found {
		return CheckResult{
			Name: "Structured logging setup", Passed: true,
			Message: "Logging framework reference found", Category: "Observability",
		}
	}
	return CheckResult{
		Name: "Structured logging setup", Passed: false,
		Message: "No structured logging framework detected (slog, structlog, pino, winston, logging)",
		Category: "Observability",
	}
}

func hasSourceFiles(repoPath string) bool {
	found := false
	_ = filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || found {
			return err
		}
		name := d.Name()
		if d.IsDir() && (namingExcludeDirs[name] || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")) {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if sourceExtensions[ext] {
			found = true
		}
		return nil
	})
	return found
}

func checkObservabilityConfig(repoPath string) CheckResult {
	configFiles := []string{
		"docker-compose.yml", "docker-compose.yaml",
		"docker-compose.observability.yml", "docker-compose.monitoring.yml",
	}

	for _, cf := range configFiles {
		path := filepath.Join(repoPath, cf)
		if containsAnyPattern(path, []string{"prometheus", "grafana", "loki", "jaeger", "tempo", "otel", "datadog", "newrelic"}) {
			return CheckResult{
				Name: "Observability tooling config", Passed: true,
				Message: "Observability tools referenced in " + cf, Category: "Observability",
			}
		}
	}

	k8sDirs := []string{"k8s", "kubernetes", "deploy", "infrastructure", "infra"}
	for _, dir := range k8sDirs {
		dirPath := filepath.Join(repoPath, dir)
		if _, err := os.Stat(dirPath); err == nil {
			if searchDirForPatterns(dirPath, []string{"prometheus", "grafana", "loki", "jaeger", "tempo", "otel"}) {
				return CheckResult{
					Name: "Observability tooling config", Passed: true,
					Message: "Observability tools referenced in " + dir + "/", Category: "Observability",
				}
			}
		}
	}

	otelSourcePatterns := []string{
		"opentelemetry", "otel.Setup", "otel.Shutdown",
		"prometheus.NewRegistry", "prometheus.MustRegister", "promhttp",
		"client_golang/prometheus", "go.opentelemetry.io/otel",
	}
	if searchInSourceFiles(repoPath, otelSourcePatterns) {
		return CheckResult{
			Name: "Observability tooling config", Passed: true,
			Message: "Observability SDK integrated in source code (OTel/Prometheus)", Category: "Observability",
		}
	}

	if !hasServiceEntrypoint(repoPath) {
		return CheckResult{
			Name: "Observability tooling config", Passed: true,
			Message: "N/A — no service entrypoint detected (library/CLI/template repo)",
			Category: "Observability",
		}
	}

	return CheckResult{
		Name: "Observability tooling config", Passed: false,
		Message: "No observability tooling found (docker-compose, k8s, or SDK integration)",
		Category: "Observability",
	}
}

// --- Helpers ---

var sourceExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true,
	".tsx": true, ".jsx": true, ".rb": true, ".java": true,
	".cs": true, ".rs": true, ".php": true, ".kt": true,
}

func searchInSourceFiles(repoPath string, patterns []string) bool {
	found := false
	_ = filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || found {
			return err
		}
		name := d.Name()
		if d.IsDir() && (namingExcludeDirs[name] || strings.HasPrefix(name, ".")) {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(name))
		if !sourceExtensions[ext] {
			return nil
		}

		if containsAnyPattern(path, patterns) {
			found = true
		}
		return nil
	})
	return found
}

func containsAnyPattern(filePath string, patterns []string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())
		for _, p := range patterns {
			if strings.Contains(line, strings.ToLower(p)) {
				return true
			}
		}
	}
	return false
}

func searchDirForPatterns(dirPath string, patterns []string) bool {
	found := false
	_ = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || found {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".toml" {
			if containsAnyPattern(path, patterns) {
				found = true
			}
		}
		return nil
	})
	return found
}

func checkFileExists(repoPath, filename, category string) CheckResult {
	path := filepath.Join(repoPath, filename)
	if _, err := os.Stat(path); err == nil {
		return CheckResult{
			Name: filename + " exists", Passed: true,
			Message: filename + " found", Category: category,
		}
	}
	return CheckResult{
		Name: filename + " exists", Passed: false,
		Message: filename + " not found", Category: category,
	}
}

func checkDirExists(repoPath, dirname, category string) CheckResult {
	path := filepath.Join(repoPath, dirname)
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return CheckResult{
			Name: dirname + "/ exists", Passed: true,
			Message: dirname + "/ found", Category: category,
		}
	}
	return CheckResult{
		Name: dirname + "/ exists", Passed: false,
		Message: dirname + "/ not found", Category: category,
	}
}

func pluralize(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	s := noun
	if !strings.HasSuffix(s, "s") && !endsWithSpecial(s) {
		s += "s"
	}
	return fmt.Sprintf("%d %s", n, s)
}

func endsWithSpecial(s string) bool {
	if len(s) == 0 {
		return false
	}
	last := rune(s[len(s)-1])
	return unicode.IsDigit(last)
}
