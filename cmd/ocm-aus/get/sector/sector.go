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

package sector

import (
	"encoding/json"
	"os"

	"github.com/app-sre/aus-cli/pkg/backend"
	"github.com/app-sre/aus-cli/pkg/output"
	"github.com/spf13/cobra"
)

var args struct {
	organizationId string
}

var Cmd = &cobra.Command{
	Use:   "sectors",
	Short: "Lists the sector configuration for an organization",
	Long:  "Lists the sector configuration for an organization",
	RunE:  run,
}

func init() {
	cmdFlags := Cmd.Flags()
	cmdFlags.StringVarP(
		&args.organizationId,
		"org-id",
		"o",
		"",
		"The ID of the OCM organization to inspect",
	)
}

func run(cmd *cobra.Command, argv []string) error {
	backendType, err := cmd.Flags().GetString("backend")
	if err != nil {
		return err
	}
	be, err := backend.NewPolicyBackend(backendType)
	if err != nil {
		return err
	}
	sectorConfiguration, err := be.ListSectorConfiguration(args.organizationId)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(sectorConfiguration)
	return output.PrettyList(os.Stdout, body)
}
