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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.cee.redhat.com/service/aus-cli/cmd/aus/apply"
	"gitlab.cee.redhat.com/service/aus-cli/cmd/aus/describe"
	"gitlab.cee.redhat.com/service/aus-cli/cmd/aus/get"
	"gitlab.cee.redhat.com/service/aus-cli/cmd/aus/login"
	"gitlab.cee.redhat.com/service/aus-cli/cmd/aus/version"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/arguments"
)

var root = &cobra.Command{
	Use:           "aus",
	Long:          "Command line tool for Advanced Upgrade Service (AUS).",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Send logs to the standard error stream by default:
	err := flag.Set("logtostderr", "true")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't set default error stream: %v\n", err)
		os.Exit(1)
	}

	// Add the command line flags:
	fs := root.PersistentFlags()
	arguments.AddDebugFlag(fs)

	root.PersistentFlags().String("backend", "ocmlabels", "Backend to store policies in. Supported: ocmlabels (requires OCM authentication)")

	// Register the subcommands:
	root.AddCommand(login.Cmd)
	root.AddCommand(get.Cmd)
	root.AddCommand(apply.Cmd)
	root.AddCommand(describe.Cmd)
	root.AddCommand(version.Cmd)
}

func main() {
	// This is needed to make `glog` believe that the flags have already been parsed, otherwise
	// every log messages is prefixed by an error message stating the the flags haven't been
	// parsed.
	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't parse empty command line to satisfy 'glog': %v\n", err)
		os.Exit(1)
	}

	// Execute the root command and exit inmediately if there was no error:
	root.SetArgs(os.Args[1:])
	err = root.Execute()
	if err == nil {
		os.Exit(0)
	}

	message := fmt.Sprintf("Error: %s", err.Error())
	fmt.Fprintf(os.Stderr, "%s\n", message)

	// Exit signaling an error:
	os.Exit(1)
}
