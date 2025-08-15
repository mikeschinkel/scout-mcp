package mcputil

import "strings"

// RequiresOneOf ensures at least one of the specified parameters is provided
type RequiresOneOf struct {
	ParamNames []string
	Message    string // Optional custom error message
}

func (r RequiresOneOf) RequirementOption() {}

func (r RequiresOneOf) IsSatisfied(req ToolRequest) (satisfied bool) {
	var args map[string]any
	var value any
	var exists bool
	var arr []any
	var ok bool

	args = req.CallToolRequest().GetArguments()
	for _, paramName := range r.ParamNames {
		value, exists = args[paramName]
		if !exists {
			continue
		}

		// Check if the value is not empty/nil
		if empty(value) {
			continue
		}

		// For arrays, check if not empty
		arr, ok = value.([]any)
		if !ok {
			// Not an array, and not empty
			satisfied = true
			goto end
		}
		if len(arr) <= 0 {
			continue
		}
		satisfied = true
		goto end
	}
end:
	return satisfied
}

func (r RequiresOneOf) Description() (description string) {
	if r.Message != "" {
		description = r.Message
		goto end
	}
	if len(r.ParamNames) == 2 {
		description = "either '" + r.ParamNames[0] + "' or '" + r.ParamNames[1] + "' parameter is required"
		goto end
	}
	description = "one of '" + strings.Join(r.ParamNames, "', '") + "' parameters is required"

end:
	return description
}

// RequiresAllOf ensures all of the specified parameters are provided
type RequiresAllOf struct {
	ParamNames []string
	Message    string // Optional custom error message
}

func (r RequiresAllOf) RequirementOption() {}

func (r RequiresAllOf) IsSatisfied(req ToolRequest) (satisfied bool) {
	var args map[string]any
	var value any
	var exists bool
	var arr []any
	var ok bool

	satisfied = true
	args = req.CallToolRequest().GetArguments()
	for _, paramName := range r.ParamNames {
		value, exists = args[paramName]
		if !exists || empty(value) {
			satisfied = false
			goto end
		}
		// For arrays, check if not empty
		arr, ok = value.([]any)
		if ok && len(arr) == 0 {
			satisfied = false
			goto end
		}
	}
end:
	return satisfied
}

func (r RequiresAllOf) Description() (description string) {
	if r.Message != "" {
		description = r.Message
		goto end
	}
	description = "all of '" + strings.Join(r.ParamNames, "', '") + "' parameters are required"

end:
	return description
}

// RequiredWhen ensures parameters are provided when a condition is met
type RequiredWhen struct {
	When       func(ToolRequest) bool
	ParamNames []string
	Message    string // Optional custom error message
}

func (r RequiredWhen) RequirementOption() {}

func (r RequiredWhen) IsSatisfied(req ToolRequest) (satisfied bool) {
	var args map[string]any
	var value any
	var exists bool

	satisfied = true
	if !r.When(req) {
		// Condition not met, so requirement doesn't apply
		goto end
	}

	args = req.CallToolRequest().GetArguments()
	for _, paramName := range r.ParamNames {
		value, exists = args[paramName]
		if !exists || value == nil || value == "" {
			satisfied = false
			goto end
		}
	}

end:
	return satisfied
}

func (r RequiredWhen) Description() (description string) {
	if r.Message != "" {
		description = r.Message
		goto end
	}
	description = "'" + strings.Join(r.ParamNames, "', '") + "' parameters are required when condition is met"

end:
	return description
}
