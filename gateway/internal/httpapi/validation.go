package httpapi

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/ratdaddy/blockcloset/gateway/internal/testutil"
)

var ErrInvalidBucketName = errors.New("invalid bucket name")

type BucketNameValidator interface {
	ValidateBucketName(string) error
}

type DefaultBucketNameValidator struct{}

var bucketShapeRE = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9.-]*[a-z0-9])?$`)
var ipv4LikeRE = regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)

func (DefaultBucketNameValidator) ValidateBucketName(s string) error {
	if n := len(s); n < 3 || n > 63 {
		return ErrInvalidBucketName
	}

	if !bucketShapeRE.MatchString(s) {
		return ErrInvalidBucketName
	}

	if strings.Contains(s, "..") {
		return ErrInvalidBucketName
	}

	if ipv4LikeRE.MatchString(s) && isValidIPv4(s) {
		return ErrInvalidBucketName
	}

	switch {
	case strings.HasPrefix(s, "xn--"),
		strings.HasPrefix(s, "sthree-"),
		strings.HasPrefix(s, "amzn-s3-demo-"):
		return ErrInvalidBucketName
	}

	switch {
	case strings.HasSuffix(s, "-s3alias"),
		strings.HasSuffix(s, "--ol-s3"),
		strings.HasSuffix(s, ".mrap"),
		strings.HasSuffix(s, "--x-s3"),
		strings.HasSuffix(s, "--table-s3"):
		return ErrInvalidBucketName
	}

	return nil
}

func isValidIPv4(s string) bool {
	if strings.Count(s, ".") != 3 {
		return false
	}

	count := 0
	for part := range strings.SplitSeq(s, ".") {
		if part == "" {
			return false
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 || n > 255 {
			return false
		}
		count++
	}
	return count == 4
}
