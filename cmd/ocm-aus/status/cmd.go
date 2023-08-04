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

	"github.com/app-sre/aus-cli/pkg/backend"
	"github.com/app-sre/aus-cli/pkg/blockedversions"
	"github.com/app-sre/aus-cli/pkg/ocm"
	"github.com/app-sre/aus-cli/pkg/output"
	"github.com/spf13/cobra"
)

var args struct {
	organizationId  string
	showAllClusters bool
}

var Cmd = &cobra.Command{
	Use:     "status",
	Short:   "Describe an organization",
	Long:    "Describe an organization with all clusters and policies",
	GroupID: "AUS commands",
	RunE:    run,
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
	flags.BoolVar(
		&args.showAllClusters,
		"show-all-clusters",
		false,
		"Show also clusters without defined upgrade policy.",
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
	organization, clusters, blockedVersions, sectors, err := be.Status(args.organizationId, args.showAllClusters)
	if err != nil {
		return err
	}
	blockedVersionExpressions, err := blockedversions.ParsedBlockedVersionExpressions(blockedVersions)
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

		w.WriteString("Sector Configuration:\t(%d in total)\n", len(sectors))
		if len(sectors) > 0 {
			w1.WriteString("Name\tDepends on\n")
			w1.WriteString("----\t----------\n")
			for _, sector := range sectors {
				w1.WriteString("%s\t%s\n", sector.Name, strings.Join(sector.Dependencies, ", "))
			}
		}

		w.WriteString("Clusters:\t(%d in total)\n", len(clusters))
		if len(clusters) > 0 {
			w1.WriteString("Cluster Name\tProduct\tVersion\tChannel\tAUS enabled\tSchedule\tSector\tMutexes\tSoak Days\tWorkloads\tBlocked Versions\tAvailable Upgrades\n")
			w1.WriteString("------------\t-------\t-------\t-------\t-----------\t--------\t------\t-------\t---------\t---------\t----------------\t------------------\n")
			for _, cluster := range clusters {
				mutexes := "<none>"
				sector := "<none>"
				if cluster.Policy.Validate() == nil {
					if len(cluster.Policy.Conditions.Mutexes) > 0 {
						mutexes = strings.Join(cluster.Policy.Conditions.Mutexes, ", ")
					}
					if cluster.Policy.Conditions.Sector != "" {
						sector = cluster.Policy.Conditions.Sector
					}
					w1.WriteString("%s\t%s\t%s\t%s\t%t\t%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
						cluster.Cluster.Name(),
						cluster.Cluster.Product().ID(),
						cluster.Cluster.Version().RawID(),
						cluster.Cluster.Version().ChannelGroup(),
						true,
						cluster.Policy.Schedule,
						sector,
						mutexes,
						cluster.Policy.Conditions.SoakDays,
						strings.Join(cluster.Policy.Workloads, ", "),
						strings.Join(cluster.Policy.Conditions.BlockedVersions, ", "),
						strings.Join(cluster.AvailableUpgrades(blockedVersionExpressions), ", "),
					)
				} else {
					w1.WriteString("%s\t%s\t%s\t%s\t%t\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						cluster.Cluster.Name(),
						cluster.Cluster.Product().ID(),
						cluster.Cluster.Version().RawID(),
						cluster.Cluster.Version().ChannelGroup(),
						false,
						"<none>",
						sector,
						mutexes,
						"<none>",
						"<none>",
						strings.Join(cluster.Policy.Conditions.BlockedVersions, ", "),
						strings.Join(cluster.AvailableUpgrades(blockedVersionExpressions), ", "),
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
