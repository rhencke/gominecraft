package error

import "os"
import "fmt"
import "runtime"

type Error interface {
	os.Error
	Inner() Error
}

type error struct {
	err   os.Error
	inner Error
}

func NewError(message string, inner os.Error) Error {
	caller, _, _, ok := runtime.Caller(1)
	var callerName string
	if !ok {
		callerName = "<unknown>"
	} else {
		callerName = runtime.FuncForPC(caller).Name()
	}
	msg := fmt.Sprint(callerName, ": ", message)
	var inerr Error
	if inner != nil {
		inerr = newOsOnly(inner)
	}
	return error{err: (os.ErrorString)(msg), inner: inerr}
}
func newOsOnly(err os.Error) Error {
	return error{err: err}
}

func (err error) String() string {
	inner := err.inner
	var str string
	if inner != nil {
		str = fmt.Sprint(err.err.String(), "\n-> ", inner.String())
	} else {
		str = err.err.String()
	}
	return str
}

func (err error) Inner() Error {
	return err.inner
}
