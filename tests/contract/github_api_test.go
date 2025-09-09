package contract

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GitHubRelease represents the structure we expect from GitHub API
type GitHubRelease struct {
	ID          int64     `json:"id"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		ID                 int64  `json:"id"`
		Name               string `json:"name"`
		Label              string `json:"label"`
		ContentType        string `json:"content_type"`
		Size               int64  `json:"size"`
		BrowserDownloadURL string `json:"browser_download_url"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
	} `json:"assets"`
}

func TestGitHubAPIContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GitHub API tests in short mode")
	}

	// Test the actual GitHub API that our implementation will use
	const (
		repoOwner = "github"
		repoName  = "spec-kit"
		apiURL    = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
	)

	t.Run("GitHub API releases endpoint responds correctly", func(t *testing.T) {
		client := &http.Client{Timeout: 30 * time.Second}
		
		resp, err := client.Get(apiURL)
		require.NoError(t, err, "should be able to contact GitHub API")
		defer resp.Body.Close()

		// API should respond with 200 OK
		assert.Equal(t, http.StatusOK, resp.StatusCode, "GitHub API should respond with 200 OK")

		// Content-Type should be JSON
		contentType := resp.Header.Get("Content-Type")
		assert.Contains(t, contentType, "application/json", "response should be JSON")

		// Should be able to parse the response
		var release GitHubRelease
		err = json.NewDecoder(resp.Body).Decode(&release)
		require.NoError(t, err, "response should be valid JSON matching our expected structure")

		// Validate required fields
		assert.NotEmpty(t, release.TagName, "release should have a tag name")
		assert.NotZero(t, release.ID, "release should have an ID")
		assert.NotEmpty(t, release.Assets, "release should have assets")

		// Validate that we have the template assets we need
		aiAssistants := []string{"claude", "gemini", "copilot"}
		for _, ai := range aiAssistants {
			found := false
			expectedPattern := "spec-kit-template-" + ai

			for _, asset := range release.Assets {
				if containsString(asset.Name, expectedPattern) && 
				   endsWithString(asset.Name, ".zip") {
					found = true

					// Validate asset properties
					assert.NotEmpty(t, asset.BrowserDownloadURL, "asset should have download URL")
					assert.Greater(t, asset.Size, int64(0), "asset should have positive size")
					assert.Contains(t, asset.ContentType, "zip", "asset should be a ZIP file")
					break
				}
			}

			assert.True(t, found, "should have template for AI assistant: %s", ai)
		}
	})

	t.Run("GitHub API rate limiting headers present", func(t *testing.T) {
		client := &http.Client{Timeout: 30 * time.Second}
		
		resp, err := client.Get(apiURL)
		require.NoError(t, err, "should be able to contact GitHub API")
		defer resp.Body.Close()

		// GitHub API should include rate limiting headers
		assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Limit"), "should include rate limit")
		assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Remaining"), "should include remaining requests")
		assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Reset"), "should include reset time")
	})

	t.Run("asset download URLs are accessible", func(t *testing.T) {
		client := &http.Client{Timeout: 30 * time.Second}
		
		// Get release info first
		resp, err := client.Get(apiURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		var release GitHubRelease
		err = json.NewDecoder(resp.Body).Decode(&release)
		require.NoError(t, err)

		// Test that at least one asset download URL is accessible
		if len(release.Assets) > 0 {
			asset := release.Assets[0]
			
			// Make HEAD request to check if asset is downloadable
			headResp, err := client.Head(asset.BrowserDownloadURL)
			require.NoError(t, err, "asset download URL should be accessible")
			defer headResp.Body.Close()

			assert.Equal(t, http.StatusOK, headResp.StatusCode, "asset should be downloadable")
			
			// Should have content-length header
			contentLength := headResp.Header.Get("Content-Length")
			assert.NotEmpty(t, contentLength, "should provide content length")
		}
	})
}

func TestGitHubAPIErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GitHub API error tests in short mode")
	}

	t.Run("non-existent repository returns 404", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Try to access a non-existent repository
		nonExistentURL := "https://api.github.com/repos/non-existent-user/non-existent-repo/releases/latest"
		resp, err := client.Get(nonExistentURL)
		require.NoError(t, err, "request should complete even for non-existent repo")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "should return 404 for non-existent repository")
	})

	t.Run("malformed API URL returns error", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Try to access a malformed URL
		malformedURL := "https://api.github.com/repos/github/spec-kit/releases/invalid-endpoint"
		resp, err := client.Get(malformedURL)
		require.NoError(t, err, "request should complete even for malformed URL")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "should return 404 for invalid endpoint")
	})
}

func TestGitHubAPIResponseStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GitHub API structure tests in short mode")
	}

	t.Run("response contains all required fields", func(t *testing.T) {
		client := &http.Client{Timeout: 30 * time.Second}
		
		resp, err := client.Get("https://api.github.com/repos/github/spec-kit/releases/latest")
		require.NoError(t, err)
		defer resp.Body.Close()

		var release GitHubRelease
		err = json.NewDecoder(resp.Body).Decode(&release)
		require.NoError(t, err, "should parse response structure")

		// Validate all fields our application depends on
		assert.NotZero(t, release.ID, "should have release ID")
		assert.NotEmpty(t, release.TagName, "should have tag name")
		assert.NotEmpty(t, release.Name, "should have release name")
		assert.False(t, release.CreatedAt.IsZero(), "should have creation date")
		assert.False(t, release.PublishedAt.IsZero(), "should have publication date")

		// Validate assets structure
		for _, asset := range release.Assets {
			assert.NotZero(t, asset.ID, "asset should have ID")
			assert.NotEmpty(t, asset.Name, "asset should have name")
			assert.Greater(t, asset.Size, int64(0), "asset should have positive size")
			assert.NotEmpty(t, asset.BrowserDownloadURL, "asset should have download URL")
			assert.NotEmpty(t, asset.ContentType, "asset should have content type")
			assert.False(t, asset.CreatedAt.IsZero(), "asset should have creation date")
			assert.False(t, asset.UpdatedAt.IsZero(), "asset should have update date")
		}
	})
}

// Helper functions
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && strings.Contains(s, substr)
}

func endsWithString(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}