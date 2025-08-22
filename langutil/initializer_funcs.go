package langutil

import (
	"errors"
)

type InitializerFunc func(Args) error

var initializerFuncs []InitializerFunc

func RegisterInitializerFunc(f InitializerFunc) {
	initializerFuncs = append(initializerFuncs, f)
}

func CallInitializerFuncs(args Args) (err error) {
	var errs []error
	for _, f := range initializerFuncs {
		errs = append(errs, f(args))
	}
	return errors.Join(errs...)
}
