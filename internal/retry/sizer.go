package retry

// SizerCheckFunc allows the creator of a Sizer to specify arbitrarily
// complex checks of the error condition to determine if the size
// needs to be retried.
type SizerCheckFunc func(error) bool

// Sizer is used to implement 'resize loops' that hides the complexity
// of the sizing and error checking away from most of the application.
//
// At the end of each iteration call Update or UpdateWants to inform
// the sizer of the result of the action(s) taken during the iteration.
// If the error is nil the update functions will return nil and halt
// the iteration. If the error is non nil but returns false from
// the check function the same error is returned and iteration will halt.
// If the check function returns true iteration continues but nil
// will be returned from the update function, effectively clearing
// the error state, *unless* the current size now exceeds maxSize.
//
// When the Sizer's Update & UpdateWants methods are called the size will be
// doubled, unless UpdateWants is provided with a size hint that is greater
// than the current size. Therefore, it is highly recommended to provide start
// and max size values that are powers of two.
type Sizer struct {
	size    int
	maxSize int
	ready   bool
	again   SizerCheckFunc
}

// NewSizer returns a new Sizer ready for use.
// The Sizer will attempt a retry on conditions where the callback
// function f returns true.
func NewSizer(startSize, maxSize int, f SizerCheckFunc) *Sizer {
	return &Sizer{
		ready:   true,
		size:    startSize,
		maxSize: maxSize,
		again:   f,
	}
}

// NewSizerEV returns a new Sizer ready for use that only checks
// for the exact error value specified.
func NewSizerEV(startSize, maxSize int, retryOn error) *Sizer {
	return NewSizer(
		startSize,
		maxSize,
		func(err error) bool { return err == retryOn },
	)
}

// Continue returns true if the application should try the desired
// action.
func (s *Sizer) Continue() bool {
	return s.ready
}

// Size currently specified by the sizer. Only changes when Update or
// UpdateWants is called, so it is safe to call multiple times within
// a single iteration.
func (s *Sizer) Size() int {
	return s.size
}

// Update the Sizer with the results of the action taken.
func (s *Sizer) Update(err error) error {
	return s.UpdateWants(err, -1)
}

// UpdateWants updates the Sizer with the results of the action taken and a
// hint for a possible size for the next iteration.
func (s *Sizer) UpdateWants(err error, hint int) error {
	switch {
	case err == nil:
		s.ready = false
		return nil
	case s.again(err):
		if hint > s.size {
			s.size = hint
		} else {
			s.size *= 2
		}
		if s.size <= s.maxSize {
			return nil
		}
		fallthrough
	default:
		s.ready = false
		return err
	}
}
