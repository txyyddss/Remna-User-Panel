package httpapi

import "strings"

// normalizeTelegramCommand accepts commands addressed to this bot and removes
// Telegram's optional @botname suffix before raffle and built-in parsing.
func normalizeTelegramCommand(text, botUsername string) (string, bool) {
	value := strings.TrimSpace(text)
	fields := strings.Fields(value)
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "/") {
		return text, true
	}
	command, target, mentioned := strings.Cut(fields[0], "@")
	if !mentioned {
		return text, true
	}
	if botUsername == "" || !strings.EqualFold(target, botUsername) {
		return "", false
	}
	return command + strings.TrimPrefix(value, fields[0]), true
}
