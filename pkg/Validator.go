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
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/silver-surfer/pkg/log"
	"github.com/getkin/kin-openapi/openapi3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
	"sort"
	"strings"
)



type kubeSpec struct {
	*openapi3.T
	kindInfoMap map[string][]*KindInfo
}

func newKubeSpec(openapi *openapi3.T) *kubeSpec {
	ks := &kubeSpec{T: openapi}
	ks.kindInfoMap = ks.buildKindInfoMap()
	return ks
}

func (ks *kubeSpec) ValidateYaml(spec string) (ValidationResult, error) {
	var err error
	jsonSpec, err := yaml.YAMLToJSON([]byte(spec))
	if err != nil {
		log.Debug(fmt.Sprintf("%v", err))
		return ValidationResult{}, err
	}
	return ks.ValidateJson(string(jsonSpec))
}

func (ks *kubeSpec) ValidateJson(spec string) (ValidationResult, error) {
	var err error
	object := make(map[string]interface{})
	err = json.Unmarshal([]byte(spec), &object)
	if err != nil {
		log.Debug(fmt.Sprintf("%v", err))
		return ValidationResult{}, err
	}
	return ks.ValidateObject(object)
}

func (ks *kubeSpec) ValidateObject(object map[string]interface{}) (ValidationResult, error) {
	validationResult, err := ks.populateValidationResult(object)
	validationResult.ValidatedAgainstSchema = true
	if err != nil {
		return validationResult, err
	}
	original, latest, err := ks.getKindsMappings(object)
	if err != nil {
		return validationResult, err
	}
	if len(original) > 0 {
		var ves []*openapi3.SchemaError
		var des []*SchemaError
		validationError, deprecated := ks.applySchema(object, original)
		if validationError != nil && len(validationError) > 0 {
			errs := []error(validationError)
			for _, e := range errs {
				if se, ok := e.(*openapi3.SchemaError); ok {
					ves = append(ves, se)
				} else if de, ok := e.(*SchemaError); ok {
					des = append(des, de)
				}
			}
		}
		validationResult.ErrorsForOriginal = ves
		validationResult.DeprecationForOriginal = des
		validationResult.Deprecated = deprecated
	} else if len(original) == 0 && len(latest) > 0 {
		validationResult.Deleted = true
	}
	if len(latest) > 0 && original != latest {
		var ves []*openapi3.SchemaError
		var des []*SchemaError
		validationError, _ := ks.applySchema(object, latest)
		if validationError != nil && len(validationError) > 0 {
			errs := []error(validationError)
			for _, e := range errs {
				if se, ok := e.(*openapi3.SchemaError); ok {
					ves = append(ves, se)
				} else if de, ok := e.(*SchemaError); ok {
					des = append(des, de)
				}
			}
		}
		validationResult.ErrorsForLatest = ves
		validationResult.DeprecationForLatest = des
		validationResult.LatestAPIVersion, err = ks.getKeyForGVFromToken(latest)
	}
	return validationResult, nil
}

func (ks *kubeSpec) buildGVKRestPathMap() map[string]string {
	pathMap := map[string]string{}
	for path, value := range ks.T.Paths {
		var method *openapi3.Operation
		if value.Post != nil {
			method = value.Post
		} else if value.Put != nil {
			method = value.Put
		}
		if method != nil {
			if gvk, ok := method.Extensions["x-kubernetes-group-version-kind"]; ok {
				gvks, err := getKeyForGVK(gvk.(json.RawMessage))
				if err != nil {
					continue
				}
				pathMap[gvks] = path
			}
		}
	}
	return pathMap
}

func (ks *kubeSpec) buildKindInfoMap() map[string][]*KindInfo {
	kindMap := map[string][]*KindInfo{}
	restPath := ks.buildGVKRestPathMap()
	for component, value := range ks.T.Components.Schemas {
		if gvk, ok := value.Value.Extensions["x-kubernetes-group-version-kind"]; ok {
			gvks, err := parseGVK(gvk.(json.RawMessage))
			if err != nil {
				continue
			}
			kind := strings.ToLower(gvks["kind"])
			if _, ok := kindMap[kind]; !ok {
				kindMap[kind] = make([]*KindInfo, 0)
			}
			ki := KindInfo{
				Version:      gvks["version"],
				Group:        gvks["group"],
				RestPath:     "",
				ComponentKey: component,
				IsGA:         getVersionType(gvks["version"]) == gaVersion,
			}
			gvkKey, err := getKeyForGVK(gvk.(json.RawMessage))
			if err != nil {
				log.Error(err)
				continue
			}
			if p, ok := restPath[gvkKey]; ok {
				ki.RestPath = p
			}
			kindMap[kind] = append(kindMap[kind], &ki)
		}
	}
	for kind, gvs := range kindMap {
		sort.Slice(gvs, func(i, j int) bool {
			return compareVersion(gvs[i].Version, gvs[j].Version)
		})
		kindMap[kind] = gvs
	}
	return kindMap
}

func (ks *kubeSpec) getLatestKinds() []schema.GroupVersionKind {
	gvkMap := make(map[string]bool, 0)
	var gvka []schema.GroupVersionKind
	for kind, info := range ks.kindInfoMap {
		last := info[len(info)-1]
		if len(last.RestPath) == 0 {
			continue
		}
		gvk := schema.GroupVersionKind{
			Group:   last.Group,
			Version: last.Version,
			Kind:    kind,
		}
		if _, ok := gvkMap[gvk.String()]; ok {
			continue
		}
		gvkMap[gvk.String()] = true
		gvka = append(gvka, gvk)
	}
	return gvka
}

func (ks *kubeSpec) isApiVersionSupported(apiVersion, kind string) bool {
	if kim, ok := ks.kindInfoMap[strings.ToLower(kind)]; ok {
		//fmt.Printf("found %s \n", kind)
		for _, ki := range kim {
			//fmt.Printf("%s:%s:%s\n", apiVersion, ki.Version, ki.RestPath)
			gv := ki.Version
			if len(ki.Group) > 0 {
				gv = fmt.Sprintf("%s/%s", ki.Group, ki.Version)
			}
			if strings.EqualFold(gv, apiVersion) && len(ki.RestPath) > 0 {
				return true
			}
		}
	}
	return false
}

func (ks *kubeSpec) populateValidationResult(object map[string]interface{}) (ValidationResult, error) {
	validationResult := ValidationResult{}
	namespace := "undefined"
	if object == nil {
		return validationResult, fmt.Errorf("missing k8s object")
	}
	apiVersion, ok := object["apiVersion"].(string)
	kind, ok := object["kind"].(string)
	if !ok {
		return validationResult, fmt.Errorf("missing kind")
	}
	metadata, ok := object["metadata"].(map[string]interface{})
	if !ok {
		return validationResult, fmt.Errorf("missing metadata")
	}
	if ns, ok := metadata["namespace"]; ok {
		namespace = ns.(string)
	}
	name, ok := metadata["name"].(string)
	if !ok {
		return validationResult, fmt.Errorf("missing resource name")
	}
	validationResult.Kind = kind
	validationResult.APIVersion = apiVersion
	validationResult.ResourceNamespace = namespace
	validationResult.ResourceName = name
	return validationResult, nil
}

func (ks *kubeSpec) applySchema(object map[string]interface{}, token string) (openapi3.MultiError, bool) {
	deprecated := false
	var validationError openapi3.MultiError
	scm, err := ks.schemaLookup(token)
	if err != nil {
		log.Debug(fmt.Sprintf("%v", err))
		validationError = append(validationError, err)
		return validationError, deprecated
	}

	opts := []openapi3.SchemaValidationOption{openapi3.MultiErrors()}
	depError := VisitJSON(scm, object, SchemaSettings{MultiError: true})
	if len(depError) > 0 {
		deprecated = true
	}
	validationError = append(validationError, depError...)

	err = scm.VisitJSON(object, opts...)
	if err != nil {
		e := err.(openapi3.MultiError)
		validationError = append(validationError, e...)
	}
	return validationError, deprecated
}

func (ks *kubeSpec) getKeyForGVFromToken(token string) (string, error) {
	scm, err := ks.schemaLookup(token)
	if err != nil {
		log.Debug(fmt.Sprintf("%v", err))
		return "", err
	}
	gv, err := getKeyForGV(scm.Extensions["x-kubernetes-group-version-kind"].(json.RawMessage))
	if err != nil {
		return "", err
	}
	return gv, nil
}

func (ks *kubeSpec) schemaLookup(token string) (*openapi3.Schema, error) {
	for ; ; {
		if strings.Index(token, "/") > 0 {
			parts := strings.Split(token, "/")
			token = parts[len(parts)-1]
		}
		dp, err := ks.Components.Schemas.JSONLookup(token)
		if err != nil {
			return nil, err
		}
		if scm, ok := dp.(*openapi3.Schema); ok {
			return scm, nil
		}
		if ref, ok := dp.(*openapi3.Ref); ok {
			token = ref.Ref
		}
	}
}

func (ks *kubeSpec) getKindsMappings(object map[string]interface{}) (original, latest string, err error) {
	original = ""
	latest = ""
	if object == nil {
		return "", "", fmt.Errorf("missing k8s object")
	}
	apiVersion, ok := object["apiVersion"].(string)
	kind, ok := object["kind"].(string)
	if !ok {
		return "", "", fmt.Errorf("missing kind")
	}
	parts := strings.Split(apiVersion, "/")
	if len(parts) == 0 {
		return "", "", fmt.Errorf("unable to parse group and version from %s", apiVersion)
	}
	if len(parts) == 1 {
		parts = []string{"", parts[0]}
	}
	if kis, ok := ks.kindInfoMap[strings.ToLower(kind)]; ok {
		for _, ki := range kis {
			if parts[0] == ki.Group && parts[1] == ki.Version && len(ki.RestPath) > 0 {
				original = ki.ComponentKey
			}
		}
		if len(kis) > 0 {
			latest = kis[len(kis)-1].ComponentKey
		}
	}
	return original, latest, nil
}
