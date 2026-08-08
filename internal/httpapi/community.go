package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"github.com/txyyddss/Remna-User-Panel/internal/questionnaires"
)

const (
	activityTimezoneSetting = "activity.timezone"
	activityRewardSetting   = "activity.daily_reward_txb"
	defaultActivityTimezone = "Asia/Shanghai"
	maxQuestionnaireCSV     = int64(5 << 20)
	maxMultipartOverhead    = int64(256 << 10)
)

var errActivityConfiguration = errors.New("invalid activity configuration")

type nullableRequestField[T any] struct {
	Set   bool
	Value *T
}

func (field *nullableRequestField[T]) UnmarshalJSON(data []byte) error {
	field.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		field.Value = nil
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	field.Value = &value
	return nil
}

func (s *Server) mountCommunity(router chi.Router) {
	router.Get("/api/v1/activity", s.activityOverview)
	router.Post("/api/v1/activity/check-ins", s.activityCheckIn)
	router.Post("/api/v1/activity/bets", s.activityBet)
	router.Post("/api/v1/activity/draws", s.activityDraw)
	router.Get("/api/v1/coupons/wallet", s.couponWallet)
	router.Post("/api/v1/coupons/redeem", s.couponRedeem)
	router.Get("/api/v1/questionnaires/active", s.activeQuestionnaire)
	router.Get("/api/v1/questionnaires/history", s.questionnaireHistory)
	router.Post("/api/v1/questionnaires/{id}/participation", s.questionnaireParticipate)
}

func (s *Server) mountCommunityAdmin(router chi.Router) {
	router.Get("/activity-settings", s.adminActivitySettings)
	router.Put("/activity-settings", s.adminUpdateActivitySettings)
	router.Get("/activity-games", s.adminActivityGames)
	router.Post("/activity-games", s.adminCreateActivityGame)
	router.Put("/activity-games/{id}", s.adminUpdateActivityGame)
	router.Get("/lucky-draw", s.adminLuckyDraws)
	router.Post("/lucky-draw", s.adminCreateLuckyDraw)
	router.Put("/lucky-draw/{id}", s.adminUpdateLuckyDraw)
	router.Get("/coupons", s.adminCoupons)
	router.Post("/coupons", s.adminCreateCoupon)
	router.Put("/coupons/{id}", s.adminUpdateCoupon)
	router.Delete("/coupons/{id}", s.adminDeactivateCoupon)
	router.Get("/questionnaires", s.adminQuestionnaires)
	router.Post("/questionnaires", s.adminCreateQuestionnaire)
	router.Put("/questionnaires/{id}", s.adminUpdateQuestionnaire)
	router.Delete("/questionnaires/{id}", s.adminCloseQuestionnaire)
	router.Post("/questionnaires/{id}/activate", s.adminActivateQuestionnaire)
	router.Post("/questionnaires/{id}/imports", s.adminUploadQuestionnaireCSV)
	router.Get("/questionnaires/{id}/imports/{importID}", s.adminQuestionnaireImport)
	router.Post("/questionnaires/{id}/imports/{importID}/analyze", s.adminAnalyzeQuestionnaireCSV)
	router.Post("/questionnaires/{id}/imports/{importID}/settle", s.adminSettleQuestionnaireCSV)
}

type activityGameResponse struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Icon                string `json:"icon"`
	Description         string `json:"description"`
	Enabled             bool   `json:"enabled"`
	WinChanceBPS        int    `json:"winChanceBps"`
	MinimumStakeMinor   string `json:"minimumStakeMinor"`
	MaximumStakeMinor   string `json:"maximumStakeMinor"`
	ReturnMultiplierBPS int64  `json:"returnMultiplierBps"`
}

type luckyDrawResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	FeeTXBMinor string `json:"feeTxbMinor"`
	Enabled     bool   `json:"enabled"`
}

type activityRewardResponse struct {
	Kind          activity.RewardKind `json:"kind"`
	TXBDeltaMinor string              `json:"txbDeltaMinor,omitempty"`
	CouponID      string              `json:"couponId,omitempty"`
	ExtensionDays int                 `json:"extensionDays,omitempty"`
}

type activityResultResponse struct {
	ID            string                 `json:"id"`
	Kind          string                 `json:"kind"`
	Outcome       string                 `json:"outcome"`
	Message       string                 `json:"message"`
	Reward        activityRewardResponse `json:"reward"`
	StakeTXBMinor string                 `json:"stakeTxbMinor,omitempty"`
	BalanceAfter  model.Money            `json:"balanceAfter"`
	CreatedAt     time.Time              `json:"createdAt"`
}

type activityOverviewResponse struct {
	Balance             model.Money              `json:"balance"`
	TimeZone            string                   `json:"timeZone"`
	CheckedInToday      bool                     `json:"checkedInToday"`
	DailyRewardTXBMinor string                   `json:"dailyRewardTxbMinor"`
	Games               []activityGameResponse   `json:"games"`
	Draws               []luckyDrawResponse      `json:"draws"`
	RecentResults       []activityResultResponse `json:"recentResults"`
}

func mapActivityGame(game activity.Game) activityGameResponse {
	return activityGameResponse{
		ID: game.ID, Name: game.Name, Icon: game.Icon, Description: game.Description, Enabled: game.Enabled,
		WinChanceBPS: game.WinChanceBPS, MinimumStakeMinor: strconv.FormatInt(game.MinimumStakeMinor, 10),
		MaximumStakeMinor: strconv.FormatInt(game.MaximumStakeMinor, 10), ReturnMultiplierBPS: game.ReturnMultiplierBPS,
	}
}

func mapLuckyDraw(draw activity.LuckyDraw) luckyDrawResponse {
	return luckyDrawResponse{ID: draw.ID, Name: draw.Name, Description: draw.Description, FeeTXBMinor: strconv.FormatInt(draw.FeeMinor, 10), Enabled: draw.Enabled}
}

func mapActivityReward(reward activity.Reward) activityRewardResponse {
	result := activityRewardResponse{Kind: reward.Kind, CouponID: reward.CouponID, ExtensionDays: reward.ExtensionDays}
	if reward.Kind == activity.RewardTXBDelta {
		result.TXBDeltaMinor = strconv.FormatInt(reward.TXBDeltaMinor, 10)
	}
	return result
}

func mapCheckInResult(result activity.DailyCheckIn) activityResultResponse {
	reward := activityRewardResponse{Kind: activity.RewardNone}
	if result.RewardMinor != 0 {
		reward = activityRewardResponse{Kind: activity.RewardTXBDelta, TXBDeltaMinor: strconv.FormatInt(result.RewardMinor, 10)}
	}
	message := "Daily reward recorded."
	if result.AlreadyClaimed {
		message = "Today's reward was already recorded."
	}
	return activityResultResponse{ID: result.ID, Kind: "check_in", Outcome: "complete", Message: message, Reward: reward,
		BalanceAfter: model.TXBMoney(result.BalanceAfterMinor), CreatedAt: result.CreatedAt}
}

func mapBetResult(result activity.BetResult) activityResultResponse {
	outcome, message := "loss", "Result recorded: no return."
	reward := activityRewardResponse{Kind: activity.RewardNone}
	if result.Won {
		outcome = "win"
		message = "Win recorded. Total return: " + model.TXBMoney(result.PayoutMinor).Display + "."
		reward = activityRewardResponse{Kind: activity.RewardTXBDelta, TXBDeltaMinor: strconv.FormatInt(result.PayoutMinor, 10)}
	}
	return activityResultResponse{ID: result.ID, Kind: "bet", Outcome: outcome, Message: message, Reward: reward,
		StakeTXBMinor: strconv.FormatInt(result.StakeMinor, 10), BalanceAfter: model.TXBMoney(result.BalanceAfterMinor), CreatedAt: result.CreatedAt}
}

func mapDrawResult(result activity.DrawResult) activityResultResponse {
	return activityResultResponse{ID: result.ID, Kind: "draw", Outcome: "complete", Message: "Draw recorded: " + result.PrizeName + ".",
		Reward: mapActivityReward(result.Reward), BalanceAfter: model.TXBMoney(result.BalanceAfterMinor), CreatedAt: result.CreatedAt}
}

func mapActivityHistory(history activity.History, limit int) []activityResultResponse {
	results := make([]activityResultResponse, 0, len(history.Bets)+len(history.CheckIns)+len(history.Draws))
	for _, result := range history.Bets {
		results = append(results, mapBetResult(result))
	}
	for _, result := range history.CheckIns {
		results = append(results, mapCheckInResult(result))
	}
	for _, result := range history.Draws {
		results = append(results, mapDrawResult(result))
	}
	sort.SliceStable(results, func(left, right int) bool { return results[left].CreatedAt.After(results[right].CreatedAt) })
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

func (s *Server) activityOverview(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	config, err := s.activityConfig(r.Context())
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	games, err := s.deps.Activity.Games(r.Context(), true)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	draws, err := s.deps.Activity.Draws(r.Context(), true)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	history, err := s.deps.Activity.History(r.Context(), user.ID, 30)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	balance, err := s.deps.Store.Balance(r.Context(), user.ID)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	gameResponses := make([]activityGameResponse, 0, len(games))
	for _, game := range games {
		gameResponses = append(gameResponses, mapActivityGame(game))
	}
	drawResponses := make([]luckyDrawResponse, 0, len(draws))
	for _, draw := range draws {
		drawResponses = append(drawResponses, mapLuckyDraw(draw))
	}
	checkedIn := false
	location, _ := time.LoadLocation(config.Timezone)
	today := time.Now().In(location).Format(time.DateOnly)
	for _, item := range history.CheckIns {
		if item.LocalDate == today {
			checkedIn = true
			break
		}
	}
	writeJSON(w, http.StatusOK, activityOverviewResponse{Balance: balance, TimeZone: config.Timezone, CheckedInToday: checkedIn,
		DailyRewardTXBMinor: strconv.FormatInt(config.RewardMinor, 10), Games: gameResponses, Draws: drawResponses,
		RecentResults: mapActivityHistory(history, 30)})
}

func (s *Server) activityCheckIn(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	if _, ok := s.requireIdempotencyKey(w, r); !ok {
		return
	}
	config, err := s.activityConfig(r.Context())
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	result, err := s.deps.Activity.CheckIn(r.Context(), user.ID, config)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapCheckInResult(result))
}

func (s *Server) activityBet(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	key, ok := s.requireIdempotencyKey(w, r)
	if !ok {
		return
	}
	var request struct {
		GameID        string `json:"gameId"`
		StakeTXBMinor string `json:"stakeTxbMinor"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_ACTIVITY_REQUEST", "A game and stake are required.")
		return
	}
	stake, err := parseMinorString(request.StakeTXBMinor, false)
	if err != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_STAKE", "Stake must be a positive decimal integer string in TXB hundredths.")
		return
	}
	result, err := s.deps.Activity.PlayBet(r.Context(), user.ID, request.GameID, stake, key)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapBetResult(result))
}

func (s *Server) activityDraw(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	key, ok := s.requireIdempotencyKey(w, r)
	if !ok {
		return
	}
	var request struct {
		DrawID string `json:"drawId"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_ACTIVITY_REQUEST", "A lucky draw is required.")
		return
	}
	result, err := s.deps.Activity.Draw(r.Context(), user.ID, request.DrawID, key)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapDrawResult(result))
}

func (s *Server) activityConfig(ctx context.Context) (activity.CheckInConfig, error) {
	timezone, err := s.deps.Settings.Optional(ctx, activityTimezoneSetting)
	if err != nil {
		return activity.CheckInConfig{}, err
	}
	if strings.TrimSpace(timezone) == "" {
		timezone = defaultActivityTimezone
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return activity.CheckInConfig{}, errActivityConfiguration
	}
	rewardValue, err := s.deps.Settings.Optional(ctx, activityRewardSetting)
	if err != nil {
		return activity.CheckInConfig{}, err
	}
	if strings.TrimSpace(rewardValue) == "" {
		rewardValue = "0"
	}
	rewardMinor, err := billing.ParseTXBMajor(rewardValue)
	if err != nil || rewardMinor < 0 {
		return activity.CheckInConfig{}, errActivityConfiguration
	}
	return activity.CheckInConfig{Timezone: timezone, RewardMinor: rewardMinor}, nil
}

func (s *Server) adminActivitySettings(w http.ResponseWriter, r *http.Request) {
	config, err := s.activityConfig(r.Context())
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"timezone":            config.Timezone,
		"dailyRewardTxb":      txbMajorString(config.RewardMinor),
		"dailyRewardTxbMinor": strconv.FormatInt(config.RewardMinor, 10),
	})
}

func (s *Server) adminUpdateActivitySettings(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Timezone       string `json:"timezone"`
		DailyRewardTXB string `json:"dailyRewardTxb"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_ACTIVITY_SETTINGS", "Timezone and daily reward are required.")
		return
	}
	request.Timezone = strings.TrimSpace(request.Timezone)
	request.DailyRewardTXB = strings.TrimSpace(request.DailyRewardTXB)
	if _, err := time.LoadLocation(request.Timezone); err != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_TIMEZONE", "Use a valid IANA timezone such as Asia/Shanghai.")
		return
	}
	minor, err := billing.ParseTXBMajor(request.DailyRewardTXB)
	if err != nil || minor < 0 {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_REWARD", "Daily reward must be a non-negative TXB amount with at most two decimals.")
		return
	}
	actorID := currentUser(r).ID
	if err := s.deps.Admin.PutSetting(r.Context(), actorID, activityTimezoneSetting, request.Timezone); err != nil {
		s.communityFailure(w, r, err)
		return
	}
	if err := s.deps.Admin.PutSetting(r.Context(), actorID, activityRewardSetting, request.DailyRewardTXB); err != nil {
		s.communityFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"timezone": request.Timezone, "dailyRewardTxb": txbMajorString(minor), "dailyRewardTxbMinor": strconv.FormatInt(minor, 10),
	})
}

func (s *Server) adminActivityGames(w http.ResponseWriter, r *http.Request) {
	games, err := s.deps.Activity.Games(r.Context(), false)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	items := make([]activityGameResponse, 0, len(games))
	for _, game := range games {
		items = append(items, mapActivityGame(game))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type activityGameRequest struct {
	Name                string `json:"name"`
	Icon                string `json:"icon"`
	Description         string `json:"description"`
	Enabled             bool   `json:"enabled"`
	WinChanceBPS        int    `json:"winChanceBps"`
	MinimumStakeMinor   string `json:"minimumStakeMinor"`
	MaximumStakeMinor   string `json:"maximumStakeMinor"`
	ReturnMultiplierBPS int64  `json:"returnMultiplierBps"`
}

func (s *Server) adminCreateActivityGame(w http.ResponseWriter, r *http.Request) {
	s.adminSaveActivityGame(w, r, "")
}

func (s *Server) adminUpdateActivityGame(w http.ResponseWriter, r *http.Request) {
	s.adminSaveActivityGame(w, r, chiURLParam(r, "id"))
}

func (s *Server) adminSaveActivityGame(w http.ResponseWriter, r *http.Request, id string) {
	var request activityGameRequest
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_ACTIVITY_GAME", "Game fields are invalid.")
		return
	}
	minimum, minimumErr := parseMinorString(request.MinimumStakeMinor, false)
	maximum, maximumErr := parseMinorString(request.MaximumStakeMinor, false)
	if minimumErr != nil || maximumErr != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_STAKE_RANGE", "Stake limits must be positive decimal integer strings in TXB hundredths.")
		return
	}
	game, err := s.deps.Activity.SaveGame(r.Context(), activity.GameInput{ID: id, Name: request.Name, Icon: request.Icon,
		Description: request.Description, Enabled: request.Enabled, WinChanceBPS: request.WinChanceBPS, MinimumStakeMinor: minimum,
		MaximumStakeMinor: maximum, ReturnMultiplierBPS: request.ReturnMultiplierBPS})
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	status := http.StatusOK
	if id == "" {
		status = http.StatusCreated
	}
	writeJSON(w, status, mapActivityGame(game))
}

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

type couponResponse struct {
	ID               string               `json:"id"`
	Code             string               `json:"code"`
	Name             string               `json:"name"`
	Kind             coupons.CouponKind   `json:"kind"`
	DiscountMode     coupons.DiscountMode `json:"discountMode,omitempty"`
	ValueMinorOrBPS  string               `json:"valueMinorOrBps"`
	PercentCapMinor  *string              `json:"percentCapMinor,omitempty"`
	EligibleComboIDs []string             `json:"eligibleComboIds"`
	EligibleSquadIDs []string             `json:"eligibleSquadIds"`
	ExpiresAt        *time.Time           `json:"expiresAt"`
	GlobalUseLimit   *int64               `json:"globalUseLimit"`
	PerUserUseLimit  *int64               `json:"perUserUseLimit"`
	Active           bool                 `json:"active"`
	CreatedAt        time.Time            `json:"createdAt"`
	UpdatedAt        time.Time            `json:"updatedAt"`
}

type couponGrantResponse struct {
	ID         string         `json:"id"`
	Coupon     couponResponse `json:"coupon"`
	SourceType string         `json:"sourceType"`
	SourceID   string         `json:"sourceId"`
	Status     string         `json:"status"`
	UseCount   int64          `json:"useCount"`
	CreatedAt  time.Time      `json:"createdAt"`
	ConsumedAt *time.Time     `json:"consumedAt"`
}

type couponRedemptionResponse struct {
	ID                string               `json:"id"`
	Coupon            couponResponse       `json:"coupon"`
	Grant             *couponGrantResponse `json:"grant"`
	BalanceDeltaMinor string               `json:"balanceDeltaMinor"`
	BalanceAfterMinor string               `json:"balanceAfterMinor"`
	IdempotencyKey    string               `json:"idempotencyKey"`
	Replayed          bool                 `json:"replayed"`
	CreatedAt         time.Time            `json:"createdAt"`
}

func mapCoupon(coupon coupons.Coupon) couponResponse {
	var capMinor *string
	if coupon.PercentCapMinor != nil {
		value := strconv.FormatInt(*coupon.PercentCapMinor, 10)
		capMinor = &value
	}
	return couponResponse{ID: coupon.ID, Code: coupon.Code, Name: coupon.Name, Kind: coupon.Kind, DiscountMode: coupon.DiscountMode,
		ValueMinorOrBPS: strconv.FormatInt(coupon.ValueMinorOrBPS, 10), PercentCapMinor: capMinor,
		EligibleComboIDs: nonNilStrings(coupon.EligibleComboIDs), EligibleSquadIDs: nonNilStrings(coupon.EligibleSquadIDs),
		ExpiresAt: coupon.ExpiresAt, GlobalUseLimit: coupon.GlobalUseLimit, PerUserUseLimit: coupon.PerUserUseLimit,
		Active: coupon.Active, CreatedAt: coupon.CreatedAt, UpdatedAt: coupon.UpdatedAt}
}

func mapCouponGrant(grant coupons.Grant) couponGrantResponse {
	return couponGrantResponse{ID: grant.ID, Coupon: mapCoupon(grant.Coupon), SourceType: grant.SourceType, SourceID: grant.SourceID,
		Status: grant.Status, UseCount: grant.UseCount, CreatedAt: grant.CreatedAt, ConsumedAt: grant.ConsumedAt}
}

func (s *Server) couponWallet(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	grants, err := s.deps.Coupons.Wallet(r.Context(), user.ID)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	items := make([]couponGrantResponse, 0, len(grants))
	for _, grant := range grants {
		items = append(items, mapCouponGrant(grant))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) couponRedeem(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	key, ok := s.requireIdempotencyKey(w, r)
	if !ok {
		return
	}
	var request struct {
		Code string `json:"code"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_COUPON_REQUEST", "A coupon code is required.")
		return
	}
	result, err := s.deps.Coupons.Redeem(r.Context(), user.ID, request.Code, key)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	response := couponRedemptionResponse{ID: result.ID, Coupon: mapCoupon(result.Coupon), BalanceDeltaMinor: strconv.FormatInt(result.BalanceDeltaMinor, 10),
		BalanceAfterMinor: strconv.FormatInt(result.BalanceAfterMinor, 10), IdempotencyKey: result.IdempotencyKey,
		Replayed: result.Replayed, CreatedAt: result.CreatedAt}
	if result.Grant != nil {
		grant := mapCouponGrant(*result.Grant)
		response.Grant = &grant
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) adminCoupons(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Coupons.List(r.Context(), false)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	response := make([]couponResponse, 0, len(items))
	for _, item := range items {
		response = append(response, mapCoupon(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

type couponRequest struct {
	Code             *string                         `json:"code"`
	Name             *string                         `json:"name"`
	Kind             *coupons.CouponKind             `json:"kind"`
	DiscountMode     *coupons.DiscountMode           `json:"discountMode"`
	ValueMinorOrBPS  *string                         `json:"valueMinorOrBps"`
	PercentCapMinor  nullableRequestField[string]    `json:"percentCapMinor"`
	EligibleComboIDs *[]string                       `json:"eligibleComboIds"`
	EligibleSquadIDs *[]string                       `json:"eligibleSquadIds"`
	ExpiresAt        nullableRequestField[time.Time] `json:"expiresAt"`
	GlobalUseLimit   nullableRequestField[int64]     `json:"globalUseLimit"`
	PerUserUseLimit  nullableRequestField[int64]     `json:"perUserUseLimit"`
	Active           *bool                           `json:"active"`
}

func (s *Server) adminCreateCoupon(w http.ResponseWriter, r *http.Request) {
	s.adminSaveCoupon(w, r, "")
}

func (s *Server) adminUpdateCoupon(w http.ResponseWriter, r *http.Request) {
	s.adminSaveCoupon(w, r, chiURLParam(r, "id"))
}

func (s *Server) adminDeactivateCoupon(w http.ResponseWriter, r *http.Request) {
	active := false
	input, err := s.couponInput(r.Context(), chiURLParam(r, "id"), couponRequest{Active: &active})
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	if _, err := s.deps.Coupons.Save(r.Context(), input); err != nil {
		s.communityFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) adminSaveCoupon(w http.ResponseWriter, r *http.Request, id string) {
	var request couponRequest
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_COUPON", "Coupon fields are invalid.")
		return
	}
	input, err := s.couponInput(r.Context(), id, request)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	item, err := s.deps.Coupons.Save(r.Context(), input)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	status := http.StatusOK
	if id == "" {
		status = http.StatusCreated
	}
	writeJSON(w, status, mapCoupon(item))
}

func (s *Server) couponInput(ctx context.Context, id string, request couponRequest) (coupons.CouponInput, error) {
	var input coupons.CouponInput
	if id != "" {
		items, err := s.deps.Coupons.List(ctx, false)
		if err != nil {
			return coupons.CouponInput{}, err
		}
		found := false
		for _, item := range items {
			if item.ID == id {
				input = item.CouponInput
				found = true
				break
			}
		}
		if !found {
			return coupons.CouponInput{}, database.ErrNotFound
		}
	}
	input.ID = id
	if request.Code != nil {
		input.Code = *request.Code
	}
	if request.Name != nil {
		input.Name = *request.Name
	}
	if request.Kind != nil {
		input.Kind = *request.Kind
	}
	if request.DiscountMode != nil {
		input.DiscountMode = *request.DiscountMode
	}
	if request.ValueMinorOrBPS != nil {
		value, err := parseMinorString(*request.ValueMinorOrBPS, true)
		if err != nil {
			return coupons.CouponInput{}, coupons.ErrInvalidInput
		}
		input.ValueMinorOrBPS = value
	}
	if request.PercentCapMinor.Set {
		input.PercentCapMinor = nil
		if request.PercentCapMinor.Value != nil {
			value, err := parseMinorString(*request.PercentCapMinor.Value, true)
			if err != nil {
				return coupons.CouponInput{}, coupons.ErrInvalidInput
			}
			input.PercentCapMinor = &value
		}
	}
	if request.EligibleComboIDs != nil {
		input.EligibleComboIDs = *request.EligibleComboIDs
	}
	if request.EligibleSquadIDs != nil {
		input.EligibleSquadIDs = *request.EligibleSquadIDs
	}
	if request.ExpiresAt.Set {
		input.ExpiresAt = request.ExpiresAt.Value
	}
	if request.GlobalUseLimit.Set {
		input.GlobalUseLimit = request.GlobalUseLimit.Value
	}
	if request.PerUserUseLimit.Set {
		input.PerUserUseLimit = request.PerUserUseLimit.Value
	}
	if request.Active != nil {
		input.Active = *request.Active
	}
	return input, nil
}

type questionnaireResponse struct {
	ID               string                `json:"id"`
	Title            string                `json:"title"`
	Description      string                `json:"description"`
	FormURL          string                `json:"formUrl"`
	RewardTXBMinor   string                `json:"rewardTxbMinor"`
	Status           questionnaires.Status `json:"status"`
	ClosesAt         *time.Time            `json:"closesAt"`
	ParticipantCount int                   `json:"participantCount"`
	RewardedCount    int                   `json:"rewardedCount"`
	CreatedAt        time.Time             `json:"createdAt"`
	UpdatedAt        time.Time             `json:"updatedAt"`
}

type questionnaireParticipationResponse struct {
	ID              string     `json:"id"`
	QuestionnaireID string     `json:"questionnaireId"`
	ValidationCode  string     `json:"validationCode"`
	AwardedAt       *time.Time `json:"awardedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type activeQuestionnaireResponse struct {
	ID             string                              `json:"id"`
	Title          string                              `json:"title"`
	Description    string                              `json:"description"`
	FormURL        string                              `json:"formUrl"`
	RewardTXBMinor string                              `json:"rewardTxbMinor"`
	ClosesAt       *time.Time                          `json:"closesAt"`
	Participation  *questionnaireParticipationResponse `json:"participation"`
}

type questionnaireHistoryResponse struct {
	Questionnaire questionnaireResponse              `json:"questionnaire"`
	Participation questionnaireParticipationResponse `json:"participation"`
}

type questionnaireSettlementResponse struct {
	ImportID            string    `json:"importId"`
	QuestionnaireID     string    `json:"questionnaireId"`
	MatchedCount        int       `json:"matchedCount"`
	DuplicateCount      int       `json:"duplicateCount"`
	UnknownCount        int       `json:"unknownCount"`
	MalformedCount      int       `json:"malformedCount"`
	AlreadyAwardedCount int       `json:"alreadyAwardedCount"`
	RewardedCount       int       `json:"rewardedCount"`
	RewardTXBMinor      string    `json:"rewardTxbMinor"`
	SettledAt           time.Time `json:"settledAt"`
	Replayed            bool      `json:"replayed"`
}

type questionnaireImportStateResponse struct {
	Preview questionnaires.ImportPreview     `json:"preview"`
	Report  *questionnaireSettlementResponse `json:"report,omitempty"`
}

func mapQuestionnaireImportState(state questionnaires.ImportState) questionnaireImportStateResponse {
	response := questionnaireImportStateResponse{Preview: mapQuestionnaireImportPreview(state.Preview)}
	if state.Report != nil {
		report := state.Report
		response.Report = &questionnaireSettlementResponse{
			ImportID: report.ImportID, QuestionnaireID: report.QuestionnaireID, MatchedCount: report.MatchedCount,
			DuplicateCount: report.DuplicateCount, UnknownCount: report.UnknownCount, MalformedCount: report.MalformedCount,
			AlreadyAwardedCount: report.AlreadyAwardedCount, RewardedCount: report.RewardedCount,
			RewardTXBMinor: strconv.FormatInt(report.RewardTXBMinor, 10), SettledAt: report.SettledAt, Replayed: report.Replayed,
		}
	}
	return response
}

func mapQuestionnaireImportPreview(preview questionnaires.ImportPreview) questionnaires.ImportPreview {
	if preview.Headers == nil {
		preview.Headers = []string{}
	}
	if preview.SampleRows == nil {
		preview.SampleRows = [][]string{}
	}
	return preview
}

func mapQuestionnaire(item questionnaires.Questionnaire) questionnaireResponse {
	return questionnaireResponse{ID: item.ID, Title: item.Title, Description: item.Description, FormURL: item.FormURL,
		RewardTXBMinor: strconv.FormatInt(item.RewardTXBMinor, 10), Status: item.Status, ClosesAt: item.ClosesAt,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func mapQuestionnaireParticipation(item questionnaires.Participant) questionnaireParticipationResponse {
	return questionnaireParticipationResponse{ID: item.ID, QuestionnaireID: item.QuestionnaireID,
		ValidationCode: item.ValidationCode, AwardedAt: item.AwardedAt, CreatedAt: item.CreatedAt}
}

func (s *Server) activeQuestionnaire(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	item, err := s.deps.Questionnaires.Active(r.Context())
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	if item == nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	var participation *questionnaireParticipationResponse
	history, err := s.deps.Questionnaires.History(r.Context(), user.ID, 200)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	for _, record := range history {
		if record.Questionnaire.ID == item.ID {
			mapped := mapQuestionnaireParticipation(record.Participation)
			participation = &mapped
			break
		}
	}
	writeJSON(w, http.StatusOK, activeQuestionnaireResponse{ID: item.ID, Title: item.Title, Description: item.Description,
		FormURL: item.FormURL, RewardTXBMinor: strconv.FormatInt(item.RewardTXBMinor, 10), ClosesAt: item.ClosesAt, Participation: participation})
}

func (s *Server) questionnaireHistory(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	items, err := s.deps.Questionnaires.History(r.Context(), user.ID, 100)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	response := make([]questionnaireHistoryResponse, 0, len(items))
	for _, item := range items {
		response = append(response, questionnaireHistoryResponse{Questionnaire: mapQuestionnaire(item.Questionnaire),
			Participation: mapQuestionnaireParticipation(item.Participation)})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (s *Server) questionnaireParticipate(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	if _, ok := s.requireIdempotencyKey(w, r); !ok {
		return
	}
	result, err := s.deps.Questionnaires.Participate(r.Context(), chiURLParam(r, "id"), user.ID)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapQuestionnaireParticipation(result))
}

func (s *Server) adminQuestionnaires(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Questionnaires.List(r.Context())
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	response := make([]questionnaireResponse, 0, len(items))
	for _, item := range items {
		response = append(response, mapQuestionnaire(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

type questionnaireRequest struct {
	Title          *string                         `json:"title"`
	Description    *string                         `json:"description"`
	FormURL        *string                         `json:"formUrl"`
	RewardTXBMinor *string                         `json:"rewardTxbMinor"`
	Status         *questionnaires.Status          `json:"status"`
	ClosesAt       nullableRequestField[time.Time] `json:"closesAt"`
}

func (s *Server) adminCreateQuestionnaire(w http.ResponseWriter, r *http.Request) {
	s.adminSaveQuestionnaire(w, r, "")
}

func (s *Server) adminUpdateQuestionnaire(w http.ResponseWriter, r *http.Request) {
	s.adminSaveQuestionnaire(w, r, chiURLParam(r, "id"))
}

func (s *Server) adminCloseQuestionnaire(w http.ResponseWriter, r *http.Request) {
	status := questionnaires.StatusClosed
	input, err := s.questionnaireInput(r.Context(), chiURLParam(r, "id"), questionnaireRequest{Status: &status})
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	if _, err := s.deps.Questionnaires.Save(r.Context(), input); err != nil {
		s.communityFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) adminSaveQuestionnaire(w http.ResponseWriter, r *http.Request, id string) {
	var request questionnaireRequest
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_QUESTIONNAIRE", "Questionnaire fields are invalid.")
		return
	}
	input, err := s.questionnaireInput(r.Context(), id, request)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	item, err := s.deps.Questionnaires.Save(r.Context(), input)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	status := http.StatusOK
	if id == "" {
		status = http.StatusCreated
	}
	writeJSON(w, status, mapQuestionnaire(item))
}

func (s *Server) questionnaireInput(ctx context.Context, id string, request questionnaireRequest) (questionnaires.QuestionnaireInput, error) {
	input := questionnaires.QuestionnaireInput{ID: id, Status: questionnaires.StatusDraft}
	if id != "" {
		items, err := s.deps.Questionnaires.List(ctx)
		if err != nil {
			return questionnaires.QuestionnaireInput{}, err
		}
		found := false
		for _, item := range items {
			if item.ID == id {
				input = item.QuestionnaireInput
				found = true
				break
			}
		}
		if !found {
			return questionnaires.QuestionnaireInput{}, database.ErrNotFound
		}
	}
	input.ID = id
	if request.Title != nil {
		input.Title = *request.Title
	}
	if request.Description != nil {
		input.Description = *request.Description
	}
	if request.FormURL != nil {
		input.FormURL = *request.FormURL
	}
	if request.RewardTXBMinor != nil {
		value, err := parseMinorString(*request.RewardTXBMinor, true)
		if err != nil {
			return questionnaires.QuestionnaireInput{}, questionnaires.ErrInvalidInput
		}
		input.RewardTXBMinor = value
	}
	if request.Status != nil {
		input.Status = *request.Status
	}
	if request.ClosesAt.Set {
		input.ClosesAt = request.ClosesAt.Value
	}
	return input, nil
}

func (s *Server) adminActivateQuestionnaire(w http.ResponseWriter, r *http.Request) {
	id := chiURLParam(r, "id")
	input, err := s.questionnaireInput(r.Context(), id, questionnaireRequest{})
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	input.Status = questionnaires.StatusActive
	item, err := s.deps.Questionnaires.Save(r.Context(), input)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapQuestionnaire(item))
}

func (s *Server) adminUploadQuestionnaireCSV(w http.ResponseWriter, r *http.Request) {
	key, err := optionalOrGeneratedIdempotencyKey(w, r)
	if err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_IDEMPOTENCY_KEY", "Idempotency-Key must contain 1 to 128 characters.")
		return
	}
	content, err := boundedMultipartFile(w, r, "file", maxQuestionnaireCSV)
	if err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_CSV_UPLOAD", "Upload one UTF-8 CSV file no larger than 5 MiB.")
		return
	}
	preview, err := s.deps.Questionnaires.UploadCSV(r.Context(), chiURLParam(r, "id"), bytes.NewReader(content), key)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapQuestionnaireImportPreview(preview))
}

func (s *Server) adminAnalyzeQuestionnaireCSV(w http.ResponseWriter, r *http.Request) {
	var request struct {
		CodeColumn string `json:"codeColumn"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_IMPORT_ANALYSIS", "Select one validation-code column.")
		return
	}
	if !s.questionnaireImportBelongs(w, r) {
		return
	}
	analysis, err := s.deps.Questionnaires.AnalyzeImport(r.Context(), chiURLParam(r, "importID"), request.CodeColumn)
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, analysis)
}

func (s *Server) adminSettleQuestionnaireCSV(w http.ResponseWriter, r *http.Request) {
	if !s.questionnaireImportBelongs(w, r) {
		return
	}
	preview, err := s.deps.Questionnaires.ConfirmImport(r.Context(), chiURLParam(r, "importID"))
	if err != nil {
		s.communityFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, mapQuestionnaireImportPreview(preview))
}

func (s *Server) adminQuestionnaireImport(w http.ResponseWriter, r *http.Request) {
	state, ok := s.loadQuestionnaireImport(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, mapQuestionnaireImportState(state))
}

func (s *Server) questionnaireImportBelongs(w http.ResponseWriter, r *http.Request) bool {
	_, ok := s.loadQuestionnaireImport(w, r)
	return ok
}

func (s *Server) loadQuestionnaireImport(w http.ResponseWriter, r *http.Request) (questionnaires.ImportState, bool) {
	state, err := s.deps.Questionnaires.Import(r.Context(), chiURLParam(r, "importID"))
	if err != nil {
		s.communityFailure(w, r, err)
		return questionnaires.ImportState{}, false
	}
	if state.Preview.QuestionnaireID != chiURLParam(r, "id") {
		s.writeError(w, r, http.StatusNotFound, "IMPORT_NOT_FOUND", "Questionnaire import not found.")
		return questionnaires.ImportState{}, false
	}
	return state, true
}

func (s *Server) requireIdempotencyKey(w http.ResponseWriter, r *http.Request) (string, bool) {
	value := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if value == "" || len(value) > 128 {
		s.writeError(w, r, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "Send an Idempotency-Key header containing 1 to 128 characters.")
		return "", false
	}
	return value, true
}

func optionalOrGeneratedIdempotencyKey(w http.ResponseWriter, r *http.Request) (string, error) {
	value := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if len(value) > 128 {
		return "", errors.New("idempotency key is too long")
	}
	if value != "" {
		return value, nil
	}
	generated, err := ids.New()
	if err != nil {
		return "", err
	}
	w.Header().Set("Idempotency-Key", generated)
	w.Header().Set("Idempotency-Key-Generated", "true")
	return generated, nil
}

func boundedMultipartFile(w http.ResponseWriter, r *http.Request, fieldName string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, errors.New("invalid upload bound")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes+maxMultipartOverhead)
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}
	var result []byte
	found := false
	for {
		part, nextErr := reader.NextPart()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			return nil, nextErr
		}
		if part.FormName() != fieldName || part.FileName() == "" || found {
			_ = part.Close()
			return nil, errors.New("multipart upload must contain exactly one file field")
		}
		content, readErr := io.ReadAll(io.LimitReader(part, maxBytes+1))
		closeErr := part.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if int64(len(content)) > maxBytes {
			return nil, errors.New("uploaded file exceeds bound")
		}
		result = content
		found = true
	}
	if !found {
		return nil, errors.New("multipart file is missing")
	}
	return result, nil
}

func parseMinorString(value string, allowZero bool) (int64, error) {
	if value == "" || strings.TrimSpace(value) != value {
		return 0, errors.New("minor amount must be a canonical decimal integer")
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return 0, errors.New("minor amount must contain decimal digits only")
		}
	}
	minor, err := strconv.ParseInt(value, 10, 64)
	if err != nil || minor < 0 || (!allowZero && minor == 0) {
		return 0, errors.New("minor amount is outside its allowed range")
	}
	return minor, nil
}

func parseSignedDecimalInt64(value string) (int64, error) {
	if value == "" || strings.TrimSpace(value) != value || value == "-" {
		return 0, errors.New("signed amount must be a canonical decimal integer")
	}
	digits := value
	if strings.HasPrefix(digits, "-") {
		digits = strings.TrimPrefix(digits, "-")
	}
	for _, character := range digits {
		if character < '0' || character > '9' {
			return 0, errors.New("signed amount must contain decimal digits only")
		}
	}
	return strconv.ParseInt(value, 10, 64)
}

func txbMajorString(minor int64) string {
	sign := ""
	if minor < 0 {
		sign = "-"
		minor = -minor
	}
	return sign + strconv.FormatInt(minor/100, 10) + "." + leftPadTwo(minor%100)
}

func leftPadTwo(value int64) string {
	if value < 10 {
		return "0" + strconv.FormatInt(value, 10)
	}
	return strconv.FormatInt(value, 10)
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func (s *Server) communityFailure(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := http.StatusInternalServerError, "COMMUNITY_OPERATION_FAILED", "The request could not be completed."
	switch {
	case errors.Is(err, database.ErrNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "The requested record was not found."
	case errors.Is(err, database.ErrInsufficientBalance):
		status, code, message = http.StatusConflict, "INSUFFICIENT_BALANCE", "Your TXB balance is too low for this action."
	case errors.Is(err, database.ErrConflict):
		status, code, message = http.StatusConflict, "CONFLICT", "The request conflicts with the current state. Refresh and retry."
	case errors.Is(err, activity.ErrInvalidInput), errors.Is(err, coupons.ErrInvalidInput), errors.Is(err, questionnaires.ErrInvalidInput):
		status, code, message = http.StatusUnprocessableEntity, "INVALID_REQUEST", "One or more request fields are invalid."
	}
	s.writeError(w, r, status, code, message)
}
