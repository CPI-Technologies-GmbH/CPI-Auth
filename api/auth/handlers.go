package auth

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/api/middleware"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/actions"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/events"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/federation"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/flows"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/oauth"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/policy"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/sessions"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/tokens"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/users"
)

// Handler holds dependencies for auth API handlers.
type Handler struct {
	oauthSvc       *oauth.Service
	userSvc        *users.Service
	tokenSvc       *tokens.Service
	sessionSvc     *sessions.Service
	mfaSvc         *flows.MFAService
	webauthnSvc    *federation.WebAuthnService
	eventSvc       *events.Service
	rbacSvc        *policy.RBACService
	actionPipeline *actions.Pipeline
	deviceCodeRepo models.DeviceCodeRepository
	tenantRepo     models.TenantRepository
	cfg            *config.Config
	logger         *zap.Logger
}

// NewHandler creates new auth handlers.
func NewHandler(
	oauthSvc *oauth.Service,
	userSvc *users.Service,
	tokenSvc *tokens.Service,
	sessionSvc *sessions.Service,
	mfaSvc *flows.MFAService,
	webauthnSvc *federation.WebAuthnService,
	eventSvc *events.Service,
	rbacSvc *policy.RBACService,
	actionPipeline *actions.Pipeline,
	deviceCodeRepo models.DeviceCodeRepository,
	tenantRepo models.TenantRepository,
	cfg *config.Config,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		oauthSvc:       oauthSvc,
		userSvc:        userSvc,
		tokenSvc:       tokenSvc,
		sessionSvc:     sessionSvc,
		mfaSvc:         mfaSvc,
		webauthnSvc:    webauthnSvc,
		eventSvc:       eventSvc,
		rbacSvc:        rbacSvc,
		actionPipeline: actionPipeline,
		deviceCodeRepo: deviceCodeRepo,
		tenantRepo:     tenantRepo,
		cfg:            cfg,
		logger:         logger,
	}
}

// RegisterRoutes registers all auth-related routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/oauth/authorize", h.Authorize)
	r.Get("/oauth/authorize", h.AuthorizeGet)
	r.Post("/oauth/token", h.Token)
	r.Post("/oauth/revoke", h.Revoke)
	r.Get("/oauth/userinfo", h.Userinfo)
	r.Get("/.well-known/openid-configuration", h.Discovery)
	r.Get("/.well-known/jwks.json", h.JWKS)

	r.Post("/passwordless/start", h.PasswordlessStart)
	r.Post("/passwordless/verify", h.PasswordlessVerify)

	r.Post("/mfa/challenge", h.MFAChallenge)
	r.Post("/mfa/verify", h.MFAVerify)

	r.Post("/webauthn/register/begin", h.WebAuthnRegisterBegin)
	r.Post("/webauthn/register/finish", h.WebAuthnRegisterFinish)
	r.Post("/webauthn/login/begin", h.WebAuthnLoginBegin)
	r.Post("/webauthn/login/finish", h.WebAuthnLoginFinish)

	r.Get("/saml/metadata", h.SAMLMetadata)
	r.Post("/saml/acs", h.SAMLACS)
	r.Get("/saml/sso", h.SAMLSSO)

	r.Post("/api/v1/auth/login", h.Login)
	r.Post("/api/v1/auth/register", h.Register)
	r.Get("/api/v1/auth/logout", h.Logout)
	r.Post("/api/v1/auth/logout", h.Logout)

	// Device Authorization Flow (RFC 8628) — public endpoints
	r.Post("/oauth/device/code", h.DeviceCode)
	r.Post("/oauth/device/token", h.DeviceToken)
}

// --- OAuth Endpoints ---

func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	var req oauth.AuthorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, err)
		return
	}

	// For POST /authorize, expect authenticated user
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		// Try to authenticate from session/form
		middleware.WriteError(w, nil)
		return
	}

	resp, err := h.oauthSvc.Authorize(r.Context(), userID, req)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) AuthorizeGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := oauth.AuthorizeRequest{
		ClientID:            r.URL.Query().Get("client_id"),
		RedirectURI:         r.URL.Query().Get("redirect_uri"),
		ResponseType:        r.URL.Query().Get("response_type"),
		Scope:               r.URL.Query().Get("scope"),
		State:               r.URL.Query().Get("state"),
		CodeChallenge:       r.URL.Query().Get("code_challenge"),
		CodeChallengeMethod: r.URL.Query().Get("code_challenge_method"),
		Nonce:               r.URL.Query().Get("nonce"),
	}

	userID := middleware.GetUserID(r.Context())

	// Check session cookie if no JWT auth
	if userID == uuid.Nil {
		if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
			sessionID, parseErr := uuid.Parse(cookie.Value)
			if parseErr == nil && h.sessionSvc != nil {
				sess, sessErr := h.sessionSvc.Validate(r.Context(), sessionID)
				if sessErr == nil && sess != nil {
					userID = sess.UserID
					// Also set tenant from session if not already resolved
					if middleware.GetTenantID(r.Context()) == uuid.Nil {
						ctx = context.WithValue(ctx, middleware.ContextKeyTenantID, sess.TenantID)
						r = r.WithContext(ctx)
					}
				}
			}
		}
	}

	if userID == uuid.Nil {
		// No valid session — redirect to the login UI with the OAuth params
		// preserved. If we know the tenant slug (because the request came in
		// via /t/{slug}/oauth/authorize OR we can resolve it from client_id),
		// route through the tenant-prefixed login URL so the user's URL bar
		// stays inside the tenant scope and the login page renders with the
		// correct branding without needing client_id juggling.
		loginPath := "/login"
		if slug := middleware.GetTenantSlug(r.Context()); slug != "" {
			loginPath = "/t/" + slug + "/login"
		} else if req.ClientID != "" && h.oauthSvc != nil {
			if appTenantID, err := h.oauthSvc.ResolveAppTenant(r.Context(), req.ClientID); err == nil && h.tenantRepo != nil {
				if t, terr := h.tenantRepo.GetByID(r.Context(), appTenantID); terr == nil && t != nil {
					loginPath = "/t/" + t.Slug + "/login"
				}
			}
		}
		http.Redirect(w, r, loginPath+"?"+r.URL.RawQuery, http.StatusFound)
		return
	}

	resp, err := h.oauthSvc.Authorize(ctx, userID, req)
	if err != nil {
		h.writeOAuthAuthorizeError(w, r, err, r.URL.Query())
		return
	}

	// Redirect with code
	redirectURL := resp.RedirectURI + "?code=" + resp.Code
	if resp.State != "" {
		redirectURL += "&state=" + resp.State
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	req := oauth.TokenRequest{
		GrantType:    r.FormValue("grant_type"),
		Code:         r.FormValue("code"),
		RedirectURI:  r.FormValue("redirect_uri"),
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
		CodeVerifier: r.FormValue("code_verifier"),
		RefreshToken: r.FormValue("refresh_token"),
		Scope:        r.FormValue("scope"),
	}

	// Try Basic auth for client credentials
	if req.ClientID == "" {
		clientID, clientSecret, ok := r.BasicAuth()
		if ok {
			req.ClientID = clientID
			req.ClientSecret = clientSecret
		}
	}

	pair, err := h.oauthSvc.Exchange(r.Context(), req)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	h.eventSvc.Publish(r.Context(), events.Event{
		Type:     events.EventTokenIssued,
		TenantID: tenantID.String(),
		IP:       extractIP(r.RemoteAddr),
		Data:     map[string]interface{}{"grant_type": req.GrantType, "client_id": req.ClientID},
	})

	middleware.WriteJSON(w, http.StatusOK, pair)
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	token := r.FormValue("token")
	tokenTypeHint := r.FormValue("token_type_hint")

	if err := h.oauthSvc.Revoke(r.Context(), token, tokenTypeHint); err != nil {
		middleware.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Userinfo(w http.ResponseWriter, r *http.Request) {
	// Userinfo requires a Bearer token (manual validation since this is a public route)
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		auth := r.Header.Get("Authorization")
		if auth == "" || len(auth) < 8 || auth[:7] != "Bearer " {
			middleware.WriteError(w, models.ErrUnauthorized)
			return
		}
		var err error
		claims, err = h.tokenSvc.ValidateAccessToken(r.Context(), auth[7:])
		if err != nil {
			middleware.WriteError(w, err)
			return
		}
	}

	info, err := h.oauthSvc.GetUserinfo(r.Context(), claims)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, info)
}

func (h *Handler) Discovery(w http.ResponseWriter, r *http.Request) {
	// When a tenant has been resolved (via /t/{slug}/, X-Tenant-ID,
	// subdomain or custom domain) emit the discovery document scoped to
	// that tenant so the issuer + endpoint URLs match where the request
	// came in. Otherwise emit the global discovery document — *not* a
	// host-derived one, because spoofing the Host header should not
	// influence the issuer claim that downstream RPs trust.
	if tid := middleware.GetTenantID(r.Context()); tid != uuid.Nil && h.tenantRepo != nil {
		if tenant, err := h.tenantRepo.GetByID(r.Context(), tid); err == nil && tenant != nil {
			middleware.WriteJSON(w, http.StatusOK, h.oauthSvc.DiscoveryDocumentForTenant(tenant))
			return
		}
	}
	middleware.WriteJSON(w, http.StatusOK, h.oauthSvc.DiscoveryDocument())
}

func (h *Handler) JWKS(w http.ResponseWriter, r *http.Request) {
	jwks := h.tokenSvc.GetJWKS()
	middleware.WriteJSON(w, http.StatusOK, jwks)
}

// --- Passwordless ---

type passwordlessStartReq struct {
	Email      string `json:"email"`
	Connection string `json:"connection"` // "email" or "sms"
	Send       string `json:"send"`       // "link" or "code"
	ClientID   string `json:"client_id"`
}

func (h *Handler) PasswordlessStart(w http.ResponseWriter, r *http.Request) {
	var req passwordlessStartReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, err)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	// Same precedence as Login/Register: client_id wins over host tenant.
	if req.ClientID != "" && h.oauthSvc != nil {
		if appTenantID, err := h.oauthSvc.ResolveAppTenant(r.Context(), req.ClientID); err == nil {
			tenantID = appTenantID
		}
	}
	user, err := h.userSvc.GetByEmail(r.Context(), tenantID, req.Email)
	if err != nil {
		// Don't reveal user existence
		middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "sent"})
		return
	}

	// Generate OTP
	code, err := h.mfaSvc.GenerateEmailOTP(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("failed to generate email OTP", zap.Error(err))
	}

	_ = code // In production, this would be sent via email service

	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (h *Handler) PasswordlessVerify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, err)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user, err := h.userSvc.GetByEmail(r.Context(), tenantID, req.Email)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	if err := h.mfaSvc.ValidateTOTP(r.Context(), user.ID, req.Code); err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":  user.ID,
		"verified": true,
	})
}

// --- MFA ---

func (h *Handler) MFAChallenge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MFAToken    string `json:"mfa_token"`
		ChallengeType string `json:"challenge_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]string{
		"challenge_type": req.ChallengeType,
		"status":         "pending",
	})
}

func (h *Handler) MFAVerify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MFAToken string `json:"mfa_token"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if err := h.mfaSvc.ValidateTOTP(r.Context(), userID, req.Code); err != nil {
		// Try recovery code
		if rcErr := h.mfaSvc.VerifyRecoveryCode(r.Context(), userID, req.Code); rcErr != nil {
			middleware.WriteError(w, err)
			return
		}
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"verified": true,
	})
}

// --- WebAuthn ---

func (h *Handler) WebAuthnRegisterBegin(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	creation, err := h.webauthnSvc.BeginRegistration(r.Context(), tenantID, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, creation)
}

func (h *Handler) WebAuthnRegisterFinish(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "registered"})
}

func (h *Handler) WebAuthnLoginBegin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, err)
		return
	}

	userID, _ := uuid.Parse(req.UserID)
	tenantID := middleware.GetTenantID(r.Context())

	assertion, err := h.webauthnSvc.BeginLogin(r.Context(), tenantID, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, assertion)
}

func (h *Handler) WebAuthnLoginFinish(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "authenticated"})
}

// --- SAML ---

func (h *Handler) SAMLMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"></EntityDescriptor>`))
}

func (h *Handler) SAMLACS(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "processed"})
}

func (h *Handler) SAMLSSO(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "sso_initiated"})
}

// extractIP returns just the IP address from RemoteAddr (strips port).
const sessionCookieName = "__cpi_auth_session"

func setSessionCookie(w http.ResponseWriter, r *http.Request, sessionID string, rememberMe bool, cfg *config.Config) {
	maxAge := 24 * 3600 // 24 hours default
	if rememberMe {
		maxAge = 30 * 24 * 3600 // 30 days
	}
	if cfg != nil && cfg.Security.SessionLifetime > 0 {
		maxAge = int(cfg.Security.SessionLifetime.Seconds())
		if rememberMe {
			maxAge = maxAge * 30
			if maxAge > 90*24*3600 {
				maxAge = 90 * 24 * 3600
			}
		}
	}
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func issuerFromRequest(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	host := r.Host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		// Keep port only if it's non-standard
		port := host[idx+1:]
		if (scheme == "https" && port == "443") || (scheme == "http" && port == "80") {
			host = host[:idx]
		}
	}
	return scheme + "://" + host
}

func extractIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// --- Login / Register (consumed by login-ui) ---

type loginRequest struct {
	Email               string `json:"email"`
	Password            string `json:"password"`
	RememberMe          bool   `json:"remember_me"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	ResponseType        string `json:"response_type"`
}

type registerRequest struct {
	Email               string                 `json:"email"`
	Password            string                 `json:"password"`
	Name                string                 `json:"name"`
	Locale              string                 `json:"locale"`
	CustomFields        map[string]interface{} `json:"custom_fields"`
	ClientID            string                 `json:"client_id"`
	RedirectURI         string                 `json:"redirect_uri"`
	Scope               string                 `json:"scope"`
	State               string                 `json:"state"`
	CodeChallenge       string                 `json:"code_challenge"`
	CodeChallengeMethod string                 `json:"code_challenge_method"`
	ResponseType        string                 `json:"response_type"`
}

type authResponse struct {
	RedirectURL  string   `json:"redirect_url,omitempty"`
	AccessToken  string   `json:"access_token,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	TokenType    string   `json:"token_type,omitempty"`
	ExpiresIn    int      `json:"expires_in,omitempty"`
	MFAToken     string   `json:"mfa_token,omitempty"`
	MFARequired  bool     `json:"mfa_required,omitempty"`
	MFAMethods   []string `json:"mfa_methods,omitempty"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid request body."))
		return
	}

	if req.Email == "" || req.Password == "" {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Email and password are required."))
		return
	}

	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)

	// If the login is part of an OAuth flow, the application's tenant must
	// take precedence over the host-derived tenant — otherwise a user with
	// the same email in the default tenant would be authenticated instead of
	// the user that actually belongs to the app's tenant.
	if req.ClientID != "" {
		appTenantID, err := h.oauthSvc.ResolveAppTenant(ctx, req.ClientID)
		if err != nil {
			middleware.WriteError(w, err)
			return
		}
		tenantID = appTenantID
		ctx = context.WithValue(ctx, middleware.ContextKeyTenantID, tenantID)
	}

	// Execute pre-login actions
	if h.actionPipeline != nil {
		preResult, err := h.actionPipeline.Execute(ctx, tenantID, actions.TriggerPreLogin, &actions.ActionContext{
			TenantID: tenantID,
			IP:       extractIP(r.RemoteAddr),
			Data:     map[string]interface{}{"email": req.Email},
		})
		if err != nil {
			h.logger.Error("pre-login action pipeline failed", zap.Error(err))
		} else if preResult != nil && !preResult.Allow {
			middleware.WriteError(w, models.ErrForbidden.WithMessage(preResult.Message))
			return
		}
	}

	user, err := h.userSvc.Authenticate(ctx, tenantID, req.Email, req.Password)
	if err != nil {
		h.eventSvc.Publish(ctx, events.Event{
			Type:     events.EventLoginFailed,
			TenantID: tenantID.String(),
			IP:       extractIP(r.RemoteAddr),
			Data:     map[string]interface{}{"email": req.Email},
		})
		middleware.WriteError(w, err)
		return
	}

	// Check MFA
	hasMFA, err := h.mfaSvc.HasVerifiedMFA(ctx, user.ID)
	if err != nil {
		h.logger.Error("failed to check MFA status", zap.Error(err))
	}
	if hasMFA {
		mfaToken, _ := crypto.GenerateOpaqueToken()
		middleware.WriteJSON(w, http.StatusOK, authResponse{
			MFARequired: true,
			MFAToken:    mfaToken,
			MFAMethods:  []string{"totp"},
		})
		return
	}

	// Create session and set cookie for browser-based flows (OAuth, SSO)
	sess, err := h.sessionSvc.Create(ctx, sessions.CreateSessionInput{
		UserID:    user.ID,
		TenantID:  tenantID,
		IP:        extractIP(r.RemoteAddr),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		h.logger.Error("failed to create session", zap.Error(err))
	}
	if sess != nil {
		setSessionCookie(w, r, sess.ID.String(), req.RememberMe, h.cfg)
	}

	h.eventSvc.Publish(ctx, events.Event{
		Type:     events.EventLoginSuccess,
		TenantID: tenantID.String(),
		ActorID:  user.ID.String(),
		IP:       extractIP(r.RemoteAddr),
		Data:     map[string]interface{}{"email": user.Email},
	})

	// If OAuth params present, do authorization code flow
	if req.ClientID != "" && req.RedirectURI != "" && req.CodeChallenge != "" {
		resp, err := h.oauthSvc.Authorize(ctx, user.ID, oauth.AuthorizeRequest{
			ClientID:            req.ClientID,
			RedirectURI:         req.RedirectURI,
			ResponseType:        req.ResponseType,
			Scope:               req.Scope,
			State:               req.State,
			CodeChallenge:       req.CodeChallenge,
			CodeChallengeMethod: req.CodeChallengeMethod,
		})
		if err != nil {
			middleware.WriteError(w, err)
			return
		}

		redirectURL := resp.RedirectURI + "?code=" + resp.Code
		if resp.State != "" {
			redirectURL += "&state=" + resp.State
		}
		middleware.WriteJSON(w, http.StatusOK, authResponse{RedirectURL: redirectURL})
		return
	}

	// Otherwise issue tokens directly
	// Load user permissions for the token
	userPerms, permErr := h.rbacSvc.GetEffectivePermissions(ctx, user.ID)
	if permErr != nil {
		h.logger.Warn("failed to load user permissions for login token", zap.Error(permErr))
	}

	pair, err := h.tokenSvc.IssueTokenPair(ctx, tokens.IssueTokenPairInput{
		UserID:      user.ID,
		TenantID:    tenantID,
		Email:       user.Email,
		Name:        user.Name,
		Scopes:      []string{"openid", "profile", "email"},
		Permissions: userPerms,
		Issuer:      issuerFromRequest(r),
	})
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	// Execute post-login actions (fire-and-forget)
	if h.actionPipeline != nil {
		userIDCopy := user.ID
		h.actionPipeline.Execute(ctx, tenantID, actions.TriggerPostLogin, &actions.ActionContext{
			TenantID: tenantID,
			UserID:   &userIDCopy,
			IP:       extractIP(r.RemoteAddr),
			Data:     map[string]interface{}{"email": user.Email, "user_id": user.ID.String()},
		})
	}

	middleware.WriteJSON(w, http.StatusOK, authResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    pair.TokenType,
		ExpiresIn:    pair.ExpiresIn,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Revoke session from cookie
	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		sessionID, parseErr := uuid.Parse(cookie.Value)
		if parseErr == nil && h.sessionSvc != nil {
			h.sessionSvc.Revoke(r.Context(), sessionID)
		}
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to post-logout URL if provided.
	if postLogoutURL := r.URL.Query().Get("post_logout_redirect_uri"); postLogoutURL != "" {
		http.Redirect(w, r, postLogoutURL, http.StatusFound)
		return
	}
	// "Switch account" flow from the OAuth error page: bounce back to a
	// same-origin path so the user re-enters the OAuth flow without their
	// previous session. Restricted to single-leading-slash paths to avoid
	// open-redirect.
	if returnTo := r.URL.Query().Get("return_to"); returnTo != "" &&
		strings.HasPrefix(returnTo, "/") && !strings.HasPrefix(returnTo, "//") {
		http.Redirect(w, r, returnTo, http.StatusFound)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid request body."))
		return
	}

	if req.Email == "" || req.Password == "" {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Email and password are required."))
		return
	}

	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)

	// Same precedence as Login: when the request carries an OAuth client_id
	// the application's tenant must take priority over the host-derived
	// tenant, otherwise registering on /t/lastsoftware/register with
	// client_id=… would create the user in the default tenant instead of
	// LastSoftware (and falsely report "user already exists" if there's a
	// same-email user in the default tenant).
	if req.ClientID != "" {
		appTenantID, err := h.oauthSvc.ResolveAppTenant(ctx, req.ClientID)
		if err != nil {
			middleware.WriteError(w, err)
			return
		}
		tenantID = appTenantID
		ctx = context.WithValue(ctx, middleware.ContextKeyTenantID, tenantID)
	}

	// Execute pre-registration actions
	if h.actionPipeline != nil {
		preResult, err := h.actionPipeline.Execute(ctx, tenantID, actions.TriggerPreRegistration, &actions.ActionContext{
			TenantID: tenantID,
			IP:       extractIP(r.RemoteAddr),
			Data:     map[string]interface{}{"email": req.Email, "name": req.Name},
		})
		if err != nil {
			h.logger.Error("pre-registration action pipeline failed", zap.Error(err))
		} else if preResult != nil && !preResult.Allow {
			middleware.WriteError(w, models.ErrForbidden.WithMessage(preResult.Message))
			return
		}
	}

	var metadata json.RawMessage
	if req.CustomFields != nil {
		metadata, _ = json.Marshal(req.CustomFields)
	}

	user, err := h.userSvc.Register(ctx, tenantID, users.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Locale:   req.Locale,
		Metadata: metadata,
	})
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	h.eventSvc.Publish(ctx, events.Event{
		Type:     events.EventUserCreated,
		TenantID: tenantID.String(),
		ActorID:  user.ID.String(),
		IP:       extractIP(r.RemoteAddr),
		Data:     map[string]interface{}{"email": user.Email},
	})

	// Create session AND set the cookie so subsequent requests on the same
	// browser see the user as authenticated. Without setSessionCookie the
	// freshly registered user lands at /login?registered=true and has to
	// type their credentials again to actually start a session.
	sess, err := h.sessionSvc.Create(ctx, sessions.CreateSessionInput{
		UserID:    user.ID,
		TenantID:  tenantID,
		IP:        extractIP(r.RemoteAddr),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		h.logger.Error("failed to create session", zap.Error(err))
	}
	if sess != nil {
		setSessionCookie(w, r, sess.ID.String(), false, h.cfg)
	}

	// Issue tokens
	pair, err := h.tokenSvc.IssueTokenPair(ctx, tokens.IssueTokenPairInput{
		UserID:   user.ID,
		TenantID: tenantID,
		Email:    user.Email,
		Name:     user.Name,
		Scopes:   []string{"openid", "profile", "email"},
	})
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	// Execute post-registration actions (fire-and-forget)
	if h.actionPipeline != nil {
		userIDCopy := user.ID
		h.actionPipeline.Execute(ctx, tenantID, actions.TriggerPostRegistration, &actions.ActionContext{
			TenantID: tenantID,
			UserID:   &userIDCopy,
			IP:       extractIP(r.RemoteAddr),
			Data:     map[string]interface{}{"email": user.Email, "user_id": user.ID.String()},
		})
	}

	// If the registration is part of an OAuth code flow, run the authorize
	// step right here so the UI can redirect to the relying party with a
	// fresh code instead of bouncing back through /login.
	if req.ClientID != "" && req.RedirectURI != "" && req.CodeChallenge != "" {
		resp, authErr := h.oauthSvc.Authorize(ctx, user.ID, oauth.AuthorizeRequest{
			ClientID:            req.ClientID,
			RedirectURI:         req.RedirectURI,
			ResponseType:        req.ResponseType,
			Scope:               req.Scope,
			State:               req.State,
			CodeChallenge:       req.CodeChallenge,
			CodeChallengeMethod: req.CodeChallengeMethod,
		})
		if authErr != nil {
			middleware.WriteError(w, authErr)
			return
		}
		redirectURL := resp.RedirectURI + "?code=" + resp.Code
		if resp.State != "" {
			redirectURL += "&state=" + resp.State
		}
		middleware.WriteJSON(w, http.StatusOK, authResponse{
			AccessToken:  pair.AccessToken,
			RefreshToken: pair.RefreshToken,
			TokenType:    pair.TokenType,
			ExpiresIn:    pair.ExpiresIn,
			RedirectURL:  redirectURL,
		})
		return
	}

	middleware.WriteJSON(w, http.StatusOK, authResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    pair.TokenType,
		ExpiresIn:    pair.ExpiresIn,
	})
}
