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

package kubedd

import (
	"bytes"
	"fmt"
	"github.com/devtron-labs/deprecation-checker/pkg"
	kLog "github.com/devtron-labs/deprecation-checker/pkg/log"
	"github.com/getkin/kin-openapi/openapi3"
	"os"
	"strings"
)

var yamlSeparator = []byte("\n---\n")

// Validate a Kubernetes YAML file, parsing out individual resources
// and validating them all according to the  relevant schemas
func Validate(input []byte, conf *pkg.Config) ([]pkg.ValidationResult, error) {
	kubeC := pkg.NewKubeCheckerImpl()
	if len(conf.TargetSchemaLocation) > 0 {
		err := kubeC.LoadFromPath(conf.TargetKubernetesVersion, conf.TargetSchemaLocation, false)
		if err != nil {
			kLog.Error(err)
			os.Exit(1)
		}
	} else {
		err := kubeC.LoadFromUrl(conf.TargetKubernetesVersion, false)
		if err != nil {
			kLog.Error(err)
			os.Exit(1)
		}
	}
	if len(conf.SourceSchemaLocation) > 0 {
		err := kubeC.LoadFromPath(conf.SourceKubernetesVersion, conf.SourceSchemaLocation, false)
		if err != nil {
			kLog.Error(err)
			os.Exit(1)
		}
	} else {
		err := kubeC.LoadFromUrl(conf.SourceKubernetesVersion, false)
		if err != nil {
			kLog.Error(err)
			os.Exit(1)
		}
	}
	if len(conf.SourceKubernetesVersion) == 0 && len(conf.TargetKubernetesVersion) != 0 {
		conf.SourceKubernetesVersion = conf.TargetKubernetesVersion
	}
	splits := bytes.Split(input, yamlSeparator)
	var validationResults []pkg.ValidationResult
	for _, split := range splits {
		validationResult, err := kubeC.ValidateYaml(string(split), conf.SourceKubernetesVersion)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			continue
		}
		validationResults = append(validationResults, validationResult)
	}
	apiVersionKindCache := make(map[string]bool, 0)
	for i, result := range validationResults {
		latestAPIVersion := result.LatestAPIVersion
		if len(result.LatestAPIVersion) == 0 {
			latestAPIVersion = result.APIVersion
		}
		if _, ok := apiVersionKindCache[fmt.Sprintf("%s/%s", latestAPIVersion, result.Kind)]; !ok {
			isSupported := kubeC.IsVersionSupported(conf.TargetKubernetesVersion,  latestAPIVersion, result.Kind)
			apiVersionKindCache[fmt.Sprintf("%s/%s", latestAPIVersion, result.Kind)] = isSupported

		}
		isSupported := apiVersionKindCache[fmt.Sprintf("%s/%s", latestAPIVersion, result.Kind)]

		if isSupported {
			result.IsVersionSupported = 1
		} else {
			result.IsVersionSupported = 2
			result.Deleted = true
		}
		if _, ok := apiVersionKindCache[fmt.Sprintf("%s/%s", result.APIVersion, result.Kind)]; !ok {
			isSupported := kubeC.IsVersionSupported(conf.TargetKubernetesVersion,  result.APIVersion, result.Kind)
			apiVersionKindCache[fmt.Sprintf("%s/%s", result.APIVersion, result.Kind)] = isSupported
		}
		isSupported = apiVersionKindCache[fmt.Sprintf("%s/%s", result.APIVersion, result.Kind)]
		result.Deleted = !isSupported
		validationResults[i] = result
	}

	for i, result := range validationResults {
		var errorsForLatest []*openapi3.SchemaError
		var errorsForOriginal []*openapi3.SchemaError
		for _, schemaError := range result.ErrorsForLatest {
			if !conf.IgnoreNullErrors || strings.TrimSpace(schemaError.Reason) != "Value is not nullable" {
				errorsForLatest = append(errorsForLatest, schemaError)
			}
		}
		result.ErrorsForLatest = errorsForLatest
		for _, schemaError := range result.ErrorsForOriginal {
			if !conf.IgnoreNullErrors  || strings.TrimSpace(schemaError.Reason) != "Value is not nullable" {
				errorsForOriginal = append(errorsForOriginal, schemaError)
			}
		}
		result.ErrorsForOriginal = errorsForOriginal
		validationResults[i] = result
	}
	return validationResults, nil
}

func ValidateCluster(cluster *pkg.Cluster, conf *pkg.Config) ([]pkg.ValidationResult, error) {
	kubeC := pkg.NewKubeCheckerImpl()
	if len(conf.TargetSchemaLocation) > 0 {
		err := kubeC.LoadFromPath(conf.TargetKubernetesVersion, conf.TargetSchemaLocation, false)
		if err != nil {
			kLog.Error(err)
			os.Exit(1)
		}
	} else {
		err := kubeC.LoadFromUrl(conf.TargetKubernetesVersion, false)
		if err != nil {
			kLog.Error(err)
			os.Exit(1)
		}
	}
	serverVersion, err := cluster.ServerVersion()
	if err != nil {
		kLog.Error( err)
		serverVersion = conf.TargetKubernetesVersion
	}
	resources, err := kubeC.GetKinds(serverVersion)
	if err != nil {
		kLog.Error(err)
		//return make([]pkg.ValidationResult, 0), nil
		resources, err = kubeC.GetKinds(conf.TargetKubernetesVersion)
		if err != nil {
			kLog.Error(err)
			return make([]pkg.ValidationResult, 0), nil
		}
	}
	objects := cluster.FetchK8sObjects(resources, conf)
	var validationResults []pkg.ValidationResult
	for _, obj := range objects {
		annon := obj.GetAnnotations()
		k8sObj := ""
		if val, ok := annon["kubectl.kubernetes.io/last-applied-configuration"]; ok {
			k8sObj = val
		} else {
			bt, err := obj.MarshalJSON()
			if err != nil {
				continue
			}
			k8sObj = string(bt)
		}
		validationResult, err := kubeC.ValidateJson(k8sObj, serverVersion)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			continue
		}
		validationResults = append(validationResults, validationResult)
	}
	apiVersionKindCache := make(map[string]bool, 0)
	for i, result := range validationResults {
		latestAPIVersion := result.LatestAPIVersion
		if len(result.LatestAPIVersion) == 0 {
			latestAPIVersion = result.APIVersion
		}
		if _, ok := apiVersionKindCache[fmt.Sprintf("%s/%s", latestAPIVersion, result.Kind)]; !ok {
			isSupported := kubeC.IsVersionSupported(conf.TargetKubernetesVersion,  latestAPIVersion, result.Kind)
			apiVersionKindCache[fmt.Sprintf("%s/%s", latestAPIVersion, result.Kind)] = isSupported

		}
		isSupported := apiVersionKindCache[fmt.Sprintf("%s/%s", latestAPIVersion, result.Kind)]

		if isSupported {
			result.IsVersionSupported = 1
		} else {
			result.IsVersionSupported = 2
			result.Deleted = true
		}
		if _, ok := apiVersionKindCache[fmt.Sprintf("%s/%s", result.APIVersion, result.Kind)]; !ok {
			isSupported := kubeC.IsVersionSupported(conf.TargetKubernetesVersion,  result.APIVersion, result.Kind)
			apiVersionKindCache[fmt.Sprintf("%s/%s", result.APIVersion, result.Kind)] = isSupported
		}
		isSupported = apiVersionKindCache[fmt.Sprintf("%s/%s", result.APIVersion, result.Kind)]
		result.Deleted = !isSupported
		validationResults[i] = result
	}

	for i, result := range validationResults {
		var errorsForLatest []*openapi3.SchemaError
		var errorsForOriginal []*openapi3.SchemaError
		for _, schemaError := range result.ErrorsForLatest {
			if !conf.IgnoreNullErrors || strings.TrimSpace(schemaError.Reason) != "Value is not nullable" {
				errorsForLatest = append(errorsForLatest, schemaError)
			}
		}
		result.ErrorsForLatest = errorsForLatest
		for _, schemaError := range result.ErrorsForOriginal {
			if !conf.IgnoreNullErrors  || strings.TrimSpace(schemaError.Reason) != "Value is not nullable" {
				errorsForOriginal = append(errorsForOriginal, schemaError)
			}
		}
		result.ErrorsForOriginal = errorsForOriginal
		validationResults[i] = result
	}
	return validationResults, nil
}
