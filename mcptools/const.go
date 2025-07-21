package mcptools

import (
	"fmt"
)

var NULL struct{}

type RelativePosition string

const (
	BeforePosition RelativePosition = "before"
	AfterPosition  RelativePosition = "after"
)

func (rp RelativePosition) Validate() (err error) {
	switch rp {
	case BeforePosition:
	case AfterPosition:
	default:
		err = fmt.Errorf("position must be '%s' or '%s', got '%s'",
			BeforePosition,
			AfterPosition,
			rp,
		)
	}
	return err
}
