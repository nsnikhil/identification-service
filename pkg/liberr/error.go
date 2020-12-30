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
	var encode func(d []pair) string
	encode = func(d []pair) string {
		b := new(strings.Builder)

		for _, v := range d {
			t, ok := v.value.([]pair)

			var val string

			if ok {
				val = encode(t)
			} else {
				val = fmt.Sprintf("%s", v.value)
			}

			b.WriteString(fmt.Sprintf(" %s:%s ", v.key, val))
		}

		return fmt.Sprintf("[%s]", strings.TrimSpace(b.String()))
	}

	return encode(e.stack())
}

type pair struct {
	key   string
	value interface{}
}

//TODO: CHANGE IMPLEMENTATION
func (e *Error) stack() []pair {
	res := make([]pair, 0)

	if len(e.kind) != 0 {
		res = append(res, pair{key: "kind", value: e.kind})
	}

	if len(e.operation) != 0 {
		res = append(res, pair{key: "operation", value: e.operation})
	}

	if len(e.severity) != 0 {
		res = append(res, pair{key: "severity", value: e.severity})
	}

	if e.cause != nil {
		t, ok := e.cause.(*Error)
		if ok {
			res = append(res, pair{key: "cause", value: t.stack()})
		} else {
			res = append(res, pair{key: "cause", value: e.cause.Error()})
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
			if len(e.operation) != 0 {
				return nil
			}

			e.operation = t
		case Kind:
			if len(e.kind) != 0 {
				return nil
			}

			e.kind = t
		case Severity:
			if len(e.severity) != 0 {
				return nil
			}

			e.severity = t
		case error:
			if e.cause != nil {
				return nil
			}

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
