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

package apply

import (
	"github.com/app-sre/aus-cli/cmd/ocm-aus/apply/blockedversions"
	"github.com/app-sre/aus-cli/cmd/ocm-aus/apply/gateagreement"
	"github.com/app-sre/aus-cli/cmd/ocm-aus/apply/inheritance"
	"github.com/app-sre/aus-cli/cmd/ocm-aus/apply/policy"
	"github.com/app-sre/aus-cli/cmd/ocm-aus/apply/sector"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:           "apply",
	Short:         "Apply an update to an AUS resource",
	Long:          "Apply an update to an AUS resource",
	GroupID:       "AUS commands",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Register the subcommands:
	Cmd.AddCommand(policy.Cmd)
	Cmd.AddCommand(sector.Cmd)
	Cmd.AddCommand(blockedversions.Cmd)
	Cmd.AddCommand(inheritance.Cmd)
	Cmd.AddCommand(gateagreement.Cmd)
}
