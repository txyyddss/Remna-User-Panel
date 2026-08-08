package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/txyyddss/Remna-User-Panel/internal/admin"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	domain "github.com/txyyddss/Remna-User-Panel/internal/emby"
	provider "github.com/txyyddss/Remna-User-Panel/internal/integrations/emby"
)

// embyAdapter resolves encrypted settings for every provider call so an
// administrator can rotate the endpoint or token without restarting.
type embyAdapter struct{ settings *admin.SettingsService }

func (a embyAdapter) client(ctx context.Context) (*provider.Client, error) {
	baseURL, err := a.settings.Plaintext(ctx, "emby.base_url")
	if err != nil {
		return nil, err
	}
	token, err := a.settings.Plaintext(ctx, "emby.api_token")
	if err != nil {
		return nil, err
	}
	return provider.NewClient(baseURL, token)
}

func (a embyAdapter) FindUserByName(ctx context.Context, name string) (domain.RemoteUser, bool, error) {
	client, err := a.client(ctx)
	if err != nil {
		return domain.RemoteUser{}, false, err
	}
	return client.FindUserByName(ctx, name)
}

func (a embyAdapter) CreateUser(ctx context.Context, name string) (domain.RemoteUser, error) {
	client, err := a.client(ctx)
	if err != nil {
		return domain.RemoteUser{}, err
	}
	return client.CreateUser(ctx, name)
}

func (a embyAdapter) GetUser(ctx context.Context, id string) (domain.RemoteUser, error) {
	client, err := a.client(ctx)
	if err != nil {
		return domain.RemoteUser{}, err
	}
	return client.GetUser(ctx, id)
}

func (a embyAdapter) SetPassword(ctx context.Context, id string, currentPassword, nextPassword []byte) error {
	client, err := a.client(ctx)
	if err != nil {
		return err
	}
	return client.SetPassword(ctx, id, currentPassword, nextPassword)
}

func (a embyAdapter) UpdatePolicy(ctx context.Context, id string, policy domain.Policy) error {
	client, err := a.client(ctx)
	if err != nil {
		return err
	}
	return client.UpdatePolicy(ctx, id, policy)
}

func (a embyAdapter) ListSelectableFolders(ctx context.Context) ([]domain.Folder, error) {
	client, err := a.client(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListSelectableFolders(ctx)
}

func (a embyAdapter) ListParentalRatings(ctx context.Context) ([]domain.ParentalRating, error) {
	client, err := a.client(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListParentalRatings(ctx)
}

func (a embyAdapter) IsNotFound(err error) bool {
	var apiError *provider.APIError
	return errors.As(err, &apiError) && apiError.HTTPStatus == http.StatusNotFound
}

func (a embyAdapter) IsTerminal(err error) bool {
	var apiError *provider.APIError
	return errors.As(err, &apiError) && apiError.HTTPStatus >= 400 && apiError.HTTPStatus < 500 &&
		apiError.HTTPStatus != http.StatusRequestTimeout && apiError.HTTPStatus != http.StatusTooManyRequests
}

func embySetupPrice(settings *admin.SettingsService) domain.PriceSource {
	return domain.PriceFunc(func(ctx context.Context) (int64, error) {
		raw, err := settings.Plaintext(ctx, "emby.setup_price_txb")
		if err != nil {
			return 0, err
		}
		minor, err := billing.ParseTXBMajor(raw)
		if err != nil || minor < 0 {
			return 0, fmt.Errorf("invalid Emby setup price")
		}
		return minor, nil
	})
}

var _ domain.Remote = embyAdapter{}
