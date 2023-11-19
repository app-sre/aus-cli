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

package gateagreement

import (
	"github.com/app-sre/aus-cli/pkg/clusters"
	"github.com/app-sre/aus-cli/pkg/ocm"
	"github.com/spf13/cobra"
)

var args struct {
	organizationId string
	clusterName    string
	version        string

	dryRun bool
}

var Cmd = &cobra.Command{
	Use:   "gate-agreement [flags]",
	Short: "Create a version gate agreement",
	Long: "Create a version gate agreement.\n" +
		"\n" +
		"AUS does not approve STS version gate agreements automatically but requires " +
		"cluster owners to follow the actions described in the agreement.\n" +
		"\n" +
		"Once these actions have been completed, the cluster owner can approve the agreement " +
		"with this command.\n",
	RunE: run,
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
		"Name of the cluster to manage a policy for. "+
			"This name needs to match the cluster name in OCM.",
	)
	flags.StringVarP(
		&args.version,
		"version",
		"",
		"",
		"The y-stream version gates to agree to.",
	)
	flags.BoolVar(
		&args.dryRun,
		"dry-run",
		false,
		"If dry-run is specified, the version gate agreements will not be applied to the cluster.",
	)
}

func run(cmd *cobra.Command, argv []string) error {
	connection, err := ocm.NewOCMConnection()
	if err != nil {
		return err
	}
	organizationId := args.organizationId
	if organizationId == "" {
		organizationId, err = ocm.CurrentOrganizationId(connection)
		if err != nil {
			return err
		}
	}

	cluster, err := clusters.GetClusterInfoByName(organizationId, args.clusterName, connection)
	if err != nil {
		return err
	}
	_, err = clusters.AckAllGatesForYStream(cluster, args.version, connection, args.dryRun)
	return err
}
