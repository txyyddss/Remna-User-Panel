package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
)

func supportListHandler(settings config.Settings, pool *pgxpool.Pool, admin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var session sessionContext
		var ok bool
		if admin {
			session, ok = requireAdmin(w, r, settings, pool, false)
			if !ok {
				return
			}
		} else {
			session, ok = requireSession(w, r, settings, pool, false)
			if !ok {
				return
			}
		}
		allTickets := supportVisibleTickets(r.Context(), pool, readSettingList(r.Context(), pool, "SUPPORT_TICKETS"), session.User.UserID, admin)
		counts := supportCounts(allTickets)
		filtered := filterSupportTickets(allTickets, r.URL.Query())
		total := len(filtered)
		filtered = paginateSupportTickets(filtered, r.URL.Query())
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "tickets": filtered, "counts": counts, "total": total})
	}
}

func supportCreateHandler(settings config.Settings, pool *pgxpool.Pool, _admin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		var payload map[string]any
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		subject := strings.TrimSpace(fmt.Sprint(payload["subject"]))
		body := strings.TrimSpace(fmt.Sprint(payload["body"]))
		if subject == "" || body == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_ticket"})
			return
		}
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		id := nextListID(tickets)
		now := time.Now().Format(time.RFC3339)
		message := map[string]any{
			"message_id":       1,
			"ticket_id":        id,
			"body":             body,
			"is_admin":         false,
			"is_internal_note": false,
			"author_role":      "user",
			"user_id":          session.User.UserID,
			"created_at":       now,
		}
		ticket := map[string]any{
			"ticket_id":            id,
			"id":                   id,
			"user_id":              session.User.UserID,
			"status":               "awaiting_admin",
			"priority":             defaultString(payload["priority"], "normal"),
			"category":             defaultString(payload["category"], "general"),
			"subject":              subject,
			"body":                 body,
			"created_at":           now,
			"updated_at":           now,
			"last_message_at":      now,
			"last_message_preview": supportPreview(body),
			"message_count":        1,
			"unread_admin_count":   1,
			"unread_user_count":    0,
			"messages":             []any{message},
		}
		tickets = append(tickets, ticket)
		_ = writeSettingList(r.Context(), pool, "SUPPORT_TICKETS", tickets)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": supportTicketResponse(r.Context(), pool, ticket, false)})
	}
}

func supportDetailHandler(settings config.Settings, pool *pgxpool.Pool, admin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var session sessionContext
		var ok bool
		if admin {
			session, ok = requireAdmin(w, r, settings, pool, false)
			if !ok {
				return
			}
		} else {
			session, ok = requireSession(w, r, settings, pool, false)
			if !ok {
				return
			}
		}
		ticket, ok := findSettingItem(r.Context(), pool, "SUPPORT_TICKETS", chi.URLParam(r, "ticket_id"), "ticket_id")
		if !ok || (!admin && !supportTicketBelongsToUser(ticket, session.User.UserID)) {
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
			return
		}
		response := map[string]any{
			"ok":       true,
			"ticket":   supportTicketResponse(r.Context(), pool, ticket, admin),
			"messages": supportVisibleMessages(ticket, admin),
		}
		if admin {
			response["user_snapshot"] = supportUserSnapshot(r.Context(), pool, int64Value(ticket, "user_id"))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func supportMessageHandler(settings config.Settings, pool *pgxpool.Pool, admin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var session sessionContext
		var ok bool
		if admin {
			session, ok = requireAdmin(w, r, settings, pool, true)
		} else {
			session, ok = requireSession(w, r, settings, pool, true)
		}
		if !ok {
			return
		}
		var payload struct {
			Body           string `json:"body"`
			IsInternalNote bool   `json:"is_internal_note"`
		}
		if err := decodeJSONBody(r, &payload); err != nil || strings.TrimSpace(payload.Body) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_message"})
			return
		}
		id := chi.URLParam(r, "ticket_id")
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		for index := range tickets {
			if fmt.Sprint(tickets[index]["ticket_id"]) == id {
				if !admin && !supportTicketBelongsToUser(tickets[index], session.User.UserID) {
					writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
					return
				}
				messages := supportAllMessages(tickets[index])
				now := time.Now().Format(time.RFC3339)
				internalNote := admin && payload.IsInternalNote
				role := "user"
				if admin {
					role = "admin"
				}
				if internalNote {
					role = "internal"
				}
				message := map[string]any{
					"message_id":       len(messages) + 1,
					"ticket_id":        id,
					"body":             strings.TrimSpace(payload.Body),
					"is_admin":         admin,
					"is_internal_note": internalNote,
					"author_role":      role,
					"user_id":          session.User.UserID,
					"created_at":       now,
				}
				tickets[index]["messages"] = append(messages, message)
				tickets[index]["updated_at"] = now
				tickets[index]["message_count"] = len(messages) + 1
				if !internalNote {
					tickets[index]["last_message_at"] = now
					tickets[index]["last_message_preview"] = supportPreview(payload.Body)
				}
				switch {
				case internalNote:
				case admin:
					tickets[index]["status"] = "awaiting_user"
					tickets[index]["unread_user_count"] = int64Value(tickets[index], "unread_user_count") + 1
				default:
					tickets[index]["status"] = "awaiting_admin"
					tickets[index]["unread_admin_count"] = int64Value(tickets[index], "unread_admin_count") + 1
				}
				_ = writeSettingList(r.Context(), pool, "SUPPORT_TICKETS", tickets)
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": supportTicketResponse(r.Context(), pool, tickets[index], admin), "message": message})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
	}
}

func supportPatchHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		id := chi.URLParam(r, "ticket_id")
		var payload map[string]any
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		for index := range tickets {
			if fmt.Sprint(tickets[index]["ticket_id"]) == id || fmt.Sprint(tickets[index]["id"]) == id {
				for k, v := range payload {
					switch k {
					case "status":
						tickets[index][k] = normalizeSupportStatus(v)
					case "priority":
						tickets[index][k] = normalizeSupportPriority(v)
					case "category":
						tickets[index][k] = defaultString(v, "general")
					case "subject":
						if subject := strings.TrimSpace(fmt.Sprint(v)); subject != "" {
							tickets[index][k] = subject
						}
					default:
						tickets[index][k] = v
					}
				}
				tickets[index]["updated_at"] = time.Now().Format(time.RFC3339)
				_ = writeSettingList(r.Context(), pool, "SUPPORT_TICKETS", tickets)
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": supportTicketResponse(r.Context(), pool, tickets[index], true)})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
	}
}

func supportReadHandler(settings config.Settings, pool *pgxpool.Pool, admin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var session sessionContext
		var ok bool
		if admin {
			session, ok = requireAdmin(w, r, settings, pool, true)
		} else {
			session, ok = requireSession(w, r, settings, pool, true)
		}
		if !ok {
			return
		}
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		for index := range tickets {
			if fmt.Sprint(tickets[index]["ticket_id"]) == chi.URLParam(r, "ticket_id") || fmt.Sprint(tickets[index]["id"]) == chi.URLParam(r, "ticket_id") {
				if !admin && !supportTicketBelongsToUser(tickets[index], session.User.UserID) {
					writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
					return
				}
				if admin {
					tickets[index]["unread_admin_count"] = 0
				} else {
					tickets[index]["unread_user_count"] = 0
				}
				_ = writeSettingList(r.Context(), pool, "SUPPORT_TICKETS", tickets)
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": supportTicketResponse(r.Context(), pool, tickets[index], admin)})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
	}
}

func supportUnreadHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, false)
		if !ok {
			return
		}
		unread := 0
		for _, ticket := range readSettingList(r.Context(), pool, "SUPPORT_TICKETS") {
			if supportTicketBelongsToUser(ticket, session.User.UserID) {
				unread += int(int64Value(ticket, "unread_user_count"))
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "unread": unread})
	}
}

func adminSupportStatsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "stats": supportCounts(tickets), "counts": supportCounts(tickets)})
	}
}

// ---------------------------------------------------------------------------
// Support helper functions
// ---------------------------------------------------------------------------

func supportVisibleTickets(ctx context.Context, pool *pgxpool.Pool, tickets []map[string]any, userID int64, admin bool) []map[string]any {
	result := make([]map[string]any, 0, len(tickets))
	for _, ticket := range tickets {
		if !admin && !supportTicketBelongsToUser(ticket, userID) {
			continue
		}
		result = append(result, supportTicketResponse(ctx, pool, ticket, admin))
	}
	return result
}

func supportTicketResponse(ctx context.Context, pool *pgxpool.Pool, ticket map[string]any, admin bool) map[string]any {
	item := make(map[string]any, len(ticket)+2)
	for key, value := range ticket {
		if key == "messages" {
			continue
		}
		item[key] = value
	}
	item["status"] = normalizeSupportStatus(item["status"])
	item["priority"] = normalizeSupportPriority(item["priority"])
	item["category"] = defaultString(item["category"], "general")
	item["message_count"] = len(supportVisibleMessages(ticket, admin))
	if item["last_message_at"] == nil || fmt.Sprint(item["last_message_at"]) == "" {
		item["last_message_at"] = item["updated_at"]
	}
	if admin {
		item["user"] = supportTicketUser(ctx, pool, int64Value(ticket, "user_id"))
	} else {
		delete(item, "unread_admin_count")
	}
	return item
}

func supportTicketUser(ctx context.Context, pool *pgxpool.Pool, userID int64) map[string]any {
	user := map[string]any{"user_id": userID}
	if pool == nil || userID == 0 {
		return user
	}
	loaded, err := loadAdminUser(ctx, pool, strconv.FormatInt(userID, 10))
	if err != nil {
		return user
	}
	loaded["name"] = strings.TrimSpace(strings.Join([]string{stringValue(loaded, "first_name"), stringValue(loaded, "last_name")}, " "))
	if loaded["name"] == "" {
		loaded["name"] = firstNonEmpty(stringValue(loaded, "username"), stringValue(loaded, "email"), strconv.FormatInt(userID, 10))
	}
	return loaded
}

func supportUserSnapshot(ctx context.Context, pool *pgxpool.Pool, userID int64) map[string]any {
	user := supportTicketUser(ctx, pool, userID)
	name := firstNonEmpty(stringValue(user, "name"), stringValue(user, "username"), stringValue(user, "email"), strconv.FormatInt(userID, 10))
	return map[string]any{
		"name":         name,
		"tariff":       "-",
		"panel_status": "-",
		"remaining":    "-",
	}
}

func supportTicketBelongsToUser(ticket map[string]any, userID int64) bool {
	return userID != 0 && int64Value(ticket, "user_id") == userID
}

func supportAllMessages(ticket map[string]any) []any {
	switch messages := ticket["messages"].(type) {
	case []any:
		return messages
	case []map[string]any:
		result := make([]any, 0, len(messages))
		for _, message := range messages {
			result = append(result, message)
		}
		return result
	default:
		return []any{}
	}
}

func supportVisibleMessages(ticket map[string]any, admin bool) []any {
	messages := supportAllMessages(ticket)
	if admin {
		return messages
	}
	result := make([]any, 0, len(messages))
	for _, message := range messages {
		if mapped, ok := message.(map[string]any); ok && supportBoolValue(mapped, "is_internal_note") {
			continue
		}
		result = append(result, message)
	}
	return result
}

func filterSupportTickets(tickets []map[string]any, query map[string][]string) []map[string]any {
	status := strings.ToLower(strings.TrimSpace(firstQuery(query, "status")))
	priority := strings.ToLower(strings.TrimSpace(firstQuery(query, "priority")))
	category := strings.ToLower(strings.TrimSpace(firstQuery(query, "category")))
	search := strings.ToLower(strings.TrimSpace(firstQuery(query, "search")))
	result := make([]map[string]any, 0, len(tickets))
	for _, ticket := range tickets {
		ticketStatus := normalizeSupportStatus(ticket["status"])
		if status != "" && status != "all" {
			active := ticketStatus != "closed" && ticketStatus != "resolved"
			if status == "active" && !active {
				continue
			}
			if status != "active" && ticketStatus != status {
				continue
			}
		}
		if priority != "" && strings.ToLower(fmt.Sprint(ticket["priority"])) != priority {
			continue
		}
		if category != "" && strings.ToLower(fmt.Sprint(ticket["category"])) != category {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(fmt.Sprint(ticket)), search) {
			continue
		}
		result = append(result, ticket)
	}
	sortSupportTickets(result, firstQuery(query, "sort"))
	return result
}

func paginateSupportTickets(tickets []map[string]any, query map[string][]string) []map[string]any {
	offset, _ := strconv.Atoi(firstQuery(query, "offset"))
	limit, _ := strconv.Atoi(firstQuery(query, "limit"))
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset >= len(tickets) {
		return []map[string]any{}
	}
	end := offset + limit
	if end > len(tickets) {
		end = len(tickets)
	}
	return tickets[offset:end]
}

func sortSupportTickets(tickets []map[string]any, rawSort string) {
	sortKey := strings.ToLower(strings.TrimSpace(rawSort))
	sort.SliceStable(tickets, func(i, j int) bool {
		left := tickets[i]
		right := tickets[j]
		switch sortKey {
		case "created_asc":
			return supportTimeValue(left, "created_at").Before(supportTimeValue(right, "created_at"))
		case "created_desc":
			return supportTimeValue(left, "created_at").After(supportTimeValue(right, "created_at"))
		case "updated_asc":
			return supportTimeValue(left, "updated_at").Before(supportTimeValue(right, "updated_at"))
		default:
			leftUnread := int64Value(left, "unread_admin_count")
			rightUnread := int64Value(right, "unread_admin_count")
			if leftUnread != rightUnread {
				return leftUnread > rightUnread
			}
			leftPriority := supportPriorityRank(left["priority"])
			rightPriority := supportPriorityRank(right["priority"])
			if leftPriority != rightPriority {
				return leftPriority > rightPriority
			}
			return supportTimeValue(left, "updated_at").After(supportTimeValue(right, "updated_at"))
		}
	})
}

func supportTimeValue(ticket map[string]any, key string) time.Time {
	value := stringValue(ticket, key)
	if value == "" && key == "updated_at" {
		value = firstNonEmpty(stringValue(ticket, "last_message_at"), stringValue(ticket, "created_at"))
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed
	}
	return time.Time{}
}

func supportPriorityRank(value any) int {
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(value))) {
	case "urgent":
		return 4
	case "high":
		return 3
	case "normal":
		return 2
	case "low":
		return 1
	default:
		return 2
	}
}

func normalizeSupportStatus(value any) string {
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(value))) {
	case "closed", "resolved", "awaiting_admin", "awaiting_user", "open":
		return strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
	default:
		return "open"
	}
}

func normalizeSupportPriority(value any) string {
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(value))) {
	case "low", "normal", "high", "urgent":
		return strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
	default:
		return "normal"
	}
}

func supportBoolValue(m map[string]any, key string) bool {
	switch value := m[key].(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(strings.TrimSpace(value), "true") || strings.TrimSpace(value) == "1"
	default:
		return false
	}
}

func supportPreview(value string) string {
	clean := strings.Join(strings.Fields(value), " ")
	runes := []rune(clean)
	if len(runes) <= 160 {
		return clean
	}
	return string(runes[:160])
}

func defaultString(value any, fallback string) string {
	clean := strings.TrimSpace(fmt.Sprint(value))
	if clean == "" || clean == "<nil>" {
		return fallback
	}
	return clean
}

func firstQuery(query map[string][]string, key string) string {
	values := query[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func supportCounts(tickets []map[string]any) map[string]int {
	counts := map[string]int{}
	for _, ticket := range tickets {
		status := normalizeSupportStatus(ticket["status"])
		counts[status]++
	}
	if _, ok := counts["open"]; !ok {
		counts["open"] = 0
	}
	return counts
}
