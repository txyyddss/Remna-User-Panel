# Display currency

`format.go` converts display-only TXB money with arbitrary-precision fixed-point arithmetic. It never changes authoritative balances, catalog prices, or payment settlement amounts.

The rate direction is `txb_per_currency`; conversion truncates toward zero at two target-currency decimal places.

`format_test.go` covers exact, fractional, signed, and invalid-rate cases.
