package fsfix

import (
	"log"
)

func must(err error) {
	if err != nil {
		log.Print(err.Error())
	}
}
