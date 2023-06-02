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

package policy

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/app-sre/aus-cli/pkg/backend"
	"github.com/app-sre/aus-cli/pkg/output"
	"github.com/app-sre/aus-cli/pkg/policy"
)

var args struct {
	organizationId string
}

var Cmd = &cobra.Command{
	Use:   "policies",
	Short: "List cluster upgrade policies",
	RunE:  run,
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(
		&args.organizationId,
		"org-id",
		"o",
		"",
		"The ID of the OCM organization that owns the cluster. "+
			"Defaults to the organization of the logged in user.",
	)
}

func run(cmd *cobra.Command, argv []string) error {
	backendType, err := cmd.Flags().GetString("backend")
	if err != nil {
		return err
	}
	fe, err := backend.NewPolicyBackend(backendType)
	if err != nil {
		return err
	}

	// todo: autodetect organization ID if not specified
	policies, err := fe.ListPolicies(args.organizationId, false)
	if err != nil {
		return err
	}

	// build a list of policies and dump them to stdout
	policiesSlice := []policy.ClusterUpgradePolicy{}

	for _, p := range policies {
		policiesSlice = append(policiesSlice, p)
	}
	body, err := json.Marshal(policiesSlice)
	if err != nil {
		return err
	}
	return output.PrettyList(os.Stdout, body)
}
