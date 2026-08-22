package billing

import (
	"fmt"
	"strings"
)

var ezpayRails = map[string]string{
	"alipay": "Alipay",
	"wxpay":  "WeChat Pay",
	// "wechat" is retained as a compatibility alias for persisted profiles.
	"wechat": "WeChat Pay",
	"qqpay":  "QQ Wallet",
	"bank":   "UnionPay",
	"jdpay":  "JD Pay",
}

var bepusdtRails = map[string]string{
	"usdt.trc20": "USDT TRC20", "usdt.erc20": "USDT ERC20",
	"usdt.polygon": "USDT Polygon", "usdt.bep20": "USDT BEP20",
	"usdt.aptos": "USDT Aptos", "usdt.solana": "USDT Solana",
	"usdt.xlayer": "USDT X-Layer", "usdt.arbitrum": "USDT Arbitrum One",
	"usdt.plasma": "USDT Plasma", "usdt.ton": "USDT TON",
	"usdc.trc20": "USDC TRC20", "usdc.erc20": "USDC ERC20",
	"usdc.polygon": "USDC Polygon", "usdc.bep20": "USDC BEP20",
	"usdc.aptos": "USDC Aptos", "usdc.solana": "USDC Solana",
	"usdc.xlayer": "USDC X-Layer", "usdc.arbitrum": "USDC Arbitrum One",
	"usdc.base": "USDC Base",
}

var paymentChannelOrder = map[string][]string{
	"ezpay": {"alipay", "wxpay", "qqpay", "bank", "jdpay"},
	"bepusdt": {
		"usdt.trc20", "usdt.erc20", "usdt.polygon", "usdt.bep20", "usdt.aptos", "usdt.solana", "usdt.xlayer", "usdt.arbitrum", "usdt.plasma", "usdt.ton",
		"usdc.trc20", "usdc.erc20", "usdc.polygon", "usdc.bep20", "usdc.aptos", "usdc.solana", "usdc.xlayer", "usdc.arbitrum", "usdc.base",
	},
}

// PaymentChannels returns the stable channel order accepted by a provider.
func PaymentChannels(provider string) []string {
	return append([]string(nil), paymentChannelOrder[provider]...)
}

// ValidatePaymentChannels validates the independently enabled provider channels.
func ValidatePaymentChannels(provider string, channels []string) error {
	allowed := ezpayRails
	if provider == "bepusdt" {
		allowed = bepusdtRails
	}
	if provider != "ezpay" && provider != "bepusdt" {
		return fmt.Errorf("unsupported provider %q", provider)
	}
	seen := make(map[string]struct{}, len(channels))
	for _, channel := range channels {
		channel = strings.ToLower(strings.TrimSpace(channel))
		if _, ok := allowed[channel]; !ok {
			return fmt.Errorf("unsupported channel %q", channel)
		}
		if _, ok := seen[channel]; ok {
			return fmt.Errorf("duplicate channel %q", channel)
		}
		seen[channel] = struct{}{}
	}
	return nil
}
