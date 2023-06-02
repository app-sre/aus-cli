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
	"github.com/app-sre/aus-cli/cmd/ocm-aus/apply"
	"github.com/app-sre/aus-cli/cmd/ocm-aus/delete"
	"github.com/app-sre/aus-cli/cmd/ocm-aus/get"
	"github.com/app-sre/aus-cli/cmd/ocm-aus/status"
	"github.com/app-sre/aus-cli/cmd/ocm-aus/version"
	"github.com/app-sre/aus-cli/pkg/arguments"
	"github.com/spf13/cobra"
	"os"
)

var root = &cobra.Command{
	Use:           "ocm aus",
	Short:         "AUS plug-in for the ocm-cli",
	Long:          "This plug-in extends the ocm-cli to provide additional commands for working with Advanced Upgrade Service",
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

	root.PersistentFlags().String("backend", "ocmlabels", "Backend to store policies in. Supported: ocmlabels")

	// Register the subcommands:
	root.AddGroup(&cobra.Group{ID: "AUS commands", Title: "AUS commands:"})
	root.AddCommand(get.Cmd)
	root.AddCommand(apply.Cmd)
	root.AddCommand(status.Cmd)
	root.AddCommand(delete.Cmd)
	root.AddCommand(version.Cmd)
}

func main() {
	// This is needed to make `glog` believe that the flags have already been parsed, otherwise
	// every log messages is prefixed by an error message stating the flags haven't been
	// parsed.
	err := flag.CommandLine.Parse([]string{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't parse empty command line to satisfy 'glog': %v\n", err)
		os.Exit(1)
	}

	// Execute the root command and exit immediately if there was no error:
	root.SetArgs(os.Args[1:])
	err = root.Execute()
	if err == nil {
		os.Exit(0)
	}

	message := fmt.Sprintf("Error: %s", err.Error())
	fmt.Fprintf(os.Stderr, "%s\n", message)
	os.Exit(1)
}
