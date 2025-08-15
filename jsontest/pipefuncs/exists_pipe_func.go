package pipefuncs

import (
	"context"

	"github.com/mikeschinkel/scout-mcp/jsontest"
	"github.com/tidwall/gjson"
)

func init() {
	jsontest.RegisterPipeFunc(&ExistsPipeFunc{
		BasePipeFunc: jsontest.NewBasePipeFunc("exists()"),
	})
}

var _ jsontest.PipeFunc = (*ExistsPipeFunc)(nil)

type ExistsPipeFunc struct {
	jsontest.BasePipeFunc
}

func (e ExistsPipeFunc) Handle(ctx context.Context, ps *jsontest.PipeState) (err error) {
	if ps.Present {
		ps.Value = gjson.Parse("true")
	} else {
		ps.Value = gjson.Parse("false")
	}
	ps.Present = true // the boolean result "exists" itself exists

	return err
}
