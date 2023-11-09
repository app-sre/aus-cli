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

package gates

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/app-sre/aus-cli/pkg/backend"
	"github.com/app-sre/aus-cli/pkg/clusters"
	"github.com/app-sre/aus-cli/pkg/ocm"
	"github.com/app-sre/aus-cli/pkg/output"
	"github.com/app-sre/aus-cli/pkg/versions"
)

var args struct {
	organizationId string
}

var Cmd = &cobra.Command{
	Use:   "gates",
	Short: "List version gates",
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
	be, err := backend.NewPolicyBackend(backendType)
	if err != nil {
		return err
	}

	// assemble data
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
	clusterInfos, err := clusters.ClusterInfosForOrganization(organizationId, "", true, connection)
	if err != nil {
		return err
	}
	clusters.SortClusters(clusterInfos)

	blockedVersions, err := be.ListBlockedVersionExpressions(args.organizationId)
	if err != nil {
		return err
	}
	blockedVersionExpressions, err := versions.ParsedBlockedVersionExpressions(blockedVersions)
	if err != nil {
		return err
	}

	organization, err := ocm.GetOrganization(args.organizationId, connection)
	if err != nil {
		return err
	}

	versionGates, err := ocm.GetVersionGates(connection)
	if err != nil {
		return err
	}

	// layout data
	description, err := output.TabbedString(func(out io.Writer) error {
		w := output.NewPrefixWriter(out, "")
		w1 := output.NewPrefixWriter(out, "  ")
		w.WriteString("Organization ID:\t%s\n", organization.ID())
		w.WriteString("Organization name:\t%s\n", organization.Name())
		w.WriteString("OCM environment:\t%s\n", connection.URL())
		output.PrintListMultiline(w, "Blocked Versions", blockedVersions)
		w.WriteString("Unacknowledged version gates:\n")
		w1.WriteString("Cluster Name\tCurrent Version\tGated version\tGate Description\tGate ID\tDocumentation\n")
		w1.WriteString("------------\t---------------\t-------------\t----------------\t-------\t-------------\n")
		for _, cluster := range clusterInfos {
			if !cluster.STSEnabled() {
				continue
			}
			missingAgreements, err := cluster.MissingSTSGateAgreements(blockedVersionExpressions, versionGates)
			if err != nil {
				return err
			}
			for _, gate := range missingAgreements {
				w1.WriteString("%s\t%s\t%s\t%s\t%s\t%s\n", cluster.Cluster.Name(), cluster.Cluster.Version().RawID(), gate.VersionRawIDPrefix(), gate.Description(), gate.ID(), gate.DocumentationURL())
			}

		}

		return nil
	})
	if err != nil {
		return err
	}
	fmt.Print(description)
	return nil
}
