package errors

import (
	"github.com/sirupsen/logrus"
)

type Operation string

type ErrorType string

const (
	NotFoundError     ErrorType = "NOT_FOUND"
	UnAuthorizedError ErrorType = "UNAUTHORIZED"
	Unexpected        ErrorType = "UNEXPECTED"
)

//
type Error struct {
	operations []Operation
	errorType  ErrorType
	error      error
	severity   logrus.Level
}

//
func NewError(operation Operation, errorType ErrorType, err error, severity logrus.Level) *Error {
	return &Error{
		operations: []Operation{operation},
		errorType:  errorType,
		error:      err,
		severity:   severity,
	}
}

//
func (e *Error) WithOperation(operation Operation) *Error {
	e.operations = append(e.operations, operation)
	return e
}

//
func (e *Error) Operations() []Operation {
	return e.operations
}

//
func (e *Error) ErrorType() ErrorType {
	return e.errorType
}

//
func (e *Error) Error() error {
	return e.error
}

//
func (e *Error) Severity() logrus.Level {
	return e.severity
}
