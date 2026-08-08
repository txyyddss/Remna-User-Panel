package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/accounts"
	"github.com/txyyddss/Remna-User-Panel/internal/admin"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/catalog"
	"github.com/txyyddss/Remna-User-Panel/internal/entitlements"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/bepusdt"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/ezpay"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/remnawave"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

type initDataAdapter struct{ verifier *telegram.InitDataVerifier }

func (a initDataAdapter) Validate(raw string) (model.TelegramProfile, error) {
	data, err := a.verifier.Verify(raw)
	if err != nil {
		return model.TelegramProfile{}, err
	}
	return model.TelegramProfile{ID: data.User.ID, FirstName: data.User.FirstName, LastName: data.User.LastName, Username: data.User.Username}, nil
}

type telegramAdapter struct{ client *telegram.Client }

func (a telegramAdapter) CreateJoinRequestInvite(ctx context.Context, chatID, name string, expiresAt time.Time) (string, error) {
	invite, err := a.client.CreateJoinRequestInvite(ctx, chatID, name, expiresAt)
	if err != nil {
		return "", err
	}
	return invite.InviteLink, nil
}

func (a telegramAdapter) GetMembership(ctx context.Context, chatID string, telegramID int64) (bool, error) {
	member, err := a.client.GetChatMember(ctx, chatID, telegramID)
	if err != nil {
		return false, err
	}
	return member.Present(), nil
}

func (a telegramAdapter) ApproveJoinRequest(ctx context.Context, chatID string, telegramID int64) error {
	return a.client.ApproveJoinRequest(ctx, chatID, telegramID)
}

func (a telegramAdapter) RevokeInviteLink(ctx context.Context, chatID, inviteLink string) error {
	_, err := a.client.RevokeInviteLink(ctx, chatID, inviteLink)
	return err
}

type remnaAdapter struct{ settings *admin.SettingsService }

func (a remnaAdapter) client(ctx context.Context) (*remnawave.Client, error) {
	baseURL, err := a.settings.Plaintext(ctx, "remnawave.base_url")
	if err != nil {
		return nil, err
	}
	token, err := a.settings.Plaintext(ctx, "remnawave.api_token")
	if err != nil {
		return nil, err
	}
	return remnawave.NewClient(baseURL, token)
}

func (a remnaAdapter) FindUserByUsername(ctx context.Context, username string) (accounts.RemoteUser, bool, error) {
	client, err := a.client(ctx)
	if err != nil {
		return accounts.RemoteUser{}, false, err
	}
	user, exists, err := client.FindUserByUsername(ctx, username)
	if err != nil || !exists {
		return accounts.RemoteUser{}, exists, err
	}
	return mapRemoteUser(*user), true, nil
}

func (a remnaAdapter) FindUserByTelegramID(ctx context.Context, telegramID int64) (accounts.RemoteUser, bool, error) {
	client, err := a.client(ctx)
	if err != nil {
		return accounts.RemoteUser{}, false, err
	}
	user, err := client.FindUserByTelegramID(ctx, telegramID)
	if err != nil {
		return accounts.RemoteUser{}, false, err
	}
	if user == nil {
		return accounts.RemoteUser{}, false, nil
	}
	return mapRemoteUser(*user), true, nil
}

func (a remnaAdapter) CreateUser(ctx context.Context, input accounts.RemoteCreateUser) (accounts.RemoteUser, error) {
	client, err := a.client(ctx)
	if err != nil {
		return accounts.RemoteUser{}, err
	}
	user, err := client.CreateUser(ctx, remnawave.CreateUserRequest{
		Username: input.Username, Status: remnawave.UserStatus(input.Status), TrafficLimitBytes: input.TrafficLimitBytes,
		TrafficLimitStrategy: remnawave.TrafficLimitStrategy(input.TrafficLimitStrategy), ExpireAt: input.ExpireAt,
		TelegramID: input.TelegramID, ActiveInternalSquads: input.ActiveInternalSquads, ExternalSquadUUID: nil,
	})
	if err != nil {
		return accounts.RemoteUser{}, err
	}
	return mapRemoteUser(*user), nil
}

func (a remnaAdapter) IsDuplicateError(err error) bool { return remnawave.IsErrorCode(err, "A019") }

func mapRemoteUser(user remnawave.User) accounts.RemoteUser {
	return accounts.RemoteUser{ID: strconv.FormatInt(user.ID, 10), Username: user.Username, TelegramID: user.TelegramID, SubscriptionURL: user.SubscriptionURL}
}

func (a remnaAdapter) Dashboard(ctx context.Context, remoteID string) (catalog.RemoteDashboard, error) {
	client, userID, err := a.clientAndID(ctx, remoteID)
	if err != nil {
		return catalog.RemoteDashboard{}, err
	}
	user, err := client.GetUserByID(ctx, userID)
	if err != nil {
		return catalog.RemoteDashboard{}, err
	}
	stats, err := client.GetUserStats(ctx, userID, time.Now().UTC().AddDate(0, -1, 0), time.Now().UTC(), 5)
	if err != nil {
		return catalog.RemoteDashboard{}, err
	}
	mapped := model.Statistics{
		UsedTrafficBytes:     strconv.FormatInt(user.UserTraffic.UsedTrafficBytes, 10),
		LifetimeTrafficBytes: strconv.FormatInt(user.UserTraffic.LifetimeUsedTrafficBytes, 10),
		TrafficLimitBytes:    strconv.FormatInt(user.TrafficLimitBytes, 10), OnlineAt: user.UserTraffic.OnlineAt,
		Categories: stats.Categories, SparklineData: make([]string, 0, len(stats.SparklineData)), TopNodes: make([]model.TopNode, 0, len(stats.TopNodes)),
	}
	for _, sample := range stats.SparklineData {
		mapped.SparklineData = append(mapped.SparklineData, strconv.FormatInt(sample, 10))
	}
	for _, node := range stats.TopNodes {
		mapped.TopNodes = append(mapped.TopNodes, model.TopNode{UUID: node.UUID, Name: node.Name, CountryCode: node.CountryCode, TotalBytes: strconv.FormatInt(node.Total, 10)})
	}
	return catalog.RemoteDashboard{Statistics: mapped, SubscriptionURL: user.SubscriptionURL}, nil
}

func (a remnaAdapter) RevokeSubscription(ctx context.Context, remoteID string) (string, error) {
	client, userID, err := a.clientAndID(ctx, remoteID)
	if err != nil {
		return "", err
	}
	user, err := client.RevokeSubscription(ctx, userID, false)
	if err != nil {
		return "", err
	}
	return user.SubscriptionURL, nil
}

func (a remnaAdapter) ApplyEntitlement(ctx context.Context, remoteID string, trafficLimitBytes int64, resetStrategy string, squadUUIDs []string) error {
	client, userID, err := a.clientAndID(ctx, remoteID)
	if err != nil {
		return err
	}
	status := remnawave.UserStatusActive
	strategy := remnawave.TrafficLimitStrategy(resetStrategy)
	expires := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)
	_, err = client.UpdateUser(ctx, remnawave.UpdateUserRequest{ID: userID, Status: &status, TrafficLimitBytes: &trafficLimitBytes,
		TrafficLimitStrategy: &strategy, ExpireAt: &expires, ActiveInternalSquads: &squadUUIDs, ClearExternalSquad: true})
	return err
}

func (a remnaAdapter) ResetTraffic(ctx context.Context, remoteID string) error {
	client, userID, err := a.clientAndID(ctx, remoteID)
	if err != nil {
		return err
	}
	_, err = client.ResetTraffic(ctx, userID)
	return err
}

func (a remnaAdapter) RemoveEntitlement(ctx context.Context, remoteID string) error {
	client, userID, err := a.clientAndID(ctx, remoteID)
	if err != nil {
		return err
	}
	status := remnawave.UserStatusActive
	limit := int64(0)
	strategy := remnawave.TrafficNoReset
	expires := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)
	squads := []string{}
	_, err = client.UpdateUser(ctx, remnawave.UpdateUserRequest{ID: userID, Status: &status, TrafficLimitBytes: &limit,
		TrafficLimitStrategy: &strategy, ExpireAt: &expires, ActiveInternalSquads: &squads, ClearExternalSquad: true})
	return err
}

func (a remnaAdapter) ListInternalSquads(ctx context.Context) ([]admin.UpstreamSquad, error) {
	client, err := a.client(ctx)
	if err != nil {
		return nil, err
	}
	squads, err := client.ListInternalSquads(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]admin.UpstreamSquad, 0, len(squads))
	for _, squad := range squads {
		result = append(result, admin.UpstreamSquad{UUID: squad.UUID, Name: squad.Name})
	}
	return result, nil
}

func (a remnaAdapter) clientAndID(ctx context.Context, remoteID string) (*remnawave.Client, int64, error) {
	client, err := a.client(ctx)
	if err != nil {
		return nil, 0, err
	}
	userID, err := strconv.ParseInt(remoteID, 10, 64)
	if err != nil || userID <= 0 {
		return nil, 0, errors.New("invalid Remnawave user id")
	}
	return client, userID, nil
}

var _ accounts.RemnawaveClient = remnaAdapter{}
var _ catalog.RemnawaveClient = remnaAdapter{}
var _ entitlements.RemnawaveClient = remnaAdapter{}
var _ admin.SquadImporter = remnaAdapter{}

type paymentAdapter struct {
	settings *admin.SettingsService
	telegram *telegram.Client
	users    interface {
		UserByID(context.Context, string) (model.User, error)
	}
}

func (a paymentAdapter) Create(ctx context.Context, request billing.ProviderCreateRequest) (billing.ProviderCheckout, error) {
	switch request.Provider {
	case "ezpay":
		return a.createEZPay(ctx, request)
	case "bepusdt":
		return a.createBEPusdt(ctx, request)
	case "stars":
		return a.createStars(ctx, request)
	default:
		return billing.ProviderCheckout{}, errors.New("unsupported provider")
	}
}

func (a paymentAdapter) RefundProvider(ctx context.Context, order model.PaymentOrder) error {
	if order.Provider != "stars" || order.Status == "refunded" {
		return nil
	}
	if order.ProviderTradeID == nil || *order.ProviderTradeID == "" {
		return errors.New("stars payment has no Telegram charge id")
	}
	user, err := a.users.UserByID(ctx, order.UserID)
	if err != nil {
		return err
	}
	return a.telegram.RefundStarPayment(ctx, user.TelegramID, *order.ProviderTradeID)
}

func (a paymentAdapter) createEZPay(ctx context.Context, request billing.ProviderCreateRequest) (billing.ProviderCheckout, error) {
	client, err := a.ezpayClient(ctx)
	if err != nil {
		return billing.ProviderCheckout{}, err
	}
	paymentType, _ := a.settings.Optional(ctx, "billing.ezpay.payment_type")
	if paymentType == "" {
		paymentType = string(ezpay.PaymentAlipay)
	}
	checkoutURL, err := client.CheckoutURL(ezpay.CheckoutRequest{Type: ezpay.PaymentType(paymentType), NotifyURL: request.NotifyURL,
		ReturnURL: request.ReturnURL, OutTradeNo: request.OrderID, Name: "TXB balance", Money: request.PayableAmount})
	if err != nil {
		return billing.ProviderCheckout{}, err
	}
	return billing.ProviderCheckout{PaymentURL: &checkoutURL, QRPayload: &checkoutURL, PayableAmount: request.PayableAmount,
		PayableCurrency: "CNY", ProviderPayload: `{}`, ExpiresAt: time.Now().UTC().Add(30 * time.Minute)}, nil
}

func (a paymentAdapter) createBEPusdt(ctx context.Context, request billing.ProviderCreateRequest) (billing.ProviderCheckout, error) {
	client, err := a.bepusdtClient(ctx)
	if err != nil {
		return billing.ProviderCheckout{}, err
	}
	tradeType, _ := a.settings.Optional(ctx, "billing.bepusdt.trade_type")
	if tradeType == "" {
		tradeType = "usdt.trc20"
	}
	transaction, err := client.CreateTransaction(ctx, bepusdt.CreateTransactionRequest{OrderID: request.OrderID, Amount: request.PayableAmount,
		Fiat: "USD", TradeType: tradeType, Name: "TXB balance", NotifyURL: request.NotifyURL, RedirectURL: request.RedirectURL, TimeoutSeconds: 1200})
	if err != nil {
		return billing.ProviderCheckout{}, err
	}
	payload, _ := json.Marshal(transaction)
	tradeID, paymentURL, address := transaction.TradeID, transaction.PaymentURL, transaction.Token
	expiresAt := transaction.ExpiresAt(time.Now().UTC())
	return billing.ProviderCheckout{TradeID: &tradeID, PaymentURL: &paymentURL, QRPayload: &address, PayableAmount: transaction.ActualAmount,
		PayableCurrency: "USDT", ProviderPayload: string(payload), ExpiresAt: expiresAt}, nil
}

func (a paymentAdapter) createStars(ctx context.Context, request billing.ProviderCreateRequest) (billing.ProviderCheckout, error) {
	amount, err := strconv.ParseInt(request.PayableAmount, 10, 64)
	if err != nil || amount <= 0 {
		return billing.ProviderCheckout{}, errors.New("invalid Stars amount")
	}
	link, err := a.telegram.CreateStarsInvoiceLink(ctx, telegram.StarsInvoiceRequest{Title: "TXB balance", Description: "Add TXB to your TX Carpool balance",
		Payload: request.OrderID, Label: "TXB credit", Amount: amount})
	if err != nil {
		return billing.ProviderCheckout{}, err
	}
	return billing.ProviderCheckout{PaymentURL: &link, QRPayload: &link, PayableAmount: request.PayableAmount,
		PayableCurrency: "XTR", ProviderPayload: `{}`, ExpiresAt: time.Now().UTC().Add(30 * time.Minute)}, nil
}

func (a paymentAdapter) ezpayClient(ctx context.Context) (*ezpay.Client, error) {
	baseURL, err := a.settings.Plaintext(ctx, "billing.ezpay.base_url")
	if err != nil {
		return nil, err
	}
	merchantID, err := a.settings.Plaintext(ctx, "billing.ezpay.merchant_id")
	if err != nil {
		return nil, err
	}
	key, err := a.settings.Plaintext(ctx, "billing.ezpay.key")
	if err != nil {
		return nil, err
	}
	return ezpay.NewClient(baseURL, merchantID, key)
}

func (a paymentAdapter) bepusdtClient(ctx context.Context) (*bepusdt.Client, error) {
	baseURL, err := a.settings.Plaintext(ctx, "billing.bepusdt.base_url")
	if err != nil {
		return nil, err
	}
	token, err := a.settings.Plaintext(ctx, "billing.bepusdt.api_token")
	if err != nil {
		return nil, err
	}
	return bepusdt.NewClient(baseURL, token)
}

func (a paymentAdapter) VerifyEZPay(ctx context.Context, values url.Values) (billing.ProviderEvent, bool, error) {
	client, err := a.ezpayClient(ctx)
	if err != nil {
		return billing.ProviderEvent{}, false, err
	}
	notification, err := client.ParseNotification(values)
	if err != nil {
		return billing.ProviderEvent{}, false, err
	}
	digest := sha256.Sum256([]byte(values.Encode()))
	event := billing.ProviderEvent{Provider: "ezpay", OrderID: notification.OutTradeNo, TradeID: notification.TradeNo,
		PayableAmount: notification.Money, PayableCurrency: "CNY", DedupeKey: notification.TradeNo, PayloadHash: hex.EncodeToString(digest[:])}
	return event, notification.Successful(), nil
}

func (a paymentAdapter) VerifyBEPusdt(ctx context.Context, body []byte) (billing.ProviderEvent, int, error) {
	client, err := a.bepusdtClient(ctx)
	if err != nil {
		return billing.ProviderEvent{}, 0, err
	}
	webhook, err := client.ParseAndVerifyWebhook(body)
	if err != nil {
		return billing.ProviderEvent{}, 0, err
	}
	digest := sha256.Sum256(body)
	dedupe := webhook.BlockTransactionID
	if dedupe == "" {
		dedupe = webhook.TradeID + ":" + strconv.Itoa(webhook.Status)
	}
	event := billing.ProviderEvent{Provider: "bepusdt", OrderID: webhook.OrderID, TradeID: webhook.TradeID,
		ChargeID: webhook.BlockTransactionID, PayableAmount: webhook.ActualAmount, PayableCurrency: "USDT",
		FiatAmount: webhook.Amount, FiatCurrency: "USD", Recipient: webhook.Token, DedupeKey: dedupe, PayloadHash: hex.EncodeToString(digest[:])}
	return event, webhook.Status, nil
}
