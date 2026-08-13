// Package onboarding validates localized welcome and agreement bundles and
// computes presentation timing without storing derived animation values.
package onboarding

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	KindWelcome    = "welcome"
	KindAgreements = "agreements"
)

var (
	ErrInvalidContent = errors.New("invalid onboarding content")
	contentIDPattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
	agreementIcons    = map[string]struct{}{"link-break": {}, "shield-check": {}, "users-three": {}, "warning": {}, "lock-key": {}, "heart": {}, "scales": {}}
	agreementColors   = map[string]struct{}{"accent": {}, "success": {}, "warning": {}, "danger": {}, "neutral": {}}
)

type WelcomeMessage struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	DurationMS int    `json:"durationMs,omitempty"`
}

type Agreement struct {
	ID        string `json:"id"`
	Icon      string `json:"icon"`
	Color     string `json:"color,omitempty"`
	PageTitle string `json:"pageTitle,omitempty"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

type Content map[string]json.RawMessage

type Bundle struct {
	Kind              string  `json:"kind"`
	Draft             Content `json:"draft"`
	Published         Content `json:"published"`
	DraftRevision     int     `json:"draftRevision"`
	PublishedRevision int     `json:"publishedRevision"`
}

type Published struct {
	Locale            string           `json:"locale"`
	WelcomeRevision   int              `json:"welcomeRevision"`
	AgreementRevision int              `json:"agreementRevision"`
	Welcome           []WelcomeMessage `json:"welcome"`
	Agreements        []Agreement      `json:"agreements"`
}

func Validate(kind string, content Content) error {
	if len(content) != 2 || content["en"] == nil || content["zh-CN"] == nil {
		return fmt.Errorf("%w: en and zh-CN are required", ErrInvalidContent)
	}
	switch kind {
	case KindWelcome:
		var english, chinese []WelcomeMessage
		if json.Unmarshal(content["en"], &english) != nil || json.Unmarshal(content["zh-CN"], &chinese) != nil {
			return fmt.Errorf("%w: welcome locales must be arrays", ErrInvalidContent)
		}
		if len(english) == 0 || len(english) > 20 || len(english) != len(chinese) {
			return fmt.Errorf("%w: welcome ordering differs", ErrInvalidContent)
		}
		seen := make(map[string]struct{}, len(english))
		for index := range english {
			if !contentIDPattern.MatchString(english[index].ID) || english[index].ID != chinese[index].ID ||
				strings.TrimSpace(english[index].Text) == "" || strings.TrimSpace(chinese[index].Text) == "" ||
				len(english[index].Text) > 1000 || len(chinese[index].Text) > 1000 {
				return fmt.Errorf("%w: invalid welcome message", ErrInvalidContent)
			}
			if _, duplicate := seen[english[index].ID]; duplicate {
				return fmt.Errorf("%w: duplicate welcome message id", ErrInvalidContent)
			}
			seen[english[index].ID] = struct{}{}
			english[index].DurationMS = 0
			chinese[index].DurationMS = 0
		}
		content["en"], _ = json.Marshal(english)
		content["zh-CN"], _ = json.Marshal(chinese)
	case KindAgreements:
		var english, chinese []Agreement
		if json.Unmarshal(content["en"], &english) != nil || json.Unmarshal(content["zh-CN"], &chinese) != nil {
			return fmt.Errorf("%w: agreement locales must be arrays", ErrInvalidContent)
		}
		if len(english) == 0 || len(english) > 30 || len(english) != len(chinese) {
			return fmt.Errorf("%w: agreement ordering differs", ErrInvalidContent)
		}
		seen := make(map[string]struct{}, len(english))
		for index := range english {
			_, iconOK := agreementIcons[english[index].Icon]
			_, translatedIconOK := agreementIcons[chinese[index].Icon]
			if english[index].Color == "" {
				english[index].Color = "warning"
			}
			if chinese[index].Color == "" {
				chinese[index].Color = "warning"
			}
			_, colorOK := agreementColors[english[index].Color]
			_, translatedColorOK := agreementColors[chinese[index].Color]
			if !contentIDPattern.MatchString(english[index].ID) || english[index].ID != chinese[index].ID || english[index].Icon != chinese[index].Icon ||
				!iconOK || !translatedIconOK || english[index].Color != chinese[index].Color || !colorOK || !translatedColorOK ||
				strings.TrimSpace(english[index].Title) == "" || strings.TrimSpace(chinese[index].Title) == "" ||
				strings.TrimSpace(english[index].Body) == "" || strings.TrimSpace(chinese[index].Body) == "" ||
				len(english[index].Title) > 200 || len(chinese[index].Title) > 200 || len(english[index].Body) > 2000 || len(chinese[index].Body) > 2000 {
				return fmt.Errorf("%w: invalid agreement", ErrInvalidContent)
			}
			if index == 0 {
				english[index].PageTitle = strings.TrimSpace(english[index].PageTitle)
				chinese[index].PageTitle = strings.TrimSpace(chinese[index].PageTitle)
				if (english[index].PageTitle == "") != (chinese[index].PageTitle == "") ||
					len(english[index].PageTitle) > 200 || len(chinese[index].PageTitle) > 200 {
					return fmt.Errorf("%w: invalid agreement page title", ErrInvalidContent)
				}
			} else if english[index].PageTitle != "" || chinese[index].PageTitle != "" {
				return fmt.Errorf("%w: agreement page title must be first", ErrInvalidContent)
			}
			if _, duplicate := seen[english[index].ID]; duplicate {
				return fmt.Errorf("%w: duplicate agreement id", ErrInvalidContent)
			}
			seen[english[index].ID] = struct{}{}
		}
		content["en"], _ = json.Marshal(english)
		content["zh-CN"], _ = json.Marshal(chinese)
	default:
		return fmt.Errorf("%w: unknown bundle kind", ErrInvalidContent)
	}
	return nil
}

func WelcomeForLocale(content Content, locale string) ([]WelcomeMessage, error) {
	locale = SupportedLocale(locale)
	var messages []WelcomeMessage
	if err := json.Unmarshal(content[locale], &messages); err != nil {
		return nil, err
	}
	for index := range messages {
		messages[index].DurationMS = MessageDurationMS(messages[index].Text)
	}
	return messages, nil
}

func AgreementsForLocale(content Content, locale string) ([]Agreement, error) {
	locale = SupportedLocale(locale)
	var agreements []Agreement
	if err := json.Unmarshal(content[locale], &agreements); err != nil {
		return nil, err
	}
	return agreements, nil
}

func AgreementIDs(content Content) ([]string, error) {
	agreements, err := AgreementsForLocale(content, "en")
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(agreements))
	for _, agreement := range agreements {
		ids = append(ids, agreement.ID)
	}
	return ids, nil
}

func SupportedLocale(locale string) string {
	if strings.EqualFold(strings.TrimSpace(locale), "zh-CN") || strings.HasPrefix(strings.ToLower(strings.TrimSpace(locale)), "zh") {
		return "zh-CN"
	}
	return "en"
}

// MessageDurationMS combines Latin word and CJK character reading time, adds
// the fixed transition allowance, and clamps to 1.8-12 seconds.
