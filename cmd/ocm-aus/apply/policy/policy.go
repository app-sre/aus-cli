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
	"fmt"
	"strings"

	"github.com/app-sre/aus-cli/pkg/backend"
	"github.com/app-sre/aus-cli/pkg/policy"
	"github.com/app-sre/aus-cli/pkg/schedule"
	"github.com/spf13/cobra"
)

var args struct {
	organizationId string
	clusterName    string
	clusterUUID    string
	schedule       string
	workloads      []string
	soakDays       int
	sector         string
	mutexes        []string

	dryRun bool
	dump   bool
}

var Cmd = &cobra.Command{
	Use:   "policies [flags] [-]",
	Short: "Create or update a cluster upgrade policy",
	Long: "Create or update a cluster upgrade policy.\n" +
		"\n" +
		"The policy is either defined by flags or is read from stdin if the - arg is present. \n" +
		"To learn about the stdin format, run this command with flags and use --dump.\n",
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
	schedulePresets := strings.Join(schedule.SupportedSchedulePresets(), ", ")
	flags.StringVarP(
		&args.schedule,
		"schedule",
		"s",
		"",
		fmt.Sprintf("A cron expression that defines when the cluster should be upgraded or a schedule preset (%s)", schedulePresets),
	)
	flags.StringArrayVarP(
		&args.workloads,
		"workload",
		"w",
		[]string{},
		"An identifier for the workload that runs on a cluster.",
	)
	flags.IntVarP(
		&args.soakDays,
		"soak-days",
		"d",
		0,
		"The number of days to wait before upgrading the cluster. Defaults to 0.",
	)
	flags.StringVar(
		&args.sector,
		"sector",
		"",
		"The sector the cluster belongs to.",
	)
	flags.StringArrayVarP(
		&args.mutexes,
		"mutex",
		"m",
		[]string{},
		"The mutexes the cluster must hold before it can start an upgrade.",
	)
	flags.BoolVar(
		&args.dryRun,
		"dry-run",
		false,
		"If dry-run is specified, the generated policy will only be printed to stdout.",
	)
	flags.BoolVar(
		&args.dump,
		"dump",
		false,
		"Dumps the policy configuration to stdout and exits without applying it.",
	)
}

func run(cmd *cobra.Command, argv []string) error {
	var policies []policy.ClusterUpgradePolicy
	var err error
	if len(argv) > 0 && argv[0] == "-" {
		policies, err = policy.NewClusterUpgradePolicyFromReader(cmd.InOrStdin())

		if err != nil {
			return fmt.Errorf("failed to decode input: %v", err)
		}
	} else {
		schedule, err := schedule.TranslateSchedule(args.schedule)
		if err != nil {
			return err
		}
		policies = []policy.ClusterUpgradePolicy{
			policy.NewClusterUpgradePolicy(
				args.clusterName,
				schedule,
				args.workloads,
				args.soakDays,
				args.sector,
				args.mutexes,
			),
		}
	}
	for _, pol := range policies {
		err = pol.Validate()
		if err != nil {
			return fmt.Errorf("invalid policy: %v", err)
		}
	}

	backendType, err := cmd.Flags().GetString("backend")
	if err != nil {
		return err
	}
	be, err := backend.NewPolicyBackend(backendType)
	if err != nil {
		return err
	}
	err = be.ApplyPolicies(args.organizationId, policies, args.dump, args.dryRun)
	return err
}

// todo verify that at least one cluster has 0 soak days
