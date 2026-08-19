package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func (s *Server) adminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.deps.Store.ListUsers(r.Context(), 200)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	type synchronization struct {
		Status string `json:"status"`
	}
	type adminUser struct {
		User            userResponse    `json:"user"`
		Balance         model.Money     `json:"balance"`
		CreatedAt       time.Time       `json:"createdAt"`
		Synchronization synchronization `json:"synchronization"`
	}
	items := make([]adminUser, 0, len(users))
	for _, user := range users {
		balance, balanceErr := s.deps.Store.Balance(r.Context(), user.ID)
		if balanceErr != nil {
			s.adminFailure(w, r, balanceErr)
			return
		}
		status := "not_provisioned"
		if user.RemnaUserID != nil {
			status = "synchronized"
		}
		items = append(items, adminUser{User: mapUser(user), Balance: balance, CreatedAt: user.CreatedAt, Synchronization: synchronization{Status: status}})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) adminUser(w http.ResponseWriter, r *http.Request) {
	userID := chiURLParam(r, "id")
	detail, err := s.deps.AdminUsers.UserDetail(r.Context(), userID)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	blocks, err := s.deps.ConnectionDrops.List(r.Context(), userID)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	response := mapAdminUserDetail(detail)
	response.IPBlocks = blocks
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) adminUnblockIP(w http.ResponseWriter, r *http.Request) {
	key, ok := s.requireIdempotencyKey(w, r)
	if !ok {
		return
	}
	receipt, err := s.deps.ConnectionDrops.Unblock(r.Context(), currentUser(r).ID,
		chiURLParam(r, "userId"), chiURLParam(r, "blockId"), key)
	if err != nil {
		s.writeIPBlockError(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, receipt)
}

func (s *Server) adminBalanceAdjustment(w http.ResponseWriter, r *http.Request) {
	var request struct {
		DeltaTXBMinor string `json:"deltaTxbMinor"`
		Reason        string `json:"reason"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_ADJUSTMENT", "Amount and reason are required.")
		return
	}
	delta, err := strconv.ParseInt(request.DeltaTXBMinor, 10, 64)
	if err != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_AMOUNT", "Adjustment must be integer hundredths of TXB.")
		return
	}
	entry, err := s.deps.Admin.AdjustBalance(r.Context(), currentUser(r).ID, chiURLParam(r, "id"), delta, request.Reason)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}

func (s *Server) adminEntitlements(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Store.ListAllPurchases(r.Context(), 200)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	type adminEntitlement struct {
		model.Purchase
		UserID string `json:"userId"`
	}
	response := make([]adminEntitlement, 0, len(items))
	for _, item := range items {
		response = append(response, adminEntitlement{Purchase: item, UserID: item.UserID})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (s *Server) adminCancelEntitlement(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_CANCELLATION", "A reason is required.")
		return
	}
	purchase, err := s.deps.Admin.CancelEntitlement(r.Context(), currentUser(r).ID, chiURLParam(r, "id"), request.Reason)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, purchase)
}
