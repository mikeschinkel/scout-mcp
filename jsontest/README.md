Got it — I’ll write this up as if we’re about to publish `jsontest` as a public Go package on GitHub, with a full README that explains:

* What the package does
* How it works conceptually
* All supported features
* Code examples for each feature
* Best practices and limitations

---

## 📄 **README.md** – *for `jsontest` package*

# jsontest

Declarative, powerful JSON assertions for Go tests

---

### Overview

`jsontest` is a small test helper for Go that lets you assert against complex JSON responses **declaratively**, without writing repetitive unmarshalling and `if` statements.

It’s ideal for REST API testing, JSON-RPC validation, and any situation where you need to check deeply nested JSON properties, enforce schema-like constraints, or compare array contents in a concise and readable way.

---

## ✨ Features

* **Declarative**: Define assertions in a `map[string]any` with GJSON paths as keys.
* **Deep path support**: Navigate JSON easily using [tidwall/gjson](https://github.com/tidwall/gjson) syntax.
* **Length checks**: Use `#` suffix in paths.
* **Map-over-array**: `[].subpath` to collect fields from every element.
* **Order-sensitive / order-insensitive** comparisons.
* **Markers**: `NotNull{}`, `NotEmpty{}` for common non-nil/empty checks.
* **Pipe functions**: Transform/assert values inline with syntax like
  `path|notEmpty()` or `path|len()`.
* **Type coercion**: Numbers, strings, and booleans are coerced to the expected type before comparing.

---

## 📦 Installation

```bash
go get github.com/yourusername/jsontest
```

---

## 🚀 Quick Start

```go
package myapi_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yourusername/jsontest"
)

func TestAPIResponse(t *testing.T) {
	body := []byte(`{
		"jsonrpc": "2.0",
		"result": {
			"content": [
				{ "type": "object", "object": { "message": "hi", "tags": ["a","b"] } },
				{ "type": "object", "object": { "message": "bye", "tags": [] } }
			]
		}
	}`)

	jsontest.TestJSON(t, body, map[string]any{
		"jsonrpc":          "2.0",
		"result.content.#": 2,

		// Enforce non-null and non-empty
		"result.content.0.object": jsontest.NotNull{},
		"result.content.0.object": jsontest.NotEmpty{},

		// Pipe functions
		"result.content.0.object.message|notEmpty()": true,
		"result.content.0.object|len()":              2,
		"result.content.0.object.tags|len()":         2,
		"result.content.1.object.tags|notEmpty()":    false,

		// Order-insensitive array comparison
		"result.content.[].type": jsontest.AnyOrder("object", "object"),
	})
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

### **Map-over-array**

Collect a property from every element in an array:

```go
"result.content.[].type": []string{"object", "object"}
```

---

## 🔄 Order-Sensitive vs Order-Insensitive

By default, array and collected-slice comparisons are **order-sensitive**.

For order-insensitive comparisons, wrap expected values with `AnyOrder` or `AnyOrderEq`:

```go
// Map-over-array, order-insensitive
"result.content.[].type": jsontest.AnyOrder("object", "object")

// Array directly at path, order-insensitive
"result.tags": jsontest.AnyOrderEq("blue", "green", "red")
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

---

## 🔧 Pipe Functions

Pipes allow you to **transform or evaluate** a value before comparison.
Syntax: `"path|func()"` or `"path|func()|subpath"` (functions before subpaths are allowed unless the function returns a scalar).

Available pipe functions:

| Function     | Returns | Description                                       |
| ------------ | ------- | ------------------------------------------------- |
| `notNull()`  | bool    | True if value exists and is not null              |
| `notEmpty()` | bool    | True if value is non-empty (same rules as marker) |
| `len()`      | int     | Length of array, object, or string (0 for others) |
| `exists()`   | bool    | True if the path exists                           |

Examples:

```go
// Assert boolean result of function
"result.content.0.object.message|notEmpty()": true

// Assert computed length
"result.content.0.object.tags|len()": 2

// Combine pipes and subpaths
"result.content.0.object|len()": 2
```

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

* Pipe functions are currently **zero-argument** only.
* Cannot apply a subpath after a scalar-returning function (e.g., `len()` on an array, then `.foo`).
* For complex validation (e.g., ranges, regex), you can add custom markers or pipe functions.

---

## 💡 Best Practices

* Use markers for `nil`/empty checks to keep your tests readable.
* Use `AnyOrder` for array assertions where order isn’t guaranteed.
* Use pipe functions for **computed assertions** (e.g., lengths) instead of manually counting in test code.
* For large API responses, group checks logically in multiple `TestJSON` calls for clarity.

---

## 📜 License

MIT © Your Name

---

If you want, I can now also write a **`doc.go`** so that if this is published as a Go module, `godoc` shows the same usage and examples directly in Go tooling. That way the package would be self-documenting for `go doc`.

Do you want me to create that `doc.go` too?
