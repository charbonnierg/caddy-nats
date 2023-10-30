//go:build arm && 386

package parseutils

// UInt32 converts int to uint32 in a safe way.
// You get error when the value is out of the 32-bit unsigned range (0 through 4294967295).
func UInt32(i int) (uint32, error) {
	if i < 0 {
		return 0, ErrUInt32OutOfRange
	}
	return uint32(i), nil
}
