package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	adminAPI "github.com/CPI-Technologies-GmbH/CPI-Auth/api/admin"
	authAPI "github.com/CPI-Technologies-GmbH/CPI-Auth/api/auth"
	mw "github.com/CPI-Technologies-GmbH/CPI-Auth/api/middleware"
	userAPI "github.com/CPI-Technologies-GmbH/CPI-Auth/api/user"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/actions"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/crypto"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/db"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/domains"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/license"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/events"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/federation"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/flows"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/oauth"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/policy"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/sessions"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/tokens"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/users"
)

func main() {
	// Load configuration
	configPath := os.Getenv("AF_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	var logger *zap.Logger
	if cfg.Logging.Level == "debug" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	defer logger.Sync()

	logger.Info("starting CPI Auth",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Connect to PostgreSQL ---
	pool, err := db.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()
	logger.Info("connected to PostgreSQL")

	// Auto-migrate database schema
	if err := db.AutoMigrate(ctx, pool, logger); err != nil {
		logger.Fatal("auto-migration failed", zap.Error(err))
	}

	// Bootstrap default tenant with runtime config (domain, etc.)
	db.Bootstrap(ctx, pool, logger)

	// --- Connect to Redis ---
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("failed to connect to Redis, continuing without cache", zap.Error(err))
	} else {
		logger.Info("connected to Redis")
	}
	defer rdb.Close()

	// --- Connect to NATS ---
	var nc *nats.Conn
	nc, err = nats.Connect(cfg.NATS.URL)
	if err != nil {
		logger.Warn("failed to connect to NATS, continuing without event bus", zap.Error(err))
	} else {
		logger.Info("connected to NATS")
		defer nc.Close()
	}

	// --- Initialize Repositories ---
	userRepo := db.NewUserRepository(pool)
	tenantRepo := db.NewTenantRepository(pool)
	appRepo := db.NewApplicationRepository(pool)
	identityRepo := db.NewIdentityRepository(pool)
	orgRepo := db.NewOrganizationRepository(pool)
	roleRepo := db.NewRoleRepository(pool)
	permRepo := db.NewPermissionRepository(pool)
	appPermRepo := db.NewApplicationPermissionRepository(pool)
	sessionRepo := db.NewSessionRepository(pool)
	grantRepo := db.NewOAuthGrantRepository(pool)
	refreshTokenRepo := db.NewRefreshTokenRepository(pool)
	mfaRepo := db.NewMFAEnrollmentRepository(pool)
	recoveryCodeRepo := db.NewRecoveryCodeRepository(pool)
	webauthnCredRepo := db.NewWebAuthnCredentialRepository(pool)
	auditLogRepo := db.NewAuditLogRepository(pool)
	webhookRepo := db.NewWebhookRepository(pool)
	actionRepo := db.NewActionRepository(pool)
	emailTemplateRepo := db.NewEmailTemplateRepository(pool)
	apiKeyRepo := db.NewAPIKeyRepository(pool)
	customFieldRepo := db.NewCustomFieldDefinitionRepository(pool)
	domainVerificationRepo := db.NewDomainVerificationRepository(pool)
	pageTemplateRepo := db.NewPageTemplateRepository(pool)
	langStringRepo := db.NewLanguageStringRepository(pool)
	licenseChecker := license.NewChecker(pool)

	// --- Initialize Crypto / Key Management ---
	var keyPair *crypto.KeyPair
	switch cfg.Security.JWTSigningAlgorithm {
	case "ES256":
		keyPair, err = crypto.GenerateECDSAKeyPair()
	default:
		keyPair, err = crypto.GenerateRSAKeyPair(2048)
	}
	if err != nil {
		logger.Fatal("failed to generate signing key pair", zap.Error(err))
	}

	// Try loading keys from file if configured
	if cfg.Security.JWTPrivateKeyPath != "" {
		privKey, loadErr := crypto.LoadRSAPrivateKeyFromFile(cfg.Security.JWTPrivateKeyPath)
		if loadErr == nil {
			keyPair.PrivateKey = privKey
			keyPair.PublicKey = &privKey.PublicKey
			keyPair.Algorithm = "RS256"
			logger.Info("loaded RSA signing key from file")
		} else {
			logger.Warn("failed to load RSA key from file, using generated key", zap.Error(loadErr))
		}
	}

	keyManager := crypto.NewKeyManager(keyPair)

	// --- Initialize Core Services ---
	tokenSvc := tokens.NewService(keyManager, refreshTokenRepo, rdb, cfg, logger)
	userSvc := users.NewService(userRepo, identityRepo, cfg, logger)
	sessionSvc := sessions.NewService(sessionRepo, rdb, cfg, logger)
	eventSvc := events.NewService(nc, auditLogRepo, webhookRepo, logger)
	mfaSvc := flows.NewMFAService(mfaRepo, recoveryCodeRepo, cfg, logger)
	rbacSvc := policy.NewRBACService(roleRepo, logger)
	oauthSvc := oauth.NewService(appRepo, grantRepo, userRepo, tokenSvc, rbacSvc, appPermRepo, cfg, logger)
	_ = policy.NewFGAService(db.NewFGATupleRepository(pool), logger)
	actionPipeline := actions.NewPipeline(actionRepo, logger)
	domainSvc := domains.NewService(domainVerificationRepo, tenantRepo, logger)

	// Initialize Federation
	federationSvc := federation.NewService(identityRepo, userRepo, cfg, logger)
	_ = federationSvc // Available for provider registration

	// Initialize WebAuthn
	webauthnSvc, err := federation.NewWebAuthnService(webauthnCredRepo, userRepo, rdb, cfg, logger)
	if err != nil {
		logger.Warn("failed to initialize WebAuthn service", zap.Error(err))
	}

	// --- Build Router ---
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RealIP)
	r.Use(mw.Recovery(logger))
	r.Use(mw.CorrelationID)
	r.Use(mw.RequestLogger(logger))
	r.Use(mw.CSPHeaders(cfg.Security.CSPHeader))

	if cfg.Security.RateLimitEnabled {
		r.Use(mw.RateLimit(cfg.Security.RateLimitRequestsPerSec))
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Security.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID", "X-Correlation-ID", "X-API-Key"},
		ExposedHeaders:   []string{"X-Correlation-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Tenant resolution
	r.Use(mw.TenantResolver(tenantRepo))

	// --- Health & Observability ---
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		mw.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		checks := map[string]string{}

		if err := pool.Ping(r.Context()); err != nil {
			checks["database"] = "error: " + err.Error()
		} else {
			checks["database"] = "ok"
		}

		if err := rdb.Ping(r.Context()).Err(); err != nil {
			checks["redis"] = "error: " + err.Error()
		} else {
			checks["redis"] = "ok"
		}

		if nc != nil && nc.IsConnected() {
			checks["nats"] = "ok"
		} else {
			checks["nats"] = "disconnected"
		}

		status := http.StatusOK
		for _, v := range checks {
			if v != "ok" && v != "disconnected" {
				status = http.StatusServiceUnavailable
				break
			}
		}

		mw.WriteJSON(w, status, map[string]interface{}{
			"status": "ready",
			"checks": checks,
		})
	})

	if cfg.Metrics.Enabled {
		r.Handle(cfg.Metrics.Path, promhttp.Handler())
	}

	// --- Auth Routes (public) ---
	authHandler := authAPI.NewHandler(oauthSvc, userSvc, tokenSvc, sessionSvc, mfaSvc, webauthnSvc, eventSvc, rbacSvc, actionPipeline, logger)
	authHandler.RegisterRoutes(r)

	// --- Admin Routes (all under /admin) ---
	r.Route("/admin", func(r chi.Router) {
		// Public auth endpoints (no auth middleware)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Token)

		// Authenticated admin auth endpoints
		r.Group(func(r chi.Router) {
			r.Use(mw.Authentication(tokenSvc))
			r.Post("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
				userID := mw.GetUserID(r.Context())
				if userID != uuid.Nil {
					sessionSvc.RevokeAllForUser(r.Context(), userID)
				}
				w.WriteHeader(http.StatusNoContent)
			})
			r.Get("/auth/me", func(w http.ResponseWriter, r *http.Request) {
				claims := mw.GetClaims(r.Context())
				if claims == nil {
					mw.WriteError(w, models.ErrUnauthorized)
					return
				}
				tenantID, _ := uuid.Parse(claims.TenantID)
				userID, _ := uuid.Parse(claims.Subject)
				user, err := userSvc.GetByID(r.Context(), tenantID, userID)
				if err != nil {
					mw.WriteError(w, err)
					return
				}
				role := "viewer"
				if user.AppMetadata != nil {
					var meta map[string]interface{}
					if jsonErr := json.Unmarshal(user.AppMetadata, &meta); jsonErr == nil {
						if meta["is_system_admin"] == true {
							role = "super_admin"
						} else if roles, ok := meta["roles"].([]interface{}); ok {
							for _, rl := range roles {
								if rl == "admin" {
									role = "admin"
									break
								}
							}
						}
					}
				}
				mw.WriteJSON(w, http.StatusOK, map[string]interface{}{
					"id":         user.ID.String(),
					"email":      user.Email,
					"name":       user.Name,
					"avatar_url": user.AvatarURL,
					"role":       role,
					"tenant_id":  user.TenantID.String(),
					"created_at": user.CreatedAt,
				})
			})
		})

		// Admin CRUD routes (authenticated + permission check)
		r.Group(func(r chi.Router) {
			r.Use(mw.Authentication(tokenSvc))
			r.Use(mw.RequirePermission(rbacSvc, "admin:access"))
			r.Use(mw.SuperAdminTenantOverride())
			adminHandler := adminAPI.NewHandler(
				userSvc, sessionSvc, eventSvc, actionPipeline,
				tokenSvc, rbacSvc,
				tenantRepo, appRepo, orgRepo, roleRepo,
				webhookRepo, emailTemplateRepo, apiKeyRepo,
				permRepo, appPermRepo, customFieldRepo, domainSvc,
				pageTemplateRepo, langStringRepo, licenseChecker, logger,
			)
			adminHandler.RegisterRoutes(r)
		})
	})

	// --- User Self-Service Routes (authenticated) ---
	r.Group(func(r chi.Router) {
		r.Use(mw.Authentication(tokenSvc))
		userHandler := userAPI.NewHandler(userSvc, sessionSvc, mfaSvc, webauthnSvc, eventSvc, logger)
		userHandler.RegisterRoutes(r)
	})

	// --- Start Server ---
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		logger.Info("received shutdown signal", zap.String("signal", sig.String()))

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown error", zap.Error(err))
		}
		cancel()
	}()

	logger.Info("CPI Auth is listening", zap.String("addr", addr))

	if cfg.Server.TLSCert != "" && cfg.Server.TLSKey != "" {
		err = srv.ListenAndServeTLS(cfg.Server.TLSCert, cfg.Server.TLSKey)
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		logger.Fatal("server failed", zap.Error(err))
	}

	logger.Info("CPI Auth shut down gracefully")
}
