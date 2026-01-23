package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/color"
	"github.com/spf13/cobra"
)

var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage custom domains",
	Long: `Manage custom domains for your tunnels.

Add, list, verify, or remove custom domains from your account.
You can use your own domain instead of random subdomains.

Examples:
  uniroute domain example.com                    # Add domain to account
  uniroute domain example.com abc123             # Add domain AND assign to tunnel
  uniroute domain list                           # List all your domains
  uniroute domain show example.com               # Show domain details
  uniroute domain verify example.com             # Verify DNS configuration
  uniroute domain resume abc123                  # Resume domain assignment by subdomain
  uniroute domain resume example.com             # Resume domain assignment by domain
  uniroute domain remove example.com             # Remove domain from account`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no args, show help
		if len(args) == 0 {
			return cmd.Help()
		}

		domain := args[0]
		tunnelID, _ := cmd.Flags().GetString("tunnel-id")
		subdomain, _ := cmd.Flags().GetString("subdomain")
		
		// If subdomain provided as second positional argument (shortcut syntax)
		// Only use if --subdomain flag is not set (flag takes precedence)
		if len(args) > 1 && subdomain == "" {
			subdomain = args[1]
		}

		// Get auth token
		token := getAuthToken()
		if token == "" {
			return fmt.Errorf("authentication required\nRun 'uniroute auth login' first")
		}

		// Determine if we're assigning to a tunnel
		assignToTunnel := false
		if tunnelID != "" || subdomain != "" {
			assignToTunnel = true
		} else {
			// Try to get from last active tunnel (optional - don't fail if not found)
			homeDir, err := os.UserHomeDir()
			if err == nil {
				statePath := filepath.Join(homeDir, ".uniroute", "tunnel-state.json")
				if data, err := os.ReadFile(statePath); err == nil {
					var state tunnel.TunnelState
					if err := json.Unmarshal(data, &state); err == nil && state.TunnelID != "" {
						tunnelID = state.TunnelID
						assignToTunnel = true
					}
				}
			}
		}

		// If subdomain is provided, look up tunnel ID
		if assignToTunnel && tunnelID == "" && subdomain != "" {
			tunnelID = lookupTunnelIDBySubdomain(subdomain, token)
			if tunnelID == "" {
				return fmt.Errorf("tunnel with subdomain '%s' not found", subdomain)
			}
		}

		// Always create/add domain to the domain management system first
		// This is the same as adding domain in dashboard - both use /auth/domains endpoint
		if err := createDomainIfNotExists(domain, token); err != nil {
			// Log warning but continue - domain might already exist
			fmt.Println(color.Yellow("⚠️  Warning: Could not add domain to management system (it may already exist)"))
		} else {
			fmt.Println(color.Green("✓") + " Domain added to your account")
			fmt.Println(color.Gray("   View and manage it: https://app.uniroute.co/dashboard/domains"))
		}

		// If tunnel specified, assign domain to tunnel
		if assignToTunnel {
			if err := setCustomDomain(tunnelID, domain, token); err != nil {
				return fmt.Errorf("failed to assign domain to tunnel: %w", err)
			}
			
			// Save domain assignment for resume functionality
			if err := saveDomainAssignment(domain, subdomain, tunnelID); err != nil {
				// Don't fail if we can't save - just log a warning
				fmt.Println(color.Gray("   (Note: Could not save for resume)"))
			}
			
			fmt.Println(color.Green("✓") + " Domain assigned to tunnel")
			fmt.Printf("   Domain: %s\n", color.Cyan(domain))
			fmt.Printf("   Tunnel ID: %s\n", color.Gray(tunnelID))
			if subdomain != "" {
				fmt.Printf("   Subdomain: %s\n", color.Gray(subdomain))
				fmt.Println(color.Gray("   Resume with: uniroute domain resume " + subdomain))
			}
		} else {
			fmt.Println(color.Green("✓") + " Domain ready to use")
			fmt.Printf("   Domain: %s\n", color.Cyan(domain))
			fmt.Println(color.Gray("   To assign to a tunnel, run:"))
			fmt.Printf(color.Gray("   uniroute domain %s <tunnel-subdomain>\n"), domain)
		}

		fmt.Println()
		fmt.Println(color.Yellow("📋 Next Steps:"))
		fmt.Println()
		fmt.Println(color.Bold("1. Configure DNS:"))
		fmt.Printf("   Add a CNAME record in your DNS provider:\n")
		fmt.Printf("   %s → tunnel.uniroute.co\n", color.Cyan(domain))
		fmt.Println()
		fmt.Println(color.Bold("2. Verify DNS configuration:"))
		fmt.Println(color.Gray("   Go to: https://app.uniroute.co/dashboard/domains"))
		fmt.Println(color.Gray("   Find your domain and click 'Verify DNS' button"))
		fmt.Println()
		fmt.Println(color.Bold("3. Once verified, your domain is ready to use!"))
		fmt.Println(color.Gray("   The domain will automatically work with your tunnel(s)"))
		fmt.Println()

		return nil
	},
}

func init() {
	domainCmd.Flags().String("tunnel-id", "", "Tunnel ID to set domain for")
	domainCmd.Flags().String("subdomain", "", "Subdomain to set domain for")
	
	// Add subcommands
	domainCmd.AddCommand(domainListCmd)
	domainCmd.AddCommand(domainShowCmd)
	domainCmd.AddCommand(domainVerifyCmd)
	domainCmd.AddCommand(domainRemoveCmd)
	domainCmd.AddCommand(domainResumeCmd)
}

// domainListCmd lists all domains
var domainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your custom domains",
	Long:  `List all custom domains in your account with their status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := getAuthToken()
		if token == "" {
			return fmt.Errorf("authentication required\nRun 'uniroute auth login' first")
		}

		domains, err := listDomains(token)
		if err != nil {
			return err
		}

		if len(domains) == 0 {
			fmt.Println(color.Yellow("No domains found. Add one with: uniroute domain <domain>"))
			return nil
		}

		fmt.Println(color.Bold("Your Custom Domains:"))
		fmt.Println()
		for _, d := range domains {
			status := color.Red("❌ Not Configured")
			if d.DNSConfigured {
				status = color.Green("✓ DNS Configured")
			}
			
			fmt.Printf("  %s %s\n", color.Cyan(d.Domain), status)
			fmt.Printf("    ID: %s\n", color.Gray(d.ID))
			if d.DNSConfigured {
				fmt.Printf("    Status: %s\n", color.Green("Ready to use"))
			} else {
				fmt.Printf("    Status: %s\n", color.Yellow("DNS not configured"))
				fmt.Printf("    CNAME: %s → tunnel.uniroute.co\n", color.Gray(d.Domain))
			}
			fmt.Println()
		}

		return nil
	},
}

// domainShowCmd shows domain details
var domainShowCmd = &cobra.Command{
	Use:   "show [domain]",
	Short: "Show details for a specific domain",
	Long:  `Show detailed information about a custom domain including DNS configuration status.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		token := getAuthToken()
		if token == "" {
			return fmt.Errorf("authentication required\nRun 'uniroute auth login' first")
		}

		domains, err := listDomains(token)
		if err != nil {
			return err
		}

		var found *DomainInfo
		for _, d := range domains {
			if d.Domain == domain {
				found = &d
				break
			}
		}

		if found == nil {
			return fmt.Errorf("domain '%s' not found in your account", domain)
		}

		fmt.Println(color.Bold("Domain Details:"))
		fmt.Println()
		fmt.Printf("  Domain: %s\n", color.Cyan(found.Domain))
		fmt.Printf("  ID: %s\n", color.Gray(found.ID))
		fmt.Printf("  DNS Configured: ")
		if found.DNSConfigured {
			fmt.Println(color.Green("Yes ✓"))
		} else {
			fmt.Println(color.Red("No ✗"))
		}
		fmt.Println()
		
		if !found.DNSConfigured {
			fmt.Println(color.Yellow("DNS Configuration Required:"))
			fmt.Printf("  Add CNAME record: %s → tunnel.uniroute.co\n", color.Cyan(found.Domain))
			fmt.Println()
			fmt.Println("  Then verify with: uniroute domain verify " + found.Domain)
		} else {
			fmt.Println(color.Green("✓ Domain is ready to use!"))
		}

		return nil
	},
}

// domainVerifyCmd verifies DNS configuration
var domainVerifyCmd = &cobra.Command{
	Use:   "verify [domain]",
	Short: "Verify DNS configuration for a domain",
	Long:  `Check if DNS (CNAME) is properly configured for your domain.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		token := getAuthToken()
		if token == "" {
			return fmt.Errorf("authentication required\nRun 'uniroute auth login' first")
		}

		// First find domain ID
		domains, err := listDomains(token)
		if err != nil {
			return err
		}

		var domainID string
		for _, d := range domains {
			if d.Domain == domain {
				domainID = d.ID
				break
			}
		}

		if domainID == "" {
			return fmt.Errorf("domain '%s' not found in your account", domain)
		}

		// Verify DNS
		result, err := verifyDomain(domainID, token)
		if err != nil {
			return err
		}

		fmt.Printf("Verifying DNS for %s...\n", color.Cyan(domain))
		fmt.Println()

		if result.DNSConfigured {
			fmt.Println(color.Green("✓ DNS is properly configured!"))
			fmt.Println(color.Green("  Your domain is ready to use."))
		} else {
			fmt.Println(color.Red("✗ DNS is not configured correctly"))
			if result.DNSError != "" {
				fmt.Printf("  Error: %s\n", color.Red(result.DNSError))
			}
			fmt.Println()
			fmt.Println(color.Yellow("Please configure your DNS:"))
			fmt.Printf("  Add CNAME record: %s → tunnel.uniroute.co\n", color.Cyan(domain))
			fmt.Println()
			fmt.Println("  After configuring, run this command again to verify.")
		}

		return nil
	},
}

// domainRemoveCmd removes a domain
var domainRemoveCmd = &cobra.Command{
	Use:   "remove [domain]",
	Short: "Remove a custom domain from your account",
	Long:  `Remove a custom domain from your account. This will also unassign it from any tunnels.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		token := getAuthToken()
		if token == "" {
			return fmt.Errorf("authentication required\nRun 'uniroute auth login' first")
		}

		// First find domain ID
		domains, err := listDomains(token)
		if err != nil {
			return err
		}

		var domainID string
		for _, d := range domains {
			if d.Domain == domain {
				domainID = d.ID
				break
			}
		}

		if domainID == "" {
			return fmt.Errorf("domain '%s' not found in your account", domain)
		}

		// Confirm removal
		fmt.Printf("Are you sure you want to remove %s? (yes/no): ", color.Cyan(domain))
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}

		// Remove domain
		if err := removeDomain(domainID, token); err != nil {
			return err
		}

		// Also remove from saved assignments
		removeDomainAssignment(domain)

		fmt.Println(color.Green("✓ Domain removed successfully"))
		return nil
	},
}

// domainResumeCmd resumes a domain assignment
var domainResumeCmd = &cobra.Command{
	Use:   "resume [domain|subdomain]",
	Short: "Resume a domain assignment to a tunnel",
	Long: `Resume a previously assigned domain to its tunnel.

You can resume by domain name or subdomain. If no argument is provided,
it will resume the most recently used domain assignment.

Examples:
  uniroute domain resume                    # Resume last used assignment
  uniroute domain resume naijacrawl.com    # Resume by domain name
  uniroute domain resume abc123              # Resume by subdomain`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var identifier string
		if len(args) > 0 {
			identifier = args[0]
		}
		token := getAuthToken()
		if token == "" {
			return fmt.Errorf("authentication required\nRun 'uniroute auth login' first")
		}

		// Load saved domain assignments
		assignments, err := loadDomainAssignments()
		if err != nil {
			return fmt.Errorf("failed to load saved assignments: %w", err)
		}

		if len(assignments) == 0 {
			return fmt.Errorf("no saved domain assignments found\nUse 'uniroute domain <domain> <subdomain>' to create one first")
		}

		// Find assignment by domain or subdomain, or use most recent
		var assignment *DomainAssignment
		if identifier == "" {
			// No identifier provided - use most recently used
			mostRecent := assignments[0]
			for _, a := range assignments {
				if a.LastUsed.After(mostRecent.LastUsed) {
					mostRecent = a
				}
			}
			assignment = &mostRecent
		} else {
			// Find by identifier
			for _, a := range assignments {
				if a.Domain == identifier || a.Subdomain == identifier {
					assignment = &a
					break
				}
			}
		}

		if assignment == nil {
			return fmt.Errorf("no saved assignment found for '%s'\nUse 'uniroute domain <domain> <subdomain>' to create one first", identifier)
		}

		fmt.Printf("Resuming domain assignment...\n")
		fmt.Printf("  Domain: %s\n", color.Cyan(assignment.Domain))
		fmt.Printf("  Subdomain: %s\n", color.Gray(assignment.Subdomain))
		fmt.Printf("  Tunnel ID: %s\n", color.Gray(assignment.TunnelID))
		fmt.Println()

		// Ensure domain exists in account
		if err := createDomainIfNotExists(assignment.Domain, token); err != nil {
			fmt.Println(color.Yellow("⚠️  Warning: Could not verify domain in account (it may already exist)"))
		}

		// Look up tunnel ID by subdomain (in case it changed)
		currentTunnelID := lookupTunnelIDBySubdomain(assignment.Subdomain, token)
		if currentTunnelID == "" {
			return fmt.Errorf("tunnel with subdomain '%s' not found. The tunnel may have been deleted.", assignment.Subdomain)
		}

		// Assign domain to tunnel
		if err := setCustomDomain(currentTunnelID, assignment.Domain, token); err != nil {
			return fmt.Errorf("failed to assign domain to tunnel: %w", err)
		}

		// Update saved assignment with current tunnel ID
		assignment.TunnelID = currentTunnelID
		assignment.LastUsed = time.Now()
		if err := saveDomainAssignment(assignment.Domain, assignment.Subdomain, assignment.TunnelID); err != nil {
			fmt.Println(color.Gray("   (Note: Could not update saved assignment)"))
		}

		fmt.Println(color.Green("✓ Domain assignment resumed successfully"))
		fmt.Printf("   Domain: %s\n", color.Cyan(assignment.Domain))
		fmt.Printf("   Tunnel: %s\n", color.Gray(assignment.Subdomain))
		fmt.Println()
		fmt.Println(color.Gray("   The domain is now assigned to your tunnel"))

		return nil
	},
}

// DomainInfo represents a domain from the API
type DomainInfo struct {
	ID             string
	Domain         string
	DNSConfigured bool
}

// listDomains fetches all domains from the API
func listDomains(token string) ([]DomainInfo, error) {
	apiURL := getAPIURL()
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auth/domains", apiURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list domains: status %d", resp.StatusCode)
	}

	var result struct {
		Domains []struct {
			ID             string `json:"id"`
			Domain         string `json:"domain"`
			DNSConfigured  bool   `json:"dns_configured"`
		} `json:"domains"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	domains := make([]DomainInfo, len(result.Domains))
	for i, d := range result.Domains {
		domains[i] = DomainInfo{
			ID:             d.ID,
			Domain:         d.Domain,
			DNSConfigured:  d.DNSConfigured,
		}
	}

	return domains, nil
}

// VerifyResult represents the result of domain verification
type VerifyResult struct {
	DNSConfigured bool   `json:"dns_configured"`
	DNSError      string `json:"dns_error"`
}

// verifyDomain verifies DNS for a domain
func verifyDomain(domainID, token string) (VerifyResult, error) {
	var result VerifyResult

	apiURL := getAPIURL()
	client := &http.Client{Timeout: 30 * time.Second} // Longer timeout for DNS checks
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/auth/domains/%s/verify", apiURL, domainID), nil)
	if err != nil {
		return result, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("failed to verify domain: status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// removeDomain removes a domain
func removeDomain(domainID, token string) error {
	apiURL := getAPIURL()
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/auth/domains/%s", apiURL, domainID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to remove domain: status %d", resp.StatusCode)
	}

	return nil
}

// getAPIURL gets the API URL from config or environment
func getAPIURL() string {
	var apiURL string
	if envURL := os.Getenv("UNIROUTE_API_URL"); envURL != "" {
		apiURL = envURL
	} else {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPath := filepath.Join(homeDir, ".uniroute", "auth.json")
			if data, err := os.ReadFile(configPath); err == nil {
				var authConfig struct {
					ServerURL string `json:"server_url"`
				}
				if err := json.Unmarshal(data, &authConfig); err == nil && authConfig.ServerURL != "" {
					apiURL = authConfig.ServerURL
				}
			}
		}
		if apiURL == "" {
			apiURL = "https://api.uniroute.co"
		}
	}

	// Ensure URL has protocol
	if !strings.HasPrefix(apiURL, "http://") && !strings.HasPrefix(apiURL, "https://") {
		apiURL = "https://" + apiURL
	}

	return apiURL
}

// DomainAssignment represents a saved domain-to-tunnel assignment
type DomainAssignment struct {
	Domain    string    `json:"domain"`
	Subdomain string    `json:"subdomain"`
	TunnelID  string    `json:"tunnel_id"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

// getDomainStatePath returns the path to the domain state file
func getDomainStatePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".uniroute")
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "domain-assignments.json")
}

// saveDomainAssignment saves a domain assignment for resume functionality
func saveDomainAssignment(domain, subdomain, tunnelID string) error {
	if domain == "" || subdomain == "" || tunnelID == "" {
		return nil // Don't save incomplete assignments
	}

	filePath := getDomainStatePath()
	
	// Load existing assignments
	assignments, _ := loadDomainAssignments()
	
	// Update or add assignment
	found := false
	for i, a := range assignments {
		if a.Domain == domain || a.Subdomain == subdomain {
			assignments[i] = DomainAssignment{
				Domain:    domain,
				Subdomain: subdomain,
				TunnelID:  tunnelID,
				CreatedAt: a.CreatedAt, // Preserve original creation time
				LastUsed:  time.Now(),
			}
			found = true
			break
		}
	}
	
	if !found {
		assignments = append(assignments, DomainAssignment{
			Domain:    domain,
			Subdomain: subdomain,
			TunnelID:  tunnelID,
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
		})
	}
	
	// Save to file
	data, err := json.MarshalIndent(assignments, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal domain assignments: %w", err)
	}
	
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write domain assignments: %w", err)
	}
	
	return nil
}

// loadDomainAssignments loads saved domain assignments
func loadDomainAssignments() ([]DomainAssignment, error) {
	filePath := getDomainStatePath()
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []DomainAssignment{}, nil // No saved assignments
		}
		return nil, fmt.Errorf("failed to read domain assignments: %w", err)
	}
	
	var assignments []DomainAssignment
	if err := json.Unmarshal(data, &assignments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal domain assignments: %w", err)
	}
	
	return assignments, nil
}

// removeDomainAssignment removes a domain assignment from saved state
func removeDomainAssignment(domain string) {
	assignments, err := loadDomainAssignments()
	if err != nil {
		return // Silently fail
	}
	
	// Filter out the domain
	filtered := []DomainAssignment{}
	for _, a := range assignments {
		if a.Domain != domain {
			filtered = append(filtered, a)
		}
	}
	
	// Save filtered list
	if len(filtered) == 0 {
		// Remove file if no assignments left
		os.Remove(getDomainStatePath())
		return
	}
	
	data, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return // Silently fail
	}
	
	os.WriteFile(getDomainStatePath(), data, 0600)
}

// lookupTunnelIDBySubdomain looks up a tunnel ID by subdomain via API
func lookupTunnelIDBySubdomain(subdomain, token string) string {
	apiURL := getAPIURL()

	// List tunnels and find by subdomain
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/tunnels", apiURL), nil)
	if err != nil {
		return ""
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		Tunnels []struct {
			ID        string `json:"id"`
			Subdomain string `json:"subdomain"`
		} `json:"tunnels"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	for _, t := range result.Tunnels {
		if t.Subdomain == subdomain {
			return t.ID
		}
	}

	return ""
}

// createDomainIfNotExists creates a domain in the domain management system
func createDomainIfNotExists(domain, token string) error {
	apiURL := getAPIURL()

	// Create request to add domain to management system
	reqBody := map[string]string{"domain": domain}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/auth/domains", apiURL), strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 409 Conflict means domain already exists, which is fine
	if resp.StatusCode == http.StatusConflict {
		return nil // Domain already exists, that's okay
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Don't fail if domain creation fails - it might already exist
		// Just return an error that will be logged as a warning
		return fmt.Errorf("domain creation returned status %d", resp.StatusCode)
	}

	return nil
}
