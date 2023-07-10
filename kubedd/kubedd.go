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
	"github.com/devtron-labs/silver-surfer/pkg"
	kLog "github.com/devtron-labs/silver-surfer/pkg/log"
	"os"
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
	//isVersionSupported := isVersionSupported()
	for _, split := range splits {
		validationResult, err := kubeC.ValidateYaml(string(split), conf.TargetKubernetesVersion)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			continue
		}
		//validationResult = isVersionSupported(validationResult, kubeC, conf)
		validationResult = pkg.FilterValidationResults(validationResult, conf)
		validationResults = append(validationResults, validationResult)
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
		kLog.Error(err)
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
	//isVersionSupported := isVersionSupported()
	for _, obj := range objects {
		annotations := obj.GetAnnotations()
		k8sObj := ""
		if val, ok := annotations["kubectl.kubernetes.io/last-applied-configuration"]; ok {
			k8sObj = val
		} else {
			bt, err := obj.MarshalJSON()
			if err != nil {
				continue
			}
			k8sObj = string(bt)
		}
		if len(k8sObj) > 0 {
			validationResult, err := kubeC.ValidateJson(k8sObj, conf.TargetKubernetesVersion)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				continue
			}
			//validationResult = isVersionSupported(validationResult, kubeC, conf)
			validationResult = pkg.FilterValidationResults(validationResult, conf)
			validationResults = append(validationResults, validationResult)
		}
	}

	return validationResults, nil
}

//func isVersionSupported() func(result pkg.ValidationResult, kubeC pkg.KubeChecker, conf *pkg.Config) pkg.ValidationResult {
//	apiVersionKindCache := make(map[string]bool, 0)
//	return func(result pkg.ValidationResult, kubeC pkg.KubeChecker, conf *pkg.Config) pkg.ValidationResult {
//		latestAPIVersion := result.LatestAPIVersion
//		if len(result.LatestAPIVersion) == 0 {
//			latestAPIVersion = result.APIVersion
//		}
//		latestAPIVersionKind := fmt.Sprintf("%s/%s", latestAPIVersion, result.Kind)
//		apiVersionKind := fmt.Sprintf("%s/%s", result.APIVersion, result.Kind)
//
//		if _, ok := apiVersionKindCache[latestAPIVersionKind]; !ok {
//			isSupported := kubeC.IsApiVersionSupported(conf.TargetKubernetesVersion, latestAPIVersion, result.Kind)
//			apiVersionKindCache[latestAPIVersionKind] = isSupported
//
//		}
//		isSupported := apiVersionKindCache[latestAPIVersionKind]
//
//		if isSupported {
//			result.IsVersionSupported = 1
//		} else {
//			result.IsVersionSupported = 2
//		}
//
//		if _, ok := apiVersionKindCache[apiVersionKind]; !ok {
//			isSupported := kubeC.IsApiVersionSupported(conf.TargetKubernetesVersion, result.APIVersion, result.Kind)
//			apiVersionKindCache[apiVersionKind] = isSupported
//		}
//		isSupported = apiVersionKindCache[apiVersionKind]
//		result.Deleted = !isSupported
//		return result
//	}
//}
