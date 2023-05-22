/*
Copyright (c) 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package output

import (
	"fmt"

	"gitlab.cee.redhat.com/service/aus-cli/pkg/debug"
)

func Debug(dryRun bool, format string, a ...interface{}) {
	if debug.Enabled() {
		Log(dryRun, format, a...)
	}
}

func Log(dryRun bool, format string, a ...interface{}) {
	if dryRun {
		fmt.Printf("[dry-run] "+format, a...)
	} else {
		fmt.Printf(format, a...)
	}

}
