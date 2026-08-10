package admin

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

// CatalogRepository is the domain-safe database surface available to admins.
type CatalogRepository interface {
	SaveCombo(context.Context, database.ComboInput) (model.Combo, error)
	DeleteCombo(context.Context, string) error
	ListCombos(context.Context, bool) ([]model.Combo, error)
	SaveSquadProduct(context.Context, database.SquadProductInput) (model.SquadProduct, error)
	SquadProductByID(context.Context, string) (model.SquadProduct, error)
	SquadProductByRemnaUUID(context.Context, string) (model.SquadProduct, error)
	ListSquadProducts(context.Context, bool) ([]model.SquadProduct, error)
	RefreshImportedSquads(context.Context, []database.ImportedSquad) error
	ListUsers(context.Context, int) ([]model.User, error)
	AdjustBalance(context.Context, string, int64, string, string, time.Time) (model.LedgerEntry, error)
	DeductBalance(context.Context, string, int64, string, string, time.Time) (model.LedgerEntry, error)
	ListPaymentOrders(context.Context, string, int) ([]model.PaymentOrder, error)
	PaymentOrderByID(context.Context, string) (model.PaymentOrder, error)
	RefundPayment(context.Context, *string, string, string, time.Time) (model.PaymentOrder, error)
	ListRefunds(context.Context, int) ([]model.Refund, error)
	ListAllPurchases(context.Context, int) ([]model.Purchase, error)
	CancelPurchase(context.Context, string, string, time.Time) (model.Purchase, error)
	ListBackupRuns(context.Context, int) ([]model.BackupRun, error)
	ListOutboxJobs(context.Context, int) ([]model.OutboxJob, error)
	RetryOutboxJob(context.Context, string, time.Time) error
	ListAuditEvents(context.Context, int) ([]model.AuditEvent, error)
	AppendAudit(context.Context, *string, string, string, string, string, time.Time) error
}

// PaymentRefunder performs provider-side refund operations that have a documented API.
type PaymentRefunder interface {
	RefundProvider(context.Context, model.PaymentOrder) error
}

// UpstreamSquad is imported from Remnawave without granting direct mutation access.
type UpstreamSquad struct {
	UUID string
	Name string
}

// SquadImporter lists Remnawave internal squads.
type SquadImporter interface {
	ListInternalSquads(context.Context) ([]UpstreamSquad, error)
}

// UpstreamNode contains only the volatile fields needed to present and apply a
// squad node selection. It is never written to SQLite.
type UpstreamNode struct {
	UUID                  string
	Name                  string
	CountryCode           string
	ConsumptionMultiplier float64
	ActiveInboundUUIDs    []string
	Disabled              bool
}

type SquadNodeManager interface {
	ListNodes(context.Context) ([]UpstreamNode, error)
	AccessibleNodeUUIDs(context.Context, string) ([]string, error)
	UpdateInternalSquadInbounds(context.Context, string, []string) error
}

// BackupRunner creates and verifies one online backup.
type BackupRunner interface {
	Run(context.Context) (model.BackupRun, error)
}

type backupDeleter interface {
	Delete(context.Context, string, string) error
}

// Service exposes audited administrative operations.
type Service struct {
	repository CatalogRepository
	settings   *SettingsService
	importer   SquadImporter
	backups    BackupRunner
	refunder   PaymentRefunder
	now        func() time.Time
}

// NewService constructs an admin service.
func NewService(repository CatalogRepository, settings *SettingsService, importer SquadImporter, backups BackupRunner, refunder PaymentRefunder) *Service {
	return &Service{repository: repository, settings: settings, importer: importer, backups: backups, refunder: refunder, now: time.Now}
}

// DeleteBackup delegates safe file containment and restore-state checks to the
// backup lifecycle service.
func (s *Service) DeleteBackup(ctx context.Context, actorID, backupID string) error {
	deleter, ok := s.backups.(backupDeleter)
	if !ok {
		return errors.New("backup deletion is unavailable")
	}
	return deleter.Delete(ctx, backupID, actorID)
}

// Settings returns masked write-only credentials.
func (s *Service) Settings(ctx context.Context) ([]model.Setting, error) {
	return s.settings.SafeList(ctx)
}

// PutSetting validates and audits one config mutation.
func (s *Service) PutSetting(ctx context.Context, actorID, key, value string) error {
	if err := s.settings.Put(ctx, actorID, key, value); err != nil {
		return err
	}
	return s.audit(ctx, actorID, "setting.update", "setting", key, map[string]any{"secret": settingDefinitions[key].Secret})
}

// SaveCombo validates a complete catalog item.
func (s *Service) SaveCombo(ctx context.Context, actorID string, input database.ComboInput) (model.Combo, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || input.PriceTXBMinor < 0 || input.PriceTXBMinor > 1_000_000_000_000 || input.ValidityDays < 1 || input.ValidityDays > 3650 || input.TrafficLimitBytes <= 0 ||
		input.RolloverMinRemainingBPS < 0 || input.RolloverMinRemainingBPS > 10_000 || input.RolloverMaxTXBMinor < 0 || input.RolloverMaxTXBMinor > 1_000_000_000_000 {
		return model.Combo{}, errors.New("invalid combo fields")
	}
	if input.ResetStrategy != "DAY" && input.ResetStrategy != "WEEK" && input.ResetStrategy != "MONTH" {
		return model.Combo{}, errors.New("invalid reset strategy")
	}
	live, err := s.liveSquads(ctx)
	if err != nil {
		return model.Combo{}, err
	}
	for _, squadUUID := range input.SquadProductIDs {
		if _, exists := live[strings.TrimSpace(squadUUID)]; !exists {
			return model.Combo{}, database.ErrNotFound
		}
	}
	combo, err := s.repository.SaveCombo(ctx, input)
	if err != nil {
		return model.Combo{}, err
	}
	if err := s.audit(ctx, actorID, "combo.save", "combo", combo.ID, map[string]any{"name": combo.Name}); err != nil {
		return model.Combo{}, err
	}
	return combo, nil
}

// DeleteCombo hides the selected combo.
func (s *Service) DeleteCombo(ctx context.Context, actorID, comboID string) error {
	if err := s.repository.DeleteCombo(ctx, comboID); err != nil {
		return err
	}
	return s.audit(ctx, actorID, "combo.hide", "combo", comboID, nil)
}

// SaveSquadProduct validates local merchandising data.
func (s *Service) SaveSquadProduct(ctx context.Context, actorID string, input database.SquadProductInput) (model.SquadProduct, error) {
	input.RemnaSquadUUID = strings.TrimSpace(input.RemnaSquadUUID)
	input.Name = strings.TrimSpace(input.Name)
	if input.RemnaSquadUUID == "" || input.PriceTXBMinor < 0 || input.PriceTXBMinor > 1_000_000_000_000 {
		return model.SquadProduct{}, errors.New("invalid squad product")
	}
	live, err := s.liveSquads(ctx)
	if err != nil {
		return model.SquadProduct{}, err
	}
	upstreamName, exists := live[input.RemnaSquadUUID]
	if !exists {
		return model.SquadProduct{}, database.ErrNotFound
	}
	input.ID = input.RemnaSquadUUID
	input.Name = upstreamName
	input.UpstreamPresent = true
	product, err := s.repository.SaveSquadProduct(ctx, input)
	if err != nil {
		return model.SquadProduct{}, err
	}
	product.Name = upstreamName
	product.UpstreamPresent = true
	if err := s.audit(ctx, actorID, "squad_product.save", "squad_product", product.ID, map[string]any{"name": upstreamName}); err != nil {
		return model.SquadProduct{}, err
	}
	return product, nil
}

// ImportSquads now returns a live upstream list overlaid with the sparse local
// overrides. The retained import call is an audited compatibility endpoint and
// no longer stores a second catalog identity table.
func (s *Service) ImportSquads(ctx context.Context, actorID string) ([]model.SquadProduct, error) {
	upstream, err := s.importer.ListInternalSquads(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.audit(ctx, actorID, "squads.import", "catalog", "squads", map[string]any{"count": len(upstream)}); err != nil {
		return nil, err
	}
	return s.Squads(ctx)
}

// Squads builds the admin catalog from live Remnawave identities plus sparse
// local merchandising overrides.
func (s *Service) Squads(ctx context.Context) ([]model.SquadProduct, error) {
	upstream, err := s.importer.ListInternalSquads(ctx)
	if err != nil {
		return nil, err
	}
	overrides, err := s.repository.ListSquadProducts(ctx, false)
	if err != nil {
		return nil, err
	}
	overrideByUUID := make(map[string]model.SquadProduct, len(overrides))
	for _, override := range overrides {
		overrideByUUID[override.RemnaSquadUUID] = override
	}
	products := make([]model.SquadProduct, 0, len(upstream))
	for _, squad := range upstream {
		product, exists := overrideByUUID[squad.UUID]
		if !exists {
			product = model.SquadProduct{ID: squad.UUID, RemnaSquadUUID: squad.UUID, Visible: false, Price: model.TXBMoney(0)}
		}
		product.ID = squad.UUID
		product.RemnaSquadUUID = squad.UUID
		product.Name = squad.Name
		product.UpstreamPresent = true
		products = append(products, product)
	}
	return products, nil
}

func (s *Service) liveSquads(ctx context.Context) (map[string]string, error) {
	upstream, err := s.importer.ListInternalSquads(ctx)
	if err != nil {
		return nil, err
	}
	live := make(map[string]string, len(upstream))
	for _, squad := range upstream {
		live[strings.TrimSpace(squad.UUID)] = strings.TrimSpace(squad.Name)
	}
	return live, nil
}

// SquadNodes returns the selectable nodes together with Remnawave's actual
// current accessibility for the squad.
func (s *Service) SquadNodes(ctx context.Context, squadUUID string) ([]model.RemnaNode, error) {
	manager, ok := s.importer.(SquadNodeManager)
	if !ok {
		return nil, errors.New("Remnawave node management is unavailable")
	}
	live, err := s.liveSquads(ctx)
	if err != nil {
		return nil, err
	}
	if _, exists := live[strings.TrimSpace(squadUUID)]; !exists {
		return nil, database.ErrNotFound
	}
	nodes, err := manager.ListNodes(ctx)
	if err != nil {
		return nil, err
	}
	accessibleUUIDs, err := manager.AccessibleNodeUUIDs(ctx, squadUUID)
	if err != nil {
		return nil, err
	}
	accessible := make(map[string]struct{}, len(accessibleUUIDs))
	for _, nodeUUID := range accessibleUUIDs {
		accessible[nodeUUID] = struct{}{}
	}
	result := make([]model.RemnaNode, 0, len(nodes))
	for _, node := range nodes {
		_, isAccessible := accessible[node.UUID]
		result = append(result, model.RemnaNode{UUID: node.UUID, Name: node.Name, CountryCode: normalizedCountryCode(node.CountryCode),
			ConsumptionMultiplier: node.ConsumptionMultiplier, ActiveInboundUUIDs: append([]string(nil), node.ActiveInboundUUIDs...), Accessible: isAccessible})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name == result[j].Name {
			return result[i].UUID < result[j].UUID
		}
		return result[i].Name < result[j].Name
	})
	return result, nil
}

// UpdateSquadNodes validates node UUIDs, unions their current active inbounds,
// patches Remnawave, then re-fetches accessibility instead of assuming the
// selected node set equals the resulting state.
func (s *Service) UpdateSquadNodes(ctx context.Context, actorID, squadUUID string, selectedNodeUUIDs []string) ([]model.RemnaNode, error) {
	manager, ok := s.importer.(SquadNodeManager)
	if !ok {
		return nil, errors.New("Remnawave node management is unavailable")
	}
	live, err := s.liveSquads(ctx)
	if err != nil {
		return nil, err
	}
	if _, exists := live[strings.TrimSpace(squadUUID)]; !exists {
		return nil, database.ErrNotFound
	}
	nodes, err := manager.ListNodes(ctx)
	if err != nil {
		return nil, err
	}
	byUUID := make(map[string]UpstreamNode, len(nodes))
	for _, node := range nodes {
		byUUID[node.UUID] = node
	}
	inboundSet := make(map[string]struct{})
	selectedSet := make(map[string]struct{}, len(selectedNodeUUIDs))
	for _, rawUUID := range selectedNodeUUIDs {
		nodeUUID := strings.TrimSpace(rawUUID)
		if _, duplicate := selectedSet[nodeUUID]; duplicate {
			continue
		}
		node, exists := byUUID[nodeUUID]
		if !exists || node.Disabled {
			return nil, database.ErrNotFound
		}
		selectedSet[nodeUUID] = struct{}{}
		for _, inboundUUID := range node.ActiveInboundUUIDs {
			inboundSet[inboundUUID] = struct{}{}
		}
	}
	inbounds := make([]string, 0, len(inboundSet))
	for inboundUUID := range inboundSet {
		inbounds = append(inbounds, inboundUUID)
	}
	sort.Strings(inbounds)
	if err := manager.UpdateInternalSquadInbounds(ctx, squadUUID, inbounds); err != nil {
		return nil, err
	}
	if err := s.audit(ctx, actorID, "squad.nodes.update", "squad_product", squadUUID, map[string]any{"selectedNodeUuids": selectedNodeUUIDs}); err != nil {
		return nil, err
	}
	return s.SquadNodes(ctx, squadUUID)
}

func normalizedCountryCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) != 2 || value[0] < 'A' || value[0] > 'Z' || value[1] < 'A' || value[1] > 'Z' {
		return ""
	}
	return value
}

// AdjustBalance appends an audited immutable ledger entry.
func (s *Service) AdjustBalance(ctx context.Context, actorID, userID string, delta int64, reason string) (model.LedgerEntry, error) {
	if delta == 0 || delta < -1_000_000_000_000 || delta > 1_000_000_000_000 || strings.TrimSpace(reason) == "" {
		return model.LedgerEntry{}, errors.New("non-zero delta and reason are required")
	}
	referenceID := fmt.Sprintf("telegram-deduct:%x", sha256.Sum256([]byte(reason)))
	entry, err := s.repository.AdjustBalance(ctx, userID, delta, referenceID, reason, s.now().UTC())
	if err != nil {
		return model.LedgerEntry{}, err
	}
	if err := s.audit(ctx, actorID, "balance.adjust", "user", userID, map[string]any{"deltaMinor": delta, "reason": reason, "ledgerEntryId": entry.ID}); err != nil {
		return model.LedgerEntry{}, err
	}
	return entry, nil
}

// DeductBalance appends an audited exact debit that cannot create debt.
func (s *Service) DeductBalance(ctx context.Context, actorID, userID string, amount int64, reason string) (model.LedgerEntry, error) {
	if amount <= 0 || amount > 1_000_000_000_000 || strings.TrimSpace(reason) == "" {
		return model.LedgerEntry{}, errors.New("positive amount and reason are required")
	}
	referenceID, err := ids.New()
	if err != nil {
		return model.LedgerEntry{}, err
	}
	entry, err := s.repository.DeductBalance(ctx, userID, amount, referenceID, reason, s.now().UTC())
	if err != nil {
		return model.LedgerEntry{}, err
	}
	if err := s.audit(ctx, actorID, "telegram.balance_deduct", "user", userID, map[string]any{"amountMinor": amount, "reason": reason, "ledgerEntryId": entry.ID}); err != nil {
		return model.LedgerEntry{}, err
	}
	return entry, nil
}

// Refund reverses one settled payment and applies the debt policy transactionally.
func (s *Service) Refund(ctx context.Context, actorID, orderID, reason string) (model.PaymentOrder, error) {
	if strings.TrimSpace(reason) == "" {
		return model.PaymentOrder{}, errors.New("refund reason is required")
	}
	order, err := s.repository.PaymentOrderByID(ctx, orderID)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	// Persist the administrator's intent before any irreversible provider call.
	// Telegram's refunded-payment webhook and reconciliation loop complete a
	// local reversal if the process stops after a successful Stars refund.
	if err := s.audit(ctx, actorID, "payment.refund", "payment", orderID, map[string]any{"reason": reason, "phase": "requested"}); err != nil {
		return model.PaymentOrder{}, err
	}
	if order.Status != "refunded" && s.refunder != nil {
		if err := s.refunder.RefundProvider(ctx, order); err != nil {
			return model.PaymentOrder{}, fmt.Errorf("refund provider payment: %w", err)
		}
	}
	order, err = s.repository.RefundPayment(ctx, &actorID, orderID, reason, s.now().UTC())
	if err != nil {
		return model.PaymentOrder{}, err
	}
	return order, nil
}

// CancelEntitlement applies a compensating credit and revokes active upstream access.
func (s *Service) CancelEntitlement(ctx context.Context, actorID, purchaseID, reason string) (model.Purchase, error) {
	if strings.TrimSpace(reason) == "" {
		return model.Purchase{}, errors.New("cancellation reason is required")
	}
	purchase, err := s.repository.CancelPurchase(ctx, purchaseID, reason, s.now().UTC())
	if err != nil {
		return model.Purchase{}, err
	}
	if err := s.audit(ctx, actorID, "entitlement.cancel", "purchase", purchaseID, map[string]any{"reason": reason}); err != nil {
		return model.Purchase{}, err
	}
	return purchase, nil
}

// RunBackup executes a verified online backup and audits the request.
func (s *Service) RunBackup(ctx context.Context, actorID string) (model.BackupRun, error) {
	run, err := s.backups.Run(ctx)
	if err != nil {
		return run, err
	}
	if err := s.audit(ctx, actorID, "backup.create", "backup", run.ID, nil); err != nil {
		return model.BackupRun{}, err
	}
	return run, nil
}

// RetryJob makes a failed synchronization eligible again.
func (s *Service) RetryJob(ctx context.Context, actorID, jobID string) error {
	if err := s.repository.RetryOutboxJob(ctx, jobID, s.now().UTC()); err != nil {
		return err
	}
	return s.audit(ctx, actorID, "job.retry", "job", jobID, nil)
}

func (s *Service) audit(ctx context.Context, actorID, action, targetType, targetID string, detail any) error {
	payload := "{}"
	if detail != nil {
		encoded, err := json.Marshal(detail)
		if err != nil {
			return fmt.Errorf("encode audit detail: %w", err)
		}
		payload = string(encoded)
	}
	return s.repository.AppendAudit(ctx, &actorID, action, targetType, targetID, payload, s.now().UTC())
}
