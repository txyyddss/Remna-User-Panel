package model

import (
	"encoding/json"
	"time"
)

// PMConversation joins persistent routing references with canonical account data.
// Conversation records contain routing references; pending outbound text is encrypted separately.
type PMConversation struct {
	ProfileState                                      string
	ID, UserID, FirstName, LastName, Username, Locale string
	TelegramID, ChatID, TopicID, ProfileMessageID     int64
	TopicState, TopicOperationID                      string
	Blocked, Muted                                    bool
	CreatedAt, UpdatedAt                              time.Time
}

type PMRelayInput struct {
	Content                                         json.RawMessage `json:"-"`
	AttributionReply                                bool            `json:"-"`
	ActorUserID, UserID                             string
	UpdateID, ChatID, SourceChatID, SourceMessageID int64
	ReplyToMessageID                                int64
	MessageAt                                       time.Time
	Inbound                                         bool
}

type PMModerationInput struct {
	ConversationID string `json:"conversationId"`
	Blocked        *bool  `json:"blocked,omitempty"`
	Muted          *bool  `json:"muted,omitempty"`
}

type PMTopicRepairInput struct {
	ConversationID   string `json:"conversationId"`
	TopicID          int64  `json:"topicId"`
	ProfileMessageID int64  `json:"profileMessageId"`
}

type PMNotice struct {
	ChatID         int64  `json:"chatId"`
	ReplyMessageID int64  `json:"replyMessageId"`
	Locale         string `json:"locale"`
	Reason         string `json:"reason"`
}

type PMDelivery struct {
	OperationID, Status, Direction, ErrorCode               string
	SourceChatID, SourceMessageID, ResultMessageID, TopicID int64
	CreatedAt                                               time.Time
	ReadAt                                                  *time.Time
	ReadSource                                              string
}
