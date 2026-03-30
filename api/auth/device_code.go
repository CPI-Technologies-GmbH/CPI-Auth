package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/api/middleware"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/tokens"
)

// userCodeAlphabet excludes ambiguous characters: I, O, 0, 1.
const userCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// generateUserCode creates a user-friendly code formatted as XXXX-XXXX.
func generateUserCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := make([]byte, 8)
	for i := range code {
		code[i] = userCodeAlphabet[int(b[i])%len(userCodeAlphabet)]
	}
	return fmt.Sprintf("%s-%s", string(code[:4]), string(code[4:])), nil
}

// generateDeviceCode creates a cryptographically random device code (32 bytes, hex encoded).
func generateDeviceCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// DeviceCode handles POST /oauth/device/code (public, no auth).
func (h *Handler) DeviceCode(w http.ResponseWriter, r *http.Request) {
	var clientID, scope string

	// Support both form and JSON body
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var req struct {
			ClientID string `json:"client_id"`
			Scope    string `json:"scope"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid request body."))
			return
		}
		clientID = req.ClientID
		scope = req.Scope
	} else {
		r.ParseForm()
		clientID = r.FormValue("client_id")
		scope = r.FormValue("scope")
	}

	if clientID == "" {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("client_id is required."))
		return
	}

	tenantID := middleware.GetTenantID(r.Context())

	deviceCode, err := generateDeviceCode()
	if err != nil {
		h.logger.Error("failed to generate device code", zap.Error(err))
		middleware.WriteError(w, models.ErrInternal)
		return
	}

	userCode, err := generateUserCode()
	if err != nil {
		h.logger.Error("failed to generate user code", zap.Error(err))
		middleware.WriteError(w, models.ErrInternal)
		return
	}

	var scopes []string
	if scope != "" {
		scopes = strings.Split(scope, " ")
	}

	dc := &models.DeviceCode{
		TenantID:     tenantID,
		DeviceCode:   deviceCode,
		UserCode:     userCode,
		ClientID:     clientID,
		Scopes:       scopes,
		Status:       "pending",
		ExpiresAt:    time.Now().UTC().Add(15 * time.Minute),
		PollInterval: 5,
	}

	if err := h.deviceCodeRepo.Create(r.Context(), dc); err != nil {
		h.logger.Error("failed to create device code", zap.Error(err))
		middleware.WriteError(w, models.ErrInternal)
		return
	}

	issuer := h.cfg.Security.Issuer

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"device_code":               deviceCode,
		"user_code":                 userCode,
		"verification_uri":          issuer + "/device",
		"verification_uri_complete": issuer + "/device?code=" + userCode,
		"expires_in":                900,
		"interval":                  5,
	})
}

// DeviceToken handles POST /oauth/device/token (public, no auth).
func (h *Handler) DeviceToken(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	deviceCode := r.FormValue("device_code")
	clientID := r.FormValue("client_id")
	grantType := r.FormValue("grant_type")

	if grantType != "urn:ietf:params:oauth:grant-type:device_code" {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "unsupported_grant_type",
		})
		return
	}

	if deviceCode == "" || clientID == "" {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid_request",
		})
		return
	}

	dc, err := h.deviceCodeRepo.GetByDeviceCode(r.Context(), deviceCode)
	if err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid_grant",
		})
		return
	}

	// Check expiry
	if time.Now().UTC().After(dc.ExpiresAt) {
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "expired_token",
		})
		return
	}

	switch dc.Status {
	case "pending":
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "authorization_pending",
		})
		return
	case "denied":
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "access_denied",
		})
		return
	case "authorized":
		// Issue tokens
		if dc.UserID == nil {
			middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "server_error",
			})
			return
		}

		tenantID := middleware.GetTenantID(r.Context())

		// Load user for token claims
		user, err := h.userSvc.GetByID(r.Context(), dc.TenantID, *dc.UserID)
		if err != nil {
			h.logger.Error("failed to load user for device token", zap.Error(err))
			middleware.WriteError(w, models.ErrInternal)
			return
		}

		// Load user permissions
		userPerms, permErr := h.rbacSvc.GetEffectivePermissions(r.Context(), user.ID)
		if permErr != nil {
			h.logger.Warn("failed to load user permissions for device token", zap.Error(permErr))
		}

		scopes := dc.Scopes
		if len(scopes) == 0 {
			scopes = []string{"openid", "profile", "email"}
		}

		pair, err := h.tokenSvc.IssueTokenPair(r.Context(), tokens.IssueTokenPairInput{
			UserID:      user.ID,
			TenantID:    tenantID,
			Email:       user.Email,
			Name:        user.Name,
			Scopes:      scopes,
			Permissions: userPerms,
		})
		if err != nil {
			h.logger.Error("failed to issue token pair for device code", zap.Error(err))
			middleware.WriteError(w, models.ErrInternal)
			return
		}

		// Delete the used device code
		h.deviceCodeRepo.DeleteExpired(r.Context())

		middleware.WriteJSON(w, http.StatusOK, pair)
		return
	default:
		middleware.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid_grant",
		})
	}
}

// DeviceAuthorize handles POST /oauth/device/authorize (authenticated).
func (h *Handler) DeviceAuthorize(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserCode string `json:"user_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid request body."))
		return
	}

	if req.UserCode == "" {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("user_code is required."))
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		middleware.WriteError(w, models.ErrUnauthorized)
		return
	}

	dc, err := h.deviceCodeRepo.GetByUserCode(r.Context(), req.UserCode)
	if err != nil {
		middleware.WriteError(w, models.ErrNotFound.WithMessage("Invalid user code."))
		return
	}

	if time.Now().UTC().After(dc.ExpiresAt) {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Device code has expired."))
		return
	}

	if err := h.deviceCodeRepo.Authorize(r.Context(), req.UserCode, userID); err != nil {
		h.logger.Error("failed to authorize device code", zap.Error(err))
		middleware.WriteError(w, models.ErrInternal)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "authorized",
		"user_code": req.UserCode,
	})
}
