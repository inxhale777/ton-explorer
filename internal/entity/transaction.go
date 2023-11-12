package entity

type Transaction struct {
	Hash        string
	Account     string
	Success     bool
	LogicalTime uint64
	TotalFee    string
	Comment     string
}
