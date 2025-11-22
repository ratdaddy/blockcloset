package validation

import (
	"errors"
	"unicode/utf8"
)

var ErrInvalidKeyName = errors.New("invalid key name")

type KeyValidator interface {
	ValidateKey(string) error
}

type DefaultKeyValidator struct{}

func (DefaultKeyValidator) ValidateKey(s string) error {
	if len(s) < 1 || len(s) > 1024 {
		return ErrInvalidKeyName
	}

	if !utf8.ValidString(s) {
		return ErrInvalidKeyName
	}

	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 0x20 && b != '\t' && b != '\n' {
			return ErrInvalidKeyName
		}
	}

	return nil
}
