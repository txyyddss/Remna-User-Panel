# Draw reward distributions

`ValueRange.Roll` and `RollAround` sample reward values after instant prize selection
or raffle stock assignment. Production uses `CryptoRandom` (`crypto/rand`); callers
inject `RandomSource` for deterministic CI coverage. This module makes no network
calls and stores no additional data.

## Units and bounds

Ranges contain inclusive integer ticks: TXB cents, percentage basis points,
multiplier ten-thousandths, GiB, or hours. Fixed ranges return their only value.
Mixed money/traffic signs and multiplier ranges crossing 1x choose the side using
`positiveChanceBps` first, exclude the neutral tick, and sample within that side.
Lower numerical values have lower operator cost, including negative rewards.

## Algorithms

- Uniform: sample every tick with equal probability.
- Gaussian: the existing truncated half-normal distribution, centered at the
  lowest-cost tick with scale equal to one third of the tick count.
- Power law: a bounded power distribution with shape `a = 1/2`. On the normalized
  range its CDF is `sqrt(x)`, so the inverse transform is `x = u²`. Draw an unbiased
  53-bit integer, divide by `2^53` to obtain `u` in `[0,1)`, scale by the inclusive
  tick count, then floor to a tick. Both configured endpoints are reachable.
  The shape follows the [power-function distribution definition](https://docs.scipy.org/doc/scipy/reference/generated/scipy.stats.powerlaw.html).

The power-law shape is a fixed product choice: the lower quarter of the range
receives approximately half the outcomes, while the mean lies approximately one
third of the way from minimum to maximum. Using the normalized range keeps this
bias stable when the reward unit or precision changes. Quantization accounts for
small differences at coarse resolutions. No SciPy dependency is used.

## 2026-10-01 correction

The previous formula was `floor(1 / (1 - u * (1 - 1/N)) - 1)`, with `N` equal to
the tick count. This applies a truncated Pareto curve directly to tick ranks:
approximately 50% of draws become the minimum, regardless of range width. It also
leaves the highest tick unreachable in exact arithmetic. More decimal precision
therefore concentrates rewards ever closer to the minimum displayed value.

For the reported 30.00–100.00% prize, `N = 7001`. The supplied 15 results include
nine exact minima, range from 30.00% to 30.13%, and average about 30.0167%. Under
the old formula, about 93.35% of draws fall at or below 30.13%; all 15 doing so
has probability about 35.6%. This is consistent with the implemented bias and
does not indicate a failure of the cryptographic random source.

The corrected curve gives approximately:

| Property | 30.00–100.00% coupon |
| --- | --- |
| Exactly 30.00% | 1.20% probability |
| Median | 47.50% discount |
| Mean | 53.33% discount |
| 90th percentile | 86.70% discount |

This intentionally increases expected rewards relative to the old implementation.
Existing `power_law` configurations use the new curve for future sampling,
including unsettled raffles. Persisted results, coupons, balances and stock
assignments are not rerolled. Uniform and Gaussian behavior remains unchanged.
Forecasts currently express expense bounds rather than distribution-weighted
means; those bounds remain valid.

## Validation and extension

`draw_random_test.go` uses fixed quantiles and stratified uniform inputs to check
shape, precision independence, endpoints, error propagation and side selection
without probabilistic test failures. `AGENTS.md` prohibits local automated tests;
execute these through hosted CI. Any future shape parameter must have explicit
validation and documented probability semantics before being exposed in a draw
configuration.
