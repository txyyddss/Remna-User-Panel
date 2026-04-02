package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	tgbot "github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
	"github.com/user/remna-user-panel/internal/services"
)

// Bot is the Telegram bot handler
type Bot struct {
	bot      *tgbot.Bot
	credit   *services.CreditService
	ipChange *services.IPChangeService
}

// NewBot creates and configures the Telegram bot
func NewBot(credit *services.CreditService, ipChange *services.IPChangeService) (*Bot, error) {
	cfg := config.Get()
	if cfg.Telegram.BotToken == "" {
		slog.Info("telegram: bot token not configured, skipping bot initialization")
		return nil, nil
	}

	b := &Bot{credit: credit, ipChange: ipChange}

	opts := []tgbot.Option{
		tgbot.WithDefaultHandler(b.handleMessage),
		tgbot.WithCallbackQueryDataHandler("agree:", tgbot.MatchTypePrefix, b.handleIPChangeVote),
		tgbot.WithCallbackQueryDataHandler("decline:", tgbot.MatchTypePrefix, b.handleIPChangeVote),
	}

	bot, err := tgbot.New(cfg.Telegram.BotToken, opts...)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	// Register commands
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/signup", tgbot.MatchTypePrefix, b.handleSignup)
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/bet", tgbot.MatchTypePrefix, b.handleBet)
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/sub", tgbot.MatchTypePrefix, b.handleSub)
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypePrefix, b.handleStart)

	b.bot = bot
	return b, nil
}

// Start starts the bot polling
func (b *Bot) Start(ctx context.Context) {
	if b == nil || b.bot == nil {
		return
	}

	// Delete any existing webhook to prevent conflict with getUpdates
	_, err := b.bot.DeleteWebhook(ctx, &tgbot.DeleteWebhookParams{
		DropPendingUpdates: true,
	})
	if err != nil {
		slog.Warn("telegram: failed to delete webhook (can be ignored if none exists)", "error", err)
	}

	slog.Info("telegram: bot started")
	b.bot.Start(ctx)
}

// --- Command Handlers ---

func (b *Bot) handleStart(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "🎉 Welcome to the panel Bot!\n\nAvailable commands:\n/signup - Daily check-in\n/bet <amount> - Bet\n/sub - View subscription status",
	})
}

func (b *Bot) handleSignup(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	telegramID := update.Message.From.ID
	userID := b.getOrCreateUserID(telegramID, update.Message.From.FirstName)
	if userID == 0 {
		return
	}

	value, newBalance, err := b.credit.Signup(ctx, userID)
	if err != nil {
		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ " + err.Error(),
			ReplyParameters: &tgmodels.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		return
	}

	text := fmt.Sprintf("🎁 Check-in successful!\nEarned: +%.2f TXB\nCurrent balance: %.2f TXB", value, newBalance)
	msg, _ := bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   text,
		ReplyParameters: &tgmodels.ReplyParameters{
			MessageID: update.Message.ID,
		},
	})

	// Auto-delete if value < 1
	if value < 1 && msg != nil {
		go func() {
			time.Sleep(10 * time.Second)
			bot.DeleteMessage(ctx, &tgbot.DeleteMessageParams{
				ChatID:    update.Message.Chat.ID,
				MessageID: msg.ID,
			})
		}()
	}
}

func (b *Bot) handleBet(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	telegramID := update.Message.From.ID
	userID := b.getOrCreateUserID(telegramID, update.Message.From.FirstName)
	if userID == 0 {
		return
	}

	// Parse amount from /bet <amount>
	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Usage: /bet <amount>",
			ReplyParameters: &tgmodels.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		return
	}

	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || amount <= 0 {
		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Please enter a valid amount",
			ReplyParameters: &tgmodels.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		return
	}

	result, newBalance, err := b.credit.Bet(ctx, userID, amount)
	if err != nil {
		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ " + err.Error(),
			ReplyParameters: &tgmodels.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		return
	}

	var emoji string
	if result > 0 {
		emoji = "🎉"
	} else {
		emoji = "💸"
	}

	text := fmt.Sprintf("%s Bet result: %+.2f TXB\nCurrent balance: %.2f TXB", emoji, result, newBalance)
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   text,
		ReplyParameters: &tgmodels.ReplyParameters{
			MessageID: update.Message.ID,
		},
	})
}

func (b *Bot) handleSub(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	cfg := config.Get()

	// Check if replying to someone else's message
	targetID := update.Message.From.ID
	if update.Message.ReplyToMessage != nil {
		targetID = update.Message.ReplyToMessage.From.ID
	}

	// Get user's Remnawave UUID
	var rwUUID string
	database.DB().QueryRow("SELECT remnawave_uuid FROM users WHERE telegram_id = ?", targetID).Scan(&rwUUID)

	if rwUUID == "" {
		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Subscription info not found",
			ReplyParameters: &tgmodels.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		return
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	rwUser, err := rwClient.GetUserByUUID(rwUUID)
	if err != nil {
		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Failed to fetch subscription info",
			ReplyParameters: &tgmodels.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		return
	}

	text := b.formatSubMessage(rwUser, rwClient)
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: tgmodels.ParseModeMarkdown,
		ReplyParameters: &tgmodels.ReplyParameters{
			MessageID: update.Message.ID,
		},
	})
}

func (b *Bot) formatSubMessage(user *remnawave.UserData, client *remnawave.Client) string {
	// Status emoji
	var statusEmoji string
	switch user.Status {
	case "ACTIVE":
		statusEmoji = "🟢"
	case "LIMITED":
		statusEmoji = "🟡"
	default:
		statusEmoji = "🔴"
	}

	// Traffic progress
	var usedPercent float64
	usedStr := formatBytes(user.UsedTrafficBytes)
	limitStr := "♾️"
	if user.TrafficLimitBytes > 0 {
		usedPercent = float64(user.UsedTrafficBytes) / float64(user.TrafficLimitBytes) * 100
		limitStr = formatBytes(user.TrafficLimitBytes)
	}
	progressBar := generateProgressBar(usedPercent, 8)

	// Days
	daysRemaining := int(time.Until(user.ExpireAt).Hours() / 24)
	daysSinceCreation := int(time.Since(user.CreatedAt).Hours() / 24)

	// Node usage (get bandwidth stats)
	nodeUsage := ""
	start := time.Now().AddDate(0, -1, 0).Format(time.RFC3339)
	end := time.Now().Format(time.RFC3339)
	stats, err := client.GetUserBandwidthStats(user.UUID, start, end)
	if err == nil && len(stats) > 0 {
		// Aggregate by node
		nodeMap := make(map[string]int64)
		nodeNames := make(map[string]string)
		nodeCodes := make(map[string]string)
		var totalTraffic int64

		for _, s := range stats {
			nodeMap[s.NodeUUID] += s.Total
			nodeNames[s.NodeUUID] = s.NodeName
			nodeCodes[s.NodeUUID] = s.CountryCode
			totalTraffic += s.Total
		}

		// Sort by usage
		var nodes []nodeEntry
		for uuid, total := range nodeMap {
			nodes = append(nodes, nodeEntry{uuid, nodeNames[uuid], nodeCodes[uuid], total})
		}
		sort.Slice(nodes, func(i, j int) bool { return nodes[i].Total > nodes[j].Total })

		// Generate node usage bar
		bars := []string{"▓", "░", "█", "▒", "▇"}
		totalBar := generateNodeBar(nodes, totalTraffic, 30)
		nodeUsage = fmt.Sprintf("%s %s\n\n", totalBar, formatBytes(totalTraffic))

		// Top 5 nodes
		for i, n := range nodes {
			if i >= 5 {
				break
			}
			pct := float64(0)
			if totalTraffic > 0 {
				pct = float64(n.Total) / float64(totalTraffic) * 100
			}
			bar := bars[i%len(bars)]
			nodeUsage += fmt.Sprintf("%s %s (%s) - %s (%.1f%%)\n", bar, n.Name, strings.ToUpper(n.Country), formatBytes(n.Total), pct)
		}
	}

	text := fmt.Sprintf("📊 My Subscription %s\n\n%s %d%% | %s/%s\n📅 Remaining %d days · Joined %d days\n",
		statusEmoji, progressBar, int(usedPercent), usedStr, limitStr, daysRemaining, daysSinceCreation)

	if nodeUsage != "" {
		text += nodeUsage
	}

	return text
}

func (b *Bot) handleIPChangeVote(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	if update.CallbackQuery == nil || b.ipChange == nil {
		return
	}

	parts := strings.SplitN(update.CallbackQuery.Data, ":", 2)
	if len(parts) != 2 {
		return
	}
	action, username := parts[0], parts[1]

	req, err := b.ipChange.ProcessVote(ctx, action, username, update.CallbackQuery.From.ID)
	answerText := "Vote accepted"
	if action == "agree" {
		answerText = "Voted Agree"
	} else if action == "decline" {
		answerText = "Voted Decline"
	}

	if err != nil {
		var apiErr *services.IPChangeAPIError
		if errors.As(err, &apiErr) {
			answerText = apiErr.Message
		} else {
			answerText = "Failed to process vote"
			slog.Error("telegram: ip change vote failed", "error", err)
		}

		_, _ = bot.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            answerText,
			ShowAlert:       false,
		})
		return
	}

	_, _ = bot.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            answerText,
		ShowAlert:       false,
	})

	if req == nil || update.CallbackQuery.Message.Message == nil {
		return
	}

	params := &tgbot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      b.ipChange.BuildTelegramMessageText(req),
		ParseMode: tgmodels.ParseModeHTML,
	}
	if req.Status == "PENDING" {
		params.ReplyMarkup = buildIPChangeKeyboard(req)
	}
	if _, err := bot.EditMessageText(ctx, params); err != nil {
		slog.Warn("telegram: failed to edit ip change vote message", "request_id", req.ID, "error", err)
	}
}

func buildIPChangeKeyboard(req *services.IPChangeRequestRecord) *tgmodels.InlineKeyboardMarkup {
	return &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{
				{
					Text:         fmt.Sprintf("Agree (%d)", req.AgreeCount),
					CallbackData: "agree:" + req.Username,
				},
				{
					Text:         fmt.Sprintf("Decline (%d)", req.DeclineCount),
					CallbackData: "decline:" + req.Username,
				},
			},
		},
	}
}

// --- Group Message Handler ---

func (b *Bot) handleMessage(ctx context.Context, bot *tgbot.Bot, update *tgmodels.Update) {
	if update.Message == nil {
		return
	}

	cfg := config.Get()
	msg := update.Message

	// Only collect group text messages
	if msg.Chat.Type != "group" && msg.Chat.Type != "supergroup" {
		return
	}

	// Skip bot messages
	if msg.From.IsBot {
		return
	}

	// Skip messages with media
	if msg.Photo != nil || msg.Video != nil || msg.Document != nil || msg.Audio != nil || msg.Voice != nil || msg.Sticker != nil {
		return
	}

	// Skip empty text
	if msg.Text == "" {
		return
	}

	// Skip commands
	if strings.HasPrefix(msg.Text, "/") {
		return
	}

	if !cfg.AI.Enabled {
		return
	}

	userID := b.getOrCreateUserID(msg.From.ID, msg.From.FirstName)
	if userID == 0 {
		return
	}

	// Store message
	database.DB().Exec(
		"INSERT INTO group_messages (user_id, telegram_msg_id, telegram_name, text, created_at) VALUES (?, ?, ?, ?, ?)",
		userID, msg.ID, msg.From.FirstName, msg.Text, time.Now(),
	)

	// Check if batch is ready
	var count int
	database.DB().QueryRow("SELECT COUNT(*) FROM group_messages").Scan(&count)

	if count >= cfg.AI.MessageBatchSize {
		go b.evaluateMessages(ctx, bot, msg.Chat.ID)
	}

	// Check leaderboard interval
	var totalProcessed int
	database.DB().QueryRow("SELECT COALESCE(SUM(1), 0) FROM credit_logs WHERE reason LIKE 'group chat score%'").Scan(&totalProcessed)
	if totalProcessed > 0 && totalProcessed%cfg.AI.LeaderboardInterval == 0 {
		go b.sendLeaderboard(ctx, bot, msg.Chat.ID)
	}
}

func (b *Bot) evaluateMessages(ctx context.Context, bot *tgbot.Bot, chatID int64) {
	cfg := config.Get()

	rows, err := database.DB().Query("SELECT id, user_id, telegram_name, text FROM group_messages ORDER BY id ASC")
	if err != nil {
		slog.Error("ai: query messages error", "error", err)
		return
	}
	defer rows.Close()

	type msg struct {
		ID   int64
		User int64
		Name string
		Text string
	}
	var messages []msg
	for rows.Next() {
		var m msg
		if err := rows.Scan(&m.ID, &m.User, &m.Name, &m.Text); err != nil {
			slog.Error("ai: scan message row", "error", err)
			continue
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		slog.Error("ai: message iteration error", "error", err)
	}

	if len(messages) == 0 {
		return
	}

	// Build prompt for AI
	var msgList strings.Builder
	for _, m := range messages {
		msgList.WriteString(fmt.Sprintf("ID:%d [%s]: %s\n", m.ID, m.Name, m.Text))
	}

	prompt := fmt.Sprintf(`You are a group chat message evaluation system. Please evaluate the value of the following group chat messages and score each one.

Scoring rules:
- Valuable shares (tech articles, resource recommendations, etc.): +2.0 ~ +3.0
- Answering questions, helping others: +1.0 ~ +2.0
- Normal chat, asking questions: +0.0 ~ +1.0
- Meaningless messages, spam: -0.5 ~ 0
- Attacking others, ads, inappropriate content: -1.0 ~ -2.0

Score range: %.1f ~ %.1f

Return only JSON format, no other text:
[{"id": <message ID>, "score": <score>}]

Message list:
%s`, cfg.AI.CreditMin, cfg.AI.CreditMax, msgList.String())

	// Call AI
	aiResp, err := callAI(cfg.AI.BaseURL, cfg.AI.APIKey, cfg.AI.Model, prompt)
	if err != nil {
		slog.Error("ai: evaluation error", "error", err)
		return
	}

	// Parse response
	var scores []struct {
		ID    int64   `json:"id"`
		Score float64 `json:"score"`
	}
	if err := json.Unmarshal([]byte(aiResp), &scores); err != nil {
		slog.Error("ai: parse scores error", "response", aiResp, "error", err)
		return
	}

	// Apply credits
	msgUserMap := make(map[int64]int64) // msgID -> userID
	for _, m := range messages {
		msgUserMap[m.ID] = m.User
	}

	for _, s := range scores {
		userID, ok := msgUserMap[s.ID]
		if !ok {
			continue
		}
		score := math.Round(s.Score*10) / 10 // 1 decimal place
		if score < cfg.AI.CreditMin {
			score = cfg.AI.CreditMin
		}
		if score > cfg.AI.CreditMax {
			score = cfg.AI.CreditMax
		}
		b.credit.AddCredit(ctx, userID, score, fmt.Sprintf("group chat score %+.1f", score))
	}

	// Delete processed messages
	for _, m := range messages {
		database.DB().Exec("DELETE FROM group_messages WHERE id = ?", m.ID)
	}
}

func (b *Bot) sendLeaderboard(ctx context.Context, bot *tgbot.Bot, chatID int64) {
	rows, err := database.DB().Query(`
		SELECT u.telegram_name, SUM(c.amount) as total
		FROM credit_logs c
		JOIN users u ON u.id = c.user_id
		WHERE c.reason LIKE 'group chat score%'
		AND c.created_at >= datetime('now', '-7 days')
		GROUP BY c.user_id
		ORDER BY total DESC
		LIMIT 5
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	var text strings.Builder
	text.WriteString("📊 Group Chat Leaderboard (last 7 days)\n\n")

	medals := []string{"🥇", "🥈", "🥉", "4️⃣", "5️⃣"}
	i := 0
	for rows.Next() {
		var name string
		var total float64
		if err := rows.Scan(&name, &total); err != nil {
			slog.Error("telegram: scan leaderboard row", "error", err)
			continue
		}
		text.WriteString(fmt.Sprintf("%s %s: %+.1f TXB\n", medals[i], name, total))
		i++
	}
	if err := rows.Err(); err != nil {
		slog.Error("telegram: leaderboard iteration error", "error", err)
	}

	if i > 0 {
		bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   text.String(),
		})
	}
}

// --- Utility ---

func (b *Bot) getOrCreateUserID(telegramID int64, name string) int64 {
	var userID int64
	err := database.DB().QueryRow("SELECT id FROM users WHERE telegram_id = ?", telegramID).Scan(&userID)
	if err != nil {
		result, err := database.DB().Exec("INSERT INTO users (telegram_id, telegram_name) VALUES (?, ?)", telegramID, name)
		if err != nil {
			return 0
		}
		userID, _ = result.LastInsertId()
	}
	return userID
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	val := float64(b) / float64(div)
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.2f %s", val, units[exp])
}

func generateProgressBar(percent float64, width int) string {
	if percent > 100 {
		percent = 100
	}
	filled := int(percent / 100 * float64(width))
	var bar strings.Builder
	for i := 0; i < width; i++ {
		if i < filled {
			if percent < 50 {
				bar.WriteString("🟩")
			} else if percent < 80 {
				bar.WriteString("🟨")
			} else {
				bar.WriteString("🟥")
			}
		} else {
			bar.WriteString("⬜️")
		}
	}
	return bar.String()
}

type nodeEntry struct {
	UUID    string
	Name    string
	Country string
	Total   int64
}

func generateNodeBar(nodes []nodeEntry, total int64, width int) string {
	bars := []string{"▓", "░", "█", "▒", "▇"}
	var result strings.Builder
	result.WriteString("[")
	for i, n := range nodes {
		if i >= 5 {
			break
		}
		pct := float64(n.Total) / float64(total)
		count := int(pct * float64(width))
		if count < 1 && n.Total > 0 {
			count = 1
		}
		for j := 0; j < count; j++ {
			result.WriteString(bars[i%len(bars)])
		}
	}
	// Fill remainder
	current := result.Len() - 1 // subtract opening bracket
	for current < width {
		result.WriteString("░")
		current++
	}
	result.WriteString("]")
	return result.String()
}

func callAI(baseURL, apiKey, model, prompt string) (string, error) {
	body := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a group chat message value evaluator. Return only JSON format scores."},
			{"role": "user", "content": prompt},
		},
		"stream": false,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	content := result.Choices[0].Message.Content
	// Clean up markdown code blocks if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	return content, nil
}
