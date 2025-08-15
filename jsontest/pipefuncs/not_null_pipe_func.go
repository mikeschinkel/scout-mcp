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

// NotNullPipeFunc implements the notNull() pipe function that checks if a value is not null.
type NotNullPipeFunc struct {
	jsontest.BasePipeFunc
}

// Handle checks if the current value exists and is not null, returning true/false accordingly.
func (n NotNullPipeFunc) Handle(ctx context.Context, ps *jsontest.PipeState) (err error) {
	if ps.Present && strings.TrimSpace(ps.Value.Raw) != "null" {
		ps.Value = gjson.Parse("true")
	} else {
		ps.Value = gjson.Parse("false")
	}
	ps.Present = true

	return err
}
