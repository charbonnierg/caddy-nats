package parseutils

import "errors"

var (
	ErrEmptyBytesSize   = errors.New("cannot parse bytes size from empty value")
	ErrInvalidBytesSize = errors.New("cannot parse bytes size from invalid value")
	ErrInt32OutOfRange  = errors.New("cannot convert out of range integer to int32")
	ErrUInt32OutOfRange = errors.New("cannot convert out of range integer to uint32")
	ErrInt16OutOfRange  = errors.New("cannot convert out of range integer to int16")
	ErrUInt16OutOfRange = errors.New("cannot convert out of range integer to uint16")
	ErrInt8OutOfRange   = errors.New("cannot convert out of range integer to int8")
	ErrUInt8OutOfRange  = errors.New("cannot convert out of range integer to uint8")
)
