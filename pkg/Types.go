/*
 * Copyright (c) 2021 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Certain portions in this file have been taken from kin-openapi and where ever
 * they are, IP and licenses of kin-openapi are applicable.
 */

package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/xeipuuv/gojsonschema"
)

var SchemaErrorDetailsDisabled = true

// ValidFormat is a type for quickly forcing
// new formats on the gojsonschema loader
type ValidFormat struct{}

// IsFormat always returns true and meets the
// gojsonschema.FormatChecker interface
func (f ValidFormat) IsFormat(input interface{}) bool {
	return true
}

// ValidationResult contains the details from
// validating a given Kubernetes resource
type ValidationResult struct {
	FileName               string
	Kind                   string
	APIVersion             string
	ValidatedAgainstSchema bool
	Errors                 []gojsonschema.ResultError
	ErrorsForOriginal      []*openapi3.SchemaError
	ErrorsForLatest        []*openapi3.SchemaError
	DeprecationForOriginal []*SchemaError
	DeprecationForLatest   []*SchemaError
	ResourceName           string
	ResourceNamespace      string
	Deleted                bool
	Deprecated             bool
	LatestAPIVersion       string
	IsVersionSupported     int
}

// VersionKind returns a string representation of this result's apiVersion and kind
func (v *ValidationResult) VersionKind() string {
	return v.APIVersion + "/" + v.Kind
}

// QualifiedName returns a string of the [namespace.]name of the k8s resource
func (v *ValidationResult) QualifiedName() string {
	if v.ResourceName == "" {
		return "unknown"
	} else if v.ResourceNamespace == "" {
		return v.ResourceName
	} else {
		return fmt.Sprintf("%s.%s", v.ResourceNamespace, v.ResourceName)
	}
}

type SchemaError struct {
	Value       interface{}
	reversePath []string
	Schema      *openapi3.Schema
	SchemaField string
	Reason      string
	Origin      error
}

func markSchemaErrorKey(err error, key string) error {
	if v, ok := err.(*SchemaError); ok {
		v.reversePath = append(v.reversePath, key)
		return v
	}
	if v, ok := err.(openapi3.MultiError); ok {
		for _, e := range v {
			_ = markSchemaErrorKey(e, key)
		}
		return v
	}
	return err
}

func markSchemaErrorIndex(err error, index int) error {
	if v, ok := err.(*SchemaError); ok {
		v.reversePath = append(v.reversePath, strconv.FormatInt(int64(index), 10))
		return v
	}
	if v, ok := err.(openapi3.MultiError); ok {
		for _, e := range v {
			_ = markSchemaErrorIndex(e, index)
		}
		return v
	}
	return err
}

func (err *SchemaError) JSONPointer() []string {
	reversePath := err.reversePath
	path := append([]string(nil), reversePath...)
	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}
	return path
}

func (err *SchemaError) Error() string {
	if err.Origin != nil {
		return err.Origin.Error()
	}

	buf := bytes.NewBuffer(make([]byte, 0, 256))
	if len(err.reversePath) > 0 {
		buf.WriteString(`Error at "`)
		reversePath := err.reversePath
		for i := len(reversePath) - 1; i >= 0; i-- {
			buf.WriteByte('/')
			buf.WriteString(reversePath[i])
		}
		buf.WriteString(`": `)
	}
	reason := err.Reason
	if reason == "" {
		buf.WriteString(`Doesn't match schema "`)
		buf.WriteString(err.SchemaField)
		buf.WriteString(`"`)
	} else {
		buf.WriteString(reason)
	}
	if !SchemaErrorDetailsDisabled {
		buf.WriteString("\nSchema:\n  ")
		encoder := json.NewEncoder(buf)
		encoder.SetIndent("  ", "  ")
		if err := encoder.Encode(err.Schema); err != nil {
			panic(err)
		}
		buf.WriteString("\nValue:\n  ")
		if err := encoder.Encode(err.Value); err != nil {
			panic(err)
		}
	}
	return buf.String()
}

func isSliceOfUniqueItems(xs []interface{}) bool {
	s := len(xs)
	m := make(map[string]struct{}, s)
	for _, x := range xs {
		// The input slice is coverted from a JSON string, there shall
		// have no error when covert it back.
		key, _ := json.Marshal(&x)
		m[string(key)] = struct{}{}
	}
	return s == len(m)
}

// SliceUniqueItemsChecker is an function used to check if an given slice
// have unique items.
type SliceUniqueItemsChecker func(items []interface{}) bool

// By default using predefined func isSliceOfUniqueItems which make use of
// json.Marshal to generate a key for map used to check if a given slice
// have unique items.
var sliceUniqueItemsChecker SliceUniqueItemsChecker = isSliceOfUniqueItems

// RegisterArrayUniqueItemsChecker is used to register a customized function
// used to check if JSON array have unique items.
func RegisterArrayUniqueItemsChecker(fn SliceUniqueItemsChecker) {
	sliceUniqueItemsChecker = fn
}

// Unused Function
//
// func unsupportedFormat(format string) error {
// 	return fmt.Errorf("unsupported 'format' value %q", format)
// }

/*
ApiVersion/Kind
RestPath - string - to find if deleted or not
ComponentKey - string -  to find schema for validation
ApiVersions - array - for migration - ga or nots
*/

type GroupVersions struct {
	GroupVersions   []string
	GAGroupVersions []string
}

type KindInfo struct {
	Version      string
	Group        string
	RestPath     string
	ComponentKey string
	IsGA         bool
}
