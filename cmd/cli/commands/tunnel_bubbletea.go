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

// Quote with weight for weighted random selection
type weightedQuote struct {
	quote    string
	weight   int
	protocol string // "http", "tcp", "tls", "udp", or "" for all protocols
}

// Protocol-specific quotes
var httpQuotes = []weightedQuote{
	{quote: "🌐 Your app is now rockin' the web!", weight: 10, protocol: "http"},
	{quote: "💬 Your web app is now accessible worldwide!", weight: 10, protocol: "http"},
	{quote: "🚀 Your local server just went global on the web!", weight: 10, protocol: "http"},
	{quote: "⚡ HTTP requests are now faster than your WiFi!", weight: 10, protocol: "http"},
	{quote: "🎯 Your web app just got its passport to the internet!", weight: 10, protocol: "http"},
	{quote: "🔥 Your web app is so hot, it's breaking the internet!", weight: 10, protocol: "http"},
	{quote: "🎉 Your website is now live and accessible globally!", weight: 10, protocol: "http"},
	{quote: "🌟 From localhost to the world wide web!", weight: 10, protocol: "http"},
	{quote: "🚦 Green light! Your HTTP tunnel is live!", weight: 10, protocol: "http"},
	{quote: "🎪 Your web app is now on the world's biggest stage!", weight: 10, protocol: "http"},
	{quote: "🎸 Your web app is now rockin' the global network!", weight: 10, protocol: "http"},
	{quote: "🏆 Your website just won the tunnel lottery!", weight: 10, protocol: "http"},
	{quote: "🎨 Your web app is now a masterpiece on display!", weight: 10, protocol: "http"},
	{quote: "🎭 The web show must go on, and it's going global!", weight: 10, protocol: "http"},
	{quote: "🎯 Bullseye! Your HTTP tunnel is perfectly aimed!", weight: 10, protocol: "http"},
	{quote: "🎪 Step right up! Your web app is now on stage!", weight: 10, protocol: "http"},
	{quote: "🎊 Party time! Your HTTP tunnel is celebrating!", weight: 10, protocol: "http"},
	{quote: "🎁 Your web app just got the best gift: global HTTP access!", weight: 10, protocol: "http"},
	{quote: "🌐 From 127.0.0.1 to the world wide web!", weight: 10, protocol: "http"},
	{quote: "🚀 Your localhost just became a world-class web citizen!", weight: 10, protocol: "http"},
	// Donation link - appears most frequently (much higher weight)
	{quote: "💝 Love UniRoute? Support us: https://polar.sh/uniroute/donate", weight: 200, protocol: "http"},
}

var tcpQuotes = []weightedQuote{
	{quote: "🔌 Your TCP connection is now globally accessible!", weight: 10, protocol: "tcp"},
	{quote: "🌐 Your TCP service just went worldwide!", weight: 10, protocol: "tcp"},
	{quote: "⚡ TCP packets are flying faster than ever!", weight: 10, protocol: "tcp"},
	{quote: "🎯 Your TCP tunnel is perfectly connected!", weight: 10, protocol: "tcp"},
	{quote: "🚀 Your TCP server just got its global passport!", weight: 10, protocol: "tcp"},
	{quote: "💪 TCP connections are now breaking down barriers!", weight: 10, protocol: "tcp"},
	{quote: "🎉 Your TCP service is now live and accessible!", weight: 10, protocol: "tcp"},
	{quote: "🌟 From localhost to global TCP access!", weight: 10, protocol: "tcp"},
	{quote: "🚦 Green light! Your TCP tunnel is live!", weight: 10, protocol: "tcp"},
	{quote: "🎪 Your TCP service is now on the world stage!", weight: 10, protocol: "tcp"},
	{quote: "🔌 TCP connections are now flowing globally!", weight: 10, protocol: "tcp"},
	{quote: "🏆 Your TCP tunnel just won the connection lottery!", weight: 10, protocol: "tcp"},
	{quote: "🎨 Your TCP service is now a global masterpiece!", weight: 10, protocol: "tcp"},
	{quote: "🎭 The TCP show must go on, and it's going global!", weight: 10, protocol: "tcp"},
	{quote: "🎯 Bullseye! Your TCP tunnel is perfectly aimed!", weight: 10, protocol: "tcp"},
	{quote: "🎪 Step right up! Your TCP service is now on stage!", weight: 10, protocol: "tcp"},
	{quote: "🎊 Party time! Your TCP tunnel is celebrating!", weight: 10, protocol: "tcp"},
	{quote: "🎁 Your TCP service just got global access!", weight: 10, protocol: "tcp"},
	{quote: "🔌 From localhost to global TCP connectivity!", weight: 10, protocol: "tcp"},
	{quote: "🚀 Your TCP server just became a global citizen!", weight: 10, protocol: "tcp"},
	// Donation link - appears most frequently (much higher weight)
	{quote: "💝 Love UniRoute? Support us: https://polar.sh/uniroute/donate", weight: 200, protocol: "tcp"},
}

var tlsQuotes = []weightedQuote{
	{quote: "🔒 Your secure TLS connection is now globally accessible!", weight: 10, protocol: "tls"},
	{quote: "🌐 Your TLS service just went worldwide with encryption!", weight: 10, protocol: "tls"},
	{quote: "⚡ Secure TLS packets are flying faster than ever!", weight: 10, protocol: "tls"},
	{quote: "🎯 Your encrypted TLS tunnel is perfectly connected!", weight: 10, protocol: "tls"},
	{quote: "🚀 Your TLS server just got its secure global passport!", weight: 10, protocol: "tls"},
	{quote: "💪 Secure TLS connections are now breaking down barriers!", weight: 10, protocol: "tls"},
	{quote: "🎉 Your encrypted TLS service is now live!", weight: 10, protocol: "tls"},
	{quote: "🌟 From localhost to global secure TLS access!", weight: 10, protocol: "tls"},
	{quote: "🚦 Green light! Your secure TLS tunnel is live!", weight: 10, protocol: "tls"},
	{quote: "🎪 Your TLS service is now on the secure world stage!", weight: 10, protocol: "tls"},
	{quote: "🔒 TLS connections are now flowing securely globally!", weight: 10, protocol: "tls"},
	{quote: "🏆 Your encrypted tunnel just won the security lottery!", weight: 10, protocol: "tls"},
	{quote: "🎨 Your TLS service is now a secure global masterpiece!", weight: 10, protocol: "tls"},
	{quote: "🎭 The secure TLS show must go on, and it's going global!", weight: 10, protocol: "tls"},
	{quote: "🎯 Bullseye! Your encrypted TLS tunnel is perfectly aimed!", weight: 10, protocol: "tls"},
	{quote: "🎪 Step right up! Your secure TLS service is now on stage!", weight: 10, protocol: "tls"},
	{quote: "🎊 Party time! Your encrypted TLS tunnel is celebrating!", weight: 10, protocol: "tls"},
	{quote: "🎁 Your TLS service just got secure global access!", weight: 10, protocol: "tls"},
	{quote: "🔒 From localhost to global encrypted TLS connectivity!", weight: 10, protocol: "tls"},
	{quote: "🚀 Your secure TLS server just became a global citizen!", weight: 10, protocol: "tls"},
	// Donation link - appears most frequently (much higher weight)
	{quote: "💝 Love UniRoute? Support us: https://polar.sh/uniroute/donate", weight: 200, protocol: "tls"},
}

var udpQuotes = []weightedQuote{
	{quote: "📡 Your UDP packets are now flying globally!", weight: 10, protocol: "udp"},
	{quote: "🌐 Your UDP service just went worldwide!", weight: 10, protocol: "udp"},
	{quote: "⚡ UDP datagrams are racing faster than ever!", weight: 10, protocol: "udp"},
	{quote: "🎯 Your UDP tunnel is perfectly connected!", weight: 10, protocol: "udp"},
	{quote: "🚀 Your UDP server just got its global passport!", weight: 10, protocol: "udp"},
	{quote: "💪 UDP packets are now breaking down barriers!", weight: 10, protocol: "udp"},
	{quote: "🎉 Your UDP service is now live and accessible!", weight: 10, protocol: "udp"},
	{quote: "🌟 From localhost to global UDP access!", weight: 10, protocol: "udp"},
	{quote: "🚦 Green light! Your UDP tunnel is live!", weight: 10, protocol: "udp"},
	{quote: "🎪 Your UDP service is now on the world stage!", weight: 10, protocol: "udp"},
	{quote: "📡 UDP datagrams are now flowing globally!", weight: 10, protocol: "udp"},
	{quote: "🏆 Your UDP tunnel just won the packet lottery!", weight: 10, protocol: "udp"},
	{quote: "🎨 Your UDP service is now a global masterpiece!", weight: 10, protocol: "udp"},
	{quote: "🎭 The UDP show must go on, and it's going global!", weight: 10, protocol: "udp"},
	{quote: "🎯 Bullseye! Your UDP tunnel is perfectly aimed!", weight: 10, protocol: "udp"},
	{quote: "🎪 Step right up! Your UDP service is now on stage!", weight: 10, protocol: "udp"},
	{quote: "🎊 Party time! Your UDP tunnel is celebrating!", weight: 10, protocol: "udp"},
	{quote: "🎁 Your UDP service just got global access!", weight: 10, protocol: "udp"},
	{quote: "📡 From localhost to global UDP connectivity!", weight: 10, protocol: "udp"},
	{quote: "🚀 Your UDP server just became a global citizen!", weight: 10, protocol: "udp"},
	// Donation link - appears most frequently (much higher weight)
	{quote: "💝 Love UniRoute? Support us: https://polar.sh/uniroute/donate", weight: 200, protocol: "udp"},
}

// Custom domain quotes (when user uses their own domain)
var domainQuotes = []weightedQuote{
	{quote: "🌐 Your custom domain is now live and accessible worldwide!", weight: 10, protocol: ""},
	{quote: "🎯 Your own domain is now rockin' the web!", weight: 10, protocol: ""},
	{quote: "🚀 Your custom domain just went global!", weight: 10, protocol: ""},
	{quote: "💎 Your domain is now a premium global citizen!", weight: 10, protocol: ""},
	{quote: "👑 Your custom domain is now the king of the internet!", weight: 10, protocol: ""},
	{quote: "🌟 Your own domain is now shining worldwide!", weight: 10, protocol: ""},
	{quote: "🎉 Congratulations! Your custom domain is live!", weight: 10, protocol: ""},
	{quote: "🏆 Your domain just won the premium tunnel lottery!", weight: 10, protocol: ""},
	{quote: "🎪 Your custom domain is now on the world's biggest stage!", weight: 10, protocol: ""},
	{quote: "🎨 Your domain is now a masterpiece on display!", weight: 10, protocol: ""},
	{quote: "🎭 The custom domain show must go on, and it's going global!", weight: 10, protocol: ""},
	{quote: "🎯 Bullseye! Your custom domain is perfectly connected!", weight: 10, protocol: ""},
	{quote: "🎪 Step right up! Your domain is now on stage!", weight: 10, protocol: ""},
	{quote: "🎊 Party time! Your custom domain is celebrating!", weight: 10, protocol: ""},
	{quote: "🎁 Your domain just got the best gift: global access!", weight: 10, protocol: ""},
	{quote: "🌐 From localhost to your own domain - you've made it!", weight: 10, protocol: ""},
	{quote: "🚀 Your custom domain just became a world-class citizen!", weight: 10, protocol: ""},
	{quote: "💎 Premium domain, premium experience!", weight: 10, protocol: ""},
	{quote: "👑 Your domain is now ruling the internet!", weight: 10, protocol: ""},
	{quote: "🌟 Your custom domain is now a global superstar!", weight: 10, protocol: ""},
	// Donation link - appears most frequently (much higher weight)
	{quote: "💝 Love UniRoute? Support us: https://polar.sh/uniroute/donate", weight: 200, protocol: ""},
}

// Generic quotes that work for all protocols
var genericQuotes = []weightedQuote{
	{quote: "💬 Your app is now accessible worldwide!", weight: 10, protocol: ""},
	{quote: "🚀 Your local server just went global!", weight: 10, protocol: ""},
	{quote: "🌍 Breaking down firewalls, one request at a time!", weight: 10, protocol: ""},
	{quote: "⚡ Your code is now faster than your WiFi!", weight: 10, protocol: ""},
	{quote: "🎯 Tunnel vision? More like tunnel success!", weight: 10, protocol: ""},
	{quote: "🔥 Your app is so hot, it's breaking the internet!", weight: 10, protocol: ""},
	{quote: "💪 Local development, global domination!", weight: 10, protocol: ""},
	{quote: "🎉 Congratulations! Your app just got a passport!", weight: 10, protocol: ""},
	{quote: "🌟 From localhost to everywhere!", weight: 10, protocol: ""},
	{quote: "🎪 Welcome to the greatest show on the internet!", weight: 10, protocol: ""},
	{quote: "🏆 You've just won the tunnel lottery!", weight: 10, protocol: ""},
	{quote: "🎪 The circus is in town, and your app is the star!", weight: 10, protocol: ""},
	{quote: "🚀 Your localhost just became a world-class citizen!", weight: 10, protocol: ""},
	{quote: "🌐 From 127.0.0.1 to infinity and beyond!", weight: 10, protocol: ""},
	{quote: "⚡ Faster than a speeding bullet, more powerful than localhost!", weight: 10, protocol: ""},
	{quote: "🏰 Your local server just became a global empire!", weight: 10, protocol: ""},
	{quote: "🎨 Picasso would be jealous of this masterpiece!", weight: 10, protocol: ""},
	{quote: "🎬 Lights, camera, action! Your app is live!", weight: 10, protocol: ""},
	{quote: "🎤 Your app just got its own world tour!", weight: 10, protocol: ""},
	{quote: "🎮 Game over for localhost limitations!", weight: 10, protocol: ""},
	{quote: "🎲 You rolled a natural 20! Tunnel connected!", weight: 10, protocol: ""},
	{quote: "🎪 The magic show begins - watch your app disappear from localhost!", weight: 10, protocol: ""},
	{quote: "🚁 Your app just got airlifted to the cloud!", weight: 10, protocol: ""},
	{quote: "🎪 Welcome to the tunnel of wonders!", weight: 10, protocol: ""},
	{quote: "🎯 Mission accomplished: Your app is now public!", weight: 10, protocol: ""},
	{quote: "🎪 Your app just joined the global network party!", weight: 10, protocol: ""},
	{quote: "🎨 From localhost to local-hero in one tunnel!", weight: 10, protocol: ""},
	{quote: "🎪 The tunnel is open - your app is free!", weight: 10, protocol: ""},
	{quote: "🚀 Houston, we have a tunnel connection!", weight: 10, protocol: ""},
	{quote: "🎪 Your app just got its VIP pass to the internet!", weight: 10, protocol: ""},
	{quote: "🎯 Direct hit! Your tunnel is perfectly connected!", weight: 10, protocol: ""},
	{quote: "🎪 The tunnel is alive! Your app is breathing the global air!", weight: 10, protocol: ""},
	// Donation link - appears most frequently (much higher weight = most likely)
	{quote: "💝 Love UniRoute? Support us: https://polar.sh/uniroute/donate", weight: 200, protocol: ""},
}

// isCustomDomain checks if the PublicURL contains a custom domain (not .localhost)
func isCustomDomain(publicURL string) bool {
	if publicURL == "" {
		return false
	}
	// Check if URL contains .localhost (default subdomain) or localhost:port
	// If it doesn't, it's likely a custom domain
	return !strings.Contains(publicURL, ".localhost") && !strings.Contains(publicURL, "localhost:")
}

// getQuotesForProtocol returns quotes appropriate for the given protocol and domain type
func getQuotesForProtocol(protocol string, hasCustomDomain bool) []weightedQuote {
	var quotes []weightedQuote

	// If custom domain is used, add domain-specific quotes first (higher priority)
	if hasCustomDomain {
		quotes = append(quotes, domainQuotes...)
	}

	// Add protocol-specific quotes
	switch protocol {
	case "http":
		quotes = append(quotes, httpQuotes...)
	case "tcp":
		quotes = append(quotes, tcpQuotes...)
	case "tls":
		quotes = append(quotes, tlsQuotes...)
	case "udp":
		quotes = append(quotes, udpQuotes...)
	}

	// Always add generic quotes (they work for all protocols)
	quotes = append(quotes, genericQuotes...)

	return quotes
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
	viewport            viewport.Model
	logs                []string
	logsMu              sync.Mutex // Mutex for thread-safe log access
	viewportNeedsUpdate bool       // Flag to track if viewport needs updating
	userHasScrolled     bool       // Track if user has manually scrolled (don't auto-scroll if true)

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
	lastStatus     string    // Track last connection status to detect changes
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

	// Get protocol from client to show protocol-specific quotes
	protocol := ""
	if client != nil {
		protocol = client.GetProtocol()
	}

	// Check if custom domain is being used
	hasCustomDomain := false
	if info != nil && info.PublicURL != "" {
		hasCustomDomain = isCustomDomain(info.PublicURL)
	}

	// Get protocol and domain-specific quotes
	quotes := getQuotesForProtocol(protocol, hasCustomDomain)

	// Pick a random quote using weighted selection
	rand.Seed(time.Now().UnixNano())

	// Calculate total weight
	totalWeight := 0
	for _, q := range quotes {
		totalWeight += q.weight
	}

	// Pick random number in range [0, totalWeight)
	randomNum := rand.Intn(totalWeight)

	// Find which quote this corresponds to
	var randomQuote string
	currentWeight := 0
	for _, q := range quotes {
		currentWeight += q.weight
		if randomNum < currentWeight {
			randomQuote = q.quote
			break
		}
	}

	// Fallback (should never happen, but safety check)
	if randomQuote == "" {
		if len(quotes) > 0 {
			randomQuote = quotes[0].quote
		} else {
			randomQuote = "🚀 Your tunnel is live!"
		}
	}

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
		// Don't update if terminated - check first to avoid any processing
		if m.terminated {
			return m, nil
		}

		// Check if tunnel was disconnected from dashboard - don't process status updates
		if m.client.ShouldExit() {
			// Should exit - don't process this status update, terminate message will be sent by monitor
			return m, nil
		}

		status := string(msg)
		// Only update if internet is online, otherwise internetStatusMsg handler will handle it
		if m.internetOnline {
			// Check if client is actually reconnecting (but only trust it if status also says reconnecting)
			// This prevents false positives where isReconnecting might be temporarily true
			isReconnecting := m.client.IsReconnecting()
			isConnected := m.client.IsConnected()

			// Trust the status from GetConnectionStatus() - it's the source of truth
			// Only show reconnecting if BOTH status says reconnecting AND client confirms it
			// This prevents showing reconnecting when connection is actually fine
			if status == "online" {
				// Connection is online - clear any reconnect flags and show connected
				// Only update if we're actually connected (prevent false positives)
				if isConnected {
					if !m.reconnectTime.IsZero() {
						m.reconnectTime = time.Time{}
					}
					m.connectionStatus = color.Green("Tunnel Connected Successfully!")
					m.sessionStatus = fmt.Sprintf("%s %s", color.Green("●"), color.Green("online"))
				}
				// If status says "online" but client says not connected, don't update (might be stale)
			} else if status == "reconnecting" && isReconnecting && !isConnected {
				// Only show reconnecting if:
				// 1. Status says reconnecting
				// 2. Client confirms it's reconnecting
				// 3. Client confirms it's NOT connected
				// This prevents false positives
				m.connectionStatus = color.Yellow("Reconnecting...")
				m.sessionStatus = fmt.Sprintf("%s %s", color.Yellow("●"), color.Yellow("reconnecting"))
			} else if status == "offline" {
				// Offline status
				m.connectionStatus = color.Red("No Internet - Connection Lost")
				m.sessionStatus = fmt.Sprintf("%s %s", color.Red("●"), color.Red("offline"))
			} else {
				// Unknown status or status doesn't match client state - don't update
				// This prevents showing incorrect status
			}
		}
		// Continue polling for latency and stats (only if not terminated)
		// Don't send status updates here - they're handled by the status change handler
		if !m.terminated {
			return m, m.updateStatus() // Continue polling for latency/stats only
		}
		return m, nil
	case sessionStatusMsg:
		// Don't update status if tunnel is terminated
		if m.terminated {
			return m, nil
		}

		// Check if tunnel was disconnected from dashboard - don't process status updates
		if m.client.ShouldExit() {
			// Should exit - don't process this status update, terminate message will be sent by monitor
			return m, nil
		}

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
			// Update lastStatus to trigger screen clear in View()
			m.lastStatus = "" // Reset to force screen clear
			// Cancel context to stop tunnel
			m.cancel()
			// Return updated model with a command to quit after showing status
			// This ensures the terminated status is rendered before exit
			return m, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
				return tea.Quit()
			})
		}
		// Already terminated, just quit
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

		// Update model state (logs array) - this is the correct pattern
		m.addRequestToLogs(event)

		// Rebuild viewport content from updated logs
		// This ensures the viewport always shows all accumulated requests
		m.updateViewportContent()

		// Return updated model - Bubble Tea will re-render
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
		// Calculate actual header height: padding (1) + title (1) + blank (1) + connection status (1) + session (1) + account (1) + blank (1) + version (1) + blank (1) + region (1) + latency (1) + blank (1) + connections header (1) + connections data (1) + blank (1) + public url label (1) + public url (1) + blank (1) + forwarding label (1) + forwarding (1) + blank (1) + quote (1) + blank (1) + ctrl+c (1) + blank (1) + http requests header (2) = ~28 lines
		headerHeight := 28

		// Calculate available height for viewport (must leave room for header)
		// Quote and Ctrl+C are now part of the header, so no separate footer
		availableHeight := msg.Height - headerHeight

		// Ensure viewport height never exceeds available space
		// This guarantees the header will always be visible
		// Use most of the available space so users can see more requests
		var viewportHeight int
		if availableHeight < 1 {
			// Terminal is too small - show at least 1 line for viewport
			viewportHeight = 1
		} else {
			// Use all available height (viewport will scroll if content exceeds it)
			// This ensures maximum visibility of requests
			viewportHeight = availableHeight
		}

		// Update viewport dimensions properly
		// Use SetSize to ensure viewport recalculates its internal state
		m.viewport.Width = msg.Width
		m.viewport.Height = viewportHeight

		// Update viewport with the window size message so it can handle resize internally
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)

		// Rebuild viewport content after size change to ensure it's correct
		m.updateViewportContent()

		return m, cmd
	}

	// Handle viewport scrolling
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	// Detect if user manually scrolled away from top
	// If user scrolls down (not at top), mark that they've scrolled so we don't auto-scroll
	if m.viewport.YOffset > 0 {
		m.userHasScrolled = true
	} else {
		// User is at top, allow auto-scrolling
		m.userHasScrolled = false
	}

	return m, cmd
}

// View function for Bubble Tea
func (m *tunnelModel) View() string {
	// Fixed header section
	header := strings.Builder{}

	// Detect status change and clear screen to prevent duplicate lines
	// This ensures old content is removed when status changes (e.g., online -> terminated)
	// Without alt screen, Bubble Tea doesn't automatically clear, so we must do it manually
	currentStatus := m.connectionStatus
	statusChanged := m.lastStatus != "" && m.lastStatus != currentStatus
	
	// Clear screen if status changed or if terminated (and lastStatus was reset)
	if statusChanged || (m.terminated && m.lastStatus == "") {
		// Status changed or terminated - clear screen to remove old content
		header.WriteString("\033[2J\033[H") // Clear screen and move cursor to top-left
	}
	
	// Update last status after clearing
	if statusChanged {
		m.lastStatus = currentStatus
	} else if m.lastStatus == "" {
		// First render or after termination reset - initialize lastStatus
		m.lastStatus = currentStatus
	}

	// Add some top padding to ensure header is visible (prevents cut-off)
	header.WriteString("\n")

	// Add title at the top
	header.WriteString(color.Cyan("Starting UniRoute Tunnel..."))
	header.WriteString("\n\n")

	// Connection Status (with label to match other fields)
	header.WriteString(fmt.Sprintf("Connection Status             %s\n", m.connectionStatus))

	// Session Status
	header.WriteString(fmt.Sprintf("Session Status                %s\n", m.sessionStatus))

	// Account
	header.WriteString(fmt.Sprintf("Account                       %s\n", color.Gray(m.account)))
	header.WriteString("\n")

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

	// Quote and Ctrl+C message (before HTTP Requests)
	header.WriteString(color.Yellow(m.quote))
	header.WriteString("\n\nPress Ctrl+C to stop\n\n")

	// HTTP Requests header
	header.WriteString("HTTP Requests\n")
	header.WriteString("-------------\n")

	// View() should be pure - just render, don't modify state
	// The viewport content is already updated in updateViewportContent()

	// Combine header (with quote/Ctrl+C) with scrolling viewport
	// HTTP Requests section is at the bottom and can scroll properly
	// View() should be pure - don't modify state here, just render
	return header.String() + m.viewport.View()
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
	// Poll more frequently (1 second) for latency and stats updates
	// Connection status is handled by the status change handler to avoid duplicates
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		// Don't poll if tunnel is terminated
		// This prevents duplicate status messages after termination
		if m.terminated {
			return nil
		}

		// Only update latency and stats - connection status is handled by status change handler
		// This prevents duplicate status messages
		latency := m.client.GetLatency()

		// Fetch stats if connected
		status := m.client.GetConnectionStatus()
		var stats *tunnel.ConnectionStats
		if status == "online" && m.info != nil {
			statsChan := make(chan *tunnel.ConnectionStats, 1)
			go func() {
				s, _ := m.client.GetConnectionStats(m.serverURL, m.info.ID)
				statsChan <- s
			}()
			select {
			case stats = <-statsChan:
			case <-time.After(1 * time.Second): // Reduced timeout for faster updates
			}
		}

		// Only return latency and stats updates - NOT connection status
		// Connection status is handled by the status change handler
		return tea.Batch(
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

// addRequestToLogs adds a request to the logs array (model state update)
// This follows Bubble Tea best practices: update state in Update, render in View
func (m *tunnelModel) addRequestToLogs(event tunnel.RequestEvent) {
	// Format request log entry

	// Format: STATUS METHOD PATH LATENCY
	// Example: 200 OK GET /auth/tunnels/stats 11ms
	// Every request is unique and should be added, even if same path/method
	// NO DEDUPLICATION - each request gets its own entry

	// Format status code and text together
	statusText := event.StatusText
	if statusText == "" {
		// Generate status text from code if not provided
		switch event.StatusCode {
		case 200:
			statusText = "OK"
		case 201:
			statusText = "Created"
		case 204:
			statusText = "No Content"
		case 400:
			statusText = "Bad Request"
		case 401:
			statusText = "Unauthorized"
		case 403:
			statusText = "Forbidden"
		case 404:
			statusText = "Not Found"
		case 405:
			statusText = "Method Not Allowed"
		case 500:
			statusText = "Internal Server Error"
		case 502:
			statusText = "Bad Gateway"
		default:
			statusText = fmt.Sprintf("Status %d", event.StatusCode)
		}
	} else {
		// Extract status text if full status line provided (e.g., "200 OK" -> "OK")
		if strings.Contains(statusText, " ") {
			parts := strings.SplitN(statusText, " ", 2)
			if len(parts) > 1 {
				statusText = parts[1]
			}
		}
	}

	statusFull := fmt.Sprintf("%d %s", event.StatusCode, statusText)

	// Color code based on status (text color, not background)
	var statusColored string
	if event.StatusCode == 200 {
		statusColored = color.Green(statusFull) // Green text for 200 OK
	} else if event.StatusCode == 201 {
		statusColored = color.Purple(statusFull) // Purple text for 201 Created
	} else if event.StatusCode >= 200 && event.StatusCode < 300 {
		statusColored = color.Green(statusFull) // Green text for 2xx
	} else if event.StatusCode >= 300 && event.StatusCode < 400 {
		statusColored = color.Yellow(statusFull) // Yellow text for 3xx
	} else if event.StatusCode >= 400 {
		statusColored = color.Red(statusFull) // Red text for 4xx/5xx
	} else {
		statusColored = color.Gray(statusFull)
	}

	// Format method (white text)
	methodColored := color.White(event.Method)

	// Format path (white text) - ensure root path is shown as "/"
	path := event.Path
	if path == "" {
		path = "/"
	}
	pathColored := color.White(path)

	// Format latency in milliseconds
	latencyStr := fmt.Sprintf("%dms", event.LatencyMs)
	latencyColored := color.White(latencyStr)

	// Format the request line: STATUS METHOD PATH LATENCY
	// Each request is unique - even same path/method gets a new entry
	// We prepend to ensure newest requests appear at the top
	// NOTE: Even if two requests have identical method/path/status/latency,
	// they are still TWO SEPARATE ENTRIES in the logs array
	// CRITICAL: Build the line properly - each colored segment already has Reset codes
	// Use very generous spacing between elements for better readability
	requestLine := fmt.Sprintf("%s          %s          %s          %s",
		statusColored,
		methodColored,
		pathColored,
		latencyColored)

	// IMPORTANT: This requestLine will be prepended to m.logs
	// Even if it's identical to an existing entry, it's a NEW entry
	// The logs array grows with EVERY request, no deduplication

	// CRITICAL: ALWAYS prepend new request at the top (streaming behavior - newest at top)
	// No deduplication - every request gets its own entry, even if identical
	// This function is called for EVERY request event, no filtering
	m.logsMu.Lock()

	// IMPORTANT: Prepend new request to the FRONT of the slice
	// This ensures newest requests appear at the top, pushing old ones down
	// Even if path/method/status are identical, this is a NEW entry
	m.logs = append([]string{requestLine}, m.logs...)

	// Limit to latest 100 requests (allows scrolling to see older requests)
	// When a new request comes in, it's added at the top and the oldest is removed from the bottom
	// This keeps the header fixed at the top and new requests appear below it
	// Users can scroll down to see older requests
	const maxRequests = 100
	if len(m.logs) > maxRequests {
		// Keep only the first 100 (newest) requests, remove oldest from the end
		m.logs = m.logs[:maxRequests]
	}

	m.logsMu.Unlock()

	// State update complete - logs array is updated
	// Viewport content will be rebuilt in updateViewportContent()
}

// updateViewportContent rebuilds viewport content from the logs array
// This implements a fixed-size window showing the latest 10 requests
// Following best practices for streaming request displays
func (m *tunnelModel) updateViewportContent() {
	m.logsMu.Lock()
	logsCopy := make([]string, len(m.logs))
	copy(logsCopy, m.logs)
	m.logsMu.Unlock()

	// Show all accumulated requests (up to 100)
	// Users can scroll through all requests to see older ones
	// Newest requests are at the top, oldest at the bottom
	displayLogs := logsCopy

	// Build content from all requests
	// Newest requests are first in the array, oldest at the end
	var contentBuilder strings.Builder
	for _, log := range displayLogs {
		// Write the log entry (already includes ANSI codes)
		contentBuilder.WriteString(log)
		// Add newline after each entry
		contentBuilder.WriteByte('\n')
	}
	content := contentBuilder.String()

	// Ensure viewport has minimum dimensions (only if not set)
	// DO NOT grow the viewport height here - it's set in WindowSizeMsg handler
	// Growing it here would push the header up
	if m.viewport.Height == 0 {
		m.viewport.Height = 10 // Small default, will be set properly by WindowSizeMsg
	}
	if m.viewport.Width == 0 {
		m.viewport.Width = 80
	}

	// Check if user is at top before updating content (for smart auto-scroll)
	wasAtTop := m.viewport.YOffset == 0

	// Update viewport with all requests
	// Newest requests are at the top of the content, oldest at bottom
	// The viewport has a FIXED height and scrolls internally
	// This keeps the header fixed at the top while requests scroll within the viewport
	m.viewport.SetContent(content)

	// Smart auto-scroll: only scroll to top if user was already at top
	// This allows users to scroll down to see older requests without being snapped back
	if wasAtTop {
		// User was at top, so auto-scroll to show newest requests
		m.viewport.GotoTop()
		m.userHasScrolled = false // Reset since we auto-scrolled
	}

	// Mark that viewport has been updated
	m.viewportNeedsUpdate = false
}

// Refactored tunnel command using Bubble Tea
func runTunnelWithBubbleTea(client *tunnel.TunnelClient, info *tunnel.TunnelInfo, accountDisplay string, serverURL string, localURL string) error {
	// Create model
	model := initialTunnelModel(client, info, accountDisplay, serverURL, localURL)

	// Channel to send request events to Bubble Tea
	// Large buffer to prevent dropping events - we want ALL requests logged
	requestChan := make(chan tunnel.RequestEvent, 1000)

	// Channel to send connection status changes to Bubble Tea for real-time updates
	statusChangeChan := make(chan string, 10)

	// Set up request handler that sends to channel
	// ALWAYS send every request event - no filtering or dropping
	requestHandler := func(event tunnel.RequestEvent) {
		// Non-blocking send - if channel is full, wait briefly then try again
		// This ensures we don't lose any requests
		select {
		case requestChan <- event:
			// Successfully sent
		default:
			// Channel full - wait a bit and try again (non-blocking)
			// This is rare but ensures we don't lose requests
			go func() {
				time.Sleep(10 * time.Millisecond)
				select {
				case requestChan <- event:
					// Successfully sent on retry
				default:
					// Still full after wait - drop event (should be extremely rare with 1000 buffer)
				}
			}()
		}
	}
	client.SetRequestHandler(requestHandler)

	// Set up connection status change handler for real-time updates
	// This provides immediate status updates when connection state changes
	// Use a debounce mechanism to prevent rapid status changes from causing duplicates
	lastStatus := ""
	lastStatusTime := time.Time{}
	statusChangeHandler := func(status string) {
		now := time.Now()

		// Only send if status actually changed AND enough time has passed (debounce)
		// This prevents rapid toggling between statuses
		if status != lastStatus {
			// If status changed, check if enough time has passed since last change
			if lastStatusTime.IsZero() || time.Since(lastStatusTime) > 500*time.Millisecond {
				lastStatus = status
				lastStatusTime = now
				// Non-blocking send to status change channel
				select {
				case statusChangeChan <- status:
					// Successfully sent
				default:
					// Channel full - drop (should be rare with 10 buffer)
				}
			} else {
				// Status changed too quickly - debounce by scheduling a delayed check
				go func() {
					time.Sleep(500 * time.Millisecond)
					// Re-check status after delay
					currentStatus := client.GetConnectionStatus()
					if currentStatus == status && currentStatus != lastStatus {
						lastStatus = currentStatus
						lastStatusTime = time.Now()
						select {
						case statusChangeChan <- currentStatus:
						default:
						}
					}
				}()
			}
		}
	}
	client.SetConnectionStatusChangeHandler(statusChangeHandler)

	// Clear screen and move cursor to top before starting
	// This ensures a clean start without duplicate content
	fmt.Print("\033[2J\033[H")        // Clear screen and move to top-left
	time.Sleep(50 * time.Millisecond) // Small delay to ensure screen is cleared

	// Create program without alt screen to ensure header is always visible
	// We'll handle screen clearing manually in the View function
	p := tea.NewProgram(model)

	// Goroutine to forward request events and status changes to Bubble Tea
	go func() {
		for {
			select {
			case event := <-requestChan:
				// Send request event to Bubble Tea
				p.Send(requestEventMsg(event))
			case status := <-statusChangeChan:
				// Send connection status change immediately for real-time updates
				// Only send if model is not terminated
				// Double-check the actual connection status before updating to prevent false positives
				if !model.terminated {
					// Check if tunnel was disconnected from dashboard - should exit
					// This check must happen FIRST before processing any status updates
					if model.client.ShouldExit() {
						// Don't process status updates if we should exit - let monitor handle termination
						// This prevents showing "online" status right before "terminated"
						continue
					}

					// Verify the status is still accurate before sending
					actualIsConnected := model.client.IsConnected()
					actualIsReconnecting := model.client.IsReconnecting()

					// Double-check ShouldExit() again after getting connection state
					// This prevents race conditions where ShouldExit() becomes true between checks
					if model.client.ShouldExit() {
						continue
					}

					// Only update if the status matches the actual connection state
					// This prevents showing "reconnecting" when connection is actually stable
					if status == "online" && actualIsConnected && !actualIsReconnecting {
						p.Send(connectionStatusMsg(status))
						p.Send(sessionStatusMsg(status))
					} else if status == "reconnecting" && actualIsReconnecting && !actualIsConnected {
						p.Send(connectionStatusMsg(status))
						p.Send(sessionStatusMsg(status))
					} else if status == "offline" && !actualIsConnected && !actualIsReconnecting {
						p.Send(connectionStatusMsg(status))
						p.Send(sessionStatusMsg(status))
					}
					// If status doesn't match actual state, ignore it (prevents false positives)
				}
			case <-model.ctx.Done():
				return
			}
		}
	}()

	// Monitor for tunnel disconnect from dashboard (should exit)
	// This is the single source of truth for detecting dashboard disconnects
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond) // Check more frequently for faster response
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if model.client.ShouldExit() && !model.terminated {
					// Tunnel was disconnected from dashboard - exit gracefully
					// Only send terminate message once - let Update handler show status and quit
					p.Send(terminateMsg{})
					model.cancel()
					// Don't call p.Quit() here - let the terminateMsg handler return tea.Quit
					// This ensures the terminated status is shown before exit
					return
				}
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
		// Cancel context to stop tunnel and goroutines
		model.cancel()
		// Give a brief moment to display terminated status, then quit
		time.Sleep(200 * time.Millisecond)
		p.Quit()
	}()

	// Run program
	if _, err := p.Run(); err != nil {
		return err
	}

	// Cleanup - close tunnel with timeout to prevent hanging
	fmt.Println()
	fmt.Println(color.Yellow("Shutting down tunnel..."))

	// Close tunnel with timeout to prevent hanging
	closeDone := make(chan error, 1)
	go func() {
		closeDone <- client.Close()
	}()

	select {
	case err := <-closeDone:
		if err != nil {
			// Log error but don't fail - connection might already be closed
			fmt.Println(color.Yellow(fmt.Sprintf("Tunnel close warning: %v", err)))
		} else {
			fmt.Println(color.Green("Tunnel closed successfully"))
		}
	case <-time.After(2 * time.Second):
		// Timeout - connection might be stuck, but we'll exit anyway
		fmt.Println(color.Yellow("Tunnel shutdown timed out, exiting..."))
	}

	return nil
}
