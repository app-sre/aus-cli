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

package status

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/backend"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/ocm"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/output"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/policy"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/sectors"
)

var args struct {
	organizationId string
}

var Cmd = &cobra.Command{
	Use:   "status",
	Short: "Describe an organization",
	Long:  "Describe an organization with all clusters and policies",
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
	organization, err := ocm.GetOrganization(args.organizationId, connection)
	if err != nil {
		return err
	}
	blockedVersions, err := be.ListBlockedVersionExpressions(args.organizationId)
	if err != nil {
		return err
	}
	policies, err := be.ListPolicies(args.organizationId, true)
	if err != nil {
		return err
	}
	policiesList := []policy.ClusterUpgradePolicy{}
	for _, p := range policies {
		policiesList = append(policiesList, p)
	}
	policy.SortPolicies(policiesList)
	sectorsConfigs, err := be.ListSectorConfiguration(args.organizationId)
	if err != nil {
		return err
	}
	sectorsConfigs = sectors.AddMissingSectors(sectorsConfigs)

	// layout data
	description, err := output.TabbedString(func(out io.Writer) error {
		w := output.NewPrefixWriter(out, "")
		w1 := output.NewPrefixWriter(out, "  ")
		w.WriteString("Organization ID:\t%s\n", organization.ID())
		w.WriteString("Organization name:\t%s\n", organization.Name())
		w.WriteString("OCM environment:\t%s\n", connection.URL())
		output.PrintListMultiline(w, "Blocked Versions", blockedVersions)

		w.WriteString("Sector Configuration:\t(%d in total)\n", len(sectorsConfigs))
		if len(sectorsConfigs) > 0 {
			w1.WriteString("Name\tDepends on\n")
			w1.WriteString("----\t----------\n")
			for _, sector := range sectorsConfigs {
				w1.WriteString("%s\t%s\n", sector.Name, strings.Join(sector.Dependencies, ", "))
			}
		}

		w.WriteString("Clusters:\t(%d in total)\n", len(policiesList))
		if len(policiesList) > 0 {
			w1.WriteString("Cluster Name\tAUS enabled\tSchedule\tSector\tMutexes\tSoak Days\tWorkloads\n")
			w1.WriteString("------------\t-----------\t--------\t------\t-------\t---------\t---------\n")
			for _, policy := range policiesList {
				mutexes := "<none>"
				if len(policy.Conditions.Mutexes) > 0 {
					mutexes = strings.Join(policy.Conditions.Mutexes, ", ")
				}
				sector := "<none>"
				if policy.Conditions.Sector != "" {
					sector = policy.Conditions.Sector
				}
				if policy.Schedule != "" {
					w1.WriteString("%s\t%t\t%s\t%s\t%s\t%d\t%s\n",
						policy.ClusterName,
						true,
						policy.Schedule,
						sector,
						mutexes,
						policy.Conditions.SoakDays,
						strings.Join(policy.Workloads, ", "),
					)
				} else {
					w1.WriteString("%s\t%t\t%s\t%s\t%s\t%s\t%s\n",
						policy.ClusterName,
						false,
						"<none>",
						sector,
						mutexes,
						"<none>",
						"<none>",
					)
				}
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
