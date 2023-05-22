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

package blockedversions

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/backend"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/blockedversions"
)

var args struct {
	organizationId     string
	blockExpressions   []string
	unblockExpressions []string
	replace            bool

	dryRun bool
	dump   bool
}

var Cmd = &cobra.Command{
	Use:   "version-blocks",
	Short: "Create or update the blocked versions for an organization",
	Long: "Create or update the blocked versions for an organization.\n" +
		"\n" +
		"The blocked versions are either defined by flags or are read from stdin in the - arg is present. \n" +
		"If - is present, --block-version and --unblock-version will be ignored.\n" +
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
		&args.blockExpressions,
		"block-version",
		"b",
		[]string{},
		"",
	)
	flags.StringArrayVarP(
		&args.unblockExpressions,
		"unblock-version",
		"u",
		[]string{},
		"",
	)
	flags.BoolVar(
		&args.replace,
		"replace",
		false,
		"Replaced all version blocks on the organization with the provided versions. "+
			"Otherwise, the provided versions will be added to the existing blocked versions.",
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
		"Dumps the blocked version configuration to stdout and exits without applying it.",
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

	var blocking, unblocking []string
	if len(argv) > 0 && argv[0] == "-" {
		blocking, err = blockedversions.ReadVersionExpressionsFromReader(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("failed to decode input: %v", err)
		}
	} else {
		if len(args.blockExpressions) == 0 && len(args.unblockExpressions) == 0 {
			return errors.New("none of block-version or unblock-version flags were provided")
		}
		blocking = args.blockExpressions
		unblocking = args.unblockExpressions
	}

	// consolidate version blocks
	var currentVersionBlocks []string = []string{}
	if !args.replace {
		currentVersionBlocks, err = be.ListBlockedVersionExpressions(args.organizationId)
		if err != nil {
			return err
		}
	}
	blockExpressions := blockedversions.ConsolidateVersionBlocks(currentVersionBlocks, blocking, unblocking)
	return be.ApplyBlockedVersionExpressions(args.organizationId, blockExpressions, args.dump, args.dryRun)
}
