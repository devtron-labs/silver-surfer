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
	"math"
	"regexp"
	"strconv"
	"strings"

	kLog "github.com/devtron-labs/silver-surfer/pkg/log"
)

const (
	gvFormat  = "%s/%s"
	gvkFormat = "%s/%s/%s"
)

func getKeyForGV(msg json.RawMessage) (string, error) {
	m, err := parseGVK(msg)
	if err != nil {
		return "", err
	}
	if len(m["group"]) == 0 {
		return m["version"], nil
	}
	return fmt.Sprintf(gvFormat, m["group"], m["version"]), nil
}

func getKeyForGVK(msg json.RawMessage) (string, error) {
	gvk, err := parseGVK(msg)
	if err != nil {
		return "", err
	}
	if g, ok := gvk["group"]; ok && len(g) > 0 {
		return strings.ToLower(fmt.Sprintf(gvkFormat, gvk["group"], gvk["version"], gvk["kind"])), nil
	}
	return strings.ToLower(fmt.Sprintf(gvFormat, gvk["version"], gvk["kind"])), nil
}

func parseGVK(msg json.RawMessage) (map[string]string, error) {
	var arr []map[string]string
	err := json.Unmarshal(msg, &arr)
	if err == nil {
		if len(arr) > 1 {
			//fmt.Printf("len >1 for %v\n", arr)
			return nil, fmt.Errorf("multiple x-kubernetes-group-version-kind hence skipping")
		}
		if len(arr) > 0 {
			return arr[0], nil
		}
	}
	var m map[string]string
	err = json.Unmarshal(msg, &m)
	if err == nil {
		return m, nil
	}
	return nil, fmt.Errorf("parsing error")
}

//func getLargerVersion(lhs, rhs string) string {
//	if compareVersion(lhs, rhs) {
//		return rhs
//	} else {
//		return lhs
//	}
//}

func compareVersion(lhs, rhs string) bool {
	if lhs == rhs {
		return false
	}
	if !isExtension(lhs) && isExtension(rhs) {
		return false
	}
	if isExtension(lhs) && !isExtension(rhs) {
		return true
	}

	isSmaller, err := isSmallerVersion(lhs, rhs)
	if err != nil {
		kLog.Debug(fmt.Sprintf("%v", err))
		return false
	}
	return isSmaller
}

func isExtension(second string) bool {
	return strings.Contains(second, "extensions")
}

func isSmallerVersion(lhs, rhs string) (bool, error) {
	var re *regexp.Regexp
	var err error
	re, err = regexp.Compile(`v(\d*)([^0-9]*)(\d*)`)
	if err != nil {
		return false, err
	}
	lhsMatch := re.FindAllStringSubmatch(lhs, -1)
	rhsMatch := re.FindAllStringSubmatch(rhs, -1)
	lhsMajorVersion, err := getMajorVersion(lhsMatch)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return false, err
	}
	rhsMajorVersion, err := getMajorVersion(rhsMatch)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return false, err
	}
	if lhsMajorVersion < rhsMajorVersion {
		return true, nil
	}
	if lhsMajorVersion > rhsMajorVersion {
		return false, nil
	}

	lhsVersionType := getVersionType(lhs)
	rhsVersionType := getVersionType(rhs)
	if lhsVersionType < rhsVersionType {
		return true, nil
	}
	if lhsVersionType > rhsVersionType {
		return false, nil
	}

	lhsMinorVersion, err := getMinorVersion(lhsMatch)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return false, err
	}
	rhsMinorVersion, err := getMinorVersion(rhsMatch)
	if err != nil {
		//kLog.Debug(fmt.Sprintf("%v", err))
		return false, err
	}

	if lhsMinorVersion <= rhsMinorVersion {
		return true, nil
	}
	if lhsMajorVersion > rhsMajorVersion {
		return false, nil
	}
	return false, nil
}

func getMajorVersion(apiVersion [][]string) (int, error) {
	majorVersion := apiVersion[0][1]
	return strconv.Atoi(majorVersion)
}

func getMinorVersion(apiVersion [][]string) (int, error) {
	minorVersion := apiVersion[0][3]
	if len(minorVersion) == 0 {
		return math.MaxInt32, nil
	}
	return strconv.Atoi(minorVersion)
}

func getVersionType(apiVersion string) int {
	if strings.Index(apiVersion, "alpha") > 0 {
		return alphaVersion
	} else if strings.Index(apiVersion, "beta") > 0 {
		return betaVersion
	}
	return gaVersion
}

func Contains(key string, patterns []string) bool {
	for _, ignoreKey := range patterns {
		if strings.EqualFold(ignoreKey, key) {
			return true
		}
		if RegexMatch(key, ignoreKey) {
			return true
		}
	}
	return false
}

func RegexMatch(s string, pattern string) bool {
	ls := strings.ToLower(s)
	lp := strings.ToLower(pattern)
	if !strings.Contains(lp, "*") {
		return ls == lp
	}
	if strings.Count(lp, "*") == 2 {
		np := strings.ReplaceAll(lp, "*", "")
		return strings.Contains(ls, np)
	}
	if strings.Index(lp, "*") == 0 {
		np := strings.ReplaceAll(lp, "*", "")
		return strings.HasSuffix(ls, np)
	}
	np := strings.ReplaceAll(lp, "*", "")
	return strings.HasPrefix(ls, np)
}
