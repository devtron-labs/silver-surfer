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
 */

package pkg

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type SchemaSettings struct {
	MultiError bool
}

func VisitJSON(schema *openapi3.Schema, value interface{}, settings SchemaSettings) openapi3.MultiError {
	return visitJSON(schema, value, settings)
}

func visitJSON(schema *openapi3.Schema, value interface{}, settings SchemaSettings) openapi3.MultiError {
	var me openapi3.MultiError
	switch value := value.(type) {
	case nil, bool, float64, string, int64:
		if strings.Contains(strings.ToLower(schema.Description), "deprecated") {
			schemaError := &SchemaError{
				Value:  "",
				Schema: schema,
				Reason: schema.Description,
			}
			me = append(me, schemaError)
		}
		return me
	case []interface{}:
		return visitJSONArray(schema, value, settings)
	case map[string]interface{}:
		return visitJSONObject(schema, value, settings)
	default:
		schemaError := &SchemaError{
			Value:  value,
			Schema: schema,
			Reason: "unhandled key",
		}
		me = append(me, schemaError)
		return me
	}
}

func visitJSONArray(schema *openapi3.Schema, object []interface{}, settings SchemaSettings) openapi3.MultiError {
	var me openapi3.MultiError
	for i, obj := range object {
		schemaError := visitJSON(schema.Items.Value, obj, settings)
		if len(schemaError) != 0 {
			if err := markSchemaErrorIndex(schemaError, i); err != nil {
				panic(err)
			}
			me = append(me, schemaError...)
		}
	}
	return me
}

func visitJSONObject(schema *openapi3.Schema, object map[string]interface{}, settings SchemaSettings) openapi3.MultiError {
	var me openapi3.MultiError
	if strings.Contains(strings.ToLower(schema.Description), "deprecated") {
		schemaError := &SchemaError{
			Value:  "",
			Schema: schema,
			Reason: schema.Description,
		}
		me = append(me, schemaError)
		if !settings.MultiError {
			return me
		}
	}
	for k, v := range object {
		if s, ok := schema.Properties[k]; ok {
			//fmt.Printf("found key %s\n", k)
			schemaError := visitJSON(s.Value, v, settings)
			if len(schemaError) != 0 {
				if err := markSchemaErrorKey(schemaError, k); err != nil {
					panic(err)
				}
				me = append(me, schemaError...)
				if !settings.MultiError {
					return me
				}
			}
		}
	}
	return me
}
