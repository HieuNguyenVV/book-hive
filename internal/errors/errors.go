package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

type AppError struct {
	StatusCode    int    `json:"-"`
	ErrorIndicate bool   `json:"error"`
	Code          int    `json:"code"`
	Message       string `json:"message"`
	TraceID       string `json:"trace_id"`

	WrappedErr  error  `json:"-"`
	CallStack   string `json:"-"`
	Name        string `json:"-"`
	Description string `json:"-"`
} // @name AppError

func (e AppError) Error() string {
	data := map[string]interface{}{}

	if e.WrappedErr != nil {
		data["wrapped_err"] = e.WrappedErr.Error()
		data["cause"] = errors.Cause(e.WrappedErr).Error()
	}

	data["callStack"] = e.CallStack
	data["error"] = e.ErrorIndicate
	data["code"] = e.Code
	data["status_code"] = e.StatusCode
	data["message"] = e.Message
	data["trace_id"] = e.TraceID
	dataBytes, _ := json.Marshal(data)
	return string(dataBytes)
}

type stack []uintptr

func (e *AppError) callersStack() {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	stackString := make([]string, 0, 5)
	for i, pc := range st {
		if i < 5 {
			f := errors.Frame(pc)
			stack := fmt.Sprintf("%+v", f)
			split := strings.Split(stack, "\t")
			stackString = append(stackString, split[1])
		}
	}
	e.CallStack = strings.Join(stackString, "::")
}

func NewAppError(code int, name string, status int, message string, description string) AppError {
	return AppError{
		Name:          name,
		ErrorIndicate: true,
		Code:          code,
		StatusCode:    status, //deprecated
		Message:       message,
		Description:   description,
	}
}

func (e AppError) Unwrap() error {
	return e.WrappedErr
}

func (e AppError) Cause() error {
	return errors.Cause(e.WrappedErr)
}

func (e AppError) Reform(msg string, args ...interface{}) AppError {
	e.callersStack()
	if len(args) > 0 {
		e.Message = fmt.Sprintf(msg, args...)
	} else {
		e.Message = msg
	}
	return e
}

func (e AppError) Wrap(err error) AppError {
	if err != nil {
		e.callersStack()
		e.WrappedErr = err
	}
	return e
}

func (e AppError) WrapString(err string) AppError {
	if len(err) != 0 {
		e.WrappedErr = errors.Wrap(e.WrappedErr, err)
	}

	return e
}

func (e AppError) AppendTraceID(traceID string) AppError {
	if len(traceID) > 0 {
		e.TraceID = traceID
	}
	return e
}

func (e AppError) SetTraceID(traceID string) AppError {
	if len(traceID) > 0 {
		e.TraceID = traceID
	}
	return e
}

func (e AppError) Is(target error) bool {
	if e.Name == "" || target == nil {
		return false
	}
	switch t := target.(type) {
	case AppError:
		return e.Name == t.Name && e.Code == t.Code
	case *AppError:
		return t != nil && e.Name == t.Name && e.Code == t.Code
	default:
		return false
	}
}
