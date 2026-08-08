package coupons

import (
	"errors"
	"math"
	"testing"
)

func TestCanonicalCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "trim and uppercase", input: "  save-20 ", want: "SAVE-20"},
		{name: "underscore", input: "draw_bonus", want: "DRAW_BONUS"},
		{name: "too short", input: "abc", wantErr: true},
		{name: "spaces", input: "bad code", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := CanonicalCode(test.input)
			if got != test.want || test.wantErr != errors.Is(err, ErrInvalidInput) {
				t.Fatalf("CanonicalCode(%q) = (%q, %v), want (%q, invalid=%t)", test.input, got, err, test.want, test.wantErr)
			}
		})
	}
}

func TestCalculateDiscount(t *testing.T) {
	t.Parallel()

	capMinor := int64(225)
	tests := []struct {
		name    string
		coupon  Coupon
		gross   int64
		want    int64
		wantErr bool
	}{
		{name: "fixed", coupon: Coupon{CouponInput: CouponInput{Kind: KindPurchaseOnce, DiscountMode: DiscountFixed, ValueMinorOrBPS: 300}}, gross: 1_000, want: 300},
		{name: "fixed clamped", coupon: Coupon{CouponInput: CouponInput{Kind: KindPurchaseOnce, DiscountMode: DiscountFixed, ValueMinorOrBPS: 3_000}}, gross: 1_000, want: 1_000},
		{name: "percentage floor", coupon: Coupon{CouponInput: CouponInput{Kind: KindPurchaseRecurring, DiscountMode: DiscountPercent, ValueMinorOrBPS: 2_500}}, gross: 999, want: 249},
		{name: "percentage cap", coupon: Coupon{CouponInput: CouponInput{Kind: KindPurchaseRecurring, DiscountMode: DiscountPercent, ValueMinorOrBPS: 5_000, PercentCapMinor: &capMinor}}, gross: 1_000, want: 225},
		{name: "wrong kind", coupon: Coupon{CouponInput: CouponInput{Kind: KindBalanceAdd}}, gross: 1_000, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := CalculateDiscount(test.coupon, test.gross)
			if got != test.want || test.wantErr != errors.Is(err, ErrInvalidInput) {
				t.Fatalf("CalculateDiscount() = (%d, %v), want (%d, invalid=%t)", got, err, test.want, test.wantErr)
			}
		})
	}
}

func TestCalculateBalanceMultiplyCredit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		balance    int64
		multiplier int64
		want       int64
		wantErr    bool
	}{
		{name: "one and a half", balance: 101, multiplier: 15_000, want: 50},
		{name: "zero balance", balance: 0, multiplier: 20_000, want: 0},
		{name: "invalid multiplier", balance: 100, multiplier: 10_000, wantErr: true},
		{name: "overflow", balance: math.MaxInt64, multiplier: 20_000, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := CalculateBalanceMultiplyCredit(test.balance, test.multiplier)
			if got != test.want || test.wantErr != errors.Is(err, ErrInvalidInput) {
				t.Fatalf("CalculateBalanceMultiplyCredit() = (%d, %v), want (%d, invalid=%t)", got, err, test.want, test.wantErr)
			}
		})
	}
}
