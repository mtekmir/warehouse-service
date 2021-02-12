package errors

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

// Error implements the errors.Error.
type Error struct {
	Op      Op     // Operation being performed
	Kind    Kind   // Kind of the error
	Message string // Error message that will be sent to the client.
	Err     error  // Underlying error
}

// Op describes an operation, usually as the package and method
type Op string

// Kind describes the kind of the error
type Kind int

// Kinds of errors
const (
	Other        Kind = iota // Unclassified err
	Unauthorized             // Unauthorized
	NotFound                 // Item does not exist
	Duplicate                // Duplicate
	Invalid                  // Invalid input
	Unavailable              // Resource unavailable
)

func (k Kind) String() string {
	switch k {
	case Other:
		return "Other error"
	case Unauthorized:
		return "Unauthorized"
	case NotFound:
		return "NotFound"
	case Duplicate:
		return "Duplicate"
	case Invalid:
		return "Invalid input"
	case Unavailable:
		return "Service unavailable"
	default:
		return "Unknown error kind"
	}
}

// Code returns http response code.
func (e *Error) Code() int {
	switch e.Kind {
	case Unauthorized:
		return 401
	case NotFound:
		return 404
	case Duplicate:
		return 409
	case Invalid:
		return 400
	case Unavailable:
		return 503
	default:
		return 500
	}
}

// Ops returns the "stack" of operations of the error.
func (e *Error) Ops() []Op {
	ops := []Op{e.Op}
	err, ok := e.Err.(*Error)
	if !ok {
		return ops
	}
	return append(err.Ops(), ops...)
}

// Cause returns the original error that was passed to the E func.
func (e *Error) Cause() error {
	if err, ok := e.Err.(*Error); ok {
		return err.Cause()
	}
	return e.Err
}

func (e *Error) Error() string {
	fields := make([]string, 0, 4)

	if e.Message != "" {
		fields = append(fields, fmt.Sprintf("Message: %s", e.Message))
	}

	fields = append(fields, fmt.Sprintf("Kind: %s", e.Kind.String()))

	if e.Cause() != nil {
		fields = append(fields, fmt.Sprintf("Cause: %v", e.Cause()))
	}

	fields = append(fields, fmt.Sprintf("StackTrace: %s", e.Ops()))

	return strings.Join(fields, ", ")
}

// Body returns http response body.
func (e *Error) Body() []byte {
	m := e.Message
	if m == "" {
		m = "Something went wrong"
	}
	return []byte(fmt.Sprintf(`{"message": "%s"}`, m))
}

// E builds an error value from its arguments.
// There must be at least one argument or E panics.
// The type of each argument determines its meaning.
// If more than one argument of a given type is presented,
// only the last one is recorded
// The types are:
// errors.Op: name of the operation
// errors.Kind: kind of the error, default is Other
// string: Message
// error: The underlying error
func E(args ...interface{}) error {
	if len(args) == 0 {
		panic("call to errors.E with no arguments")
	}
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			e.Op = arg
		case string:
			e.Message = arg
		case Kind:
			e.Kind = arg
		case *Error:
			// Make a copy
			copy := *arg
			e.Err = &copy
		case error:
			e.Err = arg
		default:
			_, file, line, _ := runtime.Caller(1)
			log.Printf("errors.E: bad call from %s:%d: %v", file, line, args)
			return fmt.Errorf("unknown type %T, value %v in error call", arg, arg)
		}
	}
	prev, ok := e.Err.(*Error)
	if !ok {
		return e
	}

	// If this error has Kind unset or Other, pull up the inner one.
	if e.Kind == Other {
		e.Kind = prev.Kind
		prev.Kind = Other
	}
	// If this error has Message unset, pull up the inner one.
	if e.Message == "" {
		e.Message = prev.Message
		prev.Message = ""
	}
	return e
}
