package liberr

import (
	"fmt"
	"strings"
)

type Error struct {
	cause     error
	operation Operation
	kind      Kind
	severity  Severity
}

func (e *Error) Kind() Kind {
	if e == nil {
		return ""
	}

	if len(e.kind) != 0 {
		return e.kind
	}

	t, ok := e.cause.(*Error)
	if !ok {
		return ""
	}

	return t.Kind()
}

//TODO: CHANGE IMPLEMENTATION
func (e *Error) EncodedStack() string {
	var encode func(d map[string]interface{}) string
	encode = func(d map[string]interface{}) string {
		b := new(strings.Builder)

		for k, v := range d {
			t, ok := v.(map[string]interface{})

			var val string

			if ok {
				val = encode(t)
			} else {
				val = fmt.Sprintf("%s", v)
			}

			b.WriteString(fmt.Sprintf(" %s:%s ", k, val))
		}

		return fmt.Sprintf("[%s]", strings.TrimSpace(b.String()))
	}

	return encode(e.Stack())
}

//TODO: CHANGE IMPLEMENTATION
func (e *Error) Stack() map[string]interface{} {
	res := make(map[string]interface{})

	if len(e.kind) != 0 {
		res["kind"] = e.kind
	}

	if len(e.operation) != 0 {
		res["operation"] = string(e.operation)
	}

	if len(e.severity) != 0 {
		res["severity"] = string(e.severity)
	}

	if e.cause != nil {
		t, ok := e.cause.(*Error)
		if ok {
			res["cause"] = t.Stack()
		} else {
			res["cause"] = e.cause.Error()
		}
	}

	return res
}

func (e *Error) Error() string {
	if e.cause == nil {
		return ""
	}

	return e.cause.Error()
}

func WithArgs(args ...interface{}) *Error {
	e := &Error{}

	//TODO: CHECK WHY DEFAULT IS NOT WORKING
	for _, arg := range args {
		switch t := arg.(type) {
		case Operation:
			e.operation = t
		case Kind:
			e.kind = t
		case Severity:
			e.severity = t
		case error:
			e.cause = t
		}
	}

	if e.cause == nil {
		return nil
	}

	return e
}

func WithOp(operation Operation, cause error) *Error {
	return WithArgs(operation, cause)
}
