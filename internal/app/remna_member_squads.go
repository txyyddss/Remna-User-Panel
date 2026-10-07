package app

import "context"

// MemberSquadNames reads the complete live list, including hidden owned squads.
func (a remnaAdapter) MemberSquadNames(ctx context.Context) (map[string]string, error) {
	return remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) (map[string]string, error) {
		items, err := client.ListInternalSquads(callCtx)
		if err != nil {
			return nil, err
		}
		names := make(map[string]string, len(items))
		for _, item := range items {
			names[item.UUID] = item.Name
		}
		return names, nil
	})
}
