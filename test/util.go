package test

import (
	"os"
)

func removeFile(fp string) (err error) {
	err = os.Remove(fp)
	if os.IsNotExist(err) {
		err = nil
		goto end
	}
end:
	return err
}
