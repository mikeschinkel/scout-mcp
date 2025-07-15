package test

import (
	"io"
	"log"
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

func mustClose(c io.Closer) {
	err := c.Close()
	logOnError(err)
}
func logOnError(err error) {
	if err != nil {
		log.Println(err)
	}
}
