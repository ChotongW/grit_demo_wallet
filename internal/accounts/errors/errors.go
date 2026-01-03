package errors

import "errors"

var (
	ErrAccountNotFound              = errors.New("account not found")
	ErrEmailAlreadyExists           = errors.New("email already exists")
	ErrInsufficientBalance          = errors.New("insufficient balance")
	ErrInvalidReferrer              = errors.New("invalid referrer")
	ErrTransferToSameAccount        = errors.New("cannot transfer to the same account")
	ErrDepositAmountMustBePositive  = errors.New("deposit amount must be positive")
	ErrWithdrawAmountMustBePositive = errors.New("withdrawal amount must be positive")
	ErrTransferAmountMustBePositive = errors.New("transfer amount must be positive")
)
