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
	"github.com/spf13/cobra"

	"github.com/app-sre/aus-cli/pkg/backend"
)

var args struct {
	organizationId string
	clusterName    string
	dryRun         bool
}

var Cmd = &cobra.Command{
	Use:   "policy",
	Short: "Delete a policy",
	RunE:  run,
}

func init() {
	flags := Cmd.Flags()
	flags.SortFlags = false
	flags.StringVarP(
		&args.organizationId,
		"org-id",
		"o",
		"",
		"The ID of the OCM organization that owns the cluster. "+
			"Defaults to the organization of the logged in user.",
	)
	flags.StringVarP(
		&args.clusterName,
		"cluster-name",
		"c",
		"",
		"Name of the cluster that holds the policy to delete. "+
			"This name needs to match the cluster name in OCM.",
	)
	_ = Cmd.MarkFlagRequired("cluster-name")
	flags.BoolVar(
		&args.dryRun,
		"dry-run",
		false,
		"If dry-run is specified, the generated policy will only be printed to stdout.",
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

	// delete policy
	return be.DeletePolicy(args.organizationId, args.clusterName, args.dryRun)
}
