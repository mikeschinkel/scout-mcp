package pipefuncs

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/scout-mcp/jsontest"
	"github.com/tidwall/gjson"
)

func init() {
	jsontest.RegisterPipeFunc(&LenPipeFunc{
		BasePipeFunc: jsontest.NewBasePipeFunc("len()"),
	})
}

var _ jsontest.PipeFunc = (*LenPipeFunc)(nil)

// LenPipeFunc implements the len() pipe function that returns the length of arrays, objects, or strings.
type LenPipeFunc struct {
	jsontest.BasePipeFunc
}

// Handle calculates the length of the current value and returns it as a number.
func (l LenPipeFunc) Handle(ctx context.Context, ps *jsontest.PipeState) (err error) {
	var n int
	switch {
	case ps.Value.IsArray():
		n = len(ps.Value.Array())
	case jsontest.IsJSONObject(ps.Value): // <-- treat {} as object with 0 keys
		n = len(ps.Value.Map())
	default:
		n = len(ps.Value.String())
	}
	ps.Value = gjson.Parse(fmt.Sprintf("%d", n))
	ps.Present = true

	return err
}
