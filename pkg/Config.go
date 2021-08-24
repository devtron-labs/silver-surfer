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
	"fmt"
	"github.com/spf13/cobra"
)

// A Config object contains various configuration data for kubedd
type Config struct {
	// DefaultNamespace is the namespace to assume in resources
	// if no namespace is set in `metadata:namespace` (as used with
	// `kubectl apply --namespace ...` or `helm install --namespace ...`,
	// for example)
	DefaultNamespace string

	// TargetKubernetesVersion represents the version of Kubernetes
	// to which we want to migrate
	TargetKubernetesVersion string

	// SourceKubernetesVersion represents the version of Kubernetes
	// on which kubernetes objects are running currently
	SourceKubernetesVersion string

	// TargetSchemaLocation is the base URL of target kubernetes version.
	// It can be either a remote location or a local directory
	TargetSchemaLocation string

	// SourceSchemaLocation is the base URL of source kubernetes versions.
	// It can be either a remote location or a local directory
	SourceSchemaLocation string

	// AdditionalSchemaLocations is a list of alternative base URLs from
	// which to search for schemas, given that the desired schema was not
	// found at TargetSchemaLocation
	AdditionalSchemaLocations []string

	// Strict tells kubedd whether to prohibit properties not in
	// the schema. The API allows them, but kubectl does not
	Strict bool

	// IgnoreMissingSchemas tells kubedd whether to skip validation
	// for resource definitions without an available schema
	IgnoreMissingSchemas bool

	// ExitOnError tells kubedd whether to halt processing upon the
	// first error encountered or to continue, aggregating all errors
	ExitOnError bool

	// FileName is the name to be displayed when testing manifests read from stdin
	FileName string

	// OutputFormat is the name of the output formatter which will be used when
	// reporting results to the user.
	OutputFormat string

	// Quiet indicates whether non-results output should be emitted to the applications
	// log.
	Quiet bool

	// InsecureSkipTLSVerify controls whether to skip TLS certificate validation
	// when retrieving schema content over HTTPS
	InsecureSkipTLSVerify bool

	// IgnoreKeysFromDeprecation is the list of keys to be skipped for depreciation check
	IgnoreKeysFromDeprecation []string

	// IgnoreKeysFromValidation is the list of keys to be skipped for validation check
	IgnoreKeysFromValidation  []string

	// SelectNamespaces is the list of namespaces to be validated, by default all namespaces are validated
	SelectNamespaces          []string

	// IgnoreNamespaces is the list of namespaces to be skipped for validation, by default none are skipped
	IgnoreNamespaces          []string

	// SelectKinds is the list of kinds to be validated, by default all kinds are validated
	SelectKinds               []string

	// IgnoreKinds is the list of kinds to be skipped for validation, by default none are skipped
	IgnoreKinds               []string

	// IgnoreNullErrors is the flag to ignore null value errors
	IgnoreNullErrors 		  bool
}

// NewDefaultConfig creates a Config with default values
func NewDefaultConfig() *Config {
	return &Config{
		DefaultNamespace:        "default",
		FileName:                "stdin",
		TargetKubernetesVersion: "master",
	}
}

// AddKubeaddFlags adds the default flags for kubedd to cmd
func AddKubeaddFlags(cmd *cobra.Command, config *Config) *cobra.Command {
	cmd.Flags().StringVarP(&config.FileName, "filename", "f", "stdin", "filename to be displayed when testing manifests read from stdin")
	cmd.Flags().StringVarP(&config.TargetSchemaLocation, "target-schema-location", "", "", "TargetSchemaLocation is the base URL of target kubernetes version.")
	cmd.Flags().StringVarP(&config.SourceSchemaLocation, "source-schema-location", "", "", "SourceSchemaLocation is the base URL of source kubernetes versions.")
	cmd.Flags().StringVarP(&config.TargetKubernetesVersion, "target-kubernetes-version", "", "1.22", "Version of Kubernetes to migrate to")
	cmd.Flags().StringVarP(&config.SourceKubernetesVersion, "source-kubernetes-version", "", "", "Version of Kubernetes on which kubernetes objects are deployed currently, ignored in case cluster is provided")
	cmd.Flags().StringVarP(&config.OutputFormat, "output", "o", "", fmt.Sprintf("The format of the output of this script. Options are: %v", validOutputs()))
	cmd.Flags().BoolVar(&config.Quiet, "quiet", false, "Silences any output aside from the direct results")
	cmd.Flags().BoolVar(&config.InsecureSkipTLSVerify, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	cmd.Flags().StringSliceVarP(&config.SelectNamespaces, "select-namespaces", "", []string{}, "A comma-separated list of namespaces to be selected, if left empty all namespaces are selected")
	cmd.Flags().StringSliceVarP(&config.IgnoreNamespaces, "ignore-namespaces", "", []string{"kube-system"}, "A comma-separated list of namespaces to be skipped")
	cmd.Flags().StringSliceVarP(&config.IgnoreKinds, "ignore-kinds", "", []string{"event","CustomResourceDefinition"}, "A comma-separated list of kinds to be skipped")
	cmd.Flags().StringSliceVarP(&config.SelectKinds, "select-kinds", "", []string{}, "A comma-separated list of kinds to be selected, if left empty all namespaces are selected")
	cmd.Flags().StringSliceVarP(&config.IgnoreKeysFromDeprecation, "ignore-keys-for-deprecation", "", []string{"metadata*", "status*"}, "A comma-separated list of keys to be ignored for depreciation check")
	cmd.Flags().StringSliceVarP(&config.IgnoreKeysFromValidation, "ignore-keys-for-validation", "", []string{"status*", "metadata*"}, "A comma-separated list of keys to be ignored for validation check")
	cmd.Flags().BoolVar(&config.IgnoreNullErrors, "ignore-null-errors", true, "Ignore null value errors")

	return cmd
}
