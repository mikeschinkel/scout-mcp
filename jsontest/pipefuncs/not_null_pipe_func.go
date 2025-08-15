package pipefuncs

import (
	"context"
	"strings"

	"github.com/mikeschinkel/scout-mcp/jsontest"
	"github.com/tidwall/gjson"
)

func init() {
	jsontest.RegisterPipeFunc(&NotNullPipeFunc{
		BasePipeFunc: jsontest.NewBasePipeFunc("notNull()"),
	})
}

var _ jsontest.PipeFunc = (*NotNullPipeFunc)(nil)

type NotNullPipeFunc struct {
	jsontest.BasePipeFunc
}

func (n NotNullPipeFunc) Handle(ctx context.Context, ps *jsontest.PipeState) (err error) {
	if ps.Present && strings.TrimSpace(ps.Value.Raw) != "null" {
		ps.Value = gjson.Parse("true")
	} else {
		ps.Value = gjson.Parse("false")
	}
	ps.Present = true

	return err
}
