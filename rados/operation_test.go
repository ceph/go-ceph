package rados

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperationError(t *testing.T) {
	oe := OperationError{
		kind:    readOp,
		OpError: fmt.Errorf("bad mojo %v", 77),
		StepErrors: map[int]error{
			1: fmt.Errorf("limit exceeded"),
			3: fmt.Errorf("weirdness"),
		},
	}
	assert.Error(t, oe)
	estr := oe.Error()
	assert.Contains(t, estr, "read operation error")
	assert.Contains(t, estr, "op=bad mojo 77")
	assert.Contains(t, estr, "Step#1=limit exceeded")
	assert.Contains(t, estr, "Step#3=weirdness")

	oe = OperationError{
		kind: writeOp,
		StepErrors: map[int]error{
			0: fmt.Errorf("unlucky"),
		},
	}
	assert.Error(t, oe)
	estr = oe.Error()
	assert.Contains(t, estr, "write operation error")
	assert.NotContains(t, estr, "op=")
	assert.Contains(t, estr, "Step#0=unlucky")
}

type fooStep struct {
	updateCount int
	freeCount   int
	failMe      error
}

func (fs *fooStep) update() error {
	fs.updateCount++
	return fs.failMe
}

func (fs *fooStep) free() {
	fs.freeCount++
}

func TestOpStepFinalizer(t *testing.T) {
	// this confirms that given a valid op step the finalizer helper
	// function calls the free method of the interface
	fs := &fooStep{}
	opStepFinalizer(fs)
	opStepFinalizer(fs)
	assert.Equal(t, 2, fs.freeCount)

	// this should not panic
	opStepFinalizer(nil)
}

func TestOperationType(t *testing.T) {
	t.Run("stepErrors", func(t *testing.T) {
		o := &operation{}
		o.steps = append(o.steps, &fooStep{})
		o.steps = append(o.steps, &fooStep{failMe: fmt.Errorf("yow")})
		o.steps = append(o.steps, &fooStep{})
		o.steps = append(o.steps, &fooStep{})
		o.steps = append(o.steps, &fooStep{failMe: fmt.Errorf("ouch")})

		err := o.update(readOp, 0)
		if assert.Error(t, err) {
			oe := err.(OperationError)
			assert.NoError(t, oe.OpError)
			assert.Len(t, oe.StepErrors, 2)
			assert.Contains(t, oe.StepErrors, 1)
			assert.Contains(t, oe.StepErrors, 4)
		}
	})

	t.Run("opError", func(t *testing.T) {
		o := &operation{}
		o.steps = append(o.steps, &fooStep{})
		o.steps = append(o.steps, &fooStep{})
		o.steps = append(o.steps, &fooStep{})
		o.steps = append(o.steps, &fooStep{})

		err := o.update(readOp, 5)
		if assert.Error(t, err) {
			oe := err.(OperationError)
			assert.Error(t, oe.OpError)
			assert.Len(t, oe.StepErrors, 0)
		}
	})

	t.Run("noErrors", func(t *testing.T) {
		o := &operation{}
		o.steps = append(o.steps, &fooStep{})
		o.steps = append(o.steps, &fooStep{})
		err := o.update(readOp, 0)
		assert.NoError(t, err)
		x := 0
		for i := range o.steps {
			x += o.steps[i].(*fooStep).updateCount
		}
		assert.Equal(t, 2, x)

		o.free()
		x = 0
		for i := range o.steps {
			x += o.steps[i].(*fooStep).updateCount
		}
		assert.Equal(t, 2, x)
	})
}

type refMock struct {
	s   string
	out *string
}

func (v refMock) Free() {
	*v.out += v.s
}

func TestOperationRefFreeOrder(t *testing.T) {
	r := withRefs{}
	var out string
	r.add(refMock{"bar", &out})
	r.add(refMock{"foo", &out})
	r.free()
	assert.Equal(t, out, "foobar")
}
