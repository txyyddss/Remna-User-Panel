package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
)

type luckyDrawAdminResponse struct {
	ID          string                        `json:"id"`
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Enabled     bool                          `json:"enabled"`
	FeeTXBMinor string                        `json:"feeTxbMinor"`
	Prizes      []luckyDrawPrizeAdminResponse `json:"prizes"`
	CreatedAt   time.Time                     `json:"createdAt"`
	UpdatedAt   time.Time                     `json:"updatedAt"`
}

type luckyDrawPrizeAdminResponse struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Weight         string                 `json:"weight"`
	StockRemaining *int64                 `json:"stockRemaining,omitempty"`
	Reward         activityRewardResponse `json:"reward"`
}

func mapLuckyDrawAdmin(draw activity.LuckyDraw) luckyDrawAdminResponse {
	prizes := make([]luckyDrawPrizeAdminResponse, 0, len(draw.Prizes))
	for _, prize := range draw.Prizes {
		prizes = append(prizes, luckyDrawPrizeAdminResponse{ID: prize.ID, Name: prize.Name, Weight: strconv.FormatInt(prize.Weight, 10),
			StockRemaining: prize.StockRemaining, Reward: mapActivityReward(prize.Reward)})
	}
	return luckyDrawAdminResponse{ID: draw.ID, Name: draw.Name, Description: draw.Description, Enabled: draw.Enabled,
		FeeTXBMinor: strconv.FormatInt(draw.FeeMinor, 10), Prizes: prizes, CreatedAt: draw.CreatedAt, UpdatedAt: draw.UpdatedAt}
}

func (s *Server) adminLuckyDraws(w http.ResponseWriter, r *http.Request) {
	draws, err := s.deps.Activity.Draws(r.Context(), false)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	items := make([]luckyDrawAdminResponse, 0, len(draws))
	for _, draw := range draws {
		items = append(items, mapLuckyDrawAdmin(draw))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type luckyDrawRequest struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Enabled     bool                    `json:"enabled"`
	FeeTXBMinor string                  `json:"feeTxbMinor"`
	Prizes      []luckyDrawPrizeRequest `json:"prizes"`
}

type luckyDrawPrizeRequest struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Weight         string `json:"weight"`
	StockRemaining *int64 `json:"stockRemaining"`
	Reward         struct {
		Kind          activity.RewardKind `json:"kind"`
		TXBDeltaMinor string              `json:"txbDeltaMinor"`
		CouponID      string              `json:"couponId"`
		ExtensionDays int                 `json:"extensionDays"`
	} `json:"reward"`
}

func (s *Server) adminCreateLuckyDraw(w http.ResponseWriter, r *http.Request) {
	s.adminSaveLuckyDraw(w, r, "")
}

func (s *Server) adminUpdateLuckyDraw(w http.ResponseWriter, r *http.Request) {
	s.adminSaveLuckyDraw(w, r, chiURLParam(r, "id"))
}

func (s *Server) adminDeleteLuckyDraw(w http.ResponseWriter, r *http.Request) {
	if err := s.deps.Store.DeleteLuckyDraw(r.Context(), currentUser(r).ID, chiURLParam(r, "id"), time.Now().UTC()); err != nil {
		s.communityFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) adminLuckyDrawStatistics(w http.ResponseWriter, r *http.Request) {
	from, to, bucket, location, ok := s.statisticsWindow(w, r)
	if !ok {
		return
	}
	statistics, err := s.deps.Store.LuckyDrawStatistics(r.Context(), chiURLParam(r, "id"), from, to, bucket, location)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, statistics)
}

func (s *Server) adminSaveLuckyDraw(w http.ResponseWriter, r *http.Request, id string) {
	var request luckyDrawRequest
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_LUCKY_DRAW", "Lucky-draw fields are invalid.")
		return
	}
	fee, err := parseMinorString(request.FeeTXBMinor, true)
	if err != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_DRAW_FEE", "Draw fee must be a non-negative decimal integer string in TXB hundredths.")
		return
	}
	prizes := make([]activity.PrizeInput, 0, len(request.Prizes))
	for _, requestPrize := range request.Prizes {
		weight, weightErr := strconv.ParseInt(requestPrize.Weight, 10, 64)
		if weightErr != nil || weight <= 0 {
			s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_DRAW_WEIGHT", "Every prize weight must be a positive decimal integer string.")
			return
		}
		delta := int64(0)
		if requestPrize.Reward.Kind == activity.RewardTXBDelta {
			delta, err = parseSignedDecimalInt64(requestPrize.Reward.TXBDeltaMinor)
			if err != nil || delta == 0 {
				s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_DRAW_REWARD", "A TXB reward needs a non-zero signed decimal integer string.")
				return
			}
		}
		prizes = append(prizes, activity.PrizeInput{ID: requestPrize.ID, Name: requestPrize.Name, Weight: weight,
			StockRemaining: requestPrize.StockRemaining, Reward: activity.Reward{Kind: requestPrize.Reward.Kind,
				TXBDeltaMinor: delta, CouponID: requestPrize.Reward.CouponID, ExtensionDays: requestPrize.Reward.ExtensionDays}})
	}
	draw, err := s.deps.Activity.SaveDraw(r.Context(), activity.LuckyDrawInput{ID: id, Name: request.Name, Description: request.Description, Enabled: request.Enabled, FeeMinor: fee, Prizes: prizes})
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	status := http.StatusOK
	if id == "" {
		status = http.StatusCreated
	}
	writeJSON(w, status, mapLuckyDrawAdmin(draw))
}
