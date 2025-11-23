package runner

import (
	"fmt"
	"regexp"
	"strings"
)

// VariableStore stores and retrieves values from server responses
type VariableStore struct {
	vars map[string]interface{}
}

// NewVariableStore creates a new variable store
func NewVariableStore() *VariableStore {
	return &VariableStore{
		vars: make(map[string]interface{}),
	}
}

// Set stores a value in the variable store
func (v *VariableStore) Set(key string, value interface{}) {
	v.vars[key] = value
}

// Get retrieves a value from the variable store
func (v *VariableStore) Get(key string) (interface{}, bool) {
	val, ok := v.vars[key]
	return val, ok
}

// Variable pattern: <variableName>
var variablePattern = regexp.MustCompile(`^<([^>]+)>$`)

// IsVariableReference checks if a string is a variable reference
func IsVariableReference(s string) (string, bool) {
	matches := variablePattern.FindStringSubmatch(s)
	if len(matches) == 2 {
		return matches[1], true
	}
	return "", false
}

// Substitute replaces variable references in a message template with stored values
func (v *VariableStore) Substitute(template map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range template {
		substituted, err := v.substituteValue(value)
		if err != nil {
			return nil, err
		}
		result[key] = substituted
	}

	return result, nil
}

// substituteValue recursively substitutes variables in a value
func (v *VariableStore) substituteValue(value interface{}) (interface{}, error) {
	switch val := value.(type) {
	case string:
		// Check if it's a variable reference
		if varName, isVar := IsVariableReference(val); isVar {
			storedVal, ok := v.Get(varName)
			if !ok {
				availableVars := make([]string, 0, len(v.vars))
				for k := range v.vars {
					availableVars = append(availableVars, k)
				}
				return nil, fmt.Errorf("variable <%s> not found in store (available: %s)", 
					varName, strings.Join(availableVars, ", "))
			}
			return storedVal, nil
		}
		// Empty strings and regular strings are preserved as-is
		return val, nil

	case map[string]interface{}:
		// Recursively substitute in nested maps
		result := make(map[string]interface{})
		for k, nestedVal := range val {
			substituted, err := v.substituteValue(nestedVal)
			if err != nil {
				return nil, err
			}
			result[k] = substituted
		}
		return result, nil

	case []interface{}:
		// Recursively substitute in arrays
		result := make([]interface{}, len(val))
		for i, item := range val {
			substituted, err := v.substituteValue(item)
			if err != nil {
				return nil, err
			}
			result[i] = substituted
		}
		return result, nil

	default:
		// Numbers, booleans, etc. are preserved as-is
		return val, nil
	}
}

// ExtractAndStore extracts specific fields from a response and stores them
func (v *VariableStore) ExtractAndStore(response map[string]interface{}) {
	// Automatically extract gameId and store as globalGame
	if gameID, ok := response["gameId"].(string); ok && gameID != "" {
		v.Set("globalGame", gameID)
	}
}
