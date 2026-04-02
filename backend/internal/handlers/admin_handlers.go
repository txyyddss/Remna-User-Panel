package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/models"
	"github.com/user/remna-user-panel/internal/sdk/jellyfin"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
	"github.com/user/remna-user-panel/internal/services"
)

const (
	adminQuery1 = "SELECT uuid, name, description, squad_uuid, traffic_gb, strategy, cycle, price_rmb, reset_price, active FROM combos WHERE active = 1"
	adminQuery2 = `INSERT INTO combos (uuid, name, description, squad_uuid, traffic_gb, strategy, cycle, price_rmb, reset_price, active, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	adminQuery3 = "SELECT COUNT(*) FROM subscriptions WHERE combo_uuid = ?"
	adminQuery4 = "SELECT COUNT(*) FROM orders WHERE metadata LIKE ?"
	adminQuery5 = "DELETE FROM combos WHERE uuid = ?"
	adminQuery6 = "UPDATE combos SET active = 0 WHERE uuid = ?"
	adminQuery7 = "SELECT uuid, name, description, squad_uuid, traffic_gb, strategy, cycle, price_rmb, reset_price, active FROM combos ORDER BY created_at DESC"
	adminQuery8 = "SELECT id, telegram_id, telegram_name, remnawave_uuid, jellyfin_user_id, credit, is_admin, created_at, updated_at FROM users WHERE telegram_name LIKE ? OR CAST(telegram_id AS TEXT) LIKE ? ORDER BY id DESC LIMIT ? OFFSET ?"
	adminQuery9 = "SELECT id, telegram_id, telegram_name, remnawave_uuid, jellyfin_user_id, credit, is_admin, created_at, updated_at FROM users ORDER BY id DESC LIMIT ? OFFSET ?"
	adminQuery10 = "SELECT COUNT(*) FROM users WHERE telegram_name LIKE ? OR CAST(telegram_id AS TEXT) LIKE ?"
	adminQuery11 = "SELECT COUNT(*) FROM users"
	adminQuery12 = "SELECT id, telegram_id, telegram_name, remnawave_uuid, jellyfin_user_id, credit, is_admin, created_at, updated_at FROM users WHERE id = ?"
	adminQuery13 = "SELECT combo_uuid, status, expires_at FROM subscriptions WHERE user_id = ? ORDER BY created_at DESC LIMIT 1"
	adminQuery14 = "SELECT jellyfin_user_id, username, parental_rating, expires_at FROM jellyfin_accounts WHERE user_id = ?"
	adminQuery15 = "SELECT credit FROM users WHERE id = ?"
	adminQuery16 = "UPDATE users SET remnawave_uuid = ?, updated_at = ? WHERE id = ?"
	adminQuery17 = "UPDATE users SET jellyfin_user_id = ?, updated_at = ? WHERE id = ?"
	adminQuery18 = "UPDATE users SET is_admin = ?, updated_at = ? WHERE id = ?"
	adminQuery19 = "SELECT remnawave_uuid FROM users WHERE id = ?"
	adminQuery20 = "UPDATE users SET remnawave_uuid = ?, updated_at = ? WHERE id = ?"
	adminQuery21 = "SELECT id FROM subscriptions WHERE user_id = ? ORDER BY created_at DESC LIMIT 1"
	adminQuery22 = "UPDATE subscriptions SET combo_uuid = ?, updated_at = ? WHERE id = ?"
	adminQuery23 = "UPDATE subscriptions SET status = ?, updated_at = ? WHERE id = ?"
	adminQuery24 = "UPDATE subscriptions SET expires_at = ?, updated_at = ? WHERE id = ?"
	adminQuery25 = "INSERT INTO subscriptions (user_id, combo_uuid, remnawave_uuid, status, expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	adminQuery26 = "SELECT squad_uuid FROM combos WHERE uuid = ?"
	adminQuery27 = "SELECT jellyfin_user_id FROM users WHERE id = ?"
	adminQuery28 = "UPDATE users SET jellyfin_user_id = ?, updated_at = ? WHERE id = ?"
	adminQuery29 = "SELECT id FROM jellyfin_accounts WHERE user_id = ?"
	adminQuery30 = "UPDATE jellyfin_accounts SET jellyfin_user_id = ? WHERE id = ?"
	adminQuery31 = "UPDATE jellyfin_accounts SET username = ? WHERE id = ?"
	adminQuery32 = "UPDATE jellyfin_accounts SET parental_rating = ? WHERE id = ?"
	adminQuery33 = "UPDATE jellyfin_accounts SET expires_at = ? WHERE id = ?"
	adminQuery34 = "INSERT INTO jellyfin_accounts (user_id, jellyfin_user_id, username, parental_rating, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?)"
)

func (h *Handler) ListCombos(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB().QueryContext(r.Context(), adminQuery1)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to list combos")
		return
	}
	defer rows.Close()

	var combos []models.Combo
	for rows.Next() {
		var c models.Combo
		if err := rows.Scan(&c.UUID, &c.Name, &c.Description, &c.SquadUUID, &c.TrafficGB, &c.Strategy, &c.Cycle, &c.PriceRMB, &c.ResetPrice, &c.Active); err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "failed to scan combo")
			return
		}
		combos = append(combos, c)
	}
	middleware.WriteSuccess(w, combos)
}

func (h *Handler) CreateCombo(w http.ResponseWriter, r *http.Request) {
	var combo models.Combo
	if err := middleware.DecodeJSON(r, &combo); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	combo.UUID = uuid.New().String()
	combo.Active = true
	combo.CreatedAt = time.Now()

	_, err := database.DB().ExecContext(r.Context(), 
		adminQuery2,
		combo.UUID, combo.Name, combo.Description, combo.SquadUUID, combo.TrafficGB,
		combo.Strategy, combo.Cycle, combo.PriceRMB, combo.ResetPrice, combo.Active, combo.CreatedAt,
	)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to create combo")
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, combo)
}

func (h *Handler) UpdateCombo(w http.ResponseWriter, r *http.Request) {
	comboUUID := chi.URLParam(r, "uuid")

	var updates map[string]interface{}
	if err := middleware.DecodeJSON(r, &updates); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// Build dynamic update
	allowed := map[string]bool{"name": true, "description": true, "squad_uuid": true, "traffic_gb": true, "strategy": true, "cycle": true, "price_rmb": true, "reset_price": true, "active": true}
	for key := range updates {
		if !allowed[key] {
			delete(updates, key)
		}
	}

	for key, val := range updates {
		database.DB().ExecContext(r.Context(), fmt.Sprintf("UPDATE combos SET %s = ? WHERE uuid = ?", key), val, comboUUID)
	}

	middleware.WriteSuccess(w, map[string]string{"uuid": comboUUID})
}

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	// Return safe copy (hide sensitive tokens)
	safeCfg := map[string]interface{}{
		"credit":    cfg.Credit,
		"ai":        map[string]interface{}{"enabled": cfg.AI.Enabled, "model": cfg.AI.Model, "message_batch_size": cfg.AI.MessageBatchSize, "credit_min": cfg.AI.CreditMin, "credit_max": cfg.AI.CreditMax, "leaderboard_interval": cfg.AI.LeaderboardInterval},
		"backup":    cfg.Backup,
		"ip_change": cfg.IPChange,
		"jellyfin":  map[string]interface{}{"monthly_price_rmb": cfg.Jellyfin.MonthlyPriceRMB},
	}
	middleware.WriteSuccess(w, safeCfg)
}

func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}
	if err := middleware.DecodeJSON(r, &updates); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	err := config.Update(func(cfg *config.Config) {
		data, _ := json.Marshal(updates)
		json.Unmarshal(data, cfg)
	})
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to update config")
		return
	}

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}

// GetInternalSquads returns internal squads from Remnawave for admin combo creation
func (h *Handler) GetInternalSquads(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	squads, err := rwClient.GetInternalSquads()
	if err != nil {
		slog.Error("admin: GetInternalSquads error", "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get internal squads: "+err.Error())
		return
	}
	if squads == nil {
		squads = []remnawave.Squad{}
	}
	middleware.WriteSuccess(w, squads)
}

// DeleteCombo deletes an unused combo, otherwise archives it.
func (h *Handler) DeleteCombo(w http.ResponseWriter, r *http.Request) {
	comboUUID := chi.URLParam(r, "uuid")

	var subscriptionRefs int
	_ = database.DB().QueryRowContext(r.Context(), adminQuery3, comboUUID).Scan(&subscriptionRefs)
	var orderRefs int
	_ = database.DB().QueryRowContext(r.Context(), adminQuery4, "%"+comboUUID+"%").Scan(&orderRefs)

	if subscriptionRefs == 0 && orderRefs == 0 {
		result, err := database.DB().ExecContext(r.Context(), adminQuery5, comboUUID)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "failed to delete combo")
			return
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			middleware.WriteError(w, http.StatusNotFound, "combo not found")
			return
		}
		middleware.WriteSuccess(w, map[string]string{"status": "deleted", "mode": "hard"})
		return
	}

	if _, err := database.DB().ExecContext(r.Context(), adminQuery6, comboUUID); err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to delete combo")
		return
	}
	middleware.WriteSuccess(w, map[string]string{"status": "deleted", "mode": "archived"})
}

// AdminListCombos lists all combos including inactive ones for admin
func (h *Handler) AdminListCombos(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB().QueryContext(r.Context(), adminQuery7)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to list combos")
		return
	}
	defer rows.Close()

	var combos []models.Combo
	for rows.Next() {
		var c models.Combo
		if err := rows.Scan(&c.UUID, &c.Name, &c.Description, &c.SquadUUID, &c.TrafficGB, &c.Strategy, &c.Cycle, &c.PriceRMB, &c.ResetPrice, &c.Active); err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "failed to scan combo")
			return
		}
		combos = append(combos, c)
	}
	if combos == nil {
		combos = []models.Combo{}
	}
	middleware.WriteSuccess(w, combos)
}

// --- Admin User Management ---

func (h *Handler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var rows *sql.Rows
	var err error

	if search != "" {
		searchPattern := "%" + search + "%"
		rows, err = database.DB().QueryContext(r.Context(), 
			adminQuery8,
			searchPattern, searchPattern, limit, offset,
		)
	} else {
		rows, err = database.DB().QueryContext(r.Context(), 
			adminQuery9,
			limit, offset,
		)
	}
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.TelegramID, &u.TelegramName, &u.RemnawaveUUID, &u.JellyfinUserID, &u.Credit, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "failed to scan user")
			return
		}
		users = append(users, u)
	}
	if users == nil {
		users = []models.User{}
	}

	// Get total count with the same filter that powers the current page.
	var total int
	if search != "" {
		searchPattern := "%" + search + "%"
		database.DB().QueryRowContext(r.Context(), 
			adminQuery10,
			searchPattern, searchPattern,
		).Scan(&total)
	} else {
		database.DB().QueryRowContext(r.Context(), adminQuery11).Scan(&total)
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"users": users,
		"total": total,
	})
}

func (h *Handler) AdminGetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var user models.User
	err = database.DB().QueryRowContext(r.Context(), 
		adminQuery12,
		userID,
	).Scan(&user.ID, &user.TelegramID, &user.TelegramName, &user.RemnawaveUUID, &user.JellyfinUserID, &user.Credit, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		middleware.WriteError(w, http.StatusInternalServerError, "failed to load user")
		return
	}

	var subscription map[string]interface{}
	var comboUUID, status string
	var expiresAt time.Time
	if err := database.DB().QueryRowContext(r.Context(), 
		adminQuery13,
		userID,
	).Scan(&comboUUID, &status, &expiresAt); err == nil {
		subscription = map[string]interface{}{
			"combo_uuid": comboUUID,
			"status":     status,
			"expires_at": expiresAt,
		}
	}

	var jellyfinAccount map[string]interface{}
	var jfUsername string
	var jfRating int
	var jfExpires time.Time
	if err := database.DB().QueryRowContext(r.Context(), 
		adminQuery14,
		userID,
	).Scan(&user.JellyfinUserID, &jfUsername, &jfRating, &jfExpires); err == nil {
		jellyfinAccount = map[string]interface{}{
			"jellyfin_user_id": user.JellyfinUserID,
			"username":         jfUsername,
			"parental_rating":  jfRating,
			"expires_at":       jfExpires,
		}
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"user":         user,
		"subscription": subscription,
		"jellyfin":     jellyfinAccount,
	})
}

func (h *Handler) AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req struct {
		Credit         *float64 `json:"credit"`
		RemnawaveUUID  *string  `json:"remnawave_uuid"`
		JellyfinUserID *string  `json:"jellyfin_user_id"`
		IsAdmin        *bool    `json:"is_admin"`
		Subscription   *struct {
			RemnawaveUUID *string `json:"remnawave_uuid"`
			ComboUUID     *string `json:"combo_uuid"`
			Status        *string `json:"status"`
			ExpiresAt     *string `json:"expires_at"`
		} `json:"subscription"`
		Jellyfin *struct {
			JellyfinUserID *string `json:"jellyfin_user_id"`
			Username       *string `json:"username"`
			ParentalRating *int    `json:"parental_rating"`
			ExpiresAt      *string `json:"expires_at"`
		} `json:"jellyfin"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Credit != nil {
		// Set absolute credit value
		var currentCredit float64
		database.DB().QueryRowContext(r.Context(), adminQuery15, userID).Scan(&currentCredit)
		diff := *req.Credit - currentCredit
		if diff != 0 {
			h.Credit.AddCredit(userID, diff, "admin adjustment")
		}
	}
	if req.RemnawaveUUID != nil {
		database.DB().ExecContext(r.Context(), adminQuery16, *req.RemnawaveUUID, time.Now(), userID)
	}
	if req.JellyfinUserID != nil {
		database.DB().ExecContext(r.Context(), adminQuery17, *req.JellyfinUserID, time.Now(), userID)
	}
	if req.IsAdmin != nil {
		adminVal := 0
		if *req.IsAdmin {
			adminVal = 1
		}
		database.DB().ExecContext(r.Context(), adminQuery18, adminVal, time.Now(), userID)
	}

	currentRemnawaveUUID := ""
	_ = database.DB().QueryRowContext(r.Context(), adminQuery19, userID).Scan(&currentRemnawaveUUID)
	if req.Subscription != nil {
		if req.Subscription.RemnawaveUUID != nil {
			currentRemnawaveUUID = *req.Subscription.RemnawaveUUID
			database.DB().ExecContext(r.Context(), adminQuery20, currentRemnawaveUUID, time.Now(), userID)
		}

		var expiresAt time.Time
		if req.Subscription.ExpiresAt != nil && *req.Subscription.ExpiresAt != "" {
			if parsed, err := time.Parse(time.RFC3339, *req.Subscription.ExpiresAt); err == nil {
				expiresAt = parsed
			}
		}

		var existingSubID int64
		err := database.DB().QueryRowContext(r.Context(), 
			adminQuery21,
			userID,
		).Scan(&existingSubID)
		if err == nil {
			if req.Subscription.ComboUUID != nil {
				database.DB().ExecContext(r.Context(), adminQuery22, *req.Subscription.ComboUUID, time.Now(), existingSubID)
			}
			if req.Subscription.Status != nil {
				database.DB().ExecContext(r.Context(), adminQuery23, *req.Subscription.Status, time.Now(), existingSubID)
			}
			if !expiresAt.IsZero() {
				database.DB().ExecContext(r.Context(), adminQuery24, expiresAt, time.Now(), existingSubID)
			}
		} else if req.Subscription.ComboUUID != nil && currentRemnawaveUUID != "" && !expiresAt.IsZero() {
			status := "active"
			if req.Subscription.Status != nil && *req.Subscription.Status != "" {
				status = *req.Subscription.Status
			}
			database.DB().ExecContext(r.Context(), 
				adminQuery25,
				userID, *req.Subscription.ComboUUID, currentRemnawaveUUID, status, expiresAt, time.Now(), time.Now(),
			)
		}

		if currentRemnawaveUUID != "" {
			cfg := config.Get()
			rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
			updateReq := remnawave.UpdateUserRequest{UUID: currentRemnawaveUUID}
			if req.Subscription.Status != nil {
				switch *req.Subscription.Status {
				case "active":
					updateReq.Status = "ACTIVE"
				case "disabled":
					updateReq.Status = "DISABLED"
				case "expired":
					updateReq.Status = "EXPIRED"
				}
			}
			if !expiresAt.IsZero() {
				updateReq.ExpireAt = expiresAt.Format(time.RFC3339)
			}
			if req.Subscription.ComboUUID != nil && *req.Subscription.ComboUUID != "" {
				var squadUUID string
				if err := database.DB().QueryRowContext(r.Context(), adminQuery26, *req.Subscription.ComboUUID).Scan(&squadUUID); err == nil && squadUUID != "" {
					updateReq.ActiveInternalSquads = []string{squadUUID}
				}
			}
			if updateReq.Status != "" || updateReq.ExpireAt != "" || len(updateReq.ActiveInternalSquads) > 0 {
				if _, err := rwClient.UpdateUser(updateReq); err != nil {
					slog.Error("admin: failed to update Remnawave user", "user_id", userID, "error", err)
				}
			}
		}
	}

	currentJellyfinUserID := ""
	_ = database.DB().QueryRowContext(r.Context(), adminQuery27, userID).Scan(&currentJellyfinUserID)
	if req.Jellyfin != nil {
		if req.Jellyfin.JellyfinUserID != nil {
			currentJellyfinUserID = *req.Jellyfin.JellyfinUserID
			database.DB().ExecContext(r.Context(), adminQuery28, currentJellyfinUserID, time.Now(), userID)
		}

		var expiresAt time.Time
		if req.Jellyfin.ExpiresAt != nil && *req.Jellyfin.ExpiresAt != "" {
			if parsed, err := time.Parse(time.RFC3339, *req.Jellyfin.ExpiresAt); err == nil {
				expiresAt = parsed
			}
		}

		var existingAccountID int64
		err := database.DB().QueryRowContext(r.Context(), adminQuery29, userID).Scan(&existingAccountID)
		if err == nil {
			if req.Jellyfin.JellyfinUserID != nil {
				database.DB().ExecContext(r.Context(), adminQuery30, *req.Jellyfin.JellyfinUserID, existingAccountID)
			}
			if req.Jellyfin.Username != nil {
				database.DB().ExecContext(r.Context(), adminQuery31, *req.Jellyfin.Username, existingAccountID)
			}
			if req.Jellyfin.ParentalRating != nil {
				database.DB().ExecContext(r.Context(), adminQuery32, *req.Jellyfin.ParentalRating, existingAccountID)
			}
			if !expiresAt.IsZero() {
				database.DB().ExecContext(r.Context(), adminQuery33, expiresAt, existingAccountID)
			}
		} else if currentJellyfinUserID != "" {
			username := ""
			if req.Jellyfin.Username != nil {
				username = *req.Jellyfin.Username
			}
			rating := 0
			if req.Jellyfin.ParentalRating != nil {
				rating = *req.Jellyfin.ParentalRating
			}
			if !expiresAt.IsZero() {
				database.DB().ExecContext(r.Context(), 
					adminQuery34,
					userID, currentJellyfinUserID, username, rating, expiresAt, time.Now(),
				)
			}
		}

		if currentJellyfinUserID != "" && req.Jellyfin.ParentalRating != nil {
			cfg := config.Get()
			jfClient := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)
			if err := jfClient.UpdateParentalRating(currentJellyfinUserID, *req.Jellyfin.ParentalRating); err != nil {
				slog.Error("admin: failed to update Jellyfin rating", "user_id", userID, "error", err)
			}
		}
	}

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}

func (h *Handler) AdminListOrders(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	orders, total, err := h.Payment.GetAdminOrders(services.OrderFilters{
		Search:        r.URL.Query().Get("search"),
		Status:        r.URL.Query().Get("status"),
		ServiceStatus: r.URL.Query().Get("service_status"),
		OrderType:     r.URL.Query().Get("order_type"),
		DateFrom:      r.URL.Query().Get("date_from"),
		DateTo:        r.URL.Query().Get("date_to"),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to load orders")
		return
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"orders": orders,
		"total":  total,
	})
}

func (h *Handler) AdminUpdateOrder(w http.ResponseWriter, r *http.Request) {
	admin := middleware.GetUser(r)
	orderUUID := chi.URLParam(r, "uuid")

	var updates services.AdminOrderUpdate
	if err := middleware.DecodeJSON(r, &updates); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	order, err := h.Payment.UpdateOrderByAdmin(orderUUID, admin.ID, updates)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	middleware.WriteSuccess(w, order)
}

func (h *Handler) AdminOrderAction(w http.ResponseWriter, r *http.Request) {
	admin := middleware.GetUser(r)
	orderUUID := chi.URLParam(r, "uuid")
	action := chi.URLParam(r, "action")

	var (
		order *models.OrderDetail
		err   error
	)

	switch action {
	case "apply-credit":
		order, err = h.Payment.ApplyCustomOrderCredit(orderUUID, admin.ID)
	case "resend-notice":
		order, err = h.Payment.ResendCustomOrderNotification(orderUUID, admin.ID)
	case "refund":
		order, err = h.Payment.RefundOrder(orderUUID, admin.ID)
	case "cancel":
		updateStatus := "cancelled"
		updateService := "cancelled"
		order, err = h.Payment.UpdateOrderByAdmin(orderUUID, admin.ID, services.AdminOrderUpdate{
			Status:        &updateStatus,
			ServiceStatus: &updateService,
		})
	default:
		middleware.WriteError(w, http.StatusBadRequest, "unknown action")
		return
	}

	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	middleware.WriteSuccess(w, order)
}
