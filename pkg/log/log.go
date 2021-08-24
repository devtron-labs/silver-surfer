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

package log

import (
	"fmt"
	"github.com/fatih/color"
	multierror "github.com/hashicorp/go-multierror"
	"strings"
)

func Success(message ...string) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s - %v\n", green("PASS"), strings.Join(message, " "))
}

func Warn(message ...string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Printf("%s - %v\n", yellow("WARN"), strings.Join(message, " "))
}

func Error(message error) {
	if merr, ok := message.(*multierror.Error); ok {
		for _, serr := range merr.Errors {
			Error(serr)
		}
	} else {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("%s - %v\n", red("ERR "), message)
	}
}

func Debug(message ...string) {
	yellow := color.New(color.FgWhite).SprintFunc()
	fmt.Printf("%s - %v\n", yellow("DEBUG"), strings.Join(message, " "))
}
