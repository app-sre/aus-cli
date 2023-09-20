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

package inheritance

import (
	"fmt"

	"github.com/app-sre/aus-cli/pkg/backend"
	"github.com/app-sre/aus-cli/pkg/versiondata"
	"github.com/spf13/cobra"
)

var args struct {
	organizationId string
	inherit        []string
	publish        []string
	replace        bool

	dryRun bool
	dump   bool
}

var Cmd = &cobra.Command{
	Use:   "inheritance",
	Short: "Create or update the cross-organization version data inheritance",
	Long: "Create or update the cross-organization version data inheritance.\n" +
		"\n" +
		"The configuration is either defined by flags or are read from stdin in the - arg is present. \n" +
		"If - is present, --inherit-from and --publish-to will be ignored.\n" +
		"To learn about the stdin format, run this command with flags and use --dump.\n",
	RunE: run,
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(
		&args.organizationId,
		"org-id",
		"o",
		"",
		"The ID of the OCM organization to manage",
	)

	flags.StringArrayVarP(
		&args.inherit,
		"inherit-from",
		"i",
		[]string{},
		"A comma-separated list of organization IDs to inherit version data from. The listed organizations "+
			"need to define a matching publish-to entry in their configuration for this to work. Otherwise AUS "+
			"will not trigger any updates for the inheriting organization and will publish a service log "+
			"on all affected clusters. The referenced organizations can be located in different OCM environments.",
	)
	flags.StringArrayVarP(
		&args.publish,
		"publish-to",
		"p",
		[]string{},
		"A comma-separated list of organization IDs to publish version data to. The listed organizations "+
			"need to define a matching inherit-from entry in their configuration for this to work. The "+
			"referenced organizations can be located in different OCM environments.",
	)
	flags.BoolVar(
		&args.replace,
		"replace",
		false,
		"Replaced the inheritance configuration on the organization with the provided configuration. "+
			"Otherwise, the provided organization IDs will be added to the existing configuration.",
	)
	flags.BoolVar(
		&args.dryRun,
		"dry-run",
		false,
		"",
	)
	flags.BoolVar(
		&args.dump,
		"dump",
		false,
		"Dumps the inheritance configuration to stdout and exits without applying it.",
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

	var config versiondata.VersionDataInheritanceConfig
	if len(argv) > 0 && argv[0] == "-" {
		config, err = versiondata.NewVersionDataInheritanceConfigFromReader(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("failed to decode input: %v", err)
		}
	} else {
		config = versiondata.VersionDataInheritanceConfig{
			InheritingFromOrgs: args.inherit,
			PublishingToOrgs:   args.publish,
		}
	}

	// consolidate configs
	var currentConfig versiondata.VersionDataInheritanceConfig
	if !args.replace {
		currentConfig, err = be.GetVersionDataInheritanceConfiguration(args.organizationId)
		if err != nil {
			return err
		}
	}
	consolidatedConfig := versiondata.ConsolidateVersionDataInheritanceConfig(currentConfig, config)
	return be.ApplyVersionDataInheritanceConfiguration(args.organizationId, consolidatedConfig, args.dump, args.dryRun)
}
