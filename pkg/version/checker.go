package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VersionInfo represents version information
type VersionInfo struct {
	LatestVersion string `json:"latest_version"`
	CurrentVersion string
	UpdateAvailable bool
	ReleaseURL      string
}

// Checker checks for new versions
type Checker struct {
	versionURL string
	client     *http.Client
}

// NewChecker creates a new version checker
func NewChecker(versionURL string) *Checker {
	return &Checker{
		versionURL: versionURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// CheckForUpdate checks if a new version is available
func (c *Checker) CheckForUpdate(currentVersion string) (*VersionInfo, error) {
	info := &VersionInfo{
		CurrentVersion: currentVersion,
		UpdateAvailable: false,
	}

	// Try to fetch latest version from GitHub releases API
	// Format: https://api.github.com/repos/{owner}/{repo}/releases/latest
	// Or use a custom version endpoint
	req, err := http.NewRequest("GET", c.versionURL, nil)
	if err != nil {
		return info, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "uniroute-cli")

	resp, err := c.client.Do(req)
	if err != nil {
		// Network error - don't fail, just return no update
		return info, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return info, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return info, nil
	}

	// Try to parse as GitHub API response
	var githubRelease struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(body, &githubRelease); err == nil && githubRelease.TagName != "" {
		// Remove 'v' prefix if present
		latestVersion := strings.TrimPrefix(githubRelease.TagName, "v")
		info.LatestVersion = latestVersion
		info.ReleaseURL = githubRelease.HTMLURL
		info.UpdateAvailable = isNewerVersion(latestVersion, currentVersion)
		return info, nil
	}

	// Try to parse as custom version endpoint
	var customVersion struct {
		Version string `json:"version"`
		URL     string `json:"url"`
	}
	if err := json.Unmarshal(body, &customVersion); err == nil && customVersion.Version != "" {
		latestVersion := strings.TrimPrefix(customVersion.Version, "v")
		info.LatestVersion = latestVersion
		info.ReleaseURL = customVersion.URL
		info.UpdateAvailable = isNewerVersion(latestVersion, currentVersion)
		return info, nil
	}

	return info, nil
}

// isNewerVersion compares two version strings
// Returns true if latest > current
func isNewerVersion(latest, current string) bool {
	// Simple comparison - can be enhanced with proper semver parsing
	// For now, just compare strings (works for most cases)
	if latest == current {
		return false
	}
	
	// Remove 'v' prefix if present
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")
	
	// Simple string comparison (works for versions like 1.0.0, 1.0.1, etc.)
	return latest > current
}

