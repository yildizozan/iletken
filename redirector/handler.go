package redirector

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"iletken/config"

	"github.com/valyala/fasthttp"
)

// RedirectHandler HTTP redirector
type RedirectHandler struct {
	rules  map[string]config.RedirectRule
	logger *slog.Logger
}

// NewRedirectHandler creates a new RedirectHandler
func NewRedirectHandler(redirects []config.RedirectRule, logger *slog.Logger) *RedirectHandler {
	rules := make(map[string]config.RedirectRule)
	
	for _, rule := range redirects {
		// Normalize host name (convert to lowercase)
		host := strings.ToLower(strings.TrimSpace(rule.From))
		rules[host] = rule
	}
	
	return &RedirectHandler{
		rules:  rules,
		logger: logger,
	}
}

// Handle processes HTTP requests
func (h *RedirectHandler) Handle(ctx *fasthttp.RequestCtx) {
	// Health check endpoint
	if string(ctx.Path()) == "/health" {
		h.handleHealthCheck(ctx)
		return
	}
	
	// Get Host header
	host := strings.ToLower(string(ctx.Host()))
	
	// Remove port number (if present)
	if colonPos := strings.Index(host, ":"); colonPos != -1 {
		host = host[:colonPos]
	}

	h.logger.Debug("Request received",
		slog.String("host", host),
		slog.String("path", string(ctx.Path())),
		slog.String("method", string(ctx.Method())),
		slog.String("user_agent", string(ctx.UserAgent())),
		slog.String("remote_addr", ctx.RemoteAddr().String()),
	)
	
	// Check if this is the default index page request (localhost or no specific host)
	if string(ctx.Path()) == "/" && (host == "localhost" || host == "127.0.0.1" || host == "") {
		h.handleIndexPage(ctx)
		return
	}
	
	// Find redirect rule
	rule, found := h.rules[host]
	if !found {
		h.logger.Warn("Redirect rule not found",
			slog.String("host", host),
			slog.String("remote_addr", ctx.RemoteAddr().String()),
		)
		
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("text/html; charset=utf-8")
		fmt.Fprintf(ctx, `<!DOCTYPE html>
<html lang="tr">
<head>
    <title>404 - Sayfa Bulunamadı</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: #333;
            margin: 0;
            padding: 20px;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.3);
            text-align: center;
            max-width: 500px;
            width: 100%%;
        }
        .logo {
            font-size: 2.5em;
            font-weight: bold;
            color: #667eea;
            margin-bottom: 20px;
        }
        .error {
            color: #d32f2f;
            font-size: 3em;
            margin-bottom: 20px;
            font-weight: bold;
        }
        .message {
            font-size: 1.2em;
            margin-bottom: 20px;
            line-height: 1.5;
        }
        .host-info {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
            font-family: 'Courier New', monospace;
            color: #495057;
        }
        .footer {
            color: #666;
            font-size: 0.9em;
            margin-top: 30px;
        }
        .back-link {
            color: #667eea;
            text-decoration: none;
            font-weight: 500;
        }
        .back-link:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">İletken</div>
        <div class="error">404</div>
        <div class="message">Sayfa Bulunamadı</div>
        <div class="host-info">
            Host: <strong>%s</strong><br>
            Bu host için yönlendirme kuralı bulunamadı.
        </div>
        <div class="footer">
            <p><a href="/" class="back-link">← Ana Sayfaya Dön</a></p>
            <p><em>İletken - HTTP Redirector</em></p>
        </div>
    </div>
</body>
</html>`, host)
		return
	}
	
	// Perform redirect
	h.logger.Info("Performing redirect",
		slog.String("from", host),
		slog.String("to", rule.To),
		slog.Int("status_code", 302),
		slog.String("remote_addr", ctx.RemoteAddr().String()),
	)
	
	ctx.SetStatusCode(302)
	ctx.Response.Header.Set("Location", rule.To)
	ctx.Response.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Response.Header.Set("Pragma", "no-cache")
	ctx.Response.Header.Set("Expires", "0")
}

// handleHealthCheck handles health check requests
func (h *RedirectHandler) handleHealthCheck(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	
	stats := h.GetStats()
	
	// Simple JSON response without external dependency
	fmt.Fprintf(ctx, `{"status":"healthy","service":"iletken","version":"1.0.0","stats":{"total_rules":%d,"configured_hosts":%d}}`,
		stats["total_rules"], len(stats["configured_hosts"].([]string)))
	
	h.logger.Debug("Health check requested",
		slog.String("remote_addr", ctx.RemoteAddr().String()),
	)
}

// handleIndexPage serves the default index page
func (h *RedirectHandler) handleIndexPage(ctx *fasthttp.RequestCtx) {
	// Try to read index.html file
	indexPath := "./index.html"
	htmlContent, err := os.ReadFile(indexPath)
	if err != nil {
		// Fallback to simple HTML if file not found
		h.logger.Warn("Index file not found, serving fallback page",
			slog.String("path", indexPath),
			slog.String("error", err.Error()),
		)
		
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("text/html; charset=utf-8")
		ctx.WriteString(`<!DOCTYPE html>
<html>
<head>
    <title>İletken - HTTP Redirector</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
        .logo { font-size: 2em; color: #667eea; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="logo">İletken</div>
    <h2>HTTP Redirector Service</h2>
    <p>Service is running. <a href="/health">Health Check</a></p>
</body>
</html>`)
		return
	}
	
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("text/html; charset=utf-8")
	ctx.Write(htmlContent)
	
	h.logger.Debug("Index page served",
		slog.String("remote_addr", ctx.RemoteAddr().String()),
		slog.String("file_path", indexPath),
	)
}

// GetStats returns redirector statistics
func (h *RedirectHandler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_rules":    len(h.rules),
		"configured_hosts": func() []string {
			hosts := make([]string, 0, len(h.rules))
			for host := range h.rules {
				hosts = append(hosts, host)
			}
			return hosts
		}(),
	}
}
