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

	"github.com/app-sre/aus-cli/pkg/ocm"
	"github.com/app-sre/aus-cli/pkg/output"
	"github.com/app-sre/aus-cli/pkg/sectors"
	"github.com/app-sre/aus-cli/pkg/utils"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
)

func (f *OCMLabelsPolicyBackend) ListSectorConfiguration(organizationId string) ([]sectors.Sector, error) {
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
	return listSectorsFromOrganizationLabels(organizationId, connection)
}

func (f *OCMLabelsPolicyBackend) ApplySectorConfiguration(organizationId string, sectors []sectors.Sector, dumpSectors bool, dryRun bool) error {
	if dumpSectors {
		body, err := json.Marshal(sectors)
		if err != nil {
			return err
		}
		err = output.PrettyList(os.Stdout, body)
		if err != nil {
			return err
		}
		return nil
	}

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

	output.Log(dryRun, "Apply sector configuration to organization %s\n", organizationId)

	sectorLabels, err := listOrganizationSectorLabels(organizationId, connection)
	if err != nil {
		return err
	}
	labelsContainer := NewOCMLabelsContainer(sectorLabels)

	for _, sector := range sectors {
		if len(sector.Dependencies) != 0 {
			l, err := sectorDependencyToLabels(sector, organizationId)
			if err != nil {
				return err
			}
			labelsContainer.AddLabel(l)
		}
		if sector.MaxParallelUpgrades != "" {
			l, err := sectorMaxParallelUpgradesToLabels(sector, organizationId)
			if err != nil {
				return err
			}
			labelsContainer.AddLabel(l)
		}
	}

	return labelsContainer.Reconcile(dryRun, connection)
}

func listOrganizationSectorDependenciesLabels(organizationId string, connection *sdk.Connection) ([]*amv1.Label, error) {
	return listOrganizationLabels(organizationId, newAusLabelKey("sector-deps."), connection)
}

func listOrganizationSectorMaxParallelUpgradesLabels(organizationId string, connection *sdk.Connection) ([]*amv1.Label, error) {
	return listOrganizationLabels(organizationId, newAusLabelKey("sector-max-parallel-upgrades."), connection)
}

func listOrganizationSectorLabels(organizationId string, connection *sdk.Connection) ([]*amv1.Label, error) {
	sectorDependenciesLabels, err := listOrganizationSectorDependenciesLabels(organizationId, connection)
	if err != nil {
		return nil, err
	}
	sectorMaxParallelUpgradesLabels, err := listOrganizationSectorMaxParallelUpgradesLabels(organizationId, connection)
	if err != nil {
		return nil, err
	}
	labels := make([]*amv1.Label, 0, len(sectorDependenciesLabels)+len(sectorMaxParallelUpgradesLabels))
	labels = append(labels, sectorDependenciesLabels...)
	labels = append(labels, sectorMaxParallelUpgradesLabels...)
	return labels, nil
}

func addOrUpdateSector(sectorMap map[string]sectors.Sector, sector sectors.Sector) {
	if existingSector, exists := sectorMap[sector.Name]; exists {
		if sector.Dependencies != nil {
			existingSector.Dependencies = sector.Dependencies
		}
		if sector.MaxParallelUpgrades != "" {
			existingSector.MaxParallelUpgrades = sector.MaxParallelUpgrades
		}
	} else {
		sectorMap[sector.Name] = sector
	}
}

func listSectorsFromOrganizationLabels(organizationId string, connection *sdk.Connection) ([]sectors.Sector, error) {
	sectorMap := make(map[string]sectors.Sector)

	labels, err := listOrganizationSectorDependenciesLabels(organizationId, connection)
	if err != nil {
		return nil, err
	}
	for _, label := range labels {
		sectorName := strings.TrimPrefix(label.Key(), newAusLabelKey("sector-deps."))
		sector := sectors.Sector{Name: sectorName, Dependencies: strings.Split(label.Value(), ",")}
		addOrUpdateSector(sectorMap, sector)
	}

	labels, err = listOrganizationSectorMaxParallelUpgradesLabels(organizationId, connection)
	if err != nil {
		return nil, err
	}
	for _, label := range labels {
		sectorName := strings.TrimPrefix(label.Key(), newAusLabelKey("sector-max-parallel-upgrades."))
		sector := sectors.Sector{Name: sectorName, MaxParallelUpgrades: label.Value()}
		addOrUpdateSector(sectorMap, sector)
	}

	sectors := make([]sectors.Sector, 0, len(sectorMap))
	for _, sector := range sectorMap {
		sectors = append(sectors, sector)
	}

	return sectors, nil
}

func sectorDependencyToLabels(sector sectors.Sector, organizationId string) (*amv1.Label, error) {
	return buildOCMLabel(
		newAusLabelKey(fmt.Sprintf("sector-deps.%s", sector.Name)), utils.StringArrayToCSV(sector.Dependencies), "", organizationId,
	)
}

func sectorMaxParallelUpgradesToLabels(sector sectors.Sector, organizationId string) (*amv1.Label, error) {
	return buildOCMLabel(
		newAusLabelKey(fmt.Sprintf("sector-max-parallel-upgrades.%s", sector.Name)), sector.MaxParallelUpgrades, "", organizationId,
	)
}
