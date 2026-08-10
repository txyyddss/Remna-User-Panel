// Package httpapi exposes the versioned JSON API and embedded Vue application.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/txyyddss/Remna-User-Panel/internal/accounts"
	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/admin"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/catalog"
	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/emby"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/questionnaires"
	"github.com/txyyddss/Remna-User-Panel/internal/requestauth"
)

const sessionCookie = "txc_session"

type contextKey string

const userContextKey contextKey = "user"

// PaymentWebhookVerifier authenticates legacy provider callbacks before domain settlement.
type PaymentWebhookVerifier interface {
	VerifyEZPay(context.Context, url.Values) (billing.ProviderEvent, bool, error)
	VerifyBEPusdt(context.Context, []byte) (billing.ProviderEvent, int, error)
}

type bepusdtUnsignedVerifier interface {
	VerifyBEPusdtUnsigned(context.Context, []byte) (billing.ProviderEvent, int, error)
}

// Dependencies contains already-constructed application services.
type Dependencies struct {
	Accounts          *accounts.Service
	Catalog           *catalog.Service
	Billing           *billing.Service
	Activity          *activity.Service
	Coupons           *coupons.Service
	Questionnaires    *questionnaires.Service
	Emby              *emby.Service
	EmbyPrice         emby.PriceSource
	Admin             *admin.Service
	Settings          *admin.SettingsService
	DatabaseAdmin     *DatabaseAdministrationHTTP
	Store             *database.Store
	Telegram          *telegram.Client
	Webhooks          PaymentWebhookVerifier
	PublicURL         *url.URL
	Static            fs.FS
	Logger            *slog.Logger
	RequestSigningKey []byte
	SessionTTL        time.Duration
	SecureCookies     bool
	AdminTelegramID   int64
}

// Server owns HTTP routing and transport-only validation.
type Server struct {
	deps     Dependencies
	requests *requestauth.Verifier
	router   http.Handler
}

// New constructs all public, authenticated, admin, and webhook routes.
func New(deps Dependencies) (*Server, error) {
	if deps.Accounts == nil || deps.Catalog == nil || deps.Billing == nil || deps.Activity == nil || deps.Coupons == nil || deps.Questionnaires == nil || deps.Emby == nil || deps.EmbyPrice == nil || deps.Admin == nil || deps.Settings == nil || deps.Store == nil || deps.Telegram == nil || deps.Webhooks == nil || deps.PublicURL == nil || deps.Logger == nil || deps.AdminTelegramID <= 0 {
		return nil, errors.New("HTTP API dependencies are incomplete")
	}
	requestVerifier, err := requestauth.New(deps.RequestSigningKey)
	if err != nil {
		return nil, err
	}
	server := &Server{deps: deps, requests: requestVerifier}
	router := chi.NewRouter()
	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)
	router.Use(server.recoverer)
	router.Use(server.securityHeaders)
	router.Use(server.accessLog)
	router.Use(server.validateRequest)

	router.Get("/healthz", server.health)
	router.Get("/readyz", server.ready)
	router.Post("/api/v1/auth/telegram", server.authenticate)
	router.Post("/api/v1/webhooks/telegram", server.telegramWebhook)
	router.Get("/api/v1/webhooks/ezpay", server.ezpayWebhook)
	router.Post("/api/v1/webhooks/bepusdt", server.bepusdtWebhook)
	router.Post("/api/v1/webhooks/bepusdt/{capability}", server.bepusdtWebhook)
	router.Get("/api/v1/payments/return/{provider}", server.paymentReturn)
	router.Get("/api/v1/payments/return/{provider}/{orderID}", server.paymentReturn)

	router.Group(func(authenticated chi.Router) {
		authenticated.Use(server.requireSignedRequest)
		authenticated.Use(server.requireSession)
		authenticated.Get("/api/v1/me", server.me)
		authenticated.Post("/api/v1/onboarding/invites", server.createInvites)
		authenticated.Post("/api/v1/onboarding/membership/check", server.checkMembership)
		authenticated.Put("/api/v1/onboarding/username", server.reserveUsername)
		authenticated.Post("/api/v1/onboarding/agreement", server.acceptAgreement)
		authenticated.Get("/api/v1/onboarding/content", server.onboardingContent)
		authenticated.Get("/api/v1/dashboard", server.dashboard)
		authenticated.Get("/api/v1/catalog", server.catalog)
		authenticated.Post("/api/v1/purchases/quote", server.purchaseQuote)
		authenticated.Post("/api/v1/purchases", server.purchase)
		authenticated.Get("/api/v1/purchases", server.purchases)
		authenticated.Post("/api/v1/subscription/revoke", server.revokeSubscription)
		authenticated.Get("/api/v1/balance", server.balance)
		authenticated.Get("/api/v1/ledger", server.ledger)
		authenticated.Post("/api/v1/payments/orders", server.createPaymentOrder)
		authenticated.Get("/api/v1/payments/orders/{id}", server.paymentOrder)
		authenticated.Post("/api/v1/payments/orders/{id}/cancel", server.cancelPaymentOrder)
		authenticated.Get("/api/v1/emby/account", server.embyAccount)
		authenticated.Post("/api/v1/emby/setup", server.setupEmby)
		authenticated.Put("/api/v1/emby/preferences", server.updateEmbyPreferences)
		authenticated.Put("/api/v1/emby/password", server.changeEmbyPassword)
		server.mountCommunity(authenticated)

		authenticated.Route("/api/v1/admin", func(adminRouter chi.Router) {
			adminRouter.Use(server.requireAdmin)
			if deps.DatabaseAdmin != nil {
				deps.DatabaseAdmin.Mount(adminRouter)
			}
			server.mountAdmin(adminRouter)
		})
	})

	if deps.Static != nil {
		router.Handle("/*", spaHandler(deps.Static))
	}
	server.router = router
	return server, nil
}

// Handler returns the fully configured router.
func (s *Server) Handler() http.Handler { return s.router }

type apiError struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Details   map[string]string `json:"details,omitempty"`
	RequestID string            `json:"requestId,omitempty"`
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, apiError{Code: code, Message: message, RequestID: middleware.GetReqID(r.Context())})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (s *Server) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			s.writeError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Open TX Carpool from Telegram to continue.")
			return
		}
		user, err := s.deps.Accounts.UserBySession(r.Context(), cookie.Value)
		if err != nil {
			s.writeError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Your session has expired. Open the Mini App again.")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
	})
}

func (s *Server) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := currentUser(r)
		if user.Role != "admin" || user.TelegramID != s.deps.AdminTelegramID {
			s.writeError(w, r, http.StatusForbidden, "ADMIN_REQUIRED", "Administrator access is required.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func currentUser(r *http.Request) model.User {
	user, _ := r.Context().Value(userContextKey).(model.User)
	return user
}

func (s *Server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				s.deps.Logger.Error("HTTP panic", "request_id", middleware.GetReqID(r.Context()), "method", r.Method, "path", r.URL.Path, "panic", recovered, "stack", string(debug.Stack()))
				s.writeError(w, r, http.StatusInternalServerError, "INTERNAL", "The request could not be completed.")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://telegram.org; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors https://web.telegram.org https://*.telegram.org")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(wrapper, r)
		path := r.URL.Path
		if strings.HasPrefix(path, "/api/v1/webhooks/bepusdt/") {
			path = "/api/v1/webhooks/bepusdt/[redacted]"
		}
		s.deps.Logger.Info("HTTP request", "request_id", middleware.GetReqID(r.Context()), "method", r.Method, "path", path, "status", wrapper.Status(), "duration_ms", time.Since(start).Milliseconds())
	})
}

func spaHandler(content fs.FS) http.Handler {
	files := http.FileServer(http.FS(content))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if file, err := content.Open(path); err == nil {
			if err := file.Close(); err != nil {
				http.Error(w, "failed to close file", http.StatusInternalServerError)
				return
			}
			files.ServeHTTP(w, r)
			return
		}
		index, err := fs.ReadFile(content, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(index)
	})
}
