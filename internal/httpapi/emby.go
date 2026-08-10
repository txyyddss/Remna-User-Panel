package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/emby"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

type embyAccountResponse struct {
	ID                 string    `json:"id"`
	Username           string    `json:"username"`
	Status             string    `json:"status"`
	MaxParentalRating  *int32    `json:"maxParentalRating"`
	DisabledLibraryIDs []string  `json:"disabledLibraryIds"`
	Retryable          bool      `json:"retryable"`
	ErrorMessage       string    `json:"errorMessage,omitempty"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type embyRatingResponse struct {
	Name  string `json:"name"`
	Value int32  `json:"value"`
}

type embyLibraryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func mapEmbyAccount(account emby.Account) embyAccountResponse {
	username := account.RemoteUsername
	if username == "" {
		username = account.CandidateUsername
	}
	if username == "" {
		username = account.BaseUsername
	}
	libraries := account.Preferences.DisabledLibraryIDs
	if libraries == nil {
		libraries = []string{}
	}
	return embyAccountResponse{ID: account.ID, Username: username, Status: account.Status,
		MaxParentalRating: account.Preferences.MaxParentalRating, DisabledLibraryIDs: libraries,
		Retryable: account.Retryable, ErrorMessage: account.LastError, UpdatedAt: account.UpdatedAt}
}

func (s *Server) embyAccount(w http.ResponseWriter, r *http.Request) {
	configured := true
	for _, key := range []string{"emby.base_url", "emby.api_token"} {
		if _, err := s.deps.Settings.Plaintext(r.Context(), key); err != nil {
			configured = false
		}
	}
	priceMinor := int64(0)
	if configured {
		var priceErr error
		priceMinor, priceErr = s.deps.EmbyPrice.EmbySetupPriceTXBMinor(r.Context())
		if priceErr != nil {
			configured = false
		}
	}
	if !configured {
		priceMinor = 0
	}
	var options emby.Options
	if configured {
		var err error
		options, err = s.deps.Emby.Options(r.Context())
		if err != nil {
			s.writeError(w, r, http.StatusBadGateway, "EMBY_UNAVAILABLE", "Emby choices are temporarily unavailable.")
			return
		}
	}
	ratings := make([]embyRatingResponse, 0, len(options.Ratings))
	for _, rating := range options.Ratings {
		ratings = append(ratings, embyRatingResponse{Name: rating.Name, Value: rating.Value})
	}
	libraries := make([]embyLibraryResponse, 0, len(options.Folders))
	for _, folder := range options.Folders {
		libraries = append(libraries, embyLibraryResponse{ID: folder.ID, Name: folder.Name})
	}
	var account *embyAccountResponse
	value, err := s.deps.Emby.Account(r.Context(), currentUser(r).ID)
	if err == nil {
		mapped := mapEmbyAccount(value)
		account = &mapped
	} else if !errors.Is(err, emby.ErrNotFound) && !errors.Is(err, database.ErrNotFound) {
		s.writeError(w, r, http.StatusInternalServerError, "EMBY_ACCOUNT_UNAVAILABLE", "Emby account state could not be loaded.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"configured": configured, "setupPrice": model.TXBMoney(priceMinor),
		"ratings": ratings, "libraries": libraries, "account": account,
	})
}

func decodeEmbyPreferences(w http.ResponseWriter, r *http.Request, withPassword bool) (emby.Preferences, string, error) {
	request := struct {
		Password           string   `json:"password"`
		MaxParentalRating  *int32   `json:"maxParentalRating"`
		DisabledLibraryIDs []string `json:"disabledLibraryIds"`
	}{}
	if err := decodeJSON(w, r, &request); err != nil {
		return emby.Preferences{}, "", err
	}
	if withPassword && request.Password == "" {
		return emby.Preferences{}, "", emby.ErrInvalidSetup
	}
	return emby.Preferences{MaxParentalRating: request.MaxParentalRating, DisabledLibraryIDs: request.DisabledLibraryIDs}, request.Password, nil
}

func (s *Server) setupEmby(w http.ResponseWriter, r *http.Request) {
	preferences, password, err := decodeEmbyPreferences(w, r, true)
	if err != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_EMBY_SETUP", "Choose a password and valid viewing preferences.")
		return
	}
	account, _, err := s.deps.Emby.Setup(r.Context(), currentUser(r).ID, password, preferences)
	if err != nil {
		status, code := http.StatusConflict, "EMBY_SETUP_FAILED"
		if errors.Is(err, database.ErrInsufficientBalance) {
			code = "INSUFFICIENT_BALANCE"
		}
		s.writeError(w, r, status, code, "Emby setup could not be queued.")
		return
	}
	writeJSON(w, http.StatusAccepted, mapEmbyAccount(account))
}

func (s *Server) updateEmbyPreferences(w http.ResponseWriter, r *http.Request) {
	preferences, _, err := decodeEmbyPreferences(w, r, false)
	if err != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_EMBY_PREFERENCES", "Choose valid Emby preferences.")
		return
	}
	account, err := s.deps.Emby.UpdatePreferences(r.Context(), currentUser(r).ID, preferences)
	if err != nil {
		s.writeError(w, r, http.StatusBadGateway, "EMBY_UPDATE_FAILED", "Emby preferences could not be updated.")
		return
	}
	writeJSON(w, http.StatusOK, mapEmbyAccount(account))
}

func (s *Server) changeEmbyPassword(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &request); err != nil || request.Password == "" {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_EMBY_PASSWORD", "Enter a new password.")
		return
	}
	if err := s.deps.Emby.ChangePassword(r.Context(), currentUser(r).ID, "", request.Password); err != nil {
		s.writeError(w, r, http.StatusBadGateway, "EMBY_PASSWORD_FAILED", "The Emby password could not be changed.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) adminEmbyAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := s.deps.Emby.ListAccounts(r.Context(), 200)
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "EMBY_ACCOUNTS_UNAVAILABLE", "Emby accounts could not be loaded.")
		return
	}
	items := make([]embyAccountResponse, 0, len(accounts))
	for _, account := range accounts {
		items = append(items, mapEmbyAccount(account))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) retryAdminEmbyAccount(w http.ResponseWriter, r *http.Request) {
	account, err := s.deps.Emby.RetryProvisioning(r.Context(), chiURLParam(r, "id"))
	if err != nil {
		s.writeError(w, r, http.StatusConflict, "EMBY_RETRY_FAILED", "This Emby setup cannot be retried.")
		return
	}
	writeJSON(w, http.StatusAccepted, mapEmbyAccount(account))
}
