package api

type TransactionDTO struct {
	Hash    string `json:"hash"`
	Account string `json:"account"`
	Success bool   `json:"success"`
	// LogicalTime uint64 `json:"logical_time"`
	Datetime string `json:"datetime"`
	TotalFee string `json:"total_fee"`
	Comment  string `json:"comment"`
}
