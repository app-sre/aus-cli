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
	"fmt"
	"os"
	"strings"

	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/ocm"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/output"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/sectors"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/utils"
)

func (f *OCMLabelsPolicyBackend) ListSectorConfiguration(organizationId string) ([]sectors.SectorDependencies, error) {
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
	return listSectorConfigurationFromOrganizationLabels(organizationId, connection)
}

func (f *OCMLabelsPolicyBackend) ApplySectorConfiguration(organizationId string, sectorDependencies []sectors.SectorDependencies, dumpSectorDeps bool, dryRun bool) error {
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

	if dumpSectorDeps {
		body, err := json.Marshal(sectorDependencies)
		if err != nil {
			return err
		}
		err = output.PrettyList(os.Stdout, body)
		if err != nil {
			return err
		}
		return nil
	}

	output.Log(dryRun, "Apply sector configuration to organization %s\n", organizationId)

	sectorLabels, err := listOrganizationSectorLabels(organizationId, connection)
	if err != nil {
		return err
	}
	labelsContainer := NewOCMLabelsContainer(sectorLabels)

	for _, sector := range sectorDependencies {
		l, err := sectorDependencyToLabels(sector, organizationId)
		if err != nil {
			return err
		}
		labelsContainer.AddLabel(l)
	}

	return labelsContainer.Reconcile(dryRun, connection)
}

func listOrganizationSectorLabels(organizationId string, connection *sdk.Connection) ([]*amv1.Label, error) {
	return listOrganizationLabels(organizationId, newAusLabelKey("sector-deps."), connection)
}

func listSectorConfigurationFromOrganizationLabels(organizationId string, connection *sdk.Connection) ([]sectors.SectorDependencies, error) {
	labels, err := listOrganizationLabels(organizationId, newAusLabelKey("sector-deps."), connection)
	if err != nil {
		return nil, err
	}
	sectorDeps := []sectors.SectorDependencies{}
	for _, sectorLabel := range labels {
		sectorName := strings.TrimPrefix(sectorLabel.Key(), newAusLabelKey("sector-deps."))
		sectorDeps = append(sectorDeps, sectors.SectorDependencies{Name: sectorName, Dependencies: strings.Split(sectorLabel.Value(), ",")})
	}
	return sectorDeps, nil
}

func sectorDependencyToLabels(sectorDep sectors.SectorDependencies, organizationId string) (*amv1.Label, error) {
	return buildOCMLabel(
		newAusLabelKey(fmt.Sprintf("sector-deps.%s", sectorDep.Name)), utils.StringArrayToCSV(sectorDep.Dependencies), "", organizationId,
	)
}
