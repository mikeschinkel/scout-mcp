package test

import (
	"io"
	"log"
)

// must terminates the program if the provided error is not nil.
func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// mustClose closes the provided io.Closer and logs any error.
func mustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Println(err.Error())
	}
}
