package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/admin"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/bepusdt"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/ezpay"
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
	client, err := a.ezpayClient(ctx, request.ProfileID, request.Rail)
	if err != nil {
		return billing.ProviderCheckout{}, err
	}
	checkoutURL, err := client.CheckoutURL(ezpay.CheckoutRequest{Type: ezpay.PaymentType(request.Rail), NotifyURL: request.NotifyURL,
		ReturnURL: request.ReturnURL, OutTradeNo: request.OrderID, Name: "TXB balance", Money: request.PayableAmount})
	if err != nil {
		return billing.ProviderCheckout{}, err
	}
	return billing.ProviderCheckout{PaymentURL: &checkoutURL, PayableAmount: request.PayableAmount,
		PayableCurrency: "CNY", ExpiresAt: time.Now().UTC().Add(30 * time.Minute)}, nil
}

func (a paymentAdapter) createBEPusdt(ctx context.Context, request billing.ProviderCreateRequest) (billing.ProviderCheckout, error) {
	client, err := a.bepusdtClient(ctx, request.ProfileID, request.Rail)
	if err != nil {
		return billing.ProviderCheckout{}, err
	}
	transaction, err := client.CreateTransaction(ctx, bepusdt.CreateTransactionRequest{OrderID: request.OrderID, Amount: request.PayableAmount,
		Fiat: "USD", TradeType: request.Rail, Name: "TXB balance", NotifyURL: request.NotifyURL, RedirectURL: request.RedirectURL, TimeoutSeconds: 1200})
	if err != nil {
		return billing.ProviderCheckout{}, err
	}
	return mapBEPusdtCheckout(transaction, request.PayableAmount, time.Now().UTC()), nil
}

func mapBEPusdtCheckout(transaction *bepusdt.Transaction, payableAmount string, createdAt time.Time) billing.ProviderCheckout {
	tradeID, paymentURL, address := transaction.TradeID, transaction.PaymentURL, transaction.Token
	actualAmount, actualCurrency := transaction.ActualAmount, "USDT"
	expiresAt := transaction.ExpiresAt(createdAt)
	return billing.ProviderCheckout{TradeID: &tradeID, PaymentURL: &paymentURL, ReceivingAddress: &address,
		ActualCryptoAmount: &actualAmount, ActualCryptoCurrency: &actualCurrency, PayableAmount: payableAmount,
		PayableCurrency: "USD", ExpiresAt: expiresAt}
}

func (a paymentAdapter) Cancel(ctx context.Context, order model.PaymentOrder) error {
	if order.Provider != "bepusdt" || order.ProviderTradeID == nil || *order.ProviderTradeID == "" {
		return nil
	}
	_, profileID, rail, parseErr := billing.ParseMethodSelection(order.MethodID)
	if parseErr != nil {
		profileID, rail = "", order.ProviderRail
	}
	client, err := a.bepusdtClient(ctx, profileID, rail)
	if err != nil {
		return err
	}
	return client.CancelTransaction(ctx, *order.ProviderTradeID)
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
	return billing.ProviderCheckout{PaymentURL: &link, PayableAmount: request.PayableAmount,
		PayableCurrency: "XTR", ExpiresAt: time.Now().UTC().Add(30 * time.Minute)}, nil
}

func (a paymentAdapter) ezpayClient(ctx context.Context, profileID, rail string) (*ezpay.Client, error) {
	var profile model.PaymentProfileRuntime
	var err error
	if profileID != "" {
		profile, err = a.settings.PaymentProfileByID(ctx, profileID, rail)
	} else {
		profile, err = a.settings.PaymentProfile(ctx, "ezpay", rail)
	}
	if err != nil {
		baseURL, fallbackErr := a.settings.Plaintext(ctx, "billing.ezpay.base_url")
		if fallbackErr != nil {
			return nil, err
		}
		merchantID, merchantErr := a.settings.Plaintext(ctx, "billing.ezpay.merchant_id")
		if merchantErr != nil {
			return nil, merchantErr
		}
		key, keyErr := a.settings.Plaintext(ctx, "billing.ezpay.key")
		if keyErr != nil {
			return nil, keyErr
		}
		return ezpay.NewClient(baseURL, merchantID, key)
	}
	return ezpay.NewClient(profile.Endpoint, profile.MerchantID, profile.CredentialPlaintext)
}

func (a paymentAdapter) bepusdtClient(ctx context.Context, profileID, rail string) (*bepusdt.Client, error) {
	var profile model.PaymentProfileRuntime
	var err error
	if profileID != "" {
		profile, err = a.settings.PaymentProfileByID(ctx, profileID, rail)
	} else {
		profile, err = a.settings.PaymentProfile(ctx, "bepusdt", rail)
	}
	if err != nil {
		baseURL, fallbackErr := a.settings.Plaintext(ctx, "billing.bepusdt.base_url")
		if fallbackErr != nil {
			return nil, err
		}
		token, tokenErr := a.settings.Plaintext(ctx, "billing.bepusdt.api_token")
		if tokenErr != nil {
			return nil, tokenErr
		}
		return bepusdt.NewClient(baseURL, token)
	}
	return bepusdt.NewClient(profile.Endpoint, profile.CredentialPlaintext)
}

func (a paymentAdapter) VerifyEZPay(ctx context.Context, values url.Values) (billing.ProviderEvent, bool, error) {
	if profiles, profilesErr := a.settings.PaymentProfileRuntimes(ctx, "ezpay"); profilesErr == nil && len(profiles) > 0 {
		var lastErr error
		for _, profile := range profiles {
			client, clientErr := ezpay.NewClient(profile.Endpoint, profile.MerchantID, profile.CredentialPlaintext)
			if clientErr != nil {
				lastErr = clientErr
				continue
			}
			notification, parseErr := client.ParseNotification(values)
			if parseErr != nil {
				lastErr = parseErr
				continue
			}
			return ezpayEvent(notification, values, profile.ID), notification.Successful(), nil
		}
		if lastErr != nil {
			return billing.ProviderEvent{}, false, lastErr
		}
	}
	client, err := a.ezpayClient(ctx, "", values.Get("type"))
	if err != nil {
		return billing.ProviderEvent{}, false, err
	}
	notification, err := client.ParseNotification(values)
	if err != nil {
		return billing.ProviderEvent{}, false, err
	}
	return ezpayEvent(notification, values, ""), notification.Successful(), nil
}

func ezpayEvent(notification *ezpay.Notification, values url.Values, profileID string) billing.ProviderEvent {
	digest := sha256.Sum256([]byte(values.Encode()))
	return billing.ProviderEvent{Provider: "ezpay", ProfileID: profileID, Rail: string(notification.Type), OrderID: notification.OutTradeNo, TradeID: notification.TradeNo,
		PayableAmount: notification.Money, PayableCurrency: "CNY", DedupeKey: notification.TradeNo, PayloadHash: hex.EncodeToString(digest[:])}
}

func (a paymentAdapter) VerifyBEPusdt(ctx context.Context, body []byte) (billing.ProviderEvent, int, error) {
	if profiles, profilesErr := a.settings.PaymentProfileRuntimes(ctx, "bepusdt"); profilesErr == nil && len(profiles) > 0 {
		var lastErr error
		for _, profile := range profiles {
			client, clientErr := bepusdt.NewClient(profile.Endpoint, profile.CredentialPlaintext)
			if clientErr != nil {
				lastErr = clientErr
				continue
			}
			webhook, verifyErr := client.ParseAndVerifyWebhook(body)
			if verifyErr != nil {
				lastErr = verifyErr
				continue
			}
			return bepusdtEvent(webhook, body, profile.ID)
		}
		if lastErr != nil {
			return billing.ProviderEvent{}, 0, lastErr
		}
	}
	client, err := a.bepusdtClient(ctx, "", "usdt.trc20")
	if err != nil {
		return billing.ProviderEvent{}, 0, err
	}
	webhook, err := client.ParseAndVerifyWebhook(body)
	if err != nil {
		return billing.ProviderEvent{}, 0, err
	}
	return bepusdtEvent(webhook, body, "")
}

func bepusdtEvent(webhook *bepusdt.Webhook, body []byte, profileID string) (billing.ProviderEvent, int, error) {
	digest := sha256.Sum256(body)
	dedupe := webhook.BlockTransactionID
	if dedupe == "" {
		dedupe = webhook.TradeID + ":" + strconv.Itoa(webhook.Status)
	}
	event := billing.ProviderEvent{Provider: "bepusdt", ProfileID: profileID, OrderID: webhook.OrderID, TradeID: webhook.TradeID,
		ChargeID: webhook.BlockTransactionID, PayableAmount: webhook.ActualAmount, PayableCurrency: "USDT",
		FiatAmount: webhook.Amount, FiatCurrency: "USD", Recipient: webhook.Token, DedupeKey: dedupe, PayloadHash: hex.EncodeToString(digest[:])}
	return event, webhook.Status, nil
}

func (a paymentAdapter) VerifyBEPusdtUnsigned(_ context.Context, body []byte) (billing.ProviderEvent, int, error) {
	webhook, err := bepusdt.ParseUnsignedWebhook(body)
	if err != nil {
		return billing.ProviderEvent{}, 0, err
	}
	digest := sha256.Sum256(body)
	dedupe := webhook.BlockTransactionID
	if dedupe == "" {
		dedupe = webhook.OrderID + ":" + strconv.Itoa(webhook.Status) + ":" + hex.EncodeToString(digest[:8])
	}
	event := billing.ProviderEvent{Provider: "bepusdt", OrderID: webhook.OrderID, TradeID: webhook.TradeID,
		ChargeID: webhook.BlockTransactionID, PayableAmount: webhook.ActualAmount, PayableCurrency: "USDT",
		FiatAmount: webhook.Amount, FiatCurrency: "USD", Recipient: webhook.Token, DedupeKey: dedupe, PayloadHash: hex.EncodeToString(digest[:])}
	return event, webhook.Status, nil
}
