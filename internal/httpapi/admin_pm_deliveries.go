package httpapi

import (
	"net/http"
	"strconv"
	"time"
)

type pmDeliveryResponse struct {
	OperationID     string    `json:"operationId"`
	Status          string    `json:"status"`
	Direction       string    `json:"direction"`
	ErrorCode       string    `json:"errorCode"`
	SourceChatID    string    `json:"sourceChatId"`
	SourceMessageID string    `json:"sourceMessageId"`
	ResultMessageID *string   `json:"resultMessageId"`
	TopicID         *string   `json:"topicId"`
	CreatedAt       time.Time `json:"createdAt"`
}

func (s *Server) adminPMDeliveries(w http.ResponseWriter, r *http.Request) {
	id := chiURLParam(r, "id")
	if _, err := s.deps.Store.PMConversation(r.Context(), id); err != nil {
		s.adminFailure(w, r, err)
		return
	}
	items, err := s.deps.Store.ListPMDeliveries(r.Context(), id)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	response := make([]pmDeliveryResponse, len(items))
	for index, item := range items {
		response[index] = pmDeliveryResponse{OperationID: item.OperationID, Status: item.Status, Direction: item.Direction, ErrorCode: item.ErrorCode,
			SourceChatID: strconv.FormatInt(item.SourceChatID, 10), SourceMessageID: strconv.FormatInt(item.SourceMessageID, 10), ResultMessageID: optionalPMID(item.ResultMessageID), TopicID: optionalPMID(item.TopicID), CreatedAt: item.CreatedAt}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}
