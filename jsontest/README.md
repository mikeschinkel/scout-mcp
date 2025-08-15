# jsontest

Declarative JSON assertions for Go tests with modular pipe functions

## Overview

`jsontest` is a JSON testing framework for Go that enables declarative assertions against complex JSON responses. It's designed for JSON-RPC validation, REST API testing, and any scenario requiring deep JSON structure validation.

The package features a modular architecture with extensible pipe functions and supports sophisticated path-based assertions without repetitive unmarshalling code.

## ✨ Features

* **Declarative Testing**: Define assertions using `map[string]any` with GJSON paths as keys
* **Deep Path Support**: Navigate JSON using [tidwall/gjson](https://github.com/tidwall/gjson) syntax
* **Modular Pipe Functions**: Extensible pipe function architecture in separate `pipefuncs` package
* **Array Operations**: `[].subpath` syntax for collecting fields from array elements
* **Order Flexibility**: Order-sensitive and order-insensitive array comparisons
* **Type Markers**: `NotNull{}`, `NotEmpty{}` for common validation patterns
* **Smart Type Coercion**: Automatic type conversion for accurate comparisons
* **Nested Arrays**: Flattened collection support for complex nested structures

## 📦 Installation & Setup

Since this is part of the Scout-MCP project, import it directly:

```go
import (
    "github.com/mikeschinkel/scout-mcp/jsontest"
    _ "github.com/mikeschinkel/scout-mcp/jsontest/pipefuncs" // For pipe functions
)
```

**Important**: Always import `pipefuncs` package as a side-effect to register pipe functions.

## 🚀 Quick Start

```go
package myapi_test

import (
	"testing"
	"github.com/mikeschinkel/scout-mcp/jsontest"
	_ "github.com/mikeschinkel/scout-mcp/jsontest/pipefuncs" // Required for pipe functions
)

func TestJSONResponse(t *testing.T) {
	body := []byte(`{
		"jsonrpc": "2.0",
		"result": {
			"content": [
				{"type": "text", "text": "Hello"},
				{"type": "text", "text": "World"}
			]
		}
	}`)

	err := jsontest.TestJSON(body, map[string]any{
		"jsonrpc":          "2.0",
		"result.content.#": 2,
		
		// Type markers
		"result.content.0": jsontest.NotNull{},
		
		// Pipe functions
		"result.content.0.text|notEmpty()": true,
		"result.content|len()":             2,
		
		// Array collection with order-insensitive comparison
		"result.content.[].type": jsontest.AnyOrder("text", "text"),
	})
	
	if err != nil {
		t.Error(err)
	}
}
```

---

## 📚 Path Syntax

We use [GJSON path syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md) with some `jsontest`-specific extensions.

### **Standard GJSON paths**

```go
"result.content.0.type": "object"
```

### **Length operator**

```go
"result.content.#": 2  // array length == 2
```

### **Array operator**

Collect a property from every element in an array:

```go
"result.content.[].type": []string{"object", "object"}
```

---

## 🔄 Order-Sensitive vs Order-Insensitive

By default, array and collected-slice comparisons are **order-sensitive**.

For order-insensitive comparisons, wrap expected values with `AnyOrder` or `AnyOrderEq`:

```go
// Array, order-insensitive
"result.content.[].type": jsontest.AnyOrder("object", "object")
```

---

## 🔧 Modular Pipe Functions

The `jsontest/pipefuncs` package provides a modular pipe function architecture. Pipe functions transform or evaluate values before comparison using the syntax: `"path|func()"`.

**Available Pipe Functions:**

| Function     | Package Location                     | Returns | Description                                         |
|--------------|-------------------------------------|---------|-----------------------------------------------------|
| `exists()`   | `pipefuncs/exists_pipe_func.go`     | bool    | True if the path exists in JSON                     |
| `notNull()`  | `pipefuncs/not_null_pipe_func.go`   | bool    | True if value exists and is not null               |
| `notEmpty()` | `pipefuncs/not_empty_pipe_func.go`  | bool    | True if value is non-empty (arrays, objects, etc.) |
| `len()`      | `pipefuncs/len_pipe_func.go`        | int     | Length of arrays, objects, or strings               |
| `json()`     | `pipefuncs/json_pipe_func.go`       | parsed  | Parse JSON strings and access nested properties    |

**Examples:**

```go
// Boolean assertions
"result.content.0.text|notEmpty()": true,
"result.missing|exists()": false,
"result.data|notNull()": true,

// Length assertions
"result.content|len()": 2,
"result.content.0.tags|len()": 3,

// JSON parsing and nested access
"result.jsonStr|json()|nested.field": "value",
```

**Adding Custom Pipe Functions:**

```go
// In your pipefuncs package
func init() {
    jsontest.RegisterPipeFunc(&CustomPipeFunc{
        BasePipeFunc: jsontest.NewBasePipeFunc("custom()"),
    })
}

type CustomPipeFunc struct {
    jsontest.BasePipeFunc
}

func (c CustomPipeFunc) Handle(ctx context.Context, ps *jsontest.PipeState) error {
    // Your custom logic here
    ps.Value = gjson.Parse("transformed_value")
    ps.Present = true
    return nil
}
```

---

## 🏷 Markers

Markers are special expected values that change assertion behavior:

| Marker       | Behavior                                                                                    |
| ------------ | ------------------------------------------------------------------------------------------- |
| `NotNull{}`  | Passes if value exists and is not JSON null                                                 |
| `NotEmpty{}` | Passes if value is: non-empty string, non-empty array, non-empty object, number, or boolean |

Example:

```go
"result.content.0.object": jsontest.NotNull{}
"result.content.0.object.message": jsontest.NotEmpty{}
```

Markers provide some type safety when compared to pipe functions, but at the expense of requiring more verbose boilerplate. 

---

## 🧪 Type Coercion

When comparing scalars, `jsontest` coerces the value to the **type of the expected value** so you can write:

```go
"result.count": 5        // coerced to int
"result.price": 19.99    // coerced to float64
"result.enabled": true   // coerced to bool
```

---

## ⚠️ Limitations

* **Pipe Function Registration**: Must import `pipefuncs` package as side-effect or call `pipefuncs.Initialize()`
* **Zero Arguments**: Pipe functions are currently zero-argument only with no plans to add arguments.
* **Scalar Subpaths**: Cannot apply subpaths after scalar-returning functions (e.g., `len()|field` is invalid)
* **Error Handling**: Pipe function errors stop processing; no fallback mechanisms

## 💡 Best Practices

* **Always Import Pipefuncs**: Use `_ "github.com/mikeschinkel/scout-mcp/jsontest/pipefuncs"` import
* **Use Type Markers**: Prefer `NotNull{}` and `NotEmpty{}` for common validation patterns
* **Order-Insensitive Arrays**: Use `AnyOrder()` when array order doesn't matter
* **Pipe Function Naming**: End custom pipe function names with `()` (enforced by framework)

## 🏗️ Architecture Notes

This package is part of the Scout-MCP project and follows its coding conventions:
- **Clear Path Style**: Single return points with `goto end` pattern
- **Modular Design**: Pipe functions in separate package for extensibility  
- **Session Integration**: Used extensively in MCP server testing framework
- **JSON-RPC Focus**: Designed specifically for JSON-RPC protocol validation

## 📜 License

This package is part of the Scout-MCP project. See the main project LICENSE file for details.
