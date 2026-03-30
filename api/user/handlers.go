package user

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/api/middleware"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/events"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/federation"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/flows"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/sessions"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/users"
)

// Handler holds dependencies for user self-service API handlers.
type Handler struct {
	userSvc     *users.Service
	sessionSvc  *sessions.Service
	mfaSvc      *flows.MFAService
	webauthnSvc *federation.WebAuthnService
	eventSvc    *events.Service
	logger      *zap.Logger
}

// NewHandler creates new user self-service handlers.
func NewHandler(
	userSvc *users.Service,
	sessionSvc *sessions.Service,
	mfaSvc *flows.MFAService,
	webauthnSvc *federation.WebAuthnService,
	eventSvc *events.Service,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		userSvc:     userSvc,
		sessionSvc:  sessionSvc,
		mfaSvc:      mfaSvc,
		webauthnSvc: webauthnSvc,
		eventSvc:    eventSvc,
		logger:      logger,
	}
}

// RegisterRoutes registers all user self-service routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/v1/users/me", h.GetMe)
	r.Patch("/v1/users/me", h.UpdateMe)
	r.Post("/v1/users/me/change-password", h.ChangePassword)
	r.Get("/v1/users/me/sessions", h.ListSessions)
	r.Delete("/v1/users/me/sessions/{sessionId}", h.RevokeSession)

	r.Get("/v1/users/me/mfa", h.ListMFA)
	r.Post("/v1/users/me/mfa/totp/enroll", h.EnrollTOTP)
	r.Post("/v1/users/me/mfa/totp/verify", h.VerifyTOTP)
	r.Delete("/v1/users/me/mfa/{enrollmentId}", h.DeleteMFA)
	r.Get("/v1/users/me/mfa/recovery-codes", h.ListRecoveryCodes)
	r.Post("/v1/users/me/mfa/recovery-codes/regenerate", h.RegenerateRecoveryCodes)

	r.Get("/v1/users/me/passkeys", h.ListPasskeys)
	r.Post("/v1/users/me/passkeys", h.CreatePasskey)
	r.Delete("/v1/users/me/passkeys/{passkeyId}", h.DeletePasskey)

	r.Get("/v1/users/me/identities", h.ListIdentities)
	r.Delete("/v1/users/me/identities/{identityId}", h.UnlinkIdentity)

	r.Post("/v1/users/me/export", h.ExportData)
	r.Delete("/v1/users/me", h.DeleteAccount)
}

// --- Profile ---

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	user, err := h.userSvc.GetByID(r.Context(), tenantID, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	user, err := h.userSvc.GetByID(r.Context(), tenantID, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	var update struct {
		Name      *string          `json:"name"`
		Phone     *string          `json:"phone"`
		AvatarURL *string          `json:"avatar_url"`
		Metadata  *json.RawMessage `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}

	if update.Name != nil {
		user.Name = *update.Name
	}
	if update.Phone != nil {
		user.Phone = *update.Phone
	}
	if update.AvatarURL != nil {
		user.AvatarURL = *update.AvatarURL
	}
	if update.Metadata != nil {
		user.Metadata = *update.Metadata
	}

	if err := h.userSvc.Update(r.Context(), user); err != nil {
		middleware.WriteError(w, err)
		return
	}

	h.eventSvc.Publish(r.Context(), events.Event{
		Type:     events.EventUserUpdated,
		TenantID: tenantID.String(),
		ActorID:  userID.String(),
		IP:       r.RemoteAddr,
	})

	middleware.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}

	if err := h.userSvc.ChangePassword(r.Context(), tenantID, userID, req.OldPassword, req.NewPassword); err != nil {
		middleware.WriteError(w, err)
		return
	}

	h.eventSvc.Publish(r.Context(), events.Event{
		Type:     events.EventPasswordChanged,
		TenantID: tenantID.String(),
		ActorID:  userID.String(),
		IP:       r.RemoteAddr,
	})

	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "password_changed"})
}

// --- Sessions ---

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	sessions, err := h.sessionSvc.ListByUser(r.Context(), userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, sessions)
}

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid session ID."))
		return
	}

	if err := h.sessionSvc.Revoke(r.Context(), sessionID); err != nil {
		middleware.WriteError(w, err)
		return
	}

	h.eventSvc.Publish(r.Context(), events.Event{
		Type:     events.EventSessionRevoked,
		TenantID: middleware.GetTenantID(r.Context()).String(),
		ActorID:  middleware.GetUserID(r.Context()).String(),
		IP:       r.RemoteAddr,
		Data:     map[string]interface{}{"session_id": sessionID.String()},
	})

	w.WriteHeader(http.StatusNoContent)
}

// --- MFA ---

func (h *Handler) ListMFA(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	enrollments, err := h.mfaSvc.ListEnrollments(r.Context(), userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	if enrollments == nil {
		enrollments = []models.MFAEnrollment{}
	}

	middleware.WriteJSON(w, http.StatusOK, enrollments)
}

func (h *Handler) EnrollTOTP(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	user, err := h.userSvc.GetByID(r.Context(), tenantID, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	enrollment, err := h.mfaSvc.EnrollTOTP(r.Context(), userID, user.Email)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, enrollment)
}

func (h *Handler) VerifyTOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EnrollmentID string `json:"enrollment_id"`
		Code         string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, models.ErrBadRequest.Wrap(err))
		return
	}

	enrollmentID, err := uuid.Parse(req.EnrollmentID)
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid enrollment ID."))
		return
	}

	if err := h.mfaSvc.VerifyTOTP(r.Context(), enrollmentID, req.Code); err != nil {
		middleware.WriteError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	h.eventSvc.Publish(r.Context(), events.Event{
		Type:     events.EventMFAEnrolled,
		TenantID: tenantID.String(),
		ActorID:  userID.String(),
		IP:       r.RemoteAddr,
		Data:     map[string]interface{}{"method": "totp"},
	})

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{"verified": true})
}

func (h *Handler) DeleteMFA(w http.ResponseWriter, r *http.Request) {
	enrollmentID, err := uuid.Parse(chi.URLParam(r, "enrollmentId"))
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid enrollment ID."))
		return
	}

	if err := h.mfaSvc.DeleteEnrollment(r.Context(), enrollmentID); err != nil {
		middleware.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	codes, err := h.mfaSvc.ListRecoveryCodes(r.Context(), userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, codes)
}

func (h *Handler) RegenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	codes, err := h.mfaSvc.GenerateRecoveryCodes(r.Context(), userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, map[string]interface{}{"codes": codes})
}

// --- Passkeys ---

func (h *Handler) ListPasskeys(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	creds, err := h.webauthnSvc.ListCredentials(r.Context(), userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, creds)
}

func (h *Handler) CreatePasskey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	creation, err := h.webauthnSvc.BeginRegistration(r.Context(), tenantID, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, creation)
}

func (h *Handler) DeletePasskey(w http.ResponseWriter, r *http.Request) {
	passkeyID, err := uuid.Parse(chi.URLParam(r, "passkeyId"))
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid passkey ID."))
		return
	}

	if err := h.webauthnSvc.DeleteCredential(r.Context(), passkeyID); err != nil {
		middleware.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Identities ---

func (h *Handler) ListIdentities(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	identities, err := h.userSvc.ListIdentities(r.Context(), userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, identities)
}

func (h *Handler) UnlinkIdentity(w http.ResponseWriter, r *http.Request) {
	identityID, err := uuid.Parse(chi.URLParam(r, "identityId"))
	if err != nil {
		middleware.WriteError(w, models.ErrBadRequest.WithMessage("Invalid identity ID."))
		return
	}

	if err := h.userSvc.UnlinkIdentity(r.Context(), identityID); err != nil {
		middleware.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Data Export & Deletion ---

func (h *Handler) ExportData(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	data, err := h.userSvc.ExportUserData(r.Context(), tenantID, userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, data)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	// Revoke all sessions
	_ = h.sessionSvc.RevokeAllForUser(r.Context(), userID)

	// Delete user
	if err := h.userSvc.Delete(r.Context(), tenantID, userID); err != nil {
		middleware.WriteError(w, err)
		return
	}

	h.eventSvc.Publish(r.Context(), events.Event{
		Type:     events.EventUserDeleted,
		TenantID: tenantID.String(),
		ActorID:  userID.String(),
		IP:       r.RemoteAddr,
	})

	w.WriteHeader(http.StatusNoContent)
}
