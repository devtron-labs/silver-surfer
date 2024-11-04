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
 * Certain portions in this file have been taken from kubeval and where ever
 * they are, IP and licenses of kubeval are applicable.
 */

package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mgutz/ansi"
	"sort"
	"strings"

	//"github.com/olekukonko/tablewriter"
	"github.com/tomlazar/table"
	"log"
	"os"
)

// OutputManager controls how results of the `kubedd` evaluation will be recorded
// and reported to the end user.
// This interface is kept private to ensure all implementations are closed within
// this package.
type OutputManager interface {
	PutBulk(r []ValidationResult) error
	Put(r ValidationResult) error
	Flush() error
	GetSummaryValidationResultBulk() []SummaryValidationResult
}

const (
	outputSTD  = "stdout"
	outputJSON = "json"
	outputTAP  = "tap"
)

var (
	hiWhite = color.New(color.FgWhite, color.Underline).SprintFunc()
	green   = color.New(color.FgHiGreen, color.Underline).SprintFunc()
)

func validOutputs() []string {
	return []string{
		outputSTD,
		outputJSON,
		outputTAP,
	}
}

func GetOutputManager(outFmt string, noColor bool) OutputManager {
	switch outFmt {
	case outputSTD:
		return newSTDOutputManager(noColor)
	case outputJSON:
		return newDefaultJSONOutputManager()
	case outputTAP:
		return newDefaultTAPOutputManager()
	default:
		return newSTDOutputManager(noColor)
	}
}

// STDOutputManager reports `kubedd` results to stdout.
type STDOutputManager struct {
	noColor bool
}

// newSTDOutputManager instantiates a new instance of STDOutputManager.
func newSTDOutputManager(noColor bool) *STDOutputManager {
	return &STDOutputManager{noColor}
}

func (s *STDOutputManager) PutBulk(results []ValidationResult) error {
	if len(results) == 0 {
		return nil
	}
	var deleted []ValidationResult
	var deprecated []ValidationResult
	var newerVersion []ValidationResult
	var unchanged []ValidationResult

	for _, result := range results {
		if len(result.Kind) == 0 {
			continue
		} else if result.Deleted {
			deleted = append(deleted, result)
			/*} else if result.Deprecated && len(result.LatestAPIVersion) > 0 {
			deprecated = append(deprecated, result)*/
		} else if result.Deprecated {
			deprecated = append(deprecated, result)
		} else if len(result.LatestAPIVersion) > 0 {
			newerVersion = append(newerVersion, result)
		} else {
			if len(result.ErrorsForOriginal) == 0 && len(result.ErrorsForLatest) == 0 &&
				len(result.DeprecationForOriginal) == 0 && len(result.DeprecationForLatest) == 0 {
				unchanged = append(unchanged, result)
			}
		}
	}
	if len(deleted) > 0 {
		sort.Slice(deleted, func(i, j int) bool {
			return len(deleted[i].ErrorsForLatest) > len(deleted[j].ErrorsForLatest)
		})
		color.NoColor = false
		red := color.New(color.FgHiRed, color.Underline).SprintFunc()
		if s.noColor {
			color.NoColor = true
		}
		fmt.Printf("%s\n", red(">>>> Removed API Version's <<<<"))
		s.SummaryTableBodyOutput(deleted)
		fmt.Println("")
		s.ValidationErrorTableBodyOutput(deleted, false)
		s.DeprecationTableBodyOutput(deleted, false)
	}
	if len(deprecated) > 0 {
		sort.Slice(deprecated, func(i, j int) bool {
			return len(deprecated[i].ErrorsForLatest) > len(deprecated[j].ErrorsForLatest)
		})
		yellow := color.New(color.FgHiYellow, color.Underline).SprintFunc()
		fmt.Printf("%s\n", yellow(">>>> Deprecated API Version's <<<<"))
		s.SummaryTableBodyOutput(deprecated)
		fmt.Println("")
		s.DeprecationTableBodyOutput(deprecated, true)
		s.ValidationErrorTableBodyOutput(deprecated, true)
		s.DeprecationTableBodyOutput(deprecated, false)
		s.ValidationErrorTableBodyOutput(deprecated, false)
	}
	if len(newerVersion) > 0 {
		sort.Slice(newerVersion, func(i, j int) bool {
			return len(newerVersion[i].ErrorsForLatest) > len(newerVersion[j].ErrorsForLatest)
		})
		yellow := color.New(color.FgHiYellow, color.Underline).SprintFunc()
		fmt.Printf("%s\n", yellow(">>>> Newer Versions available <<<<"))
		s.SummaryTableBodyOutput(newerVersion)
		fmt.Println("")
		s.DeprecationTableBodyOutput(newerVersion, true)
		s.ValidationErrorTableBodyOutput(newerVersion, true)
		s.DeprecationTableBodyOutput(newerVersion, false)
		s.ValidationErrorTableBodyOutput(newerVersion, false)
	}
	if len(unchanged) > 0 {

		fmt.Printf("%s\n", green(">>>> Unchanged API Version's <<<<"))
		//s.SummaryTableBodyOutput(unchanged)
		fmt.Println("")
		s.DeprecationTableBodyOutput(unchanged, true)
		s.ValidationErrorTableBodyOutput(unchanged, true)
	}

	if len(deleted)+len(deprecated)+len(newerVersion)+len(unchanged) == 0 {
		fmt.Printf("%s\n", green("Great!!! Everything will work as it is in new version without any changes"))
	}
	return nil
}

func (s *STDOutputManager) SummaryTableBodyOutput(results []ValidationResult) {
	t := table.Table{Headers: []string{"Namespace", "Name", "Kind", "API Version (Current Available)", "Replace With API Version (Latest Available)", "Migration Status"}}
	c := table.DefaultConfig()
	c.TitleColorCode = ansi.ColorCode("cyan+bu")
	c.AltColorCodes = []string{ansi.LightWhite, ansi.ColorCode("white+h:238")}
	c.ShowIndex = false
	for _, result := range results {
		migrationStatus := "can be migrated with just apiVersion change"
		if len(result.ErrorsForLatest) > 0 {
			migrationStatus = fmt.Sprintf("%s%d%s%s%s", "\033[31m", len(result.ErrorsForLatest), " issue(s):", "\033[97m", " fix issues before migration")
		}
		if result.IsVersionSupported == 2 {
			migrationStatus = fmt.Sprintf("%s%s", "\033[31m", fmt.Sprintf("Alert! cannot migrate kubernetes version"))
		}
		t.Rows = append(t.Rows, []string{result.ResourceNamespace, result.ResourceName, result.Kind, result.APIVersion, result.LatestAPIVersion, migrationStatus})
	}
	c.Color = !s.noColor
	t.WriteTable(os.Stdout, c)
}

func (s *STDOutputManager) DeprecationTableBodyOutput(results []ValidationResult, currentVersion bool) {
	hasData := false
	for _, result := range results {
		errors := result.DeprecationForLatest
		if currentVersion {
			errors = result.DeprecationForOriginal
		}
		if len(errors) > 0 {
			for _, e := range errors {
				if len(e.JSONPointer()) > 0 {
					hasData = true
					break
				}
			}
		}
	}
	if !hasData {
		return
	}
	if !currentVersion {
		fmt.Println(hiWhite("Deprecated fields against latest api version, recommended to resolve them before migration"))
	} else {
		fmt.Println(hiWhite("Deprecated fields against current api version, recommended to resolve them"))
	}
	apiVersionHeader := "API Version (Current Available)"
	if !currentVersion {
		apiVersionHeader = "API Version (Latest Available)"
	}
	t := table.Table{Headers: []string{"Namespace", "Name", "Kind", apiVersionHeader, "Field", "Reason"}}
	c := table.DefaultConfig()
	c.TitleColorCode = ansi.ColorCode("cyan+bu")
	c.AltColorCodes = []string{ansi.LightWhite, ansi.ColorCode("white+h:237")}
	c.ShowIndex = false
	for _, result := range results {
		errors := result.DeprecationForLatest
		apiVersion := result.LatestAPIVersion
		if currentVersion {
			apiVersion = result.APIVersion
			errors = result.DeprecationForOriginal
		}
		for _, e := range errors {
			t.Rows = append(t.Rows, []string{result.ResourceNamespace, result.ResourceName, result.Kind, apiVersion, strings.Join(e.JSONPointer(), "/"), e.Reason})
		}
	}
	c.Color = !s.noColor
	t.WriteTable(os.Stdout, c)
	fmt.Println("")
}

func (s *STDOutputManager) ValidationErrorTableBodyOutput(results []ValidationResult, currentVersion bool) {
	hasData := false
	for _, result := range results {
		errors := result.ErrorsForLatest
		if currentVersion {
			errors = result.ErrorsForOriginal
		}
		if len(errors) > 0 {
			for _, e := range errors {
				if len(e.JSONPointer()) > 0 {
					hasData = true
					break
				}
			}
		}
	}
	if !hasData {
		return
	}
	if !currentVersion {
		fmt.Println(hiWhite(">>> Validation Errors against latest api version, should be resolved before migration <<<"))
	} else {
		fmt.Println(hiWhite(">>> Validation Errors against current api version <<<"))
	}
	apiVersionHeader := "API Version (Current Available)"
	if !currentVersion {
		apiVersionHeader = "API Version (Latest Available)"
	}
	t := table.Table{Headers: []string{"Namespace", "Name", "Kind", apiVersionHeader, "Field", "Reason"}}
	c := table.DefaultConfig()
	c.TitleColorCode = ansi.ColorCode("cyan+bu")
	c.AltColorCodes = []string{ansi.LightWhite, ansi.ColorCode("white+h:237")}
	c.ShowIndex = false
	for _, result := range results {
		errors := result.ErrorsForLatest
		apiVersion := result.LatestAPIVersion
		if currentVersion {
			apiVersion = result.APIVersion
			errors = result.ErrorsForOriginal
		}
		for _, e := range errors {
			if len(e.JSONPointer()) > 0 {
				t.Rows = append(t.Rows, []string{result.ResourceNamespace, result.ResourceName, result.Kind, apiVersion, strings.Join(e.JSONPointer(), "/"), e.Reason})
			}
		}
	}
	c.Color = !s.noColor
	t.WriteTable(os.Stdout, c)
	fmt.Println("")
}

func (s *STDOutputManager) Put(result ValidationResult) error {
	openapi3.SchemaErrorDetailsDisabled = true
	return nil
}

func (s *STDOutputManager) Flush() error {
	// no op
	return nil
}

func (j *STDOutputManager) GetSummaryValidationResultBulk() []SummaryValidationResult {
	return nil
}

type status string

const (
	statusInvalid = "invalid"
	statusValid   = "valid"
	statusSkipped = "skipped"
)

type dataEvalResult struct {
	Filename string   `json:"filename"`
	Kind     string   `json:"kind"`
	Status   status   `json:"status"`
	Errors   []string `json:"errors"`
}

// jsonOutputManager reports `ccheck` results to `stdout` as a json array..
type jsonOutputManager struct {
	logger *log.Logger

	data []SummaryValidationResult
}

func newDefaultJSONOutputManager() *jsonOutputManager {
	return newJSONOutputManager(log.New(os.Stdout, "", 0))
}

func newJSONOutputManager(l *log.Logger) *jsonOutputManager {
	return &jsonOutputManager{
		logger: l,
	}
}

func getStatus(r ValidationResult) status {
	if r.Kind == "" {
		return statusSkipped
	}

	if !r.ValidatedAgainstSchema {
		return statusSkipped
	}

	if len(r.Errors) > 0 {
		return statusInvalid
	}

	return statusValid
}

func (j *jsonOutputManager) PutBulk(vrs []ValidationResult) error {
	svrs := make([]SummaryValidationResult, len(vrs))
	for _, vr := range vrs {
		if vr.Deleted == false && vr.Deprecated == false && len(vr.ErrorsForLatest) == 0 && len(vr.ErrorsForOriginal) == 0 && len(vr.DeprecationForLatest) == 0 && len(vr.DeprecationForOriginal) == 0 {
			continue
		}
		svr := SummaryValidationResult{
			Deleted:            vr.Deleted,
			Deprecated:         vr.Deprecated,
			Kind:               vr.Kind,
			ResourceName:       vr.ResourceName,
			APIVersion:         vr.APIVersion,
			FileName:           vr.FileName,
			IsVersionSupported: vr.IsVersionSupported,
			LatestAPIVersion:   vr.LatestAPIVersion,
		}
		for _, se := range vr.ErrorsForOriginal {
			sse := &SummarySchemaError{
				Path:        strings.Join(se.JSONPointer(), "/"),
				SchemaField: se.SchemaField,
				Reason:      se.Reason,
				Origin:      se.Origin,
			}
			svr.ErrorsForOriginal = append(svr.ErrorsForOriginal, sse)
		}
		for _, se := range vr.ErrorsForLatest {
			sse := &SummarySchemaError{
				Path:        strings.Join(se.JSONPointer(), "/"),
				SchemaField: se.SchemaField,
				Reason:      se.Reason,
				Origin:      se.Origin,
			}
			svr.ErrorsForLatest = append(svr.ErrorsForLatest, sse)
		}
		for _, se := range vr.DeprecationForOriginal {
			sse := &SummarySchemaError{
				Path:        strings.Join(se.JSONPointer(), "/"),
				SchemaField: se.SchemaField,
				Reason:      se.Reason,
				Origin:      se.Origin,
			}
			svr.DeprecationForOriginal = append(svr.DeprecationForOriginal, sse)
		}
		for _, se := range vr.DeprecationForLatest {
			sse := &SummarySchemaError{
				Path:        strings.Join(se.JSONPointer(), "/"),
				SchemaField: se.SchemaField,
				Reason:      se.Reason,
				Origin:      se.Origin,
			}
			svr.DeprecationForLatest = append(svr.DeprecationForLatest, sse)
		}
		svrs = append(svrs, svr)
	}
	j.data = svrs
	return nil
}

func (j *jsonOutputManager) Put(vr ValidationResult) error {
	// stringify gojsonschema errors
	// use a pre-allocated slice to ensure the json will have an
	// empty array in the "zero" case
	//errs := make([]string, 0, len(r.Errors))
	//for _, e := range r.Errors {
	//	errs = append(errs, e.String())
	//}

	svr := SummaryValidationResult{
		Deleted:            vr.Deleted,
		Deprecated:         vr.Deprecated,
		Kind:               vr.Kind,
		ResourceName:       vr.ResourceName,
		APIVersion:         vr.APIVersion,
		FileName:           vr.FileName,
		IsVersionSupported: vr.IsVersionSupported,
		LatestAPIVersion:   vr.LatestAPIVersion,
	}
	for _, se := range vr.ErrorsForOriginal {
		sse := &SummarySchemaError{
			Path:        strings.Join(se.JSONPointer(), "/"),
			SchemaField: se.SchemaField,
			Reason:      se.Reason,
			Origin:      se.Origin,
		}
		svr.ErrorsForOriginal = append(svr.ErrorsForOriginal, sse)
	}
	for _, se := range vr.ErrorsForLatest {
		sse := &SummarySchemaError{
			Path:        strings.Join(se.JSONPointer(), "/"),
			SchemaField: se.SchemaField,
			Reason:      se.Reason,
			Origin:      se.Origin,
		}
		svr.ErrorsForLatest = append(svr.ErrorsForLatest, sse)
	}
	for _, se := range vr.DeprecationForOriginal {
		sse := &SummarySchemaError{
			Path:        strings.Join(se.JSONPointer(), "/"),
			SchemaField: se.SchemaField,
			Reason:      se.Reason,
			Origin:      se.Origin,
		}
		svr.DeprecationForOriginal = append(svr.DeprecationForOriginal, sse)
	}
	for _, se := range vr.DeprecationForLatest {
		sse := &SummarySchemaError{
			Path:        strings.Join(se.JSONPointer(), "/"),
			SchemaField: se.SchemaField,
			Reason:      se.Reason,
			Origin:      se.Origin,
		}
		svr.DeprecationForLatest = append(svr.DeprecationForLatest, sse)
	}

	j.data = append(j.data, svr)

	return nil
}

func (j *jsonOutputManager) Flush() error {
	b, err := json.Marshal(j.data)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	j.logger.Print(out.String())
	return nil
}

func (j *jsonOutputManager) GetSummaryValidationResultBulk() []SummaryValidationResult {
	return j.data
}

// tapOutputManager reports `conftest` results to stdout.
type tapOutputManager struct {
	logger *log.Logger

	data []dataEvalResult
}

// newDefaultTapOutManager instantiates a new instance of tapOutputManager
// using the default logger.
func newDefaultTAPOutputManager() *tapOutputManager {
	return newTAPOutputManager(log.New(os.Stdout, "", 0))
}

// newTapOutputManager constructs an instance of tapOutputManager given a
// logger instance.
func newTAPOutputManager(l *log.Logger) *tapOutputManager {
	return &tapOutputManager{
		logger: l,
	}
}

func (j *tapOutputManager) PutBulk(r []ValidationResult) error {
	return nil
}

func (j *tapOutputManager) Put(r ValidationResult) error {
	errs := make([]string, 0, len(r.Errors))
	for _, e := range r.Errors {
		errs = append(errs, e.String())
	}

	j.data = append(j.data, dataEvalResult{
		Filename: r.FileName,
		Kind:     r.Kind,
		Status:   getStatus(r),
		Errors:   errs,
	})

	return nil
}

func (j *tapOutputManager) Flush() error {
	issues := len(j.data)
	if issues > 0 {
		total := 0
		for _, r := range j.data {
			if len(r.Errors) > 0 {
				total = total + len(r.Errors)
			} else {
				total = total + 1
			}
		}
		j.logger.Print(fmt.Sprintf("1..%d", total))
		count := 0
		for _, r := range j.data {
			count = count + 1
			var kindMarker string
			if r.Kind == "" {
				kindMarker = ""
			} else {
				kindMarker = fmt.Sprintf(" (%s)", r.Kind)
			}
			if r.Status == "valid" {
				j.logger.Print("ok ", count, " - ", r.Filename, kindMarker)
			} else if r.Status == "skipped" {
				j.logger.Print("ok ", count, " - ", r.Filename, kindMarker, " # SKIP")
			} else if r.Status == "invalid" {
				for i, e := range r.Errors {
					j.logger.Print("not ok ", count, " - ", r.Filename, kindMarker, " - ", e)

					// We have to skip adding 1 if it's the last error
					if len(r.Errors) != i+1 {
						count = count + 1
					}
				}
			}
		}
	}
	return nil
}

func (j *tapOutputManager) GetSummaryValidationResultBulk() []SummaryValidationResult {
	return nil
}
