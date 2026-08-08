package admin

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

func TestAdminSettingsAreForwardedAndAudited(t *testing.T) {
	t.Parallel()

	settingsRepository := newAdminSettingsRepository()
	settings := NewSettingsService(settingsRepository, testVault(t))
	repository := &adminCatalogRepository{}
	service := newAdminServiceForTest(repository, settings, &adminSquadImporter{}, &adminBackupRunner{}, &adminRefunder{})

	listed, err := service.Settings(context.Background())
	if err != nil || len(listed) != len(settingDefinitions) {
		t.Fatalf("Settings() = (%d settings, %v)", len(listed), err)
	}
	if err := service.PutSetting(context.Background(), "admin-1", "telegram.group_chat_id", "-1001"); err != nil {
		t.Fatalf("PutSetting(): %v", err)
	}
	if len(repository.audits) != 1 || repository.audits[0].action != "setting.update" || !strings.Contains(repository.audits[0].detail, `"secret":false`) {
		t.Fatalf("setting audit = %+v", repository.audits)
	}

	if err := service.PutSetting(context.Background(), "admin-1", "unknown", "value"); err == nil {
		t.Fatal("PutSetting() accepted unknown key")
	}
	repository.auditErr = errors.New("audit failure")
	if err := service.PutSetting(context.Background(), "admin-1", "telegram.channel_chat_id", "-1002"); err == nil {
		t.Fatal("PutSetting() ignored audit failure")
	}
}

func TestAdminSaveComboValidationAndAudit(t *testing.T) {
	t.Parallel()

	valid := database.ComboInput{Name: " Monthly ", PriceTXBMinor: 1000, ValidityDays: 30, TrafficLimitBytes: 1024, ResetStrategy: "MONTH", Active: true}
	invalid := []struct {
		name   string
		mutate func(*database.ComboInput)
	}{
		{name: "empty name", mutate: func(input *database.ComboInput) { input.Name = " " }},
		{name: "negative price", mutate: func(input *database.ComboInput) { input.PriceTXBMinor = -1 }},
		{name: "zero validity", mutate: func(input *database.ComboInput) { input.ValidityDays = 0 }},
		{name: "excess validity", mutate: func(input *database.ComboInput) { input.ValidityDays = 3651 }},
		{name: "zero traffic", mutate: func(input *database.ComboInput) { input.TrafficLimitBytes = 0 }},
		{name: "reset strategy", mutate: func(input *database.ComboInput) { input.ResetStrategy = "YEAR" }},
	}
	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			input := valid
			test.mutate(&input)
			if _, err := newAdminServiceForTest(&adminCatalogRepository{}, nil, &adminSquadImporter{}, &adminBackupRunner{}, &adminRefunder{}).SaveCombo(context.Background(), "admin", input); err == nil {
				t.Fatal("SaveCombo() unexpectedly succeeded")
			}
		})
	}

	repository := &adminCatalogRepository{savedCombo: model.Combo{ID: "combo-1", Name: "Monthly"}}
	service := newAdminServiceForTest(repository, nil, &adminSquadImporter{}, &adminBackupRunner{}, &adminRefunder{})
	combo, err := service.SaveCombo(context.Background(), "admin", valid)
	if err != nil || combo.ID != "combo-1" || repository.comboInput.Name != "Monthly" || repository.audits[0].action != "combo.save" {
		t.Fatalf("SaveCombo() = (%+v, %v), input %+v, audits %+v", combo, err, repository.comboInput, repository.audits)
	}

	repository.saveComboErr = errors.New("save failure")
	if _, err := service.SaveCombo(context.Background(), "admin", valid); err == nil {
		t.Fatal("SaveCombo() ignored repository failure")
	}
	repository.saveComboErr = nil
	repository.auditErr = errors.New("audit failure")
	if _, err := service.SaveCombo(context.Background(), "admin", valid); err == nil {
		t.Fatal("SaveCombo() ignored audit failure")
	}
}

func TestAdminDeleteComboAndSaveSquad(t *testing.T) {
	t.Parallel()

	repository := &adminCatalogRepository{savedSquad: model.SquadProduct{ID: "squad-1", RemnaSquadUUID: "uuid", Name: "Fast", UpstreamPresent: true}}
	service := newAdminServiceForTest(repository, nil, &adminSquadImporter{}, &adminBackupRunner{}, &adminRefunder{})
	if err := service.DeleteCombo(context.Background(), "admin", "combo-1"); err != nil || repository.deletedComboID != "combo-1" {
		t.Fatalf("DeleteCombo() = %v, ID %q", err, repository.deletedComboID)
	}
	product, err := service.SaveSquadProduct(context.Background(), "admin", database.SquadProductInput{RemnaSquadUUID: " uuid ", Name: "Fast", PriceTXBMinor: 10})
	if err != nil || product.ID != "squad-1" || repository.squadInput.RemnaSquadUUID != "uuid" || repository.squadInput.ID != "squad-1" {
		t.Fatalf("SaveSquadProduct() = (%+v, %v), input %+v", product, err, repository.squadInput)
	}

	for _, input := range []database.SquadProductInput{
		{Name: "name"},
		{RemnaSquadUUID: "uuid"},
		{RemnaSquadUUID: "uuid", Name: "name", PriceTXBMinor: -1},
	} {
		if _, err := service.SaveSquadProduct(context.Background(), "admin", input); err == nil {
			t.Errorf("SaveSquadProduct(%+v) unexpectedly succeeded", input)
		}
	}
	missing := &adminCatalogRepository{squadLookupErr: database.ErrNotFound}
	if _, err := newAdminServiceForTest(missing, nil, &adminSquadImporter{}, &adminBackupRunner{}, &adminRefunder{}).
		SaveSquadProduct(context.Background(), "admin", database.SquadProductInput{RemnaSquadUUID: "invented", Name: "name"}); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("SaveSquadProduct(invented UUID) = %v, want not found", err)
	}

	repository.deleteComboErr = errors.New("delete failure")
	if err := service.DeleteCombo(context.Background(), "admin", "combo-2"); err == nil {
		t.Fatal("DeleteCombo() ignored repository failure")
	}
	repository.deleteComboErr = nil
	repository.saveSquadErr = errors.New("save failure")
	if _, err := service.SaveSquadProduct(context.Background(), "admin", database.SquadProductInput{RemnaSquadUUID: "uuid", Name: "name"}); err == nil {
		t.Fatal("SaveSquadProduct() ignored repository failure")
	}
	repository.saveSquadErr = nil
	repository.auditErr = errors.New("audit failure")
	if err := service.DeleteCombo(context.Background(), "admin", "combo-3"); err == nil {
		t.Fatal("DeleteCombo() ignored audit failure")
	}
	if _, err := service.SaveSquadProduct(context.Background(), "admin", database.SquadProductInput{RemnaSquadUUID: "uuid", Name: "name"}); err == nil {
		t.Fatal("SaveSquadProduct() ignored audit failure")
	}
}

func TestAdminImportSquads(t *testing.T) {
	t.Parallel()

	upstream := []UpstreamSquad{{UUID: "uuid-1", Name: "One"}, {UUID: "uuid-2", Name: "Two"}}
	repository := &adminCatalogRepository{listedSquads: []model.SquadProduct{{ID: "one"}, {ID: "two"}}}
	importer := &adminSquadImporter{squads: upstream}
	service := newAdminServiceForTest(repository, nil, importer, &adminBackupRunner{}, &adminRefunder{})

	products, err := service.ImportSquads(context.Background(), "admin")
	if err != nil || len(products) != 2 || len(repository.importedSquads) != 2 {
		t.Fatalf("ImportSquads() = (%+v, %v), inputs %+v", products, err, repository.importedSquads)
	}
	for index, input := range repository.importedSquads {
		if input.UUID != upstream[index].UUID || input.Name != upstream[index].Name {
			t.Fatalf("imported squad = %+v, want %+v", input, upstream[index])
		}
	}
	if !strings.Contains(repository.audits[0].detail, `"count":2`) {
		t.Fatalf("import audit = %+v", repository.audits[0])
	}

	testError := errors.New("failure")
	tests := []struct {
		name       string
		repository *adminCatalogRepository
		importer   *adminSquadImporter
	}{
		{name: "upstream", repository: &adminCatalogRepository{}, importer: &adminSquadImporter{err: testError}},
		{name: "refresh", repository: &adminCatalogRepository{refreshErr: testError}, importer: &adminSquadImporter{squads: upstream}},
		{name: "audit", repository: &adminCatalogRepository{auditErr: testError}, importer: &adminSquadImporter{squads: upstream}},
		{name: "list", repository: &adminCatalogRepository{listSquadsErr: testError}, importer: &adminSquadImporter{squads: upstream}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := newAdminServiceForTest(test.repository, nil, test.importer, &adminBackupRunner{}, &adminRefunder{}).ImportSquads(context.Background(), "admin"); err == nil {
				t.Fatal("ImportSquads() unexpectedly succeeded")
			}
		})
	}
}

func TestAdminAdjustBalance(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		delta  int64
		reason string
	}{{name: "zero", reason: "reason"}, {name: "blank reason", delta: 100, reason: " "}} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := newAdminServiceForTest(&adminCatalogRepository{}, nil, &adminSquadImporter{}, &adminBackupRunner{}, &adminRefunder{}).AdjustBalance(context.Background(), "admin", "user", test.delta, test.reason); err == nil {
				t.Fatal("AdjustBalance() unexpectedly succeeded")
			}
		})
	}

	repository := &adminCatalogRepository{adjustedEntry: model.LedgerEntry{ID: "ledger-1"}}
	service := newAdminServiceForTest(repository, nil, &adminSquadImporter{}, &adminBackupRunner{}, &adminRefunder{})
	entry, err := service.AdjustBalance(context.Background(), "admin", "user", -250, " correction ")
	if err != nil || entry.ID != "ledger-1" || repository.adjustReference == "" || repository.adjustDelta != -250 || repository.audits[0].action != "balance.adjust" {
		t.Fatalf("AdjustBalance() = (%+v, %v), repo %+v", entry, err, repository)
	}
	repository.adjustErr = errors.New("adjust failure")
	if _, err := service.AdjustBalance(context.Background(), "admin", "user", 1, "reason"); err == nil {
		t.Fatal("AdjustBalance() ignored repository failure")
	}
	repository.adjustErr = nil
	repository.auditErr = errors.New("audit failure")
	if _, err := service.AdjustBalance(context.Background(), "admin", "user", 1, "reason"); err == nil {
		t.Fatal("AdjustBalance() ignored audit failure")
	}
}

func TestAdminRefund(t *testing.T) {
	t.Parallel()

	if _, err := newAdminServiceForTest(&adminCatalogRepository{}, nil, &adminSquadImporter{}, &adminBackupRunner{}, &adminRefunder{}).Refund(context.Background(), "admin", "order", " "); err == nil {
		t.Fatal("Refund() accepted blank reason")
	}

	repository := &adminCatalogRepository{paymentOrder: model.PaymentOrder{ID: "order-1", Status: "paid"}, refundedOrder: model.PaymentOrder{ID: "order-1", Status: "refunded"}}
	refunder := &adminRefunder{}
	service := newAdminServiceForTest(repository, nil, &adminSquadImporter{}, &adminBackupRunner{}, refunder)
	order, err := service.Refund(context.Background(), "admin", "order-1", "provider reversal")
	if err != nil || order.Status != "refunded" || refunder.calls != 1 || repository.refundActor == nil || *repository.refundActor != "admin" || repository.audits[0].action != "payment.refund" {
		t.Fatalf("Refund() = (%+v, %v), provider calls %d, actor %v", order, err, refunder.calls, repository.refundActor)
	}

	repository.paymentOrder.Status = "refunded"
	if _, err := service.Refund(context.Background(), "admin", "order-1", "replay"); err != nil || refunder.calls != 1 {
		t.Fatalf("Refund(replay) = %v, provider calls %d", err, refunder.calls)
	}

	testError := errors.New("failure")
	tests := []struct {
		name       string
		repository *adminCatalogRepository
		refunder   *adminRefunder
	}{
		{name: "lookup", repository: &adminCatalogRepository{paymentOrderErr: testError}, refunder: &adminRefunder{}},
		{name: "provider", repository: &adminCatalogRepository{paymentOrder: model.PaymentOrder{Status: "paid"}}, refunder: &adminRefunder{err: testError}},
		{name: "database", repository: &adminCatalogRepository{paymentOrder: model.PaymentOrder{Status: "paid"}, refundErr: testError}, refunder: &adminRefunder{}},
		{name: "audit", repository: &adminCatalogRepository{paymentOrder: model.PaymentOrder{Status: "paid"}, auditErr: testError}, refunder: &adminRefunder{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := newAdminServiceForTest(test.repository, nil, &adminSquadImporter{}, &adminBackupRunner{}, test.refunder).Refund(context.Background(), "admin", "order", "reason"); err == nil {
				t.Fatal("Refund() unexpectedly succeeded")
			}
		})
	}
}

func TestAdminCancelBackupAndRetry(t *testing.T) {
	t.Parallel()

	repository := &adminCatalogRepository{cancelledPurchase: model.Purchase{ID: "purchase-1", Status: "cancelled"}}
	backup := &adminBackupRunner{run: model.BackupRun{ID: "backup-1", Status: "complete"}}
	service := newAdminServiceForTest(repository, nil, &adminSquadImporter{}, backup, &adminRefunder{})

	if _, err := service.CancelEntitlement(context.Background(), "admin", "purchase", " "); err == nil {
		t.Fatal("CancelEntitlement() accepted blank reason")
	}
	purchase, err := service.CancelEntitlement(context.Background(), "admin", "purchase-1", "requested")
	if err != nil || purchase.Status != "cancelled" || repository.cancelledID != "purchase-1" {
		t.Fatalf("CancelEntitlement() = (%+v, %v), ID %q", purchase, err, repository.cancelledID)
	}
	run, err := service.RunBackup(context.Background(), "admin")
	if err != nil || run.ID != "backup-1" {
		t.Fatalf("RunBackup() = (%+v, %v)", run, err)
	}
	if err := service.RetryJob(context.Background(), "admin", "job-1"); err != nil || repository.retriedJobID != "job-1" {
		t.Fatalf("RetryJob() = %v, ID %q", err, repository.retriedJobID)
	}

	testError := errors.New("failure")
	repository.cancelErr = testError
	if _, err := service.CancelEntitlement(context.Background(), "admin", "purchase", "reason"); err == nil {
		t.Fatal("CancelEntitlement() ignored repository failure")
	}
	repository.cancelErr = nil
	backup.err = testError
	if got, err := service.RunBackup(context.Background(), "admin"); err == nil || got.ID != "backup-1" {
		t.Fatalf("RunBackup(error) = (%+v, %v)", got, err)
	}
	backup.err = nil
	repository.retryErr = testError
	if err := service.RetryJob(context.Background(), "admin", "job"); err == nil {
		t.Fatal("RetryJob() ignored repository failure")
	}
	repository.retryErr = nil
	repository.auditErr = testError
	if _, err := service.CancelEntitlement(context.Background(), "admin", "purchase", "reason"); err == nil {
		t.Fatal("CancelEntitlement() ignored audit failure")
	}
	if _, err := service.RunBackup(context.Background(), "admin"); err == nil {
		t.Fatal("RunBackup() ignored audit failure")
	}
	if err := service.RetryJob(context.Background(), "admin", "job"); err == nil {
		t.Fatal("RetryJob() ignored audit failure")
	}
}

type adminAuditCall struct {
	action     string
	targetType string
	targetID   string
	detail     string
}

type adminCatalogRepository struct {
	savedCombo        model.Combo
	comboInput        database.ComboInput
	saveComboErr      error
	deletedComboID    string
	deleteComboErr    error
	savedSquad        model.SquadProduct
	squadLookupErr    error
	squadInput        database.SquadProductInput
	savedSquadInputs  []database.SquadProductInput
	saveSquadErr      error
	markedMissing     bool
	markMissingErr    error
	importedSquads    []database.ImportedSquad
	refreshErr        error
	listedSquads      []model.SquadProduct
	listSquadsErr     error
	adjustedEntry     model.LedgerEntry
	adjustDelta       int64
	adjustReference   string
	adjustErr         error
	paymentOrder      model.PaymentOrder
	paymentOrderErr   error
	refundedOrder     model.PaymentOrder
	refundActor       *string
	refundErr         error
	cancelledPurchase model.Purchase
	cancelledID       string
	cancelErr         error
	retriedJobID      string
	retryErr          error
	audits            []adminAuditCall
	auditErr          error
}

func (r *adminCatalogRepository) SaveCombo(_ context.Context, input database.ComboInput) (model.Combo, error) {
	r.comboInput = input
	return r.savedCombo, r.saveComboErr
}
func (r *adminCatalogRepository) DeleteCombo(_ context.Context, id string) error {
	r.deletedComboID = id
	return r.deleteComboErr
}
func (r *adminCatalogRepository) ListCombos(context.Context, bool) ([]model.Combo, error) {
	return nil, nil
}
func (r *adminCatalogRepository) SaveSquadProduct(_ context.Context, input database.SquadProductInput) (model.SquadProduct, error) {
	r.squadInput = input
	r.savedSquadInputs = append(r.savedSquadInputs, input)
	if r.saveSquadErr != nil {
		return model.SquadProduct{}, r.saveSquadErr
	}
	if r.savedSquad.ID != "" {
		return r.savedSquad, nil
	}
	return model.SquadProduct{ID: input.RemnaSquadUUID, Name: input.Name}, nil
}
func (r *adminCatalogRepository) SquadProductByID(_ context.Context, id string) (model.SquadProduct, error) {
	if r.squadLookupErr != nil {
		return model.SquadProduct{}, r.squadLookupErr
	}
	if r.savedSquad.ID != "" {
		return r.savedSquad, nil
	}
	return model.SquadProduct{ID: id, RemnaSquadUUID: "uuid", UpstreamPresent: true}, nil
}
func (r *adminCatalogRepository) SquadProductByRemnaUUID(_ context.Context, uuid string) (model.SquadProduct, error) {
	if r.squadLookupErr != nil {
		return model.SquadProduct{}, r.squadLookupErr
	}
	if r.savedSquad.ID != "" {
		return r.savedSquad, nil
	}
	return model.SquadProduct{ID: "squad-existing", RemnaSquadUUID: uuid, UpstreamPresent: true}, nil
}
func (r *adminCatalogRepository) ListSquadProducts(context.Context, bool) ([]model.SquadProduct, error) {
	return r.listedSquads, r.listSquadsErr
}
func (r *adminCatalogRepository) MarkAllSquadsMissing(context.Context) error {
	r.markedMissing = true
	return r.markMissingErr
}
func (r *adminCatalogRepository) RefreshImportedSquads(_ context.Context, squads []database.ImportedSquad) error {
	r.importedSquads = append([]database.ImportedSquad(nil), squads...)
	return r.refreshErr
}
func (r *adminCatalogRepository) ListUsers(context.Context, int) ([]model.User, error) {
	return nil, nil
}
func (r *adminCatalogRepository) AdjustBalance(_ context.Context, _ string, delta int64, reference, _ string, _ time.Time) (model.LedgerEntry, error) {
	r.adjustDelta, r.adjustReference = delta, reference
	return r.adjustedEntry, r.adjustErr
}

func (r *adminCatalogRepository) DeductBalance(_ context.Context, _ string, amount int64, reference, _ string, _ time.Time) (model.LedgerEntry, error) {
	r.adjustDelta, r.adjustReference = -amount, reference
	return r.adjustedEntry, r.adjustErr
}
func (r *adminCatalogRepository) ListPaymentOrders(context.Context, string, int) ([]model.PaymentOrder, error) {
	return nil, nil
}
func (r *adminCatalogRepository) PaymentOrderByID(context.Context, string) (model.PaymentOrder, error) {
	return r.paymentOrder, r.paymentOrderErr
}
func (r *adminCatalogRepository) RefundPayment(_ context.Context, actor *string, _ string, _ string, _ time.Time) (model.PaymentOrder, error) {
	r.refundActor = actor
	return r.refundedOrder, r.refundErr
}
func (r *adminCatalogRepository) ListRefunds(context.Context, int) ([]model.Refund, error) {
	return nil, nil
}
func (r *adminCatalogRepository) ListAllPurchases(context.Context, int) ([]model.Purchase, error) {
	return nil, nil
}
func (r *adminCatalogRepository) CancelPurchase(_ context.Context, id, _ string, _ time.Time) (model.Purchase, error) {
	r.cancelledID = id
	return r.cancelledPurchase, r.cancelErr
}
func (r *adminCatalogRepository) ListBackupRuns(context.Context, int) ([]model.BackupRun, error) {
	return nil, nil
}
func (r *adminCatalogRepository) ListOutboxJobs(context.Context, int) ([]model.OutboxJob, error) {
	return nil, nil
}
func (r *adminCatalogRepository) RetryOutboxJob(_ context.Context, id string, _ time.Time) error {
	r.retriedJobID = id
	return r.retryErr
}
func (r *adminCatalogRepository) ListAuditEvents(context.Context, int) ([]model.AuditEvent, error) {
	return nil, nil
}
func (r *adminCatalogRepository) AppendAudit(_ context.Context, _ *string, action, targetType, targetID, detail string, _ time.Time) error {
	r.audits = append(r.audits, adminAuditCall{action: action, targetType: targetType, targetID: targetID, detail: detail})
	return r.auditErr
}

type adminSquadImporter struct {
	squads []UpstreamSquad
	err    error
}

func (i *adminSquadImporter) ListInternalSquads(context.Context) ([]UpstreamSquad, error) {
	return i.squads, i.err
}

type adminBackupRunner struct {
	run model.BackupRun
	err error
}

func (b *adminBackupRunner) Run(context.Context) (model.BackupRun, error) { return b.run, b.err }

type adminRefunder struct {
	calls int
	err   error
}

func (r *adminRefunder) RefundProvider(context.Context, model.PaymentOrder) error {
	r.calls++
	return r.err
}

func newAdminServiceForTest(repository CatalogRepository, settings *SettingsService, importer SquadImporter, backups BackupRunner, refunder PaymentRefunder) *Service {
	service := NewService(repository, settings, importer, backups, refunder)
	service.now = func() time.Time { return time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC) }
	return service
}
