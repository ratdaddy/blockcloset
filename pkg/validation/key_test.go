package validation_test

import (
	"strings"
	"testing"

	"github.com/ratdaddy/blockcloset/pkg/validation"
)

func TestDefaultKeyValidator(t *testing.T) {
	t.Parallel()
	v := validation.DefaultKeyValidator{}

	type tc struct {
		name   string
		input  string
		wantOK bool
	}

	long1024 := strings.Repeat("a", 1024)
	long1025 := strings.Repeat("a", 1025)

	cases := []tc{
		// Length validation
		{"simple key", "myfile.txt", true},
		{"empty key", "", false},
		{"max length 1024", long1024, true},
		{"too long 1025", long1025, false},

		// Valid characters (S3 allows almost anything)
		{"with space", "file name.txt", true},
		{"with unicode", "文件.txt", true},
		{"nested path", "a/b/c/file.txt", true},
		{"with tab", "file\tname.txt", true},
		{"with newline", "file\nname.txt", true},

		// Control characters (reject 0x00-0x1F except tab and newline)
		{"null byte", "file\x00name", false},
		{"control 0x1F", "file\x1Fname", false},

		// Invalid UTF-8
		{"invalid utf-8", "file\xFF\xFEname", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := v.ValidateKey(c.input)
			if c.wantOK && err != nil {
				t.Fatalf("want OK, got error: %v", err)
			}
			if !c.wantOK {
				if err == nil {
					t.Fatalf("want error %v, got nil", validation.ErrInvalidKeyName)
				}
				if err != validation.ErrInvalidKeyName {
					t.Fatalf("want %v, got %v", validation.ErrInvalidKeyName, err)
				}
			}
		})
	}
}
