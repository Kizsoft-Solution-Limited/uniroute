package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage your projects",
	Long: `Manage your UniRoute projects.

Commands:
  list    List all your projects
  show    Show details of a specific project`,
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your projects",
	Long:  `List all projects associated with your account.`,
	RunE:  runProjectsList,
}

var projectsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show project details",
	Long:  `Show detailed information about a specific project.`,
	RunE:  runProjectsShow,
}

var (
	projectsServerURL string
	projectID         string
)

func init() {
	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsShowCmd)

	projectsListCmd.Flags().StringVarP(&projectsServerURL, "server", "s", "", "UniRoute server URL (default: public server)")
	projectsShowCmd.Flags().StringVarP(&projectsServerURL, "server", "s", "", "UniRoute server URL (default: public server)")
	projectsShowCmd.Flags().StringVarP(&projectID, "id", "i", "", "Project ID")
}

func runProjectsList(cmd *cobra.Command, args []string) error {
	serverURL := projectsServerURL
	if serverURL == "" {
		serverURL = getServerURL()
	}

	token := getAuthToken()
	if token == "" {
		return fmt.Errorf("not authenticated. Run 'uniroute auth login' first")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/projects", serverURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s", string(body))
	}

	var projects []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found. Create one at https://app.uniroute.dev")
		return nil
	}

	fmt.Printf("📋 Your Projects (%d):\n\n", len(projects))
	for i, project := range projects {
		fmt.Printf("%d. %s\n", i+1, project["name"])
		if id, ok := project["id"].(string); ok {
			fmt.Printf("   ID: %s\n", id)
		}
		if desc, ok := project["description"].(string); ok && desc != "" {
			fmt.Printf("   Description: %s\n", desc)
		}
		fmt.Println()
	}

	return nil
}

func runProjectsShow(cmd *cobra.Command, args []string) error {
	if projectID == "" {
		return fmt.Errorf("project ID required. Use --id or provide as argument")
	}

	if len(args) > 0 {
		projectID = args[0]
	}

	serverURL := projectsServerURL
	if serverURL == "" {
		serverURL = getServerURL()
	}

	token := getAuthToken()
	if token == "" {
		return fmt.Errorf("not authenticated. Run 'uniroute auth login' first")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/projects/%s", serverURL, projectID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s", string(body))
	}

	var project map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Println("📋 Project Details:")
	fmt.Println()
	for key, value := range project {
		fmt.Printf("   %s: %v\n", key, value)
	}

	return nil
}

