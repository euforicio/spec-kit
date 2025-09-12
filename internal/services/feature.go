package services

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/euforicio/spec-kit/internal/models"
)

// Compiled regexes for performance
var (
	langRegex          = regexp.MustCompile(`\*\*Language/Version\*\*: (.+)`)
	depRegex           = regexp.MustCompile(`\*\*Primary Dependencies\*\*: (.+)`)
	testRegex          = regexp.MustCompile(`\*\*Testing\*\*: (.+)`)
	storageRegex       = regexp.MustCompile(`\*\*Storage\*\*: (.+)`)
	projectTypeRegex   = regexp.MustCompile(`\*\*Project Type\*\*: (.+)`)
	activeTechtRegex   = regexp.MustCompile(`(## Active Technologies\n)(.*?)(\n\n)`)
	recentChangesRegex = regexp.MustCompile(`(## Recent Changes\n)(.*?)(\n\n)`)
	updateDateRegex    = regexp.MustCompile(`Last updated: \d{4}-\d{2}-\d{2}`)
)

// Language commands map for better maintainability
var languageCommands = map[string]string{
	"Python":     "cd src && pytest && ruff check .",
	"Rust":       "cargo test && cargo clippy",
	"JavaScript": "npm test && npm run lint",
	"TypeScript": "npm test && npm run lint",
	"Go":         "go test ./... && golangci-lint run",
}

type FeatureService struct {
	filesystem FilesystemServiceInterface
	git        GitServiceInterface
}

type FeatureServiceInterface interface {
	CreateFeature(description string) (*models.FeatureCreateResult, error)
	SetupPlan() (*models.FeaturePlanResult, error)
	CheckPrerequisites() (*models.FeatureCheckResult, error)
	ValidateAgentType(agentType string) error
	UpdateContext(agentType string) (*models.FeatureContextResult, error)
	GetPaths() (*models.FeaturePathsResult, error)
}

func NewFeatureService(filesystem FilesystemServiceInterface, git GitServiceInterface) *FeatureService {
	return &FeatureService{
		filesystem: filesystem,
		git:        git,
	}
}

func (f *FeatureService) CreateFeature(description string) (*models.FeatureCreateResult, error) {
	repoRoot, err := f.git.GetRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository root: %w", err)
	}

	specsDir := filepath.Join(repoRoot, "specs")

	// Create specs directory if it doesn't exist
	if err := f.filesystem.CreateDirectory(specsDir); err != nil {
		return nil, fmt.Errorf("failed to create specs directory: %w", err)
	}

	// Find highest numbered feature directory
	highest, err := f.getHighestFeatureNumber(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get highest feature number: %w", err)
	}

	// Generate next feature number
	next := highest + 1
	featureNum := fmt.Sprintf("%03d", next)

	// Create branch name from description
	branchName := f.createBranchName(description, featureNum)

	// Create and switch to new branch
	if err := f.git.CreateBranch(branchName); err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// Create feature directory
	featureDir := filepath.Join(specsDir, branchName)
	if err := f.filesystem.CreateDirectory(featureDir); err != nil {
		return nil, fmt.Errorf("failed to create feature directory: %w", err)
	}

	// Copy template if it exists
	templatePath := filepath.Join(repoRoot, "templates", "spec-template.md")
	specFile := filepath.Join(featureDir, "spec.md")

	if exists, _ := f.filesystem.FileExists(templatePath); exists {
		if err := f.filesystem.CopyFile(templatePath, specFile); err != nil {
			return nil, fmt.Errorf("failed to copy template: %w", err)
		}
	} else {
		// Create empty spec file
		if err := f.filesystem.WriteFile(specFile, "# Feature Specification\n\nTODO: Add feature specification\n"); err != nil {
			return nil, fmt.Errorf("failed to create spec file: %w", err)
		}
	}

	return &models.FeatureCreateResult{
		BranchName: branchName,
		SpecFile:   specFile,
		FeatureNum: featureNum,
	}, nil
}

func (f *FeatureService) SetupPlan() (*models.FeaturePlanResult, error) {
	repoRoot, err := f.git.GetRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository root: %w", err)
	}

	currentBranch, err := f.git.GetCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	if !f.isFeatureBranch(currentBranch) {
		return nil, fmt.Errorf("not on a feature branch. Current branch: %s. Feature branches should be named like: 001-feature-name", currentBranch)
	}

	featureDir := filepath.Join(repoRoot, "specs", currentBranch)

	// Create feature directory if it doesn't exist
	if err := f.filesystem.CreateDirectory(featureDir); err != nil {
		return nil, fmt.Errorf("failed to create feature directory: %w", err)
	}

	// Detect AI assistant from project structure
	aiAssistant, err := f.detectAIAssistant(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to detect AI assistant: %w", err)
	}

	// Copy plan template from agent-specific directory if it exists
	templatePath := filepath.Join(repoRoot, "."+aiAssistant, "templates", "plan-template.md")
	planFile := filepath.Join(featureDir, "plan.md")

	if exists, _ := f.filesystem.FileExists(templatePath); exists {
		// Read template content
		content, err := f.filesystem.ReadFile(templatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read plan template: %w", err)
		}

		// Replace $SPECS_DIR variable with actual specs directory path
		processedContent := strings.ReplaceAll(string(content), "$SPECS_DIR", featureDir)

		// Write processed content to plan file
		if err := f.filesystem.WriteFile(planFile, processedContent); err != nil {
			return nil, fmt.Errorf("failed to write plan file: %w", err)
		}
	}

	return &models.FeaturePlanResult{
		FeatureSpec: filepath.Join(featureDir, "spec.md"),
		ImplPlan:    planFile,
		SpecsDir:    featureDir,
		Branch:      currentBranch,
	}, nil
}

func (f *FeatureService) CheckPrerequisites() (*models.FeatureCheckResult, error) {
	repoRoot, err := f.git.GetRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository root: %w", err)
	}

	currentBranch, err := f.git.GetCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	if !f.isFeatureBranch(currentBranch) {
		return nil, fmt.Errorf("not on a feature branch. Current branch: %s. Feature branches should be named like: 001-feature-name", currentBranch)
	}

	featureDir := filepath.Join(repoRoot, "specs", currentBranch)

	// Check if feature directory exists
	if exists, _ := f.filesystem.DirectoryExists(featureDir); !exists {
		return nil, fmt.Errorf("feature directory not found: %s. Run 'specify feature plan' first to create the feature structure", featureDir)
	}

	// Check for implementation plan (required)
	planFile := filepath.Join(featureDir, "plan.md")
	if exists, _ := f.filesystem.FileExists(planFile); !exists {
		return nil, fmt.Errorf("plan.md not found in %s. Run 'specify feature plan' first to create the plan", featureDir)
	}

	// Check for optional design documents
	availableDocs := []string{}

	optionalFiles := map[string]string{
		"research.md":   "research.md",
		"data-model.md": "data-model.md",
		"quickstart.md": "quickstart.md",
	}

	for file, desc := range optionalFiles {
		if exists, _ := f.filesystem.FileExists(filepath.Join(featureDir, file)); exists {
			availableDocs = append(availableDocs, desc)
		}
	}

	// Check contracts directory
	contractsDir := filepath.Join(featureDir, "contracts")
	if exists, _ := f.filesystem.DirectoryExists(contractsDir); exists {
		if empty, _ := f.filesystem.IsDirectoryEmpty(contractsDir); !empty {
			availableDocs = append(availableDocs, "contracts/")
		}
	}

	return &models.FeatureCheckResult{
		FeatureDir:    featureDir,
		AvailableDocs: availableDocs,
	}, nil
}

func (f *FeatureService) ValidateAgentType(agentType string) error {
	if agentType == "" {
		return nil // Empty is allowed for "all agents"
	}

	if models.IsValidAgent(agentType) {
		return nil
	}

	return fmt.Errorf("invalid agent type '%s', must be one of: %s",
		agentType, strings.Join(models.ListAgents(), ", "))
}

func (f *FeatureService) UpdateContext(agentType string) (*models.FeatureContextResult, error) {
	repoRoot, err := f.git.GetRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository root: %w", err)
	}

	currentBranch, err := f.git.GetCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	if !f.isFeatureBranch(currentBranch) {
		return nil, fmt.Errorf("not on a feature branch. Current branch: %s", currentBranch)
	}

	featureDir := filepath.Join(repoRoot, "specs", currentBranch)
	planFile := filepath.Join(featureDir, "plan.md")

	// Check if plan exists
	if exists, _ := f.filesystem.FileExists(planFile); !exists {
		return nil, fmt.Errorf("no plan.md found at %s", planFile)
	}

	// Read plan file to extract technology info
	planContent, err := f.filesystem.ReadFile(planFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	techInfo := f.extractTechInfo(planContent)

	// Determine which files to update
	updates := []models.ContextUpdate{}
	summary := []string{}

	if agentType == "" {
		// Update all existing files
		agentFiles := map[string]string{
			"claude":  filepath.Join(repoRoot, "CLAUDE.md"),
			"gemini":  filepath.Join(repoRoot, "GEMINI.md"),
			"copilot": filepath.Join(repoRoot, ".github", "copilot-instructions.md"),
		}

		for agent, file := range agentFiles {
			if exists, _ := f.filesystem.FileExists(file); exists {
				if err := f.updateAgentFile(file, agent, techInfo, currentBranch); err != nil {
					return nil, fmt.Errorf("failed to update %s: %w", agent, err)
				}
				updates = append(updates, models.ContextUpdate{Agent: f.getAgentDisplayName(agent)})
			}
		}

		// If no files exist, create Claude file by default
		if len(updates) == 0 {
			claudeFile := agentFiles["claude"]
			if err := f.updateAgentFile(claudeFile, "claude", techInfo, currentBranch); err != nil {
				return nil, fmt.Errorf("failed to create Claude context file: %w", err)
			}
			updates = append(updates, models.ContextUpdate{Agent: "Claude Code"})
		}
	} else {
		// Update specific agent file
		var file string
		switch agentType {
		case "claude":
			file = filepath.Join(repoRoot, "CLAUDE.md")
		case "gemini":
			file = filepath.Join(repoRoot, "GEMINI.md")
		case "copilot":
			file = filepath.Join(repoRoot, ".github", "copilot-instructions.md")
		}

		if err := f.updateAgentFile(file, agentType, techInfo, currentBranch); err != nil {
			return nil, fmt.Errorf("failed to update %s: %w", agentType, err)
		}
		updates = append(updates, models.ContextUpdate{Agent: f.getAgentDisplayName(agentType)})
	}

	// Build summary
	if techInfo.Language != "" {
		summary = append(summary, fmt.Sprintf("Added language: %s", techInfo.Language))
	}
	if techInfo.Framework != "" {
		summary = append(summary, fmt.Sprintf("Added framework: %s", techInfo.Framework))
	}
	if techInfo.Database != "" {
		summary = append(summary, fmt.Sprintf("Added database: %s", techInfo.Database))
	}

	return &models.FeatureContextResult{
		Branch:  currentBranch,
		Updates: updates,
		Summary: summary,
	}, nil
}

func (f *FeatureService) GetPaths() (*models.FeaturePathsResult, error) {
	repoRoot, err := f.git.GetRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository root: %w", err)
	}

	currentBranch, err := f.git.GetCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	if !f.isFeatureBranch(currentBranch) {
		return nil, fmt.Errorf("not on a feature branch. Current branch: %s. Feature branches should be named like: 001-feature-name", currentBranch)
	}

	featureDir := filepath.Join(repoRoot, "specs", currentBranch)

	return &models.FeaturePathsResult{
		RepoRoot:    repoRoot,
		Branch:      currentBranch,
		FeatureDir:  featureDir,
		FeatureSpec: filepath.Join(featureDir, "spec.md"),
		ImplPlan:    filepath.Join(featureDir, "plan.md"),
		Tasks:       filepath.Join(featureDir, "tasks.md"),
	}, nil
}

// Helper methods

func (f *FeatureService) getHighestFeatureNumber(specsDir string) (int, error) {
	highest := 0

	exists, err := f.filesystem.DirectoryExists(specsDir)
	if err != nil || !exists {
		return highest, nil
	}

	entries, err := f.filesystem.ListDirectory(specsDir)
	if err != nil {
		return highest, err
	}

	for _, entry := range entries {
		if len(entry) >= 3 {
			numberStr := entry[:3]
			if number, err := strconv.Atoi(numberStr); err == nil {
				if number > highest {
					highest = number
				}
			}
		}
	}

	return highest, nil
}

func (f *FeatureService) createBranchName(description, featureNum string) string {
	// Convert to lowercase and replace non-alphanumeric with hyphens
	name := strings.ToLower(description)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, name)

	// Remove multiple consecutive hyphens
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")

	// Extract first 3 meaningful words
	words := strings.Split(name, "-")
	meaningfulWords := []string{}
	for _, word := range words {
		if word != "" && len(meaningfulWords) < 3 {
			meaningfulWords = append(meaningfulWords, word)
		}
	}

	return fmt.Sprintf("%s-%s", featureNum, strings.Join(meaningfulWords, "-"))
}

func (f *FeatureService) isFeatureBranch(branch string) bool {
	if len(branch) < 4 {
		return false
	}
	// Check format: 001-feature-name
	if branch[3] != '-' {
		return false
	}
	for i := 0; i < 3; i++ {
		if branch[i] < '0' || branch[i] > '9' {
			return false
		}
	}
	return true
}

type TechInfo struct {
	Language    string
	Framework   string
	Testing     string
	Database    string
	ProjectType string
}

func (f *FeatureService) extractTechInfo(content string) TechInfo {
	info := TechInfo{}

	// Extract technology information from plan content using pre-compiled regexes
	if match := langRegex.FindStringSubmatch(content); len(match) > 1 {
		if !strings.Contains(match[1], "NEEDS CLARIFICATION") {
			info.Language = strings.TrimSpace(match[1])
		}
	}

	if match := depRegex.FindStringSubmatch(content); len(match) > 1 {
		if !strings.Contains(match[1], "NEEDS CLARIFICATION") {
			info.Framework = strings.TrimSpace(match[1])
		}
	}

	if match := testRegex.FindStringSubmatch(content); len(match) > 1 {
		if !strings.Contains(match[1], "NEEDS CLARIFICATION") {
			info.Testing = strings.TrimSpace(match[1])
		}
	}

	if match := storageRegex.FindStringSubmatch(content); len(match) > 1 {
		if !strings.Contains(match[1], "N/A") && !strings.Contains(match[1], "NEEDS CLARIFICATION") {
			info.Database = strings.TrimSpace(match[1])
		}
	}

	if match := projectTypeRegex.FindStringSubmatch(content); len(match) > 1 {
		info.ProjectType = strings.TrimSpace(match[1])
	}

	return info
}

func (f *FeatureService) updateAgentFile(filePath, agentType string, techInfo TechInfo, currentBranch string) error {
	exists, err := f.filesystem.FileExists(filePath)
	if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	if !exists {
		// Create new file from template
		return f.createAgentFile(filePath, agentType, techInfo, currentBranch)
	}

	// Update existing file
	return f.updateExistingAgentFile(filePath, techInfo, currentBranch)
}

func (f *FeatureService) createAgentFile(filePath, agentType string, techInfo TechInfo, currentBranch string) error {
	repoRoot, _ := f.git.GetRepoRoot()
	templatePath := filepath.Join(repoRoot, "templates", "agent-file-template.md")

	var content string
	if exists, _ := f.filesystem.FileExists(templatePath); exists {
		templateContent, err := f.filesystem.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template: %w", err)
		}
		content = templateContent
	} else {
		// Basic template if not found
		content = f.getBasicAgentTemplate()
	}

	// Replace placeholders
	content = strings.ReplaceAll(content, "[PROJECT NAME]", filepath.Base(repoRoot))
	content = strings.ReplaceAll(content, "[DATE]", time.Now().Format("2006-01-02"))

	if techInfo.Language != "" && techInfo.Framework != "" {
		content = strings.ReplaceAll(content, "[EXTRACTED FROM ALL PLAN.MD FILES]",
			fmt.Sprintf("- %s + %s (%s)", techInfo.Language, techInfo.Framework, currentBranch))
	}

	// Add project structure based on type
	if strings.Contains(techInfo.ProjectType, "web") {
		content = strings.ReplaceAll(content, "[ACTUAL STRUCTURE FROM PLANS]", "backend/\nfrontend/\ntests/")
	} else {
		content = strings.ReplaceAll(content, "[ACTUAL STRUCTURE FROM PLANS]", "src/\ntests/")
	}

	// Add commands based on language
	commands := f.getCommandsForLanguage(techInfo.Language)
	content = strings.ReplaceAll(content, "[ONLY COMMANDS FOR ACTIVE TECHNOLOGIES]", commands)

	// Add code style
	if techInfo.Language != "" {
		content = strings.ReplaceAll(content, "[LANGUAGE-SPECIFIC, ONLY FOR LANGUAGES IN USE]",
			fmt.Sprintf("%s: Follow standard conventions", techInfo.Language))
	}

	// Add recent changes
	if techInfo.Language != "" && techInfo.Framework != "" {
		content = strings.ReplaceAll(content, "[LAST 3 FEATURES AND WHAT THEY ADDED]",
			fmt.Sprintf("- %s: Added %s + %s", currentBranch, techInfo.Language, techInfo.Framework))
	}

	// Create directory if needed
	if dir := filepath.Dir(filePath); dir != "." {
		if err := f.filesystem.CreateDirectory(dir); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return f.filesystem.WriteFile(filePath, content)
}

func (f *FeatureService) updateExistingAgentFile(filePath string, techInfo TechInfo, currentBranch string) error {
	content, err := f.filesystem.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read existing file: %w", err)
	}

	// Simple update - add new technology to active technologies section
	if techInfo.Language != "" {
		newTech := fmt.Sprintf("- %s + %s (%s)", techInfo.Language, techInfo.Framework, currentBranch)

		// Look for Active Technologies section
		if strings.Contains(content, "## Active Technologies") {
			// Add new tech if not already present
			if !strings.Contains(content, newTech) {
				content = activeTechtRegex.ReplaceAllStringFunc(content, func(match string) string {
					parts := activeTechtRegex.FindStringSubmatch(match)
					if len(parts) == 4 {
						return parts[1] + parts[2] + "\n" + newTech + parts[3]
					}
					return match
				})
			}
		}

		// Update recent changes
		if strings.Contains(content, "## Recent Changes") {
			newChange := fmt.Sprintf("- %s: Added %s + %s", currentBranch, techInfo.Language, techInfo.Framework)
			content = recentChangesRegex.ReplaceAllStringFunc(content, func(match string) string {
				parts := recentChangesRegex.FindStringSubmatch(match)
				if len(parts) == 4 {
					// Add new change at the top, keep only last 3
					changes := strings.Split(strings.TrimSpace(parts[2]), "\n")
					changes = append([]string{newChange}, changes...)
					if len(changes) > 3 {
						changes = changes[:3]
					}
					return parts[1] + strings.Join(changes, "\n") + parts[3]
				}
				return match
			})
		}
	}

	// Update date
	content = updateDateRegex.ReplaceAllString(content, fmt.Sprintf("Last updated: %s", time.Now().Format("2006-01-02")))

	return f.filesystem.WriteFile(filePath, content)
}

func (f *FeatureService) getAgentDisplayName(agentType string) string {
	switch agentType {
	case "claude":
		return "Claude Code"
	case "gemini":
		return "Gemini CLI"
	case "copilot":
		return "GitHub Copilot"
	default:
		return agentType
	}
}

func (f *FeatureService) getCommandsForLanguage(language string) string {
	// Check each supported language
	for lang, commands := range languageCommands {
		if strings.Contains(language, lang) {
			return commands
		}
	}
	return fmt.Sprintf("# Add commands for %s", language)
}

func (f *FeatureService) getBasicAgentTemplate() string {
	return `# [PROJECT NAME]

Last updated: [DATE]

## Active Technologies
[EXTRACTED FROM ALL PLAN.MD FILES]

## Project Structure
` + "```" + `
[ACTUAL STRUCTURE FROM PLANS]
` + "```" + `

## Commands
` + "```bash" + `
[ONLY COMMANDS FOR ACTIVE TECHNOLOGIES]
` + "```" + `

## Code Style
[LANGUAGE-SPECIFIC, ONLY FOR LANGUAGES IN USE]

## Recent Changes
[LAST 3 FEATURES AND WHAT THEY ADDED]
`
}

// detectAIAssistant detects which AI assistant is being used by checking for hidden directories
func (f *FeatureService) detectAIAssistant(repoRoot string) (string, error) {
	// Check for agent-specific hidden directories
	agents := []string{"claude", "codex", "gemini", "copilot"}

	for _, agent := range agents {
		agentDir := filepath.Join(repoRoot, "."+agent)
		if exists, _ := f.filesystem.DirectoryExists(agentDir); exists {
			return agent, nil
		}
	}

	return "", fmt.Errorf("no AI assistant directory found (looking for .claude, .codex, .gemini, or .copilot)")
}
