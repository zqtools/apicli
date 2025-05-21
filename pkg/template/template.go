package template

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Renderer handles template variable substitution
type Renderer struct {
params map[string]interface{}
}

// GetParams returns the current parameter map
func (r *Renderer) GetParams() map[string]interface{} {
return r.params
}

// NewRenderer creates a new template renderer with the given parameters
func NewRenderer(params map[string]interface{}) *Renderer {
	return &Renderer{
		params: params,
	}
}

// Render substitutes template variables in the given string with their corresponding values
func (r *Renderer) Render(tmpl string) (string, error) {
	// If template contains no variables, return as is
	if !strings.Contains(tmpl, "${") {
		return tmpl, nil
	}

	// Find all variables using regex
	re := regexp.MustCompile(`\${([^}]+)}`)
	result := tmpl

	// Replace all variables
	matches := re.FindAllStringSubmatch(tmpl, -1)
	for _, match := range matches {
		fullMatch := match[0] // ${var}
		varName := match[1]   // var

		value, ok := r.params[varName]
		if !ok {
			return "", fmt.Errorf("parameter '%s' not found", varName)
		}

		// Convert value to string based on type
		strValue, err := r.convertToString(value)
		if err != nil {
			return "", fmt.Errorf("converting value for parameter '%s': %w", varName, err)
		}

		result = strings.Replace(result, fullMatch, strValue, -1)
	}

	return result, nil
}

// convertToString converts an interface{} value to a string
func (r *Renderer) convertToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case bool:
		return strconv.FormatBool(v), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case []interface{}:
		parts := make([]string, len(v))
		for i, item := range v {
			str, err := r.convertToString(item)
			if err != nil {
				return "", err
			}
			parts[i] = str
		}
		return strings.Join(parts, ","), nil
	case map[string]interface{}:
		pairs := make([]string, 0, len(v))
		for key, val := range v {
			str, err := r.convertToString(val)
			if err != nil {
				return "", err
			}
			pairs = append(pairs, fmt.Sprintf("%s:%s", key, str))
		}
		return strings.Join(pairs, ","), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// ValidateType validates if a value matches the expected type
func ValidateType(value string, typeName string) error {
	switch {
	case strings.HasPrefix(typeName, "array["):
		if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
			return fmt.Errorf("value must be an array")
		}
	case typeName == "integer":
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("value must be an integer")
		}
	case typeName == "boolean":
		if value != "true" && value != "false" {
			return fmt.Errorf("value must be a boolean")
		}
	case typeName == "number":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("value must be a number")
		}
	}
	return nil
}
