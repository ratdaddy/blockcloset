package httpapi

import (
	"errors"
)

var ErrInvalidBucketName = errors.New("invalid bucket name")

type BucketNameValidator interface {
	ValidateBucketName(string) error
}

type DefaultBucketNameValidator struct{}

func (DefaultBucketNameValidator) ValidateBucketName(s string) error {
	return nil
}
