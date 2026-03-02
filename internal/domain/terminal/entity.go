package terminal

import "time"

type Complaint struct {
	ID       int    `json:"id"`
	Category string `json:"category"`
	Text     string `json:"text"`
	Date     string `json:"date"`
	Status   string `json:"status"`
}

type Cassette struct {
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	Capacity float64 `json:"capacity"`
	Status   string  `json:"status"`
}

type ATM struct {
	ID       int     `json:"id"`
	Name     string  `json:"terminal_id"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	IsForte  bool    `json:"isForte"`
	District string  `json:"district"`
	Bank     string  `json:"bank,omitempty"`

	Address    string `json:"address"`
	Complaints string `json:"complaints"` // JSON string из базы

	EstWithdrawalKZT float64 `json:"estWithdrawalKZT,omitempty"`
	EstDepositKZT    float64 `json:"estDepositKZT,omitempty"`

	AvgCashBalanceKZT float64 `json:"avgCashBalanceKZT,omitempty"`
	TotalCashKZT      float64 `json:"totalCashKZT,omitempty"`

	WithdrawalFreqPerDay int     `json:"withdrawalFreqPerDay,omitempty"`
	DowntimePct          float64 `json:"downtimePct,omitempty"`
	EfficiencyStatus     string  `json:"efficiencyStatus,omitempty"`

	Cassettes []Cassette `json:"cassettes,omitempty"`
}

type CashBalance struct {
	TerminalID     int
	RecordTime     time.Time
	CurrentBalance int
	MaxCapacity    int
}
