package rados

import "C"

import (
	"fmt"
)

// The file operation.go exists to support both read op and write op types that
// have some pretty common behaviors between them. In C/C++ its assumed that
// the buffer types and other pointers will not be freed between passing them
// to the action setup calls (things like rados_write_op_write or
// rados_read_op_omap_get_vals2) and the call to Operate(...).  Since there's
// nothing stopping one from sleeping for hours between these calls, or passing
// the op to other functions and calling Operate there, we want a mechanism
// that's (fairly) simple to understand and won't run afoul of Go's garbage
// collection.  That's one reason the operation type tracks the elements (the
// parts that track complex inputs and outputs) so that as long as the op
// exists it will have a reference to the element, which will have references
// to the C language types.

// OperationError is an error type that may be returned an Operate call and
// captures both the error from operating any any data conversion errors that
// may have occurred.
type OperationError struct {
	flavor        string
	OpError       error
	ElementErrors []error
}

func (e OperationError) Error() string {
	s := fmt.Sprintf("%s operation error:", e.flavor)
	count := 0
	if e.OpError != nil {
		s = fmt.Sprintf("%s (op) %s", s, e.OpError.Error())
		count++
	}
	var sep string
	for i, ee := range e.ElementErrors {
		sep = ""
		if count > 0 {
			sep = ","
		}
		s = fmt.Sprintf("%s%s (%d) %s", s, sep, i+1, ee.Error())
		count++
	}
	return s
}

// Unwrap will return the first error returned from an operation or nil
// if no error is in set.
func (e OperationError) Unwrap() error {
	if e.OpError != nil {
		return e.OpError
	}
	for _, ee := range e.ElementErrors {
		return ee
	}
	return nil
}

// opElement provides an interface for types that are tied to the management of
// data being input or output from write ops and read ops. The elements are
// meant to simplify the internals of the ops themselves and be exportable when
// appropriate. If an element is not being exported it should not be returned
// from an ops action function. If the element is exported it should be
// returned from an ops action function.
//
// Not all types implementing opElement are expected to need all the functions
// in the interface. However, for the sake of simplicity on the op side, we use
// the same interface for all cases and expect those implementing opElement
// just add empty funcs. Since this in a non-public interface this should not
// be much of a burden, hopefully.
type opElement interface {
	// reset will be called before the op's Operate function.  It can be used
	// to reset state between uses of Operate, as it can be called multiple
	// times.
	reset()
	// update will be called after the op's Operate function. It can be used
	// to convert values from C and cache them and/or communicate a failure
	// of the action associated with the element.
	update() error
	// free will be called to free any resources, especially C memory, that
	// the element is managing.
	free()
}

// freeElement calls the free method of the opElement if it is valid.
func freeElement(oe opElement) {
	if oe != nil {
		oe.free()
	}
}

// operation represents some of the shared underlying mechanisms for
// both read and write op types.
type operation struct {
	elements []opElement
}

// reset all of the elements this operation contains.
func (o *operation) reset() {
	for i := range o.elements {
		o.elements[i].reset()
	}
}

// finish the operation (the Operate call) by updating all the elements and
// collecting any errors from them or the librados operate call.
func (o *operation) finish(flavor string, ret C.int) error {
	elementErrors := make([]error, 0)
	for i := range o.elements {
		err := o.elements[i].update()
		if err != nil {
			elementErrors = append(elementErrors, err)
		}
	}
	if ret == 0 && len(elementErrors) == 0 {
		return nil
	}
	return OperationError{
		flavor:        flavor,
		OpError:       getError(ret),
		ElementErrors: elementErrors,
	}
}

// freeElements will call the free method of all the elements this operation
// contains.
func (o *operation) freeElements() {
	for i := range o.elements {
		freeElement(o.elements[i])
	}
}
