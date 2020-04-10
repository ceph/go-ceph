package retry

import (
	"fmt"
)

var errTooSmall = fmt.Errorf("too small")

func fakeComplexOp(v []string) error {
	if len(v) < 30 {
		fmt.Println("too small:", len(v))
		return errTooSmall
	}
	fmt.Println("good size:", len(v))
	return nil
}

func ExampleWithSizes() {
	var err error
	WithSizes(1, 128, func(size int) Hint {
		buf := make([]string, size)
		// do something complex with buf
		err = fakeComplexOp(buf)
		return DoubleSize.If(err == errTooSmall)
	})
	// Output:
	// too small: 1
	// too small: 2
	// too small: 4
	// too small: 8
	// too small: 16
	// good size: 32
}

func fakeComplexOp2(v []string, s *int) error {
	if len(v) < 30 {
		fmt.Println("too small:", len(v))
		*s = 30
		return errTooSmall
	}
	fmt.Println("good size:", len(v))
	return nil
}

func ExampleWithSizes_hint() {
	var err error
	WithSizes(1, 128, func(size int) Hint {
		buf := make([]string, size)
		// do something complex with buf
		err = fakeComplexOp2(buf, &size)
		return Size(size).If(err == errTooSmall)
	})
	// Output:
	// too small: 1
	// good size: 30
}
