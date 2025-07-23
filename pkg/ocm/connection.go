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

package ocm

import (
	"fmt"

	"github.com/openshift-online/ocm-cli/pkg/config"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

func NewOCMConnection() (*sdk.Connection, error) {
	// Load the configuration file:
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("can't load config file: %v", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("not logged in, run the 'ocm login' command")
	}

	// Check that the configuration has credentials or tokens that don't have expired:
	armed, reason, err := cfg.Armed()
	if err != nil {
		return nil, err
	}
	if !armed {
		return nil, fmt.Errorf("not logged in, %s, run the 'ocm login' command", reason)
	}

	// Create the connection:
	connection, err := cfg.Connection()
	if err != nil {
		return nil, fmt.Errorf("can't create connection: %v", err)
	}

	return connection, nil
}
