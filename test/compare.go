package test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Compare compares two values and calls t.Error if they are different and printing the diff.
func Compare(t *testing.T, name string, expected, got interface{}, opts ...cmp.Option) {
	if diff := cmp.Diff(expected, got, opts...); diff != "" {
		_, fpath, line, _ := runtime.Caller(1)
		t.Errorf("Error on %s, line %d", filepath.Base(fpath), line)
		t.Errorf("%ss are not equal (-want +got):", name)
		t.Error(diff)
	}
}
