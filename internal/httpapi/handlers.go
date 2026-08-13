package httpapi

import (
	"errors"
	"github.com/txyyddss/Remna-User-Panel/internal/accounts"
	"github.com/txyyddss/Remna-User-Panel/internal/catalog"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"net/http"
	"time"
)

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
		Revision     int      `json:"revision"`
		AgreementIDs []string `json:"agreementIds"`
	}
	if err := decodeJSON(w, r, &request); err != nil || request.Revision <= 0 || len(request.AgreementIDs) == 0 {
		s.writeError(w, r, http.StatusBadRequest, "AGREEMENT_REQUIRED", "Select every current agreement to continue.")
		return
	}
	user, err := s.deps.Accounts.AcceptAgreementRevision(r.Context(), currentUser(r), request.Revision, request.AgreementIDs)
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
		case errors.Is(err, catalog.ErrNoAccessibleNodes):
			s.writeError(w, r, http.StatusConflict, "NO_ACCESSIBLE_NODES", "The selected catalog item has no accessible nodes.")
		case errors.Is(err, database.ErrStockUnavailable):
			s.writeError(w, r, http.StatusConflict, "SQUAD_STOCK_UNAVAILABLE", "A selected squad is currently full.")
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
