package validation_test

import (
	"strings"
	"testing"

	"github.com/ratdaddy/blockcloset/pkg/validation"
)

// Bucket name validation tests based on S3 naming rules:
//	https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html

func TestDefaultBucketNameValidator_GeneralPurposeRules(t *testing.T) {
	t.Parallel()
	v := validation.DefaultBucketNameValidator{}

	type tc struct {
		name   string
		input  string
		wantOK bool
	}

	long63 := strings.Repeat("a", 63)
	long64 := strings.Repeat("a", 64)

	cases := []tc{
		// Valid (per spec)
		{"min length 3", "abc", true},
		{"simple letters and digits", "a1b2c3", true},
		{"hyphens and dots allowed", "a-b.c-d", true},
		{"no adjacent dots but single dots ok", "a.b.c", true},
		{"must begin and end alnum (valid)", "a-.-a", true},
		{"max length 63", long63, true},

		// Invalid: length
		{"too short (2)", "ab", false},
		{"too long (64)", long64, false},

		// Invalid: characters
		{"uppercase not allowed", "Abc", false},
		{"underscore not allowed", "a_b", false},

		// Invalid: must begin/end alnum
		{"starts with hyphen", "-abc", false},
		{"ends with hyphen", "abc-", false},
		{"starts with dot", ".abc", false},
		{"ends with dot", "abc.", false},

		// Invalid: adjacent periods
		{"adjacent dots", "a..b", false},

		// Invalid: IPv4 address form
		{"looks like IPv4 address", "192.168.5.4", false},

		// Invalid: reserved prefixes
		{"reserved prefix xn--", "xn--bucket", false},
		{"reserved prefix sthree-", "sthree-bucket", false},
		{"reserved prefix amzn-s3-demo-", "amzn-s3-demo-bucket", false},

		// Invalid: reserved suffixes
		{"reserved suffix -s3alias", "my-logs-s3alias", false},
		{"reserved suffix --ol-s3", "images--ol-s3", false},
		{"reserved suffix .mrap", "data.mrap", false},
		{"reserved suffix --x-s3", "foo--x-s3", false},
		{"reserved suffix --table-s3", "bar--table-s3", false},
	}

	for _, c := range cases {
		t.Run(c.name+"="+c.input, func(t *testing.T) {
			t.Parallel()
			err := v.ValidateBucketName(c.input)
			if c.wantOK && err != nil {
				t.Fatalf("want OK, got error: %v", err)
			}
			if !c.wantOK {
				if err == nil {
					t.Fatalf("want error %v, got nil", validation.ErrInvalidBucketName)
				}
				if err != validation.ErrInvalidBucketName {
					t.Fatalf("want %v, got %v", validation.ErrInvalidBucketName, err)
				}
			}
		})
	}
}
