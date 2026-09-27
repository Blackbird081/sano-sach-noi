package setup

import (
	"errors"
	"testing"
)

func TestCheckOS(t *testing.T) {
	cases := []struct {
		goos, goarch, ver string
		bad               bool
	}{
		{"darwin", "amd64", "12.7.6", true},
		{"darwin", "amd64", "13.0", false},
		{"darwin", "amd64", "26.0.1", false},
		{"darwin", "arm64", "13.6", true},
		{"darwin", "arm64", "14.0", false},
		{"darwin", "arm64", "", false},
		{"darwin", "arm64", "lạ", false},
		{"windows", "amd64", "10.0", false},
		{"linux", "amd64", "", false},
	}
	for _, c := range cases {
		err := checkOS(c.goos, c.goarch, c.ver)
		if got := errors.Is(err, ErrOSTooOld); got != c.bad {
			t.Errorf("checkOS(%s/%s %q) = %v, muốn chặn=%v", c.goos, c.goarch, c.ver, err, c.bad)
		}
	}
}
