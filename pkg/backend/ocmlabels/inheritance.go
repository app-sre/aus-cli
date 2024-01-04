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

	"github.com/app-sre/aus-cli/pkg/ocm"
	"github.com/app-sre/aus-cli/pkg/output"
	"github.com/app-sre/aus-cli/pkg/utils"
	"github.com/app-sre/aus-cli/pkg/versiondata"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

var INHERIT_LABEL_KEY = newAusLabelKey("version-data.inherit")
var PUBLISH_LABEL_KEY = newAusLabelKey("version-data.publish")

func (f *OCMLabelsPolicyBackend) GetVersionDataInheritanceConfiguration(organizationId string) (versiondata.VersionDataInheritanceConfig, error) {
	connection, err := ocm.NewOCMConnection()
	if err != nil {
		return versiondata.VersionDataInheritanceConfig{}, err
	}

	if organizationId == "" {
		organizationId, err = ocm.CurrentOrganizationId(connection)
		if err != nil {
			return versiondata.VersionDataInheritanceConfig{}, err
		}
	}
	return listVersionDataInheritanceConfiguration(organizationId, connection)
}

func (f *OCMLabelsPolicyBackend) ApplyVersionDataInheritanceConfiguration(organizationId string, inheritance versiondata.VersionDataInheritanceConfig, dumpConfig bool, dryRun bool) error {
	if dumpConfig {
		body, err := json.Marshal(inheritance)
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

	output.Log(dryRun, "Apply version data inheritance configuration to organization %s\n", organizationId)

	labels, err := listOrganizationLabels(organizationId, newAusLabelKey("version-data."), connection)
	if err != nil {
		return err
	}
	labelsContainer := NewOCMLabelsContainer(labels)

	if len(inheritance.InheritingFromOrgs) > 0 {
		inheritLabel, err := buildOCMLabel(
			INHERIT_LABEL_KEY, utils.StringArrayToCSV(inheritance.InheritingFromOrgs), "", organizationId,
		)
		if err != nil {
			return err
		}
		labelsContainer.AddLabel(inheritLabel)
	}

	if len(inheritance.PublishingToOrgs) > 0 {
		publishLabel, err := buildOCMLabel(
			PUBLISH_LABEL_KEY, utils.StringArrayToCSV(inheritance.PublishingToOrgs), "", organizationId,
		)
		if err != nil {
			return err
		}
		labelsContainer.AddLabel(publishLabel)
	}

	return labelsContainer.Reconcile(dryRun, connection)
}

func listVersionDataInheritanceConfiguration(organizationId string, connection *sdk.Connection) (versiondata.VersionDataInheritanceConfig, error) {
	labels, err := listOrganizationLabels(organizationId, newAusLabelKey("version-data."), connection)
	if err != nil {
		return versiondata.VersionDataInheritanceConfig{}, err
	}
	inheritOrgIds := []string{}
	publishOrgIds := []string{}
	for _, versionDataLabel := range labels {
		if versionDataLabel.Key() == INHERIT_LABEL_KEY {
			inheritOrgIds = strings.Split(versionDataLabel.Value(), ",")
		}
		if versionDataLabel.Key() == PUBLISH_LABEL_KEY {
			publishOrgIds = strings.Split(versionDataLabel.Value(), ",")
		}
	}
	return versiondata.VersionDataInheritanceConfig{
		InheritingFromOrgs: inheritOrgIds,
		PublishingToOrgs:   publishOrgIds,
	}, nil
}
