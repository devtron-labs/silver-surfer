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

package main

import (
	"crypto/tls"
	"fmt"
	"github.com/devtron-labs/deprecation-checker/kubedd"
	"github.com/devtron-labs/deprecation-checker/pkg"
	log2 "github.com/devtron-labs/deprecation-checker/pkg/log"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version                   = "dev"
	commit                    = "none"
	date                      = "unknown"
	directories               = make([]string, 0)
	ignoredPathPatterns       = make([]string, 0)
	kubeconfig                = ""
	kubecontext               = ""
	ignoreKeysFromDeprecation = make([]string, 0)
	ignoreKeysFromValidation  = make([]string, 0)
	selectNamespaces          = make([]string, 0)
	ignoreNamespaces          = make([]string, 0)
	selectKinds               = make([]string, 0)
	ignoreKinds               = make([]string, 0)
	// forceColor tells kubedd to use colored output even if
	// stdout is not a TTY
	forceColor bool

	config = pkg.NewDefaultConfig()
)

/*
Deleted - Latest Version
Deprecated - Current Version Latest Version
Newer - Current Version Latest Version
Unchanged - Current Version

Field Check
Deprecated Invalid


extenstion/V1alha1 deployment - removed
apps/v1 deployment - check



*/

// RootCmd represents the the command to run when kubedd is run
var RootCmd = &cobra.Command{
	Short:   "ValidateJson a Kubernetes YAML file against the relevant apiVersion and kind",
	Long:    `ValidateJson a Kubernetes YAML file against the relevant apiVersion and kind, in case the apiVersion for the kind is deprecated or removed then it validates against the latest available apiVersion`,
	Version: fmt.Sprintf("Version: %s\nCommit: %s\nDate: %s\n", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		if config.IgnoreMissingSchemas && !config.Quiet {
			log2.Warn("Set to ignore missing schemas")
		}

		// This is not particularly secure but we highlight that with the name of
		// the config item. It would be good to also support a configurable set of
		// trusted certificate authorities as in the `--certificate-authority`
		// kubectl option.
		if config.InsecureSkipTLSVerify {
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		success := true

		// Assert that colors will definitely be used if requested
		if forceColor {
			color.NoColor = false
		}

		//if len(args) < 1 && len(directories) < 1 && len(kubeconfig) < 1 {
		//	log.Error(errors.New("at least one file or one directory or kubeconfig path should be passed as argument"))
		//	os.Exit(1)
		//}
		if len(args) > 0 || len(directories) > 0 {
			success = processFiles(args)
		} else {
			processCluster()
		}

		if !success {
			os.Exit(1)
		}
	},
}

func processFiles(args []string) bool {
	success := true
	outputManager := pkg.GetOutputManager(config.OutputFormat)
	files, err := aggregateFiles(args)
	if err != nil {
		log.Error(err)
		success = false
	}

	var aggResults []pkg.ValidationResult
	for _, fileName := range files {
		filePath, _ := filepath.Abs(fileName)
		fileContents, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Error(fmt.Errorf("Could not open file %v", fileName))
			earlyExit()
			success = false
			continue
		}
		config.FileName = fileName
		results, err := kubedd.Validate(fileContents, config)
		if err != nil {
			log.Error(err)
			earlyExit()
			success = false
			continue
		}

		fmt.Println("")
		fmt.Printf("Results for file %s\n", fileName)
		fmt.Println("-------------------------------------------")
		results = removeIgnoredKeys(results)
		outputManager.PutBulk(results)

		aggResults = append(aggResults, results...)
	}

	// only use result of hasErrors check if `success` is currently truthy
	success = success && !hasErrors(aggResults)

	// flush any final logs which may be sitting in the buffer
	err = outputManager.Flush()
	if err != nil {
		log.Error(err)
		success = false
	}
	return success
}

func processCluster() bool {
	success := true
	outputManager := pkg.GetOutputManager(config.OutputFormat)
	cluster := pkg.NewCluster(kubeconfig, kubecontext)
	results, err := kubedd.ValidateCluster(cluster, config)
	if err != nil {
		log.Error(err)
		earlyExit()
		success = false
		return success
	}

	serverVersion, _ := cluster.ServerVersion()
	fmt.Println("")
	fmt.Printf("Results for cluster at version %s to %s\n", serverVersion, config.TargetKubernetesVersion)
	fmt.Println("-------------------------------------------")
	results = removeIgnoredKeys(results)
	outputManager.PutBulk(results)

	//aggResults = append(aggResults, results...)
	success = success && !hasErrors(results)
	err = outputManager.Flush()
	if err != nil {
		log.Error(err)
		success = false
	}
	return success
}

func removeIgnoredKeys(results []pkg.ValidationResult) []pkg.ValidationResult {
	var out []pkg.ValidationResult
	for _, result := range results {
		if len(result.DeprecationForOriginal) > 0 {
			var depErr []*pkg.SchemaError
			for _, schemaError := range result.DeprecationForOriginal {
				key := strings.Join(schemaError.JSONPointer(), "/")
				if !pkg.Contains(key, ignoreKeysFromDeprecation) {
					depErr = append(depErr, schemaError)
				}
			}
			result.DeprecationForOriginal = depErr
		}
		if len(result.DeprecationForLatest) > 0 {
			var depErr []*pkg.SchemaError
			for _, schemaError := range result.DeprecationForLatest {
				key := strings.Join(schemaError.JSONPointer(), "/")
				if !pkg.Contains(key, ignoreKeysFromDeprecation) {
					depErr = append(depErr, schemaError)
				}
			}
			result.DeprecationForLatest = depErr
		}
		if len(result.ErrorsForOriginal) > 0 {
			var valErr []*openapi3.SchemaError
			for _, schemaError := range result.ErrorsForOriginal {
				key := strings.Join(schemaError.JSONPointer(), "/")
				if !pkg.Contains(key, ignoreKeysFromValidation) {
					valErr = append(valErr, schemaError)
				}
			}
			result.ErrorsForOriginal = valErr
		}
		if len(result.ErrorsForLatest) > 0 {
			var valErr []*openapi3.SchemaError
			for _, schemaError := range result.ErrorsForLatest {
				key := strings.Join(schemaError.JSONPointer(), "/")
				if !pkg.Contains(key, ignoreKeysFromValidation) {
					valErr = append(valErr, schemaError)
				}
			}
			result.ErrorsForLatest = valErr
		}
		out = append(out, result)
	}
	return out
}

// hasErrors returns truthy if any of the provided results
// contain errors.
func hasErrors(res []pkg.ValidationResult) bool {
	for _, r := range res {
		if len(r.ErrorsForOriginal) > 0 || len(r.ErrorsForLatest) > 0 {
			return true
		}
	}
	return false
}

// isIgnored returns whether the specified filename should be ignored.
func isIgnored(path string) (bool, error) {
	for _, p := range ignoredPathPatterns {
		m, err := regexp.MatchString(p, path)
		if err != nil {
			return false, err
		}
		if m {
			return true, nil
		}
	}
	return false, nil
}

func aggregateFiles(args []string) ([]string, error) {
	files := make([]string, len(args))
	copy(files, args)

	var allErrors *multierror.Error
	for _, directory := range directories {
		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			ignored, err := isIgnored(path)
			if err != nil {
				return err
			}
			if !info.IsDir() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) && !ignored {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			allErrors = multierror.Append(allErrors, err)
		}
	}

	return files, allErrors.ErrorOrNil()
}

func earlyExit() {
	if config.ExitOnError {
		os.Exit(1)
	}
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

func init() {
	rootCmdName := filepath.Base(os.Args[0])
	if strings.HasPrefix(rootCmdName, "kubectl-") {
		rootCmdName = strings.Replace(rootCmdName, "-", " ", 1)
	}
	RootCmd.Use = fmt.Sprintf("%s <file> [file...]", rootCmdName)
	pkg.AddKubeaddFlags(RootCmd, config)
	RootCmd.Flags().BoolVarP(&forceColor, "force-color", "", false, "Force colored output even if stdout is not a TTY")
	RootCmd.SetVersionTemplate(`{{.Version}}`)
	RootCmd.Flags().StringSliceVarP(&directories, "directories", "d", []string{}, "A comma-separated list of directories to recursively search for YAML documents")
	RootCmd.Flags().StringSliceVarP(&ignoredPathPatterns, "ignored-path-patterns", "i", []string{}, "A comma-separated list of regular expressions specifying paths to ignore")
	RootCmd.Flags().StringSliceVarP(&ignoredPathPatterns, "ignored-filename-patterns", "", []string{}, "An alias for ignored-path-patterns")
	RootCmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "", "", "Path of kubeconfig file of cluster to be scanned")
	RootCmd.Flags().StringVarP(&kubecontext, "kubecontext", "", "", "Kubecontext to be selected")

	viper.SetEnvPrefix("KUBEADD")
	viper.AutomaticEnv()
	viper.BindPFlag("schema_location", RootCmd.Flags().Lookup("schema-location"))
	viper.BindPFlag("filename", RootCmd.Flags().Lookup("filename"))
}

func main() {
	Execute()
}
