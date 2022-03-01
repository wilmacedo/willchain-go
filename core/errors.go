package core

import "errors"

var ErrInvalidAddress = errors.New("address is not valid")
var ErrEnoughFunds = errors.New("not enough funds")

var ErrNilPreviousTransactions = errors.New("previous transactions doest not exist")
var ErrNilTransaction = errors.New("transaction doest not exist")
