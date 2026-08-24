package app

import (
	"context"
	"errors"

	"github.com/txyyddss/Remna-User-Panel/internal/integrations/remnawave"
)

func (a remnaAdapter) AbuseRevoke(ctx context.Context, remoteID string) error {
	id, err := remnaUserID(remoteID)
	if err != nil {
		return err
	}
	return remnaExecute(ctx, a, func(callCtx context.Context, client remnaClient) error {
		_, callErr := client.RevokeSubscription(callCtx, id, false)
		return callErr
	})
}
func (a remnaAdapter) AbuseSetStatus(ctx context.Context, remoteID string, status remnawave.UserStatus) error {
	id, err := remnaUserID(remoteID)
	if err != nil {
		return err
	}
	return remnaExecute(ctx, a, func(callCtx context.Context, client remnaClient) error {
		_, callErr := client.UpdateUser(callCtx, remnawave.UpdateUserRequest{ID: id, Status: &status})
		return callErr
	})
}
func (a remnaAdapter) AbuseIPBan(ctx context.Context, remoteID string, nodes []string, all bool, durationSeconds int) error {
	id, err := remnaUserID(remoteID)
	if err != nil {
		return err
	}
	jobID, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) (string, error) {
		return client.RequestUserConnections(callCtx, id)
	})
	if err != nil {
		return err
	}
	scan, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) (remnawave.ConnectionScan, error) {
		return client.UserConnections(callCtx, jobID)
	})
	if err != nil {
		return err
	}
	if !scan.Completed {
		return errors.New("abuse connection scan is still pending")
	}
	wanted := map[string]bool{}
	for _, node := range nodes {
		wanted[node] = true
	}
	for _, node := range scan.Nodes {
		if !all && !wanted[node.UUID] {
			continue
		}
		for _, item := range node.IPs {
			if err = remnaExecute(ctx, a, func(callCtx context.Context, client remnaClient) error {
				return client.BlockIP(callCtx, item.IP, node.UUID, durationSeconds)
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
