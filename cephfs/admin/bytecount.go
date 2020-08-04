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
