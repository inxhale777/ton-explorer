package api

import "github.com/go-playground/validator/v10"

type TransactionDTO struct {
	Hash     string `json:"hash" validate:"required"`
	Account  string `json:"account" validate:"required"`
	Success  bool   `json:"success" validate:"required"`
	Datetime string `json:"datetime" validate:"required"`
	TotalFee string `json:"total_fee" validate:"required,gte=0"`
	Comment  string `json:"comment"`
}

func (t *TransactionDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(t)
}
