package tunnel

import (
	"fmt"
	"html"
	"net/http"
	"strings"
)

func (ts *TunnelServer) writeErrorPage(w http.ResponseWriter, r *http.Request, tunnel *TunnelConnection, statusCode int, title, subtitle, details string) {
	path := strings.ToLower(r.URL.Path)
	if r.URL.RawQuery != "" {
		path = path + "?" + strings.ToLower(r.URL.RawQuery)
	}

	isAsset := strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".gif") ||
		strings.HasSuffix(path, ".svg") || strings.HasSuffix(path, ".woff") ||
		strings.HasSuffix(path, ".woff2") || strings.HasSuffix(path, ".ttf") ||
		strings.HasSuffix(path, ".eot") || strings.HasSuffix(path, ".ico") ||
		strings.HasSuffix(path, ".webmanifest") || strings.Contains(path, "/browser-sync/") ||
		strings.Contains(path, "/node_modules/") || strings.Contains(path, "/assets/")

	if isAsset {
		var contentType string
		if strings.HasSuffix(path, ".js") || strings.Contains(path, "/browser-sync/") {
			contentType = "application/javascript; charset=utf-8"
		} else if strings.HasSuffix(path, ".css") {
			contentType = "text/css; charset=utf-8"
		} else if strings.HasSuffix(path, ".svg") {
			contentType = "image/svg+xml"
		} else if strings.HasSuffix(path, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(path, ".gif") {
			contentType = "image/gif"
		} else if strings.HasSuffix(path, ".woff") {
			contentType = "font/woff"
		} else if strings.HasSuffix(path, ".woff2") {
			contentType = "font/woff2"
		} else if strings.HasSuffix(path, ".ttf") {
			contentType = "font/ttf"
		} else if strings.HasSuffix(path, ".ico") {
			contentType = "image/x-icon"
		} else if strings.HasSuffix(path, ".webmanifest") {
			contentType = "application/manifest+json"
		} else {
			contentType = "application/javascript; charset=utf-8"
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(statusCode)
		return
	}

	var publicURL, localURL string
	if tunnel != nil {
		tunnel.mu.RLock()
		localURL = tunnel.LocalURL
		publicURL = fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
		tunnel.mu.RUnlock()
	} else {
		publicURL = fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
		localURL = "N/A"
	}

	publicURL = html.EscapeString(publicURL)
	localURL = html.EscapeString(localURL)
	title = html.EscapeString(title)
	subtitle = html.EscapeString(subtitle)
	details = html.EscapeString(details)

	var iconSVG, iconColor, errorCode string
	switch statusCode {
	case http.StatusNotFound:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>`
		iconColor = "#f59e0b"
		errorCode = `<div class="error-code">ERR_UNIROUTE_404</div>`
	case http.StatusBadGateway:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>`
		iconColor = "#ef4444"
		errorCode = `<div class="error-code">ERR_UNIROUTE_502</div>`
	case http.StatusServiceUnavailable:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z"></path>`
		iconColor = "#f59e0b"
		errorCode = `<div class="error-code">ERR_UNIROUTE_503</div>`
	default:
		iconSVG = `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>`
		iconColor = "#ef4444"
		errorCode = `<div class="error-code">ERR_UNIROUTE_500</div>`
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self';")

	w.WriteHeader(statusCode)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" class="dark">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s - UniRoute Tunnel</title>
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
			background: linear-gradient(135deg, #0f172a 0%%, #1e3a8a 50%%, #312e81 100%%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
			color: #f1f5f9;
			padding: 20px;
		}
		.container {
			max-width: 600px;
			width: 100%%;
			background: rgba(15, 23, 42, 0.8);
			backdrop-filter: blur(12px);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 16px;
			padding: 40px;
			box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3), 0 10px 10px -5px rgba(0, 0, 0, 0.2);
		}
		.logo {
			display: flex;
			align-items: center;
			justify-content: center;
			margin-bottom: 32px;
		}
		.logo-icon {
			width: 48px;
			height: 48px;
			background: linear-gradient(135deg, #3b82f6 0%%, #6366f1 50%%, #a855f7 100%%);
			border-radius: 12px;
			display: flex;
			align-items: center;
			justify-content: center;
			box-shadow: 0 10px 15px -3px rgba(59, 130, 246, 0.3);
			margin-right: 12px;
		}
		.logo-icon span {
			color: white;
			font-weight: bold;
			font-size: 24px;
		}
		.logo-text {
			font-size: 28px;
			font-weight: bold;
			color: white;
		}
		.error-icon {
			width: 80px;
			height: 80px;
			margin: 0 auto 24px;
			background: rgba(239, 68, 68, 0.1);
			border-radius: 50%%;
			display: flex;
			align-items: center;
			justify-content: center;
			border: 2px solid rgba(239, 68, 68, 0.3);
		}
		.error-icon svg {
			width: 48px;
			height: 48px;
			color: %s;
		}
		.error-code {
			font-size: 14px;
			font-weight: 600;
			color: #94a3b8;
			text-align: center;
			margin-bottom: 8px;
			letter-spacing: 0.5px;
		}
		h1 {
			font-size: 28px;
			font-weight: bold;
			text-align: center;
			margin-bottom: 12px;
			color: white;
		}
		.subtitle {
			text-align: center;
			color: #cbd5e1;
			margin-bottom: 32px;
			font-size: 16px;
		}
		.info-box {
			background: rgba(30, 41, 59, 0.6);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 12px;
			padding: 24px;
			margin-bottom: 24px;
		}
		.info-row {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 12px 0;
			border-bottom: 1px solid rgba(148, 163, 184, 0.1);
		}
		.info-row:last-child {
			border-bottom: none;
		}
		.info-label {
			color: #94a3b8;
			font-size: 14px;
			font-weight: 500;
		}
		.info-value {
			color: white;
			font-size: 14px;
			font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
			word-break: break-all;
			text-align: right;
		}
		.details {
			background: rgba(30, 41, 59, 0.6);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 12px;
			padding: 16px;
			margin-top: 24px;
		}
		.details-text {
			color: #cbd5e1;
			font-size: 13px;
			font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
			word-break: break-all;
		}
		.footer {
			text-align: center;
			margin-top: 32px;
			color: #64748b;
			font-size: 13px;
		}
		.footer a {
			color: #60a5fa;
			text-decoration: none;
		}
		.footer a:hover {
			text-decoration: underline;
		}
		.status-indicator {
			text-align: center;
			margin-top: 24px;
			color: #94a3b8;
			font-size: 13px;
		}
		.status-indicator.checking {
			color: #60a5fa;
		}
		.status-indicator.online {
			color: #22c55e;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="logo">
			<div class="logo-icon">
				<span>U</span>
			</div>
			<span class="logo-text">UniRoute</span>
		</div>
		
		<div class="error-icon">
			<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
				%s
			</svg>
		</div>
		
		%s
		<h1>%s</h1>
		<p class="subtitle">%s</p>
		
		<div class="info-box">
			<div class="info-row">
				<span class="info-label">Public URL:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Local URL:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Status Code:</span>
				<span class="info-value">%d</span>
			</div>
		</div>
		
		<div class="details">
			<div class="details-text">%s</div>
		</div>
		
		<div class="status-indicator" id="statusIndicator">Checking connection...</div>
		
		<div class="footer">
			Powered by <a href="https://uniroute.co" target="_blank">UniRoute</a>
		</div>
	</div>
	<script>
		(function() {
			var checkInterval = 2000;
			var statusEl = document.getElementById('statusIndicator');
			var isChecking = false;
			var isReloading = false;
			
			function checkConnection() {
				if (isChecking || isReloading) return;
				isChecking = true;
				
				statusEl.textContent = 'Checking connection...';
				statusEl.className = 'status-indicator checking';
				
				var url = window.location.href.split('?')[0] + '?t=' + Date.now();
				fetch(url, {
					method: 'HEAD',
					cache: 'no-cache',
					headers: {
						'Cache-Control': 'no-cache',
						'Pragma': 'no-cache'
					}
				}).then(function(response) {
					isChecking = false;
					if (response.status === 200 || response.status === 304) {
						statusEl.textContent = 'Connection restored! Refreshing...';
						statusEl.className = 'status-indicator online';
						isReloading = true;
						setTimeout(function() {
							window.location.reload();
						}, 500);
					} else {
						statusEl.textContent = 'Still offline. Checking again in ' + (checkInterval / 1000) + ' seconds...';
						statusEl.className = 'status-indicator';
					}
				}).catch(function(error) {
					isChecking = false;
					statusEl.textContent = 'Still offline. Checking again in ' + (checkInterval / 1000) + ' seconds...';
					statusEl.className = 'status-indicator';
				});
			}
			
			checkConnection();
			setInterval(checkConnection, checkInterval);
		})();
	</script>
</body>
</html>`, title, iconColor, iconSVG, errorCode, title, subtitle, publicURL, localURL, statusCode, details)

	w.Write([]byte(html))
}

func (ts *TunnelServer) writeConnectionRefusedError(w http.ResponseWriter, r *http.Request, tunnel *TunnelConnection, errorMsg string) {
	tunnel.mu.RLock()
	localURL := tunnel.LocalURL
	publicURL := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	tunnel.mu.RUnlock()

	publicURL = html.EscapeString(publicURL)
	localURL = html.EscapeString(localURL)

	localPort := "unknown"
	if strings.HasPrefix(localURL, "http://") {
		parts := strings.Split(localURL[7:], ":")
		if len(parts) > 1 {
			localPort = strings.Split(parts[1], "/")[0]
		}
	} else if strings.Contains(localURL, ":") {
		parts := strings.Split(localURL, ":")
		if len(parts) > 1 {
			localPort = strings.Split(parts[1], "/")[0]
		}
	}
	localPort = html.EscapeString(localPort)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'none'; img-src 'self' data:; font-src 'self' data:;")

	w.WriteHeader(http.StatusBadGateway)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" class="dark">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Connection Refused - UniRoute Tunnel</title>
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
			background: linear-gradient(135deg, #0f172a 0%%, #1e3a8a 50%%, #312e81 100%%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
			color: #f1f5f9;
			padding: 20px;
		}
		.container {
			max-width: 600px;
			width: 100%%;
			background: rgba(15, 23, 42, 0.8);
			backdrop-filter: blur(12px);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 16px;
			padding: 40px;
			box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3), 0 10px 10px -5px rgba(0, 0, 0, 0.2);
		}
		.logo {
			display: flex;
			align-items: center;
			justify-content: center;
			margin-bottom: 32px;
		}
		.logo-icon {
			width: 48px;
			height: 48px;
			background: linear-gradient(135deg, #3b82f6 0%%, #6366f1 50%%, #a855f7 100%%);
			border-radius: 12px;
			display: flex;
			align-items: center;
			justify-content: center;
			box-shadow: 0 10px 15px -3px rgba(59, 130, 246, 0.3);
			margin-right: 12px;
		}
		.logo-icon span {
			color: white;
			font-weight: bold;
			font-size: 24px;
		}
		.logo-text {
			font-size: 28px;
			font-weight: bold;
			color: white;
		}
		.error-icon {
			width: 80px;
			height: 80px;
			margin: 0 auto 24px;
			background: rgba(239, 68, 68, 0.1);
			border-radius: 50%%;
			display: flex;
			align-items: center;
			justify-content: center;
			border: 2px solid rgba(239, 68, 68, 0.3);
		}
		.error-icon svg {
			width: 48px;
			height: 48px;
			color: #ef4444;
		}
		h1 {
			font-size: 28px;
			font-weight: bold;
			text-align: center;
			margin-bottom: 12px;
			color: white;
		}
		.subtitle {
			text-align: center;
			color: #cbd5e1;
			margin-bottom: 32px;
			font-size: 16px;
		}
		.info-box {
			background: rgba(30, 41, 59, 0.6);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 12px;
			padding: 24px;
			margin-bottom: 24px;
		}
		.info-row {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 12px 0;
			border-bottom: 1px solid rgba(148, 163, 184, 0.1);
		}
		.info-row:last-child {
			border-bottom: none;
		}
		.info-label {
			color: #94a3b8;
			font-size: 14px;
			font-weight: 500;
		}
		.info-value {
			color: white;
			font-size: 14px;
			font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
			word-break: break-all;
			text-align: right;
		}
		.steps {
			background: rgba(30, 41, 59, 0.6);
			border: 1px solid rgba(148, 163, 184, 0.2);
			border-radius: 12px;
			padding: 24px;
			margin-top: 24px;
		}
		.steps h2 {
			font-size: 18px;
			font-weight: 600;
			margin-bottom: 16px;
			color: white;
		}
		.step {
			display: flex;
			align-items: flex-start;
			margin-bottom: 16px;
		}
		.step:last-child {
			margin-bottom: 0;
		}
		.step-number {
			width: 28px;
			height: 28px;
			background: linear-gradient(135deg, #3b82f6 0%%, #6366f1 100%%);
			border-radius: 50%%;
			display: flex;
			align-items: center;
			justify-content: center;
			font-weight: bold;
			font-size: 14px;
			color: white;
			flex-shrink: 0;
			margin-right: 12px;
		}
		.step-content {
			flex: 1;
			color: #cbd5e1;
			font-size: 14px;
			line-height: 1.6;
		}
		.step-content code {
			background: rgba(15, 23, 42, 0.8);
			padding: 2px 6px;
			border-radius: 4px;
			font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
			font-size: 13px;
			color: #60a5fa;
		}
		.footer {
			text-align: center;
			margin-top: 32px;
			color: #64748b;
			font-size: 13px;
		}
		.footer a {
			color: #60a5fa;
			text-decoration: none;
		}
		.footer a:hover {
			text-decoration: underline;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="logo">
			<div class="logo-icon">
				<span>U</span>
			</div>
			<span class="logo-text">UniRoute</span>
		</div>
		
		<div class="error-icon">
			<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
			</svg>
		</div>
		
		<h1>Connection Refused</h1>
		<p class="subtitle">The tunnel is connected, but your local server is not running</p>
		
		<div class="info-box">
			<div class="info-row">
				<span class="info-label">Public URL:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Local URL:</span>
				<span class="info-value">%s</span>
			</div>
			<div class="info-row">
				<span class="info-label">Local Port:</span>
				<span class="info-value">%s</span>
			</div>
		</div>
		
		<div class="steps">
			<h2>How to fix this:</h2>
			<div class="step">
				<div class="step-number">1</div>
				<div class="step-content">Make sure your local server is running on <code>%s</code></div>
			</div>
			<div class="step">
				<div class="step-number">2</div>
				<div class="step-content">Verify the port number matches your application's configuration</div>
			</div>
			<div class="step">
				<div class="step-number">3</div>
				<div class="step-content">Check that your firewall isn't blocking the connection</div>
			</div>
			<div class="step">
				<div class="step-number">4</div>
				<div class="step-content">Once your server is running, refresh this page</div>
			</div>
		</div>
		
		<div class="footer">
			Powered by <a href="https://uniroute.co" target="_blank">UniRoute</a>
		</div>
	</div>
</body>
</html>`, publicURL, localURL, localPort, localURL)

	w.Write([]byte(html))
}
