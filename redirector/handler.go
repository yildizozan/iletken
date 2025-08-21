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
	
	// Default index page for root path
	if string(ctx.Path()) == "/" {
		h.handleIndexPage(ctx)
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
<html>
<head>
    <title>404 - Page Not Found</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
        .error { color: #d32f2f; }
    </style>
</head>
<body>
    <h1 class="error">404 - Page Not Found</h1>
    <p>No redirect rule found for host (%s).</p>
    <p><em>İletken - HTTP Redirector</em></p>
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
