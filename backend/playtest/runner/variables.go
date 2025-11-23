package runner

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// VariableStore stores and retrieves values from server responses
type VariableStore struct {
	vars         map[string]interface{}
	lastResponse map[string]interface{}
}

// NewVariableStore creates a new variable store
func NewVariableStore() *VariableStore {
	return &VariableStore{
		vars:         make(map[string]interface{}),
		lastResponse: make(map[string]interface{}),
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

// SetLastResponse stores the complete last response for field path resolution
func (v *VariableStore) SetLastResponse(response map[string]interface{}) {
	v.lastResponse = response
}

// Variable pattern: <variableName> or <response.field.path>
var variablePattern = regexp.MustCompile(`^<([^>]+)>$`)

// IsVariableReference checks if a string is a variable reference
func IsVariableReference(s string) (string, bool) {
	matches := variablePattern.FindStringSubmatch(s)
	if len(matches) == 2 {
		return matches[1], true
	}
	return "", false
}

// resolveFieldPath resolves a field path like "response.gameId" or "response.players[0].id"
func (v *VariableStore) resolveFieldPath(path string) (interface{}, error) {
	// Check if it starts with "response."
	if !strings.HasPrefix(path, "response.") {
		// Try as a simple variable name (backward compatibility)
		val, ok := v.Get(path)
		if !ok {
			return nil, fmt.Errorf("variable not found: %s", path)
		}
		return val, nil
	}

	// Remove "response." prefix
	fieldPath := strings.TrimPrefix(path, "response.")
	
	// Navigate through the response object
	return v.navigatePath(v.lastResponse, fieldPath)
}

// navigatePath navigates through a nested map/array structure using a dot-notation path
func (v *VariableStore) navigatePath(data interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	// Split on dots, but handle array indices
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		// Check for array index notation: field[index]
		if strings.Contains(part, "[") {
			fieldName := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", indexStr)
			}

			// Navigate to the field first
			if fieldName != "" {
				currentMap, ok := current.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("expected map at path segment: %s", fieldName)
				}
				current, ok = currentMap[fieldName]
				if !ok {
					return nil, fmt.Errorf("field not found: %s", fieldName)
				}
			}

			// Then access the array index
			currentArray, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array at path segment: %s", part)
			}
			if index < 0 || index >= len(currentArray) {
				return nil, fmt.Errorf("array index out of bounds: %d (length: %d)", index, len(currentArray))
			}
			current = currentArray[index]
		} else {
			// Simple field access
			currentMap, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected map at path segment: %s (remaining: %s)", part, strings.Join(parts[i:], "."))
			}
			current, ok = currentMap[part]
			if !ok {
				return nil, fmt.Errorf("field not found: %s (full path: %s)", part, path)
			}
		}
	}

	return current, nil
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
		if varPath, isVar := IsVariableReference(val); isVar {
			storedVal, err := v.resolveFieldPath(varPath)
			if err != nil {
				availableVars := make([]string, 0, len(v.vars))
				for k := range v.vars {
					availableVars = append(availableVars, k)
				}
				return nil, fmt.Errorf("<%s>: %w (available vars: %s)", 
					varPath, err, strings.Join(availableVars, ", "))
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
	// Store the complete response for field path resolution
	v.SetLastResponse(response)
	
	// Automatically extract gameId and store as globalGame (backward compatibility)
	if gameID, ok := response["gameId"].(string); ok && gameID != "" {
		v.Set("globalGame", gameID)
	}
}
