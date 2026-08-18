package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/botcommands"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
)

func (a *Application) configureTelegramCommands(ctx context.Context) error {
	scopes := []telegram.BotCommandScope{{Type: "all_private_chats"}}
	groupValue, err := a.settings.Optional(ctx, "telegram.group_chat_id")
	if err != nil {
		return fmt.Errorf("load command group: %w", err)
	}
	if groupID, parseErr := strconv.ParseInt(strings.TrimSpace(groupValue), 10, 64); parseErr == nil && groupID != 0 {
		scopes = append(scopes, telegram.BotCommandScope{Type: "chat", ChatID: groupID})
	}
	registrations := []struct {
		language string
		copy     botcommands.Copy
	}{
		{copy: botcommands.Text(botcommands.English)},
		{language: "zh", copy: botcommands.Text(botcommands.Chinese)},
	}
	for _, scope := range scopes {
		for _, registration := range registrations {
			commands := make([]telegram.BotCommand, 0, len(registration.copy.Descriptions))
			for _, description := range registration.copy.Descriptions {
				commands = append(commands, telegram.BotCommand{Command: string(description.Name), Description: description.Description})
			}
			if err := a.telegram.SetMyCommands(ctx, commands, scope, registration.language); err != nil {
				return fmt.Errorf("register %s commands for %s: %w", registration.language, scope.Type, err)
			}
		}
	}
	return nil
}
