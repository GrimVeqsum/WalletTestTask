package internal

import "errors"

var ErrInvalidAmount = errors.New("amount is incorrect")
var ErrInvalidOperation = errors.New("operation is incorrect")
var ErrWalletNotFound = errors.New("wallet not found")
var ErrInsufficientFunds = errors.New("there are not enough funds in the balance")
