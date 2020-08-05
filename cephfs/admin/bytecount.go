// +build !luminous,!mimic

package admin

// ByteCount represents the size of a volume in bytes.
type ByteCount uint64

// SI byte size constants. keep these private for now.
const (
	kibiByte ByteCount = 1024
	mebiByte           = 1024 * kibiByte
	gibiByte           = 1024 * mebiByte
	tebiByte           = 1024 * gibiByte
)

// newSizeValue returns a size value as a string, as needed by the subvolume
// resize command json.
func (bc ByteCount) newSizeValue() string {
	return uint64String(uint64(bc))
}

// NewSize interface values can be used to change the size of a volume.
type NewSize interface {
	newSizeValue() string
}

// specialSize is a custom non-numeric new size value.
type specialSize string

// newSizeValue for a specialSize returns the literal string.
func (s specialSize) newSizeValue() string {
	return string(s)
}

// Infinite is a special NewSize value that can be used to clear size limits on
// a subvolume.
const Infinite = specialSize("infinite")
