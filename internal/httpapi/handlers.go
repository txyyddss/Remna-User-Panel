package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/txyyddss/Remna-User-Panel/internal/accounts"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

type userResponse struct {
	ID               string     `json:"id"`
	TelegramID       string     `json:"telegramId"`
	FirstName        string     `json:"firstName"`
	LastName         string     `json:"lastName"`
	TelegramUsername string     `json:"telegramUsername"`
	Username         *string    `json:"username"`
	Role             string     `json:"role"`
	OnboardingState  string     `json:"onboardingState"`
	GroupJoined      bool       `json:"groupJoined"`
	ChannelJoined    bool       `json:"channelJoined"`
	PolicyAcceptedAt *time.Time `json:"policyAcceptedAt"`
	RecoveryReason   string     `json:"recoveryReason"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func mapUser(user model.User) userResponse {
	return userResponse{
		ID: user.ID, TelegramID: strconv.FormatInt(user.TelegramID, 10), FirstName: user.TelegramFirstName,
		LastName: user.TelegramLastName, TelegramUsername: user.TelegramUsername, Username: user.Username, Role: user.Role,
		OnboardingState: user.OnboardingState, GroupJoined: user.GroupJoined, ChannelJoined: user.ChannelJoined,
		PolicyAcceptedAt: user.PolicyAcceptedAt, RecoveryReason: user.RecoveryReason, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}
}

type authState struct {
	Authenticated bool         `json:"authenticated"`
	User          userResponse `json:"user"`
}

func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) {
	var request struct {
		InitData string `json:"initData"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "initData is required.")
		return
	}
	user, token, expiresAt, err := s.deps.Accounts.Authenticate(r.Context(), request.InitData)
	if err != nil {
		if errors.Is(err, accounts.ErrUpstreamUnavailable) {
			s.writeError(w, r, http.StatusServiceUnavailable, "REMNAWAVE_UNAVAILABLE", "Account verification is temporarily unavailable. Please retry.")
			return
		}
		s.deps.Logger.Warn("Telegram Mini App authentication rejected", "request_id", middlewareRequestID(r), "error", err)
		s.writeError(w, r, http.StatusUnauthorized, "INVALID_TELEGRAM_DATA", "Telegram authentication could not be verified.")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", Expires: expiresAt, MaxAge: int(s.deps.SessionTTL.Seconds()),
		HttpOnly: true, Secure: s.deps.SecureCookies, SameSite: http.SameSiteLaxMode})
	writeJSON(w, http.StatusOK, authState{Authenticated: true, User: mapUser(user)})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, authState{Authenticated: true, User: mapUser(currentUser(r))})
}

func (s *Server) createInvites(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	links, expiresAt, err := s.deps.Accounts.CreateInvites(r.Context(), user)
	if err != nil {
		s.writeError(w, r, http.StatusServiceUnavailable, "INVITES_UNAVAILABLE", "Join links are temporarily unavailable.")
		return
	}
	type invite struct {
		URL       string    `json:"url"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"group":   invite{URL: links["group"], ExpiresAt: expiresAt},
		"channel": invite{URL: links["channel"], ExpiresAt: expiresAt},
	})
}

func (s *Server) checkMembership(w http.ResponseWriter, r *http.Request) {
	user, err := s.deps.Accounts.CheckMembership(r.Context(), currentUser(r))
	if err != nil {
		s.writeError(w, r, http.StatusBadGateway, "MEMBERSHIP_CHECK_FAILED", "Telegram membership could not be checked.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"groupJoined": user.GroupJoined, "channelJoined": user.ChannelJoined,
		"complete": user.GroupJoined && user.ChannelJoined, "user": mapUser(user)})
}

func (s *Server) reserveUsername(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "A username is required.")
		return
	}
	user, err := s.deps.Accounts.ReserveUsername(r.Context(), currentUser(r), request.Username)
	if err != nil {
		switch {
		case errors.Is(err, accounts.ErrUsernameUnavailable):
			s.writeError(w, r, http.StatusConflict, "USERNAME_UNAVAILABLE", "That username is unavailable.")
		case errors.Is(err, accounts.ErrMembershipRequired):
			s.writeError(w, r, http.StatusConflict, "MEMBERSHIP_REQUIRED", "Join both Telegram spaces first.")
		default:
			s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_USERNAME", err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, authState{Authenticated: true, User: mapUser(user)})
}

func (s *Server) acceptAgreement(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Accepted bool `json:"accepted"`
	}
	if err := decodeJSON(w, r, &request); err != nil || !request.Accepted {
		s.writeError(w, r, http.StatusBadRequest, "AGREEMENT_REQUIRED", "You must accept the agreement to continue.")
		return
	}
	user, err := s.deps.Accounts.AcceptAgreement(r.Context(), currentUser(r), request.Accepted)
	if err != nil {
		status, code := http.StatusBadGateway, "REMNAWAVE_UNAVAILABLE"
		if errors.Is(err, accounts.ErrUsernameUnavailable) || errors.Is(err, database.ErrConflict) {
			status, code = http.StatusConflict, "ONBOARDING_CONFLICT"
		}
		s.writeError(w, r, status, code, "The account could not be completed yet. You can safely retry.")
		return
	}
	writeJSON(w, http.StatusOK, authState{Authenticated: true, User: mapUser(user)})
}

type dashboardResponse struct {
	User              userResponse      `json:"user"`
	Balance           model.Money       `json:"balance"`
	ActivePurchase    *model.Purchase   `json:"activePurchase"`
	QueuedPurchase    *model.Purchase   `json:"queuedPurchase"`
	Statistics        *model.Statistics `json:"statistics"`
	SubscriptionURL   *string           `json:"subscriptionUrl"`
	StatisticsStale   bool              `json:"statisticsStale"`
	StatisticsWarning string            `json:"statisticsWarning"`
	FetchedAt         time.Time         `json:"fetchedAt"`
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	dashboard, err := s.deps.Catalog.Dashboard(r.Context(), user)
	if err != nil {
		s.writeError(w, r, http.StatusBadGateway, "DASHBOARD_UNAVAILABLE", "Live account data is temporarily unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, dashboardResponse{User: mapUser(dashboard.User), Balance: dashboard.Balance, ActivePurchase: dashboard.ActivePurchase,
		QueuedPurchase: dashboard.QueuedPurchase, Statistics: dashboard.Statistics, SubscriptionURL: dashboard.SubscriptionURL,
		StatisticsStale: dashboard.StatisticsStale, StatisticsWarning: dashboard.StatisticsWarning, FetchedAt: dashboard.FetchedAt})
}

func (s *Server) catalog(w http.ResponseWriter, r *http.Request) {
	if !s.requireOnboarded(w, r, currentUser(r)) {
		return
	}
	catalog, err := s.deps.Catalog.Catalog(r.Context())
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "CATALOG_UNAVAILABLE", "The catalog could not be loaded.")
		return
	}
	writeJSON(w, http.StatusOK, catalog)
}

func (s *Server) purchase(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	var request struct {
		ComboID              string   `json:"comboId"`
		AddonSquadProductIDs []string `json:"addonSquadProductIds"`
		CouponGrantID        string   `json:"couponGrantId"`
	}
	if err := decodeJSON(w, r, &request); err != nil || request.ComboID == "" {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Select a combo to continue.")
		return
	}
	idempotencyKey, err := optionalOrGeneratedIdempotencyKey(w, r)
	if err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_IDEMPOTENCY_KEY", "Idempotency-Key must contain 1 to 128 characters.")
		return
	}
	purchase, err := s.deps.Catalog.PurchaseWithCoupon(r.Context(), user, request.ComboID, request.AddonSquadProductIDs, request.CouponGrantID, idempotencyKey)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrInsufficientBalance):
			s.writeError(w, r, http.StatusConflict, "INSUFFICIENT_BALANCE", "Your TXB balance is too low for this purchase.")
		case errors.Is(err, database.ErrNotFound):
			s.writeError(w, r, http.StatusNotFound, "CATALOG_ITEM_NOT_FOUND", "The selected catalog item is unavailable.")
		default:
			s.writeError(w, r, http.StatusConflict, "PURCHASE_FAILED", "The purchase could not be completed.")
		}
		return
	}
	writeJSON(w, http.StatusCreated, purchase)
}

func (s *Server) purchases(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Catalog.Purchases(r.Context(), currentUser(r).ID)
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "PURCHASES_UNAVAILABLE", "Purchase history could not be loaded.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) revokeSubscription(w http.ResponseWriter, r *http.Request) {
	url, err := s.deps.Catalog.RevokeSubscription(r.Context(), currentUser(r))
	if err != nil {
		s.writeError(w, r, http.StatusBadGateway, "REVOKE_FAILED", "The subscription could not be rotated.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"subscriptionUrl": url})
}

func (s *Server) balance(w http.ResponseWriter, r *http.Request) {
	balance, err := s.deps.Store.Balance(r.Context(), currentUser(r).ID)
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "BALANCE_UNAVAILABLE", "Balance could not be loaded.")
		return
	}
	methods, err := s.deps.Billing.Methods(r.Context())
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "PAYMENT_METHODS_UNAVAILABLE", "Payment methods could not be loaded.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"balance": balance, "paymentMethods": methods})
}

func (s *Server) ledger(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Store.ListLedger(r.Context(), currentUser(r).ID, 100)
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "LEDGER_UNAVAILABLE", "Ledger history could not be loaded.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) createPaymentOrder(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	var request struct {
		MethodID string `json:"methodId"`
		TXBMinor string `json:"txbMinor"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Provider and TXB amount are required.")
		return
	}
	amount, err := strconv.ParseInt(request.TXBMinor, 10, 64)
	if err != nil || !billing.CanonicalMethodID(request.MethodID) {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_AMOUNT", "TXB amount must be integer hundredths.")
		return
	}
	order, err := s.deps.Billing.CreateOrder(r.Context(), user, strings.ToLower(request.MethodID), amount)
	if err != nil {
		if errors.Is(err, billing.ErrInvalidOrder) {
			s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_PAYMENT_ORDER", "The provider or TXB amount is invalid.")
		} else if errors.Is(err, billing.ErrProviderDisabled) {
			s.writeError(w, r, http.StatusConflict, "PROVIDER_DISABLED", "This payment provider is not available.")
		} else if errors.Is(err, database.ErrPaymentCapacity) {
			w.Header().Set("Retry-After", "30")
			s.writeError(w, r, http.StatusConflict, "PAYMENT_CAPACITY", "Too many unsettled payment orders. Retry after an existing order settles.")
		} else {
			s.writeError(w, r, http.StatusBadGateway, "PAYMENT_CREATE_FAILED", "The payment order could not be created.")
		}
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (s *Server) cancelPaymentOrder(w http.ResponseWriter, r *http.Request) {
	order, err := s.deps.Billing.Cancel(r.Context(), chiURLParam(r, "id"), currentUser(r).ID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			s.writeError(w, r, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment order not found.")
		case errors.Is(err, database.ErrConflict):
			s.writeError(w, r, http.StatusConflict, "PAYMENT_NOT_CANCELLABLE", "This payment can no longer be cancelled.")
		default:
			s.writeError(w, r, http.StatusInternalServerError, "PAYMENT_CANCEL_FAILED", "The payment could not be cancelled.")
		}
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *Server) paymentOrder(w http.ResponseWriter, r *http.Request) {
	order, err := s.deps.Billing.OrderForUser(r.Context(), chiURLParam(r, "id"), currentUser(r).ID)
	if err != nil {
		s.writeError(w, r, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment order not found.")
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *Server) requireOnboarded(w http.ResponseWriter, r *http.Request, user model.User) bool {
	if user.OnboardingState == "complete" {
		return true
	}
	s.writeError(w, r, http.StatusConflict, "ONBOARDING_REQUIRED", "Complete onboarding to use this feature.")
	return false
}

func chiURLParam(r *http.Request, name string) string {
	return strings.TrimSpace(chi.URLParam(r, name))
}
