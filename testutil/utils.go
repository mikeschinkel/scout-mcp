package testutil

import (
	"io/fs"
	"log"
	"os"
	"testing"
)

// MaybeRemove wraps calls to os.Remove and logs errors that are other than not-exists
func MaybeRemove(t *testing.T, fp string) {
	var ok bool

	t.Helper()
	err := os.RemoveAll(fp)
	if err == nil {
		goto end
	}
	_, ok = err.(*fs.PathError)
	if ok {
		goto end
	}

	t.Error(err)
end:
}

func must(err error) {
	if err != nil {
		log.Print(err.Error())
	}
}
