package wrapper

import (
	"errors"
	"path"
	"runtime"
	"strconv"
	"strings"
)

type UnwrapErr interface {
	Unwrap() error
}

var _ UnwrapErr = (*wrapError)(nil)

type wrapError struct {
	caller   string
	err      error
	funcName string
	message  string
}

func newError(err error, message string) error {
	caller, funcName := getCaller()
	if message != "" {
		message = "\t" + message
	}

	return &wrapError{
		caller:   caller,
		err:      err,
		message:  message,
		funcName: funcName,
	}
}

func (e *wrapError) Error() string {
	getMessage := func() string {
		if e.message == "" {
			return ""
		}

		return "\t" + e.message
	}

	str := e.caller + "\t" + e.funcName + getMessage()

	var errs *wrapError
	if errors.As(e.err, &errs) {
		return str + "\n" + e.err.Error()
	}

	return str + "\t" + e.err.Error()
}

func getCaller() (caller string, funcName string) {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return
	}

	_, fileName := path.Split(file)

	funcNameForPC := runtime.FuncForPC(pc).Name()

	ix := strings.LastIndex(funcNameForPC, ".")

	caller = funcNameForPC[0:ix] + "/" + fileName + ":" + strconv.Itoa(line)

	if len(funcNameForPC) > ix {
		funcName = funcNameForPC[ix+1:] + "()"
	}

	return
}

// Unwrap errors
func (e *wrapError) Unwrap() error {
	return e.err
}

// Is errors
func (e *wrapError) Is(target error) bool {
	return e.err.Error() == target.Error()
}

func UnwrapOriginal(err error) error {
	u, ok := err.(UnwrapErr) //nolint:errorlint
	if ok {
		return UnwrapOriginal(u.Unwrap())
	}

	return err
}
