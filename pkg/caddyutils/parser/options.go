// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"errors"
)

var (
	// ErrInvalidOption is returned when an option is not valid for
	// the provided parser.
	ErrInvalidOption      = errors.New("invalid option")
	ErrInvalidDefaultType = errors.New("invalid default value type")
)

// Option is a function that can be used to configure a parser.
type Option func(any) error

// allowMatch is an interface that can be implemented by a parser to
// indicate that it supports the MatchOr method.
type allowMatch interface {
	AddMatch(string, ...string) error
}

type allowInline interface {
	SetInline(bool)
}

// allowInplace is an interface that can be implemented by a parser to
// indicate that it supported the Inplace option.
type allowInplace interface {
	SetInplace(bool)
}

// allowSeparator is an interface that can be implemented by a parser to
// indicate that it supports the separator option.
type allowSeparator interface {
	SetSeparator(string, ...string) error
}

// allowEmpty is an interface that can be implemented by a parser to
// indicate that it supports the allowEmpty option.
type allowEmpty interface {
	SetAllowEmpty(bool)
}

type allowEmptyValues interface {
	SetAllowEmptyValues(bool)
}

// allowReverse is an interface that can be implemented by a parser to
// indicate that it supports the reverse option.
type allowReverse interface {
	SetReversed(bool)
}

// allowDefault is an interface that can be implemented by a parser to
// indicate that it supports the default option.
type allowDefault interface {
	SetDefaultValue(interface{}) error
}

// AllowEmpty is an option that can be used to not return an error
// dispenser is empty.
func AllowEmpty() Option {
	return func(p any) error {
		if p, ok := p.(allowEmpty); ok {
			p.SetAllowEmpty(true)
			return nil
		} else {
			return ErrInvalidOption
		}
	}
}

// AllowEmptyValues is an option that can be used to allow empty
// values for a parser. It is different from AllowEmpty() because it
// allows empty values for a key or an index rather than allowing empty values
// for the whole dispenser.
func AllowEmptyValues() Option {
	return func(p any) error {
		if p, ok := p.(allowEmptyValues); ok {
			p.SetAllowEmptyValues(true)
			return nil
		} else {
			return ErrInvalidOption
		}
	}
}

// ErrorIfEmpty is an option that can be used to disallow empty values
// for a parser.
func ErrorIfEmpty() Option {
	return func(p any) error {
		if p, ok := p.(allowEmpty); ok {
			p.SetAllowEmpty(false)
			return nil
		} else {
			return ErrInvalidOption
		}
	}
}

// Reverse is an option that can be used to reverse the value of a
// parser.
func Reverse() Option {
	return func(p any) error {
		if p, ok := p.(allowReverse); ok {
			p.SetReversed(true)
			return nil
		} else {
			return ErrInvalidOption
		}
	}
}

// Default is an option that can be used to set the default value of a
// parser.
func Default(value interface{}) Option {
	return func(p any) error {
		if p, ok := p.(allowDefault); ok {
			return p.SetDefaultValue(value)
		} else {
			return ErrInvalidOption
		}
	}
}

// Separator is an option that can be used to set the separator of a
// parser to the provided value.
func Separator(sep string, seps ...string) Option {
	return func(p any) error {
		if p, ok := p.(allowSeparator); ok {
			return p.SetSeparator(sep, seps...)
		} else {
			return ErrInvalidOption
		}
	}
}

// Match is an option that be used to set the match value of a parser.
func Match(value string, others ...string) Option {
	return func(p any) error {
		if p, ok := p.(allowMatch); ok {
			if err := p.AddMatch(value, others...); err != nil {
				return err
			}
			return nil
		} else {
			return ErrInvalidOption
		}
	}
}

// Inplace is an option that can be used to start parsing from the
// current position of the dispenser, rather than starting from the
// next argument (which is the default behavior).
func Inplace() Option {
	return func(p any) error {
		if p, ok := p.(allowInplace); ok {
			p.SetInplace(true)
			return nil
		} else {
			return ErrInvalidOption
		}
	}
}

// Inline is an option that can be used to parse the block inline
// rather than as a nested block. This option apply to parsers that
// parse nested blocks by default, such as ParseStringMap.
func Inline(opts ...Option) Option {
	return func(p any) error {
		for _, opt := range opts {
			if err := opt(p); err != nil {
				return err
			}
		}
		if p, ok := p.(allowInline); ok {
			p.SetInline(true)
			return nil
		} else {
			return ErrInvalidOption
		}
	}
}
