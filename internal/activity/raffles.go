package activity

// RaffleEntry is the authoritative receipt for one group message.
type RaffleEntry struct {
	DrawID    string `json:"drawId"`
	UserID    string `json:"userId"`
	Seats     int    `json:"seats"`
	Threshold int    `json:"threshold"`
	UserSeats int    `json:"userSeats"`
	FeeMinor  int64  `json:"feeMinor"`
	Replayed  bool   `json:"replayed"`
}

// RaffleSettlement groups all awarded results for one completed raffle.
type RaffleSettlement struct {
	Draw       LuckyDraw
	Results    []DrawResult
	NonWinners []string
	Reopened   bool
}
