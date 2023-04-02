package errors

type Error struct {
	message string
}

func NewError(message string) *Error {
	return &Error{
		message: message,
	}
}

func (err *Error) Error() string {
	return err.message
}

// ErrDBExists is an error that is returned when a database already exists
var ErrDBExists = NewError("database already exists")

// ErrDBDoesNotExist is an error that is returned when a database does not exist
var ErrDBDoesNotExist = NewError("database does not exist")

// ErrNotEnoughFunds is an error that is returned when a transaction does not have enough funds
var ErrNotEnoughFunds = NewError("not enough funds")

// ErrTransactionNotFound is an error that is returned when a transaction is not found
var ErrTransactionNotFound = NewError("transaction not found")

// ErrInvalidTransaction is an error that is returned when a transaction is invalid
var ErrInvalidTransaction = NewError("invalid transaction")

// ErrInvalidAddress is an error that is returned when an address is invalid
var ErrInvalidAddress = NewError("invalid address")

// ErrBlockExists is an error that is returned when a block already exists
var ErrBlockExists = NewError("block already exists")

// ErrUnknownCommand is an error that is returned when an unknown command is received
var ErrUnknownCommand = NewError("unknown command")

// ErrUnknownGetDataType is an error that is returned when an unknown getdata type is received
var ErrUnknownGetDataType = NewError("unknown getdata type")
