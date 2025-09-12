package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/euforicio/spec-kit/internal/models"
)

// GitHubService handles interactions with the GitHub API
type GitHubService struct {
	client    *http.Client
	baseURL   string
	repoOwner string
	repoName  string
}

// GitHubRelease represents a GitHub release from the API
type GitHubRelease struct {
	ID          int64         `json:"id"`
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	CreatedAt   time.Time     `json:"created_at"`
	PublishedAt time.Time     `json:"published_at"`
	Assets      []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a GitHub release asset
type GitHubAsset struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	ContentType        string    `json:"content_type"`
	Size               int64     `json:"size"`
	BrowserDownloadURL string    `json:"browser_download_url"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// NewGitHubService creates a new GitHub service instance
func NewGitHubService() *GitHubService {
	return &GitHubService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://api.github.com",
		repoOwner: "euforicio",
		repoName:  "spec-kit",
	}
}

// NewGitHubServiceWithClient creates a new GitHub service with a custom HTTP client
func NewGitHubServiceWithClient(client *http.Client, repoOwner, repoName string) *GitHubService {
	return &GitHubService{
		client:    client,
		baseURL:   "https://api.github.com",
		repoOwner: repoOwner,
		repoName:  repoName,
	}
}

// GetLatestRelease fetches the latest release from the GitHub repository
func (g *GitHubService) GetLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", g.baseURL, g.repoOwner, g.repoName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header for better API interaction
	req.Header.Set("User-Agent", "specify-cli/1.0.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrTemplateDownloadFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"%w: GitHub API returned status %d: %s",
			models.ErrTemplateNotFound,
			resp.StatusCode,
			string(body),
		)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf(
			"%w: failed to parse GitHub API response: %v",
			models.ErrTemplateCorrupted,
			err,
		)
	}

	return &release, nil
}

// FindTemplateAsset finds the template asset in the release
func (g *GitHubService) FindTemplateAsset(
	release *GitHubRelease,
) (*GitHubAsset, error) {
	if release == nil {
		return nil, fmt.Errorf("%w: release cannot be nil", models.ErrTemplateNotFound)
	}

	// Find the first zip asset (unified template)
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".zip") {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf(
		"%w: no zip asset found in release %s",
		models.ErrTemplateNotFound,
		release.TagName,
	)
}

// GetTemplates retrieves the template
func (g *GitHubService) GetTemplates() (*models.Template, error) {
	// Get latest release
	release, err := g.GetLatestRelease()
	if err != nil {
		return nil, err
	}

	// Find the appropriate asset
	asset, err := g.FindTemplateAsset(release)
	if err != nil {
		return nil, err
	}

	// Create template model
	template, err := models.NewTemplate(
		release.TagName,
		asset.BrowserDownloadURL,
		asset.Name,
		asset.Size,
		release.PublishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to create template model for version %s: %v",
			models.ErrTemplateCorrupted,
			release.TagName,
			err,
		)
	}

	return template, nil
}

// DownloadAsset downloads a GitHub asset to the specified writer
func (g *GitHubService) DownloadAsset(asset *GitHubAsset, writer io.Writer) error {
	if asset == nil {
		return fmt.Errorf("%w: asset cannot be nil", models.ErrTemplateDownloadFailed)
	}

	req, err := http.NewRequest("GET", asset.BrowserDownloadURL, nil)
	if err != nil {
		return fmt.Errorf(
			"%w: failed to create download request: %v",
			models.ErrTemplateDownloadFailed,
			err,
		)
	}

	req.Header.Set("User-Agent", "specify-cli/1.0.0")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: failed to download asset: %v", models.ErrTemplateDownloadFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"%w: download failed with status %d",
			models.ErrTemplateDownloadFailed,
			resp.StatusCode,
		)
	}

	// Copy the response body to the writer
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf(
			"%w: failed to write downloaded data: %v",
			models.ErrTemplateDownloadFailed,
			err,
		)
	}

	return nil
}

// CheckConnectivity tests if GitHub API is accessible
func (g *GitHubService) CheckConnectivity() error {
	req, err := http.NewRequest("GET", g.baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create connectivity test request: %w", err)
	}

	req.Header.Set("User-Agent", "specify-cli/1.0.0")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf(
			"%w: failed to connect to GitHub API (check your internet connection): %v",
			models.ErrInternetNotAvailable,
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		// Client error - API is reachable but request is invalid (which is expected for root endpoint)
		return nil
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf(
			"%w: GitHub API is experiencing server issues (try again later or check https://status.github.com)",
			models.ErrInternetNotAvailable,
		)
	}

	return nil
}

// GetRateLimitInfo returns information about GitHub API rate limiting
func (g *GitHubService) GetRateLimitInfo() (limit, remaining int, resetTime time.Time, err error) {
	url := fmt.Sprintf("%s/rate_limit", g.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("failed to create rate limit request: %w", err)
	}

	req.Header.Set("User-Agent", "specify-cli/1.0.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("failed to get rate limit info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, time.Time{}, fmt.Errorf("rate limit API returned status %d", resp.StatusCode)
	}

	var rateLimitResp struct {
		Resources struct {
			Core struct {
				Limit     int   `json:"limit"`
				Remaining int   `json:"remaining"`
				Reset     int64 `json:"reset"`
			} `json:"core"`
		} `json:"resources"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rateLimitResp); err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("failed to parse rate limit response: %w", err)
	}

	return rateLimitResp.Resources.Core.Limit,
		rateLimitResp.Resources.Core.Remaining,
		time.Unix(rateLimitResp.Resources.Core.Reset, 0),
		nil
}
