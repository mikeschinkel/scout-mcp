package mcputil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockRequest implements ToolRequest for testing
type mockRequest struct {
	params map[string]any
}

func (m *mockRequest) CallToolRequest() CallToolRequest {
	return CallToolRequest{
		Params: CallToolParams{
			Arguments: m.params,
		},
	}
}

func (m *mockRequest) RequireArray(_ string) ([]any, error)        { return nil, nil }
func (m *mockRequest) GetArray(_ string, defaultValue []any) []any { return defaultValue }

func newMockRequest(params map[string]any) ToolRequest {
	return &mockRequest{params: params}
}

func TestRequiredWhen_Description(t *testing.T) {
	type fields struct {
		When       func(ToolRequest) bool
		ParamNames []string
		Message    string
	}
	tests := []struct {
		name            string
		fields          fields
		wantDescription string
	}{
		{
			name: "CustomMessage_ShouldReturnCustomMessage",
			fields: fields{
				When:       nil,
				ParamNames: []string{"param1", "param2"},
				Message:    "Custom error message",
			},
			wantDescription: "Custom error message",
		},
		{
			name: "DefaultMessage_ShouldReturnFormattedMessage",
			fields: fields{
				When:       nil,
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			wantDescription: "'param1', 'param2' parameters are required when condition is met",
		},
		{
			name: "SingleParam_ShouldFormatCorrectly",
			fields: fields{
				When:       nil,
				ParamNames: []string{"param1"},
				Message:    "",
			},
			wantDescription: "'param1' parameters are required when condition is met",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RequiredWhen{
				When:       tt.fields.When,
				ParamNames: tt.fields.ParamNames,
				Message:    tt.fields.Message,
			}
			assert.Equalf(t, tt.wantDescription, r.Description(), "Description()")
		})
	}
}

func TestRequiredWhen_IsSatisfied(t *testing.T) {
	type fields struct {
		When       func(ToolRequest) bool
		ParamNames []string
		Message    string
	}
	type args struct {
		req ToolRequest
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantSatisfied bool
	}{
		{
			name: "ConditionNotMet_ShouldReturnTrue",
			fields: fields{
				When: func(req ToolRequest) bool {
					return false // Condition not met
				},
				ParamNames: []string{"param1"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{}), // No params needed when condition not met
			},
			wantSatisfied: true,
		},
		{
			name: "ConditionMetAndParamsPresent_ShouldReturnTrue",
			fields: fields{
				When: func(req ToolRequest) bool {
					return true // Condition met
				},
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "value1",
					"param2": "value2",
				}),
			},
			wantSatisfied: true,
		},
		{
			name: "ConditionMetButParamsMissing_ShouldReturnFalse",
			fields: fields{
				When: func(req ToolRequest) bool {
					return true // Condition met
				},
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "value1",
					// param2 missing
				}),
			},
			wantSatisfied: false,
		},
		{
			name: "ConditionMetButParamsEmpty_ShouldReturnFalse",
			fields: fields{
				When: func(req ToolRequest) bool {
					return true // Condition met
				},
				ParamNames: []string{"param1"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "", // Empty param
				}),
			},
			wantSatisfied: false,
		},
		{
			name: "ConditionMetButParamNil_ShouldReturnFalse",
			fields: fields{
				When: func(req ToolRequest) bool {
					return true // Condition met
				},
				ParamNames: []string{"param1"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": nil, // Nil param
				}),
			},
			wantSatisfied: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RequiredWhen{
				When:       tt.fields.When,
				ParamNames: tt.fields.ParamNames,
				Message:    tt.fields.Message,
			}
			assert.Equalf(t, tt.wantSatisfied, r.IsSatisfied(tt.args.req), "IsSatisfied(%v)", tt.args.req)
		})
	}
}

func TestRequiresAllOf_Description(t *testing.T) {
	type fields struct {
		ParamNames []string
		Message    string
	}
	tests := []struct {
		name            string
		fields          fields
		wantDescription string
	}{
		{
			name: "CustomMessage_ShouldReturnCustomMessage",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "Custom error message",
			},
			wantDescription: "Custom error message",
		},
		{
			name: "DefaultMessage_ShouldReturnFormattedMessage",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			wantDescription: "all of 'param1', 'param2' parameters are required",
		},
		{
			name: "SingleParam_ShouldFormatCorrectly",
			fields: fields{
				ParamNames: []string{"param1"},
				Message:    "",
			},
			wantDescription: "all of 'param1' parameters are required",
		},
		{
			name: "MultipleParams_ShouldJoinCorrectly",
			fields: fields{
				ParamNames: []string{"param1", "param2", "param3"},
				Message:    "",
			},
			wantDescription: "all of 'param1', 'param2', 'param3' parameters are required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RequiresAllOf{
				ParamNames: tt.fields.ParamNames,
				Message:    tt.fields.Message,
			}
			assert.Equalf(t, tt.wantDescription, r.Description(), "Description()")
		})
	}
}

func TestRequiresAllOf_IsSatisfied(t *testing.T) {
	type fields struct {
		ParamNames []string
		Message    string
	}
	type args struct {
		req ToolRequest
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantSatisfied bool
	}{
		{
			name: "AllParamsPresent_ShouldReturnTrue",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "value1",
					"param2": "value2",
				}),
			},
			wantSatisfied: true,
		},
		{
			name: "OneParamMissing_ShouldReturnFalse",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "value1",
					// param2 missing
				}),
			},
			wantSatisfied: false,
		},
		{
			name: "OneParamEmpty_ShouldReturnFalse",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "value1",
					"param2": "", // Empty param
				}),
			},
			wantSatisfied: false,
		},
		{
			name: "OneParamNil_ShouldReturnFalse",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "value1",
					"param2": nil, // Nil param
				}),
			},
			wantSatisfied: false,
		},
		{
			name: "EmptyArray_ShouldReturnFalse",
			fields: fields{
				ParamNames: []string{"param1"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": []interface{}{}, // Empty array
				}),
			},
			wantSatisfied: false,
		},
		{
			name: "NonEmptyArray_ShouldReturnTrue",
			fields: fields{
				ParamNames: []string{"param1"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": []interface{}{"item1"}, // Non-empty array
				}),
			},
			wantSatisfied: true,
		},
		{
			name: "NoParams_ShouldReturnTrue",
			fields: fields{
				ParamNames: []string{},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{}),
			},
			wantSatisfied: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RequiresAllOf{
				ParamNames: tt.fields.ParamNames,
				Message:    tt.fields.Message,
			}
			assert.Equalf(t, tt.wantSatisfied, r.IsSatisfied(tt.args.req), "IsSatisfied(%v)", tt.args.req)
		})
	}
}

func TestRequiresOneOf_Description(t *testing.T) {
	type fields struct {
		ParamNames []string
		Message    string
	}
	tests := []struct {
		name            string
		fields          fields
		wantDescription string
	}{
		{
			name: "CustomMessage_ShouldReturnCustomMessage",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "Custom error message",
			},
			wantDescription: "Custom error message",
		},
		{
			name: "TwoParams_ShouldReturnEitherOrMessage",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			wantDescription: "either 'param1' or 'param2' parameter is required",
		},
		{
			name: "SingleParam_ShouldReturnOneOfMessage",
			fields: fields{
				ParamNames: []string{"param1"},
				Message:    "",
			},
			wantDescription: "one of 'param1' parameters is required",
		},
		{
			name: "ThreeParams_ShouldReturnOneOfMessage",
			fields: fields{
				ParamNames: []string{"param1", "param2", "param3"},
				Message:    "",
			},
			wantDescription: "one of 'param1', 'param2', 'param3' parameters is required",
		},
		{
			name: "NoParams_ShouldReturnOneOfMessage",
			fields: fields{
				ParamNames: []string{},
				Message:    "",
			},
			wantDescription: "one of '' parameters is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RequiresOneOf{
				ParamNames: tt.fields.ParamNames,
				Message:    tt.fields.Message,
			}
			assert.Equalf(t, tt.wantDescription, r.Description(), "Description()")
		})
	}
}

func TestRequiresOneOf_IsSatisfied(t *testing.T) {
	type fields struct {
		ParamNames []string
		Message    string
	}
	type args struct {
		req ToolRequest
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantSatisfied bool
	}{
		{
			name: "FirstParamPresent_ShouldReturnTrue",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "value1",
					// param2 not provided, but that's OK
				}),
			},
			wantSatisfied: true,
		},
		{
			name: "SecondParamPresent_ShouldReturnTrue",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					// param1 not provided, but that's OK
					"param2": "value2",
				}),
			},
			wantSatisfied: true,
		},
		{
			name: "BothParamsPresent_ShouldReturnTrue",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "value1",
					"param2": "value2",
				}),
			},
			wantSatisfied: true,
		},
		{
			name: "NoParamsPresent_ShouldReturnFalse",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{}),
			},
			wantSatisfied: false,
		},
		{
			name: "OnlyEmptyParamsPresent_ShouldReturnFalse",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "",
					"param2": "",
				}),
			},
			wantSatisfied: false,
		},
		{
			name: "OnlyNilParamsPresent_ShouldReturnFalse",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": nil,
					"param2": nil,
				}),
			},
			wantSatisfied: false,
		},
		{
			name: "EmptyArray_ShouldReturnFalse",
			fields: fields{
				ParamNames: []string{"param1"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": []interface{}{}, // Empty array
				}),
			},
			wantSatisfied: false,
		},
		{
			name: "NonEmptyArray_ShouldReturnTrue",
			fields: fields{
				ParamNames: []string{"param1"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": []interface{}{"item1"}, // Non-empty array
				}),
			},
			wantSatisfied: true,
		},
		{
			name: "OneValidOneEmpty_ShouldReturnTrue",
			fields: fields{
				ParamNames: []string{"param1", "param2"},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "value1", // Valid
					"param2": "",       // Empty, but param1 is valid
				}),
			},
			wantSatisfied: true,
		},
		{
			name: "NoParamNames_ShouldReturnFalse",
			fields: fields{
				ParamNames: []string{},
				Message:    "",
			},
			args: args{
				req: newMockRequest(map[string]any{
					"param1": "value1",
				}),
			},
			wantSatisfied: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RequiresOneOf{
				ParamNames: tt.fields.ParamNames,
				Message:    tt.fields.Message,
			}
			assert.Equalf(t, tt.wantSatisfied, r.IsSatisfied(tt.args.req), "IsSatisfied(%v)", tt.args.req)
		})
	}
}
