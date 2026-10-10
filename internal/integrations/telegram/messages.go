package telegram

import "encoding/json"

// Message retains only the wire fields needed for existing commands and relay eligibility.
// Media stays in Telegram: its presence is checked here, then only IDs are queued.
type Message struct {
	Entities              []MessageEntity    `json:"entities,omitempty"`
	Caption               string             `json:"caption,omitempty"`
	CaptionEntities       []MessageEntity    `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool               `json:"show_caption_above_media,omitempty"`
	MessageID             int64              `json:"message_id"`
	MessageThreadID       int64              `json:"message_thread_id,omitempty"`
	SenderBoostCount      int                `json:"sender_boost_count,omitempty"`
	From                  *User              `json:"from,omitempty"`
	Chat                  Chat               `json:"chat"`
	Date                  int64              `json:"date"`
	Text                  string             `json:"text,omitempty"`
	ReplyToMessage        *Message           `json:"reply_to_message,omitempty"`
	SuccessfulPayment     *SuccessfulPayment `json:"successful_payment,omitempty"`
	RefundedPayment       *RefundedPayment   `json:"refunded_payment,omitempty"`
	Photo                 json.RawMessage    `json:"photo,omitempty"`
	Video                 json.RawMessage    `json:"video,omitempty"`
	Audio                 json.RawMessage    `json:"audio,omitempty"`
	Voice                 json.RawMessage    `json:"voice,omitempty"`
	VideoNote             json.RawMessage    `json:"video_note,omitempty"`
	Document              json.RawMessage    `json:"document,omitempty"`
	Animation             json.RawMessage    `json:"animation,omitempty"`
	Sticker               json.RawMessage    `json:"sticker,omitempty"`
	Contact               json.RawMessage    `json:"contact,omitempty"`
	Location              json.RawMessage    `json:"location,omitempty"`
	Venue                 json.RawMessage    `json:"venue,omitempty"`
	Dice                  json.RawMessage    `json:"dice,omitempty"`
	LivePhoto             json.RawMessage    `json:"live_photo,omitempty"`
	Game                  json.RawMessage    `json:"game,omitempty"`
	Story                 json.RawMessage    `json:"story,omitempty"`
	Checklist             json.RawMessage    `json:"checklist,omitempty"`
	Poll                  *RelayPoll         `json:"poll,omitempty"`
	Invoice               json.RawMessage    `json:"invoice,omitempty"`
}

type RelayPoll struct {
	Type             string `json:"type"`
	CorrectOptionIDs []int  `json:"correct_option_ids,omitempty"`
}

func (m Message) RelayContent() bool {
	if m.SuccessfulPayment != nil || m.RefundedPayment != nil || len(m.Invoice) > 0 {
		return false
	}
	if m.Poll != nil && m.Poll.Type == "quiz" && len(m.Poll.CorrectOptionIDs) == 0 {
		return false
	}
	return m.UserContent()
}

func (m Message) UserContent() bool {
	if m.Poll != nil {
		return true
	}
	if m.Text != "" {
		return true
	}
	for _, media := range []json.RawMessage{m.Photo, m.Video, m.Audio, m.Voice, m.VideoNote, m.Document, m.Animation, m.Sticker, m.Contact, m.Location, m.Venue, m.Dice, m.LivePhoto, m.Game, m.Story, m.Checklist} {
		if len(media) > 0 && string(media) != "null" && string(media) != "[]" {
			return true
		}
	}
	return false
}
