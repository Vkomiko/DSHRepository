package utils

import "fmt"

type ErrorType int

const (
	Error ErrorType = iota
	EOFError
	OverflowError
	InitializeError
)

//go:generate T:\win\bin\stringer -type=ErrorType,ControlMsg -output types_string.go

type StdError interface {
	error
	Code () int
	Type () int
}

type DSHStdError struct {
	Ecode int
	Etyp int
	Emsg string
}

func (t ErrorType) New(code int, msg string) error {
	return &DSHStdError{code, int(t), msg}
}

func (e *DSHStdError) Error() string {
	return fmt.Sprintf("%s [%d] %s", ErrorType(e.Etyp).String(), e.Ecode, e.Emsg)
}

func (e *DSHStdError) Code() int {
	return e.Ecode
}

func (e *DSHStdError) Type() int {
	return e.Etyp
}

func (e *DSHStdError) Msg() string {
	return e.Emsg
}

//----------------------------------------------------------------------------------------------------------------------

type ControlMsg int

const (
	CMsgEOF ControlMsg = iota
	CMsgStarve
)

func (m ControlMsg) Error() string {
	return fmt.Sprintf("[Control Message] %s", m.String())
}
