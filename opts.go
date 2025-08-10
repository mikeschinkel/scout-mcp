package scout

import (
	"io"
)

type Opts struct {
	OnlyMode        bool
	AdditionalPaths []string
	Stdin           io.Reader
	Stdout          io.Writer
}
