package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/accounts"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/requestauth"
)

type userResponse struct {
	ID                string     `json:"id"`
	TelegramID        string     `json:"telegramId"`
	FirstName         string     `json:"firstName"`
	LastName          string     `json:"lastName"`
	TelegramUsername  string     `json:"telegramUsername"`
	Username          *string    `json:"username"`
	Role              string     `json:"role"`
	OnboardingState   string     `json:"onboardingState"`
	GroupJoined       bool       `json:"groupJoined"`
	ChannelJoined     bool       `json:"channelJoined"`
	PolicyAcceptedAt  *time.Time `json:"policyAcceptedAt"`
	AgreementRevision int        `json:"agreementRevision"`
	RecoveryReason    string     `json:"recoveryReason"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type authState struct {
	Authenticated bool         `json:"authenticated"`
	User          userResponse `json:"user"`
}

func mapUser(user model.User) userResponse {
	return userResponse{
		ID: user.ID, TelegramID: strconv.FormatInt(user.TelegramID, 10), FirstName: user.TelegramFirstName,
		LastName: user.TelegramLastName, TelegramUsername: user.TelegramUsername, Username: user.Username, Role: user.Role,
		OnboardingState: user.OnboardingState, GroupJoined: user.GroupJoined, ChannelJoined: user.ChannelJoined,
		PolicyAcceptedAt: user.PolicyAcceptedAt, AgreementRevision: user.AgreementRevision, RecoveryReason: user.RecoveryReason,
		CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}
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
	clientKey, err := s.requests.ClientKey(token)
	if err != nil {
		s.deps.Logger.Error("derive browser request key", "request_id", middlewareRequestID(r), "error", err)
		s.writeError(w, r, http.StatusInternalServerError, "SESSION_CREATE_FAILED", "The secure session could not be created.")
		return
	}
	s.setSessionCookies(w, token, clientKey, expiresAt)
	writeJSON(w, http.StatusOK, authState{Authenticated: true, User: mapUser(user)})
}

func (s *Server) setSessionCookies(w http.ResponseWriter, token, clientKey string, expiresAt time.Time) {
	common := http.Cookie{Path: "/", Expires: expiresAt, MaxAge: int(s.deps.SessionTTL.Seconds()),
		Secure: s.deps.SecureCookies, SameSite: http.SameSiteLaxMode}
	session := common
	session.Name, session.Value, session.HttpOnly = sessionCookie, token, true
	http.SetCookie(w, &session)
	requestKey := common
	requestKey.Name, requestKey.Value = requestauth.ClientKeyCookie, clientKey
	http.SetCookie(w, &requestKey)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, authState{Authenticated: true, User: mapUser(currentUser(r))})
}
