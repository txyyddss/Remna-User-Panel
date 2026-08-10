package httpapi

import (
	"net/http"
	"strconv"
)

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
