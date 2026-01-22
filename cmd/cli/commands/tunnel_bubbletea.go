package commands

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
	"github.com/Kizsoft-Solution-Limited/uniroute/pkg/color"
	versioncheck "github.com/Kizsoft-Solution-Limited/uniroute/pkg/version"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Funny quotes to display
var tunnelQuotes = []string{
	"💬 Your app is now accessible worldwide!",
	"🚀 Your local server just went global!",
	"🌍 Breaking down firewalls, one request at a time!",
	"⚡ Your code is now faster than your WiFi!",
	"🎯 Tunnel vision? More like tunnel success!",
	"🔥 Your app is so hot, it's breaking the internet!",
	"💪 Local development, global domination!",
	"🎉 Congratulations! Your app just got a passport!",
	"🌟 From localhost to everywhere!",
	"🚦 Green light! Your tunnel is live!",
	"🎪 Welcome to the greatest show on the internet!",
	"🎸 Your app is now rockin' the web!",
	"🏆 You've just won the tunnel lottery!",
	"🎨 Your app is now a masterpiece on display!",
	"🎭 The show must go on, and it's going global!",
	"🎯 Bullseye! Your tunnel is perfectly aimed!",
	"🎪 Step right up! Your app is now on stage!",
	"🎊 Party time! Your tunnel is celebrating!",
	"🎁 Your app just got the best gift: global access!",
	"🎪 The circus is in town, and your app is the star!",
}

// Bubble Tea model for tunnel UI
type tunnelModel struct {
	// Fixed header fields
	connectionStatus string
	sessionStatus    string
	account          string
	version          string
	region           string
	latency          string
	publicURL        string
	forwarding       string
	connections      string
	quote            string // Random funny quote

	// Scrolling content
	viewport viewport.Model
	logs     []string

	// State
	client         *tunnel.TunnelClient
	info           *tunnel.TunnelInfo
	serverURL      string
	localURL       string
	updateInfo     *versioncheck.VersionInfo
	updateInfoMu   sync.Mutex
	internetOnline bool
	wasOffline     bool      // Track if we were offline to show reconnecting when back
	reconnectTime  time.Time // Time when internet came back
	terminated     bool      // Track if tunnel was terminated
	ctx            context.Context
	cancel         context.CancelFunc

	// Styles (using Lip Gloss)
	headerStyle     lipgloss.Style
	labelStyle      lipgloss.Style
	valueStyle      lipgloss.Style
	statusGreen     lipgloss.Style
	statusYellow    lipgloss.Style
	statusRed       lipgloss.Style
	timeStyle       lipgloss.Style
	methodStyle     lipgloss.Style
	pathStyle       lipgloss.Style
	statusCodeStyle lipgloss.Style
}

// Messages for Bubble Tea
type internetStatusMsg bool
type latencyUpdateMsg int64
type connectionStatusMsg string
type sessionStatusMsg string
type requestEventMsg tunnel.RequestEvent
type statsUpdateMsg *tunnel.ConnectionStats
type versionUpdateMsg *versioncheck.VersionInfo
type updateProgressMsg string
type terminateMsg struct{} // Message to signal termination

// Initial model
func initialTunnelModel(client *tunnel.TunnelClient, info *tunnel.TunnelInfo, accountDisplay string, serverURL string, localURL string) *tunnelModel {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize viewport with proper dimensions
	// Will be updated on WindowSizeMsg
	vp := viewport.New(80, 20)
	vp.SetContent("") // Start empty, will be populated with request logs

	// Get initial version
	currentVersion := GetVersion()

	// Pick a random quote
	rand.Seed(time.Now().UnixNano())
	randomQuote := tunnelQuotes[rand.Intn(len(tunnelQuotes))]

	// Initialize styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")) // Cyan
	labelStyle := lipgloss.NewStyle().Width(25).Foreground(lipgloss.Color("244"))  // Gray
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))            // Gray
	statusGreen := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))            // Green
	statusYellow := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))          // Yellow
	statusRed := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))             // Red
	timeStyle := lipgloss.NewStyle().Width(20).Foreground(lipgloss.Color("244"))
	methodStyle := lipgloss.NewStyle().Width(7).Foreground(lipgloss.Color("39"))
	pathStyle := lipgloss.NewStyle().Width(50).Foreground(lipgloss.Color("255"))
	statusCodeStyle := lipgloss.NewStyle().Width(3)

	// Get initial stats
	initialStats := &tunnel.ConnectionStats{
		Total: 0,
		Open:  0,
		RT1:   0,
		RT5:   0,
		P50:   0,
		P90:   0,
	}
	if info != nil {
		if stats, err := client.GetConnectionStats(serverURL, info.ID); err == nil {
			initialStats = stats
		}
	}

	return &tunnelModel{
		connectionStatus: color.Green("Tunnel Connected Successfully!"),
		sessionStatus:    fmt.Sprintf("%s %s", color.Green("●"), color.Green("online")),
		account:          accountDisplay,
		version:          color.Gray(currentVersion),
		region:           "Local",
		latency:          "0ms",
		publicURL:        info.PublicURL,
		forwarding:       fmt.Sprintf("%s %s %s", color.Cyan(info.PublicURL), color.Gray("->"), color.Bold(localURL)),
		connections: fmt.Sprintf("%s      %s      %s      %s      %s      %s",
			color.Gray(fmt.Sprintf("%d", initialStats.Total)),
			color.Gray(fmt.Sprintf("%d", initialStats.Open)),
			color.Gray(fmt.Sprintf("%.2f", initialStats.RT1)),
			color.Gray(fmt.Sprintf("%.2f", initialStats.RT5)),
			color.Gray(fmt.Sprintf("%.2f", initialStats.P50)),
			color.Gray(fmt.Sprintf("%.2f", initialStats.P90))),
		quote:           randomQuote,
		viewport:        vp,
		logs:            []string{},
		client:          client,
		info:            info,
		serverURL:       serverURL,
		localURL:        localURL,
		internetOnline:  true,
		wasOffline:      false,
		reconnectTime:   time.Time{},
		ctx:             ctx,
		cancel:          cancel,
		headerStyle:     headerStyle,
		labelStyle:      labelStyle,
		valueStyle:      valueStyle,
		statusGreen:     statusGreen,
		statusYellow:    statusYellow,
		statusRed:       statusRed,
		timeStyle:       timeStyle,
		methodStyle:     methodStyle,
		pathStyle:       pathStyle,
		statusCodeStyle: statusCodeStyle,
	}
}

// Init function for Bubble Tea
func (m *tunnelModel) Init() tea.Cmd {
	return tea.Batch(
		m.checkInternet(),
		m.updateStatus(),
		m.checkForUpdates(),
	)
}

// Update function for Bubble Tea
func (m *tunnelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Shutdown - send terminate message to update status first
			if !m.terminated {
				// Cancel context to stop tunnel
				m.cancel()
				// Send terminate message which will update status and quit
				return m, func() tea.Msg {
					return terminateMsg{}
				}
			}
			return m, tea.Quit
		case "ctrl+u":
			// Trigger update
			m.updateInfoMu.Lock()
			hasUpdate := m.updateInfo != nil && m.updateInfo.UpdateAvailable
			m.updateInfoMu.Unlock()

			if hasUpdate {
				return m, tea.Batch(
					m.downloadUpdate(),
					m.runUpgrade(),
				)
			}
			return m, nil
		}
	case internetStatusMsg:
		// Don't update if terminated
		if m.terminated {
			return m, nil
		}
		wasOfflineBefore := !m.internetOnline
		m.internetOnline = bool(msg)
		if !m.internetOnline {
			// Internet is offline - update both connection status and session status
			m.connectionStatus = color.Red("No Internet - Connection Lost")
			m.sessionStatus = fmt.Sprintf("%s %s", color.Red("●"), color.Red("offline"))
			m.wasOffline = true
			m.reconnectTime = time.Time{} // Reset reconnect time
		} else {
			// Internet is back
			if wasOfflineBefore || m.wasOffline {
				// We were offline, so show reconnecting first
				m.connectionStatus = color.Yellow("Reconnecting...")
				m.sessionStatus = fmt.Sprintf("%s %s", color.Yellow("●"), color.Yellow("reconnecting"))
				m.reconnectTime = time.Now()
				m.wasOffline = false // Reset flag
			} else {
				// Internet was already online, check actual connection status
				status := m.client.GetConnectionStatus()
				if status == "online" {
					m.connectionStatus = color.Green("Tunnel Connected Successfully!")
					m.sessionStatus = fmt.Sprintf("%s %s", color.Green("●"), color.Green("online"))
				} else if status == "reconnecting" {
					m.connectionStatus = color.Yellow("Reconnecting...")
					m.sessionStatus = fmt.Sprintf("%s %s", color.Yellow("●"), color.Yellow("reconnecting"))
				} else {
					m.connectionStatus = color.Yellow("Reconnecting...")
					m.sessionStatus = fmt.Sprintf("%s %s", color.Yellow("●"), color.Yellow("reconnecting"))
				}
			}
		}
		return m, m.checkInternet()
	case connectionStatusMsg:
		// Don't update if terminated
		if m.terminated {
			return m, nil
		}
		status := string(msg)
		// Only update if internet is online, otherwise internetStatusMsg handler will handle it
		if m.internetOnline {
			// If we just reconnected (within last 2 seconds), stay in reconnecting state
			// to give the tunnel time to fully establish
			if !m.reconnectTime.IsZero() && time.Since(m.reconnectTime) < 2*time.Second {
				// Still in reconnecting window, check if tunnel is actually online
				if status == "online" {
					// Wait a bit more to ensure stable connection
					m.connectionStatus = color.Yellow("Reconnecting...")
					m.sessionStatus = fmt.Sprintf("%s %s", color.Yellow("●"), color.Yellow("reconnecting"))
				} else {
					m.connectionStatus = color.Yellow("Reconnecting...")
					m.sessionStatus = fmt.Sprintf("%s %s", color.Yellow("●"), color.Yellow("reconnecting"))
				}
			} else {
				// Reconnecting window passed, update based on actual status
				if !m.reconnectTime.IsZero() {
					m.reconnectTime = time.Time{} // Clear reconnect time
				}
				switch status {
				case "online":
					m.connectionStatus = color.Green("Tunnel Connected Successfully!")
				case "reconnecting":
					m.connectionStatus = color.Yellow("Reconnecting...")
				default:
					m.connectionStatus = color.Red("No Internet - Connection Lost")
				}
			}
		}
		// Update related fields and continue polling
		currentStatus := m.client.GetConnectionStatus()
		latency := m.client.GetLatency()
		var stats *tunnel.ConnectionStats
		if currentStatus == "online" && m.info != nil {
			stats, _ = m.client.GetConnectionStats(m.serverURL, m.info.ID)
		}
		// Send updates for related fields
		return m, tea.Batch(
			func() tea.Msg { return sessionStatusMsg(currentStatus) },
			func() tea.Msg { return latencyUpdateMsg(latency) },
			func() tea.Msg { return statsUpdateMsg(stats) },
			m.updateStatus(), // Continue polling
		)
	case sessionStatusMsg:
		status := string(msg)
		// Only update session status if internet is online
		// If internet is offline, it's already handled by internetStatusMsg
		if m.internetOnline {
			// If we just reconnected (within last 2 seconds), stay in reconnecting state
			if !m.reconnectTime.IsZero() && time.Since(m.reconnectTime) < 2*time.Second {
				m.sessionStatus = fmt.Sprintf("%s %s", color.Yellow("●"), color.Yellow("reconnecting"))
			} else {
				switch status {
				case "online":
					m.sessionStatus = fmt.Sprintf("%s %s", color.Green("●"), color.Green("online"))
				case "reconnecting":
					m.sessionStatus = fmt.Sprintf("%s %s", color.Yellow("●"), color.Yellow("reconnecting"))
				default:
					m.sessionStatus = fmt.Sprintf("%s %s", color.Red("●"), color.Red("offline"))
				}
			}
		}
		return m, nil
	case terminateMsg:
		// Update status to terminated
		if !m.terminated {
			m.terminated = true
			m.sessionStatus = fmt.Sprintf("%s %s", color.Red("●"), color.Red("terminated"))
			m.connectionStatus = color.Red("Tunnel Terminated")
			// Cancel context to stop tunnel
			m.cancel()
		}
		// Quit after showing terminated status
		return m, tea.Quit
	case latencyUpdateMsg:
		latency := int64(msg)
		if latency > 0 {
			m.latency = fmt.Sprintf("%dms", latency)
		} else {
			m.latency = "0ms"
		}
		return m, nil
	case requestEventMsg:
		event := tunnel.RequestEvent(msg)
		m.addRequestLog(event)
		return m, nil
	case statsUpdateMsg:
		stats := msg
		if stats != nil {
			m.connections = fmt.Sprintf("%s      %s      %s      %s      %s      %s",
				color.Gray(fmt.Sprintf("%d", stats.Total)),
				color.Gray(fmt.Sprintf("%d", stats.Open)),
				color.Gray(fmt.Sprintf("%.2f", stats.RT1)),
				color.Gray(fmt.Sprintf("%.2f", stats.RT5)),
				color.Gray(fmt.Sprintf("%.2f", stats.P50)),
				color.Gray(fmt.Sprintf("%.2f", stats.P90)))
		}
		return m, nil
	case versionUpdateMsg:
		m.updateInfoMu.Lock()
		m.updateInfo = msg
		if msg != nil && msg.UpdateAvailable {
			m.version = fmt.Sprintf("%s %s %s",
				color.Gray(GetVersion()),
				color.Yellow("(Update available)"),
				color.Gray("Press Ctrl+U to download"))
		} else {
			m.version = color.Gray(GetVersion())
		}
		m.updateInfoMu.Unlock()
		return m, nil
	case updateProgressMsg:
		m.version = color.Yellow(string(msg))
		return m, nil
	case tea.WindowSizeMsg:
		// Update viewport size
		// Calculate actual header height: connection status (1) + blank (1) + session (1) + account (1) + version (1) + blank (1) + region (1) + latency (1) + blank (1) + connections header (1) + connections data (1) + blank (1) + public url label (1) + public url (1) + blank (1) + forwarding label (1) + forwarding (1) + blank (1) + http requests header (2) = ~20 lines
		headerHeight := 20
		if msg.Height > headerHeight {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight - 3 // -3 for footer message
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = 5 // Minimum height
		}
		return m, nil
	}

	// Handle viewport scrolling
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View function for Bubble Tea
func (m *tunnelModel) View() string {
	// Fixed header section
	header := strings.Builder{}

	// Connection status
	header.WriteString(m.connectionStatus)
	header.WriteString("\n\n")

	// Session Status
	header.WriteString(fmt.Sprintf("Session Status                %s\n", m.sessionStatus))

	// Account
	header.WriteString(fmt.Sprintf("Account                       %s\n", color.Gray(m.account)))

	// Version
	header.WriteString(fmt.Sprintf("Version                       %s\n", m.version))
	header.WriteString("\n")

	// Region
	header.WriteString(fmt.Sprintf("Region                        %s\n", color.Gray(m.region)))

	// Latency
	header.WriteString(fmt.Sprintf("Latency                       %s\n", color.Gray(m.latency)))
	header.WriteString("\n")

	// Connections
	header.WriteString("Connections                   ttl     opn     rt1     rt5     p50     p90\n")
	header.WriteString(fmt.Sprintf("                              %s\n\n", m.connections))

	// Public URL
	header.WriteString("🌍 Public URL:\n")
	header.WriteString(fmt.Sprintf("   %s\n", color.Cyan(m.publicURL)))
	header.WriteString("\n")

	// Forwarding
	header.WriteString("🔗 Forwarding:\n")
	header.WriteString(fmt.Sprintf("   %s\n", m.forwarding))
	header.WriteString("\n")

	// HTTP Requests header
	header.WriteString("HTTP Requests\n")
	header.WriteString("-------------\n")

	// Combine header with scrolling viewport
	return header.String() + m.viewport.View() + "\n\n" + color.Yellow(m.quote) + "\n\nPress Ctrl+C to stop"
}

// Commands
func (m *tunnelModel) checkInternet() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		// Check internet connectivity with a quick timeout
		client := &http.Client{Timeout: 2 * time.Second}
		_, err := client.Head("http://clients3.google.com/generate_204")
		// Also check if we can reach the tunnel server
		if err == nil {
			// Internet is up, but also check tunnel server connectivity
			tunnelClient := &http.Client{Timeout: 1 * time.Second}
			_, tunnelErr := tunnelClient.Get(fmt.Sprintf("http://%s/health", m.serverURL))
			if tunnelErr != nil {
				// Tunnel server unreachable even though internet is up
				return internetStatusMsg(false)
			}
		}
		return internetStatusMsg(err == nil)
	})
}

func (m *tunnelModel) updateStatus() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		// Check connection status
		status := m.client.GetConnectionStatus()
		if status == "" {
			return nil
		}

		// Send connection status update
		// Send session status update
		// Send latency update
		latency := m.client.GetLatency()

		// Fetch stats if connected
		var stats *tunnel.ConnectionStats
		if status == "online" && m.info != nil {
			statsChan := make(chan *tunnel.ConnectionStats, 1)
			go func() {
				s, _ := m.client.GetConnectionStats(m.serverURL, m.info.ID)
				statsChan <- s
			}()
			select {
			case stats = <-statsChan:
			case <-time.After(2 * time.Second):
			}
		}

		// Return batch of messages
		return tea.Batch(
			func() tea.Msg { return connectionStatusMsg(status) },
			func() tea.Msg { return sessionStatusMsg(status) },
			func() tea.Msg { return latencyUpdateMsg(latency) },
			func() tea.Msg { return statsUpdateMsg(stats) },
		)()
	})
}

func (m *tunnelModel) checkForUpdates() tea.Cmd {
	return func() tea.Msg {
		versionURL := os.Getenv("UNIROUTE_VERSION_URL")
		if versionURL == "" {
			versionURL = "https://api.github.com/repos/Kizsoft-Solution-Limited/uniroute/releases/latest"
		}
		checker := versioncheck.NewChecker(versionURL)
		info, _ := checker.CheckForUpdate(GetVersion())
		return versionUpdateMsg(info)
	}
}

func (m *tunnelModel) downloadUpdate() tea.Cmd {
	return func() tea.Msg {
		// Show downloading status immediately
		// This will be sent as a message
		return updateProgressMsg("Downloading update...")
	}
}

func (m *tunnelModel) runUpgrade() tea.Cmd {
	return func() tea.Msg {
		// Run upgrade command with --yes flag to auto-confirm
		upgradeCmd := exec.Command(os.Args[0], "upgrade", "--yes")
		upgradeCmd.Stdout = os.Stderr
		upgradeCmd.Stderr = os.Stderr
		upgradeCmd.Stdin = os.Stdin // Ensure stdin is connected

		// Run in background with timeout to prevent hanging
		errChan := make(chan error, 1)
		go func() {
			errChan <- upgradeCmd.Run()
		}()

		// Wait for completion or timeout (30 seconds)
		select {
		case err := <-errChan:
			if err != nil {
				return updateProgressMsg("Update failed - run 'uniroute upgrade' manually")
			}
			return updateProgressMsg("✓ Update downloaded! Close and restart to apply")
		case <-time.After(30 * time.Second):
			// Timeout - kill the process
			if upgradeCmd.Process != nil {
				upgradeCmd.Process.Kill()
			}
			return updateProgressMsg("Update timed out - run 'uniroute upgrade' manually")
		}
	}
}

func (m *tunnelModel) addRequestLog(event tunnel.RequestEvent) {
	// Get timezone abbreviation (WAT, EST, etc.)
	loc, _ := time.LoadLocation("Africa/Lagos") // Default to WAT, can be made configurable
	eventTime := event.Time.In(loc)
	timeStr := eventTime.Format("15:04:05.000 MST")

	// Truncate path if too long (max 50 chars for display)
	path := event.Path
	if len(path) > 50 {
		path = path[:47] + "..."
	}

	// Format status code and text
	statusCodeStr := fmt.Sprintf("%d", event.StatusCode)
	statusText := event.StatusText
	// Extract status text (e.g., "200 OK" -> "OK")
	if strings.Contains(statusText, " ") {
		parts := strings.SplitN(statusText, " ", 2)
		if len(parts) > 1 {
			statusText = parts[1]
		}
	}
	// Truncate status text if too long (max 10 chars)
	if len(statusText) > 10 {
		statusText = statusText[:7] + "..."
	}

	// Color code based on status
	var statusColor func(string) string
	if event.StatusCode >= 200 && event.StatusCode < 300 {
		statusColor = color.Green
	} else if event.StatusCode >= 300 && event.StatusCode < 400 {
		statusColor = color.Yellow
	} else if event.StatusCode >= 400 && event.StatusCode < 500 {
		statusColor = color.Red
	} else if event.StatusCode >= 500 {
		statusColor = color.Red
	} else {
		statusColor = color.Gray
	}

	// Helper function to strip ANSI escape sequences
	stripANSI := func(s string) int {
		var result strings.Builder
		inEscape := false
		for i := 0; i < len(s); i++ {
			if s[i] == '\033' {
				inEscape = true
			} else if inEscape {
				if s[i] == 'm' {
					inEscape = false
				}
			} else {
				result.WriteByte(s[i])
			}
		}
		return len(result.String())
	}

	// Helper function to pad colored string to fixed display width
	padColored := func(s string, width int) string {
		displayWidth := stripANSI(s)
		if displayWidth >= width {
			return s
		}
		// Add padding spaces to reach target width
		padding := strings.Repeat(" ", width-displayWidth)
		return s + padding
	}

	// Apply colors first
	timeColored := color.Gray(timeStr)
	methodColored := color.Cyan(event.Method)
	pathColored := path
	statusCodeColored := statusColor(statusCodeStr)
	statusTextColored := statusColor(statusText)

	// Pad each field to its target width (accounting for ANSI codes)
	// Format: TIME(20) METHOD(7) PATH(50) STATUS(3) TEXT(variable)
	timePadded := padColored(timeColored, 20)
	methodPadded := padColored(methodColored, 7)
	pathPadded := padColored(pathColored, 50)
	statusCodePadded := padColored(statusCodeColored, 3)

	// Format the request line with proper spacing between fields
	requestLine := fmt.Sprintf("%s %s %s %s %s",
		timePadded,
		methodPadded,
		pathPadded,
		statusCodePadded,
		statusTextColored)

	// Add to logs
	m.logs = append(m.logs, requestLine)
	if len(m.logs) > 100 {
		m.logs = m.logs[len(m.logs)-100:]
	}

	// Update viewport content (only the request logs, header is separate)
	content := ""
	for _, log := range m.logs {
		content += log + "\n"
	}
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

// Refactored tunnel command using Bubble Tea
func runTunnelWithBubbleTea(client *tunnel.TunnelClient, info *tunnel.TunnelInfo, accountDisplay string, serverURL string, localURL string) error {
	// Create model
	model := initialTunnelModel(client, info, accountDisplay, serverURL, localURL)

	// Channel to send request events to Bubble Tea
	requestChan := make(chan tunnel.RequestEvent, 100)

	// Set up request handler that sends to channel
	requestHandler := func(event tunnel.RequestEvent) {
		select {
		case requestChan <- event:
		default:
			// Channel full, drop event
		}
	}
	client.SetRequestHandler(requestHandler)

	// Create program without alt screen to ensure header is always visible
	p := tea.NewProgram(model)

	// Goroutine to forward request events to Bubble Tea
	go func() {
		for {
			select {
			case event := <-requestChan:
				p.Send(requestEventMsg(event))
			case <-model.ctx.Done():
				return
			}
		}
	}()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		// Send terminate message to update status to "terminated"
		p.Send(terminateMsg{})
		// Cancel context to stop tunnel
		model.cancel()
		// Give a moment to display terminated status
		time.Sleep(500 * time.Millisecond)
		p.Quit()
	}()

	// Run program
	if _, err := p.Run(); err != nil {
		return err
	}

	// Cleanup
	fmt.Println()
	fmt.Println(color.Yellow("Shutting down tunnel..."))
	if err := client.Close(); err != nil {
		return err
	}
	fmt.Println(color.Green("Tunnel closed successfully"))
	return nil
}
