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

// ErrorDBExists is an error that is returned when a database already exists
var ErrorDBExists = NewError("database already exists")

// ErrorDBDoesNotExist is an error that is returned when a database does not exist
var ErrorDBDoesNotExist = NewError("database does not exist")

// ErrorNotEnoughFunds is an error that is returned when a transaction does not have enough funds
var ErrorNotEnoughFunds = NewError("not enough funds")
