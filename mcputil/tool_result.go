package mcputil

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
)

// callResult wraps a ToolResult with any associated error.
type callResult struct {
	ToolResult       // The tool execution result
	Error      error // Any error that occurred during execution
}

// CallResult creates a callResult wrapping a ToolResult and error.
// If err is not an InternalError, it converts the error to a ToolResult.
// This function handles the distinction between internal system errors and application errors.
//
//goland:noinspection GoExportedFuncWithUnexportedType
func CallResult(tr ToolResult, err error) callResult {
	if err != nil {
		var intErr *InternalError
		if errors.As(err, &intErr) {
			// If it is an internal error, pass it on
			goto end
		}
		// If it is not an internal error, convert to an error result
		tr = NewToolResultError(err)
	}
end:
	return callResult{
		ToolResult: tr,
		Error:      err,
	}
}

// GetToolResult extracts and unmarshals a typed result from a callResult.
// It handles payload unmarshaling for complex result types and provides proper error handling.
// This generic function supports type-safe result extraction with payload deserialization.
func GetToolResult[R any](cr callResult, errMsg string) (r *R, err error) {
	var b []byte
	var ute *json.UnmarshalTypeError
	var t reflect.Type
	var v reflect.Value
	var dec *json.Decoder
	var c PayloadCarrier
	var tn string
	// If there's an error, return it without trying to parse JSON
	if cr.Error != nil {
		err = cr.Error
		goto end
	}

	// Only assert ToolResult is not nil when there's no error
	if cr.ToolResult == nil {
		err = errors.New(errMsg)
		goto end
	}
	r = new(R)
	b = []byte(cr.ToolResult.Value())
	err = json.Unmarshal(b, &r)
	if !errors.As(err, &ute) {
		goto end
	}
	if ute.Value != "object" {
		goto end
	}
	if ute.Type.Kind() != reflect.Interface {
		goto end
	}
	if ute.Field != "payload" {
		goto end
	}
	t, tn, c = GetPayloadInfo(r)
	if tn != "" && t == nil {
		err = ErrNoPayloadType
		goto end
	}
	v = reflect.New(t)
	dec = json.NewDecoder(bytes.NewReader(b[ute.Offset-1:]))
	err = dec.Decode(v.Interface())
	if err != nil {
		goto end
	}
	c.SetPayload(v.Elem().Interface().(Payload))
end:
	return r, err
}
