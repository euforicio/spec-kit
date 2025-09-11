package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/euforicio/spec-kit/internal/services"
)

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage template cache and synchronization",
	Long: `Manage the local template cache for Specify projects.

This command allows you to sync templates from the local templates directory
to the cache, validate cache integrity, and manage template versions.`,
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize templates to cache",
	Long: `Synchronize templates from the local template directory to the cache.

This command will:
1. Copy templates from the local /templates and /memory directories
2. Normalize naming to consistent .template.md pattern
3. Generate a manifest with current spec-kit version and file hashes
4. Create a unified cache structure at ~/.spec-kit/templates/

The cache structure will be:
  ~/.spec-kit/templates/
  ├── .manifest.json
  ├── memory/
  │   ├── constitution.md
  │   └── constitution_update_checklist.md
  ├── commands/
  │   ├── plan.md
  │   ├── tasks.md
  │   └── specify.md
  └── templates/
      ├── agent_claude.template.md
      ├── agent_codex.template.md
      ├── agent_gemini.template.md
      ├── tasks.template.md
      ├── plan.template.md
      └── spec.template.md`,
	RunE: runTemplatesSync,
}

var (
	forceSync   bool
	verboseSync bool
)

func init() {
	// Add sync command to templates command
	templatesCmd.AddCommand(syncCmd)

	// Add flags to sync command
	syncCmd.Flags().BoolVar(&forceSync, "force", false, "Force sync even if cache is up to date")
	syncCmd.Flags().BoolVar(&verboseSync, "verbose", false, "Show detailed output during sync")
}

func runTemplatesSync(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 Synchronizing templates from GitHub to cache...")

	// Initialize services
	filesystem := services.NewFilesystemService()
	github := services.NewGitHubService()
	template := services.NewTemplateService(github, filesystem)

	// Create temporary directory for downloads
	tempDir, err := filesystem.CreateTempDirectory("specify-sync-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer filesystem.RemoveDirectory(tempDir)

	if verboseSync {
		fmt.Printf("Temporary download directory: %s\n", tempDir)
	}

	// Download cache template ZIP from GitHub
	if err := downloadCacheTemplate(github, filesystem, tempDir, verboseSync); err != nil {
		return fmt.Errorf("failed to download cache template: %w", err)
	}

	// Get cache root directory
	cacheRoot, err := template.ResolveRoot()
	if err != nil {
		return fmt.Errorf("failed to resolve cache root: %w", err)
	}

	if verboseSync {
		fmt.Printf("Cache root: %s\n", cacheRoot)
	}

	// Check if cache already exists and is up to date
	if !forceSync {
		manifest, err := template.ReadManifest()
		if err == nil {
			currentVersion := template.GetSpecKitVersion()
			if manifest.IsVersionMatch(currentVersion) {
				fmt.Println("✅ Cache is already up to date")
				return nil
			}
			fmt.Printf("📦 Version mismatch: cache has %s, current is %s\n", manifest.SpecKitVersion, currentVersion)
		} else if verboseSync {
			fmt.Printf("📝 No existing manifest found: %v\n", err)
		}
	}

	// Extract and sync files from downloaded ZIP (includes pre-built manifest)
	sourceDir := filepath.Join(tempDir, "extracted")
	if err := syncFromDownloadedFiles(filesystem, template, tempDir, sourceDir, cacheRoot, verboseSync); err != nil {
		return fmt.Errorf("failed to sync from downloaded files: %w", err)
	}

	// Read the manifest that was copied from the downloaded ZIP
	manifest, err := template.ReadManifest()
	if err != nil {
		return fmt.Errorf("failed to read synced manifest: %w", err)
	}

	fmt.Printf("✅ Templates synchronized successfully!\n")
	fmt.Printf("📊 Cached %d files to %s\n", manifest.GetTemplateCount(), cacheRoot)

	return nil
}

// downloadCacheTemplate downloads the cache template ZIP from GitHub
func downloadCacheTemplate(github *services.GitHubService, filesystem *services.FilesystemService, tempDir string, verbose bool) error {
	if verbose {
		fmt.Println("🌐 Downloading cache template from GitHub...")
	}

	// Get latest release
	release, err := github.GetLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	// Look for cache template asset
	var cacheAsset *services.GitHubAsset
	for _, asset := range release.Assets {
		if asset.Name == "spec-kit-cache-template.zip" {
			cacheAsset = &asset
			break
		}
	}

	if cacheAsset == nil {
		return fmt.Errorf("cache template asset not found in release %s", release.TagName)
	}

	if verbose {
		fmt.Printf("📦 Found cache template: %s (%s)\n", cacheAsset.Name, formatBytes(cacheAsset.Size))
	}

	// Download the asset
	zipPath := filepath.Join(tempDir, cacheAsset.Name)
	if err := downloadAssetToFile(github, cacheAsset, zipPath); err != nil {
		return fmt.Errorf("failed to download cache template: %w", err)
	}

	if verbose {
		fmt.Printf("✅ Downloaded cache template to %s\n", zipPath)
	}

	return nil
}

// syncFromDownloadedFiles extracts and syncs files from the downloaded ZIP
func syncFromDownloadedFiles(filesystem *services.FilesystemService, template *services.TemplateService, tempDir, sourceDir, cacheRoot string, verbose bool) error {
	// Find the ZIP file
	zipPath := filepath.Join(tempDir, "spec-kit-cache-template.zip")

	// Extract ZIP to sourceDir
	if err := filesystem.ExtractZIPWithFlatten(zipPath, sourceDir); err != nil {
		return fmt.Errorf("failed to extract cache template: %w", err)
	}

	if verbose {
		fmt.Printf("📁 Extracted cache template to %s\n", sourceDir)
	}

	// Copy the entire extracted structure to cache (including pre-built manifest)
	if err := filesystem.MergeDirectories(sourceDir, cacheRoot); err != nil {
		return fmt.Errorf("failed to copy extracted files to cache: %w", err)
	}

	if verbose {
		fmt.Printf("📄 Copied extracted template structure to cache\n")
	}

	return nil
}

// formatBytes formats byte size for human readability
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// downloadAssetToFile downloads a GitHub asset to a file
func downloadAssetToFile(github *services.GitHubService, asset *services.GitHubAsset, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := github.DownloadAsset(asset, file); err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}

	return nil
}
