package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/middleware"
	telegramapi "github.com/user/remna-user-panel/internal/telegram"
)

type miniAppJoinSession struct {
	UserID            int64
	ChannelVerifiedAt *time.Time
	GroupInviteLink   string
	InviteCreatedAt   *time.Time
	GroupVerifiedAt   *time.Time
	UpdatedAt         time.Time
}

func (h *Handler) GetMiniAppAccessStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	channelJoined, groupJoined, inviteLink, err := h.resolveMiniAppAccessState(r.Context(), user.ID, user.TelegramID)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"user": map[string]interface{}{
			"id":            user.ID,
			"telegram_id":   user.TelegramID,
			"telegram_name": user.TelegramName,
			"is_admin":      user.IsAdmin,
		},
		"channel_joined": channelJoined,
		"group_joined":   groupJoined,
		"invite_link":    inviteLink,
		"channel_url":    cfg.Telegram.ChannelURL,
	})
}

func (h *Handler) VerifyMiniAppChannel(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()
	if cfg.Telegram.ChannelID == 0 || cfg.Telegram.ChannelURL == "" {
		middleware.WriteError(w, http.StatusInternalServerError, "channel join flow is not configured")
		return
	}
	if cfg.Telegram.GroupID == 0 {
		middleware.WriteError(w, http.StatusInternalServerError, "group join flow is not configured")
		return
	}

	channelStatus, err := telegramapi.GetChatMemberStatus(r.Context(), cfg.Telegram.BotToken, cfg.Telegram.ChannelID, user.TelegramID)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to verify channel membership")
		return
	}
	if !telegramapi.IsJoinedChatStatus(channelStatus) {
		middleware.WriteError(w, http.StatusForbidden, "join the required channel first")
		return
	}

	session, err := loadMiniAppJoinSession(r.Context(), user.ID)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to load join session")
		return
	}
	if session != nil && session.GroupInviteLink != "" {
		if err := telegramapi.RevokeChatInviteLink(r.Context(), cfg.Telegram.BotToken, cfg.Telegram.GroupID, session.GroupInviteLink); err != nil {
			// A stale invite should not block a fresh one.
		}
	}

	inviteLink, err := telegramapi.CreateChatInviteLink(r.Context(), cfg.Telegram.BotToken, cfg.Telegram.GroupID, 1)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to generate invite link")
		return
	}

	now := time.Now()
	if err := saveMiniAppJoinSession(r.Context(), miniAppJoinSession{
		UserID:            user.ID,
		ChannelVerifiedAt: &now,
		GroupInviteLink:   inviteLink,
		InviteCreatedAt:   &now,
		UpdatedAt:         now,
	}); err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to persist join session")
		return
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"channel_joined": true,
		"group_joined":   false,
		"invite_link":    inviteLink,
		"channel_url":    cfg.Telegram.ChannelURL,
	})
}

func (h *Handler) VerifyMiniAppGroup(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()
	if cfg.Telegram.GroupID == 0 {
		middleware.WriteError(w, http.StatusInternalServerError, "group join flow is not configured")
		return
	}

	groupStatus, err := telegramapi.GetChatMemberStatus(r.Context(), cfg.Telegram.BotToken, cfg.Telegram.GroupID, user.TelegramID)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to verify group membership")
		return
	}
	if !telegramapi.IsJoinedChatStatus(groupStatus) {
		middleware.WriteError(w, http.StatusForbidden, "join the required group first")
		return
	}

	session, err := loadMiniAppJoinSession(r.Context(), user.ID)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to load join session")
		return
	}

	now := time.Now()
	if session != nil && session.GroupInviteLink != "" {
		_ = telegramapi.RevokeChatInviteLink(r.Context(), cfg.Telegram.BotToken, cfg.Telegram.GroupID, session.GroupInviteLink)
	}

	if err := saveMiniAppJoinSession(r.Context(), miniAppJoinSession{
		UserID:          user.ID,
		GroupVerifiedAt: &now,
		UpdatedAt:       now,
	}); err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to finalize join session")
		return
	}

	channelJoined := cfg.Telegram.ChannelID == 0
	if cfg.Telegram.ChannelID != 0 {
		channelStatus, err := telegramapi.GetChatMemberStatus(r.Context(), cfg.Telegram.BotToken, cfg.Telegram.ChannelID, user.TelegramID)
		if err == nil {
			channelJoined = telegramapi.IsJoinedChatStatus(channelStatus)
		}
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"channel_joined": channelJoined,
		"group_joined":   true,
		"invite_link":    "",
		"channel_url":    cfg.Telegram.ChannelURL,
	})
}

func (h *Handler) resolveMiniAppAccessState(ctx context.Context, userID, telegramID int64) (bool, bool, string, error) {
	cfg := config.Get()

	groupJoined := false
	if cfg.Telegram.GroupID != 0 {
		status, err := telegramapi.GetChatMemberStatus(ctx, cfg.Telegram.BotToken, cfg.Telegram.GroupID, telegramID)
		if err != nil {
			return false, false, "", fmt.Errorf("failed to verify group membership")
		}
		groupJoined = telegramapi.IsJoinedChatStatus(status)
	}

	channelJoined := cfg.Telegram.ChannelID == 0
	if cfg.Telegram.ChannelID != 0 {
		status, err := telegramapi.GetChatMemberStatus(ctx, cfg.Telegram.BotToken, cfg.Telegram.ChannelID, telegramID)
		if err != nil {
			return false, false, "", fmt.Errorf("failed to verify channel membership")
		}
		channelJoined = telegramapi.IsJoinedChatStatus(status)
	}

	session, err := loadMiniAppJoinSession(ctx, userID)
	if err != nil {
		return false, false, "", err
	}

	inviteLink := ""
	if session != nil && !groupJoined {
		inviteLink = session.GroupInviteLink
	}
	return channelJoined, groupJoined, inviteLink, nil
}

func loadMiniAppJoinSession(ctx context.Context, userID int64) (*miniAppJoinSession, error) {
	row := database.DB().QueryRowContext(ctx,
		`SELECT user_id, channel_verified_at, group_invite_link, invite_created_at, group_verified_at, updated_at
		 FROM miniapp_join_sessions
		 WHERE user_id = ?`,
		userID,
	)

	var session miniAppJoinSession
	var channelVerifiedAt, inviteCreatedAt, groupVerifiedAt sql.NullTime
	err := row.Scan(&session.UserID, &channelVerifiedAt, &session.GroupInviteLink, &inviteCreatedAt, &groupVerifiedAt, &session.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if channelVerifiedAt.Valid {
		session.ChannelVerifiedAt = &channelVerifiedAt.Time
	}
	if inviteCreatedAt.Valid {
		session.InviteCreatedAt = &inviteCreatedAt.Time
	}
	if groupVerifiedAt.Valid {
		session.GroupVerifiedAt = &groupVerifiedAt.Time
	}
	return &session, nil
}

func saveMiniAppJoinSession(ctx context.Context, session miniAppJoinSession) error {
	_, err := database.DB().ExecContext(ctx,
		`INSERT INTO miniapp_join_sessions (user_id, channel_verified_at, group_invite_link, invite_created_at, group_verified_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET
		   channel_verified_at = COALESCE(excluded.channel_verified_at, miniapp_join_sessions.channel_verified_at),
		   group_invite_link = excluded.group_invite_link,
		   invite_created_at = excluded.invite_created_at,
		   group_verified_at = excluded.group_verified_at,
		   updated_at = excluded.updated_at`,
		session.UserID,
		nullableTime(session.ChannelVerifiedAt),
		session.GroupInviteLink,
		nullableTime(session.InviteCreatedAt),
		nullableTime(session.GroupVerifiedAt),
		session.UpdatedAt,
	)
	return err
}

func nullableTime(value *time.Time) interface{} {
	if value == nil {
		return nil
	}
	return *value
}
