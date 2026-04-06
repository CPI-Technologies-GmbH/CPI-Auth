package auth

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/api/middleware"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// oauthErrorPageData is the template payload for browser-facing OAuth error pages.
type oauthErrorPageData struct {
	Title         string
	Code          string
	Message       string
	SwitchURL     string // /login?... — re-enters the OAuth flow with a fresh login
	CancelURL     string // back to the relying party with an error parameter
	HasSwitch     bool
	HasCancel     bool
}

// oauthErrorPageTpl is the minimal HTML used when /oauth/authorize fails for a
// browser request. Inline so it cannot drift from the binary or fall through to
// a missing-template error.
var oauthErrorPageTpl = template.Must(template.New("oauth_error").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Title}}</title>
<style>
  :root { color-scheme: light dark; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", system-ui, sans-serif; margin: 0; min-height: 100vh; display: flex; align-items: center; justify-content: center; background: #0b0f1a; color: #e5e7eb; }
  .card { max-width: 480px; width: 100%; margin: 24px; padding: 32px; border-radius: 14px; background: #111827; border: 1px solid #1f2937; box-shadow: 0 12px 40px rgba(0,0,0,0.35); }
  h1 { margin: 0 0 8px; font-size: 22px; font-weight: 600; color: #f9fafb; }
  .code { display: inline-block; font-size: 11px; letter-spacing: 0.05em; text-transform: uppercase; padding: 3px 8px; border-radius: 999px; background: #7f1d1d; color: #fecaca; margin-bottom: 16px; }
  p { margin: 0 0 24px; line-height: 1.55; color: #cbd5e1; }
  .actions { display: flex; flex-direction: column; gap: 10px; }
  a.btn { display: block; text-align: center; padding: 11px 16px; border-radius: 10px; text-decoration: none; font-weight: 500; font-size: 14px; transition: opacity 120ms ease; }
  a.btn:hover { opacity: 0.88; }
  a.primary { background: #3b82f6; color: #fff; }
  a.secondary { background: transparent; color: #cbd5e1; border: 1px solid #374151; }
  @media (prefers-color-scheme: light) {
    body { background: #f3f4f6; color: #111827; }
    .card { background: #ffffff; border-color: #e5e7eb; }
    h1 { color: #111827; }
    p { color: #4b5563; }
    a.secondary { color: #374151; border-color: #d1d5db; }
  }
</style>
</head>
<body>
  <main class="card">
    <span class="code">{{.Code}}</span>
    <h1>{{.Title}}</h1>
    <p>{{.Message}}</p>
    <div class="actions">
      {{if .HasSwitch}}<a class="btn primary" href="{{.SwitchURL}}">Sign in with a different account</a>{{end}}
      {{if .HasCancel}}<a class="btn secondary" href="{{.CancelURL}}">Cancel and return to the application</a>{{end}}
    </div>
  </main>
</body>
</html>`))

// wantsHTML returns true when the client prefers an HTML response.
func wantsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

// writeOAuthAuthorizeError renders the appropriate error response for a failing
// /oauth/authorize GET. Browsers get an HTML page with recovery actions; API
// clients get the existing JSON envelope.
func (h *Handler) writeOAuthAuthorizeError(w http.ResponseWriter, r *http.Request, err error, originalQuery url.Values) {
	if !wantsHTML(r) {
		middleware.WriteError(w, err)
		return
	}

	appErr := models.GetAppError(err)
	if appErr == nil {
		appErr = models.ErrInternal
	}

	data := oauthErrorPageData{
		Code:    appErr.Code,
		Message: appErr.Message,
	}

	switch appErr.HTTPStatus {
	case http.StatusForbidden:
		data.Title = "You can't access this application"
	case http.StatusUnauthorized:
		data.Title = "Sign-in required"
	case http.StatusBadRequest:
		data.Title = "Invalid sign-in request"
	default:
		data.Title = "Sign-in error"
	}

	// "Switch account" link: log the current session out then re-enter the
	// OAuth flow with the same parameters, so the user can pick a different
	// account that does belong to the application's tenant.
	if originalQuery != nil && originalQuery.Get("client_id") != "" {
		returnTo := "/oauth/authorize?" + originalQuery.Encode()
		data.SwitchURL = "/api/v1/auth/logout?return_to=" + url.QueryEscape(returnTo)
		data.HasSwitch = true
	}

	// "Cancel" link: bounce back to the configured redirect_uri with an OAuth
	// error parameter, per RFC 6749 §4.1.2.1. Only emit if the redirect_uri
	// looks present — we don't validate it here because oauth.Authorize already
	// rejected the request before we got here.
	if originalQuery != nil {
		if rawRedirect := originalQuery.Get("redirect_uri"); rawRedirect != "" {
			cancelURL, perr := url.Parse(rawRedirect)
			if perr == nil && (cancelURL.Scheme == "https" || cancelURL.Scheme == "http") {
				q := cancelURL.Query()
				q.Set("error", oauthErrorCodeFor(appErr))
				if state := originalQuery.Get("state"); state != "" {
					q.Set("state", state)
				}
				cancelURL.RawQuery = q.Encode()
				data.CancelURL = cancelURL.String()
				data.HasCancel = true
			}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(appErr.HTTPStatus)
	_ = oauthErrorPageTpl.Execute(w, data)
}

// oauthErrorCodeFor maps an internal AppError to the OAuth 2.0 error code that
// should be returned to a relying party when the user cancels.
func oauthErrorCodeFor(appErr *models.AppError) string {
	switch appErr.Code {
	case "forbidden", "unauthorized":
		return "access_denied"
	case "invalid_client":
		return "unauthorized_client"
	case "invalid_scope":
		return "invalid_scope"
	case "bad_request", "validation_error":
		return "invalid_request"
	default:
		return "server_error"
	}
}
