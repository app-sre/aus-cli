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

package ocmlabels

import (
	"encoding/json"
	"os"
	"strings"

	sdk "github.com/openshift-online/ocm-sdk-go"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/blockedversions"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/ocm"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/output"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/utils"
)

func (f *OCMLabelsPolicyBackend) ListBlockedVersionExpressions(organizationId string) ([]string, error) {
	connection, err := ocm.NewOCMConnection()
	if err != nil {
		return nil, err
	}

	if organizationId == "" {
		organizationId, err = ocm.CurrentOrganizationId(connection)
		if err != nil {
			return nil, err
		}
	}

	return getBlockedVersionsForOrganization(organizationId, connection)
}

func (f *OCMLabelsPolicyBackend) ApplyBlockedVersionExpressions(organizationId string, blockExpressions []string, dumpVersionBlocks bool, dryRun bool) error {
	connection, err := ocm.NewOCMConnection()
	if err != nil {
		return err
	}

	if organizationId == "" {
		organizationId, err = ocm.CurrentOrganizationId(connection)
		if err != nil {
			return err
		}
	}

	if dumpVersionBlocks {
		body, err := json.Marshal(blockExpressions)
		if err != nil {
			return err
		}
		err = output.PrettyList(os.Stdout, body)
		if err != nil {
			return err
		}
		return nil
	}

	output.Log(dryRun, "Apply blocked version labels to organization %s\n", organizationId)
	label, err := buildOCMLabel(newAusLabelKey("blocked-versions"), utils.StringArrayToCSV(blockExpressions), "", organizationId)
	if err != nil {
		return err
	}
	err = applyOCMLabel(label, dryRun, connection)
	if err != nil {
		return err
	}
	return nil
}

func getBlockedVersionsForOrganization(organizationId string, connection *sdk.Connection) ([]string, error) {
	label, err := getOrganizationLabel(organizationId, newAusLabelKey("blocked-versions"), connection)
	if err != nil {
		return nil, err
	}
	if label == nil {
		return []string{}, nil
	}
	return blockedversions.SortVersionExpressions(strings.Split(label.Value(), ",")), nil
}
