package app

import (
	"context"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/integrations/remnawave"
	productstats "github.com/txyyddss/Remna-User-Panel/internal/statistics"
)

func (a remnaAdapter) Digest(ctx context.Context, start, end time.Time) (productstats.Digest, error) {
	digest, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) (remnawave.StatsDigest, error) {
		return client.GetStatsDigest(callCtx, start, end)
	})
	return productstats.Digest{CreatedUsers: digest.CreatedUsers, ExpiredUsers: digest.ExpiredUsers}, err
}

func (a remnaAdapter) Traffic(ctx context.Context, start, end time.Time) (productstats.Traffic, error) {
	usage, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) (remnawave.NodesUsage, error) {
		return client.GetNodesUsage(callCtx, start, end)
	})
	if err != nil {
		return productstats.Traffic{}, err
	}
	result := productstats.Traffic{Categories: append([]string(nil), usage.Categories...), Series: make([]productstats.TrafficSeries, 0, len(usage.Series))}
	for _, series := range usage.Series {
		result.Series = append(result.Series, productstats.TrafficSeries{UUID: series.UUID, Name: series.Name,
			CountryCode: series.CountryCode, DailyBytes: append([]int64(nil), series.Data...)})
	}
	return result, nil
}

func (a remnaAdapter) Nodes(ctx context.Context) ([]productstats.Node, error) {
	nodes, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) ([]remnawave.Node, error) {
		return client.ListNodes(callCtx)
	})
	if err != nil {
		return nil, err
	}
	metrics, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) ([]remnawave.NodeMetric, error) {
		return client.GetNodesMetrics(callCtx)
	})
	if err != nil {
		return nil, err
	}
	online := make(map[string]int64, len(metrics))
	for _, metric := range metrics {
		online[metric.UUID] = metric.UsersOnline
	}
	result := make([]productstats.Node, 0, len(nodes))
	for _, node := range nodes {
		item := productstats.Node{UUID: node.UUID, Name: node.Name, CountryCode: node.CountryCode,
			Online: node.IsConnected, UsersOnline: online[node.UUID], Multiplier: node.ConsumptionMultiplier}
		if node.System != nil && node.System.Stats.Interface != nil {
			item.RXBytesPerSecond = node.System.Stats.Interface.RXBytesPerSecond
			item.TXBytesPerSecond = node.System.Stats.Interface.TXBytesPerSecond
		}
		if node.Versions != nil {
			item.XrayVersion = node.Versions.Xray
		}
		result = append(result, item)
	}
	return result, nil
}

func (a remnaAdapter) Hosts(ctx context.Context) ([]productstats.Host, error) {
	hosts, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) ([]remnawave.Host, error) {
		return client.ListHosts(callCtx)
	})
	result := make([]productstats.Host, 0, len(hosts))
	for _, host := range hosts {
		result = append(result, productstats.Host{UUID: host.UUID, Remark: host.Remark, Nodes: append([]string(nil), host.Nodes...)})
	}
	return result, err
}

func (a remnaAdapter) UpdateHostRemark(ctx context.Context, hostUUID, remark string) error {
	return remnaExecute(ctx, a, func(callCtx context.Context, client remnaClient) error {
		return client.UpdateHostRemark(callCtx, hostUUID, remark)
	})
}

var _ productstats.Provider = remnaAdapter{}
