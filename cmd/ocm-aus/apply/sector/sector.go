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
	"fmt"

	"github.com/app-sre/aus-cli/pkg/backend"
	"github.com/app-sre/aus-cli/pkg/sectors"
	"github.com/spf13/cobra"
)

var args struct {
	organizationId            string
	sectorMaxParallelUpgrades []string
	add                       []string
	remove                    []string
	replace                   bool
	dryRun                    bool
	dump                      bool
}

var Cmd = &cobra.Command{
	Use:   "sectors",
	Short: "Create or update the sector dependencies for an organization",
	Long: "Create or update the sector dependencies for an organization.\n" +
		"\n" +
		"The sector dependencies are either defined by flags or are read from stdin if the - arg is present. \n" +
		"If - is present, --add-dep and --remove-dep will be ignored.\n" +
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
		"The ID of the OCM organization to manage",
	)

	flags.StringArrayVarP(
		&args.sectorMaxParallelUpgrades,
		"sector-max-parallel-upgrades",
		"m",
		[]string{},
		"",
	)
	flags.StringArrayVarP(
		&args.add,
		"add-dep",
		"a",
		[]string{},
		"",
	)
	flags.StringArrayVarP(
		&args.remove,
		"remove-dep",
		"r",
		[]string{},
		"",
	)
	flags.BoolVar(
		&args.replace,
		"replace",
		false,
		"Replaced all sector dependencies on the organization with the provided ones. "+
			"Otherwise, the provided dependencies will be added to the existing ones.",
	)
	flags.BoolVar(
		&args.dump,
		"dump",
		false,
		"Dumps the sector configuration to stdout and exits without applying it.",
	)
	flags.BoolVar(
		&args.dryRun,
		"dry-run",
		false,
		"",
	)
}

func run(cmd *cobra.Command, argv []string) error {
	var sectorList []sectors.Sector
	var err error

	backendType, err := cmd.Flags().GetString("backend")
	if err != nil {
		return err
	}
	be, err := backend.NewPolicyBackend(backendType)
	if err != nil {
		return err
	}

	var sectorsMaxParallelUpgrades, adding, removing []sectors.Sector
	if len(argv) > 0 && argv[0] == "-" {
		adding, err = sectors.ReadSectorsFromReader(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("failed to decode input: %v", err)
		}
	} else {
		sectorsMaxParallelUpgrades, err = sectors.NewSectorMaxParallelUpgradesList(args.sectorMaxParallelUpgrades)
		if err != nil {
			return err
		}

		adding, err = sectors.NewSectorDependenciesList(args.add)
		if err != nil {
			return err
		}

		removing, err = sectors.NewSectorDependenciesList(args.remove)
		if err != nil {
			return err
		}
	}

	// consolidate dependencies
	var currentSectors []sectors.Sector = []sectors.Sector{}
	if !args.replace {
		currentSectors, err = be.ListSectorConfiguration(args.organizationId)
		if err != nil {
			return err
		}
	}
	sectorList = sectors.ConsolidateSectorList(
		currentSectors, adding, removing, sectorsMaxParallelUpgrades,
	)

	err = be.ApplySectorConfiguration(args.organizationId, sectorList, args.dump, args.dryRun)
	if err != nil {
		return err
	}

	return nil
}
