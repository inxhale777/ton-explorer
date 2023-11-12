package wrapper

import "fmt"

func Wrap(err error) error {
	if err == nil {
		return nil
	}

	return newError(err, "")
}

func Wrapf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}

	return newError(err, fmt.Sprintf(format, a...))
}
