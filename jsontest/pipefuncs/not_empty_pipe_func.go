package pipefuncs

import (
	"context"
	"regexp"
	"strings"

	"github.com/mikeschinkel/scout-mcp/jsontest"
	"github.com/tidwall/gjson"
)

func init() {
	jsontest.RegisterPipeFunc(&NotEmptyPipeFunc{
		BasePipeFunc: jsontest.NewBasePipeFunc("notEmpty()"),
	})
}

var _ jsontest.PipeFunc = (*NotEmptyPipeFunc)(nil)

type NotEmptyPipeFunc struct {
	jsontest.BasePipeFunc
}

var notEmptyBoolOrNumRegexp = regexp.MustCompile(
	`^(?:true|false|-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?)$`,
)

func (n NotEmptyPipeFunc) isNonEmpty(v gjson.Result) (nonEmpty bool) {
	switch {
	case !v.Exists():
		goto end
	case v.IsArray():
		nonEmpty = len(v.Array()) > 0
	case jsontest.IsJSONObject(v):
		nonEmpty = len(v.Map()) > 0 // {} -> false
	default:
		s := v.String()
		if s != "" {
			nonEmpty = true
			goto end
		}
		// numbers/bools are considered non-empty
		nonEmpty = notEmptyBoolOrNumRegexp.MatchString(strings.TrimSpace(v.Raw))
	}
end:
	return nonEmpty
}

func (n NotEmptyPipeFunc) Handle(ctx context.Context, ps *jsontest.PipeState) (err error) {
	if n.isNonEmpty(ps.Value) {
		ps.Value = gjson.Parse("true")
	} else {
		ps.Value = gjson.Parse("false")
	}
	ps.Present = true

	return err
}
