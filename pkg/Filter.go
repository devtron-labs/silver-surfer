package pkg

import (
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
)

func FilterValidationResults(result ValidationResult, conf *Config) ValidationResult {
	//var vr []ValidationResult
	//for _, result := range validationResults {
		var errorsForLatest []*openapi3.SchemaError
		var errorsForOriginal []*openapi3.SchemaError
		for _, schemaError := range result.ErrorsForLatest {
			if !conf.IgnoreNullErrors || strings.TrimSpace(schemaError.Reason) != "Value is not nullable" {
				errorsForLatest = append(errorsForLatest, schemaError)
			}
		}
		result.ErrorsForLatest = errorsForLatest
		for _, schemaError := range result.ErrorsForOriginal {
			if !conf.IgnoreNullErrors || strings.TrimSpace(schemaError.Reason) != "Value is not nullable" {
				errorsForOriginal = append(errorsForOriginal, schemaError)
			}
		}
		result.ErrorsForOriginal = errorsForOriginal
		//vr = append(vr, result)
	//}
	return removeIgnoredKeys(result, conf)
}

func removeIgnoredKeys(result ValidationResult, conf *Config) ValidationResult {
	//var out []ValidationResult
	//for _, result := range results {
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
		//out = append(out, result)
	//}
	return result
}