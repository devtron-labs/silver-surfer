package pkg

import (
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
)

func FilterValidationResults(result ValidationResult, conf *Config) ValidationResult {
	//var vr []ValidationResult
	//for _, result := range validationResults {
	result.ErrorsForLatest = filterError(result.ErrorsForLatest, conf)
	result.ErrorsForOriginal = filterError(result.ErrorsForOriginal, conf)
	//vr = append(vr, result)
	//}
	return removeIgnoredKeys(result, conf)
}

func filterError(errors []*openapi3.SchemaError, conf *Config) []*openapi3.SchemaError {
	var filteredErrors []*openapi3.SchemaError
	for _, schemaError := range errors {
		penultimateValue := len(schemaError.JSONPointer()) - 2
		requestOrLimit := len(schemaError.JSONPointer()) > 2 && (schemaError.JSONPointer()[penultimateValue] == "requests" || schemaError.JSONPointer()[penultimateValue] == "limits")
		requestOrLimitIsNumber := false
		if requestOrLimit {
			switch v := schemaError.Value.(type) {
			case string:
				requestOrLimitIsNumber = v == "number, integer"
			}
		}
		if conf.IgnoreNullErrors ||
			(strings.TrimSpace(schemaError.Reason) == "Value is not nullable" && schemaError.Schema.Type == "array") ||
			Contains(schemaError.Schema.Description, []string{"RawExtension*"}) || requestOrLimitIsNumber {
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
