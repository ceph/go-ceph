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

func ExampleSizer_update() {
	var err error
	for sizer := NewSizerEV(1, 128, errTooSmall); sizer.Continue(); err = sizer.Update(err) {
		buf := make([]string, sizer.Size())
		// do something complex with buf
		err = fakeComplexOp(buf)
	}
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

func ExampleSizer_updateWants() {
	var err error
	for sizer := NewSizerEV(1, 128, errTooSmall); sizer.Continue(); {
		size := sizer.Size()
		buf := make([]string, size)
		// do something complex with buf
		err = fakeComplexOp2(buf, &size)
		err = sizer.UpdateWants(err, size)
	}
	// Output:
	// too small: 1
	// good size: 30
}
