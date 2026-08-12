package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

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
	StockLimit     *int   `json:"stockLimit"`
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
		Visible: request.Visible, UpstreamPresent: true, StockLimit: request.StockLimit})
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
