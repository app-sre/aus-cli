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
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/ocm"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/policy"
	"gitlab.cee.redhat.com/service/aus-cli/pkg/sectors"
)

func (f *OCMLabelsPolicyBackend) Status(organizationId string) (organization *amv1.Organization, policies []policy.ClusterUpgradePolicy, blockedVersions []string, sectors []sectors.SectorDependencies, err error) {
	connection, err := ocm.NewOCMConnection()
	if err != nil {
		return
	}
	organization, err = ocm.GetOrganization(organizationId, connection)
	if err != nil {
		return
	}

	blockedVersions, err = getBlockedVersionsForOrganization(organization.ID(), connection)
	if err != nil {
		return
	}
	policiesMap, err := listPoliciesInOrganization(organization.ID(), true, connection)
	if err != nil {
		return
	}
	policies = []policy.ClusterUpgradePolicy{}
	for _, p := range policiesMap {
		policies = append(policies, p)
	}
	policy.SortPolicies(policies)
	sectors, err = listSectorConfigurationFromOrganizationLabels(organization.ID(), connection)
	return
}
