package admin

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/txyyddss/Remna-User-Panel/internal/billing"
)

type adminNodeImporter struct {
	squads          []UpstreamSquad
	nodes           []UpstreamNode
	accessible      []string
	listErr         error
	accessibleErr   error
	updateErr       error
	updatedSquad    string
	updatedInbounds []string
}

func (i *adminNodeImporter) ListInternalSquads(context.Context) ([]UpstreamSquad, error) {
	return i.squads, nil
}

func (i *adminNodeImporter) ListNodes(context.Context) ([]UpstreamNode, error) {
	return i.nodes, i.listErr
}

func (i *adminNodeImporter) AccessibleNodeUUIDs(context.Context, string) ([]string, error) {
	return i.accessible, i.accessibleErr
}

func (i *adminNodeImporter) UpdateInternalSquadInbounds(_ context.Context, squad string, inbounds []string) error {
	i.updatedSquad = squad
	i.updatedInbounds = append([]string(nil), inbounds...)
	return i.updateErr
}

func TestSquadNodesNormalizesAndSortsLiveNodes(t *testing.T) {
	t.Parallel()

	importer := &adminNodeImporter{
		squads: []UpstreamSquad{{UUID: " squad-1 "}},
		nodes: []UpstreamNode{
			{UUID: "z", Name: "Zulu", CountryCode: " bad ", ActiveInboundUUIDs: []string{"in-1"}},
			{UUID: "a", Name: "Alpha", CountryCode: "us", ActiveInboundUUIDs: []string{"in-2"}},
		},
		accessible: []string{"z"},
	}
	service := newAdminServiceForTest(&adminCatalogRepository{}, nil, importer, &adminBackupRunner{}, &adminRefunder{})
	nodes, err := service.SquadNodes(context.Background(), "squad-1")
	if err != nil {
		t.Fatalf("SquadNodes(): %v", err)
	}
	if len(nodes) != 2 || nodes[0].Name != "Alpha" || nodes[0].CountryCode != "US" || nodes[0].Accessible || !nodes[1].Accessible {
		t.Fatalf("SquadNodes() = %+v", nodes)
	}
	if nodes[1].CountryCode != "" || !reflect.DeepEqual(nodes[1].ActiveInboundUUIDs, []string{"in-1"}) {
		t.Fatalf("normalized node = %+v", nodes[1])
	}
}

func TestSquadNodesAndUpdateErrors(t *testing.T) {
	t.Parallel()
	testErr := errors.New("upstream failure")
	base := func() *adminNodeImporter {
		return &adminNodeImporter{squads: []UpstreamSquad{{UUID: "squad"}}, nodes: []UpstreamNode{{UUID: "node", ActiveInboundUUIDs: []string{"b", "a"}}}, accessible: []string{"node"}}
	}

	service := newAdminServiceForTest(&adminCatalogRepository{}, nil, &adminSquadImporter{}, &adminBackupRunner{}, &adminRefunder{})
	if _, err := service.SquadNodes(context.Background(), "squad"); err == nil {
		t.Fatal("SquadNodes() accepted an importer without node management")
	}
	for name, importer := range map[string]*adminNodeImporter{
		"missing squad": {squads: []UpstreamSquad{{UUID: "other"}}},
		"node list":     {squads: []UpstreamSquad{{UUID: "squad"}}, listErr: testErr},
		"access list":   {squads: []UpstreamSquad{{UUID: "squad"}}, accessibleErr: testErr},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := newAdminServiceForTest(&adminCatalogRepository{}, nil, importer, &adminBackupRunner{}, &adminRefunder{}).SquadNodes(context.Background(), "squad"); err == nil {
				t.Fatal("SquadNodes() unexpectedly succeeded")
			}
		})
	}

	importer := base()
	repository := &adminCatalogRepository{}
	service = newAdminServiceForTest(repository, nil, importer, &adminBackupRunner{}, &adminRefunder{})
	if _, err := service.UpdateSquadNodes(context.Background(), "admin", "squad", []string{"node", "node"}); err != nil {
		t.Fatalf("UpdateSquadNodes(): %v", err)
	}
	if importer.updatedSquad != "squad" || !reflect.DeepEqual(importer.updatedInbounds, []string{"a", "b"}) {
		t.Fatalf("updated inbounds = %q/%v", importer.updatedSquad, importer.updatedInbounds)
	}
	if len(repository.audits) != 1 || repository.audits[0].action != "squad.nodes.update" {
		t.Fatalf("update audit = %+v", repository.audits)
	}

	for name, selected := range map[string][]string{"unknown node": {"missing"}, "disabled node": {"disabled"}} {
		t.Run(name, func(t *testing.T) {
			current := base()
			if name == "disabled node" {
				current.nodes = []UpstreamNode{{UUID: "disabled", Disabled: true}}
			}
			if _, err := newAdminServiceForTest(&adminCatalogRepository{}, nil, current, &adminBackupRunner{}, &adminRefunder{}).UpdateSquadNodes(context.Background(), "admin", "squad", selected); err == nil {
				t.Fatal("UpdateSquadNodes() unexpectedly succeeded")
			}
		})
	}
}

func TestSettingValidatorsCoverOptionalFormats(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		fn   func(string) error
		good string
		bad  string
	}{
		{"nonnegative TXB", validateNonnegativeTXB, "0.01", "-1"},
		{"timezone", validateTimezone, "UTC", "Not/ATimezone"},
		{"nonnegative integer", validateNonnegativeInteger, "0", "-1"},
		{"webhook secret", validateWebhookSecret, "safe_secret-1", "contains space"},
		{"BEPusdt methods", billing.ValidateBEPusdtMethods, "usdt.trc20,usdt.ton", "usdc"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.fn(test.good); err != nil {
				t.Errorf("valid value: %v", err)
			}
			if err := test.fn(test.bad); err == nil {
				t.Errorf("invalid value %q was accepted", test.bad)
			}
		})
	}
}
