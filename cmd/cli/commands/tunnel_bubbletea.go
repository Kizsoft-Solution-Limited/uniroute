package commands

import (
	"context"
	"fmt"
	"io"
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

type weightedQuote struct {
	quote    string
	weight   int
	protocol string
}

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
	{quote: "💝 Love UniRoute? Support us: https://buy.polar.sh/polar_cl_h5uF0bHhXXF6EO8Mx3tVP1Ry1G4wNWn4V8phg3rStVs", weight: 200, protocol: "http"},
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
	{quote: "💝 Love UniRoute? Support us: https://buy.polar.sh/polar_cl_h5uF0bHhXXF6EO8Mx3tVP1Ry1G4wNWn4V8phg3rStVs", weight: 200, protocol: "tcp"},
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
	{quote: "💝 Love UniRoute? Support us: https://buy.polar.sh/polar_cl_h5uF0bHhXXF6EO8Mx3tVP1Ry1G4wNWn4V8phg3rStVs", weight: 200, protocol: "tls"},
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
	{quote: "💝 Love UniRoute? Support us: https://buy.polar.sh/polar_cl_h5uF0bHhXXF6EO8Mx3tVP1Ry1G4wNWn4V8phg3rStVs", weight: 200, protocol: "udp"},
}

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
	{quote: "💝 Love UniRoute? Support us: https://buy.polar.sh/polar_cl_h5uF0bHhXXF6EO8Mx3tVP1Ry1G4wNWn4V8phg3rStVs", weight: 200, protocol: ""},
}

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
	{quote: "💝 Love UniRoute? Support us: https://buy.polar.sh/polar_cl_h5uF0bHhXXF6EO8Mx3tVP1Ry1G4wNWn4V8phg3rStVs", weight: 200, protocol: ""},
}

func isCustomDomain(publicURL string) bool {
	if publicURL == "" {
		return false
	}
	return !strings.Contains(publicURL, ".localhost") && !strings.Contains(publicURL, "localhost:")
}

func getQuotesForProtocol(protocol string, hasCustomDomain bool) []weightedQuote {
	var quotes []weightedQuote

	if hasCustomDomain {
		quotes = append(quotes, domainQuotes...)
	}

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

	quotes = append(quotes, genericQuotes...)

	return quotes
}

type tunnelModel struct {
	connectionStatus string
	sessionStatus    string
	account          string
	version          string
	region           string
	latency          string
	publicURL        string
	forwarding       string
	connections      string
	quote            string
	viewport         viewport.Model
	logs             []string
	logsMu           sync.Mutex
	viewportNeedsUpdate bool
	userHasScrolled  bool
	client           *tunnel.TunnelClient
	info             *tunnel.TunnelInfo
	serverURL        string
	localURL         string
	updateInfo       *versioncheck.VersionInfo
	updateInfoMu     sync.Mutex
	internetOnline   bool
	wasOffline       bool
	reconnectTime    time.Time
	terminated       bool
	lastStatus       string
	ctx              context.Context
	cancel           context.CancelFunc
	headerStyle      lipgloss.Style
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

type internetStatusMsg bool
type latencyUpdateMsg int64
type connectionStatusMsg string
type sessionStatusMsg string
type requestEventMsg tunnel.RequestEvent
type statsUpdateMsg *tunnel.ConnectionStats
type versionUpdateMsg *versioncheck.VersionInfo
type updateProgressMsg string
type terminateMsg struct{}

// regionDisplay returns the region string for the CLI (where the tunnel server runs). Like ngrok "Region: Europe (eu)".
func regionDisplay(info *tunnel.TunnelInfo) string {
	if info != nil && info.Region != "" {
		return info.Region
	}
	return "Local"
}

func initialTunnelModel(client *tunnel.TunnelClient, info *tunnel.TunnelInfo, accountDisplay string, serverURL string, localURL string) *tunnelModel {
	ctx, cancel := context.WithCancel(context.Background())
	vp := viewport.New(80, 20)
	vp.SetContent("")

	currentVersion := GetVersion()
	protocol := ""
	if client != nil {
		protocol = client.GetProtocol()
	}

	hasCustomDomain := false
	if info != nil && info.PublicURL != "" {
		hasCustomDomain = isCustomDomain(info.PublicURL)
	}

	quotes := getQuotesForProtocol(protocol, hasCustomDomain)

	rand.Seed(time.Now().UnixNano())
	totalWeight := 0
	for _, q := range quotes {
		totalWeight += q.weight
	}

	randomNum := rand.Intn(totalWeight)
	var randomQuote string
	currentWeight := 0
	for _, q := range quotes {
		currentWeight += q.weight
		if randomNum < currentWeight {
			randomQuote = q.quote
			break
		}
	}

	if randomQuote == "" {
		if len(quotes) > 0 {
			randomQuote = quotes[0].quote
		} else {
			randomQuote = "🚀 Your tunnel is live!"
		}
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	labelStyle := lipgloss.NewStyle().Width(25).Foreground(lipgloss.Color("244"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	statusGreen := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	statusYellow := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	statusRed := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	timeStyle := lipgloss.NewStyle().Width(20).Foreground(lipgloss.Color("244"))
	methodStyle := lipgloss.NewStyle().Width(7).Foreground(lipgloss.Color("39"))
	pathStyle := lipgloss.NewStyle().Width(50).Foreground(lipgloss.Color("255"))
	statusCodeStyle := lipgloss.NewStyle().Width(3)

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
		region:           regionDisplay(info),
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

func (m *tunnelModel) Init() tea.Cmd {
	return tea.Batch(
		m.checkInternet(),
		m.updateStatus(),
		m.checkForUpdates(),
	)
}

func (m *tunnelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if !m.terminated {
				m.cancel()
				return m, func() tea.Msg {
					return terminateMsg{}
				}
			}
			return m, tea.Quit
		case "ctrl+u":
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
		if m.terminated {
			return m, nil
		}
		m.internetOnline = bool(msg)
		if !m.internetOnline {
			m.connectionStatus = color.Red("No Internet - Connection Lost")
			m.sessionStatus = fmt.Sprintf("%s %s", color.Red("●"), color.Red("offline"))
			m.wasOffline = true
			m.reconnectTime = time.Time{}
		} else {
			m.wasOffline = false
			m.reconnectTime = time.Time{}
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
		return m, m.checkInternet()
	case connectionStatusMsg:
		if m.terminated {
			return m, nil
		}

		if m.client.ShouldExit() {
			return m, nil
		}

		status := string(msg)
		m.lastStatus = m.connectionStatus // track for clear-on-change in View
		if m.internetOnline {
			isReconnecting := m.client.IsReconnecting()
			isConnected := m.client.IsConnected()
			if status == "online" {
				if isConnected {
					if !m.reconnectTime.IsZero() {
						m.reconnectTime = time.Time{}
					}
					m.connectionStatus = color.Green("Tunnel Connected Successfully!")
					m.sessionStatus = fmt.Sprintf("%s %s", color.Green("●"), color.Green("online"))
				}
			} else if status == "reconnecting" && isReconnecting && !isConnected {
				m.connectionStatus = color.Yellow("Reconnecting...")
				m.sessionStatus = fmt.Sprintf("%s %s", color.Yellow("●"), color.Yellow("reconnecting"))
			} else if status == "offline" {
				m.connectionStatus = color.Red("No Internet - Connection Lost")
				m.sessionStatus = fmt.Sprintf("%s %s", color.Red("●"), color.Red("offline"))
			}
		}
		if !m.terminated {
			return m, m.updateStatus()
		}
		return m, nil
	case sessionStatusMsg:
		if m.terminated {
			return m, nil
		}
		if m.client.ShouldExit() {
			return m, nil
		}
		status := string(msg)
		if m.internetOnline {
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
		if !m.terminated {
			m.terminated = true
			m.sessionStatus = fmt.Sprintf("%s %s", color.Red("●"), color.Red("terminated"))
			m.connectionStatus = color.Red("Tunnel Terminated")
			m.lastStatus = ""
			m.cancel()
			return m, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
				return tea.Quit()
			})
		}
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
		m.addRequestToLogs(event)
		m.updateViewportContent()
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
		headerHeight := 28
		availableHeight := msg.Height - headerHeight

		var viewportHeight int
		if availableHeight < 1 {
			viewportHeight = 1
		} else {
			viewportHeight = availableHeight
		}

		m.viewport.Width = msg.Width
		m.viewport.Height = viewportHeight

		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		m.updateViewportContent()

		return m, cmd
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	if m.viewport.YOffset > 0 {
		m.userHasScrolled = true
	} else {
		m.userHasScrolled = false
	}

	return m, cmd
}

const tunnelHeaderLines = 28

// terminatedView uses the same header layout as the live view so Connection/Session status stay in the same place.
func (m *tunnelModel) terminatedView() string {
	var b strings.Builder
	b.WriteString("\033[2J\033[H")
	b.WriteString("\n")
	b.WriteString(color.Cyan("Starting UniRoute Tunnel..."))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Connection Status             %s\n", m.connectionStatus))
	b.WriteString(fmt.Sprintf("Session Status                %s\n", m.sessionStatus))
	b.WriteString(fmt.Sprintf("Account                       %s\n", color.Gray(m.account)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Version                       %s\n", m.version))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Region                        %s\n", color.Gray(m.region)))
	b.WriteString(fmt.Sprintf("Latency                       %s\n", color.Gray(m.latency)))
	b.WriteString("\n")
	b.WriteString("Connections                   ttl     opn     rt1     rt5     p50     p90\n")
	b.WriteString(fmt.Sprintf("                              %s\n\n", m.connections))
	b.WriteString("🌍 Public URL:\n")
	b.WriteString(fmt.Sprintf("   %s\n", color.Cyan(m.publicURL)))
	b.WriteString("\n")
	b.WriteString("🔗 Forwarding:\n")
	b.WriteString(fmt.Sprintf("   %s\n", m.forwarding))
	b.WriteString("\n")
	b.WriteString(color.Yellow(m.quote))
	b.WriteString("\n\n")
	b.WriteString(color.Gray("Tunnel stopped. Goodbye."))
	b.WriteString("\n")
	return b.String()
}

func (m *tunnelModel) View() string {
	if m.terminated {
		return m.terminatedView()
	}

	header := strings.Builder{}
	header.WriteString("\n")
	header.WriteString(color.Cyan("Starting UniRoute Tunnel..."))
	header.WriteString("\n\n")
	header.WriteString(fmt.Sprintf("Connection Status             %s\n", m.connectionStatus))
	header.WriteString(fmt.Sprintf("Session Status                %s\n", m.sessionStatus))
	header.WriteString(fmt.Sprintf("Account                       %s\n", color.Gray(m.account)))
	header.WriteString("\n")
	header.WriteString(fmt.Sprintf("Version                       %s\n", m.version))
	header.WriteString("\n")
	header.WriteString(fmt.Sprintf("Region                        %s\n", color.Gray(m.region)))
	header.WriteString(fmt.Sprintf("Latency                       %s\n", color.Gray(m.latency)))
	header.WriteString("\n")
	header.WriteString("Connections                   ttl     opn     rt1     rt5     p50     p90\n")
	header.WriteString(fmt.Sprintf("                              %s\n\n", m.connections))
	header.WriteString("🌍 Public URL:\n")
	header.WriteString(fmt.Sprintf("   %s\n", color.Cyan(m.publicURL)))
	header.WriteString("\n")
	header.WriteString("🔗 Forwarding:\n")
	header.WriteString(fmt.Sprintf("   %s\n", m.forwarding))
	header.WriteString("\n")
	header.WriteString(color.Yellow(m.quote))
	header.WriteString("\n\nPress Ctrl+C to stop\n\n")
	header.WriteString("HTTP Requests\n")
	header.WriteString("-------------\n")

	s := header.String()
	lines := strings.Count(s, "\n")
	if len(s) > 0 && !strings.HasSuffix(s, "\n") {
		lines++
	}
	if lines < tunnelHeaderLines {
		s += strings.Repeat("\n", tunnelHeaderLines-lines)
	}
	return s + m.viewport.View()
}

func (m *tunnelModel) checkInternet() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		client := &http.Client{Timeout: 2 * time.Second}
		_, err := client.Head("http://clients3.google.com/generate_204")
		return internetStatusMsg(err == nil)
	})
}

func (m *tunnelModel) updateStatus() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		if m.terminated {
			return nil
		}

		latency := m.client.GetLatency()
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
			case <-time.After(1 * time.Second):
			}
		}

		return tea.Batch(
			func() tea.Msg { return latencyUpdateMsg(latency) },
			func() tea.Msg { return statsUpdateMsg(stats) },
			func() tea.Msg { return connectionStatusMsg(status) },
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
		return updateProgressMsg("Downloading update...")
	}
}

func (m *tunnelModel) runUpgrade() tea.Cmd {
	return func() tea.Msg {
		upgradeCmd := exec.Command(os.Args[0], "upgrade", "--yes")
		upgradeCmd.Stdout = io.Discard
		upgradeCmd.Stderr = io.Discard
		upgradeCmd.Stdin = nil

		errChan := make(chan error, 1)
		go func() {
			errChan <- upgradeCmd.Run()
		}()

		select {
		case err := <-errChan:
			if err != nil {
				return updateProgressMsg("Update failed - run 'uniroute upgrade' manually")
			}
			return updateProgressMsg("✓ Update ready. Press Ctrl+C to stop the tunnel, then start it again to use the new version.")
		case <-time.After(30 * time.Second):
			if upgradeCmd.Process != nil {
				upgradeCmd.Process.Kill()
			}
			return updateProgressMsg("Update timed out - run 'uniroute upgrade' manually")
		}
	}
}

func (m *tunnelModel) addRequestToLogs(event tunnel.RequestEvent) {
	statusText := event.StatusText
	if statusText == "" {
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
		if strings.Contains(statusText, " ") {
			parts := strings.SplitN(statusText, " ", 2)
			if len(parts) > 1 {
				statusText = parts[1]
			}
		}
	}

	statusFull := fmt.Sprintf("%d %s", event.StatusCode, statusText)
	var statusColored string
	if event.StatusCode == 200 {
		statusColored = color.Green(statusFull)
	} else if event.StatusCode == 201 {
		statusColored = color.Purple(statusFull)
	} else if event.StatusCode >= 200 && event.StatusCode < 300 {
		statusColored = color.Green(statusFull)
	} else if event.StatusCode >= 300 && event.StatusCode < 400 {
		statusColored = color.Yellow(statusFull)
	} else if event.StatusCode >= 400 {
		statusColored = color.Red(statusFull)
	} else {
		statusColored = color.Gray(statusFull)
	}
	methodColored := color.White(event.Method)
	path := event.Path
	if path == "" {
		path = "/"
	}
	pathColored := color.White(path)
	latencyStr := fmt.Sprintf("%dms", event.LatencyMs)
	latencyColored := color.White(latencyStr)

	requestLine := fmt.Sprintf("%s          %s          %s          %s",
		statusColored,
		methodColored,
		pathColored,
		latencyColored)

	m.logsMu.Lock()
	m.logs = append([]string{requestLine}, m.logs...)
	const maxRequests = 100
	if len(m.logs) > maxRequests {
		m.logs = m.logs[:maxRequests]
	}

	m.logsMu.Unlock()
}

func (m *tunnelModel) updateViewportContent() {
	m.logsMu.Lock()
	logsCopy := make([]string, len(m.logs))
	copy(logsCopy, m.logs)
	m.logsMu.Unlock()
	displayLogs := logsCopy

	var contentBuilder strings.Builder
	for _, log := range displayLogs {
		contentBuilder.WriteString(log)
		contentBuilder.WriteByte('\n')
	}
	content := contentBuilder.String()

	if m.viewport.Height == 0 {
		m.viewport.Height = 10
	}
	if m.viewport.Width == 0 {
		m.viewport.Width = 80
	}

	wasAtTop := m.viewport.YOffset == 0
	m.viewport.SetContent(content)

	if wasAtTop {
		m.viewport.GotoTop()
		m.userHasScrolled = false
	}

	m.viewportNeedsUpdate = false
}

func runTunnelWithBubbleTea(client *tunnel.TunnelClient, info *tunnel.TunnelInfo, accountDisplay string, serverURL string, localURL string) error {
	model := initialTunnelModel(client, info, accountDisplay, serverURL, localURL)
	requestChan := make(chan tunnel.RequestEvent, 1000)
	statusChangeChan := make(chan string, 10)

	requestHandler := func(event tunnel.RequestEvent) {
		select {
		case requestChan <- event:
		default:
			go func() {
				time.Sleep(10 * time.Millisecond)
				select {
				case requestChan <- event:
				default:
				}
			}()
		}
	}
	client.SetRequestHandler(requestHandler)

	lastStatus := ""
	lastStatusTime := time.Time{}
	statusChangeHandler := func(status string) {
		now := time.Now()
		if status != lastStatus {
			if lastStatusTime.IsZero() || time.Since(lastStatusTime) > 500*time.Millisecond {
				lastStatus = status
				lastStatusTime = now
				select {
				case statusChangeChan <- status:
				default:
				}
			} else {
				go func() {
					time.Sleep(500 * time.Millisecond)
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

	fmt.Print("\033[2J\033[H")
	time.Sleep(50 * time.Millisecond)

	p := tea.NewProgram(model)

	go func() {
		for {
			select {
			case event := <-requestChan:
				p.Send(requestEventMsg(event))
			case status := <-statusChangeChan:
				if !model.terminated {
					if model.client.ShouldExit() {
						continue
					}

					actualIsConnected := model.client.IsConnected()
					actualIsReconnecting := model.client.IsReconnecting()

					if model.client.ShouldExit() {
						continue
					}

					if status == "online" && actualIsConnected {
						p.Send(connectionStatusMsg(status))
						p.Send(sessionStatusMsg(status))
					} else if status == "reconnecting" && actualIsReconnecting && !actualIsConnected {
						p.Send(connectionStatusMsg(status))
						p.Send(sessionStatusMsg(status))
					} else if status == "offline" && !actualIsConnected && !actualIsReconnecting {
						p.Send(connectionStatusMsg(status))
						p.Send(sessionStatusMsg(status))
					}
				}
			case <-model.ctx.Done():
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if model.client.ShouldExit() && !model.terminated {
					p.Send(terminateMsg{})
					model.cancel()
					return
				}
			case <-model.ctx.Done():
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		p.Send(terminateMsg{})
		model.cancel()
		time.Sleep(200 * time.Millisecond)
		p.Quit()
	}()

	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	// After Bubble Tea exits, re-print terminated view so it appears in main buffer (status at top)
	if tm, ok := finalModel.(*tunnelModel); ok && tm.terminated {
		fmt.Print(tm.terminatedView())
	}
	fmt.Println()
	fmt.Println(color.Yellow("Shutting down tunnel..."))
	closeDone := make(chan error, 1)
	go func() {
		closeDone <- client.Close()
	}()
	select {
	case err := <-closeDone:
		if err != nil {
			fmt.Println(color.Yellow(fmt.Sprintf("Tunnel close warning: %v", err)))
		} else {
			fmt.Println(color.Green("Tunnel closed successfully"))
		}
	case <-time.After(2 * time.Second):
		fmt.Println(color.Yellow("Tunnel shutdown timed out, exiting..."))
	}

	return nil
}
