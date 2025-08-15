// test_json_test.go
package jsontest_test

import (
	"testing"

	"github.com/mikeschinkel/scout-mcp/jsontest"

	_ "github.com/mikeschinkel/scout-mcp/jsontest/pipefuncs"
)
func TestJSONTest(t *testing.T) {
	body := []byte(`{
	  "foo": {
	    "bar": {
	      "baz": "qux",
	      "nullish": null,
	      "emptyStr": "",
	      "obj": {"a": 1, "b": 2},
	      "arr": [1, 2, 3],
	      "jsonStr": "{\"bar\":{\"baz\":\"fromjson\",\"nums\":[1,2],\"inner\":[{\"x\":\"a\"},{\"x\":\"b\"}]},\"n\":null,\"s\":\"\",\"o\":{}}"
	    }
	  },
	  "items": [
	    { "bar": [ {"baz":"A"}, {"baz":"B"} ], "tags": ["x","y"] },
	    { "bar": [ {"baz":"C"} ],             "tags": [] }
	  ],
	  "top": [
	    { "mid": [ {"leaf":1}, {"leaf":2} ] },
	    { "mid": [ {"leaf":3} ] }
	  ],
	  "maybe": {}
	}`)

	type tc struct {
		name       string
		checks     map[string]any
		shouldFail bool
	}

	cases := []tc{
		// ===== Positive: multiple nested properties =====
		{
			name: "multi-props",
			checks: map[string]any{
				"foo.bar.baz": "qux",
				"foo.bar.obj": jsontest.NotNull{},
				"foo.bar.obj|len()": 2,
				"foo.bar.arr": []int{1, 2, 3},
				"foo.bar.arr|len()": 3,
			},
		},

		// ===== Positive: arrays (single []), length operator (#) =====
		{
			name: "array-and-length",
			checks: map[string]any{
				"items.[].bar.0.baz": []string{"A", "C"}, // first bar.baz for each item
				"items.#":            2,                  // number of items
				"items.0.tags.#":     2,
				"items.1.tags.#":     0,
			},
		},

		// ===== Positive: nested [] chains (multiple arrays) =====
		{
			name: "nested-arrays-collect",
			checks: map[string]any{
				"items.[].bar.[].baz": jsontest.AnyOrder("A", "B", "C"),
				"top.[].mid.[].leaf":  jsontest.AnyOrder(1, 2, 3),
			},
		},

		// ===== Positive: pipe functions (scalar + collections) =====
		{
			name: "pipes-basic",
			checks: map[string]any{
				"foo.bar.emptyStr|notEmpty()": false,
				"foo.bar.nullish|notNull()":   false,
				"foo.bar.nullish|exists()":    true,  // path exists but is null
				"foo.missing|exists()":        false, // path does not exist
			},
		},

		// ===== Positive: json() + subpaths + further pipes =====
		{
			name: "json-pipe-and-subpaths",
			checks: map[string]any{
				"foo.bar.jsonStr|json()|bar.baz":        "fromjson",
				"foo.bar.jsonStr|json()|bar.nums|len()": 2,
				"foo.bar.jsonStr|json()|bar.inner.[].x": jsontest.AnyOrder("a", "b"),
				"foo.bar.jsonStr|json()|n|notNull()":    false,
				"foo.bar.jsonStr|json()|s|notEmpty()":   false,
				"foo.bar.jsonStr|json()|o|len()":        0,
				"foo.bar.jsonStr|json()|o|notEmpty()":   false,
			},
		},

		// ===== Positive: combinations (props + nested arrays + pipes) =====
		{
			name: "combo-all",
			checks: map[string]any{
				"foo.bar.baz":                        "qux",
				"items.[].bar.[].baz":                jsontest.AnyOrder("C", "B", "A"),
				"foo.bar.jsonStr|json()|bar.nums.#":  2,
				"top|len()":                          2,
				"top.[].mid.[].leaf":                 jsontest.AnyOrder(3, 2, 1),
			},
		},

		// ===========================
		// ===== Negative tests ======
		// ===========================

		// Invalid: scalar-returning pipe followed by subpath (disallowed)
		{
			name: "neg-scalar-then-subpath",
			checks: map[string]any{
				"foo.bar.arr|len()|oops": 0,
			},
			shouldFail: true,
		},

		// Invalid: unknown pipe function
		{
			name: "neg-unknown-pipe",
			checks: map[string]any{
				"foo.bar.arr|bogus()": true,
			},
			shouldFail: true,
		},

		// Invalid: missing path
		{
			name: "neg-missing-path",
			checks: map[string]any{
				"no.such.path": "value",
			},
			shouldFail: true,
		},

		// Invalid: wrong scalar expectation
		{
			name: "neg-wrong-scalar",
			checks: map[string]any{
				"foo.bar.baz": "NOPE",
			},
			shouldFail: true,
		},

		// Invalid: wrong collected slice (order-sensitive by default)
		{
			name: "neg-wrong-collected-order",
			checks: map[string]any{
				"items.[].bar.0.baz": []string{"C", "A"}, // actual is {"A","C"}
			},
			shouldFail: true,
		},

		// Invalid: wrong AnyOrder contents
		{
			name: "neg-wrong-AnyOrder-contents",
			checks: map[string]any{
				"items.[].bar.[].baz": jsontest.AnyOrder("A", "B", "X"),
			},
			shouldFail: true,
		},

		// Invalid: pipe boolean mismatch
		{
			name: "neg-pipe-boolean-mismatch",
			checks: map[string]any{
				"foo.bar.emptyStr|notEmpty()": true, // actually false
			},
			shouldFail: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := jsontest.TestJSON(body, tc.checks)
			if tc.shouldFail {
				if err == nil {
					t.Fatalf("expected failure but got success: %s", tc.name)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error in %s: %v", tc.name, err)
				}
			}
		})
	}
}
