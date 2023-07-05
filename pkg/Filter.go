package pkg

import (
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
)

func FilterValidationResults(result ValidationResult, conf *Config) ValidationResult {
	result.ErrorsForLatest = filterError(result.ErrorsForLatest, conf)
	result.ErrorsForOriginal = filterError(result.ErrorsForOriginal, conf)
	return removeIgnoredKeys(result, conf)
}

func filterError(errors []*openapi3.SchemaError, conf *Config) []*openapi3.SchemaError {
	var filteredErrors []*openapi3.SchemaError
	for _, schemaError := range errors {
		if conf.IgnoreNullErrors ||
			excludeArrayNullError(schemaError) ||
			excludeRawExtensionError(schemaError) || excludeCPUMemoryNumberError(schemaError) {
			continue
		}
		filteredErrors = append(filteredErrors, schemaError)
	}
	return filteredErrors
}

func removeIgnoredKeys(result ValidationResult, conf *Config) ValidationResult {
	if len(result.DeprecationForOriginal) > 0 {
		var depErr []*SchemaError
		for _, schemaError := range result.DeprecationForOriginal {
			key := strings.Join(schemaError.JSONPointer(), "/")
			if !Contains(key, conf.IgnoreKeysFromDeprecation) {
				depErr = append(depErr, schemaError)
			}
		}
		result.DeprecationForOriginal = depErr
	}
	if len(result.DeprecationForLatest) > 0 {
		var depErr []*SchemaError
		for _, schemaError := range result.DeprecationForLatest {
			key := strings.Join(schemaError.JSONPointer(), "/")
			if !Contains(key, conf.IgnoreKeysFromDeprecation) {
				depErr = append(depErr, schemaError)
			}
		}
		result.DeprecationForLatest = depErr
	}
	if len(result.ErrorsForOriginal) > 0 {
		var valErr []*openapi3.SchemaError
		for _, schemaError := range result.ErrorsForOriginal {
			key := strings.Join(schemaError.JSONPointer(), "/")
			if !Contains(key, conf.IgnoreKeysFromValidation) {
				valErr = append(valErr, schemaError)
			}
		}
		result.ErrorsForOriginal = valErr
	}
	if len(result.ErrorsForLatest) > 0 {
		var valErr []*openapi3.SchemaError
		for _, schemaError := range result.ErrorsForLatest {
			key := strings.Join(schemaError.JSONPointer(), "/")
			if !Contains(key, conf.IgnoreKeysFromValidation) {
				valErr = append(valErr, schemaError)
			}
		}
		result.ErrorsForLatest = valErr
	}
	return result
}

func excludeCPUMemoryNumberError(schemaError *openapi3.SchemaError) bool {
	penultimateValue := len(schemaError.JSONPointer()) - 2
	requestOrLimit := len(schemaError.JSONPointer()) > 1 && (schemaError.JSONPointer()[penultimateValue] == "requests" || schemaError.JSONPointer()[penultimateValue] == "limits")
	if requestOrLimit {
		switch v := schemaError.Value.(type) {
		case string:
			return v == "number, integer"
		}
	}
	return false
}

func excludeRawExtensionError(schemaError *openapi3.SchemaError) bool {
	return Contains(schemaError.Schema.Description, []string{"RawExtension*"})
}

func excludeArrayNullError(schemaError *openapi3.SchemaError) bool {
	return strings.TrimSpace(schemaError.Reason) == "Value is not nullable" && schemaError.Schema.Type == "array"
}
