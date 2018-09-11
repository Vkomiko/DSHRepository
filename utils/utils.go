package utils

type DSHHRepError struct {
	msg string
	err error
	typ string
}

func Error(msg string) *DSHHRepError {
	return &DSHHRepError{msg:msg}
}

func LibavError(avErr error, msg string) *DSHHRepError {
	return &DSHHRepError{msg:msg, err:avErr, typ:"libav"}
}

func (e *DSHHRepError) Error() string {
	if e.err == nil {
		return e.msg
	}
	return e.msg + "\n    [" + e.typ + "] " + e.err.Error()
}

