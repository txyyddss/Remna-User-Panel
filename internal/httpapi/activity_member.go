package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
)

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
	groupMessageReward, err := s.deps.Activity.GroupMessageStatus(r.Context(), user.ID, activity.GroupMessageRewardConfig{
		Timezone: config.Timezone, Threshold: config.GroupMessageThreshold, RewardMinor: config.GroupMessageRewardMinor,
	})
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
		DailyRewardMinTXBMinor: strconv.FormatInt(config.RewardMinMinor, 10), DailyRewardMaxTXBMinor: strconv.FormatInt(config.RewardMaxMinor, 10), Games: gameResponses, Draws: drawResponses,
		RecentResults: mapActivityHistory(history, 30), GroupMessageReward: groupMessageReward})
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
	rewardMinValue, err := s.deps.Settings.Optional(ctx, activityRewardMinSetting)
	if err != nil {
		return activity.CheckInConfig{}, err
	}
	if strings.TrimSpace(rewardMinValue) == "" {
		rewardMinValue = "0"
	}
	rewardMinMinor, err := billing.ParseTXBMajor(rewardMinValue)
	if err != nil || rewardMinMinor < 0 {
		return activity.CheckInConfig{}, errActivityConfiguration
	}
	rewardMaxValue, err := s.deps.Settings.Optional(ctx, activityRewardMaxSetting)
	if err != nil {
		return activity.CheckInConfig{}, err
	}
	if strings.TrimSpace(rewardMaxValue) == "" {
		rewardMaxValue = "0"
	}
	rewardMaxMinor, err := billing.ParseTXBMajor(rewardMaxValue)
	if err != nil || rewardMaxMinor < rewardMinMinor {
		return activity.CheckInConfig{}, errActivityConfiguration
	}
	thresholdValue, err := s.deps.Settings.Optional(ctx, groupMessageThresholdSetting)
	if err != nil {
		return activity.CheckInConfig{}, err
	}
	if strings.TrimSpace(thresholdValue) == "" {
		thresholdValue = "0"
	}
	threshold, err := strconv.ParseInt(thresholdValue, 10, 32)
	if err != nil || threshold < 0 {
		return activity.CheckInConfig{}, errActivityConfiguration
	}
	groupRewardValue, err := s.deps.Settings.Optional(ctx, groupMessageRewardSetting)
	if err != nil {
		return activity.CheckInConfig{}, err
	}
	if strings.TrimSpace(groupRewardValue) == "" {
		groupRewardValue = "0"
	}
	groupRewardMinor, err := billing.ParseTXBMajor(groupRewardValue)
	if err != nil || groupRewardMinor < 0 {
		return activity.CheckInConfig{}, errActivityConfiguration
	}
	return activity.CheckInConfig{Timezone: timezone, RewardMinMinor: rewardMinMinor, RewardMaxMinor: rewardMaxMinor, GroupMessageThreshold: int(threshold), GroupMessageRewardMinor: groupRewardMinor}, nil
}
