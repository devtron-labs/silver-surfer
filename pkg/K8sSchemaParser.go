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
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tidwall/sjson"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	urlTemplate       = `https://raw.githubusercontent.com/kubernetes/kubernetes/release-%s/api/openapi-spec/swagger.json`
	intOrStringPath   = "components.schemas.io\\.k8s\\.apimachinery\\.pkg\\.util\\.intstr\\.IntOrString"
	intOrStringType   = `{"oneOf":[{"type": "string"},{"type": "integer"}]}`
	intOrStringFormat = "definitions.io\\.k8s\\.apimachinery\\.pkg\\.util\\.intstr\\.IntOrString.format"
	alphaVersion      = 1
	betaVersion       = 2
	gaVersion         = 3
)

type Validator interface {
	ValidateJson(spec string, releaseVersion string) (ValidationResult, error)
	ValidateYaml(spec string, releaseVersion string) (ValidationResult, error)
	ValidateObject(spec map[string]interface{}, releaseVersion string) (ValidationResult, error)
	GetKinds(releaseVersion string) ([]schema.GroupVersionKind, error)
}

type Parser interface {
	LoadFromUrl(releaseVersion string, force bool) error
	LoadFromPath(releaseVersion string, filePath string, force bool) error
}

type KubeChecker interface {
	IsApiVersionSupported(releaseVersion, apiVersion, kind string) bool
	Parser
	Validator
}

type kubeCheckerImpl struct {
	versionMap map[string]*kubeSpec
}

func NewKubeCheckerImpl() *kubeCheckerImpl {
	return &kubeCheckerImpl{versionMap: map[string]*kubeSpec{}}
}

// Unused Function
//
// func (k *kubeCheckerImpl) hasReleaseVersion(releaseVersion string) bool {
// 	_, ok := k.versionMap[releaseVersion]
// 	return ok
// }

func (k *kubeCheckerImpl) LoadFromPath(releaseVersion string, filePath string, force bool) error {
	if _, ok := k.versionMap[releaseVersion]; ok && !force {
		return nil
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return err
	}
	return k.load(data, releaseVersion)
}

func (k *kubeCheckerImpl) LoadFromUrl(releaseVersion string, force bool) error {
	if _, ok := k.versionMap[releaseVersion]; ok && !force {
		return nil
	}
	data, err := k.downloadFile(releaseVersion)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return err
	}
	return k.load(data, releaseVersion)
}

func (k *kubeCheckerImpl) load(data []byte, releaseVersion string) error {
	openapi, err := k.loadOpenApi2(data)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return err
	}
	k.versionMap[releaseVersion] = newKubeSpec(openapi)
	return nil
}

func (k *kubeCheckerImpl) downloadFile(releaseVersion string) ([]byte, error) {
	url := fmt.Sprintf(urlTemplate, releaseVersion)
	resp, err := http.Get(url)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return []byte{}, err
	}
	defer resp.Body.Close()
	var out bytes.Buffer
	_, err = io.Copy(&out, resp.Body)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return []byte{}, err
	}
	return out.Bytes(), nil
}

func (k *kubeCheckerImpl) loadOpenApi2(data []byte) (*openapi3.T, error) {
	var err error
	stringData := string(data)
	stringData, err = sjson.Delete(stringData, intOrStringFormat)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return nil, err
	}

	ctx := context.Background()
	api := openapi2.T{}
	err = (&api).UnmarshalJSON([]byte(stringData))
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return nil, err
	}

	doc, err := openapi2conv.ToV3(&api)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return nil, err
	}

	err = doc.Validate(ctx)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return nil, err
	}

	doc3, err := doc.MarshalJSON()
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return nil, err
	}

	stringData = string(doc3)
	stringData, err = sjson.SetRaw(stringData, intOrStringPath, intOrStringType)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return nil, err
	}

	loader := &openapi3.Loader{Context: ctx}
	doc, err = loader.LoadFromData([]byte(stringData))
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return nil, err
	}

	err = doc.Validate(ctx)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return nil, err
	}
	for _, v := range doc.Components.Schemas {
		v.Value.AdditionalPropertiesAllowed = openapi3.BoolPtr(false)
	}
	return doc, nil
}

func (k *kubeCheckerImpl) ValidateYaml(spec string, releaseVersion string) (ValidationResult, error) {
	err := k.LoadFromUrl(releaseVersion, false)
	if err != nil {
		return ValidationResult{}, err
	}
	return k.versionMap[releaseVersion].ValidateYaml(spec)
}

func (k *kubeCheckerImpl) ValidateJson(spec string, releaseVersion string) (ValidationResult, error) {
	err := k.LoadFromUrl(releaseVersion, false)
	if err != nil {
		return ValidationResult{}, err
	}
	return k.versionMap[releaseVersion].ValidateJson(spec)
}

func (k *kubeCheckerImpl) ValidateObject(spec map[string]interface{}, releaseVersion string) (ValidationResult, error) {
	err := k.LoadFromUrl(releaseVersion, false)
	if err != nil {
		return ValidationResult{}, err
	}
	return k.versionMap[releaseVersion].ValidateObject(spec)
}

func (k *kubeCheckerImpl) GetKinds(releaseVersion string) ([]schema.GroupVersionKind, error) {
	err := k.LoadFromUrl(releaseVersion, false)
	if err != nil {
		return make([]schema.GroupVersionKind, 0), err
	}
	return k.versionMap[releaseVersion].getLatestKinds(), nil
}

func (k *kubeCheckerImpl) IsApiVersionSupported(releaseVersion, apiVersion, kind string) bool {
	err := k.LoadFromUrl(releaseVersion, false)
	if err != nil {
		return false
	}
	return k.versionMap[releaseVersion].isApiVersionSupported(apiVersion, kind)
}
