// Package botcommands owns parsing and localized Telegram command replies.
package botcommands

import "strings"

// Name is a supported Telegram slash command without the slash prefix.
type Name string

const (
	Sub     Name = "sub"
	Balance Name = "balance"
	SignIn  Name = "signin"
	Start   Name = "start"
	MyCombo Name = "mycombo"
	Deduct  Name = "deduct"
)

// Command is a parsed slash command. Known is false for commands which must be
// excluded from analytics but are not handled by this application.
type Command struct {
	Name  Name
	Args  []string
	Known bool
}

// Parse recognizes Telegram commands with an optional @bot suffix.
func Parse(text string) (Command, bool) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "/") {
		return Command{}, false
	}
	name := strings.TrimPrefix(strings.ToLower(fields[0]), "/")
	if at := strings.IndexByte(name, '@'); at >= 0 {
		name = name[:at]
	}
	command := Command{Name: Name(name), Args: append([]string(nil), fields[1:]...)}
	switch command.Name {
	case Sub, Balance, SignIn, Start, MyCombo, Deduct:
		command.Known = true
	}
	return command, true
}

// AllowsReplyTarget reports which commands may inspect a replied-to member.
func (c Command) AllowsReplyTarget() bool {
	return c.Name == Sub || c.Name == MyCombo
}
