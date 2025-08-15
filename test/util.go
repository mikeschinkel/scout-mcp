package test

import (
	"io"
	"log"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Println(err.Error())
	}
}
