# User preferences

`types.go` owns defaults, typed partial updates, snapshots, and the delivery reader.
`delivery.go` maps every automatic private notification to exactly one category.
The database stores nine preference flags; eligibility, language, currency, reset
automation, and Telegram tags retain their existing sources of truth. Unknown
events fail category resolution so new producers must explicitly choose a category.
`delivery_test.go` verifies category isolation for all current events.
