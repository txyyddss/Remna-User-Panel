package admin

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/abuse"
	"github.com/txyyddss/Remna-User-Panel/internal/affiliates"
	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/emby"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

// UserWorkflowRepository is the aggregate read surface for one admin profile.
type UserWorkflowRepository interface {
	UserByID(context.Context, string) (model.User, error)
	Balance(context.Context, string) (model.Money, error)
	ListPurchases(context.Context, string) ([]model.Purchase, error)
	EmbyAccountForUser(context.Context, string) (emby.Account, error)
	ListPaymentOrders(context.Context, string, int) ([]model.PaymentOrder, error)
	ListRefundsForUser(context.Context, string, int) ([]model.Refund, error)
	ListAdminOperationsForUser(context.Context, string, int) ([]model.OperationReceipt, error)
	ListCouponGrants(context.Context, string, time.Time) ([]coupons.Grant, error)
	MemberRecords(context.Context, string, string, int) (abuse.RecordPage, error)
	AffiliateReferrals(context.Context, string, int) (affiliates.ReferralPage, error)
	ComboByID(context.Context, string, bool) (model.Combo, error)
	EditAdminEntitlement(context.Context, database.AdminEntitlementEditInput, time.Time) (model.Purchase, error)
	RefundAdminEntitlement(context.Context, database.AdminEntitlementRefundInput, time.Time) (model.OperationReceipt, error)
	ReplaceAdminCombo(context.Context, database.AdminComboReplacementInput, time.Time) (model.OperationReceipt, error)
	ResolveAdminOperation(context.Context, database.AdminOperationResolutionInput, time.Time) (model.OperationReceipt, error)
	PreviewAdminBulkExtension(context.Context, database.AdminBulkExtensionFilter, time.Time) (database.AdminBulkExtensionPreview, error)
	CreateAdminBulkExtension(context.Context, database.AdminBulkExtensionInput, time.Time) (model.OperationReceipt, error)
}

// UserSynchronization summarizes the local-to-provider identity state.
type UserSynchronization struct {
	Status       string
	RemoteUserID *string
	LastError    *string
}

// UserDetail is one non-duplicated aggregate administrator projection.
type UserDetail struct {
	User             model.User
	Balance          model.Money
	Synchronization  UserSynchronization
	ActiveCombo      *model.Purchase
	Entitlements     []model.Purchase
	EmbyAccounts     []emby.Account
	Payments         []model.PaymentOrder
	Refunds          []model.Refund
	Operations       []model.OperationReceipt
	CouponWallet     []coupons.Grant
	AbuseHistory     []abuse.Record
	AffiliateHistory affiliates.ReferralPage
}

// UserWorkflows owns aggregate reads and durable administrator commands.
type UserWorkflows struct {
	repository UserWorkflowRepository
	importer   SquadImporter
	now        func() time.Time
}

// NewUserWorkflows constructs the focused administrator workflow service.
func NewUserWorkflows(repository UserWorkflowRepository, importer SquadImporter) *UserWorkflows {
	return &UserWorkflows{repository: repository, importer: importer, now: time.Now}
}

// UserDetail loads one aggregate profile without copying facts onto users.
func (s *UserWorkflows) UserDetail(ctx context.Context, userID string) (UserDetail, error) {
	user, err := s.repository.UserByID(ctx, userID)
	if err != nil {
		return UserDetail{}, err
	}
	detail := UserDetail{User: user, Synchronization: synchronizationFor(user), EmbyAccounts: []emby.Account{}, CouponWallet: []coupons.Grant{}, AbuseHistory: []abuse.Record{}, AffiliateHistory: affiliates.ReferralPage{Items: []affiliates.Referral{}}}
	if detail.Balance, err = s.repository.Balance(ctx, userID); err != nil {
		return UserDetail{}, err
	}
	if detail.Entitlements, err = s.repository.ListPurchases(ctx, userID); err != nil {
		return UserDetail{}, err
	}
	sort.SliceStable(detail.Entitlements, func(i, j int) bool {
		return detail.Entitlements[i].ValidFrom.After(detail.Entitlements[j].ValidFrom)
	})
	now := s.now().UTC()
	for index := range detail.Entitlements {
		item := &detail.Entitlements[index]
		if (item.Status == "active" || item.Status == "activating") && !now.Before(item.ValidFrom) && now.Before(item.ValidUntil) {
			copy := *item
			detail.ActiveCombo = &copy
			break
		}
	}
	account, accountErr := s.repository.EmbyAccountForUser(ctx, userID)
	if accountErr == nil {
		detail.EmbyAccounts = append(detail.EmbyAccounts, account)
	} else if !errors.Is(accountErr, emby.ErrNotFound) && !errors.Is(accountErr, database.ErrNotFound) {
		return UserDetail{}, accountErr
	}
	if detail.Payments, err = s.repository.ListPaymentOrders(ctx, userID, 200); err != nil {
		return UserDetail{}, err
	}
	if detail.Refunds, err = s.repository.ListRefundsForUser(ctx, userID, 200); err != nil {
		return UserDetail{}, err
	}
	if detail.Operations, err = s.repository.ListAdminOperationsForUser(ctx, userID, 200); err != nil {
		return UserDetail{}, err
	}
	if detail.CouponWallet, err = s.repository.ListCouponGrants(ctx, userID, now); err != nil {
		return UserDetail{}, err
	}
	if page, recordsErr := s.repository.MemberRecords(ctx, userID, "", 100); recordsErr != nil {
		return UserDetail{}, recordsErr
	} else {
		detail.AbuseHistory = page.Items
	}
	if detail.AffiliateHistory, err = s.repository.AffiliateReferrals(ctx, userID, 1); err != nil {
		return UserDetail{}, err
	}
	return detail, nil
}

func synchronizationFor(user model.User) UserSynchronization {
	state := UserSynchronization{Status: "not_provisioned", RemoteUserID: user.RemnaUserID}
	if user.RecoveryReason != "" {
		state.Status = "failed"
		state.LastError = &user.RecoveryReason
	} else if user.RemnaUserID != nil {
		state.Status = "synchronized"
	}
	return state
}
