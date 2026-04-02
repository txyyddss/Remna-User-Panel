package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/models"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
)

const (
	ipChangeStatusWaiting   = "WAITING"
	ipChangeStatusPending   = "PENDING"
	ipChangeStatusChanging  = "CHANGING"
	ipChangeStatusCompleted = "COMPLETED"
	ipChangeStatusRejected  = "REJECTED"

	ipChangeVotesNeeded   = 5
	ipChangeDeclinesLimit = 2
	ipChangeAPIUsername   = "API_AUTO"
)

var defaultIPChangeAllowedSquads = []string{
	"04d22a2e-1979-47b9-946b-8dbea5398811",
	"5899ea60-974e-4794-9f71-ed73c2f8b24c",
	"88a969cd-e313-440c-8ebd-4a53d8a79c3b",
	"225a9b69-8b7d-4c70-ab57-13f547d96f54",
	"ba55bf05-1fc8-4b8b-b4b2-b6f20715ef03",
	"e68a5b66-f684-434b-9298-65e3c3867237",
}

type IPChangeService struct{}

type IPChangeRequestRecord struct {
	ID           int64
	RequestKey   string
	UserID       sql.NullInt64
	Username     string
	ShortUUID    string
	Reason       string
	Status       string
	AgreeCount   int
	DeclineCount int
	MessageID    int
	RequestedAt  time.Time
	CompletedAt  *time.Time
	UpdatedAt    time.Time
}

type IPChangeLookupResponse struct {
	Count  int    `json:"count"`
	Status string `json:"status"`
}

type IPChangeSubmitResponse struct {
	Success bool `json:"success"`
}

type AdminIPChangeRequest struct {
	ID           int64      `json:"id"`
	RequestKey   string     `json:"request_key"`
	UserID       *int64     `json:"user_id,omitempty"`
	Username     string     `json:"username"`
	ShortUUID    string     `json:"short_uuid"`
	Reason       string     `json:"reason"`
	Status       string     `json:"status"`
	AgreeCount   int        `json:"agree_count"`
	DeclineCount int        `json:"decline_count"`
	MessageID    int        `json:"message_id"`
	MessageLink  string     `json:"message_link,omitempty"`
	RequestedAt  time.Time  `json:"requested_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type IPChangeAPIError struct {
	Status  int
	Message string
	Data    map[string]interface{}
}

func (e *IPChangeAPIError) Error() string {
	return e.Message
}

func NewIPChangeService() *IPChangeService {
	return &IPChangeService{}
}

func (s *IPChangeService) SubmitUserRequest(ctx context.Context, user *models.User, subscription, reason string) (*IPChangeSubmitResponse, error) {
	subscription = strings.TrimSpace(subscription)
	reason = strings.TrimSpace(reason)
	if subscription == "" || reason == "" {
		return nil, &IPChangeAPIError{Status: http.StatusBadRequest, Message: "subscription link and reason are required"}
	}

	rwUser, shortUUID, err := s.resolveSubscriptionUser(subscription)
	if err != nil {
		return nil, &IPChangeAPIError{Status: http.StatusBadRequest, Message: err.Error()}
	}
	rwUser.Username = fallbackString(rwUser.Username, shortUUID)
	if rwUser.Status != "ACTIVE" {
		return nil, &IPChangeAPIError{Status: http.StatusForbidden, Message: "subscription is not ACTIVE"}
	}
	if !s.hasAllowedSquad(rwUser.ActiveInternalSquads) {
		return nil, &IPChangeAPIError{Status: http.StatusForbidden, Message: "invalid product/squad for this service"}
	}

	if active, err := s.getActiveRequest(ctx); err != nil {
		return nil, err
	} else if active != nil {
		data := map[string]interface{}{}
		if messageLink := s.messageLink(active.MessageID); messageLink != "" {
			data["message_link"] = messageLink
		}
		return nil, &IPChangeAPIError{
			Status:  http.StatusTooManyRequests,
			Message: "there is already a pending IP change request",
			Data:    data,
		}
	}

	requestKey := "req_" + fallbackString(rwUser.Username, shortUUID)
	if cooldownErr := s.ensureCooldown(ctx, requestKey); cooldownErr != nil {
		return nil, cooldownErr
	}

	messageID, err := s.sendVoteMessage(ctx, rwUser.Username, reason, false)
	if err != nil {
		slog.Error("ip-change: failed to send Telegram vote message", "username", rwUser.Username, "error", err)
		messageID = 0
	}

	requestID, err := s.upsertRequest(ctx, IPChangeRequestRecord{
		RequestKey:   requestKey,
		UserID:       nullableInt64(user.ID),
		Username:     rwUser.Username,
		ShortUUID:    shortUUID,
		Reason:       reason,
		Status:       ipChangeStatusPending,
		AgreeCount:   0,
		DeclineCount: 0,
		MessageID:    messageID,
		RequestedAt:  time.Now(),
		CompletedAt:  nil,
		UpdatedAt:    time.Now(),
	})
	if err != nil {
		return nil, err
	}

	if err := s.clearVotes(ctx, requestID); err != nil {
		slog.Warn("ip-change: failed to clear previous votes", "request_id", requestID, "error", err)
	}

	return &IPChangeSubmitResponse{Success: true}, nil
}

func (s *IPChangeService) Lookup(ctx context.Context) (*IPChangeLookupResponse, error) {
	row, err := s.getLookupTarget(ctx)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return &IPChangeLookupResponse{
			Count:  0,
			Status: ipChangeStatusWaiting,
		}, nil
	}
	return &IPChangeLookupResponse{
		Count:  row.AgreeCount,
		Status: row.Status,
	}, nil
}

func (s *IPChangeService) ListAdminRequests(ctx context.Context, limit, offset int) ([]AdminIPChangeRequest, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	if err := database.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM ip_change_requests").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count ip change requests: %w", err)
	}

	rows, err := database.DB().QueryContext(ctx,
		`SELECT id, request_key, user_id, username, short_uuid, reason, status, agree_count, decline_count, message_id, requested_at, completed_at, updated_at
		 FROM ip_change_requests
		 ORDER BY CASE status WHEN ? THEN 0 WHEN ? THEN 1 ELSE 2 END, updated_at DESC, id DESC
		 LIMIT ? OFFSET ?`,
		ipChangeStatusChanging, ipChangeStatusPending, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list ip change requests: %w", err)
	}
	defer rows.Close()

	var requests []AdminIPChangeRequest
	for rows.Next() {
		req, err := scanIPChangeRequest(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan ip change request: %w", err)
		}
		requests = append(requests, toAdminIPChangeRequest(req, s.messageLink(req.MessageID)))
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate ip change requests: %w", err)
	}
	if requests == nil {
		requests = []AdminIPChangeRequest{}
	}
	return requests, total, nil
}

func (s *IPChangeService) AddAPIRequest(ctx context.Context, reason string) (string, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "API automatic trigger"
	}

	if active, err := s.getActiveRequest(ctx); err != nil {
		return "", err
	} else if active != nil {
		return "", &IPChangeAPIError{
			Status:  http.StatusConflict,
			Message: "there is already a pending IP change request",
			Data: map[string]interface{}{
				"existing_status":   active.Status,
				"existing_username": active.Username,
			},
		}
	}

	messageID, err := s.sendVoteMessage(ctx, ipChangeAPIUsername, reason, true)
	if err != nil {
		slog.Error("ip-change: failed to send API vote message", "error", err)
		messageID = 0
	}

	requestID, err := s.upsertRequest(ctx, IPChangeRequestRecord{
		RequestKey:   "req_" + ipChangeAPIUsername,
		Username:     ipChangeAPIUsername,
		ShortUUID:    "api",
		Reason:       reason,
		Status:       ipChangeStatusPending,
		AgreeCount:   0,
		DeclineCount: 0,
		MessageID:    messageID,
		RequestedAt:  time.Now(),
		CompletedAt:  nil,
		UpdatedAt:    time.Now(),
	})
	if err != nil {
		return "", err
	}

	if err := s.clearVotes(ctx, requestID); err != nil {
		slog.Warn("ip-change: failed to clear API votes", "request_id", requestID, "error", err)
	}

	return "req_" + ipChangeAPIUsername, nil
}

func (s *IPChangeService) MarkSwapCompleted(ctx context.Context) error {
	req, err := s.getChangingRequest(ctx)
	if err != nil {
		return err
	}
	if req == nil {
		return &IPChangeAPIError{Status: http.StatusNotFound, Message: "no pending tasks"}
	}

	return s.markSwapCompletedRecord(ctx, req)
}

func (s *IPChangeService) ProcessVote(ctx context.Context, action, username string, voterTelegramID int64) (*IPChangeRequestRecord, error) {
	req, err := s.getActiveRequestByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, &IPChangeAPIError{Status: http.StatusNotFound, Message: "request not found or expired"}
	}

	if req.Status != ipChangeStatusPending {
		return req, nil
	}

	switch action {
	case "agree", "decline":
	default:
		return nil, &IPChangeAPIError{Status: http.StatusBadRequest, Message: "unsupported vote action"}
	}

	if _, err := database.DB().ExecContext(ctx,
		`INSERT INTO ip_change_votes (request_id, voter_telegram_id, action, created_at)
		 VALUES (?, ?, ?, ?)`,
		req.ID, voterTelegramID, action, time.Now(),
	); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, &IPChangeAPIError{Status: http.StatusConflict, Message: "you have already voted"}
		}
		return nil, fmt.Errorf("record vote: %w", err)
	}

	switch action {
	case "agree":
		req.AgreeCount++
		if req.AgreeCount >= ipChangeVotesNeeded {
			req.Status = ipChangeStatusChanging
		}
	case "decline":
		req.DeclineCount++
		if req.DeclineCount >= ipChangeDeclinesLimit {
			req.Status = ipChangeStatusRejected
		}
	}
	req.UpdatedAt = time.Now()

	if _, err := database.DB().ExecContext(ctx,
		`UPDATE ip_change_requests
		 SET status = ?, agree_count = ?, decline_count = ?, updated_at = ?
		 WHERE id = ?`,
		req.Status, req.AgreeCount, req.DeclineCount, req.UpdatedAt, req.ID,
	); err != nil {
		return nil, fmt.Errorf("update vote counters: %w", err)
	}

	return req, nil
}

func (s *IPChangeService) buildMessageText(req *IPChangeRequestRecord) string {
	return s.buildMessageTextWithOverride(req, "")
}

func (s *IPChangeService) buildMessageTextWithOverride(req *IPChangeRequestRecord, statusOverride string) string {
	requestedAt := req.RequestedAt.In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05")
	statusLine := fmt.Sprintf("Status: Pending (%d/%d)", req.AgreeCount, ipChangeVotesNeeded)

	if statusOverride != "" {
		statusLine = statusOverride
	} else {
		switch req.Status {
		case ipChangeStatusChanging:
			statusLine = fmt.Sprintf("Status: Changing (%d/%d)", ipChangeVotesNeeded, ipChangeVotesNeeded)
		case ipChangeStatusCompleted:
			statusLine = "Status: IP swapped successfully"
		case ipChangeStatusRejected:
			statusLine = "Status: Rejected"
		}
	}

	return fmt.Sprintf(
		"<b>IP Change Request</b>\n\n<b>User:</b> <code>%s</code>\n<b>Submitted At:</b> %s\n<b>Reason:</b>\n<blockquote>%s</blockquote>\n\n<b>%s</b>",
		req.Username,
		requestedAt,
		escapeTelegramHTML(req.Reason),
		statusLine,
	)
}

func (s *IPChangeService) BuildTelegramMessageText(req *IPChangeRequestRecord) string {
	return s.buildMessageText(req)
}

func (s *IPChangeService) buildInlineKeyboard(req *IPChangeRequestRecord) map[string]interface{} {
	if req.Status != ipChangeStatusPending {
		return nil
	}

	return map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{
				{
					"text":          fmt.Sprintf("Agree (%d)", req.AgreeCount),
					"callback_data": "agree:" + req.Username,
				},
				{
					"text":          fmt.Sprintf("Decline (%d)", req.DeclineCount),
					"callback_data": "decline:" + req.Username,
				},
			},
		},
	}
}

func (s *IPChangeService) AdminApproveRequest(ctx context.Context, requestID int64) (*AdminIPChangeRequest, error) {
	req, err := s.getRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, &IPChangeAPIError{Status: http.StatusNotFound, Message: "request not found"}
	}
	if req.Status != ipChangeStatusPending {
		return nil, &IPChangeAPIError{Status: http.StatusBadRequest, Message: "only pending requests can be approved"}
	}

	req.Status = ipChangeStatusChanging
	req.AgreeCount = ipChangeVotesNeeded
	req.UpdatedAt = time.Now()

	if err := s.persistRequestState(ctx, req); err != nil {
		return nil, err
	}
	s.syncTelegramRequestState(ctx, req, "")

	result := toAdminIPChangeRequest(req, s.messageLink(req.MessageID))
	return &result, nil
}

func (s *IPChangeService) AdminDeclineRequest(ctx context.Context, requestID int64) (*AdminIPChangeRequest, error) {
	req, err := s.getRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, &IPChangeAPIError{Status: http.StatusNotFound, Message: "request not found"}
	}
	if req.Status != ipChangeStatusPending {
		return nil, &IPChangeAPIError{Status: http.StatusBadRequest, Message: "only pending requests can be declined"}
	}

	req.Status = ipChangeStatusRejected
	req.DeclineCount = ipChangeDeclinesLimit
	req.UpdatedAt = time.Now()

	if err := s.persistRequestState(ctx, req); err != nil {
		return nil, err
	}
	s.syncTelegramRequestState(ctx, req, "")

	result := toAdminIPChangeRequest(req, s.messageLink(req.MessageID))
	return &result, nil
}

func (s *IPChangeService) AdminCompleteRequest(ctx context.Context, requestID int64) (*AdminIPChangeRequest, error) {
	req, err := s.getRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, &IPChangeAPIError{Status: http.StatusNotFound, Message: "request not found"}
	}
	if req.Status != ipChangeStatusChanging {
		return nil, &IPChangeAPIError{Status: http.StatusBadRequest, Message: "only changing requests can be completed"}
	}

	if err := s.markSwapCompletedRecord(ctx, req); err != nil {
		return nil, err
	}
	refreshed, err := s.getRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if refreshed == nil {
		return nil, &IPChangeAPIError{Status: http.StatusNotFound, Message: "request not found"}
	}
	result := toAdminIPChangeRequest(refreshed, s.messageLink(refreshed.MessageID))
	return &result, nil
}

func (s *IPChangeService) AdminDeleteRequest(ctx context.Context, requestID int64) error {
	req, err := s.getRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req == nil {
		return &IPChangeAPIError{Status: http.StatusNotFound, Message: "request not found"}
	}
	if req.Status == ipChangeStatusCompleted {
		return &IPChangeAPIError{Status: http.StatusBadRequest, Message: "completed requests cannot be deleted"}
	}

	s.syncTelegramRequestState(ctx, req, "Status: Deleted by admin")

	if _, err := database.DB().ExecContext(ctx, "DELETE FROM ip_change_requests WHERE id = ?", requestID); err != nil {
		return fmt.Errorf("delete ip change request: %w", err)
	}
	return nil
}

func (s *IPChangeService) resolveSubscriptionUser(subscription string) (*remnawave.UserData, string, error) {
	shortUUID := extractIPChangeShortUUID(subscription)
	if shortUUID == "" {
		return nil, "", fmt.Errorf("invalid subscription URL format")
	}

	cfg := config.Get()
	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	rwUser, err := rwClient.GetUserByShortUUID(shortUUID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch user info")
	}
	if rwUser == nil {
		return nil, "", fmt.Errorf("user not found")
	}
	return rwUser, shortUUID, nil
}

func (s *IPChangeService) hasAllowedSquad(squads []remnawave.Squad) bool {
	allowed := config.Get().IPChange.AllowedSquadUUIDs
	if len(allowed) == 0 {
		allowed = defaultIPChangeAllowedSquads
	}

	allowedSet := make(map[string]struct{}, len(allowed))
	for _, squadUUID := range allowed {
		allowedSet[squadUUID] = struct{}{}
	}
	for _, squad := range squads {
		if _, ok := allowedSet[squad.UUID]; ok {
			return true
		}
	}
	return false
}

func (s *IPChangeService) ensureCooldown(ctx context.Context, requestKey string) error {
	req, err := s.getRequestByKey(ctx, requestKey)
	if err != nil {
		return err
	}
	if req == nil || req.Status != ipChangeStatusCompleted {
		return nil
	}

	cooldownHours := config.Get().IPChange.CooldownHours
	if cooldownHours <= 0 {
		cooldownHours = 6
	}

	completionTime := req.RequestedAt
	if req.CompletedAt != nil {
		completionTime = *req.CompletedAt
	}
	diff := time.Since(completionTime)
	if diff < time.Duration(cooldownHours)*time.Hour {
		return &IPChangeAPIError{
			Status:  http.StatusForbidden,
			Message: fmt.Sprintf("cooldown active. please wait %d hours between IP swaps", cooldownHours),
		}
	}
	return nil
}

func (s *IPChangeService) sendVoteMessage(ctx context.Context, username, reason string, fromAPI bool) (int, error) {
	cfg := config.Get()
	if cfg.Telegram.BotToken == "" || cfg.Telegram.GroupID == 0 {
		return 0, fmt.Errorf("telegram bot token or group id not configured")
	}

	title := "IP Change Request"
	if fromAPI {
		title = "IP Change Request (API)"
	}
	now := time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05")
	text := fmt.Sprintf(
		"<b>%s</b>\n\n<b>User:</b> <code>%s</code>\n<b>Submitted At:</b> %s\n<b>Reason:</b>\n<blockquote>%s</blockquote>\n\n<b>Status: Pending (0/%d)</b>",
		title,
		username,
		now,
		escapeTelegramHTML(reason),
		ipChangeVotesNeeded,
	)
	replyMarkup := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{
				{"text": "Agree (0)", "callback_data": "agree:" + username},
				{"text": "Decline (0)", "callback_data": "decline:" + username},
			},
		},
	}

	var payload struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		Result      struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
	}

	if err := s.doTelegramJSON(ctx, "sendMessage", map[string]interface{}{
		"chat_id":      cfg.Telegram.GroupID,
		"text":         text,
		"parse_mode":   "HTML",
		"reply_markup": replyMarkup,
	}, &payload); err != nil {
		return 0, err
	}
	if !payload.OK {
		return 0, fmt.Errorf(payload.Description)
	}
	return payload.Result.MessageID, nil
}

func (s *IPChangeService) editTelegramMessage(ctx context.Context, messageID int, text string, replyMarkup map[string]interface{}) error {
	cfg := config.Get()
	if cfg.Telegram.BotToken == "" || cfg.Telegram.GroupID == 0 || messageID <= 0 {
		return nil
	}

	body := map[string]interface{}{
		"chat_id":    cfg.Telegram.GroupID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if replyMarkup != nil {
		body["reply_markup"] = replyMarkup
	}

	var payload struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := s.doTelegramJSON(ctx, "editMessageText", body, &payload); err != nil {
		return err
	}
	if !payload.OK {
		return fmt.Errorf(payload.Description)
	}
	return nil
}

func (s *IPChangeService) markSwapCompletedRecord(ctx context.Context, req *IPChangeRequestRecord) error {
	now := time.Now()
	req.Status = ipChangeStatusCompleted
	req.AgreeCount = 0
	req.CompletedAt = &now
	req.UpdatedAt = now

	if err := s.persistRequestState(ctx, req); err != nil {
		return err
	}
	s.syncTelegramRequestState(ctx, req, "")
	return nil
}

func (s *IPChangeService) doTelegramJSON(ctx context.Context, method string, body interface{}, out interface{}) error {
	rawBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+config.Get().Telegram.BotToken+"/"+method, bytes.NewReader(rawBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode telegram %s response: %w", method, err)
	}
	return nil
}

func (s *IPChangeService) getActiveRequest(ctx context.Context) (*IPChangeRequestRecord, error) {
	row := database.DB().QueryRowContext(ctx,
		`SELECT id, request_key, user_id, username, short_uuid, reason, status, agree_count, decline_count, message_id, requested_at, completed_at, updated_at
		 FROM ip_change_requests
		 WHERE status IN (?, ?)
		 ORDER BY CASE status WHEN ? THEN 0 ELSE 1 END, requested_at ASC
		 LIMIT 1`,
		ipChangeStatusPending, ipChangeStatusChanging, ipChangeStatusChanging,
	)
	req, err := scanIPChangeRequest(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query active request: %w", err)
	}
	return req, nil
}

func (s *IPChangeService) getRequestByID(ctx context.Context, requestID int64) (*IPChangeRequestRecord, error) {
	row := database.DB().QueryRowContext(ctx,
		`SELECT id, request_key, user_id, username, short_uuid, reason, status, agree_count, decline_count, message_id, requested_at, completed_at, updated_at
		 FROM ip_change_requests
		 WHERE id = ?`,
		requestID,
	)
	req, err := scanIPChangeRequest(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query request by id: %w", err)
	}
	return req, nil
}

func (s *IPChangeService) getChangingRequest(ctx context.Context) (*IPChangeRequestRecord, error) {
	row := database.DB().QueryRowContext(ctx,
		`SELECT id, request_key, user_id, username, short_uuid, reason, status, agree_count, decline_count, message_id, requested_at, completed_at, updated_at
		 FROM ip_change_requests
		 WHERE status = ?
		 ORDER BY updated_at ASC
		 LIMIT 1`,
		ipChangeStatusChanging,
	)
	req, err := scanIPChangeRequest(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query changing request: %w", err)
	}
	return req, nil
}

func (s *IPChangeService) getLookupTarget(ctx context.Context) (*IPChangeRequestRecord, error) {
	row := database.DB().QueryRowContext(ctx,
		`SELECT id, request_key, user_id, username, short_uuid, reason, status, agree_count, decline_count, message_id, requested_at, completed_at, updated_at
		 FROM ip_change_requests
		 WHERE status = ?
		 ORDER BY updated_at ASC
		 LIMIT 1`,
		ipChangeStatusChanging,
	)
	req, err := scanIPChangeRequest(row)
	if err == nil {
		return req, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("query lookup changing request: %w", err)
	}

	row = database.DB().QueryRowContext(ctx,
		`SELECT id, request_key, user_id, username, short_uuid, reason, status, agree_count, decline_count, message_id, requested_at, completed_at, updated_at
		 FROM ip_change_requests
		 WHERE status = ? AND agree_count >= 1
		 ORDER BY updated_at ASC
		 LIMIT 1`,
		ipChangeStatusPending,
	)
	req, err = scanIPChangeRequest(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query lookup pending request: %w", err)
	}
	return req, nil
}

func (s *IPChangeService) getRequestByKey(ctx context.Context, requestKey string) (*IPChangeRequestRecord, error) {
	row := database.DB().QueryRowContext(ctx,
		`SELECT id, request_key, user_id, username, short_uuid, reason, status, agree_count, decline_count, message_id, requested_at, completed_at, updated_at
		 FROM ip_change_requests
		 WHERE request_key = ?`,
		requestKey,
	)
	req, err := scanIPChangeRequest(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query request by key: %w", err)
	}
	return req, nil
}

func (s *IPChangeService) getActiveRequestByUsername(ctx context.Context, username string) (*IPChangeRequestRecord, error) {
	row := database.DB().QueryRowContext(ctx,
		`SELECT id, request_key, user_id, username, short_uuid, reason, status, agree_count, decline_count, message_id, requested_at, completed_at, updated_at
		 FROM ip_change_requests
		 WHERE username = ? AND status IN (?, ?)
		 ORDER BY requested_at DESC
		 LIMIT 1`,
		username, ipChangeStatusPending, ipChangeStatusChanging,
	)
	req, err := scanIPChangeRequest(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query request by username: %w", err)
	}
	return req, nil
}

func (s *IPChangeService) upsertRequest(ctx context.Context, req IPChangeRequestRecord) (int64, error) {
	if _, err := database.DB().ExecContext(ctx,
		`INSERT INTO ip_change_requests (request_key, user_id, username, short_uuid, reason, status, agree_count, decline_count, message_id, requested_at, completed_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(request_key) DO UPDATE SET
		   user_id = excluded.user_id,
		   username = excluded.username,
		   short_uuid = excluded.short_uuid,
		   reason = excluded.reason,
		   status = excluded.status,
		   agree_count = excluded.agree_count,
		   decline_count = excluded.decline_count,
		   message_id = excluded.message_id,
		   requested_at = excluded.requested_at,
		   completed_at = excluded.completed_at,
		   updated_at = excluded.updated_at`,
		req.RequestKey, nullableValue(req.UserID), req.Username, req.ShortUUID, req.Reason, req.Status, req.AgreeCount, req.DeclineCount, req.MessageID, req.RequestedAt, nullableTimeValue(req.CompletedAt), req.UpdatedAt,
	); err != nil {
		return 0, fmt.Errorf("upsert ip change request: %w", err)
	}

	row := database.DB().QueryRowContext(ctx, "SELECT id FROM ip_change_requests WHERE request_key = ?", req.RequestKey)
	var requestID int64
	if err := row.Scan(&requestID); err != nil {
		return 0, fmt.Errorf("load upserted request id: %w", err)
	}
	return requestID, nil
}

func (s *IPChangeService) clearVotes(ctx context.Context, requestID int64) error {
	_, err := database.DB().ExecContext(ctx, "DELETE FROM ip_change_votes WHERE request_id = ?", requestID)
	return err
}

func (s *IPChangeService) persistRequestState(ctx context.Context, req *IPChangeRequestRecord) error {
	if _, err := database.DB().ExecContext(ctx,
		`UPDATE ip_change_requests
		 SET status = ?, agree_count = ?, decline_count = ?, completed_at = ?, updated_at = ?
		 WHERE id = ?`,
		req.Status, req.AgreeCount, req.DeclineCount, nullableTimeValue(req.CompletedAt), req.UpdatedAt, req.ID,
	); err != nil {
		return fmt.Errorf("update ip change request: %w", err)
	}
	return nil
}

func (s *IPChangeService) syncTelegramRequestState(ctx context.Context, req *IPChangeRequestRecord, statusOverride string) {
	if req.MessageID <= 0 {
		return
	}

	replyMarkup := emptyTelegramInlineKeyboard()
	if req.Status == ipChangeStatusPending && statusOverride == "" {
		replyMarkup = s.buildInlineKeyboard(req)
	}

	if err := s.editTelegramMessage(ctx, req.MessageID, s.buildMessageTextWithOverride(req, statusOverride), replyMarkup); err != nil {
		slog.Warn("ip-change: failed to sync Telegram request state", "request_id", req.ID, "error", err)
	}
}

func (s *IPChangeService) messageLink(messageID int) string {
	if messageID <= 0 {
		return ""
	}

	chatID := fmt.Sprintf("%d", config.Get().Telegram.GroupID)
	if chatID == "0" {
		return ""
	}
	if strings.HasPrefix(chatID, "-100") {
		chatID = chatID[4:]
	} else {
		chatID = strings.TrimPrefix(chatID, "-")
	}
	if chatID == "" {
		return ""
	}
	return fmt.Sprintf("https://t.me/c/%s/%d", chatID, messageID)
}

func toAdminIPChangeRequest(req *IPChangeRequestRecord, messageLink string) AdminIPChangeRequest {
	var userID *int64
	if req.UserID.Valid {
		userID = &req.UserID.Int64
	}

	return AdminIPChangeRequest{
		ID:           req.ID,
		RequestKey:   req.RequestKey,
		UserID:       userID,
		Username:     req.Username,
		ShortUUID:    req.ShortUUID,
		Reason:       req.Reason,
		Status:       req.Status,
		AgreeCount:   req.AgreeCount,
		DeclineCount: req.DeclineCount,
		MessageID:    req.MessageID,
		MessageLink:  messageLink,
		RequestedAt:  req.RequestedAt,
		CompletedAt:  req.CompletedAt,
		UpdatedAt:    req.UpdatedAt,
	}
}

func emptyTelegramInlineKeyboard() map[string]interface{} {
	return map[string]interface{}{
		"inline_keyboard": [][]map[string]string{},
	}
}

func scanIPChangeRequest(scanner interface {
	Scan(dest ...interface{}) error
}) (*IPChangeRequestRecord, error) {
	var req IPChangeRequestRecord
	var userID sql.NullInt64
	var completedAt sql.NullTime
	if err := scanner.Scan(
		&req.ID,
		&req.RequestKey,
		&userID,
		&req.Username,
		&req.ShortUUID,
		&req.Reason,
		&req.Status,
		&req.AgreeCount,
		&req.DeclineCount,
		&req.MessageID,
		&req.RequestedAt,
		&completedAt,
		&req.UpdatedAt,
	); err != nil {
		return nil, err
	}

	req.UserID = userID
	if completedAt.Valid {
		req.CompletedAt = &completedAt.Time
	}
	return &req, nil
}

func extractIPChangeShortUUID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.ContainsAny(raw, "/?#") {
		return raw
	}

	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '/' || r == '?' || r == '#'
	})
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func escapeTelegramHTML(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(value)
}

func fallbackString(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

func nullableInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}

func nullableValue(value sql.NullInt64) interface{} {
	if value.Valid {
		return value.Int64
	}
	return nil
}

func nullableTimeValue(value *time.Time) interface{} {
	if value != nil {
		return *value
	}
	return nil
}
