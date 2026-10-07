package abuse

import (
	"context"
	"errors"
	"strings"
)

// ErrNodeMissing rejects credentials for a node absent from a complete live list.
var ErrNodeMissing = errors.New("abuse reporting node no longer exists")

type nodeCredentialReconciler interface {
	ReconcileNodeCredentials(context.Context, []Node) error
}

func (s *Service) liveNodes(ctx context.Context) ([]Node, error) {
	if s.nodes == nil {
		return nil, ErrInvalid
	}
	nodes, err := s.nodes.AbuseNodes(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(nodes))
	for index := range nodes {
		nodes[index].UUID = strings.TrimSpace(nodes[index].UUID)
		if nodes[index].UUID == "" || strings.TrimSpace(nodes[index].Name) == "" || seen[nodes[index].UUID] {
			return nil, ErrInvalid
		}
		seen[nodes[index].UUID] = true
	}
	return nodes, nil
}

func (s *Service) requireLiveNode(ctx context.Context, uuid string) error {
	nodes, err := s.liveNodes(ctx)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if node.UUID == uuid {
			return nil
		}
	}
	return ErrNodeMissing
}
