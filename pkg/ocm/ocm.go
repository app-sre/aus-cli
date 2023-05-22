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
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
)

func NewOCMConnection() (*sdk.Connection, error) {
	// Load the configuration file:
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("Can't load config file: %v", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("Not logged in, run the 'ocm login' command")
	}

	// Check that the configuration has credentials or tokens that don't have expired:
	armed, reason, err := cfg.Armed()
	if err != nil {
		return nil, err
	}
	if !armed {
		return nil, fmt.Errorf("Not logged in, %s, run the 'ocm login' command", reason)
	}

	// Create the connection:
	connection, err := cfg.Connection()
	if err != nil {
		return nil, fmt.Errorf("Can't create connection: %v", err)
	}

	return connection, nil
}

func Whoami(connection *sdk.Connection) (*amv1.Account, error) {
	response, err := connection.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		return nil, err
	}
	return response.Body(), nil
}

func CurrentOrganizationId(connection *sdk.Connection) (string, error) {
	account, err := Whoami(connection)
	if err != nil {
		return "", err
	}
	return account.Organization().ID(), nil
}

func GetOrganization(organizationId string, connection *sdk.Connection) (*amv1.Organization, error) {
	var err error
	if organizationId == "" {
		organizationId, err = CurrentOrganizationId(connection)
		if err != nil {
			return nil, err
		}
	}
	response, err := connection.AccountsMgmt().V1().Organizations().Organization(organizationId).Get().Send()
	if err != nil {
		return nil, err
	}
	return response.Body(), nil
}
