package billing

import "testing"

func TestParseDecimal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		raw       string
		canonical string
		wantErr   bool
	}{
		{name: "zero", raw: "0", canonical: "0"},
		{name: "integer", raw: "42", canonical: "42"},
		{name: "trims whitespace", raw: " 001.2300 ", canonical: "1.23"},
		{name: "fraction below one", raw: "0.0050", canonical: "0.005"},
		{name: "empty", raw: "", wantErr: true},
		{name: "leading decimal point", raw: ".5", wantErr: true},
		{name: "trailing decimal point", raw: "1.", wantErr: true},
		{name: "negative", raw: "-1", wantErr: true},
		{name: "explicit positive sign", raw: "+1", wantErr: true},
		{name: "exponent", raw: "1e3", wantErr: true},
		{name: "multiple points", raw: "1.2.3", wantErr: true},
		{name: "non digit", raw: "1,25", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			decimal, err := ParseDecimal(test.raw)
			if test.wantErr {
				if err == nil {
					t.Fatalf("ParseDecimal(%q) unexpectedly succeeded", test.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDecimal(%q): %v", test.raw, err)
			}
			if got := decimal.Canonical(); got != test.canonical {
				t.Fatalf("Canonical() = %q, want %q", got, test.canonical)
			}
		})
	}
}

func TestPayableRoundsUpToProviderPrecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		txbMinor  int64
		rate      string
		precision int
		want      string
	}{
		{name: "exact CNY hundredths", txbMinor: 101, rate: "1", precision: 2, want: "1.01"},
		{name: "CNY rounds fractional cent upward", txbMinor: 1, rate: "0.333", precision: 2, want: "0.01"},
		{name: "CNY carry into integer", txbMinor: 123, rate: "6.5", precision: 2, want: "8.00"},
		{name: "USD preserves provider precision", txbMinor: 25, rate: "0.04", precision: 2, want: "0.01"},
		{name: "Stars exact integer", txbMinor: 200, rate: "3", precision: 0, want: "6"},
		{name: "Stars round upward", txbMinor: 1, rate: "1", precision: 0, want: "1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rate, err := ParseDecimal(test.rate)
			if err != nil {
				t.Fatalf("ParseDecimal(%q): %v", test.rate, err)
			}
			got, err := Payable(test.txbMinor, rate, test.precision)
			if err != nil {
				t.Fatalf("Payable(): %v", err)
			}
			if got != test.want {
				t.Fatalf("Payable(%d, %q, %d) = %q, want %q", test.txbMinor, test.rate, test.precision, got, test.want)
			}
		})
	}
}

func TestPayableRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	positive, err := ParseDecimal("1")
	if err != nil {
		t.Fatalf("ParseDecimal(): %v", err)
	}
	zero, err := ParseDecimal("0")
	if err != nil {
		t.Fatalf("ParseDecimal(): %v", err)
	}

	tests := []struct {
		name      string
		txbMinor  int64
		rate      Decimal
		precision int
	}{
		{name: "zero TXB", txbMinor: 0, rate: positive, precision: 2},
		{name: "negative TXB", txbMinor: -1, rate: positive, precision: 2},
		{name: "zero rate", txbMinor: 100, rate: zero, precision: 2},
		{name: "negative precision", txbMinor: 100, rate: positive, precision: -1},
		{name: "excessive precision", txbMinor: 100, rate: positive, precision: 13},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := Payable(test.txbMinor, test.rate, test.precision); err == nil {
				t.Fatal("Payable() unexpectedly succeeded")
			}
		})
	}
}

func TestEquivalent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		left       string
		right      string
		equivalent bool
	}{
		{name: "trailing zeroes", left: "1.00", right: "1", equivalent: true},
		{name: "different scales", left: "0.010", right: "0.01", equivalent: true},
		{name: "different values", left: "1.01", right: "1.001", equivalent: false},
		{name: "invalid left", left: "NaN", right: "0", equivalent: false},
		{name: "invalid right", left: "0", right: "-0", equivalent: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := Equivalent(test.left, test.right); got != test.equivalent {
				t.Fatalf("Equivalent(%q, %q) = %t, want %t", test.left, test.right, got, test.equivalent)
			}
		})
	}
}
