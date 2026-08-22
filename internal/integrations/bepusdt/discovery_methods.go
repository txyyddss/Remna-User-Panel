package bepusdt

import (
	"fmt"
	"strings"
)

var networkRailSuffixes = map[string]string{
	"tron": "trc20", "ethereum": "erc20", "polygon": "polygon", "bsc": "bep20",
	"aptos": "aptos", "solana": "solana", "x-layer": "xlayer", "xlayer": "xlayer",
	"arbitrum-one": "arbitrum", "arbitrum": "arbitrum", "base": "base",
	"plasma": "plasma", "ton": "ton",
}

// TradeType converts a discovered currency/network pair to BEPusdt's direct rail.
func (m AvailableMethod) TradeType() (string, error) {
	currency := strings.ToLower(strings.TrimSpace(m.Currency))
	if currency != "usdt" && currency != "usdc" {
		return "", fmt.Errorf("unsupported BEPusdt discovery currency %q", m.Currency)
	}
	network := strings.ToLower(strings.TrimSpace(m.Network))
	suffix := networkRailSuffixes[network]
	if suffix == "" {
		return "", fmt.Errorf("unsupported BEPusdt discovery network %q", m.Network)
	}
	return currency + "." + suffix, nil
}
