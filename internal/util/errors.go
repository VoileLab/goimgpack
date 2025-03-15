package util

import (
	"errors"
	"fmt"
	"runtime"
)

// Errorf is a wrapper of fmt.Errorf, it will add the caller function name as prefix.
func Errorf(format string, args ...any) error {
	pc, _, _, ok := runtime.Caller(1)
	prefix := "unknow: "
	if ok {
		prefix = fmt.Sprintf("%s: ", runtime.FuncForPC(pc).Name())
	}
	return fmt.Errorf(prefix+format, args...)
}

// NewError is a wrapper of errors.New, it will add the caller function name as prefix.
func NewError(info string) error {
	pc, _, _, ok := runtime.Caller(1)
	prefix := "unknow: "
	if ok {
		prefix = fmt.Sprintf("%s: ", runtime.FuncForPC(pc).Name())
	}
	return errors.New(prefix + info)
}
