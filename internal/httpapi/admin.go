package httpapi

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/backup"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

func (s *Server) mountAdmin(router chi.Router) {
	s.mountCommunityAdmin(router)
	router.Get("/emby-accounts", s.adminEmbyAccounts)
	router.Post("/emby-accounts/{id}/retry", s.retryAdminEmbyAccount)
	router.Get("/settings", s.adminSettings)
	router.Post("/settings", s.adminCreateSetting)
	router.Put("/settings/{key}", s.adminUpdateSetting)
	router.Get("/combos", s.adminCombos)
	router.Post("/combos", s.adminCreateCombo)
	router.Put("/combos/{id}", s.adminUpdateCombo)
	router.Delete("/combos/{id}", s.adminDeleteCombo)
	router.Get("/combos/{id}/statistics", s.adminComboStatistics)
	router.Get("/squad-products", s.adminSquadProducts)
	router.Post("/squad-products", s.adminCreateSquadProduct)
	router.Put("/squad-products/{id}", s.adminUpdateSquadProduct)
	router.Get("/squad-products/{id}/nodes", s.adminSquadNodes)
	router.Put("/squad-products/{id}/nodes", s.adminUpdateSquadNodes)
	router.Get("/squad-products/{id}/statistics", s.adminSquadStatistics)
	router.Post("/squad-products/import", s.adminImportSquads)
	router.Get("/users", s.adminUsers)
	router.Get("/users/{id}", s.adminUser)
	router.Post("/users/{id}/balance-adjustments", s.adminBalanceAdjustment)
	router.Get("/entitlements", s.adminEntitlements)
	router.Post("/entitlements/{id}/cancel", s.adminCancelEntitlement)
	router.Get("/payments", s.adminPayments)
	router.Post("/payments/{id}/refund", s.adminRefundPayment)
	router.Get("/refunds", s.adminRefunds)
	router.Get("/backups", s.adminBackups)
	router.Post("/backups", s.adminCreateBackup)
	router.Delete("/backups/{id}", s.adminDeleteBackup)
	router.Get("/jobs", s.adminJobs)
	router.Post("/jobs/{id}/retry", s.adminRetryJob)
	router.Delete("/jobs/{id}", s.adminDeleteJob)
	router.Get("/audit-events", s.adminAuditEvents)
	router.Get("/onboarding/content/{kind}", s.adminOnboardingBundle)
	router.Put("/onboarding/content/{kind}/draft", s.adminSaveOnboardingDraft)
	router.Post("/onboarding/content/{kind}/publish", s.adminPublishOnboarding)
}

func (s *Server) adminSettings(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Admin.Settings(r.Context())
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) adminCreateSetting(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := decodeJSON(w, r, &request); err != nil || request.Key == "" {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_SETTING", "A known setting key and value are required.")
		return
	}
	if err := s.deps.Admin.PutSetting(r.Context(), currentUser(r).ID, request.Key, request.Value); err != nil {
		s.adminFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) adminUpdateSetting(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Value string `json:"value"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_SETTING", "A setting value is required.")
		return
	}
	if err := s.deps.Admin.PutSetting(r.Context(), currentUser(r).ID, chiURLParam(r, "key"), request.Value); err != nil {
		s.adminFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) adminCombos(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Store.ListCombos(r.Context(), false)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type comboRequest struct {
	Name                    string   `json:"name"`
	Description             string   `json:"description"`
	PriceTXBMinor           string   `json:"priceTxbMinor"`
	ValidityDays            int      `json:"validityDays"`
	TrafficLimitBytes       string   `json:"trafficLimitBytes"`
	ResetStrategy           string   `json:"resetStrategy"`
	Active                  bool     `json:"active"`
	SquadProductIDs         []string `json:"squadProductIds"`
	IncludedSquadIDs        []string `json:"includedSquadIds"`
	RolloverMinRemainingBPS int      `json:"rolloverMinRemainingBps"`
	RolloverMaxTXBMinor     string   `json:"rolloverMaxTxbMinor"`
}

func (s *Server) adminCreateCombo(w http.ResponseWriter, r *http.Request) {
	s.adminSaveCombo(w, r, "")
}

func (s *Server) adminUpdateCombo(w http.ResponseWriter, r *http.Request) {
	s.adminSaveCombo(w, r, chiURLParam(r, "id"))
}

func (s *Server) adminSaveCombo(w http.ResponseWriter, r *http.Request, id string) {
	var request comboRequest
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_COMBO", "Combo fields are invalid.")
		return
	}
	price, priceErr := strconv.ParseInt(request.PriceTXBMinor, 10, 64)
	traffic, trafficErr := strconv.ParseInt(request.TrafficLimitBytes, 10, 64)
	rolloverMax, rolloverErr := strconv.ParseInt(request.RolloverMaxTXBMinor, 10, 64)
	if request.RolloverMaxTXBMinor == "" {
		rolloverMax, rolloverErr = 0, nil
	}
	if priceErr != nil || trafficErr != nil || rolloverErr != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_COMBO", "Price and traffic must be decimal integer strings.")
		return
	}
	squadIDs := request.SquadProductIDs
	if len(squadIDs) == 0 {
		squadIDs = request.IncludedSquadIDs
	}
	combo, err := s.deps.Admin.SaveCombo(r.Context(), currentUser(r).ID, database.ComboInput{ID: id, Name: request.Name,
		Description: request.Description, PriceTXBMinor: price, ValidityDays: request.ValidityDays, TrafficLimitBytes: traffic,
		ResetStrategy: request.ResetStrategy, Active: request.Active, SquadProductIDs: squadIDs,
		RolloverMinRemainingBPS: request.RolloverMinRemainingBPS, RolloverMaxTXBMinor: rolloverMax})
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	status := http.StatusOK
	if id == "" {
		status = http.StatusCreated
	}
	writeJSON(w, status, combo)
}

func (s *Server) adminDeleteCombo(w http.ResponseWriter, r *http.Request) {
	if err := s.deps.Admin.DeleteCombo(r.Context(), currentUser(r).ID, chiURLParam(r, "id")); err != nil {
		s.adminFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) adminComboStatistics(w http.ResponseWriter, r *http.Request) {
	from, to, bucket, location, ok := s.statisticsWindow(w, r)
	if !ok {
		return
	}
	statistics, err := s.deps.Store.ComboStatistics(r.Context(), chiURLParam(r, "id"), from, to, bucket, location)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, statistics)
}

func (s *Server) adminSquadProducts(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Admin.Squads(r.Context())
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type squadProductRequest struct {
	RemnaSquadUUID string `json:"remnaSquadUuid"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	PriceTXBMinor  string `json:"priceTxbMinor"`
	Visible        bool   `json:"visible"`
}

func (s *Server) adminCreateSquadProduct(w http.ResponseWriter, r *http.Request) {
	s.adminSaveSquadProduct(w, r, "")
}

func (s *Server) adminUpdateSquadProduct(w http.ResponseWriter, r *http.Request) {
	s.adminSaveSquadProduct(w, r, chiURLParam(r, "id"))
}

func (s *Server) adminSaveSquadProduct(w http.ResponseWriter, r *http.Request, id string) {
	var request squadProductRequest
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_SQUAD_PRODUCT", "Squad product fields are invalid.")
		return
	}
	price, err := strconv.ParseInt(request.PriceTXBMinor, 10, 64)
	if err != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_PRICE", "Price must be integer hundredths of TXB.")
		return
	}
	if request.RemnaSquadUUID == "" {
		request.RemnaSquadUUID = id
	}
	product, err := s.deps.Admin.SaveSquadProduct(r.Context(), currentUser(r).ID, database.SquadProductInput{ID: id,
		RemnaSquadUUID: request.RemnaSquadUUID, Name: request.Name, Description: request.Description, PriceTXBMinor: price,
		Visible: request.Visible, UpstreamPresent: true})
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	status := http.StatusOK
	if id == "" {
		status = http.StatusCreated
	}
	writeJSON(w, status, product)
}

func (s *Server) adminImportSquads(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Admin.ImportSquads(r.Context(), currentUser(r).ID)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) adminSquadNodes(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Admin.SquadNodes(r.Context(), chiURLParam(r, "id"))
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) adminUpdateSquadNodes(w http.ResponseWriter, r *http.Request) {
	var request struct {
		NodeUUIDs []string `json:"nodeUuids"`
	}
	if err := decodeJSON(w, r, &request); err != nil || len(request.NodeUUIDs) > 500 {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_NODE_SELECTION", "Select valid Remnawave nodes.")
		return
	}
	items, err := s.deps.Admin.UpdateSquadNodes(r.Context(), currentUser(r).ID, chiURLParam(r, "id"), request.NodeUUIDs)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) adminSquadStatistics(w http.ResponseWriter, r *http.Request) {
	from, to, bucket, location, ok := s.statisticsWindow(w, r)
	if !ok {
		return
	}
	statistics, err := s.deps.Store.SquadStatistics(r.Context(), chiURLParam(r, "id"), from, to, bucket, location)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, statistics)
}

func (s *Server) statisticsWindow(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, string, *time.Location, bool) {
	zone := strings.TrimSpace(r.URL.Query().Get("timeZone"))
	if zone == "" {
		zone = defaultActivityTimezone
	}
	location, err := time.LoadLocation(zone)
	if err != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_TIMEZONE", "Use a valid IANA timezone.")
		return time.Time{}, time.Time{}, "", nil, false
	}
	today := time.Now().In(location)
	toDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, location)
	fromDate := toDate.AddDate(0, 0, -29)
	if value := strings.TrimSpace(r.URL.Query().Get("from")); value != "" {
		fromDate, err = time.ParseInLocation(time.DateOnly, value, location)
		if err != nil {
			s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_DATE_RANGE", "Dates must use YYYY-MM-DD.")
			return time.Time{}, time.Time{}, "", nil, false
		}
	}
	if value := strings.TrimSpace(r.URL.Query().Get("to")); value != "" {
		toDate, err = time.ParseInLocation(time.DateOnly, value, location)
		if err != nil {
			s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_DATE_RANGE", "Dates must use YYYY-MM-DD.")
			return time.Time{}, time.Time{}, "", nil, false
		}
	}
	if toDate.Before(fromDate) || toDate.Sub(fromDate) > 731*24*time.Hour {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_DATE_RANGE", "Choose a range of at most two years.")
		return time.Time{}, time.Time{}, "", nil, false
	}
	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		bucket = "daily"
	}
	if bucket != "daily" && bucket != "weekly" {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_BUCKET", "Bucket must be daily or weekly.")
		return time.Time{}, time.Time{}, "", nil, false
	}
	return fromDate.UTC(), toDate.AddDate(0, 0, 1).UTC(), bucket, location, true
}

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
	user, err := s.deps.Store.UserByID(r.Context(), chiURLParam(r, "id"))
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapUser(user))
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

func (s *Server) adminPayments(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Store.ListPaymentOrders(r.Context(), "", 200)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	type adminPayment struct {
		model.PaymentOrder
		UserID string `json:"userId"`
	}
	response := make([]adminPayment, 0, len(items))
	for _, item := range items {
		response = append(response, adminPayment{PaymentOrder: item, UserID: item.UserID})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (s *Server) adminRefundPayment(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_REFUND", "A refund reason is required.")
		return
	}
	order, err := s.deps.Admin.Refund(r.Context(), currentUser(r).ID, chiURLParam(r, "id"), request.Reason)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *Server) adminRefunds(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Store.ListRefunds(r.Context(), 200)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) adminBackups(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Store.ListBackupRuns(r.Context(), 100)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	response := make([]backupRunResponse, 0, len(items))
	for _, item := range items {
		response = append(response, mapBackupRun(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (s *Server) adminCreateBackup(w http.ResponseWriter, r *http.Request) {
	run, err := s.deps.Admin.RunBackup(r.Context(), currentUser(r).ID)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapBackupRun(run))
}

func (s *Server) adminDeleteBackup(w http.ResponseWriter, r *http.Request) {
	if err := s.deps.Admin.DeleteBackup(r.Context(), currentUser(r).ID, chiURLParam(r, "id")); err != nil {
		s.adminFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type backupRunResponse struct {
	ID          string     `json:"id"`
	Path        string     `json:"path"`
	SizeBytes   string     `json:"sizeBytes"`
	Status      string     `json:"status"`
	Error       string     `json:"error"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

func mapBackupRun(run model.BackupRun) backupRunResponse {
	return backupRunResponse{ID: run.ID, Path: filepath.Base(run.Path), SizeBytes: strconv.FormatInt(run.SizeBytes, 10),
		Status: run.Status, Error: run.Error, CreatedAt: run.CreatedAt, CompletedAt: run.CompletedAt}
}

func (s *Server) adminJobs(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Store.ListOutboxJobs(r.Context(), 200)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) adminRetryJob(w http.ResponseWriter, r *http.Request) {
	if err := s.deps.Admin.RetryJob(r.Context(), currentUser(r).ID, chiURLParam(r, "id")); err != nil {
		s.adminFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) adminDeleteJob(w http.ResponseWriter, r *http.Request) {
	if err := s.deps.Store.DeleteOutboxJob(r.Context(), chiURLParam(r, "id")); err != nil {
		s.adminFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) adminAuditEvents(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Store.ListAuditEvents(r.Context(), 200)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) adminFailure(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := http.StatusUnprocessableEntity, "ADMIN_OPERATION_FAILED", err.Error()
	switch {
	case errors.Is(err, database.ErrNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "The requested record was not found."
	case errors.Is(err, database.ErrConflict), errors.Is(err, backup.ErrRestoreConflict):
		status, code, message = http.StatusConflict, "CONFLICT", "The operation conflicts with current state."
	case strings.Contains(strings.ToLower(err.Error()), "remnawave") || strings.Contains(strings.ToLower(err.Error()), "provider"):
		status, code, message = http.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "The upstream operation failed."
	}
	s.writeError(w, r, status, code, message)
}
